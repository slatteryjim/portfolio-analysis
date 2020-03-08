package portfolio_analysis

type Permutation struct {
	Assets      []string
	Percentages []Percent
}

func Permutations(assets []string, percentages []Percent) []Permutation {
	perms := generatePermutations(assets, percentages)
	// fix up the percentages as a last step, translating them into allocation amounts for each asset
	for _, p := range perms {
		translatePercentages(p.Percentages)
	}
	return perms
}

func generatePermutations(assets []string, percentages []Percent) []Permutation {
	if len(assets) == 0 || len(percentages) == 0 {
		return nil
	}
	var res []Permutation
	thisAsset, remainingAssets := assets[0], assets[1:]
	// append all of the downstream permutations where this asset is simply skipped
	res = append(res, generatePermutations(remainingAssets, percentages)...)
	for i := range percentages {
		if i == len(percentages)-1 {
			// if this is the last percentage, add one permutation where this asset takes it all
			res = append(res, Permutation{
				Assets:      []string{thisAsset},
				Percentages: []Percent{percentages[i]},
			})
		}
		// add on this asset and percentage to all of the other downstream permutations
		for _, remainingPermutation := range generatePermutations(remainingAssets, percentages[i+1:]) {
			res = append(res, Permutation{
				Assets:      append([]string{thisAsset}, remainingPermutation.Assets...),
				Percentages: append([]Percent{percentages[i]}, remainingPermutation.Percentages...),
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
