# Architecture

## Overview

Zero-Ops-Bot is a Discord bot that bridges Discord commands/messages to n8n workflows via webhooks.

```
Discord User → Bot → n8n Webhook → n8n Workflow → (SSH/API/etc)
                                              ↓
Discord User ← Bot ← Webhook Response ←──────┘
```

## Design Decisions

### 1. Language: Go

**Decision**: Use Go instead of Node.js/TypeScript

**Rationale**:
- Single binary deployment (no node_modules)
- Lower memory footprint for home lab
- Strong typing with simpler toolchain
- discordgo is mature and well-maintained

### 2. SSH Handling: n8n (not bot)

**Decision**: Bot triggers webhooks, n8n executes SSH commands

**Rationale**:
- SSH credentials isolated in n8n credentials store
- Bot remains stateless and simple
- n8n provides visual debugging and logging
- If bot is compromised, attacker doesn't get SSH access
- Easy to add approval workflows in n8n

**Alternative Considered**: Bot connects to SSH directly
- Rejected: Higher security risk, more complex bot code

### 3. Command Registration: Guild-scoped

**Decision**: Register commands to specific guild, not globally

**Rationale**:
- Instant command updates (global takes up to 1 hour)
- Suitable for personal/home lab use
- Can easily switch to global for multi-server deployment

### 4. Message Forwarding: Optional

**Decision**: Forward chat messages to n8n for AI/processing

**Rationale**:
- Enables conversational AI via n8n
- Channel filtering prevents spam
- Non-blocking (fire-and-forget for most messages)

## Package Structure

```
internal/
├── config/      Configuration loading, validation
├── services/    External service clients
│   └── n8n.go   Webhook HTTP client
├── commands/    Slash command definitions
│   ├── commands.go   Registry and interface
│   └── health.go     /check-health implementation
├── handlers/    Discord event handlers
│   ├── interaction.go   Slash command routing
│   └── message.go       Message forwarding
└── bot/         Discord session management
    └── bot.go   Client setup, lifecycle
```

## Data Flow

### Slash Command Flow

1. User types `/check-health` in Discord
2. Discord sends interaction to bot
3. `handlers/interaction.go` routes to command handler
4. Command calls `services/n8n.TriggerWebhook()`
5. n8n workflow executes, returns result
6. Bot replies to interaction with result

### Message Flow

1. User sends message in allowed channel
2. `handlers/message.go` checks filters
3. Forwards to n8n webhook (non-blocking)
4. n8n can optionally reply via Discord API

## Security Considerations

1. **Token Storage**: Use environment variables, never commit
2. **Webhook Secret**: Optional header for n8n authentication
3. **Channel Filtering**: Limit message forwarding scope
4. **No SSH in Bot**: Credentials stay in n8n
