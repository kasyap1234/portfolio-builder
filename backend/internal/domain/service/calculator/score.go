package calculator

import (
	service "smart-alert/internal/domain/service/data_fetcher"
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

	// 1. Market Sentiment (Weight: 0.25)
	// If Nifty is below 200 DMA, it's a "Fear" market - better time to accumulate
	niftyPrice, err1 := s.fetcher.FetchCurrentPrice("^NSEI")
	niftyDMA200, err2 := s.fetcher.FetchDMA200("^NSEI")
	if err1 == nil && err2 == nil {
		if niftyPrice < niftyDMA200 {
			score += 0.25
		}
	}

	// 2. Asset Discount (Weight: 0.45)
	// Buying below 200 DMA is a classic Value/Accumulation signal
	if currentPrice < dma200 {
		score += 0.45
	}

	// 3. Stability Scoring (Weight: up to 0.30)
	// Favor Large Caps for stability, Mid/Small caps require more "crash" points to pass
	marketCap, err := s.fetcher.FetchMarketCap(assetName)
	if err == nil {
		if marketCap >= 1000000000000 {
			// Large Cap (> 1L Cr)
			score += 0.30
		} else if marketCap >= 300000000000 {
			// Mid Cap (30k - 1L Cr)
			score += 0.20
		} else {
			// Small Cap (< 30k Cr)
			score += 0.10
		}
	}
	// 4. Recent Fall (Drawdown) (Weight: up to 0.20)
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
				score += 0.20 // Deep Correction (20%+)
			} else if drawdown >= 0.10 {
				score += 0.10 // Healthy Correction (10-20%)
			} else if drawdown >= 0.05 {
				score += 0.05 // Minor Dip (5-10%)
			}
		}
	}

	return score
}


