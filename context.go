package logging

import (
	"context"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// LoggerContextKey is the key used to store logger in context
	LoggerContextKey contextKey = "logger"
	// FieldsContextKey is the key used to store fields in context
	FieldsContextKey contextKey = "fields"
)

// FromContext returns a logger from the context, or creates a new one if not found
func FromContext(ctx context.Context) Logger {
	// Get existing logger from context
	var logger Logger
	if existingLogger, ok := ctx.Value(LoggerContextKey).(Logger); ok {
		logger = existingLogger
	} else {
		logger = NewLogger()
	}

	// Get fields from context and add them to the logger
	fields := GetFieldsFromContext(ctx)
	if len(fields) > 0 {
		logger = logger.WithFields(fields)
	}

	return logger
}

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, LoggerContextKey, logger)
}

// WithFields adds fields to the context
func WithFields(ctx context.Context, fields Fields) context.Context {
	existingFields, _ := ctx.Value(FieldsContextKey).(Fields)
	if existingFields == nil {
		existingFields = make(Fields)
	}

	// Merge fields
	for k, v := range fields {
		existingFields[k] = v
	}

	return context.WithValue(ctx, FieldsContextKey, existingFields)
}

// GetFieldsFromContext returns fields from the context
func GetFieldsFromContext(ctx context.Context) Fields {
	if fields, ok := ctx.Value(FieldsContextKey).(Fields); ok {
		return fields
	}
	return make(Fields)
}

// ContextLogger is a logger that automatically includes context fields
type ContextLogger struct {
	Logger
}

// NewContextLogger creates a new context-aware logger
func NewContextLogger() *ContextLogger {
	return &ContextLogger{
		Logger: NewLogger(),
	}
}

// FromContext creates a context logger from a context
func (cl *ContextLogger) FromContext(ctx context.Context) Logger {
	fields := GetFieldsFromContext(ctx)
	if len(fields) == 0 {
		return cl.Logger
	}
	return cl.Logger.WithFields(fields)
}
