package data

import (
	"encoding/csv"
	"fmt"
	"math"
	"strconv"
	"strings"

	. "github.com/slatteryjim/portfolio-analysis/types"
)

// MustFind returns the Series for the given asset name, or panics if it is not found.
func MustFind(name string) Series {
	s, ok := _seriesByName[name]
	if !ok {
		panic(fmt.Sprintf("Did not find series with name %q", name))
	}
	return s
}

// Assets returns a list of all the available asset names.
func Names() []string {
	res := make([]string, 0, len(_seriesByName))
	for k := range _seriesByName {
		res = append(res, k)
	}
	return res
}

// PortfolioReturnsList returns a list of returns for the given assets, for the years that they overlap.
func PortfolioReturnsList(assetNames ...string) [][]Percent {
	var (
		series       = make([]Series, len(assetNames))
		maxFirstYear = math.MinInt64
		minLastYear  = math.MaxInt64
	)
	for i := 0; i < len(assetNames); i++ {
		s := MustFind(assetNames[i])
		series[i] = s
		if s.FirstYear > maxFirstYear {
			maxFirstYear = s.FirstYear
		}
		if s.LastYear < minLastYear {
			minLastYear = s.LastYear
		}
	}
	res := make([][]Percent, len(assetNames))
	{
		years := minLastYear - maxFirstYear
		for i := 0; i < len(assetNames); i++ {
			s := series[i]
			index := s.IndexOfYear(maxFirstYear)
			res[i] = s.AnnualReturns[index : index+years+1]
		}
	}
	return res
}

// _seriesByName has all of the Series organized by asset name
var _seriesByName map[string]Series

func init() {
	var err error
	_seriesByName, err = parseSimbaTSV(simbaBacktestingSpreadsheetRev20aTSV)
	if err != nil {
		panic(err.Error())
	}
}

// parseSimbaTSV parses TSV content from the Simba Backtesting Spreadsheet,
// and returns a map of the series identified by name.
func parseSimbaTSV(tsv string) (map[string]Series, error) {
	const (
		expectedLastYear = 2020
	)
	normalizeNames := map[string]string{
		"TSM (US)": "TSM",
		"TBM (US)": "TBM",
	}

	reader := csv.NewReader(strings.NewReader(tsv))
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) != 152 {
		return nil, fmt.Errorf("unexpected records length: %d", len(records))
	}

	rows := transpose(records)
	if len(rows) != 78 {
		return nil, fmt.Errorf("unexpected rows length: %d", len(rows))
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
	if lastYear := yearNumbers[len(yearNumbers)-1]; lastYear != expectedLastYear {
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
		if newName, ok := normalizeNames[name]; ok {
			name = newName
		}
		var (
			annualReturns []Percent
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
			annualReturns = append(annualReturns, ReadablePercent(returnThisYearFloat))
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
		if lastYear != expectedLastYear {
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
