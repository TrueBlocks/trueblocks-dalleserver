package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
)

// Enhanced timeout constants
var (
	enhanceDeadline = 60 * time.Second
)

// OpenAIClient provides resilient OpenAI API operations
type OpenAIClient struct {
	httpClient     *http.Client
	circuitBreaker *CircuitBreaker
	retryConfig    RetryConfig
	apiKey         string
}

// NewOpenAIClient creates a new resilient OpenAI client
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		httpClient: &http.Client{
			Timeout: enhanceDeadline + 10*time.Second, // Buffer beyond context timeout
		},
		circuitBreaker: DefaultOpenAICircuitBreaker,
		retryConfig:    OpenAIRetryConfig,
		apiKey:         apiKey,
	}
}

// EnhancePromptWithResilience enhances a prompt with retry and circuit breaker protection
func (c *OpenAIClient) EnhancePromptWithResilience(prompt, authorType, requestID string) (string, error) {
	var result string

	operation := func() error {
		enhanced, err := c.enhancePromptAttempt(prompt, authorType, requestID)
		if err != nil {
			return err
		}
		result = enhanced
		return nil
	}

	// Execute with circuit breaker
	err := c.circuitBreaker.Execute(func() error {
		// Execute with retry logic
		return RetryableHTTPOperation(c.retryConfig, requestID, func() (int, error) {
			if err := operation(); err != nil {
				// Extract status code if available
				if apiErr, ok := err.(*OpenAIAPIError); ok {
					return apiErr.StatusCode, err
				}
				return 0, err
			}
			return 200, nil
		})
	})

	if err != nil {
		// Check if circuit breaker blocked the request
		if cbErr, ok := err.(*CircuitBreakerError); ok && cbErr.IsCircuitBreakerOpen() {
			logger.InfoR(fmt.Sprintf("[%s] OpenAI circuit breaker is open, using non-enhanced prompt", requestID))
			return prompt, nil // Graceful degradation
		}

		// Log error but return original prompt for graceful degradation
		logger.InfoR(fmt.Sprintf("[%s] OpenAI enhancement failed, using original prompt", requestID), "error", err)
		return prompt, nil
	}

	return result, nil
}

// enhancePromptAttempt performs a single attempt to enhance a prompt
func (c *OpenAIClient) enhancePromptAttempt(prompt, authorType, requestID string) (string, error) {
	_ = authorType // delint
	url := "https://api.openai.com/v1/chat/completions"

	payload := map[string]interface{}{
		"model":       "gpt-4",
		"temperature": 0.2,
		"seed":        1337,
		"messages": []map[string]string{
			{"role": "system", "content": prompt},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), enhanceDeadline)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-Request-ID", requestID)

	start := time.Now()
	logger.Info(fmt.Sprintf("[%s] OpenAI enhance request starting", requestID))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", &OpenAIAPIError{
			Message:    fmt.Sprintf("HTTP request failed: %v", err),
			StatusCode: 0,
			RequestID:  requestID,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &OpenAIAPIError{
			Message:    fmt.Sprintf("read response body: %v", err),
			StatusCode: resp.StatusCode,
			RequestID:  requestID,
		}
	}

	duration := time.Since(start)
	logger.Info(fmt.Sprintf("[%s] OpenAI enhance request completed", requestID),
		"durMs", duration.Milliseconds(), "status", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		// Truncate long error responses
		errorBody := string(body)
		if len(errorBody) > 512 {
			errorBody = errorBody[:512] + "..."
		}

		return "", &OpenAIAPIError{
			Message:    fmt.Sprintf("OpenAI API error: %s", errorBody),
			StatusCode: resp.StatusCode,
			RequestID:  requestID,
		}
	}

	// Parse response
	type dalleResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error,omitempty"`
	}

	var response dalleResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", &OpenAIAPIError{
			Message:    fmt.Sprintf("parse response: %v", err),
			StatusCode: resp.StatusCode,
			RequestID:  requestID,
		}
	}

	if response.Error != nil {
		return "", &OpenAIAPIError{
			Message:    fmt.Sprintf("OpenAI API error: %s", response.Error.Message),
			StatusCode: resp.StatusCode,
			RequestID:  requestID,
		}
	}

	if len(response.Choices) == 0 {
		logger.InfoR(fmt.Sprintf("[%s] OpenAI returned no choices", requestID))
		return prompt, nil // Return original
	}

	content := response.Choices[0].Message.Content
	if content == "" {
		logger.InfoR(fmt.Sprintf("[%s] OpenAI returned empty content", requestID))
		return prompt, nil // Return original
	}

	logger.InfoG(fmt.Sprintf("[%s] OpenAI enhancement successful", requestID),
		"originalLen", len(prompt), "enhancedLen", len(content))

	return content, nil
}

// GetCircuitBreakerMetrics returns current circuit breaker metrics
func (c *OpenAIClient) GetCircuitBreakerMetrics() CircuitBreakerMetrics {
	return c.circuitBreaker.GetMetrics()
}

// ResetCircuitBreaker manually resets the circuit breaker
func (c *OpenAIClient) ResetCircuitBreaker() {
	c.circuitBreaker.Reset()
}

// OpenAIAPIError represents an error from the OpenAI API
type OpenAIAPIError struct {
	Message    string
	StatusCode int
	RequestID  string
}

func (e *OpenAIAPIError) Error() string {
	return fmt.Sprintf("[%s] OpenAI API error (status %d): %s", e.RequestID, e.StatusCode, e.Message)
}

// IsRetryable determines if this error should be retried
func (e *OpenAIAPIError) IsRetryable() bool {
	return IsOpenAIRetryableError(fmt.Errorf(e.Message), e.StatusCode)
}

// Global OpenAI client instance
var globalOpenAIClient *OpenAIClient

// GetOpenAIClient returns the global OpenAI client, creating it if necessary
func GetOpenAIClient() *OpenAIClient {
	if globalOpenAIClient == nil {
		apiKey := getOpenAIAPIKey() // Need to implement this
		globalOpenAIClient = NewOpenAIClient(apiKey)
	}
	return globalOpenAIClient
}

// getOpenAIAPIKey safely retrieves the OpenAI API key
func getOpenAIAPIKey() string {
	// This will be imported from environment or config
	// For now, return empty - the client will handle missing keys gracefully
	return ""
}
