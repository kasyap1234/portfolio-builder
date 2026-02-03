package telegram

import (
	"fmt"
	"strings"

	"smart-alert/internal/domain/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Notifier handles Telegram message sending
type Notifier struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

// NewNotifier creates a new Telegram notifier
func NewNotifier(botToken string, chatID int64) (*Notifier, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}
	return &Notifier{bot: bot, chatID: chatID}, nil
}

// SendMessage sends a text message to the configured chat
func (n *Notifier) SendMessage(text string) error {
	msg := tgbotapi.NewMessage(n.chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := n.bot.Send(msg)
	return err
}

// FormatRecommendations converts allocation recommendations to a formatted message
func FormatRecommendations(recs []models.AllocationRecommendation, regime string, totalAmount float64) string {
	var sb strings.Builder

	sb.WriteString("📊 *Monthly SIP Allocation*\n")
	sb.WriteString(fmt.Sprintf("💰 Amount: ₹%.0f\n", totalAmount))
	sb.WriteString(fmt.Sprintf("📈 Market Regime: *%s*\n\n", regime))

	// Group by type
	var equityRecs, debtRecs, panicRecs []models.AllocationRecommendation
	for _, r := range recs {
		if strings.HasPrefix(r.Reason, "PANIC BUY") {
			panicRecs = append(panicRecs, r)
		} else if r.AssetType == models.AssetTypeDebt {
			debtRecs = append(debtRecs, r)
		} else {
			equityRecs = append(equityRecs, r)
		}
	}

	// Panic buys (if any)
	if len(panicRecs) > 0 {
		sb.WriteString("🚨 *PANIC BUY DEPLOYMENT*\n")
		for _, r := range panicRecs {
			sb.WriteString(fmt.Sprintf("  • %s: ₹%.0f\n", r.AssetSymbol, r.Amount))
		}
		sb.WriteString("\n")
	}

	// Equity allocations
	if len(equityRecs) > 0 {
		sb.WriteString("📈 *Equity*\n")
		for _, r := range equityRecs {
			sb.WriteString(fmt.Sprintf("  • %s: ₹%.0f\n", r.AssetSymbol, r.Amount))
		}
		sb.WriteString("\n")
	}

	// Debt allocations
	if len(debtRecs) > 0 {
		sb.WriteString("🏦 *Debt*\n")
		for _, r := range debtRecs {
			sb.WriteString(fmt.Sprintf("  • %s: ₹%.0f\n", r.AssetSymbol, r.Amount))
		}
	}

	return sb.String()
}
