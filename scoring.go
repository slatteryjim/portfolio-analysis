package portfolio_analysis

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
)

var (
	// TODO: use an enum for the assets, so it's just an int under the covers, but has a nice String method,
	//  and maybe even a Returns() method that returns the appropriate []float64?
	assetMap = map[string][]float64{
		"TSM": TSM,
		"SCV": SCV,
		"LTT": LTT,
		"STT": STT,
		"GLD": GLD,
	}
)

type (
	PortfolioStat struct {
		// describe portfolio assets and percentages
		Assets      []string
		Percentages []float64

		// stats on the portfolio performance
		PWR20                float64
		PWR30                float64
		UlcerScore           float64
		DeepestDrawdown      float64
		LongestDrawdown      int
		StartDateSensitivity float64

		// This portfolio's rank on various stats
		PWR20Rank                Rank
		PWR30Rank                Rank
		UlcerScoreRank           Rank
		DeepestDrawdownRank      Rank
		LongestDrawdownRank      Rank
		StartDateSensitivityRank Rank
	}

	Rank struct {
		Ordinal    int
		Percentage float64
	}
)

func (p PortfolioStat) String() string {
	return fmt.Sprintf("%v %v PWR20:%0.3f%%(%d) PWR30:%0.3f%%(%d) Ulcer:%0.1f(%d) DeepestDrawdown:%0.2f%%(%d) LongestDrawdown:%d(%d), StartDateSensitivity:%0.2f%%(%d)",
		p.Assets,
		p.Percentages,
		p.PWR20*100,
		p.PWR20Rank.Ordinal,
		p.PWR30*100,
		p.PWR30Rank.Ordinal,
		p.UlcerScore,
		p.UlcerScoreRank.Ordinal,
		p.DeepestDrawdown*100,
		p.DeepestDrawdownRank.Ordinal,
		p.LongestDrawdown,
		p.LongestDrawdownRank.Ordinal,
		p.StartDateSensitivity*100,
		p.StartDateSensitivityRank.Ordinal,
	)
}

func (p PortfolioStat) ComparePerformance(other PortfolioStat) PortfolioStat {
	copied := *p.Clone()
	copied.PWR20 -= other.PWR20
	copied.PWR30 -= other.PWR30
	copied.UlcerScore -= other.UlcerScore
	copied.DeepestDrawdown -= other.DeepestDrawdown
	copied.LongestDrawdown -= other.LongestDrawdown
	copied.StartDateSensitivity -= other.StartDateSensitivity
	return copied
}

// Clone returns a deep copy.
func (p PortfolioStat) Clone() *PortfolioStat {
	// deep copy the slices
	assets := make([]string, len(p.Assets))
	copy(assets, p.Assets)
	percentages := make([]float64, len(p.Percentages))
	copy(percentages, p.Percentages)

	return &PortfolioStat{
		Assets:                   assets,
		Percentages:              percentages,
		PWR20:                    p.PWR20,
		PWR30:                    p.PWR30,
		UlcerScore:               p.UlcerScore,
		DeepestDrawdown:          p.DeepestDrawdown,
		LongestDrawdown:          p.LongestDrawdown,
		StartDateSensitivity:     p.StartDateSensitivity,
		PWR20Rank:                p.PWR20Rank,
		PWR30Rank:                p.PWR30Rank,
		UlcerScoreRank:           p.UlcerScoreRank,
		DeepestDrawdownRank:      p.DeepestDrawdownRank,
		LongestDrawdownRank:      p.LongestDrawdownRank,
		StartDateSensitivityRank: p.StartDateSensitivityRank,
	}
}

func CopyAll(ps []*PortfolioStat) []*PortfolioStat {
	res := make([]*PortfolioStat, len(ps))
	for i, p := range ps {
		res[i] = p.Clone()
	}
	return res
}

// EvaluatePortfolios evaluates the portfolio for each of the given permutations, returning a slice of stats.
// It processes in parallel using multiple CPUs as needed.
func EvaluatePortfolios(perms []Permutation, assetMap map[string][]float64) ([]*PortfolioStat, error) {
	res := make([]*PortfolioStat, len(perms))
	var (
		wg sync.WaitGroup

		mu       sync.Mutex
		finalErr error
	)
	startIndex := 0
	for i, index := range segmentIndexes(len(perms), runtime.NumCPU()) {
		wg.Add(1)
		go func(startIndex, endIndex int) {
			defer wg.Done()
			// evaluate this portion of the perms
			stats, err := evaluatePortfolios(perms[startIndex:endIndex], assetMap)
			if err != nil {
				mu.Lock()
				mu.Unlock()
				finalErr = fmt.Errorf("error in segment %d, perms offset %d: %w", i+1, startIndex, err)
				return
			}
			// copy over the PortfolioStats to the appropriate part of the final slice
			for i := startIndex; i < endIndex; i++ {
				res[i] = stats[i-startIndex]
			}
		}(startIndex, index)
		startIndex = index // the last index is the next startIndex
	}
	wg.Wait()
	if finalErr != nil {
		return nil, finalErr
	}
	return res, nil
}

// evaluatePortfolios evaluates the portfolio for each of the given permutations, returning a slice of stats.
func evaluatePortfolios(perms []Permutation, assetMap map[string][]float64) ([]*PortfolioStat, error) {
	// define this array to be reused
	var returnsList [][]float64

	results := make([]*PortfolioStat, 0, len(perms))
	for i, p := range perms {
		// populate returnsList from p.Assets and assetMap
		{
			returnsList = returnsList[:0] // zero out cached returnsList
			for _, a := range p.Assets {
				returns, ok := assetMap[a]
				if !ok {
					return nil, fmt.Errorf("perm #%d, asset not found in map: %q", i+1, a)
				}
				returnsList = append(returnsList, returns)
			}
		}
		portfolioReturns, err := portfolioReturns(returnsList, p.Percentages)
		if err != nil {
			return nil, fmt.Errorf("perm #%d, error calculating portfolio returns for %+v: %w", i+1, p, err)
		}

		minPWR20, _ := minPWR(portfolioReturns, 20)
		minPWR30, _ := minPWR(portfolioReturns, 30)
		maxUlcerScore, deepestDrawdown, longestDrawdown := drawdownScores(portfolioReturns)

		results = append(results, &PortfolioStat{
			Assets:               p.Assets,
			Percentages:          p.Percentages,
			PWR20:                minPWR20,
			PWR30:                minPWR30,
			UlcerScore:           maxUlcerScore,
			DeepestDrawdown:      deepestDrawdown,
			LongestDrawdown:      longestDrawdown,
			StartDateSensitivity: startDateSensitivity(portfolioReturns),
		})
	}
	return results, nil
}

// RankPortfoliosInPlace this is a "destructive" operation, reordering the list and mutating the ***Rank fields.
// It sorts the list by the various performance metrics and populates the corresponding Rank field for each.
// It finishes up by sorting them by their "overall" rank, considering all of the performance metrics equally.
func RankPortfoliosInPlace(results []*PortfolioStat) {
	// Calculate rank scores for the portfolios
	{
		// rank by PWR20
		RankAll(results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.PWR20 },
			LessIsBetter: false,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.PWR20Rank = rank },
		})
		// rank by PWR30
		RankAll(results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.PWR30 },
			LessIsBetter: false,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.PWR30Rank = rank },
		})
		// rank by UlcerScore
		RankAll(results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.UlcerScore },
			LessIsBetter: true,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.UlcerScoreRank = rank },
		})
		// rank by DeepestDrawdown
		RankAll(results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.DeepestDrawdown },
			LessIsBetter: false,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.DeepestDrawdownRank = rank },
		})
		// rank by LongestDrawdown
		RankAll(results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return float64(stat.LongestDrawdown) },
			LessIsBetter: true,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.LongestDrawdownRank = rank },
		})
		// rank by StartDateSensitivity
		RankAll(results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.StartDateSensitivity },
			LessIsBetter: true,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.StartDateSensitivityRank = rank },
		})
	}

	// rank by all their ranks (equally weighted)
	{
		// with simply summing up the ranks:
		// #1: [TSM SCV LTT STT GLD] [30 5  5 40 20] PWR30: 3.958% (1483) Ulcer:2.4(316) DeepestDrawdown:-11.73%(199) LongestDrawdown:3(199)
		// #2: [TSM SCV LTT STT GLD] [25 10 5 40 20] PWR30: 4.035% (1259) Ulcer:2.5(427) DeepestDrawdown:-12.24%(263) LongestDrawdown:3(263)
		// #3: [TSM SCV LTT STT GLD] [25 5  5 45 20] PWR30: 3.862% (1762) Ulcer:2.1(205) DeepestDrawdown:-11.07%(128) LongestDrawdown:3(128)

		// with sum of (each rank^2)
		// #1: [TSM SCV LTT STT GLD] [35 5 5 30 25] PWR30: 4.232% (790) Ulcer:3.0(730) DeepestDrawdown:-13.32%(420) LongestDrawdown:3(420)
		// #2: [TSM SCV LTT STT GLD] [15 20 5 40 20] PWR30: 4.180% (894) Ulcer:2.9(658) DeepestDrawdown:-13.28%(412) LongestDrawdown:3(412)
		// #3: [TSM SCV LTT STT GLD] [30 5 10 30 25] PWR30: 4.142% (985) Ulcer:2.8(596) DeepestDrawdown:-13.13%(386) LongestDrawdown:3(386)
		sumRanks := func(p *PortfolioStat) float64 {
			return math.Pow(p.PWR20Rank.Percentage, 2) +
				math.Pow(p.PWR30Rank.Percentage, 2) +
				math.Pow(p.UlcerScoreRank.Percentage, 2) +
				math.Pow(p.LongestDrawdownRank.Percentage, 2) +
				math.Pow(p.DeepestDrawdownRank.Percentage, 2) +
				math.Pow(p.StartDateSensitivityRank.Percentage, 2)
		}
		sort.Slice(results, func(i, j int) bool {
			return sumRanks(results[i]) < sumRanks(results[j])
		})
	}
}

type RankAllParams struct {
	Metric       func(*PortfolioStat) float64
	LessIsBetter bool
	SetRank      func(stat *PortfolioStat, rank Rank)
}

func RankAll(
	results []*PortfolioStat,
	params RankAllParams,
) {
	if params.LessIsBetter {
		sort.Slice(results, func(i, j int) bool { return params.Metric(results[i]) < params.Metric(results[j]) })
	} else {
		sort.Slice(results, func(i, j int) bool { return params.Metric(results[i]) > params.Metric(results[j]) })
	}
	ranks := make([]int, len(results))
	var (
		rank      = 0
		lastValue = 0.0
	)
	for i, portfolioStat := range results {
		value := params.Metric(portfolioStat)
		if i == 0 || lastValue != value {
			rank++
			lastValue = value
		}
		ranks[i] = rank
	}
	maxRank := float64(rank)
	for i, portfolioStat := range results {
		rank := ranks[i]
		rankPercentage := float64(rank)/maxRank*99 + 1
		params.SetRank(portfolioStat, Rank{Ordinal: rank, Percentage: rankPercentage})
	}
}

func FindOne(results []*PortfolioStat, pred func(p *PortfolioStat) bool) *PortfolioStat {
	for _, p := range results {
		if pred(p) {
			return p
		}
	}
	return nil
}

func FindMany(results []*PortfolioStat, pred func(p *PortfolioStat) bool) []*PortfolioStat {
	var res []*PortfolioStat
	for _, p := range results {
		if pred(p) {
			res = append(res, p)
		}
	}
	return res
}

// segmentIndexes splits the `count` number of items into the given number of segments.
// It returns the array of indexes to refer to those segments.
func segmentIndexes(count, segments int) []int {
	if segments < 1 {
		panic(fmt.Sprintf("segments must be greater than zero but got %d", segments))
	}
	if count <= segments {
		res := make([]int, 0, count)
		for i := 1; i <= count; i++ {
			res = append(res, i)
		}
		return res
	}
	res := make([]int, 0, segments)
	step := float64(count) / float64(segments)
	for i := 1; i <= segments; i++ {
		res = append(res, int(float64(i)*step))
	}
	return res
}
