package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/logger"
	"github.com/TrueBlocks/trueblocks-dalle/v6/pkg/prompt"
	"github.com/TrueBlocks/trueblocks-dalle/v6/pkg/storage"
)

func main() {
	app := NewApp()

	// Fail fast if required OpenAI key missing (before starting server)
	if os.Getenv("OPENAI_API_KEY") == "" {
		panic("OPENAI_API_KEY is required but not set. Please configure your OpenAI API key and try again.")
	}

	// Initialize circuit breaker for OpenAI
	circuitBreaker := NewCircuitBreaker(5, 30*time.Second)

	// Initialize health checker with circuit breaker
	GetHealthChecker().SetCircuitBreaker(circuitBreaker)

	printStartupReport()

	mux := http.NewServeMux()

	// Apply middleware to all handlers
	mux.HandleFunc("/", WrapWithMiddleware(app.handleDefault, circuitBreaker))
	mux.HandleFunc("/dalle/", WrapWithMiddleware(app.handleDalleDress, circuitBreaker))
	mux.HandleFunc("/series", WrapWithMiddleware(app.handleSeries, circuitBreaker))
	mux.HandleFunc("/series/", WrapWithMiddleware(app.handleSeries, circuitBreaker))
	mux.HandleFunc("/health", WrapWithMiddleware(app.handleHealth, circuitBreaker))
	mux.HandleFunc("/metrics", WrapWithMiddleware(app.handleMetrics, circuitBreaker))
	mux.HandleFunc("/preview", WrapWithMiddleware(app.handlePreview, circuitBreaker))
	mux.HandleFunc("/errors", WrapWithMiddleware(handleErrors, circuitBreaker))
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

// Build-time variables (set via ldflags during build)
var (
	BuildTime   = "unknown"
	BuildCommit = "unknown"
	BuildBranch = "unknown"
	Version     = "development"
)

// printStartupReport displays build and runtime information when the server starts
func printStartupReport() {
	logger.InfoG("=== TrueBlocks DALLE Server Startup Report ===")
	logger.InfoG(fmt.Sprintf("Version: %s", Version))
	logger.InfoG(fmt.Sprintf("Build Time: %s", BuildTime))
	logger.InfoG(fmt.Sprintf("Build Commit: %s", BuildCommit))
	logger.InfoG(fmt.Sprintf("Build Branch: %s", BuildBranch))
	logger.InfoG(fmt.Sprintf("Go Version: %s", runtime.Version()))
	logger.InfoG(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	logger.InfoG(fmt.Sprintf("Start Time: %s", time.Now().Format("2006-01-02 15:04:05 MST")))

	logger.InfoG("--- Database Information ---")
	cm := storage.GetCacheManager()
	if err := cm.LoadOrBuild(); err != nil {
		logger.InfoG(fmt.Sprintf("ERROR: Failed to load cache: %v", err))
		return
	}

	totalRecords := 0
	for _, dbName := range prompt.DatabaseNames {
		if dbIndex, err := cm.GetDatabase(dbName); err == nil {
			recordCount := len(dbIndex.Records)
			totalRecords += recordCount
			logger.InfoG(fmt.Sprintf("Database: %-12s Records: %3d Version: %s", dbName, recordCount, dbIndex.Version))
		} else {
			logger.InfoG(fmt.Sprintf("Database: %-12s Error: %v", dbName, err))
		}
	}

	logger.InfoG(fmt.Sprintf("Total Records: %d across %d databases", totalRecords, len(prompt.DatabaseNames)))
	logger.InfoG("============================================")
}
