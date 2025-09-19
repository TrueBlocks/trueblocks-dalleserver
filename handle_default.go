package main

import (
	"fmt"
	"net/http"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
)

func (a *App) handleDefault(w http.ResponseWriter, r *http.Request) {
	logger.Info(fmt.Sprintf("Received request: %s %s", r.Method, r.URL.Path))
	if _, err := fmt.Fprintln(w, "Endpoints:"); err != nil {
		logger.Error("Failed to write response:", err)
		return
	}
	if _, err := fmt.Fprintln(w, "  /series?[details=series] - list valid series"); err != nil {
		logger.Error("Failed to write response:", err)
		return
	}
	if _, err := fmt.Fprintln(w, "  /dalle/<series>/<address>?[generate|remove] - request or remove an image"); err != nil {
		logger.Error("Failed to write response:", err)
		return
	}
	if _, err := fmt.Fprintln(w, "  /preview - HTML gallery of generated annotated images"); err != nil {
		logger.Error("Failed to write response:", err)
		return
	}
	if _, err := fmt.Fprintln(w, "  /health | /metrics"); err != nil {
		logger.Error("Failed to write response:", err)
		return
	}
}
