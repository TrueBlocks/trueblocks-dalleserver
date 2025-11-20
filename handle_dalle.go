package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/file"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/logger"
	dalle "github.com/TrueBlocks/trueblocks-dalle/v6"
	"github.com/TrueBlocks/trueblocks-dalle/v6/pkg/progress"
	"github.com/TrueBlocks/trueblocks-dalle/v6/pkg/prompt"
	"github.com/TrueBlocks/trueblocks-dalle/v6/pkg/storage"
)

var isDebugging = false

// indirection for easier test injection of failures
var generateAnnotatedImage = dalle.GenerateAnnotatedImage

func (a *App) handleDalleDress(w http.ResponseWriter, r *http.Request) {
	logger.Info(fmt.Sprintf("Received request: %s %s", r.Method, r.URL.Path))
	req, apiErr := a.parseRequest(r)
	if apiErr != nil {
		WriteErrorResponse(w, apiErr, http.StatusBadRequest)
		return
	}
	req.Respond(w, r)
}

func (req *Request) Respond(w io.Writer, r *http.Request) {
	filePath := filepath.Join(storage.OutputDir(), req.series, "annotated", req.address+".png")
	exists := file.FileExists(filePath)
	if exists && req.remove {
		_ = os.Remove(filePath)
		if _, err := fmt.Fprintln(w, "image removed", filePath); err != nil {
			// Log error or handle as appropriate for your application
			_ = err
		}
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
			logger.Info(fmt.Sprintf("[%s] generation already active; not spawning duplicate goroutine", req.requestID))
		} else {
			logger.Info(fmt.Sprintf("[%s] starting generation goroutine (if lock acquired)", req.requestID))
			go func(series, addr, requestID string) {
				start := time.Now()

				if path, err := generateAnnotatedImage(series, addr, req.app.Config.SkipImage || os.Getenv("TB_DALLE_SKIP_IMAGE") == "1", req.app.Config.LockTTL); err != nil {
					logger.InfoR(fmt.Sprintf("[%s] error generating image:", requestID), err)
					GetMetricsCollector().RecordError("GENERATION_ERROR", "/dalle/", requestID)
				} else {
					if file.FileExists(path) {
						logger.InfoG(fmt.Sprintf("[%s] generated image for %s/%s in %s", requestID, series, addr, time.Since(start)))
					} else {
						logger.Info(fmt.Sprintf("[%s] generation in progress (lock contention) for %s/%s elapsed %s", requestID, series, addr, time.Since(start)))
					}
				}
			}(req.series, req.address, req.requestID)
		}
	} else {
		if _, err := generateAnnotatedImage(req.series, req.address, req.app.Config.SkipImage || os.Getenv("TB_DALLE_SKIP_IMAGE") == "1", req.app.Config.LockTTL); err != nil {
			logger.InfoR(fmt.Sprintf("[%s] error generating image:", req.requestID), err)
			GetMetricsCollector().RecordError("GENERATION_ERROR", "/dalle/", req.requestID)
		}
	}

	if pr == nil {
		// Return empty progress with request ID
		if rw, ok := w.(http.ResponseWriter); ok {
			WriteSuccessResponse(rw, map[string]interface{}{}, req.requestID)
		} else {
			if _, err := fmt.Fprintln(w, "{}"); err != nil {
				// Log error or handle as appropriate for your application
				_ = err
			}
		}
		return
	}

	// Add request ID to progress response
	if rw, ok := w.(http.ResponseWriter); ok {
		WriteSuccessResponse(rw, pr, req.requestID)
	} else {
		// Fallback for non-HTTP writers (tests)
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(pr)
	}
}
