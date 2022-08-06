package v2

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	pa "github.com/slatteryjim/portfolio-analysis"
	"github.com/slatteryjim/portfolio-analysis/data"
	. "github.com/slatteryjim/portfolio-analysis/types"
)

// GoFindKAssetsBetterThanX will spin up multiple goroutines to look at all `k` combination of the given names.
// Any combinations that have stats better than the given ideal will be written to the returned channel.
// When all combinations have been evaluated, the returned channel will be closed.
func GoFindKAssetsBetterThanX(ideal *pa.PortfolioStat, k int, names []string) <-chan *pa.PortfolioStat {
	var resultsCh = make(chan *pa.PortfolioStat, 10)
	go func() {
		defer close(resultsCh)
		startAt := time.Now()
		// look at all `k` combinations of assets
		targetAllocations := make([]Percent, k)
		for i := 0; i < k; i++ {
			targetAllocations[i] = Percent(1.0 / float64(k))
		}
		nCr := pa.Binomial(len(names), k)
		fmt.Println()
		fmt.Println(time.Now(), "k =", k, "nCr =", nCr, "TargetAllocations", targetAllocations)

		GoEvaluateAndFindBetterThan := func(assetCombinationBatches <-chan [][]string) <-chan *pa.PortfolioStat {
			out := make(chan *pa.PortfolioStat, 10)
			go func() {
				defer close(out)
				for batch := range assetCombinationBatches {
					for _, assets := range batch {
						returnsList := data.PortfolioReturnsList(assets...)
						returns, err := pa.PortfolioReturns(returnsList, targetAllocations)
						if err != nil {
							panic(err.Error())
						}
						combination := pa.Combination{Assets: assets, Percentages: targetAllocations}
						var stat *pa.PortfolioStat
						if ideal != nil {
							stat = pa.EvaluatePortfolioIfAsGoodOrBetterThan(returns, combination, ideal)
						} else {
							stat = pa.EvaluatePortfolio(returns, combination)
						}
						if stat != nil {
							out <- stat
						}
					}
				}
			}()
			return out
		}

		combinationsCh := pa.GoEnumerateCombinations(names, k, 10_000)
		// fan out to multiple workers, 9 workers was a sweet spot
		var workersOutput []<-chan *pa.PortfolioStat
		for i := 0; i < 9; i++ {
			results := GoEvaluateAndFindBetterThan(combinationsCh)
			workersOutput = append(workersOutput, results)
		}
		// merge workers' output
		for result := range pa.GoMerge(workersOutput...) {
			resultsCh <- result
		}

		elapsed := time.Since(startAt)
		fmt.Printf("Finished evaluating %d portfolios in %v (%d portfolios per second)\n",
			nCr, elapsed, int(float64(nCr)/elapsed.Seconds()))
	}()
	return resultsCh
}

func EncodeResultsToSQLite(sqliteFile string, results <-chan *pa.PortfolioStat) error {
	db, err := sql.Open("sqlite3", sqliteFile+"?mode=rwc")
	if err != nil {
		return err
	}
	defer db.Close()

	// create table
	sqlStmt := `
		DROP TABLE IF EXISTS portfolios;
		CREATE TABLE IF NOT EXISTS portfolios (
			assets                TEXT NOT NULL,
			percentages           TEXT NOT NULL,
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
			pwr5                  REAL,
			pwr10                 REAL,
			pwr10_stdev           REAL,
			pwr10_slope           REAL,
			pwr30_stdev           REAL,
			pwr30_slope           REAL,
			percent_tsm           REAL,
			percent_scv           REAL,
			percent_ltt           REAL,
			percent_stt           REAL,
			percent_gold          REAL,
			percent_reit          REAL
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
				percentages,                    
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
				pwr5,
				pwr10,
				pwr10_stdev,
				pwr10_slope,
				pwr30_stdev,
				pwr30_slope,
				percent_tsm,
				percent_scv,
				percent_ltt,
				percent_stt,
				percent_gold,
				percent_reit
			)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		minPWR5, _ := pa.MinPWR(returns, 5)
		pwrs10 := pa.AllPWRs(returns, 10)
		pwrs30 := pa.AllPWRs(returns, 30)
		percentTSM, _ := stat.Percentage("TSM")
		percentSCV, _ := stat.Percentage("SCV")
		percentLTT, _ := stat.Percentage("LTT")
		percentSTT, _ := stat.Percentage("STT")
		percentGold, _ := stat.Percentage("Gold")
		percentREIT, _ := stat.Percentage("REIT")
		_, err = stmt.Exec(
			"|"+strings.Join(stat.Assets, "|")+"|",               // encode as string
			"|"+strings.Join(Strings(stat.Percentages), "|")+"|", // encode as string
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
			minPWR5.Float(),
			minPWR10.Float(),
			pa.StandardDeviation(pwrs10).Float(),
			pa.Slope(pwrs10).Float(),
			pa.StandardDeviation(pwrs30).Float(),
			pa.Slope(pwrs30).Float(),
			percentTSM.Float(),
			percentSCV.Float(),
			percentLTT.Float(),
			percentSTT.Float(),
			percentGold.Float(),
			percentREIT.Float(),
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

func Strings[T fmt.Stringer](values []T) []string {
	var strings []string
	for _, value := range values {
		strings = append(strings, value.String())
	}
	return strings
}
