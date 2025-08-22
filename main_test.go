package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMainRequestRespond(t *testing.T) {
	_ = os.Setenv("DALLESERVER_SKIP_IMAGE", "1")
	defer os.Unsetenv("DALLESERVER_SKIP_IMAGE")
	cwd, _ := os.Getwd()
	fmt.Println(cwd)
	isDebugging = true
	_ = withTempDataDir(t, map[string]string{"simple": `{"suffix":"simple"}`})
	app := NewApp()
	app.StartLogging()
	defer app.StopLogging()
	series := "simple"
	addr := "0xf503017d7baf7fbc0fff7492b751025c6a78179b"
	req := Request{
		series:   series,
		address:  addr,
		filePath: "testing",
		generate: true,
		app:      app,
	}
	req.filePath = filepath.Join(app.OutputDir(), req.series, req.address+".png")
	req.Respond(os.Stdout, nil)
}
