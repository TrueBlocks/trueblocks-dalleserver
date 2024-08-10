package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(t *testing.T) {
	cwd, _ := os.Getwd()
	fmt.Println(cwd)
	isDebugging = true
	app := NewApp()
	app.StartLogging()
	defer app.StopLogging()
	req := Request{
		series:   "five-tone-postal-protozoa",
		address:  "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
		filePath: "testing",
		generate: true,
		app:      app,
	}
	req.filePath = filepath.Join("./output", req.series, req.address+".png")
	defer os.Remove(filepath.Join("./pending", req.series+"-"+req.address+".lck"))
	req.Respond(os.Stdout, nil)
}
