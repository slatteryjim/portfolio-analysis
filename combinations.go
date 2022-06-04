package portfolio_analysis

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/slatteryjim/portfolio-analysis/data"
	. "github.com/slatteryjim/portfolio-analysis/types"
)

type Combination struct {
	Assets      []string
	Percentages []Percent
}

func Combinations(assets []string, percentages []Percent) []Combination {
	perms := generateCombinations(assets, percentages)
	// fix up the percentages as a last step, translating them into allocation amounts for each asset
	for _, p := range perms {
		translatePercentages(p.Percentages)
	}
	return perms
}

func generateCombinations(assets []string, percentages []Percent) []Combination {
	if len(assets) == 0 || len(percentages) == 0 {
		return nil
	}
	var res []Combination
	thisAsset, remainingAssets := assets[0], assets[1:]
	// append all of the downstream combinations where this asset is simply skipped
	res = append(res, generateCombinations(remainingAssets, percentages)...)
	for i := range percentages {
		if i == len(percentages)-1 {
			// if this is the last percentage, add one combination where this asset takes it all
			res = append(res, Combination{
				Assets:      []string{thisAsset},
				Percentages: []Percent{percentages[i]},
			})
		}
		// add on this asset and percentage to all of the other downstream combinations
		for _, remainingCombination := range generateCombinations(remainingAssets, percentages[i+1:]) {
			res = append(res, Combination{
				Assets:      append([]string{thisAsset}, remainingCombination.Assets...),
				Percentages: append([]Percent{percentages[i]}, remainingCombination.Percentages...),
			})
		}
	}
	return res
}

// translatePercentages mutates the given slice of percentages, making them represent allocation sizes
// of each asset. So a sequence like [25, 50, 75, 100] would be translated into [25, 25, 25, 25].
func translatePercentages(ps []Percent) {
	var prev Percent = 0
	for i, p := range ps {
		prev, ps[i] = p, p-prev
	}
}

var ErrEndEnumeration = errors.New("end enumeration")

// EnumerateCombinations will enumerate all combinations of k values of xs.
// Trying to make this function screaming fast by not doing any allocations.
// It calls the callback for each enumeration that is available, and the enumerated
// value will be populated in kBuffer each time.
// The callback function may return ErrEndEnumeration to stop enumerations.
func EnumerateCombinations(xs []string, k int, kBuffer []string, combination func() error) error {
	if len(xs) == 0 || k <= 0 {
		return nil
	}
	if len(xs) < k {
		return nil
	}
	if len(xs) == k {
		copy(kBuffer, xs)
		if err := combination(); err != nil {
			if err == ErrEndEnumeration {
				return nil
			}
			return err
		}
		return nil
	}
	// so len(xs) must be > k; we'll have results to return
	var (
		first, rest = xs[0], xs[1:]
		kBufferRest = kBuffer[1:]
	)
	// return cases that include `first`
	kBuffer[0] = first
	// setup a function we might use twice; slight speedup to factor it out
	passThroughAllCombinations := func() error {
		if err := combination(); err != nil {
			if err == ErrEndEnumeration {
				return nil
			}
			return err
		}
		return nil
	}
	if k == 1 {
		// publish this entry
		if err := combination(); err != nil {
			if err == ErrEndEnumeration {
				return nil
			}
			return err
		}
	} else {
		err := EnumerateCombinations(rest, k-1, kBufferRest, passThroughAllCombinations)
		if err != nil {
			return err
		}
	}

	// return cases that don't include `first`
	return EnumerateCombinations(rest, k, kBuffer, passThroughAllCombinations)
}

// GoEnumerateCombinations calls EnumerateCombinations on a goroutine, and returns
// a channel on which it publishes the combinations in batches if the given size.
func GoEnumerateCombinations(xs []string, k, batchSize int) <-chan [][]string {
	out := make(chan [][]string, 1000)
	total := Binomial(len(xs), k)
	go func() {
		defer close(out)
		var (
			kBuffer = make([]string, k)
			batch   = make([][]string, 0, batchSize)
		)
		count := 0
		// ignore returned error, our callback function won't return anything worth propagating
		_ = EnumerateCombinations(xs, k, kBuffer, func() error {
			count++
			if count%20_000_000 == 0 {
				fmt.Printf(" - combination #%d of %d (%0.1f%%) at %s\n", count, total, float64(count)/float64(total)*100, time.Now())
			}

			combination := make([]string, k)
			copy(combination, kBuffer)

			// add combination to batch
			batch = append(batch, combination)

			// dump batch if full
			if len(batch) == batchSize {
				// try to write this batch to channel
				out <- batch
				// reset batch
				batch = make([][]string, 0, batchSize)
			}
			return nil
		})
		// dump any partial batch remaining
		if len(batch) > 0 {
			// try to write this batch to channel
			out <- batch
		}
	}()
	return out
}

// GoMerge merges the output from multiple channels into a single output channel.
// Once all of the given channels are closed, the returned output channel is closed.
func GoMerge(channels ...<-chan *PortfolioStat) <-chan *PortfolioStat {
	var wg sync.WaitGroup
	out := make(chan *PortfolioStat, 1000)

	worker := func(c <-chan *PortfolioStat) {
		defer wg.Done()
		// pump data into shared output
		for x := range c {
			out <- x
		}
	}

	wg.Add(len(channels))
	for _, c := range channels {
		go worker(c)
	}

	// close once all the channels are done
	go func() {
		defer close(out)
		wg.Wait()
	}()
	return out
}

func ConsistsOf(assets []string, expectedAssets map[string]bool) bool {
	if len(assets) != len(expectedAssets) {
		return false
	}
	for _, asset := range assets {
		if !expectedAssets[asset] {
			return false
		}
	}
	return true
}

// The following Binomial function is adapted from:
//   https://github.com/gonum/gonum/blob/889a9573ff783c5b75e4528dbd304233c58ce0c4/stat/combin/combin.go

const (
	errNegInput = "combin: negative input"
	badSetSize  = "combin: n < k"
)

// Binomial returns the binomial coefficient of (n,k), also commonly referred to
// as "n choose k".
//
// The binomial coefficient, C(n,k), is the number of unordered combinations of
// k elements in a set that is n elements big, and is defined as
//
//  C(n,k) = n!/((n-k)!k!)
//
// n and k must be non-negative with n >= k, otherwise Binomial will panic.
// No check is made for overflow.
func Binomial(n, k int) int {

	if n < 0 || k < 0 {
		panic(errNegInput)
	}
	if n < k {
		panic(badSetSize)
	}
	// (n,k) = (n, n-k)
	if k > n/2 {
		k = n - k
	}
	b := 1
	for i := 1; i <= k; i++ {
		b = (n - k + i) * b / i
	}
	return b
}

func MustGoldenButterflyStat() *PortfolioStat {
	assets := []string{"LTT", "Gold", "STT", "SCV", "TSM"}
	targetAllocations := ReadablePercents(20, 20, 20, 20, 20)
	returnsList := data.PortfolioReturnsList(assets...)
	returns, err := PortfolioReturns(returnsList, targetAllocations)
	if err != nil {
		panic(err.Error())
	}
	stat := EvaluatePortfolio(returns, Combination{Assets: assets, Percentages: targetAllocations})
	return stat
}
