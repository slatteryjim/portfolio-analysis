package portfolio_analysis

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/guptarohit/asciigraph"
	. "github.com/onsi/gomega"

	"github.com/slatteryjim/portfolio-analysis/data"
	. "github.com/slatteryjim/portfolio-analysis/types"
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

		_, err = portfolioReturnsProxy(nil, ReadablePercents(100))
		g.Expect(err).To(MatchError("lists must have the same length: targetAllocations (1), returnsList (0)"))
	})

	t.Run("success", func(t *testing.T) {
		g := NewGomegaWithT(t)

		// simply one asset, one year
		g.Expect(portfolioReturnsProxy(
			[][]Percent{
				ReadablePercents(1),
			},
			ReadablePercents(100)),
		).To(Equal(
			ReadablePercents(1),
		))

		// two assets, two years, 50%/50%
		g.Expect(portfolioReturnsProxy(
			[][]Percent{
				ReadablePercents(10, 20),
				ReadablePercents(5, 10),
			},
			ReadablePercents(50, 50)),
		).To(Equal(
			[]Percent{0.07500000000000001, 0.15000000000000002},
		))

		// TSM asset, 100%, yields itself
		g.Expect(portfolioReturnsProxy(
			[][]Percent{
				TSM,
			},
			ReadablePercents(100)),
		).To(Equal(
			TSM,
		))

		// GoldenButterfly
		g.Expect(portfolioReturnsProxy([][]Percent{TSM, SCV, LTT, STT, GLD}, ReadablePercents(20, 20, 20, 20, 20))).To(Equal(
			[]Percent{-0.1533383268282135, 0.017318791285211466, 0.10068425974797353, 0.11372919971053391, -0.0172148539804938, -0.06461024302567049, 0.09360907230753582, 0.139978985084822, 0.009677225905924784, 0.03404942084475439, 0.2372421712101816, 0.025928538460314007, -0.09709304924704815, 0.1973312916341973, 0.07495518258141035, -0.010392136396391902, 0.18336903125620316, 0.14078544233466855, 0.00018309646857563727, 0.04669750085940259, 0.08673884623758171, -0.08572038845018258, 0.1563762834314272, 0.061295238653148135, 0.10983322532725735, -0.0465004614702163, 0.178993634975939, 0.052560744788289336, 0.11162614426521761, 0.05817517787259961, 0.016929554860230407, 0.03104069332156112, 0.017056422822255918, -0.0017220660217116823, 0.16666679718569166, 0.06405102737142347, 0.03846823294327486, 0.09777651111177979, 0.04997737064863614, -0.06695505567337855, 0.11050169878101092, 0.14604978374972077, 0.04114653873382587, 0.07603962935631807, 0.04372680283039169, 0.08816734725542127, -0.04048477144240979, 0.07444466516990281, 0.08438538586983815, -0.05602189725865632, 0.1534744248725398},
		))
	})
}

func TestPortfolioTradingSimulation(t *testing.T) {
	var (
		assets = [][]Percent{
			ReadablePercents(20, 20, 20), // first asset greatly outperforms second
			ReadablePercents(0, 0, 0),
		}
		targetAllocations = ReadablePercents(50, 50)

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
	sampleReturns = ReadablePercents(-10.28, 0.90, 17.63, 17.93, -18.55, -28.42, 38.42, 26.54, -2.68, 9.23, 25.51, 33.62, -3.79, 18.66, 23.42, 3.01, 32.51, 16.05, 2.23, 17.89, 29.12, -6.22, 34.15, 8.92, 10.62, -0.17, 35.79, 20.96, 30.99, 23.26, 23.81, -10.57, -10.89, -20.95, 31.42, 12.61, 6.09, 15.63, 5.57, -36.99, 28.83, 17.26, 1.08, 16.38, 33.52, 12.56, 0.39, 12.66, 21.17, -5.17, 30.80)
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
		g.Expect(swrProxy(ReadablePercents(10))).To(Equal(Percent(fixedWithdrawal)))
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
		g.Expect(pwrProxy(ReadablePercents(10))).To(Equal(Percent(fixedWithdrawal)))
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
	verify(ReadablePercents(10), 1, pwr(ReadablePercents(10)), 0)
	verify(ReadablePercents(20), 1, pwr(ReadablePercents(20)), 0)

	// length 2
	verify(ReadablePercents(10, 20), 1, pwr(ReadablePercents(10)), 0)
	verify(ReadablePercents(10, 20), 2, pwr(ReadablePercents(10, 20)), 0)

	// length 3
	verify(ReadablePercents(10, -5, 30), 1, pwr(ReadablePercents(-5)), 1)
	verify(ReadablePercents(10, -5, 30), 2, pwr(ReadablePercents(10, -5)), 0)
	verify(ReadablePercents(10, -5, 30), 3, pwr(ReadablePercents(10, -5, 30)), 0)

	// length 4
	verify(ReadablePercents(10, -5, 10, -20), 1, pwr(ReadablePercents(-20)), 3)
	verify(ReadablePercents(10, -5, 10, -20), 2, pwr(ReadablePercents(10, -20)), 2)
	verify(ReadablePercents(10, -5, 10, -20), 3, pwr(ReadablePercents(-5, 10, -20)), 1)
	verify(ReadablePercents(10, -5, 10, -20), 4, pwr(ReadablePercents(10, -5, 10, -20)), 0)

	verify(TSM, 30, 0.03237787319823412, 0)
	verify(SCV, 30, 0.03803405103934034, 0)
	verify(GLD, 30, -0.015630815039841064, 6)
	verify(LTT, 30, 0.021678521798131487, 0)
	verify(STT, 30, 0.017145731503677816, 21)
	verify(STB, 30, 0.022672387011138363, 0)

	verify(GoldenButterfly, 10, 0.019455306405833442, 0)
	verify(GoldenButterfly, 20, 0.038591150235617475, 0)
	verify(GoldenButterfly, 30, 0.04224373554338784, 0)
	verify(GoldenButterfly, 40, 0.04205689393031536, 0)
	verify(GoldenButterfly, 50, 0.042883248739577114, 0)
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
	verify(ReadablePercents(10), 1, swr(ReadablePercents(10)), 0)
	verify(ReadablePercents(20), 1, swr(ReadablePercents(20)), 0)

	// length 2
	verify(ReadablePercents(10, 20), 1, swr(ReadablePercents(10)), 0)
	verify(ReadablePercents(10, 20), 2, swr(ReadablePercents(10, 20)), 0)

	// length 3
	verify(ReadablePercents(10, -5, 30), 1, swr(ReadablePercents(-5)), 1)
	verify(ReadablePercents(10, -5, 30), 2, swr(ReadablePercents(10, -5)), 0)
	verify(ReadablePercents(10, -5, 30), 3, swr(ReadablePercents(10, -5, 30)), 0)

	// length 4
	verify(ReadablePercents(10, -5, 10, -20), 1, swr(ReadablePercents(-20)), 3)
	verify(ReadablePercents(10, -5, 10, -20), 2, swr(ReadablePercents(10, -20)), 2)
	verify(ReadablePercents(10, -5, 10, -20), 3, swr(ReadablePercents(-5, 10, -20)), 1)
	verify(ReadablePercents(10, -5, 10, -20), 4, swr(ReadablePercents(10, -5, 10, -20)), 0)

	verify(TSM, 30, 0.03786147243197951, 0)
	verify(SCV, 30, 0.0429043601737704, 0)
	verify(GLD, 30, 0.012374613192608942, 11)
	verify(LTT, 30, 0.03418591730792702, 0)
	verify(STT, 30, 0.039739924674330344, 3)
	verify(STB, 30, 0.039649617433661334, 0)

	verify(GoldenButterfly, 10, 0.09331551382163732, 0)
	verify(GoldenButterfly, 20, 0.0626136479380333, 0)
	verify(GoldenButterfly, 30, 0.053048942980622904, 0)
	verify(GoldenButterfly, 40, 0.048790172577370526, 0)
	verify(GoldenButterfly, 50, 0.04665049136149236, 0)
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
	g.Expect(cagr(ReadablePercents(1))).To(Equal(Percent(0.010000000000000009)))
	// prove it's correct
	{
		cagrValue := Percent(0.029805806936433976)
		g.Expect(cagr(ReadablePercents(1, 5))).To(Equal(cagrValue))
		cumulativeTwoYears := Percent(1.0605)
		// compound the CAGR value
		g.Expect((1 + cagrValue) * (1 + cagrValue)).To(Equal(cumulativeTwoYears))
		// compound the original returns, arrive at the same cumulative value
		g.Expect(1.01 * 1.05).To(Equal(cumulativeTwoYears.Float()))
	}
	g.Expect(cagr(ReadablePercents(5, 5, 5, 5, 5))).To(Equal(Percent(0.050000000000000044)))

	g.Expect(cagr(TSM)).To(Equal(Percent(0.05924856139463475)))
	g.Expect(cagr(SCV)).To(Equal(Percent(0.07363869666341749)))
	g.Expect(cagr(LTT)).To(Equal(Percent(0.036995981175477866)))
	g.Expect(cagr(STT)).To(Equal(Percent(0.01821005315305091)))
	g.Expect(cagr(STB)).To(Equal(Percent(0.02224589590692516)))
	g.Expect(cagr(GLD)).To(Equal(Percent(0.029261555900620406)))
	g.Expect(cagr(GoldenButterfly)).To(Equal(Percent(0.053522786534198286)))
}

func Test_average(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(math.IsNaN(average(nil).Float())).To(BeTrue())
	g.Expect(math.IsNaN(average([]Percent{}).Float())).To(BeTrue())
	g.Expect(average(ReadablePercents(1))).To(Equal(Percent(0.01)))
	g.Expect(average(ReadablePercents(1, 2))).To(Equal(Percent(0.015)))
	g.Expect(average(ReadablePercents(1, 2, -3))).To(BeNumerically("~", 0))
	g.Expect(average(ReadablePercents(series(0, 100, 1)...))).To(Equal(Percent(0.50)))
}

func Test_startDateSensitivity(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("one 20-year segment", func(t *testing.T) {
		g := NewGomegaWithT(t)

		g.Expect(startDateSensitivity(repeat(0, 20))).To(BeZero())

		// 5% improvment
		returns := append(
			repeat(ReadablePercent(5), 10),
			repeat(ReadablePercent(10), 10)...)
		g.Expect(startDateSensitivity(returns)).To(Equal(Percent(0.050000000000000044)))

		// 5% shortfall
		returns = append(
			repeat(ReadablePercent(10), 10),
			repeat(ReadablePercent(5), 10)...)
		g.Expect(startDateSensitivity(returns)).To(Equal(Percent(0.050000000000000044)))
	})

	// long steadily increasing set of returns
	returns := ReadablePercents(series(0, 100, 1)...)
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
			baselineReturn(ReadablePercents(10), 1, -0.0001)
		}).To(Panic())
		g.Expect(func() {
			baselineReturn(ReadablePercents(10), 1, 100)
		}).To(Panic())
	})

	t.Run("empty returns ok", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(baselineReturn(nil, 0, ReadablePercent(0))).To(Equal(Percent(0.0)))
		g.Expect(baselineReturn(nil, 1, ReadablePercent(50))).To(Equal(Percent(0.0)))
	})

	t.Run("one return", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(baselineReturn(ReadablePercents(10), 1, ReadablePercent(0))).To(Equal(Percent(0.10000000000000009)))
		g.Expect(baselineReturn(ReadablePercents(10), 1, ReadablePercent(50))).To(Equal(Percent(0.10000000000000009)))
		g.Expect(baselineReturn(ReadablePercents(10), 1, ReadablePercent(99.999))).To(Equal(Percent(0.10000000000000009)))
	})

	t.Run("two returns", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(baselineReturn(ReadablePercents(10, 20), 1, ReadablePercent(0))).To(Equal(Percent(0.10000000000000009)))
		g.Expect(baselineReturn(ReadablePercents(10, 20), 1, ReadablePercent(49))).To(Equal(Percent(0.10000000000000009)))
		g.Expect(baselineReturn(ReadablePercents(10, 20), 1, ReadablePercent(50))).To(Equal(Percent(0.19999999999999996)))
		g.Expect(baselineReturn(ReadablePercents(10, 20), 1, ReadablePercent(99.999))).To(Equal(Percent(0.19999999999999996)))
	})

	t.Run("three returns", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(baselineReturn(ReadablePercents(10, 20, 30), 2, ReadablePercent(0))).To(Equal(Percent(0.14891252930760568)))
		g.Expect(baselineReturn(ReadablePercents(10, 20, 30), 2, ReadablePercent(49))).To(Equal(Percent(0.14891252930760568)))
		g.Expect(baselineReturn(ReadablePercents(10, 20, 30), 2, ReadablePercent(50))).To(Equal(Percent(0.24899959967967966)))
		g.Expect(baselineReturn(ReadablePercents(10, 20, 30), 2, ReadablePercent(99.999))).To(Equal(Percent(0.24899959967967966)))

		t.Run("CAGRs are sorted", func(t *testing.T) {
			g := NewGomegaWithT(t)
			g.Expect(baselineReturn(ReadablePercents(-10, 20, -30), 1, ReadablePercent(0))).To(Equal(Percent(-0.30000000000000004)))
			g.Expect(baselineReturn(ReadablePercents(-10, 20, -30), 1, ReadablePercent(33.3))).To(Equal(Percent(-0.30000000000000004)))
			g.Expect(baselineReturn(ReadablePercents(-10, 20, -30), 1, ReadablePercent(33.4))).To(Equal(Percent(-0.09999999999999998)))
			g.Expect(baselineReturn(ReadablePercents(-10, 20, -30), 1, ReadablePercent(66.6))).To(Equal(Percent(-0.09999999999999998)))
			g.Expect(baselineReturn(ReadablePercents(-10, 20, -30), 1, ReadablePercent(66.7))).To(Equal(Percent(0.19999999999999996)))
			g.Expect(baselineReturn(ReadablePercents(-10, 20, -30), 1, ReadablePercent(99.999))).To(Equal(Percent(0.19999999999999996)))
		})
	})
}

func Test_baselineLongTermReturn(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(baselineLongTermReturn(TSM)).To(Equal(Percent(0.0306081363792714)))
	g.Expect(baselineLongTermReturn(SCV)).To(Equal(Percent(0.05925736306162355)))
	g.Expect(baselineLongTermReturn(LTT)).To(Equal(Percent(0.022165919650537713)))
	g.Expect(baselineLongTermReturn(STT)).To(Equal(Percent(0.007220749317469188)))
	g.Expect(baselineLongTermReturn(STB)).To(Equal(Percent(0.013944874877032776)))
	g.Expect(baselineLongTermReturn(GLD)).To(Equal(Percent(-0.0520425731161559)))
	g.Expect(baselineLongTermReturn(GoldenButterfly)).To(Equal(Percent(0.05240937633854492)))
}

func Test_baselineShortTermReturn(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(baselineShortTermReturn(TSM)).To(Equal(Percent(-0.02907904796851324)))
	g.Expect(baselineShortTermReturn(SCV)).To(Equal(Percent(0.01936917994777443)))
	g.Expect(baselineShortTermReturn(LTT)).To(Equal(Percent(0.00639165311066181)))
	g.Expect(baselineShortTermReturn(STT)).To(Equal(Percent(-0.013316758228289038)))
	g.Expect(baselineShortTermReturn(STB)).To(Equal(Percent(-0.006415578705477043)))
	g.Expect(baselineShortTermReturn(GLD)).To(Equal(Percent(-0.1135918966323598)))
	g.Expect(baselineShortTermReturn(GoldenButterfly)).To(Equal(Percent(0.028491192128850873)))
}

func Test_standardDeviation(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(func() {
		standardDeviation(nil)
	}).To(Panic())

	t.Run("matches STDEVP function in google spreadsheet", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(standardDeviation([]Percent{1})).To(Equal(ReadablePercent(0.0)))
		g.Expect(standardDeviation([]Percent{1, 2})).To(Equal(Percent(0.5)))
		g.Expect(standardDeviation([]Percent{1, 2, 3, 4})).To(Equal(Percent(1.118033988749895)))
		g.Expect(standardDeviation([]Percent{1, -2, 3, -4})).To(Equal(Percent(2.692582403567252)))
	})

	g.Expect(standardDeviation(TSM)).To(Equal(ReadablePercent(17.165466213991304)))
	g.Expect(standardDeviation(SCV)).To(Equal(ReadablePercent(19.3942212424017)))
	g.Expect(standardDeviation(LTT)).To(Equal(Percent(0.12326313856389255)))
	g.Expect(standardDeviation(STT)).To(Equal(Percent(0.043862677714401194)))
	g.Expect(standardDeviation(STB)).To(Equal(Percent(0.04824512942467834)))
	g.Expect(standardDeviation(GLD)).To(Equal(ReadablePercent(23.857120994357875)))
	g.Expect(standardDeviation(GoldenButterfly)).To(Equal(Percent(0.08102969356581732)))
}

func Test_readablePercent(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(float64(ReadablePercent(33))).To(Equal(0.33))
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
		gbCombination = Combination{
			Assets:      []string{"TSM", "SCV", "LTT", "STT", "GLD"},
			Percentages: ReadablePercents(20, 20, 20, 20, 20),
		}
	)

	// see how the rebalanceFactor can affect the GoldenButterfly portfolio results.
	var results []*PortfolioStat
	for rebalanceFactor := 0.0; rebalanceFactor <= 1.975; rebalanceFactor += 0.001 {
		returns, err := PortfolioTradingSimulation(gbAssets, gbCombination.Percentages, rebalanceFactor)
		g.Expect(err).To(Succeed())

		stat := evaluatePortfolio(returns, gbCombination)
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

func TestTSMPerformance(t *testing.T) {
	tsmCombination := Combination{Assets: []string{"TSM"}, Percentages: ReadablePercents(100)}

	// 1969 start date, using new TSV data source
	stat := evaluatePortfolio(TSM, tsmCombination)
	fmt.Println(stat)

	// 1871 start date
	stat = evaluatePortfolio(data.MustFind("TSM").AnnualReturns, tsmCombination)
	fmt.Println(stat)

	// Output:
	// [TSM] [100%] (0) RF:0.00 AvgReturn:7.453%(0) BLT:3.061%(0) BST:-2.908%(0) PWR:3.238%(0) SWR:3.786%(0) StdDev:17.165%(0) Ulcer:27.0(0) DeepestDrawdown:-52.25%(0) LongestDrawdown:13(0), StartDateSensitivity:31.65%(0)
	// [TSM] [100%] (0) RF:0.00 AvgReturn:8.481%(0) BLT:2.363%(0) BST:-1.339%(0) PWR:2.715%(0) SWR:3.578%(0) StdDev:18.115%(0) Ulcer:27.0(0) DeepestDrawdown:-57.56%(0) LongestDrawdown:13(0), StartDateSensitivity:36.48%(0)
}

func Test_allPWRs(t *testing.T) {

	t.Run("GoldenButterfly", func(t *testing.T) {
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
		g := NewGomegaWithT(t)

		allReturns, err := portfolioReturns(
			data.PortfolioReturnsList(ParseAssets(`|ST Invest. Grade|Int'l Small|T-Bill|Wellesley|TIPS|REIT|LT STRIPS|Wellington|`)...),
			equalWeightAllocations(8))
		g.Expect(err).To(Succeed())
		ExpectPlot(t, allPWRs(allReturns, 10), `
 0.080 ┼╮                         
 0.069 ┤╰╮   ╭╮  ╭╮               
 0.058 ┤ │╭─╮│╰─╮│╰─╮   ╭─╮       
 0.047 ┤ ╰╯ ╰╯  ╰╯  ╰╮╭─╯ ╰──╮╭── 
 0.035 ┤             ╰╯      ╰╯
`)
		ExpectPlot(t, allPWRs(allReturns, 20), `
 0.079 ┼╮               
 0.067 ┤╰╮              
 0.056 ┤ ╰──╮╭────╮   ╭ 
 0.045 ┤    ╰╯    ╰───╯
`)
		ExpectPlot(t, allPWRs(allReturns, 30), `
 0.074 ┼╮     
 0.062 ┤╰╮╭╮  
 0.050 ┤ ╰╯╰─
`)
	})
}

func ExpectPlot(t testing.TB, data []Percent, expectedResult string) {
	t.Helper()
	plot := " " + strings.TrimSpace(asciigraph.Plot(Floats(data...)))
	expectedResult = " " + strings.TrimSpace(expectedResult)

	if plot != expectedResult {
		fmt.Println("Got:")
		fmt.Println(plot)
		fmt.Println("Expected:")
		fmt.Println(expectedResult)
		t.Fatal("Plot didn't match")
	}
}
