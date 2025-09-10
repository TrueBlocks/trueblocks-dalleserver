package main

import (
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // Normal operation
	CircuitOpen                         // Failing, rejecting requests
	CircuitHalfOpen                     // Testing if service recovered
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "CLOSED"
	case CircuitOpen:
		return "OPEN"
	case CircuitHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           CircuitState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	lastSuccessTime time.Time

	// Configuration
	failureThreshold int           // Number of failures before opening
	resetTimeout     time.Duration // Time to wait before attempting half-open
	successThreshold int           // Successes needed in half-open to close

	// Metrics
	totalRequests  int64
	totalFailures  int64
	totalSuccesses int64
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            CircuitClosed,
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		successThreshold: 2, // Require 2 successes to close circuit
	}
}

// DefaultOpenAICircuitBreaker creates a circuit breaker optimized for OpenAI API
var DefaultOpenAICircuitBreaker = NewCircuitBreaker(5, 60*time.Second)

// Execute runs the given operation through the circuit breaker
func (cb *CircuitBreaker) Execute(operation func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalRequests++

	// Check if circuit should transition from open to half-open
	if cb.state == CircuitOpen && time.Since(cb.lastFailureTime) > cb.resetTimeout {
		cb.state = CircuitHalfOpen
		cb.successCount = 0
		fmt.Printf("Circuit breaker transitioning to HALF_OPEN after %v\n", cb.resetTimeout)
	}

	// Reject requests if circuit is open
	if cb.state == CircuitOpen {
		return &CircuitBreakerError{
			Message: "circuit breaker is OPEN - service unavailable",
			State:   cb.state,
		}
	}

	// Execute the operation
	err := operation()

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

// onFailure handles a failure case
func (cb *CircuitBreaker) onFailure() {
	cb.totalFailures++
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	// Transition to open if failure threshold exceeded
	if cb.state == CircuitClosed && cb.failureCount >= cb.failureThreshold {
		cb.state = CircuitOpen
		fmt.Printf("Circuit breaker OPENED after %d failures\n", cb.failureCount)
	} else if cb.state == CircuitHalfOpen {
		// Failed while testing - go back to open
		cb.state = CircuitOpen
		cb.failureCount = cb.failureThreshold // Reset to threshold
		fmt.Printf("Circuit breaker returned to OPEN state after half-open failure\n")
	}
}

// onSuccess handles a success case
func (cb *CircuitBreaker) onSuccess() {
	cb.totalSuccesses++
	cb.lastSuccessTime = time.Now()

	switch cb.state {
	case CircuitHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			// Enough successes - close the circuit
			cb.state = CircuitClosed
			cb.failureCount = 0
			cb.successCount = 0
			fmt.Printf("Circuit breaker CLOSED after %d successful requests\n", cb.successThreshold)
		}
	case CircuitClosed:
		// Reset failure count on success
		cb.failureCount = 0
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitBreakerMetrics{
		State:            cb.state,
		TotalRequests:    cb.totalRequests,
		TotalFailures:    cb.totalFailures,
		TotalSuccesses:   cb.totalSuccesses,
		FailureCount:     cb.failureCount,
		SuccessCount:     cb.successCount,
		LastFailureTime:  cb.lastFailureTime,
		LastSuccessTime:  cb.lastSuccessTime,
		FailureThreshold: cb.failureThreshold,
		ResetTimeout:     cb.resetTimeout,
	}
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitClosed
	cb.failureCount = 0
	cb.successCount = 0
	fmt.Printf("Circuit breaker manually reset to CLOSED state\n")
}

// CircuitBreakerMetrics holds metrics about circuit breaker performance
type CircuitBreakerMetrics struct {
	State            CircuitState  `json:"state"`
	TotalRequests    int64         `json:"total_requests"`
	TotalFailures    int64         `json:"total_failures"`
	TotalSuccesses   int64         `json:"total_successes"`
	FailureCount     int           `json:"current_failure_count"`
	SuccessCount     int           `json:"current_success_count"`
	LastFailureTime  time.Time     `json:"last_failure_time"`
	LastSuccessTime  time.Time     `json:"last_success_time"`
	FailureThreshold int           `json:"failure_threshold"`
	ResetTimeout     time.Duration `json:"reset_timeout"`
}

// CircuitBreakerError represents an error from circuit breaker
type CircuitBreakerError struct {
	Message string
	State   CircuitState
}

func (e *CircuitBreakerError) Error() string {
	return fmt.Sprintf("circuit breaker error [%s]: %s", e.State, e.Message)
}

// IsCircuitBreakerOpen checks if the circuit breaker is preventing operations
func (e *CircuitBreakerError) IsCircuitBreakerOpen() bool {
	return e.State == CircuitOpen
}
