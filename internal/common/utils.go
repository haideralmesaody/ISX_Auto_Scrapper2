package common

import (
	"encoding/csv"
	"os"
)

// TickerInfo represents complete ticker information
type TickerInfo struct {
	Symbol      string    `csv:"Ticker" json:"symbol"`
	Sector      string    `csv:"Sector" json:"sector"`
	CompanyName string    `csv:"Name" json:"name"`
	Price       float64   `json:"price"`
	Change      float64   `json:"change"`
	Volume      int64     `json:"volume"`
	Open        float64   `json:"open"`
	High        float64   `json:"high"`
	Low         float64   `json:"low"`
	Value       float64   `json:"value"`
	Sparkline   []float64 `json:"sparkline"`
}

// LoadTickers loads ticker symbols from CSV file
func LoadTickers(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var tickers []string
	// Skip header row
	for i := 1; i < len(records); i++ {
		if len(records[i]) > 0 {
			tickers = append(tickers, records[i][0])
		}
	}

	return tickers, nil
}

// LoadTickersWithInfo loads complete ticker information from CSV file
func LoadTickersWithInfo(filename string) ([]TickerInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var tickers []TickerInfo
	// Skip header row (Ticker,Sector,Name)
	for i := 1; i < len(records); i++ {
		if len(records[i]) >= 3 {
			tickers = append(tickers, TickerInfo{
				Symbol:      records[i][0],
				Sector:      records[i][1],
				CompanyName: records[i][2],
			})
		}
	}

	return tickers, nil
}
