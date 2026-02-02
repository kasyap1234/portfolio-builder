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
