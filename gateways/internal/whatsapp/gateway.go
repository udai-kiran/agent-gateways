package whatsapp

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

// Webhook payload types for Meta Cloud API.
type webhookPayload struct {
	Object string  `json:"object"`
	Entry  []entry `json:"entry"`
}

type entry struct {
	ID      string   `json:"id"`
	Changes []change `json:"changes"`
}

type change struct {
	Value changeValue `json:"value"`
	Field string      `json:"field"`
}

type changeValue struct {
	MessagingProduct string           `json:"messaging_product"`
	Metadata         metadata         `json:"metadata"`
	Messages         []incomingMsg    `json:"messages"`
	Contacts         []contactInfo    `json:"contacts"`
}

type metadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type incomingMsg struct {
	From      string   `json:"from"`
	ID        string   `json:"id"`
	Timestamp string   `json:"timestamp"`
	Text      *textObj `json:"text"`
	Type      string   `json:"type"`
}

type textObj struct {
	Body string `json:"body"`
}

type contactInfo struct {
	Profile struct {
		Name string `json:"name"`
	} `json:"profile"`
	WaID string `json:"wa_id"`
}

// sendMessagePayload is the request body for WhatsApp Cloud API.
type sendMessagePayload struct {
	MessagingProduct string      `json:"messaging_product"`
	To               string      `json:"to"`
	Type             string      `json:"type"`
	Text             sendTextObj `json:"text"`
}

type sendTextObj struct {
	Body string `json:"body"`
}

// Gateway implements core.Gateway for WhatsApp.
type Gateway struct {
	config Config
	agent  core.AgentClient
	server *http.Server
}

// NewGateway creates a new WhatsApp gateway.
func NewGateway(cfg Config, agent core.AgentClient) *Gateway {
	return &Gateway{
		config: cfg,
		agent:  agent,
	}
}

// Start begins listening for WhatsApp webhook events.
func (g *Gateway) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", g.handleWebhook)

	g.server = &http.Server{
		Addr:    ":" + g.config.WebhookPort,
		Handler: mux,
	}

	log.Printf("WhatsApp gateway listening on :%s", g.config.WebhookPort)
	if err := g.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Stop gracefully shuts down the server.
func (g *Gateway) Stop(ctx context.Context) error {
	log.Println("Shutting down WhatsApp gateway...")
	return g.server.Shutdown(ctx)
}

func (g *Gateway) handleWebhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		g.handleVerification(w, r)
	case http.MethodPost:
		g.handleMessage(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (g *Gateway) handleVerification(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	if mode == "subscribe" && token == g.config.VerifyToken {
		log.Println("Webhook verified")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, challenge)
		return
	}

	http.Error(w, "forbidden", http.StatusForbidden)
}

func (g *Gateway) handleMessage(w http.ResponseWriter, r *http.Request) {
	var payload webhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Failed to decode payload: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Always return 200 to Meta immediately.
	w.WriteHeader(http.StatusOK)

	for _, e := range payload.Entry {
		for _, c := range e.Changes {
			for _, m := range c.Value.Messages {
				if m.Type != "text" || m.Text == nil {
					continue
				}

				ts, _ := time.Parse("1136239445", m.Timestamp)

				msg := core.IncomingMessage{
					Platform:       "whatsapp",
					SenderID:       m.From,
					ConversationID: m.From,
					Text:           m.Text.Body,
					Timestamp:      ts,
				}

				resp, err := g.agent.Process(r.Context(), msg)
				if err != nil {
					log.Printf("Agent error: %v", err)
					continue
				}

				if err := g.sendMessage(m.From, resp.Text); err != nil {
					log.Printf("Failed to send reply: %v", err)
				}
			}
		}
	}
}

func (g *Gateway) sendMessage(to, text string) error {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", g.config.PhoneNumberID)

	payload, err := json.Marshal(sendMessagePayload{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             "text",
		Text:             sendTextObj{Body: text},
	})
	if err != nil {
		return fmt.Errorf("marshal send request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.config.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("whatsapp API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
