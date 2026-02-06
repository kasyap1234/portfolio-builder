package main

import (
	"log"
	"time"

	"smart-alert/configs"
	"smart-alert/internal/domain/models"
	"smart-alert/internal/domain/service/allocator"
	service "smart-alert/internal/domain/service/data_fetcher"
	"smart-alert/internal/scheduler"
	"smart-alert/pkg/telegram"
)

func main() {
	log.Println("SIP Allocator Job starting...")

	// Load configuration
	cfg := configs.LoadConfig()

	// Validate required config
	if cfg.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}
	if cfg.TelegramChatID == 0 {
		log.Fatal("TELEGRAM_CHAT_ID environment variable is required")
	}

	// Check if today is first trading day of month
	ist, _ := time.LoadLocation("Asia/Kolkata")
	now := time.Now().In(ist)

	if !scheduler.IsFirstTradingDayOfMonth(now) {
		log.Printf("Today (%s) is not the first trading day of the month. Exiting.", now.Format("2006-01-02"))
		return
	}

	log.Println("First trading day of month detected - running allocation")

	// Create allocator with config
	fetcher := service.NewUniversalFetcher()
	alloc := allocator.NewAllocator(fetcher, allocator.WithPETriggerThreshold(cfg.PETriggerThreshold))

	// Get allocation recommendations
	recs, err := alloc.Allocate(
		cfg.MonthlySIPAmount,
		nil,
		cfg.DebtReserve,
		nil,
	)
	if err != nil {
		log.Fatalf("Allocation failed: %v", err)
	}

	// Determine regime from the same metrics used during allocation
	niftyMetrics, err := alloc.CalculateDropMetrics("^NSEI")
	regime := "UNKNOWN"
	if err != nil {
		log.Printf("Warning: failed to fetch Nifty metrics for display: %v", err)
	} else {
		regime = string(alloc.DetermineMarketRegime(niftyMetrics))
	}

	// Fetch Nifty PE for notification
	niftyPE, err := fetcher.FetchPE("^NSEI")
	if err != nil {
		log.Printf("Warning: failed to fetch Nifty PE: %v", err)
	}

	// Format message
	message := telegram.FormatRecommendations(recs, regime, cfg.MonthlySIPAmount, niftyPE)

	// Send to Telegram
	notifier, err := telegram.NewNotifier(cfg.TelegramBotToken, cfg.TelegramChatID)
	if err != nil {
		log.Fatalf("Failed to create Telegram notifier: %v", err)
	}

	if err := notifier.SendMessage(message); err != nil {
		log.Fatalf("Failed to send Telegram message: %v", err)
	}

	log.Println("Monthly allocation sent to Telegram successfully")
	if niftyPE > 0 {
		log.Printf("Current Nifty PE: %.2f", niftyPE)
	}
	printSummary(recs, cfg.MonthlySIPAmount)
}

func printSummary(recs []models.AllocationRecommendation, sipAmount float64) {
	log.Println("=== Allocation Summary ===")
	var sipTotal, reserveTotal float64
	for _, r := range recs {
		log.Printf("  %s (%s): ₹%.0f", r.AssetSymbol, r.AssetType, r.Amount)
		if r.AssetType == models.AssetTypeDebt || r.AssetType == models.AssetTypeMF || r.AssetType == models.AssetTypeStock {
			// Exclude panic buy / PE trigger (reserve deployments) from SIP total
			if len(r.Reason) >= 9 && (r.Reason[:9] == "PANIC BUY" || r.Reason[:10] == "PE_TRIGGER") {
				reserveTotal += r.Amount
			} else {
				sipTotal += r.Amount
			}
		} else if r.AssetType == models.AssetTypeETF {
			reserveTotal += r.Amount
		}
	}
	log.Printf("SIP Total: ₹%.0f (budget: ₹%.0f)", sipTotal, sipAmount)
	if reserveTotal > 0 {
		log.Printf("Debt Reserve Deployment: ₹%.0f", reserveTotal)
	}
}
