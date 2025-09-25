package aggregation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func GenerateEventsList() error {
	err := os.MkdirAll("files/lists", 0755)
	if err != nil {
		return fmt.Errorf("failed to create lists directory: %w", err)
	}

	eventFiles, err := filepath.Glob("files/events/*.json")
	if err != nil {
		return fmt.Errorf("failed to read event files: %w", err)
	}

	events := []EventListItem{}

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

		// Create an event list item with name and date (same value for both)
		event := EventListItem{
			Name: eventData.Name,
			Date: eventData.Date,
			URL:  "/events/" + eventData.Name,
		}

		events = append(events, event)
	}

	// Sort events in reverse alphabetical order
	sort.Slice(events, func(i, j int) bool {
		return events[i].Name > events[j].Name
	})

	// Write the sorted events list to events.json
	eventsJSON, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal events list: %w", err)
	}

	if err := os.WriteFile("files/lists/events.json", eventsJSON, 0644); err != nil {
		return fmt.Errorf("failed to write events.json: %w", err)
	}

	return nil
}

// UpdateEvents generates the events list and should be called
// after any changes to event files
func UpdateEvents() error {
	return GenerateEventsList()
}
