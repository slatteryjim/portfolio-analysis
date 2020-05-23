package types

import (
	"math"
)

// PercentSlice attaches the methods of sort.Interface to []PercentSlice, sorting in increasing order
// (not-a-number values are treated as less than other values).
type PercentSlice []Percent

func (p PercentSlice) Len() int {
	return len(p)
}
func (p PercentSlice) Less(i, j int) bool {
	return p[i] < p[j] || math.IsNaN(p[i].Float()) && !math.IsNaN(p[j].Float())
}
func (p PercentSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
