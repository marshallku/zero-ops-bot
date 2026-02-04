# Zero-Ops-Bot

Discord bot for home lab maintenance automation and personal secretary via n8n webhooks.

## Quick Start

```bash
# Set environment variables
cp .env.example .env
# Edit .env with your values

# Copy and edit metadata config (schedules, prompts, repos)
cp metadata.example.yaml metadata.yaml

# Run
go run ./cmd/bot

# Build
go build -o bot ./cmd/bot
```

## Documentation

- [docs/architecture.md](docs/architecture.md) - Design decisions and structure
- [docs/working-guide.md](docs/working-guide.md) - Development workflow and profiles

## Architecture

### Key Decisions

1. **Go over Node.js** - Lower memory footprint, single binary deployment
2. **n8n handles SSH** - Bot only triggers webhooks, SSH credentials stay in n8n
3. **Global commands by default** - Guild ID optional, omit for global registration
4. **Schedules are data** - Defined in metadata.yaml, not code. Adding a schedule = YAML entry + n8n workflow
5. **Notes as markdown** - Human-readable files, editable outside the bot

### Project Structure

```
cmd/bot/          Entry point
internal/
  config/         Environment configuration
  services/       External service clients (n8n)
  commands/       Slash command definitions (/repo, /note, /schedule)
  handlers/       Discord event handlers
  scheduler/      Cron-based scheduler (replaces heartbeat)
  notes/          Markdown-based note storage
  metadata/       YAML config (repos, schedules, prompts)
  bot/            Discord client setup
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| DISCORD_TOKEN | Yes | Bot token from Discord Developer Portal |
| DISCORD_APP_ID | Yes | Application ID |
| DISCORD_GUILD_ID | No | Guild ID for instant command updates (omit for global) |
| N8N_WEBHOOK_URL | Yes | n8n webhook endpoint |
| N8N_WEBHOOK_SECRET | No | Webhook auth (sent as `x-discord-api-key` header) |
| ALLOWED_CHANNELS | No | Comma-separated channel IDs for message forwarding |
| TZ | No | Timezone for scheduler (default: `Local`, e.g. `Asia/Seoul`) |
| NOTES_DIR | No | Directory for markdown notes (default: `./notes`) |

## Schedules

Schedules are defined in `metadata.yaml`. Each schedule triggers an n8n webhook on a cron expression:

```yaml
schedules:
  - name: morning-briefing
    cron: "30 7 * * 1-5"
    channel_id: "123456"
    command: briefing
    include_notes: true
    include_repos: true
    prompt: "Generate a daily briefing..."
```

Manage via `/schedule list|add|remove` or edit `metadata.yaml` directly.

## Notes

Notes stored as markdown in `NOTES_DIR`:
- `daily/YYYY-MM-DD.md` - Daily notes
- `categories/{name}.md` - Category-specific notes

Manage via `/note add|today|list|remove|search` or mention the bot with "remember ...".
