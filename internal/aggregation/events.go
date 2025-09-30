package aggregation

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"premodernonsdagar/internal/utils"
	"sort"
	"strings"
)

func generateEventsList() error {
	err := os.MkdirAll("files/lists", 0755)
	if err != nil {
		return fmt.Errorf("failed to create lists directory: %w", err)
	}

	err = os.MkdirAll("files/events", 0755)
	if err != nil {
		return fmt.Errorf("failed to create lists directory: %w", err)
	}

	eventFiles, err := filepath.Glob("input/events/*.json")
	if err != nil {
		return fmt.Errorf("failed to read event files: %w", err)
	}

	eventsOutputData := EventListStats{
		Count:  len(eventFiles),
		Events: []EventListItem{},
	}

	attendances := []int{}

	// Collect existing event JSON files
	existingEventFiles := make(map[string]bool)
	err = filepath.WalkDir("files/events", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), ".json") {
			existingEventFiles[path] = true
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to collect existing event files: %w", err)
	}

	for _, eventFile := range eventFiles {
		// Read event file
		data, err := os.ReadFile(eventFile)
		if err != nil {
			return fmt.Errorf("failed to read event file %s: %w", eventFile, err)
		}

		// Parse event data to get the date
		var eventData InputEvent
		if err := json.Unmarshal(data, &eventData); err != nil {
			return fmt.Errorf("failed to parse event file %s: %w", eventFile, err)
		}

		uniquePlayers := make(map[string]struct{})
		for match := range eventData.Matches {
			uniquePlayers[eventData.Matches[match].Player1] = struct{}{}
			uniquePlayers[eventData.Matches[match].Player2] = struct{}{}
		}
		attendance := len(uniquePlayers)

		if attendance > eventsOutputData.MaxAttendance {
			eventsOutputData.MaxAttendance = attendance
		}
		if eventsOutputData.MinAttendance == 0 || attendance < eventsOutputData.MinAttendance {
			eventsOutputData.MinAttendance = attendance
		}

		attendances = append(attendances, attendance)

		// Create an event list item with name and date (same value for both)
		event := EventListItem{
			Name: eventData.Name,
			Date: eventData.Date,
			URL:  "/events/" + eventData.Date,
		}

		eventsOutputData.Events = append(eventsOutputData.Events, event)

		wins := make(map[string]int)
		losses := make(map[string]int)
		draws := make(map[string]int)
		points := make(map[string]int)
		matches := make(map[string]int)

		for _, match := range eventData.Matches {
			result := ParseMatchResult(match)

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
				Deck:     eventData.PlayerInfo[key].Deck,
				Decklist: eventData.PlayerInfo[key].Decklist,
				URL:      "/players/" + utils.Slugify(key),
			})
		}

		outputEvent := Event{
			Name:    eventData.Name,
			Date:    eventData.Date,
			Rounds:  eventData.Rounds,
			Matches: eventData.Matches,
			Results: results,
		}

		// Save the updated event data back to its file
		updatedEventJSON, err := json.MarshalIndent(outputEvent, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal updated event data for %s: %w", eventFile, err)
		}

		outputFilePath := "files/events/" + outputEvent.Date + ".json"
		if err := os.WriteFile(outputFilePath, updatedEventJSON, 0644); err != nil {
			return fmt.Errorf("failed to write updated event data to %s: %w", eventFile, err)
		}

		// Mark this file as regenerated
		delete(existingEventFiles, outputFilePath)
	}

	// Remove any leftover files that were not regenerated
	for filePath := range existingEventFiles {
		err := os.Remove(filePath)
		if err != nil {
			return fmt.Errorf("failed to remove old event file %s: %w", filePath, err)
		}
	}

	// Sort events in reverse alphabetical order
	sort.Slice(eventsOutputData.Events, func(i, j int) bool {
		return eventsOutputData.Events[i].Name > eventsOutputData.Events[j].Name
	})

	// Calculate average attendance
	totalAttendance := 0
	for _, att := range attendances {
		totalAttendance += att
	}
	if len(attendances) > 0 {
		eventsOutputData.AverageAttendance = float64(totalAttendance) / float64(len(attendances))
	}

	// Write the sorted events list to events.json
	eventsJSON, err := json.MarshalIndent(eventsOutputData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal events list: %w", err)
	}

	if err := os.WriteFile("files/lists/events.json", eventsJSON, 0644); err != nil {
		return fmt.Errorf("failed to write events.json: %w", err)
	}

	return nil
}
