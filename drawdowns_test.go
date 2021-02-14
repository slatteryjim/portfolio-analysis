package portfolio_analysis

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	. "github.com/slatteryjim/portfolio-analysis/types"
)

func Test_leadingDrawdownSequence(t *testing.T) {
	g := NewGomegaWithT(t)

	verify := func(returns []Percent, expectedSequence []GrowthMultiplier, expectedEnded bool) {
		t.Helper()
		sequence, ended := leadingDrawdownSequence(returns)
		g.Expect(sequence).To(Equal(expectedSequence), "sequence")
		g.Expect(ended).To(Equal(expectedEnded), "ended")
	}

	// empty
	verify(nil, []GrowthMultiplier{}, false)
	verify([]Percent{}, []GrowthMultiplier{}, false)

	// doesn't start with a drawdown
	verify(ReadablePercents(1), []GrowthMultiplier{}, true)
	verify(ReadablePercents(0, -1, 2), []GrowthMultiplier{}, true)

	// starts with a drawdown
	verify(ReadablePercents(-1), []GrowthMultiplier{0.99}, false)
	verify(ReadablePercents(-1, 2), []GrowthMultiplier{0.99}, true)
	verify(ReadablePercents(-1, -1, 3), []GrowthMultiplier{0.99, 0.9801}, true)
	verify(ReadablePercents(-1, -1, 3, -5, -50), []GrowthMultiplier{0.99, 0.9801}, true)

	verify(ReadablePercents(-50, 100), []GrowthMultiplier{0.50}, true)

}

func Test_drawdowns(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(drawdowns(nil)).To(BeEmpty())
	g.Expect(drawdowns(ReadablePercents(0))).To(BeEmpty())
	g.Expect(drawdowns(ReadablePercents(1))).To(BeEmpty())
	g.Expect(drawdowns(ReadablePercents(-1))).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []GrowthMultiplier{0.99}, recovered: false},
	}))
	g.Expect(drawdowns(ReadablePercents(-1, 2))).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []GrowthMultiplier{0.99}, recovered: true},
	}))
	g.Expect(drawdowns(ReadablePercents(-1, -1, 3))).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []GrowthMultiplier{0.99, 0.9801}, recovered: true},
		{startIndex: 1, cumulativeReturns: []GrowthMultiplier{0.99}, recovered: true},
	}))
	g.Expect(drawdowns(ReadablePercents(-1, -1, 1))).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []GrowthMultiplier{0.99, 0.9801, 0.989901}, recovered: false},
		{startIndex: 1, cumulativeReturns: []GrowthMultiplier{0.99, 0.9999}, recovered: false},
	}))
	g.Expect(drawdowns(ReadablePercents(-1, 3, -1, -1, 3))).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []GrowthMultiplier{0.99}, recovered: true},
		{startIndex: 2, cumulativeReturns: []GrowthMultiplier{0.99, 0.9801}, recovered: true},
		{startIndex: 3, cumulativeReturns: []GrowthMultiplier{0.99}, recovered: true},
	}))
	g.Expect(drawdowns(ReadablePercents(-1, -1, 3, -5, -50, 100))).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []GrowthMultiplier{0.99, 0.9801}, recovered: true},
		{startIndex: 1, cumulativeReturns: []GrowthMultiplier{0.99}, recovered: true},
		{startIndex: 3, cumulativeReturns: []GrowthMultiplier{0.95, 0.475, 0.95}, recovered: false},
		{startIndex: 4, cumulativeReturns: []GrowthMultiplier{0.50}, recovered: true},
	}))
}

func Test_ulcerScore(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(ulcerScore(nil, false)).To(Equal(0.0))
	g.Expect(ulcerScore(nil, true)).To(Equal(0.0))
	g.Expect(ulcerScore([]GrowthMultiplier{0.99}, true)).To(Equal(0.10000000000000009))
	g.Expect(ulcerScore([]GrowthMultiplier{0.99}, false)).To(Equal(0.20000000000000018))
	g.Expect(ulcerScore([]GrowthMultiplier{0.90}, true)).To(Equal(0.9999999999999998))
	g.Expect(ulcerScore([]GrowthMultiplier{0.90, 0.90}, true)).To(Equal(1.9999999999999996))
	g.Expect(ulcerScore([]GrowthMultiplier{0.90, 0.80}, true)).To(Equal(2.999999999999999))

	t.Run("GoldenButterfly", func(t *testing.T) {
		dd, _ := leadingDrawdownSequence(GoldenButterfly)
		var sb strings.Builder
		sb.WriteString("Draw-down sequence:\n")
		for _, m := range dd {
			sb.WriteString(fmt.Sprintf("  %17v\n", Percent(m)))
		}
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("Ulcer score: %17v", ulcerScore(dd, true)))
		ExpectMatchesGoldenFile(t, sb.String())
	})
}

func Test_drawdownScores(t *testing.T) {
	g := NewGomegaWithT(t)

	verify := func(returns []Percent, expectedUlcer float64, expectedMaxDrawdown Percent, expectedMaxDuration int) {
		t.Helper()
		ulcer, maxDD, maxDur := drawdownScores(returns)
		g.Expect(ulcer).To(Equal(expectedUlcer), "maxUlcerScore")
		g.Expect(maxDD).To(Equal(expectedMaxDrawdown), "maxDrawdown")
		g.Expect(maxDur).To(Equal(expectedMaxDuration), "maxDuration")
	}

	verify(nil, 0, 0, 0)
	verify([]Percent{}, 0, 0, 0)
	verify(ReadablePercents(-1), 0.20000000000000018, -0.010000000000000009, 1)
	verify(ReadablePercents(-1, 2), 0.10000000000000009, -0.010000000000000009, 1)
	verify(ReadablePercents(-1, 2, -1, -3), 0.9940000000000015, -0.03970000000000007, 2)
	verify(ReadablePercents(-1, 2, -1, -3, 10), 0.4970000000000008, -0.03970000000000007, 2)

	verify(ReadablePercents(-10, 30), 0.9999999999999998, -0.09999999999999998, 1)
	verify(ReadablePercents(-20, 30), 1.9999999999999996, -0.19999999999999996, 1)
	verify(ReadablePercents(-10, -20, 40), 3.799999999999999, -0.2799999999999999, 2)
	verify(ReadablePercents(-10, 30, -10, -20, 30), 8.879999999999995, -0.2799999999999999, 3)

	t.Run("example series", func(t *testing.T) {
		var sb strings.Builder
		reportLine := func(name string, returns []Percent) string {
			ulcer, maxDD, maxDur := drawdownScores(returns)
			return fmt.Sprintf("%16s: %8.4f ulcer, maxDrawDown: %17v, maxDuration: %2d\n", name, ulcer, maxDD, maxDur)
		}
		sb.WriteString(reportLine("TSM", TSM))
		sb.WriteString(reportLine("SCV", SCV))
		sb.WriteString(reportLine("GLD", GLD))
		sb.WriteString(reportLine("LTT", LTT))
		sb.WriteString(reportLine("STT", STT))
		sb.WriteString(reportLine("STB", STB))
		sb.WriteString("\n")
		sb.WriteString(reportLine("GoldenButterfly", GoldenButterfly))

		ExpectMatchesGoldenFile(t, sb.String())
	})
}
