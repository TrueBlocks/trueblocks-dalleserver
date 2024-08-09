package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	app := NewApp()
	app.StartLogging()
	defer app.StopLogging()

	fmt.Println("Server started at :8080")

	http.HandleFunc("/dalle/", app.handleDalleDress)
	http.HandleFunc("/series", app.handleSeries)
	http.HandleFunc("/series/", app.handleSeries)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
