package main

import (
	"fmt"
	"log"
	"strings"

	"smart-alert/internal/domain/models"
	"smart-alert/internal/domain/service/allocator"
	service "smart-alert/internal/domain/service/data_fetcher"
	"smart-alert/pkg/allocationreport"
)

func main() {
	fmt.Println("Smart SIP Allocator — Demo")
	fmt.Println()

	fetcher := service.NewUniversalFetcher()
	alloc := allocator.NewAllocator(fetcher)

	fmt.Println("Fetching market data…")
	niftyMetrics, err := alloc.CalculateDropMetrics("^NSEI")
	if err != nil {
		log.Fatalf("Failed to get Nifty metrics: %v", err)
	}

	niftyPE, peErr := fetcher.FetchPE("^NSEI")
	regimePE := 0.0
	if peErr == nil {
		regimePE = niftyPE
	}
	regime := alloc.DetermineMarketRegime(niftyMetrics, regimePE)

	fmt.Println()
	fmt.Print(formatMarketStatus(niftyMetrics, niftyPE, peErr, regime))
	fmt.Println()

	sipAmount := 45000.0
	debtReserve := 120000.0

	recommendations, err := alloc.Allocate(sipAmount, nil, debtReserve, nil)
	if err != nil {
		log.Fatalf("Allocation failed: %v", err)
	}

	opts := allocationreport.Options{
		SIPAmount:   sipAmount,
		DebtReserve: debtReserve,
		Regime:      string(regime),
		Title:       "SIP Allocation (demo)",
	}
	if peErr == nil {
		opts.NiftyPE = niftyPE
	}

	fmt.Println(allocationreport.FormatConsole(recommendations, opts))
}

func formatMarketStatus(m *models.DropMetrics, niftyPE float64, peErr error, regime allocator.MarketRegime) string {
	var sb strings.Builder
	sb.WriteString(strings.Repeat("═", 44) + "\n")
	sb.WriteString(" Market snapshot (Nifty 50)\n")
	sb.WriteString(strings.Repeat("═", 44) + "\n")
	sb.WriteString(fmt.Sprintf(" Price:          ₹%.2f\n", m.CurrentPrice))
	sb.WriteString(fmt.Sprintf(" 200-DMA:        ₹%.2f\n", m.DMA200))
	sb.WriteString(fmt.Sprintf(" vs 200-DMA:     %.2f%%  %s\n", m.DMADistance, dmaLabel(m.DMADistance)))
	sb.WriteString(fmt.Sprintf(" Week / month:   %.2f%% / %.2f%%\n", m.WeekDrop, m.MonthDrop))
	sb.WriteString(fmt.Sprintf(" From 52w high:  %.2f%%\n", m.HighDrop))
	sb.WriteString(fmt.Sprintf(" Composite:      %.2f\n", m.CompositeScore))
	if peErr != nil {
		sb.WriteString(fmt.Sprintf(" Nifty PE:       unavailable (%v)\n", peErr))
	} else {
		effectiveDMA := allocator.EffectiveDMADistance(m.DMADistance, niftyPE)
		sb.WriteString(fmt.Sprintf(" Nifty PE:       %.2f  (fear boost %+.1f%% → effective DMA %.1f%%)\n",
			niftyPE, allocator.PEFearBoost(niftyPE), effectiveDMA))
	}
	sb.WriteString(fmt.Sprintf(" Regime:         %s\n", regime))
	return sb.String()
}

func dmaLabel(distance float64) string {
	switch {
	case distance >= 10:
		return "deep fear"
	case distance > 0:
		return "fear"
	case distance >= -5:
		return "neutral"
	default:
		return "greed"
	}
}
