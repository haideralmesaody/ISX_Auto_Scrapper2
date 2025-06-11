package indicators

import (
	"fmt"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/shopspring/decimal"

	"isx-auto-scrapper/internal/common"
)

// CSVDate is a custom type for parsing dates from CSV
type CSVDate struct {
	time.Time
}

// UnmarshalCSV implements the CSVUnmarshaler interface
func (date *CSVDate) UnmarshalCSV(csv string) error {
	t, err := time.Parse("2006-01-02", csv)
	if err != nil {
		return err
	}
	date.Time = t
	return nil
}

// MarshalCSV implements the CSVMarshaler interface
func (date CSVDate) MarshalCSV() (string, error) {
	return date.Time.Format("2006-01-02"), nil
}

// StockDataCSV represents stock data for CSV parsing
type StockDataCSV struct {
	Date          CSVDate         `csv:"Date"`
	Close         decimal.Decimal `csv:"Close"`
	Open          decimal.Decimal `csv:"Open"`
	High          decimal.Decimal `csv:"High"`
	Low           decimal.Decimal `csv:"Low"`
	Change        decimal.Decimal `csv:"Change"`
	ChangePercent string          `csv:"Change%"`  // String because it has % symbol
	TShares       string          `csv:"T.Shares"` // String because it might be empty
	Volume        int64           `csv:"Volume"`
	NoTrades      string          `csv:"No. Trades"` // String because it might be empty
}

// StockDataWithIndicators extends StockData with technical indicators
type StockDataWithIndicators struct {
	common.StockData

	// Additional RSI periods not in StockData
	RSI9  decimal.Decimal `csv:"RSI_9"`
	RSI25 decimal.Decimal `csv:"RSI_25"`

	// Additional indicators not in StockData
	CMF20 decimal.Decimal `csv:"CMF_20"`

	// OBV additional indicators
	OBVSMADiff decimal.Decimal `csv:"OBV_SMA_Diff"`
	OBVRoC     decimal.Decimal `csv:"OBV_RoC"`

	// Additional PSAR
	PSAR1 decimal.Decimal `csv:"PSARl_0.02_0.2"`

	// Additional Rolling Standard Deviation periods
	RollingStd10 decimal.Decimal `csv:"Rolling_Std_10"`
	RollingStd50 decimal.Decimal `csv:"Rolling_Std_50"`
}

// IndicatorsCalculator handles technical indicator calculations
type IndicatorsCalculator struct {
	logger     *common.Logger
	indicators *TechnicalIndicators
}

// NewIndicatorsCalculator creates a new IndicatorsCalculator instance
func NewIndicatorsCalculator() *IndicatorsCalculator {
	return &IndicatorsCalculator{
		logger:     common.NewLogger(),
		indicators: NewTechnicalIndicators(),
	}
}

// CalculateAll calculates all technical indicators for a ticker
func (ic *IndicatorsCalculator) CalculateAll(ticker string) error {
	ic.logger.Info("Calculating indicators for ticker %s", ticker)

	// Check if the raw data CSV file exists
	rawFilePath := fmt.Sprintf("raw_%s.csv", ticker)
	if _, err := os.Stat(rawFilePath); os.IsNotExist(err) {
		ic.logger.Info("File %s does not exist.", rawFilePath)
		return fmt.Errorf("raw data file does not exist: %s", rawFilePath)
	}

	// Read the raw data from CSV file
	stockData, err := ic.loadStockData(rawFilePath)
	if err != nil {
		ic.logger.Error("Failed to load stock data: %v", err)
		return err
	}

	if len(stockData) == 0 {
		ic.logger.Error("The DataFrame from raw data is empty.")
		return fmt.Errorf("no stock data found")
	}

	// Define the path for the indicators CSV file
	indicatorsFilePath := fmt.Sprintf("indicators_%s.csv", ticker)

	// Check if the indicators CSV file already exists and is up-to-date
	if ic.isDataUpToDate(indicatorsFilePath, stockData) {
		ic.logger.Info("The data is up to date.")
		return nil
	}

	// Calculate all the indicators
	ic.logger.Info("Calculating technical indicators...")

	// Calculate SMA indicators
	if err := ic.calculateSMA(stockData); err != nil {
		return fmt.Errorf("failed to calculate SMA: %w", err)
	}

	// Calculate RSI indicators
	if err := ic.calculateRSI(stockData); err != nil {
		return fmt.Errorf("failed to calculate RSI: %w", err)
	}

	// Calculate Stochastic Oscillator
	if err := ic.calculateStochasticOscillator(stockData); err != nil {
		return fmt.Errorf("failed to calculate Stochastic: %w", err)
	}

	// Calculate CMF
	if err := ic.calculateCMF(stockData); err != nil {
		return fmt.Errorf("failed to calculate CMF: %w", err)
	}

	// Calculate MACD
	if err := ic.calculateMACD(stockData); err != nil {
		return fmt.Errorf("failed to calculate MACD: %w", err)
	}

	// Calculate OBV
	if err := ic.calculateOBV(stockData); err != nil {
		return fmt.Errorf("failed to calculate OBV: %w", err)
	}

	// Calculate EMA indicators
	if err := ic.calculateEMAs(stockData); err != nil {
		return fmt.Errorf("failed to calculate EMAs: %w", err)
	}

	// Calculate PSAR
	if err := ic.calculatePSAR(stockData); err != nil {
		return fmt.Errorf("failed to calculate PSAR: %w", err)
	}

	// Calculate ATR
	if err := ic.calculateATR(stockData); err != nil {
		return fmt.Errorf("failed to calculate ATR: %w", err)
	}

	// Calculate Rolling Standard Deviation
	if err := ic.calculateRollingStd(stockData); err != nil {
		return fmt.Errorf("failed to calculate Rolling Std: %w", err)
	}

	// Add descriptions for full mode
	ic.addDescriptions(stockData)

	// Save the updated data to CSV file
	if err := ic.saveIndicatorsData(stockData, indicatorsFilePath); err != nil {
		return fmt.Errorf("failed to save indicators data: %w", err)
	}

	ic.logger.Info("Data calculation completed and saved to %s.", indicatorsFilePath)
	return nil
}

// loadStockData loads stock data from CSV file
func (ic *IndicatorsCalculator) loadStockData(filePath string) ([]*StockDataWithIndicators, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rawData []*StockDataCSV
	if err := gocsv.UnmarshalFile(file, &rawData); err != nil {
		return nil, err
	}

	// Convert to StockDataWithIndicators
	stockData := make([]*StockDataWithIndicators, len(rawData))
	for i, data := range rawData {
		stockData[i] = &StockDataWithIndicators{
			StockData: common.StockData{
				Date:   data.Date.Time,
				Close:  data.Close,
				Open:   data.Open,
				High:   data.High,
				Low:    data.Low,
				Volume: data.Volume,
				Change: data.Change,
				// Parse ChangePercent (remove % and convert)
				ChangePercent: ParsePercentage(data.ChangePercent),

				// Initialize all technical indicator fields
				SMA10:      decimal.Zero,
				SMA50:      decimal.Zero,
				SMA200:     decimal.Zero,
				EMA5:       decimal.Zero,
				EMA10:      decimal.Zero,
				EMA20:      decimal.Zero,
				EMA50:      decimal.Zero,
				EMA200:     decimal.Zero,
				RSI14:      decimal.Zero,
				StochK:     decimal.Zero,
				StochD:     decimal.Zero,
				MACD:       decimal.Zero,
				MACDSignal: decimal.Zero,
				MACDHist:   decimal.Zero,
				CMF:        decimal.Zero,
				OBV:        decimal.Zero,
				ATR:        decimal.Zero,
				PSAR:       decimal.Zero,
				PSAR2:      decimal.Zero,
				RollingStd: decimal.Zero,

				// Initialize crossover signals
				GoldenCross:          false,
				DeathCross:           false,
				PriceCrossSMA10Up:    false,
				PriceCrossSMA10Down:  false,
				PriceCrossSMA50Up:    false,
				PriceCrossSMA50Down:  false,
				PriceCrossSMA200Up:   false,
				PriceCrossSMA200Down: false,

				// Initialize trend indicators
				SMA10Up:  false,
				SMA50Up:  false,
				SMA200Up: false,

				// Initialize distance calculations
				PriceDistanceSMA10:  decimal.Zero,
				PriceDistanceSMA50:  decimal.Zero,
				PriceDistanceSMA200: decimal.Zero,

				// Initialize relationship indicators
				SMA50AboveSMA200: false,

				// Initialize description fields
				GoldenDeathCrossDesc:    "",
				PriceSMA10CrossoverDesc: "",
				PriceCrossoverDesc:      "",
				RSIDesc:                 "",
				StochasticDesc:          "",
				CMFDesc:                 "",
				MACDDesc:                "",
				OBVDesc:                 "",
				PSARDesc:                "",
				ATRDesc:                 "",
			},
			// Initialize indicator fields with zero values
			RSI9:         decimal.Zero,
			RSI25:        decimal.Zero,
			CMF20:        decimal.Zero,
			OBVSMADiff:   decimal.Zero,
			OBVRoC:       decimal.Zero,
			PSAR1:        decimal.Zero,
			RollingStd10: decimal.Zero,
			RollingStd50: decimal.Zero,
		}
	}

	return stockData, nil
}

// parsePercentage converts percentage string like "2.15%" to decimal
func ParsePercentage(percentStr string) decimal.Decimal {
	if percentStr == "" {
		return decimal.Zero
	}

	// Remove % symbol
	if len(percentStr) > 0 && percentStr[len(percentStr)-1] == '%' {
		percentStr = percentStr[:len(percentStr)-1]
	}

	// Parse as decimal
	if val, err := decimal.NewFromString(percentStr); err == nil {
		return val
	}

	return decimal.Zero
}

// isDataUpToDate checks if the indicators file is up-to-date
func (ic *IndicatorsCalculator) isDataUpToDate(indicatorsFilePath string, stockData []*StockDataWithIndicators) bool {
	if _, err := os.Stat(indicatorsFilePath); os.IsNotExist(err) {
		return false
	}

	// Load existing indicators file
	existingData, err := ic.loadExistingIndicators(indicatorsFilePath)
	if err != nil || len(existingData) == 0 {
		return false
	}

	// Compare last dates
	lastExistingDate := existingData[len(existingData)-1].Date
	lastNewDate := stockData[len(stockData)-1].Date

	return lastExistingDate.Equal(lastNewDate)
}

// loadExistingIndicators loads existing indicators from CSV file
func (ic *IndicatorsCalculator) loadExistingIndicators(filePath string) ([]*StockDataWithIndicators, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data []*StockDataWithIndicators
	if err := gocsv.UnmarshalFile(file, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// saveIndicatorsData saves the indicators data to CSV file
func (ic *IndicatorsCalculator) saveIndicatorsData(stockData []*StockDataWithIndicators, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return gocsv.MarshalFile(stockData, file)
}

// calculateSMA calculates Simple Moving Averages and related indicators
func (ic *IndicatorsCalculator) calculateSMA(stockData []*StockDataWithIndicators) error {
	if len(stockData) == 0 {
		return fmt.Errorf("no data provided for SMA calculation")
	}

	// Calculate SMA10, SMA50, SMA200
	for i := range stockData {
		// SMA10
		if i >= 9 { // Need at least 10 data points
			sum := decimal.Zero
			for j := i - 9; j <= i; j++ {
				sum = sum.Add(stockData[j].Close)
			}
			stockData[i].SMA10 = sum.Div(decimal.NewFromInt(10)).Round(2)
		}

		// SMA50
		if i >= 49 { // Need at least 50 data points
			sum := decimal.Zero
			for j := i - 49; j <= i; j++ {
				sum = sum.Add(stockData[j].Close)
			}
			stockData[i].SMA50 = sum.Div(decimal.NewFromInt(50)).Round(2)
		}

		// SMA200
		if i >= 199 { // Need at least 200 data points
			sum := decimal.Zero
			for j := i - 199; j <= i; j++ {
				sum = sum.Add(stockData[j].Close)
			}
			stockData[i].SMA200 = sum.Div(decimal.NewFromInt(200)).Round(2)
		}
	}

	// Calculate crossover signals and other SMA-related indicators
	for i := 1; i < len(stockData); i++ {
		current := stockData[i]
		previous := stockData[i-1]

		// Golden Cross and Death Cross
		if !current.SMA50.IsZero() && !current.SMA200.IsZero() &&
			!previous.SMA50.IsZero() && !previous.SMA200.IsZero() {

			// Golden Cross: SMA50 crosses above SMA200
			if current.SMA50.GreaterThan(current.SMA200) &&
				previous.SMA50.LessThanOrEqual(previous.SMA200) {
				current.GoldenCross = true
			}

			// Death Cross: SMA50 crosses below SMA200
			if current.SMA50.LessThan(current.SMA200) &&
				previous.SMA50.GreaterThanOrEqual(previous.SMA200) {
				current.DeathCross = true
			}
		}

		// Price and SMA crossovers
		if !current.SMA10.IsZero() && !previous.SMA10.IsZero() {
			// Price crosses above SMA10
			if current.Close.GreaterThan(current.SMA10) &&
				previous.Close.LessThanOrEqual(previous.SMA10) {
				current.PriceCrossSMA10Up = true
			}

			// Price crosses below SMA10
			if current.Close.LessThan(current.SMA10) &&
				previous.Close.GreaterThanOrEqual(previous.SMA10) {
				current.PriceCrossSMA10Down = true
			}
		}

		// Price and SMA50 crossovers
		if !current.SMA50.IsZero() && !previous.SMA50.IsZero() {
			// Price crosses above SMA50
			if current.Close.GreaterThan(current.SMA50) &&
				previous.Close.LessThanOrEqual(previous.SMA50) {
				current.PriceCrossSMA50Up = true
			}

			// Price crosses below SMA50
			if current.Close.LessThan(current.SMA50) &&
				previous.Close.GreaterThanOrEqual(previous.SMA50) {
				current.PriceCrossSMA50Down = true
			}
		}

		// Calculate SMA slopes
		if !current.SMA10.IsZero() && !previous.SMA10.IsZero() {
			current.SMA10Up = current.SMA10.GreaterThan(previous.SMA10)
		}
		if !current.SMA50.IsZero() && !previous.SMA50.IsZero() {
			current.SMA50Up = current.SMA50.GreaterThan(previous.SMA50)
		}
		if !current.SMA200.IsZero() && !previous.SMA200.IsZero() {
			current.SMA200Up = current.SMA200.GreaterThan(previous.SMA200)
		}

		// Calculate distances between price and SMAs
		if !current.SMA10.IsZero() {
			current.PriceDistanceSMA10 = current.Close.Sub(current.SMA10).Round(2)
		}
		if !current.SMA50.IsZero() {
			current.PriceDistanceSMA50 = current.Close.Sub(current.SMA50).Round(2)
		}
		if !current.SMA200.IsZero() {
			current.PriceDistanceSMA200 = current.Close.Sub(current.SMA200).Round(2)
		}

		// SMA50 above SMA200
		if !current.SMA50.IsZero() && !current.SMA200.IsZero() {
			current.SMA50AboveSMA200 = current.SMA50.GreaterThan(current.SMA200)
		}
	}

	return nil
}

// calculateRSI calculates RSI indicators for periods 9, 14, and 25
func (ic *IndicatorsCalculator) calculateRSI(stockData []*StockDataWithIndicators) error {
	periods := []int{9, 14, 25}

	for _, period := range periods {
		if len(stockData) < period+1 {
			continue // Not enough data for this RSI period
		}

		// Calculate RSI for this period
		for i := period; i < len(stockData); i++ {
			gains := decimal.Zero
			losses := decimal.Zero

			// Calculate average gains and losses over the period
			for j := i - period + 1; j <= i; j++ {
				if j > 0 {
					change := stockData[j].Close.Sub(stockData[j-1].Close)
					if change.GreaterThan(decimal.Zero) {
						gains = gains.Add(change)
					} else {
						losses = losses.Add(change.Abs())
					}
				}
			}

			avgGain := gains.Div(decimal.NewFromInt(int64(period)))
			avgLoss := losses.Div(decimal.NewFromInt(int64(period)))

			if !avgLoss.IsZero() {
				rs := avgGain.Div(avgLoss)
				rsi := decimal.NewFromInt(100).Sub(decimal.NewFromInt(100).Div(decimal.NewFromInt(1).Add(rs)))

				// Assign to appropriate field
				switch period {
				case 9:
					stockData[i].RSI9 = rsi.Round(2)
				case 14:
					stockData[i].RSI14 = rsi.Round(2)
				case 25:
					stockData[i].RSI25 = rsi.Round(2)
				}
			}
		}
	}

	return nil
}

// calculateBasicIndicators calculates basic indicators without descriptions (for calculate_num mode)
func (ic *IndicatorsCalculator) calculateBasicIndicators(stockData []*StockDataWithIndicators) error {
	// Calculate SMA
	if err := ic.calculateSMA(stockData); err != nil {
		return err
	}

	// Calculate RSI
	if err := ic.calculateRSI(stockData); err != nil {
		return err
	}

	// Calculate basic MACD
	if err := ic.calculateMACD(stockData); err != nil {
		return err
	}

	// Calculate basic EMAs
	if err := ic.calculateEMAs(stockData); err != nil {
		return err
	}

	// Calculate other basic indicators
	if err := ic.calculateStochasticOscillator(stockData); err != nil {
		return err
	}

	if err := ic.calculateCMF(stockData); err != nil {
		return err
	}

	if err := ic.calculateOBV(stockData); err != nil {
		return err
	}

	if err := ic.calculatePSAR(stockData); err != nil {
		return err
	}

	if err := ic.calculateATR(stockData); err != nil {
		return err
	}

	if err := ic.calculateRollingStd(stockData); err != nil {
		return err
	}

	return nil
}

// calculateEMAs calculates Exponential Moving Averages
func (ic *IndicatorsCalculator) calculateEMAs(stockData []*StockDataWithIndicators) error {
	if len(stockData) == 0 {
		return nil
	}

	// Calculate EMA5, EMA10, EMA20, EMA50, EMA200
	periods := []int{5, 10, 20, 50, 200}

	for _, period := range periods {
		if len(stockData) < period {
			continue
		}

		multiplier := decimal.NewFromFloat(2.0).Div(decimal.NewFromInt(int64(period + 1)))

		// Initialize first EMA value with first close price
		var emaValues []decimal.Decimal
		emaValues = append(emaValues, stockData[0].Close)

		for i := 1; i < len(stockData); i++ {
			ema := stockData[i].Close.Mul(multiplier).Add(emaValues[i-1].Mul(decimal.NewFromInt(1).Sub(multiplier)))
			emaValues = append(emaValues, ema)

			// Assign to appropriate field
			switch period {
			case 5:
				stockData[i].EMA5 = ema.Round(2)
			case 10:
				stockData[i].EMA10 = ema.Round(2)
			case 20:
				stockData[i].EMA20 = ema.Round(2)
			case 50:
				stockData[i].EMA50 = ema.Round(2)
			case 200:
				stockData[i].EMA200 = ema.Round(2)
			}
		}
	}

	return nil
}

// calculateMACD calculates MACD indicator
func (ic *IndicatorsCalculator) calculateMACD(stockData []*StockDataWithIndicators) error {
	if len(stockData) < 26 {
		return nil // Not enough data for MACD
	}

	// Calculate EMA12 and EMA26 for MACD
	ema12 := make([]decimal.Decimal, len(stockData))
	ema26 := make([]decimal.Decimal, len(stockData))

	// Initialize first EMA values
	ema12[0] = stockData[0].Close
	ema26[0] = stockData[0].Close

	multiplier12 := decimal.NewFromFloat(2.0).Div(decimal.NewFromInt(13)) // 2/(12+1)
	multiplier26 := decimal.NewFromFloat(2.0).Div(decimal.NewFromInt(27)) // 2/(26+1)

	for i := 1; i < len(stockData); i++ {
		// EMA12
		ema12[i] = stockData[i].Close.Mul(multiplier12).Add(ema12[i-1].Mul(decimal.NewFromInt(1).Sub(multiplier12)))

		// EMA26
		ema26[i] = stockData[i].Close.Mul(multiplier26).Add(ema26[i-1].Mul(decimal.NewFromInt(1).Sub(multiplier26)))

		// MACD Line = EMA12 - EMA26
		if i >= 25 { // Start calculating MACD after we have enough data for EMA26
			stockData[i].MACD = ema12[i].Sub(ema26[i]).Round(4)
		}
	}

	// Calculate MACD Signal Line (EMA9 of MACD)
	if len(stockData) >= 34 { // Need at least 34 periods for signal line
		macdSignal := make([]decimal.Decimal, len(stockData))
		multiplier9 := decimal.NewFromFloat(2.0).Div(decimal.NewFromInt(10)) // 2/(9+1)

		// Find first non-zero MACD value for initialization
		firstMACDIndex := 25
		macdSignal[firstMACDIndex] = stockData[firstMACDIndex].MACD

		for i := firstMACDIndex + 1; i < len(stockData); i++ {
			if !stockData[i].MACD.IsZero() {
				macdSignal[i] = stockData[i].MACD.Mul(multiplier9).Add(macdSignal[i-1].Mul(decimal.NewFromInt(1).Sub(multiplier9)))
				stockData[i].MACDSignal = macdSignal[i].Round(4)

				// MACD Histogram = MACD - Signal
				stockData[i].MACDHist = stockData[i].MACD.Sub(stockData[i].MACDSignal).Round(4)
			}
		}
	}

	return nil
}

// Placeholder implementations for remaining indicators
func (ic *IndicatorsCalculator) calculateStochasticOscillator(stockData []*StockDataWithIndicators) error {
	// Basic Stochastic Oscillator implementation
	period := 9
	if len(stockData) < period {
		return nil
	}

	for i := period - 1; i < len(stockData); i++ {
		// Find highest high and lowest low in the period
		highestHigh := stockData[i-period+1].High
		lowestLow := stockData[i-period+1].Low

		for j := i - period + 2; j <= i; j++ {
			if stockData[j].High.GreaterThan(highestHigh) {
				highestHigh = stockData[j].High
			}
			if stockData[j].Low.LessThan(lowestLow) {
				lowestLow = stockData[j].Low
			}
		}

		// Calculate %K
		if !highestHigh.Equal(lowestLow) {
			k := stockData[i].Close.Sub(lowestLow).Div(highestHigh.Sub(lowestLow)).Mul(decimal.NewFromInt(100))
			stockData[i].StochK = k.Round(2)
		}
	}

	// Calculate %D (3-period SMA of %K)
	for i := period + 1; i < len(stockData); i++ {
		if i >= period+2 { // Need at least 3 %K values
			sum := stockData[i].StochK.Add(stockData[i-1].StochK).Add(stockData[i-2].StochK)
			stockData[i].StochD = sum.Div(decimal.NewFromInt(3)).Round(2)
		}
	}

	return nil
}

func (ic *IndicatorsCalculator) calculateCMF(stockData []*StockDataWithIndicators) error {
	// Chaikin Money Flow implementation
	period := 20
	if len(stockData) < period {
		return nil
	}

	for i := period - 1; i < len(stockData); i++ {
		var sumMFV, sumVolume decimal.Decimal

		for j := i - period + 1; j <= i; j++ {
			// Money Flow Multiplier = ((Close - Low) - (High - Close)) / (High - Low)
			if !stockData[j].High.Equal(stockData[j].Low) {
				mfm := stockData[j].Close.Sub(stockData[j].Low).Sub(stockData[j].High.Sub(stockData[j].Close)).Div(stockData[j].High.Sub(stockData[j].Low))
				mfv := mfm.Mul(decimal.NewFromInt(stockData[j].Volume))
				sumMFV = sumMFV.Add(mfv)
			}
			sumVolume = sumVolume.Add(decimal.NewFromInt(stockData[j].Volume))
		}

		if !sumVolume.IsZero() {
			stockData[i].CMF20 = sumMFV.Div(sumVolume).Round(4)
		}
	}

	return nil
}

func (ic *IndicatorsCalculator) calculateOBV(stockData []*StockDataWithIndicators) error {
	// On-Balance Volume implementation
	if len(stockData) == 0 {
		return nil
	}

	stockData[0].OBV = decimal.NewFromInt(stockData[0].Volume)

	for i := 1; i < len(stockData); i++ {
		if stockData[i].Close.GreaterThan(stockData[i-1].Close) {
			stockData[i].OBV = stockData[i-1].OBV.Add(decimal.NewFromInt(stockData[i].Volume))
		} else if stockData[i].Close.LessThan(stockData[i-1].Close) {
			stockData[i].OBV = stockData[i-1].OBV.Sub(decimal.NewFromInt(stockData[i].Volume))
		} else {
			stockData[i].OBV = stockData[i-1].OBV
		}
	}

	// Calculate OBV Rate of Change
	for i := 10; i < len(stockData); i++ {
		if !stockData[i-10].OBV.IsZero() {
			roc := stockData[i].OBV.Sub(stockData[i-10].OBV).Div(stockData[i-10].OBV.Abs()).Mul(decimal.NewFromInt(100))
			stockData[i].OBVRoC = roc.Round(2)
		}
	}

	return nil
}

func (ic *IndicatorsCalculator) calculatePSAR(stockData []*StockDataWithIndicators) error {
	// Parabolic SAR implementation (basic version)
	if len(stockData) < 2 {
		return nil
	}

	// PSAR with acceleration 0.02, maximum 0.2
	acceleration := decimal.NewFromFloat(0.02)
	maxAcceleration := decimal.NewFromFloat(0.2)

	// Initialize
	stockData[0].PSAR1 = stockData[0].Low
	isUptrend := true
	af := acceleration
	ep := stockData[0].High // Extreme Point

	for i := 1; i < len(stockData); i++ {
		// Calculate PSAR
		psar := stockData[i-1].PSAR1.Add(af.Mul(ep.Sub(stockData[i-1].PSAR1)))

		if isUptrend {
			if stockData[i].Low.LessThan(psar) {
				// Trend reversal
				isUptrend = false
				psar = ep
				ep = stockData[i].Low
				af = acceleration
			} else {
				if stockData[i].High.GreaterThan(ep) {
					ep = stockData[i].High
					af = af.Add(acceleration)
					if af.GreaterThan(maxAcceleration) {
						af = maxAcceleration
					}
				}
			}
		} else {
			if stockData[i].High.GreaterThan(psar) {
				// Trend reversal
				isUptrend = true
				psar = ep
				ep = stockData[i].High
				af = acceleration
			} else {
				if stockData[i].Low.LessThan(ep) {
					ep = stockData[i].Low
					af = af.Add(acceleration)
					if af.GreaterThan(maxAcceleration) {
						af = maxAcceleration
					}
				}
			}
		}

		stockData[i].PSAR1 = psar.Round(4)
	}

	// Calculate PSAR2 with different parameters (0.01, 0.1)
	acceleration2 := decimal.NewFromFloat(0.01)
	maxAcceleration2 := decimal.NewFromFloat(0.1)

	stockData[0].PSAR2 = stockData[0].Low
	isUptrend2 := true
	af2 := acceleration2
	ep2 := stockData[0].High

	for i := 1; i < len(stockData); i++ {
		psar := stockData[i-1].PSAR2.Add(af2.Mul(ep2.Sub(stockData[i-1].PSAR2)))

		if isUptrend2 {
			if stockData[i].Low.LessThan(psar) {
				isUptrend2 = false
				psar = ep2
				ep2 = stockData[i].Low
				af2 = acceleration2
			} else {
				if stockData[i].High.GreaterThan(ep2) {
					ep2 = stockData[i].High
					af2 = af2.Add(acceleration2)
					if af2.GreaterThan(maxAcceleration2) {
						af2 = maxAcceleration2
					}
				}
			}
		} else {
			if stockData[i].High.GreaterThan(psar) {
				isUptrend2 = true
				psar = ep2
				ep2 = stockData[i].High
				af2 = acceleration2
			} else {
				if stockData[i].Low.LessThan(ep2) {
					ep2 = stockData[i].Low
					af2 = af2.Add(acceleration2)
					if af2.GreaterThan(maxAcceleration2) {
						af2 = maxAcceleration2
					}
				}
			}
		}

		stockData[i].PSAR2 = psar.Round(4)
	}

	return nil
}

func (ic *IndicatorsCalculator) calculateATR(stockData []*StockDataWithIndicators) error {
	// Average True Range implementation
	period := 14
	if len(stockData) < period {
		return nil
	}

	// Calculate True Range for each period
	trueRanges := make([]decimal.Decimal, len(stockData))

	for i := 1; i < len(stockData); i++ {
		// TR = max(High - Low, |High - PrevClose|, |Low - PrevClose|)
		hl := stockData[i].High.Sub(stockData[i].Low)
		hpc := stockData[i].High.Sub(stockData[i-1].Close).Abs()
		lpc := stockData[i].Low.Sub(stockData[i-1].Close).Abs()

		tr := hl
		if hpc.GreaterThan(tr) {
			tr = hpc
		}
		if lpc.GreaterThan(tr) {
			tr = lpc
		}

		trueRanges[i] = tr
	}

	// Calculate ATR (Simple Moving Average of True Range)
	for i := period; i < len(stockData); i++ {
		sum := decimal.Zero
		for j := i - period + 1; j <= i; j++ {
			sum = sum.Add(trueRanges[j])
		}
		stockData[i].ATR = sum.Div(decimal.NewFromInt(int64(period))).Round(4)
	}

	return nil
}

func (ic *IndicatorsCalculator) calculateRollingStd(stockData []*StockDataWithIndicators) error {
	// Rolling Standard Deviation implementation
	periods := []int{10, 50}

	for _, period := range periods {
		if len(stockData) < period {
			continue
		}

		for i := period - 1; i < len(stockData); i++ {
			// Calculate mean
			sum := decimal.Zero
			for j := i - period + 1; j <= i; j++ {
				sum = sum.Add(stockData[j].Close)
			}
			mean := sum.Div(decimal.NewFromInt(int64(period)))

			// Calculate variance
			variance := decimal.Zero
			for j := i - period + 1; j <= i; j++ {
				diff := stockData[j].Close.Sub(mean)
				variance = variance.Add(diff.Mul(diff))
			}
			variance = variance.Div(decimal.NewFromInt(int64(period)))

			// Standard deviation is square root of variance
			// Using approximation for decimal square root
			std := ic.sqrt(variance)

			switch period {
			case 10:
				stockData[i].RollingStd10 = std.Round(4)
			case 50:
				stockData[i].RollingStd50 = std.Round(4)
			}
		}
	}

	return nil
}

// sqrt calculates square root using Newton's method for decimal
func (ic *IndicatorsCalculator) sqrt(x decimal.Decimal) decimal.Decimal {
	if x.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}

	// Newton's method for square root
	guess := x.Div(decimal.NewFromInt(2))
	for i := 0; i < 10; i++ { // 10 iterations should be enough for precision
		newGuess := guess.Add(x.Div(guess)).Div(decimal.NewFromInt(2))
		if newGuess.Sub(guess).Abs().LessThan(decimal.NewFromFloat(0.0001)) {
			break
		}
		guess = newGuess
	}

	return guess
}

// addDescriptions adds descriptive text for indicators (for full mode)
func (ic *IndicatorsCalculator) addDescriptions(stockData []*StockDataWithIndicators) {
	for _, data := range stockData {
		// Golden/Death Cross descriptions
		if data.GoldenCross {
			data.GoldenDeathCrossDesc = "Buy - Golden Cross Detected: The shorter-term SMA50 has crossed above the longer-term SMA200; a bullish signal. Historically this pattern indicates the early stages of a prolonged bull market. Interpretation: The asset's price might rise; suggesting a favorable time to buy or add to your position."
		} else if data.DeathCross {
			data.GoldenDeathCrossDesc = "Sell - Death Cross Detected: The shorter-term SMA50 has crossed below the longer-term SMA200; a bearish signal. This pattern often precedes a forthcoming bear market or a prolonged period of selling. Interpretation: Exercise caution; consider reducing exposure or hedge against losses."
		} else {
			data.GoldenDeathCrossDesc = "Neutral - No Significant Cross Detected: The market shows no clear bullish or bearish signals at the moment. This can indicate a period of consolidation or sideways movement. Interpretation: Monitor other indicators; stay updated with market news and maintain a diversified strategy."
		}

		// Price SMA10 crossover descriptions
		if data.PriceCrossSMA10Up {
			data.PriceSMA10CrossoverDesc = "Buy - Price Crossed Above SMA10: The asset's price has surged above its 10-period average. This upward crossover is historically a sign of short-term bullish momentum. Interpretation: It might be an opportunity to capitalize on the momentum; but also consider other indicators for confirmation."
		} else if data.PriceCrossSMA10Down {
			data.PriceSMA10CrossoverDesc = "Sell - Price Crossed Below SMA10: The asset's price is dipping below its recent 10-period average. This can hint at a short-term decline or a potential pullback. Interpretation: It might be wise to exercise caution; adjust strategies or set stop losses."
		} else {
			data.PriceSMA10CrossoverDesc = "Neutral - Price Oscillating Around SMA10: The asset's price is weaving around its 10-period average; indicating market indecision. This pattern could be a sign of consolidation. Interpretation: Stay alert; monitor other indicators and be ready for a potential breakout."
		}

		// Price crossover descriptions (SMA50 and SMA200)
		if data.PriceCrossSMA50Up || data.PriceCrossSMA200Up {
			data.PriceCrossoverDesc = "Buy; Price Crossed Above Major SMA: The asset's price has crossed above a significant moving average - indicating strong bullish momentum. This suggests institutional interest and potential trend continuation. Interpretation: Consider entering long positions with proper risk management."
		} else if data.PriceCrossSMA50Down || data.PriceCrossSMA200Down {
			data.PriceCrossoverDesc = "Sell; Price Crossed Below Major SMA: The asset's price has fallen below a key moving average - signaling potential weakness. This could indicate the start of a downtrend. Interpretation: Consider reducing positions or implementing protective stops."
		} else {
			data.PriceCrossoverDesc = "Neutral; Price Respecting SMA Levels: The asset's price is trading in line with major moving averages - showing balanced market conditions. Interpretation: Wait for clearer directional signals before making significant position changes."
		}

		// RSI descriptions
		if !data.RSI14.IsZero() {
			rsi := data.RSI14
			if rsi.GreaterThan(decimal.NewFromInt(70)) {
				data.RSIDesc = "Sell - RSI Overbought: RSI is above 70; indicating the asset may be overbought. This suggests selling pressure could emerge soon. Interpretation: Consider taking profits or reducing positions; but watch for trend continuation in strong markets."
			} else if rsi.LessThan(decimal.NewFromInt(30)) {
				data.RSIDesc = "Buy - RSI Oversold: RSI is below 30; indicating the asset may be oversold. This suggests a potential bounce or reversal. Interpretation: Look for buying opportunities; but confirm with other indicators and price action."
			} else if rsi.GreaterThan(decimal.NewFromInt(50)) {
				data.RSIDesc = "Neutral-Bullish - RSI Above Midline: RSI is above 50; showing bullish momentum but not extreme. The trend appears healthy with room for further upside. Interpretation: Maintain bullish bias but monitor for overbought conditions."
			} else {
				data.RSIDesc = "Neutral-Bearish - RSI Below Midline: RSI is below 50; indicating bearish momentum but not extreme oversold. The trend shows weakness with potential for further decline. Interpretation: Exercise caution and look for confirmation before buying."
			}
		}

		// Stochastic descriptions
		if !data.StochK.IsZero() && !data.StochD.IsZero() {
			stochK := data.StochK
			stochD := data.StochD
			if stochK.GreaterThan(decimal.NewFromInt(80)) && stochD.GreaterThan(decimal.NewFromInt(80)) {
				data.StochasticDesc = "Sell; Stochastic Overbought: Both %K and %D are above 80 - indicating overbought conditions. Momentum may be slowing. Interpretation: Consider profit-taking or tightening stops as a pullback may be imminent."
			} else if stochK.LessThan(decimal.NewFromInt(20)) && stochD.LessThan(decimal.NewFromInt(20)) {
				data.StochasticDesc = "Buy; Stochastic Oversold: Both %K and %D are below 20 - indicating oversold conditions. A bounce may be developing. Interpretation: Look for buying opportunities on confirmation of upward momentum."
			} else if stochK.GreaterThan(stochD) {
				data.StochasticDesc = "Neutral-Bullish; Stochastic Bullish Crossover: %K is above %D - indicating bullish momentum. The trend appears to be strengthening. Interpretation: Monitor for continuation of upward movement."
			} else {
				data.StochasticDesc = "Neutral-Bearish; Stochastic Bearish Crossover: %K is below %D - indicating bearish momentum. Weakness may be developing. Interpretation: Exercise caution and consider defensive positioning."
			}
		}

		// CMF descriptions
		if !data.CMF.IsZero() {
			cmf := data.CMF
			if cmf.GreaterThan(decimal.NewFromFloat(0.1)) {
				data.CMFDesc = "Buy; Strong Money Flow: CMF is positive and strong - indicating accumulation by institutional investors. Money is flowing into the asset. Interpretation: This supports bullish price action and suggests buying interest."
			} else if cmf.LessThan(decimal.NewFromFloat(-0.1)) {
				data.CMFDesc = "Sell; Weak Money Flow: CMF is negative and weak - indicating distribution by institutional investors. Money is flowing out of the asset. Interpretation: This supports bearish price action and suggests selling pressure."
			} else if cmf.GreaterThan(decimal.Zero) {
				data.CMFDesc = "Neutral-Bullish, Mild Accumulation: CMF is slightly positive; showing mild buying interest. The trend is supported but not strongly. Interpretation: Cautiously bullish - but look for stronger confirmation."
			} else {
				data.CMFDesc = "Neutral-Bearish; Mild Distribution: CMF is slightly negative - showing mild selling pressure. The trend shows some weakness. Interpretation: Exercise caution and monitor for trend deterioration."
			}
		}

		// MACD descriptions
		if !data.MACD.IsZero() && !data.MACDSignal.IsZero() {
			macd := data.MACD
			signal := data.MACDSignal
			histogram := data.MACDHist

			if macd.GreaterThan(signal) && histogram.GreaterThan(decimal.Zero) {
				data.MACDDesc = "Buy; MACD Bullish: MACD line is above signal line with positive histogram - indicating strong bullish momentum. The trend is accelerating upward. Interpretation: Consider entering long positions or adding to existing bullish positions."
			} else if macd.LessThan(signal) && histogram.LessThan(decimal.Zero) {
				data.MACDDesc = "Sell; MACD Bearish: MACD line is below signal line with negative histogram - indicating strong bearish momentum. The trend is accelerating downward. Interpretation: Consider reducing positions or entering short positions."
			} else if macd.GreaterThan(signal) {
				data.MACDDesc = "Neutral-Bullish - MACD Above Signal: MACD is above signal line but momentum is weakening. Bullish trend may be losing steam. Interpretation: Maintain bullish bias but watch for potential reversal signals."
			} else {
				data.MACDDesc = "Neutral-Bearish - MACD Below Signal: MACD is below signal line but momentum is weakening. Bearish trend may be losing steam. Interpretation: Maintain bearish bias but watch for potential reversal signals."
			}
		}

		// OBV descriptions
		if !data.OBV.IsZero() {
			obvRoc := data.OBVRoC
			if obvRoc.GreaterThan(decimal.NewFromInt(10)) {
				data.OBVDesc = "Buy; Strong Volume Accumulation: OBV is rising strongly - indicating heavy accumulation. Smart money is buying aggressively. Interpretation: This supports bullish price action and suggests strong institutional interest."
			} else if obvRoc.LessThan(decimal.NewFromInt(-10)) {
				data.OBVDesc = "Sell; Strong Volume Distribution: OBV is falling strongly - indicating heavy distribution. Smart money is selling aggressively. Interpretation: This supports bearish price action and suggests institutional selling."
			} else if obvRoc.GreaterThan(decimal.Zero) {
				data.OBVDesc = "Neutral-Bullish; Mild Volume Accumulation: OBV is rising moderately - showing steady accumulation. Buying interest is present but not overwhelming. Interpretation: Cautiously bullish volume pattern."
			} else {
				data.OBVDesc = "Neutral-Bearish; Mild Volume Distribution: OBV is declining moderately - showing steady distribution. Selling pressure is present but not overwhelming. Interpretation: Cautiously bearish volume pattern."
			}
		}

		// PSAR descriptions
		if !data.PSAR.IsZero() {
			if data.Close.GreaterThan(data.PSAR) {
				data.PSARDesc = "Buy; PSAR Bullish: Price is above PSAR - indicating an uptrend. The parabolic SAR suggests continued bullish momentum. Interpretation: Trend following systems suggest maintaining long positions with PSAR as trailing stop."
			} else {
				data.PSARDesc = "Sell; PSAR Bearish: Price is below PSAR - indicating a downtrend. The parabolic SAR suggests continued bearish momentum. Interpretation: Trend following systems suggest maintaining short positions or avoiding long positions."
			}
		}

		// ATR descriptions
		if !data.ATR.IsZero() && !data.Close.IsZero() {
			// ATR is used for volatility assessment rather than directional signals
			atr := data.ATR
			pricePercent := atr.Div(data.Close).Mul(decimal.NewFromInt(100))

			if pricePercent.GreaterThan(decimal.NewFromInt(5)) {
				data.ATRDesc = "High Volatility Warning: ATR indicates high volatility (>5% of price). Market conditions are unstable with large price swings. Interpretation: Use wider stops; reduce position sizes - and expect increased risk."
			} else if pricePercent.GreaterThan(decimal.NewFromInt(2)) {
				data.ATRDesc = "Moderate Volatility: ATR shows moderate volatility (2-5% of price). Normal market conditions with reasonable price movement. Interpretation: Standard risk management applies - monitor for volatility changes."
			} else {
				data.ATRDesc = "Low Volatility: ATR indicates low volatility (<2% of price). Market is relatively calm with small price movements. Interpretation: Consider tighter stops - but watch for potential volatility breakouts."
			}
		}
	}
}
