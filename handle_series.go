package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/TrueBlocks/trueblocks-dalleserver/dd"
)

func (a *App) handleSeries(w http.ResponseWriter, r *http.Request) {
	seriesList := dd.SeriesList()
	bytes, _ := json.MarshalIndent(seriesList, "", "  ")
	fmt.Fprintln(w, "Available series: ", string(bytes))
}
