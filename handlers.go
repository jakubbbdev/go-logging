package logging

import (
	"fmt"
	"os"
	"sync"
)

// ConsoleHandler handles logging to console
type ConsoleHandler struct {
	formatter Formatter
	mu        sync.Mutex
}

// NewConsoleHandler creates a new console handler
func NewConsoleHandler() Handler {
	return &ConsoleHandler{
		formatter: NewTextFormatter(),
	}
}

// Handle implements the Handler interface for console output
func (h *ConsoleHandler) Handle(entry *Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	formatted, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(os.Stdout, string(formatted))
	return err
}

// SetFormatter sets the formatter for the console handler
func (h *ConsoleHandler) SetFormatter(formatter Formatter) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.formatter = formatter
}

// FileHandler handles logging to a file
type FileHandler struct {
	file      *os.File
	formatter Formatter
	mu        sync.Mutex
}

// NewFileHandler creates a new file handler
func NewFileHandler(filename string) (Handler, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileHandler{
		file:      file,
		formatter: NewTextFormatter(),
	}, nil
}

// Handle implements the Handler interface for file output
func (h *FileHandler) Handle(entry *Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	formatted, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = h.file.Write(append(formatted, '\n'))
	return err
}

// SetFormatter sets the formatter for the file handler
func (h *FileHandler) SetFormatter(formatter Formatter) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.formatter = formatter
}

// Close closes the file handler
func (h *FileHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.file.Close()
}

// MultiHandler handles logging to multiple handlers
type MultiHandler struct {
	handlers []Handler
	mu       sync.RWMutex
}

// NewMultiHandler creates a new multi handler
func NewMultiHandler(handlers ...Handler) Handler {
	return &MultiHandler{
		handlers: handlers,
	}
}

// Handle implements the Handler interface for multiple handlers
func (h *MultiHandler) Handle(entry *Entry) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var lastErr error
	for _, handler := range h.handlers {
		if err := handler.Handle(entry); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// AddHandler adds a new handler to the multi handler
func (h *MultiHandler) AddHandler(handler Handler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers = append(h.handlers, handler)
}

// RemoveHandler removes a handler from the multi handler
func (h *MultiHandler) RemoveHandler(handler Handler) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, existingHandler := range h.handlers {
		if existingHandler == handler {
			h.handlers = append(h.handlers[:i], h.handlers[i+1:]...)
			break
		}
	}
}

// JSONHandler handles logging in JSON format
type JSONHandler struct {
	handler Handler
}

// NewJSONHandler creates a new JSON handler
func NewJSONHandler(handler Handler) Handler {
	return &JSONHandler{
		handler: handler,
	}
}

// Handle implements the Handler interface for JSON output
func (h *JSONHandler) Handle(entry *Entry) error {
	// Create a copy of the entry with JSON formatter
	jsonEntry := &Entry{
		Level:   entry.Level,
		Message: entry.Message,
		Fields:  entry.Fields,
		Time:    entry.Time,
		Caller:  entry.Caller,
		Context: entry.Context,
	}

	// Use JSON formatter
	formatter := NewJSONFormatter()
	formatted, err := formatter.Format(jsonEntry)
	if err != nil {
		return err
	}

	// Write to the underlying handler
	return h.handler.Handle(&Entry{
		Level:   entry.Level,
		Message: string(formatted),
		Time:    entry.Time,
	})
}
