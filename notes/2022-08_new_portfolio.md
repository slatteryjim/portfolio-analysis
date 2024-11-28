### New Portfolio - 2022 August

Playing around with portfolio allocations (see the [2022-08 notebooks here](../v2/)), I [found a new one](../v2/2022-08-07_refine_GB+REIT_allocations.ipynb) that I liked:
- 22% TSM (Total Stock Market)
- 23% SCV (Small-cap Value)
- 10% LTT (Long-term Treasuries)
- 25% REIT (Real Estate Investment Trust)
- 20% GLD (Gold)

Here were the stats from [that run](../v2/2022-08-07_refine_GB+REIT_allocations.ipynb):

| assets                 | \|TSM\|SCV\|LTT\|Gold\|REIT\| |
| ---------------------- | ----------------------------- |
| percentages            | \|22%\|23%\|10%\|20%\|25%\|   |
| num\_assets            | 5                             |
| num\_years             | 50                            |
| avg\_return            | 0.081298                      |
| baseline\_lt\_return   | 0.064707                      |
| baseline\_st\_return   | 0.036117                      |
| pwr30                  | 0.057678                      |
| swr30                  | 0.066947                      |
| std\_dev               | 0.111316                      |
| ulcer\_score           | 3.434891                      |
| deepest\_drawdown      | -0.224129                     |
| longest\_drawdown      | 3                             |
| startdate\_sensitivity | 0.089617                      |
| pwr5                   | 0.012002                      |
| pwr10                  | 0.040567                      |
| pwr10\_stdev           | 0.013823                      |
| pwr10\_slope           | -0.022336                     |
| pwr30\_stdev           | 0.009041                      |
| pwr30\_slope           | -0.040522                     |

You can plug those numbers into the [PortfolioCharts.com](https://portfoliocharts.com/) tools:
- the [My Portfolio](https://portfoliocharts.com/portfolio/my-portfolio/) tool to see its stats 
([shortcode here](https://portfoliocharts.com/portfolio/my-portfolio/?shortcode=ZUSAXAA22AI23AK10AO25GB20Z))
- or the [Portfolio Matrix](https://portfoliocharts.com/portfolio/portfolio-matrix/)
to compare it against other portfolios
([shortcode here](https://portfoliocharts.com/portfolio/portfolio-matrix/?shortcode=ZUSAXAA22AI23AK10AO25GB20Z))
.

Here are some highlights vs. Golden Butterfly:

| Metrics                                             | New Portfolio | Golden Butterfly |
| --------------------------------------------------- | ------------: | ---------------: |
| Minimum 30-year PWR<br/>(Perpetual Withdrawal Rate) |          5.9% |             5.1% |
| Baseline Long-Term Return                           |          6.8% |             5.6% |
| Average Return                                      |          8.2% |             6.3% |
| Ulcer index                                         |           5.2 |              2.5 |

I wanted to explore alternatives to the Golden Butterfly portfolio with reduced exposure to Long-Term Treasuries
(see Ray Dalio's article: [Why in the World Would You Own Bonds When...](https://www.linkedin.com/pulse/why-world-would-you-own-bonds-when-ray-dalio/)).

I found REIT turned out to be a strong asset to include.

And surprisingly, this new portfolio entirely dropped the "cash" asset, STT (Short-term Treasuries),
yet its 10-year minimum PWR was 4%, as high as Golden Butterfly!

Here are a few screenshots showing the new portfolio's performance:

The [Portfolio Matrix](https://portfoliocharts.com/portfolio/portfolio-matrix/?shortcode=ZUSAXAA22AI23AK10AO25GB20Z)
ranking the portfolios by Perpetual Withdrawal Rate:
> ![](portfoliocharts-newPortfolio2022-08_portfolioMatrix.png)

Its standard deviation ranking is lower, but I don't mind the increased volatility, 
especially when the ulcer score is still so low.

Here's a tall screenshot comparing the new portfolio on the left, to the Golden Butterfly portfolio on the right:
> Note the different y-axis scales in several graphs, their 
> differences are larger than they appear: Portfolio Growth, Rolling Returns, and Retirement Spending.
> ![](portfoliocharts-compareNewVsGoldenButterfly.png)

[comment]: # (Command line to generate animated GIF:)
[comment]: # (  convert -delay 160 -loop 0 ?.png output.gif)

And [here's a tall animated GIF](portfoliocharts-compareNewVsGoldenButterfly_animation.gif) switching between the portfolios for easier comparison.