package types

// GrowthMultiplier is a float focused around 1.00. So a 5% return would be represented
// as a 1.05 GrowthMultiplier.
type GrowthMultiplier float64

func (g GrowthMultiplier) Float() float64 {
	return float64(g)
}
