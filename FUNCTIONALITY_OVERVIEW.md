# ISX Auto Scrapper - Functionality Overview

This document summarises all major features provided by the **ISX Auto Scrapper** application. Use it as a quick reference to understand what the project can do and how the different modules fit together.

## 1. Command Line Modes

The application exposes several modes via the `--mode` flag. Each mode performs a specific part of the data pipeline:

| Mode | Purpose |
|------|---------|
| `web` | Launch the interactive web dashboard with real‑time charts and API endpoints. |
| `single` | Prompt for a ticker and scrape only that one. Useful for manual updates. |
| `auto` | Run the full end‑to‑end pipeline for every ticker listed in `TICKERS.csv`. Generates raw data, indicators, strategies and reports. |
| `calculate` | Compute all technical indicators **with descriptions** for one or all tickers. |
| `calculate_num` | Compute indicators **without descriptions** (numerical only) for faster processing. |
| `liquidity` | Calculate enhanced liquidity scores for all tickers. |
| `strategies` | Apply trading strategies and generate strategy sheets and summary JSON. |
| `simulate` | Perform comprehensive backtesting/Monte‑Carlo simulation based on `backtest_config.json`. |

## 2. Data Fetching

* Uses `chromedp` to control Microsoft Edge (Chromium) and scrape historical price tables directly from the ISX website.
* Handles pop‑ups and AJAX loads automatically.
* Generates detailed `Processing_Report_*` and `Timing_Analysis_*` CSV files describing pages loaded, rows processed and performance metrics.
* Raw data is stored as `raw_<TICKER>.csv` files.

## 3. Technical Indicator Calculation

* `indicators_calculator.go` and `technical_indicators.go` implement more than 25 indicators such as SMA, EMA, RSI, MACD, Stochastic, CMF, OBV, PSAR, ATR and rolling standard deviation.
* Results with textual descriptions are written to `indicators_<TICKER>.csv`.
* `numerical_indicators_calculator.go` performs the same calculations but skips descriptions. Output is `Indicators2_<TICKER>.csv`.

## 4. Liquidity Analysis

* `liquidity_calculator.go` computes an enhanced liquidity score per ticker based on volume, volatility and trading activity.
* Results are saved in `liquidity_scores.csv`.

## 5. Trading Strategies

* `strategies.go` applies multiple trading strategies including RSI, MACD, CMF, OBV, EMA5+PSAR and rolling standard‑deviation based rules.
* Each strategy produces Buy/Sell/Hold signals with seven strength levels (Strong Buy → Strong Sell).
* Strategy results per ticker are stored in `Strategies_<TICKER>.csv` and summarised across tickers in `Strategy_Summary.json`.

## 6. Backtesting Engine

* The `simulate` mode runs a full portfolio backtest via `StrategyTester` and `BacktestEngine`.
* Parameters such as starting capital, commission, position sizing and risk limits are configured in `backtest_config.json`.
* Generates reports like `backtest_results.csv`, `backtest_results.json`, `backtest_summary.json`, individual trade logs and portfolio value histories.

## 7. Web Dashboard

* Served by `web_server.go` when running `--mode web`.
* Features professional candlestick charts, technical indicators, strategy signals and market overviews.
* Auto-refreshes every five minutes and exposes a REST API:
  - `GET /api/tickers` – list tickers with latest prices
  - `GET /api/ticker/<SYMBOL>?type=price|indicators` – price or indicator data
  - `GET /api/strategies` – strategy summary
  - `POST /api/backtest` – trigger backtesting
  - `POST /api/refresh` – refresh data
* Static assets live under the `web/` directory.

## 8. Logging

* All operations log to console and `stock_analysis.log` via `logger.go`.
* Processing and timing reports aid in troubleshooting and performance tuning.

## 9. Typical Workflow

1. Run `./isx-auto-scrapper.exe --mode auto` to fetch data for all tickers and generate indicators.
2. Optionally run `./isx-auto-scrapper.exe --mode strategies` to compute trading signals.
3. Start the dashboard with `./isx-auto-scrapper.exe --mode web` and explore results in the browser.
4. Use `./isx-auto-scrapper.exe --mode simulate` to backtest strategies.

The repository also contains example CSV outputs and several markdown guides (`README.md`, `MODE_REFERENCE.md`, `DASHBOARD_DEMO.md`, `DROPDOWN_UPDATE.md`, `TESTING_PLAN.md`) for further reference.

---

*Happy Trading!* 
