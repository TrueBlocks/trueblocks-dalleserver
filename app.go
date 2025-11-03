package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
	dalle "github.com/TrueBlocks/trueblocks-dalle/v6"
	"github.com/TrueBlocks/trueblocks-dalle/v6/pkg/storage"
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
	series    string
	address   string
	generate  bool
	remove    bool
	app       *App
	requestID string
}

func (r *Request) String() string {
	return fmt.Sprintf(`{
  "series": "%s",
  "address": "%s",
  "generate": %t,
  "remove": %t,
  "request_id": "%s"
}`, r.series, r.address, r.generate, r.remove, r.requestID)
}

func (a *App) parseRequest(r *http.Request) (Request, *APIError) {
	requestID := GenerateRequestID()
	path := strings.TrimPrefix(r.URL.Path, "/dalle/")

	segments := strings.SplitN(path, "/", 2)
	if len(segments) < 2 {
		return Request{}, NewAPIError(
			ErrorInvalidRequest,
			"Invalid request path",
			fmt.Sprintf("Path '%s' does not match expected format /dalle/{series}/{address}", r.URL.Path),
		).WithRequestID(requestID)
	}

	series := strings.ToLower(segments[0])
	if len(series) == 0 {
		return Request{}, ErrorMissingRequiredParameter("series").WithRequestID(requestID)
	}
	if !dalle.IsValidSeries(series, a.ValidSeries) {
		return Request{}, ErrorInvalidSeriesName(series).WithRequestID(requestID)
	}

	address := strings.ToLower(segments[1])
	if len(address) == 0 {
		return Request{}, ErrorMissingRequiredParameter("address").WithRequestID(requestID)
	}
	if !base.IsValidAddress(address) {
		return Request{}, ErrorInvalidAddressFormat(address).WithRequestID(requestID)
	}

	return Request{
		series:    series,
		address:   address,
		generate:  r.URL.Query().Get("generate") == "1",
		remove:    r.URL.Query().Has("remove"),
		app:       a,
		requestID: requestID,
	}, nil
}
