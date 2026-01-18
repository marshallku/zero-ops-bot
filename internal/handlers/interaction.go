package handlers

import (
    "log"

    "github.com/bwmarrin/discordgo"
    "github.com/marshall/zero-ops-bot/internal/commands"
    "github.com/marshall/zero-ops-bot/internal/services"
)

func NewInteractionHandler(n8n *services.N8nClient) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        if i.Type != discordgo.InteractionApplicationCommand {
            return
        }

        cmd, ok := commands.Registry[i.ApplicationCommandData().Name]
        if !ok {
            log.Printf("Unknown command: %s", i.ApplicationCommandData().Name)
            return
        }

        cmd.Handler(s, i, n8n)
    }
}
