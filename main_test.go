package main

import (
	"fmt"
	"os"
	"testing"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

func TestMainRequestRespond(t *testing.T) {
	cwd, _ := os.Getwd()
	fmt.Println(cwd)
	isDebugging = true
	dalle.SetupTest(t, dalle.SetupTestOptions{Series: []string{"simple"}})
	app := NewApp()
	app.StartLogging()
	defer app.StopLogging()
	series := "simple"
	addr := "0xf503017d7baf7fbc0fff7492b751025c6a78179b"
	req := Request{
		series:   series,
		address:  addr,
		generate: true,
		app:      app,
	}
	req.Respond(os.Stdout, nil)
}
