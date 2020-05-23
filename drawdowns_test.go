package portfolio_analysis

import (
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

	dd, _ := leadingDrawdownSequence(GoldenButterfly)
	g.Expect(dd).To(Equal([]GrowthMultiplier{0.8466616731717865, 0.8613248299786367, 0.9480466828875849}))
	g.Expect(ulcerScore(dd, true)).To(Equal(3.4396681396199194))
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

	verify(TSM, 26.990140749914836, -0.5225438230658536, 13)
	verify(SCV, 27.53985971540292, -0.6102169347344113, 10)
	verify(GLD, 393.21757430117924, -0.783196539516205, 40)
	verify(LTT, 28.38445160979715, -0.45411601643765687, 13)
	verify(STT, 11.51644603701871, -0.19005483707691806, 11)
	verify(STB, 10.082153041903915, -0.2005766291394362, 10)
	verify(GoldenButterfly, 3.4396681396199194, -0.15333832682821347, 3)
}
