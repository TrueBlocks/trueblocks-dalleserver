package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
	"golang.org/x/term"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type App struct {
	ValidSeries []string
	Logger      *log.Logger
	logCloser   io.Closer
	Config      Config
}

// colorStripWriter removes ANSI escape sequences before writing (used for file logs).
type colorStripWriter struct{ w io.Writer }

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func (c colorStripWriter) Write(p []byte) (int, error) {
	clean := ansiRegexp.ReplaceAll(p, nil)
	return c.w.Write(clean)
}

func NewApp() *App {
	app := App{Config: MustLoadConfig()}
	_ = os.MkdirAll(dalle.OutputDir(), 0o750)
	app.ValidSeries = dalle.ListSeries()
	return &app
}

func logsDir() string { return filepath.Join(dalle.DataDir(), "logs") }

// StartLogging initializes the rotating logger. Optionally pass a positive override size (MB) for tests.
func (a *App) StartLogging(optionalMaxSize ...int) {
	_ = os.MkdirAll(logsDir(), 0o750)
	lfPath := filepath.Join(logsDir(), "server.log")
	maxSize := 50 // default MB
	if envSz := os.Getenv("TB_DALLE_LOG_MAX_MB"); envSz != "" {
		if v, err := strconv.Atoi(envSz); err == nil && v > 0 {
			maxSize = v
		}
	}
	if len(optionalMaxSize) > 0 && optionalMaxSize[0] > 0 {
		maxSize = optionalMaxSize[0]
	}
	rotator := &lumberjack.Logger{
		Filename:   lfPath,
		MaxSize:    maxSize, // megabytes
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   false,
	}
	a.logCloser = rotator

	// Writer for file (strip ANSI). Always strip for file.
	fileWriter := colorStripWriter{w: rotator}

	// Optional silent mode for tests to avoid flooding `go test` output (which caused
	// unexplained non-zero exit when very large log output was emitted). When
	// TB_DALLE_SILENT_LOG=1 we do not mirror logs to stderr.
	silent := os.Getenv("TB_DALLE_SILENT_LOG") == "1"
	var serverMW io.Writer
	if silent {
		serverMW = fileWriter
	} else {
		stderrWriter := io.Writer(os.Stderr)
		if !term.IsTerminal(int(os.Stderr.Fd())) { // redirected
			stderrWriter = colorStripWriter{w: os.Stderr}
		}
		serverMW = io.MultiWriter(fileWriter, stderrWriter)
	}
	a.Logger = log.New(serverMW, "", log.LstdFlags|log.Lmicroseconds)
	a.Logger.Printf("logging started (rotating): %s", lfPath)

	// Redirect core logger to same sinks (with file color stripping already handled)
	logger.SetLoggerWriter(serverMW)
}

func (a *App) StopLogging() {
	if a.logCloser != nil {
		_ = a.logCloser.Close()
	}
}

func (a *App) Logf(format string, args ...any) { // convenience
	if a.Logger != nil {
		a.Logger.Printf(format, args...)
	}
}

type Request struct {
	series   string
	address  string
	generate bool
	remove   bool
	app      *App
}

func (r *Request) String() string {
	bytes, _ := json.MarshalIndent(r, "", "  ")
	return string(bytes)
}

func (a *App) parseRequest(r *http.Request) (Request, error) {
	path := strings.TrimPrefix(r.URL.Path, "/dalle/")

	segments := strings.SplitN(path, "/", 2)
	if len(segments) < 2 {
		return Request{}, fmt.Errorf("invalid request: %s", r.URL.Path)
	}

	series := strings.ToLower(segments[0])
	if len(series) == 0 {
		return Request{}, fmt.Errorf("series is required")
	}
	if !dalle.IsValidSeries(series, a.ValidSeries) {
		return Request{}, fmt.Errorf("invalid series")
	}

	address := strings.ToLower(segments[1])
	if len(address) == 0 {
		return Request{}, fmt.Errorf("address is required")
	}
	if !base.IsValidAddress(address) {
		return Request{}, fmt.Errorf("invalid address")
	}
	// Normalize to checksum format for canonical storage / filenames (EIP-55 via go-ethereum)
	addr := base.HexToAddress(address)
	address = addr.Hex()

	generate := r.URL.Query().Has("generate")
	remove := r.URL.Query().Has("remove")

	return Request{
		series:   series,
		address:  address,
		generate: generate,
		remove:   remove,
		app:      a,
	}, nil
}
