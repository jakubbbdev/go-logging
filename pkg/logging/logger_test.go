package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// TestLogger tests basic logging functionality
func TestLogger(t *testing.T) {
	var buf bytes.Buffer

	// Create a custom handler that writes to buffer
	handler := &testHandler{buf: &buf}

	logger := NewLogger()
	logger.SetHandler(handler)
	logger.SetLevel(DebugLevel)

	// Test basic logging
	logger.Info("test message")

	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("Expected log to contain 'test message', got: %s", buf.String())
	}

	// Test formatted logging
	buf.Reset()
	logger.Infof("formatted %s", "message")

	if !strings.Contains(buf.String(), "formatted message") {
		t.Errorf("Expected log to contain 'formatted message', got: %s", buf.String())
	}
}

// TestFastLogging tests the new fast logging methods
func TestFastLogging(t *testing.T) {
	var buf bytes.Buffer
	handler := &testHandler{buf: &buf}

	logger := NewLogger()
	logger.SetHandler(handler)
	logger.SetLevel(InfoLevel)

	// Test fast logging methods
	logger.InfoFast("fast message")

	if !strings.Contains(buf.String(), "fast message") {
		t.Errorf("Expected log to contain 'fast message', got: %s", buf.String())
	}

	// Test that fast logging respects log levels
	buf.Reset()
	logger.DebugFast("debug fast message")
	if strings.Contains(buf.String(), "debug fast message") {
		t.Error("Debug fast message should not be logged at Info level")
	}
}

// TestLogLevels tests log level filtering
func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	handler := &testHandler{buf: &buf}

	logger := NewLogger()
	logger.SetHandler(handler)

	// Test Info level (default)
	logger.SetLevel(InfoLevel)

	buf.Reset()
	logger.Debug("debug message")
	if strings.Contains(buf.String(), "debug message") {
		t.Error("Debug message should not be logged at Info level")
	}

	buf.Reset()
	logger.Info("info message")
	if !strings.Contains(buf.String(), "info message") {
		t.Error("Info message should be logged at Info level")
	}

	buf.Reset()
	logger.Warn("warn message")
	if !strings.Contains(buf.String(), "warn message") {
		t.Error("Warn message should be logged at Info level")
	}
}

// TestWithFields tests structured logging with fields
func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	handler := &testHandler{buf: &buf}
	handler.setFormatter(NewJSONFormatter())

	logger := NewLogger()
	logger.SetHandler(handler)
	logger.SetFormatter(NewJSONFormatter())

	// Test with fields
	fields := Fields{
		"user_id": 123,
		"action":  "login",
	}

	logger.WithFields(fields).Info("user action")

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if logEntry["user_id"] != float64(123) {
		t.Errorf("Expected user_id to be 123, got: %v", logEntry["user_id"])
	}

	if logEntry["action"] != "login" {
		t.Errorf("Expected action to be 'login', got: %v", logEntry["action"])
	}
}

// TestContext tests context-based logging
func TestContext(t *testing.T) {
	var buf bytes.Buffer
	handler := &testHandler{buf: &buf}
	handler.setFormatter(NewJSONFormatter())

	logger := NewLogger()
	logger.SetHandler(handler)
	logger.SetFormatter(NewJSONFormatter())

	// Test context with fields
	ctx := context.Background()
	ctx = WithFields(ctx, Fields{"request_id": "abc123"})

	contextLogger := FromContext(ctx)
	contextLogger.SetHandler(handler)
	contextLogger.SetFormatter(NewJSONFormatter())

	contextLogger.Info("request processed")

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if logEntry["request_id"] != "abc123" {
		t.Errorf("Expected request_id to be 'abc123', got: %v", logEntry["request_id"])
	}
}

// TestFileHandler tests file logging
func TestFileHandler(t *testing.T) {
	filename := "test.log"
	defer os.Remove(filename)

	handler, err := NewFileHandler(filename)
	if err != nil {
		t.Fatalf("Failed to create file handler: %v", err)
	}

	logger := NewLogger()
	logger.SetHandler(handler)

	logger.Info("test file logging")

	// Read file content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "test file logging") {
		t.Errorf("Expected log file to contain 'test file logging', got: %s", string(content))
	}
}

// TestRotatingFileHandler tests rotating file logging
func TestRotatingFileHandler(t *testing.T) {
	filename := "test_rotate.log"
	defer func() {
		os.Remove(filename)
		os.Remove(filename + ".1")
		os.Remove(filename + ".2")
	}()

	// Create rotating file handler with small max size
	handler, err := NewRotatingFileHandler(filename, 100, 2)
	if err != nil {
		t.Fatalf("Failed to create rotating file handler: %v", err)
	}

	logger := NewLogger()
	logger.SetHandler(handler)

	// Write enough logs to trigger rotation
	for i := 0; i < 10; i++ {
		logger.Info("test rotating file logging message", i)
	}

	// Check that files were created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Expected current log file to exist")
	}

	if _, err := os.Stat(filename + ".1"); os.IsNotExist(err) {
		t.Error("Expected rotated log file to exist")
	}
}

// TestHTTPHandler tests HTTP logging
func TestHTTPHandler(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	handler := NewHTTPHandler(server.URL)
	logger := NewLogger()
	logger.SetHandler(handler)

	// Test HTTP logging
	logger.Info("test HTTP logging")
	// Note: HTTP logging is asynchronous, so we don't check for errors here
}

// TestAsyncHandler tests async logging
func TestAsyncHandler(t *testing.T) {
	var buf bytes.Buffer
	baseHandler := &testHandler{buf: &buf}

	asyncHandler := NewAsyncHandler(baseHandler, 10, 2)
	logger := NewLogger()
	logger.SetHandler(asyncHandler)

	// Test async logging
	for i := 0; i < 5; i++ {
		logger.Info("async message", i)
	}

	// Give more time for async processing
	time.Sleep(500 * time.Millisecond)

	// Check if any logs were processed
	if buf.Len() == 0 {
		t.Error("Expected async logs to be processed")
	}

	// Stop the async handler
	if stopHandler, ok := asyncHandler.(*AsyncHandler); ok {
		stopHandler.Stop()
	}
}

// TestSamplingHandler tests sampling logging
func TestSamplingHandler(t *testing.T) {
	var buf bytes.Buffer
	baseHandler := &testHandler{buf: &buf}

	// Create sampling handler with 20% rate (more reliable for testing)
	samplingHandler := NewSamplingHandler(baseHandler, 0.2)
	logger := NewLogger()
	logger.SetHandler(samplingHandler)

	// Log many messages with varying lengths to ensure sampling works
	for i := 0; i < 50; i++ {
		logger.Info("sampled message", i)
	}

	// Check that some messages were logged (not all due to sampling)
	logCount := strings.Count(buf.String(), "sampled message")
	if logCount == 0 {
		t.Error("Expected some sampled messages to be logged")
	}
	// With 20% sampling, we expect roughly 10 logs out of 50
	// Allow some variance but ensure it's not all logs
	if logCount >= 50 {
		t.Errorf("Expected sampling to reduce log count, got %d out of 50", logCount)
	}
}

// TestMultiHandler tests multiple handlers
func TestMultiHandler(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	handler1 := &testHandler{buf: &buf1}
	handler2 := &testHandler{buf: &buf2}

	multiHandler := NewMultiHandler(handler1, handler2)

	logger := NewLogger()
	logger.SetHandler(multiHandler)

	logger.Info("multi handler test")

	if !strings.Contains(buf1.String(), "multi handler test") {
		t.Error("First handler should contain the message")
	}

	if !strings.Contains(buf2.String(), "multi handler test") {
		t.Error("Second handler should contain the message")
	}
}

// TestTextFormatter tests text formatting
func TestTextFormatter(t *testing.T) {
	formatter := NewTextFormatter()

	entry := &Entry{
		Level:   InfoLevel,
		Message: "test message",
		Time:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Fields: Fields{
			"key": "value",
		},
	}

	formatted, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Failed to format entry: %v", err)
	}

	formattedStr := string(formatted)
	if !strings.Contains(formattedStr, "test message") {
		t.Errorf("Expected formatted string to contain 'test message', got: %s", formattedStr)
	}

	if !strings.Contains(formattedStr, "key=value") {
		t.Errorf("Expected formatted string to contain 'key=value', got: %s", formattedStr)
	}
}

// TestJSONFormatter tests JSON formatting
func TestJSONFormatter(t *testing.T) {
	formatter := NewJSONFormatter()

	entry := &Entry{
		Level:   InfoLevel,
		Message: "test message",
		Time:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Fields: Fields{
			"key": "value",
		},
	}

	formatted, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Failed to format entry: %v", err)
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal(formatted, &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if logEntry["message"] != "test message" {
		t.Errorf("Expected message to be 'test message', got: %v", logEntry["message"])
	}

	if logEntry["level"] != "info" {
		t.Errorf("Expected level to be 'info', got: %v", logEntry["level"])
	}

	if logEntry["key"] != "value" {
		t.Errorf("Expected key to be 'value', got: %v", logEntry["key"])
	}
}

// testHandler is a test handler that writes to a buffer
type testHandler struct {
	buf       *bytes.Buffer
	formatter Formatter
}

func (h *testHandler) Handle(entry *Entry) error {
	formatter := h.formatter
	if formatter == nil {
		formatter = NewTextFormatter()
	}

	formatted, err := formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = h.buf.Write(formatted)
	return err
}

// setFormatter sets the formatter for the test handler
func (h *testHandler) setFormatter(formatter Formatter) {
	h.formatter = formatter
}

// BenchmarkLogger benchmarks the logger performance
func BenchmarkLogger(b *testing.B) {
	logger := NewLogger()
	logger.SetLevel(InfoLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}

// BenchmarkFastLogger benchmarks the fast logger performance
func BenchmarkFastLogger(b *testing.B) {
	logger := NewLogger()
	logger.SetLevel(InfoLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.InfoFast("benchmark fast message")
	}
}

// BenchmarkLoggerWithFields benchmarks logging with fields
func BenchmarkLoggerWithFields(b *testing.B) {
	logger := NewLogger()
	logger.SetLevel(InfoLevel)

	fields := Fields{
		"user_id": 123,
		"action":  "benchmark",
		"count":   b.N,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(fields).Info("benchmark message with fields")
	}
}

// BenchmarkAsyncLogger benchmarks async logging
func BenchmarkAsyncLogger(b *testing.B) {
	var buf bytes.Buffer
	baseHandler := &testHandler{buf: &buf}
	asyncHandler := NewAsyncHandler(baseHandler, 1000, 4)
	defer asyncHandler.(*AsyncHandler).Stop()

	logger := NewLogger()
	logger.SetHandler(asyncHandler)
	logger.SetLevel(InfoLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark async message")
	}
}
