package main

import (
	"fmt"
	"net/http"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v6"
)

func (a *App) handleSeries(w http.ResponseWriter, r *http.Request) {
	requestID := GenerateRequestID()
	logInfo(fmt.Sprintf("[%s] Received request: %s %s", requestID, r.Method, r.URL.Path))

	seriesList := dalle.ListSeries()
	data := map[string]interface{}{
		"series": seriesList,
		"count":  len(seriesList),
	}

	WriteSuccessResponse(w, data, requestID)
}
