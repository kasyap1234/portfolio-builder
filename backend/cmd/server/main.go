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

	// Create allocator
	fetcher := service.NewUniversalFetcher()
	alloc := allocator.NewAllocator(fetcher)

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

	// Get market regime
	niftyMetrics, _ := alloc.CalculateDropMetrics("^NSEI")
	regime := "UNKNOWN"
	if niftyMetrics != nil {
		regime = string(alloc.DetermineMarketRegime(niftyMetrics))
	}

	// Format message
	message := telegram.FormatRecommendations(recs, regime, cfg.MonthlySIPAmount)

	// Send to Telegram
	notifier, err := telegram.NewNotifier(cfg.TelegramBotToken, cfg.TelegramChatID)
	if err != nil {
		log.Fatalf("Failed to create Telegram notifier: %v", err)
	}

	if err := notifier.SendMessage(message); err != nil {
		log.Fatalf("Failed to send Telegram message: %v", err)
	}

	log.Println("Monthly allocation sent to Telegram successfully")
	printSummary(recs, cfg.MonthlySIPAmount)
}

func printSummary(recs []models.AllocationRecommendation, total float64) {
	log.Println("=== Allocation Summary ===")
	for _, r := range recs {
		log.Printf("  %s (%s): ₹%.0f", r.AssetSymbol, r.AssetType, r.Amount)
	}
	log.Printf("Total: ₹%.0f", total)
}
