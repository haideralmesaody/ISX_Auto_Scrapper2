# ISX Auto Scrapper - Comprehensive Testing Plan

## Prerequisites
- Auto mode has been completed successfully
- All necessary files are in place (TICKERS.csv, raw_*.csv files, etc.)

## Phase 1: Auto Mode Results Verification

### 1.1 Check Generated Files
```bash
# List all generated files
dir raw_*.csv
dir indicators_*.csv
dir liquidity_scores.csv
dir Strategy_Summary.json
dir Processing_Report_*.csv
dir Timing_Analysis_*.csv
```

### 1.2 Validate Data Integrity
```bash
# Check file sizes (should not be empty)
for %f in (raw_*.csv) do echo %f && wc -l "%f"

# Verify CSV headers
head -1 raw_AAHP.csv
head -1 indicators_AAHP.csv
head -1 liquidity_scores.csv
```

### 1.3 Review Processing Reports
```bash
# Open latest processing report
notepad Processing_Report_*.csv
# Check for:
# - Status: SUCCESS/ERROR/PARTIAL/UP_TO_DATE
# - Error messages
# - Data quality scores
# - Processing times
```

## Phase 2: Individual Mode Testing

### 2.1 Single Mode Test
```bash
# Test single ticker fetch
./isx-auto-scrapper.exe --mode single
# Enter ticker: AAHP
# Expected: Creates/updates raw_AAHP.csv
```

**Validation Steps:**
- [ ] Check if raw_AAHP.csv is created/updated
- [ ] Verify data has recent dates
- [ ] Check for proper OHLC values
- [ ] Validate volume data

### 2.2 Calculate Mode Test
```bash
# Test indicator calculations with descriptions
./isx-auto-scrapper.exe --mode calculate AAHP
# Expected: Creates indicators_AAHP.csv with full descriptions
```

**Validation Steps:**
- [ ] Verify indicators_AAHP.csv is created
- [ ] Check all technical indicators are calculated (SMA, EMA, RSI, MACD, etc.)
- [ ] Verify description columns are populated
- [ ] Check crossover signals (Golden Cross, Death Cross)
- [ ] Validate boolean flags

### 2.3 Calculate_Num Mode Test
```bash
# Test numerical indicators only
./isx-auto-scrapper.exe --mode calculate_num AAHP
# Expected: Creates Indicators2_AAHP.csv without descriptions
```

**Validation Steps:**
- [ ] Verify Indicators2_AAHP.csv is created
- [ ] Compare numerical values with indicators_AAHP.csv
- [ ] Ensure no description columns present
- [ ] Check file size is smaller than full indicators file

### 2.4 Calculate_Num All Tickers Test
```bash
# Test batch numerical calculations
./isx-auto-scrapper.exe --mode calculate_num
# Expected: Creates Indicators2_*.csv for all tickers
```

**Validation Steps:**
- [ ] Count Indicators2_*.csv files matches ticker count
- [ ] Verify processing completes without errors
- [ ] Check file timestamps are recent

### 2.5 Liquidity Mode Test
```bash
# Re-run liquidity calculations
./isx-auto-scrapper.exe --mode liquidity
# Expected: Updates liquidity_scores.csv
```

**Validation Steps:**
- [ ] Verify liquidity_scores.csv is updated
- [ ] Check liquidity scores are calculated for all tickers
- [ ] Validate score ranges (0-100)
- [ ] Verify ranking order
- [ ] Check volume consistency scores

## Phase 3: Strategy & Analysis Testing

### 3.1 Strategy Mode Test
```bash
# Test trading strategies
./isx-auto-scrapper.exe --mode strategies
# Expected: Creates strategies_*.csv files and Strategy_Summary.json
```

**Validation Steps:**
- [ ] Verify strategies_*.csv files are created
- [ ] Check Strategy_Summary.json is populated
- [ ] Validate trading signals (BUY/SELL/HOLD)
- [ ] Verify strategy counts and summaries
- [ ] Check alternative strategy states

### 3.2 Simulate Mode Test
```bash
# Test backtesting simulations
./isx-auto-scrapper.exe --mode simulate
# Expected: Generates simulation results and summaries
```

**Validation Steps:**
- [ ] Verify simulation completes without errors
- [ ] Check performance metrics are calculated
- [ ] Validate return calculations
- [ ] Verify risk metrics (Sharpe ratio, max drawdown)
- [ ] Check trade statistics

### 3.3 Additional Verification
```bash
# Verify all generated files are accessible
dir Indicators2_*.csv
dir strategies_*.csv
# Expected: All files are properly formatted and readable
```

**Validation Steps:**
- [ ] Verify all CSV files can be opened
- [ ] Check file sizes are reasonable
- [ ] Validate data formats are consistent
- [ ] Ensure no corrupted files

## Phase 4: Edge Case Testing

### 4.1 Missing Data Test
```bash
# Test with non-existent ticker
./isx-auto-scrapper.exe --mode single
# Enter ticker: INVALID
# Expected: Graceful error handling
```

### 4.2 Corrupted File Test
```bash
# Backup and corrupt a file
copy raw_AAHP.csv raw_AAHP.csv.backup
echo "corrupted" > raw_AAHP.csv

# Test calculate mode
./isx-auto-scrapper.exe --mode calculate AAHP
# Expected: Error handling and logging

# Restore file
copy raw_AAHP.csv.backup raw_AAHP.csv
```

### 4.3 Empty Dataset Test
```bash
# Create empty CSV with headers only
echo "Date,Close,Open,High,Low,Change,Change%,Volume,T.Shares,No. Trades" > raw_TEST.csv

# Test calculations
./isx-auto-scrapper.exe --mode calculate TEST
# Expected: Proper error handling
```

## Phase 5: Integration Testing

### 5.1 End-to-End Workflow Test
```bash
# Complete workflow for one ticker
./isx-auto-scrapper.exe --mode single
# Enter: BASH

./isx-auto-scrapper.exe --mode calculate BASH
./isx-auto-scrapper.exe --mode calculate_num BASH
./isx-auto-scrapper.exe --mode report BASH
```

### 5.2 Data Consistency Test
```bash
# Compare indicator values between different calculation modes
# Manual verification required
```

### 5.3 Performance Test
```bash
# Time different operations
measure-command { ./isx-auto-scrapper.exe --mode calculate_num }
measure-command { ./isx-auto-scrapper.exe --mode liquidity }
```

## Validation Checklist

### File Output Validation
- [ ] All expected CSV files are created
- [ ] File sizes are reasonable (not empty, not excessively large)
- [ ] Headers match expected format
- [ ] Data types are correct (dates, decimals, integers)

### Data Quality Validation
- [ ] No missing dates in sequence
- [ ] OHLC values are logical (High >= Low, etc.)
- [ ] Volume values are non-negative
- [ ] Indicator values are within expected ranges
- [ ] Crossover signals are logically correct

### Performance Validation
- [ ] Processing times are reasonable
- [ ] Memory usage is acceptable
- [ ] No memory leaks during long operations
- [ ] Error recovery works properly

### Data Output Validation
- [ ] CSV files are properly formatted
- [ ] All data columns are present
- [ ] Values are within expected ranges
- [ ] Files can be imported into analysis tools

## Troubleshooting Guide

### Common Issues and Solutions

1. **File Not Found Errors**
   - Check if TICKERS.csv exists
   - Verify raw_*.csv files are present
   - Ensure proper file permissions

2. **Calculation Errors**
   - Verify input data quality
   - Check for sufficient data points
   - Validate date formats

3. **Report Generation Issues**
   - Check if indicators_*.csv exists
   - Verify template files are present
   - Check file permissions for output directory

4. **Performance Issues**
   - Monitor system resources
   - Check for large file sizes
   - Verify efficient processing

## Test Result Documentation

Create a test results log with:
- Test case name
- Expected result
- Actual result
- Pass/Fail status
- Notes/Issues found

Example:
```
Test: Single Mode AAHP
Expected: raw_AAHP.csv created with recent data
Actual: File created, 2000+ records, data up to current date
Status: PASS
Notes: Processing took 45 seconds
```

## Final Validation

After completing all tests:
1. [ ] All modes work without errors
2. [ ] Data integrity is maintained across operations
3. [ ] Reports generate correctly
4. [ ] Performance is acceptable
5. [ ] Error handling works properly
6. [ ] File outputs are consistent and complete 