package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/shopspring/decimal"
)

// LiquidityScoreRecord represents a liquidity score record for a ticker
type LiquidityScoreRecord struct {
	Ticker                  string          `csv:"Ticker"`
	AverageVolume           decimal.Decimal `csv:"Average Volume"`
	AverageTradedVolume     decimal.Decimal `csv:"Average Traded Volume"`
	TimeWeightedVolume      decimal.Decimal `csv:"Time Weighted Volume"`
	VolumeSTD               decimal.Decimal `csv:"Volume STD"`
	DaysTraded              int             `csv:"Days Traded"`
	DaysTraded2             int             `csv:"Days Traded2"`
	ZeroVolumeDays          int             `csv:"Zero Volume Days"`
	TradingActivityScore    decimal.Decimal `csv:"Trading Activity Score"`
	VolumeConsistencyScore  decimal.Decimal `csv:"Volume Consistency Score"`
	ZeroVolumePenalty       decimal.Decimal `csv:"Zero Volume Penalty"`
	MarketImpactScore       decimal.Decimal `csv:"Market Impact Score"`
	IntradayVolatilityScore decimal.Decimal `csv:"Intraday Volatility Score"`
	RelativeVolumeScore     decimal.Decimal `csv:"Relative Volume Score"`
	LiquidityScore          decimal.Decimal `csv:"Enhanced Liquidity Score"`
	LiquidityScorePercent   decimal.Decimal `csv:"Liquidity Score%"`
}

// LiquidityCalc handles liquidity score calculations (separate from the stub in data_calculator.go)
type LiquidityCalc struct {
	logger *Logger
}

// NewLiquidityCalc creates a new LiquidityCalc instance
func NewLiquidityCalc() *LiquidityCalc {
	return &LiquidityCalc{
		logger: NewLogger(),
	}
}

// CalculateScores calculates liquidity scores for all tickers
func (lc *LiquidityCalc) CalculateScores() error {
	lc.logger.Info("Starting liquidity score calculation")

	// Read tickers from CSV
	tickers, err := lc.loadTickers()
	if err != nil {
		return fmt.Errorf("failed to load tickers: %w", err)
	}

	var liquidityScores []*LiquidityScoreRecord

	// Process each ticker
	for _, ticker := range tickers {
		score, err := lc.calculateTickerLiquidity(ticker)
		if err != nil {
			lc.logger.Error("Failed to calculate liquidity for ticker %s: %v", ticker, err)
			continue
		}
		if score != nil {
			liquidityScores = append(liquidityScores, score)
		}
	}

	if len(liquidityScores) == 0 {
		lc.logger.Error("No liquidity scores calculated")
		return fmt.Errorf("no liquidity scores calculated")
	}

	// Calculate relative scores
	lc.calculateRelativeScores(liquidityScores)

	// Save results
	if err := lc.saveLiquidityScores(liquidityScores); err != nil {
		return fmt.Errorf("failed to save liquidity scores: %w", err)
	}

	lc.logger.Info("Liquidity scores successfully calculated and saved")
	return nil
}

// loadTickers loads ticker symbols from TICKERS.csv
func (lc *LiquidityCalc) loadTickers() ([]string, error) {
	file, err := os.Open("TICKERS.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tickers []TickerInfo
	if err := gocsv.UnmarshalFile(file, &tickers); err != nil {
		return nil, err
	}

	var tickerSymbols []string
	for _, ticker := range tickers {
		tickerSymbols = append(tickerSymbols, ticker.Symbol)
	}

	return tickerSymbols, nil
}

// calculateTickerLiquidity calculates liquidity metrics for a single ticker
func (lc *LiquidityCalc) calculateTickerLiquidity(ticker string) (*LiquidityScoreRecord, error) {
	// Check if raw data file exists
	rawFilePath := fmt.Sprintf("raw_%s.csv", ticker)
	if _, err := os.Stat(rawFilePath); os.IsNotExist(err) {
		lc.logger.Info("Data file for %s does not exist", ticker)
		return nil, nil
	}

	// Load stock data
	stockData, err := lc.loadStockDataForLiquidity(rawFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load stock data: %w", err)
	}

	if len(stockData) == 0 {
		lc.logger.Info("No trading data for %s", ticker)
		return nil, nil
	}

	// Filter data to last 12 months
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	var last12MonthsData []*StockDataForLiquidity
	for _, data := range stockData {
		if data.Date.After(oneYearAgo) {
			last12MonthsData = append(last12MonthsData, data)
		}
	}

	if len(last12MonthsData) == 0 {
		lc.logger.Info("No trading data for %s in the past 12 months", ticker)
		return nil, nil
	}

	// Calculate basic metrics
	daysTraded := len(last12MonthsData)
	averageVolume := lc.calculateAverageVolume(last12MonthsData)
	volumeSTD := lc.calculateVolumeSTD(last12MonthsData)

	// Remove outliers (top and bottom 5%)
	filteredData := lc.removeVolumeOutliers(last12MonthsData)
	daysTraded2 := len(filteredData)
	averageTradedVolume := lc.calculateAverageVolume(filteredData)

	// Calculate enhanced metrics
	timeWeightedVolume := lc.calculateTimeWeightedVolume(last12MonthsData)
	zeroVolumeDays := lc.countZeroVolumeDays(last12MonthsData)
	volumeConsistencyScore := lc.calculateVolumeConsistencyScore(last12MonthsData)
	zeroVolumePenalty := lc.calculateZeroVolumePenalty(last12MonthsData)
	marketImpactScore := lc.calculateMarketImpactScore(last12MonthsData)
	intradayVolatilityScore := lc.calculateIntradayVolatilityScore(last12MonthsData)

	lc.logger.Info("Average volume traded for %s is %s", ticker, averageTradedVolume.String())

	// Calculate trading activity score
	tradingActivityScore := decimal.NewFromInt(int64(daysTraded)).Div(decimal.NewFromInt(252))
	if daysTraded < 100 {
		tradingActivityScore = decimal.Zero
	}

	return &LiquidityScoreRecord{
		Ticker:                  ticker,
		AverageVolume:           averageVolume,
		AverageTradedVolume:     averageTradedVolume,
		TimeWeightedVolume:      timeWeightedVolume,
		VolumeSTD:               volumeSTD,
		DaysTraded:              daysTraded,
		DaysTraded2:             daysTraded2,
		ZeroVolumeDays:          zeroVolumeDays,
		TradingActivityScore:    tradingActivityScore,
		VolumeConsistencyScore:  volumeConsistencyScore,
		ZeroVolumePenalty:       zeroVolumePenalty,
		MarketImpactScore:       marketImpactScore,
		IntradayVolatilityScore: intradayVolatilityScore,
		RelativeVolumeScore:     decimal.Zero, // Will be calculated later
		LiquidityScore:          decimal.Zero, // Will be calculated later
		LiquidityScorePercent:   decimal.Zero, // Will be calculated later
	}, nil
}

// StockDataForLiquidity represents stock data needed for liquidity calculations
type StockDataForLiquidity struct {
	Date          time.Time       `csv:"Date"`
	Close         decimal.Decimal `csv:"Close"`
	Open          decimal.Decimal `csv:"Open"`
	High          decimal.Decimal `csv:"High"`
	Low           decimal.Decimal `csv:"Low"`
	Change        decimal.Decimal `csv:"Change"`
	ChangePercent decimal.Decimal `csv:"Change%"`
	Volume        int64           `csv:"Volume"`
}

// loadStockDataForLiquidity loads stock data specifically for liquidity calculations
func (lc *LiquidityCalc) loadStockDataForLiquidity(filePath string) ([]*StockDataForLiquidity, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rawData []*StockDataCSV
	if err := gocsv.UnmarshalFile(file, &rawData); err != nil {
		return nil, err
	}

	// Convert to liquidity-specific data structure
	var stockData []*StockDataForLiquidity
	for _, data := range rawData {
		stockData = append(stockData, &StockDataForLiquidity{
			Date:          data.Date.Time,
			Close:         data.Close,
			Open:          data.Open,
			High:          data.High,
			Low:           data.Low,
			Change:        data.Change,
			ChangePercent: parsePercentage(data.ChangePercent),
			Volume:        data.Volume,
		})
	}

	// Sort by date
	sort.Slice(stockData, func(i, j int) bool {
		return stockData[i].Date.Before(stockData[j].Date)
	})

	return stockData, nil
}

// calculateAverageVolume calculates the average volume for a dataset
func (lc *LiquidityCalc) calculateAverageVolume(data []*StockDataForLiquidity) decimal.Decimal {
	if len(data) == 0 {
		return decimal.Zero
	}

	total := decimal.Zero
	for _, d := range data {
		total = total.Add(decimal.NewFromInt(d.Volume))
	}

	return total.Div(decimal.NewFromInt(int64(len(data))))
}

// calculateVolumeSTD calculates the standard deviation of volume percentage changes
func (lc *LiquidityCalc) calculateVolumeSTD(data []*StockDataForLiquidity) decimal.Decimal {
	if len(data) <= 1 {
		return decimal.Zero
	}

	// Calculate percentage changes
	var pctChanges []decimal.Decimal
	for i := 1; i < len(data); i++ {
		if data[i-1].Volume > 0 {
			prevVolume := decimal.NewFromInt(data[i-1].Volume)
			currVolume := decimal.NewFromInt(data[i].Volume)
			pctChange := currVolume.Sub(prevVolume).Div(prevVolume)
			pctChanges = append(pctChanges, pctChange)
		}
	}

	if len(pctChanges) == 0 {
		return decimal.Zero
	}

	// Calculate mean
	sum := decimal.Zero
	for _, change := range pctChanges {
		sum = sum.Add(change)
	}
	mean := sum.Div(decimal.NewFromInt(int64(len(pctChanges))))

	// Calculate variance
	variance := decimal.Zero
	for _, change := range pctChanges {
		diff := change.Sub(mean)
		variance = variance.Add(diff.Mul(diff))
	}
	variance = variance.Div(decimal.NewFromInt(int64(len(pctChanges))))

	// Return standard deviation (square root of variance)
	return lc.sqrt(variance)
}

// removeVolumeOutliers removes the top and bottom 5% of volume data
func (lc *LiquidityCalc) removeVolumeOutliers(data []*StockDataForLiquidity) []*StockDataForLiquidity {
	if len(data) <= 10 { // Need at least 10 data points for meaningful percentiles
		return data
	}

	// Create a copy and sort by volume
	volumes := make([]int64, len(data))
	for i, d := range data {
		volumes[i] = d.Volume
	}
	sort.Slice(volumes, func(i, j int) bool {
		return volumes[i] < volumes[j]
	})

	// Calculate 5th and 95th percentiles
	p5Index := int(float64(len(volumes)) * 0.05)
	p95Index := int(float64(len(volumes)) * 0.95)

	if p95Index >= len(volumes) {
		p95Index = len(volumes) - 1
	}

	p5Value := volumes[p5Index]
	p95Value := volumes[p95Index]

	// Filter original data
	var filtered []*StockDataForLiquidity
	for _, d := range data {
		if d.Volume >= p5Value && d.Volume <= p95Value {
			filtered = append(filtered, d)
		}
	}

	return filtered
}

// calculateRelativeScores calculates relative volume scores and enhanced final liquidity scores
func (lc *LiquidityCalc) calculateRelativeScores(scores []*LiquidityScoreRecord) {
	if len(scores) == 0 {
		return
	}

	// Find maximum average traded volume for normalization
	maxVolume := decimal.Zero
	for _, score := range scores {
		if score.AverageTradedVolume.GreaterThan(maxVolume) {
			maxVolume = score.AverageTradedVolume
		}
	}

	// Calculate relative volume scores and enhanced liquidity scores
	totalLiquidityScore := decimal.Zero
	for _, score := range scores {
		// Relative Volume Score (normalized to 0-1)
		if !maxVolume.IsZero() {
			score.RelativeVolumeScore = score.AverageTradedVolume.Div(maxVolume)
		}

		// Enhanced Liquidity Score using weighted combination
		score.LiquidityScore = lc.calculateEnhancedLiquidityScore(score)
		totalLiquidityScore = totalLiquidityScore.Add(score.LiquidityScore)
	}

	// Calculate liquidity score percentages
	for _, score := range scores {
		if !totalLiquidityScore.IsZero() {
			score.LiquidityScorePercent = score.LiquidityScore.Div(totalLiquidityScore)
		}
	}
}

// calculateEnhancedLiquidityScore calculates the final weighted liquidity score
func (lc *LiquidityCalc) calculateEnhancedLiquidityScore(score *LiquidityScoreRecord) decimal.Decimal {
	// Weighted combination of all factors
	weights := map[string]decimal.Decimal{
		"tradingActivity":       decimal.NewFromFloat(0.25), // 25% - trading frequency
		"volumeConsistency":     decimal.NewFromFloat(0.20), // 20% - predictable volume
		"relativeVolume":        decimal.NewFromFloat(0.20), // 20% - volume size
		"marketImpact":          decimal.NewFromFloat(0.15), // 15% - price impact
		"intradayVolatility":    decimal.NewFromFloat(0.10), // 10% - intraday stability
		"timeWeightedRelevance": decimal.NewFromFloat(0.10), // 10% - recent activity
	}

	// Calculate time-weighted relevance score (based on recent vs average volume)
	timeWeightedRelevance := decimal.Zero
	if !score.AverageTradedVolume.IsZero() {
		timeWeightedRelevance = score.TimeWeightedVolume.Div(score.AverageTradedVolume)
		if timeWeightedRelevance.GreaterThan(decimal.NewFromInt(1)) {
			timeWeightedRelevance = decimal.NewFromInt(1)
		}
	}

	// Calculate weighted score
	enhancedScore := weights["tradingActivity"].Mul(score.TradingActivityScore).
		Add(weights["volumeConsistency"].Mul(score.VolumeConsistencyScore)).
		Add(weights["relativeVolume"].Mul(score.RelativeVolumeScore)).
		Add(weights["marketImpact"].Mul(score.MarketImpactScore)).
		Add(weights["intradayVolatility"].Mul(score.IntradayVolatilityScore)).
		Add(weights["timeWeightedRelevance"].Mul(timeWeightedRelevance))

	// Apply zero-volume penalty (multiply by penalty factor)
	enhancedScore = enhancedScore.Mul(score.ZeroVolumePenalty)

	// Scale to 0-100
	return enhancedScore.Mul(decimal.NewFromInt(100))
}

// saveLiquidityScores saves liquidity scores to CSV file
func (lc *LiquidityCalc) saveLiquidityScores(scores []*LiquidityScoreRecord) error {
	file, err := os.Create("liquidity_scores.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	return gocsv.MarshalFile(scores, file)
}

// countZeroVolumeDays counts days with zero trading volume
func (lc *LiquidityCalc) countZeroVolumeDays(data []*StockDataForLiquidity) int {
	count := 0
	for _, d := range data {
		if d.Volume == 0 {
			count++
		}
	}
	return count
}

// calculateTimeWeightedVolume calculates volume with exponential time weighting
func (lc *LiquidityCalc) calculateTimeWeightedVolume(data []*StockDataForLiquidity) decimal.Decimal {
	if len(data) == 0 {
		return decimal.Zero
	}

	totalWeightedVolume := decimal.Zero
	totalWeight := decimal.Zero

	for i, d := range data {
		// More recent data gets higher weight (exponential decay with 90-day half-life)
		daysFromMostRecent := len(data) - i - 1
		weight := decimal.NewFromFloat(math.Exp(-float64(daysFromMostRecent) / 90.0))

		weightedVolume := decimal.NewFromInt(d.Volume).Mul(weight)
		totalWeightedVolume = totalWeightedVolume.Add(weightedVolume)
		totalWeight = totalWeight.Add(weight)
	}

	if totalWeight.IsZero() {
		return decimal.Zero
	}

	return totalWeightedVolume.Div(totalWeight)
}

// calculateVolumeConsistencyScore calculates how consistent the volume is
func (lc *LiquidityCalc) calculateVolumeConsistencyScore(data []*StockDataForLiquidity) decimal.Decimal {
	if len(data) <= 1 {
		return decimal.Zero
	}

	avgVolume := lc.calculateAverageVolume(data)
	if avgVolume.IsZero() {
		return decimal.Zero
	}

	stdDev := lc.calculateVolumeSTD(data)

	// Coefficient of Variation
	cv := stdDev.Div(avgVolume)

	// Convert to consistency score: 1/(1+CV) gives 0-1 scale where 1 = perfect consistency
	return decimal.NewFromInt(1).Div(decimal.NewFromInt(1).Add(cv))
}

// calculateZeroVolumePenalty calculates penalty for days with zero volume
func (lc *LiquidityCalc) calculateZeroVolumePenalty(data []*StockDataForLiquidity) decimal.Decimal {
	if len(data) == 0 {
		return decimal.Zero
	}

	zeroVolumeDays := lc.countZeroVolumeDays(data)
	penaltyRate := decimal.NewFromInt(int64(zeroVolumeDays)).Div(decimal.NewFromInt(int64(len(data))))

	// Return multiplier (1.0 = no penalty, 0.0 = maximum penalty)
	return decimal.NewFromInt(1).Sub(penaltyRate)
}

// calculateMarketImpactScore calculates how much volume affects price
func (lc *LiquidityCalc) calculateMarketImpactScore(data []*StockDataForLiquidity) decimal.Decimal {
	if len(data) <= 1 {
		return decimal.Zero
	}

	var impacts []decimal.Decimal
	avgVolume := lc.calculateAverageVolume(data)

	if avgVolume.IsZero() {
		return decimal.Zero
	}

	for _, d := range data {
		if d.Volume > 0 && !d.ChangePercent.IsZero() {
			// Price change percentage (absolute value)
			priceChangePercent := d.ChangePercent.Abs()

			// Volume relative to average
			volumeRatio := decimal.NewFromInt(d.Volume).Div(avgVolume)

			// Market impact = price change per unit of relative volume
			if !volumeRatio.IsZero() {
				impact := priceChangePercent.Div(volumeRatio)
				impacts = append(impacts, impact)
			}
		}
	}

	if len(impacts) == 0 {
		return decimal.Zero
	}

	// Calculate average impact
	sum := decimal.Zero
	for _, impact := range impacts {
		sum = sum.Add(impact)
	}
	avgImpact := sum.Div(decimal.NewFromInt(int64(len(impacts))))

	// Convert to 0-1 score where lower impact = higher liquidity
	// Use exponential decay to convert to score
	impactFloat, _ := avgImpact.Float64()
	score := math.Exp(-impactFloat / 2.0)

	return decimal.NewFromFloat(score)
}

// calculateIntradayVolatilityScore calculates intraday price volatility
func (lc *LiquidityCalc) calculateIntradayVolatilityScore(data []*StockDataForLiquidity) decimal.Decimal {
	if len(data) == 0 {
		return decimal.Zero
	}

	var ranges []decimal.Decimal

	for _, d := range data {
		if !d.High.IsZero() && !d.Low.IsZero() && !d.Close.IsZero() {
			// True Range calculation as percentage of close price
			hl := d.High.Sub(d.Low)
			rangePct := hl.Div(d.Close)
			ranges = append(ranges, rangePct)
		}
	}

	if len(ranges) == 0 {
		return decimal.Zero
	}

	// Calculate average range
	sum := decimal.Zero
	for _, r := range ranges {
		sum = sum.Add(r)
	}
	avgRange := sum.Div(decimal.NewFromInt(int64(len(ranges))))

	// Convert to 0-1 score where lower volatility = higher liquidity
	rangeFloat, _ := avgRange.Float64()
	score := math.Exp(-rangeFloat * 10.0)

	return decimal.NewFromFloat(score)
}

// sqrt calculates square root using Newton's method for decimal
func (lc *LiquidityCalc) sqrt(x decimal.Decimal) decimal.Decimal {
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
