package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/logger"
)

// ResponseWriterWrapper wraps http.ResponseWriter to capture status code and response size
type ResponseWriterWrapper struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
}

func (rw *ResponseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriterWrapper) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.responseSize += int64(size)
	return size, err
}

// MetricsMiddleware wraps HTTP handlers to collect metrics
func MetricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := GenerateRequestID()
		start := time.Now()

		// Add request ID to context or headers for downstream use
		r.Header.Set("X-Request-ID", requestID)

		// Wrap response writer to capture metrics
		wrapper := &ResponseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default to 200
		}

		// Add request ID header to response
		wrapper.Header().Set("X-Request-ID", requestID)

		// Log request start
		logger.Info(fmt.Sprintf("[%s] %s %s - Request started", requestID, r.Method, r.URL.Path))

		// Execute the handler
		next(wrapper, r)

		// Calculate duration
		duration := time.Since(start)

		// Record metrics
		collector := GetMetricsCollector()
		collector.RecordResponseTime(duration.Milliseconds(), requestID)

		// Record error metrics if this was an error response
		if wrapper.statusCode >= 400 {
			endpoint := getEndpointName(r.URL.Path)
			errorCode := getErrorCodeFromStatus(wrapper.statusCode)
			collector.RecordError(errorCode, endpoint, requestID)
		}

		// Log request completion
		logger.Info(fmt.Sprintf("[%s] %s %s - %d (%v, %d bytes)",
			requestID, r.Method, r.URL.Path, wrapper.statusCode, duration, wrapper.responseSize))
	}
}

// StructuredLoggingMiddleware provides enhanced logging with request context
func StructuredLoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := GenerateRequestID()
		start := time.Now()

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Enhanced request logging
		userAgent := r.Header.Get("User-Agent")
		clientIP := getClientIP(r)

		logger.Info(fmt.Sprintf("[%s] REQUEST: %s %s | IP: %s | User-Agent: %s | Content-Length: %s",
			requestID, r.Method, r.URL.RequestURI(), clientIP, userAgent, r.Header.Get("Content-Length")))

		// Wrap response writer to capture status
		wrapper := &ResponseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Execute handler
		next(wrapper, r)

		// Enhanced response logging
		duration := time.Since(start)
		logger.Info(fmt.Sprintf("[%s] RESPONSE: %d | Size: %d bytes | Duration: %v | Endpoint: %s",
			requestID, wrapper.statusCode, wrapper.responseSize, duration, getEndpointName(r.URL.Path)))

		// Log any concerning patterns
		if duration > 5*time.Second {
			logger.Warn(fmt.Sprintf("[%s] SLOW REQUEST: %v duration for %s %s",
				requestID, duration, r.Method, r.URL.Path))
		}

		if wrapper.statusCode >= 500 {
			logger.Error(fmt.Sprintf("[%s] SERVER ERROR: %d for %s %s",
				requestID, wrapper.statusCode, r.Method, r.URL.Path))
		}
	}
}

// CircuitBreakerMiddleware integrates circuit breaker status into request handling
func CircuitBreakerMiddleware(circuitBreaker *CircuitBreaker) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			requestID := getRequestIDFromHeaders(r)

			// Update metrics with current circuit breaker state
			if circuitBreaker != nil {
				metrics := circuitBreaker.GetMetrics()
				GetMetricsCollector().UpdateCircuitBreakerMetrics(metrics)

				// For OpenAI-dependent endpoints, check circuit breaker
				if isOpenAIDependentEndpoint(r.URL.Path) && metrics.State == CircuitOpen {
					logger.Warn(fmt.Sprintf("[%s] Circuit breaker OPEN - degraded mode for %s", requestID, r.URL.Path))
					// Continue with request but it will fail gracefully in handler
				}
			}

			next(w, r)
		}
	}
}

// Helper functions

func getEndpointName(path string) string {
	if strings.HasPrefix(path, "/dalle/") {
		return "dalle"
	} else if strings.HasPrefix(path, "/series") {
		return "series"
	} else if strings.HasPrefix(path, "/health") {
		return "health"
	} else if strings.HasPrefix(path, "/metrics") {
		return "metrics"
	} else if strings.HasPrefix(path, "/preview") {
		return "preview"
	} else if path == "/readiness" {
		return "readiness"
	} else if path == "/liveness" {
		return "liveness"
	}
	return "other"
}

func getErrorCodeFromStatus(statusCode int) string {
	switch statusCode {
	case 400:
		return "BAD_REQUEST"
	case 401:
		return "UNAUTHORIZED"
	case 403:
		return "FORBIDDEN"
	case 404:
		return "NOT_FOUND"
	case 408:
		return "REQUEST_TIMEOUT"
	case 429:
		return "RATE_LIMITED"
	case 500:
		return "INTERNAL_ERROR"
	case 502:
		return "BAD_GATEWAY"
	case 503:
		return "SERVICE_UNAVAILABLE"
	case 504:
		return "GATEWAY_TIMEOUT"
	default:
		if statusCode >= 400 && statusCode < 500 {
			return "CLIENT_ERROR"
		}
		return "SERVER_ERROR"
	}
}

func getClientIP(r *http.Request) string {
	// Check various headers for the real client IP
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	cfConnectingIP := r.Header.Get("CF-Connecting-IP")
	if cfConnectingIP != "" {
		return cfConnectingIP
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}

func getRequestIDFromHeaders(r *http.Request) string {
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = GenerateRequestID()
	}
	return requestID
}

func isOpenAIDependentEndpoint(path string) bool {
	// Only /dalle/ endpoints depend on OpenAI
	return strings.HasPrefix(path, "/dalle/")
}

// WrapWithMiddleware applies all monitoring middleware to a handler
func WrapWithMiddleware(handler http.HandlerFunc, circuitBreaker *CircuitBreaker) http.HandlerFunc {
	// Apply middleware in reverse order (last applied = first executed)
	wrapped := handler
	wrapped = MetricsMiddleware(wrapped)
	wrapped = StructuredLoggingMiddleware(wrapped)

	if circuitBreaker != nil {
		wrapped = CircuitBreakerMiddleware(circuitBreaker)(wrapped)
	}

	return wrapped
}
