package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/file"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

var isDebugging = false

// indirection for easier test injection of failures
var generateAnnotatedImage = dalle.GenerateAnnotatedImage

func (a *App) handleDalleDress(w http.ResponseWriter, r *http.Request) {
	logger.Info(fmt.Sprintf("Received request: %s %s", r.Method, r.URL.Path))
	req, err := a.parseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Respond(w, r)
}

func (req *Request) Respond(w io.Writer, r *http.Request) {
	series := req.series
	addr := req.address
	filePath := filepath.Join(dalle.OutputDir(), series, "annotated", addr+".png")
	exists := file.FileExists(filePath)
	if req.remove {
		if !exists {
			fmt.Fprintln(w, "Image not found")
			return
		}
		_ = os.Remove(filePath)
		fmt.Fprintln(w, "Image removed")
		return
	}

	if exists && !req.generate {
		if rw, ok := w.(http.ResponseWriter); ok {
			filePath := filepath.Join(dalle.OutputDir(), series, "annotated", addr+".png")
			http.ServeFile(rw, r, filePath)
			return
		}
	}

	if !isDebugging {
		go func(series, addr string) {
			start := time.Now()
			if _, err := generateAnnotatedImage(series, addr, req.app.Config.SkipImage || os.Getenv("TB_DALLE_SKIP_IMAGE") == "1", req.app.Config.LockTTL); err != nil {
				logger.InfoR("error generating image:", err)
			} else {
				logger.InfoG(fmt.Sprintf("generated image for %s/%s in %s", series, addr, time.Since(start)))
			}
		}(series, addr)
	} else {
		if _, err := generateAnnotatedImage(series, addr, req.app.Config.SkipImage || os.Getenv("TB_DALLE_SKIP_IMAGE") == "1", req.app.Config.LockTTL); err != nil {
			logger.InfoR("error generating image:", err)
		}
	}

	pr := dalle.GetProgress(series, addr)
	if pr == nil {
		fmt.Fprintln(w, "{}")
		return
	}

	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Set("Content-Type", "application/json")
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(pr)
}
