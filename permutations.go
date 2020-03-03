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
	res = append(res, permutations(remainingAssets, percentages)...)
	for i := range percentages {
		if i == len(percentages)-1 {
			res = append(res, permutation{
				assets:      []string{thisAsset},
				percentages: []float64{percentages[i]},
			})
		}
		for _, remainingPermutation := range permutations(remainingAssets, percentages[i+1:]) {
			res = append(res, permutation{
				assets:      append([]string{thisAsset}, remainingPermutation.assets...),
				percentages: append([]float64{percentages[i]}, remainingPermutation.percentages...),
			})
		}
	}
	return res
}
