package portfolio_analysis

import (
	"math"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_portfolioReturns(t *testing.T) {

	t.Run("errors", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, err := portfolioReturns(nil, nil)
		g.Expect(err).To(MatchError("percentages must sum to 100%, got 0"))

		_, err = portfolioReturns(nil, []float64{100})
		g.Expect(err).To(MatchError("lists must have the same length: percentages (1), returnsList (0)"))
	})

	t.Run("success", func(t *testing.T) {
		g := NewGomegaWithT(t)

		// simply one asset, one year
		g.Expect(portfolioReturns(
			[][]float64{
				{1},
			},
			[]float64{100}),
		).To(Equal(
			[]float64{1},
		))

		// two assets, two years, 50%/50%
		g.Expect(portfolioReturns(
			[][]float64{
				{10, 20},
				{5, 10},
			},
			[]float64{50, 50}),
		).To(Equal(
			[]float64{7.5, 15},
		))

		// TSM asset, 100%, yields itself
		g.Expect(portfolioReturns(
			[][]float64{
				TSM,
			},
			[]float64{100}),
		).To(Equal(
			TSM,
		))

		// GoldenButterfly
		g.Expect(portfolioReturns([][]float64{TSM, SCV, LTT, STT, GLD}, []float64{20, 20, 20, 20, 20})).To(Equal(
			[]float64{-15.334, 1.7320000000000002, 10.07, 11.374, -1.7219999999999995, -6.462, 9.36, 14.0, 0.9680000000000004, 3.404, 23.724000000000004, 2.592, -9.71, 19.733999999999998, 7.496, -1.0420000000000007, 18.338, 14.08, 0.018000000000000682, 4.668000000000002, 8.674, -8.572, 15.636000000000001, 6.130000000000001, 10.984000000000002, -4.6499999999999995, 17.900000000000002, 5.255999999999999, 11.162000000000003, 5.816, 1.6939999999999995, 3.1079999999999997, 1.7060000000000002, -0.1719999999999997, 16.666, 6.406000000000001, 3.846, 9.778000000000002, 4.998, -6.696, 11.048000000000002, 14.606, 4.112000000000001, 7.602, 4.372, 8.816, -4.050000000000001, 7.442, 8.438, -5.602, 15.346000000000004},
		))
	})
}

func Test_harmonicMean(t *testing.T) {

	t.Run("errors", func(t *testing.T) {
		g := NewGomegaWithT(t)

		// NaN
		g.Expect(math.IsNaN(harmonicMean(nil))).To(BeTrue())
		g.Expect(math.IsNaN(harmonicMean([]float64{}))).To(BeTrue())

		// panic
		g.Expect(func() {
			harmonicMean([]float64{0})
		}).To(Panic())
	})

	t.Run("success", func(t *testing.T) {
		g := NewGomegaWithT(t)
		// got these numbers using the HARMMEAN function in a Google Spreadsheets
		g.Expect(harmonicMean([]float64{1})).To(Equal(1.0))
		g.Expect(harmonicMean([]float64{1, 2})).To(Equal(1.3333333333333333))
		g.Expect(harmonicMean([]float64{1, 2, 3, 4, 5})).To(Equal(2.18978102189781))
	})
}

var (
	// tested in spreadsheet: https://docs.google.com/spreadsheets/d/14bxTQncj8BIUtQghpiQM0ZyeGsNUYQGFtNx2yOIkEEE/edit#gid=1013681666
	sampleReturns = []float64{-10.28, 0.90, 17.63, 17.93, -18.55, -28.42, 38.42, 26.54, -2.68, 9.23, 25.51, 33.62, -3.79, 18.66, 23.42, 3.01, 32.51, 16.05, 2.23, 17.89, 29.12, -6.22, 34.15, 8.92, 10.62, -0.17, 35.79, 20.96, 30.99, 23.26, 23.81, -10.57, -10.89, -20.95, 31.42, 12.61, 6.09, 15.63, 5.57, -36.99, 28.83, 17.26, 1.08, 16.38, 33.52, 12.56, 0.39, 12.66, 21.17, -5.17, 30.80}
)

func Test_swr(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(swr([]float64{})).To(Equal(1.0))

	// briefly prove SWR for a 1 year 10% return
	{
		fixedWithdrawal := 0.5238095238095238
		g.Expect(swr([]float64{10})).To(Equal(fixedWithdrawal))
		initial := 1.0
		remaining := initial - fixedWithdrawal // withdraw first year's amount
		remaining *= 1.10                      // apply first year's growth
		remaining -= fixedWithdrawal           // withdraw second year's amount
		g.Expect(remaining).To(Equal(0.0))     // we exactly exhausted the account
	}

	g.Expect(swr(sampleReturns)).To(Equal(0.06622907313022618))
}

func Test_pwr(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(pwr([]float64{})).To(Equal(0.0))

	// briefly prove PWR for a 1 year 10% return
	{
		fixedWithdrawal := 0.04761904761904764
		g.Expect(pwr([]float64{10})).To(Equal(fixedWithdrawal))
		initial := 1.0
		remaining := initial - fixedWithdrawal // withdraw first year's amount
		remaining *= 1.10                      // apply first year's growth
		remaining -= fixedWithdrawal           // withdraw second year's amount
		g.Expect(remaining).To(Equal(initial)) // we exactly exhausted the account
	}

	g.Expect(pwr(sampleReturns)).To(Equal(0.06574303881824275))
}

func Test_minPWR(t *testing.T) {
	g := NewGomegaWithT(t)

	verify := func(returns []float64, nYears int, expectedPWR float64, expectedIndex int) {
		t.Helper()
		rate, n := minPWR(returns, nYears)
		g.Expect(rate).To(Equal(expectedPWR), "rate")
		g.Expect(n).To(Equal(expectedIndex), "index")
	}

	g.Expect(func() {
		minPWR(nil, 1)
	}).To(Panic())

	verify(nil, 0, 0, 0)

	// length 1
	verify([]float64{10}, 1, pwr([]float64{10}), 0)
	verify([]float64{20}, 1, pwr([]float64{20}), 0)

	// length 2
	verify([]float64{10, 20}, 1, pwr([]float64{10}), 0)
	verify([]float64{10, 20}, 2, pwr([]float64{10, 20}), 0)

	// length 3
	verify([]float64{10, -5, 30}, 1, pwr([]float64{-5}), 1)
	verify([]float64{10, -5, 30}, 2, pwr([]float64{10, -5}), 0)
	verify([]float64{10, -5, 30}, 3, pwr([]float64{10, -5, 30}), 0)

	// length 4
	verify([]float64{10, -5, 10, -20}, 1, pwr([]float64{-20}), 3)
	verify([]float64{10, -5, 10, -20}, 2, pwr([]float64{10, -20}), 2)
	verify([]float64{10, -5, 10, -20}, 3, pwr([]float64{-5, 10, -20}), 1)
	verify([]float64{10, -5, 10, -20}, 4, pwr([]float64{10, -5, 10, -20}), 0)

	verify(GoldenButterfly, 10, 0.01945631963862428, 0)
	verify(GoldenButterfly, 20, 0.038590835351436564, 0)
	verify(GoldenButterfly, 30, 0.04224334655073258, 0)
	verify(GoldenButterfly, 40, 0.042057016507784345, 0)
	verify(GoldenButterfly, 50, 0.04288283017428213, 0)
}

func Test_minSWR(t *testing.T) {
	g := NewGomegaWithT(t)

	verify := func(returns []float64, nYears int, expectedSWR float64, expectedIndex int) {
		t.Helper()
		rate, n := minSWR(returns, nYears)
		g.Expect(rate).To(Equal(expectedSWR), "rate")
		g.Expect(n).To(Equal(expectedIndex), "index")
	}

	g.Expect(func() {
		minSWR(nil, 1)
	}).To(Panic())

	verify(nil, 0, 0, 0)

	// length 1
	verify([]float64{10}, 1, swr([]float64{10}), 0)
	verify([]float64{20}, 1, swr([]float64{20}), 0)

	// length 2
	verify([]float64{10, 20}, 1, swr([]float64{10}), 0)
	verify([]float64{10, 20}, 2, swr([]float64{10, 20}), 0)

	// length 3
	verify([]float64{10, -5, 30}, 1, swr([]float64{-5}), 1)
	verify([]float64{10, -5, 30}, 2, swr([]float64{10, -5}), 0)
	verify([]float64{10, -5, 30}, 3, swr([]float64{10, -5, 30}), 0)

	// length 4
	verify([]float64{10, -5, 10, -20}, 1, swr([]float64{-20}), 3)
	verify([]float64{10, -5, 10, -20}, 2, swr([]float64{10, -20}), 2)
	verify([]float64{10, -5, 10, -20}, 3, swr([]float64{-5, 10, -20}), 1)
	verify([]float64{10, -5, 10, -20}, 4, swr([]float64{10, -5, 10, -20}), 0)

	verify(GoldenButterfly, 10, 0.09331636419042066, 0)
	verify(GoldenButterfly, 20, 0.06261394175862438, 0)
	verify(GoldenButterfly, 30, 0.05304896125102126, 0)
	verify(GoldenButterfly, 40, 0.04879022342090543, 0)
	verify(GoldenButterfly, 50, 0.04665043650688064, 0)
}

func Test_subSlices(t *testing.T) {
	g := NewGomegaWithT(t)

	// length 0
	g.Expect(subSlices(nil, 0)).To(BeEmpty())

	// length 1
	g.Expect(subSlices([]float64{1}, 1)).To(Equal([][]float64{{1}}))
	g.Expect(func() {
		subSlices([]float64{1}, 2) // n is greater than length of the slice
	}).To(Panic())

	// length 2
	g.Expect(subSlices([]float64{1, 2}, 1)).To(Equal([][]float64{{1}, {2}}))
	g.Expect(subSlices([]float64{1, 2}, 2)).To(Equal([][]float64{{1, 2}}))

	// length 3
	g.Expect(subSlices([]float64{1, 2, 3}, 1)).To(Equal([][]float64{{1}, {2}, {3}}))
	g.Expect(subSlices([]float64{1, 2, 3}, 2)).To(Equal([][]float64{{1, 2}, {2, 3}}))
	g.Expect(subSlices([]float64{1, 2, 3}, 3)).To(Equal([][]float64{{1, 2, 3}}))
}

func Test_leadingDrawdownSequence(t *testing.T) {
	g := NewGomegaWithT(t)

	verify := func(returns, expectedSequence []float64, expectedEnded bool) {
		t.Helper()
		sequence, ended := leadingDrawdownSequence(returns)
		g.Expect(sequence).To(Equal(expectedSequence), "sequence")
		g.Expect(ended).To(Equal(expectedEnded), "ended")
	}

	// empty
	verify(nil, []float64{}, false)
	verify([]float64{}, []float64{}, false)

	// doesn't start with a drawdown
	verify([]float64{1}, []float64{}, true)
	verify([]float64{0, -1, 2}, []float64{}, true)

	// starts with a drawdown
	verify([]float64{-1}, []float64{0.99}, false)
	verify([]float64{-1, 2}, []float64{0.99}, true)
	verify([]float64{-1, -1, 3}, []float64{0.99, 0.9801}, true)
	verify([]float64{-1, -1, 3, -5, -50}, []float64{0.99, 0.9801}, true)

	verify([]float64{-50, 100}, []float64{0.50}, true)

}

func Test_drawdowns(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(drawdowns(nil)).To(BeEmpty())
	g.Expect(drawdowns([]float64{0})).To(BeEmpty())
	g.Expect(drawdowns([]float64{1})).To(BeEmpty())
	g.Expect(drawdowns([]float64{-1})).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []float64{0.99}, recovered: false},
	}))
	g.Expect(drawdowns([]float64{-1, 2})).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []float64{0.99}, recovered: true},
	}))
	g.Expect(drawdowns([]float64{-1, -1, 3})).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []float64{0.99, 0.9801}, recovered: true},
		{startIndex: 1, cumulativeReturns: []float64{0.99}, recovered: true},
	}))
	g.Expect(drawdowns([]float64{-1, -1, 1})).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []float64{0.99, 0.9801, 0.989901}, recovered: false},
		{startIndex: 1, cumulativeReturns: []float64{0.99, 0.9999}, recovered: false},
	}))
	g.Expect(drawdowns([]float64{-1, 3, -1, -1, 3})).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []float64{0.99}, recovered: true},
		{startIndex: 2, cumulativeReturns: []float64{0.99, 0.9801}, recovered: true},
		{startIndex: 3, cumulativeReturns: []float64{0.99}, recovered: true},
	}))
	g.Expect(drawdowns([]float64{-1, -1, 3, -5, -50, 100})).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []float64{0.99, 0.9801}, recovered: true},
		{startIndex: 1, cumulativeReturns: []float64{0.99}, recovered: true},
		{startIndex: 3, cumulativeReturns: []float64{0.95, 0.475, 0.95}, recovered: false},
		{startIndex: 4, cumulativeReturns: []float64{0.50}, recovered: true},
	}))
}

func Test_ulcerScore(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(ulcerScore(nil, false)).To(Equal(0.0))
	g.Expect(ulcerScore(nil, true)).To(Equal(0.0))
	g.Expect(ulcerScore([]float64{0.99}, true)).To(Equal(0.10000000000000009))
	g.Expect(ulcerScore([]float64{0.99}, false)).To(Equal(0.20000000000000018))
	g.Expect(ulcerScore([]float64{0.90}, true)).To(Equal(0.9999999999999998))
	g.Expect(ulcerScore([]float64{0.90, 0.90}, true)).To(Equal(1.9999999999999996))
	g.Expect(ulcerScore([]float64{0.90, 0.80}, true)).To(Equal(2.999999999999999))

	dd, _ := leadingDrawdownSequence(GoldenButterfly)
	g.Expect(dd).To(Equal([]float64{0.84666, 0.8613241511999999, 0.9480594932258399}))
	g.Expect(ulcerScore(dd, true)).To(Equal(3.4395635557416018))
}

func Test_drawdownScores(t *testing.T) {
	g := NewGomegaWithT(t)

	verify := func(returns []float64, expectedUlcer, expectedMaxDrawdown float64, expectedMaxDuration int) {
		ulcer, maxDD, maxDur := drawdownScores(returns)
		g.Expect(ulcer).To(Equal(expectedUlcer), "maxUlcerScore")
		g.Expect(maxDD).To(Equal(expectedMaxDrawdown), "maxDrawdown")
		g.Expect(maxDur).To(Equal(expectedMaxDuration), "maxDuration")
	}

	verify(nil, 0, 0, 0)
	verify([]float64{}, 0, 0, 0)
	verify([]float64{-1}, 0.20000000000000018, -0.010000000000000009, 1)
	verify([]float64{-1, 2}, 0.10000000000000009, -0.010000000000000009, 1)
	verify([]float64{-1, 2, -1, -3}, 0.9940000000000015, -0.03970000000000007, 2)
	verify([]float64{-1, 2, -1, -3, 10}, 0.4970000000000008, -0.03970000000000007, 2)

	verify([]float64{-10, 30}, 0.9999999999999998, -0.09999999999999998, 1)
	verify([]float64{-20, 30}, 1.9999999999999996, -0.19999999999999996, 1)
	verify([]float64{-10, -20, 40}, 3.799999999999999, -0.2799999999999999, 2)
	verify([]float64{-10, 30, -10, -20, 30}, 8.879999999999995, -0.2799999999999999, 3)
}

func Test_cagr(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(cagr(nil)).To(Equal(0.0))
	g.Expect(cagr([]float64{})).To(Equal(0.0))
	g.Expect(cagr([]float64{1})).To(Equal(0.010000000000000009))
	// prove it's correct
	{
		cagrValue := 0.029805806936433976
		g.Expect(cagr([]float64{1, 5})).To(Equal(cagrValue))
		cumulativeTwoYears := 1.0605
		// compound the CAGR value
		g.Expect((1 + cagrValue) * (1 + cagrValue)).To(Equal(cumulativeTwoYears))
		// compound the original returns, arrive at the same cumulative value
		g.Expect(1.01 * 1.05).To(Equal(cumulativeTwoYears))
	}
	g.Expect(cagr([]float64{5, 5, 5, 5, 5})).To(Equal(0.050000000000000044))
}
