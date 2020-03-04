package portfolio_analysis

type Permutation struct {
	Assets      []string
	Percentages []float64
}

func Permutations(assets []string, percentages []float64) []Permutation {
	if len(assets) == 0 || len(percentages) == 0 {
		return nil
	}
	var res []Permutation
	thisAsset, remainingAssets := assets[0], assets[1:]
	// append all of the downstream permutations where this asset is simply skipped
	res = append(res, Permutations(remainingAssets, percentages)...)
	for i := range percentages {
		if i == len(percentages)-1 {
			// if this is the last percentage, add one permutation where this asset takes it all
			res = append(res, Permutation{
				Assets:      []string{thisAsset},
				Percentages: []float64{percentages[i]},
			})
		}
		// add on this asset and percentage to all of the other downstream permutations
		for _, remainingPermutation := range Permutations(remainingAssets, percentages[i+1:]) {
			res = append(res, Permutation{
				Assets:      append([]string{thisAsset}, remainingPermutation.Assets...),
				Percentages: append([]float64{percentages[i]}, remainingPermutation.Percentages...),
			})
		}
	}
	return res
}
