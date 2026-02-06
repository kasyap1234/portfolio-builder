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

func TestIsNSEHoliday(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{"Republic Day 2025", time.Date(2025, 1, 26, 10, 0, 0, 0, ist), true},
		{"Republic Day 2026", time.Date(2026, 1, 26, 10, 0, 0, 0, ist), true},
		{"Independence Day 2025", time.Date(2025, 8, 15, 10, 0, 0, 0, ist), true},
		{"Christmas 2025", time.Date(2025, 12, 25, 10, 0, 0, 0, ist), true},
		{"Diwali 2025", time.Date(2025, 10, 20, 10, 0, 0, 0, ist), true},
		{"Holi 2026", time.Date(2026, 3, 4, 10, 0, 0, 0, ist), true},
		{"Regular day", time.Date(2025, 6, 10, 10, 0, 0, 0, ist), false},
		{"May Day 2026", time.Date(2026, 5, 1, 10, 0, 0, 0, ist), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNSEHoliday(tt.date)
			if result != tt.expected {
				t.Errorf("IsNSEHoliday(%s) = %v, want %v", tt.date.Format("2006-01-02"), result, tt.expected)
			}
		})
	}
}

func TestIsNonTradingDay(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{"Saturday", time.Date(2026, 2, 7, 12, 0, 0, 0, ist), true},
		{"Sunday", time.Date(2026, 2, 8, 12, 0, 0, 0, ist), true},
		{"Holiday on weekday", time.Date(2026, 1, 26, 12, 0, 0, 0, ist), true},
		{"Normal weekday", time.Date(2026, 2, 9, 12, 0, 0, 0, ist), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNonTradingDay(tt.date)
			if result != tt.expected {
				t.Errorf("IsNonTradingDay(%s) = %v, want %v", tt.date.Format("2006-01-02"), result, tt.expected)
			}
		})
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

		// May 2026: May 1 is Friday but it's May Day (NSE holiday), first trading is May 4 (Mon)
		{"May 1 2026 (May Day holiday)", time.Date(2026, 5, 1, 12, 0, 0, 0, ist), false},
		{"May 4 2026 (Monday - first trading)", time.Date(2026, 5, 4, 12, 0, 0, 0, ist), true},

		// January 2026: Jan 1 is Thursday (not a holiday), so first trading day is Jan 1
		{"Jan 1 2026 (Thursday - first trading)", time.Date(2026, 1, 1, 12, 0, 0, 0, ist), true},

		// October 2025: Oct 1 is Wednesday but Dussehra holiday, Oct 2 is Gandhi Jayanti, first trading is Oct 3
		{"Oct 1 2025 (Dussehra holiday)", time.Date(2025, 10, 1, 12, 0, 0, 0, ist), false},
		{"Oct 2 2025 (Gandhi Jayanti)", time.Date(2025, 10, 2, 12, 0, 0, 0, ist), false},
		{"Oct 3 2025 (Friday - first trading)", time.Date(2025, 10, 3, 12, 0, 0, 0, ist), true},
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

func TestGetFirstTradingDayOfMonth(t *testing.T) {
	tests := []struct {
		name     string
		year     int
		month    time.Month
		expected time.Time
	}{
		{"May 2026 skips May Day", 2026, time.May, time.Date(2026, 5, 4, 0, 0, 0, 0, ist)},
		{"Oct 2025 skips Dussehra+Gandhi Jayanti", 2025, time.October, time.Date(2025, 10, 3, 0, 0, 0, 0, ist)},
		{"Feb 2026 skips weekend", 2026, time.February, time.Date(2026, 2, 2, 0, 0, 0, 0, ist)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFirstTradingDayOfMonth(tt.year, tt.month, ist)
			if !result.Equal(tt.expected) {
				t.Errorf("GetFirstTradingDayOfMonth(%d, %s) = %s, want %s",
					tt.year, tt.month, result.Format("2006-01-02"), tt.expected.Format("2006-01-02"))
			}
		})
	}
}
