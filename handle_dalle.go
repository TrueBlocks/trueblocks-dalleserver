package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
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
	a.Logger.Printf("Req: %s", req.String())
	if err != nil {
		a.Logger.Printf("Err: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Respond(w, r)
}

func (req *Request) Respond(w io.Writer, r *http.Request) {
	exists := file.FileExists(req.filePath)
	if req.remove {
		if !exists {
			fmt.Fprintln(w, "Image not found")
			return
		}
		_ = os.Remove(req.filePath)
		fmt.Fprintln(w, "Image removed")
		return
	}

	// If the file already exists and we're not told to generate it, serve it
	req.app.Logger.Println("exists:", exists)
	req.app.Logger.Println("generate:", req.generate)
	if exists && !req.generate {
		rw, ok := w.(http.ResponseWriter)
		if ok {
			req.app.Logger.Println("the image exists, serving it")
			http.ServeFile(rw, r, req.filePath)
			return
		}
	}

	if !isDebugging {
		// Fire off async generation respecting in-memory lock
		req.app.Logger.Println("calling the go routine to generate the image")
		go func(series, addr, fp string) {
			start := time.Now()
			// OUTPUT_DIR
			if _, err := generateAnnotatedImage(series, addr, "output", req.app.Config.SkipImage || os.Getenv("DALLESERVER_SKIP_IMAGE") == "1", req.app.Config.LockTTL); err != nil {
				req.app.Logger.Println("error generating image:", err)
			} else {
				req.app.Logger.Printf("generated image for %s/%s in %s", series, addr, time.Since(start))
			}
		}(req.series, req.address, req.filePath)
	} else {
		// OUTPUT_DIR
		if _, err := generateAnnotatedImage(req.series, req.address, "output", req.app.Config.SkipImage || os.Getenv("DALLESERVER_SKIP_IMAGE") == "1", req.app.Config.LockTTL); err != nil {
			req.app.Logger.Println("error generating image:", err)
		}
	}

	fmt.Fprintln(w, "Your image will be ready shortly")

}
