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
	assetMap = map[string][]Percent{
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
		Percentages []Percent

		// stats on the portfolio performance
		AvgReturn            Percent
		BaselineLTReturn     Percent
		BaselineSTReturn     Percent
		PWR30                Percent
		SWR30                Percent
		StdDev               Percent
		UlcerScore           float64
		DeepestDrawdown      Percent
		LongestDrawdown      int
		StartDateSensitivity Percent

		// This portfolio's rank on various stats
		AvgReturnRank            Rank
		BaselineLTReturnRank     Rank
		BaselineSTReturnRank     Rank
		PWR30Rank                Rank
		SWR30Rank                Rank
		StdDevRank               Rank
		UlcerScoreRank           Rank
		DeepestDrawdownRank      Rank
		LongestDrawdownRank      Rank
		StartDateSensitivityRank Rank

		// Score the rankings!
		OverallRankScore float64

		// rank based on overall rank! (头晕了)
		OverallRankScoreRank Rank
	}

	Rank struct {
		Ordinal    int
		Percentage float64
	}
)

func (p PortfolioStat) String() string {
	return fmt.Sprintf("%v %v (%d) AvgReturn:%0.3f%%(%d) BLT:%0.3f%%(%d) BST:%0.3f%%(%d) PWR:%0.3f%%(%d) SWR:%0.3f%%(%d) StdDev:%0.3f%%(%d) Ulcer:%0.1f(%d) DeepestDrawdown:%0.2f%%(%d) LongestDrawdown:%d(%d), StartDateSensitivity:%0.2f%%(%d)",
		p.Assets,
		p.Percentages,
		p.OverallRankScoreRank.Ordinal,
		p.AvgReturn*100,
		p.AvgReturnRank.Ordinal,
		p.BaselineLTReturn*100,
		p.BaselineLTReturnRank.Ordinal,
		p.BaselineSTReturn*100,
		p.BaselineSTReturnRank.Ordinal,
		p.PWR30*100,
		p.PWR30Rank.Ordinal,
		p.SWR30*100,
		p.SWR30Rank.Ordinal,
		p.StdDev*100,
		p.StdDevRank.Ordinal,
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
	copied.AvgReturn -= other.AvgReturn
	copied.BaselineLTReturn -= other.BaselineLTReturn
	copied.BaselineSTReturn -= other.BaselineSTReturn
	copied.PWR30 -= other.PWR30
	copied.SWR30 -= other.SWR30
	copied.StdDev -= other.StdDev
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
	percentages := make([]Percent, len(p.Percentages))
	copy(percentages, p.Percentages)

	return &PortfolioStat{
		Assets:                   assets,
		Percentages:              percentages,
		AvgReturn:                p.AvgReturn,
		BaselineLTReturn:         p.BaselineLTReturn,
		BaselineSTReturn:         p.BaselineSTReturn,
		PWR30:                    p.PWR30,
		SWR30:                    p.SWR30,
		StdDev:                   p.StdDev,
		UlcerScore:               p.UlcerScore,
		DeepestDrawdown:          p.DeepestDrawdown,
		LongestDrawdown:          p.LongestDrawdown,
		StartDateSensitivity:     p.StartDateSensitivity,
		AvgReturnRank:            p.AvgReturnRank,
		PWR30Rank:                p.PWR30Rank,
		SWR30Rank:                p.SWR30Rank,
		StdDevRank:               p.StdDevRank,
		UlcerScoreRank:           p.UlcerScoreRank,
		DeepestDrawdownRank:      p.DeepestDrawdownRank,
		LongestDrawdownRank:      p.LongestDrawdownRank,
		StartDateSensitivityRank: p.StartDateSensitivityRank,
		OverallRankScore:         p.OverallRankScore,
		OverallRankScoreRank:     p.OverallRankScoreRank,
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
func EvaluatePortfolios(perms []Permutation, assetMap map[string][]Percent) ([]*PortfolioStat, error) {
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
func evaluatePortfolios(perms []Permutation, assetMap map[string][]Percent) ([]*PortfolioStat, error) {
	// define this array to be reused
	var returnsList [][]Percent

	results := make([]*PortfolioStat, 0, len(perms))
	for i, p := range perms {
		// populate returnsList from p.Assets and assetMap
		{
			// zero out cached returnsList
			returnsList = returnsList[:0]
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

		minPWR30, _ := minPWR(portfolioReturns, 30)
		minSWR30, _ := minSWR(portfolioReturns, 30)
		maxUlcerScore, deepestDrawdown, longestDrawdown := drawdownScores(portfolioReturns)

		results = append(results, &PortfolioStat{
			Assets:               p.Assets,
			Percentages:          p.Percentages,
			AvgReturn:            averageReturn(portfolioReturns),
			BaselineLTReturn:     baselineLongTermReturn(portfolioReturns),
			BaselineSTReturn:     baselineShortTermReturn(portfolioReturns),
			PWR30:                minPWR30,
			SWR30:                minSWR30,
			StdDev:               standardDeviation(portfolioReturns),
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
	fmt.Println("...Calculate rank scores for the portfolios")
	{
		// rank by AvgReturn
		RankAll(results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.AvgReturn.Float() },
			LessIsBetter: false,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.AvgReturnRank = rank },
		})
		// rank by BaselineLTReturn
		RankAll(results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.BaselineLTReturn.Float() },
			LessIsBetter: false,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.BaselineLTReturnRank = rank },
		})
		// rank by BaselineSTReturn
		RankAll(results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.BaselineSTReturn.Float() },
			LessIsBetter: false,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.BaselineSTReturnRank = rank },
		})
		// rank by PWR30
		RankAll(results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.PWR30.Float() },
			LessIsBetter: false,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.PWR30Rank = rank },
		})
		// rank by SWR30
		RankAll(results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.SWR30.Float() },
			LessIsBetter: false,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.SWR30Rank = rank },
		})
		// rank by StdDev
		RankAll(results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.StdDev.Float() },
			LessIsBetter: true,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.StdDevRank = rank },
		})
		// rank by UlcerScore
		RankAll(results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.UlcerScore },
			LessIsBetter: true,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.UlcerScoreRank = rank },
		})
		// rank by DeepestDrawdown
		RankAll(results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.DeepestDrawdown.Float() },
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
			Metric:       func(stat *PortfolioStat) float64 { return stat.StartDateSensitivity.Float() },
			LessIsBetter: true,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.StartDateSensitivityRank = rank },
		})
	}

	fmt.Println("...rank by all their ranks (equally weighted)")
	{
		// populate the OverallRankScore for all
		for _, p := range results {
			p.OverallRankScore = math.Pow(p.AvgReturnRank.Percentage, 2) +
				math.Pow(p.BaselineLTReturnRank.Percentage, 2) +
				math.Pow(p.BaselineSTReturnRank.Percentage, 2) +
				math.Pow(p.PWR30Rank.Percentage, 2) +
				math.Pow(p.SWR30Rank.Percentage, 2) +
				math.Pow(p.StdDevRank.Percentage, 2) +
				math.Pow(p.UlcerScoreRank.Percentage, 2) +
				math.Pow(p.LongestDrawdownRank.Percentage, 2) +
				math.Pow(p.DeepestDrawdownRank.Percentage, 2) +
				math.Pow(p.StartDateSensitivityRank.Percentage, 2)
		}
		// rank by OverallRankScore
		RankAll(results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.OverallRankScore },
			LessIsBetter: true,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.OverallRankScoreRank = rank },
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
