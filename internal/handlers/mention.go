package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/marshall/zero-ops-bot/internal/metadata"
	"github.com/marshall/zero-ops-bot/internal/notes"
	"github.com/marshall/zero-ops-bot/internal/services"
	"github.com/marshall/zero-ops-bot/internal/state"
	"github.com/marshall/zero-ops-bot/internal/utils"
)

type noteAction struct {
	Action   string `json:"action"`
	Text     string `json:"text"`
	Category string `json:"category"`
}

func NewMentionHandler(n8n *services.N8nClient, noteStore *notes.Store) func(s *discordgo.Session, m *discordgo.MessageCreate) {
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

		analyzed, err := n8n.TriggerWebhookJSON(ctx, services.WebhookPayload{
			Type:    "mention",
			Command: "analyze",
			Content: "Given the following system context and user message, determine the user's intent, select appropriate tools, and build a prompt to execute their request.\n\nSystem context:\n" + meta.SystemPrompt + "\n\nUser message:\n" + content,
			UserID:    m.Author.ID,
			UserName:  m.Author.Username,
			ChannelID: m.ChannelID,
			ThreadID:  threadID,
			SessionID: sessionID,
			MessageID: m.ID,
			Repos:     repos,
		})
		if err != nil {
			s.MessageReactionRemove(m.ChannelID, m.ID, "👀", s.State.User.ID)
			s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
			s.ChannelMessageSend(threadID, "Sorry, I encountered an error: "+err.Error())
			return
		}

		if analyzed.Command == "note" && noteStore != nil {
			handleNoteAction(s, m, threadID, analyzed.Content, noteStore)
			return
		}

		executionContent := analyzed.Content
		if noteStore != nil {
			today := time.Now().Format("2006-01-02")
			executionContent += fmt.Sprintf("\n\nNotes directory: %s\nToday's notes: daily/%s.md\nCategories directory: %s/categories/", noteStore.BaseDir(), today, noteStore.BaseDir())
		}

		result, err := n8n.TriggerWebhook(ctx, services.WebhookPayload{
			Type:      "mention",
			Command:   analyzed.Command,
			Content:   executionContent,
			UserID:    m.Author.ID,
			UserName:  m.Author.Username,
			ChannelID: m.ChannelID,
			ThreadID:  threadID,
			SessionID: sessionID,
			MessageID: m.ID,
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

func handleNoteAction(s *discordgo.Session, m *discordgo.MessageCreate, threadID, content string, store *notes.Store) {
	var action noteAction
	if err := json.Unmarshal([]byte(content), &action); err != nil {
		log.Printf("Failed to parse note action: %v", err)
		s.MessageReactionRemove(m.ChannelID, m.ID, "👀", s.State.User.ID)
		s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
		s.ChannelMessageSend(threadID, "Sorry, I couldn't understand the note request.")
		return
	}

	switch action.Action {
	case "add":
		if err := store.Add(action.Text, action.Category); err != nil {
			s.MessageReactionRemove(m.ChannelID, m.ID, "👀", s.State.User.ID)
			s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
			s.ChannelMessageSend(threadID, "Failed to save note: "+err.Error())
			return
		}

		s.MessageReactionRemove(m.ChannelID, m.ID, "👀", s.State.User.ID)
		s.MessageReactionAdd(m.ChannelID, m.ID, "✅")

		label := "daily"
		if action.Category != "" && action.Category != "daily" {
			label = action.Category
		}
		s.ChannelMessageSend(threadID, fmt.Sprintf("Got it, noted in **%s**: %s", label, action.Text))

	default:
		s.MessageReactionRemove(m.ChannelID, m.ID, "👀", s.State.User.ID)
		s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
		s.ChannelMessageSend(threadID, "Unknown note action: "+action.Action)
	}
}
