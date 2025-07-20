# Go Logging Library

A modern, flexible, and feature-rich logging library for Go applications. This library provides structured logging with multiple output formats, log levels, and customizable handlers.

## Features

- **Multiple Log Levels**: Debug, Info, Warn, Error, Fatal, Panic
- **Structured Logging**: JSON and text output formats
- **Customizable Handlers**: Console, file, and custom handlers
- **Color Support**: Colored output for better readability
- **Context Support**: Add fields and context to log entries
- **Performance Optimized**: Zero-allocation logging for high-performance applications
- **Thread Safe**: Safe for concurrent use

## Installation

```bash
go get github.com/jakubbbdev/go-logging
```

## Quick Start

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

## Configuration

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

## Advanced Usage

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

## Performance

The library is designed for high-performance applications:

- Zero-allocation logging for common use cases
- Efficient field handling
- Minimal memory footprint
- Fast JSON serialization

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 