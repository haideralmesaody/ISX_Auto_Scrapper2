package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var (
	mode string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "isx-auto-scrapper",
		Short: "ISX Auto Scrapper - Iraq Stock Exchange Data Analysis Tool",
		Long: `A comprehensive tool for scraping, analyzing, and backtesting 
Iraq Stock Exchange (ISX) stock data with technical indicators and trading strategies.`,
		Run: runApp,
	}

	rootCmd.Flags().StringVar(&mode, "mode", "", "The mode to run the script in (required)")
	rootCmd.MarkFlagRequired("mode")

	// Add mode validation
	cobra.CheckErr(rootCmd.Execute())
}

func runApp(cmd *cobra.Command, args []string) {
	logger := NewLogger()

	// Initialize components
	dataFetcher := NewDataFetcher()
	indicatorsCalculator := NewIndicatorsCalculator()
	liquidityCalc := NewLiquidityCalc()
	strategies := NewStrategies()
	strategyTester := NewStrategyTester()

	switch mode {
	case "web":
		// Start web dashboard
		port := 8080
		if len(args) > 0 {
			if p, err := strconv.Atoi(args[0]); err == nil && p > 0 && p < 65536 {
				port = p
			}
		}

		webServer := NewWebServer(port)
		logger.Info("Starting ISX Auto Scrapper Web Dashboard...")
		logger.Info("Open your browser and navigate to: http://localhost:%d", port)

		if err := webServer.Start(); err != nil {
			logger.Error("Web server failed to start: %v", err)
			os.Exit(1)
		}

	case "single":
		fmt.Print("Enter the ticker: ")
		var ticker string
		fmt.Scanln(&ticker)

		err := dataFetcher.FetchData(ticker)
		if err != nil {
			logger.Error("Failed to fetch data for ticker %s: %v", ticker, err)
			os.Exit(1)
		}

	case "auto":
		tickers, err := LoadTickersWithInfo("TICKERS.csv")
		if err != nil {
			logger.Error("Failed to load tickers: %v", err)
			os.Exit(1)
		}

		numTickers := len(tickers)
		var reports []ProcessingReport
		var timingReports []TimingReport

		logger.Info("Starting auto mode processing for %d tickers", numTickers)
		overallStartTime := time.Now()

		for i, tickerInfo := range tickers {
			logger.Info("Processing %s (%d/%d) - %s", tickerInfo.Symbol, i+1, numTickers, tickerInfo.CompanyName)

			// Fetch data with detailed reporting
			report, err := dataFetcher.FetchDataWithReport(tickerInfo.Symbol, tickerInfo.Sector, tickerInfo.CompanyName)
			if err != nil {
				logger.Error("Failed to fetch data for %s: %v", tickerInfo.Symbol, err)
				// Create error report
				if report == nil {
					report = &ProcessingReport{
						Ticker:           tickerInfo.Symbol,
						Sector:           tickerInfo.Sector,
						CompanyName:      tickerInfo.CompanyName,
						Status:           "ERROR",
						ErrorMessage:     err.Error(),
						Recommendation:   "Manually check ticker - failed to fetch data. Consider removing from tickers file if consistently failing.",
						DataQualityScore: "FAILED",
					}
				}
				dataFetcher.finalizeReport(err)
				report = dataFetcher.currentReport
			}

			if report != nil {
				reports = append(reports, *report)
			}

			// Collect timing report
			if timingReport := dataFetcher.GetTimingReport(); timingReport != nil {
				timingReports = append(timingReports, *timingReport)
			}

			// Continue with calculations if data was fetched successfully
			if err == nil {
				if calcErr := indicatorsCalculator.CalculateAll(tickerInfo.Symbol); calcErr != nil {
					logger.Error("Failed to calculate indicators for %s: %v", tickerInfo.Symbol, calcErr)
				}

			}
		}

		// Save processing report
		reportFilename := fmt.Sprintf("Processing_Report_%s.csv", time.Now().Format("2006-01-02_15-04-05"))
		if err := SaveProcessingReport(reports, reportFilename); err != nil {
			logger.Error("Failed to save processing report: %v", err)
		} else {
			logger.Info("Processing report saved to %s", reportFilename)
		}

		// Save timing analysis report
		timingFilename := fmt.Sprintf("Timing_Analysis_%s.csv", time.Now().Format("2006-01-02_15-04-05"))
		if err := SaveTimingReport(timingReports, timingFilename); err != nil {
			logger.Error("Failed to save timing report: %v", err)
		} else {
			logger.Info("Timing analysis report saved to %s", timingFilename)
		}

		// Generate summary statistics
		totalProcessed := len(reports)
		successful := 0
		errors := 0
		partial := 0
		upToDate := 0

		// Performance statistics
		var totalProcessingTime time.Duration
		var excellentPerf, goodPerf, avgPerf, poorPerf int
		var totalPages int

		for _, report := range reports {
			switch report.Status {
			case "SUCCESS":
				successful++
			case "ERROR":
				errors++
			case "PARTIAL":
				partial++
			case "UP_TO_DATE":
				upToDate++
			}
		}

		for _, timing := range timingReports {
			totalProcessingTime += timing.TotalProcessingTime
			totalPages += timing.PagesProcessed

			switch timing.PerformanceScore {
			case "EXCELLENT":
				excellentPerf++
			case "GOOD":
				goodPerf++
			case "AVERAGE":
				avgPerf++
			case "POOR":
				poorPerf++
			}
		}

		overallDuration := time.Since(overallStartTime)
		avgProcessingTime := time.Duration(0)
		if len(timingReports) > 0 {
			avgProcessingTime = totalProcessingTime / time.Duration(len(timingReports))
		}

		logger.Info("Auto mode completed in %s", overallDuration.String())
		logger.Info("Summary: %d total, %d successful, %d up-to-date, %d partial, %d errors",
			totalProcessed, successful, upToDate, partial, errors)
		logger.Info("Performance: %d excellent, %d good, %d average, %d poor",
			excellentPerf, goodPerf, avgPerf, poorPerf)
		logger.Info("Timing: Total pages processed: %d, Average processing time per ticker: %s",
			totalPages, avgProcessingTime.String())

		// Run additional analysis only for successful downloads
		if successful > 0 {
			logger.Info("Running additional analysis...")
			liquidityCalc.CalculateScores()
			strategies.ApplyStrategiesAndSave()
			strategies.ApplyAlternativeStrategyStates()
			strategies.SummarizeStrategyActions()
		}

	case "liquidity":
		liquidityCalc.CalculateScores()

	case "strategies":
		strategies.ApplyStrategiesAndSave()
		strategies.ApplyAlternativeStrategyStates()
		strategies.SummarizeStrategyActions()

	case "simulate":
		strategyTester.SimulateStrategyResults()
		strategyTester.SummarizeSimulatedStrategyResults()

	case "calculate":
		// Calculate indicators for a single ticker (with descriptions)
		if len(args) > 0 {
			// Calculate indicators for a single ticker
			ticker := args[0]
			logger.Info("Calculating full indicators for ticker %s", ticker)
			if err := indicatorsCalculator.CalculateAll(ticker); err != nil {
				logger.Error("Failed to calculate indicators for %s: %v", ticker, err)
				os.Exit(1)
			}
			logger.Info("Indicators calculation completed for %s", ticker)
		} else {
			// Calculate indicators for all tickers
			tickers, err := LoadTickers("TICKERS.csv")
			if err != nil {
				logger.Error("Failed to load tickers: %v", err)
				os.Exit(1)
			}

			numTickers := len(tickers)
			for i, ticker := range tickers {
				logger.Info("Calculating full indicators for %s (%d/%d)", ticker, i+1, numTickers)
				if err := indicatorsCalculator.CalculateAll(ticker); err != nil {
					logger.Error("Failed to calculate indicators for %s: %v", ticker, err)
				}
			}
		}

	case "calculate_num":
		// Use the dedicated NumericalIndicatorsCalculator for numerical calculations
		numericalIndicatorsCalculator := NewNumericalIndicatorsCalculator()

		if len(args) > 0 {
			// Calculate numerical indicators for a single ticker
			ticker := args[0]
			logger.Info("Calculating numerical indicators for ticker %s", ticker)
			if err := numericalIndicatorsCalculator.CalculateAllNums(ticker); err != nil {
				logger.Error("Failed to calculate indicators for %s: %v", ticker, err)
				os.Exit(1)
			}
			logger.Info("Numerical indicators calculation completed for %s", ticker)
		} else {
			// Calculate numerical indicators for all tickers
			tickers, err := LoadTickers("TICKERS.csv")
			if err != nil {
				logger.Error("Failed to load tickers: %v", err)
				os.Exit(1)
			}

			numTickers := len(tickers)
			for i, ticker := range tickers {
				logger.Info("Calculating indicators for %s (%d/%d)", ticker, i+1, numTickers)
				if err := numericalIndicatorsCalculator.CalculateAllNums(ticker); err != nil {
					logger.Error("Failed to calculate indicators for %s: %v", ticker, err)
				}
			}
		}

	default:
		log.Fatalf("Invalid mode: %s. Valid modes are: web, single, auto, liquidity, strategies, simulate, calculate, calculate_num", mode)
	}
}
