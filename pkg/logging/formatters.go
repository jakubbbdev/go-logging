package logging

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// TextFormatter formats log entries as text
type TextFormatter struct {
	UseColors bool
	Timestamp bool
}

// NewTextFormatter creates a new text formatter
func NewTextFormatter() Formatter {
	return &TextFormatter{
		UseColors: true,
		Timestamp: true,
	}
}

// Format implements the Formatter interface for text output
func (f *TextFormatter) Format(entry *Entry) ([]byte, error) {
	var parts []string

	// Add timestamp
	if f.Timestamp {
		parts = append(parts, entry.Time.Format("2006-01-02 15:04:05"))
	}

	// Add level
	levelStr := entry.Level.String()
	if f.UseColors {
		levelStr = f.colorizeLevel(entry.Level, levelStr)
	}
	parts = append(parts, fmt.Sprintf("[%s]", strings.ToUpper(levelStr)))

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

// colorizeLevel adds colors to the level string
func (f *TextFormatter) colorizeLevel(level Level, levelStr string) string {
	if !f.UseColors {
		return levelStr
	}

	switch level {
	case DebugLevel:
		return color.CyanString(levelStr)
	case InfoLevel:
		return color.GreenString(levelStr)
	case WarnLevel:
		return color.YellowString(levelStr)
	case ErrorLevel:
		return color.RedString(levelStr)
	case FatalLevel:
		return color.MagentaString(levelStr)
	case PanicLevel:
		return color.MagentaString(levelStr)
	default:
		return levelStr
	}
}

// formatFields formats the fields as a string
func (f *TextFormatter) formatFields(fields Fields) string {
	if len(fields) == 0 {
		return ""
	}

	var parts []string
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
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
