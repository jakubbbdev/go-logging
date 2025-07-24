package logging

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// MetricsCollector collects logging metrics
type MetricsCollector struct {
	logCount    map[Level]int64
	logErrors   int64
	logDuration map[Level]time.Duration
	lastLogTime time.Time
	mu          sync.RWMutex
	enabled     bool
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		logCount:    make(map[Level]int64),
		logDuration: make(map[Level]time.Duration),
		enabled:     true,
	}
}

// RecordLog records a log entry
func (m *MetricsCollector) RecordLog(level Level, duration time.Duration) {
	if !m.enabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.logCount[level]++
	m.logDuration[level] += duration
	m.lastLogTime = time.Now()
}

// RecordError records a logging error
func (m *MetricsCollector) RecordError() {
	if !m.enabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.logErrors++
}

// GetStats returns current statistics
func (m *MetricsCollector) GetStats() LogStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := LogStats{
		LogCount:    make(map[Level]int64),
		LogDuration: make(map[Level]time.Duration),
		LogErrors:   m.logErrors,
		LastLogTime: m.lastLogTime,
	}

	for level, count := range m.logCount {
		stats.LogCount[level] = count
	}

	for level, duration := range m.logDuration {
		stats.LogDuration[level] = duration
	}

	return stats
}

// Reset resets all metrics
func (m *MetricsCollector) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logCount = make(map[Level]int64)
	m.logDuration = make(map[Level]time.Duration)
	m.logErrors = 0
	m.lastLogTime = time.Time{}
}

// Enable enables metrics collection
func (m *MetricsCollector) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable disables metrics collection
func (m *MetricsCollector) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// LogStats represents logging statistics
type LogStats struct {
	LogCount    map[Level]int64
	LogDuration map[Level]time.Duration
	LogErrors   int64
	LastLogTime time.Time
}

// MetricsHandler creates an HTTP handler for exposing metrics
func (m *MetricsCollector) MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := m.GetStats()

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		// Prometheus-style metrics format
		fmt.Fprintf(w, "# HELP logging_total Total number of log entries by level\n")
		fmt.Fprintf(w, "# TYPE logging_total counter\n")

		for level, count := range stats.LogCount {
			fmt.Fprintf(w, "logging_total{level=\"%s\"} %d\n", level.String(), count)
		}

		fmt.Fprintf(w, "\n# HELP logging_errors_total Total number of logging errors\n")
		fmt.Fprintf(w, "# TYPE logging_errors_total counter\n")
		fmt.Fprintf(w, "logging_errors_total %d\n", stats.LogErrors)

		fmt.Fprintf(w, "\n# HELP logging_duration_seconds Total time spent logging by level\n")
		fmt.Fprintf(w, "# TYPE logging_duration_seconds counter\n")

		for level, duration := range stats.LogDuration {
			fmt.Fprintf(w, "logging_duration_seconds{level=\"%s\"} %.6f\n",
				level.String(), duration.Seconds())
		}

		if !stats.LastLogTime.IsZero() {
			fmt.Fprintf(w, "\n# HELP logging_last_log_timestamp Last log entry timestamp\n")
			fmt.Fprintf(w, "# TYPE logging_last_log_timestamp gauge\n")
			fmt.Fprintf(w, "logging_last_log_timestamp %d\n", stats.LastLogTime.Unix())
		}
	}
}

// StartMetricsServer starts an HTTP server for metrics
func (m *MetricsCollector) StartMetricsServer(addr string, path string) error {
	http.HandleFunc(path, m.MetricsHandler())
	return http.ListenAndServe(addr, nil)
}

// MetricsHook creates a hook that records metrics
func NewMetricsHook(collector *MetricsCollector) Hook {
	return func(entry *Entry) {
		start := time.Now()
		// Simulate processing time for metrics
		duration := time.Since(start)
		collector.RecordLog(entry.Level, duration)
	}
}
