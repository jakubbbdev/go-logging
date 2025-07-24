package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mathrand "math/rand"
	"strings"
)

// TraceContext represents tracing information
type TraceContext struct {
	TraceID   string
	SpanID    string
	RequestID string
	UserID    string
	SessionID string
}

// ContextKey type for context keys
type ContextKey string

const (
	// TraceContextKey is the context key for trace context
	TraceContextKey ContextKey = "trace_context"
)

// GenerateTraceID generates a new trace ID
func GenerateTraceID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("trace_%d", mathrand.Int63())
	}
	return hex.EncodeToString(bytes)
}

// GenerateSpanID generates a new span ID
func GenerateSpanID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("span_%d", mathrand.Int63())
	}
	return hex.EncodeToString(bytes)
}

// GenerateRequestID generates a new request ID
func GenerateRequestID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("req_%d", mathrand.Int63())
	}
	return strings.ToUpper(hex.EncodeToString(bytes))
}

// WithTraceContext adds trace context to the context
func WithTraceContext(ctx context.Context, tc *TraceContext) context.Context {
	return context.WithValue(ctx, TraceContextKey, tc)
}

// TraceFromContext extracts trace context from context
func TraceFromContext(ctx context.Context) *TraceContext {
	if tc, ok := ctx.Value(TraceContextKey).(*TraceContext); ok {
		return tc
	}
	return nil
}

// WithTrace returns a logger with tracing context
func (l *logger) WithTrace(ctx context.Context) Logger {
	tc := TraceFromContext(ctx)
	if tc == nil {
		return l
	}

	fields := make(Fields)
	for k, v := range l.fields {
		fields[k] = v
	}

	if tc.TraceID != "" {
		fields["trace_id"] = tc.TraceID
	}
	if tc.SpanID != "" {
		fields["span_id"] = tc.SpanID
	}
	if tc.RequestID != "" {
		fields["request_id"] = tc.RequestID
	}
	if tc.UserID != "" {
		fields["user_id"] = tc.UserID
	}
	if tc.SessionID != "" {
		fields["session_id"] = tc.SessionID
	}

	return &logger{
		level:     l.level,
		handler:   l.handler,
		formatter: l.formatter,
		fields:    fields,
		hooks:     l.hooks,
	}
}

// NewTraceContext creates a new trace context
func NewTraceContext() *TraceContext {
	return &TraceContext{
		TraceID:   GenerateTraceID(),
		SpanID:    GenerateSpanID(),
		RequestID: GenerateRequestID(),
	}
}

// WithUserID adds user ID to trace context
func (tc *TraceContext) WithUserID(userID string) *TraceContext {
	tc.UserID = userID
	return tc
}

// WithSessionID adds session ID to trace context
func (tc *TraceContext) WithSessionID(sessionID string) *TraceContext {
	tc.SessionID = sessionID
	return tc
}

// String returns a string representation of the trace context
func (tc *TraceContext) String() string {
	var parts []string
	if tc.TraceID != "" {
		parts = append(parts, fmt.Sprintf("trace=%s", tc.TraceID))
	}
	if tc.SpanID != "" {
		parts = append(parts, fmt.Sprintf("span=%s", tc.SpanID))
	}
	if tc.RequestID != "" {
		parts = append(parts, fmt.Sprintf("req=%s", tc.RequestID))
	}
	return strings.Join(parts, " ")
}

// TracingHook creates a hook that adds tracing information
func NewTracingHook() Hook {
	return func(entry *Entry) {
		if entry.Context != nil {
			tc := TraceFromContext(entry.Context)
			if tc != nil {
				if entry.Fields == nil {
					entry.Fields = make(Fields)
				}

				if tc.TraceID != "" {
					entry.Fields["trace_id"] = tc.TraceID
				}
				if tc.SpanID != "" {
					entry.Fields["span_id"] = tc.SpanID
				}
				if tc.RequestID != "" {
					entry.Fields["request_id"] = tc.RequestID
				}
				if tc.UserID != "" {
					entry.Fields["user_id"] = tc.UserID
				}
				if tc.SessionID != "" {
					entry.Fields["session_id"] = tc.SessionID
				}
			}
		}
	}
}
