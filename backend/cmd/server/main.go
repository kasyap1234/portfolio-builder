package main

import (
	"log"
	"time"

	"smart-alert/configs"
	"smart-alert/internal/domain/service/allocator"
	service "smart-alert/internal/domain/service/data_fetcher"
	"smart-alert/internal/scheduler"
	"smart-alert/pkg/allocationreport"
	"smart-alert/pkg/telegram"
)

func main() {
	log.Println("SIP Allocator Job starting...")

	cfg := configs.LoadConfig()

	if cfg.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}
	if cfg.TelegramChatID == 0 {
		log.Fatal("TELEGRAM_CHAT_ID environment variable is required")
	}

	ist, _ := time.LoadLocation("Asia/Kolkata")
	now := time.Now().In(ist)

	if !scheduler.IsFirstTradingDayOfMonth(now) {
		log.Printf("Today (%s) is not the first trading day of the month. Exiting.", now.Format("2006-01-02"))
		return
	}

	log.Println("First trading day of month detected - running allocation")

	fetcher := service.NewUniversalFetcher()
	alloc := allocator.NewAllocator(fetcher, allocator.WithPETriggerThreshold(cfg.PETriggerThreshold))

	recs, err := alloc.Allocate(
		cfg.MonthlySIPAmount,
		nil,
		cfg.DebtReserve,
		nil,
	)
	if err != nil {
		log.Fatalf("Allocation failed: %v", err)
	}

	niftyMetrics, err := alloc.CalculateDropMetrics("^NSEI")
	niftyPE, peErr := fetcher.FetchPE("^NSEI")
	regime := "UNKNOWN"
	if err != nil {
		log.Printf("Warning: failed to fetch Nifty metrics for display: %v", err)
	} else {
		regime = string(alloc.DetermineMarketRegime(niftyMetrics, niftyPE))
	}
	if peErr != nil {
		log.Printf("Warning: failed to fetch Nifty PE: %v", peErr)
	}

	opts := allocationreport.Options{
		SIPAmount:   cfg.MonthlySIPAmount,
		DebtReserve: cfg.DebtReserve,
		Regime:      regime,
	}
	if peErr == nil {
		opts.NiftyPE = niftyPE
	}

	message := telegram.FormatRecommendations(recs, opts)

	notifier, err := telegram.NewNotifier(cfg.TelegramBotToken, cfg.TelegramChatID)
	if err != nil {
		log.Fatalf("Failed to create Telegram notifier: %v", err)
	}

	if err := notifier.SendMessage(message); err != nil {
		log.Fatalf("Failed to send Telegram message: %v", err)
	}

	log.Println("Monthly allocation sent to Telegram successfully")
	log.Print(allocationreport.FormatConsole(recs, opts))
}
