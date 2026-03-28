package twitter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/udai-kiran/agent-gateways/gateways/internal/core"
)

// Account Activity API webhook event types.
type webhookEvent struct {
	ForUserID           string              `json:"for_user_id"`
	DirectMessageEvents []directMessageEvent `json:"direct_message_events"`
}

type directMessageEvent struct {
	Type             string           `json:"type"`
	ID               string           `json:"id"`
	CreatedTimestamp  string           `json:"created_timestamp"`
	MessageCreate    messageCreate    `json:"message_create"`
}

type messageCreate struct {
	Target      target      `json:"target"`
	SenderID    string      `json:"sender_id"`
	MessageData messageData `json:"message_data"`
}

type target struct {
	RecipientID string `json:"recipient_id"`
}

type messageData struct {
	Text string `json:"text"`
}

// DM API v2 request body.
type dmRequest struct {
	Text string `json:"text"`
}

// Gateway implements core.Gateway for Twitter/X.
type Gateway struct {
	config Config
	agent  core.AgentClient
	server *http.Server
}

// NewGateway creates a new Twitter gateway.
func NewGateway(cfg Config, agent core.AgentClient) *Gateway {
	return &Gateway{
		config: cfg,
		agent:  agent,
	}
}

// Start begins listening for Twitter webhook events.
func (g *Gateway) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", g.handleWebhook)

	g.server = &http.Server{
		Addr:    ":" + g.config.WebhookPort,
		Handler: mux,
	}

	log.Printf("Twitter gateway listening on :%s", g.config.WebhookPort)
	if err := g.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Stop gracefully shuts down the server.
func (g *Gateway) Stop(ctx context.Context) error {
	log.Println("Shutting down Twitter gateway...")
	return g.server.Shutdown(ctx)
}

func (g *Gateway) handleWebhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		g.handleCRC(w, r)
	case http.MethodPost:
		g.handleEvent(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (g *Gateway) handleCRC(w http.ResponseWriter, r *http.Request) {
	crcToken := r.URL.Query().Get("crc_token")
	if crcToken == "" {
		http.Error(w, "missing crc_token", http.StatusBadRequest)
		return
	}

	responseToken := CRCToken(crcToken, g.config.APISecret)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"response_token": responseToken,
	})
}

func (g *Gateway) handleEvent(w http.ResponseWriter, r *http.Request) {
	var event webhookEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Printf("Failed to decode event: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	for _, dm := range event.DirectMessageEvents {
		if dm.Type != "message_create" {
			continue
		}

		// Filter out own messages to prevent echo loop.
		if dm.MessageCreate.SenderID == event.ForUserID {
			continue
		}

		ts, _ := time.Parse("1136239445000", dm.CreatedTimestamp)

		msg := core.IncomingMessage{
			Platform:       "twitter",
			SenderID:       dm.MessageCreate.SenderID,
			ConversationID: dm.MessageCreate.SenderID,
			Text:           dm.MessageCreate.MessageData.Text,
			Timestamp:      ts,
		}

		resp, err := g.agent.Process(r.Context(), msg)
		if err != nil {
			log.Printf("Agent error: %v", err)
			continue
		}

		if err := g.sendDM(dm.MessageCreate.SenderID, resp.Text); err != nil {
			log.Printf("Failed to send DM: %v", err)
		}
	}
}

func (g *Gateway) sendDM(recipientID, text string) error {
	url := fmt.Sprintf("https://api.twitter.com/2/dm_conversations/with/%s/messages", recipientID)

	payload, err := json.Marshal(dmRequest{Text: text})
	if err != nil {
		return fmt.Errorf("marshal DM request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	SignRequest(req, g.config, string(payload))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send DM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("twitter API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
