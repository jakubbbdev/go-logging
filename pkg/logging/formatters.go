package logging

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
)

// TextFormatterOption is a functional option for TextFormatter configuration.
type TextFormatterOption func(*TextFormatter)

// TextFormatter formats log entries as text.
type TextFormatter struct {
	UseColors       bool
	Timestamp       bool
	TimestampFormat string
	LevelPadding    int
	LevelColors     map[Level]*color.Color
	LevelPrefix     map[Level]string
	LevelSuffix     map[Level]string
	FieldOrder      []string
}

// NewTextFormatter creates a new text formatter with options.
func NewTextFormatter(opts ...TextFormatterOption) Formatter {
	f := &TextFormatter{
		UseColors:       true,
		Timestamp:       true,
		TimestampFormat: "2006-01-02 15:04:05",
		LevelPadding:    5,
		LevelColors: map[Level]*color.Color{
			DebugLevel: color.New(color.FgCyan),
			InfoLevel:  color.New(color.FgGreen),
			WarnLevel:  color.New(color.FgYellow),
			ErrorLevel: color.New(color.FgRed),
			FatalLevel: color.New(color.FgMagenta),
			PanicLevel: color.New(color.FgHiMagenta),
		},
		LevelPrefix: map[Level]string{},
		LevelSuffix: map[Level]string{},
		FieldOrder:  nil,
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// WithTextFormatterColors allows custom colors per log level.
func WithTextFormatterColors(colors map[Level]*color.Color) TextFormatterOption {
	return func(f *TextFormatter) {
		for lvl, c := range colors {
			f.LevelColors[lvl] = c
		}
	}
}

// WithTextFormatterTimestampFormat sets the timestamp format.
func WithTextFormatterTimestampFormat(format string) TextFormatterOption {
	return func(f *TextFormatter) {
		f.TimestampFormat = format
	}
}

// WithTextFormatterLevelPadding sets the padding for the level string.
func WithTextFormatterLevelPadding(pad int) TextFormatterOption {
	return func(f *TextFormatter) {
		f.LevelPadding = pad
	}
}

// WithTextFormatterPrefix sets a custom prefix for a log level.
func WithTextFormatterPrefix(level Level, prefix string) TextFormatterOption {
	return func(f *TextFormatter) {
		f.LevelPrefix[level] = prefix
	}
}

// WithTextFormatterSuffix sets a custom suffix for a log level.
func WithTextFormatterSuffix(level Level, suffix string) TextFormatterOption {
	return func(f *TextFormatter) {
		f.LevelSuffix[level] = suffix
	}
}

// WithTextFormatterFieldOrder sets the order of fields in output.
func WithTextFormatterFieldOrder(order []string) TextFormatterOption {
	return func(f *TextFormatter) {
		f.FieldOrder = order
	}
}

// Format implements the Formatter interface for text output
func (f *TextFormatter) Format(entry *Entry) ([]byte, error) {
	var parts []string

	// Add timestamp
	if f.Timestamp {
		parts = append(parts, entry.Time.Format(f.TimestampFormat))
	}

	// Add level
	levelStr := strings.ToUpper(entry.Level.String())
	if f.LevelPadding > 0 {
		levelStr = fmt.Sprintf("%-*s", f.LevelPadding, levelStr)
	}
	if f.UseColors {
		if c, ok := f.LevelColors[entry.Level]; ok {
			levelStr = c.Sprint(levelStr)
		}
	}
	if prefix, ok := f.LevelPrefix[entry.Level]; ok && prefix != "" {
		levelStr = prefix + levelStr
	}
	if suffix, ok := f.LevelSuffix[entry.Level]; ok && suffix != "" {
		levelStr = levelStr + suffix
	}
	parts = append(parts, fmt.Sprintf("[%s]", levelStr))

	// Add message
	parts = append(parts, entry.Message)

	// Add fields
	if len(entry.Fields) > 0 {
		fieldStr := f.formatFields(entry.Fields)
		parts = append(parts, fieldStr)
	}

	// Add caller if available
	if entry.Caller != "" {
		parts = append(parts, fmt.Sprintf("(%s)", entry.Caller))
	}

	result := strings.Join(parts, " ")
	return []byte(result), nil
}

// formatFields formats the fields as a string
func (f *TextFormatter) formatFields(fields Fields) string {
	if len(fields) == 0 {
		return ""
	}

	var parts []string
	if f.FieldOrder != nil && len(f.FieldOrder) > 0 {
		for _, k := range f.FieldOrder {
			if v, ok := fields[k]; ok {
				parts = append(parts, fmt.Sprintf("%s=%v", k, v))
			}
		}
		// Add any remaining fields not in FieldOrder
		for k, v := range fields {
			found := false
			for _, ok := range f.FieldOrder {
				if k == ok {
					found = true
					break
				}
			}
			if !found {
				parts = append(parts, fmt.Sprintf("%s=%v", k, v))
			}
		}
	} else {
		// Default: sort keys alphabetically
		keys := make([]string, 0, len(fields))
		for k := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%v", k, fields[k]))
		}
	}

	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}

// JSONFormatter formats log entries as JSON
type JSONFormatter struct {
	PrettyPrint bool
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() Formatter {
	return &JSONFormatter{
		PrettyPrint: false,
	}
}

// Format implements the Formatter interface for JSON output
func (f *JSONFormatter) Format(entry *Entry) ([]byte, error) {
	data := map[string]interface{}{
		"level":   entry.Level.String(),
		"message": entry.Message,
		"time":    entry.Time.Format(time.RFC3339),
	}

	// Add fields
	if len(entry.Fields) > 0 {
		for k, v := range entry.Fields {
			data[k] = v
		}
	}

	// Add caller if available
	if entry.Caller != "" {
		data["caller"] = entry.Caller
	}

	if f.PrettyPrint {
		return json.MarshalIndent(data, "", "  ")
	}

	return json.Marshal(data)
}

// SetPrettyPrint enables or disables pretty printing for JSON
func (f *JSONFormatter) SetPrettyPrint(pretty bool) {
	f.PrettyPrint = pretty
}
