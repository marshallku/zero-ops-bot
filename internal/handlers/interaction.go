package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/marshall/zero-ops-bot/internal/commands"
)

type InteractionHandlers struct {
	Note     func(s *discordgo.Session, i *discordgo.InteractionCreate)
	Schedule func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func NewInteractionHandler(h InteractionHandlers) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		switch i.ApplicationCommandData().Name {
		case "repo":
			commands.HandleRepoCommand(s, i)
		case "note":
			if h.Note != nil {
				h.Note(s, i)
			}
		case "schedule":
			if h.Schedule != nil {
				h.Schedule(s, i)
			}
		}
	}
}
