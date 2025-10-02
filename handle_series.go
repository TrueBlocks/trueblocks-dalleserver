package main

import (
	"fmt"
	"net/http"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

func (a *App) handleSeries(w http.ResponseWriter, r *http.Request) {
	requestID := GenerateRequestID()
	logger.Info(fmt.Sprintf("[%s] Received request: %s %s", requestID, r.Method, r.URL.Path))

	seriesList := dalle.ListSeries()
	logger.Info(fmt.Sprintf("[%s] ListSeries returned %d series: %v", requestID, len(seriesList), seriesList))

	data := map[string]interface{}{
		"series": seriesList,
		"count":  len(seriesList),
	}

	logger.Info(fmt.Sprintf("[%s] Sending response with data: %+v", requestID, data))
	WriteSuccessResponse(w, data, requestID)
}
