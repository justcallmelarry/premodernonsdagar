package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"premodernonsdagar/internal/aggregation"
	"premodernonsdagar/internal/templates"
	"premodernonsdagar/internal/utils"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	nextEvent := utils.NextEvent(time.Now())
	weekNumber := utils.SwedishWeekNumber(nextEvent)

	eventString := nextEvent.Format("2006-01-02")
	if nextEvent.Format("2006-01-02") == time.Now().Format("2006-01-02") {
		eventString = "Today!"
	}

	templateData := map[string]interface{}{
		"ActivePage":          "index",
		"NextEventDate":       eventString,
		"NextEventWeekNumber": weekNumber,
		"Scheme":              templates.ColorScheme(),
	}
	templates.RenderTemplate(w, "index.tmpl", templateData)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	templateData := map[string]interface{}{
		"ActivePage": "404",
		"Scheme":     templates.ColorScheme(),
	}
	templates.RenderTemplate(w, "404.tmpl", templateData)
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	templateData := map[string]interface{}{
		"ActivePage": "about",
		"Scheme":     templates.ColorScheme(),
		"Rules":      templates.Rules(),
	}
	templates.RenderTemplate(w, "about.tmpl", templateData)
}

func EventsHandler(w http.ResponseWriter, r *http.Request) {
	fileContent, err := os.ReadFile("files/lists/events.json")
	if err != nil {
		http.Error(w, "Error reading events file", http.StatusInternalServerError)
		return
	}

	var eventsData aggregation.EventListStats
	err = json.Unmarshal(fileContent, &eventsData)
	if err != nil {
		log.Printf("Error unmarshalling events data: %v", err)
		http.Error(w, "Error loading events data", http.StatusInternalServerError)
		return
	}

	stats := map[string]interface{}{
		"Total Events": map[string]interface{}{
			"Value": eventsData.Count,
			"Icon":  "calendar_check",
		},
		"Average Turnout": map[string]interface{}{
			"Value": fmt.Sprintf("%.2f", float64(eventsData.AverageAttendance)),
			"Icon":  "person",
		},
		"Max Turnout": map[string]interface{}{
			"Value": eventsData.MaxAttendance,
			"Icon":  "person",
		},
		"Min Turnout": map[string]interface{}{
			"Value": eventsData.MinAttendance,
			"Icon":  "person",
		},
	}
	templateData := map[string]interface{}{
		"ActivePage": "events",
		"Scheme":     templates.ColorScheme(),
		"Stats":      stats,
		"Events":     eventsData.Events,
	}

	templates.RenderTemplate(w, "events.tmpl", templateData)
}

func EventDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Extract event ID from URL path
	eventID := r.URL.Path[len("/events/"):]

	filePath := "files/events/" + eventID + ".json"
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading event file %s: %v", filePath, err)
		http.Error(w, "Error reading events file", http.StatusInternalServerError)
		return
	}

	var eventsData aggregation.Event
	err = json.Unmarshal(fileContent, &eventsData)
	if err != nil {
		log.Printf("Error unmarshalling events data: %v", err)
		http.Error(w, "Error loading events data", http.StatusInternalServerError)
		return
	}

	templateData := map[string]interface{}{
		"ActivePage": "events",
		"Scheme":     templates.ColorScheme(),
		"Event":      eventsData,
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
		"Scheme":     templates.ColorScheme(),
		"Players":    playersData,
	}
	templates.RenderTemplate(w, "players.tmpl", templateData)
}

func PlayerDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Extract player ID from URL path
	playerID := r.URL.Path[len("/players/"):]

	filePath := "files/players/" + playerID + ".json"
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	var playerData aggregation.Player
	err = json.Unmarshal(fileContent, &playerData)
	if err != nil {
		http.Error(w, "Error parsing player data", http.StatusInternalServerError)
		return
	}

	templateData := map[string]interface{}{
		"ActivePage": "players",
		"Scheme":     templates.ColorScheme(),
		"Player":     playerData,
	}
	templates.RenderTemplate(w, "player.tmpl", templateData)
}

func LeaderboardsHandler(w http.ResponseWriter, r *http.Request) {
	fileContent, err := os.ReadFile("files/lists/leaderboards.json")
	if err != nil {
		http.Error(w, "Error reading leaderboards file", http.StatusInternalServerError)
		return
	}

	var leaderboardsData []aggregation.LeaderboardContainer
	err = json.Unmarshal(fileContent, &leaderboardsData)
	if err != nil {
		http.Error(w, "Error loading leaderboards data", http.StatusInternalServerError)
		return
	}

	showCount := 8
	if countParam := r.URL.Query().Get("show"); countParam != "" {
		if count, err := strconv.Atoi(countParam); err == nil && count > 0 {
			showCount = count
		}
	}

	templateData := map[string]interface{}{
		"ActivePage":   "leaderboards",
		"Scheme":       templates.ColorScheme(),
		"Leaderboards": leaderboardsData,
		"ShowCount":    showCount,
	}
	templates.RenderTemplate(w, "leaderboards.tmpl", templateData)
}

func DecklistHandler(w http.ResponseWriter, r *http.Request) {
	// Extract decklist ID from URL path
	decklistID := r.URL.Path[len("/decklists/"):]

	filePath := "files/decklists/" + decklistID + ".json"
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	var decklistData aggregation.Decklist
	err = json.Unmarshal(fileContent, &decklistData)
	if err != nil {
		http.Error(w, "Error parsing decklist data", http.StatusInternalServerError)
		return
	}

	templateData := map[string]interface{}{
		"ActivePage": "",
		"Scheme":     templates.ColorScheme(),
		"Decklist":   decklistData,
	}

	templates.RenderTemplate(w, "decklist.tmpl", templateData)
}

func ImagesHandler(w http.ResponseWriter, r *http.Request) {
	scryfallURL := r.URL.Query().Get("url")
	if scryfallURL == "" {
		http.Error(w, "Missing 'url' query parameter", http.StatusBadRequest)
		return
	}
	parsedURL, err := url.Parse(scryfallURL)
	if err != nil {
		http.Error(w, "Invalid Scryfall URL", http.StatusBadRequest)
		return
	}

	parsedURL.RawQuery = ""
	parsedURL.Fragment = ""
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) == 0 {
		http.Error(w, "Invalid Scryfall URL", http.StatusBadRequest)
		return
	}
	filename := pathParts[len(pathParts)-1]

	// Ensure the directory exists
	if _, err := os.Stat("static/images/scryfall"); os.IsNotExist(err) {
		err = os.MkdirAll("static/images/scryfall", 0755)
		if err != nil {
			http.Error(w, "Failed to create image directory", http.StatusInternalServerError)
			return
		}
	}

	imagePath := "static/images/scryfall/" + filename

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		resp, err := http.Get(scryfallURL)
		if err != nil {
			http.Error(w, "Failed to fetch image from Scryfall", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Save the image to the local file system
		out, err := os.Create(imagePath)
		if err != nil {
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/"+imagePath, http.StatusFound)
}
