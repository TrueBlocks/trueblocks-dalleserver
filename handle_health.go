package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/logger"
)

// Enhanced health handler with support for liveness and readiness checks via query parameters
func (app *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	requestID := GenerateRequestID()
	start := time.Now()

	// Check for query parameters to determine the type of health check
	checkType := r.URL.Query().Get("check")

	switch checkType {
	case "liveness":
		// Simple liveness check - if we can respond, we're alive
		w.WriteHeader(http.StatusOK)
		WriteSuccessResponse(w, map[string]interface{}{
			"status":    "alive",
			"timestamp": time.Now(),
			"uptime":    time.Since(globalHealthChecker.startTime).Seconds(),
		}, requestID)

		logger.Info(fmt.Sprintf("[%s] Liveness check completed", requestID))
		return

	case "readiness":
		// Readiness check - perform health check but only return ready/not ready
		healthCheck := GetHealthChecker().CheckHealth(requestID)

		// Service is ready if not unhealthy
		if healthCheck.Status == HealthStatusUnhealthy {
			WriteErrorResponse(w, NewAPIError("SERVICE_UNAVAILABLE", "Service not ready", "Health check failed"), http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		WriteSuccessResponse(w, map[string]interface{}{
			"status":    "ready",
			"timestamp": time.Now(),
		}, requestID)

		logger.Info(fmt.Sprintf("[%s] Readiness check completed: ready", requestID))
		return

	default:
		// Full enhanced health check (default behavior)
		// Perform health check
		healthCheck := GetHealthChecker().CheckHealth(requestID)

		// Record response time
		duration := time.Since(start)
		GetMetricsCollector().RecordResponseTime(duration.Milliseconds(), requestID)

		// Set appropriate HTTP status code based on health
		var statusCode int
		switch healthCheck.Status {
		case HealthStatusHealthy:
			statusCode = http.StatusOK
		case HealthStatusDegraded:
			statusCode = http.StatusOK // Still OK but with warnings
		case HealthStatusUnhealthy:
			statusCode = http.StatusServiceUnavailable
		default:
			statusCode = http.StatusInternalServerError
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		WriteSuccessResponse(w, healthCheck, requestID)

		logger.Info(fmt.Sprintf("[%s] Enhanced health check completed: %s (took %v)",
			requestID, healthCheck.Status, duration))
	}
}
