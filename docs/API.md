# Go Logging Library API Documentation

## Overview

The Go Logging Library provides a flexible, structured logging solution for Go applications. It supports multiple log levels, output formats, and handlers while maintaining high performance and thread safety.

## Core Types

### Logger Interface

The main interface for logging operations:

```go
type Logger interface {
    // Basic logging methods
    Debug(args ...interface{})
    Info(args ...interface{})
    Warn(args ...interface{})
    Error(args ...interface{})
    Fatal(args ...interface{})
    Panic(args ...interface{})

    // Formatted logging methods
    Debugf(format string, args ...interface{})
    Infof(format string, args ...interface{})
    Warnf(format string, args ...interface{})
    Errorf(format string, args ...interface{})
    Fatalf(format string, args ...interface{})
    Panicf(format string, args ...interface{})

    // Context and field management
    WithFields(fields Fields) Logger
    WithContext(ctx context.Context) Logger
    
    // Configuration
    SetLevel(level Level)
    SetHandler(handler Handler)
    SetFormatter(formatter Formatter)
}
```

### Log Levels

```go
const (
    DebugLevel Level = iota  // Most verbose
    InfoLevel                // Default level
    WarnLevel                // Warnings and above
    ErrorLevel               // Errors only
    FatalLevel               // Fatal errors only
    PanicLevel               // Panic only
)
```

### Fields

```go
type Fields map[string]interface{}
```

Fields represent key-value pairs for structured logging.

### Entry

```go
type Entry struct {
    Level     Level
    Message   string
    Fields    Fields
    Time      time.Time
    Caller    string
    Context   context.Context
}
```

## Handlers

### ConsoleHandler

Outputs logs to the console (stdout).

```go
handler := logging.NewConsoleHandler()
```

### FileHandler

Outputs logs to a file.

```go
handler, err := logging.NewFileHandler("app.log")
if err != nil {
    // Handle error
}
defer handler.(*logging.FileHandler).Close()
```

### MultiHandler

Combines multiple handlers.

```go
multiHandler := logging.NewMultiHandler(
    logging.NewConsoleHandler(),
    fileHandler,
)
```

## Formatters

### TextFormatter

Formats logs as human-readable text with optional colors.

```go
formatter := logging.NewTextFormatter()
formatter.UseColors = true  // Enable colored output
formatter.Timestamp = true  // Include timestamps
```

### JSONFormatter

Formats logs as JSON for machine processing.

```go
formatter := logging.NewJSONFormatter()
formatter.SetPrettyPrint(true)  // Enable pretty printing
```

## Context Support

### WithFields

Add fields to a context for automatic inclusion in logs.

```go
ctx := logging.WithFields(context.Background(), logging.Fields{
    "request_id": "abc123",
    "user_id":    456,
})
```

### FromContext

Get a logger with context fields automatically included.

```go
logger := logging.FromContext(ctx)
logger.Info("Processing request")  // Includes request_id and user_id
```

### WithLogger

Add a logger to a context.

```go
ctx := logging.WithLogger(ctx, logger)
```

## Usage Examples

### Basic Usage

```go
logger := logging.NewLogger()
logger.SetLevel(logging.InfoLevel)

logger.Info("Application started")
logger.Warn("This is a warning")
logger.Error("An error occurred")
```

### Structured Logging

```go
logger := logging.NewLogger()
logger.WithFields(logging.Fields{
    "user_id": 123,
    "action":  "login",
    "ip":      "192.168.1.1",
}).Info("User logged in successfully")
```

### JSON Output

```go
logger := logging.NewLogger()
logger.SetFormatter(logging.NewJSONFormatter())
logger.Info("This will be logged as JSON")
```

### File Logging

```go
fileHandler, err := logging.NewFileHandler("app.log")
if err != nil {
    log.Fatal(err)
}
defer fileHandler.(*logging.FileHandler).Close()

logger := logging.NewLogger()
logger.SetHandler(fileHandler)
logger.Info("This will be logged to file")
```

### Multiple Outputs

```go
consoleHandler := logging.NewConsoleHandler()
fileHandler, _ := logging.NewFileHandler("app.log")

multiHandler := logging.NewMultiHandler(consoleHandler, fileHandler)
logger := logging.NewLogger()
logger.SetHandler(multiHandler)

logger.Info("This will be logged to both console and file")
```

### Context-Aware Logging

```go
// Add fields to context
ctx := logging.WithFields(context.Background(), logging.Fields{
    "request_id": "abc123",
})

// Get logger with context fields
logger := logging.FromContext(ctx)
logger.Info("Processing request")  // Includes request_id automatically
```

### Custom Handler

```go
type CustomHandler struct{}

func (h *CustomHandler) Handle(entry *logging.Entry) error {
    // Custom handling logic
    fmt.Printf("Custom: %s\n", entry.Message)
    return nil
}

logger := logging.NewLogger()
logger.SetHandler(&CustomHandler{})
```

### Custom Formatter

```go
type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logging.Entry) ([]byte, error) {
    return []byte(fmt.Sprintf("CUSTOM: %s - %s", entry.Level, entry.Message)), nil
}

logger := logging.NewLogger()
logger.SetFormatter(&CustomFormatter{})
```

## Performance Considerations

- The library is designed for zero-allocation logging in common use cases
- Use appropriate log levels to avoid unnecessary processing
- Consider using structured logging for better performance in production
- File handlers should be properly closed to avoid resource leaks

## Thread Safety

All logger operations are thread-safe and can be used concurrently from multiple goroutines.

## Error Handling

- File handlers return errors that should be checked
- Formatters may return errors that are handled internally
- Fatal and Panic methods will terminate the program or panic respectively 