package whatsapp

import (
	"fmt"
	"os"
)

type Config struct {
	AccessToken   string
	VerifyToken   string
	PhoneNumberID string
	WebhookPort   string
}

func LoadConfig() (Config, error) {
	accessToken := os.Getenv("WHATSAPP_ACCESS_TOKEN")
	if accessToken == "" {
		return Config{}, fmt.Errorf("WHATSAPP_ACCESS_TOKEN is required")
	}

	verifyToken := os.Getenv("WHATSAPP_VERIFY_TOKEN")
	if verifyToken == "" {
		return Config{}, fmt.Errorf("WHATSAPP_VERIFY_TOKEN is required")
	}

	phoneNumberID := os.Getenv("WHATSAPP_PHONE_NUMBER_ID")
	if phoneNumberID == "" {
		return Config{}, fmt.Errorf("WHATSAPP_PHONE_NUMBER_ID is required")
	}

	port := os.Getenv("WEBHOOK_PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		AccessToken:   accessToken,
		VerifyToken:   verifyToken,
		PhoneNumberID: phoneNumberID,
		WebhookPort:   port,
	}, nil
}
