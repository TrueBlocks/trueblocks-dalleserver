package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/file"
	"github.com/TrueBlocks/trueblocks-dalleserver/pkg/dd"
)

var isDebugging = false

func (a *App) handleDalleDress(w http.ResponseWriter, r *http.Request) {
	a.Logger.Printf("Received request: %s %s", r.Method, r.URL.Path)
	req, err := a.parseRequest(r)
	if err != nil {
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
		os.Remove(req.filePath)
		fmt.Fprintln(w, "Image removed")
		return
	}

	// If the file already exists and we're not told to generate it, serve it
	if exists && !req.generate {
		rw, ok := w.(http.ResponseWriter)
		if ok {
			http.ServeFile(rw, r, req.filePath)
			return
		}
	}

	if !isDebugging {
		// If the file is currently being generated, return a waiting message
		lockFilePath := filepath.Join("./pending", req.series+"-"+req.address+".lck")
		if _, err := os.Stat(lockFilePath); err == nil {
			fmt.Fprintln(w, "Your image will be ready shortly")
			return
		}
	}

	// Otherwise, kick off a generation job and return a waiting message
	if isDebugging {
		dd.CreateDalleDress(req.series, req.address, req.filePath)
	} else {
		go dd.CreateDalleDress(req.series, req.address, req.filePath)
	}
	fmt.Fprintln(w, "Your image will be ready shortly")
}
