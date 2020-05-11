package portfolio_analysis

import (
	"fmt"
	"math"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestPermutations(t *testing.T) {
	g := NewGomegaWithT(t)

	dumpAll := func(perms []Permutation) {
		for _, p := range perms {
			fmt.Println(p.Assets, p.Percentages)
		}
	}

	perms := Permutations([]string{"A"}, readablePercents(100))
	dumpAll(perms)
	g.Expect(perms).To(Equal([]Permutation{
		{[]string{"A"}, readablePercents(100)},
	}))

	perms = Permutations([]string{"A", "B"}, readablePercents(50, 100))
	g.Expect(perms).To(ConsistOf([]Permutation{
		{[]string{"A"}, readablePercents(100)},
		{[]string{"A", "B"}, readablePercents(50, 50)},
		{[]string{"B"}, readablePercents(100)},
	}))

	perms = Permutations([]string{"A", "B", "C"}, readablePercents(33, 66, 100))
	g.Expect(perms).To(ConsistOf([]Permutation{
		{[]string{"A"}, []Percent{1.00}},
		{[]string{"A", "B"}, []Percent{0.66, 0.33999999999999997}},
		{[]string{"A", "C"}, []Percent{0.66, 0.33999999999999997}},
		{[]string{"A", "B"}, []Percent{0.33, 0.6699999999999999}},
		{[]string{"A", "B", "C"}, []Percent{0.33, 0.33, 0.33999999999999997}},
		{[]string{"A", "C"}, []Percent{0.33, 0.6699999999999999}},
		{[]string{"B"}, []Percent{1.00}},
		{[]string{"B", "C"}, []Percent{0.66, 0.33999999999999997}},
		{[]string{"B", "C"}, []Percent{0.33, 0.6699999999999999}},
		{[]string{"C"}, []Percent{1.00}},
	}))

	perms = Permutations([]string{"A", "B", "C"}, readablePercents(series(1, 100, 1)...))
	g.Expect(len(perms)).To(Equal(5151))

	perms = Permutations([]string{"A", "B", "C", "D"}, readablePercents(series(1, 100, 1)...))
	g.Expect(len(perms)).To(Equal(176_851))

	// perms = Permutations([]string{"A", "B", "C", "D", "E"}, floats(1, 100, 1))
	// g.Expect(len(perms)).To(Equal(4_598_126))

	perms = Permutations([]string{"A", "B", "C", "D", "E"}, readablePercents(series(2.5, 100, 2.5)...))
	g.Expect(len(perms)).To(Equal(135_751))

	// perms = Permutations([]string{"A", "B", "C", "D", "E", "F"}, floats(2.5, 100, 2.5))
	// g.Expect(len(perms)).To(Equal(1_221_759))

	// perms = Permutations([]string{"A", "B", "C", "D", "E", "F", "G"}, floats(2.5, 100, 2.5))
	// g.Expect(len(perms)).To(Equal(9_366_819))
}

func Test_translatePercentages(t *testing.T) {
	g := NewGomegaWithT(t)

	verify := func(ps, expected []Percent) {
		t.Helper()
		translatePercentages(ps)
		g.Expect(ps).To(Equal(expected))
	}

	verify(nil, nil)
	verify([]Percent{}, []Percent{})

	verify([]Percent{25}, []Percent{25})

	verify([]Percent{25, 100}, []Percent{25, 75})

	verify([]Percent{25, 50, 75, 100}, []Percent{25, 25, 25, 25})
}

func TestPortfolioPermutations_GoldenButterflyAssets(t *testing.T) {
	g := NewGomegaWithT(t)

	// GoldenButterfly advertised on: https://portfoliocharts.com/portfolio/golden-butterfly/
	// Standard: [TSM SCV LTT STT GLD] [20% 20% 20% 20% 20%] (66) RF:0.00 AvgReturn:5.668%(5300) BLT:5.241%(2449) BST:2.848%(929) PWR:4.224%(1853) SWR:5.305%(1699) StdDev:8.103%(2383) Ulcer:3.4(2258) DeepestDrawdown:-15.33%(1835) LongestDrawdown:3(2), StartDateSensitivity:7.71%(756)
	//
	// Check out the results using 1% increments:
	// MaxPWR:   [TSM SCV LTT STT GLD] [1 65 1  1 32] PWR30: 5.404% Ulcer:8.1 DeepestDrawdown:-26.24% LongestDrawdown:6
	// MinUlcer: [TSM SCV LTT STT GLD] [7  1 3 80  9] PWR30: 2.470% Ulcer:0.6 DeepestDrawdown: -5.53% LongestDrawdown:4
	//
	// Using 5% increments:
	// Best PWR30: [SCV GLD] [70% 30%] (4281) RF:0.00 AvgReturn:8.068%(178) BLT:5.896%(810) BST:2.385%(2518) PWR:5.364%(1) SWR:6.148%(8) StdDev:13.709%(9495) Ulcer:9.5(6975) DeepestDrawdown:-27.10%(7168) LongestDrawdown:6(5), StartDateSensitivity:16.21%(6723)
	// Best UlcerScore: [TSM STT GLD] [10% 80% 10%] (6679) RF:0.00 AvgReturn:2.809%(10564) BLT:2.069%(9640) BST:0.385%(8246) PWR:2.477%(9484) SWR:4.582%(5739) StdDev:3.929%(7) Ulcer:0.6(1) DeepestDrawdown:-5.59%(4) LongestDrawdown:2(1), StartDateSensitivity:8.44%(1193)
	//
	// Timing/log for GoldenButterfly assets, 1% step permutations:
	//   Generated 4,598,126 permutations in 10.9s
	//   ...culled down to 81.9% permutations in 82ms
	//   ...Evaluating 3,764,376 permutations.
	//   Done evaluating portfolios in 2m40s
	//   Ranked portfolios in 50.4s
	startAt := time.Now()
	perms := Permutations([]string{"TSM", "SCV", "LTT", "STT", "GLD"}, readablePercents(seriesRange(5)...))
	// g.Expect(len(perms)).To(Equal(10_626)) // only 3,876 include all five.
	fmt.Println("Generated", len(perms), "permutations in", time.Since(startAt))

	// filter to only include permutations where all 5 assets are used/
	// (See: https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating)
	// {
	// 	startAt := time.Now()
	// 	numberOfAssets := 5
	// 	filtered := perms[:0]
	// 	for _, p := range perms {
	// 		// this cuts 10,626 permutations down to 3,876
	// 		if len(p.Assets) == numberOfAssets {
	// 			filtered = append(filtered, p)
	// 		}
	// 	}
	// 	for i := len(filtered); i < len(perms); i++ {
	// 		perms[i] = Permutation{}
	// 	}
	// 	fmt.Printf("...culled down to %0.1f%% permutations in %s\n", float64(len(filtered))/float64(len(perms))*100, time.Since(startAt))
	// 	perms = filtered
	// }
	//g.Expect(len(perms)).To(Equal(3_876))
	startAt = time.Now()
	fmt.Println("...Evaluating", len(perms), "permutations.")

	results, err := EvaluatePortfolios(perms, assetMap)
	g.Expect(err).ToNot(HaveOccurred())
	elapsed := time.Since(startAt)
	fmt.Println("Done evaluating portfolios in", elapsed, "or", int(float64(len(results))/elapsed.Seconds()), "portfolios/second")

	startAt = time.Now()
	RankPortfoliosInPlace(results)
	fmt.Println("Ranked portfolios in", time.Since(startAt))

	// print best:
	fmt.Println("Best combined overall ranks:")
	fmt.Println("#1:", results[0])
	fmt.Println("#2:", results[1])
	fmt.Println("#3:", results[2])

	PrintBestByEachRanking(results)

	startAt = time.Now()
	gbStat := FindOne(results, func(p *PortfolioStat) bool {
		if len(p.Percentages) != 5 {
			return false
		}
		for _, pct := range p.Percentages {
			if !approxEqual(pct.Float(), 0.20, 0.001) {
				return false
			}
		}
		return true
	})
	g.Expect(gbStat).ToNot(BeNil())
	fmt.Println("\nGoldenButterfly:", gbStat)
	// find as good or better than GoldenButterfly
	betterThanGB := CopyAll(FindMany(results, AsGoodOrBetterThan(gbStat)))
	RankPortfoliosInPlace(betterThanGB)
	fmt.Println("As good or better than GoldenButterfly:", len(betterThanGB))
	PrintBestByEachRanking(betterThanGB)
	fmt.Println("\nAll as good or better:")
	for i, p := range betterThanGB[:min(len(betterThanGB), 5)] {
		fmt.Println(" ", i, p.ComparePerformance(*gbStat))
	}
	fmt.Println("Finished GB analysis in", time.Since(startAt))
}

func PrintBestByEachRanking(results []*PortfolioStat) {
	fmt.Println("\nBest by each ranking:")
	fmt.Println("Best AvgReturn:", FindOne(results, func(p *PortfolioStat) bool { return p.AvgReturnRank.Ordinal == 1 }))
	fmt.Println("Best BaselineLTReturn:", FindOne(results, func(p *PortfolioStat) bool { return p.BaselineLTReturnRank.Ordinal == 1 }))
	fmt.Println("Best BaselineSTReturn:", FindOne(results, func(p *PortfolioStat) bool { return p.BaselineSTReturnRank.Ordinal == 1 }))
	fmt.Println("Best PWR30:", FindOne(results, func(p *PortfolioStat) bool { return p.PWR30Rank.Ordinal == 1 }))
	fmt.Println("Best SWR30:", FindOne(results, func(p *PortfolioStat) bool { return p.SWR30Rank.Ordinal == 1 }))
	fmt.Println("Best StdDev:", FindOne(results, func(p *PortfolioStat) bool { return p.StdDevRank.Ordinal == 1 }))
	fmt.Println("Best UlcerScore:", FindOne(results, func(p *PortfolioStat) bool { return p.UlcerScoreRank.Ordinal == 1 }))
	fmt.Println("Best DeepestDrawdown:", FindOne(results, func(p *PortfolioStat) bool { return p.DeepestDrawdownRank.Ordinal == 1 }))
	fmt.Println("Best LongestDrawdown:", FindOne(results, func(p *PortfolioStat) bool { return p.LongestDrawdownRank.Ordinal == 1 }))
	fmt.Println("Best StartDateSensitivity:", FindOne(results, func(p *PortfolioStat) bool { return p.StartDateSensitivityRank.Ordinal == 1 }))
}

func approxEqual(x, y, tolerance float64) bool {
	return math.Abs(x-y) < tolerance
}

func Test_seriesRange(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(seriesRange(25)).To(Equal([]float64{25, 50, 75, 100}))
	g.Expect(seriesRange(10)).To(Equal([]float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}))
	g.Expect(seriesRange(33.333333333333333)).To(Equal([]float64{33.333333333333336, 66.66666666666667, 100}))
}

func Test_series(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(series(25, 100, 25)).To(Equal([]float64{25, 50, 75, 100}))
	g.Expect(series(12.5, 100, 12.5)).To(Equal([]float64{12.5, 25, 37.5, 50, 62.5, 75, 87.5, 100}))
}

func seriesRange(step float64) []float64 {
	return series(step, 100, step)
}

func series(start, end, step float64) []float64 {
	var res []float64
	for i := start; i <= end; i += step {
		res = append(res, i)
	}
	return res
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
