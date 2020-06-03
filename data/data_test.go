package data

import (
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

func TestPortfolioReturnsList(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(PortfolioReturnsList()).To(BeEmpty())

	// one series, just returns itself
	g.Expect(PortfolioReturnsList("TSM")).To(Equal([][]Percent{
		MustFind("TSM").AnnualReturns,
	}))

	// two series, LCV has less data
	g.Expect(PortfolioReturnsList("TSM", "LCV")).To(Equal([][]Percent{
		MustFind("TSM").AnnualReturnsStartingIn(1927),
		MustFind("LCV").AnnualReturnsStartingIn(1927),
	}))

	// three series, Gold has the least data
	g.Expect(PortfolioReturnsList("TSM", "Gold", "LCV")).To(Equal([][]Percent{
		MustFind("TSM").AnnualReturnsStartingIn(1969),
		MustFind("Gold").AnnualReturnsStartingIn(1969),
		MustFind("LCV").AnnualReturnsStartingIn(1969),
	}))
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
