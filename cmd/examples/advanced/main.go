package main

import (
	"fmt"
	"time"

	"github.com/jakubbbdev/go-logging/pkg/logging"
)

func main() {
	fmt.Println("🚀 Advanced Go Logging Library Demo")
	fmt.Println("=====================================")

	// 1. Performance Optimizations Demo
	demoPerformanceOptimizations()

	// 2. Advanced Handlers Demo
	demoAdvancedHandlers()

	// 3. Async Logging Demo
	demoAsyncLogging()

	// 4. Sampling Demo
	demoSampling()

	// 5. Rotating File Demo
	demoRotatingFile()

	fmt.Println("\n✅ All demos completed successfully!")
}

func demoPerformanceOptimizations() {
	fmt.Println("\n📈 1. Performance Optimizations Demo")
	fmt.Println("-----------------------------------")

	logger := logging.NewLogger()
	logger.SetLevel(logging.InfoLevel)

	// Regular logging
	start := time.Now()
	for i := 0; i < 1000; i++ {
		logger.Info("regular message", i)
	}
	regularDuration := time.Since(start)

	// Fast logging (zero-allocation)
	start = time.Now()
	for i := 0; i < 1000; i++ {
		logger.InfoFast("fast message")
	}
	fastDuration := time.Since(start)

	fmt.Printf("Regular logging: %v\n", regularDuration)
	fmt.Printf("Fast logging: %v\n", fastDuration)
	fmt.Printf("Performance improvement: %.2fx faster\n", float64(regularDuration)/float64(fastDuration))
}

func demoAdvancedHandlers() {
	fmt.Println("\n🔧 2. Advanced Handlers Demo")
	fmt.Println("----------------------------")

	// Create multiple handlers
	consoleHandler := logging.NewConsoleHandler()
	fileHandler, _ := logging.NewFileHandler("advanced_demo.log")
	defer fileHandler.(*logging.FileHandler).Close()

	// Create rotating file handler
	rotatingHandler, _ := logging.NewRotatingFileHandler("rotating.log", 1024, 3)
	defer rotatingHandler.(*logging.RotatingFileHandler).Close()

	// Create multi handler
	multiHandler := logging.NewMultiHandler(
		consoleHandler,
		fileHandler,
		rotatingHandler,
	)

	logger := logging.NewLogger()
	logger.SetHandler(multiHandler)
	logger.SetFormatter(logging.NewJSONFormatter())

	// Log with structured data
	logger.WithFields(logging.Fields{
		"demo":      "advanced_handlers",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	}).Info("Testing multiple handlers")

	fmt.Println("✅ Logged to console, file, and rotating file")
}

func demoAsyncLogging() {
	fmt.Println("\n⚡ 3. Async Logging Demo")
	fmt.Println("------------------------")

	// Create base handler
	baseHandler := logging.NewConsoleHandler()

	// Create async handler with buffer size 100 and 4 workers
	asyncHandler := logging.NewAsyncHandler(baseHandler, 100, 4)
	defer asyncHandler.(*logging.AsyncHandler).Stop()

	logger := logging.NewLogger()
	logger.SetHandler(asyncHandler)
	logger.SetFormatter(logging.NewJSONFormatter())

	// Send many logs asynchronously
	fmt.Println("Sending 50 logs asynchronously...")
	start := time.Now()
	for i := 0; i < 50; i++ {
		logger.WithFields(logging.Fields{
			"async_id": i,
			"worker":   i % 4,
		}).Info("Async log message")
	}
	sendDuration := time.Since(start)

	// Give time for processing
	time.Sleep(200 * time.Millisecond)

	fmt.Printf("✅ Sent 50 logs in %v (non-blocking)\n", sendDuration)
}

func demoSampling() {
	fmt.Println("\n🎯 4. Sampling Demo")
	fmt.Println("-------------------")

	// Create base handler
	baseHandler := logging.NewConsoleHandler()

	// Create sampling handler with 30% rate
	samplingHandler := logging.NewSamplingHandler(baseHandler, 0.3)

	logger := logging.NewLogger()
	logger.SetHandler(samplingHandler)
	logger.SetFormatter(logging.NewJSONFormatter())

	// Send many logs (only 30% should be logged)
	fmt.Println("Sending 100 logs with 30% sampling...")
	for i := 0; i < 100; i++ {
		logger.WithFields(logging.Fields{
			"sample_id": i,
			"rate":      0.3,
		}).Info("Sampled log message")
	}

	fmt.Println("✅ Sampling completed (only ~30% of logs were processed)")
}

func demoRotatingFile() {
	fmt.Println("\n🔄 5. Rotating File Demo")
	fmt.Println("-------------------------")

	// Create rotating file handler with small max size to trigger rotation
	rotatingHandler, _ := logging.NewRotatingFileHandler("demo_rotate.log", 200, 3)
	defer rotatingHandler.(*logging.RotatingFileHandler).Close()

	logger := logging.NewLogger()
	logger.SetHandler(rotatingHandler)
	logger.SetFormatter(logging.NewJSONFormatter())

	// Write enough logs to trigger rotation
	fmt.Println("Writing logs to trigger file rotation...")
	for i := 0; i < 20; i++ {
		logger.WithFields(logging.Fields{
			"rotation_demo": true,
			"message_id":    i,
			"timestamp":     time.Now().Unix(),
		}).Info("This is a long message that will trigger file rotation when the file gets too large")
	}

	fmt.Println("✅ File rotation demo completed")
	fmt.Println("   Check for files: demo_rotate.log, demo_rotate.log.1, demo_rotate.log.2")
}

// Example of using all features together
func demoProductionSetup() {
	fmt.Println("\n🏭 Production Setup Example")
	fmt.Println("----------------------------")

	// Create production-ready logger
	logger := logging.NewLogger()
	logger.SetLevel(logging.InfoLevel)

	// Create handlers
	consoleHandler := logging.NewConsoleHandler()
	rotatingHandler, _ := logging.NewRotatingFileHandler("production.log", 10*1024*1024, 5) // 10MB, 5 files
	defer rotatingHandler.(*logging.RotatingFileHandler).Close()

	// Create async multi handler
	multiHandler := logging.NewMultiHandler(consoleHandler, rotatingHandler)
	asyncHandler := logging.NewAsyncHandler(multiHandler, 1000, 8)
	defer asyncHandler.(*logging.AsyncHandler).Stop()

	logger.SetHandler(asyncHandler)
	logger.SetFormatter(logging.NewJSONFormatter())

	// Simulate production workload
	fmt.Println("Simulating production workload...")
	for i := 0; i < 100; i++ {
		logger.WithFields(logging.Fields{
			"service":     "api",
			"endpoint":    "/users",
			"user_id":     i,
			"duration_ms": time.Duration(i*10) * time.Millisecond,
		}).Info("API request processed")

		// Simulate some errors
		if i%10 == 0 {
			logger.WithFields(logging.Fields{
				"service": "api",
				"error":   "database_connection_timeout",
			}).Error("Database connection failed")
		}
	}

	fmt.Println("✅ Production setup demo completed")
}
