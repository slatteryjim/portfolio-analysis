package types

import (
	"strconv"
	"strings"
)

// Percent is a percentage using the range 0.00 - 1.00.
type Percent float64

func (p Percent) String() string {
	return formatFloat(p.Float()*100, 12) + "%"
}

// GrowthMultiplier is a focused around 1.00. So a 5% return would be represented
// as a 1.05 GrowthMultiplier.
func (p Percent) GrowthMultiplier() GrowthMultiplier {
	return GrowthMultiplier(p + 1)
}

func (p Percent) Float() float64 {
	return float64(p)
}

// ReadablePercents takes easy-to-read percentages using the range 0 - 100, and returns
// a slice of Percent (each using the range 0.00 - 1.00).
func ReadablePercents(xs ...float64) []Percent {
	res := make([]Percent, len(xs))
	for i, x := range xs {
		res[i] = ReadablePercent(x)
	}
	return res
}

// ReadablePercent takes an easy-to-read percentage using the range 0 - 100, and returns
// a Percent (using the range 0.00 - 1.00).
func ReadablePercent(x float64) Percent {
	return Percent(x / 100)
}

func formatFloat(f float64, prec int) string {
	s := strconv.FormatFloat(f, 'f', prec, 64)
	s = strings.TrimRight(s, "0")
	return strings.TrimRight(s, ".")
}

func Floats(xs ...Percent) []float64 {
	res := make([]float64, len(xs))
	for i, x := range xs {
		res[i] = x.Float()
	}
	return res
}
