package logging

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// SlogHandler implements slog.Handler interface for compatibility with Go 1.21+ slog
type SlogHandler struct {
	logger Logger
	level  slog.Level
	attrs  []slog.Attr
	group  string
}

// NewSlogHandler creates a new slog-compatible handler
func NewSlogHandler(logger Logger) slog.Handler {
	return &SlogHandler{
		logger: logger,
		level:  slog.LevelInfo,
		attrs:  make([]slog.Attr, 0),
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *SlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle handles the Record.
func (h *SlogHandler) Handle(ctx context.Context, record slog.Record) error {
	// Convert slog level to our level
	var logLevel Level
	switch record.Level {
	case slog.LevelDebug:
		logLevel = DebugLevel
	case slog.LevelInfo:
		logLevel = InfoLevel
	case slog.LevelWarn:
		logLevel = WarnLevel
	case slog.LevelError:
		logLevel = ErrorLevel
	default:
		logLevel = InfoLevel
	}

	// Build fields from attributes and record
	fields := make(Fields)

	// Add handler attributes
	for _, attr := range h.attrs {
		h.addAttrToFields(fields, attr, h.group)
	}

	// Add record attributes
	record.Attrs(func(attr slog.Attr) bool {
		h.addAttrToFields(fields, attr, "")
		return true
	})

	// Add built-in fields
	fields["time"] = record.Time.Format(time.RFC3339Nano)
	if record.PC != 0 {
		// Add source information if available
		// Note: In real implementation, you'd extract file:line from PC
		fields["source"] = "available"
	}

	// Log using our logger
	logger := h.logger.WithFields(fields).WithContext(ctx)
	logger.Log(logLevel, record.Message)

	return nil
}

// WithAttrs returns a new Handler whose attributes consist of both the receiver's attributes and the arguments.
func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &SlogHandler{
		logger: h.logger,
		level:  h.level,
		attrs:  newAttrs,
		group:  h.group,
	}
}

// WithGroup returns a new Handler with the given group appended to the receiver's existing groups.
func (h *SlogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	group := name
	if h.group != "" {
		group = h.group + "." + name
	}

	return &SlogHandler{
		logger: h.logger,
		level:  h.level,
		attrs:  h.attrs,
		group:  group,
	}
}

// addAttrToFields converts slog.Attr to our Fields format
func (h *SlogHandler) addAttrToFields(fields Fields, attr slog.Attr, group string) {
	key := attr.Key
	if group != "" {
		key = group + "." + key
	}

	switch attr.Value.Kind() {
	case slog.KindString:
		fields[key] = attr.Value.String()
	case slog.KindInt64:
		fields[key] = attr.Value.Int64()
	case slog.KindUint64:
		fields[key] = attr.Value.Uint64()
	case slog.KindFloat64:
		fields[key] = attr.Value.Float64()
	case slog.KindBool:
		fields[key] = attr.Value.Bool()
	case slog.KindDuration:
		fields[key] = attr.Value.Duration()
	case slog.KindTime:
		fields[key] = attr.Value.Time()
	case slog.KindAny:
		fields[key] = attr.Value.Any()
	case slog.KindGroup:
		// Handle group attributes
		attrs := attr.Value.Group()
		for _, groupAttr := range attrs {
			h.addAttrToFields(fields, groupAttr, key)
		}
	default:
		fields[key] = attr.Value.String()
	}
}

// SlogLogger wraps our logger to provide slog.Logger functionality
type SlogLogger struct {
	handler slog.Handler
	*slog.Logger
}

// NewSlogLogger creates a new slog.Logger using our handler
func NewSlogLogger(logger Logger) *SlogLogger {
	handler := NewSlogHandler(logger)
	slogLogger := slog.New(handler)

	return &SlogLogger{
		handler: handler,
		Logger:  slogLogger,
	}
}

// ToSlog converts our logger to a slog.Logger
func (l *logger) ToSlog() *slog.Logger {
	handler := NewSlogHandler(l)
	return slog.New(handler)
}

// FromSlog creates our logger from a slog.Logger (by wrapping it)
func FromSlog(slogLogger *slog.Logger) Logger {
	return &slogWrapper{slogLogger: slogLogger}
}

// slogWrapper wraps slog.Logger to implement our Logger interface
type slogWrapper struct {
	slogLogger *slog.Logger
}

func (sw *slogWrapper) Debug(args ...interface{}) { sw.slogLogger.Debug(fmt.Sprint(args...)) }
func (sw *slogWrapper) Info(args ...interface{})  { sw.slogLogger.Info(fmt.Sprint(args...)) }
func (sw *slogWrapper) Warn(args ...interface{})  { sw.slogLogger.Warn(fmt.Sprint(args...)) }
func (sw *slogWrapper) Error(args ...interface{}) { sw.slogLogger.Error(fmt.Sprint(args...)) }
func (sw *slogWrapper) Fatal(args ...interface{}) {
	sw.slogLogger.Error(fmt.Sprint(args...))
	panic("fatal")
}
func (sw *slogWrapper) Panic(args ...interface{}) { panic(fmt.Sprint(args...)) }
func (sw *slogWrapper) Debugf(format string, args ...interface{}) {
	sw.slogLogger.Debug(fmt.Sprintf(format, args...))
}
func (sw *slogWrapper) Infof(format string, args ...interface{}) {
	sw.slogLogger.Info(fmt.Sprintf(format, args...))
}
func (sw *slogWrapper) Warnf(format string, args ...interface{}) {
	sw.slogLogger.Warn(fmt.Sprintf(format, args...))
}
func (sw *slogWrapper) Errorf(format string, args ...interface{}) {
	sw.slogLogger.Error(fmt.Sprintf(format, args...))
}
func (sw *slogWrapper) Fatalf(format string, args ...interface{}) {
	sw.slogLogger.Error(fmt.Sprintf(format, args...))
	panic("fatal")
}
func (sw *slogWrapper) Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
func (sw *slogWrapper) DebugFast(msg string) { sw.slogLogger.Debug(msg) }
func (sw *slogWrapper) InfoFast(msg string)  { sw.slogLogger.Info(msg) }
func (sw *slogWrapper) WarnFast(msg string)  { sw.slogLogger.Warn(msg) }
func (sw *slogWrapper) ErrorFast(msg string) { sw.slogLogger.Error(msg) }
func (sw *slogWrapper) Log(level Level, args ...interface{}) {
	sw.slogLogger.Log(context.Background(), slog.Level(level.Value), fmt.Sprint(args...))
}
func (sw *slogWrapper) Logf(level Level, format string, args ...interface{}) {
	sw.slogLogger.Log(context.Background(), slog.Level(level.Value), fmt.Sprintf(format, args...))
}
func (sw *slogWrapper) LogFast(level Level, msg string) {
	sw.slogLogger.Log(context.Background(), slog.Level(level.Value), msg)
}
func (sw *slogWrapper) WithFields(fields Fields) Logger        { return sw } // Simplified
func (sw *slogWrapper) WithContext(ctx context.Context) Logger { return sw } // Simplified
func (sw *slogWrapper) WithTrace(ctx context.Context) Logger   { return sw } // Simplified
func (sw *slogWrapper) SetLevel(level Level)                   {}            // Simplified
func (sw *slogWrapper) SetHandler(handler Handler)             {}            // Simplified
func (sw *slogWrapper) SetFormatter(formatter Formatter)       {}            // Simplified
func (sw *slogWrapper) AddHook(hook Hook)                      {}            // Simplified
