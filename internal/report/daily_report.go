package report

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"isx-auto-scrapper/internal/common"

	"github.com/xuri/excelize/v2"
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
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	LastTraded   string    `json:"last_traded"`
	Sparkline    []float64 `json:"sparkline"`
	Open         float64   `json:"open"`
	High         float64   `json:"high"`
	Low          float64   `json:"low"`
	AvgPrice     float64   `json:"avg_price"`
	PrevAvgPrice float64   `json:"prev_avg_price"`
	Close        float64   `json:"close"`
	PrevClose    float64   `json:"prev_close"`
	ChangePct    float64   `json:"change_pct"`
	Trades       int64     `json:"trades"`
	Volume       int64     `json:"volume"`
	Value        float64   `json:"value"`
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

	// -------------------------------------------------
	// Pass 1: discover the latest date that had ANY trade
	// -------------------------------------------------
	latestTradeDate := ""

	for _, t := range tickers {
		filename := fmt.Sprintf("raw_%s.csv", t.Symbol)
		content, err := os.ReadFile(filename)
		if err != nil {
			continue // skip missing files
		}
		lines := strings.Split(string(content), "\n")

		// Scan from newest to oldest until we find a row with volume>0
		for i := len(lines) - 1; i >= 0; i-- {
			row := strings.TrimSpace(lines[i])
			if row == "" {
				continue
			}
			parts := strings.Split(row, ",")
			if len(parts) < 9 {
				continue // malformed
			}
			closeVal, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			volumeVal, _ := strconv.ParseInt(strings.TrimSpace(parts[8]), 10, 64)
			if volumeVal > 0 || closeVal > 0 {
				dateStr := strings.TrimSpace(parts[0])
				if dateStr > latestTradeDate {
					latestTradeDate = dateStr
				}
				break // done with this ticker
			}
		}
	}

	if latestTradeDate == "" {
		return nil, fmt.Errorf("could not determine latest trade date")
	}

	// -------------------------------------------------
	// Pass 2: collect data for that latestTradeDate
	// -------------------------------------------------
	var traded []CompanyData
	var nonTraded []CompanyData

	for _, t := range tickers {
		filename := fmt.Sprintf("raw_%s.csv", t.Symbol)
		content, err := os.ReadFile(filename)
		if err != nil {
			continue
		}
		lines := strings.Split(string(content), "\n")

		var targetLine string
		var prevLine string
		targetIdx := -1

		// Iterate bottom-up to find the line that matches latestTradeDate
		for i := len(lines) - 1; i >= 0; i-- {
			if strings.TrimSpace(lines[i]) == "" {
				continue
			}
			parts := strings.Split(lines[i], ",")
			if len(parts) < 9 {
				continue
			}
			dateStr := strings.TrimSpace(parts[0])
			if dateStr == latestTradeDate {
				targetLine = lines[i]
				targetIdx = i
				// the row immediately before targetLine (older in time)
				for j := i - 1; j >= 0; j-- {
					if strings.TrimSpace(lines[j]) != "" {
						prevLine = lines[j]
						break
					}
				}
				break
			}
		}

		if targetLine == "" {
			// No record on latestTradeDate → non-traded

			// search for the most recent line with volume>0 (bottom-up)
			var lastTradeLine string
			for i := len(lines) - 1; i >= 0; i-- {
				if strings.TrimSpace(lines[i]) == "" {
					continue
				}
				parts := strings.Split(lines[i], ",")
				if len(parts) < 9 {
					continue
				}
				closeVal, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				volumeVal, _ := strconv.ParseInt(strings.TrimSpace(parts[8]), 10, 64)
				if volumeVal > 0 || closeVal > 0 {
					lastTradeLine = lines[i]
					break
				}
			}

			cd := CompanyData{Code: t.Symbol, Name: t.CompanyName}

			if lastTradeLine != "" {
				lp := strings.Split(lastTradeLine, ",")
				cd.LastTraded = strings.TrimSpace(lp[0])

				cd.Close, _ = strconv.ParseFloat(strings.TrimSpace(lp[1]), 64)
				cd.Open, _ = strconv.ParseFloat(strings.TrimSpace(lp[2]), 64)
				cd.High, _ = strconv.ParseFloat(strings.TrimSpace(lp[3]), 64)
				cd.Low, _ = strconv.ParseFloat(strings.TrimSpace(lp[4]), 64)
				cd.Volume, _ = strconv.ParseInt(strings.TrimSpace(lp[8]), 10, 64)
				if len(lp) > 9 {
					cd.Trades, _ = strconv.ParseInt(strings.TrimSpace(lp[9]), 10, 64)
				}
				cd.AvgPrice = (cd.Open + cd.High + cd.Low + cd.Close) / 4

				// Build sparkline: gather last 7 closes ending at lastTradeLine index
				// first, find index of lastTradeLine
				idx := -1
				for z := len(lines) - 1; z >= 0; z-- {
					if strings.TrimSpace(lines[z]) == strings.TrimSpace(lastTradeLine) {
						idx = z
						break
					}
				}
				if idx >= 0 {
					var spark []float64
					cnt := 0
					for k := idx; k >= 0 && cnt < 7; k-- {
						row := strings.TrimSpace(lines[k])
						if row == "" {
							continue
						}
						partsK := strings.Split(row, ",")
						if len(partsK) < 2 {
							continue
						}
						cv, _ := strconv.ParseFloat(strings.TrimSpace(partsK[1]), 64)
						if cv > 0 {
							spark = append([]float64{cv}, spark...)
							cnt++
						}
					}
					cd.Sparkline = spark
				}
				cd.Value = cd.Close * float64(cd.Volume)
			}

			nonTraded = append(nonTraded, cd)
			continue
		}

		parts := strings.Split(targetLine, ",")
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

		// collect sparkline closes (last 7 closes including today)
		var spark []float64
		// iterate backwards from i (targetLine index) up to 6 previous non-empty lines
		count := 0
		for k := targetIdx; k >= 0 && count < 7; k-- {
			row := strings.TrimSpace(lines[k])
			if row == "" {
				continue
			}
			partsK := strings.Split(row, ",")
			if len(partsK) < 2 {
				continue
			}
			cVal, _ := strconv.ParseFloat(strings.TrimSpace(partsK[1]), 64)
			if cVal > 0 {
				spark = append([]float64{cVal}, spark...) // prepend to keep chronological order
				count++
			}
		}

		// Previous day numbers (if available)
		prevClose := 0.0
		prevAvg := 0.0
		if prevLine != "" {
			pp := strings.Split(prevLine, ",")
			if len(pp) >= 5 {
				pClose, _ := strconv.ParseFloat(strings.TrimSpace(pp[1]), 64)
				pOpen, _ := strconv.ParseFloat(strings.TrimSpace(pp[2]), 64)
				pHigh, _ := strconv.ParseFloat(strings.TrimSpace(pp[3]), 64)
				pLow, _ := strconv.ParseFloat(strings.TrimSpace(pp[4]), 64)
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
			LastTraded:   latestTradeDate,
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
			Sparkline:    spark,
		}

		// Presence of a row on latestTradeDate means the company traded that day.
		traded = append(traded, cd)
	}

	// Always return non-nil slices to avoid null in JSON
	if traded == nil {
		traded = []CompanyData{}
	}
	if nonTraded == nil {
		nonTraded = []CompanyData{}
	}

	// Persist CSVs for downstream calculations
	_ = saveCompaniesCSV(traded, fmt.Sprintf("traded_%s.csv", latestTradeDate))
	_ = saveCompaniesCSV(nonTraded, fmt.Sprintf("non_traded_%s.csv", latestTradeDate))

	buildTop := func(src []CompanyData, less func(a, b CompanyData) bool) []ReportEntry {
		temp := make([]CompanyData, len(src))
		copy(temp, src)

		sort.Slice(temp, func(i, j int) bool {
			return less(temp[i], temp[j])
		})

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

	topVol := buildTop(traded, func(a, b CompanyData) bool { return a.Volume > b.Volume })
	topVal := buildTop(traded, func(a, b CompanyData) bool { return a.Value > b.Value })
	topGain := buildTop(traded, func(a, b CompanyData) bool { return a.ChangePct > b.ChangePct })
	topLoss := buildTop(traded, func(a, b CompanyData) bool { return a.ChangePct < b.ChangePct })

	return &DailyReport{
		Date:      latestTradeDate,
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

// saveCompaniesCSV writes full rows to a CSV file
func saveCompaniesCSV(list []CompanyData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	header := []string{"Code", "Company", "Open", "High", "Low", "AvgPrice", "PrevAvg", "Close", "PrevClose", "ChangePct", "Trades", "Volume", "Value"}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, c := range list {
		rec := []string{
			c.Code,
			c.Name,
			fmt.Sprintf("%f", c.Open),
			fmt.Sprintf("%f", c.High),
			fmt.Sprintf("%f", c.Low),
			fmt.Sprintf("%f", c.AvgPrice),
			fmt.Sprintf("%f", c.PrevAvgPrice),
			fmt.Sprintf("%f", c.Close),
			fmt.Sprintf("%f", c.PrevClose),
			fmt.Sprintf("%f", c.ChangePct),
			fmt.Sprintf("%d", c.Trades),
			fmt.Sprintf("%d", c.Volume),
			fmt.Sprintf("%f", c.Value),
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}
	return nil
}
