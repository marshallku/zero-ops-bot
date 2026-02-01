package handlers

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/marshall/zero-ops-bot/internal/metadata"
	"github.com/marshall/zero-ops-bot/internal/services"
	"github.com/marshall/zero-ops-bot/internal/state"
	"github.com/marshall/zero-ops-bot/internal/utils"
)

func NewMentionHandler(n8n *services.N8nClient) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.Bot {
			return
		}

		channel, err := s.Channel(m.ChannelID)
		if err != nil {
			log.Printf("Failed to get channel: %v", err)
			return
		}

		isInActiveThread := channel.IsThread() && state.IsActiveThread(m.ChannelID)
		isBotMentioned := isMentioned(s, m)

		if !isBotMentioned && !isInActiveThread {
			return
		}

		s.MessageReactionAdd(m.ChannelID, m.ID, "👀")

		var threadID string
		if channel.IsThread() {
			threadID = m.ChannelID
			if !state.IsActiveThread(threadID) {
				state.AddThread(threadID)
			}
		} else {
			thread, err := s.MessageThreadStart(m.ChannelID, m.ID, "Chat", 60)
			if err != nil {
				log.Printf("Failed to create thread: %v", err)
				return
			}
			threadID = thread.ID
			state.AddThread(threadID)
		}

		content := stripMention(s, m.Content)

		sessionID := state.ThreadIDToSessionID(threadID)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		meta := metadata.Get()
		repos := make([]services.RepoMeta, len(meta.Repos))
		for i, r := range meta.Repos {
			repos[i] = services.RepoMeta{
				Name:        r.Name,
				Description: r.Description,
				Path:        r.Path,
			}
		}

		result, err := n8n.TriggerWebhook(ctx, services.WebhookPayload{
			Type:         "mention",
			Command:      "chat",
			Content:      content,
			UserID:       m.Author.ID,
			UserName:     m.Author.Username,
			ChannelID:    m.ChannelID,
			ThreadID:     threadID,
			SessionID:    sessionID,
			MessageID:    m.ID,
			SystemPrompt: meta.SystemPrompt,
			Repos:        repos,
		})

		if err != nil {
			s.MessageReactionRemove(m.ChannelID, m.ID, "👀", s.State.User.ID)
			s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
			s.ChannelMessageSend(threadID, "Sorry, I encountered an error: "+err.Error())
			return
		}

		s.MessageReactionRemove(m.ChannelID, m.ID, "👀", s.State.User.ID)
		s.MessageReactionAdd(m.ChannelID, m.ID, "✅")

		if result.Message != "" {
			chunks := utils.SplitMessage(result.Message)
			for _, chunk := range chunks {
				s.ChannelMessageSend(threadID, chunk)
			}
		}
	}
}

func isMentioned(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID {
			return true
		}
	}
	return false
}

func stripMention(s *discordgo.Session, content string) string {
	botID := s.State.User.ID
	content = strings.ReplaceAll(content, "<@"+botID+">", "")
	content = strings.ReplaceAll(content, "<@!"+botID+">", "")
	return strings.TrimSpace(content)
}
