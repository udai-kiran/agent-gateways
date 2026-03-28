package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/udai-kiran/agent-gateways/gateways/internal/core"
)

// Update represents a Telegram Bot API Update object.
type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message"`
}

// Message represents a Telegram message.
type Message struct {
	MessageID int    `json:"message_id"`
	From      *User  `json:"from"`
	Chat      Chat   `json:"chat"`
	Date      int64  `json:"date"`
	Text      string `json:"text"`
}

// User represents a Telegram user.
type User struct {
	ID int64 `json:"id"`
}

// Chat represents a Telegram chat.
type Chat struct {
	ID int64 `json:"id"`
}

// sendMessageRequest is the payload for Telegram's sendMessage API.
type sendMessageRequest struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

// Gateway implements core.Gateway for Telegram.
type Gateway struct {
	config Config
	agent  core.AgentClient
	server *http.Server
}

// NewGateway creates a new Telegram gateway.
func NewGateway(cfg Config, agent core.AgentClient) *Gateway {
	return &Gateway{
		config: cfg,
		agent:  agent,
	}
}

// Start begins listening for Telegram webhook updates.
func (g *Gateway) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", g.handleWebhook)

	g.server = &http.Server{
		Addr:    ":" + g.config.WebhookPort,
		Handler: mux,
	}

	log.Printf("Telegram gateway listening on :%s", g.config.WebhookPort)
	if err := g.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Stop gracefully shuts down the server.
func (g *Gateway) Stop(ctx context.Context) error {
	log.Println("Shutting down Telegram gateway...")
	return g.server.Shutdown(ctx)
}

func (g *Gateway) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("Failed to decode update: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if update.Message == nil || update.Message.Text == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	msg := core.IncomingMessage{
		Platform:       "telegram",
		SenderID:       strconv.FormatInt(update.Message.From.ID, 10),
		ConversationID: strconv.FormatInt(update.Message.Chat.ID, 10),
		Text:           update.Message.Text,
		Timestamp:      time.Unix(update.Message.Date, 0),
	}

	resp, err := g.agent.Process(r.Context(), msg)
	if err != nil {
		log.Printf("Agent error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := g.sendMessage(update.Message.Chat.ID, resp.Text); err != nil {
		log.Printf("Failed to send reply: %v", err)
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Gateway) sendMessage(chatID int64, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", g.config.BotToken)

	payload, err := json.Marshal(sendMessageRequest{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		return fmt.Errorf("marshal send request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	return nil
}
