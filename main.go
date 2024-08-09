package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	app := NewApp()
	http.HandleFunc("/dalle/", app.handleDalleDress)
	http.HandleFunc("/series", app.handleSeries)
	http.HandleFunc("/series/", app.handleSeries)
	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
