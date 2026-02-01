package handlers

import (
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/marshall/zero-ops-bot/internal/services"
)

func NewMessageHandler(n8n *services.N8nClient, allowedChannels []string) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore bot messages
		if m.Author.Bot {
			return
		}

		// Skip if bot is mentioned (handled by mention handler)
		for _, mention := range m.Mentions {
			if mention.ID == s.State.User.ID {
				return
			}
		}

		// Check channel filter
		if len(allowedChannels) > 0 && !slices.Contains(allowedChannels, m.ChannelID) {
			return
		}

		// Forward to n8n asynchronously
		n8n.TriggerWebhookAsync(services.WebhookPayload{
			Type:      "message",
			Command:   "chat",
			Content:   m.Content,
			UserID:    m.Author.ID,
			UserName:  m.Author.Username,
			ChannelID: m.ChannelID,
			MessageID: m.ID,
		})
	}
}
