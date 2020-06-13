# portfolio-analysis
Analyzing Portfolios, inspired by PortfolioCharts.com

## Golden Butterfly Portfolio
The [Golden Butterfly Portfolio](https://portfoliocharts.com/portfolio/golden-butterfly/)
is one of the most compelling portfolios, with
an unusually high perpetual withdrawal rate, and
an unusually low ulcer index.

It's a simple mix of 20% each of:
- Total Stock Market Index
- Small Cap Value Stocks Index
- Long Term Treasuries
- Short Term Treasuries
- Gold

### Verifying the Advertised Performance
In this repo I wrote some Go code to recreate most of the portfolio metrics
used on PortfolioCharts.com.

Frustratingly, the historical data used on the PortfolioCharts website is not accessible,
so it's hard to verify what is advertised.


I want to see how portfolios will perform using real ETFs that I have access to.
I found a wonderful resource, [Simba's backtesting spreadsheet](https://www.bogleheads.org/wiki/Simba%27s_backtesting_spreadsheet), 
containing inflation-adjusted return data for many Vanguard funds.

I loaded the historical data for the five Golden Butterfly assets from 1969.

Here are the results of evaluating that portfolio:

| Metrics  | Golden Butterfly using Vanguard Funds | Golden Butterfly as Advertised on PortfolioCharts.com  |
|---|---:|---:|
| Average Return             |  5.668% |      6.4% |
| Baseline Return (15years)  |  5.241% |      5.5% | 
| Baseline Return (3years)   |  2.848% |      2.8% |    
| Perpetual Withdrawal Rate  |  4.224% |      5.3% | 
| Safe Withdrawal Rate       |  5.305% |      6.4% |    
| Standard Deviation         |  8.103% |      7.9% | 
| Ulcer Index                |     3.4 |       2.7 |    
| Deepest Drawdown           | -15.33% |      -11% | 
| Longest Drawdown           | 3 years | 2.8 years |    
| Start Date Sensitivity     |   7.71% |      6.7% | 

It was a little disappointing to see the diminished performance of the real portfolio compared to
what was advertised on portfolio charts.
Notably:
- The Perpetual Withdrawal Rate was a full percentage point lower: 4.2% rather than 5.3%!
- The Deepest Drawdown was also a little scarier: -15% rather than -11%.

Still, the portfolio still seems solid, and likely one of the best.

### Better than Golden Butterfly?
#### Same assets, different allocation
I tried experimenting with variations on the equal-parts 20% allocation, 
trying all possible ways to combine the five assets, in 0.1% increments.

Somewhat incredibly, NONE of the variations could perform as well as or better than the 
standard Golden Butterfly in ALL of the ten metrics.

The variations could only perform better in one or more metrics, at the expense of other metrics.

The Golden Butterfly is amazingly robust, and tough to beat! 

### Any other assets?
Is there any combination of assets that can perform as well as or better than Golden Butterfly
in ALL of the given metrics?

The Simba spreadsheet has a wealth of other assets, we can try it out!

I was able to evaluate over 4 billion portfolios (all combinations of up to 9 assets!).
Out of that, 0.2% of them were actually as good or better than the Golden Butterfly
portfolio in all of the ten metrics.

That is over **8 million portfolios** that are "on paper" better than Golden Butterfly!

Now:
- How to sort through them and find one we would really like better?
- Can we refine the portfolios even further by tweaking their allocation percentages
(they were all simply equally-weighted, 1/N portfolios).

To be determined...