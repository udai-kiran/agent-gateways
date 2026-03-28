# Commands

All commands run from the repo root.

## Build

```bash
# Build all
go build ./...

# Build individual gateways
go build -o telegram-gw ./gateways/cmd/telegram-gateway
go build -o whatsapp-gw ./gateways/cmd/whatsapp-gateway
go build -o twitter-gw ./gateways/cmd/twitter-gateway
```

## Run from Terminal

```bash
# Telegram
export AGENT_ENDPOINT=http://localhost:9000/process
export TELEGRAM_BOT_TOKEN=your-bot-token
./telegram-gw

# WhatsApp
export AGENT_ENDPOINT=http://localhost:9000/process
export WHATSAPP_ACCESS_TOKEN=your-access-token
export WHATSAPP_VERIFY_TOKEN=your-verify-token
export WHATSAPP_PHONE_NUMBER_ID=your-phone-number-id
./whatsapp-gw

# Twitter/X
export AGENT_ENDPOINT=http://localhost:9000/process
export TWITTER_API_KEY=your-api-key
export TWITTER_API_SECRET=your-api-secret
export TWITTER_ACCESS_TOKEN=your-access-token
export TWITTER_ACCESS_SECRET=your-access-secret
export TWITTER_ENV_NAME=your-env-name
./twitter-gw
```

## Docker

```bash
# Build images (from repo root)
docker build -f gateways/docker/Dockerfile.telegram -t telegram-gateway .
docker build -f gateways/docker/Dockerfile.whatsapp -t whatsapp-gateway .
docker build -f gateways/docker/Dockerfile.twitter -t twitter-gateway .

# Run individual container
docker run -p 8080:8080 --env-file .env telegram-gateway
docker run -p 8080:8080 --env-file .env whatsapp-gateway
docker run -p 8080:8080 --env-file .env twitter-gateway

# Run all with docker compose
docker compose -f gateways/docker-compose.yml up --build
```
