package main

import (
	"time"

	"github.com/fatih/color"
	"github.com/jakubbbdev/go-logging/pkg/logging"
)

func main() {
	// 1. Custom Level registration
	AuditLevel := logging.RegisterLevel("audit", 25)
	TraceLevel := logging.RegisterLevel("trace", 5)

	// 2. Logger mit allen modernen Features
	logger := logging.NewLogger(
		logging.WithLevel(TraceLevel), // Zeige alles ab Trace
		logging.WithCaller(true),
		logging.WithStacktrace(true),
		logging.WithFormatter(logging.NewTextFormatter(
			logging.WithTextFormatterEmojis(map[logging.Level]string{
				logging.DebugLevel: "🐛 ",
				logging.InfoLevel:  "ℹ️ ",
				logging.WarnLevel:  "⚠️ ",
				logging.ErrorLevel: "❌ ",
				AuditLevel:         "🔍 ",
				TraceLevel:         "🔎 ",
			}),
			logging.WithTextFormatterColors(map[logging.Level]*color.Color{
				logging.InfoLevel:  color.New(color.FgHiBlue, color.Bold),
				logging.ErrorLevel: color.New(color.FgHiRed, color.Bold, color.BgBlack),
				AuditLevel:         color.New(color.FgHiCyan, color.Bold),
				TraceLevel:         color.New(color.FgHiWhite, color.Bold),
			}),
			logging.WithTextFormatterTimestampFormat("15:04:05.000"),
			logging.WithTextFormatterLevelPadding(7),
			logging.WithTextFormatterFieldOrder([]string{"user_id", "action", "ip", "password", "token"}),
			logging.WithTextFormatterFieldMasking([]string{"password", "token"}, "****"),
		)),
	)

	// 3. Verschiedene Log-Levels
	logger.Log(TraceLevel, "Trace message for debugging")
	logger.Debug("Debugging info")
	logger.Info("Application started")
	logger.Warn("This is a warning!")
	logger.Error("Something went wrong!")
	logger.Log(AuditLevel, "User audit event", 123)

	// 4. Mit Feldern, Masking und Stacktrace
	logger.WithFields(logging.Fields{
		"user_id":  42,
		"action":   "login",
		"ip":       "127.0.0.1",
		"password": "supersecret",
		"token":    "abcdefg",
	}).Error("Login failed!")

	// 5. Stacktrace bei Panic
	// logger.Panic("This is a panic!") // (auskommentiert, sonst bricht das Beispiel ab)

	// 6. Formatiertes Audit-Log
	logger.Logf(AuditLevel, "Audit for user %d at %s", 42, time.Now().Format(time.RFC822))
}
