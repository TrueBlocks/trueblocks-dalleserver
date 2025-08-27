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
	dalle.SetupTest(t, dalle.SetupTestOptions{Series: []string{"empty"}})
	app := NewApp()
	series := "empty"
	addr := "0xf503017d7baf7fbc0fff7492b751025c6a78179b"
	req := Request{
		series:   series,
		address:  addr,
		generate: true,
		app:      app,
	}
	req.Respond(os.Stdout, nil)
}
