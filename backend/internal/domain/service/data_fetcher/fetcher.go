package service

import (
	"fmt"
	"smart-alert/pkg/mf"
	"smart-alert/pkg/yahoo"
)

// DataFetcher defines the contract for any price data source.
// This interface allows the domain layer to remain agnostic of the underlying API.
type DataFetcher interface {
	FetchCurrentPrice(symbol string) (float64, error)
	FetchHistoricalData(symbol string, rangeStr string, interval string) ([]int64, []float64, error)
	FetchDMA200(symbol string) (float64, error)
	FetchMarketCap(symbol string) (float64, error)
	FetchPE(symbol string) (float64, error)
	FetchPriceNDaysAgo(symbol string, days int) (float64, error)
	Fetch52WeekHigh(symbol string) (float64, error)
}

// Internal implementation structs (unexported to encourage interface usage)
type stockFetcher struct{}
type mfFetcher struct{}

// --- Constructors ---

// NewStockFetcher returns a fetcher for standard equities (Yahoo Finance).
func NewStockFetcher() DataFetcher {
	return &stockFetcher{}
}

// NewMFFetcher returns a fetcher for Indian Mutual Funds (MFapi).
func NewMFFetcher() DataFetcher {
	return &mfFetcher{}
}

// GetFetcherBySymbol is a factory helper that detects the asset type from the symbol.
// Use this when the asset type is determined at runtime (e.g., from user input).
func GetFetcherBySymbol(symbol string) DataFetcher {
	if symbol == "" {
		return NewStockFetcher()
	}
	// AMFI Mutual Fund codes are numeric (e.g., "120503")
	numeric := true
	for _, r := range symbol {
		if r < '0' || r > '9' {
			numeric = false
			break
		}
	}
	if numeric {
		return NewMFFetcher()
	}
	return NewStockFetcher()
}

// --- stockFetcher Implementation ---

func (f *stockFetcher) FetchCurrentPrice(symbol string) (float64, error) {
	return yahoo.FetchCurrentPrice(symbol)
}

func (f *stockFetcher) FetchHistoricalData(symbol string, rangeStr string, interval string) ([]int64, []float64, error) {
	return yahoo.FetchHistoricalData(symbol, rangeStr, interval)
}

func (f *stockFetcher) FetchDMA200(symbol string) (float64, error) {
	return yahoo.FetchDMA200(symbol)
}

func (f *stockFetcher) FetchMarketCap(symbol string) (float64, error) {
	return yahoo.FetchMarketCap(symbol)
}

func (f *stockFetcher) FetchPE(symbol string) (float64, error) {
	return yahoo.FetchPE(symbol)
}

func (f *stockFetcher) FetchPriceNDaysAgo(symbol string, days int) (float64, error) {
	return yahoo.FetchPriceNDaysAgo(symbol, days)
}

func (f *stockFetcher) Fetch52WeekHigh(symbol string) (float64, error) {
	return yahoo.Fetch52WeekHigh(symbol)
}

// --- mfFetcher Implementation ---

func (f *mfFetcher) FetchCurrentPrice(symbol string) (float64, error) {
	return mf.FetchLatestNAV(symbol)
}

func (f *mfFetcher) FetchHistoricalData(symbol string, rangeStr string, interval string) ([]int64, []float64, error) {
	return mf.FetchHistoricalData(symbol, rangeStr, interval)
}

func (f *mfFetcher) FetchDMA200(symbol string) (float64, error) {
	return mf.FetchDMA200(symbol)
}

func (f *mfFetcher) FetchMarketCap(symbol string) (float64, error) {
	// Market cap is not typically used for Mutual Funds in this context
	return 0, fmt.Errorf("market cap not available for mutual funds")
}

func (f *mfFetcher) FetchPE(symbol string) (float64, error) {
	// PE ratio is not typically directly applicable or available via this simple API for MFs
	return 0, fmt.Errorf("PE ratio not available for mutual funds")
}

func (f *mfFetcher) FetchPriceNDaysAgo(symbol string, days int) (float64, error) {
	return mf.FetchNAVNDaysAgo(symbol, days)
}

func (f *mfFetcher) Fetch52WeekHigh(symbol string) (float64, error) {
	return mf.Fetch52WeekHigh(symbol)
}

// --- UniversalFetcher Implementation ---
// UniversalFetcher automatically routes requests to the appropriate fetcher
// based on symbol detection (numeric = MF, otherwise = Stock)

type universalFetcher struct {
	stockFetcher DataFetcher
	mfFetcher    DataFetcher
}

// NewUniversalFetcher creates a fetcher that can handle both stocks and MFs
func NewUniversalFetcher() DataFetcher {
	return &universalFetcher{
		stockFetcher: NewStockFetcher(),
		mfFetcher:    NewMFFetcher(),
	}
}

// isNumericSymbol checks if a symbol is all digits (MF AMFI code)
func isNumericSymbol(symbol string) bool {
	if symbol == "" {
		return false
	}
	for _, r := range symbol {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (f *universalFetcher) getFetcher(symbol string) DataFetcher {
	if isNumericSymbol(symbol) {
		return f.mfFetcher
	}
	return f.stockFetcher
}

func (f *universalFetcher) FetchCurrentPrice(symbol string) (float64, error) {
	return f.getFetcher(symbol).FetchCurrentPrice(symbol)
}

func (f *universalFetcher) FetchHistoricalData(symbol string, rangeStr string, interval string) ([]int64, []float64, error) {
	return f.getFetcher(symbol).FetchHistoricalData(symbol, rangeStr, interval)
}

func (f *universalFetcher) FetchDMA200(symbol string) (float64, error) {
	return f.getFetcher(symbol).FetchDMA200(symbol)
}

func (f *universalFetcher) FetchMarketCap(symbol string) (float64, error) {
	return f.getFetcher(symbol).FetchMarketCap(symbol)
}

func (f *universalFetcher) FetchPE(symbol string) (float64, error) {
	return f.getFetcher(symbol).FetchPE(symbol)
}

func (f *universalFetcher) FetchPriceNDaysAgo(symbol string, days int) (float64, error) {
	return f.getFetcher(symbol).FetchPriceNDaysAgo(symbol, days)
}

func (f *universalFetcher) Fetch52WeekHigh(symbol string) (float64, error) {
	return f.getFetcher(symbol).Fetch52WeekHigh(symbol)
}
