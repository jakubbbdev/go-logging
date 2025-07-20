package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jakubbbdev/go-logging/pkg/logging"
)

// RequestLogger middleware for logging HTTP requests
func RequestLogger(logger logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create context with request fields
			ctx := logging.WithFields(r.Context(), logging.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"user_agent": r.UserAgent(),
				"remote_ip":  r.RemoteAddr,
			})

			// Add logger to context
			ctx = logging.WithLogger(ctx, logger)
			r = r.WithContext(ctx)

			// Call next handler
			next.ServeHTTP(w, r)

			// Log request completion
			duration := time.Since(start)
			logger.WithFields(logging.Fields{
				"duration_ms": duration.Milliseconds(),
				"status":      200, // You might want to capture actual status
			}).Info("Request completed")
		})
	}
}

// HomeHandler handles the home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.FromContext(r.Context())

	logger.Info("Serving home page")

	// Simulate some work
	time.Sleep(50 * time.Millisecond)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head><title>Go Logging Example</title></head>
		<body>
			<h1>Welcome to Go Logging Example</h1>
			<p>Check the console and logs/app.log for logging output.</p>
		</body>
		</html>
	`)
}

// APIHandler handles API requests
func APIHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.FromContext(r.Context())

	// Simulate API processing
	logger.WithFields(logging.Fields{
		"endpoint": "/api/data",
		"params":   r.URL.Query(),
	}).Info("Processing API request")

	// Simulate some processing time
	time.Sleep(100 * time.Millisecond)

	// Simulate occasional errors
	if r.URL.Query().Get("error") == "true" {
		logger.Error("Simulated API error")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "success", "message": "API response"}`)
}

func main() {
	// Create logger with JSON formatting
	logger := logging.NewLogger()
	logger.SetLevel(logging.DebugLevel)
	logger.SetFormatter(logging.NewJSONFormatter())

	// Create file handler for logs
	fileHandler, err := logging.NewFileHandler("logs/app.log")
	if err != nil {
		logger.Error("Failed to create file handler:", err)
		return
	}

	// Use multi handler for console and file output
	multiHandler := logging.NewMultiHandler(
		logging.NewConsoleHandler(),
		fileHandler,
	)
	logger.SetHandler(multiHandler)

	logger.Info("Starting web server")

	// Create mux and add routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", HomeHandler)
	mux.HandleFunc("/api/data", APIHandler)

	// Add request logging middleware
	handler := RequestLogger(logger)(mux)

	// Start server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.WithFields(logging.Fields{
		"port": 8080,
		"host": "localhost",
	}).Info("Server started")

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error:", err)
		}
	}()

	// Wait for interrupt signal
	logger.Info("Press Ctrl+C to stop the server")
	select {}
}
