# Code Structure Guide

This document explains the overall layout of the Go code so that new contributors can quickly find the relevant packages.

```
ISX_Auto_Scrapper2/
├── cmd/               # CLI entry points
│   └── isx-scraper/   # main application
├── internal/          # private application packages
│   ├── common/        # shared utilities, configuration and types
│   ├── scraper/       # web scraping logic
│   ├── indicators/    # indicator calculations
│   ├── liquidity/     # liquidity scoring
│   ├── strategies/    # trading strategies and backtesting
│   └── server/        # HTTP dashboard and API
└── web/               # static assets served by the dashboard
```

Each subpackage only exposes a minimal API to keep dependencies clear.  The `cmd/isx-scraper` folder contains the Cobra-based CLI which wires everything together.

- **common** – logging, configuration and data structures used across the project.
- **scraper** – drives a headless browser via `chromedp` and produces `raw_*.csv` files along with detailed processing reports.
- **indicators** – calculates technical indicators and writes both descriptive and numerical CSVs.
- **liquidity** – derives enhanced liquidity scores from historical price data.
- **strategies** – implements trading strategies and a small backtesting engine.
- **server** – serves the interactive web dashboard and exposes a REST API.

The static files under `web/` are embedded at runtime and include HTML, JavaScript and CSS for the dashboard.
