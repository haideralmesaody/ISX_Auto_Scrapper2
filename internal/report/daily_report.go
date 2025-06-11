package report

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"isx-auto-scrapper/internal/common"
)

// ReportEntry represents a single row in the top mover tables.
type ReportEntry struct {
	Ticker    string  `json:"ticker"`
	Name      string  `json:"name"`
	Close     float64 `json:"close"`
	ChangePct float64 `json:"change_pct"`
	Volume    int64   `json:"volume"`
	Value     float64 `json:"value"`
}

// CompanyData holds full trading information for a company.
type CompanyData struct {
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Open         float64 `json:"open"`
	High         float64 `json:"high"`
	Low          float64 `json:"low"`
	AvgPrice     float64 `json:"avg_price"`
	PrevAvgPrice float64 `json:"prev_avg_price"`
	Close        float64 `json:"close"`
	PrevClose    float64 `json:"prev_close"`
	ChangePct    float64 `json:"change_pct"`
	Trades       int64   `json:"trades"`
	Volume       int64   `json:"volume"`
	Value        float64 `json:"value"`
}

// DailyReport aggregates all sections for the daily market report.
type DailyReport struct {
	Date      string        `json:"date"`
	TopVolume []ReportEntry `json:"top_volume"`
	TopValue  []ReportEntry `json:"top_value"`
	TopGain   []ReportEntry `json:"top_gain"`
	TopLoss   []ReportEntry `json:"top_loss"`
	Traded    []CompanyData `json:"traded"`
	NonTraded []CompanyData `json:"non_traded"`
}

// GenerateDailyReport builds a DailyReport from the latest raw_*.csv files.
func GenerateDailyReport(_ time.Time) (*DailyReport, error) {
	tickers, err := common.LoadTickersWithInfo("TICKERS.csv")
	if err != nil {
		return nil, err
	}

	var traded []CompanyData
	var nonTraded []CompanyData
	latestDate := ""

	for _, t := range tickers {
		filename := fmt.Sprintf("raw_%s.csv", t.Symbol)
		content, err := os.ReadFile(filename)
		if err != nil {
			continue
		}
		lines := strings.Split(string(content), "\n")
		if len(lines) < 2 {
			continue
		}
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
			continue
		}
		parts := strings.Split(lastLine, ",")
		if len(parts) < 10 {
			continue
		}
		dateStr := strings.TrimSpace(parts[0])
		if dateStr > latestDate {
			latestDate = dateStr
		}
		closeVal, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		openVal, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
		highVal, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		lowVal, _ := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)
		volumeVal, _ := strconv.ParseInt(strings.TrimSpace(parts[8]), 10, 64)
		tradesVal := int64(0)
		if len(parts) > 9 {
			tradesVal, _ = strconv.ParseInt(strings.TrimSpace(parts[9]), 10, 64)
		}
		avgPrice := (openVal + highVal + lowVal + closeVal) / 4
		value := closeVal * float64(volumeVal)

		prevClose := 0.0
		prevAvg := 0.0
		if prevLine != "" {
			prevParts := strings.Split(prevLine, ",")
			if len(prevParts) >= 5 {
				pClose, _ := strconv.ParseFloat(strings.TrimSpace(prevParts[1]), 64)
				pOpen, _ := strconv.ParseFloat(strings.TrimSpace(prevParts[2]), 64)
				pHigh, _ := strconv.ParseFloat(strings.TrimSpace(prevParts[3]), 64)
				pLow, _ := strconv.ParseFloat(strings.TrimSpace(prevParts[4]), 64)
				prevClose = pClose
				prevAvg = (pOpen + pHigh + pLow + pClose) / 4
			}
		}
		changePct := 0.0
		if prevClose != 0 {
			changePct = ((closeVal - prevClose) / prevClose) * 100
		}

		cd := CompanyData{
			Code:         t.Symbol,
			Name:         t.CompanyName,
			Open:         openVal,
			High:         highVal,
			Low:          lowVal,
			AvgPrice:     avgPrice,
			PrevAvgPrice: prevAvg,
			Close:        closeVal,
			PrevClose:    prevClose,
			ChangePct:    changePct,
			Trades:       tradesVal,
			Volume:       volumeVal,
			Value:        value,
		}
		if volumeVal > 0 {
			traded = append(traded, cd)
		} else {
			nonTraded = append(nonTraded, cd)
		}
	}

	buildTop := func(src []CompanyData, less func(i, j int) bool) []ReportEntry {
		temp := make([]CompanyData, len(src))
		copy(temp, src)
		sort.Slice(temp, less)
		if len(temp) > 5 {
			temp = temp[:5]
		}
		out := make([]ReportEntry, len(temp))
		for i, c := range temp {
			out[i] = ReportEntry{
				Ticker:    c.Code,
				Name:      c.Name,
				Close:     c.Close,
				ChangePct: c.ChangePct,
				Volume:    c.Volume,
				Value:     c.Value,
			}
		}
		return out
	}

	topVol := buildTop(traded, func(i, j int) bool { return traded[i].Volume > traded[j].Volume })
	topVal := buildTop(traded, func(i, j int) bool { return traded[i].Value > traded[j].Value })
	topGain := buildTop(traded, func(i, j int) bool { return traded[i].ChangePct > traded[j].ChangePct })
	topLoss := buildTop(traded, func(i, j int) bool { return traded[i].ChangePct < traded[j].ChangePct })

	return &DailyReport{
		Date:      latestDate,
		TopVolume: topVol,
		TopValue:  topVal,
		TopGain:   topGain,
		TopLoss:   topLoss,
		Traded:    traded,
		NonTraded: nonTraded,
	}, nil
}

// SaveDailyReportExcel writes the report to an Excel file with one sheet per section.
func SaveDailyReportExcel(r *DailyReport, path string) error {
	f := excelize.NewFile()
	// Top Volume
	f.NewSheet("TopVolume")
	f.SetSheetRow("TopVolume", "A1", &[]string{"Ticker", "Company", "Close", "Change%", "Volume", "Value"})
	for i, row := range r.TopVolume {
		f.SetSheetRow("TopVolume", fmt.Sprintf("A%d", i+2), &[]interface{}{row.Ticker, row.Name, row.Close, row.ChangePct, row.Volume, row.Value})
	}
	// Top Value
	f.NewSheet("TopValue")
	f.SetSheetRow("TopValue", "A1", &[]string{"Ticker", "Company", "Close", "Change%", "Volume", "Value"})
	for i, row := range r.TopValue {
		f.SetSheetRow("TopValue", fmt.Sprintf("A%d", i+2), &[]interface{}{row.Ticker, row.Name, row.Close, row.ChangePct, row.Volume, row.Value})
	}
	// Top Gain
	f.NewSheet("TopGain")
	f.SetSheetRow("TopGain", "A1", &[]string{"Ticker", "Company", "Close", "Change%", "Volume", "Value"})
	for i, row := range r.TopGain {
		f.SetSheetRow("TopGain", fmt.Sprintf("A%d", i+2), &[]interface{}{row.Ticker, row.Name, row.Close, row.ChangePct, row.Volume, row.Value})
	}
	// Top Loss
	f.NewSheet("TopLoss")
	f.SetSheetRow("TopLoss", "A1", &[]string{"Ticker", "Company", "Close", "Change%", "Volume", "Value"})
	for i, row := range r.TopLoss {
		f.SetSheetRow("TopLoss", fmt.Sprintf("A%d", i+2), &[]interface{}{row.Ticker, row.Name, row.Close, row.ChangePct, row.Volume, row.Value})
	}
	// Traded
	f.NewSheet("Traded")
	f.SetSheetRow("Traded", "A1", &[]string{"Code", "Company", "Open", "High", "Low", "AvgPrice", "PrevAvg", "Close", "PrevClose", "Change%", "Trades", "Volume", "Value"})
	for i, row := range r.Traded {
		f.SetSheetRow("Traded", fmt.Sprintf("A%d", i+2), &[]interface{}{row.Code, row.Name, row.Open, row.High, row.Low, row.AvgPrice, row.PrevAvgPrice, row.Close, row.PrevClose, row.ChangePct, row.Trades, row.Volume, row.Value})
	}
	// Non traded
	f.NewSheet("NonTraded")
	f.SetSheetRow("NonTraded", "A1", &[]string{"Code", "Company", "Open", "High", "Low", "AvgPrice", "PrevAvg", "Close", "PrevClose", "Change%", "Trades", "Volume", "Value"})
	for i, row := range r.NonTraded {
		f.SetSheetRow("NonTraded", fmt.Sprintf("A%d", i+2), &[]interface{}{row.Code, row.Name, row.Open, row.High, row.Low, row.AvgPrice, row.PrevAvgPrice, row.Close, row.PrevClose, row.ChangePct, row.Trades, row.Volume, row.Value})
	}

	if idx, err := f.GetSheetIndex("TopVolume"); err == nil {
		f.SetActiveSheet(idx)
	}
	return f.SaveAs(path)
}
