# ISX Auto Scrapper - Mode Reference Guide

## Quick Mode Overview

| Mode | Purpose | Input | Output | Use When |
|------|---------|-------|--------|----------|
| `web` | Interactive dashboard | Existing data files | Web interface | Real-time analysis and visualization |
| `auto` | Complete data pipeline | TICKERS.csv | All files | Full analysis of all stocks |
| `single` | Fetch one ticker | User input | raw_*.csv | Testing or single stock update |
| `calculate` | Full indicators + descriptions | raw_*.csv | indicators_*.csv | Need detailed analysis with explanations |
| `calculate_num` | Numerical indicators only | raw_*.csv | Indicators2_*.csv | Performance analysis or data processing |
| `liquidity` | Volume analysis | raw_*.csv | liquidity_scores.csv | Assess market liquidity |
| `strategies` | Trading signals | indicators_*.csv | strategies_*.csv, Strategy_Summary.json | Generate trading recommendations |
| `simulate` | Backtest strategies | strategies_*.csv | Performance reports | Test strategy effectiveness |


## Detailed Mode Descriptions

### 🌐 `web` Mode - Interactive Dashboard
```bash
./isx-auto-scrapper.exe --mode web [port]
# Default port: 8080
# Custom port example: ./isx-auto-scrapper.exe --mode web 3000
```

**What it does:**
- Starts an HTTP web server with interactive dashboard
- Provides real-time data visualization and analysis
- Serves professional candlestick charts with multiple timeframes
- Displays technical indicators and trading signals
- Integrates backtest results and market analytics
- Auto-refreshes data every 5 minutes

**Dashboard Features:**

🎯 **Main Interface:**
- **Ticker Selection**: Search and select from all available tickers
- **Interactive Charts**: Candlestick and line charts with zoom/pan
- **Multiple Timeframes**: 1D, 1W, 1M, 3M, 6M, 1Y views
- **Real-time Prices**: Live ticker prices with change indicators

📊 **Technical Analysis Panel:**
- **RSI**: Relative Strength Index with bullish/bearish signals
- **MACD**: Moving Average Convergence Divergence
- **OBV**: On-Balance Volume analysis
- **CMF**: Chaikin Money Flow indicator
- **Moving Averages**: EMA and SMA overlays

📈 **Strategy Signals:**
- **Current Recommendations**: Buy/Sell/Hold signals
- **Signal Strength**: Strong, Weak, and neutral classifications
- **Strategy Performance**: Real-time strategy rankings
- **Market Overview**: Top movers and market statistics

🔧 **Interactive Controls:**
- **Refresh Data**: Manual data refresh button
- **Run Backtest**: Trigger backtesting from web interface
- **Chart Types**: Switch between candlestick and line charts
- **Search Functionality**: Filter tickers by symbol

**Prerequisites:**
- Existing data files in working directory:
  - `raw_*.csv` - Price data
  - `indicators_*.csv` - Technical indicators
  - `Strategies_*.csv` - Strategy signals
  - `Strategy_Summary.json` - Strategy summaries
  - `TICKERS.csv` - Ticker list

**API Endpoints:**
- `GET /api/tickers` - List all tickers with current prices
- `GET /api/ticker/[SYMBOL]?type=price` - Price data for ticker
- `GET /api/ticker/[SYMBOL]?type=indicators` - Technical indicators
- `GET /api/strategies` - Strategy summary data
- `POST /api/backtest` - Trigger backtesting
- `POST /api/refresh` - Refresh all data

**Web Server Details:**
- **Technology**: Go HTTP server with CORS enabled
- **Static Files**: HTML, CSS, JavaScript served from `web/` directory
- **Charts**: Chart.js with financial plugin for candlestick charts
- **Responsive Design**: Works on desktop, tablet, and mobile
- **Modern UI**: Dark theme with professional styling

**Browser Compatibility:**
- Chrome/Edge 90+
- Firefox 88+
- Safari 14+
- Mobile browsers supported

**Network Access:**
```bash
# Local access only
http://localhost:8080

# Allow network access (if firewall permits)
http://[YOUR_IP]:8080
```

**Use When:**
- Interactive market analysis
- Real-time trading signal monitoring
- Presentation and demonstration
- Client reports and meetings
- Educational purposes
- Portfolio monitoring

**Performance Notes:**
- Dashboard updates automatically every 5 minutes
- Chart rendering is optimized for large datasets
- API responses are cached for performance
- Supports concurrent users
- Minimal memory footprint

**Security Considerations:**
- Web server binds to all interfaces (0.0.0.0)
- No authentication required (local use)
- Consider firewall rules for network access
- HTTPS not enabled by default

---

### 🚀 `auto` Mode - The Complete Pipeline
```bash
./isx-auto-scrapper.exe --mode auto
```
**What it does:**
- Loads all tickers from TICKERS.csv
- Scrapes data for each ticker from ISX website
- Calculates technical indicators with descriptions
- Generates liquidity scores
- Applies trading strategies
- Creates strategy summaries
- Generates processing and timing reports

**Output Files:**
- `raw_*.csv` - Raw stock data for each ticker
- `indicators_*.csv` - Technical indicators with descriptions
- `liquidity_scores.csv` - Liquidity analysis
- `strategies_*.csv` - Trading strategy results
- `Strategy_Summary.json` - Strategy summaries
- `Processing_Report_*.csv` - Processing statistics
- `Timing_Analysis_*.csv` - Performance metrics

**Use When:**
- First time setup
- Daily/weekly data updates
- Complete market analysis

---

### 📊 `single` Mode - Individual Ticker Fetch
```bash
./isx-auto-scrapper.exe --mode single
# Enter ticker when prompted
```
**What it does:**
- Prompts for ticker symbol
- Scrapes data for that specific ticker
- Updates or creates raw_*.csv file

**Output Files:**
- `raw_[TICKER].csv` - Raw stock data

**Use When:**
- Testing data fetching
- Updating specific ticker
- Troubleshooting scraping issues

---

### 🧮 `calculate` Mode - Full Technical Analysis
```bash
./isx-auto-scrapper.exe --mode calculate [TICKER]
```
**What it does:**
- Loads raw data from raw_*.csv
- Calculates all technical indicators
- Adds human-readable descriptions
- Identifies crossover signals and trends

**Output Files:**
- `indicators_[TICKER].csv` - Complete technical analysis

**Use When:**
- Need detailed technical analysis
- Generating reports for humans
- Educational purposes
- Manual analysis

---

### 📈 `calculate_num` Mode - Numerical Analysis Only
```bash
./isx-auto-scrapper.exe --mode calculate_num [TICKER]
# Or for all tickers:
./isx-auto-scrapper.exe --mode calculate_num
```
**What it does:**
- Calculates technical indicators (numbers only)
- No descriptions or explanations
- Faster processing
- Smaller file sizes

**Output Files:**
- `Indicators2_[TICKER].csv` - Numerical indicators only

**Use When:**
- High-performance analysis
- Automated processing
- Data feeds for other systems
- Memory/storage optimization

---

### 💧 `liquidity` Mode - Market Liquidity Analysis
```bash
./isx-auto-scrapper.exe --mode liquidity
```
**What it does:**
- Analyzes volume patterns for all tickers
- Calculates liquidity scores
- Ranks stocks by liquidity
- Identifies trading activity patterns

**Output Files:**
- `liquidity_scores.csv` - Liquidity rankings and scores

**Use When:**
- Portfolio construction
- Risk assessment
- Market making decisions
- Investment planning

---

### 📈 `strategies` Mode - Trading Signal Generation
```bash
./isx-auto-scrapper.exe --mode strategies
```
**What it does:**
- Applies multiple trading strategies
- Generates BUY/SELL/HOLD signals
- Creates strategy alternatives
- Summarizes strategy performance

**Output Files:**
- `strategies_*.csv` - Strategy signals for each ticker
- `Strategy_Summary.json` - Aggregated strategy results

**Use When:**
- Generating trading signals
- Strategy development
- Market screening
- Investment decisions

---

### 🎯 `simulate` Mode - Strategy Backtesting
```bash
./isx-auto-scrapper.exe --mode simulate
```
**What it does:**
- Runs comprehensive backtesting and Monte Carlo simulation of all trading strategies
- Loads backtesting configuration from `backtest_config.json`
- Loads all strategy signals from `Strategies_*.csv` files  
- Simulates trading for each strategy with realistic:
  - Portfolio management (cash, positions, equity tracking)
  - Position sizing based on configuration
  - Commission and transaction costs
  - Stop-loss and take-profit orders
  - Maximum holding periods
  - Risk management rules
- Calculates comprehensive performance metrics
- Generates detailed trading reports and analysis

**Prerequisites**: 
- Strategy files (`Strategies_*.csv`) must exist (run `--mode strategies` first)
- `backtest_config.json` configuration file (created automatically with defaults)

**Configuration (`backtest_config.json`)**:
```json
{
  "initial_cash": 100000.0,           // Starting capital
  "commission_per_trade": 50.0,       // Fixed commission per trade
  "commission_percent": 0.0025,       // Percentage commission (0.25%)
  "max_positions": 10,                // Maximum concurrent positions
  "position_size_percent": 10.0,      // % of portfolio per position
  "risk_per_trade_percent": 2.0,      // Maximum risk per trade
  "stop_loss_percent": 5.0,           // Stop loss percentage
  "take_profit_percent": 15.0,        // Take profit percentage
  "max_holding_days": 90,             // Maximum days to hold position
  "use_signal_strength": true,        // Use signal strength for position sizing
  "strategies": ["RSI Strategy", ...], // Strategies to backtest
  "tickers": [],                      // Specific tickers (empty = all)
  "benchmark": "TASC"                 // Benchmark ticker for comparison
}
```

**Output Files**:

1. **`backtest_results.csv`** - Summary results for each strategy:
   - Total Return, Win Rate, Max Drawdown
   - Sharpe Ratio, Profit Factor
   - Trade statistics (total, winning, losing)
   - Average trade duration and P&L metrics

2. **`backtest_results.json`** - Same data in JSON format for programmatic access

3. **`backtest_summary.json`** - High-level summary with best/worst performing strategies

4. **`backtest_trades_[Strategy].csv`** - Detailed trade history for each strategy:
   - Entry/exit dates and prices
   - P&L calculations and percentages
   - Hold duration and exit reasons
   - Commission costs

5. **`backtest_portfolio_[Strategy].csv`** - Daily portfolio value history:
   - Cash, equity, and total portfolio value
   - Daily returns and drawdown tracking
   - Active positions count

**Key Metrics Calculated**:

- **Total Return**: Overall portfolio performance percentage
- **Annualized Return**: Return adjusted for time period
- **Win Rate**: Percentage of profitable trades
- **Profit Factor**: Gross profit divided by gross losses
- **Sharpe Ratio**: Risk-adjusted return metric
- **Maximum Drawdown**: Worst peak-to-trough decline
- **Average Trade Duration**: Typical holding period
- **Recovery Factor**: Total return / Maximum drawdown
- **Calmar Ratio**: Annualized return / Maximum drawdown

**Risk Management Features**:
- Position size limits (max positions, % of portfolio)
- Automatic stop-loss and take-profit execution
- Time-based position exits (max holding days)
- Signal strength-based position sizing
- Commission and slippage modeling

**Sample Usage**:
```bash
# Run backtesting with default configuration
./isx-auto-scrapper.exe --mode simulate

# Edit backtest_config.json to customize parameters, then re-run
./isx-auto-scrapper.exe --mode simulate
```

**Example Output Summary**:
```
Strategy: RSI Strategy
Total Return: 15.32%
Win Rate: 62.5%
Max Drawdown: -8.45%
Sharpe Ratio: 1.42
Total Trades: 47
Profit Factor: 1.89
```

**Notes**:
- Backtesting uses realistic trading assumptions (commissions, slippage)
- Results are based on historical data and don't guarantee future performance
- Each strategy is tested independently with the same starting capital
- Signal strength affects position sizing when `use_signal_strength` is enabled
- Stop-loss and take-profit orders are executed at exact levels (no slippage modeled)

**Use When:**
- Strategy validation
- Risk assessment
- Performance analysis
- Strategy optimization

## Typical Workflows

### 🔄 Daily Update Workflow
```bash
# 1. Update all data
./isx-auto-scrapper.exe --mode auto

# 2. Generate additional analysis if needed
./isx-auto-scrapper.exe --mode strategies
```

### 🔍 Individual Stock Analysis
```bash
# 1. Fetch latest data
./isx-auto-scrapper.exe --mode single
# Enter: BASH

# 2. Calculate indicators
./isx-auto-scrapper.exe --mode calculate BASH

# 3. Apply strategies
./isx-auto-scrapper.exe --mode strategies
```

### ⚡ Performance Analysis
```bash
# 1. Generate numerical indicators for all stocks
./isx-auto-scrapper.exe --mode calculate_num

# 2. Calculate liquidity scores
./isx-auto-scrapper.exe --mode liquidity

# 3. Apply strategies
./isx-auto-scrapper.exe --mode strategies
```

### 🧪 Strategy Development
```bash
# 1. Ensure indicators are calculated
./isx-auto-scrapper.exe --mode calculate_num

# 2. Apply strategies
./isx-auto-scrapper.exe --mode strategies

# 3. Backtest performance
./isx-auto-scrapper.exe --mode simulate
```

## File Dependencies

```
TICKERS.csv (input)
    ↓
raw_*.csv (from single/auto)
    ↓
indicators_*.csv (from calculate)
Indicators2_*.csv (from calculate_num)
    ↓
liquidity_scores.csv (from liquidity)
strategies_*.csv (from strategies)
Strategy_Summary.json (from strategies)
```

## Performance Tips

1. **Use `calculate_num` for bulk processing** - Much faster than full calculations
2. **Run `auto` mode during off-market hours** - Reduces website load
3. **Use `single` mode for testing** - Before running full auto mode
4. **Use strategies mode for trading signals** - Generates actionable insights

## Troubleshooting

### Mode Fails to Start
- Check if required input files exist
- Verify TICKERS.csv format
- Check file permissions

### Slow Performance
- Use `calculate_num` instead of `calculate`
- Run during off-peak hours
- Check available memory

### Missing Output Files
- Check logs for error messages
- Verify input data quality
- Ensure sufficient disk space

### Data Inconsistencies
- Re-run `auto` mode to refresh all data
- Check for corrupted CSV files
- Validate date ranges 