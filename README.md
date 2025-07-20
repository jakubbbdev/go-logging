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

## 🚀 Quick Start

```go
package main

import (
    "github.com/jakubbbdev/go-logging/pkg/logging"
)

func main() {
    // Create a new logger
    logger := logging.NewLogger()
    
    // Set log level
    logger.SetLevel(logging.InfoLevel)
    
    // Log messages
    logger.Info("Application started")
    logger.Warn("This is a warning message")
    logger.Error("An error occurred")
    
    // Structured logging with fields
    logger.WithFields(logging.Fields{
        "user_id": 123,
        "action":  "login",
    }).Info("User logged in successfully")
}
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
logger := logging.NewLogger()
logger.SetLevel(logging.InfoLevel)

// Create handlers
consoleHandler := logging.NewConsoleHandler()
rotatingHandler, _ := logging.NewRotatingFileHandler("production.log", 10*1024*1024, 5)
defer rotatingHandler.(*logging.RotatingFileHandler).Close()

// Create async multi handler
multiHandler := logging.NewMultiHandler(consoleHandler, rotatingHandler)
asyncHandler := logging.NewAsyncHandler(multiHandler, 1000, 8)
defer asyncHandler.(*logging.AsyncHandler).Stop()

logger.SetHandler(asyncHandler)
logger.SetFormatter(logging.NewJSONFormatter())
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
- **Web Server Example**: `go run cmd/examples/web/main.go`
- **Advanced Features**: `go run cmd/examples/advanced/main.go`

## 🧪 Testing

```bash
# Run all tests
go test ./pkg/logging

# Run tests with coverage
go test -cover ./pkg/logging

# Run benchmarks
go test -bench=. ./pkg/logging
```

## 🛠️ Development

```bash
# Build the library
go build ./pkg/logging

# Run tests
go test ./pkg/logging

# Format code
go fmt ./pkg/logging ./cmd/examples

# Run examples
go run cmd/examples/basic/main.go
go run cmd/examples/advanced/main.go
```

## 📈 Performance

The library is designed for high-performance applications:

- **Zero-allocation logging** for common use cases
- **Entry pooling** to reduce memory allocations
- **Async logging** for non-blocking operations
- **Efficient field handling**
- **Fast JSON serialization**
- **Configurable sampling** to reduce log volume

### Benchmark Results
```bash
# Run benchmarks to see performance metrics
go test -bench=. -benchmem ./pkg/logging
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- [GitHub Repository](https://github.com/jakubbbdev/go-logging)
- [API Documentation](docs/API.md)
- [Changelog](CHANGELOG.md) 