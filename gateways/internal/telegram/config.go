package telegram

import (
	"fmt"
	"os"
)

type Config struct {
	BotToken    string
	WebhookPort string
}

func LoadConfig() (Config, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return Config{}, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	port := os.Getenv("WEBHOOK_PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		BotToken:    token,
		WebhookPort: port,
	}, nil
}
