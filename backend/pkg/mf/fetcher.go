package mf

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type mfResponse struct {
	Meta struct {
		SchemeName string `json:"scheme_name"`
		SchemeCode int    `json:"scheme_code"`
	} `json:"meta"`
	Data []struct {
		Date string `json:"date"`
		Nav  string `json:"nav"`
	} `json:"data"`
	Status string `json:"status"`
}

// FetchLatestNAV retrieves the latest NAV for a given mutual fund scheme code.
func FetchLatestNAV(schemeCode string) (float64, error) {
	url := fmt.Sprintf("https://api.mfapi.in/mf/%s/latest", schemeCode)

	data, err := fetchMFData(url)
	if err != nil {
		return 0, err
	}

	if len(data.Data) == 0 {
		return 0, fmt.Errorf("no NAV data found for scheme %s", schemeCode)
	}

	nav, err := strconv.ParseFloat(data.Data[0].Nav, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing NAV: %v", err)
	}

	return nav, nil
}

// FetchHistoricalNAV retrieves historical NAVs for a given mutual fund scheme code.
func FetchHistoricalNAV(schemeCode string) ([]string, []float64, error) {
	url := fmt.Sprintf("https://api.mfapi.in/mf/%s", schemeCode)

	data, err := fetchMFData(url)
	if err != nil {
		return nil, nil, err
	}

	var dates []string
	var navs []float64

	for _, d := range data.Data {
		nav, err := strconv.ParseFloat(d.Nav, 64)
		if err != nil {
			continue // Skip invalid entries
		}
		dates = append(dates, d.Date)
		navs = append(navs, nav)
	}

	return dates, navs, nil
}

// FetchHistoricalData retrieves historical timestamps and NAVs for a given range and interval.
// Supported ranges: "5d", "1mo", "3mo", "6mo", "1y", "2y", "5y", "max"
// Interval is currently ignored for MF as the API only provides daily data.
func FetchHistoricalData(schemeCode string, rangeStr string, interval string) ([]int64, []float64, error) {
	url := fmt.Sprintf("https://api.mfapi.in/mf/%s", schemeCode)

	data, err := fetchMFData(url)
	if err != nil {
		return nil, nil, err
	}

	var timestamps []int64
	var navs []float64

	// Determine cutoff time for range filtering
	cutoff := time.Time{}
	if rangeStr != "max" {
		now := time.Now()
		switch rangeStr {
		case "5d":
			cutoff = now.AddDate(0, 0, -5)
		case "1mo":
			cutoff = now.AddDate(0, -1, 0)
		case "3mo":
			cutoff = now.AddDate(0, -3, 0)
		case "6mo":
			cutoff = now.AddDate(0, -6, 0)
		case "1y":
			cutoff = now.AddDate(-1, 0, 0)
		case "2y":
			cutoff = now.AddDate(-2, 0, 0)
		case "5y":
			cutoff = now.AddDate(-5, 0, 0)
		default:
			return nil, nil, fmt.Errorf("unsupported range: %s", rangeStr)
		}
	}

	// MF API returns data in reverse chronological order (newest first)
	// We iterate backwards to return chronological order (oldest first) like Yahoo does
	for i := len(data.Data) - 1; i >= 0; i-- {
		d := data.Data[i]
		t, err := time.Parse("02-01-2006", d.Date)
		if err != nil {
			continue
		}

		if !cutoff.IsZero() && t.Before(cutoff) {
			continue
		}

		nav, err := strconv.ParseFloat(d.Nav, 64)
		if err != nil {
			continue
		}

		timestamps = append(timestamps, t.Unix())
		navs = append(navs, nav)
	}

	return timestamps, navs, nil
}

// FetchDMA200 calculates the 200-day Daily Moving Average for a mutual fund.
func FetchDMA200(schemeCode string) (float64, error) {
	// We need 200 points, so we'll just fetch all data and take the last 200 from the RAW response (newest first)
	url := fmt.Sprintf("https://api.mfapi.in/mf/%s", schemeCode)
	data, err := fetchMFData(url)
	if err != nil {
		return 0, err
	}

	if len(data.Data) < 200 {
		return 0, fmt.Errorf("insufficient data for DMA200: found %d points, need 200", len(data.Data))
	}

	sum := 0.0
	count := 0
	for i := 0; i < len(data.Data) && count < 200; i++ {
		nav, err := strconv.ParseFloat(data.Data[i].Nav, 64)
		if err != nil {
			continue
		}
		sum += nav
		count++
	}

	if count < 200 {
		return 0, fmt.Errorf("insufficient valid data points for DMA200")
	}

	return sum / 200.0, nil
}

// FetchNAVNDaysAgo retrieves the NAV from approximately N days ago
func FetchNAVNDaysAgo(schemeCode string, days int) (float64, error) {
	url := fmt.Sprintf("https://api.mfapi.in/mf/%s", schemeCode)
	data, err := fetchMFData(url)
	if err != nil {
		return 0, err
	}

	if len(data.Data) <= days {
		// Not enough data, return the oldest available
		if len(data.Data) > 0 {
			nav, err := strconv.ParseFloat(data.Data[len(data.Data)-1].Nav, 64)
			if err != nil {
				return 0, fmt.Errorf("error parsing NAV: %v", err)
			}
			return nav, nil
		}
		return 0, fmt.Errorf("no NAV data found for scheme %s", schemeCode)
	}

	// MF API returns data in reverse chronological order (newest first)
	// Index 0 = today, index 7 = ~7 days ago (accounting for weekends)
	nav, err := strconv.ParseFloat(data.Data[days].Nav, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing NAV: %v", err)
	}

	return nav, nil
}

// FetchRecentHigh retrieves the highest NAV in the given range
func FetchRecentHigh(schemeCode string, rangeStr string) (float64, error) {
	_, navs, err := FetchHistoricalData(schemeCode, rangeStr, "1d")
	if err != nil {
		return 0, err
	}

	highest := 0.0
	for _, n := range navs {
		if n > highest {
			highest = n
		}
	}

	if highest == 0 {
		return 0, fmt.Errorf("no valid NAV data found for scheme %s", schemeCode)
	}

	return highest, nil
}

// Fetch52WeekHigh is a convenience function for getting the 52-week high NAV
func Fetch52WeekHigh(schemeCode string) (float64, error) {
	return FetchRecentHigh(schemeCode, "1y")
}

func fetchMFData(url string) (*mfResponse, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch MF data: status code %d", resp.StatusCode)
	}

	var data mfResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	if data.Status != "SUCCESS" {
		return nil, fmt.Errorf("API returned status: %s", data.Status)
	}

	return &data, nil
}
