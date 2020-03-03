package portfolio_analysis

type drawdownSequence struct {
	startIndex        int
	cumulativeReturns []float64
	// did the sequence recover by the end?
	recovered bool
}

// drawdowns returns a list of all the sequences of negative cumulative returns.
func drawdowns(returns []float64) []drawdownSequence {
	if len(returns) == 0 {
		return nil
	}
	var res []drawdownSequence
	for i := range returns {
		drawdownSeq, recovered := leadingDrawdownSequence(returns[i:])
		if len(drawdownSeq) > 0 {
			res = append(res, drawdownSequence{
				startIndex:        i,
				cumulativeReturns: drawdownSeq,
				recovered:         recovered,
			})
		}
	}
	return res
}

// leadingDrawdownSequence returns the drawdown sequence, if this list starts with one.
// Also returns a boolean indicating whether the drawdown sequence ended in the end,
// or false if it never ended.
func leadingDrawdownSequence(returns []float64) ([]float64, bool) {
	end := -1
	cumulativeReturns := cumulativeList(returns)
	for i, value := range cumulativeReturns {
		if value >= 1 {
			end = i
			break
		}
	}
	if end == -1 {
		return cumulativeReturns, false
	}
	return cumulativeReturns[0:end], true
}

func ulcerScore(cumulativeReturns []float64, recovered bool) float64 {
	if len(cumulativeReturns) == 0 {
		return 0
	}
	score := 0.0
	for _, x := range cumulativeReturns {
		score += (1 - x) * 10
	}
	if !recovered {
		// that's scary, we ended without a recovery, let's increase the score
		score *= 2
	}
	return score
}

func drawdownScores(returns []float64) (maxUlcerScore, deepestDrawdown float64, longestDrawdown int) {
	maxUlcerScore = 0.0
	deepestDrawdown = 0.0
	longestDrawdown = 0
	for _, dd := range drawdowns(returns) {
		score := ulcerScore(dd.cumulativeReturns, dd.recovered)
		if score > maxUlcerScore {
			maxUlcerScore = score
		}
		lowestPoint := minFloats(dd.cumulativeReturns) - 1
		if lowestPoint < deepestDrawdown {
			deepestDrawdown = lowestPoint
		}
		length := len(dd.cumulativeReturns)
		if length > longestDrawdown {
			longestDrawdown = length
		}
	}
	return maxUlcerScore, deepestDrawdown, longestDrawdown
}
