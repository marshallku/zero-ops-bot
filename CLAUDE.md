# Zero-Ops-Bot

Discord bot for home lab maintenance automation via n8n webhooks.

## Quick Start

```bash
# Set environment variables
cp .env.example .env
# Edit .env with your values

# Run
go run ./cmd/bot

# Build
go build -o bot ./cmd/bot
```

## Architecture

See [docs/architecture.md](docs/architecture.md) for detailed design decisions.

### Key Decisions

1. **Go over Node.js** - Lower memory footprint, single binary deployment
2. **n8n handles SSH** - Bot only triggers webhooks, SSH credentials stay in n8n
3. **Guild-scoped commands** - Instant updates, suitable for home lab

### Project Structure

```
cmd/bot/          Entry point
internal/
  config/         Environment configuration
  services/       External service clients (n8n)
  commands/       Slash command definitions
  handlers/       Discord event handlers
  bot/            Discord client setup
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| DISCORD_TOKEN | Yes | Bot token from Discord Developer Portal |
| DISCORD_APP_ID | Yes | Application ID |
| DISCORD_GUILD_ID | No | Guild ID for guild-scoped commands |
| N8N_WEBHOOK_URL | Yes | n8n webhook endpoint |
| N8N_WEBHOOK_SECRET | No | Optional webhook auth header |
| ALLOWED_CHANNELS | No | Comma-separated channel IDs for message forwarding |
