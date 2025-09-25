package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"premodernonsdagar/internal/templates"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	templateData := map[string]interface{}{
		"ActivePage": "index",
	}
	templates.RenderTemplate(w, "index.tmpl", templateData)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	templateData := map[string]interface{}{
		"ActivePage": "404",
	}
	templates.RenderTemplate(w, "404.tmpl", templateData)
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	templateData := map[string]interface{}{
		"ActivePage":       "about",
		"maintainer_email": "test@example.com",
	}
	templates.RenderTemplate(w, "about.tmpl", templateData)
}

func EventsHandler(w http.ResponseWriter, r *http.Request) {
	fileContent, err := os.ReadFile("files/lists/events.json")
	if err != nil {
		http.Error(w, "Error reading events file", http.StatusInternalServerError)
		return
	}

	var eventsData []map[string]interface{}
	err = json.Unmarshal(fileContent, &eventsData)
	if err != nil {
		http.Error(w, "Error loading events data", http.StatusInternalServerError)
		return
	}
	templateData := map[string]interface{}{
		"ActivePage": "events",
		"Events":     eventsData,
	}

	templates.RenderTemplate(w, "events.tmpl", templateData)
}

func EventDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Extract event ID from URL path
	eventID := r.URL.Path[len("/events/"):]

	players := []struct {
		Name   string
		Result string
	}{
		{Name: "Player 1", Result: "2-0"},
		{Name: "Player 2", Result: "1-1"},
		{Name: "Player 3", Result: "0.1"},
	}

	matches := []struct {
		Player1 string
		Player2 string
		Score   string
	}{
		{Player1: "Player 1", Player2: "Player 2", Score: "2-0"},
		{Player1: "Player 1", Player2: "Player 3", Score: "1-1"},
		{Player1: "Player 2", Player2: "Player 3", Score: "0-2"},
	}

	templateData := map[string]interface{}{
		"ActivePage": "events",
		"Date":       eventID,
		"Players":    players,
		"Matches":    matches,
	}
	templates.RenderTemplate(w, "event.tmpl", templateData)
}

func PlayersHandler(w http.ResponseWriter, r *http.Request) {
	fileContent, err := os.ReadFile("files/lists/players.json")
	if err != nil {
		http.Error(w, "Error reading players file", http.StatusInternalServerError)
		return
	}

	var playersData []map[string]interface{}
	err = json.Unmarshal(fileContent, &playersData)
	if err != nil {
		http.Error(w, "Error loading players data", http.StatusInternalServerError)
		return
	}
	templateData := map[string]interface{}{
		"ActivePage": "players",
		"Players":    playersData,
	}
	templates.RenderTemplate(w, "players.tmpl", templateData)
}

func PlayerDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Extract player ID from URL path
	playerID := r.URL.Path[len("/players/"):]

	// Construct the file path for this player
	playerFilePath := "files/players/" + playerID + ".json"

	// Read the player details file
	fileContent, err := os.ReadFile(playerFilePath)
	if err != nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	// Parse the player data
	var playerData map[string]interface{}
	err = json.Unmarshal(fileContent, &playerData)
	if err != nil {
		http.Error(w, "Error parsing player data", http.StatusInternalServerError)
		return
	}

	templateData := map[string]interface{}{
		"ActivePage": "players",
		"Player":     playerData,
	}
	templates.RenderTemplate(w, "player.tmpl", templateData)
}
