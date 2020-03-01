package portfolio_analysis

import (
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
