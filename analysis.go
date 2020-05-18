package portfolio_analysis

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/slatteryjim/portfolio-analysis/data"
)

type (
	// Percent is a percentage using the range 0.00 - 1.00.
	Percent float64

	// GrowthMultiplier is a focused around 1.00. So a 5% return would be represented
	// as a 1.05 GrowthMultiplier.
	GrowthMultiplier float64
)

func (p Percent) String() string {
	return formatFloat(p.Float()*100, 12) + "%"
}

// readablePercents takes easy-to-read percentages using the range 0 - 100, and returns
// a slice of Percent (each using the range 0.00 - 1.00).
func readablePercents(xs ...float64) []Percent {
	res := make([]Percent, len(xs))
	for i, x := range xs {
		res[i] = readablePercent(x)
	}
	return res
}

// readablePercent takes an easy-to-read percentage using the range 0 - 100, and returns
// a Percent (using the range 0.00 - 1.00).
func readablePercent(x float64) Percent {
	return Percent(x / 100)
}

func (g GrowthMultiplier) Float() float64 { return float64(g) }

// GrowthMultiplier is a focused around 1.00. So a 5% return would be represented
// as a 1.05 GrowthMultiplier.
func (p Percent) GrowthMultiplier() GrowthMultiplier {
	return GrowthMultiplier(p + 1)
}

func (p Percent) Float() float64 { return float64(p) }

// PercentSlice attaches the methods of Interface to []PercentSlice, sorting in increasing order
// (not-a-number values are treated as less than other values).
type PercentSlice []Percent

func (p PercentSlice) Len() int { return len(p) }
func (p PercentSlice) Less(i, j int) bool {
	return p[i] < p[j] || math.IsNaN(p[i].Float()) && !math.IsNaN(p[j].Float())
}
func (p PercentSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// some data from Boglehead's "Simba Spreadsheet"
// 1969 start date - returns (percentage as a float, 100.0 == 100%)
var (
	TSM = readablePercents(data.MustFind("TSM").AnnualReturnsStartingIn(1969)...)
	SCV = readablePercents(data.MustFind("SCV").AnnualReturnsStartingIn(1969)...)
	LTT = readablePercents(data.MustFind("LTT").AnnualReturnsStartingIn(1969)...)
	STT = readablePercents(data.MustFind("STT").AnnualReturnsStartingIn(1969)...)
	STB = readablePercents(data.MustFind("STB").AnnualReturnsStartingIn(1969)...)
	GLD = readablePercents(data.MustFind("Gold").AnnualReturnsStartingIn(1969)...)

	GoldenButterfly, _ = portfolioReturns([][]Percent{TSM, SCV, LTT, STT, GLD}, readablePercents(20, 20, 20, 20, 20))
)

// take a list of multiple asset returns, and the percentage to rebalance each year. Returns the resultant set of returns.
// Example:
//     portfolio_returns([TSM, ITB], [60, 40])
func portfolioReturns(returnsList [][]Percent, targetAllocations []Percent) ([]Percent, error) {
	if math.Abs(sum(targetAllocations).Float()-1.00) > 0.00000000000001 {
		return nil, fmt.Errorf("targetAllocations must sum to 100%%, got %v", sum(targetAllocations))
	}
	if len(targetAllocations) != len(returnsList) {
		return nil, fmt.Errorf("lists must have the same length: targetAllocations (%d), returnsList (%d)", len(targetAllocations), len(returnsList))
	}
	res := make([]Percent, 0, len(returnsList[0]))
	zipWalk(returnsList, func(yearsReturns []Percent) {
		var sum Percent
		for i := range yearsReturns {
			sum += yearsReturns[i] * targetAllocations[i]
		}
		res = append(res, sum)
	})

	return res, nil
}

// PortfolioTradingSimulation takes a list of multiple asset returns, and the percentage to rebalance each year.
// Returns the resultant set of returns.
// I want to play with rebalance_factor. Instead of rebalancing exactly, we can overshoot or undershoot the
// transactions, to "juice" it up.
// I want to see how that tweak in rebalancing strategy affects the performance of various portfolios.
// Example:
//     portfolio_trading_simulation([TSM, ITB], [60, 40])
func PortfolioTradingSimulation(returnsList [][]Percent, targetAllocations []Percent, rebalanceFactor float64) ([]Percent, error) {
	if math.Abs(sum(targetAllocations).Float()-1.00) > 0.00000000000001 {
		return nil, fmt.Errorf("targetAllocations must sum to 100%%, got %v", sum(targetAllocations))
	}
	if len(targetAllocations) != len(returnsList) {
		return nil, fmt.Errorf("lists must have the same length: targetAllocations (%d), returnsList (%d)", len(targetAllocations), len(returnsList))
	}
	var cumulativeReturnsL = make([]Percent, 0, len(returnsList[0]))
	{
		var (
			// our initial allocation of 1.000 will simply be according to the target allocations.
			// Shorthand to clone targetAllocations. See: https://github.com/go101/go101/wiki/How-to-efficiently-clone-a-slice%3F
			allocations = append(targetAllocations[:0:0], targetAllocations...)
			// slice to reuse for calculations in each iteration
			eoyAllocation = make([]Percent, len(targetAllocations))
		)
		zipWalk(returnsList, func(oneReturnSet []Percent) {
			// fmt.Println("\noneReturnSet", fmt.Sprintf("%v", oneReturnSet))
			// apply the returns to the current allocation
			startSum := sum(allocations)
			var eoySum Percent
			for i := range allocations {
				value := allocations[i] * (oneReturnSet[i] + 1)
				eoyAllocation[i] = value
				eoySum += value
			}
			cumulativeReturnsL = append(cumulativeReturnsL, (eoySum/startSum)-1)
			// fmt.Println("eoyAllocation", fmt.Sprintf("%v", eoyAllocation), "eoySum", eoySum)

			// update allocations -- calculate the transactions to perform, and apply
			for i := range targetAllocations {
				target := targetAllocations[i] * eoySum
				transaction := (target - eoyAllocation[i]) * Percent(rebalanceFactor)
				// fmt.Println("transaction", i+1, transaction)
				allocation := eoyAllocation[i] + transaction
				if allocation < 0 {
					panic("no allocation can go below zero! maybe rebalanceFactor is too extreme?")
				}
				allocations[i] = allocation
			}
			// fmt.Println("post-transaction allocation", fmt.Sprintf("%v", allocations))
			return
		})
	}
	return cumulativeReturnsL, nil
}

// calculates the cumulative growth of the returns
func cumulative(returns []Percent) GrowthMultiplier {
	var product GrowthMultiplier = 1
	for _, x := range returns {
		product *= x.GrowthMultiplier()
	}
	return product
}

// cumulativeList calculates the cumulative growth of the returns.
// It returns a list that always starts with `1`.
func cumulativeList(returns []Percent) []GrowthMultiplier {
	res := make([]GrowthMultiplier, 0, len(returns)+1)
	var acc GrowthMultiplier = 1
	res = append(res, acc)
	for _, r := range returns {
		acc *= r.GrowthMultiplier()
		res = append(res, acc)
	}
	return res
}

// CAGR - calculates the compound annual growth rate of the returns
func cagr(returns []Percent) Percent {
	n := float64(len(returns))
	return Percent(math.Pow(cumulative(returns).Float(), 1/n) - 1)
}

// averageReturn returns the average of the given returns.
// See: https://portfoliocharts.com/portfolio/annual-returns/
func averageReturn(returns []Percent) Percent {
	return sum(returns) / Percent(len(returns))
}

// baselineLongTermReturn returns the:
// "Conservative practical long-term compound return excluding
// the worst outliers (15th percentile 15-year real CAGR)"
// See: https://portfoliocharts.com/portfolio/long-term-returns/
func baselineLongTermReturn(returns []Percent) Percent {
	return baselineReturn(returns, 15, readablePercent(15))
}

// baselineShortTermReturn returns the:
// "Conservative practical short-term compound return excluding
// the worst outliers (15th percentile 3-year real CAGR)"
// See: https://portfoliocharts.com/portfolio/long-term-returns/
func baselineShortTermReturn(returns []Percent) Percent {
	return baselineReturn(returns, 3, readablePercent(15))
}

func baselineReturn(returns []Percent, nYears int, percentile Percent) Percent {
	if len(returns) == 0 {
		return 0
	}
	if percentile < 0 || percentile >= 1.00 {
		panic(fmt.Sprintf("percentile must be in the range [0,1.00) but got %f", percentile))
	}
	cagrs := make([]Percent, 0, len(returns)-nYears)
	for _, slice := range subSlices(returns, nYears) {
		cagrs = append(cagrs, cagr(slice))
	}
	sort.Sort(PercentSlice(cagrs))
	return cagrs[int(Percent(len(cagrs))*percentile)]
}

// standardDeviation returns "The statistical uncertainty of the average real return"
// See: https://portfoliocharts.com/portfolio/annual-returns/
func standardDeviation(xs []Percent) Percent {
	n := Percent(len(xs))
	if n == 0 {
		panic("returns list must not be empty")
	}
	var sumOfSquaredDiffs Percent
	{
		avg := sum(xs) / n
		for _, x := range xs {
			sumOfSquaredDiffs += Percent(math.Pow((x - avg).Float(), 2))
		}
	}
	return Percent(math.Sqrt(((1 / n) * sumOfSquaredDiffs).Float()))
}

// swr returns the Safe-withdrawal rate
func swr(returns []Percent) Percent {
	cumulativeGrowth := cumulativeList(returns)
	return harmonicMean(cumulativeGrowth) / Percent(len(cumulativeGrowth))
}

// pwr returns the perpetual withdrawal rate.
// The amount that can be safely withdrawn annually (before growth), such that at the end of the series,
// the account balance will match what we started with.
func pwr(returns []Percent) Percent {
	var preservationPercent = 1.00
	return swr(returns) * Percent(preservationPercent-1/cumulative(returns).Float())
}

// minPWR looks at all of the nYears-long periods and evaluates their PWR. Returns the min PWR.
func minPWR(returns []Percent, nYears int) (rate Percent, startAtIndex int) {
	if nYears == 0 {
		return 0, 0
	}

	rate = math.MaxFloat64
	startAtIndex = math.MaxInt64
	for i, slice := range subSlices(returns, nYears) {
		thisPWR := pwr(slice)
		if thisPWR < rate {
			rate = thisPWR
			startAtIndex = i
		}
	}
	return rate, startAtIndex
}

// minSWR looks at all of the nYears-long periods and evaluates their SWR. Returns the min SWR.
func minSWR(returns []Percent, nYears int) (rate Percent, startAtIndex int) {
	if nYears == 0 {
		return 0, 0
	}
	rate = Percent(math.MaxFloat64)
	startAtIndex = math.MaxInt64
	for i, slice := range subSlices(returns, nYears) {
		thisSWR := swr(slice)
		if thisSWR < rate {
			rate = thisSWR
			startAtIndex = i
		}
	}
	return rate, startAtIndex
}

// minPWRAndSWR calculates both PWR and SWR at the same time, for efficiency.
func minPWRAndSWR(returns []Percent, nYears int) (Percent, Percent) {
	if nYears == 0 {
		return 0, 0
	}
	var (
		minPerpetual = Percent(math.MaxFloat64)
		minSafe      = Percent(math.MaxFloat64)
	)
	for _, slice := range subSlices(returns, nYears) {
		thisPWR, thisSWR := pwrAndSWR(slice)
		if thisSWR < minSafe {
			minSafe = thisSWR
		}
		if thisPWR < minPerpetual {
			minPerpetual = thisPWR
		}
	}
	return minPerpetual, minSafe
}

// pwrAndSWR calculates both PWR and SWR at the same time, for efficiency.
func pwrAndSWR(returns []Percent) (Percent, Percent) {
	cumulativeGrowth := cumulativeList(returns)
	var swr = harmonicMean(cumulativeGrowth) / Percent(len(cumulativeGrowth))

	var pwr Percent
	{
		const preservationPercent = 1.00
		cumulativeReturn := cumulativeGrowth[len(cumulativeGrowth)-1]
		pwr = swr * Percent(preservationPercent-1/cumulativeReturn.Float())
	}
	return pwr, swr
}

// startDateSensitivity is a simple quantitative way to measure the dependability of a portfolio.
// See: https://portfoliocharts.com/portfolio/start-date-sensitivity/
func startDateSensitivity(returns []Percent) Percent {
	var (
		worstShortfall  Percent
		bestImprovement Percent
	)
	for _, twentyYears := range subSlices(returns, 20) {
		firstTwenty := cagr(twentyYears[:10])
		secondTwenty := cagr(twentyYears[10:])
		diff := secondTwenty - firstTwenty
		// fmt.Println(i+1, firstTwenty, secondTwenty, "diff:", diff)
		// shortfall?
		if firstTwenty > secondTwenty {
			if diff < worstShortfall {
				worstShortfall = diff
			}
		} else {
			// improvement
			if diff > bestImprovement {
				bestImprovement = diff
			}
		}
	}
	return bestImprovement - worstShortfall
}

// returns all of the sub-slices of length n.
func subSlices(orig []Percent, n int) [][]Percent {
	length := len(orig)
	if n > length {
		panic(fmt.Sprintf("n (%d) cannot be greater than the length of the original slice (%d)", n, length))
	}
	if length == 0 {
		return nil
	}
	res := make([][]Percent, 0, length)
	if len(orig) <= n {
		res = append(res, orig)
		return res
	}
	start, end := 0, n
	for end <= length {
		res = append(res, orig[start:end])
		start++
		end++
	}
	return res
}

// harmonicMean returns the harmonic mean of the given numbers, which must all be greater than zero.
// See: https://en.wikipedia.org/wiki/Harmonic_mean#Definition
func harmonicMean(xs []GrowthMultiplier) Percent {
	var acc GrowthMultiplier
	for i, x := range xs {
		if x <= 0 {
			panic(fmt.Sprintf("harmonicMean requires inputs greater than zero, but element #%x is %v", i+1, x))
		}
		acc += 1 / x
	}
	return Percent(len(xs)) / Percent(acc)
}

// zipWalk zips together the streams, calling fn for each set of numbers.
// Really this is a like a transpose operation.
// Example: zip([[1,2,3],[4,5,6]]) => [[1,4],[2,5],[3,6]]
func zipWalk(streams [][]Percent, fn func([]Percent)) {
	var width = len(streams)
	if width == 0 {
		return
	}
	length := len(streams[0])
	for i := range streams {
		if len(streams[i]) != length {
			panic(fmt.Sprintf("expected stream %d to have length %d", i+1, length))
		}
	}
	var (
		// reuse this slice for each iteration
		snapshot = make([]Percent, width)
	)
	for i := range streams[0] {
		// build slice of all the elements at position i
		for j := 0; j < width; j++ {
			snapshot[j] = streams[j][i]
		}
		fn(snapshot)
	}
	return
}

// sum returns the sum of the given float64 values.
func sum(xs []Percent) Percent {
	var sum Percent
	for _, x := range xs {
		sum += x
	}
	return sum
}

func minFloats(xs []GrowthMultiplier) GrowthMultiplier {
	if len(xs) == 0 {
		panic("can't return the minimum of an empty slice")
	}
	var min GrowthMultiplier = math.MaxFloat64
	for _, x := range xs {
		if x < min {
			min = x
		}
	}
	return min
}

func formatFloat(f float64, prec int) string {
	s := strconv.FormatFloat(f, 'f', prec, 64)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	return s
}
