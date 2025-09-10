package main

import (
	"fmt"
	"net/http"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
)

// Metrics handler
func (app *App) handleMetrics(w http.ResponseWriter, r *http.Request) {
	requestID := GenerateRequestID()

	if r.URL.Query().Get("format") == "json" {
		// Return JSON format
		w.Header().Set("Content-Type", "application/json")
		metrics := GetMetricsCollector().GetMetrics()

		WriteSuccessResponse(w, metrics, requestID)
	} else {
		// Return Prometheus format (default)
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		prometheusMetrics := GetMetricsCollector().PrometheusMetrics()

		// Add request ID as comment
		_, _ = w.Write([]byte(fmt.Sprintf("# Request ID: %s\n", requestID)))
		_, _ = w.Write([]byte(prometheusMetrics))
	}

	logger.Info(fmt.Sprintf("[%s] Metrics request served", requestID))
}
