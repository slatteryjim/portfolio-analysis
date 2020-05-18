package data

import (
	"encoding/csv"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
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

func TestData(t *testing.T) {
	g := NewGomegaWithT(t)
	reader := csv.NewReader(strings.NewReader(simbaBacktestingSpreadsheetRev19bTSV))
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(records).To(HaveLen(151))

	rows := transpose(records)
	g.Expect(rows).To(HaveLen(80))
	for _, s := range rows {
		fmt.Println(s)
	}

	// extract the year numbers from the first row
	var yearNumbers []int
	for i, yearString := range rows[0][2:] {
		yearNumber, err := strconv.Atoi(yearString)
		g.Expect(err).ToNot(HaveOccurred())
		yearNumbers = append(yearNumbers, yearNumber)
		if i > 0 {
			// should always be increasing by 1
			g.Expect(yearNumber).To(Equal(yearNumbers[i-1]+1), "year #"+strconv.Itoa(i+1))
		}
	}
	// first and last year
	g.Expect(yearNumbers[0]).To(Equal(1871))
	g.Expect(yearNumbers[len(yearNumbers)-1]).To(Equal(2019))

	// Parse all of the assets' data series
	var series []Series
	for i, row := range rows[1:] {
		// we'll stop processing on blank rows
		if i == 53 {
			g.Expect(row[0]).To(BeEmpty())
			g.Expect(row[1]).To(BeEmpty())
			break
		}
		var (
			name   = strings.TrimSpace(row[0])
			symbol = strings.TrimSpace(row[1])
		)
		g.Expect(name).ToNot(BeEmpty())
		g.Expect(symbol).ToNot(BeEmpty())
		var (
			annualReturns []float64
			firstYear     *int
			lastYear      int
		)
		for j, returnThisYear := range row[2:] {
			returnThisYear = strings.TrimSpace(returnThisYear)
			if len(returnThisYear) == 0 {
				g.Expect(firstYear).To(BeNil(), fmt.Sprintf("row #%d, year #%d", i+1, j+1))
				continue
			}
			// populate the float value
			returnThisYearFloat, err := strconv.ParseFloat(returnThisYear, 64)
			g.Expect(err).ToNot(HaveOccurred())
			annualReturns = append(annualReturns, returnThisYearFloat)
			// update firstYear/lastYear as appropriate
			thisYear := yearNumbers[j]
			if firstYear == nil {
				firstYear = &thisYear
			}
			lastYear = thisYear
		}
		g.Expect(firstYear).ToNot(BeNil(), fmt.Sprintf("row #%d", i+1))
		g.Expect(lastYear).To(Equal(2019), "all have data up to the same year")
		g.Expect(len(annualReturns)).To(BeNumerically(">=", 35), "should have this many years of data for all assets")
		series = append(series, Series{
			Name:          name,
			Symbol:        symbol,
			FirstYear:     *firstYear,
			LastYear:      lastYear,
			AnnualReturns: annualReturns,
		})
	}

	// build a lookup map of series by name
	seriesByName := map[string]Series{}
	for _, s := range series {
		seriesByName[s.Name] = s
	}
	g.Expect(seriesByName).To(HaveLen(len(series)), "no duplicate IDs")
}

func TestSeries_AnnualReturnsStartingIn(t *testing.T) {
	g := NewGomegaWithT(t)

	s := Series{
		FirstYear:     2010,
		LastYear:      2013,
		AnnualReturns: []float64{0, 1, 2, 3},
	}
	g.Expect(s.AnnualReturnsStartingIn(math.MinInt64)).To(Equal([]float64{0, 1, 2, 3}))
	g.Expect(s.AnnualReturnsStartingIn(0)).To(Equal([]float64{0, 1, 2, 3}))
	g.Expect(s.AnnualReturnsStartingIn(2009)).To(Equal([]float64{0, 1, 2, 3}))
	g.Expect(s.AnnualReturnsStartingIn(2010)).To(Equal([]float64{0, 1, 2, 3}))
	g.Expect(s.AnnualReturnsStartingIn(2011)).To(Equal([]float64{1, 2, 3}))
	g.Expect(s.AnnualReturnsStartingIn(2012)).To(Equal([]float64{2, 3}))
	g.Expect(s.AnnualReturnsStartingIn(2013)).To(Equal([]float64{3}))
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
