package strategies

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/shopspring/decimal"

	"isx-auto-scrapper/internal/common"
	"isx-auto-scrapper/internal/indicators"
)

// Strategies handles trading strategy analysis
type Strategies struct {
	logger *common.Logger
	config StrategyConfig
}

// NewStrategies creates a new Strategies instance
func NewStrategies() *Strategies {
	cfg, err := loadStrategyConfig("strategy_config.json")
	if err != nil {
		cfg = defaultStrategyConfig
	}
	return &Strategies{
		logger: common.NewLogger(),
		config: cfg,
	}
}

// StrategyData represents stock data with applied strategies
type StrategyData struct {
	indicators.StockDataWithIndicators
	RSIStrategy          string `csv:"RSI Strategy"`
	RSIStrategy2         string `csv:"RSI Strategy2"`
	RSI14OBVRoCStrategy  string `csv:"RSI14_OBV_RoC Strategy"`
	RSIMACDStrategy      string `csv:"RSIMACD Strategy"`
	RSICMFStrategy       string `csv:"RSICMF Strategy"`
	RSIOBVStrategy       string `csv:"RSI OBV Strategy"`
	OBVStrategy          string `csv:"OBV Strategy"`
	MACDStrategy         string `csv:"MACD Strategy"`
	CMFStrategy          string `csv:"CMF Strategy"`
	EMA5PSARStrategy     string `csv:"EMA5 PSAR Strategy"`
	EMA5PSARStrategy2    string `csv:"EMA5 PSAR Strategy2"`
	RollingStd10Strategy string `csv:"Rolling Std10 Strategy"`
	RollingStd50Strategy string `csv:"Rolling Std50 Strategy"`
}

// Strategy signal levels with intermediate states
const (
	StrongBuy  = "Strong Buy"
	Buy        = "Buy"
	WeakBuy    = "Weak Buy"
	Hold       = "Hold"
	WeakSell   = "Weak Sell"
	Sell       = "Sell"
	StrongSell = "Strong Sell"
)

// Levels holds threshold levels for a strategy
type Levels struct {
	StrongBuy  float64 `json:"strong_buy"`
	Buy        float64 `json:"buy"`
	WeakBuy    float64 `json:"weak_buy"`
	WeakSell   float64 `json:"weak_sell"`
	Sell       float64 `json:"sell"`
	StrongSell float64 `json:"strong_sell"`
}

// MACDHistLevels defines thresholds for MACD histogram strength
type MACDHistLevels struct {
	Strong float64 `json:"strong"`
	Buy    float64 `json:"buy"`
}

// StrategyConfig defines tunable strategy thresholds
type StrategyConfig struct {
	RSI      Levels         `json:"rsi"`
	RSI2     Levels         `json:"rsi2"`
	CMF      Levels         `json:"cmf"`
	OBVRoC   Levels         `json:"obvroc"`
	MACDHist MACDHistLevels `json:"macd_hist"`
}

// defaultStrategyConfig provides sane defaults if no file is found
var defaultStrategyConfig = StrategyConfig{
	RSI:      Levels{20, 30, 40, 60, 70, 80},
	RSI2:     Levels{15, 25, 35, 65, 75, 85},
	CMF:      Levels{0.2, 0.1, 0.05, -0.05, -0.1, -0.2},
	OBVRoC:   Levels{10, 5, 2, -2, -5, -10},
	MACDHist: MACDHistLevels{0.1, 0.05},
}

// loadStrategyConfig reads configuration from a JSON file
func loadStrategyConfig(path string) (StrategyConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return StrategyConfig{}, err
	}
	defer f.Close()

	var cfg StrategyConfig
	dec := json.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return StrategyConfig{}, err
	}
	return cfg, nil
}

// ApplyStrategiesAndSave applies strategies and saves results
func (s *Strategies) ApplyStrategiesAndSave() error {
	s.logger.Info("Applying strategies and saving results")

	// Load tickers
	tickers, err := common.LoadTickers("TICKERS.csv")
	if err != nil {
		return fmt.Errorf("failed to load tickers: %w", err)
	}

	if len(tickers) == 0 {
		s.logger.Info("TICKERS.csv is empty. No tickers to process.")
		return nil
	}

	for _, ticker := range tickers {
		filePath := fmt.Sprintf("indicators_%s.csv", ticker)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			s.logger.Error("indicators_%s.csv does not exist", ticker)
			continue
		}

		// Load indicators data
		indicatorData, err := s.loadIndicatorData(filePath)
		if err != nil {
			s.logger.Error("Error loading indicators for %s: %v", ticker, err)
			continue
		}

		// Filter to last 12 months
		oneYearAgo := time.Now().AddDate(-1, 0, 0)
		var filteredData []*indicators.StockDataWithIndicators
		for _, data := range indicatorData {
			if data.Date.After(oneYearAgo) {
				filteredData = append(filteredData, data)
			}
		}

		if len(filteredData) == 0 {
			s.logger.Info("No trading data for %s in the past 12 months.", ticker)
			continue
		}

		// Apply strategies
		strategyData, err := s.applyTradingStrategies(filteredData)
		if err != nil {
			s.logger.Error("Error applying strategies for %s: %v", ticker, err)
			continue
		}

		// Save strategies data
		strategiesFilePath := fmt.Sprintf("Strategies_%s.csv", ticker)
		if err := s.saveStrategiesData(strategyData, strategiesFilePath); err != nil {
			s.logger.Error("Error saving strategies for %s: %v", ticker, err)
			continue
		}

		s.logger.Info("Trading strategies successfully added and saved for %s.", ticker)
	}

	s.logger.Info("Trading strategies successfully added and saved for all processed tickers.")
	return nil
}

// ApplyAlternativeStrategyStates applies alternative strategy states
func (s *Strategies) ApplyAlternativeStrategyStates() error {
	s.logger.Info("Applying alternative strategy states")

	// Load tickers
	tickers, err := common.LoadTickers("TICKERS.csv")
	if err != nil {
		return fmt.Errorf("failed to load tickers: %w", err)
	}

	// List of all strategy columns
	strategies := []string{
		"RSI Strategy", "RSI Strategy2", "RSI14_OBV_RoC Strategy", "RSIMACD Strategy",
		"RSICMF Strategy", "RSI OBV Strategy", "OBV Strategy", "MACD Strategy",
		"CMF Strategy", "EMA5 PSAR Strategy", "EMA5 PSAR Strategy2",
		"Rolling Std10 Strategy", "Rolling Std50 Strategy",
	}

	for _, ticker := range tickers {
		strategiesFilePath := fmt.Sprintf("Strategies_%s.csv", ticker)
		if _, err := os.Stat(strategiesFilePath); os.IsNotExist(err) {
			s.logger.Error("Strategies_%s.csv does not exist", ticker)
			continue
		}

		// Load and process alternative states
		if err := s.processAlternativeStates(strategiesFilePath, strategies); err != nil {
			s.logger.Error("Error processing alternative states for %s: %v", ticker, err)
			continue
		}

		s.logger.Info("Alternative strategy states applied for %s", ticker)
	}

	return nil
}

// SummarizeStrategyActions summarizes strategy actions
func (s *Strategies) SummarizeStrategyActions() error {
	s.logger.Info("Summarizing strategy actions")

	// Load tickers
	tickers, err := common.LoadTickers("TICKERS.csv")
	if err != nil {
		return fmt.Errorf("failed to load tickers: %w", err)
	}

	allSummaries := make(map[string]interface{})

	for _, ticker := range tickers {
		strategiesFilePath := fmt.Sprintf("Strategies_%s.csv", ticker)
		if _, err := os.Stat(strategiesFilePath); os.IsNotExist(err) {
			continue
		}

		summary, err := s.generateStrategySummary(ticker, strategiesFilePath)
		if err != nil {
			s.logger.Error("Error generating summary for %s: %v", ticker, err)
			continue
		}

		allSummaries[ticker] = summary
	}

	// Save summary to file
	summaryFilePath := "Strategy_Summary.json"
	if err := s.saveSummaryToFile(allSummaries, summaryFilePath); err != nil {
		return fmt.Errorf("failed to save strategy summary: %w", err)
	}

	s.logger.Info("Strategy actions summarized and saved to %s", summaryFilePath)
	return nil
}

// loadIndicatorData loads indicator data from CSV file
func (s *Strategies) loadIndicatorData(filePath string) ([]*indicators.StockDataWithIndicators, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var stockData []*indicators.StockDataWithIndicators
	if err := gocsv.UnmarshalFile(file, &stockData); err != nil {
		return nil, err
	}

	return stockData, nil
}

// applyTradingStrategies applies trading strategies to the data with intermediate states
func (s *Strategies) applyTradingStrategies(data []*indicators.StockDataWithIndicators) ([]*StrategyData, error) {
	var strategyData []*StrategyData

	for _, stock := range data {
		strategy := &StrategyData{
			StockDataWithIndicators: *stock,
			// Initialize all strategies to "Hold"
			RSIStrategy:          Hold,
			RSIStrategy2:         Hold,
			RSI14OBVRoCStrategy:  Hold,
			RSIMACDStrategy:      Hold,
			RSICMFStrategy:       Hold,
			RSIOBVStrategy:       Hold,
			OBVStrategy:          Hold,
			MACDStrategy:         Hold,
			CMFStrategy:          Hold,
			EMA5PSARStrategy:     Hold,
			EMA5PSARStrategy2:    Hold,
			RollingStd10Strategy: Hold,
			RollingStd50Strategy: Hold,
		}
		strategyData = append(strategyData, strategy)
	}

	// Apply all strategies with enhanced signal levels
	s.applyRSIStrategy(strategyData)
	s.applyMACDStrategy(strategyData)
	s.applyCMFStrategy(strategyData)
	s.applyOBVStrategy(strategyData)
	s.applyEMA5PSARStrategy(strategyData)
	s.applyRollingStdStrategies(strategyData)

	return strategyData, nil
}

// applyRSIStrategy applies RSI-based strategies with intermediate states
func (s *Strategies) applyRSIStrategy(data []*StrategyData) {
	for _, d := range data {
		rsi := d.RSI14
		if rsi.IsZero() {
			continue
		}

		// RSI Strategy: Enhanced with intermediate states
		if rsi.LessThan(decimal.NewFromFloat(s.config.RSI.StrongBuy)) {
			d.RSIStrategy = StrongBuy
		} else if rsi.LessThan(decimal.NewFromFloat(s.config.RSI.Buy)) {
			d.RSIStrategy = Buy
		} else if rsi.LessThan(decimal.NewFromFloat(s.config.RSI.WeakBuy)) {
			d.RSIStrategy = WeakBuy
		} else if rsi.GreaterThan(decimal.NewFromFloat(s.config.RSI.StrongSell)) {
			d.RSIStrategy = StrongSell
		} else if rsi.GreaterThan(decimal.NewFromFloat(s.config.RSI.Sell)) {
			d.RSIStrategy = Sell
		} else if rsi.GreaterThan(decimal.NewFromFloat(s.config.RSI.WeakSell)) {
			d.RSIStrategy = WeakSell
		}

		// RSI Strategy2: More conservative thresholds
		if rsi.LessThan(decimal.NewFromFloat(s.config.RSI2.StrongBuy)) {
			d.RSIStrategy2 = StrongBuy
		} else if rsi.LessThan(decimal.NewFromFloat(s.config.RSI2.Buy)) {
			d.RSIStrategy2 = Buy
		} else if rsi.LessThan(decimal.NewFromFloat(s.config.RSI2.WeakBuy)) {
			d.RSIStrategy2 = WeakBuy
		} else if rsi.GreaterThan(decimal.NewFromFloat(s.config.RSI2.StrongSell)) {
			d.RSIStrategy2 = StrongSell
		} else if rsi.GreaterThan(decimal.NewFromFloat(s.config.RSI2.Sell)) {
			d.RSIStrategy2 = Sell
		} else if rsi.GreaterThan(decimal.NewFromFloat(s.config.RSI2.WeakSell)) {
			d.RSIStrategy2 = WeakSell
		}
	}
}

// applyMACDStrategy applies MACD-based strategies with signal strength
func (s *Strategies) applyMACDStrategy(data []*StrategyData) {
	for _, d := range data {
		macd := d.MACD
		signal := d.MACDSignal
		histogram := d.MACDHist

		if macd.IsZero() || signal.IsZero() {
			continue
		}

		// Calculate signal strength based on MACD histogram
		histogramAbs := histogram.Abs()

		// MACD Strategy: Enhanced with signal strength
		if macd.GreaterThan(signal) {
			if histogramAbs.GreaterThan(decimal.NewFromFloat(s.config.MACDHist.Strong)) {
				d.MACDStrategy = StrongBuy
			} else if histogramAbs.GreaterThan(decimal.NewFromFloat(s.config.MACDHist.Buy)) {
				d.MACDStrategy = Buy
			} else {
				d.MACDStrategy = WeakBuy
			}
		} else if macd.LessThan(signal) {
			if histogramAbs.GreaterThan(decimal.NewFromFloat(s.config.MACDHist.Strong)) {
				d.MACDStrategy = StrongSell
			} else if histogramAbs.GreaterThan(decimal.NewFromFloat(s.config.MACDHist.Buy)) {
				d.MACDStrategy = Sell
			} else {
				d.MACDStrategy = WeakSell
			}
		}

		// Combined RSI+MACD Strategy
		rsi := d.RSI14
		if !rsi.IsZero() {
			if macd.GreaterThan(signal) && rsi.LessThan(decimal.NewFromInt(70)) {
				if rsi.LessThan(decimal.NewFromInt(30)) && histogramAbs.GreaterThan(decimal.NewFromFloat(0.1)) {
					d.RSIMACDStrategy = StrongBuy
				} else if rsi.LessThan(decimal.NewFromInt(50)) {
					d.RSIMACDStrategy = Buy
				} else {
					d.RSIMACDStrategy = WeakBuy
				}
			} else if macd.LessThan(signal) && rsi.GreaterThan(decimal.NewFromInt(30)) {
				if rsi.GreaterThan(decimal.NewFromInt(70)) && histogramAbs.GreaterThan(decimal.NewFromFloat(0.1)) {
					d.RSIMACDStrategy = StrongSell
				} else if rsi.GreaterThan(decimal.NewFromInt(50)) {
					d.RSIMACDStrategy = Sell
				} else {
					d.RSIMACDStrategy = WeakSell
				}
			}
		}
	}
}

// applyCMFStrategy applies CMF-based strategies with enhanced signal levels
func (s *Strategies) applyCMFStrategy(data []*StrategyData) {
	for _, d := range data {
		cmf := d.CMF
		if cmf.IsZero() {
			continue
		}

		// CMF Strategy: Enhanced money flow signals
		if cmf.GreaterThan(decimal.NewFromFloat(s.config.CMF.StrongBuy)) {
			d.CMFStrategy = StrongBuy
		} else if cmf.GreaterThan(decimal.NewFromFloat(s.config.CMF.Buy)) {
			d.CMFStrategy = Buy
		} else if cmf.GreaterThan(decimal.NewFromFloat(s.config.CMF.WeakBuy)) {
			d.CMFStrategy = WeakBuy
		} else if cmf.LessThan(decimal.NewFromFloat(s.config.CMF.StrongSell)) {
			d.CMFStrategy = StrongSell
		} else if cmf.LessThan(decimal.NewFromFloat(s.config.CMF.Sell)) {
			d.CMFStrategy = Sell
		} else if cmf.LessThan(decimal.NewFromFloat(s.config.CMF.WeakSell)) {
			d.CMFStrategy = WeakSell
		}

		// Combined RSI+CMF Strategy
		rsi := d.RSI14
		if !rsi.IsZero() {
			if cmf.GreaterThan(decimal.NewFromFloat(s.config.CMF.Buy)) && rsi.LessThan(decimal.NewFromInt(70)) {
				if rsi.LessThan(decimal.NewFromInt(30)) {
					d.RSICMFStrategy = StrongBuy
				} else if rsi.LessThan(decimal.NewFromInt(50)) {
					d.RSICMFStrategy = Buy
				} else {
					d.RSICMFStrategy = WeakBuy
				}
			} else if cmf.LessThan(decimal.NewFromFloat(s.config.CMF.Sell)) && rsi.GreaterThan(decimal.NewFromInt(30)) {
				if rsi.GreaterThan(decimal.NewFromInt(70)) {
					d.RSICMFStrategy = StrongSell
				} else if rsi.GreaterThan(decimal.NewFromInt(50)) {
					d.RSICMFStrategy = Sell
				} else {
					d.RSICMFStrategy = WeakSell
				}
			}
		}
	}
}

// applyOBVStrategy applies OBV-based strategies with volume momentum levels
func (s *Strategies) applyOBVStrategy(data []*StrategyData) {
	for _, d := range data {
		obvRoc := d.OBVRoC
		if obvRoc.IsZero() {
			continue
		}

		// OBV Strategy: Enhanced volume momentum signals
		if obvRoc.GreaterThan(decimal.NewFromFloat(s.config.OBVRoC.StrongBuy)) {
			d.OBVStrategy = StrongBuy
		} else if obvRoc.GreaterThan(decimal.NewFromFloat(s.config.OBVRoC.Buy)) {
			d.OBVStrategy = Buy
		} else if obvRoc.GreaterThan(decimal.NewFromFloat(s.config.OBVRoC.WeakBuy)) {
			d.OBVStrategy = WeakBuy
		} else if obvRoc.LessThan(decimal.NewFromFloat(s.config.OBVRoC.StrongSell)) {
			d.OBVStrategy = StrongSell
		} else if obvRoc.LessThan(decimal.NewFromFloat(s.config.OBVRoC.Sell)) {
			d.OBVStrategy = Sell
		} else if obvRoc.LessThan(decimal.NewFromFloat(s.config.OBVRoC.WeakSell)) {
			d.OBVStrategy = WeakSell
		}

		// Combined RSI+OBV Strategy
		rsi := d.RSI14
		if !rsi.IsZero() {
			if obvRoc.GreaterThan(decimal.NewFromFloat(s.config.OBVRoC.Buy)) && rsi.LessThan(decimal.NewFromInt(70)) {
				if rsi.LessThan(decimal.NewFromInt(30)) {
					d.RSIOBVStrategy = StrongBuy
					d.RSI14OBVRoCStrategy = StrongBuy
				} else if rsi.LessThan(decimal.NewFromInt(50)) {
					d.RSIOBVStrategy = Buy
					d.RSI14OBVRoCStrategy = Buy
				} else {
					d.RSIOBVStrategy = WeakBuy
					d.RSI14OBVRoCStrategy = WeakBuy
				}
			} else if obvRoc.LessThan(decimal.NewFromFloat(s.config.OBVRoC.Sell)) && rsi.GreaterThan(decimal.NewFromInt(30)) {
				if rsi.GreaterThan(decimal.NewFromInt(70)) {
					d.RSIOBVStrategy = StrongSell
					d.RSI14OBVRoCStrategy = StrongSell
				} else if rsi.GreaterThan(decimal.NewFromInt(50)) {
					d.RSIOBVStrategy = Sell
					d.RSI14OBVRoCStrategy = Sell
				} else {
					d.RSIOBVStrategy = WeakSell
					d.RSI14OBVRoCStrategy = WeakSell
				}
			}
		}
	}
}

// applyEMA5PSARStrategy applies EMA5+PSAR trend following strategies
func (s *Strategies) applyEMA5PSARStrategy(data []*StrategyData) {
	for _, d := range data {
		ema5 := d.EMA5
		psar := d.PSAR
		price := d.Close

		if ema5.IsZero() || psar.IsZero() || price.IsZero() {
			continue
		}

		// Calculate how far price is from EMA5 for signal strength
		emaDistance := price.Sub(ema5).Div(ema5).Mul(decimal.NewFromInt(100))

		// EMA5 PSAR Strategy: Enhanced trend following
		if price.GreaterThan(ema5) && price.GreaterThan(psar) {
			if emaDistance.GreaterThan(decimal.NewFromInt(5)) {
				d.EMA5PSARStrategy = StrongBuy
				d.EMA5PSARStrategy2 = StrongBuy
			} else if emaDistance.GreaterThan(decimal.NewFromInt(2)) {
				d.EMA5PSARStrategy = Buy
				d.EMA5PSARStrategy2 = Buy
			} else {
				d.EMA5PSARStrategy = WeakBuy
				d.EMA5PSARStrategy2 = WeakBuy
			}
		} else if price.LessThan(ema5) && price.LessThan(psar) {
			if emaDistance.LessThan(decimal.NewFromInt(-5)) {
				d.EMA5PSARStrategy = StrongSell
				d.EMA5PSARStrategy2 = StrongSell
			} else if emaDistance.LessThan(decimal.NewFromInt(-2)) {
				d.EMA5PSARStrategy = Sell
				d.EMA5PSARStrategy2 = Sell
			} else {
				d.EMA5PSARStrategy = WeakSell
				d.EMA5PSARStrategy2 = WeakSell
			}
		}
	}
}

// applyRollingStdStrategies applies volatility-based strategies with band distance
func (s *Strategies) applyRollingStdStrategies(data []*StrategyData) {
	for _, d := range data {
		price := d.Close
		sma10 := d.SMA10
		sma50 := d.SMA50
		std10 := d.RollingStd10
		std50 := d.RollingStd50

		if price.IsZero() || sma10.IsZero() || std10.IsZero() {
			continue
		}

		// Rolling Std10 Strategy: Enhanced Bollinger Band-like signals
		upperBand10 := sma10.Add(std10.Mul(decimal.NewFromInt(2)))
		lowerBand10 := sma10.Sub(std10.Mul(decimal.NewFromInt(2)))
		extremeUpper10 := sma10.Add(std10.Mul(decimal.NewFromFloat(2.5)))
		extremeLower10 := sma10.Sub(std10.Mul(decimal.NewFromFloat(2.5)))

		if price.LessThan(extremeLower10) {
			d.RollingStd10Strategy = StrongBuy
		} else if price.LessThan(lowerBand10) {
			d.RollingStd10Strategy = Buy
		} else if price.LessThan(sma10.Sub(std10.Mul(decimal.NewFromFloat(0.5)))) {
			d.RollingStd10Strategy = WeakBuy
		} else if price.GreaterThan(extremeUpper10) {
			d.RollingStd10Strategy = StrongSell
		} else if price.GreaterThan(upperBand10) {
			d.RollingStd10Strategy = Sell
		} else if price.GreaterThan(sma10.Add(std10.Mul(decimal.NewFromFloat(0.5)))) {
			d.RollingStd10Strategy = WeakSell
		}

		// Rolling Std50 Strategy: Similar logic with SMA50
		if !sma50.IsZero() && !std50.IsZero() {
			upperBand50 := sma50.Add(std50.Mul(decimal.NewFromInt(2)))
			lowerBand50 := sma50.Sub(std50.Mul(decimal.NewFromInt(2)))
			extremeUpper50 := sma50.Add(std50.Mul(decimal.NewFromFloat(2.5)))
			extremeLower50 := sma50.Sub(std50.Mul(decimal.NewFromFloat(2.5)))

			if price.LessThan(extremeLower50) {
				d.RollingStd50Strategy = StrongBuy
			} else if price.LessThan(lowerBand50) {
				d.RollingStd50Strategy = Buy
			} else if price.LessThan(sma50.Sub(std50.Mul(decimal.NewFromFloat(0.5)))) {
				d.RollingStd50Strategy = WeakBuy
			} else if price.GreaterThan(extremeUpper50) {
				d.RollingStd50Strategy = StrongSell
			} else if price.GreaterThan(upperBand50) {
				d.RollingStd50Strategy = Sell
			} else if price.GreaterThan(sma50.Add(std50.Mul(decimal.NewFromFloat(0.5)))) {
				d.RollingStd50Strategy = WeakSell
			}
		}
	}
}

// saveStrategiesData saves strategy data to CSV file
func (s *Strategies) saveStrategiesData(data []*StrategyData, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return gocsv.MarshalFile(data, file)
}

// processAlternativeStates processes alternative states for strategies
func (s *Strategies) processAlternativeStates(filePath string, strategies []string) error {
	s.logger.Info("Processing alternative states for file: %s", filePath)

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	var data []*StrategyData
	if err := gocsv.UnmarshalFile(f, &data); err != nil {
		return err
	}

	for _, row := range data {
		signals := []*string{
			&row.RSIStrategy, &row.RSIStrategy2, &row.RSI14OBVRoCStrategy,
			&row.RSIMACDStrategy, &row.RSICMFStrategy, &row.RSIOBVStrategy,
			&row.OBVStrategy, &row.MACDStrategy, &row.CMFStrategy,
			&row.EMA5PSARStrategy, &row.EMA5PSARStrategy2,
			&row.RollingStd10Strategy, &row.RollingStd50Strategy,
		}

		buyCount := 0
		sellCount := 0
		for _, sig := range signals {
			if strings.Contains(*sig, "Buy") {
				buyCount++
			} else if strings.Contains(*sig, "Sell") {
				sellCount++
			}
		}

		if buyCount >= 3 {
			for _, sig := range signals {
				if *sig == Hold {
					*sig = WeakBuy
				}
			}
		}

		if sellCount >= 3 {
			for _, sig := range signals {
				if *sig == Hold {
					*sig = WeakSell
				}
			}
		}
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()
	return gocsv.MarshalFile(&data, out)
}

// generateStrategySummary generates a summary for a ticker's strategies
func (s *Strategies) generateStrategySummary(ticker, filePath string) (map[string]interface{}, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data []*StrategyData
	if err := gocsv.UnmarshalFile(f, &data); err != nil {
		return nil, err
	}

	stats := make(map[string]map[string]int)
	for _, d := range data {
		update := func(name, val string) {
			if _, ok := stats[name]; !ok {
				stats[name] = make(map[string]int)
			}
			stats[name][val]++
		}

		update("RSI Strategy", d.RSIStrategy)
		update("RSI Strategy2", d.RSIStrategy2)
		update("RSI14_OBV_RoC Strategy", d.RSI14OBVRoCStrategy)
		update("RSIMACD Strategy", d.RSIMACDStrategy)
		update("RSICMF Strategy", d.RSICMFStrategy)
		update("RSI OBV Strategy", d.RSIOBVStrategy)
		update("OBV Strategy", d.OBVStrategy)
		update("MACD Strategy", d.MACDStrategy)
		update("CMF Strategy", d.CMFStrategy)
		update("EMA5 PSAR Strategy", d.EMA5PSARStrategy)
		update("EMA5 PSAR Strategy2", d.EMA5PSARStrategy2)
		update("Rolling Std10 Strategy", d.RollingStd10Strategy)
		update("Rolling Std50 Strategy", d.RollingStd50Strategy)
	}

	summary := map[string]interface{}{
		"ticker":      ticker,
		"file_path":   filePath,
		"stats":       stats,
		"last_update": time.Now().Format("2006-01-02 15:04:05"),
	}
	return summary, nil
}

// saveSummaryToFile saves summary data to JSON file
func (s *Strategies) saveSummaryToFile(summaries map[string]interface{}, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(summaries)
}

// StrategyTester handles strategy testing and backtesting
type StrategyTester struct {
	logger *common.Logger
}

// NewStrategyTester creates a new StrategyTester instance
func NewStrategyTester() *StrategyTester {
	return &StrategyTester{
		logger: common.NewLogger(),
	}
}

// BacktestEngine represents the main backtesting engine
type BacktestEngine struct {
	config           common.BacktestConfig
	portfolio        common.Portfolio
	trades           []common.Trade
	positions        map[string]common.Position
	portfolioHistory []common.Portfolio
	tradeCounter     int
	logger           *common.Logger
}

// NewBacktestEngine creates a new backtesting engine
func NewBacktestEngine(config common.BacktestConfig) *BacktestEngine {
	return &BacktestEngine{
		config:           config,
		positions:        make(map[string]common.Position),
		trades:           make([]common.Trade, 0),
		portfolioHistory: make([]common.Portfolio, 0),
		tradeCounter:     0,
		logger:           common.NewLogger(),
		portfolio: common.Portfolio{
			Cash:       config.InitialCash,
			TotalValue: config.InitialCash,
		},
	}
}

// BacktestAllStrategies backtests all strategies
func (st *StrategyTester) BacktestAllStrategies() error {
	st.logger.Info("Starting comprehensive backtesting of all strategies")

	// Load configuration
	config, err := st.loadBacktestConfig()
	if err != nil {
		return fmt.Errorf("failed to load backtest config: %w", err)
	}

	// Load tickers if not specified in config
	if len(config.Tickers) == 0 {
		tickers, err := common.LoadTickers("TICKERS.csv")
		if err != nil {
			return fmt.Errorf("failed to load tickers: %w", err)
		}
		config.Tickers = tickers
	}

	// Backtest each strategy
	allResults := make(map[string]*common.BacktestResult)

	for _, strategy := range config.Strategies {
		st.logger.Info("Backtesting strategy: %s", strategy)

		result, err := st.backtestSingleStrategy(strategy, config.Tickers, config)
		if err != nil {
			st.logger.Error("Error backtesting %s: %v", strategy, err)
			continue
		}

		allResults[strategy] = result
		st.logger.Info("Strategy %s: Total Return %.2f%%, Win Rate %.2f%%",
			strategy, result.TotalReturn.InexactFloat64(), result.WinRate.InexactFloat64())
	}

	// Save results
	return st.saveBacktestResults(allResults)
}

// loadBacktestConfig loads the backtesting configuration
func (st *StrategyTester) loadBacktestConfig() (common.BacktestConfig, error) {
	var config common.BacktestConfig

	file, err := os.Open("backtest_config.json")
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}

// backtestSingleStrategy backtests a single strategy across all tickers
func (st *StrategyTester) backtestSingleStrategy(strategy string, tickers []string, config common.BacktestConfig) (*common.BacktestResult, error) {
	engine := NewBacktestEngine(config)

	// Combine all ticker data into chronological order
	allData, err := st.loadAndMergeTickerData(tickers, strategy)
	if err != nil {
		return nil, err
	}

	// Filter by date range
	filteredData := st.filterByDateRange(allData, config.StartDate, config.EndDate)

	if len(filteredData) == 0 {
		return nil, fmt.Errorf("no data available for strategy %s in specified date range", strategy)
	}

	// Run the backtest
	err = engine.runBacktest(filteredData, strategy)
	if err != nil {
		return nil, err
	}

	// Calculate performance metrics
	result := engine.calculatePerformanceMetrics(strategy)

	// Save detailed results
	st.saveDetailedResults(engine, strategy)

	return result, nil
}

// loadAndMergeTickerData loads strategy data for all tickers and merges chronologically
func (st *StrategyTester) loadAndMergeTickerData(tickers []string, strategy string) ([]StrategyDataPoint, error) {
	var allData []StrategyDataPoint

	for _, ticker := range tickers {
		filePath := fmt.Sprintf("Strategies_%s.csv", ticker)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			st.logger.Info("Strategy file not found for %s, skipping", ticker)
			continue
		}

		tickerData, err := st.loadTickerStrategyData(filePath, ticker, strategy)
		if err != nil {
			st.logger.Error("Error loading strategy data for %s: %v", ticker, err)
			continue
		}

		allData = append(allData, tickerData...)
	}

	// Sort by date
	sort.Slice(allData, func(i, j int) bool {
		return allData[i].Date.Before(allData[j].Date)
	})

	return allData, nil
}

// StrategyDataPoint represents a single data point for backtesting
type StrategyDataPoint struct {
	Ticker string
	Date   time.Time
	Open   decimal.Decimal
	High   decimal.Decimal
	Low    decimal.Decimal
	Close  decimal.Decimal
	Volume int64
	Signal string
}

// loadTickerStrategyData loads strategy data for a specific ticker
func (st *StrategyTester) loadTickerStrategyData(filePath, ticker, strategy string) ([]StrategyDataPoint, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var strategyData []*StrategyData
	if err := gocsv.UnmarshalFile(file, &strategyData); err != nil {
		return nil, err
	}

	var dataPoints []StrategyDataPoint
	for _, data := range strategyData {
		signal := st.getStrategySignal(data, strategy)

		dataPoint := StrategyDataPoint{
			Ticker: ticker,
			Date:   data.Date,
			Open:   data.Open,
			High:   data.High,
			Low:    data.Low,
			Close:  data.Close,
			Volume: data.Volume,
			Signal: signal,
		}
		dataPoints = append(dataPoints, dataPoint)
	}

	return dataPoints, nil
}

// getStrategySignal extracts the signal for a specific strategy
func (st *StrategyTester) getStrategySignal(data *StrategyData, strategy string) string {
	switch strategy {
	case "RSI Strategy":
		return data.RSIStrategy
	case "RSI Strategy2":
		return data.RSIStrategy2
	case "MACD Strategy":
		return data.MACDStrategy
	case "CMF Strategy":
		return data.CMFStrategy
	case "OBV Strategy":
		return data.OBVStrategy
	case "RSI OBV Strategy":
		return data.RSIOBVStrategy
	case "RSIMACD Strategy":
		return data.RSIMACDStrategy
	case "RSICMF Strategy":
		return data.RSICMFStrategy
	case "EMA5 PSAR Strategy":
		return data.EMA5PSARStrategy
	case "EMA5 PSAR Strategy2":
		return data.EMA5PSARStrategy2
	case "Rolling Std10 Strategy":
		return data.RollingStd10Strategy
	case "Rolling Std50 Strategy":
		return data.RollingStd50Strategy
	case "RSI14_OBV_RoC Strategy":
		return data.RSI14OBVRoCStrategy
	default:
		return Hold
	}
}

// filterByDateRange filters data points by date range
func (st *StrategyTester) filterByDateRange(data []StrategyDataPoint, startDate, endDate time.Time) []StrategyDataPoint {
	var filtered []StrategyDataPoint
	for _, point := range data {
		if point.Date.After(startDate) && point.Date.Before(endDate) {
			filtered = append(filtered, point)
		}
	}
	return filtered
}

// runBacktest executes the backtest simulation
func (be *BacktestEngine) runBacktest(data []StrategyDataPoint, strategy string) error {
	be.logger.Info("Running backtest for strategy %s with %d data points", strategy, len(data))

	for i, dataPoint := range data {
		// Update portfolio date
		be.portfolio.Date = dataPoint.Date
		be.portfolio.DaysSinceStart = i

		// Update existing positions with current prices
		be.updatePositions(dataPoint)

		// Process trading signal
		err := be.processSignal(dataPoint, strategy)
		if err != nil {
			be.logger.Error("Error processing signal for %s on %s: %v",
				dataPoint.Ticker, dataPoint.Date.Format("2006-01-02"), err)
		}

		// Update portfolio metrics
		be.updatePortfolioMetrics()

		// Save portfolio snapshot
		be.portfolioHistory = append(be.portfolioHistory, be.portfolio)
	}

	be.logger.Info("Backtest completed. Total trades: %d, Final portfolio value: %.2f",
		len(be.trades), be.portfolio.TotalValue.InexactFloat64())

	return nil
}

// updatePositions updates existing positions with current market prices
func (be *BacktestEngine) updatePositions(dataPoint StrategyDataPoint) {
	for ticker, position := range be.positions {
		if ticker == dataPoint.Ticker {
			position.CurrentPrice = dataPoint.Close
			position.CurrentValue = dataPoint.Close.Mul(decimal.NewFromInt(int64(position.Quantity)))
			position.UnrealizedPnL = position.CurrentValue.Sub(
				position.EntryPrice.Mul(decimal.NewFromInt(int64(position.Quantity))))

			be.positions[ticker] = position
		}
	}
}

// processSignal processes a trading signal and executes trades
func (be *BacktestEngine) processSignal(dataPoint StrategyDataPoint, strategy string) error {
	action, confidence := be.translateSignalToAction(dataPoint.Signal)

	switch action {
	case "BUY":
		return be.enterLongPosition(dataPoint, strategy, confidence)
	case "SELL":
		return be.exitPosition(dataPoint.Ticker, dataPoint, "SIGNAL")
	case "HOLD":
		// Check for stop loss or take profit
		return be.checkExitConditions(dataPoint)
	}

	return nil
}

// translateSignalToAction converts strategy signals to trading actions
func (be *BacktestEngine) translateSignalToAction(signal string) (action string, confidence decimal.Decimal) {
	switch signal {
	case "Strong Buy":
		return "BUY", decimal.NewFromFloat(1.0)
	case "Buy":
		return "BUY", decimal.NewFromFloat(0.8)
	case "Weak Buy":
		return "BUY", decimal.NewFromFloat(0.6)
	case "Strong Sell":
		return "SELL", decimal.NewFromFloat(1.0)
	case "Sell":
		return "SELL", decimal.NewFromFloat(0.8)
	case "Weak Sell":
		return "SELL", decimal.NewFromFloat(0.6)
	default: // Hold
		return "HOLD", decimal.NewFromFloat(0.0)
	}
}

// enterLongPosition enters a new long position
func (be *BacktestEngine) enterLongPosition(dataPoint StrategyDataPoint, strategy string, confidence decimal.Decimal) error {
	// Check if we already have a position in this ticker
	if _, exists := be.positions[dataPoint.Ticker]; exists {
		return nil // Already have position
	}

	// Check position limits
	if len(be.positions) >= be.config.MaxPositions {
		return nil // Too many positions
	}

	// Calculate position size
	positionValue := be.portfolio.Cash.Mul(be.config.PositionSize).Div(decimal.NewFromInt(100))
	if be.config.UseSignalStrength {
		positionValue = positionValue.Mul(confidence)
	}

	quantity := int(positionValue.Div(dataPoint.Close).IntPart())
	if quantity <= 0 {
		return nil // Insufficient cash or price too high
	}

	totalCost := dataPoint.Close.Mul(decimal.NewFromInt(int64(quantity))).Add(be.config.Commission)

	if totalCost.GreaterThan(be.portfolio.Cash) {
		// Adjust quantity to fit available cash
		availableCash := be.portfolio.Cash.Sub(be.config.Commission)
		quantity = int(availableCash.Div(dataPoint.Close).IntPart())
		if quantity <= 0 {
			return nil // Still insufficient cash
		}
		totalCost = dataPoint.Close.Mul(decimal.NewFromInt(int64(quantity))).Add(be.config.Commission)
	}

	// Calculate stop loss and take profit levels
	stopLoss := dataPoint.Close.Mul(decimal.NewFromInt(1).Sub(be.config.StopLoss.Div(decimal.NewFromInt(100))))
	takeProfit := dataPoint.Close.Mul(decimal.NewFromInt(1).Add(be.config.TakeProfit.Div(decimal.NewFromInt(100))))

	// Create position
	position := common.Position{
		Ticker:       dataPoint.Ticker,
		Strategy:     strategy,
		EntryDate:    dataPoint.Date,
		EntryPrice:   dataPoint.Close,
		Quantity:     quantity,
		CurrentPrice: dataPoint.Close,
		StopLoss:     stopLoss,
		TakeProfit:   takeProfit,
		TradeType:    "LONG",
	}

	be.positions[dataPoint.Ticker] = position
	be.portfolio.Cash = be.portfolio.Cash.Sub(totalCost)

	be.logger.Info("Entered position: %s at %.2f, quantity: %d, total cost: %.2f",
		dataPoint.Ticker, dataPoint.Close.InexactFloat64(), quantity, totalCost.InexactFloat64())

	return nil
}

// exitPosition exits an existing position
func (be *BacktestEngine) exitPosition(ticker string, dataPoint StrategyDataPoint, reason string) error {
	position, exists := be.positions[ticker]
	if !exists {
		return nil // No position to exit
	}

	// Calculate trade result
	proceeds := dataPoint.Close.Mul(decimal.NewFromInt(int64(position.Quantity))).Sub(be.config.Commission)
	cost := position.EntryPrice.Mul(decimal.NewFromInt(int64(position.Quantity))).Add(be.config.Commission)
	pnl := proceeds.Sub(cost)
	pnlPercent := pnl.Div(cost).Mul(decimal.NewFromInt(100))

	// Create trade record
	be.tradeCounter++
	trade := common.Trade{
		TradeID:     be.tradeCounter,
		Ticker:      ticker,
		Strategy:    position.Strategy,
		EntryDate:   position.EntryDate,
		ExitDate:    dataPoint.Date,
		EntryPrice:  position.EntryPrice,
		ExitPrice:   dataPoint.Close,
		EntrySignal: "", // Could be enhanced to track entry signal
		ExitSignal:  dataPoint.Signal,
		Quantity:    position.Quantity,
		PnL:         pnl,
		PnLPercent:  pnlPercent,
		HoldingDays: int(dataPoint.Date.Sub(position.EntryDate).Hours() / 24),
		TradeType:   position.TradeType,
		Commission:  be.config.Commission.Mul(decimal.NewFromInt(2)), // Entry + Exit
		ExitReason:  reason,
	}

	be.trades = append(be.trades, trade)
	be.portfolio.Cash = be.portfolio.Cash.Add(proceeds)
	delete(be.positions, ticker)

	be.logger.Info("Exited position: %s at %.2f, PnL: %.2f (%.2f%%), reason: %s",
		ticker, dataPoint.Close.InexactFloat64(), pnl.InexactFloat64(), pnlPercent.InexactFloat64(), reason)

	return nil
}

// checkExitConditions checks for stop loss, take profit, or time-based exits
func (be *BacktestEngine) checkExitConditions(dataPoint StrategyDataPoint) error {
	position, exists := be.positions[dataPoint.Ticker]
	if !exists {
		return nil
	}

	// Check stop loss
	if dataPoint.Low.LessThanOrEqual(position.StopLoss) {
		return be.exitPosition(dataPoint.Ticker, StrategyDataPoint{
			Ticker: dataPoint.Ticker,
			Date:   dataPoint.Date,
			Close:  position.StopLoss,
			Signal: "STOP_LOSS",
		}, "STOP_LOSS")
	}

	// Check take profit
	if dataPoint.High.GreaterThanOrEqual(position.TakeProfit) {
		return be.exitPosition(dataPoint.Ticker, StrategyDataPoint{
			Ticker: dataPoint.Ticker,
			Date:   dataPoint.Date,
			Close:  position.TakeProfit,
			Signal: "TAKE_PROFIT",
		}, "TAKE_PROFIT")
	}

	// Check time limit
	holdingDays := int(dataPoint.Date.Sub(position.EntryDate).Hours() / 24)
	if holdingDays >= be.config.MaxHoldingDays {
		return be.exitPosition(dataPoint.Ticker, dataPoint, "TIME_LIMIT")
	}

	return nil
}

// updatePortfolioMetrics updates portfolio performance metrics
func (be *BacktestEngine) updatePortfolioMetrics() {
	// Calculate total equity value
	equityValue := decimal.Zero
	for _, position := range be.positions {
		equityValue = equityValue.Add(position.CurrentValue)
	}

	be.portfolio.EquityValue = equityValue
	be.portfolio.TotalValue = be.portfolio.Cash.Add(equityValue)
	be.portfolio.ActivePositions = len(be.positions)

	// Calculate returns
	if len(be.portfolioHistory) > 0 {
		previousValue := be.portfolioHistory[len(be.portfolioHistory)-1].TotalValue
		if !previousValue.IsZero() {
			be.portfolio.DailyReturn = be.portfolio.TotalValue.Sub(previousValue).Div(previousValue).Mul(decimal.NewFromInt(100))
		}
	}

	be.portfolio.TotalReturn = be.portfolio.TotalValue.Sub(be.config.InitialCash).Div(be.config.InitialCash).Mul(decimal.NewFromInt(100))

	// Calculate drawdown
	if len(be.portfolioHistory) > 0 {
		maxValue := be.config.InitialCash
		for _, portfolio := range be.portfolioHistory {
			if portfolio.TotalValue.GreaterThan(maxValue) {
				maxValue = portfolio.TotalValue
			}
		}
		if be.portfolio.TotalValue.GreaterThan(maxValue) {
			maxValue = be.portfolio.TotalValue
		}

		be.portfolio.Drawdown = be.portfolio.TotalValue.Sub(maxValue).Div(maxValue).Mul(decimal.NewFromInt(100))
	}
}

// calculatePerformanceMetrics calculates comprehensive performance metrics
func (be *BacktestEngine) calculatePerformanceMetrics(strategy string) *common.BacktestResult {
	if len(be.trades) == 0 {
		return &common.BacktestResult{
			Strategy:      strategy,
			TotalReturn:   decimal.Zero,
			WinRate:       decimal.Zero,
			MaxDrawdown:   decimal.Zero,
			SharpeRatio:   decimal.Zero,
			ProfitFactor:  decimal.Zero,
			TotalTrades:   0,
			WinningTrades: 0,
			LosingTrades:  0,
			InitialCash:   be.config.InitialCash,
			FinalValue:    be.portfolio.TotalValue,
		}
	}

	// Basic metrics
	totalTrades := len(be.trades)
	winningTrades := 0
	totalPnL := decimal.Zero
	totalWinPnL := decimal.Zero
	totalLossPnL := decimal.Zero
	totalDays := decimal.Zero
	maxWin := decimal.Zero
	maxLoss := decimal.Zero

	for _, trade := range be.trades {
		totalPnL = totalPnL.Add(trade.PnL)
		totalDays = totalDays.Add(decimal.NewFromInt(int64(trade.HoldingDays)))

		if trade.PnL.GreaterThan(decimal.Zero) {
			winningTrades++
			totalWinPnL = totalWinPnL.Add(trade.PnL)
			if trade.PnL.GreaterThan(maxWin) {
				maxWin = trade.PnL
			}
		} else {
			totalLossPnL = totalLossPnL.Add(trade.PnL.Abs())
			if trade.PnL.LessThan(maxLoss) {
				maxLoss = trade.PnL
			}
		}
	}

	losingTrades := totalTrades - winningTrades
	winRate := decimal.NewFromInt(int64(winningTrades)).Div(decimal.NewFromInt(int64(totalTrades))).Mul(decimal.NewFromInt(100))
	avgTradeDays := totalDays.Div(decimal.NewFromInt(int64(totalTrades)))

	// Average win/loss
	avgWin := decimal.Zero
	avgLoss := decimal.Zero
	if winningTrades > 0 {
		avgWin = totalWinPnL.Div(decimal.NewFromInt(int64(winningTrades)))
	}
	if losingTrades > 0 {
		avgLoss = totalLossPnL.Div(decimal.NewFromInt(int64(losingTrades)))
	}

	// Profit factor
	profitFactor := decimal.Zero
	if !totalLossPnL.IsZero() {
		profitFactor = totalWinPnL.Div(totalLossPnL)
	}

	// Calculate maximum drawdown
	maxDrawdown := be.calculateMaxDrawdown()

	// Calculate Sharpe ratio
	sharpeRatio := be.calculateSharpeRatio()

	// Dates
	startDate := be.portfolioHistory[0].Date
	endDate := be.portfolioHistory[len(be.portfolioHistory)-1].Date

	return &common.BacktestResult{
		Strategy:      strategy,
		TotalReturn:   be.portfolio.TotalReturn,
		WinRate:       winRate,
		MaxDrawdown:   maxDrawdown,
		SharpeRatio:   sharpeRatio,
		ProfitFactor:  profitFactor,
		TotalTrades:   totalTrades,
		WinningTrades: winningTrades,
		LosingTrades:  losingTrades,
		AvgTradeDays:  avgTradeDays,
		AvgWin:        avgWin,
		AvgLoss:       avgLoss,
		MaxWin:        maxWin,
		MaxLoss:       maxLoss,
		StartDate:     startDate,
		EndDate:       endDate,
		InitialCash:   be.config.InitialCash,
		FinalValue:    be.portfolio.TotalValue,
	}
}

// calculateMaxDrawdown calculates the maximum drawdown
func (be *BacktestEngine) calculateMaxDrawdown() decimal.Decimal {
	if len(be.portfolioHistory) == 0 {
		return decimal.Zero
	}

	maxDrawdown := decimal.Zero
	peak := be.portfolioHistory[0].TotalValue

	for _, portfolio := range be.portfolioHistory {
		if portfolio.TotalValue.GreaterThan(peak) {
			peak = portfolio.TotalValue
		}

		drawdown := peak.Sub(portfolio.TotalValue).Div(peak).Mul(decimal.NewFromInt(100))
		if drawdown.GreaterThan(maxDrawdown) {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}

// calculateSharpeRatio calculates the Sharpe ratio
func (be *BacktestEngine) calculateSharpeRatio() decimal.Decimal {
	if len(be.portfolioHistory) < 2 {
		return decimal.Zero
	}

	// Calculate daily returns
	var returns []decimal.Decimal
	for i := 1; i < len(be.portfolioHistory); i++ {
		prevValue := be.portfolioHistory[i-1].TotalValue
		currValue := be.portfolioHistory[i].TotalValue
		if !prevValue.IsZero() {
			dailyReturn := currValue.Sub(prevValue).Div(prevValue)
			returns = append(returns, dailyReturn)
		}
	}

	if len(returns) == 0 {
		return decimal.Zero
	}

	// Calculate mean return
	sum := decimal.Zero
	for _, ret := range returns {
		sum = sum.Add(ret)
	}
	meanReturn := sum.Div(decimal.NewFromInt(int64(len(returns))))

	// Calculate standard deviation
	varianceSum := decimal.Zero
	for _, ret := range returns {
		diff := ret.Sub(meanReturn)
		varianceSum = varianceSum.Add(diff.Mul(diff))
	}

	if len(returns) <= 1 {
		return decimal.Zero
	}

	variance := varianceSum.Div(decimal.NewFromInt(int64(len(returns) - 1)))

	// Convert to float for sqrt calculation
	varianceFloat, _ := variance.Float64()
	if varianceFloat <= 0 {
		return decimal.Zero
	}

	stdDev := decimal.NewFromFloat(math.Sqrt(varianceFloat))

	if stdDev.IsZero() {
		return decimal.Zero
	}

	// Annualized Sharpe ratio (assuming 252 trading days)
	annualizedReturn := meanReturn.Mul(decimal.NewFromInt(252))
	annualizedVolatility := stdDev.Mul(decimal.NewFromFloat(math.Sqrt(252)))

	return annualizedReturn.Div(annualizedVolatility)
}

// saveDetailedResults saves detailed backtesting results
func (st *StrategyTester) saveDetailedResults(engine *BacktestEngine, strategy string) error {
	// Save trades
	tradesFileName := fmt.Sprintf("backtest_trades_%s.csv", strings.ReplaceAll(strategy, " ", "_"))
	if err := st.saveTradesCSV(engine.trades, tradesFileName); err != nil {
		st.logger.Error("Failed to save trades for %s: %v", strategy, err)
	}

	// Save portfolio history
	portfolioFileName := fmt.Sprintf("backtest_portfolio_%s.csv", strings.ReplaceAll(strategy, " ", "_"))
	if err := st.savePortfolioHistoryCSV(engine.portfolioHistory, portfolioFileName); err != nil {
		st.logger.Error("Failed to save portfolio history for %s: %v", strategy, err)
	}

	return nil
}

// saveTradesCSV saves trades to CSV file
func (st *StrategyTester) saveTradesCSV(trades []common.Trade, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return gocsv.MarshalFile(trades, file)
}

// savePortfolioHistoryCSV saves portfolio history to CSV file
func (st *StrategyTester) savePortfolioHistoryCSV(history []common.Portfolio, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return gocsv.MarshalFile(history, file)
}

// saveBacktestResults saves overall backtest results
func (st *StrategyTester) saveBacktestResults(results map[string]*common.BacktestResult) error {
	// Convert map to slice for CSV output
	var resultSlice []*common.BacktestResult
	for _, result := range results {
		resultSlice = append(resultSlice, result)
	}

	// Save to CSV
	file, err := os.Create("backtest_results.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	err = gocsv.MarshalFile(resultSlice, file)
	if err != nil {
		return err
	}

	// Also save as JSON for easier reading
	jsonFile, err := os.Create("backtest_results.json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "  ")

	return encoder.Encode(results)
}

// SimulateStrategyResults simulates strategy results
func (st *StrategyTester) SimulateStrategyResults() error {
	st.logger.Info("Running strategy backtesting simulation")
	return st.BacktestAllStrategies()
}

// SummarizeSimulatedStrategyResults summarizes simulated strategy results
func (st *StrategyTester) SummarizeSimulatedStrategyResults() error {
	st.logger.Info("Generating backtest summary report")

	// Load results
	results, err := st.loadBacktestResults()
	if err != nil {
		return fmt.Errorf("failed to load backtest results: %w", err)
	}

	// Generate summary
	summary := st.generateBacktestSummary(results)

	// Save summary
	return st.saveBacktestSummary(summary)
}

// loadBacktestResults loads backtest results from file
func (st *StrategyTester) loadBacktestResults() (map[string]*common.BacktestResult, error) {
	file, err := os.Open("backtest_results.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results map[string]*common.BacktestResult
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&results)

	return results, err
}

// generateBacktestSummary generates a comprehensive summary
func (st *StrategyTester) generateBacktestSummary(results map[string]*common.BacktestResult) map[string]interface{} {
	summary := map[string]interface{}{
		"total_strategies": len(results),
		"best_strategy":    "",
		"best_return":      decimal.Zero,
		"worst_strategy":   "",
		"worst_return":     decimal.Zero,
		"avg_return":       decimal.Zero,
		"strategies":       results,
		"generated_at":     time.Now().Format("2006-01-02 15:04:05"),
	}

	if len(results) == 0 {
		return summary
	}

	// Find best and worst strategies
	bestReturn := decimal.NewFromFloat(-999999)
	worstReturn := decimal.NewFromFloat(999999)
	totalReturn := decimal.Zero

	for strategy, result := range results {
		totalReturn = totalReturn.Add(result.TotalReturn)

		if result.TotalReturn.GreaterThan(bestReturn) {
			bestReturn = result.TotalReturn
			summary["best_strategy"] = strategy
			summary["best_return"] = bestReturn
		}

		if result.TotalReturn.LessThan(worstReturn) {
			worstReturn = result.TotalReturn
			summary["worst_strategy"] = strategy
			summary["worst_return"] = worstReturn
		}
	}

	// Calculate average return
	avgReturn := totalReturn.Div(decimal.NewFromInt(int64(len(results))))
	summary["avg_return"] = avgReturn

	return summary
}

// saveBacktestSummary saves the backtest summary
func (st *StrategyTester) saveBacktestSummary(summary map[string]interface{}) error {
	file, err := os.Create("backtest_summary.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(summary)
}
