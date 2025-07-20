package logging

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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

// RotatingFileHandler handles logging to rotating files
type RotatingFileHandler struct {
	filename    string
	maxSize     int64
	maxFiles    int
	currentFile *os.File
	formatter   Formatter
	mu          sync.Mutex
	currentSize int64
}

// NewRotatingFileHandler creates a new rotating file handler
func NewRotatingFileHandler(filename string, maxSize int64, maxFiles int) (Handler, error) {
	handler := &RotatingFileHandler{
		filename:  filename,
		maxSize:   maxSize,
		maxFiles:  maxFiles,
		formatter: NewTextFormatter(),
	}

	if err := handler.openFile(); err != nil {
		return nil, err
	}

	return handler, nil
}

// openFile opens the current log file
func (h *RotatingFileHandler) openFile() error {
	file, err := os.OpenFile(h.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// Get current file size
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}

	h.currentFile = file
	h.currentSize = info.Size()
	return nil
}

// rotate rotates the log file
func (h *RotatingFileHandler) rotate() error {
	if h.currentFile != nil {
		h.currentFile.Close()
	}

	// Rotate existing files
	for i := h.maxFiles - 1; i > 0; i-- {
		oldName := h.filename + "." + strconv.Itoa(i)
		newName := h.filename + "." + strconv.Itoa(i+1)

		if _, err := os.Stat(oldName); err == nil {
			os.Rename(oldName, newName)
		}
	}

	// Rename current file
	if _, err := os.Stat(h.filename); err == nil {
		os.Rename(h.filename, h.filename+".1")
	}

	return h.openFile()
}

// Handle implements the Handler interface for rotating file output
func (h *RotatingFileHandler) Handle(entry *Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	formatted, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	// Check if we need to rotate
	if h.currentSize+int64(len(formatted)+1) > h.maxSize {
		if err := h.rotate(); err != nil {
			return err
		}
	}

	_, err = h.currentFile.Write(append(formatted, '\n'))
	if err == nil {
		h.currentSize += int64(len(formatted) + 1)
	}
	return err
}

// SetFormatter sets the formatter for the rotating file handler
func (h *RotatingFileHandler) SetFormatter(formatter Formatter) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.formatter = formatter
}

// Close closes the rotating file handler
func (h *RotatingFileHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.currentFile != nil {
		return h.currentFile.Close()
	}
	return nil
}

// HTTPHandler handles logging via HTTP requests
type HTTPHandler struct {
	endpoint  string
	client    *http.Client
	formatter Formatter
	mu        sync.Mutex
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(endpoint string) Handler {
	return &HTTPHandler{
		endpoint:  endpoint,
		client:    &http.Client{Timeout: 10 * time.Second},
		formatter: NewJSONFormatter(),
	}
}

// Handle implements the Handler interface for HTTP output
func (h *HTTPHandler) Handle(entry *Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	formatted, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", h.endpoint, strings.NewReader(string(formatted)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "go-logging/1.0")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	return nil
}

// SetFormatter sets the formatter for the HTTP handler
func (h *HTTPHandler) SetFormatter(formatter Formatter) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.formatter = formatter
}

// AsyncHandler handles logging asynchronously
type AsyncHandler struct {
	handler Handler
	buffer  chan *Entry
	workers int
	wg      sync.WaitGroup
	stop    chan struct{}
}

// NewAsyncHandler creates a new async handler
func NewAsyncHandler(handler Handler, bufferSize int, workers int) Handler {
	async := &AsyncHandler{
		handler: handler,
		buffer:  make(chan *Entry, bufferSize),
		workers: workers,
		stop:    make(chan struct{}),
	}

	async.start()
	return async
}

// start starts the async workers
func (h *AsyncHandler) start() {
	for i := 0; i < h.workers; i++ {
		h.wg.Add(1)
		go h.worker()
	}
}

// worker processes entries from the buffer
func (h *AsyncHandler) worker() {
	defer h.wg.Done()

	for {
		select {
		case entry := <-h.buffer:
			if entry != nil {
				h.handler.Handle(entry)
			}
		case <-h.stop:
			return
		}
	}
}

// Handle implements the Handler interface for async output
func (h *AsyncHandler) Handle(entry *Entry) error {
	select {
	case h.buffer <- entry:
		return nil
	default:
		// Buffer is full, log synchronously
		return h.handler.Handle(entry)
	}
}

// Stop stops the async handler
func (h *AsyncHandler) Stop() {
	close(h.stop)
	h.wg.Wait()
}

// SamplingHandler handles logging with sampling
type SamplingHandler struct {
	handler Handler
	rate    float64
	mu      sync.Mutex
	counter int
}

// NewSamplingHandler creates a new sampling handler
func NewSamplingHandler(handler Handler, rate float64) Handler {
	return &SamplingHandler{
		handler: handler,
		rate:    rate,
		counter: 0,
	}
}

// Handle implements the Handler interface for sampling output
func (h *SamplingHandler) Handle(entry *Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.counter++

	// Use counter-based sampling for deterministic behavior
	if h.rate < 1.0 {
		if float64(h.counter%100) >= h.rate*100 {
			return nil // Skip this log entry
		}
	}

	return h.handler.Handle(entry)
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
