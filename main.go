package main

import (
	"context"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/TrueBlocks/trueblocks-dalle/v6/pkg/prompt"
	"github.com/TrueBlocks/trueblocks-dalle/v6/pkg/storage"
)

const (
	brightYellow = "\033[93m"
	colorOff     = "\033[0m"
)

func logInfo(args ...any) {
	stdlog.Println(args...)
}

func logWarn(args ...any) {
	stdlog.Println(args...)
}

func logError(args ...any) {
	stdlog.Println(args...)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

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
	mux.HandleFunc("/v1/images/generate", WrapWithMiddleware(app.handleV1ImagesGenerate, circuitBreaker))
	mux.HandleFunc("/v1/images/preview", WrapWithMiddleware(app.handleV1ImagesPreview, circuitBreaker))
	mux.HandleFunc("/v1/images/", WrapWithMiddleware(app.handleV1Image, circuitBreaker))
	mux.HandleFunc("/v1/images", WrapWithMiddleware(app.handleV1Images, circuitBreaker))
	mux.HandleFunc("/v1/series/", WrapWithMiddleware(app.handleV1SeriesItem, circuitBreaker))
	mux.HandleFunc("/v1/series", WrapWithMiddleware(app.handleV1Series, circuitBreaker))
	mux.HandleFunc("/v1/databases/", WrapWithMiddleware(app.handleV1Database, circuitBreaker))
	mux.HandleFunc("/v1/databases", WrapWithMiddleware(app.handleV1Databases, circuitBreaker))
	mux.HandleFunc("/v1/validate", WrapWithMiddleware(app.handleV1Validate, circuitBreaker))
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
		logInfo(fmt.Sprintf("Starting server on %s", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logInfo(fmt.Sprintf("Server error: %v", err))
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
	logInfo("=== TrueBlocks DALLE Server Startup Report ===")
	logInfo(fmt.Sprintf("Version: %s", Version))
	logInfo(fmt.Sprintf("Build Time: %s", BuildTime))
	logInfo(fmt.Sprintf("Build Commit: %s", BuildCommit))
	logInfo(fmt.Sprintf("Build Branch: %s", BuildBranch))
	logInfo(fmt.Sprintf("Go Version: %s", runtime.Version()))
	logInfo(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	logInfo(fmt.Sprintf("Start Time: %s", time.Now().Format("2006-01-02 15:04:05 MST")))

	logInfo("--- Database Information ---")
	cm := storage.GetCacheManager()
	if err := cm.LoadOrBuild(); err != nil {
		logInfo(fmt.Sprintf("ERROR: Failed to load cache: %v", err))
		return
	}

	totalRecords := 0
	for _, dbName := range prompt.DatabaseNames {
		if dbIndex, err := cm.GetDatabase(dbName); err == nil {
			recordCount := len(dbIndex.Records)
			totalRecords += recordCount
			logInfo(fmt.Sprintf("Database: %-12s Records: %3d Version: %s", dbName, recordCount, dbIndex.Version))
		} else {
			logInfo(fmt.Sprintf("Database: %-12s Error: %v", dbName, err))
		}
	}

	logInfo(fmt.Sprintf("Total Records: %d across %d databases", totalRecords, len(prompt.DatabaseNames)))
	logInfo("============================================")
}
