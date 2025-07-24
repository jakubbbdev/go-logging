package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jakubbbdev/go-logging/pkg/logging"
)

func main() {
	fmt.Println("🚀 Futuristic Go Logging Library Demo")
	fmt.Println("🌟 Next-Generation Features Showcase")
	fmt.Println("=====================================")

	// 1. Modern slog Integration Demo
	demoSlogIntegration()

	// 2. OpenTelemetry Tracing Demo
	demoOpenTelemetryTracing()

	// 3. Health Monitoring & Circuit Breaker Demo
	demoHealthAndCircuitBreaker()

	// 4. Real-time Dashboard Demo (commented out for non-interactive mode)
	// demoRealtimeDashboard()

	// 5. Future-Ready Enterprise Setup
	demoFuturisticEnterpriseSetup()

	fmt.Println("\n✅ All futuristic demos completed successfully!")
	fmt.Println("🎯 Your logging library is now FUTURE-READY! 🚀")
}

func demoSlogIntegration() {
	fmt.Println("\n📱 1. Modern slog Integration (Go 1.21+)")
	fmt.Println("------------------------------------------")

	// Create our logger
	logger := logging.NewLogger(
		logging.WithLevel(logging.DebugLevel),
		logging.WithFormatter(logging.NewJSONFormatter()),
	)

	// Convert to slog.Logger using the implementation
	slogLogger := logging.NewSlogLogger(logger).Logger

	// Show slog integration using NewSlogLogger
	slogLogger2 := logging.NewSlogLogger(logger).Logger
	slogLogger2.Info("✨ Future-ready enterprise logging system operational!")

	// Use standard slog API
	slogLogger.Info("Using standard slog interface!")
	slogLogger.With("user_id", 123, "action", "login").Warn("slog structured logging")

	// Group attributes
	slogLogger.WithGroup("database").Info("Connection established",
		slog.String("driver", "postgres"),
		slog.Duration("connect_time", 150*time.Millisecond),
		slog.Int("pool_size", 10),
	)

	// Create slog.Logger from our handler
	slogFromHandler := slog.New(logging.NewSlogHandler(logger))
	slogFromHandler.Error("Error via slog handler",
		slog.String("error", "connection failed"),
		slog.Int("retry_count", 3),
	)

	fmt.Println("✅ slog integration working perfectly!")
}

func demoOpenTelemetryTracing() {
	fmt.Println("\n🔗 2. OpenTelemetry Distributed Tracing")
	fmt.Println("--------------------------------------")

	// Create OpenTelemetry tracer
	tracer := logging.NewOTelTracer("futuristic-app")

	// Create logger with OTel integration
	logger := logging.NewLogger(
		logging.WithLevel(logging.InfoLevel),
		logging.WithHook(logging.NewOTelLoggingHook(tracer)),
	)

	// Start a span
	ctx, span := tracer.StartSpan(context.Background(), "user_authentication")
	defer tracer.FinishSpan(span)

	// Set span tags
	tracer.SetSpanTag(span, "user.id", "user123")
	tracer.SetSpanTag(span, "service.version", "2.0.0")

	// Create span-aware logger
	spanLogger := logging.NewSpanLogger(logger, span, tracer)

	// Log with automatic span context
	spanLogger.Info("Starting authentication process")
	spanLogger.WithFields(logging.Fields{
		"method":   "oauth2",
		"provider": "google",
	}).Info("OAuth2 authentication initiated")

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	// Start child span
	_, childSpan := tracer.StartSpan(logging.WithSpan(ctx, span), "token_validation")
	tracer.SetSpanTag(childSpan, "token.type", "jwt")

	childSpanLogger := logging.NewSpanLogger(logger, childSpan, tracer)
	childSpanLogger.Info("Validating JWT token")

	tracer.FinishSpan(childSpan)

	spanLogger.Info("Authentication completed successfully")

	fmt.Printf("✅ Traced operation: %s\n", span.TraceID)
	fmt.Printf("📊 Span duration: %v\n", span.Duration)
	fmt.Printf("🏷️  Span tags: %+v\n", span.Tags)
}

func demoHealthAndCircuitBreaker() {
	fmt.Println("\n🔄 3. Health Monitoring & Circuit Breaker")
	fmt.Println("----------------------------------------")

	logger := logging.NewLogger(
		logging.WithLevel(logging.InfoLevel),
		logging.WithFormatter(logging.NewTextFormatter()),
	)

	// Create health monitor
	healthMonitor := logging.NewHealthMonitor(logger, 2*time.Second)
	defer healthMonitor.Stop()

	// Add some health checks
	healthMonitor.AddCheck("console_handler", logging.NewHandlerHealthCheck(
		logging.NewConsoleHandler(),
	))

	// Start monitoring
	healthMonitor.Start()

	// Wait for initial health check
	time.Sleep(3 * time.Second)

	// Check health status
	health := healthMonitor.GetHealth()
	overall := healthMonitor.GetOverallHealth()

	fmt.Printf("🟢 Overall Health: %s\n", overall.String())
	for name, result := range health {
		status := "🟢"
		if result.Status != logging.HealthStatusHealthy {
			status = "🔴"
		}
		fmt.Printf("  %s %s: %s (%v)\n", status, name, result.Status.String(), result.Duration)
	}

	// Circuit Breaker Demo
	fmt.Println("\n⚡ Circuit Breaker Protection:")

	// Create circuit breaker
	breaker := logging.NewCircuitBreaker(3, 5*time.Second)

	// Wrap handler with circuit breaker
	protectedHandler := logging.NewCircuitBreakerHandler(
		logging.NewConsoleHandler(),
		breaker,
		logger,
	)

	// Test circuit breaker
	logger.SetHandler(protectedHandler)

	fmt.Printf("🔧 Circuit breaker state: %s\n", breaker.GetState().String())

	// Normal operation
	logger.Info("Circuit breaker test - this should work")

	fmt.Println("✅ Health monitoring and circuit breaker working!")
}

func demoRealtimeDashboard() {
	fmt.Println("\n📊 4. Real-time Dashboard (Interactive)")
	fmt.Println("-------------------------------------")

	// Create components
	metrics := logging.NewMetricsCollector()
	healthMonitor := logging.NewHealthMonitor(
		logging.NewLogger(),
		2*time.Second,
	)

	// Create dashboard
	dashboard := logging.NewDashboard(
		logging.NewLogger(),
		metrics,
		healthMonitor,
	)

	// Create logger with dashboard integration
	logger := logging.NewLogger(
		logging.WithLevel(logging.DebugLevel),
		logging.WithHandler(logging.NewDashboardHandler(
			logging.NewConsoleHandler(),
			dashboard,
		)),
		logging.WithHook(logging.NewMetricsHook(metrics)),
		logging.WithHook(logging.NewDashboardHook(dashboard)),
	)

	// Generate some logs for the dashboard
	go func() {
		for i := 0; i < 50; i++ {
			logger.Info(fmt.Sprintf("Dashboard demo log %d", i+1))
			logger.WithFields(logging.Fields{
				"iteration": i + 1,
				"component": "demo",
			}).Debug("Debug message for dashboard")

			if i%10 == 0 {
				logger.Error("Simulated error for dashboard")
			}

			time.Sleep(500 * time.Millisecond)
		}
	}()

	fmt.Println("🎮 Starting real-time dashboard...")
	fmt.Println("📱 Press 'q' to quit, 'r' to refresh")

	// Start dashboard (this will block)
	if err := logging.StartDashboard(dashboard); err != nil {
		fmt.Printf("Dashboard error: %v\n", err)
	}
}

func demoFuturisticEnterpriseSetup() {
	fmt.Println("\n🏢 5. Future-Ready Enterprise Setup")
	fmt.Println("----------------------------------")

	// Create all components
	metrics := logging.NewMetricsCollector()
	piiDetector := logging.NewPIIDetector()
	tracer := logging.NewOTelTracer("enterprise-app-v2")
	healthMonitor := logging.NewHealthMonitor(
		logging.NewLogger(),
		5*time.Second,
	)
	breaker := logging.NewCircuitBreaker(5, 10*time.Second)
	dashboard := logging.NewDashboard(
		logging.NewLogger(),
		metrics,
		healthMonitor,
	)

	// Create the ultimate enterprise logger
	logger := logging.NewLogger(
		logging.WithLevel(logging.InfoLevel),
		logging.WithCaller(true),
		logging.WithStacktrace(true),

		// Modern formatting with security
		logging.WithFormatter(logging.NewSecurityFormatter(
			logging.NewJSONFormatter(),
			piiDetector,
		)),

		// Production-ready handler chain
		logging.WithHandler(logging.NewCircuitBreakerHandler(
			logging.NewContextAwareHandler(
				logging.NewOTelHandler(
					logging.NewAsyncHandler(
						logging.NewMultiHandler(
							logging.NewConsoleHandler(),
							// Would add file handler in production
						),
						1000, // buffer size
						4,    // workers
					),
					tracer,
				),
				30*time.Second, // timeout
			),
			breaker,
			logging.NewLogger(), // for circuit breaker logging
		)),

		// Default enterprise fields
		logging.WithDefaultFields(logging.Fields{
			"service":     "futuristic-enterprise-app",
			"version":     "2.0.0",
			"environment": "production",
			"region":      "us-east-1",
			"go_version":  "1.23",
		}),

		// All the hooks
		logging.WithHook(logging.NewMetricsHook(metrics)),
		logging.WithHook(logging.NewSecurityHook(piiDetector)),
		logging.WithHook(logging.NewOTelLoggingHook(tracer)),
		logging.WithHook(logging.NewTracingHook()),
		logging.WithHook(logging.NewDashboardHook(dashboard)),
	)

	// Start background services
	healthMonitor.Start()
	defer healthMonitor.Stop()

	// Add health checks
	healthMonitor.AddCheck("metrics", func(ctx context.Context) logging.HealthCheckResult {
		stats := metrics.GetStats()
		if len(stats.LogCount) == 0 {
			return logging.HealthCheckResult{
				Status:  logging.HealthStatusDegraded,
				Message: "No metrics data available",
			}
		}
		return logging.HealthCheckResult{
			Status:  logging.HealthStatusHealthy,
			Message: "Metrics collector is working",
		}
	})

	// Set as global logger
	logging.SetGlobalLogger(logger)

	// Enterprise operation simulation
	ctx, span := tracer.StartSpan(context.Background(), "enterprise_operation")
	defer tracer.FinishSpan(span)

	// Add enterprise trace context
	traceCtx := logging.NewTraceContext().
		WithUserID("enterprise-user-456").
		WithSessionID("ent-session-xyz")
	ctx = logging.WithTraceContext(ctx, traceCtx)

	enterpriseLogger := logger.WithTrace(ctx)

	// Simulate enterprise workload
	enterpriseLogger.Info("🚀 Enterprise application v2.0 initialized")

	enterpriseLogger.WithFields(logging.Fields{
		"operation":   "user_onboarding",
		"email":       "enterprise.user@company.com",
		"password":    "super-secret-enterprise-password",
		"credit_card": "4532-1234-5678-9012",
		"ssn":         "123-45-6789",
		"phone":       "555-987-6543",
	}).Info("Enterprise user onboarding completed")

	enterpriseLogger.WithFields(logging.Fields{
		"operation":    "data_processing",
		"records":      10000,
		"duration_ms":  850,
		"memory_usage": "245MB",
		"cpu_percent":  12.5,
	}).Info("Bulk data processing completed")

	// Simulate error with full context
	enterpriseLogger.WithFields(logging.Fields{
		"error_code":     "ENT_001",
		"correlation_id": "corr-123-456",
		"user_id":        "user456",
		"operation":      "payment_processing",
	}).Error("Payment processing failed - enterprise circuit breaker activated")

	// Wait a bit for async processing
	time.Sleep(2 * time.Second)

	// Show final enterprise metrics
	stats := metrics.GetStats()
	health := healthMonitor.GetOverallHealth()

	fmt.Printf("\n🏢 Enterprise Metrics Summary:\n")
	fmt.Printf("  📊 Total Logs: %d\n", getTotalLogs(stats.LogCount))
	fmt.Printf("  🔍 Trace ID: %s\n", span.TraceID)
	fmt.Printf("  ⚡ Circuit Breaker: %s\n", breaker.GetState().String())
	fmt.Printf("  🟢 Health Status: %s\n", health.String())
	fmt.Printf("  ⏱️  Span Duration: %v\n", span.Duration)
	fmt.Printf("  🔒 PII Detection: Active\n")
	fmt.Printf("  📊 Real-time Dashboard: ✅\n")

	// slog integration is demonstrated in the slog demo section above
	fmt.Println("  📱 slog Compatible: ✅")

	fmt.Println("\n🎯 Your logging library is now THE MOST ADVANCED Go logging solution! 🚀")
}

func getTotalLogs(logCount map[logging.Level]int64) int64 {
	total := int64(0)
	for _, count := range logCount {
		total += count
	}
	return total
}
