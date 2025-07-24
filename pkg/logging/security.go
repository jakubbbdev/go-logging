package logging

import (
	"regexp"
	"strings"
)

// PIIDetector detects and sanitizes personally identifiable information
type PIIDetector struct {
	patterns map[string]*regexp.Regexp
	enabled  bool
}

// NewPIIDetector creates a new PII detector
func NewPIIDetector() *PIIDetector {
	return &PIIDetector{
		patterns: map[string]*regexp.Regexp{
			"email":       regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
			"phone":       regexp.MustCompile(`(\+?1[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}`),
			"ssn":         regexp.MustCompile(`\b\d{3}-?\d{2}-?\d{4}\b`),
			"credit_card": regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`),
			"ip_address":  regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`),
			"password":    regexp.MustCompile(`(?i)(password|pwd|pass|secret|token|key)\s*[:=]\s*\S+`),
		},
		enabled: true,
	}
}

// Enable enables PII detection
func (p *PIIDetector) Enable() {
	p.enabled = true
}

// Disable disables PII detection
func (p *PIIDetector) Disable() {
	p.enabled = false
}

// AddPattern adds a custom PII detection pattern
func (p *PIIDetector) AddPattern(name string, pattern *regexp.Regexp) {
	p.patterns[name] = pattern
}

// RemovePattern removes a PII detection pattern
func (p *PIIDetector) RemovePattern(name string) {
	delete(p.patterns, name)
}

// SanitizeString sanitizes a string by replacing PII with masked values
func (p *PIIDetector) SanitizeString(input string) string {
	if !p.enabled {
		return input
	}

	result := input
	for name, pattern := range p.patterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			return p.maskValue(name, match)
		})
	}
	return result
}

// SanitizeFields sanitizes all fields in a Fields map
func (p *PIIDetector) SanitizeFields(fields Fields) Fields {
	if !p.enabled || fields == nil {
		return fields
	}

	sanitized := make(Fields)
	for key, value := range fields {
		if str, ok := value.(string); ok {
			sanitized[key] = p.SanitizeString(str)
		} else {
			sanitized[key] = value
		}
	}
	return sanitized
}

// maskValue returns a masked version of the detected PII
func (p *PIIDetector) maskValue(piiType, value string) string {
	switch piiType {
	case "email":
		parts := strings.Split(value, "@")
		if len(parts) == 2 {
			username := parts[0]
			domain := parts[1]
			if len(username) > 2 {
				username = username[:2] + strings.Repeat("*", len(username)-2)
			}
			return username + "@" + domain
		}
		return "***@***.***"
	case "phone":
		return "***-***-" + value[len(value)-4:]
	case "ssn":
		return "***-**-" + value[len(value)-4:]
	case "credit_card":
		clean := strings.ReplaceAll(strings.ReplaceAll(value, "-", ""), " ", "")
		if len(clean) >= 4 {
			return "****-****-****-" + clean[len(clean)-4:]
		}
		return "****-****-****-****"
	case "ip_address":
		parts := strings.Split(value, ".")
		if len(parts) == 4 {
			return parts[0] + ".***.***.***"
		}
		return "***.***.***.***"
	case "password":
		if colonIndex := strings.Index(value, ":"); colonIndex != -1 {
			return value[:colonIndex+1] + " [REDACTED]"
		}
		if equalIndex := strings.Index(value, "="); equalIndex != -1 {
			return value[:equalIndex+1] + " [REDACTED]"
		}
		return "[REDACTED]"
	default:
		return "[REDACTED]"
	}
}

// SecurityHook creates a hook that sanitizes PII in log entries
func NewSecurityHook(detector *PIIDetector) Hook {
	return func(entry *Entry) {
		if detector != nil && detector.enabled {
			// Sanitize message
			entry.Message = detector.SanitizeString(entry.Message)

			// Sanitize fields
			entry.Fields = detector.SanitizeFields(entry.Fields)
		}
	}
}

// SecurityFormatter wraps another formatter and sanitizes output
type SecurityFormatter struct {
	formatter Formatter
	detector  *PIIDetector
}

// NewSecurityFormatter creates a new security formatter
func NewSecurityFormatter(formatter Formatter, detector *PIIDetector) Formatter {
	return &SecurityFormatter{
		formatter: formatter,
		detector:  detector,
	}
}

// Format formats the entry and sanitizes the output
func (sf *SecurityFormatter) Format(entry *Entry) ([]byte, error) {
	// Create a copy of the entry for sanitization
	sanitizedEntry := &Entry{
		Level:   entry.Level,
		Message: entry.Message,
		Fields:  entry.Fields,
		Time:    entry.Time,
		Caller:  entry.Caller,
		Context: entry.Context,
	}

	if sf.detector != nil && sf.detector.enabled {
		sanitizedEntry.Message = sf.detector.SanitizeString(sanitizedEntry.Message)
		sanitizedEntry.Fields = sf.detector.SanitizeFields(sanitizedEntry.Fields)
	}

	return sf.formatter.Format(sanitizedEntry)
}

// Common PII patterns for easy access
var (
	EmailPattern      = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	PhonePattern      = regexp.MustCompile(`(\+?1[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}`)
	SSNPattern        = regexp.MustCompile(`\b\d{3}-?\d{2}-?\d{4}\b`)
	CreditCardPattern = regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`)
	IPAddressPattern  = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	PasswordPattern   = regexp.MustCompile(`(?i)(password|pwd|pass|secret|token|key)\s*[:=]\s*\S+`)
)
