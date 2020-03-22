package portfolio_analysis

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_portfolioReturns(t *testing.T) {
	g := NewGomegaWithT(t)

	// portfolioReturns and PortfolioTradingSimulation should always return identical results, for rebalanceFactor=1
	portfolioReturnsProxy := func(returnsList [][]Percent, targetAllocations []Percent) ([]Percent, error) {
		t.Helper()
		a, errA := portfolioReturns(returnsList, targetAllocations)
		b, errB := PortfolioTradingSimulation(returnsList, targetAllocations, 1)
		g.Expect(len(a)).To(Equal(len(b)), "returns length")
		for i := range a {
			g.Expect(a[i]).To(BeNumerically("~", b[i], 0.000000000000001), "returns element "+strconv.Itoa(i))
		}
		if errA == nil {
			g.Expect(errB).To(BeNil(), "returned error is nil")
		} else {
			g.Expect(errA).To(Equal(errB), "returned error")
		}
		return a, errA
	}

	t.Run("errors", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, err := portfolioReturnsProxy(nil, nil)
		g.Expect(err).To(MatchError("targetAllocations must sum to 100%, got 0%"))

		_, err = portfolioReturnsProxy(nil, readablePercents(100))
		g.Expect(err).To(MatchError("lists must have the same length: targetAllocations (1), returnsList (0)"))
	})

	t.Run("success", func(t *testing.T) {
		g := NewGomegaWithT(t)

		// simply one asset, one year
		g.Expect(portfolioReturnsProxy(
			[][]Percent{
				readablePercents(1),
			},
			readablePercents(100)),
		).To(Equal(
			readablePercents(1),
		))

		// two assets, two years, 50%/50%
		g.Expect(portfolioReturnsProxy(
			[][]Percent{
				readablePercents(10, 20),
				readablePercents(5, 10),
			},
			readablePercents(50, 50)),
		).To(Equal(
			[]Percent{0.07500000000000001, 0.15000000000000002},
		))

		// TSM asset, 100%, yields itself
		g.Expect(portfolioReturnsProxy(
			[][]Percent{
				TSM,
			},
			readablePercents(100)),
		).To(Equal(
			TSM,
		))

		// GoldenButterfly
		g.Expect(portfolioReturnsProxy([][]Percent{TSM, SCV, LTT, STT, GLD}, readablePercents(20, 20, 20, 20, 20))).To(Equal(
			[]Percent{-0.15334, 0.017320000000000002, 0.10070000000000001, 0.11374000000000001, -0.01721999999999997, -0.06462000000000001, 0.09359999999999999, 0.14, 0.009680000000000005, 0.03404, 0.23724000000000003, 0.025920000000000012, -0.0971, 0.19734, 0.07496000000000001, -0.010419999999999999, 0.18338000000000004, 0.14079999999999998, 0.00018000000000000654, 0.04668, 0.08674000000000001, -0.08571999999999999, 0.15636, 0.06130000000000001, 0.10984000000000001, -0.0465, 0.17900000000000002, 0.05256, 0.11162000000000002, 0.05815999999999999, 0.016940000000000004, 0.03108, 0.017060000000000006, -0.0017199999999999924, 0.16666, 0.06406, 0.03846000000000001, 0.09778, 0.04998, -0.06696000000000002, 0.11048000000000002, 0.14606000000000002, 0.04112000000000001, 0.07602, 0.04372000000000001, 0.08816000000000002, -0.0405, 0.07442000000000001, 0.08438000000000001, -0.05602, 0.15345999999999999},
		))
	})
}

func TestPortfolioTradingSimulation(t *testing.T) {
	var (
		assets = [][]Percent{
			readablePercents(20, 20, 20), // first asset greatly outperforms second
			readablePercents(0, 0, 0),
		}
		targetAllocations = readablePercents(50, 50)

		// first year's returns is always the same
		firstReturn Percent = 0.10000000000000009
	)
	t.Run("rebalanceFactor = 0, so no rebalancing -- the first asset can grow faster untouched", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(PortfolioTradingSimulation(assets, targetAllocations, 0)).To(Equal(
			[]Percent{firstReturn, 0.1090909090909089, 0.11803278688524577},
		))
	})
	t.Run("rebalanceFactor = 0.5", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(PortfolioTradingSimulation(assets, targetAllocations, 0.5)).To(Equal(
			[]Percent{firstReturn, 0.10454545454545427, 0.10679012345679006},
		))
	})
	t.Run("rebalanceFactor = 1", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(PortfolioTradingSimulation(assets, targetAllocations, 1)).To(Equal(
			[]Percent{firstReturn, 0.09999999999999987, 0.10000000000000009},
		))
		// just a sanity check, this matches portfolioReturns, minus some float weirdness
		g.Expect(portfolioReturns(assets, targetAllocations)).To(Equal(
			[]Percent{0.1, 0.1, 0.1},
		))
	})
	t.Run("rebalanceFactor = 1.5", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(PortfolioTradingSimulation(assets, targetAllocations, 1.5)).To(Equal(
			[]Percent{firstReturn, 0.09545454545454546, 0.09771784232365155},
		))
	})
	t.Run("rebalanceFactor = 2.0", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(PortfolioTradingSimulation(assets, targetAllocations, 2.0)).To(Equal(
			[]Percent{firstReturn, 0.09090909090909105, 0.10000000000000009},
		))
	})
	t.Run("rebalanceFactor = 3.0", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(PortfolioTradingSimulation(assets, targetAllocations, 3.0)).To(Equal(
			[]Percent{firstReturn, 0.08181818181818179, 0.11848739495798322},
		))
	})
	t.Run("rebalanceFactor = 4.00 -- too extreme, panic as allocation goes below zero", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(func() {
			PortfolioTradingSimulation(assets, targetAllocations, 4.0)
		}).To(Panic()) //
	})
}

func Test_harmonicMean(t *testing.T) {

	t.Run("errors", func(t *testing.T) {
		g := NewGomegaWithT(t)

		// NaN
		g.Expect(math.IsNaN(harmonicMean(nil).Float())).To(BeTrue())
		g.Expect(math.IsNaN(harmonicMean([]GrowthMultiplier{}).Float())).To(BeTrue())

		// panic
		g.Expect(func() {
			harmonicMean([]GrowthMultiplier{0})
		}).To(Panic())
	})

	t.Run("success", func(t *testing.T) {
		g := NewGomegaWithT(t)
		// got these numbers using the HARMMEAN function in a Google Spreadsheets
		g.Expect(harmonicMean([]GrowthMultiplier{1})).To(Equal(Percent(1.0)))
		g.Expect(harmonicMean([]GrowthMultiplier{1, 2})).To(Equal(Percent(1.3333333333333333)))
		g.Expect(harmonicMean([]GrowthMultiplier{1, 2, 3, 4, 5})).To(Equal(Percent(2.18978102189781)))
	})
}

var (
	// tested in spreadsheet: https://docs.google.com/spreadsheets/d/14bxTQncj8BIUtQghpiQM0ZyeGsNUYQGFtNx2yOIkEEE/edit#gid=1013681666
	sampleReturns = readablePercents(-10.28, 0.90, 17.63, 17.93, -18.55, -28.42, 38.42, 26.54, -2.68, 9.23, 25.51, 33.62, -3.79, 18.66, 23.42, 3.01, 32.51, 16.05, 2.23, 17.89, 29.12, -6.22, 34.15, 8.92, 10.62, -0.17, 35.79, 20.96, 30.99, 23.26, 23.81, -10.57, -10.89, -20.95, 31.42, 12.61, 6.09, 15.63, 5.57, -36.99, 28.83, 17.26, 1.08, 16.38, 33.52, 12.56, 0.39, 12.66, 21.17, -5.17, 30.80)
)

func Test_swr(t *testing.T) {
	g := NewGomegaWithT(t)

	// swrProxy calls both swr and pwrAndSWR, making sure they both calculate an identical value.
	swrProxy := func(returns []Percent) Percent {
		t.Helper()
		a := swr(returns)
		_, b := pwrAndSWR(returns)
		g.Expect(a).To(Equal(b), "swr equal pwrAndSWR")
		return a
	}

	g.Expect(swrProxy([]Percent{})).To(Equal(Percent(1.0)))

	// briefly prove SWR for a 1 year 10% return
	{
		fixedWithdrawal := 0.5238095238095238
		g.Expect(swrProxy(readablePercents(10))).To(Equal(Percent(fixedWithdrawal)))
		initial := 1.0
		remaining := initial - fixedWithdrawal // withdraw first year's amount
		remaining *= 1.10                      // apply first year's growth
		remaining -= fixedWithdrawal           // withdraw second year's amount
		g.Expect(remaining).To(Equal(0.0))     // we exactly exhausted the account
	}

	g.Expect(swrProxy(sampleReturns)).To(Equal(Percent(0.06622907313022616)))
}

func Test_pwr(t *testing.T) {
	g := NewGomegaWithT(t)

	// pwrProxy calls both swr and pwrAndSWR, making sure they both calculate an identical value.
	pwrProxy := func(returns []Percent) Percent {
		t.Helper()
		a := pwr(returns)
		b, _ := pwrAndSWR(returns)
		g.Expect(a).To(Equal(b), "pwr equal pwrAndSWR")
		return a
	}

	g.Expect(pwrProxy([]Percent{})).To(Equal(Percent(0.0)))

	// briefly prove PWR for a 1 year 10% return
	{
		fixedWithdrawal := 0.04761904761904764
		g.Expect(pwrProxy(readablePercents(10))).To(Equal(Percent(fixedWithdrawal)))
		initial := 1.0
		remaining := initial - fixedWithdrawal // withdraw first year's amount
		remaining *= 1.10                      // apply first year's growth
		remaining -= fixedWithdrawal           // withdraw second year's amount
		g.Expect(remaining).To(Equal(initial)) // we exactly exhausted the account
	}

	g.Expect(pwrProxy(sampleReturns)).To(Equal(Percent(0.06574303881824274)))
}

func Test_minPWR(t *testing.T) {
	g := NewGomegaWithT(t)

	// minPWRProxy calls both minPWR and minPWRAndSWR, making sure they both calculate an identical value.
	minPWRProxy := func(returns []Percent, nYears int) (Percent, int) {
		t.Helper()
		a, n := minPWR(returns, nYears)
		b, _ := minPWRAndSWR(returns, nYears)
		g.Expect(a).To(Equal(b), "minPWR equal minPWRAndSWR")
		return a, n
	}

	verify := func(returns []Percent, nYears int, expectedPWR Percent, expectedIndex int) {
		t.Helper()
		rate, n := minPWRProxy(returns, nYears)
		g.Expect(rate).To(Equal(expectedPWR), "rate")
		g.Expect(n).To(Equal(expectedIndex), "index")
	}

	g.Expect(func() {
		minPWR(nil, 1)
	}).To(Panic())
	g.Expect(func() {
		minPWRAndSWR(nil, 1)
	}).To(Panic())

	verify(nil, 0, 0, 0)

	// length 1
	verify(readablePercents(10), 1, pwr(readablePercents(10)), 0)
	verify(readablePercents(20), 1, pwr(readablePercents(20)), 0)

	// length 2
	verify(readablePercents(10, 20), 1, pwr(readablePercents(10)), 0)
	verify(readablePercents(10, 20), 2, pwr(readablePercents(10, 20)), 0)

	// length 3
	verify(readablePercents(10, -5, 30), 1, pwr(readablePercents(-5)), 1)
	verify(readablePercents(10, -5, 30), 2, pwr(readablePercents(10, -5)), 0)
	verify(readablePercents(10, -5, 30), 3, pwr(readablePercents(10, -5, 30)), 0)

	// length 4
	verify(readablePercents(10, -5, 10, -20), 1, pwr(readablePercents(-20)), 3)
	verify(readablePercents(10, -5, 10, -20), 2, pwr(readablePercents(10, -20)), 2)
	verify(readablePercents(10, -5, 10, -20), 3, pwr(readablePercents(-5, 10, -20)), 1)
	verify(readablePercents(10, -5, 10, -20), 4, pwr(readablePercents(10, -5, 10, -20)), 0)

	verify(TSM, 30, 0.03237620200614041, 0)
	verify(SCV, 30, 0.038033553022899465, 0)
	verify(GLD, 30, -0.015629074083395443, 6)
	verify(LTT, 30, 0.02167080631193789, 0)
	verify(STT, 30, 0.017146827504420623, 21)
	verify(STB, 30, 0.022664798870331706, 0)

	verify(GoldenButterfly, 10, 0.01945631963862426, 0)
	verify(GoldenButterfly, 20, 0.038590835351436564, 0)
	verify(GoldenButterfly, 30, 0.04224334655073258, 0)
	verify(GoldenButterfly, 40, 0.042057016507784345, 0)
	verify(GoldenButterfly, 50, 0.04288283017428213, 0)
}

func Test_minSWR(t *testing.T) {
	g := NewGomegaWithT(t)

	// minSWRProxy calls both minSWR and minPWRAndSWR, making sure they both calculate an identical value.
	minSWRProxy := func(returns []Percent, nYears int) (Percent, int) {
		t.Helper()
		a, n := minSWR(returns, nYears)
		_, b := minPWRAndSWR(returns, nYears)
		g.Expect(a).To(Equal(b), "minSWR equal minPWRAndSWR")
		return a, n
	}

	verify := func(returns []Percent, nYears int, expectedSWR Percent, expectedIndex int) {
		t.Helper()
		rate, n := minSWRProxy(returns, nYears)
		g.Expect(rate).To(Equal(expectedSWR), "rate")
		g.Expect(n).To(Equal(expectedIndex), "index")
	}

	g.Expect(func() {
		minSWR(nil, 1)
	}).To(Panic())
	g.Expect(func() {
		minPWRAndSWR(nil, 1)
	}).To(Panic())

	verify(nil, 0, 0, 0)

	// length 1
	verify(readablePercents(10), 1, swr(readablePercents(10)), 0)
	verify(readablePercents(20), 1, swr(readablePercents(20)), 0)

	// length 2
	verify(readablePercents(10, 20), 1, swr(readablePercents(10)), 0)
	verify(readablePercents(10, 20), 2, swr(readablePercents(10, 20)), 0)

	// length 3
	verify(readablePercents(10, -5, 30), 1, swr(readablePercents(-5)), 1)
	verify(readablePercents(10, -5, 30), 2, swr(readablePercents(10, -5)), 0)
	verify(readablePercents(10, -5, 30), 3, swr(readablePercents(10, -5, 30)), 0)

	// length 4
	verify(readablePercents(10, -5, 10, -20), 1, swr(readablePercents(-20)), 3)
	verify(readablePercents(10, -5, 10, -20), 2, swr(readablePercents(10, -20)), 2)
	verify(readablePercents(10, -5, 10, -20), 3, swr(readablePercents(-5, 10, -20)), 1)
	verify(readablePercents(10, -5, 10, -20), 4, swr(readablePercents(10, -5, 10, -20)), 0)

	verify(TSM, 30, 0.037860676066939845, 0)
	verify(SCV, 30, 0.04290428968981266, 0)
	verify(GLD, 30, 0.012375454483082474, 11)
	verify(LTT, 30, 0.034180844746692515, 0)
	verify(STT, 30, 0.039744529728239206, 3)
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
	g.Expect(subSlices([]Percent{1}, 1)).To(Equal([][]Percent{{1}}))
	g.Expect(func() {
		subSlices([]Percent{1}, 2) // n is greater than length of the slice
	}).To(Panic())

	// length 2
	g.Expect(subSlices([]Percent{1, 2}, 1)).To(Equal([][]Percent{{1}, {2}}))
	g.Expect(subSlices([]Percent{1, 2}, 2)).To(Equal([][]Percent{{1, 2}}))

	// length 3
	g.Expect(subSlices([]Percent{1, 2, 3}, 1)).To(Equal([][]Percent{{1}, {2}, {3}}))
	g.Expect(subSlices([]Percent{1, 2, 3}, 2)).To(Equal([][]Percent{{1, 2}, {2, 3}}))
	g.Expect(subSlices([]Percent{1, 2, 3}, 3)).To(Equal([][]Percent{{1, 2, 3}}))
}

func Test_cagr(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(cagr(nil)).To(Equal(Percent(0.0)))
	g.Expect(cagr([]Percent{})).To(Equal(Percent(0.0)))
	g.Expect(cagr(readablePercents(1))).To(Equal(Percent(0.010000000000000009)))
	// prove it's correct
	{
		cagrValue := Percent(0.029805806936433976)
		g.Expect(cagr(readablePercents(1, 5))).To(Equal(cagrValue))
		cumulativeTwoYears := Percent(1.0605)
		// compound the CAGR value
		g.Expect((1 + cagrValue) * (1 + cagrValue)).To(Equal(cumulativeTwoYears))
		// compound the original returns, arrive at the same cumulative value
		g.Expect(1.01 * 1.05).To(Equal(cumulativeTwoYears.Float()))
	}
	g.Expect(cagr(readablePercents(5, 5, 5, 5, 5))).To(Equal(Percent(0.050000000000000044)))

	g.Expect(cagr(TSM)).To(Equal(Percent(0.059240605917942224)))
	g.Expect(cagr(SCV)).To(Equal(Percent(0.07363836101130472)))
	g.Expect(cagr(LTT)).To(Equal(Percent(0.036986778393646835)))
	g.Expect(cagr(STT)).To(Equal(Percent(0.018215249078317397)))
	g.Expect(cagr(STB)).To(Equal(Percent(0.02224127904840234)))
	g.Expect(cagr(GLD)).To(Equal(Percent(0.029259375673007515)))
	g.Expect(cagr(GoldenButterfly)).To(Equal(Percent(0.05352050963712207)))
}

func Test_averageReturn(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(math.IsNaN(averageReturn(nil).Float())).To(BeTrue())
	g.Expect(math.IsNaN(averageReturn([]Percent{}).Float())).To(BeTrue())
	g.Expect(averageReturn(readablePercents(1))).To(Equal(Percent(0.01)))
	g.Expect(averageReturn(readablePercents(1, 2))).To(Equal(Percent(0.015)))
	g.Expect(averageReturn(readablePercents(1, 2, -3))).To(BeNumerically("~", 0))
	g.Expect(averageReturn(readablePercents(series(0, 100, 1)...))).To(Equal(Percent(0.50)))
}

func Test_startDateSensitivity(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("one 20-year segment", func(t *testing.T) {
		g := NewGomegaWithT(t)

		g.Expect(startDateSensitivity(repeat(0, 20))).To(BeZero())

		// 5% improvment
		returns := append(
			repeat(readablePercent(5), 10),
			repeat(readablePercent(10), 10)...)
		g.Expect(startDateSensitivity(returns)).To(Equal(Percent(0.050000000000000044)))

		// 5% shortfall
		returns = append(
			repeat(readablePercent(10), 10),
			repeat(readablePercent(5), 10)...)
		g.Expect(startDateSensitivity(returns)).To(Equal(Percent(0.050000000000000044)))
	})

	// long steadily increasing set of returns
	returns := readablePercents(series(0, 100, 1)...)
	g.Expect(startDateSensitivity(returns)).To(Equal(Percent(0.10003452051664796)))
}

func repeat(x Percent, count int) []Percent {
	res := make([]Percent, count)
	for i := range res {
		res[i] = x
	}
	return res
}

func Test_baselineReturn(t *testing.T) {

	t.Run("percentile out of range", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(func() {
			baselineReturn(readablePercents(10), 1, -0.0001)
		}).To(Panic())
		g.Expect(func() {
			baselineReturn(readablePercents(10), 1, 100)
		}).To(Panic())
	})

	t.Run("empty returns ok", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(baselineReturn(nil, 0, readablePercent(0))).To(Equal(Percent(0.0)))
		g.Expect(baselineReturn(nil, 1, readablePercent(50))).To(Equal(Percent(0.0)))
	})

	t.Run("one return", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(baselineReturn(readablePercents(10), 1, readablePercent(0))).To(Equal(Percent(0.10000000000000009)))
		g.Expect(baselineReturn(readablePercents(10), 1, readablePercent(50))).To(Equal(Percent(0.10000000000000009)))
		g.Expect(baselineReturn(readablePercents(10), 1, readablePercent(99.999))).To(Equal(Percent(0.10000000000000009)))
	})

	t.Run("two returns", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(baselineReturn(readablePercents(10, 20), 1, readablePercent(0))).To(Equal(Percent(0.10000000000000009)))
		g.Expect(baselineReturn(readablePercents(10, 20), 1, readablePercent(49))).To(Equal(Percent(0.10000000000000009)))
		g.Expect(baselineReturn(readablePercents(10, 20), 1, readablePercent(50))).To(Equal(Percent(0.19999999999999996)))
		g.Expect(baselineReturn(readablePercents(10, 20), 1, readablePercent(99.999))).To(Equal(Percent(0.19999999999999996)))
	})

	t.Run("three returns", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(baselineReturn(readablePercents(10, 20, 30), 2, readablePercent(0))).To(Equal(Percent(0.14891252930760568)))
		g.Expect(baselineReturn(readablePercents(10, 20, 30), 2, readablePercent(49))).To(Equal(Percent(0.14891252930760568)))
		g.Expect(baselineReturn(readablePercents(10, 20, 30), 2, readablePercent(50))).To(Equal(Percent(0.24899959967967966)))
		g.Expect(baselineReturn(readablePercents(10, 20, 30), 2, readablePercent(99.999))).To(Equal(Percent(0.24899959967967966)))

		t.Run("CAGRs are sorted", func(t *testing.T) {
			g := NewGomegaWithT(t)
			g.Expect(baselineReturn(readablePercents(-10, 20, -30), 1, readablePercent(0))).To(Equal(Percent(-0.30000000000000004)))
			g.Expect(baselineReturn(readablePercents(-10, 20, -30), 1, readablePercent(33.3))).To(Equal(Percent(-0.30000000000000004)))
			g.Expect(baselineReturn(readablePercents(-10, 20, -30), 1, readablePercent(33.4))).To(Equal(Percent(-0.09999999999999998)))
			g.Expect(baselineReturn(readablePercents(-10, 20, -30), 1, readablePercent(66.6))).To(Equal(Percent(-0.09999999999999998)))
			g.Expect(baselineReturn(readablePercents(-10, 20, -30), 1, readablePercent(66.7))).To(Equal(Percent(0.19999999999999996)))
			g.Expect(baselineReturn(readablePercents(-10, 20, -30), 1, readablePercent(99.999))).To(Equal(Percent(0.19999999999999996)))
		})
	})
}

func Test_baselineLongTermReturn(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(baselineLongTermReturn(TSM)).To(Equal(Percent(0.030599414622012988)))
	g.Expect(baselineLongTermReturn(SCV)).To(Equal(Percent(0.05925630873497112)))
	g.Expect(baselineLongTermReturn(LTT)).To(Equal(Percent(0.022152065956750455)))
	g.Expect(baselineLongTermReturn(STT)).To(Equal(Percent(0.007233206367346812)))
	g.Expect(baselineLongTermReturn(STB)).To(Equal(Percent(0.013932133854550166)))
	g.Expect(baselineLongTermReturn(GLD)).To(Equal(Percent(-0.052036901573972894)))
	g.Expect(baselineLongTermReturn(GoldenButterfly)).To(Equal(Percent(0.05240715018337849)))
}

func Test_baselineShortTermReturn(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(baselineShortTermReturn(TSM)).To(Equal(Percent(-0.02905140217935165)))
	g.Expect(baselineShortTermReturn(SCV)).To(Equal(Percent(0.019349617074645886)))
	g.Expect(baselineShortTermReturn(LTT)).To(Equal(Percent(0.006370681347895868)))
	g.Expect(baselineShortTermReturn(STT)).To(Equal(Percent(-0.013333955977865353)))
	g.Expect(baselineShortTermReturn(STB)).To(Equal(Percent(-0.006417877051557608)))
	g.Expect(baselineShortTermReturn(GLD)).To(Equal(Percent(-0.11357718226127445)))
	g.Expect(baselineShortTermReturn(GoldenButterfly)).To(Equal(Percent(0.028483535122723058)))
}

func Test_standardDeviation(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(func() {
		standardDeviation(nil)
	}).To(Panic())

	t.Run("matches STDEVP function in google spreadsheet", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(standardDeviation([]Percent{1})).To(Equal(readablePercent(0.0)))
		g.Expect(standardDeviation([]Percent{1, 2})).To(Equal(Percent(0.5)))
		g.Expect(standardDeviation([]Percent{1, 2, 3, 4})).To(Equal(Percent(1.118033988749895)))
		g.Expect(standardDeviation([]Percent{1, -2, 3, -4})).To(Equal(Percent(2.692582403567252)))
	})

	g.Expect(standardDeviation(TSM)).To(Equal(readablePercent(17.165685399889558)))
	g.Expect(standardDeviation(SCV)).To(Equal(readablePercent(19.394760597619787)))
	g.Expect(standardDeviation(LTT)).To(Equal(Percent(0.12327064319013904)))
	g.Expect(standardDeviation(STT)).To(Equal(Percent(0.04387180601846474)))
	g.Expect(standardDeviation(STB)).To(Equal(Percent(0.0482507718671391)))
	g.Expect(standardDeviation(GLD)).To(Equal(readablePercent(23.857500543433524)))
	g.Expect(standardDeviation(GoldenButterfly)).To(Equal(Percent(0.08103170495645956)))
}

func Test_readablePercent(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(float64(readablePercent(33))).To(Equal(0.33))
}

func Test_cumulative_and_cumulativeList(t *testing.T) {
	g := NewGomegaWithT(t)

	// The last item returned by cumulativeList should always equal what cumulative returns for the same input.
	verify := func(returns []Percent, expectedList []GrowthMultiplier) {
		t.Helper()
		g.Expect(cumulative(returns)).To(Equal(expectedList[len(expectedList)-1]), "cumulative")
		g.Expect(cumulativeList(returns)).To(Equal(expectedList), "cumulativeList")
	}

	verify([]Percent{}, []GrowthMultiplier{1})
	verify([]Percent{0.2}, []GrowthMultiplier{1, 1.2})
	verify([]Percent{0.2, -0.2}, []GrowthMultiplier{1, 1.2, 0.96})
}

func Test_rebalanceFactor_effect(t *testing.T) {
	t.Skip("Can run this test manually, when desired.")

	g := NewGomegaWithT(t)

	var (
		gbAssets      = [][]Percent{TSM, SCV, LTT, STT, GLD}
		gbPermutation = Permutation{
			Assets:      []string{"TSM", "SCV", "LTT", "STT", "GLD"},
			Percentages: readablePercents(20, 20, 20, 20, 20),
		}
	)

	// see how the rebalanceFactor can affect the GoldenButterfly portfolio results.
	var results []*PortfolioStat
	for rebalanceFactor := 0.0; rebalanceFactor <= 1.975; rebalanceFactor += 0.001 {
		returns, err := PortfolioTradingSimulation(gbAssets, gbPermutation.Percentages, rebalanceFactor)
		g.Expect(err).To(Succeed())

		stat := evaluatePortfolio(returns, gbPermutation)
		stat.RebalanceFactor = rebalanceFactor
		results = append(results, stat)
	}

	RankPortfoliosInPlace(results)

	gbStat := FindOne(results, func(p *PortfolioStat) bool { return math.Abs(p.RebalanceFactor-1.0) < 0.00001 })
	fmt.Println("RebalanceFactor=1:", gbStat)

	fmt.Println("Best combined overall ranks:")
	fmt.Println("#1:", results[0])
	fmt.Println("#2:", results[1])
	fmt.Println("#3:", results[2])

	fmt.Println("\nBest by each ranking:")
	PrintBestByEachRanking(results)

	// find as good or better than GoldenButterfly
	betterThanGB := CopyAll(FindMany(results, AsGoodOrBetterThan(gbStat)))
	RankPortfoliosInPlace(betterThanGB)
	fmt.Println("As good or better than GoldenButterfly:", len(betterThanGB))
	PrintBestByEachRanking(betterThanGB)
}
