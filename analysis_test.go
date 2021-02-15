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

		t.Run("GoldenButterfly", func(t *testing.T) {
			g := NewGomegaWithT(t)
			returns, err := portfolioReturnsProxy([][]Percent{TSM, SCV, LTT, STT, GLD}, ReadablePercents(20, 20, 20, 20, 20))
			g.Expect(err).To(Succeed())
			ExpectMatchesGoldenFile(t, formatPercents(returns))
		})
	})
}

// formatPercents returns the percents, one per line, right-justified.
func formatPercents(returns []Percent) string {
	var sb strings.Builder
	sb.WriteString("Percents:\n")
	for _, r := range returns {
		sb.WriteString(fmt.Sprintf("  %18v\n", r.String()))
	}
	return sb.String()
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

	t.Run("example series", func(t *testing.T) {
		var sb strings.Builder
		reportLine := func(name string, returns []Percent, nYears int) string {
			minPWR, index := minPWRProxy(returns, nYears)
			return fmt.Sprintf("%s: %2d years minPWR: %16v at index %2d\n", name, nYears, minPWR, index)
		}
		sb.WriteString(reportLine("TSM", TSM, 30))
		sb.WriteString(reportLine("SCV", SCV, 30))
		sb.WriteString(reportLine("GLD", GLD, 30))
		sb.WriteString(reportLine("LTT", LTT, 30))
		sb.WriteString(reportLine("STT", STT, 30))
		sb.WriteString(reportLine("STB", STB, 30))
		sb.WriteString("\n")
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly, 10))
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly, 20))
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly, 30))
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly, 40))
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly, 50))

		ExpectMatchesGoldenFile(t, sb.String())
	})
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

	t.Run("example series", func(t *testing.T) {
		var sb strings.Builder
		reportLine := func(name string, returns []Percent, nYears int) string {
			gotMinSWR, index := minSWRProxy(returns, nYears)
			return fmt.Sprintf("%s: %2d years minSWR: %16v at index %2d\n", name, nYears, gotMinSWR, index)
		}
		sb.WriteString(reportLine("TSM", TSM, 30))
		sb.WriteString(reportLine("SCV", SCV, 30))
		sb.WriteString(reportLine("GLD", GLD, 30))
		sb.WriteString(reportLine("LTT", LTT, 30))
		sb.WriteString(reportLine("STT", STT, 30))
		sb.WriteString(reportLine("STB", STB, 30))
		sb.WriteString("\n")
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly, 10))
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly, 20))
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly, 30))
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly, 40))
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly, 50))

		ExpectMatchesGoldenFile(t, sb.String())
	})
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

	t.Run("example series", func(t *testing.T) {
		reportLine := func(name string, returns []Percent) string {
			return fmt.Sprintf("%s CAGR: %14v\n", name, cagr(returns))
		}
		var sb strings.Builder
		sb.WriteString(reportLine("TSM", TSM))
		sb.WriteString(reportLine("SCV", SCV))
		sb.WriteString(reportLine("LTT", LTT))
		sb.WriteString(reportLine("STT", STT))
		sb.WriteString(reportLine("STB", STB))
		sb.WriteString(reportLine("GLD", GLD))
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly))
		ExpectMatchesGoldenFile(t, sb.String())
	})
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
	t.Run("example data", func(t *testing.T) {
		var sb strings.Builder
		reportLine := func(name string, returns []Percent) string {
			ltReturn := baselineLongTermReturn(returns)
			return fmt.Sprintf("%s: baselineLongTermReturn: %16v\n", name, ltReturn)
		}
		sb.WriteString(reportLine("TSM", TSM))
		sb.WriteString(reportLine("SCV", SCV))
		sb.WriteString(reportLine("GLD", GLD))
		sb.WriteString(reportLine("LTT", LTT))
		sb.WriteString(reportLine("STT", STT))
		sb.WriteString(reportLine("STB", STB))
		sb.WriteString("\n")
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly))

		ExpectMatchesGoldenFile(t, sb.String())
	})
}

func Test_baselineShortTermReturn(t *testing.T) {
	t.Run("example data", func(t *testing.T) {
		var sb strings.Builder
		reportLine := func(name string, returns []Percent) string {
			stReturn := baselineShortTermReturn(returns)
			return fmt.Sprintf("%s: baselineShortTermReturn: %16v\n", name, stReturn)
		}
		sb.WriteString(reportLine("TSM", TSM))
		sb.WriteString(reportLine("SCV", SCV))
		sb.WriteString(reportLine("GLD", GLD))
		sb.WriteString(reportLine("LTT", LTT))
		sb.WriteString(reportLine("STT", STT))
		sb.WriteString(reportLine("STB", STB))
		sb.WriteString("\n")
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly))

		ExpectMatchesGoldenFile(t, sb.String())
	})
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

	t.Run("example data", func(t *testing.T) {
		var sb strings.Builder
		reportLine := func(name string, returns []Percent) string {
			return fmt.Sprintf("%s standardDeviation: %16v\n", name, standardDeviation(returns))
		}
		sb.WriteString(reportLine("TSM", TSM))
		sb.WriteString(reportLine("SCV", SCV))
		sb.WriteString(reportLine("GLD", GLD))
		sb.WriteString(reportLine("LTT", LTT))
		sb.WriteString(reportLine("STT", STT))
		sb.WriteString(reportLine("STB", STB))
		sb.WriteString("\n")
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly))
		ExpectMatchesGoldenFile(t, sb.String())
	})
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
	Log(t, "RebalanceFactor=1:", gbStat)

	Log(t, "Best combined overall ranks:")
	Log(t, "#1:", results[0])
	Log(t, "#2:", results[1])
	Log(t, "#3:", results[2])

	Log(t, "\nBest by each ranking:")
	PrintBestByEachRanking(t, results)

	// find as good or better than GoldenButterfly
	betterThanGB := CopyAll(FindMany(results, AsGoodOrBetterThan(gbStat)))
	RankPortfoliosInPlace(betterThanGB)
	Log(t, "As good or better than GoldenButterfly:", len(betterThanGB))
	PrintBestByEachRanking(t, betterThanGB)
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

	plot := func(name string, returns []Percent, nYears int) string {
		data := allPWRs(returns, nYears)
		return fmt.Sprintf("%s, %d-year PWR's:\n", name, nYears) +
			asciigraph.Plot(Floats(data...)) +
			"\n\n"
	}

	t.Run("GoldenButterfly", func(t *testing.T) {
		var sb strings.Builder
		sb.WriteString(plot("GoldenButterfly", GoldenButterfly, 10))
		sb.WriteString(plot("GoldenButterfly", GoldenButterfly, 20))
		sb.WriteString(plot("GoldenButterfly", GoldenButterfly, 30))
		ExpectMatchesGoldenFile(t, sb.String())
	})

	t.Run("8-way", func(t *testing.T) {
		g := NewGomegaWithT(t)

		allReturns, err := portfolioReturns(
			data.PortfolioReturnsList(ParseAssets(`|ST Invest. Grade|Int'l Small|T-Bill|Wellesley|TIPS|REIT|LT STRIPS|Wellington|`)...),
			equalWeightAllocations(8))
		g.Expect(err).To(Succeed())

		var sb strings.Builder
		sb.WriteString(plot("8-way", allReturns, 10))
		sb.WriteString(plot("8-way", allReturns, 20))
		sb.WriteString(plot("8-way", allReturns, 30))
		ExpectMatchesGoldenFile(t, sb.String())
	})
}

func Test_slope(t *testing.T) {
	// 1.00 slope (100%)
	ExpectRoughPercent(t, slope(ReadablePercents(1, 2, 3, 4)), 100)
	ExpectRoughPercent(t, slope(ReadablePercents(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)), 100)
	// different y-intercept, but same 100% slope!
	ExpectRoughPercent(t, slope(ReadablePercents(5, 6, 7, 8, 9, 10)), 100)
	// 50% slope
	ExpectRoughPercent(t, slope(ReadablePercents(1, 1.5, 2, 2.5, 3, 3.5, 4)), 50)
	// negative slope
	ExpectRoughPercent(t, slope(ReadablePercents(-1, -2, -3, -4)), -100)
	ExpectRoughPercent(t, slope(ReadablePercents(-1, -1.5, -2, -2.5, -3, -3.5, -4)), -50)

	t.Run("example series", func(t *testing.T) {
		reportLine := func(name string, returns []Percent) string {
			return fmt.Sprintf("%s slope: %17v\n", name, slope(returns))
		}
		var sb strings.Builder
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly))
		sb.WriteString(reportLine("TSM", TSM))
		sb.WriteString(reportLine("SCV", SCV))
		sb.WriteString(reportLine("LTT", LTT))
		sb.WriteString(reportLine("STT", STT))
		sb.WriteString(reportLine("STB", STB))
		sb.WriteString(reportLine("GLD", GLD))
		ExpectMatchesGoldenFile(t, sb.String())
	})
}

func ExpectRoughPercent(t testing.TB, got Percent, expected Percent, note ...interface{}) {
	t.Helper()
	g := NewGomegaWithT(t)
	g.Expect(got*100).To(BeNumerically("~", expected, 0.0005), note...)
}
