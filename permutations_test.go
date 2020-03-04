package portfolio_analysis

import (
	"fmt"
	"math"
	"sort"
	"testing"

	. "github.com/onsi/gomega"
)

func TestPermutations(t *testing.T) {
	g := NewGomegaWithT(t)

	translate := func(perms []Permutation) []Permutation {
		res := make([]Permutation, 0, len(perms))
		for _, p := range perms {
			res = append(res, Permutation{
				Assets:      p.Assets,
				Percentages: translatePercentages(p.Percentages),
			})
		}
		return res
	}

	dumpAll := func(perms []Permutation) {
		for _, p := range perms {
			fmt.Println(p.Assets, p.Percentages)
		}
	}

	perms := translate(Permutations([]string{"A"}, []float64{100}))
	dumpAll(perms)
	g.Expect(perms).To(Equal([]Permutation{
		{[]string{"A"}, []float64{100}},
	}))

	perms = translate(Permutations([]string{"A", "B"}, []float64{50, 100}))
	g.Expect(perms).To(ConsistOf([]Permutation{
		{[]string{"A"}, []float64{100}},
		{[]string{"A", "B"}, []float64{50, 50}},
		{[]string{"B"}, []float64{100}},
	}))

	perms = translate(Permutations([]string{"A", "B", "C"}, []float64{33, 66, 100}))
	g.Expect(perms).To(ConsistOf([]Permutation{
		{[]string{"A"}, []float64{100}},
		{[]string{"A", "B"}, []float64{66, 34}},
		{[]string{"A", "C"}, []float64{66, 34}},
		{[]string{"A", "B"}, []float64{33, 67}},
		{[]string{"A", "B", "C"}, []float64{33, 33, 34}},
		{[]string{"A", "C"}, []float64{33, 67}},
		{[]string{"B"}, []float64{100}},
		{[]string{"B", "C"}, []float64{66, 34}},
		{[]string{"B", "C"}, []float64{33, 67}},
		{[]string{"C"}, []float64{100}},
	}))

	perms = Permutations([]string{"A", "B", "C"}, floats(1, 100, 1))
	g.Expect(len(perms)).To(Equal(5151))

	perms = Permutations([]string{"A", "B", "C", "D"}, floats(1, 100, 1))
	g.Expect(len(perms)).To(Equal(176_851))

	// perms = Permutations([]string{"A", "B", "C", "D", "E"}, floats(1, 100, 1))
	// g.Expect(len(perms)).To(Equal(4_598_126))

	perms = Permutations([]string{"A", "B", "C", "D", "E"}, floats(2.5, 100, 2.5))
	g.Expect(len(perms)).To(Equal(135_751))

	// perms = Permutations([]string{"A", "B", "C", "D", "E", "F"}, floats(2.5, 100, 2.5))
	// g.Expect(len(perms)).To(Equal(1_221_759))

	// perms = Permutations([]string{"A", "B", "C", "D", "E", "F", "G"}, floats(2.5, 100, 2.5))
	// g.Expect(len(perms)).To(Equal(9_366_819))
}

type PortfolioStat struct {
	// describe portfolio assets and percentages
	Assets      []string
	Percentages []float64

	// stats on the portfolio performance
	MinPWR30        float64
	MaxUlcerScore   float64
	DeepestDrawdown float64
	LongestDrawdown int

	// This portfolio's rank on various stats
	MinPWR30Rank        int
	MaxUlcerScoreRank   int
	DeepestDrawdownRank int
	LongestDrawdownRank int
}

func (p PortfolioStat) String() string {
	return fmt.Sprintf("%v %v PWR30: %0.3f%% (%d) Ulcer:%0.1f(%d) DeepestDrawdown:%0.2f%%(%d) LongestDrawdown:%d(%d)",
		p.Assets,
		p.Percentages,
		p.MinPWR30*100,
		p.MinPWR30Rank,
		p.MaxUlcerScore,
		p.MaxUlcerScoreRank,
		p.DeepestDrawdown*100,
		p.DeepestDrawdownRank,
		p.LongestDrawdown,
		p.LongestDrawdownRank,
	)
}

func TestPortfolioPermutations(t *testing.T) {
	g := NewGomegaWithT(t)

	// Check out the results using 1% increments:
	// MaxPWR:   [TSM SCV LTT STT GLD] [1 65 1  1 32] PWR30: 5.404% Ulcer:8.1 DeepestDrawdown:-26.24% LongestDrawdown:6
	// MinUlcer: [TSM SCV LTT STT GLD] [7  1 3 80  9] PWR30: 2.470% Ulcer:0.6 DeepestDrawdown: -5.53% LongestDrawdown:4
	//
	// Using 5% increments:
	// MaxPWR:   [TSM SCV LTT STT GLD] [5 55 5  5 30] PWR30: 5.221% Ulcer:5.9 DeepestDrawdown:-23.66% LongestDrawdown:3
	// MinUlcer: [TSM SCV LTT STT GLD] [5  5 5 75 10] PWR30: 2.740% Ulcer:0.8 DeepestDrawdown: -6.58% LongestDrawdown:3

	perms := Permutations([]string{"TSM", "SCV", "LTT", "STT", "GLD"}, percentageRange(5))
	// g.Expect(len(perms)).To(Equal(10_626)) // only 3,876 include all five.

	// TODO: use an enum for the assets, so it's just an int under the covers, but has a nice String method,
	//  and maybe even a Returns() method that returns the appropriate []float64?
	assetMap := map[string][]float64{
		"TSM": TSM,
		"SCV": SCV,
		"LTT": LTT,
		"STT": STT,
		"GLD": GLD,
	}

	numberOfAssets := 5

	// allocate this array to be reused
	returnsList := make([][]float64, numberOfAssets)

	results := make([]*PortfolioStat, 0, len(perms))
	for _, p := range perms {
		if len(p.Assets) != numberOfAssets {
			continue // this cuts 10,626 permutations down to 3,876
		}
		for i, a := range p.Assets {
			returns, ok := assetMap[a]
			g.Expect(ok).To(BeTrue())
			returnsList[i] = returns
		}
		translatedPercentages := translatePercentages(p.Percentages)
		// fmt.Println(p.Assets, translatedPercentages)
		portfolioReturns, err := portfolioReturns(returnsList, translatedPercentages)

		// fmt.Println(portfolioReturns)

		minPWR30, _ := minPWR(portfolioReturns, 30)
		maxUlcerScore, deepestDrawdown, longestDrawdown := drawdownScores(portfolioReturns)

		g.Expect(err).ToNot(HaveOccurred())
		results = append(results, &PortfolioStat{
			Assets:          p.Assets,
			Percentages:     translatedPercentages,
			MinPWR30:        minPWR30,
			MaxUlcerScore:   maxUlcerScore,
			DeepestDrawdown: deepestDrawdown,
			LongestDrawdown: longestDrawdown,
		})
	}

	// Calculate rank scores for the portfolios
	{
		// TODO: factor out some better ranking logic. Take a function that returns the value of the parameter to rank on.
		//  I think we should let different numbers tie for the same place! Don't increment the rank number for no reason
		//  In that case the ranking numbers should probably then be translated into a percentage (1 to 100.0?)
		//  So that all the rankings operate on the same scale, regardless of ow many unique ranked values there were.

		// rank by MinPWR
		{
			sort.Slice(results, func(i, j int) bool { return results[i].MinPWR30 > results[j].MinPWR30 })
			for i, portfolioStat := range results {
				portfolioStat.MinPWR30Rank = i + 1
			}
		}
		// rank by MinPWR
		{
			sort.Slice(results, func(i, j int) bool { return results[i].MaxUlcerScore < results[j].MaxUlcerScore })
			for i, portfolioStat := range results {
				portfolioStat.MaxUlcerScoreRank = i + 1
			}
		}
		// rank by DeepestDrawdown
		{
			sort.Slice(results, func(i, j int) bool { return results[i].DeepestDrawdown > results[j].DeepestDrawdown })
			for i, portfolioStat := range results {
				portfolioStat.DeepestDrawdownRank = i + 1
			}
		}
		// rank by LongestDrawdown
		{
			sort.Slice(results, func(i, j int) bool { return results[i].LongestDrawdown < results[j].LongestDrawdown })
			for i, portfolioStat := range results {
				portfolioStat.LongestDrawdownRank = i + 1
			}
		}
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
			return math.Pow(float64(p.MinPWR30Rank), 2) +
				math.Pow(float64(p.MaxUlcerScoreRank), 2) +
				math.Pow(float64(p.LongestDrawdownRank), 2) +
				math.Pow(float64(p.DeepestDrawdownRank), 2)
		}
		sort.Slice(results, func(i, j int) bool {
			return sumRanks(results[i]) < sumRanks(results[j])
		})
		// print best:
		fmt.Println("Best combined overall ranks:")
		fmt.Println("#1:", results[0])
		fmt.Println("#2:", results[1])
		fmt.Println("#3:", results[2])
	}

	fmt.Println("\nBest by each ranking:")
	fmt.Println("Best MinPWR30:", findOne(results, func(p *PortfolioStat) bool { return p.MinPWR30Rank == 1 }))
	fmt.Println("Best MaxUlcerScore:", findOne(results, func(p *PortfolioStat) bool { return p.MaxUlcerScoreRank == 1 }))
	fmt.Println("Best DeepestDrawdown:", findOne(results, func(p *PortfolioStat) bool { return p.DeepestDrawdownRank == 1 }))
	fmt.Println("Best LongestDrawdown:", findOne(results, func(p *PortfolioStat) bool { return p.LongestDrawdownRank == 1 }))

	goldenButterflyStat := findOne(results, func(p *PortfolioStat) bool {
		for _, pct := range p.Percentages {
			if pct != 20 {
				return false
			}
		}
		return true
	})
	fmt.Println("\nGoldenButterfly:", goldenButterflyStat)
	// find as good or better than GoldenButterfly
	betterThanGB := findMany(results, func(p *PortfolioStat) bool {
		return p.DeepestDrawdown >= goldenButterflyStat.DeepestDrawdown &&
			p.LongestDrawdown <= goldenButterflyStat.LongestDrawdown &&
			p.MaxUlcerScore <= goldenButterflyStat.MaxUlcerScore &&
			p.MinPWR30 >= goldenButterflyStat.MinPWR30
	})
	fmt.Println("As good or better than GoldenButterfly:", len(betterThanGB))
	for i, p := range betterThanGB {
		fmt.Println(" ", i, p)
	}
}

func findOne(results []*PortfolioStat, pred func(p *PortfolioStat) bool) *PortfolioStat {
	for _, p := range results {
		if pred(p) {
			return p
		}
	}
	return nil
}

func findMany(results []*PortfolioStat, pred func(p *PortfolioStat) bool) []*PortfolioStat {
	var res []*PortfolioStat
	for _, p := range results {
		if pred(p) {
			res = append(res, p)
		}
	}
	return res
}

func Test_percentageRange(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(percentageRange(25)).To(Equal([]float64{25, 50, 75, 100}))
	g.Expect(percentageRange(10)).To(Equal([]float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}))
	g.Expect(percentageRange(33.333333333333333)).To(Equal([]float64{33.333333333333336, 66.66666666666667, 100}))
}

func Test_floats(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(floats(25, 100, 25)).To(Equal([]float64{25, 50, 75, 100}))
	g.Expect(floats(12.5, 100, 12.5)).To(Equal([]float64{12.5, 25, 37.5, 50, 62.5, 75, 87.5, 100}))
}

func translatePercentages(ps []float64) []float64 {
	res := make([]float64, 0, len(ps))
	prev := 0.0
	for _, p := range ps {
		res = append(res, p-prev)
		prev = p
	}
	return res
}

func percentageRange(step float64) []float64 {
	return floats(step, 100, step)
}

func floats(start, end, step float64) []float64 {
	var res []float64
	for i := start; i <= end; i += step {
		res = append(res, i)
	}
	return res
}
