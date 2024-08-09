package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-dalleserver/dd"
)

type App struct {
	ValidSeries []string
}

func NewApp() *App {
	app := App{}
	app.ValidSeries = dd.SeriesList()
	return &app
}

func (a *App) handleDalleDress(w http.ResponseWriter, r *http.Request) {
	filePath, series, address, generate, err := a.parseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// If the file already exists and we're not being forced to generate, serve it
	if _, err := os.Stat(filePath); err == nil && !generate {
		http.ServeFile(w, r, filePath)
		return
	}

	// If the file is currently being generated, return a waiting message
	lockFilePath := filepath.Join("./pending", series+"-"+address+".lck")
	if _, err := os.Stat(lockFilePath); err == nil {
		fmt.Fprintln(w, "Your image will be ready in", dd.Waiting, "seconds")
		return
	}

	// Otherwise, kick off a generation job and return a waiting message
	go dd.CreateDalleDress(series, address, filePath)
	fmt.Fprintln(w, "Your image will be ready in", dd.Waiting, "seconds")
}

func (a *App) parseRequest(r *http.Request) (string, string, string, bool, error) {
	path := strings.TrimPrefix(r.URL.Path, "/dalle/")

	segments := strings.SplitN(path, "/", 2)
	if len(segments) < 2 {
		return "", "", "", false, fmt.Errorf("invalid request: %s", r.URL.Path)
	}

	series := segments[0]
	if len(series) == 0 {
		return "", "", "", false, fmt.Errorf("series is required")
	}
	if !dd.IsValidSeries(series, a.ValidSeries) {
		return "", "", "", false, fmt.Errorf("invalid series")
	}

	address := segments[1]
	if len(address) == 0 {
		return "", "", "", false, fmt.Errorf("address is required")
	}
	if !base.IsValidAddress(address) {
		return "", "", "", false, fmt.Errorf("invalid address")
	}

	generate := r.URL.Query().Has("generate")

	filePath := filepath.Join("./output", series, address+".png")
	return filePath, series, address, generate, nil
}
