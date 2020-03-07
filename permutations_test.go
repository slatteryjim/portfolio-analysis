package portfolio_analysis

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
)

func TestPermutations(t *testing.T) {
	g := NewGomegaWithT(t)

	dumpAll := func(perms []Permutation) {
		for _, p := range perms {
			fmt.Println(p.Assets, p.Percentages)
		}
	}

	perms := Permutations([]string{"A"}, []float64{100})
	dumpAll(perms)
	g.Expect(perms).To(Equal([]Permutation{
		{[]string{"A"}, []float64{100}},
	}))

	perms = Permutations([]string{"A", "B"}, []float64{50, 100})
	g.Expect(perms).To(ConsistOf([]Permutation{
		{[]string{"A"}, []float64{100}},
		{[]string{"A", "B"}, []float64{50, 50}},
		{[]string{"B"}, []float64{100}},
	}))

	perms = Permutations([]string{"A", "B", "C"}, []float64{33, 66, 100})
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

func Test_translatePercentages(t *testing.T) {
	g := NewGomegaWithT(t)

	verify := func(ps, expected []float64) {
		t.Helper()
		translatePercentages(ps)
		g.Expect(ps).To(Equal(expected))
	}

	verify(nil, nil)
	verify([]float64{}, []float64{})

	verify([]float64{25}, []float64{25})

	verify([]float64{25, 100}, []float64{25, 75})

	verify([]float64{25, 50, 75, 100}, []float64{25, 25, 25, 25})
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
	fmt.Println("Generated", len(perms), "permutations.")

	// filter to only include permutations where all 5 assets are used/
	// (See: https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating)
	{
		numberOfAssets := 5
		filtered := perms[:0]
		for _, p := range perms {
			// this cuts 10,626 permutations down to 3,876
			if len(p.Assets) == numberOfAssets {
				filtered = append(filtered, p)
			}
		}
		for i := len(filtered); i < len(perms); i++ {
			perms[i] = Permutation{}
		}
		perms = filtered
	}
	//g.Expect(len(perms)).To(Equal(3_876))
	fmt.Println("...Culled down. Evaluating", len(perms), "permutations.")

	results, err := EvaluatePortfolios(perms, assetMap)
	g.Expect(err).ToNot(HaveOccurred())
	fmt.Println("Done evaluating portfolios.")

	RankPortfoliosInPlace(results)

	// print best:
	fmt.Println("Best combined overall ranks:")
	fmt.Println("#1:", results[0])
	fmt.Println("#2:", results[1])
	fmt.Println("#3:", results[2])

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

	gbStat := FindOne(results, func(p *PortfolioStat) bool {
		return reflect.DeepEqual(p.Percentages, []float64{20, 20, 20, 20, 20})
	})
	fmt.Println("\nGoldenButterfly:", gbStat)
	// find as good or better than GoldenButterfly
	betterThanGB := CopyAll(FindMany(results, func(p *PortfolioStat) bool {
		return p.AvgReturnRank.Ordinal <= gbStat.AvgReturnRank.Ordinal &&
			p.BaselineLTReturnRank.Ordinal <= gbStat.BaselineLTReturnRank.Ordinal &&
			p.BaselineSTReturnRank.Ordinal <= gbStat.BaselineSTReturnRank.Ordinal &&
			p.PWR30Rank.Ordinal <= gbStat.PWR30Rank.Ordinal &&
			p.SWR30Rank.Ordinal <= gbStat.SWR30Rank.Ordinal &&
			p.StdDevRank.Ordinal <= gbStat.StdDevRank.Ordinal &&
			p.UlcerScoreRank.Ordinal <= gbStat.UlcerScoreRank.Ordinal &&
			p.DeepestDrawdownRank.Ordinal <= gbStat.DeepestDrawdownRank.Ordinal &&
			p.LongestDrawdownRank.Ordinal <= gbStat.LongestDrawdownRank.Ordinal &&
			p.StartDateSensitivityRank.Ordinal <= gbStat.StartDateSensitivityRank.Ordinal
	}))
	RankPortfoliosInPlace(betterThanGB)
	fmt.Println("As good or better than GoldenButterfly:", len(betterThanGB))
	fmt.Println("Best AvgReturn:", FindOne(betterThanGB, func(p *PortfolioStat) bool { return p.AvgReturnRank.Ordinal == 1 }))
	fmt.Println("Best BaselineLTReturn:", FindOne(betterThanGB, func(p *PortfolioStat) bool { return p.BaselineLTReturnRank.Ordinal == 1 }))
	fmt.Println("Best BaselineSTReturn:", FindOne(betterThanGB, func(p *PortfolioStat) bool { return p.BaselineSTReturnRank.Ordinal == 1 }))
	fmt.Println("Best PWR30:", FindOne(betterThanGB, func(p *PortfolioStat) bool { return p.PWR30Rank.Ordinal == 1 }))
	fmt.Println("Best SWR30:", FindOne(betterThanGB, func(p *PortfolioStat) bool { return p.SWR30Rank.Ordinal == 1 }))
	fmt.Println("Best StdDev:", FindOne(betterThanGB, func(p *PortfolioStat) bool { return p.StdDevRank.Ordinal == 1 }))
	fmt.Println("Best UlcerScore:", FindOne(betterThanGB, func(p *PortfolioStat) bool { return p.UlcerScoreRank.Ordinal == 1 }))
	fmt.Println("Best DeepestDrawdown:", FindOne(betterThanGB, func(p *PortfolioStat) bool { return p.DeepestDrawdownRank.Ordinal == 1 }))
	fmt.Println("Best LongestDrawdown:", FindOne(betterThanGB, func(p *PortfolioStat) bool { return p.LongestDrawdownRank.Ordinal == 1 }))
	fmt.Println("Best StartDateSensitivity:", FindOne(betterThanGB, func(p *PortfolioStat) bool { return p.StartDateSensitivityRank.Ordinal == 1 }))
	fmt.Println("\nAll as good or better:")
	for i, p := range betterThanGB[:min(len(betterThanGB), 5)] {
		fmt.Println(" ", i, p.ComparePerformance(*gbStat))
	}
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
