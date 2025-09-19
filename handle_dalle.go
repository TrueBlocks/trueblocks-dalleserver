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
	"github.com/TrueBlocks/trueblocks-dalle/v2/pkg/progress"
	"github.com/TrueBlocks/trueblocks-dalle/v2/pkg/storage"
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
	filePath := filepath.Join(storage.OutputDir(), req.series, "annotated", req.address+".png")
	exists := file.FileExists(filePath)
	if exists && req.remove {
		_ = os.Remove(filePath)
		fmt.Fprintln(w, "image removed", filePath)
		return
	}

	if req.generate {
		dalle.Clean(req.series, req.address)
	} else if exists {
		if rw, ok := w.(http.ResponseWriter); ok {
			filePath := filepath.Join(storage.OutputDir(), req.series, "annotated", req.address+".png")
			http.ServeFile(rw, r, filePath)
			return
		}
	}

	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Set("Content-Type", "application/json")
	}

	pr := progress.GetProgress(req.series, req.address)
	if !isDebugging {
		if pr != nil && !pr.Done && !req.generate {
			logger.Info("generation already active; not spawning duplicate goroutine")
		} else {
			logger.Info("starting generation goroutine (if lock acquired)")
			go func(series, addr string) {
				start := time.Now()
				if path, err := generateAnnotatedImage(series, addr, req.app.Config.SkipImage || os.Getenv("TB_DALLE_SKIP_IMAGE") == "1", req.app.Config.LockTTL); err != nil {
					logger.InfoR("error generating image:", err)
				} else {
					if file.FileExists(path) {
						logger.InfoG(fmt.Sprintf("generated image for %s/%s in %s", series, addr, time.Since(start)))
					} else {
						logger.Info(fmt.Sprintf("generation in progress (lock contention) for %s/%s elapsed %s", series, addr, time.Since(start)))
					}
				}
			}(req.series, req.address)
		}
	} else {
		if _, err := generateAnnotatedImage(req.series, req.address, req.app.Config.SkipImage || os.Getenv("TB_DALLE_SKIP_IMAGE") == "1", req.app.Config.LockTTL); err != nil {
			logger.InfoR("error generating image:", err)
		}
	}

	if pr == nil {
		fmt.Fprintln(w, "{}")
		return
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(pr)
}
