package parser

import (
	"errors"
	"sync"
	"time"

	"webtracker-bot/internal/logger"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements an exponential backoff circuit breaker.
type CircuitBreaker struct {
	mu sync.Mutex

	state        CircuitState
	failures     int
	maxFailures  int
	openDuration time.Duration
	openedAt     time.Time

	backoff    time.Duration
	maxBackoff time.Duration
}

var ErrCircuitOpen = errors.New("circuit breaker is open")

// NewCircuitBreaker creates a new CircuitBreaker.
func NewCircuitBreaker(maxFailures int, initialOpenDuration, maxBackoff time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        StateClosed,
		maxFailures:  maxFailures,
		openDuration: initialOpenDuration,
		backoff:      initialOpenDuration,
		maxBackoff:   maxBackoff,
	}
}

// Allow checks if a request is allowed to proceed.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil
	case StateOpen:
		if time.Since(cb.openedAt) >= cb.backoff {
			cb.state = StateHalfOpen
			logger.Warn().Msg("Circuit breaker transitioned to Half-Open")
			return nil
		}
		return ErrCircuitOpen
	case StateHalfOpen:
		// Only allow one request in half-open state, but returning nil here allows it.
		// If multiple requests hit exactly at the same time, they all might be allowed,
		// but that's acceptable for a soft half-open.
		return nil
	}

	return nil
}

// RecordSuccess records a successful request.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen || cb.failures > 0 {
		logger.Info().Msg("Circuit breaker reset to Closed")
		cb.state = StateClosed
		cb.failures = 0
		cb.backoff = cb.openDuration // reset backoff
	}
}

// RecordFailure records a failed request.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++

	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		cb.openedAt = time.Now()
		// Exponential backoff
		cb.backoff *= 2
		if cb.backoff > cb.maxBackoff {
			cb.backoff = cb.maxBackoff
		}
		logger.Warn().Dur("backoff", cb.backoff).Msg("Circuit breaker failed in Half-Open, returning to Open")
		return
	}

	if cb.state == StateClosed && cb.failures >= cb.maxFailures {
		cb.state = StateOpen
		cb.openedAt = time.Now()
		logger.Error().Int("failures", cb.failures).Msg("Circuit breaker tripped to Open")
	}
}
