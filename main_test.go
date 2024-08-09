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
	req := Request{
		series:   "simple",
		address:  "0x1234567890123456789012345678901234567890",
		filePath: "testing",
		generate: true,
	}
	req.filePath = filepath.Join("./output", req.series, req.address+".png")
	defer os.Remove(filepath.Join("./pending", req.series+"-"+req.address+".lck"))
	req.Respond(os.Stdout, nil)
}
