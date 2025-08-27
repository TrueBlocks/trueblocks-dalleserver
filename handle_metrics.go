package main

import "net/http"

// handleMetrics exposes very basic counters (placeholder).
func (app *App) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte("dalleserver_up 1\n"))
}
