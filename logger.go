package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger provides logging functionality similar to the Python version
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	file        *os.File
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	// Create or open log file using config
	file, err := os.OpenFile(AppConfig.LogFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file:", err)
	}

	return &Logger{
		infoLogger:  log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		file:        file,
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.infoLogger.Println(message)
	fmt.Printf("[%s] INFO: %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.errorLogger.Println(message)
	fmt.Printf("[%s] ERROR: %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

// Close closes the log file
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}
