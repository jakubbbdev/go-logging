package logging

import (
	"context"
	"fmt"
	"time"
)

// OTelConfig represents OpenTelemetry configuration
type OTelConfig struct {
	ServiceName    string
	ServiceVersion string
	TraceExporter  string // "jaeger", "otlp", "console"
	TraceEndpoint  string
	MetricExporter string // "prometheus", "otlp", "console"
	MetricEndpoint string
	Enabled        bool
}

// OTelSpan represents a span for tracing
type OTelSpan struct {
	TraceID   string
	SpanID    string
	ParentID  string
	Operation string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Tags      map[string]interface{}
	Logs      []OTelLog
	Status    SpanStatus
	Error     error
}

// OTelLog represents a log entry within a span
type OTelLog struct {
	Timestamp time.Time
	Level     Level
	Message   string
	Fields    Fields
}

// SpanStatus represents the status of a span
type SpanStatus int

const (
	SpanStatusOK SpanStatus = iota
	SpanStatusError
	SpanStatusTimeout
)

// OTelTracer handles OpenTelemetry-style tracing
type OTelTracer struct {
	serviceName string
	spans       map[string]*OTelSpan
	hooks       []OTelHook
}

// OTelHook is called when span events occur
type OTelHook func(span *OTelSpan, event string)

// NewOTelTracer creates a new OpenTelemetry-compatible tracer
func NewOTelTracer(serviceName string) *OTelTracer {
	return &OTelTracer{
		serviceName: serviceName,
		spans:       make(map[string]*OTelSpan),
		hooks:       make([]OTelHook, 0),
	}
}

// AddHook adds a hook to the tracer
func (t *OTelTracer) AddHook(hook OTelHook) {
	t.hooks = append(t.hooks, hook)
}

// StartSpan starts a new span
func (t *OTelTracer) StartSpan(ctx context.Context, operation string) (context.Context, *OTelSpan) {
	span := &OTelSpan{
		TraceID:   GenerateTraceID(),
		SpanID:    GenerateSpanID(),
		Operation: operation,
		StartTime: time.Now(),
		Tags:      make(map[string]interface{}),
		Logs:      make([]OTelLog, 0),
		Status:    SpanStatusOK,
	}

	// Check for parent span in context
	if parentSpan := SpanFromContext(ctx); parentSpan != nil {
		span.ParentID = parentSpan.SpanID
		span.TraceID = parentSpan.TraceID
	}

	t.spans[span.SpanID] = span

	// Call hooks
	for _, hook := range t.hooks {
		hook(span, "start")
	}

	// Add span to context
	ctx = WithSpan(ctx, span)
	return ctx, span
}

// FinishSpan finishes a span
func (t *OTelTracer) FinishSpan(span *OTelSpan) {
	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)

	// Call hooks
	for _, hook := range t.hooks {
		hook(span, "finish")
	}
}

// LogToSpan adds a log entry to a span
func (t *OTelTracer) LogToSpan(span *OTelSpan, level Level, message string, fields Fields) {
	if span == nil {
		return
	}

	log := OTelLog{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}

	span.Logs = append(span.Logs, log)

	// Call hooks
	for _, hook := range t.hooks {
		hook(span, "log")
	}
}

// SetSpanTag sets a tag on a span
func (t *OTelTracer) SetSpanTag(span *OTelSpan, key string, value interface{}) {
	if span == nil {
		return
	}
	span.Tags[key] = value
}

// SetSpanError sets an error on a span
func (t *OTelTracer) SetSpanError(span *OTelSpan, err error) {
	if span == nil {
		return
	}
	span.Error = err
	span.Status = SpanStatusError
	span.Tags["error"] = true
	span.Tags["error.message"] = err.Error()
}

// Context keys for spans
type spanContextKey struct{}

// WithSpan adds a span to context
func WithSpan(ctx context.Context, span *OTelSpan) context.Context {
	return context.WithValue(ctx, spanContextKey{}, span)
}

// SpanFromContext extracts a span from context
func SpanFromContext(ctx context.Context) *OTelSpan {
	if span, ok := ctx.Value(spanContextKey{}).(*OTelSpan); ok {
		return span
	}
	return nil
}

// OTelHandler wraps another handler and adds OpenTelemetry integration
type OTelHandler struct {
	handler Handler
	tracer  *OTelTracer
}

// NewOTelHandler creates a new OpenTelemetry-integrated handler
func NewOTelHandler(handler Handler, tracer *OTelTracer) Handler {
	return &OTelHandler{
		handler: handler,
		tracer:  tracer,
	}
}

// Handle implements the Handler interface with OTel integration
func (h *OTelHandler) Handle(entry *Entry) error {
	// Add span information to entry if available
	if entry.Context != nil {
		if span := SpanFromContext(entry.Context); span != nil {
			if entry.Fields == nil {
				entry.Fields = make(Fields)
			}
			entry.Fields["otel.trace_id"] = span.TraceID
			entry.Fields["otel.span_id"] = span.SpanID
			if span.ParentID != "" {
				entry.Fields["otel.parent_id"] = span.ParentID
			}
			entry.Fields["otel.operation"] = span.Operation

			// Log to span
			h.tracer.LogToSpan(span, entry.Level, entry.Message, entry.Fields)

			// Set error on span if it's an error log
			if entry.Level == ErrorLevel || entry.Level == FatalLevel || entry.Level == PanicLevel {
				h.tracer.SetSpanError(span, fmt.Errorf("%s", entry.Message))
			}
		}
	}

	return h.handler.Handle(entry)
}

// OTelHookFactory creates hooks for OpenTelemetry integration
func NewOTelLoggingHook(tracer *OTelTracer) Hook {
	return func(entry *Entry) {
		if entry.Context != nil {
			if span := SpanFromContext(entry.Context); span != nil {
				// Add OpenTelemetry fields to log entry
				if entry.Fields == nil {
					entry.Fields = make(Fields)
				}
				entry.Fields["otel.trace_id"] = span.TraceID
				entry.Fields["otel.span_id"] = span.SpanID
				entry.Fields["otel.operation"] = span.Operation

				// Log to span
				tracer.LogToSpan(span, entry.Level, entry.Message, entry.Fields)
			}
		}
	}
}

// SpanLogger wraps a logger with span context
type SpanLogger struct {
	logger Logger
	span   *OTelSpan
	tracer *OTelTracer
}

// NewSpanLogger creates a logger that automatically adds span context
func NewSpanLogger(logger Logger, span *OTelSpan, tracer *OTelTracer) Logger {
	return &SpanLogger{
		logger: logger,
		span:   span,
		tracer: tracer,
	}
}

// Implement Logger interface for SpanLogger
func (sl *SpanLogger) Debug(args ...interface{}) {
	msg := fmt.Sprint(args...)
	sl.tracer.LogToSpan(sl.span, DebugLevel, msg, nil)
	sl.logger.Debug(args...)
}

func (sl *SpanLogger) Info(args ...interface{}) {
	msg := fmt.Sprint(args...)
	sl.tracer.LogToSpan(sl.span, InfoLevel, msg, nil)
	sl.logger.Info(args...)
}

func (sl *SpanLogger) Warn(args ...interface{}) {
	msg := fmt.Sprint(args...)
	sl.tracer.LogToSpan(sl.span, WarnLevel, msg, nil)
	sl.logger.Warn(args...)
}

func (sl *SpanLogger) Error(args ...interface{}) {
	msg := fmt.Sprint(args...)
	sl.tracer.LogToSpan(sl.span, ErrorLevel, msg, nil)
	sl.tracer.SetSpanError(sl.span, fmt.Errorf("%s", msg))
	sl.logger.Error(args...)
}

func (sl *SpanLogger) Fatal(args ...interface{}) {
	msg := fmt.Sprint(args...)
	sl.tracer.LogToSpan(sl.span, FatalLevel, msg, nil)
	sl.tracer.SetSpanError(sl.span, fmt.Errorf("%s", msg))
	sl.logger.Fatal(args...)
}

func (sl *SpanLogger) Panic(args ...interface{}) {
	msg := fmt.Sprint(args...)
	sl.tracer.LogToSpan(sl.span, PanicLevel, msg, nil)
	sl.tracer.SetSpanError(sl.span, fmt.Errorf("%s", msg))
	sl.logger.Panic(args...)
}

// Implement rest of Logger interface (simplified)
func (sl *SpanLogger) Debugf(format string, args ...interface{}) {
	sl.Debug(fmt.Sprintf(format, args...))
}
func (sl *SpanLogger) Infof(format string, args ...interface{}) {
	sl.Info(fmt.Sprintf(format, args...))
}
func (sl *SpanLogger) Warnf(format string, args ...interface{}) {
	sl.Warn(fmt.Sprintf(format, args...))
}
func (sl *SpanLogger) Errorf(format string, args ...interface{}) {
	sl.Error(fmt.Sprintf(format, args...))
}
func (sl *SpanLogger) Fatalf(format string, args ...interface{}) {
	sl.Fatal(fmt.Sprintf(format, args...))
}
func (sl *SpanLogger) Panicf(format string, args ...interface{}) {
	sl.Panic(fmt.Sprintf(format, args...))
}
func (sl *SpanLogger) DebugFast(msg string)                 { sl.Debug(msg) }
func (sl *SpanLogger) InfoFast(msg string)                  { sl.Info(msg) }
func (sl *SpanLogger) WarnFast(msg string)                  { sl.Warn(msg) }
func (sl *SpanLogger) ErrorFast(msg string)                 { sl.Error(msg) }
func (sl *SpanLogger) Log(level Level, args ...interface{}) { sl.logger.Log(level, args...) }
func (sl *SpanLogger) Logf(level Level, format string, args ...interface{}) {
	sl.logger.Logf(level, format, args...)
}
func (sl *SpanLogger) LogFast(level Level, msg string)        { sl.logger.LogFast(level, msg) }
func (sl *SpanLogger) WithFields(fields Fields) Logger        { return sl.logger.WithFields(fields) }
func (sl *SpanLogger) WithContext(ctx context.Context) Logger { return sl.logger.WithContext(ctx) }
func (sl *SpanLogger) WithTrace(ctx context.Context) Logger   { return sl.logger.WithTrace(ctx) }
func (sl *SpanLogger) SetLevel(level Level)                   { sl.logger.SetLevel(level) }
func (sl *SpanLogger) SetHandler(handler Handler)             { sl.logger.SetHandler(handler) }
func (sl *SpanLogger) SetFormatter(formatter Formatter)       { sl.logger.SetFormatter(formatter) }
func (sl *SpanLogger) AddHook(hook Hook)                      { sl.logger.AddHook(hook) }
