package main

import (
	"log"
	"net/http"

	"premodernonsdagar/internal/aggregation"
	"premodernonsdagar/internal/config"
	"premodernonsdagar/internal/handlers"
	"premodernonsdagar/internal/templates"
)

func main() {
	if err := aggregation.AggregateStats(); err != nil {
		log.Fatalf("Error aggregating player stats: %v", err)
	}
	log.Println("Stats aggregated successfully.")

	config := config.GetConfig()
	if config.DevelopmentEnvironment {
		err := templates.RenderAllTemplates()
		if err != nil {
			log.Fatalf("Error rendering templates: %v", err)
		}
		log.Println("All templates rendered successfully.")
	}

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
