package main

import (
	"crypto/rand"
	"fmt"
	"math"
	"time"

	"github.com/TrueBlocks/trueblocks-dalle/v2/pkg/prompt"
)

// RetryConfig defines configuration for retry operations
type RetryConfig struct {
	MaxAttempts   int           `json:"max_attempts"`
	BaseDelay     time.Duration `json:"base_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// DefaultRetryConfig provides sensible defaults for retry operations
var DefaultRetryConfig = RetryConfig{
	MaxAttempts:   3,
	BaseDelay:     1 * time.Second,
	MaxDelay:      30 * time.Second,
	BackoffFactor: 2.0,
}

// OpenAIRetryConfig provides specific retry settings for OpenAI API calls
var OpenAIRetryConfig = RetryConfig{
	MaxAttempts:   3,
	BaseDelay:     2 * time.Second,
	MaxDelay:      60 * time.Second,
	BackoffFactor: 2.0,
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err       error
	Retryable bool
	Temporary bool
}

func (r *RetryableError) Error() string {
	return r.Err.Error()
}

func (r *RetryableError) Unwrap() error {
	return r.Err
}

// RetryWithBackoff executes a function with exponential backoff retry logic
func RetryWithBackoff(config RetryConfig, operation func() error) error {
	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if this is the last attempt
		if attempt >= config.MaxAttempts {
			break
		}

		// Calculate delay with exponential backoff and jitter
		delay := calculateBackoffDelay(config, attempt)

		// Log retry attempt
		fmt.Printf("Retry attempt %d/%d after error: %v (waiting %v)\n",
			attempt, config.MaxAttempts, err, delay)

		time.Sleep(delay)
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// calculateBackoffDelay calculates delay with exponential backoff and jitter
func calculateBackoffDelay(config RetryConfig, attempt int) time.Duration {
	// Exponential backoff: base * factor^(attempt-1)
	delay := float64(config.BaseDelay) * math.Pow(config.BackoffFactor, float64(attempt-1))

	// Cap at max delay
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	// Add jitter (±25% randomization) using crypto/rand
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to no jitter if crypto/rand fails
		return time.Duration(delay)
	}

	// Convert to float64 between 0 and 1
	randomUint64 := uint64(0)
	for i, b := range randomBytes {
		randomUint64 |= uint64(b) << (i * 8)
	}
	randomFloat := float64(randomUint64) / float64(^uint64(0))

	// Scale to ±25%
	jitter := delay * 0.25 * (randomFloat*2 - 1) // Random between -0.25 and +0.25
	finalDelay := delay + jitter

	// Ensure minimum delay
	if finalDelay < float64(config.BaseDelay)/2 {
		finalDelay = float64(config.BaseDelay) / 2
	}

	return time.Duration(finalDelay)
}

// RetryableHTTPOperation wraps an HTTP operation with retry logic
func RetryableHTTPOperation(config RetryConfig, requestID string, operation func() (int, error)) error {
	return RetryWithBackoff(config, func() error {
		statusCode, err := operation()
		if err != nil {
			if prompt.IsOpenAIRetryableError(err, statusCode) {
				return fmt.Errorf("[%s] retryable error (status %d): %w", requestID, statusCode, err)
			}
			return fmt.Errorf("[%s] non-retryable error (status %d): %w", requestID, statusCode, err)
		}
		return nil
	})
}
