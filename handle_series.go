package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/TrueBlocks/trueblocks-dalleserver/pkg/dd"
)

func (a *App) handleSeries(w http.ResponseWriter, r *http.Request) {
	a.Logger.Printf("Received request: %s %s", r.Method, r.URL.Path)
	seriesList := dd.SeriesList()
	bytes, _ := json.MarshalIndent(seriesList, "", "  ")
	fmt.Fprintln(w, "Available series: ", string(bytes))
}
