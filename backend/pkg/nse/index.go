package nse

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ErrNotIndexSymbol is returned when a symbol is not mapped to an NSE index.
var ErrNotIndexSymbol = errors.New("symbol is not a supported NSE index")

const nifty50Index = "NIFTY 50"

// allIndicesURL is a var so tests can override it with httptest.
var allIndicesURL = "https://www.nseindia.com/api/allIndices"

type allIndicesResponse struct {
	Data []indexQuote `json:"data"`
}

type indexQuote struct {
	Index string `json:"index"`
	PE    string `json:"pe"`
}

var indexSymbolMap = map[string]string{
	"^NSEI":   nifty50Index,
	"NSEI":    nifty50Index,
	"NIFTY50": nifty50Index,
	"NIFTY 50": nifty50Index,
}

// IsIndexSymbol reports whether symbol maps to an NSE index with published PE.
func IsIndexSymbol(symbol string) bool {
	_, ok := ResolveIndexName(symbol)
	return ok
}

// ResolveIndexName maps Yahoo or shorthand symbols to NSE index names.
func ResolveIndexName(symbol string) (string, bool) {
	key := strings.TrimSpace(strings.ToUpper(symbol))
	if name, ok := indexSymbolMap[key]; ok {
		return name, true
	}
	// Allow direct NSE index names (e.g. "NIFTY BANK")
	if key != "" && strings.HasPrefix(key, "NIFTY") {
		return strings.ReplaceAll(key, "  ", " "), true
	}
	return "", false
}

// FetchIndexPE returns the trailing P/E for an NSE index (e.g. "NIFTY 50").
func FetchIndexPE(indexName string) (float64, error) {
	indexName = strings.TrimSpace(indexName)
	if indexName == "" {
		return 0, errors.New("index name cannot be empty")
	}

	quotes, err := fetchAllIndices()
	if err != nil {
		return 0, err
	}

	target := strings.ToUpper(indexName)
	for _, q := range quotes {
		if strings.ToUpper(strings.TrimSpace(q.Index)) == target {
			return parsePE(q.PE, q.Index)
		}
	}

	return 0, fmt.Errorf("PE not found for index %q", indexName)
}

// FetchPEForSymbol returns index PE when symbol is a known index; otherwise ErrNotIndexSymbol.
func FetchPEForSymbol(symbol string) (float64, error) {
	indexName, ok := ResolveIndexName(symbol)
	if !ok {
		return 0, ErrNotIndexSymbol
	}
	return FetchIndexPE(indexName)
}

func parsePE(raw, index string) (float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "-" || raw == "0" {
		return 0, fmt.Errorf("PE unavailable for index %s", index)
	}
	pe, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid PE %q for index %s: %w", raw, index, err)
	}
	if pe <= 0 {
		return 0, fmt.Errorf("invalid PE %.4f for index %s", pe, index)
	}
	return pe, nil
}

func fetchAllIndices() ([]indexQuote, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest(http.MethodGet, allIndicesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create NSE request: %w", err)
	}
	setNSEHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NSE allIndices returned status %d", resp.StatusCode)
	}

	var payload allIndicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode NSE response: %w", err)
	}
	if len(payload.Data) == 0 {
		return nil, errors.New("NSE allIndices returned no data")
	}

	return payload.Data, nil
}

func setNSEHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://www.nseindia.com/market-data/live-market-indices")
}
