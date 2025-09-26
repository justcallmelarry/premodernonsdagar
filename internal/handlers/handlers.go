package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"

	"premodernonsdagar/internal/aggregation"
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

	var eventsData aggregation.EventData
	err = json.Unmarshal(fileContent, &eventsData)
	if err != nil {
		log.Printf("Error unmarshalling events data: %v", err)
		http.Error(w, "Error loading events data", http.StatusInternalServerError)
		return
	}

	wins := make(map[string]int)
	losses := make(map[string]int)
	draws := make(map[string]int)
	points := make(map[string]int)
	matches := make(map[string]int)

	for _, match := range eventsData.Matches {
		result := aggregation.ParseMatchResult(match)

		if result.Draw {
			draws[match.Player1]++
			draws[match.Player2]++
			points[match.Player1] += 1
			points[match.Player2] += 1
		} else {
			wins[result.Winner]++
			losses[result.Loser]++
			points[result.Winner] += 3
			points[result.Loser] += 0
		}
		matches[match.Player1]++
		matches[match.Player2]++
	}

	type PlayerResult struct {
		Name     string
		Result   string
		Deck     string
		Decklist string
	}
	results := []PlayerResult{}

	keys := make([]string, 0, len(points))
	for key := range points {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		pi, pj := points[keys[i]], points[keys[j]]
		if pi != pj {
			return pi > pj // primary: points desc
		}
		if matches[keys[i]] != matches[keys[j]] {
			return matches[keys[i]] > matches[keys[j]] // secondary: matches desc
		}
		return keys[i] < keys[j] // thirdly: name asc
	})

	for _, key := range keys {
		result := fmt.Sprintf("%d-%d", wins[key], losses[key])
		if draws[key] > 0 {
			result += fmt.Sprintf("-%d", draws[key])
		}
		results = append(results, PlayerResult{
			Name:     key,
			Result:   result,
			Deck:     eventsData.PlayerInfo[key].Deck,
			Decklist: eventsData.PlayerInfo[key].Decklist,
		})
	}

	templateData := map[string]interface{}{
		"ActivePage": "events",
		"Name":       eventsData.Name,
		"Date":       eventID,
		"Results":    results,
		"Matches":    eventsData.Matches,
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

	filePath := "files/players/" + playerID + ".json"
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Player not found", http.StatusNotFound)
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

	templateData := map[string]interface{}{
		"ActivePage":   "leaderboards",
		"Leaderboards": leaderboardsData,
	}
	templates.RenderTemplate(w, "leaderboards.tmpl", templateData)
}
