package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/TrueBlocks/trueblocks-art/packages/ai"
	"github.com/TrueBlocks/trueblocks-art/packages/creds"
)

const openAIEnhancementModel = "gpt-4"
const openAIEnhancementOperation = "openai_chat_completions"

var enhanceDeadline = 60 * time.Second

type OpenAIClient struct {
	httpClient     *http.Client
	circuitBreaker *CircuitBreaker
	retryConfig    RetryConfig
	apiKey         string
	model          string
	endpoint       string
	timeout        time.Duration
	metrics        *MetricsCollector
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		httpClient:     &http.Client{Timeout: enhanceDeadline + 10*time.Second},
		circuitBreaker: DefaultOpenAICircuitBreaker,
		retryConfig:    OpenAIRetryConfig,
		apiKey:         apiKey,
		model:          openAIEnhancementModel,
		timeout:        enhanceDeadline,
		metrics:        GetMetricsCollector(),
	}
}

func (c *OpenAIClient) EnhancePromptWithResilience(prmt, authorType, requestID string) (string, error) {
	return c.EnhancePromptWithContext(context.Background(), prmt, authorType, requestID)
}

func (c *OpenAIClient) EnhancePromptWithContext(ctx context.Context, prmt, _ string, requestID string) (string, error) {
	content := prmt
	err := c.circuitBreaker.Execute(func() error {
		if c.timeout <= 0 {
			return context.DeadlineExceeded
		}
		spec, ok := ai.LookupModel(c.model)
		if !ok || spec.Provider != ai.ProviderOpenAI || !spec.Writes {
			return fmt.Errorf("unsupported enhancement model %q in shared registry", c.model)
		}
		seed, temperature := 1337, 0.2
		attempts := max(1, c.retryConfig.MaxAttempts)
		timeout := time.Duration(attempts)*c.timeout + time.Duration(attempts-1)*c.retryConfig.MaxDelay*5/4
		provider := &ai.OpenAI{
			APIKey: c.apiKey, HTTPClient: c.httpClient, ChatURL: c.endpoint,
			MaxRetries: attempts, Pricing: ai.ProviderPricing(ai.ProviderOpenAI),
			RetryDelay: func(attempt int) time.Duration { return calculateBackoffDelay(c.retryConfig, attempt) },
			OnAttempt:  func(attempt int, _ *ai.Result, err error) { c.recordAttempt(attempt, err, requestID) },
		}
		start := time.Now()
		result, err := provider.Call(ctx, c.model, "", ai.CallOptions{
			System: prmt, SystemOnly: true, Seed: &seed, Temperature: &temperature,
			MaxTokens: -1, Timeout: timeout, AttemptTimeout: c.timeout, RequestID: requestID,
		})
		if err != nil && !errors.Is(err, ai.ErrNoResponse) {
			return err
		}
		if result.UsageReported {
			if err := ai.RecordCall("dalleserver", result, time.Since(start).Seconds()); err != nil {
				logInfo("record enhancement usage", "error", err)
			}
		}
		if result.Content != "" {
			content = result.Content
		}
		return nil
	})
	c.metrics.UpdateCircuitBreakerMetrics(c.circuitBreaker.GetMetrics())
	if err != nil {
		logInfo(fmt.Sprintf("[%s] OpenAI enhancement unavailable, using original prompt", requestID), "error", err)
		return prmt, nil
	}
	return content, nil
}

func (c *OpenAIClient) recordAttempt(attempt int, err error, requestID string) {
	if attempt > 1 {
		c.metrics.RecordRetry(openAIEnhancementOperation, requestID)
	}
	success := err == nil || errors.Is(err, ai.ErrNoResponse)
	timeout := errors.Is(err, context.DeadlineExceeded)
	code := "OPENAI_ERROR"
	var apiErr *ai.APIError
	if errors.As(err, &apiErr) {
		timeout = timeout || apiErr.StatusCode == http.StatusGatewayTimeout
		if apiErr.Code != "" {
			code = apiErr.Code
		}
		if apiErr.RequestID != "" {
			logInfo(fmt.Sprintf("[%s] OpenAI attempt completed", requestID), "providerRequestID", apiErr.RequestID, "attempt", attempt)
		}
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		timeout = timeout || netErr.Timeout()
	}
	c.metrics.RecordOpenAIRequest(success, timeout, requestID)
	if !success {
		c.metrics.RecordError(code, openAIEnhancementOperation, requestID)
	}
}

func (c *OpenAIClient) GetCircuitBreakerMetrics() CircuitBreakerMetrics {
	return c.circuitBreaker.GetMetrics()
}

func (c *OpenAIClient) ResetCircuitBreaker() {
	c.circuitBreaker.Reset()
}

var globalOpenAIClient = sync.OnceValue(func() *OpenAIClient {
	return NewOpenAIClient(getOpenAIAPIKey())
})

func GetOpenAIClient() *OpenAIClient {
	return globalOpenAIClient()
}

func getOpenAIAPIKey() string {
	return creds.MustGet("OPENAI_API_KEY")
}
