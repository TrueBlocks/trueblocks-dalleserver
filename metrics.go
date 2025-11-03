package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/logger"
)

// ErrorMetrics tracks error-related metrics
type ErrorMetrics struct {
	mu sync.RWMutex

	// Error counts by type
	TotalErrors      int64            `json:"total_errors"`
	ErrorsByCode     map[string]int64 `json:"errors_by_code"`
	ErrorsByEndpoint map[string]int64 `json:"errors_by_endpoint"`

	// Circuit breaker metrics
	CircuitBreakerState     string `json:"circuit_breaker_state"`
	CircuitBreakerFailures  int64  `json:"circuit_breaker_failures"`
	CircuitBreakerSuccesses int64  `json:"circuit_breaker_successes"`

	// Retry metrics
	TotalRetries       int64            `json:"total_retries"`
	RetriesByOperation map[string]int64 `json:"retries_by_operation"`

	// Response time metrics (in milliseconds)
	ResponseTimes *ResponseTimeMetrics `json:"response_times"`

	// OpenAI API metrics
	OpenAIRequests int64 `json:"openai_requests"`
	OpenAIErrors   int64 `json:"openai_errors"`
	OpenAITimeouts int64 `json:"openai_timeouts"`

	// File system metrics
	FileOperations      int64 `json:"file_operations"`
	FileOperationErrors int64 `json:"file_operation_errors"`

	LastUpdated time.Time `json:"last_updated"`
}

// ErrorMetricsSnapshot is a serializable version without mutex
type ErrorMetricsSnapshot struct {
	// Error counts by type
	TotalErrors      int64            `json:"total_errors"`
	ErrorsByCode     map[string]int64 `json:"errors_by_code"`
	ErrorsByEndpoint map[string]int64 `json:"errors_by_endpoint"`

	// Circuit breaker metrics
	CircuitBreakerState     string `json:"circuit_breaker_state"`
	CircuitBreakerFailures  int64  `json:"circuit_breaker_failures"`
	CircuitBreakerSuccesses int64  `json:"circuit_breaker_successes"`

	// Retry metrics
	TotalRetries       int64            `json:"total_retries"`
	RetriesByOperation map[string]int64 `json:"retries_by_operation"`

	// Response time metrics (in milliseconds)
	ResponseTimes *ResponseTimeMetrics `json:"response_times"`

	// OpenAI API metrics
	OpenAIRequests int64 `json:"openai_requests"`
	OpenAIErrors   int64 `json:"openai_errors"`
	OpenAITimeouts int64 `json:"openai_timeouts"`

	// File system metrics
	FileOperations      int64 `json:"file_operations"`
	FileOperationErrors int64 `json:"file_operation_errors"`

	LastUpdated time.Time `json:"last_updated"`
}

// ResponseTimeMetrics tracks response time statistics
type ResponseTimeMetrics struct {
	Count int64   `json:"count"`
	Sum   int64   `json:"sum_ms"`
	Min   int64   `json:"min_ms"`
	Max   int64   `json:"max_ms"`
	Avg   float64 `json:"avg_ms"`
	P95   int64   `json:"p95_ms"`
	P99   int64   `json:"p99_ms"`

	// Sliding window for percentile calculation
	samples []int64
}

// MetricsCollector manages all error and performance metrics
type MetricsCollector struct {
	metrics ErrorMetrics
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: ErrorMetrics{
			ErrorsByCode:       make(map[string]int64),
			ErrorsByEndpoint:   make(map[string]int64),
			RetriesByOperation: make(map[string]int64),
			ResponseTimes: &ResponseTimeMetrics{
				Min:     int64(^uint64(0) >> 1), // Max int64
				samples: make([]int64, 0, 1000), // Keep last 1000 samples
			},
			LastUpdated: time.Now(),
		},
	}
}

// RecordError records an error occurrence
func (mc *MetricsCollector) RecordError(errorCode, endpoint, requestID string) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.TotalErrors++
	mc.metrics.ErrorsByCode[errorCode]++
	mc.metrics.ErrorsByEndpoint[endpoint]++
	mc.metrics.LastUpdated = time.Now()

	logger.Info(fmt.Sprintf("[%s] Error recorded: %s on %s", requestID, errorCode, endpoint))
}

// RecordRetry records a retry attempt
func (mc *MetricsCollector) RecordRetry(operation, requestID string) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.TotalRetries++
	mc.metrics.RetriesByOperation[operation]++
	mc.metrics.LastUpdated = time.Now()

	logger.Info(fmt.Sprintf("[%s] Retry recorded for operation: %s", requestID, operation))
}

// RecordResponseTime records a response time measurement
func (mc *MetricsCollector) RecordResponseTime(durationMs int64, requestID string) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	rt := mc.metrics.ResponseTimes
	rt.Count++
	rt.Sum += durationMs

	if durationMs < rt.Min {
		rt.Min = durationMs
	}
	if durationMs > rt.Max {
		rt.Max = durationMs
	}

	rt.Avg = float64(rt.Sum) / float64(rt.Count)

	// Add to samples for percentile calculation
	rt.samples = append(rt.samples, durationMs)
	if len(rt.samples) > 1000 {
		// Keep only last 1000 samples
		rt.samples = rt.samples[1:]
	}

	// Calculate percentiles if we have enough samples
	if len(rt.samples) >= 20 {
		rt.P95 = mc.calculatePercentile(rt.samples, 95)
		rt.P99 = mc.calculatePercentile(rt.samples, 99)
	}

	mc.metrics.LastUpdated = time.Now()
}

// RecordOpenAIRequest records an OpenAI API request
func (mc *MetricsCollector) RecordOpenAIRequest(success bool, timeout bool, requestID string) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.OpenAIRequests++
	if !success {
		mc.metrics.OpenAIErrors++
	}
	if timeout {
		mc.metrics.OpenAITimeouts++
	}
	mc.metrics.LastUpdated = time.Now()
}

// RecordFileOperation records a file system operation
func (mc *MetricsCollector) RecordFileOperation(success bool, requestID string) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.FileOperations++
	if !success {
		mc.metrics.FileOperationErrors++
	}
	mc.metrics.LastUpdated = time.Now()
}

// UpdateCircuitBreakerMetrics updates circuit breaker state
func (mc *MetricsCollector) UpdateCircuitBreakerMetrics(metrics CircuitBreakerMetrics) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.CircuitBreakerState = metrics.State.String()
	mc.metrics.CircuitBreakerFailures = metrics.TotalFailures
	mc.metrics.CircuitBreakerSuccesses = metrics.TotalSuccesses
	mc.metrics.LastUpdated = time.Now()
}

// GetMetrics returns a copy of current metrics
func (mc *MetricsCollector) GetMetrics() ErrorMetricsSnapshot {
	mc.metrics.mu.RLock()
	defer mc.metrics.mu.RUnlock()

	// Deep copy maps
	errorsByCode := make(map[string]int64)
	for k, v := range mc.metrics.ErrorsByCode {
		errorsByCode[k] = v
	}

	errorsByEndpoint := make(map[string]int64)
	for k, v := range mc.metrics.ErrorsByEndpoint {
		errorsByEndpoint[k] = v
	}

	retriesByOperation := make(map[string]int64)
	for k, v := range mc.metrics.RetriesByOperation {
		retriesByOperation[k] = v
	}

	// Deep copy response times
	rt := &ResponseTimeMetrics{
		Count: mc.metrics.ResponseTimes.Count,
		Sum:   mc.metrics.ResponseTimes.Sum,
		Min:   mc.metrics.ResponseTimes.Min,
		Max:   mc.metrics.ResponseTimes.Max,
		Avg:   mc.metrics.ResponseTimes.Avg,
		P95:   mc.metrics.ResponseTimes.P95,
		P99:   mc.metrics.ResponseTimes.P99,
	}

	return ErrorMetricsSnapshot{
		TotalErrors:             mc.metrics.TotalErrors,
		ErrorsByCode:            errorsByCode,
		ErrorsByEndpoint:        errorsByEndpoint,
		CircuitBreakerState:     mc.metrics.CircuitBreakerState,
		CircuitBreakerFailures:  mc.metrics.CircuitBreakerFailures,
		CircuitBreakerSuccesses: mc.metrics.CircuitBreakerSuccesses,
		TotalRetries:            mc.metrics.TotalRetries,
		RetriesByOperation:      retriesByOperation,
		ResponseTimes:           rt,
		OpenAIRequests:          mc.metrics.OpenAIRequests,
		OpenAIErrors:            mc.metrics.OpenAIErrors,
		OpenAITimeouts:          mc.metrics.OpenAITimeouts,
		FileOperations:          mc.metrics.FileOperations,
		FileOperationErrors:     mc.metrics.FileOperationErrors,
		LastUpdated:             mc.metrics.LastUpdated,
	}
}

// calculatePercentile calculates the specified percentile from samples
func (mc *MetricsCollector) calculatePercentile(samples []int64, percentile int) int64 {
	if len(samples) == 0 {
		return 0
	}

	// Simple percentile calculation (could be optimized with proper sorting)
	sortedSamples := make([]int64, len(samples))
	copy(sortedSamples, samples)

	// Simple bubble sort for now (good enough for 1000 samples)
	for i := 0; i < len(sortedSamples); i++ {
		for j := 0; j < len(sortedSamples)-1; j++ {
			if sortedSamples[j] > sortedSamples[j+1] {
				sortedSamples[j], sortedSamples[j+1] = sortedSamples[j+1], sortedSamples[j]
			}
		}
	}

	index := (percentile * len(sortedSamples)) / 100
	if index >= len(sortedSamples) {
		index = len(sortedSamples) - 1
	}

	return sortedSamples[index]
}

// PrometheusMetrics generates Prometheus-format metrics
func (mc *MetricsCollector) PrometheusMetrics() string {
	metrics := mc.GetMetrics()

	var result string

	// Basic counters
	result += fmt.Sprintf("dalleserver_errors_total %d\n", metrics.TotalErrors)
	result += fmt.Sprintf("dalleserver_retries_total %d\n", metrics.TotalRetries)
	result += fmt.Sprintf("dalleserver_openai_requests_total %d\n", metrics.OpenAIRequests)
	result += fmt.Sprintf("dalleserver_openai_errors_total %d\n", metrics.OpenAIErrors)
	result += fmt.Sprintf("dalleserver_openai_timeouts_total %d\n", metrics.OpenAITimeouts)
	result += fmt.Sprintf("dalleserver_file_operations_total %d\n", metrics.FileOperations)
	result += fmt.Sprintf("dalleserver_file_operation_errors_total %d\n", metrics.FileOperationErrors)

	// Circuit breaker state (1 for current state, 0 for others)
	states := []string{"CLOSED", "OPEN", "HALF_OPEN"}
	for _, state := range states {
		value := 0
		if metrics.CircuitBreakerState == state {
			value = 1
		}
		result += fmt.Sprintf("dalleserver_circuit_breaker_state{state=\"%s\"} %d\n", state, value)
	}

	// Response time metrics
	if metrics.ResponseTimes.Count > 0 {
		result += fmt.Sprintf("dalleserver_response_time_ms{quantile=\"avg\"} %.2f\n", metrics.ResponseTimes.Avg)
		result += fmt.Sprintf("dalleserver_response_time_ms{quantile=\"min\"} %d\n", metrics.ResponseTimes.Min)
		result += fmt.Sprintf("dalleserver_response_time_ms{quantile=\"max\"} %d\n", metrics.ResponseTimes.Max)
		result += fmt.Sprintf("dalleserver_response_time_ms{quantile=\"0.95\"} %d\n", metrics.ResponseTimes.P95)
		result += fmt.Sprintf("dalleserver_response_time_ms{quantile=\"0.99\"} %d\n", metrics.ResponseTimes.P99)
	}

	// Error breakdowns by code
	for code, count := range metrics.ErrorsByCode {
		// sanitize code for label value (very simple: replace quotes)
		safeCode := strings.ReplaceAll(code, "\"", "")
		result += fmt.Sprintf("dalleserver_error_code_total{code=\"%s\"} %d\n", safeCode, count)
	}
	// Error breakdowns by endpoint
	for ep, count := range metrics.ErrorsByEndpoint {
		// basic sanitation
		safeEp := strings.ReplaceAll(ep, "\"", "")
		result += fmt.Sprintf("dalleserver_error_endpoint_total{endpoint=\"%s\"} %d\n", safeEp, count)
	}

	// Always include the basic up metric
	result += "dalleserver_up 1\n"

	return result
}

// Global metrics collector
var globalMetricsCollector = NewMetricsCollector()

// GetMetricsCollector returns the global metrics collector
func GetMetricsCollector() *MetricsCollector {
	return globalMetricsCollector
}
