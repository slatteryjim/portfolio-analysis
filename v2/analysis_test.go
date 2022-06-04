package v2

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	. "github.com/onsi/gomega"

	pa "github.com/slatteryjim/portfolio-analysis"
	"github.com/slatteryjim/portfolio-analysis/data"
)

func TestAllKAssetPortfolios(t *testing.T) {
	t.Skip("Run manually")
	var (
	// sqliteFileBetterThanGB = func(k int) string {
	// 	return fmt.Sprintf("testdata/TestAllKAssetPortfolios_PortfolioStats_k%d_betterThanGoldenButterfly.sqlite", k)
	// }
	)
	t.Run("Evaluate", func(t *testing.T) {
		t.Skip("Run manually")
		g := NewGomegaWithT(t)

		fmt.Println("Starting with", len(data.Names()), "assets in list.")
		assetNames := map[string]struct{}{}
		for _, name := range data.Names() {
			assetNames[name] = struct{}{}
			// fmt.Println("-", name)
		}
		// delete assets to exclude
		{
			// GB assets
			// mustDelete(t, assetNames, "Gold")
			// mustDelete(t, assetNames, "SCV")
			// mustDelete(t, assetNames, "TSM")
			// mustDelete(t, assetNames, "LTT")
			// mustDelete(t, assetNames, "STT")
			// Bond assets
			// mustDelete(t, assetNames, "Int'l Bd")
			// mustDelete(t, assetNames, "Global Bd")
			// mustDelete(t, assetNames, "Hi-Yield Corp Bd")
			// mustDelete(t, assetNames, "LT STRIPS")
			// mustDelete(t, assetNames, "T-Bill")
			// mustDelete(t, assetNames, "STB")
			// mustDelete(t, assetNames, "ITB")
			// mustDelete(t, assetNames, "IT Corp")
			// mustDelete(t, assetNames, "ITT")
			// mustDelete(t, assetNames, "TBM")
			// mustDelete(t, assetNames, "ST Invest. Grade")
			// mustDelete(t, assetNames, "ST Munis")
			// mustDelete(t, assetNames, "IT Munis")
			// mustDelete(t, assetNames, "LT Munis")
		}
		var names []string
		for n := range assetNames {
			names = append(names, n)
		}
		sort.Strings(names)
		fmt.Println("After deletes, have", len(names), "assets in list:")
		for _, name := range names {
			fmt.Println("-", name)
		}

		fmt.Println("Evaluating possible portfolio combinations...")

		// generate portfolio combinations
		// across N goroutines -- evaluate combination
		//  -- if it's better than GoldenButterfly, save it
		//  -- writer channel writes to Sqlite file

		// minStat := pa.MustGoldenButterflyStat()
		var minStat *pa.PortfolioStat // nil; accept all portfolios

		resultsCh := make(chan *pa.PortfolioStat, 10)
		go func() {
			defer close(resultsCh)
			for k := 1; k <= 5; k++ {
				count := 0
				for result := range GoFindKAssetsBetterThanX(minStat, k, names) {
					count++
					resultsCh <- result
				}
				fmt.Printf("k=%d result count: %d\n", k, count)
			}
		}()

		// just count results
		// CountResults(resultsCh)
		err := EncodeResultsToSQLite("output/portfolios.sqlite", resultsCh)
		g.Expect(err).To(Succeed())
	})
	t.Run("Evaluate trimmed down list", func(t *testing.T) {
		t.Skip("Run manually")
		// g := NewGomegaWithT(t)

		fmt.Println("Starting with", len(data.Names()), "assets in list.")
		assetNames := map[string]struct{}{}
		for _, name := range data.Names() {
			assetNames[name] = struct{}{}
			// fmt.Println("-", name)
		}
		// delete assets to exclude
		{
			// GB assets
			// mustDelete(t, assetNames, "Gold")
			mustDelete(t, assetNames, "SCV")
			mustDelete(t, assetNames, "TSM")
			mustDelete(t, assetNames, "LTT")
			mustDelete(t, assetNames, "STT")
			// bond bond assets
			mustDelete(t, assetNames, "Int'l Bd")
			mustDelete(t, assetNames, "Global Bd")
			mustDelete(t, assetNames, "Hi-Yield Corp Bd")
			mustDelete(t, assetNames, "LT STRIPS")
			mustDelete(t, assetNames, "T-Bill")
			mustDelete(t, assetNames, "STB")
			mustDelete(t, assetNames, "ITB")
			mustDelete(t, assetNames, "IT Corp")
			mustDelete(t, assetNames, "ITT")
			mustDelete(t, assetNames, "TBM")
			mustDelete(t, assetNames, "ST Invest. Grade")
			mustDelete(t, assetNames, "ST Munis")
			mustDelete(t, assetNames, "IT Munis")
			mustDelete(t, assetNames, "LT Munis")
		}
		var names []string
		for n := range assetNames {
			names = append(names, n)
		}
		sort.Strings(names)
		fmt.Println("After deletes, have", len(names), "assets in list:")
		for _, name := range names {
			fmt.Println("-", name)
		}

		fmt.Println("Evaluating possible portfolio combinations...")

		// generate portfolio combinations
		// across N goroutines -- evaluate combination
		//  -- if it's better than GoldenButterfly, save it
		//  -- writer channel writes to Sqlite file

		gbStat := pa.MustGoldenButterflyStat()

		// without GB and bond assets, didn't find anything for k from 5 to 13
		//    9: Finished evaluating (wrong number) portfolios in    47s
		//   10: Finished evaluating (wrong number) portfolios in  2m21s
		//   11: Finished evaluating (wrong number) portfolios in  5m14s
		//   12: Finished evaluating (wrong number) portfolios in  9m55s
		//   13: Finished evaluating (wrong number) portfolios in 17m42s

		// Adding back Gold as an allowed asset
		//    8: Finished evaluating  23,535,820 portfolios in  19s (1,237,431 portfolios per second)
		//    9: Finished evaluating  70,607,460 portfolios in 1m9s (1,026,176 portfolios per second)
		//   10: Finished evaluating 183,579,396 portfolios in 3m7s   (984,209 portfolios per second)

		resultsCh := make(chan *pa.PortfolioStat, 10)
		go func() {
			defer close(resultsCh)
			for k := 11; k <= 11; k++ {
				count := 0
				for result := range GoFindKAssetsBetterThanX(gbStat, k, names) {
					count++
					resultsCh <- result
				}
				fmt.Printf("k=%d result count: %d\n", k, count)
			}
		}()

		// just count results
		count := 0
		for result := range resultsCh {
			count++
			fmt.Println("Found better:", result)
		}
		fmt.Println("\nOverall result count:", count)
	})
	/*
		t.Run("parse GOBL.gz", func(t *testing.T) {
			goblFileBetterThanGB = func(k int) string {
				return fmt.Sprintf("testdata/snapshot/TestAllKAssetPortfolios_PortfolioStats_k%d_betterThanGoldenButterfly.gobl.gz", k)
			}
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
	*/
}

func EncodeResultsToSQLite(sqliteFile string, results chan *pa.PortfolioStat) error {
	db, err := sql.Open("sqlite3", sqliteFile+"?mode=rwc")
	if err != nil {
		return err
	}
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
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
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
	if err != nil {
		return err
	}
	var totalRows int
	for stat := range results {
		returnsList := data.PortfolioReturnsList(stat.Assets...)
		returns, err := pa.PortfolioReturns(returnsList, stat.Percentages)
		if err != nil {
			return err
		}
		minPWR10, _ := pa.MinPWR(returns, 10)
		pwrs10 := pa.AllPWRs(returns, 10)
		pwrs30 := pa.AllPWRs(returns, 30)
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
			pa.StandardDeviation(pwrs10).Float(),
			pa.Slope(pwrs10).Float(),
			pa.StandardDeviation(pwrs30).Float(),
			pa.Slope(pwrs30).Float(),
		)
		if err != nil {
			return err
		}
		totalRows++
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if err := stmt.Close(); err != nil {
		return err
	}

	fmt.Println("Wrote total rows:", totalRows)
	return nil
}

func CountResults(resultsCh <-chan *pa.PortfolioStat) {
	count := 0
	for range resultsCh {
		count++
		// fmt.Println("Found:", result)
	}
	fmt.Println("\nOverall result count:", count)
}

func mustDelete(t *testing.T, names map[string]struct{}, name string) {
	t.Helper()
	g := NewGomegaWithT(t)
	g.Expect(names).To(HaveKey(name))
	delete(names, name)
}
