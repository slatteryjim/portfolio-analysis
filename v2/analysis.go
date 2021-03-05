package v2

import (
	"fmt"
	"time"

	pa "github.com/slatteryjim/portfolio-analysis"
	"github.com/slatteryjim/portfolio-analysis/data"
	. "github.com/slatteryjim/portfolio-analysis/types"
)

// GoFindKAssetsBetterThanX will spin up multiple goroutines to look at all `k` combination of the given names.
// Any combinations that have stats better than the given ideal will be written to the returned channel.
// When all combinations have been evaluated, the returned channel will be closed.
func GoFindKAssetsBetterThanX(ideal *pa.PortfolioStat, k int, names []string) <-chan *pa.PortfolioStat {
	var resultsCh = make(chan *pa.PortfolioStat, 10)
	go func() {
		defer close(resultsCh)
		startAt := time.Now()
		// look at all `k` combinations of assets
		targetAllocations := make([]Percent, k)
		for i := 0; i < k; i++ {
			targetAllocations[i] = Percent(1.0 / float64(k))
		}
		nCr := pa.Binomial(len(names), k)
		fmt.Println()
		fmt.Println(time.Now(), "k =", k, "nCr =", nCr, "TargetAllocations", targetAllocations)

		GoEvaluateAndFindBetterThan := func(assetCombinationBatches <-chan [][]string) <-chan *pa.PortfolioStat {
			out := make(chan *pa.PortfolioStat, 10)
			go func() {
				defer close(out)
				for batch := range assetCombinationBatches {
					for _, assets := range batch {
						returnsList := data.PortfolioReturnsList(assets...)
						returns, err := pa.PortfolioReturns(returnsList, targetAllocations)
						if err != nil {
							panic(err.Error())
						}
						combination := pa.Combination{Assets: assets, Percentages: targetAllocations}
						statIfBetter := pa.EvaluatePortfolioIfAsGoodOrBetterThan(returns, combination, ideal)
						if statIfBetter != nil {
							out <- statIfBetter
						}
					}
				}
			}()
			return out
		}

		combinationsCh := pa.GoEnumerateCombinations(names, k, 10_000)
		// fan out to multiple workers, 9 workers was a sweet spot
		var workersOutput []<-chan *pa.PortfolioStat
		for i := 0; i < 9; i++ {
			results := GoEvaluateAndFindBetterThan(combinationsCh)
			workersOutput = append(workersOutput, results)
		}
		// merge workers' output
		for result := range pa.GoMerge(workersOutput...) {
			resultsCh <- result
		}

		elapsed := time.Since(startAt)
		fmt.Printf("Finished evaluating %d portfolios in %v (%d portfolios per second)\n",
			nCr, elapsed, int(float64(nCr)/elapsed.Seconds()))
	}()
	return resultsCh
}
