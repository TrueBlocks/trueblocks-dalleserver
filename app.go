package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
	"github.com/TrueBlocks/trueblocks-dalle/v2/pkg/storage"
)

type App struct {
	ValidSeries []string
	Config      Config
}

func NewApp() *App {
	app := App{Config: MustLoadConfig()}
	_ = os.MkdirAll(storage.OutputDir(), 0o750)
	app.ValidSeries = dalle.ListSeries()
	return &app
}

type Request struct {
	series   string
	address  string
	generate bool
	remove   bool
	app      *App
}

func (r *Request) String() string {
	return fmt.Sprintf(`{
  "series": "%s",
  "address": "%s",
  "generate": %t,
  "remove": %t
}`, r.series, r.address, r.generate, r.remove)
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

	return Request{
		series:   series,
		address:  address,
		generate: r.URL.Query().Get("generate") == "1",
		remove:   r.URL.Query().Has("remove"),
		app:      a,
	}, nil
}
