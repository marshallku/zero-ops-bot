package heartbeat

import (
    "context"
    "log"
    "time"

    "github.com/bwmarrin/discordgo"
    "github.com/marshall/zero-ops-bot/internal/metadata"
    "github.com/marshall/zero-ops-bot/internal/services"
    "github.com/marshall/zero-ops-bot/internal/utils"
)

type Heartbeat struct {
    session   *discordgo.Session
    n8nClient *services.N8nClient
    channelID string
    interval  time.Duration
}

func New(session *discordgo.Session, n8nClient *services.N8nClient, channelID string, interval time.Duration) *Heartbeat {
    return &Heartbeat{
        session:   session,
        n8nClient: n8nClient,
        channelID: channelID,
        interval:  interval,
    }
}

func (h *Heartbeat) Start(ctx context.Context) {
    ticker := time.NewTicker(h.interval)
    defer ticker.Stop()

    log.Printf("Heartbeat started (interval: %s, channel: %s)", h.interval, h.channelID)

    for {
        select {
        case <-ctx.Done():
            log.Println("Heartbeat stopped")
            return
        case <-ticker.C:
            h.beat(ctx)
        }
    }
}

func (h *Heartbeat) beat(ctx context.Context) {
    meta := metadata.Get()
    repos := make([]services.RepoMeta, len(meta.Repos))
    for i, r := range meta.Repos {
        repos[i] = services.RepoMeta{
            Name:        r.Name,
            Description: r.Description,
            Command:     r.Command,
        }
    }

    beatCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    result, err := h.n8nClient.TriggerWebhook(beatCtx, services.WebhookPayload{
        Type:         "heartbeat",
        ChannelID:    h.channelID,
        SystemPrompt: meta.SystemPrompt,
        Repos:        repos,
    })
    if err != nil {
        log.Printf("Heartbeat webhook failed: %v", err)
        return
    }

    if result.Message == "" {
        return
    }

    for _, chunk := range utils.SplitMessage(result.Message) {
        if _, err := h.session.ChannelMessageSend(h.channelID, chunk); err != nil {
            log.Printf("Heartbeat message send failed: %v", err)
            return
        }
    }
}
