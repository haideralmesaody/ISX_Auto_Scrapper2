package common

import (
	"os"
	"path/filepath"
)

// Config holds all application configuration
type Config struct {
	// Logging Configuration
	Debug       bool
	LogFilename string
	LogLevel    string
	LogFormat   string

	// Directory Configuration
	CurrentDirectory string

	// Edge Driver Configuration
	EdgeDriverPath string

	// URL Configuration
	BaseURL     string
	BaseURLASE  string
	DefaultDate string

	// Table Configuration
	TableSelector string

	// WebDriver Wait Configuration
	WebDriverWaitTime int

	// Data Fetching Configuration
	DefaultSMAPeriod int
	DefaultRowCount  int

	// Excel Configuration
	ExcelEngine string
}

// NewConfig creates a new configuration instance with default values
func NewConfig() *Config {
	// Get current directory
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		currentDir = "."
	}

	// Read Edge Driver path from environment variable, fall back to default
	edgeDriverPath := os.Getenv("EDGE_DRIVER_PATH")
	if edgeDriverPath == "" {
		edgeDriverPath = filepath.Join(currentDir, "msedgedriver.exe")
	}

	return &Config{
		// Logging Configuration
		Debug:       true,
		LogFilename: "stock_analysis.log",
		LogLevel:    "DEBUG",        // or "ERROR" based on Debug flag
		LogFormat:   "%s - %s - %s", // timestamp - level - message

		// Directory Configuration
		CurrentDirectory: currentDir,

		// Edge Driver Configuration
		EdgeDriverPath: edgeDriverPath,

		// URL Configuration
		BaseURL:     "http://www.isx-iq.net/isxportal/portal/companyprofilecontainer.html",
		BaseURLASE:  "https://www.ase.com.jo/en/company_historical/",
		DefaultDate: "06/10/2010",

		// Table Configuration
		TableSelector: "#dispTable",

		// WebDriver Wait Configuration
		WebDriverWaitTime: 10,

		// Data Fetching Configuration
		DefaultSMAPeriod: 10,
		DefaultRowCount:  600,

		// Excel Configuration
		ExcelEngine: "openpyxl",
	}
}

// GetLogLevel returns the appropriate log level based on debug flag
func (c *Config) GetLogLevel() string {
	if c.Debug {
		return "DEBUG"
	}
	return "ERROR"
}

// Global configuration instance
var AppConfig = NewConfig()
