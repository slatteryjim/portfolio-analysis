package portfolio_analysis

type permutation struct {
	assets      []string
	percentages []float64
}

func permutations(assets []string, percentages []float64) []permutation {
	if len(assets) == 0 || len(percentages) == 0 {
		return nil
	}
	var res []permutation
	thisAsset, remainingAssets := assets[0], assets[1:]
	// append all of the downstream permutations where this asset is simply skipped
	res = append(res, permutations(remainingAssets, percentages)...)
	for i := range percentages {
		if i == len(percentages)-1 {
			// if this is the last percentage, add one permutation where this asset takes it all
			res = append(res, permutation{
				assets:      []string{thisAsset},
				percentages: []float64{percentages[i]},
			})
		}
		// add on this asset and percentage to all of the other downstream permutations
		for _, remainingPermutation := range permutations(remainingAssets, percentages[i+1:]) {
			res = append(res, permutation{
				assets:      append([]string{thisAsset}, remainingPermutation.assets...),
				percentages: append([]float64{percentages[i]}, remainingPermutation.percentages...),
			})
		}
	}
	return res
}
