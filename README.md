# Agent Gateways

Platform adapters that connect AI agents to messaging platforms. Each gateway receives messages from a platform (Telegram, WhatsApp, Twitter/X), forwards them to your AI agent via HTTP, and sends the agent's response back to the user.

```
User ──► Platform ──► Gateway ──► Agent Endpoint
                        │                │
                        ◄────────────────┘
                        │
User ◄── Platform ◄────┘
```

## Prerequisites

- Go 1.22+
- Docker & Docker Compose (for containerized deployment)

## Project Structure

```
cmd/
  telegram-gateway/    # Telegram bot entrypoint
  whatsapp-gateway/    # WhatsApp Cloud API entrypoint
  twitter-gateway/     # Twitter/X DM entrypoint
internal/
  core/                # Shared types, interfaces, HTTP agent client
  telegram/            # Telegram gateway implementation
  whatsapp/            # WhatsApp gateway implementation
  twitter/             # Twitter/X gateway implementation
docker/                # Dockerfiles for each gateway
config/                # Example environment files
```

## Setup

### 1. Agent Endpoint

All gateways forward messages to an HTTP endpoint you provide via `AGENT_ENDPOINT`. Your agent should accept a POST with this JSON body:

```json
{
  "platform": "telegram",
  "sender_id": "12345",
  "conversation_id": "12345",
  "text": "Hello",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

And respond with:

```json
{
  "text": "Hi there!"
}
```

### 2. Platform API Keys

**Telegram:**
1. Message [@BotFather](https://t.me/BotFather) on Telegram
2. Create a new bot with `/newbot`
3. Copy the bot token to `TELEGRAM_BOT_TOKEN`
4. Set up a webhook pointing to `https://your-domain/webhook`

**WhatsApp:**
1. Create an app at [Meta for Developers](https://developers.facebook.com/)
2. Add the WhatsApp product to your app
3. Get your access token, verify token, and phone number ID
4. Configure the webhook URL to `https://your-domain/webhook`

**Twitter/X:**
1. Create a project at [Twitter Developer Portal](https://developer.twitter.com/)
2. Enable the Account Activity API
3. Get your API key, API secret, access token, and access secret
4. Set up a dev environment and note the environment name

### 3. Environment Variables

Copy the example env file and fill in your values:

```bash
cp config/.env.example .env
```

| Variable | Required By | Description |
|---|---|---|
| `AGENT_ENDPOINT` | All | URL of your AI agent endpoint |
| `AGENT_TIMEOUT` | All | Request timeout (default: `30s`) |
| `TELEGRAM_BOT_TOKEN` | Telegram | Bot token from BotFather |
| `WEBHOOK_PORT` | All | HTTP port (default: `8080`) |
| `WHATSAPP_ACCESS_TOKEN` | WhatsApp | Meta API access token |
| `WHATSAPP_VERIFY_TOKEN` | WhatsApp | Webhook verification token |
| `WHATSAPP_PHONE_NUMBER_ID` | WhatsApp | Your WhatsApp phone number ID |
| `TWITTER_API_KEY` | Twitter | Consumer/API key |
| `TWITTER_API_SECRET` | Twitter | Consumer/API secret |
| `TWITTER_ACCESS_TOKEN` | Twitter | OAuth access token |
| `TWITTER_ACCESS_SECRET` | Twitter | OAuth access secret |
| `TWITTER_ENV_NAME` | Twitter | Account Activity API dev environment |

## Running

### With Docker Compose (all gateways)

```bash
cp config/.env.example .env
# Edit .env with your values
docker compose up --build
```

This starts all three gateways on ports 8081 (Telegram), 8082 (WhatsApp), and 8083 (Twitter).

### Individual Gateway

```bash
# Build
go build -o telegram-gw ./cmd/telegram-gateway

# Run
export AGENT_ENDPOINT=http://localhost:9000/process
export TELEGRAM_BOT_TOKEN=your-token
./telegram-gw
```

Replace `telegram` with `whatsapp` or `twitter` for the other gateways.

### Docker (single gateway)

```bash
docker build -f docker/Dockerfile.telegram -t telegram-gateway .
docker run -p 8080:8080 --env-file .env telegram-gateway
```

## License

See [LICENSE](LICENSE).
