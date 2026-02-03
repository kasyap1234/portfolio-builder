package scheduler

import (
	"time"
)

// IsWeekend checks if the given date is a weekend
func IsWeekend(t time.Time) bool {
	day := t.Weekday()
	return day == time.Saturday || day == time.Sunday
}

// IsFirstTradingDayOfMonth checks if today is the first trading day of the current month
func IsFirstTradingDayOfMonth(t time.Time) bool {
	// Get the first day of the current month
	firstOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())

	// Find the first trading day (skip weekends)
	firstTradingDay := firstOfMonth
	for IsWeekend(firstTradingDay) {
		firstTradingDay = firstTradingDay.AddDate(0, 0, 1)
	}

	// Compare just the date (ignore time)
	return t.Year() == firstTradingDay.Year() &&
		t.Month() == firstTradingDay.Month() &&
		t.Day() == firstTradingDay.Day()
}

// GetFirstTradingDayOfMonth returns the first trading day of the given month
func GetFirstTradingDayOfMonth(year int, month time.Month, loc *time.Location) time.Time {
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, loc)

	for IsWeekend(firstOfMonth) {
		firstOfMonth = firstOfMonth.AddDate(0, 0, 1)
	}

	return firstOfMonth
}
