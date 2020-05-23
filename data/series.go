package data

import (
	. "github.com/slatteryjim/portfolio-analysis/types"
)

// Series represents the historical annual returns of an asset.
type Series struct {
	Name          string
	Symbol        string
	FirstYear     int
	LastYear      int
	AnnualReturns []Percent
}

func (s Series) AnnualReturnsStartingIn(year int) []Percent {
	if year < s.FirstYear {
		year = s.FirstYear
	}
	if year > s.LastYear {
		return nil
	}
	return s.AnnualReturns[year-s.FirstYear:]
}
