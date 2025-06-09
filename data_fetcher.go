package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/shopspring/decimal"
)

// DataFetcher handles web scraping of stock data
type DataFetcher struct {
	logger *Logger
	// Reporting fields
	currentReport *ProcessingReport
	startTime     time.Time
	pagesLoaded   int
	newRowsCount  int

	// Timing tracking fields
	timingReport      *TimingReport
	pageStartTimes    []time.Time
	pageDurations     []time.Duration
	ajaxCallCount     int
	totalAjaxWaitTime time.Duration
}

// NewDataFetcher creates a new DataFetcher instance
func NewDataFetcher() *DataFetcher {
	return &DataFetcher{
		logger: NewLogger(),
	}
}

// FetchData scrapes stock data for a given ticker
func (df *DataFetcher) FetchData(ticker string) error {
	_, err := df.FetchDataWithReport(ticker, "", "")
	return err
}

// FetchDataWithReport scrapes stock data for a given ticker and tracks detailed statistics
func (df *DataFetcher) FetchDataWithReport(ticker, sector, companyName string) (*ProcessingReport, error) {
	// Initialize reporting
	df.startTime = time.Now()
	df.pagesLoaded = 0
	df.newRowsCount = 0
	df.currentReport = &ProcessingReport{
		Ticker:      ticker,
		Sector:      sector,
		CompanyName: companyName,
		StartTime:   df.startTime,
		Status:      "PROCESSING",
	}

	// Initialize timing tracking
	df.timingReport = &TimingReport{
		Ticker:              ticker,
		TotalProcessingTime: 0,
		FastestPageTime:     time.Hour, // Initialize to max value
		SlowestPageTime:     0,
		PagesProcessed:      0,
		TotalAjaxCalls:      0,
	}
	df.pageStartTimes = []time.Time{}
	df.pageDurations = []time.Duration{}
	df.ajaxCallCount = 0
	df.totalAjaxWaitTime = 0

	filename := fmt.Sprintf("raw_%s.csv", ticker)

	// Check existing file and count rows
	pagesBeforeUpdate := 0
	if _, err := os.Stat(filename); err == nil {
		df.logger.Info("File '%s' already exists", filename)
		df.currentReport.Status = "UPDATING"

		// Count existing rows
		existingData, err := df.loadExistingData(filename)
		if err != nil {
			df.currentReport.ErrorMessage = fmt.Sprintf("Failed to load existing data: %v", err)
		} else {
			pagesBeforeUpdate = len(existingData)
			df.currentReport.PagesBeforeUpdate = pagesBeforeUpdate

			if len(existingData) > 0 {
				lastDate := existingData[len(existingData)-1].Date
				currentDate := time.Now()
				daysDiff := int(currentDate.Sub(lastDate).Hours() / 24)

				if daysDiff <= 1 {
					df.logger.Info("Data is up to date for ticker %s", ticker)
					df.currentReport.Status = "UP_TO_DATE"
					df.currentReport.EndTime = time.Now()
					df.currentReport.ProcessingDuration = time.Since(df.startTime).String()
					df.currentReport.TotalRowsInCSV = len(existingData)
					df.currentReport.DataQualityScore = "EXCELLENT"
					df.currentReport.Recommendation = "No action needed - data is current"
					return df.currentReport, nil
				}
			}
		}
	} else {
		df.currentReport.Status = "NEW"
	}

	// Setup chrome options for better performance
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // Disable headless mode to show browser window
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-features", "TranslateUI"),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("disable-images", true),      // Disable images for better performance
		chromedp.Flag("disable-javascript", false), // Keep JS enabled for AJAX
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-translate", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.WindowSize(1024, 768), // Smaller window for better performance
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second) // Increased from 45 to 60 seconds due to optimizations working well
	defer cancel()

	// Set up JavaScript dialog handler IMMEDIATELY before any navigation
	// This catches popups that appear during page load
	df.logger.Info("Setting up early popup detection before navigation...")
	dialogHandled := make(chan bool, 1)

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *page.EventJavascriptDialogOpening:
			df.logger.Info("EARLY POPUP DETECTED: Type=%s, Message='%s'", ev.Type, ev.Message)

			// Automatically accept the dialog immediately
			go func() {
				err := chromedp.Run(ctx, page.HandleJavaScriptDialog(true))
				if err != nil {
					df.logger.Error("Failed to handle early popup: %v", err)
				} else {
					df.logger.Info("Successfully auto-accepted early popup")
					select {
					case dialogHandled <- true:
					default:
					}
				}
			}()
		}
	})

	// Navigate to the company profile page with Performance tab active
	url := fmt.Sprintf("%s?currLanguage=en&companyCode=%s&activeTab=0", AppConfig.BaseURL, ticker)

	df.logger.Info("Fetching data from URL %s for ticker %s", url, ticker)

	var stockData []StockData

	// Navigate to page and wait for it to fully load
	navigationStart := time.Now()
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
	)

	// Check if a dialog was handled during navigation
	select {
	case <-dialogHandled:
		df.logger.Info("Early popup was automatically handled during navigation")
	default:
		// No early popup detected
	}

	// Wait for page to be completely ready using event-driven detection
	if err == nil {
		err = df.waitForPageComplete(ctx, 15*time.Second) // Max 15 seconds for initial page load
	}

	// Handle any popups that might appear (like year validation popup)
	if err == nil {
		err = df.handleInitialPopups(ctx)
	}

	// Set the fromDate and trigger search after popup handling
	if err == nil {
		err = df.setDateAndSearch(ctx, ticker)
	}

	df.timingReport.NavigationTime = time.Since(navigationStart)

	if err != nil {
		df.logger.Error("Failed to navigate to URL or handle popups: %v", err)
		return nil, fmt.Errorf("failed to navigate to URL or handle popups: %w", err)
	}

	// After setDateAndSearch, the page should already have the correct data loaded
	// So we skip the manual company code setting and AJAX triggering
	df.logger.Info("Skipping manual company code and AJAX setup since search was already performed")

	// Set timing placeholders since we skipped those steps
	df.timingReport.CompanyCodeSetTime = 0
	df.timingReport.AjaxTriggerTime = 0

	df.logger.Info("Page loaded successfully, extracting historical data...")

	// Extract data from the page
	dataExtractionStart := time.Now()
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Save HTML content for inspection
			var htmlContent string
			chromedp.Run(ctx,
				chromedp.Evaluate(`document.documentElement.outerHTML`, &htmlContent),
			)

			// Save to file
			htmlFile := fmt.Sprintf("final_page_%s.html", ticker)
			if err := os.WriteFile(htmlFile, []byte(htmlContent), 0644); err != nil {
				df.logger.Error("Failed to save HTML content: %v", err)
			} else {
				df.logger.Info("HTML content saved to %s for inspection", htmlFile)
			}

			return nil
		}),
		// Wait for data table to be populated with actual data
		chromedp.ActionFunc(func(ctx context.Context) error {
			return df.waitForDataTablePopulated(ctx, 8*time.Second) // Max 8 seconds wait
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return df.extractDataFromAllPages(ctx, &stockData, ticker)
		}),
	)
	df.timingReport.DataExtractionTime = time.Since(dataExtractionStart)

	// Handle temp file for testing - DO NOT rename
	tempFilename := fmt.Sprintf("raw_%s_temp.csv", ticker)
	finalFilename := fmt.Sprintf("raw_%s.csv", ticker)

	if err != nil {
		df.logger.Error("Extraction failed, but checking for temp file to save: %v", err)

		// Check if temp file exists and has data
		if _, statErr := os.Stat(tempFilename); statErr == nil {
			df.logger.Info("Temp file exists, keeping for testing (not renaming to %s)", finalFilename)
			// Commented out for testing:
			// if renameErr := os.Rename(tempFilename, finalFilename); renameErr != nil {
			//     df.logger.Error("Failed to rename temp file: %v", renameErr)
			// } else {
			//     df.logger.Info("Successfully saved partial data to %s", finalFilename)
			//     // Don't return the extraction error if we successfully saved partial data
			//     err = nil
			// }
		}

		if err != nil {
			return nil, fmt.Errorf("failed to scrape data for ticker %s: %w", ticker, err)
		}
	}

	if len(stockData) == 0 {
		df.logger.Info("No data found, trying alternative extraction method...")

		// Try to extract data directly from any visible tables
		err = chromedp.Run(ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				return df.extractDataDirectly(ctx, &stockData)
			}),
		)

		if err != nil {
			return nil, fmt.Errorf("failed to extract data using alternative method: %w", err)
		}
	}

	if len(stockData) == 0 {
		return nil, fmt.Errorf("no stock data found for ticker %s", ticker)
	}

	// Save data to CSV
	if err := df.saveDataToCSV(stockData, filename); err != nil {
		return nil, fmt.Errorf("failed to save data to CSV: %w", err)
	}

	df.logger.Info("Successfully fetched and saved %d records for ticker %s", len(stockData), ticker)
	df.currentReport.EndTime = time.Now()
	df.currentReport.ProcessingDuration = time.Since(df.startTime).String()
	df.currentReport.TotalRowsInCSV = len(stockData)
	df.currentReport.DataQualityScore = "EXCELLENT"
	df.currentReport.Recommendation = "No action needed - data is current"

	// Finalize report
	df.finalizeReport(nil)
	return df.currentReport, nil
}

// extractDataFromAllPages extracts data from all pages
func (df *DataFetcher) extractDataFromAllPages(ctx context.Context, stockData *[]StockData, ticker string) error {
	pageNum := 1
	maxRows := 2500 // Based on the HTML showing 2,379 total records

	// Load existing CSV data to check for overlaps
	existingData := make(map[string]bool)
	filename := fmt.Sprintf("raw_%s.csv", ticker)

	// First check if CSV file exists and load existing dates
	df.logger.Info("Attempting to load existing CSV file: %s", filename)
	if existingCSVData, err := df.loadExistingData(filename); err == nil {
		for _, data := range existingCSVData {
			dateKey := data.Date.Format("2006-01-02")
			existingData[dateKey] = true
		}

		// Log existing CSV data range for analysis
		if len(existingCSVData) > 0 {
			firstDate := existingCSVData[0].Date.Format("2006-01-02")
			lastDate := existingCSVData[len(existingCSVData)-1].Date.Format("2006-01-02")
			df.logger.Info("Loaded %d existing records from CSV file for overlap detection", len(existingData))
			df.logger.Info("CSV data range: %s to %s", firstDate, lastDate)
		} else {
			df.logger.Info("CSV file exists but contains no data")
		}

		// Also add any data already in memory
		for _, data := range *stockData {
			dateKey := data.Date.Format("2006-01-02")
			existingData[dateKey] = true
		}
	} else {
		// If no CSV file exists, just track in-memory data
		df.logger.Error("Failed to load existing CSV file %s: %v", filename, err)
		for _, data := range *stockData {
			dateKey := data.Date.Format("2006-01-02")
			existingData[dateKey] = true
		}
		df.logger.Info("No existing CSV file found, tracking %d in-memory records for overlap detection", len(existingData))
	}

	for len(*stockData) < maxRows {
		df.logger.Info("Extracting data from page %d", pageNum)

		// Track page processing time
		pageStart := time.Now()

		// Extract data from current page
		pageData, err := df.extractDataFromCurrentPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to extract data from page %d: %w", pageNum, err)
		}

		if len(pageData) == 0 {
			df.logger.Info("No data found on page %d, stopping", pageNum)
			break // No more data
		}

		// Log page data range for analysis
		if len(pageData) > 0 {
			firstPageDate := pageData[0].Date.Format("2006-01-02")
			lastPageDate := pageData[len(pageData)-1].Date.Format("2006-01-02")
			df.logger.Info("Page %d data range: %s to %s (%d records)", pageNum, firstPageDate, lastPageDate, len(pageData))
		}

		// Check for overlap with existing data and filter duplicates
		overlapCount := 0
		newDataCount := 0
		var newRecords []StockData
		var overlappingDates []string

		for _, data := range pageData {
			dateKey := data.Date.Format("2006-01-02")
			if existingData[dateKey] {
				overlapCount++
				overlappingDates = append(overlappingDates, dateKey)
			} else {
				newDataCount++
				newRecords = append(newRecords, data)
				existingData[dateKey] = true // Add to existing data map
			}
		}

		df.logger.Info("Page %d: %d new records, %d overlapping records", pageNum, newDataCount, overlapCount)

		// Log overlapping dates for analysis
		if overlapCount > 0 {
			df.logger.Info("Overlapping dates on page %d: %v", pageNum, overlappingDates)
		}

		// Add only the new records from this page
		*stockData = append(*stockData, newRecords...)
		df.logger.Info("Extracted %d records from page %d, total so far: %d", len(newRecords), pageNum, len(*stockData))

		// If we have any overlap with existing CSV data, this means we've reached data we already have
		// So we should stop pagination to avoid unnecessary processing
		if overlapCount > 0 {
			df.logger.Info("Overlap detected with existing data (%d overlapping records), stopping pagination", overlapCount)
			break
		}

		// Track statistics for reporting
		df.pagesLoaded = pageNum
		df.newRowsCount += newDataCount

		// Record page timing
		pageDuration := time.Since(pageStart)
		df.pageDurations = append(df.pageDurations, pageDuration)

		// Update fastest/slowest page times
		if pageDuration < df.timingReport.FastestPageTime {
			df.timingReport.FastestPageTime = pageDuration
		}
		if pageDuration > df.timingReport.SlowestPageTime {
			df.timingReport.SlowestPageTime = pageDuration
		}

		// Save CSV after each page to prevent data loss
		csvSaveStart := time.Now()
		tempFilename := fmt.Sprintf("raw_%s_temp.csv", ticker)

		// Sort data by date and recalculate changes before saving
		sortStart := time.Now()
		sortedData := df.sortAndRecalculateChanges(*stockData)
		df.timingReport.SortingTime += time.Since(sortStart)

		if err := df.saveDataToCSV(sortedData, tempFilename); err != nil {
			df.logger.Error("Failed to save temporary CSV after page %d: %v", pageNum, err)
		} else {
			df.logger.Info("Saved %d sorted records to %s after page %d", len(sortedData), tempFilename, pageNum)
		}
		df.timingReport.CSVSaveTime += time.Since(csvSaveStart)

		// Try to navigate to next page using AJAX
		paginationStart := time.Now()
		nextPageNum := pageNum + 1
		hasNextPage, err := df.navigateToNextPageAjax(ctx, nextPageNum)
		df.timingReport.PaginationTime += time.Since(paginationStart)
		df.ajaxCallCount++

		if err != nil {
			return fmt.Errorf("failed to navigate to page %d: %w", nextPageNum, err)
		}

		if hasNextPage {
			df.logger.Info("Navigating to page %d using AJAX...", nextPageNum)
			// Wait for AJAX pagination to complete using event-driven detection
			ajaxWaitStart := time.Now()
			if err := df.waitForDataTablePopulated(ctx, 8*time.Second); err != nil {
				df.logger.Error("AJAX pagination timeout, proceeding anyway: %v", err)
			}
			df.totalAjaxWaitTime += time.Since(ajaxWaitStart)
		} else {
			df.logger.Info("No more pages available")
		}

		pageNum++
		// No artificial delay needed - event-driven waiting handles timing
	}

	df.logger.Info("Finished extracting data. Total records: %d", len(*stockData))

	// Remove any duplicates that might have been added
	deduplicationStart := time.Now()
	*stockData = df.removeDuplicates(*stockData)
	df.timingReport.DeduplicationTime = time.Since(deduplicationStart)
	df.logger.Info("After deduplication: %d unique records", len(*stockData))

	// Finalize timing report
	df.finalizeTimingReport()

	// Keep temp file for testing - DO NOT rename
	tempFilename := fmt.Sprintf("raw_%s_temp.csv", ticker)
	finalFilename := fmt.Sprintf("raw_%s.csv", ticker)

	if _, err := os.Stat(tempFilename); err == nil {
		df.logger.Info("Keeping temp file %s for testing (not renaming to %s)", tempFilename, finalFilename)
		// Commented out for testing:
		// if err := os.Rename(tempFilename, finalFilename); err != nil {
		//     df.logger.Error("Failed to rename temp file to final name: %v", err)
		// } else {
		//     df.logger.Info("Successfully renamed %s to %s", tempFilename, finalFilename)
		// }
	}

	return nil
}

// extractDataFromCurrentPage extracts data from the current page
func (df *DataFetcher) extractDataFromCurrentPage(ctx context.Context) ([]StockData, error) {
	var rows []map[string]string

	// Wait for data table to be populated before extracting
	err := df.waitForDataTablePopulated(ctx, 5*time.Second) // Max 5 seconds wait
	if err != nil {
		df.logger.Error("Data table not populated, proceeding anyway: %v", err)
	}

	err = chromedp.Run(ctx,
		chromedp.Evaluate(`
			(function() {
				let dataRows = [];
				
				// Optimized direct element query - target specific table immediately
				const dispTable = document.getElementById('dispTable');
				if (!dispTable) return dataRows;
				
				const tbody = dispTable.querySelector('tbody');
				if (!tbody) return dataRows;
				
				const rows = tbody.rows; // Use native HTMLCollection for better performance
				
				for (let i = 0; i < rows.length; i++) {
					const cells = rows[i].cells;
					if (cells.length >= 10) {
						// Fast cell access using native HTMLCollection
						// Index mapping: 9:Date, 8:Close, 7:Open, 6:High, 5:Low, 4:Change, 3:Change%, 2:T.Shares, 1:Volume, 0:No.Trades
						dataRows.push({
							date: cells[9].textContent.trim(),
							close: cells[8].textContent.trim(),
							open: cells[7].textContent.trim(),
							high: cells[6].textContent.trim(),
							low: cells[5].textContent.trim(),
							change: cells[4].textContent.trim(),
							changePercent: cells[3].textContent.trim(),
							shares: cells[2].textContent.trim(),
							volume: cells[1].textContent.trim(),
							trades: cells[0].textContent.trim()
						});
					}
				}
				
				return dataRows;
			})();
		`, &rows),
	)

	if err != nil {
		return nil, err
	}

	df.logger.Info("Found %d data rows on current page", len(rows))

	var stockData []StockData
	for _, row := range rows {
		if row["date"] == "" {
			continue
		}

		data, err := df.parseRowData(row)
		if err != nil {
			df.logger.Error("Failed to parse row data: %v", err)
			continue
		}

		stockData = append(stockData, data)
	}

	return stockData, nil
}

// extractDataDirectly tries to extract data from any visible tables on the page
func (df *DataFetcher) extractDataDirectly(ctx context.Context, stockData *[]StockData) error {
	var rows []map[string]string

	err := chromedp.Run(ctx,
		chromedp.Evaluate(`
			(function() {
				// Look for any tables on the page that might contain stock data
				let allTables = document.querySelectorAll('table');
				let dataRows = [];
				
				for (let table of allTables) {
					let rows = table.querySelectorAll('tr');
					for (let row of rows) {
						let cells = Array.from(row.querySelectorAll('td, th'));
						if (cells.length >= 5) {
							// Check if this looks like a data row (has date-like content)
							let firstCell = cells[0]?.textContent?.trim() || '';
							if (firstCell.match(/\d{1,2}\/\d{1,2}\/\d{4}/) || firstCell.match(/\d{4}-\d{2}-\d{2}/)) {
								dataRows.push({
									date: cells[0]?.textContent?.trim() || '',
									close: cells[1]?.textContent?.trim() || '',
									open: cells[2]?.textContent?.trim() || '',
									high: cells[3]?.textContent?.trim() || '',
									low: cells[4]?.textContent?.trim() || '',
									change: cells[5]?.textContent?.trim() || '',
									changePercent: cells[6]?.textContent?.trim() || '',
									shares: cells[7]?.textContent?.trim() || '',
									volume: cells[8]?.textContent?.trim() || '',
									trades: cells[9]?.textContent?.trim() || ''
								});
							}
						}
					}
				}
				
				return dataRows;
			})();
		`, &rows),
	)

	if err != nil {
		return err
	}

	df.logger.Info("Found %d potential data rows using direct extraction", len(rows))

	for _, row := range rows {
		if row["date"] == "" {
			continue
		}

		data, err := df.parseRowData(row)
		if err != nil {
			df.logger.Error("Failed to parse row data: %v", err)
			continue
		}

		*stockData = append(*stockData, data)
	}

	return nil
}

// parseRowData parses a single row of data
func (df *DataFetcher) parseRowData(row map[string]string) (StockData, error) {
	// Parse date
	date, err := time.Parse("2/1/2006", row["date"])
	if err != nil {
		return StockData{}, fmt.Errorf("failed to parse date %s: %w", row["date"], err)
	}

	// Parse decimal values
	open, err := df.parseDecimal(row["open"])
	if err != nil {
		return StockData{}, fmt.Errorf("failed to parse open price: %w", err)
	}

	high, err := df.parseDecimal(row["high"])
	if err != nil {
		return StockData{}, fmt.Errorf("failed to parse high price: %w", err)
	}

	low, err := df.parseDecimal(row["low"])
	if err != nil {
		return StockData{}, fmt.Errorf("failed to parse low price: %w", err)
	}

	close, err := df.parseDecimal(row["close"])
	if err != nil {
		return StockData{}, fmt.Errorf("failed to parse close price: %w", err)
	}

	// Parse volume
	volume, err := df.parseInt(row["volume"])
	if err != nil {
		return StockData{}, fmt.Errorf("failed to parse volume: %w", err)
	}

	return StockData{
		Date:   date,
		Open:   open,
		High:   high,
		Low:    low,
		Close:  close,
		Volume: volume,
	}, nil
}

// parseDecimal parses a decimal value from string
func (df *DataFetcher) parseDecimal(value string) (decimal.Decimal, error) {
	// Remove commas and clean the string
	cleaned := strings.ReplaceAll(strings.TrimSpace(value), ",", "")
	if cleaned == "" || cleaned == "-" {
		return decimal.Zero, nil
	}

	return decimal.NewFromString(cleaned)
}

// parseInt parses an integer value from string
func (df *DataFetcher) parseInt(value string) (int64, error) {
	// Remove commas and clean the string
	cleaned := strings.ReplaceAll(strings.TrimSpace(value), ",", "")
	if cleaned == "" || cleaned == "-" {
		return 0, nil
	}

	return strconv.ParseInt(cleaned, 10, 64)
}

// navigateToNextPageAjax tries to navigate to the next page using AJAX
func (df *DataFetcher) navigateToNextPageAjax(ctx context.Context, pageNum int) (bool, error) {
	var hasNextPage bool

	err := chromedp.Run(ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			// Try to construct the AJAX call based on the pattern we found
			// Pattern: doAjax('companyperformancehistoryfilter.html','fromDate=05%%2F02%%2F2013&d-6716032-p=%d&toDate=05%%2F06%%2F2025&1749297467924=&companyCode=TASC','ajxDspId')
			
			// First, check if doAjax function exists
			if (typeof doAjax === 'function') {
				// Get the current date range from the form
				const fromDate = document.getElementById('fromDate') ? document.getElementById('fromDate').value : '05/02/2013';
				const toDate = document.getElementById('toDate') ? document.getElementById('toDate').value : '05/06/2025';
				const companyCode = document.getElementById('companyCode') ? document.getElementById('companyCode').value : 'TASC';
				
				// Encode the dates for URL
				const fromDateEncoded = encodeURIComponent(fromDate);
				const toDateEncoded = encodeURIComponent(toDate);
				
				// Construct the parameters
				const params = 'fromDate=' + fromDateEncoded + '&d-6716032-p=%d&toDate=' + toDateEncoded + '&companyCode=' + companyCode;
				
				// Call doAjax
				doAjax('companyperformancehistoryfilter.html', params, 'ajxDspId');
				true;
			} else {
				// Fallback: try to click the next page link
				const nextLink = document.querySelector('a[href*="doAjax"] img[src*="next.gif"]');
				if (nextLink && nextLink.parentElement) {
					nextLink.parentElement.click();
					true;
				} else {
					false;
				}
			}
		`, pageNum, pageNum), &hasNextPage),
	)

	if err != nil {
		return false, err
	}

	if hasNextPage {
		df.logger.Info("Successfully triggered AJAX navigation to page %d", pageNum)
		// No artificial delay needed - event-driven waiting will handle detection
	} else {
		df.logger.Info("No more pages available")
	}

	return hasNextPage, nil
}

// waitForNetworkIdle waits for network activity to cease
func (df *DataFetcher) waitForNetworkIdle(ctx context.Context, maxWaitTime time.Duration) error {
	start := time.Now()

	// Setup network monitoring
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`
			window.networkMonitor = {
				requestCount: 0,
				lastActivity: Date.now(),
				isIdle: false
			};
			
			// Monitor fetch requests
			const originalFetch = window.fetch;
			window.fetch = function() {
				window.networkMonitor.requestCount++;
				window.networkMonitor.lastActivity = Date.now();
				window.networkMonitor.isIdle = false;
				
				return originalFetch.apply(this, arguments).finally(() => {
					window.networkMonitor.requestCount--;
					if (window.networkMonitor.requestCount <= 0) {
						setTimeout(() => {
							if (Date.now() - window.networkMonitor.lastActivity > 500) {
								window.networkMonitor.isIdle = true;
							}
						}, 500);
					}
				});
			};
		`, nil),
	)

	if err != nil {
		return err
	}

	for time.Since(start) < maxWaitTime {
		var networkState map[string]interface{}

		err := chromedp.Run(ctx,
			chromedp.Evaluate(`window.networkMonitor`, &networkState),
		)

		if err == nil && networkState != nil {
			if getBoolFromMap(networkState, "isIdle") {
				df.logger.Info("Network idle achieved in %v", time.Since(start))
				return nil
			}
		}

		time.Sleep(50 * time.Millisecond) // Faster polling for network detection
	}

	return fmt.Errorf("timeout waiting for network idle after %v", maxWaitTime)
}

// loadExistingData loads existing data from CSV file
func (df *DataFetcher) loadExistingData(filename string) ([]StockData, error) {
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

	if len(records) < 2 {
		return []StockData{}, nil
	}

	var stockData []StockData
	// Skip header row
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 6 {
			continue
		}

		// Parse the existing CSV format: Date,Close,Open,High,Low,Change,Change%,T.Shares,Volume,No. Trades
		date, err := time.Parse("2006-01-02", record[0])
		if err != nil {
			df.logger.Error("Failed to parse date %s: %v", record[0], err)
			continue
		}

		close, err := df.parseDecimal(record[1])
		if err != nil {
			df.logger.Error("Failed to parse close price: %v", err)
			continue
		}

		open, err := df.parseDecimal(record[2])
		if err != nil {
			df.logger.Error("Failed to parse open price: %v", err)
			continue
		}

		high, err := df.parseDecimal(record[3])
		if err != nil {
			df.logger.Error("Failed to parse high price: %v", err)
			continue
		}

		low, err := df.parseDecimal(record[4])
		if err != nil {
			df.logger.Error("Failed to parse low price: %v", err)
			continue
		}

		volume := int64(0)
		if len(record) > 8 {
			volume, _ = df.parseInt(record[8])
		}

		stockData = append(stockData, StockData{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	return stockData, nil
}

// saveDataToCSV saves stock data to CSV file
func (df *DataFetcher) saveDataToCSV(stockData []StockData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header - matching the original ISX format
	header := []string{"Date", "Close", "Open", "High", "Low", "Change", "Change%", "T.Shares", "Volume", "No. Trades"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data - matching the original ISX format
	for _, data := range stockData {
		record := []string{
			data.Date.Format("2006-01-02"),
			data.Close.String(),
			data.Open.String(),
			data.High.String(),
			data.Low.String(),
			data.Change.String(),
			data.ChangePercent.StringFixed(2) + "%", // Format as percentage with 2 decimal places
			"",                                      // T.Shares - not available in our data
			strconv.FormatInt(data.Volume, 10),
			"", // No. Trades - not available in our data
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// removeDuplicates removes duplicate records based on date
func (df *DataFetcher) removeDuplicates(stockData []StockData) []StockData {
	seen := make(map[string]bool)
	var unique []StockData

	for _, data := range stockData {
		dateKey := data.Date.Format("2006-01-02")
		if !seen[dateKey] {
			seen[dateKey] = true
			unique = append(unique, data)
		}
	}

	return unique
}

// sortAndRecalculateChanges sorts data by date and recalculates change/change% values
func (df *DataFetcher) sortAndRecalculateChanges(stockData []StockData) []StockData {
	// Sort by date (oldest first)
	sort.Slice(stockData, func(i, j int) bool {
		return stockData[i].Date.Before(stockData[j].Date)
	})

	// Recalculate change and change percentage for each day
	for i := 0; i < len(stockData); i++ {
		if i == 0 {
			// First day has no previous day to compare
			stockData[i].Change = decimal.Zero
			stockData[i].ChangePercent = decimal.Zero
		} else {
			// Calculate change from previous day
			prevClose := stockData[i-1].Close
			currentClose := stockData[i].Close

			// Change = Current Close - Previous Close
			stockData[i].Change = currentClose.Sub(prevClose)

			// Change Percent = (Change / Previous Close) * 100
			if !prevClose.IsZero() {
				stockData[i].ChangePercent = stockData[i].Change.Div(prevClose).Mul(decimal.NewFromInt(100))
			} else {
				stockData[i].ChangePercent = decimal.Zero
			}
		}
	}

	return stockData
}

// finalizeReport completes the processing report with final statistics
func (df *DataFetcher) finalizeReport(err error) {
	if df.currentReport == nil {
		return
	}

	df.currentReport.EndTime = time.Now()
	df.currentReport.ProcessingDuration = time.Since(df.startTime).String()
	df.currentReport.PagesLoaded = df.pagesLoaded
	df.currentReport.NewRowsCount = df.newRowsCount

	// Determine status and recommendations
	if err != nil {
		df.currentReport.Status = "ERROR"
		df.currentReport.ErrorMessage = err.Error()
		df.currentReport.DataQualityScore = "FAILED"

		if df.pagesLoaded > 0 {
			df.currentReport.Status = "PARTIAL"
			df.currentReport.DataQualityScore = "POOR"
			df.currentReport.Recommendation = "Manually check ticker - partial data downloaded. Consider investigating connection issues."
		} else {
			df.currentReport.Recommendation = "Manually check ticker - no data downloaded. Consider removing from tickers file if consistently failing."
		}
	} else {
		df.currentReport.Status = "SUCCESS"

		// Determine data quality score
		if df.currentReport.TotalRowsInCSV > 200 {
			df.currentReport.DataQualityScore = "EXCELLENT"
			df.currentReport.Recommendation = "No action needed - excellent data coverage"
		} else if df.currentReport.TotalRowsInCSV > 50 {
			df.currentReport.DataQualityScore = "GOOD"
			df.currentReport.Recommendation = "Good data coverage - monitor for updates"
		} else {
			df.currentReport.DataQualityScore = "POOR"
			df.currentReport.Recommendation = "Limited data available - manually verify ticker is actively traded"
		}
	}

	// Get file statistics
	filename := fmt.Sprintf("raw_%s.csv", df.currentReport.Ticker)
	if stat, statErr := os.Stat(filename); statErr == nil {
		df.currentReport.FileSize = stat.Size()

		// Get date range from file
		if data, loadErr := df.loadExistingData(filename); loadErr == nil && len(data) > 0 {
			df.currentReport.FirstDataDate = data[0].Date.Format("2006-01-02")
			df.currentReport.LastDataDate = data[len(data)-1].Date.Format("2006-01-02")
			df.currentReport.DaysLoaded = len(data)
			df.currentReport.TotalRowsInCSV = len(data)
		}
	}
}

// SaveProcessingReport saves a processing report to CSV
func SaveProcessingReport(reports []ProcessingReport, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Ticker", "Sector", "Company_Name", "Status", "Start_Time", "End_Time",
		"Processing_Duration", "Pages_Before_Update", "Pages_Loaded", "Days_Loaded",
		"New_Rows_Count", "Total_Rows_In_CSV", "Error_Message", "Recommendation",
		"File_Size_Bytes", "Last_Data_Date", "First_Data_Date", "Data_Quality_Score",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, report := range reports {
		record := []string{
			report.Ticker,
			report.Sector,
			report.CompanyName,
			report.Status,
			report.StartTime.Format("2006-01-02 15:04:05"),
			report.EndTime.Format("2006-01-02 15:04:05"),
			report.ProcessingDuration,
			strconv.Itoa(report.PagesBeforeUpdate),
			strconv.Itoa(report.PagesLoaded),
			strconv.Itoa(report.DaysLoaded),
			strconv.Itoa(report.NewRowsCount),
			strconv.Itoa(report.TotalRowsInCSV),
			report.ErrorMessage,
			report.Recommendation,
			strconv.FormatInt(report.FileSize, 10),
			report.LastDataDate,
			report.FirstDataDate,
			report.DataQualityScore,
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// finalizeTimingReport completes the timing analysis with performance metrics
func (df *DataFetcher) finalizeTimingReport() {
	if df.timingReport == nil {
		return
	}

	// Calculate total processing time
	df.timingReport.TotalProcessingTime = time.Since(df.startTime)
	df.timingReport.PagesProcessed = df.pagesLoaded
	df.timingReport.TotalAjaxCalls = df.ajaxCallCount
	df.timingReport.AjaxWaitTime = df.totalAjaxWaitTime

	// Calculate average page time
	if len(df.pageDurations) > 0 {
		var totalPageTime time.Duration
		for _, duration := range df.pageDurations {
			totalPageTime += duration
		}
		df.timingReport.AveragePageTime = totalPageTime / time.Duration(len(df.pageDurations))
	}

	// Calculate browser overhead (time not spent in measured operations)
	measuredTime := df.timingReport.NavigationTime +
		df.timingReport.CompanyCodeSetTime +
		df.timingReport.AjaxTriggerTime +
		df.timingReport.DataExtractionTime +
		df.timingReport.PaginationTime +
		df.timingReport.SortingTime +
		df.timingReport.CSVSaveTime +
		df.timingReport.DeduplicationTime +
		df.timingReport.AjaxWaitTime

	df.timingReport.BrowserOverheadTime = df.timingReport.TotalProcessingTime - measuredTime

	// Determine bottleneck function
	timings := map[string]time.Duration{
		"Navigation":       df.timingReport.NavigationTime,
		"Company Code Set": df.timingReport.CompanyCodeSetTime,
		"AJAX Trigger":     df.timingReport.AjaxTriggerTime,
		"Data Extraction":  df.timingReport.DataExtractionTime,
		"Pagination":       df.timingReport.PaginationTime,
		"Sorting":          df.timingReport.SortingTime,
		"CSV Save":         df.timingReport.CSVSaveTime,
		"Deduplication":    df.timingReport.DeduplicationTime,
		"AJAX Wait":        df.timingReport.AjaxWaitTime,
		"Browser Overhead": df.timingReport.BrowserOverheadTime,
	}

	var bottleneck string
	var maxTime time.Duration
	for function, duration := range timings {
		if duration > maxTime {
			maxTime = duration
			bottleneck = function
		}
	}
	df.timingReport.BottleneckFunction = bottleneck

	// Determine performance score and optimization suggestions
	totalSeconds := df.timingReport.TotalProcessingTime.Seconds()
	avgPageSeconds := df.timingReport.AveragePageTime.Seconds()

	if totalSeconds < 30 && avgPageSeconds < 3 {
		df.timingReport.PerformanceScore = "EXCELLENT"
		df.timingReport.OptimizationSuggestion = "Performance is excellent - no optimization needed"
	} else if totalSeconds < 60 && avgPageSeconds < 5 {
		df.timingReport.PerformanceScore = "GOOD"
		df.timingReport.OptimizationSuggestion = "Good performance - minor optimizations possible"
	} else if totalSeconds < 120 && avgPageSeconds < 8 {
		df.timingReport.PerformanceScore = "AVERAGE"
		df.timingReport.OptimizationSuggestion = fmt.Sprintf("Average performance - focus on optimizing %s", bottleneck)
	} else {
		df.timingReport.PerformanceScore = "POOR"
		df.timingReport.OptimizationSuggestion = fmt.Sprintf("Poor performance - urgent optimization needed for %s", bottleneck)
	}

	// Add specific optimization suggestions based on bottleneck
	switch bottleneck {
	case "AJAX Wait":
		df.timingReport.OptimizationSuggestion += " - Reduce AJAX wait times or implement smarter waiting"
	case "Pagination":
		df.timingReport.OptimizationSuggestion += " - Optimize pagination logic or reduce page navigation overhead"
	case "Data Extraction":
		df.timingReport.OptimizationSuggestion += " - Optimize data parsing or reduce DOM queries"
	case "Browser Overhead":
		df.timingReport.OptimizationSuggestion += " - Consider headless mode or reduce browser operations"
	case "CSV Save":
		df.timingReport.OptimizationSuggestion += " - Optimize file I/O or reduce save frequency"
	}
}

// SaveTimingReport saves timing analysis reports to CSV
func SaveTimingReport(reports []TimingReport, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Ticker", "Total_Processing_Time_Ms", "Navigation_Time_Ms", "Page_Load_Time_Ms",
		"Company_Code_Set_Time_Ms", "Ajax_Trigger_Time_Ms", "Data_Extraction_Time_Ms",
		"Pagination_Time_Ms", "Data_Parsing_Time_Ms", "Sorting_Time_Ms", "Change_Calculation_Time_Ms",
		"CSV_Save_Time_Ms", "Deduplication_Time_Ms", "File_Operations_Time_Ms",
		"Average_Page_Time_Ms", "Slowest_Page_Time_Ms", "Fastest_Page_Time_Ms",
		"Ajax_Wait_Time_Ms", "Browser_Overhead_Time_Ms", "Pages_Processed", "Total_Ajax_Calls",
		"Performance_Score", "Bottleneck_Function", "Optimization_Suggestion",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, report := range reports {
		record := []string{
			report.Ticker,
			fmt.Sprintf("%.2f", report.TotalProcessingTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.NavigationTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.PageLoadTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.CompanyCodeSetTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.AjaxTriggerTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.DataExtractionTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.PaginationTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.DataParsingTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.SortingTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.ChangeCalculationTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.CSVSaveTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.DeduplicationTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.FileOperationsTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.AveragePageTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.SlowestPageTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.FastestPageTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.AjaxWaitTime.Seconds()*1000),
			fmt.Sprintf("%.2f", report.BrowserOverheadTime.Seconds()*1000),
			strconv.Itoa(report.PagesProcessed),
			strconv.Itoa(report.TotalAjaxCalls),
			report.PerformanceScore,
			report.BottleneckFunction,
			report.OptimizationSuggestion,
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// GetTimingReport returns the current timing report
func (df *DataFetcher) GetTimingReport() *TimingReport {
	return df.timingReport
}

// waitForPageComplete waits for all page operations to complete using multiple detection methods
func (df *DataFetcher) waitForPageComplete(ctx context.Context, maxWaitTime time.Duration) error {
	start := time.Now()

	// Setup performance monitoring
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`
			// Only initialize if not already done
			if (!window.pageReadyState) {
				window.pageReadyState = {
					networkIdle: false,
					domReady: false,
					ajaxComplete: false,
					dataTableReady: false,
					startTime: Date.now()
				};
				
				// Monitor network activity
				window.pendingRequests = 0;
				window.networkIdleTimer = null;
			
				// Override XMLHttpRequest to track AJAX
				const originalXHR = window.XMLHttpRequest;
				window.XMLHttpRequest = function() {
					const xhr = new originalXHR();
					window.pendingRequests++;
					window.pageReadyState.ajaxComplete = false;
					console.log('AJAX request started, pending:', window.pendingRequests);
					
					const updateNetworkState = () => {
						window.pendingRequests--;
						console.log('AJAX request finished, pending:', window.pendingRequests);
						if (window.pendingRequests <= 0) {
							clearTimeout(window.networkIdleTimer);
							window.networkIdleTimer = setTimeout(() => {
								window.pageReadyState.networkIdle = true;
								window.pageReadyState.ajaxComplete = true;
								console.log('Network idle achieved');
							}, 200); // 200ms of network silence
						}
					};
					
					xhr.addEventListener('load', updateNetworkState);
					xhr.addEventListener('error', updateNetworkState);
					xhr.addEventListener('abort', updateNetworkState);
					
					return xhr;
				};
			
			// Monitor DOM changes
			const observer = new MutationObserver((mutations) => {
				const hasDataTable = document.querySelector('#ajxDspId table#dispTable tbody tr');
				if (hasDataTable) {
					window.pageReadyState.dataTableReady = true;
					console.log('Data table detected');
				}
			});
			
			observer.observe(document.body, {
				childList: true,
				subtree: true,
				attributes: false
			});
			
			// Check document ready state
			const checkDOMReady = () => {
				if (document.readyState === 'complete') {
					window.pageReadyState.domReady = true;
					console.log('DOM ready');
				}
			};
			
			document.addEventListener('DOMContentLoaded', checkDOMReady);
			if (document.readyState !== 'loading') {
				checkDOMReady();
			}
			
				// Set initial states if already ready
				if (document.readyState === 'complete') {
					window.pageReadyState.domReady = true;
				}
				
				// Set network idle initially if no active requests
				if (window.pendingRequests === 0) {
					setTimeout(() => {
						window.pageReadyState.networkIdle = true;
						window.pageReadyState.ajaxComplete = true;
					}, 500);
				}
			}
		`, nil),
	)

	if err != nil {
		return err
	}

	// Poll for completion with intelligent checks
	lastLogTime := time.Now()
	for time.Since(start) < maxWaitTime {
		var pageState map[string]interface{}

		err := chromedp.Run(ctx,
			chromedp.Evaluate(`window.pageReadyState`, &pageState),
		)

		if err == nil && pageState != nil {
			networkIdle := getBoolFromMap(pageState, "networkIdle")
			domReady := getBoolFromMap(pageState, "domReady")
			ajaxComplete := getBoolFromMap(pageState, "ajaxComplete")
			dataTableReady := getBoolFromMap(pageState, "dataTableReady")

			// Log progress every 2 seconds for debugging
			if time.Since(lastLogTime) >= 2*time.Second {
				df.logger.Info("Page completion status: network=%v dom=%v ajax=%v table=%v (elapsed: %v)",
					networkIdle, domReady, ajaxComplete, dataTableReady, time.Since(start))
				lastLogTime = time.Now()
			}

			// For initial page load, we only need DOM ready and network idle
			// Data table is not expected on the initial page
			if networkIdle && domReady {
				df.logger.Info("Initial page load completed in %v", time.Since(start))
				return nil
			}
		}

		time.Sleep(50 * time.Millisecond) // Ultra-fast polling
	}

	df.logger.Error("Timeout waiting for page completion after %v, proceeding anyway", maxWaitTime)
	return nil // Don't fail hard, just proceed
}

// Helper function to safely get boolean from map
func getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return false
}

// waitForAjaxOperation waits for specific AJAX operation to complete
func (df *DataFetcher) waitForAjaxOperation(ctx context.Context, operationName string, maxWaitTime time.Duration) error {
	start := time.Now()

	// Setup AJAX operation monitoring
	err := chromedp.Run(ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			window.ajaxOperation_%s = {
				started: false,
				completed: false,
				startTime: Date.now()
			};
			
			// Monitor for the specific AJAX operation
			const originalSend = XMLHttpRequest.prototype.send;
			XMLHttpRequest.prototype.send = function(data) {
				console.log('XMLHttpRequest send called for:', this.responseURL);
				if (this.responseURL && this.responseURL.includes('%s')) {
					window.ajaxOperation_%s.started = true;
					console.log('Target AJAX operation started:', '%s');
					
					this.addEventListener('load', () => {
						window.ajaxOperation_%s.completed = true;
						console.log('Target AJAX operation completed:', '%s');
					});
					this.addEventListener('error', () => {
						window.ajaxOperation_%s.completed = true;
						console.log('Target AJAX operation error:', '%s');
					});
				}
				
				return originalSend.call(this, data);
			};
		`, operationName, operationName, operationName, operationName, operationName, operationName, operationName, operationName), nil),
	)

	if err != nil {
		df.logger.Error("Failed to setup AJAX monitoring: %v", err)
		return nil // Don't fail hard
	}

	// Poll for AJAX completion
	lastLogTime := time.Now()
	for time.Since(start) < maxWaitTime {
		var operationState map[string]interface{}

		err := chromedp.Run(ctx,
			chromedp.Evaluate(fmt.Sprintf(`window.ajaxOperation_%s`, operationName), &operationState),
		)

		if err == nil && operationState != nil {
			started := getBoolFromMap(operationState, "started")
			completed := getBoolFromMap(operationState, "completed")

			// Log progress every 2 seconds
			if time.Since(lastLogTime) >= 2*time.Second {
				df.logger.Info("AJAX operation '%s': started=%v completed=%v (elapsed: %v)",
					operationName, started, completed, time.Since(start))
				lastLogTime = time.Now()
			}

			if completed {
				df.logger.Info("AJAX operation '%s' completed in %v", operationName, time.Since(start))
				return nil
			}

			// If operation hasn't started after 3 seconds, assume it's not going to happen
			if time.Since(start) > 3*time.Second && !started {
				df.logger.Info("AJAX operation '%s' didn't start, assuming no AJAX needed", operationName)
				return nil
			}
		}

		time.Sleep(25 * time.Millisecond) // Ultra-fast polling for AJAX
	}

	df.logger.Error("Timeout waiting for AJAX operation '%s' after %v, proceeding anyway", operationName, maxWaitTime)
	return nil // Don't fail hard
}

// waitForDataTablePopulated waits for data table to be populated with actual data
func (df *DataFetcher) waitForDataTablePopulated(ctx context.Context, maxWaitTime time.Duration) error {
	start := time.Now()
	previousRowCount := 0
	stableCount := 0

	df.logger.Info("Waiting for data table to be populated...")

	for time.Since(start) < maxWaitTime {
		var currentRowCount int

		err := chromedp.Run(ctx,
			chromedp.Evaluate(`
				(function() {
					// Try multiple selectors for data table
					let table = document.querySelector('#ajxDspId table#dispTable tbody');
					if (!table) {
						table = document.querySelector('#dispTable tbody');
					}
					if (!table) {
						table = document.querySelector('table tbody');
					}
					
					if (table) {
						const rows = table.querySelectorAll('tr');
						console.log('Found table with', rows.length, 'rows');
						return rows.length;
					}
					
					console.log('No data table found');
					return 0;
				})();
			`, &currentRowCount),
		)

		if err == nil {
			// Log progress every 2 seconds
			if int(time.Since(start).Seconds())%2 == 0 && time.Since(start) >= 2*time.Second {
				df.logger.Info("Data table check: %d rows found (elapsed: %v)", currentRowCount, time.Since(start))
			}

			if currentRowCount > 0 {
				// Check if row count is stable (not changing)
				if currentRowCount == previousRowCount {
					stableCount++
					if stableCount >= 3 { // Stable for 3 checks (300ms)
						df.logger.Info("Data table populated with %d rows in %v", currentRowCount, time.Since(start))
						return nil
					}
				} else {
					stableCount = 0 // Reset stability counter
					previousRowCount = currentRowCount
					df.logger.Info("Data table rows changed to %d, waiting for stability", currentRowCount)
				}
			}
		} else {
			df.logger.Error("Error checking data table: %v", err)
		}

		time.Sleep(50 * time.Millisecond) // Faster polling for table detection
	}

	df.logger.Error("Timeout waiting for data table population after %v, proceeding anyway", maxWaitTime)
	return nil // Don't fail hard, let the extraction attempt to proceed
}

// handleInitialPopups handles any initial popups that might appear (like year validation popup)
func (df *DataFetcher) handleInitialPopups(ctx context.Context) error {
	df.logger.Info("Checking for initial popups...")

	// Check immediately first (some popups appear instantly with invalid dates)
	if df.tryDismissPopup(ctx) {
		return nil
	}

	// Wait a moment for any delayed popups to appear, then check again
	time.Sleep(1 * time.Second)
	if df.tryDismissPopup(ctx) {
		return nil
	}

	df.logger.Info("No popups detected after initial checks")
	return nil
}

// tryDismissPopup attempts to detect and dismiss any popup, returns true if successful
func (df *DataFetcher) tryDismissPopup(ctx context.Context) bool {
	df.logger.Info("Starting synchronous popup detection and dismissal...")

	// Channel to track dialog events
	dialogDetected := make(chan bool, 1)

	// Set up a listener for JavaScript dialog events (like browser alerts)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *page.EventJavascriptDialogOpening:
			df.logger.Info("JavaScript dialog detected: Type=%s, Message='%s'", ev.Type, ev.Message)

			// Handle the dialog immediately and synchronously
			err := chromedp.Run(ctx, page.HandleJavaScriptDialog(true))
			if err != nil {
				df.logger.Error("Failed to handle JavaScript dialog: %v", err)
			} else {
				df.logger.Info("Successfully accepted JavaScript dialog")
				select {
				case dialogDetected <- true:
				default:
				}
			}
		}
	})

	// Immediate check for popup conditions using synchronous JavaScript execution
	df.logger.Info("Checking for popup indicators in page content...")

	var popupResultStr string

	err := chromedp.Run(ctx,
		chromedp.Evaluate(`
			(function() {
				var bodyText = document.body ? (document.body.innerText || document.body.textContent || '') : '';
				var result = {
					hasPopup: false,
					popupType: '',
					popupText: '',
					bodyPreview: bodyText.substring(0, 200)
				};
				
				// Check for specific popup error messages
				if (bodyText.includes('Please enter a valid 4 digit year between 1900 and 2100')) {
					result.hasPopup = true;
					result.popupType = 'year_validation';
					result.popupText = 'Year validation error detected';
				} else if (bodyText.includes('invalid year') || bodyText.includes('Please enter a valid')) {
					result.hasPopup = true;
					result.popupType = 'validation_error';
					result.popupText = 'General validation error detected';
				}
				
				// Check for visible modal dialogs
				var modals = document.querySelectorAll('[role="dialog"], .modal, .ui-dialog, .popup, .alert');
				for (var i = 0; i < modals.length; i++) {
					var style = window.getComputedStyle(modals[i]);
					if (style.display !== 'none' && style.visibility !== 'hidden') {
						result.hasPopup = true;
						result.popupType = 'modal_dialog';
						result.popupText = 'Visible modal dialog found';
						break;
					}
				}
				
				return JSON.stringify(result);
			})()
		`, &popupResultStr),
	)

	var popupResult struct {
		HasPopup    bool   `json:"hasPopup"`
		PopupType   string `json:"popupType"`
		PopupText   string `json:"popupText"`
		BodyPreview string `json:"bodyPreview"`
	}

	if err != nil {
		df.logger.Error("Failed to check for popup indicators: %v", err)
	} else {
		// Parse the JSON result
		if parseErr := json.Unmarshal([]byte(popupResultStr), &popupResult); parseErr != nil {
			df.logger.Error("Failed to parse popup result: %v", parseErr)
			df.logger.Info("Raw result: %s", popupResultStr)
		} else {
			df.logger.Info("Popup check results: HasPopup=%v, Type=%s, Text=%s",
				popupResult.HasPopup, popupResult.PopupType, popupResult.PopupText)
			df.logger.Info("Page body preview: %s", popupResult.BodyPreview)
		}
	}

	// If popup detected, attempt to dismiss it
	if popupResult.HasPopup {
		df.logger.Info("Popup detected, attempting immediate dismissal...")

		// Strategy 1: Try keyboard shortcuts (most reliable)
		shortcuts := []string{"Enter", "Escape"}
		for _, key := range shortcuts {
			df.logger.Info("Trying %s key...", key)
			keyErr := chromedp.Run(ctx, chromedp.KeyEvent(key))
			if keyErr == nil {
				df.logger.Info("Successfully sent %s key", key)
				time.Sleep(500 * time.Millisecond) // Brief wait for response

				// Check if popup is gone
				var stillHasPopup bool
				checkErr := chromedp.Run(ctx,
					chromedp.Evaluate(`
						(function() {
							var bodyText = document.body ? (document.body.innerText || document.body.textContent || '') : '';
							return bodyText.includes('Please enter a valid') || bodyText.includes('invalid year');
						})()
					`, &stillHasPopup),
				)
				if checkErr == nil && !stillHasPopup {
					df.logger.Info("Popup successfully dismissed with %s key", key)
					return true
				}
			}
		}

		// Strategy 2: Try clicking OK buttons
		df.logger.Info("Keyboard shortcuts failed, trying to click OK buttons...")
		var buttonClicked bool
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`
				(function() {
					var buttons = document.querySelectorAll('button, input[type="button"], input[type="submit"]');
					for (var i = 0; i < buttons.length; i++) {
						var btn = buttons[i];
						var text = (btn.textContent || btn.value || '').toLowerCase().trim();
						var style = window.getComputedStyle(btn);
						
						if ((text === 'ok' || text === 'close' || text === 'dismiss' || text === 'continue') &&
							style.display !== 'none' && style.visibility !== 'hidden') {
							btn.click();
							return true;
						}
					}
					
					// If no specific button found, try the first visible button
					for (var i = 0; i < buttons.length; i++) {
						var btn = buttons[i];
						var style = window.getComputedStyle(btn);
						if (style.display !== 'none' && style.visibility !== 'hidden') {
							btn.click();
							return true;
						}
					}
					
					return false;
				})()
			`, &buttonClicked),
		)

		if err == nil && buttonClicked {
			df.logger.Info("Successfully clicked a button to dismiss popup")
			time.Sleep(500 * time.Millisecond)
			return true
		}
	}

	// Check if a JavaScript dialog was detected during our execution
	select {
	case <-dialogDetected:
		df.logger.Info("JavaScript dialog was handled during execution")
		return true
	default:
		// No dialog detected
	}

	// Wait briefly for any delayed dialogs
	df.logger.Info("Waiting briefly for any delayed dialogs...")
	select {
	case <-dialogDetected:
		df.logger.Info("Delayed JavaScript dialog was handled")
		return true
	case <-time.After(1 * time.Second):
		// Timeout - no dialog appeared
	}

	if popupResult.HasPopup {
		df.logger.Error("Popup was detected but could not be dismissed")
		return false
	}

	df.logger.Info("No popup detected")
	return false
}

// setDateAndSearch sets the fromDate field to 01/01/2010 and clicks the Search button
func (df *DataFetcher) setDateAndSearch(ctx context.Context, ticker string) error {
	df.logger.Info("Setting company code to %s, fromDate to 01/01/2010 and triggering search", ticker)

	// Set the company code, fromDate field and click Search button
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			// First, set the company code in the hidden field
			err := chromedp.Run(ctx,
				chromedp.SetValue(`#companyCode`, ticker, chromedp.ByID),
			)
			if err != nil {
				df.logger.Error("Failed to set company code field: %v", err)
				return err
			}

			df.logger.Info("Successfully set company code to %s", ticker)

			// Set the fromDate field to 1/1/2010 exactly like Python version
			df.logger.Info("Setting fromDate field to 1/1/2010 (matching Python implementation)...")

			// First, wait for the fromDate input field to be present (like Python does)
			err = chromedp.Run(ctx,
				chromedp.WaitVisible(`#fromDate`, chromedp.ByID),
			)
			if err != nil {
				df.logger.Error("Failed to wait for fromDate field: %v", err)
				return err
			}

			// Use JavaScript to set the value exactly like Python: document.querySelector("#fromDate").value = "1/1/2010"
			// Clear field first, then set the value to avoid validation issues
			err = chromedp.Run(ctx,
				chromedp.Evaluate(`
					var fromDateField = document.querySelector("#fromDate");
					if (fromDateField) {
						fromDateField.value = "";
						fromDateField.focus();
						fromDateField.value = "1/1/2010";
						fromDateField.blur();
					}
				`, nil),
			)
			if err != nil {
				df.logger.Error("Failed to set fromDate value with JavaScript: %v", err)
				return err
			}

			// Small delay to allow any validation popups to appear
			time.Sleep(500 * time.Millisecond)

			// Verify the value was set correctly (like Python assertion)
			var actualValue string
			err = chromedp.Run(ctx,
				chromedp.Evaluate(`document.querySelector("#fromDate").value`, &actualValue),
			)
			if err != nil {
				df.logger.Error("Failed to verify fromDate value: %v", err)
				return err
			}

			if actualValue != "1/1/2010" {
				df.logger.Error("Failed to set the fromDate input value. Expected '1/1/2010', got '%s'", actualValue)
				return fmt.Errorf("failed to set the fromDate input value")
			}

			df.logger.Info("Successfully set and verified fromDate to 1/1/2010")

			// Find and click the search button with id="button"
			df.logger.Info("Finding and clicking the search button with id='button'...")

			// First try to click the correct search button that calls submitForm()
			var clickSuccess bool
			err = chromedp.Run(ctx,
				chromedp.Evaluate(`
					(function() {
						// Find the search button that calls submitForm() (not doSearch())
						var buttons = document.querySelectorAll('input[type="button"][value="Search"]');
						for (var i = 0; i < buttons.length; i++) {
							var onclick = buttons[i].getAttribute('onclick');
							if (onclick && onclick.includes('submitForm()')) {
								buttons[i].click();
								return true;
							}
						}
						return false;
					})()
				`, &clickSuccess),
			)

			if err != nil || !clickSuccess {
				df.logger.Info("submitForm() button click failed, trying alternative approaches...")

				// Try calling submitForm() directly since that's what the onclick does
				err = chromedp.Run(ctx,
					chromedp.Evaluate(`
						(function() {
							if (typeof submitForm === 'function') {
								submitForm();
								return true;
							}
							return false;
						})()
					`, &clickSuccess),
				)
				if err == nil && clickSuccess {
					df.logger.Info("Successfully called submitForm() directly")
				} else {
					// Try finding button by name="Search"
					err = chromedp.Run(ctx,
						chromedp.Evaluate(`
							(function() {
								var searchButton = document.querySelector('input[name="Search"][onclick*="submitForm"]');
								if (searchButton) {
									searchButton.click();
									return true;
								}
								return false;
							})()
						`, &clickSuccess),
					)
					if err == nil && clickSuccess {
						df.logger.Info("Successfully clicked search button by name and onclick")
					} else {
						// Last resort: try to find any button with submitForm in onclick
						err = chromedp.Run(ctx,
							chromedp.Evaluate(`
								(function() {
									var allButtons = document.querySelectorAll('input[type="button"]');
									for (var i = 0; i < allButtons.length; i++) {
										var onclick = allButtons[i].getAttribute('onclick');
										if (onclick && onclick.includes('submitForm()')) {
											allButtons[i].click();
											return true;
										}
									}
									return false;
								})()
							`, &clickSuccess),
						)
						if err == nil && clickSuccess {
							df.logger.Info("Successfully clicked button with submitForm onclick")
						} else {
							df.logger.Error("Failed to click search button with any method")
							return fmt.Errorf("failed to click search button")
						}
					}
				}
			} else {
				df.logger.Info("Successfully clicked search button that calls submitForm()")
			}
			return nil
		}),
	)

	if err != nil {
		df.logger.Error("Error in setDateAndSearch: %v", err)
		return err
	}

	// Wait for the search results to load - specifically wait for the data table like Python does
	df.logger.Info("Waiting for data table to load after search (like Python implementation)...")
	err = chromedp.Run(ctx,
		chromedp.WaitVisible(`#dispTable`, chromedp.ByID),
	)
	if err != nil {
		df.logger.Error("Failed to wait for data table: %v", err)
		// Don't fail hard, continue with page completion wait
	} else {
		df.logger.Info("Data table is now visible")
	}

	// Also wait for page completion as backup
	err = df.waitForPageComplete(ctx, 5*time.Second)
	if err != nil {
		df.logger.Error("Failed to wait for page completion: %v", err)
		// Don't fail hard, data table wait might be sufficient
	}

	df.logger.Info("Search completed successfully, returning to normal process")
	return nil
}
