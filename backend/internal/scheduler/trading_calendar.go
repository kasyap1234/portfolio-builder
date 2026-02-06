package scheduler

import (
	"time"
)

var ist = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		panic("failed to load IST location: " + err.Error())
	}
	return loc
}()

var nseHolidays = func() map[time.Time]bool {
	holidays := map[time.Time]bool{}
	dates := []time.Time{
		// 2025
		time.Date(2025, time.January, 26, 0, 0, 0, 0, ist),  // Republic Day
		time.Date(2025, time.February, 26, 0, 0, 0, 0, ist),  // Maha Shivaratri
		time.Date(2025, time.March, 14, 0, 0, 0, 0, ist),     // Holi
		time.Date(2025, time.March, 31, 0, 0, 0, 0, ist),     // Id-Ul-Fitr (Eid)
		time.Date(2025, time.April, 10, 0, 0, 0, 0, ist),     // Shri Mahavir Jayanti
		time.Date(2025, time.April, 14, 0, 0, 0, 0, ist),     // Dr. Ambedkar Jayanti
		time.Date(2025, time.April, 18, 0, 0, 0, 0, ist),     // Good Friday
		time.Date(2025, time.May, 1, 0, 0, 0, 0, ist),        // May Day
		time.Date(2025, time.June, 7, 0, 0, 0, 0, ist),       // Eid-Ul-Adha (Bakri Eid)
		time.Date(2025, time.August, 15, 0, 0, 0, 0, ist),    // Independence Day
		time.Date(2025, time.August, 27, 0, 0, 0, 0, ist),    // Ganesh Chaturthi
		time.Date(2025, time.October, 1, 0, 0, 0, 0, ist),    // Dussehra
		time.Date(2025, time.October, 2, 0, 0, 0, 0, ist),    // Mahatma Gandhi Jayanti
		time.Date(2025, time.October, 20, 0, 0, 0, 0, ist),   // Diwali (Laxmi Puja)
		time.Date(2025, time.October, 21, 0, 0, 0, 0, ist),   // Diwali Balipratipada
		time.Date(2025, time.November, 5, 0, 0, 0, 0, ist),   // Gurunanak Jayanti
		time.Date(2025, time.December, 25, 0, 0, 0, 0, ist),  // Christmas

		// 2026
		time.Date(2026, time.January, 26, 0, 0, 0, 0, ist),   // Republic Day
		time.Date(2026, time.February, 17, 0, 0, 0, 0, ist),  // Maha Shivaratri
		time.Date(2026, time.March, 4, 0, 0, 0, 0, ist),      // Holi
		time.Date(2026, time.March, 20, 0, 0, 0, 0, ist),     // Id-Ul-Fitr (Eid)
		time.Date(2026, time.March, 25, 0, 0, 0, 0, ist),     // Shri Mahavir Jayanti
		time.Date(2026, time.April, 3, 0, 0, 0, 0, ist),      // Good Friday
		time.Date(2026, time.April, 14, 0, 0, 0, 0, ist),     // Dr. Ambedkar Jayanti
		time.Date(2026, time.May, 1, 0, 0, 0, 0, ist),        // May Day
		time.Date(2026, time.May, 28, 0, 0, 0, 0, ist),       // Eid-Ul-Adha (Bakri Eid)
		time.Date(2026, time.August, 15, 0, 0, 0, 0, ist),    // Independence Day
		time.Date(2026, time.August, 17, 0, 0, 0, 0, ist),    // Ganesh Chaturthi
		time.Date(2026, time.October, 2, 0, 0, 0, 0, ist),    // Mahatma Gandhi Jayanti
		time.Date(2026, time.October, 9, 0, 0, 0, 0, ist),    // Diwali (Laxmi Puja)
		time.Date(2026, time.October, 20, 0, 0, 0, 0, ist),   // Dussehra
		time.Date(2026, time.November, 24, 0, 0, 0, 0, ist),  // Gurunanak Jayanti
		time.Date(2026, time.December, 25, 0, 0, 0, 0, ist),  // Christmas
	}
	for _, d := range dates {
		holidays[d] = true
	}
	return holidays
}()

func IsWeekend(t time.Time) bool {
	day := t.Weekday()
	return day == time.Saturday || day == time.Sunday
}

func IsNSEHoliday(t time.Time) bool {
	normalized := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, ist)
	return nseHolidays[normalized]
}

func IsNonTradingDay(t time.Time) bool {
	return IsWeekend(t) || IsNSEHoliday(t)
}

func IsFirstTradingDayOfMonth(t time.Time) bool {
	firstOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	firstTradingDay := firstOfMonth
	for IsNonTradingDay(firstTradingDay) {
		firstTradingDay = firstTradingDay.AddDate(0, 0, 1)
	}
	return t.Year() == firstTradingDay.Year() &&
		t.Month() == firstTradingDay.Month() &&
		t.Day() == firstTradingDay.Day()
}

func GetFirstTradingDayOfMonth(year int, month time.Month, loc *time.Location) time.Time {
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	for IsNonTradingDay(firstOfMonth) {
		firstOfMonth = firstOfMonth.AddDate(0, 0, 1)
	}
	return firstOfMonth
}
