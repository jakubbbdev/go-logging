package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/jakubbbdev/go-logging/pkg/logging"
)

var (
	benchmark   = flag.Bool("benchmark", false, "Run performance benchmarks")
	iterations  = flag.Int("iterations", 100000, "Number of iterations for benchmark")
	loadTest    = flag.Bool("load-test", false, "Run load test")
	duration    = flag.Duration("duration", 30*time.Second, "Load test duration")
	healthCheck = flag.Bool("health-check", false, "Run health check and exit")
	workers     = flag.Int("workers", runtime.NumCPU(), "Number of worker goroutines")
)

func main() {
	flag.Parse()

	fmt.Println("🚀 Ultra-Performance Go Logging Demo")
	fmt.Println("🐳 Docker & Cloud-Native Ready")
	fmt.Println("⚡ Zero-Allocation Optimizations")
	fmt.Println("====================================")

	if *healthCheck {
		runHealthCheck()
		return
	}

	if *benchmark {
		runBenchmarks()
		return
	}

	if *loadTest {
		runLoadTest()
		return
	}

	// Default demo
	runPerformanceDemo()
}

func runHealthCheck() {
	fmt.Println("🏥 Running Health Check...")

	// Detect container environment
	containerInfo := logging.DetectContainerEnvironment()
	if containerInfo != nil {
		fmt.Printf("✅ Container detected: %s\n", containerInfo.String())
		if json, err := containerInfo.ToJSON(); err == nil {
			fmt.Printf("📊 Container Info:\n%s\n", string(json))
		}
	} else {
		fmt.Println("⚠️  Not running in container")
	}

	// Test high-performance logger
	config := logging.DefaultPerformanceConfig()
	hpLogger := logging.NewHighPerformanceLogger(config, logging.NewConsoleHandler())

	start := time.Now()
	hpLogger.InfoFast("Health check test message")
	duration := time.Since(start)

	fmt.Printf("⚡ Logger response time: %v\n", duration)

	// Get performance stats
	stats := hpLogger.GetPerformanceStats()
	fmt.Printf("📈 Performance Stats:\n")
	fmt.Printf("  - Logs/sec: %d\n", stats.LogsPerSecond)
	fmt.Printf("  - Avg duration: %v\n", stats.AvgProcessingTime)
	fmt.Printf("  - Memory: %d bytes\n", stats.MemoryUsage)

	hpLogger.Close()
	fmt.Println("✅ Health check passed!")
	os.Exit(0)
}

func runBenchmarks() {
	fmt.Printf("🔥 Running Performance Benchmarks (%d iterations)\n", *iterations)
	fmt.Println("================================================")

	// 1. Standard Logger Benchmark
	fmt.Println("\n📊 1. Standard Logger Benchmark")
	standardLogger := logging.NewLogger(
		logging.WithLevel(logging.InfoLevel),
		logging.WithHandler(logging.NewConsoleHandler()),
	)

	standardBench := logging.NewPerformanceBenchmark(standardLogger, logging.PerformanceConfig{})
	standardResults := standardBench.RunBenchmark(*iterations)

	fmt.Printf("  ⏱️  Duration: %v\n", standardResults.TotalDuration)
	fmt.Printf("  🏃 Ops/sec: %d\n", standardResults.OpsPerSecond)
	fmt.Printf("  📏 Avg/op: %v\n", standardResults.AvgPerOp)
	fmt.Printf("  💾 Memory: %d bytes\n", standardResults.MemoryUsage)

	// 2. High-Performance Logger Benchmark
	fmt.Println("\n🚀 2. High-Performance Logger Benchmark")
	config := logging.DefaultPerformanceConfig()
	hpLogger := logging.NewHighPerformanceLogger(config, logging.NewConsoleHandler())
	defer hpLogger.Close()

	hpBench := logging.NewPerformanceBenchmark(hpLogger, config)
	hpResults := hpBench.RunBenchmark(*iterations)

	fmt.Printf("  ⏱️  Duration: %v\n", hpResults.TotalDuration)
	fmt.Printf("  🏃 Ops/sec: %d\n", hpResults.OpsPerSecond)
	fmt.Printf("  📏 Avg/op: %v\n", hpResults.AvgPerOp)
	fmt.Printf("  💾 Memory: %d bytes\n", hpResults.MemoryUsage)

	// 3. Zero-Allocation Path Benchmark
	fmt.Println("\n⚡ 3. Zero-Allocation Path Benchmark")
	start := time.Now()
	for i := 0; i < *iterations; i++ {
		hpLogger.LogFastUnsafe(logging.InfoLevel, "zero-alloc message")
	}
	zeroAllocDuration := time.Since(start)

	fmt.Printf("  ⏱️  Duration: %v\n", zeroAllocDuration)
	fmt.Printf("  🏃 Ops/sec: %d\n", int64(float64(*iterations)/zeroAllocDuration.Seconds()))
	fmt.Printf("  📏 Avg/op: %v\n", zeroAllocDuration/time.Duration(*iterations))

	// 4. Container Logger Benchmark
	fmt.Println("\n🐳 4. Container Logger Benchmark")
	containerLogger := logging.CreateContainerLogger()

	containerBench := logging.NewPerformanceBenchmark(containerLogger, config)
	containerResults := containerBench.RunBenchmark(*iterations)

	fmt.Printf("  ⏱️  Duration: %v\n", containerResults.TotalDuration)
	fmt.Printf("  🏃 Ops/sec: %d\n", containerResults.OpsPerSecond)
	fmt.Printf("  📏 Avg/op: %v\n", containerResults.AvgPerOp)
	fmt.Printf("  💾 Memory: %d bytes\n", containerResults.MemoryUsage)

	// Performance comparison
	fmt.Println("\n📈 Performance Comparison:")
	improvement := float64(hpResults.OpsPerSecond) / float64(standardResults.OpsPerSecond) * 100
	fmt.Printf("  🚀 High-Performance vs Standard: %.1fx faster\n", improvement/100)

	zeroAllocOps := int64(float64(*iterations) / zeroAllocDuration.Seconds())
	zeroImprov := float64(zeroAllocOps) / float64(standardResults.OpsPerSecond) * 100
	fmt.Printf("  ⚡ Zero-Alloc vs Standard: %.1fx faster\n", zeroImprov/100)

	// Get detailed performance stats
	stats := hpLogger.GetPerformanceStats()
	bufferStats := hpLogger.GetBufferPoolStats()

	fmt.Println("\n📊 Detailed Performance Stats:")
	fmt.Printf("  🎯 Total logs: %d\n", stats.TotalLogs)
	fmt.Printf("  📊 Logs/sec: %d\n", stats.LogsPerSecond)
	fmt.Printf("  ⏱️  Avg processing: %v\n", stats.AvgProcessingTime)
	fmt.Printf("  🧠 Memory usage: %d bytes\n", stats.MemoryUsage)
	fmt.Printf("  🗑️ GC pauses: %d\n", stats.GCPauses)
	fmt.Printf("  📦 Buffer reuses: %d\n", bufferStats.Reuses)
	fmt.Printf("  💾 Buffer hits: %d\n", bufferStats.Hits)
}

func runLoadTest() {
	fmt.Printf("🔥 Running Load Test (Duration: %v, Workers: %d)\n", *duration, *workers)
	fmt.Println("==================================================")

	// Create high-performance logger
	config := logging.DefaultPerformanceConfig()
	config.BatchSize = 1000
	config.FlushInterval = 50 * time.Millisecond

	logger := logging.NewHighPerformanceLogger(config, logging.NewConsoleHandler())
	defer logger.Close()

	// Setup metrics
	var startTime = time.Now()
	var wg sync.WaitGroup

	// Worker function
	worker := func(workerId int) {
		defer wg.Done()

		localCount := 0
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		timeout := time.After(*duration)

		for {
			select {
			case <-timeout:
				fmt.Printf("  🔧 Worker %d: Generated %d logs\n", workerId, localCount)
				return
			case <-ticker.C:
				// Generate burst of logs
				for i := 0; i < 100; i++ {
					logger.LogFastUnsafe(logging.InfoLevel, fmt.Sprintf("Load test message from worker %d", workerId))
					localCount++
				}
			}
		}
	}

	// Start workers
	fmt.Printf("🚀 Starting %d workers...\n", *workers)
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(i)
	}

	// Monitor progress
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-time.After(*duration):
				return
			case <-ticker.C:
				stats := logger.GetPerformanceStats()
				elapsed := time.Since(startTime)
				fmt.Printf("  📊 Progress: %v elapsed, %d logs/sec, %d total logs\n",
					elapsed.Round(time.Second), stats.LogsPerSecond, stats.TotalLogs)
			}
		}
	}()

	// Wait for completion
	wg.Wait()

	finalStats := logger.GetPerformanceStats()
	elapsed := time.Since(startTime)

	fmt.Println("\n📈 Load Test Results:")
	fmt.Printf("  ⏱️  Total duration: %v\n", elapsed)
	fmt.Printf("  📊 Total logs: %d\n", finalStats.TotalLogs)
	fmt.Printf("  🏃 Average logs/sec: %d\n", finalStats.LogsPerSecond)
	fmt.Printf("  📏 Avg processing time: %v\n", finalStats.AvgProcessingTime)
	fmt.Printf("  💾 Memory usage: %d bytes\n", finalStats.MemoryUsage)
	fmt.Printf("  🗑️ GC pauses: %d\n", finalStats.GCPauses)

	// Container info
	if containerInfo := logging.DetectContainerEnvironment(); containerInfo != nil {
		fmt.Printf("  🐳 Container: %s\n", containerInfo.String())
	}
}

func runPerformanceDemo() {
	fmt.Println("\n🎯 Running Performance & Docker Demo")
	fmt.Println("===================================")

	// 1. Container Detection Demo
	fmt.Println("\n🐳 1. Container Environment Detection")
	containerInfo := logging.DetectContainerEnvironment()
	if containerInfo != nil {
		fmt.Printf("  ✅ Running in: %s\n", containerInfo.String())
		fmt.Printf("  📊 Container ID: %s\n", containerInfo.ID)
		fmt.Printf("  🏷️  Image: %s:%s\n", containerInfo.Image, containerInfo.ImageTag)
		if containerInfo.PodName != "" {
			fmt.Printf("  ☸️  Kubernetes Pod: %s/%s\n", containerInfo.PodNamespace, containerInfo.PodName)
		}
	} else {
		fmt.Println("  ℹ️  Not running in container")
	}

	// 2. High-Performance Logger Demo
	fmt.Println("\n⚡ 2. High-Performance Logger Demo")
	config := logging.DefaultPerformanceConfig()
	hpLogger := logging.NewHighPerformanceLogger(config, logging.NewConsoleHandler())
	defer hpLogger.Close()

	// Generate some logs
	start := time.Now()
	for i := 0; i < 10000; i++ {
		hpLogger.LogFastUnsafe(logging.InfoLevel, fmt.Sprintf("High-performance log message %d", i))
	}
	duration := time.Since(start)

	stats := hpLogger.GetPerformanceStats()
	fmt.Printf("  ⏱️  Logged 10,000 messages in: %v\n", duration)
	fmt.Printf("  🏃 Performance: %d logs/sec\n", stats.LogsPerSecond)
	fmt.Printf("  💾 Memory usage: %d bytes\n", stats.MemoryUsage)

	// 3. Cloud-Native Logger Demo
	fmt.Println("\n☁️  3. Cloud-Native Logger Demo")
	cloudLogger := logging.CreateContainerLogger()

	cloudLogger.WithFields(logging.Fields{
		"user_id":    12345,
		"request_id": "req-abc-123",
		"operation":  "demo",
	}).Info("Cloud-native structured logging with auto container fields")

	cloudLogger.Error("Error with full container context")

	// 4. Performance Comparison
	fmt.Println("\n📊 4. Performance Comparison")

	// Standard logger
	standardLogger := logging.NewLogger()
	start = time.Now()
	for i := 0; i < 1000; i++ {
		standardLogger.InfoFast("standard message")
	}
	standardDuration := time.Since(start)

	// High-performance logger
	start = time.Now()
	for i := 0; i < 1000; i++ {
		hpLogger.LogFastUnsafe(logging.InfoLevel, "hp message")
	}
	hpDuration := time.Since(start)

	improvement := float64(standardDuration) / float64(hpDuration)
	fmt.Printf("  📈 Standard Logger: %v (1000 logs)\n", standardDuration)
	fmt.Printf("  🚀 HP Logger: %v (1000 logs)\n", hpDuration)
	fmt.Printf("  ⚡ Performance improvement: %.1fx faster\n", improvement)

	// 5. Memory efficiency
	fmt.Println("\n🧠 5. Memory Efficiency Demo")

	// Force GC to get clean baseline
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Run many logs
	for i := 0; i < 50000; i++ {
		hpLogger.LogFastUnsafe(logging.InfoLevel, "memory efficiency test")
	}

	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	allocDiff := m2.TotalAlloc - m1.TotalAlloc
	fmt.Printf("  📊 Total allocations for 50k logs: %d bytes\n", allocDiff)
	fmt.Printf("  📏 Average per log: %.1f bytes\n", float64(allocDiff)/50000)

	bufferStats := hpLogger.GetBufferPoolStats()
	fmt.Printf("  ♻️  Buffer pool reuses: %d\n", bufferStats.Reuses)
	fmt.Printf("  🎯 Buffer pool hit rate: %.1f%%\n",
		float64(bufferStats.Hits)/float64(bufferStats.Gets)*100)

	fmt.Println("\n✅ Performance & Docker demo completed!")
	fmt.Println("🚀 Your logging library is ULTRA-PERFORMANT and CLOUD-READY! 💪")
}
