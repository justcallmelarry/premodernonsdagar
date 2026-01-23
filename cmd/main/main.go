package main

import (
	"log"
	"net/http"
	"os"
	"slices"

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

	buildFlag := false
	if len(os.Args) > 1 {
		if slices.Contains(os.Args[1:], "--build") {
			buildFlag = true
		}
	}
	if config.DevelopmentEnvironment || buildFlag {
		err := templates.RenderAllTemplates()
		if err != nil {
			log.Fatalf("Error rendering templates: %v", err)
		}
		log.Println("All templates rendered successfully.")
	}

	if buildFlag {
		return
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
