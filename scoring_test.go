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
			AvgReturn:            0.0745235294117647,
			BaselineLTReturn:     0.030599414622012988,
			BaselineSTReturn:     -0.02905140217935165,
			PWR30:                0.03237620200614041,
			SWR30:                0.037860676066939845,
			StdDev:               0.17165685399889558,
			UlcerScore:           26.989112643639167,
			DeepestDrawdown:      -0.5225460399999999,
			LongestDrawdown:      13,
			StartDateSensitivity: 0.3164468857748042,
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
			AvgReturn:            0.0745235294117647,
			BaselineLTReturn:     0.030599414622012988,
			BaselineSTReturn:     -0.02905140217935165,
			PWR30:                0.03237620200614041,
			SWR30:                0.037860676066939845,
			StdDev:               0.17165685399889558,
			UlcerScore:           26.989112643639167,
			DeepestDrawdown:      -0.5225460399999999,
			LongestDrawdown:      13,
			StartDateSensitivity: 0.3164468857748042,
		},
		{
			Assets:               []string{"TSM", "GLD"},
			Percentages:          readablePercents(50, 50),
			AvgReturn:            0.06387058823529411,
			BaselineLTReturn:     0.03547746755212633,
			BaselineSTReturn:     -0.00519786860232474,
			PWR30:                0.028200924227015447,
			SWR30:                0.04066238595346762,
			StdDev:               0.13238467721907654,
			UlcerScore:           9.965815083766593,
			DeepestDrawdown:      -0.2593209743740048,
			LongestDrawdown:      6,
			StartDateSensitivity: 0.21609785983534202,
		},
	}))
}
