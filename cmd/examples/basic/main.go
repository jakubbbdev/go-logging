package main

import (
	"context"
	"time"

	"github.com/jakubbbdev/go-logging/pkg/logging"
)

func main() {
	// Create a new logger
	logger := logging.NewLogger()

	// Set log level
	logger.SetLevel(logging.InfoLevel)

	// Basic logging
	logger.Info("Application started")
	logger.Warn("This is a warning message")
	logger.Error("An error occurred")

	// Formatted logging
	logger.Infof("Server running on port %d", 8080)
	logger.Errorf("Failed to connect to database: %s", "connection timeout")

	// Structured logging with fields
	logger.WithFields(logging.Fields{
		"user_id": 123,
		"action":  "login",
		"ip":      "192.168.1.1",
	}).Info("User logged in successfully")

	// JSON formatting
	logger.SetFormatter(logging.NewJSONFormatter())
	logger.Info("This will be logged as JSON")

	// Context logging
	ctx := context.WithValue(context.Background(), "request_id", "abc123")
	ctx = logging.WithFields(ctx, logging.Fields{
		"user_agent": "Mozilla/5.0",
		"method":     "GET",
	})

	contextLogger := logging.FromContext(ctx)
	contextLogger.Info("Processing request")

	// File logging
	fileHandler, err := logging.NewFileHandler("app.log")
	if err != nil {
		logger.Error("Failed to create file handler:", err)
		return
	}

	// Use multiple handlers (console and file)
	multiHandler := logging.NewMultiHandler(
		logging.NewConsoleHandler(),
		fileHandler,
	)
	logger.SetHandler(multiHandler)

	logger.Info("This will be logged to both console and file")

	// Performance logging
	start := time.Now()
	time.Sleep(100 * time.Millisecond)
	duration := time.Since(start)

	logger.WithFields(logging.Fields{
		"duration_ms": duration.Milliseconds(),
		"operation":   "database_query",
	}).Info("Database query completed")

	// Different log levels
	logger.SetLevel(logging.DebugLevel)
	logger.Debug("This debug message will be shown")

	logger.SetLevel(logging.WarnLevel)
	logger.Debug("This debug message will be hidden")
	logger.Warn("This warning will be shown")

	// Fatal and panic (commented out to avoid program termination)
	// logger.Fatal("This would terminate the program")
	// logger.Panic("This would panic the program")
}
