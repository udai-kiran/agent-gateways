package core

import "time"

// IncomingMessage represents a normalized message from any platform.
type IncomingMessage struct {
	Platform       string    `json:"platform"`
	SenderID       string    `json:"sender_id"`
	ConversationID string    `json:"conversation_id"`
	Text           string    `json:"text"`
	Timestamp      time.Time `json:"timestamp"`
}

// AgentResponse is what the agent returns.
type AgentResponse struct {
	Text string `json:"text"`
}
