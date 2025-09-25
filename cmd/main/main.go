package main

import (
	"log"
	"net/http"

	"premodernonsdagar/internal/aggregation"
	"premodernonsdagar/internal/handlers"
)

func main() {
	if err := aggregation.AggregateStats(); err != nil {
		log.Fatalf("Error aggregating player stats: %v", err)
	}
	log.Println("Stats aggregated successfully.")

	// Start the web server
	mux := handlers.SetupRoutes()

	// Start the server
	serverAddr := ":8080"
	log.Println("Server started at http://localhost:8080")
	err := http.ListenAndServe(serverAddr, mux)
	if err != nil {
		log.Fatal(err)
	}
}
