# Go Logging Library

A modern, flexible, and feature-rich logging library for Go applications. This library provides structured logging with multiple output formats, log levels, and customizable handlers with **performance optimizations** and **advanced features**.

## 🚀 Features

- **Multiple Log Levels**: Debug, Info, Warn, Error, Fatal, Panic
- **Structured Logging**: JSON and text output formats
- **Customizable Handlers**: Console, file, rotating file, HTTP, and custom handlers
- **Performance Optimized**: Zero-allocation logging with entry pooling
- **Async Logging**: Non-blocking logging with worker pools
- **Log Sampling**: Reduce log volume with configurable sampling rates
- **Color Support**: Colored output for better readability
- **Context Support**: Add fields and context to log entries
- **Hooks**: Add custom hooks for metrics, audit, etc.
- **Thread Safe**: Safe for concurrent use

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

## 📚 Examples

Check out the examples in the `cmd/examples/` directory:

- **Basic Example**: `go run cmd/examples/basic/main.go`
- **Web Server Example**: `