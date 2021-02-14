package portfolio_analysis

import (
	"fmt"
	"strings"
	"testing"

	"github.com/guptarohit/asciigraph"
	"github.com/kr/pretty"
	. "github.com/onsi/gomega"

	"github.com/slatteryjim/portfolio-analysis/data"
	. "github.com/slatteryjim/portfolio-analysis/types"
)

func Test_segmentIndexes(t *testing.T) {
	t.Run("segments must be > 0", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(func() {
			segmentIndexes(0, 0)
		}).To(Panic())
		g.Expect(func() {
			segmentIndexes(1, 0)
		}).To(Panic())
	})
	t.Run("count less than or equal to number of segments", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(segmentIndexes(0, 3)).To(Equal([]int{}))
		g.Expect(segmentIndexes(1, 3)).To(Equal([]int{1}))
		g.Expect(segmentIndexes(2, 3)).To(Equal([]int{1, 2}))
		g.Expect(segmentIndexes(3, 3)).To(Equal([]int{1, 2, 3}))
	})
	t.Run("count greater than than number of segments", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(segmentIndexes(2, 1)).To(Equal([]int{2}))
		slice := []string{"a", "b"}
		g.Expect(slice[:2]).To(Equal([]string{"a", "b"}))

		g.Expect(segmentIndexes(3, 2)).To(Equal([]int{1, 3}))
		slice = []string{"a", "b", "c"}
		g.Expect(slice[:1]).To(Equal([]string{"a"}))
		g.Expect(slice[1:3]).To(Equal([]string{"b", "c"}))

		g.Expect(segmentIndexes(4, 2)).To(Equal([]int{2, 4}))
		slice = []string{"a", "b", "c", "d"}
		g.Expect(slice[:2]).To(Equal([]string{"a", "b"}))
		g.Expect(slice[2:4]).To(Equal([]string{"c", "d"}))

		g.Expect(segmentIndexes(4, 3)).To(Equal([]int{1, 2, 4}))
		slice = []string{"a", "b", "c", "d"}
		g.Expect(slice[:1]).To(Equal([]string{"a"}))
		g.Expect(slice[1:2]).To(Equal([]string{"b"}))
		g.Expect(slice[2:4]).To(Equal([]string{"c", "d"}))

		g.Expect(segmentIndexes(7, 3)).To(Equal([]int{2, 4, 7}))
		slice = []string{"a", "b", "c", "d", "e", "f", "g"}
		g.Expect(slice[:2]).To(Equal([]string{"a", "b"}))
		g.Expect(slice[2:4]).To(Equal([]string{"c", "d"}))
		g.Expect(slice[4:7]).To(Equal([]string{"e", "f", "g"}))

		g.Expect(segmentIndexes(1000, 1)).To(Equal([]int{1000}))
		g.Expect(segmentIndexes(1000, 2)).To(Equal([]int{500, 1000}))
		g.Expect(segmentIndexes(1000, 3)).To(Equal([]int{333, 666, 1000}))
		g.Expect(segmentIndexes(1000, 4)).To(Equal([]int{250, 500, 750, 1000}))
		g.Expect(segmentIndexes(1000, 5)).To(Equal([]int{200, 400, 600, 800, 1000}))
		g.Expect(segmentIndexes(1000, 6)).To(Equal([]int{166, 333, 500, 666, 833, 1000}))
		g.Expect(segmentIndexes(1000, 7)).To(Equal([]int{142, 285, 428, 571, 714, 857, 1000}))
	})
}

func TestEvaluatePortfolios(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(EvaluatePortfolios(nil, nil)).To(BeEmpty())
	g.Expect(EvaluatePortfolios([]Combination{}, nil)).To(BeEmpty())

	t.Run("TSM", func(t *testing.T) {
		g := NewGomegaWithT(t)
		res, err := EvaluatePortfolios([]Combination{
			{Assets: []string{"TSM"}, Percentages: ReadablePercents(100)},
		}, assetMap)
		g.Expect(err).To(Succeed())
		ExpectMatchesGoldenFile(t, pretty.Sprint(res))
	})

	t.Run("two combinations", func(t *testing.T) {
		g := NewGomegaWithT(t)
		// two combinations will exercise two goroutines
		res, err := EvaluatePortfolios([]Combination{
			{Assets: []string{"TSM"}, Percentages: ReadablePercents(100)},
			{Assets: []string{"TSM", "GLD"}, Percentages: ReadablePercents(50, 50)},
		}, assetMap)
		g.Expect(err).To(Succeed())
		ExpectMatchesGoldenFile(t, pretty.Sprint(res))
	})
}

var (
	combinationsGoldenButterfly = []Combination{
		{
			Assets:      []string{"TSM", "SCV", "LTT", "STT", "GLD"},
			Percentages: []Percent{0.2, 0.2, 0.2, 0.2, 0.2},
		},
	}

	combinationsTSM = []Combination{{Assets: []string{"TSM"}, Percentages: []Percent{1}}}
)

func TestExtraPWRMetrics(t *testing.T) {
	g := NewGomegaWithT(t)
	statsReport := func(name string, returns []Percent) string {
		pwrs10 := allPWRs(returns, 10)
		pwrs30 := allPWRs(returns, 30)
		actualMinPWR10, _ := minPWR(returns, 10)
		actualMinPWR30, _ := minPWR(returns, 30)

		var sb strings.Builder
		sb.WriteString(name + ":\n\n")
		sb.WriteString(fmt.Sprintf("10-year PWRs:\n"))
		sb.WriteString(fmt.Sprintf("     min: %15v\n", actualMinPWR10))
		sb.WriteString(fmt.Sprintf("     avg: %15v\n", average(pwrs10)))
		sb.WriteString(fmt.Sprintf("  stdDev: %15v\n", standardDeviation(pwrs10)))
		sb.WriteString(fmt.Sprintf("\n"))
		sb.WriteString(fmt.Sprintf("30-year PWRs:\n"))
		sb.WriteString(fmt.Sprintf("     min: %15v\n", actualMinPWR30))
		sb.WriteString(fmt.Sprintf("     avg: %15v\n", average(pwrs30)))
		sb.WriteString(fmt.Sprintf("  stdDev: %15v\n", standardDeviation(pwrs30)))
		sb.WriteString(fmt.Sprintf("\n"))
		sb.WriteString(fmt.Sprintf("10-year PWR slope: %16v\n", slope(pwrs10)))
		sb.WriteString(fmt.Sprintf("30-year PWR slope: %16v\n", slope(pwrs30)))
		sb.WriteString(fmt.Sprintf("\n"))
		sb.WriteString(fmt.Sprintf("Overall portfolio slope: %15v\n", slope(returns)))
		return sb.String()
	}

	// 8 asset portfolio with no GoldenButterfly assets and no bonds other than very short & secure
	portfolio8way, err := portfolioReturns(
		data.PortfolioReturnsList(ParseAssets(`|ST Invest. Grade|Int'l Small|T-Bill|Wellesley|TIPS|REIT|LT STRIPS|Wellington|`)...),
		equalWeightAllocations(8))
	g.Expect(err).To(Succeed())

	t.Run("GoldenButterfly", func(t *testing.T) {
		ExpectMatchesGoldenFile(t, statsReport(t.Name(), GoldenButterfly))
	})
	t.Run("portfolio8way", func(t *testing.T) {
		ExpectMatchesGoldenFile(t, statsReport(t.Name(), portfolio8way))
	})

	t.Run("GoldenButterfly", func(t *testing.T) {
		t.Run("basic components", func(t *testing.T) {
			ExpectPlot(t, GoldenButterfly, `
  0.24 ┤         ╭╮ ╭╮                                     
  0.11 ┼ ╭─╮ ╭─╮ ││ │╰╮╭─╮ ╭╮╭╮╭╮╭╮╭╮    ╭╮ ╭╮ ╭─╮╭╮╭╮╭─╮╭ 
 -0.02 ┤╭╯ ╰─╯ ╰─╯╰╮│ ╰╯ ╰─╯││╰╯╰╯╰╯╰────╯╰─╯╰╮│ ╰╯╰╯╰╯ ╰╯ 
 -0.15 ┼╯          ╰╯       ╰╯                ╰╯           `)
			ExpectPlot(t, TSM, `
  0.32 ┤     ╭╮              ╭╮  ╭╮      ╭╮        ╭╮      
  0.21 ┤     │╰╮  ╭╮ ╭╮╭╮  ╭╮││  ││╭──╮  ││    ╭╮  ││  ╭╮╭ 
  0.09 ┤ ╭─╮ │ │ ╭╯│╭╯││╰╮╭╯││╰─╮│╰╯  │  │╰╮╭╮ │╰╮╭╯╰╮╭╯││ 
 -0.02 ┼╭╯ │ │ │╭╯ ││ ╰╯ ╰╯ ││  ╰╯    │  │ ╰╯╰╮│ ╰╯  ╰╯ ││ 
 -0.14 ┼╯  │ │ ╰╯  ╰╯       ╰╯        ╰─╮│    ││        ╰╯ 
 -0.25 ┤   ╰╮│                          ╰╯    ││           
 -0.37 ┤    ╰╯                                ╰╯            `)
			ExpectPlot(t, SCV, `
  0.45 ┤     ╭─╮                                           
  0.34 ┤     │ │     ╭╮      ╭╮    ╭╮    ╭╮        ╭╮      
  0.23 ┤     │ │ ╭╮ ╭╯│╭╮ ╭╮ │╰─╮╭─╯│ ╭╮ │╰╮   ╭─╮╭╯│ ╭╮ ╭ 
  0.12 ┼ ╭╮  │ ╰─╯╰╮│ ││╰╮│╰╮│  ││  │ │╰╮│ │╭╮ │ ││ ╰╮│╰╮│ 
  0.01 ┤╭╯╰╮ │     ╰╯ ╰╯ ││ ││  ╰╯  │╭╯ ││ ╰╯│ │ ││  ╰╯ ││ 
 -0.10 ┤│  │ │           ╰╯ ││      ╰╯  ╰╯   ╰╮│ ╰╯     ╰╯ 
 -0.21 ┼╯  │ │              ╰╯                ││           
 -0.32 ┤   ╰─╯                                ╰╯            `)
			ExpectPlot(t, LTT, `
  0.36 ┤            ╭╮                                     
  0.26 ┼            ││ ╭╮        ╭╮                        
  0.16 ┤            ││ │╰╮       ││   ╭╮      ╭╮ ╭╮ ╭╮     
  0.05 ┤╭─╮   ╭╮    ││╭╯ │ ╭╮╭╮╭╮││╭─╮││╭╮   ╭╯│╭╯│ ││ ╭╮╭ 
 -0.05 ┤│ │  ╭╯│    │╰╯  │╭╯╰╯╰╯││╰╯ ││╰╯╰───╯ ││ ╰╮│╰─╯╰╯ 
 -0.15 ┼╯ ╰──╯ ╰────╯    ╰╯     ╰╯   ╰╯        ╰╯  ╰╯        `)
			ExpectPlot(t, STT, `
  0.17 ┼            ╭╮                                     
  0.06 ┤╭╮          ││╭──╮ ╭╮╭╮  ╭╮    ╭╮     ╭╮           
 -0.05 ┼╯╰──────────╯╰╯  ╰─╯╰╯╰──╯╰────╯╰─────╯╰────────── `)
			ExpectPlot(t, GLD, `
  0.97 ┤                   ╭╮                                                                                 
  0.84 ┤                   ││                                                                                 
  0.71 ┤                   │╰╮                                                                                
  0.58 ┼      ╭───╮       ╭╯ │                                                                                
  0.45 ┤     ╭╯   │       │  │                                                                                
  0.32 ┤    ╭╯    │      ╭╯  │                                            ╭─╮     ╭──╮  ╭──╮                  
  0.19 ┤   ╭╯     ╰╮   ╭─╯   ╰╮         ╭────╮         ╭─╮               ╭╯ ╰─╮╭──╯  ╰──╯  ╰╮         ╭──╮  ╭ 
  0.06 ┤ ╭─╯       │  ╭╯      │  ╭─╮   ╭╯    │        ╭╯ ╰────╮   ╭──╮ ╭─╯    ╰╯            ╰──╮  ╭╮ ╭╯  ╰──╯ 
 -0.07 ┤╭╯         │ ╭╯       ╰╮ │ ╰╮ ╭╯     ╰────────╯       ╰───╯  ╰─╯                       ╰╮╭╯╰─╯        
 -0.20 ┼╯          ╰─╯         │╭╯  ╰─╯                                                         ╰╯            
 -0.33 ┤                       ╰╯
`,
				asciigraph.Height(10), asciigraph.Width(len(GLD)*2))
		})

		ExpectPlot(t, allPWRs(GoldenButterfly, 10), `
 0.070 ┤         ╭╮ ╭╮                            
 0.060 ┤     ╭───╯│ ││╭─╮    ╭╮  ╭╮      ╭╮       
 0.050 ┤╭─╮ ╭╯    ╰╮│╰╯ ╰╮╭──╯╰──╯╰─╮  ╭─╯╰──╮╭── 
 0.040 ┤│ ╰─╯      ╰╯    ╰╯         ╰╮╭╯     ╰╯   
 0.030 ┤│                            ╰╯           
 0.019 ┼╯                                         
`)
		ExpectPlot(t, allPWRs(GoldenButterfly, 20), `
 0.067 ┤         ╭╮ ╭╮                  
 0.053 ┤╭──╮╭────╯╰─╯╰───╮╭─╮╭──────╮   
 0.039 ┼╯  ╰╯            ╰╯ ╰╯      ╰──
`)
		ExpectPlot(t, allPWRs(GoldenButterfly, 30), `
 0.066 ┤            ╭╮        
 0.054 ┤╭─╮ ╭─────╮╭╯╰──╮     
 0.042 ┼╯ ╰─╯     ╰╯    ╰────
`)
	})

	t.Run("8-way", func(t *testing.T) {
		ExpectPlot(t, allPWRs(portfolio8way, 10), `
 0.080 ┼╮                         
 0.069 ┤╰╮   ╭╮  ╭╮               
 0.058 ┤ │╭─╮│╰─╮│╰─╮   ╭─╮       
 0.047 ┤ ╰╯ ╰╯  ╰╯  ╰╮╭─╯ ╰──╮╭── 
 0.035 ┤             ╰╯      ╰╯
`)
		ExpectPlot(t, allPWRs(portfolio8way, 20), `
 0.079 ┼╮               
 0.067 ┤╰╮              
 0.056 ┤ ╰──╮╭────╮   ╭ 
 0.045 ┤    ╰╯    ╰───╯
`)
		ExpectPlot(t, allPWRs(portfolio8way, 30), `
 0.074 ┼╮     
 0.062 ┤╰╮╭╮  
 0.050 ┤ ╰╯╰─
`)
	})
}

func equalWeightAllocations(n int) []Percent {
	res := make([]Percent, n)
	amount := 1.0 / Percent(n)
	for i := 0; i < n; i++ {
		res[i] = amount
	}
	return res
}

func ParseAssets(assetsPipeDelimited string) []string {
	res := assetsPipeDelimited
	res = strings.TrimPrefix(res, "|")
	res = strings.TrimSuffix(res, "|")
	return strings.Split(res, "|")
}

func BenchmarkCopy(b *testing.B) {
	b.Run("Built-in", func(b *testing.B) {
		from := make([]byte, b.N)
		to := make([]byte, b.N)
		b.ReportAllocs()
		b.ResetTimer()
		b.SetBytes(1)
		copy(to, from)
	})
	b.Run("manual", func(b *testing.B) {
		from := make([]byte, b.N)
		to := make([]byte, b.N)
		b.ReportAllocs()
		b.ResetTimer()
		b.SetBytes(1)
		for i := 0; i < b.N; i++ {
			to[i] = from[i]
		}
	})
	b.Run("manual-append", func(b *testing.B) {
		from := make([]byte, b.N)
		to := make([]byte, 0, b.N)
		b.ReportAllocs()
		b.ResetTimer()
		b.SetBytes(1)
		for i := 0; i < b.N; i++ {
			to = append(to, from[i])
		}
	})
}

// go test -run=^$ -bench=Benchmark_evaluatePortfolios_GoldenButterfly$ --benchtime=10s
//
// Benchmark_evaluatePortfolios_GoldenButterfly-12           431221             27260 ns/op
func Benchmark_evaluatePortfolios_GoldenButterfly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := evaluatePortfolios(combinationsGoldenButterfly, assetMap)
		if err != nil {
			b.Fatal(err.Error())
		}
	}
}

// go test -run=^$ -bench=Benchmark_evaluatePortfolios_TSM$ --benchtime=10s
//
// Benchmark_evaluatePortfolios_TSM-12       423140             26658 ns/op
func Benchmark_evaluatePortfolios_TSM(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := evaluatePortfolios(combinationsTSM, assetMap)
		if err != nil {
			b.Fatal(err.Error())
		}
	}
}

// Benchmark some portfolio evaluation metrics, want to order them cheapest to most expensive.
//
// $ go test -bench ^BenchmarkPortfolioEvaluationMetrics$ -run ^$ -benchtime=5s
//
// BenchmarkPortfolioEvaluationMetrics/average-12                         189537374                31.6 ns/op
// BenchmarkPortfolioEvaluationMetrics/standardDeviation-12                 3170172              1887 ns/op
// BenchmarkPortfolioEvaluationMetrics/minPWRAndSWR30-12                    2038736              2944 ns/op
// BenchmarkPortfolioEvaluationMetrics/baselineLongTermReturn-12            1299754              4564 ns/op
// BenchmarkPortfolioEvaluationMetrics/drawdownScores-12                    1205773              4992 ns/op
// BenchmarkPortfolioEvaluationMetrics/baselineShortTermReturn-12            999006              5614 ns/op
// BenchmarkPortfolioEvaluationMetrics/startDateSensitivity-12               998498              5825 ns/op
func BenchmarkPortfolioEvaluationMetrics(b *testing.B) {
	gbReturns := mustGoldenButterflyStat().MustReturns()
	b.Run("minPWRAndSWR30", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			minPWRAndSWR(gbReturns, 30)
		}
	})
	b.Run("drawdownScores", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			drawdownScores(gbReturns)
		}
	})
	b.Run("average", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			average(gbReturns)
		}
	})
	b.Run("baselineLongTermReturn", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			baselineLongTermReturn(gbReturns)
		}
	})
	b.Run("baselineShortTermReturn", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			baselineShortTermReturn(gbReturns)
		}
	})
	b.Run("standardDeviation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			standardDeviation(gbReturns)
		}
	})
	b.Run("startDateSensitivity", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			startDateSensitivity(gbReturns)
		}
	})
}
