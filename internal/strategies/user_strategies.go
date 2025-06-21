package strategies

import (
	"encoding/json"
	"os"
)

// userStrategies holds custom strategies loaded from JSON.
var userStrategies []StrategyDefinition

// loadUserStrategies reads strategies_user.json if present.
func loadUserStrategies(path string) error {
	f, err := os.Open(path)
	if err != nil {
		// file optional
		return nil
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	return dec.Decode(&userStrategies)
}

// applyCustomRSI iterates over data slice and fills CustomSignal based on first custom strategy using RSI.
func applyCustomRSI(data []*StrategyData, indValues []float64, strat StrategyDefinition) {
	if len(strat.BuyRules) == 0 || len(strat.SellRules) == 0 {
		return
	}
	for i, v := range indValues {
		buy := evalChain(v, strat.BuyRules)
		sell := evalChain(v, strat.SellRules)
		switch {
		case buy:
			data[i].RSIStrategy = "Buy"
		case sell:
			data[i].RSIStrategy = "Sell"
		default:
			data[i].RSIStrategy = "Hold"
		}
	}
}
