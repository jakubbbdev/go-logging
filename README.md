# Go Logging Library

A modern, flexible, and **enterprise-ready** logging library for Go applications. This library provides structured logging with multiple output formats, log levels, customizable handlers, **performance optimizations**, **security features**, **metrics collection**, **distributed tracing**, and **configuration management**.

## 🚀 Features

### Core Features
- **Multiple Log Levels**: Debug, Info, Warn, Error, Fatal, Panic + Custom Levels
- **Structured Logging**: JSON and text output formats
- **Customizable Handlers**: Console, file, rotating file, HTTP, async, and custom handlers
- **Performance Optimized**: Zero-allocation logging with entry pooling
- **Thread Safe**: Safe for concurrent use

### Enterprise Features
- **📋 Configuration Management**: YAML, JSON, and environment variable support
- **🌍 Global Logger**: Singleton pattern with convenient helper functions
- **📊 Metrics Collection**: Prometheus-compatible metrics and HTTP endpoint
- **🔍 Distributed Tracing**: Request IDs, trace IDs, span IDs, and user context
- **🔒 Security & PII Detection**: Automatic sanitization of sensitive data
- **⚡ Performance Features**: Async logging, sampling, and pooling
- **🎨 Rich Formatting**: Colors, emojis, custom timestamps, field ordering
- **🔗 Context Support**: Add fields and context to log entries
- **🪝 Hooks System**: Custom hooks for metrics, audit, external integrations

## 📦 Installation

```bash
go get github.com/jakubbbdev/go-logging
```

## 🏗️ Project Structure

```
go-logging/
├── pkg/logging/          # Main library package
│   ├── logging.go        # Package entry point
│   ├── logger.go         # Core logger interface and implementation
│   ├── handlers.go       # Console, file, rotating file, HTTP, async handlers
│   ├── formatters.go     # Text and JSON formatters
│   ├── context.go        # Context support and utilities
│   └── logger_test.go    # Test files
├── cmd/examples/         # Example applications
│   ├── basic/            # Basic usage examples
│   ├── web/              # Web server examples
│   └── advanced/         # Advanced features demo
├── docs/                 # Documentation
├── Makefile              # Build and development tools
├── LICENSE               # MIT License
└── README.md             # This file
```

## 🚀 Quick Start (Functional Options)

```go
package main

import (
    "github.com/jakubbbdev/go-logging/pkg/logging"
)

func main() {
    logger := logging.NewLogger(
        logging.WithLevel(logging.DebugLevel),
        logging.WithFormatter(logging.NewJSONFormatter()),
        logging.WithHandler(logging.NewRotatingFileHandler("app.log", 10*1024*1024, 5)),
        logging.WithDefaultFields(logging.Fields{"service": "api"}),
        logging.WithHook(func(entry *logging.Entry) {
            // Custom hook: z.B. Metrics, Audit, Sentry, ...
        }),
    )

    logger.Info("Application started!")
    logger.WithFields(logging.Fields{"user_id": 42}).Warn("User warning!")
}
```

## 📝 GoDoc Example

```go
// ExampleLogger demonstrates the new functional options API and hooks.
func ExampleLogger() {
    logger := logging.NewLogger(
        logging.WithLevel(logging.InfoLevel),
        logging.WithFormatter(logging.NewTextFormatter()),
        logging.WithDefaultFields(logging.Fields{"env": "dev"}),
        logging.WithHook(func(entry *logging.Entry) {
            if entry.Level == logging.ErrorLevel {
                // Send to Sentry, Prometheus, etc.
            }
        }),
    )
    logger.Info("Hello, world!")
    // Output: [INFO] Hello, world! {env=dev}
}
```

## 🎨 Custom Colors, Timestamp & Field Order

```go
import "github.com/fatih/color"

logger := logging.NewLogger(
    logging.WithFormatter(logging.NewTextFormatter(
        logging.WithTextFormatterColors(map[logging.Level]*color.Color{
            logging.InfoLevel:  color.New(color.FgHiBlue, color.Bold),
            logging.ErrorLevel: color.New(color.FgHiRed, color.Bold, color.BgBlack),
        }),
        logging.WithTextFormatterTimestampFormat("15:04:05"),
        logging.WithTextFormatterLevelPadding(7),
        logging.WithTextFormatterPrefix(logging.ErrorLevel, "🔥 "),
        logging.WithTextFormatterSuffix(logging.InfoLevel, " ℹ️"),
        logging.WithTextFormatterFieldOrder([]string{"user_id", "action", "ip"}),
    )),
)

logger.WithFields(logging.Fields{
    "user_id": 123,
    "action":  "login",
    "ip":      "192.168.1.1",
}).Error("Custom colored error!")
```

## 🧑‍💻 Caller, Stacktrace, Emojis & Field Masking

```go
logger := logging.NewLogger(
    logging.WithCaller(true),
    logging.WithStacktrace(true),
    logging.WithFormatter(logging.NewTextFormatter(
        logging.WithTextFormatterEmojis(map[logging.Level]string{
            logging.DebugLevel: "🐛 ",
            logging.InfoLevel:  "ℹ️ ",
            logging.WarnLevel:  "⚠️ ",
            logging.ErrorLevel: "❌ ",
        }),
        logging.WithTextFormatterFieldMasking([]string{"password", "token"}, "****"),
    )),
)

logger.WithFields(logging.Fields{
    "user_id": 123,
    "password": "supersecret",
    "token":    "abcdefg",
}).Error("Login failed!")
```

## ⚡ Performance Features

### Fast Logging Methods
```go
// Zero-allocation logging for performance-critical applications
logger.InfoFast("fast message")
logger.DebugFast("debug message")
logger.WarnFast("warning message")
logger.ErrorFast("error message")
```

### Entry Pooling
The library uses object pooling to reduce memory allocations:
```go
// Automatic entry pooling - no manual configuration needed
logger.Info("This uses pooled entries automatically")
```

## 🔧 Advanced Handlers

### Rotating File Handler
```go
// Automatically rotate log files when they reach a certain size
rotatingHandler, err := logging.NewRotatingFileHandler("app.log", 10*1024*1024, 5) // 10MB, 5 files
if err != nil {
    log.Fatal(err)
}
defer rotatingHandler.(*logging.RotatingFileHandler).Close()

logger.SetHandler(rotatingHandler)
```

### Async Handler
```go
// Non-blocking logging with worker pools
baseHandler := logging.NewConsoleHandler()
asyncHandler := logging.NewAsyncHandler(baseHandler, 1000, 4) // buffer size, workers
defer asyncHandler.(*logging.AsyncHandler).Stop()

logger.SetHandler(asyncHandler)
```

### HTTP Handler
```go
// Send logs to remote servers via HTTP
httpHandler := logging.NewHTTPHandler("https://logs.example.com/api/logs")
logger.SetHandler(httpHandler)
```

### Sampling Handler
```go
// Reduce log volume with sampling
baseHandler := logging.NewConsoleHandler()
samplingHandler := logging.NewSamplingHandler(baseHandler, 0.1) // 10% of logs
logger.SetHandler(samplingHandler)
```

### Multi Handler
```go
// Log to multiple destinations simultaneously
multiHandler := logging.NewMultiHandler(
    logging.NewConsoleHandler(),
    fileHandler,
    rotatingHandler,
)
logger.SetHandler(multiHandler)
```

## ⚙️ Configuration

### Log Levels

```go
logger.SetLevel(logging.DebugLevel)  // Most verbose
logger.SetLevel(logging.InfoLevel)   // Default
logger.SetLevel(logging.WarnLevel)   // Warnings and above
logger.SetLevel(logging.ErrorLevel)  // Errors only
logger.SetLevel(logging.FatalLevel)  // Fatal errors only
logger.SetLevel(logging.PanicLevel)  // Panic only
```

### Output Formats

```go
// JSON format
logger.SetFormatter(logging.NewJSONFormatter())

// Text format (default)
logger.SetFormatter(logging.NewTextFormatter())

// Custom formatter
logger.SetFormatter(&MyCustomFormatter{})
```

## 🔧 Advanced Usage

### Production Setup
```go
// Production-ready logging setup
logger := logging.NewLogger(
    logging.WithLevel(logging.InfoLevel),
    logging.WithHandler(logging.NewRotatingFileHandler("production.log", 10*1024*1024, 5)),
    logging.WithFormatter(logging.NewJSONFormatter()),
)
```

### Custom Fields

```go
logger := logging.NewLogger().WithFields(logging.Fields{
    "service": "api",
    "version": "1.0.0",
})

logger.Info("Service initialized")
// Output: {"level":"info","message":"Service initialized","service":"api","version":"1.0.0"}
```

### Context Logging

```go
ctx := context.WithValue(context.Background(), "request_id", "abc123")
logger := logging.FromContext(ctx)

logger.Info("Processing request")
// Output: {"level":"info","message":"Processing request","request_id":"abc123"}
```

### Custom Handlers

```go
type CustomHandler struct{}

func (h *CustomHandler) Handle(entry *logging.Entry) error {
    // Custom handling logic
    return nil
}

logger.SetHandler(&CustomHandler{})
```

## 🆕 Enterprise Features

### 📋 Configuration Management

```go
// From YAML file
config, err := logging.LoadConfigFromFile("config.yaml")
if err != nil {
    log.Fatal(err)
}

logger, err := config.ToLogger()
if err != nil {
    log.Fatal(err)
}

// From environment variables
config := logging.LoadConfigFromEnv()
logger, err := config.ToLogger()

// Initialize global logger from config
err = logging.InitGlobalLogger(config)
```

**Sample YAML Configuration:**
```yaml
level: "info"
format: "json"
output: "file"
include_caller: true
include_stack: true

default_fields:
  service: "my-app"
  version: "1.0.0"

file:
  path: "logs/app.log"
  rotate: true
  max_size: 10485760  # 10MB
  max_files: 5

metrics:
  enabled: true
  port: 8080
  path: "/metrics"
```

### 🌍 Global Logger

```go
// Set global logger
globalLogger := logging.NewLogger(
    logging.WithLevel(logging.InfoLevel),
    logging.WithDefaultFields(logging.Fields{"app": "myapp"}),
)
logging.SetGlobalLogger(globalLogger)

// Use global functions anywhere in your app
logging.Info("Application started")
logging.WithGlobalFields(logging.Fields{"module": "auth"}).Error("Auth failed")
logging.Errorf("Error: %v", err)
```

### 📊 Metrics Collection

```go
// Create metrics collector
metrics := logging.NewMetricsCollector()

// Add metrics hook to logger
logger := logging.NewLogger(
    logging.WithHook(logging.NewMetricsHook(metrics)),
)

// Start metrics HTTP server
go metrics.StartMetricsServer(":8080", "/metrics")

// Get statistics
stats := metrics.GetStats()
fmt.Printf("Total logs: %d\n", getTotalLogs(stats.LogCount))
```

**Metrics Endpoint Output:**
```
# HELP logging_total Total number of log entries by level
# TYPE logging_total counter
logging_total{level="info"} 42
logging_total{level="error"} 3

# HELP logging_duration_seconds Total time spent logging by level
# TYPE logging_duration_seconds counter
logging_duration_seconds{level="info"} 0.002341
```

### 🔍 Distributed Tracing

```go
// Create trace context
traceCtx := logging.NewTraceContext().
    WithUserID("user123").
    WithSessionID("session456")

ctx := logging.WithTraceContext(context.Background(), traceCtx)

// Log with tracing
logger := logging.NewLogger(
    logging.WithHook(logging.NewTracingHook()),
)

tracedLogger := logger.WithTrace(ctx)
tracedLogger.Info("Request processed")
// Output: {..., "trace_id": "abc123", "user_id": "user123", "request_id": "req456"}
```

### 🔒 Security & PII Detection

```go
// Create PII detector
piiDetector := logging.NewPIIDetector()

// Use security hook
logger := logging.NewLogger(
    logging.WithHook(logging.NewSecurityHook(piiDetector)),
)

// Or wrap formatter
secureFormatter := logging.NewSecurityFormatter(
    logging.NewJSONFormatter(), 
    piiDetector,
)

// PII data is automatically sanitized
logger.WithFields(logging.Fields{
    "email":    "user@example.com",      // -> "us**@example.com"
    "phone":    "555-123-4567",          // -> "***-***-4567"
    "password": "secret123",             // -> "[REDACTED]"
    "ssn":      "123-45-6789",          // -> "***-**-6789"
}).Info("User data logged safely")
```

### 🏢 Enterprise Setup Example

```go
// Full enterprise configuration
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

// Start metrics server
go metrics.StartMetricsServer(":8080", "/metrics")

// Set as global logger
logging.SetGlobalLogger(logger)
```

## 🆕 Eigene Logging-Levels

```go
// Registriere ein neues Level
var AuditLevel = logging.RegisterLevel("audit", 25)

logger := logging.NewLogger(
    logging.WithLevel(logging.DebugLevel),
)

logger.Log(AuditLevel, "User audit event!", 123)
logger.Logf(AuditLevel, "Audit for user %d", 123)
logger.Log(logging.RegisterLevel("trace", 5), "Trace message!")
```

## 🔧 Environment Variables

Configure the library using environment variables:

```bash
# Basic configuration
export LOG_LEVEL=debug
export LOG_FORMAT=json
export LOG_OUTPUT=file
export LOG_INCLUDE_CALLER=true
export LOG_INCLUDE_STACK=true

# Default fields (comma-separated key=value pairs)
export LOG_DEFAULT_FIELDS="service=myapp,version=1.0.0"

# File configuration
export LOG_FILE_PATH=logs/app.log
export LOG_FILE_ROTATE=true
export LOG_FILE_MAX_SIZE=10485760
export LOG_FILE_MAX_FILES=5

# Metrics configuration
export LOG_METRICS_ENABLED=true
export LOG_METRICS_PORT=8080
export LOG_METRICS_PATH=/metrics
```

## 📚 Examples

Check out the examples in the `cmd/examples/` directory:

- **Basic Example**: `go run cmd/examples/basic/main.go`
- **Web Server Example**: `go run cmd/examples/web/main.go`
- **Advanced Features**: `go run cmd/examples/advanced/main.go`
- **Modern Features**: `go run cmd/examples/modern/main.go`
- **Enterprise Setup**: `go run cmd/examples/enterprise/main.go`