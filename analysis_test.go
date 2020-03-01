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

	// tested in spreadsheet: https://docs.google.com/spreadsheets/d/14bxTQncj8BIUtQghpiQM0ZyeGsNUYQGFtNx2yOIkEEE/edit#gid=1013681666
	sample := []float64{-10.28, 0.90, 17.63, 17.93, -18.55, -28.42, 38.42, 26.54, -2.68, 9.23, 25.51, 33.62, -3.79, 18.66, 23.42, 3.01, 32.51, 16.05, 2.23, 17.89, 29.12, -6.22, 34.15, 8.92, 10.62, -0.17, 35.79, 20.96, 30.99, 23.26, 23.81, -10.57, -10.89, -20.95, 31.42, 12.61, 6.09, 15.63, 5.57, -36.99, 28.83, 17.26, 1.08, 16.38, 33.52, 12.56, 0.39, 12.66, 21.17, -5.17, 30.80}
	g.Expect(swr(sample)).To(Equal(0.06622907313022618))
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

	// tested in spreadsheet: https://docs.google.com/spreadsheets/d/14bxTQncj8BIUtQghpiQM0ZyeGsNUYQGFtNx2yOIkEEE/edit#gid=1013681666
	sample := []float64{-10.28, 0.90, 17.63, 17.93, -18.55, -28.42, 38.42, 26.54, -2.68, 9.23, 25.51, 33.62, -3.79, 18.66, 23.42, 3.01, 32.51, 16.05, 2.23, 17.89, 29.12, -6.22, 34.15, 8.92, 10.62, -0.17, 35.79, 20.96, 30.99, 23.26, 23.81, -10.57, -10.89, -20.95, 31.42, 12.61, 6.09, 15.63, 5.57, -36.99, 28.83, 17.26, 1.08, 16.38, 33.52, 12.56, 0.39, 12.66, 21.17, -5.17, 30.80}
	g.Expect(pwr(sample)).To(Equal(0.06574303881824275))
}
