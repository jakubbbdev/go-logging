package logging

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

// Level represents a logging level (built-in or custom).
//
// Use RegisterLevel to add your own levels.
type Level struct {
	Name  string
	Value int
}

var (
	DebugLevel   = Level{"debug", 10}
	InfoLevel    = Level{"info", 20}
	WarnLevel    = Level{"warn", 30}
	ErrorLevel   = Level{"error", 40}
	FatalLevel   = Level{"fatal", 50}
	PanicLevel   = Level{"panic", 60}
	customLevels = make(map[string]Level)
	levelOrder   = []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel, PanicLevel}
	levelMu      sync.RWMutex
)

// RegisterLevel registers a new custom log level.
func RegisterLevel(name string, value int) Level {
	levelMu.Lock()
	defer levelMu.Unlock()
	lvl := Level{Name: name, Value: value}
	customLevels[name] = lvl
	levelOrder = append(levelOrder, lvl)
	return lvl
}

// ParseLevel returns a Level by name (case-insensitive).
func ParseLevel(name string) (Level, bool) {
	levelMu.RLock()
	defer levelMu.RUnlock()
	for _, lvl := range levelOrder {
		if lvl.Name == name {
			return lvl, true
		}
	}
	return Level{"unknown", 0}, false
}

// String returns the string representation of the level
func (l Level) String() string {
	return l.Name
}

// Fields represents key-value pairs for structured logging
type Fields map[string]interface{}

// Entry represents a log entry
//
// Contains all information about a single log event.
type Entry struct {
	Level   Level
	Message string
	Fields  Fields
	Time    time.Time
	Caller  string
	Context context.Context
}

// Reset resets the entry for reuse in the pool
func (e *Entry) Reset() {
	e.Level = InfoLevel
	e.Message = ""
	e.Fields = nil
	e.Time = time.Time{}
	e.Caller = ""
	e.Context = nil
}

// Logger is the main logging interface.
//
// Use NewLogger(...) to create a new logger instance.
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	// Fast logging methods for performance-critical applications
	DebugFast(msg string)
	InfoFast(msg string)
	WarnFast(msg string)
	ErrorFast(msg string)

	// Log with a custom level
	Log(level Level, args ...interface{})
	Logf(level Level, format string, args ...interface{})
	LogFast(level Level, msg string)

	WithFields(fields Fields) Logger
	WithContext(ctx context.Context) Logger
	WithTrace(ctx context.Context) Logger
	SetLevel(level Level)
	SetHandler(handler Handler)
	SetFormatter(formatter Formatter)

	// AddHook adds a hook function that is called for every log entry
	AddHook(hook Hook)
}

// Handler interface for handling log entries
type Handler interface {
	Handle(entry *Entry) error
}

// Formatter interface for formatting log entries
type Formatter interface {
	Format(entry *Entry) ([]byte, error)
}

// Hook is a function that is called for every log entry before it is handled.
type Hook func(entry *Entry)

// Option is a functional option for configuring the logger.
type Option func(*logger)

// WithLevel sets the log level for the logger.
func WithLevel(level Level) Option {
	return func(l *logger) {
		l.level = level
	}
}

// WithHandler sets the handler for the logger.
func WithHandler(handler Handler) Option {
	return func(l *logger) {
		l.handler = handler
	}
}

// WithFormatter sets the formatter for the logger.
func WithFormatter(formatter Formatter) Option {
	return func(l *logger) {
		l.formatter = formatter
	}
}

// WithDefaultFields sets default fields for the logger.
func WithDefaultFields(fields Fields) Option {
	return func(l *logger) {
		for k, v := range fields {
			l.fields[k] = v
		}
	}
}

// WithHook adds a hook to the logger.
func WithHook(hook Hook) Option {
	return func(l *logger) {
		l.hooks = append(l.hooks, hook)
	}
}

// WithCaller enables caller (file:line) information in log entries.
func WithCaller(enabled bool) Option {
	return func(l *logger) {
		l.includeCaller = enabled
	}
}

// WithStacktrace enables stacktrace for error/fatal/panic logs.
func WithStacktrace(enabled bool) Option {
	return func(l *logger) {
		l.includeStacktrace = enabled
	}
}

// Entry pool for reducing allocations
var entryPool = sync.Pool{
	New: func() interface{} {
		return &Entry{}
	},
}

// logger implements the Logger interface
type logger struct {
	level             Level
	handler           Handler
	formatter         Formatter
	fields            Fields
	hooks             []Hook
	mu                sync.RWMutex
	includeCaller     bool
	includeStacktrace bool
}

// NewLogger creates a new logger instance with optional configuration.
//
// Example:
//
//	logger := logging.NewLogger(
//	    logging.WithLevel(logging.DebugLevel),
//	    logging.WithFormatter(logging.NewJSONFormatter()),
//	    logging.WithHandler(logging.NewRotatingFileHandler("app.log", 10*1024*1024, 5)),
//	)
func NewLogger(opts ...Option) Logger {
	l := &logger{
		level:     InfoLevel,
		handler:   NewConsoleHandler(),
		formatter: NewTextFormatter(),
		fields:    make(Fields),
		hooks:     make([]Hook, 0),
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// SetLevel sets the logging level
func (l *logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetHandler sets the log handler
func (l *logger) SetHandler(handler Handler) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.handler = handler
}

// SetFormatter sets the log formatter
func (l *logger) SetFormatter(formatter Formatter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.formatter = formatter
}

// AddHook adds a hook function to the logger.
func (l *logger) AddHook(hook Hook) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks = append(l.hooks, hook)
}

// WithFields returns a new logger with the given fields
func (l *logger) WithFields(fields Fields) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newFields := make(Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &logger{
		level:     l.level,
		handler:   l.handler,
		formatter: l.formatter,
		fields:    newFields,
		hooks:     l.hooks,
	}
}

// WithContext returns a new logger with the given context
func (l *logger) WithContext(ctx context.Context) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return &logger{
		level:     l.level,
		handler:   l.handler,
		formatter: l.formatter,
		fields:    l.fields,
		hooks:     l.hooks,
	}
}

// getEntryFromPool gets an entry from the pool
func getEntryFromPool() *Entry {
	return entryPool.Get().(*Entry)
}

// putEntryToPool returns an entry to the pool
func putEntryToPool(entry *Entry) {
	entry.Reset()
	entryPool.Put(entry)
}

// runHooks runs all hooks for the entry
func (l *logger) runHooks(entry *Entry) {
	for _, hook := range l.hooks {
		hook(entry)
	}
}

// log creates and handles a log entry
func (l *logger) log(level Level, args ...interface{}) {
	if level.Value < l.level.Value {
		return
	}

	entry := getEntryFromPool()
	entry.Level = level
	entry.Message = fmt.Sprint(args...)
	entry.Fields = l.fields
	entry.Time = time.Now()

	if l.includeCaller {
		entry.Caller = callerString()
	}

	l.mu.RLock()
	handler := l.handler
	hooks := l.hooks
	l.mu.RUnlock()

	for _, hook := range hooks {
		hook(entry)
	}

	if handler != nil {
		handler.Handle(entry)
	}

	if l.includeStacktrace && (level == ErrorLevel || level == FatalLevel || level == PanicLevel) {
		entry.Fields["stacktrace"] = stacktraceString()
	}

	putEntryToPool(entry)
}

// logf creates and handles a formatted log entry
func (l *logger) logf(level Level, format string, args ...interface{}) {
	if level.Value < l.level.Value {
		return
	}

	entry := getEntryFromPool()
	entry.Level = level
	entry.Message = fmt.Sprintf(format, args...)
	entry.Fields = l.fields
	entry.Time = time.Now()

	if l.includeCaller {
		entry.Caller = callerString()
	}

	l.mu.RLock()
	handler := l.handler
	hooks := l.hooks
	l.mu.RUnlock()

	for _, hook := range hooks {
		hook(entry)
	}

	if handler != nil {
		handler.Handle(entry)
	}

	if l.includeStacktrace && (level == ErrorLevel || level == FatalLevel || level == PanicLevel) {
		entry.Fields["stacktrace"] = stacktraceString()
	}

	putEntryToPool(entry)
}

// logFast creates and handles a log entry without string formatting (for performance)
func (l *logger) logFast(level Level, msg string) {
	if level.Value < l.level.Value {
		return
	}

	entry := getEntryFromPool()
	entry.Level = level
	entry.Message = msg
	entry.Fields = l.fields
	entry.Time = time.Now()

	if l.includeCaller {
		entry.Caller = callerString()
	}

	l.mu.RLock()
	handler := l.handler
	hooks := l.hooks
	l.mu.RUnlock()

	for _, hook := range hooks {
		hook(entry)
	}

	if handler != nil {
		handler.Handle(entry)
	}

	if l.includeStacktrace && (level == ErrorLevel || level == FatalLevel || level == PanicLevel) {
		entry.Fields["stacktrace"] = stacktraceString()
	}

	putEntryToPool(entry)
}

// Log logs a message at a custom level.
func (l *logger) Log(level Level, args ...interface{}) {
	l.log(level, args...)
}

// Logf logs a formatted message at a custom level.
func (l *logger) Logf(level Level, format string, args ...interface{}) {
	l.logf(level, format, args...)
}

// LogFast logs a message at a custom level (fast path).
func (l *logger) LogFast(level Level, msg string) {
	l.logFast(level, msg)
}

// Debug logs a debug message
func (l *logger) Debug(args ...interface{}) { l.log(DebugLevel, args...) }
func (l *logger) Info(args ...interface{})  { l.log(InfoLevel, args...) }
func (l *logger) Warn(args ...interface{})  { l.log(WarnLevel, args...) }
func (l *logger) Error(args ...interface{}) { l.log(ErrorLevel, args...) }
func (l *logger) Fatal(args ...interface{}) { l.log(FatalLevel, args...); os.Exit(1) }
func (l *logger) Panic(args ...interface{}) { l.log(PanicLevel, args...); panic(fmt.Sprint(args...)) }

func (l *logger) Debugf(format string, args ...interface{}) { l.logf(DebugLevel, format, args...) }
func (l *logger) Infof(format string, args ...interface{})  { l.logf(InfoLevel, format, args...) }
func (l *logger) Warnf(format string, args ...interface{})  { l.logf(WarnLevel, format, args...) }
func (l *logger) Errorf(format string, args ...interface{}) { l.logf(ErrorLevel, format, args...) }
func (l *logger) Fatalf(format string, args ...interface{}) {
	l.logf(FatalLevel, format, args...)
	os.Exit(1)
}
func (l *logger) Panicf(format string, args ...interface{}) {
	l.logf(PanicLevel, format, args...)
	panic(fmt.Sprintf(format, args...))
}

// Fast logging methods for performance-critical applications
func (l *logger) DebugFast(msg string) { l.logFast(DebugLevel, msg) }
func (l *logger) InfoFast(msg string)  { l.logFast(InfoLevel, msg) }
func (l *logger) WarnFast(msg string)  { l.logFast(WarnLevel, msg) }
func (l *logger) ErrorFast(msg string) { l.logFast(ErrorLevel, msg) }

// Helper functions:
func callerString() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func stacktraceString() string {
	buf := make([]byte, 2048)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}
