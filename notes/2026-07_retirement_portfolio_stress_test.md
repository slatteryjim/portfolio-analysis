# Retirement Portfolio Review and Stress Test — 2026 July

This note evaluates whether the [2022 portfolio](2022-08_new_portfolio.md) is a
reasonable foundation for a 50-year retirement with annual rebalancing:

- 22% Total Stock Market (TSM)
- 23% Small-Cap Value (SCV)
- 10% Long-Term Treasuries (LTT)
- 25% Real Estate Investment Trusts (REIT)
- 20% Gold

All results use inflation-adjusted annual returns from Simba Rev25c through 2025.
`PortfolioReturns` rebalances to the target allocation annually. PWR is this repo's
finite-period principal-preservation calculation; it is not proof that a withdrawal
can continue forever.

## Executive conclusion

The portfolio is historically excellent and more robust to nearby allocation changes
than might be expected. It performed well through multiple inflation, recession,
stock-market, and interest-rate regimes. It also compares favorably with the famous
portfolios examined below.

That does **not** make its 5.66% historical 30-year PWR a prudent retirement spending
rate. The portfolio was selected from billions of combinations using substantially the
same history now used to evaluate it. Taxes, costs, personal spending shocks, and a
future worse than the sample are absent. A 30-year result is also not a 50-year or
indefinite guarantee.

A planning range around 2.75% fixed-real, or roughly 3.0%-3.25% with genuine spending
guardrails and dependable income covering essential expenses, remains much more
defensible. Portfolio construction should complement an inflation-protected spending
floor, adequate insurance, and a willingness to reduce discretionary spending after
poor markets.

## Updated base results

The portfolio has 54 annual observations from 1972 through 2025:

| Metric | Result |
| --- | ---: |
| Average real return | 7.72% |
| 15-year baseline return | 6.44% |
| 10-year minimum PWR | 4.02% |
| 30-year minimum PWR | 5.66% |
| 50-year minimum PWR | 5.92% |
| 30-year minimum SWR | 6.59% |
| Standard deviation | 11.48% |
| Ulcer index | 3.57 |
| Deepest drawdown | -22.42% |
| Longest drawdown | 3 years |
| Start-date sensitivity | 9.04% |

The 50-year figure is especially weak evidence: 54 years contain only five overlapping
50-year periods. The 30-year result has only 25 overlapping periods, which are also far
from independent.

## Comparison with famous portfolios

Each portfolio below uses its maximum available history, so the rows are useful regime
stress tests but are not perfectly apples-to-apples. The [Portfolio Charts Portfolio
Matrix](https://portfoliocharts.com/charts/portfolio-matrix/) similarly emphasizes that
every allocation has tradeoffs.

| Portfolio | History | Avg real | 15yr baseline | 10yr PWR | 30yr PWR | Std dev | Ulcer | Deepest DD |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 2022 portfolio | 1972-2025 | **7.72%** | **6.44%** | **4.02%** | **5.66%** | 11.48% | 3.57 | -22.42% |
| [Pinwheel](https://portfoliocharts.com/portfolios/pinwheel-portfolio/?home=usa) | 1976-2025 | 6.69% | 4.84% | 2.51% | 4.75% | 10.24% | 3.42 | -23.31% |
| [Swensen](https://portfoliocharts.com/portfolios/swensen-portfolio/), actual TIPS | 1985-2025 | 7.33% | 4.78% | 2.00% | 5.39% | 11.19% | 4.42 | -25.57% |
| Swensen, Portfolio Charts bond proxy | 1976-2025 | 6.93% | 4.81% | 2.03% | 5.41% | 10.42% | 4.34 | -23.00% |
| [Core Four](https://portfoliocharts.com/portfolios/core-four-portfolio/) | 1972-2025 | 6.66% | 4.02% | -1.11% | 3.39% | 13.22% | 22.98 | -41.45% |
| [Three Fund](https://portfoliocharts.com/portfolios/three-fund-portfolio/) | 1970-2025 | 6.81% | 3.75% | -0.94% | 3.27% | 13.59% | 23.32 | -43.56% |
| [Classic 60/40](https://portfoliocharts.com/portfolios/classic-60-40-portfolio/) | 1871-2025 | 5.96% | 2.15% | -4.04% | 2.33% | 11.79% | 23.72 | -41.01% |
| Golden Butterfly | 1969-2025 | 5.54% | 4.91% | 1.97% | 4.18% | 8.37% | **3.35** | **-17.61%** |
| [Permanent Portfolio](https://portfoliocharts.com/portfolios/permanent-portfolio/) | 1969-2025 | 4.48% | 3.65% | 1.36% | 3.56% | **7.75%** | 3.61 | -18.70% |

The 2022 portfolio's lead is broad rather than limited to average return. Golden
Butterfly remains the best defensive comparison, while Pinwheel is the most natural
globally diversified alternative with smaller tilts to SCV, REIT, and gold.

## Allocation perturbation

To test whether 22/23/10/25/20 is a fragile optimum, every asset was moved up, down, or
left unchanged while requiring the changes to net to zero and preserving positive
allocations.

### Five-point neighborhood

There are 51 valid combinations when each asset may move by 5 percentage points.

| Metric | Worst nearby result | Best nearby result | Base |
| --- | ---: | ---: | ---: |
| Average real return | 7.33% | 8.10% | 7.72% |
| 15-year baseline | 5.72% | 6.71% | 6.44% |
| 10-year PWR | 3.29% | 4.06% | 4.02% |
| 30-year PWR | 5.20% | 5.83% | 5.66% |
| 50-year PWR | 5.45% | 6.39% | 5.92% |
| Standard deviation | 10.59% | 12.72% | 11.48% |
| Ulcer index | 2.80 | 5.74 | 3.57 |
| Deepest drawdown | -29.54% | -19.48% | -22.42% |

The lowest 30-year PWR was 5.20% at 17% TSM / 18% SCV / 15% LTT / 25% REIT /
25% gold. The favorable metrics therefore do not depend on hitting the original weights
exactly.

### Ten-point neighborhood

Across 35 valid broader perturbations, 30-year PWR ranged from 4.61% to 5.80%.
The worst PWR allocation was 22% TSM / 13% SCV / 20% LTT / 15% REIT / 30% gold.
Worst drawdown across the neighborhood reached -34.55%, and the ulcer index reached
11.75. The design remains viable, but larger shifts can materially weaken its behavior.

These ranges report the minimum and maximum of each metric independently; the extrema
do not generally belong to one portfolio.

## Asset-return stress tests

The following deterministic tests reduce annual real returns by fixed percentage points
or remove a historical diversification benefit. They are deliberately severe and do not
model changing volatility or correlations.

| Scenario | Avg real | 15yr baseline | 10yr PWR | 30yr PWR | 50yr PWR | Ulcer | Deepest DD |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Base history | 7.72% | 6.44% | 4.02% | 5.66% | 5.92% | 3.57 | -22.42% |
| Every asset loses 1 point/year | 6.72% | 5.43% | 3.09% | 4.85% | 5.12% | 4.15 | -24.18% |
| TSM, SCV, and REIT lose 2 points/year | 6.32% | 5.03% | 2.71% | 4.52% | 4.79% | 4.38 | -24.88% |
| LTT loses 1 and gold loses 2 points/year | 7.22% | 5.94% | 3.56% | 5.25% | 5.52% | 3.86 | -23.30% |
| Combined haircuts: TSM -1; all others -2 except LTT -1 | 6.04% | 4.75% | 2.45% | 4.29% | 4.55% | 4.54 | -25.37% |
| SCV behaves like TSM | 7.41% | 5.94% | 2.79% | 4.82% | 5.13% | 3.85 | -23.36% |
| REIT behaves like TSM | 7.84% | 5.91% | 2.09% | 5.25% | 5.49% | 3.70 | -21.63% |
| Gold earns 0% real every year | 6.36% | 5.38% | 1.28% | 4.22% | 4.43% | 9.54 | -37.56% |

The portfolio survives every test in the narrow sense that long-horizon PWR remains
positive, but its historical excellence is meaningfully dependent on gold. Setting gold
to a constant zero real return causes the largest deterioration in short-horizon PWR,
ulcer index, and drawdown. Removing the SCV premium also reduces 30-year PWR by about
0.84 percentage points.

Even the combined haircut leaves a historical 30-year PWR of 4.29%, but this should not
be converted directly into a recommended withdrawal rate. Fixed haircuts cannot produce
new crises, correlation changes, tax surprises, or spending shocks.

## Exploring 100+ years of data

This is worth doing, with an important limitation: the 100+ year universe contains many
closely related or duplicated series, not a broad collection of independent assets.

The usable economic building blocks with 1871-2025 histories are primarily:

- TSM or LCB stocks (substantially overlapping choices)
- Total bonds
- Long-, intermediate-, and short-term Treasuries
- Intermediate and short-term broad bonds
- Intermediate corporate bonds
- T-bills or cash

Other 155-year columns are total-return, price-return, or reference variants of the same
assets. There is no 100-year history in this sheet for gold, REIT, international stocks,
small value, commodities, or TIPS.

An exhaustive 5%-increment search of TSM/LTT/ITT/STT over all 155 years produced:

| Objective | Best allocation (TSM/LTT/ITT/STT) | Result |
| --- | ---: | ---: |
| Maximum 30-year PWR | 70/0/0/30 | 2.64% |
| Maximum 50-year PWR | 80/0/10/10 | 3.18% |
| Maximum 15-year baseline | 80/0/15/5 | 2.65% |
| Lowest ulcer index | 50/0/0/50 | 18.47 |

Even the best century portfolios had a negative 10-year PWR near -3.5% to -4.0%, meaning
that their worst real decade lost so much purchasing power that preserving the starting
principal would have required contributions rather than withdrawals. The long sample
contains severe regimes that disappear from post-1970 comparisons.

This supports two conclusions:

1. Long history is extremely valuable for exposing stock/bond portfolios to hostile
   inflation and valuation regimes.
2. It cannot by itself validate the newer diversifiers, because substituting assets that
   did not exist changes the economic question.

The century dataset should be used as a harsh stock/bond baseline and as a source of
historical macro regimes—not as evidence that the post-1970 assets would have produced
their later returns in the nineteenth or early twentieth century.

## Remaining tests worth implementing

### Walk-forward selection

At several historical cutoffs, search for portfolios using only data available at that
time and score them only on subsequent years. This directly attacks the selection bias
created by the original multi-billion-portfolio search.

### Leave-one-regime-out analysis

Exclude a major regime such as the 1970s or 2000s during selection, then test the chosen
portfolio on the omitted years. A sound economic design should not require seeing the
test crisis in advance.

### Block bootstrap and synthetic regimes

Resample multi-year blocks so inflation, returns, and correlations retain some serial
structure. Add explicit regimes in which stocks, REIT, gold, and bonds become more
correlated, rather than merely reducing their average returns independently.

### Withdrawal-policy simulation

Compare fixed-real spending with guardrails, skipped inflation adjustments after losing
years, Social Security timing, and an essential-spending floor. Portfolio failure and
spending failure should be measured separately.

### Taxes, fees, and account placement

Apply fund expenses, turnover, taxable distributions, and realistic account locations.
Large REIT, gold, and annual-rebalancing positions can behave differently after tax than
the frictionless index series.

### International validation

Test the design using non-US home-country returns and inflation. That is a stronger test
of the economic concept than adding more variations to the same US history.

## Practical interpretation

The heavy assault does not invalidate the portfolio. It instead identifies the right
questions:

- The allocation is not a knife-edge optimum; nearby allocations remain strong.
- Gold is the most consequential historical diversifier and therefore the largest model
  risk.
- The small-value premium matters, but losing it does not destroy the portfolio.
- REIT's distinct historical behavior improves short-horizon consistency more than its
  average return alone suggests.
- A century of stock/bond history argues for conservative withdrawal assumptions even
  when a post-1970 diversified backtest looks outstanding.
- The next priority is genuine out-of-sample and regime testing, not further optimization
  of the target percentages.

The portfolio remains a leading candidate, but the retirement plan should be constructed
so that it can survive even if its historical 5.66% PWR proves far too optimistic.
