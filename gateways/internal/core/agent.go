package core

import "context"

// AgentClient sends messages to the AI agent and gets responses.
type AgentClient interface {
	Process(ctx context.Context, msg IncomingMessage) (AgentResponse, error)
}
