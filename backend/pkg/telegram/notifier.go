package telegram

import (
	"fmt"

	"smart-alert/internal/domain/models"
	"smart-alert/pkg/allocationreport"

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

// FormatRecommendations converts allocation recommendations to a formatted Telegram message.
func FormatRecommendations(recs []models.AllocationRecommendation, opts allocationreport.Options) string {
	return allocationreport.FormatTelegram(recs, opts)
}
