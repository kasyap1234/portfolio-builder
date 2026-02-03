package scheduler

import (
	"testing"
	"time"
)

func TestIsWeekend(t *testing.T) {
	tests := []struct {
		date     time.Time
		expected bool
	}{
		{time.Date(2026, 2, 7, 12, 0, 0, 0, time.UTC), true},   // Saturday
		{time.Date(2026, 2, 8, 12, 0, 0, 0, time.UTC), true},   // Sunday
		{time.Date(2026, 2, 9, 12, 0, 0, 0, time.UTC), false},  // Monday
		{time.Date(2026, 2, 13, 12, 0, 0, 0, time.UTC), false}, // Friday
	}

	for _, tt := range tests {
		result := IsWeekend(tt.date)
		if result != tt.expected {
			t.Errorf("IsWeekend(%s) = %v, want %v", tt.date.Format("Mon"), result, tt.expected)
		}
	}
}

func TestIsFirstTradingDayOfMonth(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		// February 2026: Feb 1 is Sunday, so first trading day is Feb 2 (Monday)
		{"Feb 1 2026 (Sunday)", time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC), false},
		{"Feb 2 2026 (Monday - first trading)", time.Date(2026, 2, 2, 12, 0, 0, 0, time.UTC), true},
		{"Feb 3 2026 (Tuesday)", time.Date(2026, 2, 3, 12, 0, 0, 0, time.UTC), false},

		// March 2026: Mar 1 is Sunday, first trading day is Mar 2
		{"Mar 1 2026 (Sunday)", time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC), false},
		{"Mar 2 2026 (Monday - first trading)", time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC), true},

		// April 2026: Apr 1 is Wednesday
		{"Apr 1 2026 (Wednesday - first trading)", time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC), true},
		{"Apr 2 2026 (Thursday)", time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFirstTradingDayOfMonth(tt.date)
			if result != tt.expected {
				t.Errorf("IsFirstTradingDayOfMonth(%s) = %v, want %v", tt.date.Format("2006-01-02"), result, tt.expected)
			}
		})
	}
}
