package calculator

import (
	service "smart-alert/internal/domain/service/data_fetcher"
	"sync"
)

type ScoreCalculator interface {
	CalculateScore(assetName string, currentPrice float64, dma200 float64, historicalData []float64) float64
}

type SimpleScoreCalculator struct {
	fetcher service.DataFetcher
}

func NewSimpleScoreCalculator(fetcher service.DataFetcher) *SimpleScoreCalculator {
	return &SimpleScoreCalculator{
		fetcher: fetcher,
	}
}

const threshold = 0.85

func (s *SimpleScoreCalculator) CalculateScore(assetName string, currentPrice float64, dma200 float64, historicalData []float64) float64 {
	score := 0.0

	var (
		niftyPrice, niftyDMA200, marketCap float64
		niftyPriceErr, niftyDMAErr, mcErr  error
		wg                                 sync.WaitGroup
	)

	wg.Add(3)
	go func() { defer wg.Done(); niftyPrice, niftyPriceErr = s.fetcher.FetchCurrentPrice("^NSEI") }()
	go func() { defer wg.Done(); niftyDMA200, niftyDMAErr = s.fetcher.FetchDMA200("^NSEI") }()
	go func() { defer wg.Done(); marketCap, mcErr = s.fetcher.FetchMarketCap(assetName) }()
	wg.Wait()

	// 1. Market Sentiment (Weight: 0.20)
	if niftyPriceErr == nil && niftyDMAErr == nil {
		if niftyPrice < niftyDMA200 {
			score += 0.20
		}
	}

	// 2. Asset Discount (Weight: 0.40)
	if currentPrice < dma200 {
		score += 0.40
	}

	// 3. Stability Scoring (Weight: up to 0.25)
	if mcErr == nil {
		if marketCap >= 1000000000000 {
			score += 0.25
		} else if marketCap >= 300000000000 {
			score += 0.15
		} else {
			score += 0.08
		}
	}
	// 4. Recent Fall (Drawdown) (Weight: up to 0.15)
	// Calculate fall from recent High (ATH in the provided historical window)
	if len(historicalData) > 0 {
		highest := 0.0
		for _, p := range historicalData {
			if p > highest {
				highest = p
			}
		}

		if highest > 0 {
			drawdown := (highest - currentPrice) / highest
			if drawdown >= 0.20 {
				score += 0.15 // Deep Correction (20%+)
			} else if drawdown >= 0.10 {
				score += 0.08 // Healthy Correction (10-20%)
			} else if drawdown >= 0.05 {
				score += 0.04 // Minor Dip (5-10%)
			}
		}
	}

	if score > 1.0 {
		score = 1.0
	} else if score < 0.0 {
		score = 0.0
	}

	return score
}


