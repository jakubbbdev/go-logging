package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jakubbbdev/go-logging/pkg/logging"
)

func main() {
	fmt.Println("🏢 Enterprise Go Logging Library Demo")
	fmt.Println("=====================================")

	// 1. Configuration Management Demo
	demoConfigurationManagement()

	// 2. Global Logger Demo
	demoGlobalLogger()

	// 3. Metrics Collection Demo
	demoMetricsCollection()

	// 4. Tracing & Context Demo
	demoTracingAndContext()

	// 5. Security & PII Detection Demo
	demoSecurityFeatures()

	// 6. Full Enterprise Setup
	demoEnterpriseSetup()

	fmt.Println("\n✅ All enterprise demos completed successfully!")
}

func demoConfigurationManagement() {
	fmt.Println("\n📋 1. Configuration Management")
	fmt.Println("------------------------------")

	// Configuration from environment variables
	config := logging.LoadConfigFromEnv()
	config.Level = "debug"
	config.Format = "json"
	config.IncludeCaller = true
	config.DefaultFields = map[string]string{
		"service": "enterprise-demo",
		"version": "1.0.0",
	}

	logger, err := config.ToLogger()
	if err != nil {
		log.Fatalf("Failed to create logger from config: %v", err)
	}

	logger.Info("Configuration-based logger initialized!")
	logger.WithFields(logging.Fields{"feature": "config"}).Debug("Config loaded successfully")
}

func demoGlobalLogger() {
	fmt.Println("\n🌍 2. Global Logger")
	fmt.Println("-------------------")

	// Set up global logger
	globalLogger := logging.NewLogger(
		logging.WithLevel(logging.InfoLevel),
		logging.WithFormatter(logging.NewTextFormatter()),
		logging.WithDefaultFields(logging.Fields{"app": "global-demo"}),
	)

	logging.SetGlobalLogger(globalLogger)

	// Use global logging functions
	logging.Info("This is using the global logger!")
	logging.WithGlobalFields(logging.Fields{"module": "auth"}).Warn("Global warning!")
	logging.Errorf("Global error: %s", "something went wrong")
}

func demoMetricsCollection() {
	fmt.Println("\n📊 3. Metrics Collection")
	fmt.Println("-----------------------")

	// Create metrics collector
	metrics := logging.NewMetricsCollector()

	// Create logger with metrics hook
	logger := logging.NewLogger(
		logging.WithLevel(logging.DebugLevel),
		logging.WithHook(logging.NewMetricsHook(metrics)),
	)

	// Generate some logs for metrics
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")

	// Show metrics
	stats := metrics.GetStats()
	fmt.Printf("Log Counts: %+v\n", stats.LogCount)
	fmt.Printf("Total Errors: %d\n", stats.LogErrors)
	fmt.Printf("Last Log Time: %v\n", stats.LastLogTime)

	// Note: In production, you'd start metrics server with:
	// go metrics.StartMetricsServer(":8080", "/metrics")
}

func demoTracingAndContext() {
	fmt.Println("\n🔍 4. Tracing & Context")
	fmt.Println("----------------------")

	// Create trace context
	traceCtx := logging.NewTraceContext().
		WithUserID("user123").
		WithSessionID("session456")

	// Add to context
	ctx := logging.WithTraceContext(context.Background(), traceCtx)

	// Create logger with tracing hook
	logger := logging.NewLogger(
		logging.WithLevel(logging.InfoLevel),
		logging.WithHook(logging.NewTracingHook()),
	)

	// Log with trace context
	tracedLogger := logger.WithTrace(ctx)
	tracedLogger.Info("This log has tracing information!")
	tracedLogger.WithFields(logging.Fields{"action": "login"}).Info("User login attempt")

	fmt.Printf("Trace Context: %s\n", traceCtx.String())
}

func demoSecurityFeatures() {
	fmt.Println("\n🔒 5. Security & PII Detection")
	fmt.Println("------------------------------")

	// Create PII detector
	piiDetector := logging.NewPIIDetector()

	// Create logger with security hook
	logger := logging.NewLogger(
		logging.WithLevel(logging.InfoLevel),
		logging.WithHook(logging.NewSecurityHook(piiDetector)),
	)

	// Log with PII data (will be automatically sanitized)
	logger.WithFields(logging.Fields{
		"email":    "user@example.com",
		"phone":    "555-123-4567",
		"password": "secret123",
		"ssn":      "123-45-6789",
	}).Warn("User data logged - PII should be sanitized!")

	// Test string sanitization
	sensitiveMessage := "User email: john.doe@example.com and phone: 555-987-6543"
	sanitized := piiDetector.SanitizeString(sensitiveMessage)
	fmt.Printf("Original: %s\n", sensitiveMessage)
	fmt.Printf("Sanitized: %s\n", sanitized)
}

func demoEnterpriseSetup() {
	fmt.Println("\n🏢 6. Full Enterprise Setup")
	fmt.Println("---------------------------")

	// Create enterprise-grade logger
	metrics := logging.NewMetricsCollector()
	piiDetector := logging.NewPIIDetector()

	logger := logging.NewLogger(
		logging.WithLevel(logging.InfoLevel),
		logging.WithCaller(true),
		logging.WithStacktrace(true),
		logging.WithFormatter(logging.NewSecurityFormatter(
			logging.NewJSONFormatter(),
			piiDetector,
		)),
		logging.WithDefaultFields(logging.Fields{
			"service":     "enterprise-app",
			"version":     "1.0.0",
			"environment": "production",
		}),
		logging.WithHook(logging.NewMetricsHook(metrics)),
		logging.WithHook(logging.NewSecurityHook(piiDetector)),
		logging.WithHook(logging.NewTracingHook()),
	)

	// Create trace context for request
	traceCtx := logging.NewTraceContext().
		WithUserID("enterprise-user").
		WithSessionID("ent-session-789")

	ctx := logging.WithTraceContext(context.Background(), traceCtx)

	// Enterprise logging in action
	enterpriseLogger := logger.WithTrace(ctx)

	enterpriseLogger.Info("Enterprise application started")

	enterpriseLogger.WithFields(logging.Fields{
		"operation": "user_authentication",
		"email":     "enterprise.user@company.com",
		"password":  "super-secret-password",
	}).Info("User authentication successful")

	enterpriseLogger.WithFields(logging.Fields{
		"operation":   "data_processing",
		"records":     1000,
		"duration_ms": 250,
	}).Info("Data processing completed")

	// Simulate error with stack trace
	enterpriseLogger.WithFields(logging.Fields{
		"error_code": "AUTH_001",
		"user_id":    "user123",
	}).Error("Authentication failed - account locked")

	// Show final metrics
	time.Sleep(100 * time.Millisecond) // Allow async processing
	stats := metrics.GetStats()
	fmt.Printf("\nFinal Enterprise Metrics:\n")
	fmt.Printf("- Total Logs: %d\n", getTotalLogs(stats.LogCount))
	fmt.Printf("- Errors: %d\n", stats.LogErrors)
	fmt.Printf("- Last Activity: %v\n", stats.LastLogTime.Format(time.RFC3339))
}

func getTotalLogs(logCount map[logging.Level]int64) int64 {
	total := int64(0)
	for _, count := range logCount {
		total += count
	}
	return total
}
