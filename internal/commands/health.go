package commands

import (
    "context"
    "time"

    "github.com/bwmarrin/discordgo"
    "github.com/marshall/zero-ops-bot/internal/services"
    "github.com/marshall/zero-ops-bot/internal/utils"
)

var CheckHealth = &Command{
    Definition: &discordgo.ApplicationCommand{
        Name:        "check-health",
        Description: "Check system health via n8n workflow",
    },
    Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate, n8n *services.N8nClient) {
        // Defer reply for long-running operations
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
        })

        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        result, err := n8n.TriggerWebhook(ctx, services.WebhookPayload{
            Type:      "command",
            Command:   "check-health",
            UserID:    i.Member.User.ID,
            ChannelID: i.ChannelID,
        })

        var content string
        if err != nil {
            content = "Health check failed: " + err.Error()
        } else if result.Message != "" {
            content = result.Message
        } else {
            content = "Health check completed"
        }

        chunks := utils.SplitMessage(content)
        s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
            Content: &chunks[0],
        })

        for _, chunk := range chunks[1:] {
            s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
                Content: chunk,
            })
        }
    },
}
