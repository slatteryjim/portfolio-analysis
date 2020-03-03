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

	verify(TSM, 30, 0.03237620200614041, 0)
	verify(SCV, 30, 0.03803355302289947, 0)
	verify(GLD, 30, -0.015629074083395443, 6)
	verify(LTT, 30, 0.02167080631193789, 0)
	verify(STT, 30, 0.01714682750442063, 21)
	verify(STB, 30, 0.0226647988703317, 0)

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

	verify(TSM, 30, 0.037860676066939845, 0)
	verify(SCV, 30, 0.04290428968981267, 0)
	verify(GLD, 30, 0.012375454483082474, 11)
	verify(LTT, 30, 0.034180844746692515, 0)
	verify(STT, 30, 0.03974452972823921, 3)
	verify(STB, 30, 0.0396442756302357, 0)

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

	g.Expect(cagr(TSM)).To(Equal(0.059240605917942224))
	g.Expect(cagr(SCV)).To(Equal(0.07363836101130472))
	g.Expect(cagr(LTT)).To(Equal(0.036986778393646835))
	g.Expect(cagr(STT)).To(Equal(0.018215249078317397))
	g.Expect(cagr(STB)).To(Equal(0.02224127904840234))
	g.Expect(cagr(GLD)).To(Equal(0.029259375673007515))
	g.Expect(cagr(GoldenButterfly)).To(Equal(0.05352050963712207))
}
