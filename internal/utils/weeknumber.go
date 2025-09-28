package utils

import "time"

// SwedishWeekNumber calculates the Swedish calendar week number for a given date.
func SwedishWeekNumber(date time.Time) int {
	_, week := date.ISOWeek()
	return week
}

func nextWednesday(date time.Time) time.Time {
	daysUntilWednesday := (3 - int(date.Weekday()) + 7) % 7
	nextWednesday := date.AddDate(0, 0, daysUntilWednesday)
	return nextWednesday
}

func NextEvent(date time.Time) time.Time {
	// Find the next Wednesday
	wednesday := nextWednesday(date)
	weekNo := SwedishWeekNumber(wednesday)
	if weekNo%2 != 0 {
		wednesday = wednesday.AddDate(0, 0, 1)
		wednesday = nextWednesday(wednesday)
	}
	return wednesday
}
