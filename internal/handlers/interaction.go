package handlers

import (
    "github.com/bwmarrin/discordgo"
    "github.com/marshall/zero-ops-bot/internal/commands"
)

func NewInteractionHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        if i.Type != discordgo.InteractionApplicationCommand {
            return
        }

        switch i.ApplicationCommandData().Name {
        case "repo":
            commands.HandleRepoCommand(s, i)
        }
    }
}
