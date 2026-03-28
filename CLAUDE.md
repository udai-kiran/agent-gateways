# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

### Gateways (Go)

```bash
go build ./...          # Build all binaries
go vet ./...            # Lint
go test ./...           # Run all tests
docker compose -f gateways/docker-compose.yml config      # Validate docker-compose
docker compose -f gateways/docker-compose.yml up --build   # Run all gateways (ports 8081/8082/8083)
```

Individual gateway: `go build -o gw ./gateways/cmd/telegram-gateway` (or whatsapp-gateway, twitter-gateway).

Docker single build: `docker build -f gateways/docker/Dockerfile.telegram .`

### Agents (Python)

```bash
cd agents/simple-agent
cp .env.example .env    # Fill in OPENAI_API_KEY
uv sync                 # Install dependencies
uv run uvicorn main:app --host 0.0.0.0 --port 9000   # Run agent
docker build -t simple-agent .                         # Docker build
```

## Architecture

Monorepo with two top-level directories: `gateways/` (Go platform adapters) and `agents/` (Python AI agents).

### Gateways

Three independent gateway binaries share a common core. Each gateway receives messages from a messaging platform via webhook, normalizes them into `core.IncomingMessage`, forwards to an AI agent via HTTP POST to `AGENT_ENDPOINT`, and sends the `core.AgentResponse` back to the platform.

**Core layer** (`gateways/internal/core/`): Defines `IncomingMessage`/`AgentResponse` types, `Gateway` interface (Start/Stop), `AgentClient` interface (Process), and `HTTPAgentClient` which POSTs JSON to a configurable `AGENT_ENDPOINT`.

**Gateway layer** (`gateways/internal/{telegram,whatsapp,twitter}/`): Each package implements `core.Gateway` with platform-specific webhook handling and reply sending. Each has a `config.go` (reads env vars) and `gateway.go` (HTTP server + handlers). Twitter additionally has `auth.go` for OAuth 1.0a signing and CRC validation.

**Entrypoints** (`gateways/cmd/{name}-gateway/main.go`): All follow the same pattern — create agent client, load config, create gateway, start with signal-based graceful shutdown.

### Agents

Python FastAPI apps that receive `IncomingMessage` via POST `/process` and return `AgentResponse`. Each agent is a standalone service with its own `pyproject.toml` and virtual environment managed by `uv`.

**simple-agent** (`agents/simple-agent/`): Strands-based agent using OpenAI. Key files:
- `models.py` — Pydantic models (`Settings`, `IncomingMessage`, `AgentResponse`) with `pydantic-settings` for `.env` loading.
- `agent.py` — Strands agent creation and execution (isolated from web framework).
- `history.py` — `ConversationHistory` ABC with `RedisConversationHistory` implementation for per-conversation state with TTL.
- `main.py` — FastAPI app wiring Redis, agent, and HTTP endpoint.
- `logging_config.py` — structlog setup with JSON output.

## Key Patterns

- Gateway config via environment variables (`os.Getenv`), no config libraries. Each platform's `LoadConfig()` validates required vars.
- Agent config via `pydantic-settings` loading from `.env` files and environment variables.
- All gateways listen on `WEBHOOK_PORT` (default 8080). In docker-compose, host ports are mapped to 8081/8082/8083.
- Agents listen on port 9000 by default. Gateways connect via `AGENT_ENDPOINT` (e.g., `http://localhost:9000/process`).
- No external Go dependencies — stdlib only.
- Webhook endpoints are always at `/webhook` (GET for verification/CRC, POST for events).
- Agent conversation history stored in Redis with TTL-based expiration, enabling horizontal scaling.
- Blocking Strands agent calls run via `asyncio.to_thread` to keep the FastAPI event loop free.
