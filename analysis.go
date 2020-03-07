package portfolio_analysis

import (
	"fmt"
	"math"
	"sort"
)

// some data from Boglehead's "Simba Spreadsheet"
// returns (percentage as a float, 100.0 == 100%) - starting in 1969
var (
	TSM_nonInflationAdjusted = []float64{-10.28, 0.9, 17.63, 17.93, -18.55, -28.42, 38.42, 26.54, -2.68, 9.23, 25.51, 33.62, -3.79, 18.66, 23.42, 3.01, 32.51, 16.05, 2.23, 17.89, 29.12, -6.22, 34.15, 8.92, 10.62, -0.17, 35.79, 20.96, 30.99, 23.26, 23.81, -10.57, -10.89, -20.95, 31.42, 12.61, 6.09, 15.63, 5.57, -36.99, 28.83, 17.26, 1.08, 16.38, 33.52, 12.56, 0.39, 12.66, 21.17, -5.17, 30.8}

	TSM = []float64{-15.51, -4.43, 13.91, 14.05, -25.07, -36.28, 29.44, 20.67, -8.79, 0.20, 10.78, 18.75, -11.67, 14.29, 18.91, -0.91, 27.66, 14.79, -2.11, 12.90, 23.39, -11.62, 30.16, 5.85, 7.67, -2.77, 32.42, 17.07, 28.80, 21.31, 20.58, -13.49, -12.25, -22.79, 28.99, 9.06, 2.58, 12.76, 1.43, -37.05, 25.41, 15.53, -1.83, 14.39, 31.55, 11.71, -0.34, 10.37, 18.67, -6.95, 27.88}
	SCV = []float64{-25.83, -5.18, 14.55, 4.9, -31.81, -32.36, 44.92, 43.35, 7.7, 7.51, 19.41, 11.36, 5.37, 23.69, 33.48, -1.68, 26.13, 6.17, -11.12, 23.9, 7.36, -26.33, 37.39, 25.41, 16.82, -3.32, 23.32, 19.47, 32.38, -6.69, 0.65, 17.89, 11.96, -16.19, 34.66, 19.66, 2.57, 16.29, -10.72, -32.12, 26.88, 22.98, -6.85, 16.73, 34.56, 9.72, -5.34, 22.24, 9.49, -13.87, 20.02}
	LTT = []float64{-11.76, 11.1, 5.59, -5.54, -7.03, -6.58, 1.23, 11.59, -6.39, -9.66, -12.27, -13.8, -7.91, 36.44, -1.81, 10.35, 26.66, 22.66, -6.87, 4.49, 13.58, 0.11, 14.88, 4.84, 14.03, -10.11, 27.37, -4.13, 13.07, 11.64, -11.19, 16.25, 2.55, 14, 0.52, 4.24, 2.91, -0.74, 5.43, 23.83, -15.29, 7.78, 25.17, 1.72, -14.03, 24.09, -2.06, -0.76, 6.44, -3.51, 11.74}
	STT = []float64{-2.41, 6.86, 3.58, 0.38, -3.9, -5.41, 1.44, 3.13, -2.74, -5.07, -4.81, -3.09, 3.52, 16.65, 5.06, 9.31, 9.54, 8.83, 1.17, 1.61, 5.85, 3.31, 8.27, 3.23, 2.48, -2.18, 8.04, 1.61, 4.8, 5.25, 0.14, 4.44, 6.66, 3.34, -0.03, -2.34, -1.8, 1.28, 3.03, 6.5, -1.94, 0.69, -1.46, -1.35, -1.19, -0.25, -0.22, -1.27, -1.68, -0.44, 1.2}
	STB = []float64{-3.29, 5.22, 4.14, 0.00, -3.96, -5.91, 1.06, 5.65, -2.96, -5.62, -5.54, -4.24, 2.55, 18.98, 4.91, 9.77, 11.22, 10.10, 0.48, 1.72, 6.66, 3.31, 9.70, 3.74, 4.16, -3.37, 10.09, 1.19, 5.25, 5.92, -0.59, 5.28, 7.22, 3.69, 1.52, -1.44, -1.97, 1.58, 3.10, 5.41, 1.62, 2.50, 0.11, 0.30, -1.32, 0.50, 0.19, -0.57, -0.91, -0.55, 2.51}
	GLD = []float64{-21.16, 0.31, 12.72, 43.08, 59.2, 48.32, -30.23, -8.74, 15.06, 24.04, 105.51, -0.26, -37.86, 7.6, -18.16, -22.28, 1.7, 17.95, 19.02, -19.56, -6.81, -8.33, -12.52, -8.68, 13.92, -4.87, -1.65, -7.74, -23.24, -2.43, -1.71, -9.55, -0.39, 20.78, 19.19, 1.41, 12.97, 19.3, 25.82, 5.36, 20.18, 26.05, 5.53, 6.52, -29.03, -1.19, -12.29, 6.63, 9.27, -3.24, 15.89}

	GoldenButterfly, _ = portfolioReturns([][]float64{TSM, SCV, LTT, STT, GLD}, []float64{20, 20, 20, 20, 20})
)

// take a list of multiple asset returns, and the percentage to rebalance each year. Returns the resultant set of returns.
// Example:
//     portfolio_returns([TSM, ITB], [60, 40])
func portfolioReturns(returnsList [][]float64, percentages []float64) ([]float64, error) {
	if sum(percentages) != 100 {
		return nil, fmt.Errorf("percentages must sum to 100%%, got %v", sum(percentages))
	}
	if len(percentages) != len(returnsList) {
		return nil, fmt.Errorf("lists must have the same length: percentages (%d), returnsList (%d)", len(percentages), len(returnsList))
	}
	percentageMultipliers := mapFloats(percentages, func(x float64) float64 {
		return x / 100
	})
	res := make([]float64, 0, len(returnsList[0]))
	zipWalk(returnsList, func(yearsReturns []float64) {
		sum := 0.0
		for i := range yearsReturns {
			sum += yearsReturns[i] * percentageMultipliers[i]
		}
		res = append(res, sum)
	})

	return res, nil
}

// percentages convert numbers from 0-100.0 to a "growth multiplier" percentage based around 1.00.
// Example: 20 => 1.20
func percentages(xs []float64) []float64 {
	return mapFloats(xs, func(x float64) float64 {
		return (100 + x) / 100
	})
}

// calculates the cumulative growth of the returns
func cumulative(returns []float64) float64 {
	return product(percentages(returns))
}

// # calculates the cumulative growth of the returns
func cumulativeList(returns []float64) []float64 {
	res := make([]float64, len(returns))
	acc := 1.0
	for i, r := range percentages(returns) {
		acc *= r
		res[i] = acc
	}
	return res
}

// CAGR - calculates the compound annual growth rate of the returns
func cagr(returns []float64) float64 {
	n := float64(len(returns))
	return math.Pow(cumulative(returns), 1/n) - 1
}

// averageReturn returns the average of the given returns.
// See: https://portfoliocharts.com/portfolio/annual-returns/
func averageReturn(returns []float64) float64 {
	return sum(returns) / float64(len(returns)) / 100
}

// baselineLongTermReturn returns the:
// "Conservative practical long-term compound return excluding
// the worst outliers (15th percentile 15-year real CAGR)"
// See: https://portfoliocharts.com/portfolio/long-term-returns/
func baselineLongTermReturn(returns []float64) float64 {
	return baselineReturn(returns, 15, 15)
}

// baselineShortTermReturn returns the:
// "Conservative practical short-term compound return excluding
// the worst outliers (15th percentile 3-year real CAGR)"
// See: https://portfoliocharts.com/portfolio/long-term-returns/
func baselineShortTermReturn(returns []float64) float64 {
	return baselineReturn(returns, 3, 15)
}

func baselineReturn(returns []float64, nYears int, percentile float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	if percentile < 0 || percentile >= 100 {
		panic(fmt.Sprintf("percentile must be in the range [0,100) but got %f", percentile))
	}
	cagrs := make([]float64, 0, len(returns)-nYears)
	for _, slice := range subSlices(returns, nYears) {
		cagrs = append(cagrs, cagr(slice))
	}
	sort.Sort(sort.Float64Slice(cagrs))
	return cagrs[int(float64(len(cagrs))*percentile/100)]
}

// standardDeviation returns "The statistical uncertainty of the average real return"
// See: https://portfoliocharts.com/portfolio/annual-returns/
func standardDeviation(xs []float64) float64 {
	n := float64(len(xs))
	if n == 0 {
		panic("returns list must not be empty")
	}
	var sumOfSquaredDiffs float64
	{
		avg := sum(xs) / n
		for _, x := range xs {
			sumOfSquaredDiffs += math.Pow(x-avg, 2)
		}
	}
	return math.Sqrt((1 / n) * sumOfSquaredDiffs)
}

// swr returns the Safe-withdrawal rate
func swr(returns []float64) float64 {
	// prepend 1.0 to the list of returns
	cumulativeGrowth := make([]float64, 0, len(returns)+1)
	cumulativeGrowth = append(cumulativeGrowth, 1.0)
	cumulativeGrowth = append(cumulativeGrowth, cumulativeList(returns)...)

	return harmonicMean(cumulativeGrowth) / float64(len(cumulativeGrowth))
}

// pwr returns the perpetual withdrawal rate.
// The amount that can be safely withdrawn annually (before growth), such that at the end of the series,
// the account balance will match what we started with.
func pwr(returns []float64) float64 {
	preservationPercent := 1.00
	return swr(returns) * (preservationPercent - 1/cumulative(returns))
}

// minPWR looks at all of the nYears-long periods and evaluates their PWR. Returns the min PWR.
func minPWR(returns []float64, nYears int) (rate float64, startAtIndex int) {
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
func minSWR(returns []float64, nYears int) (rate float64, startAtIndex int) {
	if nYears == 0 {
		return 0, 0
	}

	rate = math.MaxFloat64
	startAtIndex = math.MaxInt64
	for i, slice := range subSlices(returns, nYears) {
		thisPWR := swr(slice)
		if thisPWR < rate {
			rate = thisPWR
			startAtIndex = i
		}
	}
	return rate, startAtIndex
}

// startDateSensitivity is a simple quantitative way to measure the dependability of a portfolio.
// See: https://portfoliocharts.com/portfolio/start-date-sensitivity/
func startDateSensitivity(returns []float64) float64 {
	var (
		worstShortfall  = 0.0
		bestImprovement = 0.0
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
func subSlices(orig []float64, n int) [][]float64 {
	length := len(orig)
	if n > length {
		panic(fmt.Sprintf("n (%d) cannot be greater than the length of the original slice (%d)", n, length))
	}
	if length == 0 {
		return nil
	}
	res := make([][]float64, 0, length)
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
func harmonicMean(xs []float64) float64 {
	acc := 0.0
	for i, x := range xs {
		if x <= 0 {
			panic(fmt.Sprintf("harmonicMean requires inputs greater than zero, but element #%x is %v", i+1, x))
		}
		acc += 1 / x
	}
	return float64(len(xs)) / acc
}

// zipWalk zips together the streams, calling fn for each set of numbers.
// Really this is a like a transpose operation.
// Example: zip([[1,2,3],[4,5,6]]) => [[1,4],[2,5],[3,6]]
func zipWalk(streams [][]float64, fn func([]float64)) {
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
		snapshot = make([]float64, width)
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
func sum(xs []float64) float64 {
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum
}

// product returns the product of the given float64 values.
func product(xs []float64) float64 {
	product := 1.0
	for _, x := range xs {
		product *= x
	}
	return product
}

func reduceFloats(init float64, xs []float64, fn func(x, y float64) float64) float64 {
	res := init
	for _, x := range xs {
		res = fn(res, x)
	}
	return res
}

func mapFloats(xs []float64, fn func(float64) float64) []float64 {
	res := make([]float64, len(xs))
	for i := range xs {
		res[i] = fn(xs[i])
	}
	return res
}

func minFloats(xs []float64) float64 {
	if len(xs) == 0 {
		panic("can't return the minimum of an empty slice")
	}
	min := math.MaxFloat64
	for _, x := range xs {
		if x < min {
			min = x
		}
	}
	return min
}
