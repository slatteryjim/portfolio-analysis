package portfolio_analysis

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
)

func TestPermutations(t *testing.T) {
	g := NewGomegaWithT(t)
	translatePercentages := func(ps []float64) []float64 {
		res := make([]float64, 0, len(ps))
		prev := 0.0
		for _, p := range ps {
			res = append(res, p-prev)
			prev = p
		}
		return res
	}

	translate := func(perms []permutation) []permutation {
		res := make([]permutation, 0, len(perms))
		for _, p := range perms {
			res = append(res, permutation{
				assets:      p.assets,
				percentages: translatePercentages(p.percentages),
			})
		}
		return res
	}

	dumpAll := func(perms []permutation) {
		for _, p := range perms {
			fmt.Println(p.assets, p.percentages)
		}
	}

	perms := translate(permutations([]string{"A"}, []float64{100}))
	dumpAll(perms)
	g.Expect(perms).To(Equal([]permutation{
		{[]string{"A"}, []float64{100}},
	}))

	perms = translate(permutations([]string{"A", "B"}, []float64{50, 100}))
	g.Expect(perms).To(ConsistOf([]permutation{
		{[]string{"A"}, []float64{100}},
		{[]string{"A", "B"}, []float64{50, 50}},
		{[]string{"B"}, []float64{100}},
	}))

	perms = translate(permutations([]string{"A", "B", "C"}, []float64{33, 66, 100}))
	g.Expect(perms).To(ConsistOf([]permutation{
		{[]string{"A"}, []float64{100}},
		{[]string{"A", "B"}, []float64{66, 34}},
		{[]string{"A", "C"}, []float64{66, 34}},
		{[]string{"A", "B"}, []float64{33, 67}},
		{[]string{"A", "B", "C"}, []float64{33, 33, 34}},
		{[]string{"A", "C"}, []float64{33, 67}},
		{[]string{"B"}, []float64{100}},
		{[]string{"B", "C"}, []float64{66, 34}},
		{[]string{"B", "C"}, []float64{33, 67}},
		{[]string{"C"}, []float64{100}},
	}))

	g.Expect(floats(25, 100, 25)).To(Equal([]float64{25, 50, 75, 100}))

	perms = permutations([]string{"A", "B", "C"}, floats(1, 100, 1))
	g.Expect(len(perms)).To(Equal(5151))

	perms = permutations([]string{"A", "B", "C", "D"}, floats(1, 100, 1))
	g.Expect(len(perms)).To(Equal(176_851))

	// perms = permutations([]string{"A", "B", "C", "D", "E"}, floats(1, 100, 1))
	// g.Expect(len(perms)).To(Equal(4_598_126))

	perms = permutations([]string{"A", "B", "C", "D", "E"}, floats(2.5, 100, 2.5))
	g.Expect(len(perms)).To(Equal(135_751))

	// perms = permutations([]string{"A", "B", "C", "D", "E", "F"}, floats(2.5, 100, 2.5))
	// g.Expect(len(perms)).To(Equal(1_221_759))

	// perms = permutations([]string{"A", "B", "C", "D", "E", "F", "G"}, floats(2.5, 100, 2.5))
	// g.Expect(len(perms)).To(Equal(9_366_819))
}

func floats(start, end, step float64) []float64 {
	var res []float64
	for i := start; i <= end; i += step {
		res = append(res, i)
	}
	return res
}
