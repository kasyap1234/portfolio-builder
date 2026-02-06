package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"smart-alert/internal/domain/models"
	"smart-alert/internal/domain/service/allocator"
	service "smart-alert/internal/domain/service/data_fetcher"
	"smart-alert/pkg/telegram"

	"github.com/robfig/cron/v3"
)

// Config holds scheduler configuration
type Config struct {
	MonthlySIPAmount   float64
	DebtReserve        float64 // Current debt reserve available
	TelegramBotToken   string
	TelegramChatID     int64
	PETriggerThreshold float64
}

// Scheduler handles periodic allocation tasks
type Scheduler struct {
	config    Config
	allocator allocator.Allocator
	fetcher   service.DataFetcher
	notifier  *telegram.Notifier
	cron      *cron.Cron
}

// NewScheduler creates a new scheduler instance
func NewScheduler(config Config) (*Scheduler, error) {
	// Create allocator with universal fetcher
	fetcher := service.NewUniversalFetcher()
	var opts []allocator.AllocatorOption
	if config.PETriggerThreshold > 0 {
		opts = append(opts, allocator.WithPETriggerThreshold(config.PETriggerThreshold))
	}
	alloc := allocator.NewAllocator(fetcher, opts...)

	// Create Telegram notifier
	notifier, err := telegram.NewNotifier(config.TelegramBotToken, config.TelegramChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram notifier: %w", err)
	}

	// Create cron with IST timezone
	ist, _ := time.LoadLocation("Asia/Kolkata")
	c := cron.New(cron.WithLocation(ist))

	return &Scheduler{
		config:    config,
		allocator: alloc,
		fetcher:   fetcher,
		notifier:  notifier,
		cron:      c,
	}, nil
}

// Start begins the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	// Schedule daily at 12:00 PM IST
	_, err := s.cron.AddFunc("0 12 * * *", s.dailyCheck)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	log.Println("Scheduler started - running daily at 12:00 PM IST")
	s.cron.Start()

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("Scheduler stopping...")
	s.cron.Stop()
	return nil
}

// dailyCheck runs the daily allocation check
func (s *Scheduler) dailyCheck() {
	ist, _ := time.LoadLocation("Asia/Kolkata")
	now := time.Now().In(ist)

	log.Printf("Daily check triggered at %s", now.Format("2006-01-02 15:04:05"))

	if !IsFirstTradingDayOfMonth(now) {
		log.Println("Not first trading day of month - skipping allocation")
		return
	}

	log.Println("First trading day of month - running allocation")
	s.runMonthlyAllocation()
}

// runMonthlyAllocation executes the monthly SIP allocation
func (s *Scheduler) runMonthlyAllocation() {
	// Get allocation recommendations
	recs, err := s.allocator.Allocate(
		s.config.MonthlySIPAmount,
		nil, // No existing portfolio needed for recommendations
		s.config.DebtReserve,
		nil, // Using default watchlist
	)
	if err != nil {
		log.Printf("Allocation failed: %v", err)
		s.notifier.SendMessage(fmt.Sprintf("❌ Allocation failed: %v", err))
		return
	}

	// Get market regime
	niftyMetrics, err := s.allocator.CalculateDropMetrics("^NSEI")
	regime := "UNKNOWN"
	if err == nil {
		regime = string(s.allocator.DetermineMarketRegime(niftyMetrics))
	}

	// Get Nifty PE for notification
	niftyPE, _ := s.fetcher.FetchPE("^NSEI")

	// Format and send message
	message := telegram.FormatRecommendations(recs, regime, s.config.MonthlySIPAmount, niftyPE)
	if err := s.notifier.SendMessage(message); err != nil {
		log.Printf("Failed to send Telegram message: %v", err)
		return
	}

	log.Println("Monthly allocation sent to Telegram successfully")
}

// RunNow triggers an immediate allocation (for testing)
func (s *Scheduler) RunNow() {
	s.runMonthlyAllocation()
}

// GetNextRun returns the next scheduled run time
func (s *Scheduler) GetNextRun() time.Time {
	entries := s.cron.Entries()
	if len(entries) > 0 {
		return entries[0].Next
	}
	return time.Time{}
}

// Helper to calculate total by type
func sumByType(recs []models.AllocationRecommendation, assetType models.AssetType) float64 {
	var total float64
	for _, r := range recs {
		if r.AssetType == assetType {
			total += r.Amount
		}
	}
	return total
}
