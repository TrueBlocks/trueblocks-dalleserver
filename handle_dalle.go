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
	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

var isDebugging = false

// indirection for easier test injection of failures
var generateAnnotatedImage = dalle.GenerateAnnotatedImage

func (a *App) handleDalleDress(w http.ResponseWriter, r *http.Request) {
	a.Logger.Printf("Received request: %s %s", r.Method, r.URL.Path)
	req, err := a.parseRequest(r)
	if err != nil {
		a.Logger.Printf("Err: %s", err.Error())
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

	req.app.Logger.Println("exists:", exists)
	req.app.Logger.Println("generate:", req.generate)
	if exists && !req.generate {
		if rw, ok := w.(http.ResponseWriter); ok {
			req.app.Logger.Println("the image exists, serving it")
			filePath := filepath.Join(dalle.OutputDir(), series, "annotated", addr+".png")
			http.ServeFile(rw, r, filePath)
			return
		}
	}

	if !isDebugging {
		req.app.Logger.Println("starting generation goroutine (if lock acquired)")
		go func(series, addr string) {
			start := time.Now()
			if _, err := generateAnnotatedImage(series, addr, req.app.Config.SkipImage || os.Getenv("TB_DALLE_SKIP_IMAGE") == "1", req.app.Config.LockTTL); err != nil {
				req.app.Logger.Println("error generating image:", err)
			} else {
				req.app.Logger.Printf("generated image for %s/%s in %s", series, addr, time.Since(start))
			}
		}(series, addr)
	} else {
		if _, err := generateAnnotatedImage(series, addr, req.app.Config.SkipImage || os.Getenv("TB_DALLE_SKIP_IMAGE") == "1", req.app.Config.LockTTL); err != nil {
			req.app.Logger.Println("error generating image:", err)
		}
	}

	// Always attempt to fetch progress (may be nil if run not started yet)
	pr := dalle.GetProgress(series, addr)
	if pr == nil {
		fmt.Fprintln(w, "{}")
		return
	}
	// Force headers if possible
	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Set("Content-Type", "application/json")
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(pr)

}
