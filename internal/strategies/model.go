package strategies

// Rule represents a single comparison row defined in Strategy Builder.
// Link is empty for the first rule in a chain otherwise "AND" or "OR".
// Indicator naming convention follows front-end e.g. "RSI_14", "SMA_50".
type Rule struct {
	Link      string  `json:"link"`      // "", "AND", "OR"
	Indicator string  `json:"indicator"` // indicatorId[_param]
	Operator  string  `json:"operator"`  // >, <, >=, <=, ==, !=
	Target    float64 `json:"target"`    // numeric value (only used when Target dropdown = Value)
}

// StrategyDefinition matches the JSON saved by the front-end.
// BuyRules and SellRules are evaluated separately on each bar.
type StrategyDefinition struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BuyRules  []Rule `json:"buy_rules"`
	SellRules []Rule `json:"sell_rules"`
}
