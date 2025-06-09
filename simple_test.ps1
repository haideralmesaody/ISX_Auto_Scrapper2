# Simple ISX Auto Scrapper Testing Script
# Compatible with all PowerShell versions

param(
    [string]$TestTicker = "AAHP"
)

Write-Host "=== ISX Auto Scrapper Simple Test Suite ===" -ForegroundColor Cyan
Write-Host "Testing with ticker: $TestTicker" -ForegroundColor Yellow
Write-Host ""

# Function to test file existence and content
function Test-OutputFile {
    param($FilePath, $Description, $MinLines = 1)
    
    if (Test-Path $FilePath) {
        try {
            $lineCount = (Get-Content $FilePath | Measure-Object -Line).Lines
            if ($lineCount -gt $MinLines) {
                Write-Host "[PASS] $Description - $lineCount lines" -ForegroundColor Green
                return $true
            } else {
                Write-Host "[FAIL] $Description - Only $lineCount lines" -ForegroundColor Red
                return $false
            }
        } catch {
            Write-Host "[FAIL] $Description - Error reading file" -ForegroundColor Red
            return $false
        }
    } else {
        Write-Host "[FAIL] $Description - File not found: $FilePath" -ForegroundColor Red
        return $false
    }
}

# Function to run mode and check result
function Test-Mode {
    param($Mode, $Description, $Args = "")
    
    Write-Host "Testing $Mode mode..." -ForegroundColor Cyan
    
    try {
        if ($Args) {
            $result = & ".\isx-auto-scrapper.exe" --mode $Mode $Args 2>&1
        } else {
            $result = & ".\isx-auto-scrapper.exe" --mode $Mode 2>&1
        }
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "[PASS] $Description completed successfully" -ForegroundColor Green
            return $true
        } else {
            Write-Host "[FAIL] $Description failed with exit code: $LASTEXITCODE" -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Host "[FAIL] $Description error: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# Initialize test results
$testResults = @()

# Test 1: Check if auto mode results exist
Write-Host "=== Phase 1: Auto Mode Results Check ===" -ForegroundColor Yellow
$rawFiles = Get-ChildItem "raw_*.csv" -ErrorAction SilentlyContinue
$indicatorFiles = Get-ChildItem "indicators_*.csv" -ErrorAction SilentlyContinue

Write-Host "Found $($rawFiles.Count) raw CSV files" -ForegroundColor Gray
Write-Host "Found $($indicatorFiles.Count) indicator CSV files" -ForegroundColor Gray

$autoResults = Test-OutputFile "liquidity_scores.csv" "Liquidity Scores" 5
$testResults += $autoResults

if (Test-Path "Strategy_Summary.json") {
    Write-Host "[PASS] Strategy Summary JSON exists" -ForegroundColor Green
    $testResults += $true
} else {
    Write-Host "[FAIL] Strategy Summary JSON missing" -ForegroundColor Red
    $testResults += $false
}

# Test 2: Calculate Mode
Write-Host ""
Write-Host "=== Phase 2: Calculate Mode Test ===" -ForegroundColor Yellow
$calcResult = Test-Mode "calculate" "Calculate Mode" $TestTicker
$testResults += $calcResult

if ($calcResult) {
    $indicatorResult = Test-OutputFile "indicators_$TestTicker.csv" "Indicators File" 50
    $testResults += $indicatorResult
}

# Test 3: Calculate_Num Mode
Write-Host ""
Write-Host "=== Phase 3: Calculate_Num Mode Test ===" -ForegroundColor Yellow
$calcNumResult = Test-Mode "calculate_num" "Calculate_Num Mode" $TestTicker
$testResults += $calcNumResult

if ($calcNumResult) {
    $indicators2Result = Test-OutputFile "Indicators2_$TestTicker.csv" "Numerical Indicators File" 50
    $testResults += $indicators2Result
}

# Test 4: Liquidity Mode
Write-Host ""
Write-Host "=== Phase 4: Liquidity Mode Test ===" -ForegroundColor Yellow
$liquidityResult = Test-Mode "liquidity" "Liquidity Mode"
$testResults += $liquidityResult

# Test 5: Strategies Mode
Write-Host ""
Write-Host "=== Phase 5: Strategies Mode Test ===" -ForegroundColor Yellow
$strategiesResult = Test-Mode "strategies" "Strategies Mode"
$testResults += $strategiesResult

# Test 6: Simulate Mode
Write-Host ""
Write-Host "=== Phase 6: Simulate Mode Test ===" -ForegroundColor Yellow
$simulateResult = Test-Mode "simulate" "Simulate Mode"
$testResults += $simulateResult

if ($simulateResult) {
    # Check for backtest output files
    Write-Host "Checking backtest output files..." -ForegroundColor Gray
    
    $backtestResults = Test-OutputFile "backtest_results.csv" "Backtest Results" 1
    $testResults += $backtestResults
    
    $backtestJSON = Test-OutputFile "backtest_results.json" "Backtest JSON" 1
    $testResults += $backtestJSON
    
    $backtestSummary = Test-OutputFile "backtest_summary.json" "Backtest Summary" 1
    $testResults += $backtestSummary
    
    # Count strategy-specific files
    $tradeFiles = Get-ChildItem "backtest_trades_*.csv" -ErrorAction SilentlyContinue
    $portfolioFiles = Get-ChildItem "backtest_portfolio_*.csv" -ErrorAction SilentlyContinue
    
    if ($tradeFiles.Count -gt 0) {
        Write-Host "[PASS] Generated $($tradeFiles.Count) trade files" -ForegroundColor Green
        $testResults += $true
    } else {
        Write-Host "[FAIL] No trade files generated" -ForegroundColor Red
        $testResults += $false
    }
    
    if ($portfolioFiles.Count -gt 0) {
        Write-Host "[PASS] Generated $($portfolioFiles.Count) portfolio files" -ForegroundColor Green
        $testResults += $true
    } else {
        Write-Host "[FAIL] No portfolio files generated" -ForegroundColor Red
        $testResults += $false
    }
}

# Test 7: Data Consistency Check
Write-Host ""
Write-Host "=== Phase 7: Data Consistency Check ===" -ForegroundColor Yellow

if ((Test-Path "indicators_$TestTicker.csv") -and (Test-Path "Indicators2_$TestTicker.csv")) {
    try {
        $fullData = Import-Csv "indicators_$TestTicker.csv" | Select-Object -First 1
        $numData = Import-Csv "Indicators2_$TestTicker.csv" | Select-Object -First 1
        
        if ($fullData.SMA10 -eq $numData.SMA10) {
            Write-Host "[PASS] Data consistency check - SMA10 values match" -ForegroundColor Green
            $testResults += $true
        } else {
            Write-Host "[FAIL] Data consistency check - SMA10 values differ" -ForegroundColor Red
            Write-Host "  Full: $($fullData.SMA10), Num: $($numData.SMA10)" -ForegroundColor Gray
            $testResults += $false
        }
    } catch {
        Write-Host "[FAIL] Data consistency check - Error comparing files" -ForegroundColor Red
        $testResults += $false
    }
} else {
    Write-Host "[SKIP] Data consistency check - Required files missing" -ForegroundColor Yellow
}

# Final Summary
Write-Host ""
Write-Host "=== Test Summary ===" -ForegroundColor Cyan
$passCount = ($testResults | Where-Object { $_ -eq $true }).Count
$totalCount = $testResults.Count

Write-Host "Test Results: $passCount/$totalCount tests passed" -ForegroundColor $(if ($passCount -eq $totalCount) { "Green" } else { "Yellow" })

if ($passCount -eq $totalCount) {
    Write-Host ""
    Write-Host "[SUCCESS] All tests passed! Your ISX Auto Scrapper is working correctly." -ForegroundColor Green
    Write-Host "You can now use all modes confidently." -ForegroundColor Gray
} elseif ($passCount -gt ($totalCount * 0.7)) {
    Write-Host ""
    Write-Host "[PARTIAL] Most tests passed. Some minor issues detected." -ForegroundColor Yellow
    Write-Host "Check the failed tests above for details." -ForegroundColor Gray
} else {
    Write-Host ""
    Write-Host "[WARNING] Multiple tests failed. Please investigate:" -ForegroundColor Red
    Write-Host "- Check log files for error details" -ForegroundColor Gray
    Write-Host "- Verify input data files exist" -ForegroundColor Gray
    Write-Host "- Ensure proper file permissions" -ForegroundColor Gray
}

Write-Host ""
Write-Host "For detailed testing procedures, see TESTING_PLAN.md" -ForegroundColor Cyan 