package portfolio_analysis

import (
	"compress/gzip"
	"database/sql"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	. "github.com/onsi/gomega"

	"github.com/slatteryjim/portfolio-analysis/data"
	. "github.com/slatteryjim/portfolio-analysis/types"
)

func TestCombinations(t *testing.T) {
	g := NewGomegaWithT(t)

	dumpAll := func(perms []Combination) {
		for _, p := range perms {
			fmt.Println(p.Assets, p.Percentages)
		}
	}

	perms := Combinations([]string{"A"}, ReadablePercents(100))
	dumpAll(perms)
	g.Expect(perms).To(Equal([]Combination{
		{[]string{"A"}, ReadablePercents(100)},
	}))

	perms = Combinations([]string{"A", "B"}, ReadablePercents(50, 100))
	g.Expect(perms).To(ConsistOf([]Combination{
		{[]string{"A"}, ReadablePercents(100)},
		{[]string{"A", "B"}, ReadablePercents(50, 50)},
		{[]string{"B"}, ReadablePercents(100)},
	}))

	perms = Combinations([]string{"A", "B", "C"}, ReadablePercents(33, 66, 100))
	g.Expect(perms).To(ConsistOf([]Combination{
		{[]string{"A"}, []Percent{1.00}},
		{[]string{"A", "B"}, []Percent{0.66, 0.33999999999999997}},
		{[]string{"A", "C"}, []Percent{0.66, 0.33999999999999997}},
		{[]string{"A", "B"}, []Percent{0.33, 0.6699999999999999}},
		{[]string{"A", "B", "C"}, []Percent{0.33, 0.33, 0.33999999999999997}},
		{[]string{"A", "C"}, []Percent{0.33, 0.6699999999999999}},
		{[]string{"B"}, []Percent{1.00}},
		{[]string{"B", "C"}, []Percent{0.66, 0.33999999999999997}},
		{[]string{"B", "C"}, []Percent{0.33, 0.6699999999999999}},
		{[]string{"C"}, []Percent{1.00}},
	}))

	perms = Combinations([]string{"A", "B", "C"}, ReadablePercents(series(1, 100, 1)...))
	g.Expect(len(perms)).To(Equal(5151))

	perms = Combinations([]string{"A", "B", "C", "D"}, ReadablePercents(series(1, 100, 1)...))
	g.Expect(len(perms)).To(Equal(176_851))

	// perms = Combinations([]string{"A", "B", "C", "D", "E"}, floats(1, 100, 1))
	// g.Expect(len(perms)).To(Equal(4_598_126))

	perms = Combinations([]string{"A", "B", "C", "D", "E"}, ReadablePercents(series(2.5, 100, 2.5)...))
	g.Expect(len(perms)).To(Equal(135_751))

	// perms = Combinations([]string{"A", "B", "C", "D", "E", "F"}, floats(2.5, 100, 2.5))
	// g.Expect(len(perms)).To(Equal(1_221_759))

	// perms = Combinations([]string{"A", "B", "C", "D", "E", "F", "G"}, floats(2.5, 100, 2.5))
	// g.Expect(len(perms)).To(Equal(9_366_819))
}

func Test_translatePercentages(t *testing.T) {
	g := NewGomegaWithT(t)

	verify := func(ps, expected []Percent) {
		t.Helper()
		translatePercentages(ps)
		g.Expect(ps).To(Equal(expected))
	}

	verify(nil, nil)
	verify([]Percent{}, []Percent{})

	verify([]Percent{25}, []Percent{25})

	verify([]Percent{25, 100}, []Percent{25, 75})

	verify([]Percent{25, 50, 75, 100}, []Percent{25, 25, 25, 25})
}

func TestPortfolioCombinations_GoldenButterflyAssets(t *testing.T) {
	g := NewGomegaWithT(t)

	// GoldenButterfly advertised on: https://portfoliocharts.com/portfolio/golden-butterfly/
	// GoldenButterfly: [TSM SCV LTT STT GLD] [20% 20% 20% 20% 20%] (64) RF:0.00 AvgReturn:5.669%(5299) BLT:5.241%(2450) BST:2.849%(927) PWR:4.224%(1853) SWR:5.305%(1699) StdDev:8.103%(2383) Ulcer:3.4(2258) DeepestDrawdown:-15.33%(1862) LongestDrawdown:3(2), StartDateSensitivity:7.71%(756)
	//
	// Check out the results using 1% increments:
	// Best PWR30: [TSM SCV GLD] [1% 66% 33%] (2042098) RF:0.00 AvgReturn:7.932%(64438) BLT:5.536%(704794) BST:2.885%(431554) PWR:5.450%(1) SWR:6.284%(38) StdDev:13.397%(4192102) Ulcer:8.2(3097695) DeepestDrawdown:-26.73%(3323261) LongestDrawdown:6(5), StartDateSensitivity:16.97%(3381664)
	// Best UlcerScore: [TSM LTT STT GLD] [8% 3% 80% 9%] (3209303) RF:0.00 AvgReturn:2.738%(4589590) BLT:1.948%(4292822) BST:0.194%(3917744) PWR:2.456%(4228073) SWR:4.504%(2983311) StdDev:3.976%(1530) Ulcer:0.6(1) DeepestDrawdown:-5.43%(798) LongestDrawdown:4(3), StartDateSensitivity:9.10%(803328)
	//
	// Using 5% increments:
	// Best PWR30: [SCV GLD] [70% 30%] (4285) RF:0.00 AvgReturn:8.068%(178) BLT:5.896%(811) BST:2.386%(2517) PWR:5.364%(1) SWR:6.148%(8) StdDev:13.708%(9495) Ulcer:9.5(6975) DeepestDrawdown:-27.10%(7246) LongestDrawdown:6(5), StartDateSensitivity:16.20%(6723)
	// Best UlcerScore: [TSM STT GLD] [10% 80% 10%] (6678) RF:0.00 AvgReturn:2.808%(10564) BLT:2.070%(9640) BST:0.386%(8245) PWR:2.477%(9484) SWR:4.581%(5745) StdDev:3.928%(7) Ulcer:0.6(1) DeepestDrawdown:-5.60%(4) LongestDrawdown:2(1), StartDateSensitivity:8.44%(1191)
	//
	// Timing/log for GoldenButterfly assets, 1% step combinations:
	//   Generated 4598126 combinations in 6.551823599s
	//   ...Evaluating 4598126 combinations.
	//   Done evaluating portfolios in 53.007350212s or 86745 portfolios/second
	//   ...Calculate rank scores for the portfolios
	//   ...rank by all their ranks (equally weighted)
	//   Ranked portfolios in 1m8.660651682s
	startAt := time.Now()
	perms := Combinations([]string{"TSM", "SCV", "LTT", "STT", "GLD"}, ReadablePercents(seriesRange(5)...))
	// g.Expect(len(perms)).To(Equal(10_626)) // only 3,876 include all five.
	fmt.Println("Generated", len(perms), "combinations in", time.Since(startAt))

	// filter to only include combinations where all 5 assets are used/
	// (See: https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating)
	// {
	// 	startAt := time.Now()
	// 	numberOfAssets := 5
	// 	filtered := perms[:0]
	// 	for _, p := range perms {
	// 		// this cuts 10,626 combinations down to 3,876
	// 		if len(p.Assets) == numberOfAssets {
	// 			filtered = append(filtered, p)
	// 		}
	// 	}
	// 	for i := len(filtered); i < len(perms); i++ {
	// 		perms[i] = Combination{}
	// 	}
	// 	fmt.Printf("...culled down to %0.1f%% combinations in %s\n", float64(len(filtered))/float64(len(perms))*100, time.Since(startAt))
	// 	perms = filtered
	// }
	//g.Expect(len(perms)).To(Equal(3_876))
	startAt = time.Now()
	fmt.Println("...Evaluating", len(perms), "combinations.")

	results, err := EvaluatePortfolios(perms, assetMap)
	g.Expect(err).ToNot(HaveOccurred())
	elapsed := time.Since(startAt)
	fmt.Println("Done evaluating portfolios in", elapsed, "or", int(float64(len(results))/elapsed.Seconds()), "portfolios/second")

	startAt = time.Now()
	RankPortfoliosInPlace(results)
	fmt.Println("Ranked portfolios in", time.Since(startAt))

	// print best:
	fmt.Println("Best combined overall ranks:")
	fmt.Println("#1:", results[0])
	fmt.Println("#2:", results[1])
	fmt.Println("#3:", results[2])

	PrintBestByEachRanking(results)

	gbStat := FindOne(results, func(p *PortfolioStat) bool {
		if len(p.Percentages) != 5 {
			return false
		}
		for _, pct := range p.Percentages {
			if !approxEqual(pct.Float(), 0.20, 0.001) {
				return false
			}
		}
		return true
	})
	g.Expect(gbStat).ToNot(BeNil())
	findBetterThanGoldenButterfly(gbStat, results)
}

func TestPortfolioCombinations_AnythingBetterThanGoldenButterfly(t *testing.T) {
	// g := NewGomegaWithT(t)

	// need an n-choose-r algorithm
	// we'll just do an "n-choose-1" for the moment
	var results []*PortfolioStat
	for _, n := range data.Names() {
		p := Combination{
			Assets:      []string{n},
			Percentages: ReadablePercents(100),
		}
		stat := evaluatePortfolio(data.MustFind(n).AnnualReturns, p)
		results = append(results, stat)
	}
	RankPortfoliosInPlace(results)

	// print best:
	fmt.Println("Best combined overall ranks:")
	for i := 0; i < 10; i++ {
		fmt.Printf("#%d: %s\n", i+1, results[i])
	}

	PrintBestByEachRanking(results)
}

// Wow, found 4 thousand portfolios that were better than GoldenButterfly!
// Sample output:
//
//   === RUN   TestAllKAssetPortfolios
//   ...Calculate rank scores for the portfolios
//   ...rank by all their ranks (equally weighted)
//
//   Best by each ranking:
//   Best AvgReturn: [Momentum Factor Quality Factor Health Care Emerging LT STRIPS] [20% 20% 20% 20% 20%] (470027) RF:0.00 AvgReturn:12.624%(1) BLT:8.229%(518) BST:2.633%(531923) PWR:10.237%(1) SWR:10.804%(2) StdDev:13.804%(2426953) Ulcer:2.6(749516) DeepestDrawdown:-17.69%(686945) LongestDrawdown:2(2), StartDateSensitivity:16.82%(2278676)
//   Best BaselineLTReturn: [Energy Health Care Emerging LT STRIPS REIT] [20% 20% 20% 20% 20%] (91517) RF:0.00 AvgReturn:11.488%(3207) BLT:9.368%(1) BST:4.133%(24907) PWR:9.177%(1499) SWR:9.824%(1738) StdDev:12.851%(2113132) Ulcer:2.0(485828) DeepestDrawdown:-19.18%(833783) LongestDrawdown:2(2), StartDateSensitivity:7.82%(512200)
//   Best BaselineSTReturn: [Momentum Factor Precious Metals Dividend Growth Health Care SCV] [20% 20% 20% 20% 20%] (385089) RF:0.00 AvgReturn:10.822%(24705) BLT:7.348%(10930) BST:6.820%(1) PWR:9.022%(2721) SWR:9.685%(3068) StdDev:11.945%(1750769) Ulcer:4.0(1471427) DeepestDrawdown:-27.75%(1819033) LongestDrawdown:2(2), StartDateSensitivity:12.92%(1617187)
//   Best PWR30: [Momentum Factor Quality Factor Health Care Emerging LT STRIPS] [20% 20% 20% 20% 20%] (470027) RF:0.00 AvgReturn:12.624%(1) BLT:8.229%(518) BST:2.633%(531923) PWR:10.237%(1) SWR:10.804%(2) StdDev:13.804%(2426953) Ulcer:2.6(749516) DeepestDrawdown:-17.69%(686945) LongestDrawdown:2(2), StartDateSensitivity:16.82%(2278676)
//   Best SWR30: [Momentum Factor Quality Factor Dividend Growth Health Care Emerging] [20% 20% 20% 20% 20%] (1864275) RF:0.00 AvgReturn:12.453%(11) BLT:6.824%(42334) BST:1.029%(1749316) PWR:10.201%(2) SWR:10.828%(1) StdDev:15.238%(2692925) Ulcer:5.9(2117092) DeepestDrawdown:-34.06%(2434559) LongestDrawdown:4(4), StartDateSensitivity:22.85%(2771259)
//   Best StdDev: [Int'l Bd T-Bill STT ST Munis ST Invest. Grade] [20% 20% 20% 20% 20%] (2435071) RF:0.00 AvgReturn:2.079%(2869592) BLT:0.564%(2865458) BST:-0.108%(2366870) PWR:1.756%(2865253) SWR:4.501%(2813937) StdDev:2.588%(1) Ulcer:0.3(1345) DeepestDrawdown:-2.38%(289) LongestDrawdown:4(4), StartDateSensitivity:2.67%(2575)
//   Best UlcerScore: [Health Care Gold STB STT ST Invest. Grade] [20% 20% 20% 20% 20%] (1282952) RF:0.00 AvgReturn:5.068%(2724456) BLT:4.025%(2024102) BST:2.437%(672868) PWR:4.370%(2281119) SWR:5.972%(2150791) StdDev:4.602%(3529) Ulcer:0.1(1) DeepestDrawdown:-1.28%(2) LongestDrawdown:2(2), StartDateSensitivity:3.03%(6221)
//   Best DeepestDrawdown: [Int'l Bd Health Care T-Bill Gold STT] [20% 20% 20% 20% 20%] (1451846) RF:0.00 AvgReturn:4.792%(2773942) BLT:3.875%(2169737) BST:2.705%(482818) PWR:4.148%(2412086) SWR:5.802%(2291009) StdDev:4.184%(1314) Ulcer:0.1(3) DeepestDrawdown:-1.28%(1) LongestDrawdown:2(2), StartDateSensitivity:3.14%(7592)
//   Best LongestDrawdown: [Precious Metals Dividend Growth Health Care Wellesley LTT] [20% 20% 20% 20% 20%] (1) RF:0.00 AvgReturn:8.967%(468145) BLT:6.483%(90317) BST:4.851%(2565) PWR:7.533%(115913) SWR:8.442%(107771) StdDev:8.941%(518423) Ulcer:1.0(131713) DeepestDrawdown:-10.33%(177206) LongestDrawdown:1(1), StartDateSensitivity:5.40%(132427)
//   Best StartDateSensitivity: [Global Bd Int'l Bd LT Munis Gold Windsor] [20% 20% 20% 20% 20%] (1868143) RF:0.00 AvgReturn:4.713%(2784986) BLT:3.649%(2362688) BST:1.761%(1206746) PWR:3.596%(2665201) SWR:5.264%(2599384) StdDev:5.596%(23466) Ulcer:0.7(38336) DeepestDrawdown:-6.54%(43125) LongestDrawdown:1(1), StartDateSensitivity:1.96%(1)
//
//   Best combined overall ranks:
//   #1: [Precious Metals Dividend Growth Health Care Wellesley LTT] [20% 20% 20% 20% 20%] (1) RF:0.00 AvgReturn:8.967%(468145) BLT:6.483%(90317) BST:4.851%(2565) PWR:7.533%(115913) SWR:8.442%(107771) StdDev:8.941%(518423) Ulcer:1.0(131713) DeepestDrawdown:-10.33%(177206) LongestDrawdown:1(1), StartDateSensitivity:5.40%(132427)
//   #2: [Health Care Wellesley Emerging LTT ITT] [20% 20% 20% 20% 20%] (2) RF:0.00 AvgReturn:8.988%(458121) BLT:6.368%(114617) BST:4.465%(9337) PWR:7.782%(72347) SWR:8.787%(49126) StdDev:9.098%(563633) Ulcer:0.9(84118) DeepestDrawdown:-8.55%(105344) LongestDrawdown:2(2), StartDateSensitivity:5.27%(117162)
//   #3: [Momentum Factor Precious Metals Health Care LTT ITT] [20% 20% 20% 20% 20%] (3) RF:0.00 AvgReturn:8.797%(552752) BLT:6.335%(122335) BST:4.313%(14720) PWR:7.550%(112471) SWR:8.467%(102123) StdDev:8.673%(445944) Ulcer:0.8(78139) DeepestDrawdown:-8.33%(96843) LongestDrawdown:1(1), StartDateSensitivity:6.16%(229073)
//   #4: [Int'l Bd Health Care Emerging STB LT STRIPS] [20% 20% 20% 20% 20%] (4) RF:0.00 AvgReturn:9.059%(425539) BLT:6.443%(98233) BST:4.618%(5761) PWR:7.656%(92220) SWR:8.658%(66821) StdDev:9.280%(619976) Ulcer:0.7(52665) DeepestDrawdown:-7.27%(61420) LongestDrawdown:2(2), StartDateSensitivity:4.83%(75743)
//   #5: [Global Bd Int'l Bd Health Care Emerging LT STRIPS] [20% 20% 20% 20% 20%] (5) RF:0.00 AvgReturn:9.278%(334699) BLT:6.707%(55454) BST:5.029%(1380) PWR:7.797%(70129) SWR:8.750%(53652) StdDev:9.463%(681863) Ulcer:0.8(64981) DeepestDrawdown:-7.82%(78463) LongestDrawdown:2(2), StartDateSensitivity:5.02%(91945)
//   #6: [Int'l Bd Health Care Wellesley Emerging LTT] [20% 20% 20% 20% 20%] (6) RF:0.00 AvgReturn:8.902%(499330) BLT:6.296%(132094) BST:4.354%(13036) PWR:7.715%(82387) SWR:8.711%(59078) StdDev:8.917%(511932) Ulcer:1.0(132392) DeepestDrawdown:-10.36%(178269) LongestDrawdown:1(1), StartDateSensitivity:5.97%(203443)
//   #7: [Global Bd Health Care Wellesley Emerging LTT] [20% 20% 20% 20% 20%] (7) RF:0.00 AvgReturn:8.951%(475608) BLT:6.322%(125452) BST:4.218%(19421) PWR:7.751%(76847) SWR:8.749%(53758) StdDev:9.043%(547475) Ulcer:1.0(132995) DeepestDrawdown:-10.38%(179180) LongestDrawdown:1(1), StartDateSensitivity:5.86%(188741)
//   #8: [Int'l Bd Health Care Emerging LT STRIPS ST Invest. Grade] [20% 20% 20% 20% 20%] (8) RF:0.00 AvgReturn:9.098%(408101) BLT:6.505%(86322) BST:4.638%(5395) PWR:7.671%(89680) SWR:8.662%(66124) StdDev:9.358%(645824) Ulcer:0.7(49771) DeepestDrawdown:-7.13%(57634) LongestDrawdown:2(2), StartDateSensitivity:4.86%(78222)
//   #9: [Precious Metals Quality Factor Health Care LT STRIPS STT] [20% 20% 20% 20% 20%] (9) RF:0.00 AvgReturn:9.123%(397034) BLT:6.451%(96499) BST:4.166%(22693) PWR:7.518%(119173) SWR:8.423%(112328) StdDev:9.413%(664308) Ulcer:0.6(36903) DeepestDrawdown:-6.47%(41406) LongestDrawdown:1(1), StartDateSensitivity:5.38%(130165)
//   #10: [Precious Metals Quality Factor Health Care STB LT STRIPS] [20% 20% 20% 20% 20%] (10) RF:0.00 AvgReturn:9.256%(342868) BLT:6.598%(70595) BST:4.118%(25927) PWR:7.604%(101740) SWR:8.473%(100661) StdDev:9.525%(704471) Ulcer:0.7(41549) DeepestDrawdown:-6.71%(47065) LongestDrawdown:1(1), StartDateSensitivity:5.46%(139420)
//
//   GoldenButterfly: [SCV Gold TSM STT LTT] [20% 20% 20% 20% 20%] (1268826) RF:0.00 AvgReturn:5.669%(2551324) BLT:5.241%(720646) BST:2.849%(394012) PWR:4.224%(2369362) SWR:5.305%(2581584) StdDev:8.103%(306429) Ulcer:3.4(1214869) DeepestDrawdown:-15.33%(484662) LongestDrawdown:3(3), StartDateSensitivity:7.71%(490731)
//   ...Calculate rank scores for the portfolios
//   ...rank by all their ranks (equally weighted)
//   As good or better than GoldenButterfly: 4463
//
//   Best by each ranking:
//   Best AvgReturn: [Momentum Factor Health Care Gold IT Corp LTT] [20% 20% 20% 20% 20%] (1157) RF:0.00 AvgReturn:8.419%(1) BLT:6.737%(35) BST:3.834%(352) PWR:6.663%(175) SWR:7.604%(327) StdDev:7.992%(3956) Ulcer:0.7(1215) DeepestDrawdown:-7.48%(1255) LongestDrawdown:2(2), StartDateSensitivity:7.15%(4134)
//   Best BaselineLTReturn: [Energy Health Care Wellesley Gold LTT] [20% 20% 20% 20% 20%] (877) RF:0.00 AvgReturn:8.094%(27) BLT:7.305%(1) BST:3.220%(2427) PWR:6.613%(207) SWR:7.589%(346) StdDev:7.891%(3534) Ulcer:0.8(1673) DeepestDrawdown:-8.41%(1737) LongestDrawdown:1(1), StartDateSensitivity:6.45%(3576)
//   Best BaselineSTReturn: [Precious Metals Int'l Bd Dividend Growth Health Care LTT] [20% 20% 20% 20% 20%] (393) RF:0.00 AvgReturn:8.291%(5) BLT:6.026%(728) BST:4.663%(1) PWR:7.058%(9) SWR:8.059%(32) StdDev:8.073%(4314) Ulcer:0.7(1130) DeepestDrawdown:-7.25%(1170) LongestDrawdown:1(1), StartDateSensitivity:4.96%(1859)
//   Best PWR30: [Energy Health Care Wellesley STB LTT] [20% 20% 20% 20% 20%] (1385) RF:0.00 AvgReturn:8.117%(21) BLT:5.960%(911) BST:3.129%(2877) PWR:7.149%(1) SWR:8.231%(4) StdDev:8.060%(4255) Ulcer:0.8(1671) DeepestDrawdown:-8.40%(1734) LongestDrawdown:1(1), StartDateSensitivity:6.54%(3655)
//   Best SWR30: [Global Bd Health Care Emerging STB LTT] [20% 20% 20% 20% 20%] (752) RF:0.00 AvgReturn:8.105%(24) BLT:5.655%(1943) BST:3.675%(657) PWR:7.121%(4) SWR:8.281%(1) StdDev:8.095%(4429) Ulcer:0.7(1156) DeepestDrawdown:-7.32%(1196) LongestDrawdown:2(2), StartDateSensitivity:4.40%(1063)
//   Best StdDev: [Global Bd Int'l Bd Health Care Wellesley Gold] [20% 20% 20% 20% 20%] (2233) RF:0.00 AvgReturn:6.395%(4298) BLT:5.313%(4025) BST:3.607%(804) PWR:5.298%(3637) SWR:6.556%(3493) StdDev:5.710%(1) Ulcer:0.4(72) DeepestDrawdown:-3.71%(81) LongestDrawdown:1(1), StartDateSensitivity:3.41%(243)
//   Best UlcerScore: [Wellington Health Care T-Bill Gold LTT] [20% 20% 20% 20% 20%] (1672) RF:0.00 AvgReturn:6.625%(3854) BLT:5.530%(2656) BST:3.739%(520) PWR:5.368%(3456) SWR:6.611%(3331) StdDev:6.202%(75) Ulcer:0.3(1) DeepestDrawdown:-2.54%(2) LongestDrawdown:2(2), StartDateSensitivity:3.01%(91)
//   Best DeepestDrawdown: [Health Care Min Vol Factor Gold ST Munis LTT] [20% 20% 20% 20% 20%] (1145) RF:0.00 AvgReturn:6.890%(2957) BLT:5.675%(1859) BST:3.434%(1433) PWR:5.522%(2990) SWR:6.728%(2907) StdDev:6.561%(318) Ulcer:0.4(107) DeepestDrawdown:-2.47%(1) LongestDrawdown:2(2), StartDateSensitivity:5.02%(1929)
//   Best LongestDrawdown: [Int'l Bd Dividend Growth Health Care Gold LTT] [20% 20% 20% 20% 20%] (1) RF:0.00 AvgReturn:7.682%(355) BLT:6.144%(498) BST:4.120%(106) PWR:6.246%(719) SWR:7.310%(875) StdDev:6.968%(949) Ulcer:0.3(45) DeepestDrawdown:-3.48%(55) LongestDrawdown:1(1), StartDateSensitivity:4.64%(1414)
//   Best StartDateSensitivity: [IT Munis Global Bd Health Care Gold LTT] [20% 20% 20% 20% 20%] (3103) RF:0.00 AvgReturn:6.261%(4419) BLT:5.350%(3759) BST:3.734%(532) PWR:5.025%(4274) SWR:6.342%(4150) StdDev:6.050%(31) Ulcer:0.4(91) DeepestDrawdown:-3.82%(102) LongestDrawdown:2(2), StartDateSensitivity:2.22%(1)
//
//   All as good or better:
//    0 [Int'l Bd Dividend Growth Health Care Gold LTT] [20% 20% 20% 20% 20%] (1) RF:0.00 AvgReturn:2.013%(355) BLT:0.903%(0) BST:1.271%(0) PWR:2.022%(719) SWR:2.005%(875) StdDev:-1.135%(949) Ulcer:-3.1(45) DeepestDrawdown:11.85%(55) LongestDrawdown:-2(1), StartDateSensitivity:-3.07%(1414)
//    1 [Global Bd Dividend Growth Health Care Gold LTT] [20% 20% 20% 20% 20%] (2) RF:0.00 AvgReturn:2.062%(285) BLT:0.952%(0) BST:1.377%(0) PWR:2.037%(681) SWR:2.021%(827) StdDev:-0.990%(1276) Ulcer:-3.1(71) DeepestDrawdown:11.64%(80) LongestDrawdown:-2(1), StartDateSensitivity:-3.17%(1281)
//    2 [Health Care Gold MCV STB LTT] [20% 20% 20% 20% 20%] (3) RF:0.00 AvgReturn:1.891%(596) BLT:1.197%(0) BST:1.261%(0) PWR:1.880%(1069) SWR:1.833%(1343) StdDev:-0.950%(1361) Ulcer:-3.0(207) DeepestDrawdown:10.92%(226) LongestDrawdown:-2(1), StartDateSensitivity:-4.00%(400)
//    3 [Int'l Bd Health Care Gold MCV LTT] [20% 20% 20% 20% 20%] (4) RF:0.00 AvgReturn:2.061%(286) BLT:1.428%(0) BST:0.971%(0) PWR:1.970%(838) SWR:1.879%(1213) StdDev:-0.859%(1578) Ulcer:-3.0(200) DeepestDrawdown:10.96%(219) LongestDrawdown:-2(1), StartDateSensitivity:-3.48%(871)
//    4 [Health Care Gold STB Value Factor LTT] [20% 20% 20% 20% 20%] (5) RF:0.00 AvgReturn:1.826%(772) BLT:1.068%(0) BST:1.239%(0) PWR:1.846%(1160) SWR:1.826%(1366) StdDev:-1.086%(1051) Ulcer:-3.0(166) DeepestDrawdown:11.14%(186) LongestDrawdown:-2(1), StartDateSensitivity:-3.49%(853)
//   Finished GB analysis in 147.352553ms
//   --- PASS: TestAllKAssetPortfolios (103.94s)
func TestAllKAssetPortfolios(t *testing.T) {
	t.Skip("Run manually")
	var (
		goblFileBetterThanGB = func(k int) string {
			return fmt.Sprintf("testdata/TestAllKAssetPortfolios_PortfolioStats_k%d_betterThanGoldenButterfly.gobl.gz", k)
		}
	)
	t.Run("Evaluate", func(t *testing.T) {
		t.Skip("Run manually")
		// g := NewGomegaWithT(t)

		fmt.Println("Evaluating possible portfolio combinations...")

		// generate portfolio combinations
		// across N goroutines -- evaluate combination
		//  -- if it's better than GoldenButterfly, save it
		//  -- writer channel writes to GOBL file

		// some sample output:
		// k=8:
		// Finished writing 1,726,099 rows in 13m55s
		// Finished evaluating 886322710 portfolios in 13m55s (1,061,504 portfolios per second)
		// k=10:
		// Finished writing 35,896,713 rows in 6h31m13s
		// Finished evaluating 19,499,099,620 portfolios in 6h31m13s (830,700 portfolios per second)
		// k=11
		// Finished writing 129,460,837 rows in 30h42m48s
		// Finished evaluating 76,223,753,060 portfolios in 30h42m48s (689,383 portfolios per second)

		for k := 5; k <= 5; k++ {
			// look at all `k` combinations of assets
			startAt := time.Now()
			targetAllocations := make([]Percent, k)
			for i := 0; i < k; i++ {
				targetAllocations[i] = Percent(1.0 / float64(k))
			}
			nCr := Binomial(len(data.Names()), k)
			fmt.Println()
			fmt.Println(time.Now(), "k =", k, "nCr =", nCr, "TargetAllocations", targetAllocations)

			gbAssets := make(map[string]bool, 5)
			for _, a := range []string{"TSM", "SCV", "Gold", "LTT", "STT"} {
				gbAssets[a] = true
			}
			gbStat := mustGoldenButterflyStat()

			GoEvaluateAndFindBetterThanGB := func(assetCombinationBatches <-chan [][]string) <-chan *PortfolioStat {
				out := make(chan *PortfolioStat, 10)
				go func() {
					defer close(out)
					for batch := range assetCombinationBatches {
						for _, assets := range batch {
							returnsList := data.PortfolioReturnsList(assets...)
							returns, err := portfolioReturns(returnsList, targetAllocations)
							if err != nil {
								panic(err.Error())
							}
							combination := Combination{Assets: assets, Percentages: targetAllocations}
							statIfBetter := evaluatePortfolioIfAsGoodOrBetterThan(returns, combination, gbStat)
							if statIfBetter != nil {
								out <- statIfBetter
							}
							// check if these are GB assets
							if consistsOf(assets, gbAssets) {
								fmt.Println("Seen GB assets!", assets, "StatIfBetter:", statIfBetter)
							}
						}
					}
				}()
				return out
			}
			combinationsCh := GoEnumerateCombinations(data.Names(), k, 10_000)
			// fan out to multiple workers
			// k=6
			//  6 workers: 1278067 portfolios per second
			//  7 workers: 1273075    1426317
			//  8 workers: 1366385    1060228  1478084
			//  9 workers: 1430209    1219254  1226238 1468269
			// 10 workers: 1398095    1223920
			// 11 workers: 1418906    1254684  1193651
			// 12 workers: 1312534

			var workersOutput []<-chan *PortfolioStat
			for i := 0; i < 9; i++ {
				results := GoEvaluateAndFindBetterThanGB(combinationsCh)
				workersOutput = append(workersOutput, results)
			}
			// merge workers' output
			resultsCh := GoMerge(workersOutput...)
			// encode to file
			// err := goblEncodeToFile(goblFileBetterThanGB(k), resultsCh)
			// g.Expect(err).To(Succeed())

			// just count results
			count := 0
			for range resultsCh {
				count++
			}
			fmt.Println("Result count:", count)

			elapsed := time.Since(startAt)
			fmt.Printf("Finished evaluating %d portfolios in %v (%d portfolios per second)\n",
				nCr, elapsed, int(float64(nCr)/elapsed.Seconds()))
		}
	})
	t.Run("inspect better than GoldenButterfly", func(t *testing.T) {
		g := NewGomegaWithT(t)
		startAt := time.Now()
		var betterThanGB []*PortfolioStat
		err := goblDecodeFromFile(goblFileBetterThanGB(3), func(stat *PortfolioStat) (shouldContinue bool) {
			betterThanGB = append(betterThanGB, stat)
			return true
		})
		g.Expect(err).To(Succeed())
		fmt.Println("As good or better than GoldenButterfly:", len(betterThanGB), "found in", time.Since(startAt))

		// err := csvEncodeToFile(betterThanGB, csvFileBetterThanGB)
		// g.Expect(err).To(Succeed())
		// return

		// One looks noticeably good:
		// ["Gold","Wellington","Health Care","Global Bd","Int'l Bd"]  PRW30: 5.49%  StdDev:5.99%

		fmt.Println(len(betterThanGB), "portfolios better than GoldenButterfly")

		gbStat := mustGoldenButterflyStat()
		fmt.Println("GoldenButterfly: ", gbStat)

		// fmt.Println("\nAll as good or better:")
		// for i, p := range betterThanGB[:min(len(betterThanGB), 5)] {
		// 	fmt.Println(" ", i+1, p.DiffPerformance(*gbStat))
		// }

		count := 0
		for _, s := range betterThanGB {
			if !contains(s.Assets, "Health Care") {
				count++
			}
		}
		fmt.Println("Counted", count, "portfolios without Health Care.")

		// 19 portfolios use only 3 assets!
		{
			n := 4
			fmt.Println("Less than", n, "assets:")
			count = 0
			for _, s := range betterThanGB {
				if len(s.Assets) < 4 {
					// fmt.Println(" -", s)
					count++
				}
			}
			fmt.Println("Counted", count, "portfolios with less than", n, "assets.")
		}

		// what's the longest history we have? -- 51 years
		// a lot of 48+ years.. seems like a lot of variations of GoldenButterfly
		fmt.Println("Longest histories...")
		maxHistoricalYears := 0
		for _, s := range betterThanGB {
			years := len(data.PortfolioReturnsList(s.Assets...)[0])
			if years > maxHistoricalYears {
				maxHistoricalYears = years
			}
			// if years > 45 {
			// 	fmt.Println(" -", years, "years: ", s)
			// }
		}
		fmt.Println("Max historical years of any portfolio:", maxHistoricalYears)

		// which assets are never included in any portfolios?
		// used 52 assets, never used: "Commodities"
		{
			seenAssets := map[string]bool{}
			for _, s := range betterThanGB {
				for _, a := range s.Assets {
					seenAssets[a] = true
				}
			}
			fmt.Println("Portfolios use", len(seenAssets), "assets.")
			for _, name := range data.Names() {
				if !seenAssets[name] {
					fmt.Println(" - never used:", name)
				}
			}
		}
	})
	t.Run("parse GOBL.gz", func(t *testing.T) {
		goblFileBetterThanGB = func(k int) string {
			return fmt.Sprintf("testdata/snapshot/TestAllKAssetPortfolios_PortfolioStats_k%d_betterThanGoldenButterfly.gobl.gz", k)
		}
		t.Run("save asset lists", func(t *testing.T) {
			t.Skip()
			g := NewGomegaWithT(t)

			var (
				output = "testdata/betterThanGoldenButterflyPortfolios.jsonl.gz"
			)
			jsonl := make(chan interface{})
			jsonlErrCh := GoJSONLEncodeToFile(jsonl, output)
			for k := 1; k <= 11; k++ {
				input := goblFileBetterThanGB(k)
				err := goblDecodeFromFile(input, func(stat *PortfolioStat) bool {
					jsonl <- stat.Assets
					return true
				})
				g.Expect(err).To(Succeed())
			}

			close(jsonl)
			g.Expect(<-jsonlErrCh).To(Succeed())
		})
		t.Run("convert to CSV", func(t *testing.T) {
			g := NewGomegaWithT(t)

			// which assets have >N years data?
			const minYears = 45
			allowAssets := map[string]bool{}
			for _, name := range data.Names() {
				s := data.MustFind(name)
				numYears := s.YearsOfData()
				if numYears > minYears {
					allowAssets[name] = true
					fmt.Println(numYears, name)
				}
			}
			fmt.Println("Found", len(allowAssets), "matching")
			// return
			const minK, maxK = 1, 11
			var (
				output    = fmt.Sprintf("testdata/betterThanGoldenButterflyPortfolios_k%d-%d_over%dyears_sensitive05_ulcer13.csv", minK, maxK, minYears)
				totalRows = 0
			)
			// write
			var itemsCh = make(chan PortfolioStat)
			errorCh := GoCSVEncodeToFile(itemsCh, output)

			for k := minK; k <= maxK; k++ {
				input := goblFileBetterThanGB(k)
				err := goblDecodeFromFile(input, func(stat *PortfolioStat) bool {
					// filtering
					if stat.StartDateSensitivity > 0.05 ||
						stat.UlcerScore > 1.3 {
						return true
					}
					for _, asset := range stat.Assets {
						if !allowAssets[asset] {
							return true
						}
					}
					// try to write to channel
					select {
					case itemsCh <- *stat:
					case err := <-errorCh:
						g.Expect(err).To(Succeed())
					}
					// fmt.Println(" - ", stat)
					totalRows++

					return true
				})
				g.Expect(err).To(Succeed())
			}
			close(itemsCh)
			g.Expect(<-errorCh).To(Succeed())
			fmt.Println("Wrote total rows:", totalRows)
		})
		t.Run("convert to SQLite", func(t *testing.T) {
			g := NewGomegaWithT(t)
			const sqliteFile = "testdata/betterThanGoldenButterfly_extraMetrics.sqlite"
			db, err := sql.Open("sqlite3", sqliteFile+"?mode=rwc")
			g.Expect(err).To(Succeed())
			defer db.Close()

			// create table
			sqlStmt := `
				CREATE TABLE IF NOT EXISTS portfolios (
				    assets                TEXT NOT NULL,
				    num_assets            INTEGER,
				    num_years             INTEGER,
				    avg_return            REAL,
				    baseline_lt_return    REAL,
				    baseline_st_return    REAL,
				    pwr30                 REAL,
				    swr30                 REAL,
				    std_dev               REAL,
				    ulcer_score           REAL,
				    deepest_drawdown      REAL,
				    longest_drawdown      REAL,
				    startdate_sensitivity REAL,
				    pwr10                 REAL,
				    pwr10_stdev           REAL,
				    pwr10_slope           REAL,
				    pwr30_stdev           REAL,
				    pwr30_slope           REAL
				    );
			`
			_, err = db.Exec(sqlStmt)
			g.Expect(err).To(Succeed())

			// spawn goroutine to insert data
			items := make(chan *PortfolioStat, 100)
			errorCh := make(chan error)
			go func() (err error) {
				defer func() {
					errorCh <- err
				}()
				tx, err := db.Begin()
				g.Expect(err).To(Succeed())
				stmt, err := tx.Prepare(`
					INSERT INTO portfolios (
						assets,
						num_assets,
						num_years,
						avg_return,
						baseline_lt_return,
						baseline_st_return,
						pwr30,
						swr30,
						std_dev,
						ulcer_score,
						deepest_drawdown,
						longest_drawdown,
						startdate_sensitivity,
						pwr10,
						pwr10_stdev,
						pwr10_slope,
						pwr30_stdev,
						pwr30_slope
					)
					VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				`)
				g.Expect(err).To(Succeed())
				for stat := range items {
					returnsList := data.PortfolioReturnsList(stat.Assets...)
					returns, err := portfolioReturns(returnsList, stat.Percentages)
					minPWR10, _ := minPWR(returns, 10)
					pwrs10 := allPWRs(returns, 10)
					pwrs30 := allPWRs(returns, 30)
					_, err = stmt.Exec(
						"|"+strings.Join(stat.Assets, "|")+"|", // encode as string
						len(stat.Assets), // NumAssets
						len(returns),     // NumYears
						stat.AvgReturn.Float(),
						stat.BaselineLTReturn.Float(),
						stat.BaselineSTReturn.Float(),
						stat.PWR30.Float(),
						stat.SWR30.Float(),
						stat.StdDev.Float(),
						stat.UlcerScore,
						stat.DeepestDrawdown.Float(),
						stat.LongestDrawdown,
						stat.StartDateSensitivity.Float(),
						minPWR10.Float(),
						standardDeviation(pwrs10).Float(),
						slope(pwrs10).Float(),
						standardDeviation(pwrs30).Float(),
						slope(pwrs30).Float(),
					)
					if err != nil {
						return err
					}
				}
				g.Expect(tx.Commit()).To(Succeed())
				g.Expect(stmt.Close()).To(Succeed())
				return nil
			}()

			const minK, maxK = 1, 10
			var (
				totalRows = 0
			)
			for k := minK; k <= maxK; k++ {
				input := goblFileBetterThanGB(k)
				// batch := make([]*PortfolioStat, 0, batchSize)
				err = goblDecodeFromFile(input, func(stat *PortfolioStat) bool {
					// filtering
					// if stat.StartDateSensitivity > 0.05 ||
					// 	stat.UlcerScore > 1.3 {
					// 	return true
					// }
					// for _, asset := range stat.Assets {
					// 	if !allowAssets[asset] {
					// 		return true
					// 	}
					// }
					// try to write to channel
					items <- stat
					totalRows++
					return true
				})
				g.Expect(err).To(Succeed())
			}
			fmt.Println("Done writing items to channel.")
			close(items)
			fmt.Println("Waiting for goroutine to finish...")
			g.Expect(<-errorCh).To(Succeed())
			fmt.Println("Wrote total rows:", totalRows)
		})
		t.Run("some extra metrics", func(t *testing.T) {
			g := NewGomegaWithT(t)

			var (
				input = goblFileBetterThanGB(8)

				minPWRn = func(returns []Percent, nYears int) Percent {
					val, _ := minPWR(returns, nYears)
					return val
				}

				gbStat    = mustGoldenButterflyStat()
				gbReturns = gbStat.MustReturns()
				gbPWR3    = minPWRn(gbReturns, 3)
				gbPWR10   = minPWRn(gbReturns, 10)
			)
			fmt.Println("GoldenButterfly PWR3:", gbPWR3, "PWR10:", gbPWR10)
			var (
				total          = 0
				asGoodOrBetter = 0

				betterByNMetricsCounts = map[int]int{}
			)
			err := goblDecodeFromFile(input, func(stat *PortfolioStat) bool {
				total++
				var (
					returns = stat.MustReturns()
					pwr3    = minPWRn(returns, 3)
					pwr10   = minPWRn(returns, 10)
				)
				betterByNMetricsCounts[CountBetterMetrics(stat, gbStat)]++
				if pwr3 >= gbPWR3 && pwr10 >= gbPWR10 {
					// fmt.Printf(" - PWR3:%0.1f%% PWR10:%0.1f%%: %v\n", pwr3*100, pwr10*100, stat)
					asGoodOrBetter++
				}
				return true
			})
			g.Expect(err).To(Succeed())
			PrintMemUsage()
			fmt.Printf("Of %d, found %d as good or better (%0.1f%%)\n",
				total, asGoodOrBetter, float64(asGoodOrBetter)/float64(total)*100)
			fmt.Printf("BetterByMetricsCounts: %v\n", betterByNMetricsCounts)
		})
	})
}

func consistsOf(assets []string, expectedAssets map[string]bool) bool {
	if len(assets) != len(expectedAssets) {
		return false
	}
	for _, asset := range assets {
		if !expectedAssets[asset] {
			return false
		}
	}
	return true
}

func CountBetterMetrics(stat, other *PortfolioStat) int {
	n := 0
	if stat.AvgReturn > other.AvgReturn {
		n++
	}
	if stat.BaselineLTReturn > other.BaselineLTReturn {
		n++
	}
	if stat.BaselineSTReturn > other.BaselineSTReturn {
		n++
	}
	if stat.PWR30 > other.PWR30 {
		n++
	}
	if stat.SWR30 > other.SWR30 {
		n++
	}
	if stat.StdDev < other.StdDev {
		n++
	}
	if stat.UlcerScore < other.UlcerScore {
		n++
	}
	if stat.DeepestDrawdown > other.DeepestDrawdown {
		n++
	}
	if stat.LongestDrawdown < other.LongestDrawdown {
		n++
	}
	if stat.StartDateSensitivity < other.StartDateSensitivity {
		n++
	}
	return n
}

func mustGoldenButterflyStat() *PortfolioStat {
	assets := []string{"LTT", "Gold", "STT", "SCV", "TSM"}
	targetAllocations := ReadablePercents(20, 20, 20, 20, 20)
	returnsList := data.PortfolioReturnsList(assets...)
	returns, err := portfolioReturns(returnsList, targetAllocations)
	if err != nil {
		panic(err.Error())
	}
	stat := evaluatePortfolio(returns, Combination{Assets: assets, Percentages: targetAllocations})
	return stat
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc: %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc: %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys: %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC: %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func GoCSVEncodeToFile(rows <-chan PortfolioStat, filename string) <-chan error {
	startAt := time.Now()

	errorCh := make(chan error)

	go func() (err error) {
		defer func() {
			errorCh <- err
			close(errorCh)
		}()
		fmt.Println("CSV-encoding rows to", filename)
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		csvWriter := csv.NewWriter(f)
		err = csvWriter.Write([]string{
			"Assets",
			"NumAssets",
			"NumYears",
			"AvgReturn",
			"BaselineLTReturn",
			"BaselineSTReturn",
			"PWR30",
			"SWR30",
			"StdDev",
			"UlcerScore",
			"DeepestDrawdown",
			"LongestDrawdown",
			"StartDateSensitivity",
		})
		if err != nil {
			return err
		}
		var count int
		for row := range rows {
			count++
			returnsList := data.PortfolioReturnsList(row.Assets...)
			returns, err := portfolioReturns(returnsList, row.Percentages)
			if err != nil {
				return err
			}
			err = csvWriter.Write([]string{
				strings.Join(row.Assets, ";"), // encode as JS array of strings
				strconv.Itoa(len(row.Assets)), // NumAssets
				strconv.Itoa(len(returns)),    // NumYears
				fmt.Sprintf("%f", row.AvgReturn.Float()),
				fmt.Sprintf("%f", row.BaselineLTReturn.Float()),
				fmt.Sprintf("%f", row.BaselineSTReturn.Float()),
				fmt.Sprintf("%f", row.PWR30.Float()),
				fmt.Sprintf("%f", row.SWR30.Float()),
				fmt.Sprintf("%f", row.StdDev.Float()),
				fmt.Sprintf("%f", row.UlcerScore),
				fmt.Sprintf("%f", row.DeepestDrawdown.Float()),
				strconv.Itoa(row.LongestDrawdown),
				fmt.Sprintf("%f", row.StartDateSensitivity.Float()),
			})
			if err != nil {
				return fmt.Errorf("error writing row #%d (%s): %w", count, row, err)
			}
		}
		csvWriter.Flush()
		fmt.Println("Finished writing in", time.Since(startAt))
		return nil
	}()
	return errorCh
}

func TestSingle(t *testing.T) {
	g := NewGomegaWithT(t)
	for _, name := range data.Names() {
		series := data.MustFind(name)
		returns, err := portfolioReturns([][]Percent{series.AnnualReturns}, []Percent{1})
		g.Expect(err).To(Succeed())
		stat := evaluatePortfolio(returns, Combination{
			Assets:      []string{name},
			Percentages: []Percent{1},
		})
		fmt.Println(stat)
	}
}

func GoJSONLEncodeToFile(items <-chan interface{}, filename string) <-chan error {
	startAt := time.Now()

	errorCh := make(chan error)

	go func() (err error) {
		defer func() {
			errorCh <- err
			close(errorCh)
		}()
		var count int
		fmt.Println("JSONL-encoding data to", filename)
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		gzipWriter := gzip.NewWriter(f)
		defer gzipWriter.Close()
		jsonWriter := json.NewEncoder(gzipWriter)
		for item := range items {
			count++
			err = jsonWriter.Encode(item)
			if err != nil {
				return fmt.Errorf("error writing row #%d (%s): %w", count, item, err)
			}
		}
		fmt.Println("Finished writing in", time.Since(startAt))
		return nil
	}()
	return errorCh
}

func mustJSONMarshal(obj interface{}) string {
	marshal, err := json.Marshal(obj)
	if err != nil {
		panic(err.Error())
	}
	return string(marshal)
}

// gobDecodeFromFile reads the GOB-encoded object from a file.
func gobDecodeFromFile(g *GomegaWithT, filename string) []*PortfolioStat {
	startAt := time.Now()
	fmt.Println("GOB-decoding from", filename)
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	g.Expect(err).To(Succeed())

	gzipReader, err := gzip.NewReader(f)
	g.Expect(err).To(Succeed())

	var obj []*PortfolioStat
	err = gob.NewDecoder(gzipReader).Decode(&obj)
	g.Expect(err).To(Succeed())
	fmt.Println("Finished decoding", len(obj), "rows in", time.Since(startAt))
	return obj
}

// goblEncodeToFile writes the GOB-encoded object to a file.
func goblEncodeToFile(filename string, obj <-chan *PortfolioStat) error {
	startAt := time.Now()
	fmt.Println("GOBL-encoding rows to", filename)
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	gzipWriter := gzip.NewWriter(f)
	defer gzipWriter.Close()
	encoder := gob.NewEncoder(gzipWriter)
	n := 0
	for o := range obj {
		n++
		if err = encoder.Encode(o); err != nil {
			return err
		}
	}
	fmt.Println("Finished writing", n, "rows in", time.Since(startAt))
	return nil
}

// goblDecodeFromFile reads the GOB-encoded object from a file.
func goblDecodeFromFile(filename string, handle func(*PortfolioStat) (shouldContinue bool)) error {
	startAt := time.Now()
	fmt.Println("GOBL-decoding from", filename)
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	decoder := gob.NewDecoder(gzipReader)

	count := 0
	for {
		var obj *PortfolioStat
		err = decoder.Decode(&obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		count++
		if shouldContinue := handle(obj); !shouldContinue {
			break
		}
	}
	elapsed := time.Since(startAt)
	fmt.Printf("Finished decoding %d objects in %v (%d/sec)\n",
		count, elapsed, int(float64(count)/elapsed.Seconds()))
	return nil
}

func findBetterThanGoldenButterfly(gbStat *PortfolioStat, results []*PortfolioStat) []*PortfolioStat {
	startAt := time.Now()
	fmt.Println("\nGoldenButterfly:", gbStat)
	// find as good or better than GoldenButterfly
	betterThanGB := CopyAll(FindMany(results, AsGoodOrBetterThan(gbStat)))
	fmt.Println("As good or better than GoldenButterfly:", len(betterThanGB), "found in", time.Since(startAt))

	RankPortfoliosInPlace(betterThanGB)
	PrintBestByEachRanking(betterThanGB)
	fmt.Println("\nAll as good or better:")
	for i, p := range betterThanGB[:min(len(betterThanGB), 5)] {
		fmt.Println(" ", i+1, p.DiffPerformance(*gbStat))
	}
	return betterThanGB
}
func contains(slice []string, element string) bool {
	for _, a := range slice {
		if a == element {
			return true
		}
	}
	return false
}

func PrintBestByEachRanking(results []*PortfolioStat) {
	fmt.Println("\nBest by each ranking:")
	fmt.Println("Best AvgReturn:", FindOne(results, func(p *PortfolioStat) bool { return p.AvgReturnRank.Ordinal == 1 }))
	fmt.Println("Best BaselineLTReturn:", FindOne(results, func(p *PortfolioStat) bool { return p.BaselineLTReturnRank.Ordinal == 1 }))
	fmt.Println("Best BaselineSTReturn:", FindOne(results, func(p *PortfolioStat) bool { return p.BaselineSTReturnRank.Ordinal == 1 }))
	fmt.Println("Best PWR30:", FindOne(results, func(p *PortfolioStat) bool { return p.PWR30Rank.Ordinal == 1 }))
	fmt.Println("Best SWR30:", FindOne(results, func(p *PortfolioStat) bool { return p.SWR30Rank.Ordinal == 1 }))
	fmt.Println("Best StdDev:", FindOne(results, func(p *PortfolioStat) bool { return p.StdDevRank.Ordinal == 1 }))
	fmt.Println("Best UlcerScore:", FindOne(results, func(p *PortfolioStat) bool { return p.UlcerScoreRank.Ordinal == 1 }))
	fmt.Println("Best DeepestDrawdown:", FindOne(results, func(p *PortfolioStat) bool { return p.DeepestDrawdownRank.Ordinal == 1 }))
	fmt.Println("Best LongestDrawdown:", FindOne(results, func(p *PortfolioStat) bool { return p.LongestDrawdownRank.Ordinal == 1 }))
	fmt.Println("Best StartDateSensitivity:", FindOne(results, func(p *PortfolioStat) bool { return p.StartDateSensitivityRank.Ordinal == 1 }))
}

func approxEqual(x, y, tolerance float64) bool {
	return math.Abs(x-y) < tolerance
}

func Test_seriesRange(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(seriesRange(25)).To(Equal([]float64{25, 50, 75, 100}))
	g.Expect(seriesRange(10)).To(Equal([]float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}))
	g.Expect(seriesRange(33.333333333333333)).To(Equal([]float64{33.333333333333336, 66.66666666666667, 100}))
}

func Test_series(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(series(25, 100, 25)).To(Equal([]float64{25, 50, 75, 100}))
	g.Expect(series(12.5, 100, 12.5)).To(Equal([]float64{12.5, 25, 37.5, 50, 62.5, 75, 87.5, 100}))
}

func seriesRange(step float64) []float64 {
	return series(step, 100, step)
}

func series(start, end, step float64) []float64 {
	var res []float64
	for i := start; i <= end; i += step {
		res = append(res, i)
	}
	return res
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestEnumerateCombinations(t *testing.T) {

	makeAccumulateAllFn := func(buffer []string) (*[][]string, func() error) {
		var result [][]string
		return &result, func() error {
			fmt.Println(buffer)
			copyBuffer := make([]string, len(buffer))
			copy(copyBuffer, buffer)
			result = append(result, copyBuffer)
			return nil
		}
	}

	t.Run("no combinations", func(t *testing.T) {
		g := NewGomegaWithT(t)
		var nilBuffer []string
		result, fn := makeAccumulateAllFn(nilBuffer)
		g.Expect(EnumerateCombinations(nil, 0, nilBuffer, fn)).To(Succeed())
		g.Expect(*result).To(HaveLen(0))
	})

	t.Run("k=1", func(t *testing.T) {
		var (
			kOne      = 1
			bufferOne = make([]string, kOne)
		)
		t.Run("one combination", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferOne)
			g.Expect(EnumerateCombinations([]string{"1"}, kOne, bufferOne, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{{"1"}}))
		})
		t.Run("two combinations", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferOne)
			g.Expect(EnumerateCombinations([]string{"1", "2"}, kOne, bufferOne, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{{"1"}, {"2"}}))
		})
		t.Run("three combinations", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferOne)
			g.Expect(EnumerateCombinations([]string{"1", "2", "3"}, kOne, bufferOne, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{{"1"}, {"2"}, {"3"}}))
		})
	})

	t.Run("k=2", func(t *testing.T) {
		var (
			kTwo      = 2
			bufferTwo = make([]string, kTwo)
		)
		t.Run("one combination", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferTwo)
			g.Expect(EnumerateCombinations([]string{"1", "2"}, kTwo, bufferTwo, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{{"1", "2"}}))
		})
		t.Run("three combinations", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferTwo)
			g.Expect(EnumerateCombinations([]string{"1", "2", "3"}, kTwo, bufferTwo, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{
				{"1", "2"}, {"1", "3"},
				{"2", "3"}}))
		})
		t.Run("six combinations", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferTwo)
			g.Expect(EnumerateCombinations([]string{"1", "2", "3", "4"}, kTwo, bufferTwo, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{
				{"1", "2"}, {"1", "3"}, {"1", "4"},
				{"2", "3"}, {"2", "4"},
				{"3", "4"}}))
		})
	})
	t.Run("k=3", func(t *testing.T) {
		var (
			kThree      = 3
			bufferThree = make([]string, kThree)
		)
		t.Run("one combination", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferThree)
			g.Expect(EnumerateCombinations([]string{"1", "2", "3"}, kThree, bufferThree, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{{"1", "2", "3"}}))
		})
		t.Run("four combination", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferThree)
			g.Expect(EnumerateCombinations([]string{"1", "2", "3", "4"}, kThree, bufferThree, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{
				{"1", "2", "3"}, {"1", "2", "4"},
				{"1", "3", "4"}, {"2", "3", "4"}}))
		})
		t.Run("ten combination", func(t *testing.T) {
			g := NewGomegaWithT(t)
			result, fn := makeAccumulateAllFn(bufferThree)
			g.Expect(EnumerateCombinations([]string{"1", "2", "3", "4", "5"}, kThree, bufferThree, fn)).To(Succeed())
			g.Expect(*result).To(Equal([][]string{
				{"1", "2", "3"}, {"1", "2", "4"}, {"1", "2", "5"},
				{"1", "3", "4"}, {"1", "3", "5"}, {"1", "4", "5"},
				{"2", "3", "4"}, {"2", "3", "5"}, {"2", "4", "5"},
				{"3", "4", "5"}}))
		})
	})
	t.Run("validate some counts", func(t *testing.T) {
		g := NewGomegaWithT(t)

		makeCounterFn := func() (*int, func() error) {
			var count int
			return &count, func() error {
				count++
				return nil
			}
		}

		verify := func(n, r, expectedCount int) {
			var (
				xs        = make([]string, n)
				buffer    = make([]string, r)
				count, fn = makeCounterFn()
			)
			// t.Log(fmt.Sprintf("n=%d r=%d count=%d", n, r, expectedCount))
			g.Expect(EnumerateCombinations(xs, r, buffer, fn)).To(Succeed())
			g.Expect(*count).To(Equal(expectedCount))
			if n >= r && r > 0 {
				g.Expect(Binomial(n, r)).To(Equal(expectedCount)) // verify the counts would match nCr result
			}
		}
		// expectedCount=0
		verify(0, 0, 0)
		verify(0, 1, 0)
		verify(1, 2, 0)
		// n=1
		verify(1, 1, 1)
		// n=2
		verify(2, 1, 2)
		verify(2, 2, 1)
		// n=3
		verify(3, 1, 3)
		verify(3, 2, 3)
		verify(3, 3, 1)
		// n=10, see the nice symmetric curve
		verify(10, 1, 10)
		verify(10, 2, 45)
		verify(10, 3, 120)
		verify(10, 4, 210)
		verify(10, 5, 252) // peak
		verify(10, 6, 210)
		verify(10, 7, 120)
		verify(10, 8, 45)
		verify(10, 9, 10)
		// sample a few larger-size results
		verify(100, 2, 4_950)
		verify(100, 3, 161_700)
		verify(50, 4, 230_300)
		verify(25, 6, 177_100)
	})
	t.Run("can end early", func(t *testing.T) {
		// TODO: verify we can end at any point by returning ErrEndEnumeration from our callback
	})
}

// BenchmarkEnumerateCombinations/100_choose_2-12         	1000000000	         0.00119 ns/op
// BenchmarkEnumerateCombinations/100_choose_3-12         	1000000000	         0.0298 ns/op
// BenchmarkEnumerateCombinations/50_choose_4-12          	1000000000	         0.0303 ns/op
// BenchmarkEnumerateCombinations/50_choose_5-12          	1000000000	         0.269 ns/op
func BenchmarkEnumerateCombinations(b *testing.B) {
	makeCounterFn := func() (*int, func() error) {
		var count int
		return &count, func() error {
			count++
			return nil
		}
	}

	benchmark := func(n, r, expectedCount int) {
		var (
			xs        = make([]string, n)
			buffer    = make([]string, r)
			count, fn = makeCounterFn()
		)
		b.ResetTimer()
		err := EnumerateCombinations(xs, r, buffer, fn)
		if err != nil {
			b.Fatal(err)
		}
		if *count != expectedCount {
			b.Fatalf("expected %d to equal %d", *count, expectedCount)
		}
	}
	b.Run("100 choose 2", func(b *testing.B) {
		benchmark(100, 2, 4_950)
	})
	b.Run("100 choose 3", func(b *testing.B) {
		benchmark(100, 3, 161_700)
	})
	b.Run("50 choose 4", func(b *testing.B) {
		benchmark(50, 4, 230_300)
	})
	b.Run("50 choose 5", func(b *testing.B) {
		benchmark(50, 5, 2_118_760)
	})
}

// Benchmark_portfolioReturnsAltogether-12    	  954178	      1263 ns/op
func Benchmark_portfolioReturnsAltogether(b *testing.B) {
	assets := []string{"TSM", "SCV", "Gold", "LTT", "STT"}
	targetAllocations := ReadablePercents(20, 20, 20, 20, 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		returnsList := data.PortfolioReturnsList(assets...)
		_, err := portfolioReturns(returnsList, targetAllocations)
		if err != nil {
			b.Fatal(err)
		}
	}
}
