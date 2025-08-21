package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
)

func main() {
	app := NewApp()
	app.StartLogging()
	defer app.StopLogging()

	// Fail fast if required OpenAI key missing (before starting server)
	if os.Getenv("OPENAI_API_KEY") == "" {
		fmt.Fprintln(os.Stderr, "FATAL: OPENAI_API_KEY not set in environment (.env not loaded or missing). Exiting.")
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", app.handleDefault)
	mux.HandleFunc("/dalle/", app.handleDalleDress)
	mux.HandleFunc("/series", app.handleSeries)
	mux.HandleFunc("/series/", app.handleSeries)
	mux.HandleFunc("/healthz", handleHealth)
	mux.HandleFunc("/metrics", handleMetrics)
	mux.HandleFunc("/preview", app.handlePreview)
	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(app.OutputDir()))))

	port := getPort()
	srv := &http.Server{
		Addr:              port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second, // mitigates Slowloris (gosec G112)
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		app.Logf("Starting server on %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	app.Logf("Shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		app.Logf("Graceful shutdown failed: %v", err)
		_ = srv.Close()
	}
}

// handleHealth provides a simple health probe.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// handleMetrics exposes very basic counters (placeholder).
func handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte("dalleserver_up 1\n"))
}

func getPort() string {
	port := ":8080"
	if len(os.Args) > 1 && strings.Contains(os.Args[1], "--port=") {
		isNumeric := func(s string) bool {
			_, err := strconv.ParseFloat(s, 64)
			return err == nil
		}
		n := strings.ReplaceAll(os.Args[1], "--port=", "")
		if !isNumeric(n) {
			logger.Fatal("Invalid port number: " + n)
		}
		port = ":" + n
	}
	return port
}
