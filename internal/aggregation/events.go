package aggregation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

	eventFiles, err := filepath.Glob("files/events/*.json")
	if err != nil {
		return fmt.Errorf("failed to read event files: %w", err)
	}

	eventsOutputData := EventListStats{
		Count:  len(eventFiles),
		Events: []EventListItem{},
	}

	attendances := []int{}

	for _, eventFile := range eventFiles {
		// Read event file
		data, err := os.ReadFile(eventFile)
		if err != nil {
			return fmt.Errorf("failed to read event file %s: %w", eventFile, err)
		}

		// Parse event data to get the date
		var eventData EventData
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
