package main

import (
	"fmt"
	"net/http"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
)

func (a *App) handleDefault(w http.ResponseWriter, r *http.Request) {
	logger.Info(fmt.Sprintf("Received request: %s %s", r.Method, r.URL.Path))
	fmt.Fprintln(w, "Endpoints:")
	fmt.Fprintln(w, "  /series?[details=series] - list valid series")
	fmt.Fprintln(w, "  /dalle/<series>/<address>?[generate|remove] - request or remove an image")
	fmt.Fprintln(w, "  /preview - HTML gallery of generated annotated images")
	fmt.Fprintln(w, "  /health | /metrics")
}
