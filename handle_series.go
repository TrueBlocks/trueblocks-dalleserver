package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

func (a *App) handleSeries(w http.ResponseWriter, r *http.Request) {
	a.Logger.Printf("Received request: %s %s", r.Method, r.URL.Path)
	// OUTPUT_DIR
	seriesList := dalle.ListSeries("output")
	bytes, _ := json.MarshalIndent(seriesList, "", "  ")
	fmt.Fprintln(w, "Available series: ", string(bytes))
}
