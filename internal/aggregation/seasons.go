package aggregation

import (
	"fmt"
	"sort"
	"time"
)

func GetSeason(date string, firstEventDate string) (string, error) {
	eventDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", fmt.Errorf("failed to parse event date: %w", err)
	}

	firstDate, err := time.Parse("2006-01-02", firstEventDate)
	if err != nil {
		return "", fmt.Errorf("failed to parse first event date: %w", err)
	}

	// Seasons are January-June or July-December, starting from the first event
	firstMonth := firstDate.Month()
	firstYear := firstDate.Year()

	var season1Start time.Time
	if firstMonth <= 6 {
		season1Start = time.Date(firstYear, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		season1Start = time.Date(firstYear, 7, 1, 0, 0, 0, 0, time.UTC)
	}

	seasonNumber := 1
	currentSeasonStart := season1Start

	for {
		var nextSeasonStart time.Time
		if currentSeasonStart.Month() == 1 {
			nextSeasonStart = time.Date(currentSeasonStart.Year(), 7, 1, 0, 0, 0, 0, time.UTC)
		} else {
			nextSeasonStart = time.Date(currentSeasonStart.Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)
		}

		if eventDate.Before(nextSeasonStart) {
			return fmt.Sprintf("s%02d", seasonNumber), nil
		}

		currentSeasonStart = nextSeasonStart
		seasonNumber++
	}
}

func GetAllSeasons(eventDates []string) ([]string, error) {
	if len(eventDates) == 0 {
		return []string{}, nil
	}

	sortedDates := make([]string, len(eventDates))
	copy(sortedDates, eventDates)
	sort.Strings(sortedDates)
	firstEventDate := sortedDates[0]

	seasonsMap := make(map[string]bool)
	for _, date := range eventDates {
		season, err := GetSeason(date, firstEventDate)
		if err != nil {
			return nil, err
		}
		seasonsMap[season] = true
	}

	seasons := make([]string, 0, len(seasonsMap))
	for season := range seasonsMap {
		seasons = append(seasons, season)
	}
	sort.Strings(seasons)

	return seasons, nil
}

func GetCurrentSeason(eventDates []string) (string, error) {
	if len(eventDates) == 0 {
		return "", fmt.Errorf("no events available")
	}

	sortedDates := make([]string, len(eventDates))
	copy(sortedDates, eventDates)
	sort.Strings(sortedDates)

	firstEventDate := sortedDates[0]
	lastEventDate := sortedDates[len(sortedDates)-1]

	return GetSeason(lastEventDate, firstEventDate)
}
