package portfolio_analysis

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_leadingDrawdownSequence(t *testing.T) {
	g := NewGomegaWithT(t)

	verify := func(returns, expectedSequence []float64, expectedEnded bool) {
		t.Helper()
		sequence, ended := leadingDrawdownSequence(returns)
		g.Expect(sequence).To(Equal(expectedSequence), "sequence")
		g.Expect(ended).To(Equal(expectedEnded), "ended")
	}

	// empty
	verify(nil, []float64{}, false)
	verify([]float64{}, []float64{}, false)

	// doesn't start with a drawdown
	verify([]float64{1}, []float64{}, true)
	verify([]float64{0, -1, 2}, []float64{}, true)

	// starts with a drawdown
	verify([]float64{-1}, []float64{0.99}, false)
	verify([]float64{-1, 2}, []float64{0.99}, true)
	verify([]float64{-1, -1, 3}, []float64{0.99, 0.9801}, true)
	verify([]float64{-1, -1, 3, -5, -50}, []float64{0.99, 0.9801}, true)

	verify([]float64{-50, 100}, []float64{0.50}, true)

}

func Test_drawdowns(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(drawdowns(nil)).To(BeEmpty())
	g.Expect(drawdowns([]float64{0})).To(BeEmpty())
	g.Expect(drawdowns([]float64{1})).To(BeEmpty())
	g.Expect(drawdowns([]float64{-1})).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []float64{0.99}, recovered: false},
	}))
	g.Expect(drawdowns([]float64{-1, 2})).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []float64{0.99}, recovered: true},
	}))
	g.Expect(drawdowns([]float64{-1, -1, 3})).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []float64{0.99, 0.9801}, recovered: true},
		{startIndex: 1, cumulativeReturns: []float64{0.99}, recovered: true},
	}))
	g.Expect(drawdowns([]float64{-1, -1, 1})).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []float64{0.99, 0.9801, 0.989901}, recovered: false},
		{startIndex: 1, cumulativeReturns: []float64{0.99, 0.9999}, recovered: false},
	}))
	g.Expect(drawdowns([]float64{-1, 3, -1, -1, 3})).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []float64{0.99}, recovered: true},
		{startIndex: 2, cumulativeReturns: []float64{0.99, 0.9801}, recovered: true},
		{startIndex: 3, cumulativeReturns: []float64{0.99}, recovered: true},
	}))
	g.Expect(drawdowns([]float64{-1, -1, 3, -5, -50, 100})).To(Equal([]drawdownSequence{
		{startIndex: 0, cumulativeReturns: []float64{0.99, 0.9801}, recovered: true},
		{startIndex: 1, cumulativeReturns: []float64{0.99}, recovered: true},
		{startIndex: 3, cumulativeReturns: []float64{0.95, 0.475, 0.95}, recovered: false},
		{startIndex: 4, cumulativeReturns: []float64{0.50}, recovered: true},
	}))
}

func Test_ulcerScore(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(ulcerScore(nil, false)).To(Equal(0.0))
	g.Expect(ulcerScore(nil, true)).To(Equal(0.0))
	g.Expect(ulcerScore([]float64{0.99}, true)).To(Equal(0.10000000000000009))
	g.Expect(ulcerScore([]float64{0.99}, false)).To(Equal(0.20000000000000018))
	g.Expect(ulcerScore([]float64{0.90}, true)).To(Equal(0.9999999999999998))
	g.Expect(ulcerScore([]float64{0.90, 0.90}, true)).To(Equal(1.9999999999999996))
	g.Expect(ulcerScore([]float64{0.90, 0.80}, true)).To(Equal(2.999999999999999))

	dd, _ := leadingDrawdownSequence(GoldenButterfly)
	g.Expect(dd).To(Equal([]float64{0.84666, 0.8613241511999999, 0.9480594932258399}))
	g.Expect(ulcerScore(dd, true)).To(Equal(3.4395635557416018))
}

func Test_drawdownScores(t *testing.T) {
	g := NewGomegaWithT(t)

	verify := func(returns []float64, expectedUlcer, expectedMaxDrawdown float64, expectedMaxDuration int) {
		ulcer, maxDD, maxDur := drawdownScores(returns)
		g.Expect(ulcer).To(Equal(expectedUlcer), "maxUlcerScore")
		g.Expect(maxDD).To(Equal(expectedMaxDrawdown), "maxDrawdown")
		g.Expect(maxDur).To(Equal(expectedMaxDuration), "maxDuration")
	}

	verify(nil, 0, 0, 0)
	verify([]float64{}, 0, 0, 0)
	verify([]float64{-1}, 0.20000000000000018, -0.010000000000000009, 1)
	verify([]float64{-1, 2}, 0.10000000000000009, -0.010000000000000009, 1)
	verify([]float64{-1, 2, -1, -3}, 0.9940000000000015, -0.03970000000000007, 2)
	verify([]float64{-1, 2, -1, -3, 10}, 0.4970000000000008, -0.03970000000000007, 2)

	verify([]float64{-10, 30}, 0.9999999999999998, -0.09999999999999998, 1)
	verify([]float64{-20, 30}, 1.9999999999999996, -0.19999999999999996, 1)
	verify([]float64{-10, -20, 40}, 3.799999999999999, -0.2799999999999999, 2)
	verify([]float64{-10, 30, -10, -20, 30}, 8.879999999999995, -0.2799999999999999, 3)

	verify(TSM, 26.98911264363915, -0.5225460399999999, 13)
	verify(SCV, 27.539774130167125, -0.6102168140239335, 10)
	verify(GLD, 393.20116848234824, -0.7831742646171225, 40)
	verify(LTT, 28.3941804140844, -0.4541977153865162, 13)
	verify(STT, 11.543618482934454, -0.19004565800151896, 11)
	verify(STB, 10.091173245580162, -0.20070470834693865, 10)
	verify(GoldenButterfly, 3.4395635557416018, -0.15334000000000003, 3)
}
