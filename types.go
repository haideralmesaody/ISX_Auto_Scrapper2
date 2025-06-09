package main

import (
	"time"

	"github.com/shopspring/decimal"
)

// StockData represents a single day's stock data
type StockData struct {
	Date   time.Time       `csv:"Date"`
	Open   decimal.Decimal `csv:"Open"`
	High   decimal.Decimal `csv:"High"`
	Low    decimal.Decimal `csv:"Low"`
	Close  decimal.Decimal `csv:"Close"`
	Volume int64           `csv:"Volume"`

	// Price change calculations
	Change        decimal.Decimal `csv:"Change"`
	ChangePercent decimal.Decimal `csv:"Change_Percent"`

	// Technical Indicators
	SMA10  decimal.Decimal `csv:"SMA10"`
	SMA50  decimal.Decimal `csv:"SMA50"`
	SMA200 decimal.Decimal `csv:"SMA200"`

	EMA5   decimal.Decimal `csv:"EMA5"`
	EMA10  decimal.Decimal `csv:"EMA10"`
	EMA20  decimal.Decimal `csv:"EMA20"`
	EMA50  decimal.Decimal `csv:"EMA50"`
	EMA200 decimal.Decimal `csv:"EMA200"`

	RSI14  decimal.Decimal `csv:"RSI_14"`
	StochK decimal.Decimal `csv:"STOCHk_9_6_3"`
	StochD decimal.Decimal `csv:"STOCHd_9_6_3"`

	MACD       decimal.Decimal `csv:"MACD_12_26_9"`
	MACDSignal decimal.Decimal `csv:"MACDs_12_26_9"`
	MACDHist   decimal.Decimal `csv:"MACDh_12_26_9"`

	CMF   decimal.Decimal `csv:"CMF"`
	OBV   decimal.Decimal `csv:"OBV"`
	ATR   decimal.Decimal `csv:"ATR"`
	PSAR  decimal.Decimal `csv:"PSAR"`
	PSAR2 decimal.Decimal `csv:"PSARl_0.01_0.1"`

	RollingStd decimal.Decimal `csv:"Rolling_Std"`

	// Crossover signals
	GoldenCross          bool `csv:"Golden_Cross"`
	DeathCross           bool `csv:"Death_Cross"`
	PriceCrossSMA10Up    bool `csv:"Price_Cross_SMA10_Up"`
	PriceCrossSMA10Down  bool `csv:"Price_Cross_SMA10_Down"`
	PriceCrossSMA50Up    bool `csv:"Price_Cross_SMA50_Up"`
	PriceCrossSMA50Down  bool `csv:"Price_Cross_SMA50_Down"`
	PriceCrossSMA200Up   bool `csv:"Price_Cross_SMA200_Up"`
	PriceCrossSMA200Down bool `csv:"Price_Cross_SMA200_Down"`

	// Trend indicators
	SMA10Up  bool `csv:"SMA10_Up"`
	SMA50Up  bool `csv:"SMA50_Up"`
	SMA200Up bool `csv:"SMA200_Up"`

	// Distance calculations
	PriceDistanceSMA10  decimal.Decimal `csv:"Price_Distance_SMA10"`
	PriceDistanceSMA50  decimal.Decimal `csv:"Price_Distance_SMA50"`
	PriceDistanceSMA200 decimal.Decimal `csv:"Price_Distance_SMA200"`

	// Relationship indicators
	SMA50AboveSMA200 bool `csv:"SMA50_Above_SMA200"`

	// Descriptions
	GoldenDeathCrossDesc    string `csv:"Golden_Death_Cross_Desc"`
	PriceSMA10CrossoverDesc string `csv:"Price_SMA10_Crossover_Desc"`
	PriceCrossoverDesc      string `csv:"Price_Crossover_Desc"`
	RSIDesc                 string `csv:"RSI_Desc"`
	StochasticDesc          string `csv:"Stochastic_Desc"`
	CMFDesc                 string `csv:"CMF_Desc"`
	MACDDesc                string `csv:"MACD_Desc"`
	OBVDesc                 string `csv:"OBV_Desc"`
	PSARDesc                string `csv:"PSAR_Desc"`
	ATRDesc                 string `csv:"ATR_Desc"`
}

// Ticker represents a stock ticker
type Ticker struct {
	Symbol string `csv:"Ticker"`
}

// LiquidityScore represents liquidity calculation for a ticker
type LiquidityScore struct {
	Ticker         string          `csv:"Ticker"`
	LiquidityScore decimal.Decimal `csv:"Liquidity_Score"`
	Rank           int             `csv:"Rank"`
}

// StrategyResult represents the result of a trading strategy
type StrategyResult struct {
	Ticker     string          `csv:"Ticker"`
	Strategy   string          `csv:"Strategy"`
	Action     string          `csv:"Action"`
	Date       time.Time       `csv:"Date"`
	Price      decimal.Decimal `csv:"Price"`
	Quantity   int             `csv:"Quantity"`
	PnL        decimal.Decimal `csv:"PnL"`
	TotalValue decimal.Decimal `csv:"Total_Value"`
}

// BacktestResult represents backtesting results
type BacktestResult struct {
	Ticker        string          `csv:"Ticker"`
	Strategy      string          `csv:"Strategy"`
	TotalReturn   decimal.Decimal `csv:"Total_Return"`
	WinRate       decimal.Decimal `csv:"Win_Rate"`
	MaxDrawdown   decimal.Decimal `csv:"Max_Drawdown"`
	SharpeRatio   decimal.Decimal `csv:"Sharpe_Ratio"`
	ProfitFactor  decimal.Decimal `csv:"Profit_Factor"`
	TotalTrades   int             `csv:"Total_Trades"`
	WinningTrades int             `csv:"Winning_Trades"`
	LosingTrades  int             `csv:"Losing_Trades"`
	AvgTradeDays  decimal.Decimal `csv:"Avg_Trade_Days"`
	AvgWin        decimal.Decimal `csv:"Avg_Win"`
	AvgLoss       decimal.Decimal `csv:"Avg_Loss"`
	MaxWin        decimal.Decimal `csv:"Max_Win"`
	MaxLoss       decimal.Decimal `csv:"Max_Loss"`
	StartDate     time.Time       `csv:"Start_Date"`
	EndDate       time.Time       `csv:"End_Date"`
	InitialCash   decimal.Decimal `csv:"Initial_Cash"`
	FinalValue    decimal.Decimal `csv:"Final_Value"`
}

// Position represents an open trading position
type Position struct {
	Ticker        string          `csv:"Ticker"`
	Strategy      string          `csv:"Strategy"`
	EntryDate     time.Time       `csv:"Entry_Date"`
	EntryPrice    decimal.Decimal `csv:"Entry_Price"`
	Quantity      int             `csv:"Quantity"`
	CurrentPrice  decimal.Decimal `csv:"Current_Price"`
	CurrentValue  decimal.Decimal `csv:"Current_Value"`
	UnrealizedPnL decimal.Decimal `csv:"Unrealized_PnL"`
	StopLoss      decimal.Decimal `csv:"Stop_Loss"`
	TakeProfit    decimal.Decimal `csv:"Take_Profit"`
	TradeType     string          `csv:"Trade_Type"` // LONG, SHORT
}

// Trade represents a completed trade
type Trade struct {
	TradeID     int             `csv:"Trade_ID"`
	Ticker      string          `csv:"Ticker"`
	Strategy    string          `csv:"Strategy"`
	EntryDate   time.Time       `csv:"Entry_Date"`
	ExitDate    time.Time       `csv:"Exit_Date"`
	EntryPrice  decimal.Decimal `csv:"Entry_Price"`
	ExitPrice   decimal.Decimal `csv:"Exit_Price"`
	EntrySignal string          `csv:"Entry_Signal"`
	ExitSignal  string          `csv:"Exit_Signal"`
	Quantity    int             `csv:"Quantity"`
	PnL         decimal.Decimal `csv:"PnL"`
	PnLPercent  decimal.Decimal `csv:"PnL_Percent"`
	HoldingDays int             `csv:"Holding_Days"`
	TradeType   string          `csv:"Trade_Type"` // LONG, SHORT
	Commission  decimal.Decimal `csv:"Commission"`
	ExitReason  string          `csv:"Exit_Reason"` // SIGNAL, STOP_LOSS, TAKE_PROFIT, TIME_LIMIT
}

// Portfolio represents portfolio state at a point in time
type Portfolio struct {
	Date            time.Time       `csv:"Date"`
	Cash            decimal.Decimal `csv:"Cash"`
	EquityValue     decimal.Decimal `csv:"Equity_Value"`
	TotalValue      decimal.Decimal `csv:"Total_Value"`
	DailyReturn     decimal.Decimal `csv:"Daily_Return"`
	TotalReturn     decimal.Decimal `csv:"Total_Return"`
	Drawdown        decimal.Decimal `csv:"Drawdown"`
	ActivePositions int             `csv:"Active_Positions"`
	DaysSinceStart  int             `csv:"Days_Since_Start"`
}

// BacktestConfig represents backtesting configuration
type BacktestConfig struct {
	InitialCash       decimal.Decimal `json:"initial_cash"`
	Commission        decimal.Decimal `json:"commission_per_trade"`
	CommissionPercent decimal.Decimal `json:"commission_percent"`
	MaxPositions      int             `json:"max_positions"`
	PositionSize      decimal.Decimal `json:"position_size_percent"`  // Percentage of portfolio per trade
	RiskPerTrade      decimal.Decimal `json:"risk_per_trade_percent"` // Maximum risk per trade
	StopLoss          decimal.Decimal `json:"stop_loss_percent"`      // Stop loss percentage
	TakeProfit        decimal.Decimal `json:"take_profit_percent"`    // Take profit percentage
	MaxHoldingDays    int             `json:"max_holding_days"`       // Maximum days to hold a position
	StartDate         time.Time       `json:"start_date"`
	EndDate           time.Time       `json:"end_date"`
	Strategies        []string        `json:"strategies"`          // Which strategies to test
	Tickers           []string        `json:"tickers"`             // Which tickers to include
	UseSignalStrength bool            `json:"use_signal_strength"` // Whether to use signal strength (Strong Buy vs Buy)
	ReinvestDividends bool            `json:"reinvest_dividends"`
	Benchmark         string          `json:"benchmark"` // Benchmark ticker for comparison
}

// StrategyPerformance represents performance metrics for a single strategy
type StrategyPerformance struct {
	Strategy          string          `csv:"Strategy"`
	Ticker            string          `csv:"Ticker"`
	TotalReturn       decimal.Decimal `csv:"Total_Return"`
	AnnualizedReturn  decimal.Decimal `csv:"Annualized_Return"`
	Volatility        decimal.Decimal `csv:"Volatility"`
	SharpeRatio       decimal.Decimal `csv:"Sharpe_Ratio"`
	MaxDrawdown       decimal.Decimal `csv:"Max_Drawdown"`
	WinRate           decimal.Decimal `csv:"Win_Rate"`
	ProfitFactor      decimal.Decimal `csv:"Profit_Factor"`
	TotalTrades       int             `csv:"Total_Trades"`
	AvgTradeDuration  decimal.Decimal `csv:"Avg_Trade_Duration"`
	BestTrade         decimal.Decimal `csv:"Best_Trade"`
	WorstTrade        decimal.Decimal `csv:"Worst_Trade"`
	ConsecutiveWins   int             `csv:"Consecutive_Wins"`
	ConsecutiveLosses int             `csv:"Consecutive_Losses"`
	RecoveryFactor    decimal.Decimal `csv:"Recovery_Factor"`
	CalmarRatio       decimal.Decimal `csv:"Calmar_Ratio"`
}

// ProcessingReport represents detailed processing statistics for each ticker
type ProcessingReport struct {
	Ticker             string    `csv:"Ticker"`
	Sector             string    `csv:"Sector"`
	CompanyName        string    `csv:"Company_Name"`
	Status             string    `csv:"Status"` // SUCCESS, ERROR, PARTIAL
	StartTime          time.Time `csv:"Start_Time"`
	EndTime            time.Time `csv:"End_Time"`
	ProcessingDuration string    `csv:"Processing_Duration"`
	PagesBeforeUpdate  int       `csv:"Pages_Before_Update"`
	PagesLoaded        int       `csv:"Pages_Loaded"`
	DaysLoaded         int       `csv:"Days_Loaded"`
	NewRowsCount       int       `csv:"New_Rows_Count"`
	TotalRowsInCSV     int       `csv:"Total_Rows_In_CSV"`
	ErrorMessage       string    `csv:"Error_Message"`
	Recommendation     string    `csv:"Recommendation"`
	FileSize           int64     `csv:"File_Size_Bytes"`
	LastDataDate       string    `csv:"Last_Data_Date"`
	FirstDataDate      string    `csv:"First_Data_Date"`
	DataQualityScore   string    `csv:"Data_Quality_Score"` // EXCELLENT, GOOD, POOR, FAILED
}

// TimingReport represents detailed timing analysis for performance optimization
type TimingReport struct {
	Ticker                 string        `csv:"Ticker"`
	TotalProcessingTime    time.Duration `csv:"Total_Processing_Time_Ms"`
	NavigationTime         time.Duration `csv:"Navigation_Time_Ms"`
	PageLoadTime           time.Duration `csv:"Page_Load_Time_Ms"`
	CompanyCodeSetTime     time.Duration `csv:"Company_Code_Set_Time_Ms"`
	AjaxTriggerTime        time.Duration `csv:"Ajax_Trigger_Time_Ms"`
	DataExtractionTime     time.Duration `csv:"Data_Extraction_Time_Ms"`
	PaginationTime         time.Duration `csv:"Pagination_Time_Ms"`
	DataParsingTime        time.Duration `csv:"Data_Parsing_Time_Ms"`
	SortingTime            time.Duration `csv:"Sorting_Time_Ms"`
	ChangeCalculationTime  time.Duration `csv:"Change_Calculation_Time_Ms"`
	CSVSaveTime            time.Duration `csv:"CSV_Save_Time_Ms"`
	DeduplicationTime      time.Duration `csv:"Deduplication_Time_Ms"`
	FileOperationsTime     time.Duration `csv:"File_Operations_Time_Ms"`
	AveragePageTime        time.Duration `csv:"Average_Page_Time_Ms"`
	SlowestPageTime        time.Duration `csv:"Slowest_Page_Time_Ms"`
	FastestPageTime        time.Duration `csv:"Fastest_Page_Time_Ms"`
	AjaxWaitTime           time.Duration `csv:"Ajax_Wait_Time_Ms"`
	BrowserOverheadTime    time.Duration `csv:"Browser_Overhead_Time_Ms"`
	PagesProcessed         int           `csv:"Pages_Processed"`
	TotalAjaxCalls         int           `csv:"Total_Ajax_Calls"`
	PerformanceScore       string        `csv:"Performance_Score"` // EXCELLENT, GOOD, AVERAGE, POOR
	BottleneckFunction     string        `csv:"Bottleneck_Function"`
	OptimizationSuggestion string        `csv:"Optimization_Suggestion"`
}
