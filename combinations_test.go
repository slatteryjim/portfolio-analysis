package portfolio_analysis

import (
	"fmt"
	"math"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/slatteryjim/portfolio-analysis/data"
	. "github.com/slatteryjim/portfolio-analysis/types"
)

func TestCombinations(t *testing.T) {
	g := NewGomegaWithT(t)

	dumpAll := func(perms []Combination) {
		for _, p := range perms {
			fmt.Println(p.Assets, p.Percentages)
		}
	}

	perms := Combinations([]string{"A"}, ReadablePercents(100))
	dumpAll(perms)
	g.Expect(perms).To(Equal([]Combination{
		{[]string{"A"}, ReadablePercents(100)},
	}))

	perms = Combinations([]string{"A", "B"}, ReadablePercents(50, 100))
	g.Expect(perms).To(ConsistOf([]Combination{
		{[]string{"A"}, ReadablePercents(100)},
		{[]string{"A", "B"}, ReadablePercents(50, 50)},
		{[]string{"B"}, ReadablePercents(100)},
	}))

	perms = Combinations([]string{"A", "B", "C"}, ReadablePercents(33, 66, 100))
	g.Expect(perms).To(ConsistOf([]Combination{
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

	perms = Combinations([]string{"A", "B", "C"}, ReadablePercents(series(1, 100, 1)...))
	g.Expect(len(perms)).To(Equal(5151))

	perms = Combinations([]string{"A", "B", "C", "D"}, ReadablePercents(series(1, 100, 1)...))
	g.Expect(len(perms)).To(Equal(176_851))

	// perms = Combinations([]string{"A", "B", "C", "D", "E"}, floats(1, 100, 1))
	// g.Expect(len(perms)).To(Equal(4_598_126))

	perms = Combinations([]string{"A", "B", "C", "D", "E"}, ReadablePercents(series(2.5, 100, 2.5)...))
	g.Expect(len(perms)).To(Equal(135_751))

	// perms = Combinations([]string{"A", "B", "C", "D", "E", "F"}, floats(2.5, 100, 2.5))
	// g.Expect(len(perms)).To(Equal(1_221_759))

	// perms = Combinations([]string{"A", "B", "C", "D", "E", "F", "G"}, floats(2.5, 100, 2.5))
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

func TestPortfolioCombinations_GoldenButterflyAssets(t *testing.T) {
	g := NewGomegaWithT(t)

	// GoldenButterfly advertised on: https://portfoliocharts.com/portfolio/golden-butterfly/
	// GoldenButterfly: [TSM SCV LTT STT GLD] [20% 20% 20% 20% 20%] (64) RF:0.00 AvgReturn:5.669%(5299) BLT:5.241%(2450) BST:2.849%(927) PWR:4.224%(1853) SWR:5.305%(1699) StdDev:8.103%(2383) Ulcer:3.4(2258) DeepestDrawdown:-15.33%(1862) LongestDrawdown:3(2), StartDateSensitivity:7.71%(756)
	//
	// Check out the results using 1% increments:
	// Best PWR30: [TSM SCV GLD] [1% 66% 33%] (2042098) RF:0.00 AvgReturn:7.932%(64438) BLT:5.536%(704794) BST:2.885%(431554) PWR:5.450%(1) SWR:6.284%(38) StdDev:13.397%(4192102) Ulcer:8.2(3097695) DeepestDrawdown:-26.73%(3323261) LongestDrawdown:6(5), StartDateSensitivity:16.97%(3381664)
	// Best UlcerScore: [TSM LTT STT GLD] [8% 3% 80% 9%] (3209303) RF:0.00 AvgReturn:2.738%(4589590) BLT:1.948%(4292822) BST:0.194%(3917744) PWR:2.456%(4228073) SWR:4.504%(2983311) StdDev:3.976%(1530) Ulcer:0.6(1) DeepestDrawdown:-5.43%(798) LongestDrawdown:4(3), StartDateSensitivity:9.10%(803328)
	//
	// Using 5% increments:
	// Best PWR30: [SCV GLD] [70% 30%] (4285) RF:0.00 AvgReturn:8.068%(178) BLT:5.896%(811) BST:2.386%(2517) PWR:5.364%(1) SWR:6.148%(8) StdDev:13.708%(9495) Ulcer:9.5(6975) DeepestDrawdown:-27.10%(7246) LongestDrawdown:6(5), StartDateSensitivity:16.20%(6723)
	// Best UlcerScore: [TSM STT GLD] [10% 80% 10%] (6678) RF:0.00 AvgReturn:2.808%(10564) BLT:2.070%(9640) BST:0.386%(8245) PWR:2.477%(9484) SWR:4.581%(5745) StdDev:3.928%(7) Ulcer:0.6(1) DeepestDrawdown:-5.60%(4) LongestDrawdown:2(1), StartDateSensitivity:8.44%(1191)
	//
	// Timing/log for GoldenButterfly assets, 1% step combinations:
	//   Generated 4598126 combinations in 6.551823599s
	//   ...Evaluating 4598126 combinations.
	//   Done evaluating portfolios in 53.007350212s or 86745 portfolios/second
	//   ...Calculate rank scores for the portfolios
	//   ...rank by all their ranks (equally weighted)
	//   Ranked portfolios in 1m8.660651682s
	startAt := time.Now()
	perms := Combinations([]string{"TSM", "SCV", "LTT", "STT", "GLD"}, ReadablePercents(seriesRange(5)...))
	// g.Expect(len(perms)).To(Equal(10_626)) // only 3,876 include all five.
	fmt.Println("Generated", len(perms), "combinations in", time.Since(startAt))

	// filter to only include combinations where all 5 assets are used/
	// (See: https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating)
	// {
	// 	startAt := time.Now()
	// 	numberOfAssets := 5
	// 	filtered := perms[:0]
	// 	for _, p := range perms {
	// 		// this cuts 10,626 combinations down to 3,876
	// 		if len(p.Assets) == numberOfAssets {
	// 			filtered = append(filtered, p)
	// 		}
	// 	}
	// 	for i := len(filtered); i < len(perms); i++ {
	// 		perms[i] = Combination{}
	// 	}
	// 	fmt.Printf("...culled down to %0.1f%% combinations in %s\n", float64(len(filtered))/float64(len(perms))*100, time.Since(startAt))
	// 	perms = filtered
	// }
	//g.Expect(len(perms)).To(Equal(3_876))
	startAt = time.Now()
	fmt.Println("...Evaluating", len(perms), "combinations.")

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

func TestPortfolioCombinations_AnythingBetterThanGoldenButterfly(t *testing.T) {
	// g := NewGomegaWithT(t)

	// need an n-choose-r algorithm
	// we'll just do an "n-choose-1" for the moment
	var results []*PortfolioStat
	for _, n := range data.Names() {
		p := Combination{
			Assets:      []string{n},
			Percentages: ReadablePercents(100),
		}
		stat := evaluatePortfolio(data.MustFind(n).AnnualReturns, p)
		results = append(results, stat)
	}
	RankPortfoliosInPlace(results)

	// print best:
	fmt.Println("Best combined overall ranks:")
	for i := 0; i < 10; i++ {
		fmt.Printf("#%d: %s\n", i+1, results[i])
	}

	PrintBestByEachRanking(results)
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

func TestEnumerateCombinations(t *testing.T) {

	makeAccumulateAllFn := func(buffer []string) (*[][]string, func() error) {
		var result [][]string
		return &result, func() error {
			fmt.Println(buffer)
			copyBuffer := make([]string, len(buffer))
			copy(copyBuffer, buffer)
			result = append(result, copyBuffer)
			return nil
		}
	}

	t.Run("no combinations", func(t *testing.T) {
		g := NewGomegaWithT(t)
		var nilBuffer []string
		result, fn := makeAccumulateAllFn(nilBuffer)
		g.Expect(EnumerateCombinations(nil, 0, nilBuffer, fn)).To(Succeed())
		g.Expect(*result).To(HaveLen(0))
	})

	t.Run("k=1", func(t *testing.T) {
		var (
			kOne      = 1
			bufferOne = make([]string, kOne)
		)
		t.Run("one combination", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferOne)
			g.Expect(EnumerateCombinations([]string{"1"}, kOne, bufferOne, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{{"1"}}))
		})
		t.Run("two combinations", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferOne)
			g.Expect(EnumerateCombinations([]string{"1", "2"}, kOne, bufferOne, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{{"1"}, {"2"}}))
		})
		t.Run("three combinations", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferOne)
			g.Expect(EnumerateCombinations([]string{"1", "2", "3"}, kOne, bufferOne, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{{"1"}, {"2"}, {"3"}}))
		})
	})

	t.Run("k=2", func(t *testing.T) {
		var (
			kTwo      = 2
			bufferTwo = make([]string, kTwo)
		)
		t.Run("one combination", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferTwo)
			g.Expect(EnumerateCombinations([]string{"1", "2"}, kTwo, bufferTwo, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{{"1", "2"}}))
		})
		t.Run("three combinations", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferTwo)
			g.Expect(EnumerateCombinations([]string{"1", "2", "3"}, kTwo, bufferTwo, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{
				{"1", "2"}, {"1", "3"},
				{"2", "3"}}))
		})
		t.Run("six combinations", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferTwo)
			g.Expect(EnumerateCombinations([]string{"1", "2", "3", "4"}, kTwo, bufferTwo, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{
				{"1", "2"}, {"1", "3"}, {"1", "4"},
				{"2", "3"}, {"2", "4"},
				{"3", "4"}}))
		})
	})
	t.Run("k=3", func(t *testing.T) {
		var (
			kThree      = 3
			bufferThree = make([]string, kThree)
		)
		t.Run("one combination", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferThree)
			g.Expect(EnumerateCombinations([]string{"1", "2", "3"}, kThree, bufferThree, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{{"1", "2", "3"}}))
		})
		t.Run("four combination", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferThree)
			g.Expect(EnumerateCombinations([]string{"1", "2", "3", "4"}, kThree, bufferThree, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{
				{"1", "2", "3"}, {"1", "2", "4"},
				{"1", "3", "4"}, {"2", "3", "4"}}))
		})
		t.Run("ten combination", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferThree)
			g.Expect(EnumerateCombinations([]string{"1", "2", "3", "4", "5"}, kThree, bufferThree, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{
				{"1", "2", "3"}, {"1", "2", "4"}, {"1", "2", "5"},
				{"1", "3", "4"}, {"1", "3", "5"}, {"1", "4", "5"},
				{"2", "3", "4"}, {"2", "3", "5"}, {"2", "4", "5"},
				{"3", "4", "5"}}))
		})
	})
	t.Run("validate some counts", func(t *testing.T) {
		g := NewGomegaWithT(t)

		makeCounterFn := func() (*int, func() error) {
			var count int
			return &count, func() error {
				count++
				return nil
			}
		}

		verify := func(n, r, expectedCount int) {
			var (
				xs        = make([]string, n)
				buffer    = make([]string, r)
				count, fn = makeCounterFn()
			)
			g.Expect(EnumerateCombinations(xs, r, buffer, fn)).To(Succeed())
			g.Expect(*count).To(Equal(expectedCount))
		}
		// expectedCount=0
		verify(0, 0, 0)
		verify(0, 1, 0)
		verify(1, 2, 0)
		// n=1
		verify(1, 1, 1)
		// n=2
		verify(2, 1, 2)
		verify(2, 2, 1)
		// n=3
		verify(3, 1, 3)
		verify(3, 2, 3)
		verify(3, 3, 1)
		// n=10, see the nice symetric curve
		verify(10, 1, 10)
		verify(10, 2, 45)
		verify(10, 3, 120)
		verify(10, 4, 210)
		verify(10, 5, 252) // peak
		verify(10, 6, 210)
		verify(10, 7, 120)
		verify(10, 8, 45)
		verify(10, 9, 10)
		// sample a few larger-size results
		verify(100, 2, 4_950)
		verify(100, 3, 161_700)
		verify(50, 4, 230_300)
		verify(25, 6, 177_100)
	})
	t.Run("can end early", func(t *testing.T) {
		// TODO: verify we can end at any point by returning ErrEndEnumeration from our callback
	})
}

// BenchmarkEnumerateCombinations/100_choose_2-12         	1000000000	         0.00119 ns/op
// BenchmarkEnumerateCombinations/100_choose_3-12         	1000000000	         0.0298 ns/op
// BenchmarkEnumerateCombinations/50_choose_4-12          	1000000000	         0.0303 ns/op
// BenchmarkEnumerateCombinations/50_choose_5-12          	1000000000	         0.269 ns/op
func BenchmarkEnumerateCombinations(b *testing.B) {
	makeCounterFn := func() (*int, func() error) {
		var count int
		return &count, func() error {
			count++
			return nil
		}
	}

	benchmark := func(n, r, expectedCount int) {
		var (
			xs        = make([]string, n)
			buffer    = make([]string, r)
			count, fn = makeCounterFn()
		)
		b.ResetTimer()
		err := EnumerateCombinations(xs, r, buffer, fn)
		if err != nil {
			b.Fatal(err)
		}
		if *count != expectedCount {
			b.Fatalf("expected %d to equal %d", *count, expectedCount)
		}
	}
	b.Run("100 choose 2", func(b *testing.B) {
		benchmark(100, 2, 4_950)
	})
	b.Run("100 choose 3", func(b *testing.B) {
		benchmark(100, 3, 161_700)
	})
	b.Run("50 choose 4", func(b *testing.B) {
		benchmark(50, 4, 230_300)
	})
	b.Run("50 choose 5", func(b *testing.B) {
		benchmark(50, 5, 2_118_760)
	})
}
