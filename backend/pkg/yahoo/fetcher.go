package yahoo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type chartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				Symbol             string  `json:"symbol"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Close []float64 `json:"close"`
				} `json:"quote"`
				Adjclose []struct {
					Adjclose []float64 `json:"adjclose"`
				} `json:"adjclose"`
			} `json:"indicators"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"chart"`
}

type quoteResponse struct {
	QuoteResponse struct {
		Result []struct {
			MarketCap  float64 `json:"marketCap"`
			TrailingPE float64 `json:"trailingPE"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"quoteResponse"`
}

// FetchCurrentPrice retrieves the current market price (LTP) for a given symbol.
func FetchCurrentPrice(symbol string) (float64, error) {
	if symbol == "" {
		return 0, errors.New("symbol cannot be empty")
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1m&range=1d", symbol)

	data, err := fetchChartData(url)
	if err != nil {
		return 0, err
	}

	return data.Chart.Result[0].Meta.RegularMarketPrice, nil
}

// FetchMarketCap retrieves the market capitalization for a given symbol.
func FetchMarketCap(symbol string) (float64, error) {
	if symbol == "" {
		return 0, errors.New("symbol cannot be empty")
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?symbols=%s", symbol)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to fetch data: status code %d", resp.StatusCode)
	}

	var data quoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("error decoding response: %v", err)
	}

	if data.QuoteResponse.Error != nil {
		return 0, fmt.Errorf("yahoo finance error: %v", data.QuoteResponse.Error)
	}

	if len(data.QuoteResponse.Result) == 0 {
		return 0, fmt.Errorf("no data found for symbol %s", symbol)
	}

	return data.QuoteResponse.Result[0].MarketCap, nil
}

// FetchPE retrieves the trailing PE ratio for a given symbol.
func FetchPE(symbol string) (float64, error) {
	if symbol == "" {
		return 0, errors.New("symbol cannot be empty")
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?symbols=%s", symbol)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to fetch data: status code %d", resp.StatusCode)
	}

	var data quoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("error decoding response: %v", err)
	}

	if data.QuoteResponse.Error != nil {
		return 0, fmt.Errorf("yahoo finance error: %v", data.QuoteResponse.Error)
	}

	if len(data.QuoteResponse.Result) == 0 {
		return 0, fmt.Errorf("no data found for symbol %s", symbol)
	}

	return data.QuoteResponse.Result[0].TrailingPE, nil
}

// FetchHistoricalData retrieves historical timestamps and closing prices for a given range and interval.
// Example range: "5d", "1mo", "1y", "2y", "max"
// Example interval: "1m", "5m", "1d", "1wk", "1mo"
func FetchHistoricalData(symbol string, rangeStr string, interval string) ([]int64, []float64, error) {
	if symbol == "" {
		return nil, nil, errors.New("symbol cannot be empty")
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?range=%s&interval=%s", symbol, rangeStr, interval)

	data, err := fetchChartData(url)
	if err != nil {
		return nil, nil, err
	}

	result := data.Chart.Result[0]
	if len(result.Timestamp) == 0 {
		return nil, nil, fmt.Errorf("no historical data found for symbol %s in range %s", symbol, rangeStr)
	}

	// Use closing prices from the first quote indicator
	if len(result.Indicators.Quote) == 0 || len(result.Indicators.Quote[0].Close) == 0 {
		return nil, nil, fmt.Errorf("no closing prices found for symbol %s", symbol)
	}

	return result.Timestamp, result.Indicators.Quote[0].Close, nil
}

// FetchDMA200 calculates the 200-day Daily Moving Average.
func FetchDMA200(symbol string) (float64, error) {
	// Fetching 2 years of daily data to ensure we have enough points (approx 252 trading days per year)
	_, prices, err := FetchHistoricalData(symbol, "2y", "1d")
	if err != nil {
		return 0, err
	}

	// Filter out null/zero prices if any (Yahoo sometimes returns nulls for some days)
	var validPrices []float64
	for _, p := range prices {
		if p > 0 {
			validPrices = append(validPrices, p)
		}
	}

	if len(validPrices) < 200 {
		return 0, fmt.Errorf("insufficient data for DMA200: found %d valid points, need 200", len(validPrices))
	}

	// Calculate average of the last 200 points
	sum := 0.0
	count := 0
	for i := len(validPrices) - 1; i >= 0 && count < 200; i-- {
		sum += validPrices[i]
		count++
	}

	return sum / float64(count), nil
}

// FetchPriceNDaysAgo retrieves the closing price from approximately N days ago
func FetchPriceNDaysAgo(symbol string, days int) (float64, error) {
	rangeStr := "1mo"
	if days > 30 {
		rangeStr = "3mo"
	}
	if days > 90 {
		rangeStr = "1y"
	}

	_, prices, err := FetchHistoricalData(symbol, rangeStr, "1d")
	if err != nil {
		return 0, err
	}

	// Filter out zero/invalid prices
	var validPrices []float64
	for _, p := range prices {
		if p > 0 {
			validPrices = append(validPrices, p)
		}
	}

	if len(validPrices) == 0 {
		return 0, fmt.Errorf("no valid prices found for %s", symbol)
	}

	// Get the price from approximately N days ago
	// Trading days are roughly 5 per week, so approximate index
	targetIndex := len(validPrices) - 1 - days
	if targetIndex < 0 {
		targetIndex = 0
	}

	return validPrices[targetIndex], nil
}

// FetchRecentHigh retrieves the highest price in the given range (e.g., "1y" for 52-week high)
func FetchRecentHigh(symbol string, rangeStr string) (float64, error) {
	_, prices, err := FetchHistoricalData(symbol, rangeStr, "1d")
	if err != nil {
		return 0, err
	}

	highest := 0.0
	for _, p := range prices {
		if p > highest {
			highest = p
		}
	}

	if highest == 0 {
		return 0, fmt.Errorf("no valid prices found for %s", symbol)
	}

	return highest, nil
}

// Fetch52WeekHigh is a convenience function for getting the 52-week high
func Fetch52WeekHigh(symbol string) (float64, error) {
	return FetchRecentHigh(symbol, "1y")
}

// fetchChartData is a private helper to perform the external API call
func fetchChartData(url string) (*chartResponse, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data: status code %d", resp.StatusCode)
	}

	var data chartResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	if data.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo finance error: %v", data.Chart.Error)
	}

	if len(data.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data found in response")
	}

	return &data, nil
}
