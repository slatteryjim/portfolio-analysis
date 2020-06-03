package data

import (
	"math"
	"testing"

	. "github.com/onsi/gomega"

	. "github.com/slatteryjim/portfolio-analysis/types"
)

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
