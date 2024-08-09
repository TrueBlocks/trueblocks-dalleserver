package main

import (
	"fmt"
    "encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-dalleserver/pkg/dd"
)

type App struct {
	ValidSeries []string
	Logger      *log.Logger
	LogFile     *os.File
}

func NewApp() *App {
	app := App{}
	app.ValidSeries = dd.SeriesList()
	return &app
}

func (a *App) StartLogging() {
	var err error
	a.LogFile, err = os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	a.Logger = log.New(a.LogFile, "LOG: ", log.LstdFlags|log.Lshortfile)
}

func (a *App) StopLogging() {
	a.LogFile.Close()
}

type Request struct {
	filePath string
	series   string
	address  string
	generate bool
	remove   bool
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

	series := segments[0]
	if len(series) == 0 {
		return Request{}, fmt.Errorf("series is required")
	}
	if !dd.IsValidSeries(series, a.ValidSeries) {
		return Request{}, fmt.Errorf("invalid series")
	}

	address := segments[1]
	if len(address) == 0 {
		return Request{}, fmt.Errorf("address is required")
	}
	if !base.IsValidAddress(address) {
		return Request{}, fmt.Errorf("invalid address")
	}

	generate := r.URL.Query().Has("generate")
	remove := r.URL.Query().Has("remove")

	filePath := filepath.Join("./output", series, "annotated", address+".png")
	return Request{
		filePath: filePath,
		series:   series,
		address:  address,
		generate: generate,
		remove:   remove,
	}, nil
}
