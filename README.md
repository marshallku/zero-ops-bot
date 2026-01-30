# Zero-Ops-Bot

Discord bot for home lab maintenance automation. Triggers n8n workflows via webhooks.

## Features

- `/check-health` - Trigger health check workflow
- `/infra` - Trigger infrastructure workflow
- Message forwarding to n8n for AI/automation
- Proactive heartbeat messaging — bot periodically checks in with n8n and posts to a channel
- Extensible command system

## Setup

1. Create a Discord application at [Discord Developer Portal](https://discord.com/developers/applications)
2. Enable "Message Content Intent" in Bot settings
3. Invite bot to your server with `applications.commands` and `bot` scopes

```bash
cp .env.example .env
# Edit .env with your values

go run ./cmd/bot
```

## Docker

```bash
docker build -t zero-ops-bot .
docker run --env-file .env zero-ops-bot
```

## Documentation

- [Architecture](docs/architecture.md) - Design decisions and structure
- [Working Guide](docs/working-guide.md) - Development workflow and profiles
- [CLAUDE.md](CLAUDE.md) - Quick reference for development
