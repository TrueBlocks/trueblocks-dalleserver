package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

func (a *App) handleSeries(w http.ResponseWriter, r *http.Request) {
	logger.Info(fmt.Sprintf("Received request: %s %s", r.Method, r.URL.Path))
	seriesList := dalle.ListSeries()
	bytes, _ := json.MarshalIndent(seriesList, "", "  ")
	fmt.Fprintln(w, "Available series: ", string(bytes))
}
