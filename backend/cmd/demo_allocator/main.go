package main

import (
	"fmt"
	"log"
	"smart-alert/internal/domain/service/allocator"
	service "smart-alert/internal/domain/service/data_fetcher"
)

func main() {
	fmt.Println("=== Smart SIP Allocator Demo ===")
	fmt.Println()

	// Create the universal fetcher (handles both stocks and MFs)
	fetcher := service.NewUniversalFetcher()

	// Create the allocator
	alloc := allocator.NewAllocator(fetcher)

	// Get current market status
	fmt.Println("Fetching market data...")
	niftyMetrics, err := alloc.CalculateDropMetrics("^NSEI")
	if err != nil {
		log.Fatalf("Failed to get Nifty metrics: %v", err)
	}

	fmt.Println()
	fmt.Println("=== Market Status (Nifty 50) ===")
	fmt.Printf("Current Price: %.2f\n", niftyMetrics.CurrentPrice)
	fmt.Printf("200 DMA:       %.2f\n", niftyMetrics.DMA200)
	fmt.Printf("DMA Distance:  %.2f%% %s\n", niftyMetrics.DMADistance, getDMALabel(niftyMetrics.DMADistance))
	fmt.Printf("Week Drop:     %.2f%%\n", niftyMetrics.WeekDrop)
	fmt.Printf("Month Drop:    %.2f%%\n", niftyMetrics.MonthDrop)
	fmt.Printf("High Drop:     %.2f%%\n", niftyMetrics.HighDrop)
	fmt.Printf("Composite:     %.2f\n", niftyMetrics.CompositeScore)
	fmt.Println()

	regime := alloc.DetermineMarketRegime(niftyMetrics)
	fmt.Printf("Market Regime: %s\n", regime)
	fmt.Println()

	// Example SIP amount and debt reserve
	sipAmount := 35000.0
	debtReserve := 100000.0 // Example: ₹1 lakh in debt reserve for panic buying
	fmt.Printf("=== SIP Allocation for ₹%.0f (Debt Reserve: ₹%.0f) ===\n", sipAmount, debtReserve)
	fmt.Println()

	recommendations, err := alloc.Allocate(sipAmount, nil, debtReserve, nil)
	if err != nil {
		log.Fatalf("Allocation failed: %v", err)
	}

	// Display recommendations by type
	var mfTotal, stockTotal, debtTotal float64
	var mfRecs, stockRecs, debtRecs []string

	for _, rec := range recommendations {
		switch rec.AssetType {
		case "MF":
			mfTotal += rec.Amount
			mfRecs = append(mfRecs, fmt.Sprintf("  %s: ₹%.0f - %s", rec.AssetSymbol, rec.Amount, rec.Reason))
		case "STOCK":
			stockTotal += rec.Amount
			stockRecs = append(stockRecs, fmt.Sprintf("  %s: ₹%.0f - %s", rec.AssetSymbol, rec.Amount, rec.Reason))
		case "DEBT":
			debtTotal += rec.Amount
			debtRecs = append(debtRecs, fmt.Sprintf("  %s: ₹%.0f - %s", rec.AssetSymbol, rec.Amount, rec.Reason))
		}
	}

	fmt.Println("--- Mutual Funds ---")
	if len(mfRecs) > 0 {
		for _, r := range mfRecs {
			fmt.Println(r)
		}
		fmt.Printf("  Total MF: ₹%.0f (%.1f%%)\n", mfTotal, mfTotal/sipAmount*100)
	} else {
		fmt.Println("  (none)")
	}
	fmt.Println()

	fmt.Println("--- Individual Stocks ---")
	if len(stockRecs) > 0 {
		for _, r := range stockRecs {
			fmt.Println(r)
		}
		fmt.Printf("  Total Stocks: ₹%.0f (%.1f%%)\n", stockTotal, stockTotal/sipAmount*100)
	} else {
		fmt.Println("  (none qualified - stocks not down enough vs Nifty)")
	}
	fmt.Println()

	fmt.Println("--- Debt ---")
	if len(debtRecs) > 0 {
		for _, r := range debtRecs {
			fmt.Println(r)
		}
		fmt.Printf("  Total Debt: ₹%.0f (%.1f%%)\n", debtTotal, debtTotal/sipAmount*100)
	} else {
		fmt.Println("  (none)")
	}
	fmt.Println()

	fmt.Println("=== Summary ===")
	fmt.Printf("Equity (MF + Stocks): ₹%.0f (%.1f%%)\n", mfTotal+stockTotal, (mfTotal+stockTotal)/sipAmount*100)
	fmt.Printf("Debt:                 ₹%.0f (%.1f%%)\n", debtTotal, debtTotal/sipAmount*100)
	fmt.Printf("Total:                ₹%.0f (100%%)\n", mfTotal+stockTotal+debtTotal)
}

func getDMALabel(distance float64) string {
	if distance >= 10 {
		return "(DEEP FEAR - below 200 DMA)"
	} else if distance > 0 {
		return "(FEAR - below 200 DMA)"
	} else if distance >= -5 {
		return "(NEUTRAL - near 200 DMA)"
	}
	return "(GREED - above 200 DMA)"
}
