package data

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
)

type (
	Series struct {
		Name          string
		Symbol        string
		FirstYear     int
		LastYear      int
		AnnualReturns []float64
	}
)

func MustFind(name string) Series {
	s, ok := _seriesByName[name]
	if !ok {
		panic(fmt.Sprintf("Did not find series with name %q", name))
	}
	return s
}

// _seriesByName has all of the Series organized by asset name
var _seriesByName map[string]Series

func init() {
	var err error
	_seriesByName, err = parseSimbaTSV(simbaBacktestingSpreadsheetRev19bTSV)
	if err != nil {
		panic(err.Error())
	}
}

// parseSimbaTSV parses TSV content from the Simba Backtesting Spreadsheet,
// and returns a map of the series identified by name.
func parseSimbaTSV(tsv string) (map[string]Series, error) {
	reader := csv.NewReader(strings.NewReader(tsv))
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) != 151 {
		return nil, fmt.Errorf("unexpected records length)")
	}

	rows := transpose(records)
	if len(rows) != 80 {
		return nil, fmt.Errorf("unexpected records length")
	}
	// for _, s := range rows {
	// 	fmt.Println(s)
	// }

	// extract the year numbers from the first row
	var yearNumbers []int
	for i, yearString := range rows[0][2:] {
		yearNumber, err := strconv.Atoi(yearString)
		if err != nil {
			return nil, err
		}
		yearNumbers = append(yearNumbers, yearNumber)
		if i > 0 {
			// year should always be increasing by 1
			if yearNumber != yearNumbers[i-1]+1 {
				return nil, fmt.Errorf("unexpected year #%d: should always be increasing by 1", i+1)
			}
		}
	}
	// first and last year
	if firstYear := yearNumbers[0]; firstYear != 1871 {
		return nil, fmt.Errorf("unexpected first year: %d", firstYear)
	}
	if lastYear := yearNumbers[len(yearNumbers)-1]; lastYear != 2019 {
		return nil, fmt.Errorf("unexpected last year: %d", lastYear)
	}

	// Parse all of the assets' data series
	var series []Series
	for i, row := range rows[1:] {
		// we'll stop processing on blank rows
		if i == 53 {
			if row[0] != "" || row[1] != "" {
				return nil, fmt.Errorf("expected empty row #%d", i+1)
			}
			break
		}
		var (
			name   = strings.TrimSpace(row[0])
			symbol = strings.TrimSpace(row[1])
		)
		if name == "" || symbol == "" {
			return nil, fmt.Errorf("name or symbol should not be empty in row #%d", i+1)
		}
		var (
			annualReturns []float64
			firstYear     *int
			lastYear      int
		)
		for j, returnThisYear := range row[2:] {
			returnThisYear = strings.TrimSpace(returnThisYear)
			if returnThisYear == "" {
				if firstYear != nil {
					return nil, fmt.Errorf("row #%d, year #%d: first year should be nil", i+1, j+1)
				}
				continue
			}
			// populate the float value
			returnThisYearFloat, err := strconv.ParseFloat(returnThisYear, 64)
			if err != nil {
				return nil, err
			}
			annualReturns = append(annualReturns, returnThisYearFloat)
			// update firstYear/lastYear as appropriate
			thisYear := yearNumbers[j]
			if firstYear == nil {
				firstYear = &thisYear
			}
			lastYear = thisYear
		}
		if firstYear == nil {
			return nil, fmt.Errorf("row #%d: firstYear should not be nil", i+1)
		}
		if lastYear != 2019 {
			return nil, fmt.Errorf("row #%d: all rows should have data up to the same year", i+1)
		}
		if len(annualReturns) < 35 {
			return nil, fmt.Errorf("row #%d: expected asset to have a minimum number of years", i+1)
		}
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
	if len(seriesByName) != len(series) {
		return nil, fmt.Errorf("shouldn't have any duplicate IDs")
	}
	return seriesByName, nil
}

func (s Series) AnnualReturnsStartingIn(year int) []float64 {
	if year < s.FirstYear {
		year = s.FirstYear
	}
	if year > s.LastYear {
		return nil
	}
	return s.AnnualReturns[year-s.FirstYear:]
}

// transpose returns a transposed version of the two-dimensional slice.
func transpose(data [][]string) [][]string {
	if len(data) == 0 {
		return nil
	}
	var (
		height = len(data)
		width  = len(data[0])
	)
	result := make([][]string, width)
	for x := 0; x < width; x++ {
		column := make([]string, height)
		for y := 0; y < height; y++ {
			column[y] = data[y][x]
		}
		result[x] = column
	}
	return result
}
