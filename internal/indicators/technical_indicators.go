package indicators

import (
	"math"

	"github.com/shopspring/decimal"
)

// TechnicalIndicators provides technical analysis calculations
type TechnicalIndicators struct{}

// NewTechnicalIndicators creates a new TechnicalIndicators instance
func NewTechnicalIndicators() *TechnicalIndicators {
	return &TechnicalIndicators{}
}

// CalculateSMA calculates Simple Moving Average
func (ti *TechnicalIndicators) CalculateSMA(prices []decimal.Decimal, period int) []decimal.Decimal {
	if len(prices) < period {
		return make([]decimal.Decimal, len(prices))
	}

	sma := make([]decimal.Decimal, len(prices))

	for i := period - 1; i < len(prices); i++ {
		sum := decimal.Zero
		for j := i - period + 1; j <= i; j++ {
			sum = sum.Add(prices[j])
		}
		sma[i] = sum.Div(decimal.NewFromInt(int64(period)))
	}

	return sma
}

// CalculateEMA calculates Exponential Moving Average
func (ti *TechnicalIndicators) CalculateEMA(prices []decimal.Decimal, period int) []decimal.Decimal {
	if len(prices) == 0 {
		return []decimal.Decimal{}
	}

	ema := make([]decimal.Decimal, len(prices))
	multiplier := decimal.NewFromFloat(2.0).Div(decimal.NewFromInt(int64(period + 1)))

	// First EMA value is the first price
	ema[0] = prices[0]

	for i := 1; i < len(prices); i++ {
		ema[i] = prices[i].Mul(multiplier).Add(ema[i-1].Mul(decimal.NewFromInt(1).Sub(multiplier)))
	}

	return ema
}

// CalculateRSI calculates Relative Strength Index
func (ti *TechnicalIndicators) CalculateRSI(prices []decimal.Decimal, period int) []decimal.Decimal {
	if len(prices) < period+1 {
		return make([]decimal.Decimal, len(prices))
	}

	rsi := make([]decimal.Decimal, len(prices))
	gains := make([]decimal.Decimal, len(prices))
	losses := make([]decimal.Decimal, len(prices))

	// Calculate price changes
	for i := 1; i < len(prices); i++ {
		change := prices[i].Sub(prices[i-1])
		if change.GreaterThan(decimal.Zero) {
			gains[i] = change
		} else {
			losses[i] = change.Abs()
		}
	}

	// Calculate initial average gain and loss
	avgGain := decimal.Zero
	avgLoss := decimal.Zero

	for i := 1; i <= period; i++ {
		avgGain = avgGain.Add(gains[i])
		avgLoss = avgLoss.Add(losses[i])
	}

	avgGain = avgGain.Div(decimal.NewFromInt(int64(period)))
	avgLoss = avgLoss.Div(decimal.NewFromInt(int64(period)))

	// Calculate RSI
	for i := period; i < len(prices); i++ {
		if i > period {
			avgGain = avgGain.Mul(decimal.NewFromInt(int64(period - 1))).Add(gains[i]).Div(decimal.NewFromInt(int64(period)))
			avgLoss = avgLoss.Mul(decimal.NewFromInt(int64(period - 1))).Add(losses[i]).Div(decimal.NewFromInt(int64(period)))
		}

		if avgLoss.Equal(decimal.Zero) {
			rsi[i] = decimal.NewFromInt(100)
		} else {
			rs := avgGain.Div(avgLoss)
			rsi[i] = decimal.NewFromInt(100).Sub(decimal.NewFromInt(100).Div(decimal.NewFromInt(1).Add(rs)))
		}
	}

	return rsi
}

// CalculateMACD calculates Moving Average Convergence Divergence
func (ti *TechnicalIndicators) CalculateMACD(prices []decimal.Decimal, fastPeriod, slowPeriod, signalPeriod int) ([]decimal.Decimal, []decimal.Decimal, []decimal.Decimal) {
	fastEMA := ti.CalculateEMA(prices, fastPeriod)
	slowEMA := ti.CalculateEMA(prices, slowPeriod)

	macd := make([]decimal.Decimal, len(prices))
	for i := 0; i < len(prices); i++ {
		macd[i] = fastEMA[i].Sub(slowEMA[i])
	}

	signal := ti.CalculateEMA(macd, signalPeriod)

	histogram := make([]decimal.Decimal, len(prices))
	for i := 0; i < len(prices); i++ {
		histogram[i] = macd[i].Sub(signal[i])
	}

	return macd, signal, histogram
}

// CalculateStochastic calculates Stochastic Oscillator
func (ti *TechnicalIndicators) CalculateStochastic(highs, lows, closes []decimal.Decimal, kPeriod, dPeriod int) ([]decimal.Decimal, []decimal.Decimal) {
	if len(highs) != len(lows) || len(lows) != len(closes) {
		return nil, nil
	}

	k := make([]decimal.Decimal, len(closes))

	for i := kPeriod - 1; i < len(closes); i++ {
		// Find highest high and lowest low in the period
		highestHigh := highs[i-kPeriod+1]
		lowestLow := lows[i-kPeriod+1]

		for j := i - kPeriod + 2; j <= i; j++ {
			if highs[j].GreaterThan(highestHigh) {
				highestHigh = highs[j]
			}
			if lows[j].LessThan(lowestLow) {
				lowestLow = lows[j]
			}
		}

		// Calculate %K
		if highestHigh.Equal(lowestLow) {
			k[i] = decimal.NewFromInt(50) // Avoid division by zero
		} else {
			k[i] = closes[i].Sub(lowestLow).Div(highestHigh.Sub(lowestLow)).Mul(decimal.NewFromInt(100))
		}
	}

	// Calculate %D (SMA of %K)
	d := ti.CalculateSMA(k, dPeriod)

	return k, d
}

// CalculateATR calculates Average True Range
func (ti *TechnicalIndicators) CalculateATR(highs, lows, closes []decimal.Decimal, period int) []decimal.Decimal {
	if len(highs) != len(lows) || len(lows) != len(closes) || len(closes) < 2 {
		return make([]decimal.Decimal, len(closes))
	}

	trueRanges := make([]decimal.Decimal, len(closes))

	for i := 1; i < len(closes); i++ {
		tr1 := highs[i].Sub(lows[i])
		tr2 := highs[i].Sub(closes[i-1]).Abs()
		tr3 := lows[i].Sub(closes[i-1]).Abs()

		trueRanges[i] = decimal.Max(tr1, decimal.Max(tr2, tr3))
	}

	return ti.CalculateEMA(trueRanges, period)
}

// CalculateOBV calculates On-Balance Volume
func (ti *TechnicalIndicators) CalculateOBV(closes []decimal.Decimal, volumes []int64) []int64 {
	if len(closes) != len(volumes) || len(closes) < 2 {
		return make([]int64, len(closes))
	}

	obv := make([]int64, len(closes))
	obv[0] = volumes[0]

	for i := 1; i < len(closes); i++ {
		if closes[i].GreaterThan(closes[i-1]) {
			obv[i] = obv[i-1] + volumes[i]
		} else if closes[i].LessThan(closes[i-1]) {
			obv[i] = obv[i-1] - volumes[i]
		} else {
			obv[i] = obv[i-1]
		}
	}

	return obv
}

// CalculateCMF calculates Chaikin Money Flow
func (ti *TechnicalIndicators) CalculateCMF(highs, lows, closes []decimal.Decimal, volumes []int64, period int) []decimal.Decimal {
	if len(highs) != len(lows) || len(lows) != len(closes) || len(closes) != len(volumes) {
		return make([]decimal.Decimal, len(closes))
	}

	cmf := make([]decimal.Decimal, len(closes))

	for i := period - 1; i < len(closes); i++ {
		var sumMoneyFlowVolume decimal.Decimal
		var sumVolume int64

		for j := i - period + 1; j <= i; j++ {
			if highs[j].Equal(lows[j]) {
				continue // Avoid division by zero
			}

			moneyFlowMultiplier := closes[j].Sub(lows[j]).Sub(highs[j].Sub(closes[j])).Div(highs[j].Sub(lows[j]))
			moneyFlowVolume := moneyFlowMultiplier.Mul(decimal.NewFromInt(volumes[j]))

			sumMoneyFlowVolume = sumMoneyFlowVolume.Add(moneyFlowVolume)
			sumVolume += volumes[j]
		}

		if sumVolume > 0 {
			cmf[i] = sumMoneyFlowVolume.Div(decimal.NewFromInt(sumVolume))
		}
	}

	return cmf
}

// CalculatePSAR calculates Parabolic SAR
func (ti *TechnicalIndicators) CalculatePSAR(highs, lows []decimal.Decimal, acceleration, maximum decimal.Decimal) []decimal.Decimal {
	if len(highs) != len(lows) || len(highs) < 2 {
		return make([]decimal.Decimal, len(highs))
	}

	psar := make([]decimal.Decimal, len(highs))
	af := acceleration
	ep := highs[0] // Extreme point
	isUptrend := true

	psar[0] = lows[0]

	for i := 1; i < len(highs); i++ {
		psar[i] = psar[i-1].Add(af.Mul(ep.Sub(psar[i-1])))

		if isUptrend {
			if lows[i].LessThan(psar[i]) {
				isUptrend = false
				psar[i] = ep
				ep = lows[i]
				af = acceleration
			} else {
				if highs[i].GreaterThan(ep) {
					ep = highs[i]
					af = decimal.Min(af.Add(acceleration), maximum)
				}
			}
		} else {
			if highs[i].GreaterThan(psar[i]) {
				isUptrend = true
				psar[i] = ep
				ep = highs[i]
				af = acceleration
			} else {
				if lows[i].LessThan(ep) {
					ep = lows[i]
					af = decimal.Min(af.Add(acceleration), maximum)
				}
			}
		}
	}

	return psar
}

// CalculateRollingStd calculates rolling standard deviation
func (ti *TechnicalIndicators) CalculateRollingStd(prices []decimal.Decimal, period int) []decimal.Decimal {
	if len(prices) < period {
		return make([]decimal.Decimal, len(prices))
	}

	std := make([]decimal.Decimal, len(prices))

	for i := period - 1; i < len(prices); i++ {
		// Calculate mean
		sum := decimal.Zero
		for j := i - period + 1; j <= i; j++ {
			sum = sum.Add(prices[j])
		}
		mean := sum.Div(decimal.NewFromInt(int64(period)))

		// Calculate variance
		variance := decimal.Zero
		for j := i - period + 1; j <= i; j++ {
			diff := prices[j].Sub(mean)
			variance = variance.Add(diff.Mul(diff))
		}
		variance = variance.Div(decimal.NewFromInt(int64(period)))

		// Calculate standard deviation
		stdFloat, _ := variance.Float64()
		std[i] = decimal.NewFromFloat(math.Sqrt(stdFloat))
	}

	return std
}
