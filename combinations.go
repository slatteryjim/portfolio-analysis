package portfolio_analysis

import (
	"errors"

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
