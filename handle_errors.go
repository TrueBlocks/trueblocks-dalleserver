package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// handleErrors shows error information from the metrics system
func handleErrors(w http.ResponseWriter, r *http.Request) {
	// Check for clear parameter
	if r.URL.Query().Get("clear") != "" {
		// Reset the metrics collector (create a new one)
		globalMetricsCollector = NewMetricsCollector()
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Error metrics cleared.\n")
		return
	}

	// Get current metrics
	metrics := GetMetricsCollector().GetMetrics()

	// Check if JSON format is requested
	if r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
		return
	}

	// Serve human-readable format
	w.Header().Set("Content-Type", "text/plain")

	if metrics.TotalErrors == 0 {
		fmt.Fprintf(w, "No errors recorded yet.\n")
		return
	}

	fmt.Fprintf(w, "=== DALLE Server Error Report ===\n")
	fmt.Fprintf(w, "Total Errors: %d\n", metrics.TotalErrors)
	fmt.Fprintf(w, "Last Updated: %s\n\n", metrics.LastUpdated.Format("2006-01-02 15:04:05"))

	if len(metrics.ErrorsByCode) > 0 {
		fmt.Fprintf(w, "Errors by Code:\n")
		for code, count := range metrics.ErrorsByCode {
			fmt.Fprintf(w, "  %s: %d\n", code, count)
		}
		fmt.Fprintf(w, "\n")
	}

	if len(metrics.ErrorsByEndpoint) > 0 {
		fmt.Fprintf(w, "Errors by Endpoint:\n")
		for endpoint, count := range metrics.ErrorsByEndpoint {
			fmt.Fprintf(w, "  %s: %d\n", endpoint, count)
		}
		fmt.Fprintf(w, "\n")
	}

	if metrics.OpenAIErrors > 0 {
		fmt.Fprintf(w, "OpenAI API Errors: %d (out of %d requests)\n", metrics.OpenAIErrors, metrics.OpenAIRequests)
	}

	if metrics.FileOperationErrors > 0 {
		fmt.Fprintf(w, "File Operation Errors: %d (out of %d operations)\n", metrics.FileOperationErrors, metrics.FileOperations)
	}

	fmt.Fprintf(w, "\nUse ?clear to reset error metrics\n")
	fmt.Fprintf(w, "Use ?format=json for JSON output\n")
}
