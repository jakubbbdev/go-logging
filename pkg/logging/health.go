package logging

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus int

const (
	HealthStatusHealthy HealthStatus = iota
	HealthStatusDegraded
	HealthStatusUnhealthy
)

func (hs HealthStatus) String() string {
	switch hs {
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusDegraded:
		return "degraded"
	case HealthStatusUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// HealthCheck represents a health check function
type HealthCheck func(ctx context.Context) HealthCheckResult

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status    HealthStatus
	Message   string
	Duration  time.Duration
	Timestamp time.Time
	Error     error
}

// HealthMonitor monitors the health of logging components
type HealthMonitor struct {
	checks   map[string]HealthCheck
	results  map[string]HealthCheckResult
	mu       sync.RWMutex
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	logger   Logger
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(logger Logger, interval time.Duration) *HealthMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &HealthMonitor{
		checks:   make(map[string]HealthCheck),
		results:  make(map[string]HealthCheckResult),
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
	}
}

// AddCheck adds a health check
func (hm *HealthMonitor) AddCheck(name string, check HealthCheck) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.checks[name] = check
}

// RemoveCheck removes a health check
func (hm *HealthMonitor) RemoveCheck(name string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	delete(hm.checks, name)
}

// Start starts the health monitor
func (hm *HealthMonitor) Start() {
	go hm.monitor()
}

// Stop stops the health monitor
func (hm *HealthMonitor) Stop() {
	hm.cancel()
}

// GetHealth returns the current health status
func (hm *HealthMonitor) GetHealth() map[string]HealthCheckResult {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	results := make(map[string]HealthCheckResult)
	for name, result := range hm.results {
		results[name] = result
	}
	return results
}

// GetOverallHealth returns the overall health status
func (hm *HealthMonitor) GetOverallHealth() HealthStatus {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if len(hm.results) == 0 {
		return HealthStatusUnhealthy
	}

	hasUnhealthy := false
	hasDegraded := false

	for _, result := range hm.results {
		switch result.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	}
	if hasDegraded {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

// monitor runs the health checks periodically
func (hm *HealthMonitor) monitor() {
	ticker := time.NewTicker(hm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-hm.ctx.Done():
			return
		case <-ticker.C:
			hm.runChecks()
		}
	}
}

// runChecks runs all health checks
func (hm *HealthMonitor) runChecks() {
	hm.mu.RLock()
	checks := make(map[string]HealthCheck)
	for name, check := range hm.checks {
		checks[name] = check
	}
	hm.mu.RUnlock()

	for name, check := range checks {
		go hm.runCheck(name, check)
	}
}

// runCheck runs a single health check
func (hm *HealthMonitor) runCheck(name string, check HealthCheck) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(hm.ctx, 30*time.Second)
	defer cancel()

	result := check(ctx)
	result.Duration = time.Since(start)
	result.Timestamp = time.Now()

	hm.mu.Lock()
	hm.results[name] = result
	hm.mu.Unlock()

	// Log health check result
	fields := Fields{
		"check":    name,
		"status":   result.Status.String(),
		"duration": result.Duration,
	}

	if result.Error != nil {
		fields["error"] = result.Error.Error()
		hm.logger.WithFields(fields).Error("Health check failed")
	} else {
		hm.logger.WithFields(fields).Debug("Health check completed")
	}
}

// HealthHandler creates an HTTP handler for health checks
func (hm *HealthMonitor) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := hm.GetHealth()
		overall := hm.GetOverallHealth()

		// Set HTTP status based on health
		switch overall {
		case HealthStatusHealthy:
			w.WriteHeader(http.StatusOK)
		case HealthStatusDegraded:
			w.WriteHeader(http.StatusOK) // 200 but with degraded status
		case HealthStatusUnhealthy:
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		w.Header().Set("Content-Type", "application/json")

		// Simple JSON response
		fmt.Fprintf(w, `{
  "status": "%s",
  "timestamp": "%s",
  "checks": {`, overall.String(), time.Now().Format(time.RFC3339))

		first := true
		for name, result := range health {
			if !first {
				fmt.Fprintf(w, ",")
			}
			first = false

			fmt.Fprintf(w, `
    "%s": {
      "status": "%s",
      "message": "%s",
      "duration": "%s",
      "timestamp": "%s"`,
				name,
				result.Status.String(),
				result.Message,
				result.Duration,
				result.Timestamp.Format(time.RFC3339))

			if result.Error != nil {
				fmt.Fprintf(w, `,
      "error": "%s"`, result.Error.Error())
			}

			fmt.Fprintf(w, `
    }`)
		}

		fmt.Fprintf(w, `
  }
}`)
	}
}

// CircuitBreaker implements the circuit breaker pattern for logging handlers
type CircuitBreaker struct {
	maxFailures   int64
	timeout       time.Duration
	failures      int64
	lastFailure   time.Time
	state         CircuitState
	mu            sync.RWMutex
	onStateChange func(from, to CircuitState)
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitStateClosed CircuitState = iota
	CircuitStateOpen
	CircuitStateHalfOpen
)

func (cs CircuitState) String() string {
	switch cs {
	case CircuitStateClosed:
		return "closed"
	case CircuitStateOpen:
		return "open"
	case CircuitStateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int64, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
		state:       CircuitStateClosed,
	}
}

// SetStateChangeCallback sets a callback for state changes
func (cb *CircuitBreaker) SetStateChangeCallback(callback func(from, to CircuitState)) {
	cb.onStateChange = callback
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.RLock()
	state := cb.state
	cb.mu.RUnlock()

	if state == CircuitStateOpen {
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.setState(CircuitStateHalfOpen)
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	}

	err := fn()

	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// recordFailure records a failure
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.state == CircuitStateHalfOpen {
		cb.setState(CircuitStateOpen)
	} else if cb.failures >= cb.maxFailures {
		cb.setState(CircuitStateOpen)
	}
}

// recordSuccess records a success
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0

	if cb.state == CircuitStateHalfOpen {
		cb.setState(CircuitStateClosed)
	}
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState CircuitState) {
	oldState := cb.state
	cb.state = newState

	if cb.onStateChange != nil {
		cb.onStateChange(oldState, newState)
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// CircuitBreakerHandler wraps a handler with circuit breaker protection
type CircuitBreakerHandler struct {
	handler Handler
	breaker *CircuitBreaker
	logger  Logger
}

// NewCircuitBreakerHandler creates a new circuit breaker handler
func NewCircuitBreakerHandler(handler Handler, breaker *CircuitBreaker, logger Logger) Handler {
	cbh := &CircuitBreakerHandler{
		handler: handler,
		breaker: breaker,
		logger:  logger,
	}

	// Set up state change logging
	breaker.SetStateChangeCallback(func(from, to CircuitState) {
		logger.WithFields(Fields{
			"component":  "circuit_breaker",
			"from_state": from.String(),
			"to_state":   to.String(),
		}).Warn("Circuit breaker state changed")
	})

	return cbh
}

// Handle implements the Handler interface with circuit breaker protection
func (cbh *CircuitBreakerHandler) Handle(entry *Entry) error {
	return cbh.breaker.Execute(func() error {
		return cbh.handler.Handle(entry)
	})
}

// Common health checks
func NewHandlerHealthCheck(handler Handler) HealthCheck {
	return func(ctx context.Context) HealthCheckResult {
		// Create a test entry
		testEntry := &Entry{
			Level:   InfoLevel,
			Message: "health check test",
			Time:    time.Now(),
			Fields:  Fields{"test": true},
		}

		// Try to handle the entry
		start := time.Now()
		err := handler.Handle(testEntry)
		duration := time.Since(start)

		if err != nil {
			return HealthCheckResult{
				Status:  HealthStatusUnhealthy,
				Message: "Handler failed test",
				Error:   err,
			}
		}

		// Check response time
		if duration > 5*time.Second {
			return HealthCheckResult{
				Status:  HealthStatusDegraded,
				Message: "Handler response time is slow",
			}
		}

		return HealthCheckResult{
			Status:  HealthStatusHealthy,
			Message: "Handler is working correctly",
		}
	}
}

// Context-aware handler that respects cancellation
type ContextAwareHandler struct {
	handler Handler
	timeout time.Duration
}

// NewContextAwareHandler creates a new context-aware handler
func NewContextAwareHandler(handler Handler, timeout time.Duration) Handler {
	return &ContextAwareHandler{
		handler: handler,
		timeout: timeout,
	}
}

// Handle implements the Handler interface with context awareness
func (cah *ContextAwareHandler) Handle(entry *Entry) error {
	if entry.Context == nil {
		// No context, handle normally
		return cah.handler.Handle(entry)
	}

	// Create a timeout context if one doesn't exist
	ctx := entry.Context
	if cah.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cah.timeout)
		defer cancel()
	}

	// Handle with context cancellation support
	done := make(chan error, 1)
	go func() {
		done <- cah.handler.Handle(entry)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
