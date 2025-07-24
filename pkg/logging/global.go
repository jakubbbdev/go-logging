package logging

import (
	"sync"
)

var (
	// globalLogger is the singleton logger instance
	globalLogger Logger
	globalMu     sync.RWMutex
	globalOnce   sync.Once
)

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger Logger) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() Logger {
	globalMu.RLock()
	defer globalMu.RUnlock()

	if globalLogger == nil {
		// Initialize default logger if none set
		globalOnce.Do(func() {
			globalLogger = NewLogger(
				WithLevel(InfoLevel),
				WithFormatter(NewTextFormatter()),
				WithHandler(NewConsoleHandler()),
			)
		})
	}

	return globalLogger
}

// InitGlobalLogger initializes the global logger from configuration
func InitGlobalLogger(config *Config) error {
	logger, err := config.ToLogger()
	if err != nil {
		return err
	}

	SetGlobalLogger(logger)
	return nil
}

// Global logging functions using the global logger

// Debug logs a debug message using the global logger
func Debug(args ...interface{}) {
	GetGlobalLogger().Debug(args...)
}

// Info logs an info message using the global logger
func Info(args ...interface{}) {
	GetGlobalLogger().Info(args...)
}

// Warn logs a warning message using the global logger
func Warn(args ...interface{}) {
	GetGlobalLogger().Warn(args...)
}

// Error logs an error message using the global logger
func Error(args ...interface{}) {
	GetGlobalLogger().Error(args...)
}

// Fatal logs a fatal message using the global logger
func Fatal(args ...interface{}) {
	GetGlobalLogger().Fatal(args...)
}

// Panic logs a panic message using the global logger
func Panic(args ...interface{}) {
	GetGlobalLogger().Panic(args...)
}

// Debugf logs a formatted debug message using the global logger
func Debugf(format string, args ...interface{}) {
	GetGlobalLogger().Debugf(format, args...)
}

// Infof logs a formatted info message using the global logger
func Infof(format string, args ...interface{}) {
	GetGlobalLogger().Infof(format, args...)
}

// Warnf logs a formatted warning message using the global logger
func Warnf(format string, args ...interface{}) {
	GetGlobalLogger().Warnf(format, args...)
}

// Errorf logs a formatted error message using the global logger
func Errorf(format string, args ...interface{}) {
	GetGlobalLogger().Errorf(format, args...)
}

// Fatalf logs a formatted fatal message using the global logger
func Fatalf(format string, args ...interface{}) {
	GetGlobalLogger().Fatalf(format, args...)
}

// Panicf logs a formatted panic message using the global logger
func Panicf(format string, args ...interface{}) {
	GetGlobalLogger().Panicf(format, args...)
}

// WithGlobalFields returns a logger with the given fields using the global logger
func WithGlobalFields(fields Fields) Logger {
	return GetGlobalLogger().WithFields(fields)
}

// SetLevel sets the log level for the global logger
func SetLevel(level Level) {
	GetGlobalLogger().SetLevel(level)
}

// SetHandler sets the handler for the global logger
func SetHandler(handler Handler) {
	GetGlobalLogger().SetHandler(handler)
}

// SetFormatter sets the formatter for the global logger
func SetFormatter(formatter Formatter) {
	GetGlobalLogger().SetFormatter(formatter)
}

// AddHook adds a hook to the global logger
func AddHook(hook Hook) {
	GetGlobalLogger().AddHook(hook)
}
