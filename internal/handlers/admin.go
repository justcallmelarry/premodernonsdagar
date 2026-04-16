// Package handlers provides admin-specific HTTP handlers for event management.
package handlers

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"premodernonsdagar/internal/aggregation"
	"premodernonsdagar/internal/templates"
)

// getAvailablePlayerNames reads and extracts player names from the players.json file
func getAvailablePlayerNames() ([]string, error) {
	fileContent, err := os.ReadFile("files/lists/players.json")
	if err != nil {
		// If players file doesn't exist, continue with empty list
		log.Printf("Warning: Could not read players file: %v", err)
		return []string{}, nil
	}

	var playersData []map[string]interface{}
	err = json.Unmarshal(fileContent, &playersData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling players data: %w", err)
	}

	// Extract player names
	var playerNames []string
	for _, player := range playersData {
		if name, ok := player["name"].(string); ok {
			playerNames = append(playerNames, name)
		}
	}

	return playerNames, nil
}

// parseMatchResults extracts match data from form submission
func parseMatchResults(r *http.Request) ([]aggregation.Match, error) {
	var matches []aggregation.Match
	for i := 0; ; i++ {
		player1Key := fmt.Sprintf("matches[%d][player_1]", i)
		player2Key := fmt.Sprintf("matches[%d][player_2]", i)
		resultKey := fmt.Sprintf("matches[%d][result]", i)
		extraKey := fmt.Sprintf("matches[%d][extra_match]", i)

		player1 := r.FormValue(player1Key)
		player2 := r.FormValue(player2Key)
		result := r.FormValue(resultKey)

		if player1 == "" || player2 == "" || result == "" {
			break
		}

		// Validate player names are not empty and different
		if strings.TrimSpace(player1) == "" || strings.TrimSpace(player2) == "" {
			return nil, fmt.Errorf("player names cannot be empty")
		}
		if player1 == player2 {
			return nil, fmt.Errorf("a player cannot play against themselves")
		}

		// Parse extra match players (comma-separated)
		var extraMatch []string
		if extraPlayers := r.FormValue(extraKey); extraPlayers != "" {
			for _, player := range strings.Split(extraPlayers, ",") {
				if trimmed := strings.TrimSpace(player); trimmed != "" {
					extraMatch = append(extraMatch, trimmed)
				}
			}
		}

		matches = append(matches, aggregation.Match{
			Player1:    player1,
			Player2:    player2,
			Result:     result,
			ExtraMatch: extraMatch,
		})
	}

	return matches, nil
}

func AdminEventsListHandler(w http.ResponseWriter, r *http.Request) {
	// Read events from input/events directory
	inputEventsDir := "input/events"
	var eventItems []aggregation.EventListItem

	// Check if directory exists
	if _, err := os.Stat(inputEventsDir); os.IsNotExist(err) {
		// Directory doesn't exist, show empty list
		log.Printf("Warning: input/events directory does not exist")
	} else {
		// Read all JSON files in the directory
		err := filepath.WalkDir(inputEventsDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && strings.HasSuffix(d.Name(), ".json") {
				// Read and parse the event file
				fileContent, readErr := os.ReadFile(path)
				if readErr != nil {
					log.Printf("Error reading event file %s: %v", path, readErr)
					return nil
				}

				var event aggregation.InputEvent
				if parseErr := json.Unmarshal(fileContent, &event); parseErr != nil {
					log.Printf("Error parsing event file %s: %v", path, parseErr)
					return nil
				}

				// Create event list item
				// Extract year from date as season for now
				season := "2026"
				if len(event.Date) >= 4 {
					season = event.Date[:4]
				}

				eventItem := aggregation.EventListItem{
					Name:   event.Name,
					Date:   event.Date,
					Season: season,
					URL:    fmt.Sprintf("/admin/events/edit/%s", event.Date),
				}

				eventItems = append(eventItems, eventItem)
			}
			return nil
		})

		if err != nil {
			log.Printf("Error walking input/events directory: %v", err)
			http.Error(w, "Error reading events directory", http.StatusInternalServerError)
			return
		}
	}

	// Sort events by date (newest first)
	sort.Slice(eventItems, func(i, j int) bool {
		return eventItems[i].Date > eventItems[j].Date
	})

	// Calculate stats
	stats := map[string]interface{}{
		"Total Events": map[string]interface{}{
			"Value": len(eventItems),
			"Icon":  "calendar_check",
		},
	}

	templateData := map[string]interface{}{
		"ActivePage": "admin",
		"Scheme":     templates.ColorScheme(),
		"Stats":      stats,
		"Events":     eventItems,
		"IsAdmin":    true,
	}

	templates.RenderTemplate(w, "admin_events.tmpl", templateData)
}

func EventEntryHandler(w http.ResponseWriter, r *http.Request) {
	playerNames, err := getAvailablePlayerNames()
	if err != nil {
		http.Error(w, "Error loading players data", http.StatusInternalServerError)
		return
	}

	templateData := map[string]interface{}{
		"ActivePage":  "events",
		"Scheme":      templates.ColorScheme(),
		"PlayerNames": playerNames,
	}
	templates.RenderTemplate(w, "admin_event.tmpl", templateData)
}

func EventEntryPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	eventName := r.FormValue("event_name")
	eventDate := r.FormValue("event_date")
	if eventName == "" || eventDate == "" {
		http.Error(w, "Event name and date are required", http.StatusBadRequest)
		return
	}

	// Parse match results from form
	matches, err := parseMatchResults(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Extract unique players from matches
	playerSet := make(map[string]bool)
	for _, match := range matches {
		playerSet[match.Player1] = true
		playerSet[match.Player2] = true
	}

	// Create event structure using InputEvent
	event := aggregation.InputEvent{
		Name:       eventName,
		Date:       eventDate,
		Rounds:     4,
		PlayerInfo: make(map[string]aggregation.PlayerEventInfo),
		Matches:    matches,
	}

	// Initialize player info for all players
	for player := range playerSet {
		deckName := r.FormValue(fmt.Sprintf("player_deck_%s", player))
		event.PlayerInfo[player] = aggregation.PlayerEventInfo{
			Deck:     deckName,
			Decklist: "",
		}
	}

	// Save the event to a JSON file
	eventJSON, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		http.Error(w, "Error creating event data", http.StatusInternalServerError)
		return
	}

	// Create events directory if it doesn't exist
	if _, err := os.Stat("input/events"); os.IsNotExist(err) {
		err = os.MkdirAll("input/events", 0755)
		if err != nil {
			http.Error(w, "Error creating events directory", http.StatusInternalServerError)
			return
		}
	}

	filePath := fmt.Sprintf("input/events/%s.json", eventDate)
	err = os.WriteFile(filePath, eventJSON, 0644)
	if err != nil {
		http.Error(w, "Error saving event", http.StatusInternalServerError)
		return
	}

	// Redirect to admin events page
	http.Redirect(w, r, "/admin/events", http.StatusSeeOther)
}

func EventEditHandler(w http.ResponseWriter, r *http.Request) {
	// Extract event date from URL path
	eventDate := r.URL.Path[len("/admin/events/edit/"):]

	// Read the existing event file
	filePath := fmt.Sprintf("input/events/%s.json", eventDate)
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading event file %s: %v", filePath, err)
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	var existingEvent aggregation.InputEvent
	err = json.Unmarshal(fileContent, &existingEvent)
	if err != nil {
		log.Printf("Error unmarshalling event data: %v", err)
		http.Error(w, "Error parsing event data", http.StatusInternalServerError)
		return
	}

	playerNames, err := getAvailablePlayerNames()
	if err != nil {
		http.Error(w, "Error loading players data", http.StatusInternalServerError)
		return
	}

	templateData := map[string]interface{}{
		"ActivePage":    "events",
		"Scheme":        templates.ColorScheme(),
		"PlayerNames":   playerNames,
		"ExistingEvent": existingEvent,
		"IsEdit":        true,
	}
	templates.RenderTemplate(w, "admin_event.tmpl", templateData)
}

func EventEditPostHandler(w http.ResponseWriter, r *http.Request) {
	// Extract event date from URL path
	eventDate := r.URL.Path[len("/admin/events/edit/"):]

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	eventName := r.FormValue("event_name")
	eventDateForm := r.FormValue("event_date")
	if eventName == "" || eventDateForm == "" {
		http.Error(w, "Event name and date are required", http.StatusBadRequest)
		return
	}

	// Parse match results from form
	matches, err := parseMatchResults(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Extract unique players from matches
	playerSet := make(map[string]bool)
	for _, match := range matches {
		playerSet[match.Player1] = true
		playerSet[match.Player2] = true
	}

	// Create event structure using InputEvent
	event := aggregation.InputEvent{
		Name:       eventName,
		Date:       eventDateForm,
		Rounds:     4,
		PlayerInfo: make(map[string]aggregation.PlayerEventInfo),
		Matches:    matches,
	}

	// Initialize player info for all players
	for player := range playerSet {
		deckName := r.FormValue(fmt.Sprintf("player_deck_%s", player))
		event.PlayerInfo[player] = aggregation.PlayerEventInfo{
			Deck:     deckName,
			Decklist: "",
		}
	}

	// Save the event to a JSON file
	eventJSON, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		http.Error(w, "Error creating event data", http.StatusInternalServerError)
		return
	}

	// Create events directory if it doesn't exist
	if _, err := os.Stat("input/events"); os.IsNotExist(err) {
		err = os.MkdirAll("input/events", 0755)
		if err != nil {
			http.Error(w, "Error creating events directory", http.StatusInternalServerError)
			return
		}
	}

	// Use the form date for the file path (in case date was changed)
	newFilePath := fmt.Sprintf("input/events/%s.json", eventDateForm)

	// If the date changed, remove the old file
	oldFilePath := fmt.Sprintf("input/events/%s.json", eventDate)
	if oldFilePath != newFilePath {
		os.Remove(oldFilePath)
	}

	err = os.WriteFile(newFilePath, eventJSON, 0644)
	if err != nil {
		http.Error(w, "Error saving event", http.StatusInternalServerError)
		return
	}

	// Redirect to admin events page
	http.Redirect(w, r, "/admin/events", http.StatusSeeOther)
}