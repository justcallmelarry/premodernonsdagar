package utils

import (
	"testing"
	"time"
)

func TestSwedishWeekNumber(t *testing.T) {
	tests := []struct {
		date     time.Time
		expected int
	}{
		{time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), 52},
		// Regular year, first week
		{time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), 1},
		// Regular year, second week
		{time.Date(2023, 1, 9, 0, 0, 0, 0, time.UTC), 2},
		// Leap year, after February 29th
		{time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), 9},
		// End of year, week 52
		{time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC), 52},
		// Start of year, week 1
		{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 1},
	}

	for _, test := range tests {
		result := SwedishWeekNumber(test.date)
		if result != test.expected {
			t.Errorf("For date %v, expected week %d, but got %d", test.date, test.expected, result)
		}
	}
}

func TestNextEvent(t *testing.T) {
	tests := []struct {
		date     time.Time
		expected time.Time
	}{
		{time.Date(2025, 9, 28, 0, 0, 0, 0, time.UTC), time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)},
		{time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)},
		{time.Date(2025, 10, 2, 0, 0, 0, 0, time.UTC), time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC)},
	}

	for _, test := range tests {
		result := NextEvent(test.date)
		if !result.Equal(test.expected) {
			t.Errorf("For date %v, expected next event on %v, but got %v", test.date, test.expected, result)
		}
	}
}
