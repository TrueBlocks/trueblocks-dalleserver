package main

import (
	"fmt"
	"net/http"
)

func (a *App) handleDefault(w http.ResponseWriter, r *http.Request) {
	a.Logger.Printf("Received request: %s %s", r.Method, r.URL.Path)
	fmt.Fprintln(w, "Request either:")
	fmt.Fprintln(w, "   http://localhost:8080/series?[details=series], or")
	fmt.Fprintln(w, "   http://localhost:8080/dalle/<series>/<address>?[generate|remove]")
}
