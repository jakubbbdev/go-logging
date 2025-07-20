package logging

import (
	"bytes"
	"context"
	"encoding/json"
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
