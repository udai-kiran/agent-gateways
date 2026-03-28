package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// HTTPAgentClient implements AgentClient by POSTing to an HTTP endpoint.
type HTTPAgentClient struct {
	endpoint   string
	httpClient *http.Client
}

// NewHTTPAgentClient creates a new HTTP-based agent client.
// It reads AGENT_ENDPOINT and AGENT_TIMEOUT from environment variables.
func NewHTTPAgentClient() (*HTTPAgentClient, error) {
	endpoint := os.Getenv("AGENT_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("AGENT_ENDPOINT is required")
	}

	timeout := 30 * time.Second
	if t := os.Getenv("AGENT_TIMEOUT"); t != "" {
		d, err := time.ParseDuration(t)
		if err != nil {
			return nil, fmt.Errorf("invalid AGENT_TIMEOUT %q: %w", t, err)
		}
		timeout = d
	}

	return &HTTPAgentClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Process sends an IncomingMessage to the agent and returns the response.
func (c *HTTPAgentClient) Process(ctx context.Context, msg IncomingMessage) (AgentResponse, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return AgentResponse{}, fmt.Errorf("marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return AgentResponse{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return AgentResponse{}, fmt.Errorf("agent request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return AgentResponse{}, fmt.Errorf("agent returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var agentResp AgentResponse
	if err := json.NewDecoder(resp.Body).Decode(&agentResp); err != nil {
		return AgentResponse{}, fmt.Errorf("decode agent response: %w", err)
	}

	return agentResp, nil
}
