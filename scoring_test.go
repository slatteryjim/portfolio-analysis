package portfolio_analysis

import (
	"testing"

	. "github.com/onsi/gomega"
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
	g.Expect(EvaluatePortfolios([]Permutation{}, nil)).To(BeEmpty())

	g.Expect(EvaluatePortfolios([]Permutation{
		{Assets: []string{"TSM"}, Percentages: readablePercents(100)},
	}, assetMap)).To(Equal([]*PortfolioStat{
		{
			Assets:               []string{"TSM"},
			Percentages:          readablePercents(100),
			AvgReturn:            0.07453082614598813,
			BaselineLTReturn:     0.0306081363792714,
			BaselineSTReturn:     -0.02907904796851324,
			PWR30:                0.03237787319823412,
			SWR30:                0.03786147243197951,
			StdDev:               0.17165466213991304,
			UlcerScore:           26.990140749914836,
			DeepestDrawdown:      -0.5225438230658536,
			LongestDrawdown:      13,
			StartDateSensitivity: 0.3164541256493081,
		},
	}))

	// two permutations will exercise two goroutines
	g.Expect(EvaluatePortfolios([]Permutation{
		{Assets: []string{"TSM"}, Percentages: readablePercents(100)},
		{Assets: []string{"TSM", "GLD"}, Percentages: readablePercents(50, 50)},
	}, assetMap)).To(Equal([]*PortfolioStat{
		{
			Assets:               []string{"TSM"},
			Percentages:          readablePercents(100),
			AvgReturn:            0.07453082614598813,
			BaselineLTReturn:     0.0306081363792714,
			BaselineSTReturn:     -0.02907904796851324,
			PWR30:                0.03237787319823412,
			SWR30:                0.03786147243197951,
			StdDev:               0.17165466213991304,
			UlcerScore:           26.990140749914836,
			DeepestDrawdown:      -0.5225438230658536,
			LongestDrawdown:      13,
			StartDateSensitivity: 0.3164541256493081,
		},
		{
			Assets:               []string{"TSM", "GLD"},
			Percentages:          readablePercents(50, 50),
			AvgReturn:            0.06387494536221017,
			BaselineLTReturn:     0.035477861130724264,
			BaselineSTReturn:     -0.0051889628058078285,
			PWR30:                0.028202600645605407,
			SWR30:                0.04066373587232011,
			StdDev:               0.13238370526536952,
			UlcerScore:           9.965222445166578,
			DeepestDrawdown:      -0.25929582951888575,
			LongestDrawdown:      6,
			StartDateSensitivity: 0.21610371811517437,
		},
	}))
}

var (
	permutationsGoldenButterfly = []Permutation{
		{
			Assets:      []string{"TSM", "SCV", "LTT", "STT", "GLD"},
			Percentages: []Percent{0.2, 0.2, 0.2, 0.2, 0.2},
		},
	}

	permutationsTSM = []Permutation{{Assets: []string{"TSM"}, Percentages: []Percent{1}}}
)

// go test -run=^$ -bench=Benchmark_evaluatePortfolios_GoldenButterfly$ --benchtime=10s
//
// Benchmark_evaluatePortfolios_GoldenButterfly-12           431221             27260 ns/op
func Benchmark_evaluatePortfolios_GoldenButterfly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := evaluatePortfolios(permutationsGoldenButterfly, assetMap)
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
		_, err := evaluatePortfolios(permutationsTSM, assetMap)
		if err != nil {
			b.Fatal(err.Error())
		}
	}
}
