package data

import (
	"math"
	"testing"

	. "github.com/onsi/gomega"

	. "github.com/slatteryjim/portfolio-analysis/types"
)

func Test_transpose(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(transpose(nil)).To(BeEmpty())
	g.Expect(transpose([][]string{})).To(BeEmpty())

	g.Expect(transpose([][]string{{"a"}})).To(Equal([][]string{{"a"}}))

	g.Expect(transpose([][]string{
		{"a", "b"},
	})).To(Equal([][]string{
		{"a"},
		{"b"},
	}))

	g.Expect(transpose([][]string{
		{"a", "b"},
		{"1", "2"},
	})).To(Equal([][]string{
		{"a", "1"},
		{"b", "2"},
	}))
}

func TestSeries_AnnualReturnsStartingIn(t *testing.T) {
	g := NewGomegaWithT(t)

	s := Series{
		FirstYear:     2010,
		LastYear:      2013,
		AnnualReturns: []Percent{0, 1, 2, 3},
	}
	g.Expect(s.AnnualReturnsStartingIn(math.MinInt64)).To(Equal([]Percent{0, 1, 2, 3}))
	g.Expect(s.AnnualReturnsStartingIn(0)).To(Equal([]Percent{0, 1, 2, 3}))
	g.Expect(s.AnnualReturnsStartingIn(2009)).To(Equal([]Percent{0, 1, 2, 3}))
	g.Expect(s.AnnualReturnsStartingIn(2010)).To(Equal([]Percent{0, 1, 2, 3}))
	g.Expect(s.AnnualReturnsStartingIn(2011)).To(Equal([]Percent{1, 2, 3}))
	g.Expect(s.AnnualReturnsStartingIn(2012)).To(Equal([]Percent{2, 3}))
	g.Expect(s.AnnualReturnsStartingIn(2013)).To(Equal([]Percent{3}))
	g.Expect(s.AnnualReturnsStartingIn(2014)).To(BeNil())
	g.Expect(s.AnnualReturnsStartingIn(2015)).To(BeNil())
	g.Expect(s.AnnualReturnsStartingIn(math.MaxInt64)).To(BeNil())
}

// Wow, 1 millisecond (1 million nanoseconds)
// Benchmark_parseSimbaTSV-12    	    1117	   1063301 ns/op
func Benchmark_parseSimbaTSV(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := parseSimbaTSV(simbaBacktestingSpreadsheetRev19bTSV)
		if err != nil {
			b.Fatal(err)
		}
	}
}
