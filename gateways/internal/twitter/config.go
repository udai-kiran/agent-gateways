package twitter

import (
	"fmt"
	"os"
)

type Config struct {
	APIKey       string
	APISecret    string
	AccessToken  string
	AccessSecret string
	EnvName      string
	WebhookPort  string
}

func LoadConfig() (Config, error) {
	apiKey := os.Getenv("TWITTER_API_KEY")
	if apiKey == "" {
		return Config{}, fmt.Errorf("TWITTER_API_KEY is required")
	}

	apiSecret := os.Getenv("TWITTER_API_SECRET")
	if apiSecret == "" {
		return Config{}, fmt.Errorf("TWITTER_API_SECRET is required")
	}

	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	if accessToken == "" {
		return Config{}, fmt.Errorf("TWITTER_ACCESS_TOKEN is required")
	}

	accessSecret := os.Getenv("TWITTER_ACCESS_SECRET")
	if accessSecret == "" {
		return Config{}, fmt.Errorf("TWITTER_ACCESS_SECRET is required")
	}

	envName := os.Getenv("TWITTER_ENV_NAME")
	if envName == "" {
		return Config{}, fmt.Errorf("TWITTER_ENV_NAME is required")
	}

	port := os.Getenv("WEBHOOK_PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		APIKey:       apiKey,
		APISecret:    apiSecret,
		AccessToken:  accessToken,
		AccessSecret: accessSecret,
		EnvName:      envName,
		WebhookPort:  port,
	}, nil
}
