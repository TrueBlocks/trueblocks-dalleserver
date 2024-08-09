package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
)

func main() {
	app := NewApp()
	app.StartLogging()
	defer app.StopLogging()

	http.HandleFunc("/{$}", app.handleDefault)
	http.HandleFunc("/dalle/", app.handleDalleDress)
	http.HandleFunc("/series", app.handleSeries)
	http.HandleFunc("/series/", app.handleSeries)

	port := getPort()
	fmt.Println("Starting server on", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Println("Error starting server:", err)
		fmt.Println("Run with " + colors.BrightYellow + "--port=n" + colors.Off + ".")
	}
}

func getPort() string {
	port := ":8080"
	if len(os.Args) > 1 && strings.Contains(os.Args[1], "--port=") {
		isNumeric := func(s string) bool {
			_, err := strconv.ParseFloat(s, 64)
			return err == nil
		}
		n := strings.ReplaceAll(os.Args[1], "--port=", "")
		if !isNumeric(n) {
			logger.Fatal("Invalid port number: " + n)
		}
		port = ":" + n
	}
	return port
}
