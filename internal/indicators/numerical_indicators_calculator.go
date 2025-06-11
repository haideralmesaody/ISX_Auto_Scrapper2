package indicators

import (
	"fmt"
	"os"

	"isx-auto-scrapper/internal/common"
)

// NumericalIndicatorsCalculator handles numerical technical indicator calculations (no descriptions)
// It reuses the existing IndicatorsCalculator methods and only adds the numerical-specific workflow
type NumericalIndicatorsCalculator struct {
	indicatorsCalculator *IndicatorsCalculator
	logger               *common.Logger
}

// NewNumericalIndicatorsCalculator creates a new NumericalIndicatorsCalculator instance
func NewNumericalIndicatorsCalculator() *NumericalIndicatorsCalculator {
	return &NumericalIndicatorsCalculator{
		indicatorsCalculator: NewIndicatorsCalculator(),
		logger:               common.NewLogger(),
	}
}

// CalculateAllNums calculates all numerical technical indicators for a ticker
// This method reuses existing calculation logic but saves to Indicators2_ files without descriptions
func (nic *NumericalIndicatorsCalculator) CalculateAllNums(ticker string) error {
	nic.logger.Info("Calculating numerical indicators for ticker %s", ticker)

	// Check if the raw data CSV file exists
	rawFilePath := fmt.Sprintf("raw_%s.csv", ticker)
	if _, err := os.Stat(rawFilePath); os.IsNotExist(err) {
		nic.logger.Info("File %s does not exist.", rawFilePath)
		return fmt.Errorf("raw data file does not exist: %s", rawFilePath)
	}

	// Load stock data using the existing DataCalculator method
	stockData, err := nic.indicatorsCalculator.loadStockData(rawFilePath)
	if err != nil {
		nic.logger.Error("Failed to load stock data: %v", err)
		return err
	}

	if len(stockData) == 0 {
		nic.logger.Error("The DataFrame from raw data is empty.")
		return fmt.Errorf("no stock data found")
	}

	// Define the path for the numerical indicators CSV file (Indicators2_ instead of indicators_)
	indicatorsFilePath := fmt.Sprintf("Indicators2_%s.csv", ticker)

	// Check if the indicators CSV file already exists and is up-to-date
	if nic.isDataUpToDate(indicatorsFilePath, stockData) {
		nic.logger.Info("The data is up to date.")
		return nil
	}

	// Calculate all indicators using existing DataCalculator methods
	nic.logger.Info("Calculating numerical technical indicators...")

	// Reuse all the existing calculation methods from DataCalculator
	if err := nic.indicatorsCalculator.calculateSMA(stockData); err != nil {
		return fmt.Errorf("failed to calculate SMA: %w", err)
	}

	if err := nic.indicatorsCalculator.calculateRSI(stockData); err != nil {
		return fmt.Errorf("failed to calculate RSI: %w", err)
	}

	if err := nic.indicatorsCalculator.calculateEMAs(stockData); err != nil {
		return fmt.Errorf("failed to calculate EMAs: %w", err)
	}

	if err := nic.indicatorsCalculator.calculateStochasticOscillator(stockData); err != nil {
		return fmt.Errorf("failed to calculate Stochastic: %w", err)
	}

	if err := nic.indicatorsCalculator.calculateCMF(stockData); err != nil {
		return fmt.Errorf("failed to calculate CMF: %w", err)
	}

	if err := nic.indicatorsCalculator.calculateMACD(stockData); err != nil {
		return fmt.Errorf("failed to calculate MACD: %w", err)
	}

	if err := nic.indicatorsCalculator.calculateOBV(stockData); err != nil {
		return fmt.Errorf("failed to calculate OBV: %w", err)
	}

	if err := nic.indicatorsCalculator.calculatePSAR(stockData); err != nil {
		return fmt.Errorf("failed to calculate PSAR: %w", err)
	}

	if err := nic.indicatorsCalculator.calculateATR(stockData); err != nil {
		return fmt.Errorf("failed to calculate ATR: %w", err)
	}

	if err := nic.indicatorsCalculator.calculateRollingStd(stockData); err != nil {
		return fmt.Errorf("failed to calculate Rolling Std: %w", err)
	}

	nic.logger.Info("Data calculation completed.")

	// Save the data without descriptions (skip the addDescriptions step)
	if err := nic.indicatorsCalculator.saveIndicatorsData(stockData, indicatorsFilePath); err != nil {
		return fmt.Errorf("failed to save indicators data: %w", err)
	}

	nic.logger.Info("Numerical indicators calculation completed and saved to %s.", indicatorsFilePath)
	return nil
}

// isDataUpToDate checks if the indicators file is up-to-date
// This is the only method we need to implement separately because it uses a different file path
func (nic *NumericalIndicatorsCalculator) isDataUpToDate(indicatorsFilePath string, stockData []*StockDataWithIndicators) bool {
	if _, err := os.Stat(indicatorsFilePath); os.IsNotExist(err) {
		return false
	}

	// Load existing indicators file using the DataCalculator method
	existingData, err := nic.indicatorsCalculator.loadExistingIndicators(indicatorsFilePath)
	if err != nil || len(existingData) == 0 {
		return false
	}

	// Compare last dates
	lastExistingDate := existingData[len(existingData)-1].Date
	lastNewDate := stockData[len(stockData)-1].Date

	return lastExistingDate.Equal(lastNewDate)
}
