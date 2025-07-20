# Go Logging Library

A modern, flexible, and feature-rich logging library for Go applications. This library provides structured logging with multiple output formats, log levels, and customizable handlers.

## 🚀 Features

- **Multiple Log Levels**: Debug, Info, Warn, Error, Fatal, Panic
- **Structured Logging**: JSON and text output formats
- **Customizable Handlers**: Console, file, and custom handlers
- **Color Support**: Colored output for better readability
- **Context Support**: Add fields and context to log entries
- **Performance Optimized**: Zero-allocation logging for high-performance applications
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
│   ├── handlers.go       # Console, file, and multi handlers
│   ├── formatters.go     # Text and JSON formatters
│   └── context.go        # Context support and utilities
├── cmd/examples/         # Example applications
│   ├── basic/            # Basic usage examples
│   └── web/              # Web server examples
├── internal/tests/       # Test files
├── docs/                 # Documentation
├── Makefile              # Build and development tools
├── LICENSE               # MIT License
└── README.md             # This file
```

## 🚀 Quick Start

```go
package main

import (
    "github.com/jakubbbdev/go-logging"
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

### Handlers

```go
// Console handler (default)
logger.SetHandler(logging.NewConsoleHandler())

// File handler
fileHandler := logging.NewFileHandler("app.log")
logger.SetHandler(fileHandler)

// Multiple handlers
logger.SetHandler(logging.NewMultiHandler(
    logging.NewConsoleHandler(),
    logging.NewFileHandler("app.log"),
))
```

## 🔧 Advanced Usage

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

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...
```

## 🛠️ Development

```bash
# Build the library
make build

# Run tests
make test

# Format code
make fmt

# Run examples
make examples
```

## 📈 Performance

The library is designed for high-performance applications:

- Zero-allocation logging for common use cases
- Efficient field handling
- Minimal memory footprint
- Fast JSON serialization

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