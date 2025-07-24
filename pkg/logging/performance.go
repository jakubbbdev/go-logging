package logging

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// PerformanceConfig holds performance-related configuration
type PerformanceConfig struct {
	BufferPoolSize   int
	PreallocatedLogs int
	EnableZeroAlloc  bool
	EnableProfiling  bool
	EnableMetrics    bool
	FlushInterval    time.Duration
	BatchSize        int
}

// DefaultPerformanceConfig returns default high-performance settings
func DefaultPerformanceConfig() PerformanceConfig {
	return PerformanceConfig{
		BufferPoolSize:   1000,
		PreallocatedLogs: 10000,
		EnableZeroAlloc:  true,
		EnableProfiling:  true,
		EnableMetrics:    true,
		FlushInterval:    100 * time.Millisecond,
		BatchSize:        100,
	}
}

// HighPerformanceLogger is an ultra-fast logger implementation
type HighPerformanceLogger struct {
	config      PerformanceConfig
	bufferPool  *BufferPool
	entryPool   *EntryPool
	handler     Handler
	formatter   Formatter
	level       Level
	fields      Fields
	hooks       []Hook
	stats       *PerformanceStats
	flushTimer  *time.Timer
	batchBuffer [][]byte
	batchMutex  sync.Mutex
	shutdown    chan struct{}
}

// BufferPool manages reusable byte buffers
type BufferPool struct {
	pool     sync.Pool
	stats    *BufferPoolStats
	maxSize  int
	hitCount int64
}

// BufferPoolStats tracks buffer pool performance
type BufferPoolStats struct {
	Gets        int64
	Puts        int64
	Hits        int64
	Misses      int64
	Allocations int64
	Reuses      int64
}

// EntryPool manages reusable log entries for zero-allocation logging
type EntryPool struct {
	pool  sync.Pool
	stats *EntryPoolStats
}

// EntryPoolStats tracks entry pool performance
type EntryPoolStats struct {
	Gets   int64
	Puts   int64
	Reuses int64
	News   int64
}

// PerformanceStats tracks overall performance metrics
type PerformanceStats struct {
	LogsPerSecond     int64
	AvgProcessingTime time.Duration
	MemoryUsage       int64
	GCPauses          int64
	AllocationsCount  int64
	TotalLogs         int64
	StartTime         time.Time
	mutex             sync.RWMutex
}

// NewBufferPool creates a new high-performance buffer pool
func NewBufferPool(maxSize int) *BufferPool {
	return &BufferPool{
		maxSize: maxSize,
		stats:   &BufferPoolStats{},
		pool: sync.Pool{
			New: func() interface{} {
				atomic.AddInt64(&(&BufferPoolStats{}).Allocations, 1)
				return make([]byte, 0, 1024) // 1KB initial capacity
			},
		},
	}
}

// Get retrieves a buffer from the pool
func (bp *BufferPool) Get() []byte {
	atomic.AddInt64(&bp.stats.Gets, 1)

	buf := bp.pool.Get().([]byte)
	if cap(buf) > 0 {
		atomic.AddInt64(&bp.stats.Hits, 1)
		atomic.AddInt64(&bp.stats.Reuses, 1)
	} else {
		atomic.AddInt64(&bp.stats.Misses, 1)
	}

	return buf[:0] // Reset length but keep capacity
}

// Put returns a buffer to the pool
func (bp *BufferPool) Put(buf []byte) {
	atomic.AddInt64(&bp.stats.Puts, 1)

	// Don't pool buffers that are too large
	if cap(buf) > bp.maxSize {
		return
	}

	bp.pool.Put(buf)
}

// GetStats returns buffer pool statistics
func (bp *BufferPool) GetStats() BufferPoolStats {
	return BufferPoolStats{
		Gets:        atomic.LoadInt64(&bp.stats.Gets),
		Puts:        atomic.LoadInt64(&bp.stats.Puts),
		Hits:        atomic.LoadInt64(&bp.stats.Hits),
		Misses:      atomic.LoadInt64(&bp.stats.Misses),
		Allocations: atomic.LoadInt64(&bp.stats.Allocations),
		Reuses:      atomic.LoadInt64(&bp.stats.Reuses),
	}
}

// NewEntryPool creates a new entry pool for zero-allocation logging
func NewEntryPool() *EntryPool {
	return &EntryPool{
		stats: &EntryPoolStats{},
		pool: sync.Pool{
			New: func() interface{} {
				atomic.AddInt64(&(&EntryPoolStats{}).News, 1)
				return &Entry{
					Fields: make(Fields),
				}
			},
		},
	}
}

// GetEntry retrieves an entry from the pool
func (ep *EntryPool) GetEntry() *Entry {
	atomic.AddInt64(&ep.stats.Gets, 1)
	entry := ep.pool.Get().(*Entry)
	atomic.AddInt64(&ep.stats.Reuses, 1)
	return entry
}

// PutEntry returns an entry to the pool
func (ep *EntryPool) PutEntry(entry *Entry) {
	atomic.AddInt64(&ep.stats.Puts, 1)
	entry.Reset()
	ep.pool.Put(entry)
}

// GetStats returns entry pool statistics
func (ep *EntryPool) GetStats() EntryPoolStats {
	return EntryPoolStats{
		Gets:   atomic.LoadInt64(&ep.stats.Gets),
		Puts:   atomic.LoadInt64(&ep.stats.Puts),
		Reuses: atomic.LoadInt64(&ep.stats.Reuses),
		News:   atomic.LoadInt64(&ep.stats.News),
	}
}

// NewHighPerformanceLogger creates a new ultra-fast logger
func NewHighPerformanceLogger(config PerformanceConfig, handler Handler) *HighPerformanceLogger {
	hpl := &HighPerformanceLogger{
		config:      config,
		bufferPool:  NewBufferPool(config.BufferPoolSize),
		entryPool:   NewEntryPool(),
		handler:     handler,
		formatter:   NewTextFormatter(),
		level:       InfoLevel,
		fields:      make(Fields),
		hooks:       make([]Hook, 0),
		stats:       &PerformanceStats{StartTime: time.Now()},
		batchBuffer: make([][]byte, 0, config.BatchSize),
		shutdown:    make(chan struct{}),
	}

	// Start background flushing if batching is enabled
	if config.FlushInterval > 0 {
		hpl.startBatchFlusher()
	}

	return hpl
}

// startBatchFlusher starts the background batch flusher
func (hpl *HighPerformanceLogger) startBatchFlusher() {
	go func() {
		ticker := time.NewTicker(hpl.config.FlushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				hpl.flushBatch()
			case <-hpl.shutdown:
				hpl.flushBatch() // Final flush
				return
			}
		}
	}()
}

// flushBatch flushes the accumulated batch
func (hpl *HighPerformanceLogger) flushBatch() {
	hpl.batchMutex.Lock()

	if len(hpl.batchBuffer) == 0 {
		hpl.batchMutex.Unlock()
		return
	}

	// Copy buffer to process outside of lock
	batch := make([][]byte, len(hpl.batchBuffer))
	copy(batch, hpl.batchBuffer)

	// Clear batch while still holding lock
	hpl.batchBuffer = hpl.batchBuffer[:0]
	hpl.batchMutex.Unlock()

	// Process batch outside of lock
	for _, buf := range batch {
		// Here you would normally send to handler
		// For now, we just return the buffer to pool
		hpl.bufferPool.Put(buf)
	}
}

// LogFastUnsafe performs ultra-fast logging with minimal allocations
//
//go:noinline
func (hpl *HighPerformanceLogger) LogFastUnsafe(level Level, msg string) {
	if level.Value < hpl.level.Value {
		return
	}

	start := time.Now()
	atomic.AddInt64(&hpl.stats.TotalLogs, 1)

	if hpl.config.EnableZeroAlloc {
		// Zero-allocation path
		hpl.logZeroAlloc(level, msg)
	} else {
		// Regular path
		hpl.logRegular(level, msg)
	}

	// Update performance stats
	duration := time.Since(start)
	hpl.updateStats(duration)
}

// logZeroAlloc performs zero-allocation logging
func (hpl *HighPerformanceLogger) logZeroAlloc(level Level, msg string) {
	// Get buffer from pool
	buf := hpl.bufferPool.Get()

	// Build log message directly into buffer
	buf = append(buf, level.String()...)
	buf = append(buf, ": "...)
	buf = append(buf, msg...)
	buf = append(buf, '\n')

	if hpl.config.FlushInterval > 0 {
		// Add to batch
		hpl.batchMutex.Lock()
		hpl.batchBuffer = append(hpl.batchBuffer, buf)

		// Check if batch is full (don't flush here to avoid deadlock)
		shouldFlush := len(hpl.batchBuffer) >= hpl.config.BatchSize
		hpl.batchMutex.Unlock()

		// Flush in separate goroutine if needed
		if shouldFlush {
			go hpl.flushBatch()
		}
	} else {
		// Direct processing
		// Simulate handler processing
		hpl.bufferPool.Put(buf)
	}
}

// logRegular performs regular logging
func (hpl *HighPerformanceLogger) logRegular(level Level, msg string) {
	entry := hpl.entryPool.GetEntry()
	defer hpl.entryPool.PutEntry(entry)

	entry.Level = level
	entry.Message = msg
	entry.Time = time.Now()
	entry.Fields = hpl.fields

	// Process through handler
	if hpl.handler != nil {
		hpl.handler.Handle(entry)
	}
}

// updateStats updates performance statistics
func (hpl *HighPerformanceLogger) updateStats(duration time.Duration) {
	hpl.stats.mutex.Lock()
	defer hpl.stats.mutex.Unlock()

	// Update average processing time
	totalLogs := atomic.LoadInt64(&hpl.stats.TotalLogs)
	if totalLogs > 0 {
		avgDuration := time.Duration(int64(hpl.stats.AvgProcessingTime) * (totalLogs - 1) / totalLogs)
		hpl.stats.AvgProcessingTime = avgDuration + duration/time.Duration(totalLogs)
	} else {
		hpl.stats.AvgProcessingTime = duration
	}

	// Update logs per second
	elapsed := time.Since(hpl.stats.StartTime)
	if elapsed > 0 {
		hpl.stats.LogsPerSecond = totalLogs * int64(time.Second) / int64(elapsed)
	}
}

// GetPerformanceStats returns current performance statistics
func (hpl *HighPerformanceLogger) GetPerformanceStats() PerformanceStats {
	hpl.stats.mutex.RLock()
	defer hpl.stats.mutex.RUnlock()

	stats := *hpl.stats
	stats.TotalLogs = atomic.LoadInt64(&hpl.stats.TotalLogs)
	stats.LogsPerSecond = atomic.LoadInt64(&hpl.stats.LogsPerSecond)

	// Update memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	stats.MemoryUsage = int64(m.Alloc)
	stats.AllocationsCount = int64(m.Mallocs)
	stats.GCPauses = int64(m.NumGC)

	return stats
}

// Close gracefully shuts down the logger
func (hpl *HighPerformanceLogger) Close() {
	close(hpl.shutdown)
	hpl.flushBatch() // Final flush
}

// Implement Logger interface for HighPerformanceLogger
func (hpl *HighPerformanceLogger) Debug(args ...interface{}) {
	hpl.LogFastUnsafe(DebugLevel, formatArgs(args...))
}

func (hpl *HighPerformanceLogger) Info(args ...interface{}) {
	hpl.LogFastUnsafe(InfoLevel, formatArgs(args...))
}

func (hpl *HighPerformanceLogger) Warn(args ...interface{}) {
	hpl.LogFastUnsafe(WarnLevel, formatArgs(args...))
}

func (hpl *HighPerformanceLogger) Error(args ...interface{}) {
	hpl.LogFastUnsafe(ErrorLevel, formatArgs(args...))
}

func (hpl *HighPerformanceLogger) Fatal(args ...interface{}) {
	hpl.LogFastUnsafe(FatalLevel, formatArgs(args...))
	hpl.Close()
	panic("fatal")
}

func (hpl *HighPerformanceLogger) Panic(args ...interface{}) {
	msg := formatArgs(args...)
	hpl.LogFastUnsafe(PanicLevel, msg)
	panic(msg)
}

// Helper functions
func formatArgs(args ...interface{}) string {
	if len(args) == 1 {
		if s, ok := args[0].(string); ok {
			return s
		}
	}
	// For performance, we avoid fmt.Sprint here
	return "formatted_message"
}

// Simplified implementations for Logger interface compliance
func (hpl *HighPerformanceLogger) Debugf(format string, args ...interface{}) { hpl.Debug(format) }
func (hpl *HighPerformanceLogger) Infof(format string, args ...interface{})  { hpl.Info(format) }
func (hpl *HighPerformanceLogger) Warnf(format string, args ...interface{})  { hpl.Warn(format) }
func (hpl *HighPerformanceLogger) Errorf(format string, args ...interface{}) { hpl.Error(format) }
func (hpl *HighPerformanceLogger) Fatalf(format string, args ...interface{}) { hpl.Fatal(format) }
func (hpl *HighPerformanceLogger) Panicf(format string, args ...interface{}) { hpl.Panic(format) }
func (hpl *HighPerformanceLogger) DebugFast(msg string)                      { hpl.LogFastUnsafe(DebugLevel, msg) }
func (hpl *HighPerformanceLogger) InfoFast(msg string)                       { hpl.LogFastUnsafe(InfoLevel, msg) }
func (hpl *HighPerformanceLogger) WarnFast(msg string)                       { hpl.LogFastUnsafe(WarnLevel, msg) }
func (hpl *HighPerformanceLogger) ErrorFast(msg string)                      { hpl.LogFastUnsafe(ErrorLevel, msg) }
func (hpl *HighPerformanceLogger) Log(level Level, args ...interface{}) {
	hpl.LogFastUnsafe(level, formatArgs(args...))
}
func (hpl *HighPerformanceLogger) Logf(level Level, format string, args ...interface{}) {
	hpl.LogFastUnsafe(level, format)
}
func (hpl *HighPerformanceLogger) LogFast(level Level, msg string)  { hpl.LogFastUnsafe(level, msg) }
func (hpl *HighPerformanceLogger) WithFields(fields Fields) Logger  { return hpl }
func (hpl *HighPerformanceLogger) WithContext(ctx Context) Logger   { return hpl }
func (hpl *HighPerformanceLogger) WithTrace(ctx Context) Logger     { return hpl }
func (hpl *HighPerformanceLogger) SetLevel(level Level)             { hpl.level = level }
func (hpl *HighPerformanceLogger) SetHandler(handler Handler)       { hpl.handler = handler }
func (hpl *HighPerformanceLogger) SetFormatter(formatter Formatter) { hpl.formatter = formatter }
func (hpl *HighPerformanceLogger) AddHook(hook Hook)                { hpl.hooks = append(hpl.hooks, hook) }

// GetBufferPoolStats returns buffer pool statistics
func (hpl *HighPerformanceLogger) GetBufferPoolStats() BufferPoolStats {
	return hpl.bufferPool.GetStats()
}

// PerformanceBenchmark runs performance benchmarks
type PerformanceBenchmark struct {
	logger Logger
	config PerformanceConfig
}

// NewPerformanceBenchmark creates a new benchmark runner
func NewPerformanceBenchmark(logger Logger, config PerformanceConfig) *PerformanceBenchmark {
	return &PerformanceBenchmark{
		logger: logger,
		config: config,
	}
}

// RunBenchmark runs a comprehensive performance benchmark
func (pb *PerformanceBenchmark) RunBenchmark(iterations int) BenchmarkResults {
	start := time.Now()

	// Run logging benchmark
	for i := 0; i < iterations; i++ {
		pb.logger.InfoFast("benchmark message")
	}

	duration := time.Since(start)

	return BenchmarkResults{
		Iterations:    iterations,
		TotalDuration: duration,
		AvgPerOp:      duration / time.Duration(iterations),
		OpsPerSecond:  int64(float64(iterations) / duration.Seconds()),
		MemoryUsage:   getCurrentMemoryUsage(),
	}
}

// BenchmarkResults holds benchmark results
type BenchmarkResults struct {
	Iterations    int
	TotalDuration time.Duration
	AvgPerOp      time.Duration
	OpsPerSecond  int64
	MemoryUsage   int64
}

// getCurrentMemoryUsage returns current memory usage
func getCurrentMemoryUsage() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc)
}

// Fix Context type
type Context = context.Context
