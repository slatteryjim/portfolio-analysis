package portfolio_analysis

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/slatteryjim/portfolio-analysis/data"
	. "github.com/slatteryjim/portfolio-analysis/types"
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
		Assets      []string  `parquet:"name=assets,                 type=LIST, valuetype=UTF8"`
		Percentages []Percent `parquet:"name=allocation_percentages, type=LIST, valuetype=DOUBLE"`

		RebalanceFactor float64

		// stats on the portfolio performance
		AvgReturn            Percent `parquet:"name=avg_return,            type=DOUBLE"`
		BaselineLTReturn     Percent `parquet:"name=baseline_lt_return,    type=DOUBLE"`
		BaselineSTReturn     Percent `parquet:"name=baseline_st_return,    type=DOUBLE"`
		PWR30                Percent `parquet:"name=pwr30,                 type=DOUBLE"`
		SWR30                Percent `parquet:"name=swr30,                 type=DOUBLE"`
		StdDev               Percent `parquet:"name=std_dev,               type=DOUBLE"`
		UlcerScore           float64 `parquet:"name=ulcer_score,           type=DOUBLE"`
		DeepestDrawdown      Percent `parquet:"name=deepest_drawdown,      type=DOUBLE"`
		LongestDrawdown      int     `parquet:"name=longest_drawdown,      type=DOUBLE"`
		StartDateSensitivity Percent `parquet:"name=startdate_sensitivity, type=DOUBLE"`

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

	// PortfolioStatParquet is the PortfolioStat
	PortfolioStatParquet struct {
		// describe portfolio assets and percentages
		Assets []string `parquet:"name=assets,                 type=LIST, valuetype=UTF8"`
		// Percentages []float64 `parquet:"name=allocation_percentages, type=LIST, valuetype=DOUBLE"`

		// YearlyReturns []float64 `parquet:"name=yearly_returns,       type=LIST, valuetype=DOUBLE"`

		// some attributes derived from Assets and YearlyReturns values
		NumAssets uint8 `parquet:"name=num_assets, type=UINT_8"`
		NumYears  uint8 `parquet:"name=num_years,  type=UINT_8"`

		// stats on the portfolio performance
		AvgReturn            float64 `parquet:"name=avg_return,            type=DOUBLE"`
		BaselineLTReturn     float64 `parquet:"name=baseline_lt_return,    type=DOUBLE"`
		BaselineSTReturn     float64 `parquet:"name=baseline_st_return,    type=DOUBLE"`
		PWR30                float64 `parquet:"name=pwr30,                 type=DOUBLE"`
		SWR30                float64 `parquet:"name=swr30,                 type=DOUBLE"`
		StdDev               float64 `parquet:"name=std_dev,               type=DOUBLE"`
		UlcerScore           float64 `parquet:"name=ulcer_score,           type=DOUBLE"`
		DeepestDrawdown      float64 `parquet:"name=deepest_drawdown,      type=DOUBLE"`
		LongestDrawdown      float64 `parquet:"name=longest_drawdown,      type=DOUBLE"`
		StartDateSensitivity float64 `parquet:"name=startdate_sensitivity, type=DOUBLE"`
	}
	Rank struct {
		Ordinal    int
		Percentage float64
	}
)

func (p PortfolioStat) String() string {
	return fmt.Sprintf("%v %v (%d) RF:%0.2f AvgReturn:%0.3f%%(%d) BLT:%0.3f%%(%d) BST:%0.3f%%(%d) PWR:%0.3f%%(%d) SWR:%0.3f%%(%d) StdDev:%0.3f%%(%d) Ulcer:%0.1f(%d) DeepestDrawdown:%0.2f%%(%d) LongestDrawdown:%d(%d), StartDateSensitivity:%0.2f%%(%d)",
		p.Assets,
		p.Percentages,
		p.OverallRankScoreRank.Ordinal,
		p.RebalanceFactor,
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

func (p PortfolioStat) DiffPerformance(other PortfolioStat) PortfolioStat {
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

func (p *PortfolioStat) AsGoodOrBetterThan(other *PortfolioStat) bool {
	return p.AvgReturn >= other.AvgReturn &&
		p.BaselineLTReturn >= other.BaselineLTReturn &&
		p.BaselineSTReturn >= other.BaselineSTReturn &&
		p.PWR30 >= other.PWR30 &&
		p.SWR30 >= other.SWR30 &&
		p.StdDev <= other.StdDev &&
		p.UlcerScore <= other.UlcerScore &&
		p.DeepestDrawdown >= other.DeepestDrawdown &&
		p.LongestDrawdown <= other.LongestDrawdown &&
		p.StartDateSensitivity <= other.StartDateSensitivity
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
		RebalanceFactor:          p.RebalanceFactor,
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

func (p PortfolioStat) MustReturns() []Percent {
	assetReturns := data.PortfolioReturnsList(p.Assets...)
	returns, err := portfolioReturns(assetReturns, p.Percentages)
	if err != nil {
		panic(err.Error())
	}
	return returns

}

// Parquet returns a struct ready to be serialized as Parquet.
func (p PortfolioStat) Parquet() PortfolioStatParquet {
	yearlyReturns := Floats(p.MustReturns()...)
	return PortfolioStatParquet{
		Assets: p.Assets,
		// Percentages: Floats(p.Percentages...),
		// YearlyReturns:        yearlyReturns,
		NumAssets:            uint8(len(p.Assets)),
		NumYears:             uint8(len(yearlyReturns)),
		AvgReturn:            p.AvgReturn.Float(),
		BaselineLTReturn:     p.BaselineLTReturn.Float(),
		BaselineSTReturn:     p.BaselineSTReturn.Float(),
		PWR30:                p.PWR30.Float(),
		SWR30:                p.SWR30.Float(),
		StdDev:               p.StdDev.Float(),
		UlcerScore:           p.UlcerScore,
		DeepestDrawdown:      p.DeepestDrawdown.Float(),
		LongestDrawdown:      float64(p.LongestDrawdown),
		StartDateSensitivity: p.StartDateSensitivity.Float(),
	}
}

func CopyAll(ps []*PortfolioStat) []*PortfolioStat {
	res := make([]*PortfolioStat, len(ps))
	for i, p := range ps {
		res[i] = p.Clone()
	}
	return res
}

// EvaluatePortfolios evaluates the portfolio for each of the given combinations, returning a slice of stats.
// It processes in parallel using multiple CPUs as needed.
func EvaluatePortfolios(perms []Combination, assetMap map[string][]Percent) ([]*PortfolioStat, error) {
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

// evaluatePortfolios evaluates the portfolio for each of the given combinations, returning a slice of stats.
func evaluatePortfolios(perms []Combination, assetMap map[string][]Percent) ([]*PortfolioStat, error) {
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
		results = append(results, evaluatePortfolio(portfolioReturns, p))
	}
	return results, nil
}

func evaluatePortfolio(portfolioReturns []Percent, p Combination) *PortfolioStat {
	minPWR30, minSWR30 := minPWRAndSWR(portfolioReturns, 30)
	maxUlcerScore, deepestDrawdown, longestDrawdown := drawdownScores(portfolioReturns)

	return &PortfolioStat{
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
	}
}

// evaluatePortfolioIfAsGoodOrBetterThan evaluates the given portfolioReturns and returns
// a non-nil PortfolioStat only if the performance metrics are all as good or better than the given
// otherStat porformance.
// It can return early if any of the metrics aren't as good.
func evaluatePortfolioIfAsGoodOrBetterThan(portfolioReturns []Percent, p Combination, other *PortfolioStat) *PortfolioStat {
	avgReturn := averageReturn(portfolioReturns)
	if avgReturn < other.AvgReturn {
		return nil
	}
	stdDev := standardDeviation(portfolioReturns)
	if stdDev > other.StdDev {
		return nil
	}
	minPWR30, minSWR30 := minPWRAndSWR(portfolioReturns, 30)
	if minPWR30 < other.PWR30 {
		return nil
	}
	if minSWR30 < other.SWR30 {
		return nil
	}
	baselineLT := baselineLongTermReturn(portfolioReturns)
	if baselineLT < other.BaselineLTReturn {
		return nil
	}
	maxUlcerScore, deepestDrawdown, longestDrawdown := drawdownScores(portfolioReturns)
	if maxUlcerScore > other.UlcerScore {
		return nil
	}
	if deepestDrawdown < other.DeepestDrawdown {
		return nil
	}
	if longestDrawdown > other.LongestDrawdown {
		return nil
	}
	baselineST := baselineShortTermReturn(portfolioReturns)
	if baselineST < other.BaselineSTReturn {
		return nil
	}
	sensitivity := startDateSensitivity(portfolioReturns)
	if sensitivity > other.StartDateSensitivity {
		return nil
	}
	return &PortfolioStat{
		Assets:               p.Assets,
		Percentages:          p.Percentages,
		AvgReturn:            avgReturn,
		BaselineLTReturn:     baselineLT,
		BaselineSTReturn:     baselineST,
		PWR30:                minPWR30,
		SWR30:                minSWR30,
		StdDev:               stdDev,
		UlcerScore:           maxUlcerScore,
		DeepestDrawdown:      deepestDrawdown,
		LongestDrawdown:      longestDrawdown,
		StartDateSensitivity: sensitivity,
	}
}

// RankPortfoliosInPlace this is a "destructive" operation, reordering the list and mutating the ***Rank fields.
// It sorts the list by the various performance metrics and populates the corresponding Rank field for each.
// It finishes up by sorting them by their "overall" rank, considering all of the performance metrics equally.
func RankPortfoliosInPlace(results []*PortfolioStat) {
	startAt := time.Now()
	fmt.Println("...Calculate rank scores for the portfolios")
	{
		RankAll("AvgReturn", results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.AvgReturn.Float() },
			LessIsBetter: false,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.AvgReturnRank = rank },
		})
		RankAll("BaselineLTReturn", results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.BaselineLTReturn.Float() },
			LessIsBetter: false,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.BaselineLTReturnRank = rank },
		})
		RankAll("BaselineSTReturn", results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.BaselineSTReturn.Float() },
			LessIsBetter: false,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.BaselineSTReturnRank = rank },
		})
		RankAll("PWR30", results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.PWR30.Float() },
			LessIsBetter: false,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.PWR30Rank = rank },
		})
		RankAll("SWR30", results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.SWR30.Float() },
			LessIsBetter: false,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.SWR30Rank = rank },
		})
		RankAll("StdDev", results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.StdDev.Float() },
			LessIsBetter: true,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.StdDevRank = rank },
		})
		RankAll("UlcerScore", results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.UlcerScore },
			LessIsBetter: true,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.UlcerScoreRank = rank },
		})
		RankAll("DeepestDrawdown", results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.DeepestDrawdown.Float() },
			LessIsBetter: false,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.DeepestDrawdownRank = rank },
		})
		RankAll("LongestDrawdown", results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return float64(stat.LongestDrawdown) },
			LessIsBetter: true,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.LongestDrawdownRank = rank },
		})
		RankAll("StartDateSensitivity", results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.StartDateSensitivity.Float() },
			LessIsBetter: true,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.StartDateSensitivityRank = rank },
		})
	}
	fmt.Println("Finished basic rank scores in", time.Since(startAt))
	startAt = time.Now()
	fmt.Println("...rank by all their ranks (equally weighted)")
	{
		// populate the OverallRankScore for all
		for i, p := range results {
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
			if i%10_000_000 == 0 {
				fmt.Println(" - populating OverallRankScore row", i+1)
			}
		}
		fmt.Println("populated OverallRankScore for all in", time.Since(startAt))
		RankAll("OverallRankScore", results, RankAllParams{
			Metric:       func(stat *PortfolioStat) float64 { return stat.OverallRankScore },
			LessIsBetter: true,
			SetRank:      func(stat *PortfolioStat, rank Rank) { stat.OverallRankScoreRank = rank },
		})
	}
	fmt.Println("Elapsed:", time.Since(startAt))
}

type RankAllParams struct {
	Metric       func(*PortfolioStat) float64
	LessIsBetter bool
	SetRank      func(stat *PortfolioStat, rank Rank)
}

func RankAll(
	name string,
	results []*PortfolioStat,
	params RankAllParams,
) {
	startAt := time.Now()
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
	fmt.Println("Finished ranking by", name, "in", time.Since(startAt))
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

// AsGoodOrBetterThan returns a function that returns true if p is as good or better
// than the model, in all ranking ordinals.
// Assumes that the model and all tested inputs have already been ranked against one another.
func AsGoodOrBetterThan(model *PortfolioStat) func(p *PortfolioStat) bool {
	return func(p *PortfolioStat) bool {
		return p.AvgReturnRank.Ordinal <= model.AvgReturnRank.Ordinal &&
			p.BaselineLTReturnRank.Ordinal <= model.BaselineLTReturnRank.Ordinal &&
			p.BaselineSTReturnRank.Ordinal <= model.BaselineSTReturnRank.Ordinal &&
			p.PWR30Rank.Ordinal <= model.PWR30Rank.Ordinal &&
			p.SWR30Rank.Ordinal <= model.SWR30Rank.Ordinal &&
			p.StdDevRank.Ordinal <= model.StdDevRank.Ordinal &&
			p.UlcerScoreRank.Ordinal <= model.UlcerScoreRank.Ordinal &&
			p.DeepestDrawdownRank.Ordinal <= model.DeepestDrawdownRank.Ordinal &&
			p.LongestDrawdownRank.Ordinal <= model.LongestDrawdownRank.Ordinal &&
			p.StartDateSensitivityRank.Ordinal <= model.StartDateSensitivityRank.Ordinal
	}
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
