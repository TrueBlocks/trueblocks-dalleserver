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
	"github.com/TrueBlocks/trueblocks-dalle/v2/pkg/storage"
)

func main() {
	app := NewApp()

	// Fail fast if required OpenAI key missing (before starting server)
	if os.Getenv("OPENAI_API_KEY") == "" {
		fmt.Fprintln(os.Stderr, "WARNING: OPENAI_API_KEY not set; image generation will be skipped.")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.handleDefault)
	mux.HandleFunc("/dalle/", app.handleDalleDress)
	mux.HandleFunc("/series", app.handleSeries)
	mux.HandleFunc("/series/", app.handleSeries)
	mux.HandleFunc("/healthz", app.handleHealth)
	mux.HandleFunc("/metrics", app.handleMetrics)
	mux.HandleFunc("/preview", app.handlePreview)
	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(storage.OutputDir()))))

	startStatusPrinter(0)

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
		logger.InfoG(fmt.Sprintf("Starting server on %s", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.InfoR(fmt.Sprintf("Server error: %v", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		_ = srv.Close()
	}
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
			fmt.Fprintln(os.Stderr, "WARNING: invalid port number, falling back to :8080 =>", n)
		} else {
			port = ":" + n
		}
	}
	return port
}
