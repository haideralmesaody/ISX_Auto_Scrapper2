# ISX Auto Scrapper - Automated Testing Script
# Run this script after completing auto mode to test all functionalities

param(
    [string]$TestTicker = "AAHP",
    [switch]$SkipSingle,
    [switch]$Verbose
)

# Colors for output
$SuccessColor = "Green"
$ErrorColor = "Red"
$InfoColor = "Cyan"
$WarningColor = "Yellow"

function Write-TestResult($TestName, $Success, $Message = "") {
    $status = if ($Success) { "PASS" } else { "FAIL" }
    $color = if ($Success) { $SuccessColor } else { $ErrorColor }
    
    Write-Host "[$status] $TestName" -ForegroundColor $color
    if ($Message) {
        Write-Host "      $Message" -ForegroundColor Gray
    }
}

function Test-FileExists($FilePath, $Description) {
    $exists = Test-Path $FilePath
    Write-TestResult "File Check: $Description" $exists "Path: $FilePath"
    return $exists
}

function Get-FileLineCount($FilePath) {
    if (Test-Path $FilePath) {
        return (Get-Content $FilePath | Measure-Object -Line).Lines
    }
    return 0
}

function Test-CSVContent($FilePath, $ExpectedMinLines = 1) {
    if (-not (Test-Path $FilePath)) {
        return $false
    }
    
    $lineCount = Get-FileLineCount $FilePath
    $success = $lineCount -gt $ExpectedMinLines
    Write-TestResult "CSV Content: $(Split-Path $FilePath -Leaf)" $success "$lineCount lines"
    return $success
}

Write-Host "=== ISX Auto Scrapper Testing Suite ===" -ForegroundColor $InfoColor
Write-Host "Test Ticker: $TestTicker" -ForegroundColor $InfoColor
Write-Host ""

# Phase 1: Verify Auto Mode Results
Write-Host "Phase 1: Auto Mode Results Verification" -ForegroundColor $InfoColor
Write-Host "----------------------------------------" -ForegroundColor $InfoColor

$autoModeResults = @{
    "Raw Data Files" = Test-Path "raw_*.csv"
    "Indicators Files" = Test-Path "indicators_*.csv"
    "Liquidity Scores" = Test-FileExists "liquidity_scores.csv" "Liquidity Scores"
    "Strategy Summary" = Test-FileExists "Strategy_Summary.json" "Strategy Summary"
    "Processing Report" = Test-Path "Processing_Report_*.csv"
    "Timing Analysis" = Test-Path "Timing_Analysis_*.csv"
}

# Count files
$rawFiles = Get-ChildItem "raw_*.csv" -ErrorAction SilentlyContinue
$indicatorFiles = Get-ChildItem "indicators_*.csv" -ErrorAction SilentlyContinue

Write-Host "Raw CSV files found: $($rawFiles.Count)" -ForegroundColor Gray
Write-Host "Indicator CSV files found: $($indicatorFiles.Count)" -ForegroundColor Gray

# Phase 2: Individual Mode Testing
Write-Host ""
Write-Host "Phase 2: Individual Mode Testing" -ForegroundColor $InfoColor
Write-Host "--------------------------------" -ForegroundColor $InfoColor

# Test Single Mode (optional)
if (-not $SkipSingle) {
    Write-Host "Testing Single Mode..." -ForegroundColor $InfoColor
    $singleResult = & ".\isx-auto-scrapper.exe" --mode single 2>&1
    if ($LASTEXITCODE -eq 0) {
        $rawFile = "raw_$TestTicker.csv"
        Test-CSVContent $rawFile 100
    } else {
        Write-TestResult "Single Mode" $false "Exit code: $LASTEXITCODE"
    }
}

# Test Calculate Mode
Write-Host "Testing Calculate Mode..." -ForegroundColor $InfoColor
try {
    $calculateResult = & ".\isx-auto-scrapper.exe" --mode calculate $TestTicker 2>&1
    if ($LASTEXITCODE -eq 0) {
        $indicatorFile = "indicators_$TestTicker.csv"
        Test-CSVContent $indicatorFile 100
        
        # Check for key columns
        if (Test-Path $indicatorFile) {
            $header = Get-Content $indicatorFile -First 1
            $hasDescriptions = $header -match "Desc"
            Write-TestResult "Calculate Mode - Descriptions" $hasDescriptions
        }
    } else {
        Write-TestResult "Calculate Mode" $false "Exit code: $LASTEXITCODE"
    }
} catch {
    Write-TestResult "Calculate Mode" $false $_.Exception.Message
}

# Test Calculate_Num Mode
Write-Host "Testing Calculate_Num Mode..." -ForegroundColor $InfoColor
try {
    $calculateNumResult = & ".\isx-auto-scrapper.exe" --mode calculate_num $TestTicker 2>&1
    if ($LASTEXITCODE -eq 0) {
        $indicators2File = "Indicators2_$TestTicker.csv"
        Test-CSVContent $indicators2File 100
        
        # Compare file sizes
        if ((Test-Path "indicators_$TestTicker.csv") -and (Test-Path $indicators2File)) {
            $fullSize = (Get-Item "indicators_$TestTicker.csv").Length
            $numSize = (Get-Item $indicators2File).Length
            $smaller = $numSize -lt $fullSize
            Write-TestResult "Calculate_Num - Smaller File" $smaller "Full: $fullSize bytes, Num: $numSize bytes"
        }
    } else {
        Write-TestResult "Calculate_Num Mode" $false "Exit code: $LASTEXITCODE"
    }
} catch {
    Write-TestResult "Calculate_Num Mode" $false $_.Exception.Message
}

# Test Liquidity Mode
Write-Host "Testing Liquidity Mode..." -ForegroundColor $InfoColor
try {
    $liquidityResult = & ".\isx-auto-scrapper.exe" --mode liquidity 2>&1
    if ($LASTEXITCODE -eq 0) {
        Test-CSVContent "liquidity_scores.csv" 10
    } else {
        Write-TestResult "Liquidity Mode" $false "Exit code: $LASTEXITCODE"
    }
} catch {
    Write-TestResult "Liquidity Mode" $false $_.Exception.Message
}

# Phase 3: Strategy & Analysis Testing
Write-Host ""
Write-Host "Phase 3: Strategy & Analysis Testing" -ForegroundColor $InfoColor
Write-Host "------------------------------------" -ForegroundColor $InfoColor

# Test Strategies Mode
Write-Host "Testing Strategies Mode..." -ForegroundColor $InfoColor
try {
    $strategiesResult = & ".\isx-auto-scrapper.exe" --mode strategies 2>&1
    if ($LASTEXITCODE -eq 0) {
        $strategyFiles = Get-ChildItem "strategies_*.csv" -ErrorAction SilentlyContinue
        Write-TestResult "Strategies Mode - Files Created" ($strategyFiles.Count -gt 0) "$($strategyFiles.Count) strategy files"
        Test-FileExists "Strategy_Summary.json" "Strategy Summary JSON"
    } else {
        Write-TestResult "Strategies Mode" $false "Exit code: $LASTEXITCODE"
    }
} catch {
    Write-TestResult "Strategies Mode" $false $_.Exception.Message
}

# Test Simulate Mode
Write-Host "Testing Simulate Mode..." -ForegroundColor $InfoColor
try {
    $simulateResult = & ".\isx-auto-scrapper.exe" --mode simulate 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-TestResult "Simulate Mode" $true "Backtesting completed"
        
        # Check for backtest output files
        $backtestResults = Test-FileExists "backtest_results.csv" "Backtest Results CSV"
        $backtestJSON = Test-FileExists "backtest_results.json" "Backtest Results JSON"
        $backtestSummary = Test-FileExists "backtest_summary.json" "Backtest Summary JSON"
        
        # Check for strategy-specific files
        $tradeFiles = Get-ChildItem "backtest_trades_*.csv" -ErrorAction SilentlyContinue
        $portfolioFiles = Get-ChildItem "backtest_portfolio_*.csv" -ErrorAction SilentlyContinue
        
        Write-TestResult "Backtest Trade Files" ($tradeFiles.Count -gt 0) "$($tradeFiles.Count) trade files generated"
        Write-TestResult "Backtest Portfolio Files" ($portfolioFiles.Count -gt 0) "$($portfolioFiles.Count) portfolio files generated"
        
        # Validate backtest results content
        if (Test-Path "backtest_results.csv") {
            try {
                $backtestData = Import-Csv "backtest_results.csv"
                $strategiesCount = $backtestData.Count
                Write-TestResult "Backtest Strategies Count" ($strategiesCount -gt 0) "$strategiesCount strategies backtested"
                
                # Check for key metrics
                $hasReturns = $backtestData | Where-Object { $_.Total_Return -ne "" -and $_.Total_Return -ne "0" }
                Write-TestResult "Backtest Returns Data" ($hasReturns.Count -gt 0) "$($hasReturns.Count) strategies with return data"
                
                $hasTrades = $backtestData | Where-Object { $_.Total_Trades -ne "" -and [int]$_.Total_Trades -gt 0 }
                Write-TestResult "Backtest Trades Data" ($hasTrades.Count -gt 0) "$($hasTrades.Count) strategies with trades"
                
            } catch {
                Write-TestResult "Backtest Results Validation" $false $_.Exception.Message
            }
        }
        
        # Validate configuration
        if (Test-Path "backtest_config.json") {
            try {
                $config = Get-Content "backtest_config.json" | ConvertFrom-Json
                Write-TestResult "Backtest Config Valid" ($config.initial_cash -gt 0) "Initial cash: $($config.initial_cash)"
                Write-TestResult "Backtest Strategies Config" ($config.strategies.Count -gt 0) "$($config.strategies.Count) strategies configured"
            } catch {
                Write-TestResult "Backtest Config Validation" $false $_.Exception.Message
            }
        }
        
    } else {
        Write-TestResult "Simulate Mode" $false "Exit code: $LASTEXITCODE"
    }
} catch {
    Write-TestResult "Simulate Mode" $false $_.Exception.Message
}

# HTML Report functionality has been removed
Write-Host "HTML Report functionality removed from codebase" -ForegroundColor $InfoColor

# Phase 4: Data Validation
Write-Host ""
Write-Host "Phase 4: Data Validation" -ForegroundColor $InfoColor
Write-Host "------------------------" -ForegroundColor $InfoColor

# Validate data consistency between files
if ((Test-Path "indicators_$TestTicker.csv") -and (Test-Path "Indicators2_$TestTicker.csv")) {
    Write-Host "Comparing indicator values between full and numerical files..." -ForegroundColor $InfoColor
    
    try {
        $fullData = Import-Csv "indicators_$TestTicker.csv" | Select-Object -First 5
        $numData = Import-Csv "Indicators2_$TestTicker.csv" | Select-Object -First 5
        
        $smaMatch = $fullData[0].SMA10 -eq $numData[0].SMA10
        Write-TestResult "Data Consistency - SMA10" $smaMatch "Full: $($fullData[0].SMA10), Num: $($numData[0].SMA10)"
        
        $rsiMatch = $fullData[0].RSI_14 -eq $numData[0].RSI_14
        Write-TestResult "Data Consistency - RSI14" $rsiMatch "Full: $($fullData[0].RSI_14), Num: $($numData[0].RSI_14)"
    } catch {
        Write-TestResult "Data Consistency Check" $false $_.Exception.Message
    }
}

# Performance Summary
Write-Host ""
Write-Host "Phase 5: Performance Summary" -ForegroundColor $InfoColor
Write-Host "---------------------------" -ForegroundColor $InfoColor

# Check log file for performance metrics
if (Test-Path "stock_analysis.log") {
    $logSize = (Get-Item "stock_analysis.log").Length
    Write-Host "Log file size: $([math]::Round($logSize/1KB, 2)) KB" -ForegroundColor Gray
    
    # Look for error messages in log
    $errors = Select-String -Path "stock_analysis.log" -Pattern "ERROR" -ErrorAction SilentlyContinue
    if ($errors) {
        Write-TestResult "Log Error Check" $false "$($errors.Count) errors found in log"
    } else {
        Write-TestResult "Log Error Check" $true "No errors found in log"
    }
}

# Final Summary
Write-Host ""
Write-Host "=== Testing Summary ===" -ForegroundColor $InfoColor

$testResults = @{
    "Raw Data Available" = (Get-ChildItem "raw_*.csv" -ErrorAction SilentlyContinue).Count -gt 0
    "Indicators Calculated" = Test-Path "indicators_$TestTicker.csv"
    "Numerical Indicators" = Test-Path "Indicators2_$TestTicker.csv"
    "Liquidity Scores" = Test-Path "liquidity_scores.csv"
    "Strategy Analysis" = Test-Path "Strategy_Summary.json"
}

$passCount = ($testResults.Values | Where-Object { $_ -eq $true }).Count
$totalCount = $testResults.Count

Write-Host "Overall Test Results: $passCount/$totalCount tests passed" -ForegroundColor $(if ($passCount -eq $totalCount) { $SuccessColor } else { $WarningColor })

foreach ($result in $testResults.GetEnumerator()) {
    $status = if ($result.Value) { "[PASS]" } else { "[FAIL]" }
    $color = if ($result.Value) { $SuccessColor } else { $ErrorColor }
    Write-Host "$status $($result.Key)" -ForegroundColor $color
}

# Recommendations
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor $InfoColor

if ($passCount -eq $totalCount) {
    Write-Host "[SUCCESS] All tests passed! Your ISX Auto Scrapper is working correctly." -ForegroundColor $SuccessColor
    Write-Host "  - You can now use all modes confidently" -ForegroundColor Gray
    Write-Host "  - Consider setting up automated scheduled runs" -ForegroundColor Gray
} else {
    Write-Host "[WARNING] Some tests failed. Please check:" -ForegroundColor $WarningColor
    Write-Host "  - Log files for error details" -ForegroundColor Gray
    Write-Host "  - File permissions and disk space" -ForegroundColor Gray
    Write-Host "  - Input data quality" -ForegroundColor Gray
}

Write-Host ""
Write-Host "For detailed testing procedures, see TESTING_PLAN.md" -ForegroundColor $InfoColor 