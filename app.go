package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
	"go.uber.org/zap"
)

type App struct {
	ValidSeries []string
	Logger      *log.Logger
	LogFile     *os.File
	SLogger     *zap.SugaredLogger
	Config      Config
}

func NewApp() *App {
	app := App{Config: LoadConfig()}
	// OUTPUT_DIR
	app.ValidSeries = dalle.ListSeries("output")
	return &app
}

func (a *App) StartLogging() {
	var err error
	// Use restrictive permissions (0600) per gosec recommendation (was 0666)
	a.LogFile, err = os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	a.Logger = log.New(a.LogFile, "LOG: ", log.LstdFlags|log.Lshortfile)
	var zl *zap.Logger
	if a.Config.LogJSON {
		zl, err = zap.NewProduction()
	} else {
		zl, err = zap.NewDevelopment()
	}
	if err == nil {
		a.SLogger = zl.Sugar()
	}
}

func (a *App) StopLogging() {
	_ = a.LogFile.Close()
	if a.SLogger != nil {
		_ = a.SLogger.Sync()
	}
}

type Request struct {
	filePath string
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

	// OUTPUT_DIR
	filePath := filepath.Join("./output", series, "annotated", address+".png")
	return Request{
		filePath: filePath,
		series:   series,
		address:  address,
		generate: generate,
		remove:   remove,
		app:      a,
	}, nil
}
