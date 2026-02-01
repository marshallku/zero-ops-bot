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
	next := nextAlignedTick(time.Now(), h.interval)
	log.Printf("Heartbeat started (interval: %s, channel: %s, next: %s)", h.interval, h.channelID, next.Format(time.RFC3339))

	timer := time.NewTimer(time.Until(next))
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Heartbeat stopped")
			return
		case <-timer.C:
			h.beat(ctx)
			next = nextAlignedTick(time.Now(), h.interval)
			timer.Reset(time.Until(next))
		}
	}
}

func nextAlignedTick(now time.Time, interval time.Duration) time.Time {
	intervalSec := int64(interval.Seconds())
	nowUnix := now.Unix()
	next := nowUnix + intervalSec - (nowUnix % intervalSec)
	return time.Unix(next, 0)
}

func (h *Heartbeat) beat(ctx context.Context) {
	meta := metadata.Get()
	repos := make([]services.RepoMeta, len(meta.Repos))
	for i, r := range meta.Repos {
		repos[i] = services.RepoMeta{
			Name:        r.Name,
			Description: r.Description,
			Path:        r.Path,
		}
	}

	beatCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	result, err := h.n8nClient.TriggerWebhook(beatCtx, services.WebhookPayload{
		Type:      "command",
		Command:   "heartbeat",
		ChannelID: h.channelID,
		Content:   meta.HeartbeatPrompt,
		Repos:     repos,
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
