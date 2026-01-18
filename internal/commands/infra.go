package commands

import (
    "context"
    "time"

    "github.com/bwmarrin/discordgo"
    "github.com/marshall/zero-ops-bot/internal/services"
)

var Infra = &Command{
    Definition: &discordgo.ApplicationCommand{
        Name:        "infra",
        Description: "Trigger infrastructure workflow via n8n",
    },
    Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate, n8n *services.N8nClient) {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
        })

        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        result, err := n8n.TriggerWebhook(ctx, services.WebhookPayload{
            Type:      "command",
            Command:   "infra",
            UserID:    i.Member.User.ID,
            ChannelID: i.ChannelID,
        })

        var content string
        if err != nil {
            content = "Infra workflow failed: " + err.Error()
        } else if result.Message != "" {
            content = result.Message
        } else {
            content = "Infra workflow triggered"
        }

        s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
            Content: &content,
        })
    },
}
