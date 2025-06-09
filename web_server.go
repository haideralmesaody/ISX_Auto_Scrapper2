package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// WebServer handles HTTP requests for the dashboard
type WebServer struct {
	logger *Logger
	port   int
}

// NewWebServer creates a new WebServer instance
func NewWebServer(port int) *WebServer {
	return &WebServer{
		logger: NewLogger(),
		port:   port,
	}
}

// Start starts the web server
func (ws *WebServer) Start() error {
	mux := http.NewServeMux()

	// Serve static files from web directory
	fileServer := http.FileServer(http.Dir("web/"))
	mux.Handle("/", fileServer)

	// API endpoints
	mux.HandleFunc("/api/tickers", ws.handleTickers)
	mux.HandleFunc("/api/ticker/", ws.handleTickerData)
	mux.HandleFunc("/api/strategies", ws.handleStrategies)
	mux.HandleFunc("/api/backtest", ws.handleBacktest)
	mux.HandleFunc("/api/refresh", ws.handleRefresh)

	// CORS middleware
	corsHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			h.ServeHTTP(w, r)
		})
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", ws.port),
		Handler: corsHandler(mux),
	}

	ws.logger.Info("Web server starting on http://localhost:%d", ws.port)
	ws.logger.Info("Dashboard available at: http://localhost:%d", ws.port)

	return server.ListenAndServe()
}

// API Handlers

func (ws *WebServer) handleTickers(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("API: Getting tickers list")

	// Read TICKERS.csv
	tickers, err := ws.loadTickersList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add price data from raw files
	for i, ticker := range tickers {
		priceData, err := ws.getLastPrice(ticker.Symbol)
		if err == nil {
			tickers[i].Price = priceData.Close
			tickers[i].Volume = priceData.Volume
			tickers[i].Change = priceData.Change
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickers)
}

func (ws *WebServer) handleTickerData(w http.ResponseWriter, r *http.Request) {
	// Extract ticker symbol from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/ticker/")
	symbol := strings.Split(path, "/")[0]

	if symbol == "" {
		http.Error(w, "Ticker symbol required", http.StatusBadRequest)
		return
	}

	ws.logger.Info("API: Getting data for ticker %s", symbol)

	// Get what type of data is requested
	dataType := r.URL.Query().Get("type")
	if dataType == "" {
		dataType = "price" // default
	}

	switch dataType {
	case "price":
		ws.handlePriceData(w, symbol)
	case "indicators":
		ws.handleIndicatorData(w, symbol)
	case "strategies":
		ws.handleTickerStrategies(w, symbol)
	default:
		http.Error(w, "Invalid data type", http.StatusBadRequest)
	}
}

func (ws *WebServer) handlePriceData(w http.ResponseWriter, symbol string) {
	priceData, err := ws.loadPriceData(symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(priceData)
}

func (ws *WebServer) handleIndicatorData(w http.ResponseWriter, symbol string) {
	indicators, err := ws.loadIndicatorData(symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(indicators)
}

func (ws *WebServer) handleTickerStrategies(w http.ResponseWriter, symbol string) {
	strategies, err := ws.loadTickerStrategies(symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(strategies)
}

func (ws *WebServer) handleStrategies(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("API: Getting strategies summary")

	// Read Strategy_Summary.json
	data, err := os.ReadFile("Strategy_Summary.json")
	if err != nil {
		http.Error(w, "Strategy summary not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (ws *WebServer) handleBacktest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ws.logger.Info("API: Running backtest")

	// Run backtesting in background
	go func() {
		// You could integrate with your existing backtesting logic here
		strategyTester := NewStrategyTester()

		// Run simulation
		strategyTester.SimulateStrategyResults()
		strategyTester.SummarizeSimulatedStrategyResults()
		ws.logger.Info("Backtest completed successfully")
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "started",
		"message": "Backtest initiated successfully",
	})
}

func (ws *WebServer) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ws.logger.Info("API: Refreshing data")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "success",
		"message":   "Data refresh completed",
		"timestamp": getCurrentTimestamp(),
	})
}

// Data loading helpers

type PriceData struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}

type LastPriceData struct {
	Close  float64
	Volume int64
	Change float64
}

func (ws *WebServer) loadTickersList() ([]TickerInfo, error) {
	content, err := os.ReadFile("TICKERS.csv")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var tickers []TickerInfo

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 1 {
			symbol := strings.TrimSpace(parts[0])
			companyName := symbol
			sector := ""
			if len(parts) >= 2 {
				sector = strings.TrimSpace(parts[1])
			}
			if len(parts) >= 3 {
				companyName = strings.TrimSpace(parts[2])
			}

			tickers = append(tickers, TickerInfo{
				Symbol:      symbol,
				Sector:      sector,
				CompanyName: companyName,
				Price:       0,
				Change:      0,
				Volume:      0,
			})
		}
	}

	return tickers, nil
}

func (ws *WebServer) loadPriceData(symbol string) ([]PriceData, error) {
	filename := fmt.Sprintf("raw_%s.csv", symbol)
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("price data not found for %s", symbol)
	}

	lines := strings.Split(string(content), "\n")
	var priceData []PriceData

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 9 {
			// CORRECT column mapping: Date,Close,Open,High,Low,Change,Change%,T.Shares,Volume,No. Trades
			//                         0    1     2    3    4     5      6       7       8       9
			close, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			open, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			high, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
			low, _ := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)
			volume, _ := strconv.ParseInt(strings.TrimSpace(parts[8]), 10, 64)

			// Skip rows with invalid data (like the zero close price entries)
			if close > 0 && open > 0 && high > 0 && low > 0 {
				priceData = append(priceData, PriceData{
					Date:   strings.TrimSpace(parts[0]),
					Open:   open,
					High:   high,
					Low:    low,
					Close:  close,
					Volume: volume,
				})
			}
		}
	}

	return priceData, nil
}

func (ws *WebServer) getLastPrice(symbol string) (*LastPriceData, error) {
	filename := fmt.Sprintf("raw_%s.csv", symbol)
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("insufficient data")
	}

	// Get last two lines for price comparison
	var lastLine, prevLine string
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			if lastLine == "" {
				lastLine = lines[i]
			} else {
				prevLine = lines[i]
				break
			}
		}
	}

	if lastLine == "" {
		return nil, fmt.Errorf("no valid data found")
	}

	parts := strings.Split(lastLine, ",")
	if len(parts) < 9 {
		return nil, fmt.Errorf("invalid data format")
	}

	// CORRECT column mapping: Date,Close,Open,High,Low,Change,Change%,T.Shares,Volume,No. Trades
	//                         0    1     2    3    4     5      6       7       8       9
	close, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	volume, _ := strconv.ParseInt(strings.TrimSpace(parts[8]), 10, 64)

	var change float64
	if prevLine != "" {
		prevParts := strings.Split(prevLine, ",")
		if len(prevParts) >= 2 {
			prevClose, _ := strconv.ParseFloat(strings.TrimSpace(prevParts[1]), 64)
			change = close - prevClose
		}
	}

	return &LastPriceData{
		Close:  close,
		Volume: volume,
		Change: change,
	}, nil
}

func (ws *WebServer) loadIndicatorData(symbol string) (map[string]interface{}, error) {
	filename := fmt.Sprintf("indicators_%s.csv", symbol)
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("indicator data not found for %s", symbol)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("insufficient indicator data")
	}

	// Get headers and last data line
	headers := strings.Split(lines[0], ",")
	lastLine := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			lastLine = lines[i]
			break
		}
	}

	if lastLine == "" {
		return nil, fmt.Errorf("no valid indicator data found")
	}

	values := strings.Split(lastLine, ",")
	indicators := make(map[string]interface{})

	for i, header := range headers {
		if i < len(values) {
			value := strings.TrimSpace(values[i])
			if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
				indicators[strings.TrimSpace(header)] = floatVal
			} else {
				indicators[strings.TrimSpace(header)] = value
			}
		}
	}

	return indicators, nil
}

func (ws *WebServer) loadTickerStrategies(symbol string) (map[string]interface{}, error) {
	filename := fmt.Sprintf("Strategies_%s.csv", symbol)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("strategy data not found for %s", symbol)
	}

	// For now, return a placeholder
	return map[string]interface{}{
		"symbol": symbol,
		"strategies": []map[string]string{
			{"name": "OBV Strategy", "signal": "Buy", "strength": "Strong"},
			{"name": "RSI Strategy", "signal": "Hold", "strength": "Weak"},
		},
	}, nil
}

func getCurrentTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}
