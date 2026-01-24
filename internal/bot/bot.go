package bot

import (
    "fmt"
    "log"

    "github.com/bwmarrin/discordgo"
    "github.com/marshall/zero-ops-bot/internal/config"
    "github.com/marshall/zero-ops-bot/internal/handlers"
    "github.com/marshall/zero-ops-bot/internal/services"
)

type Bot struct {
    session *discordgo.Session
    config  *config.Config
}

func New(cfg *config.Config) (*Bot, error) {
    session, err := discordgo.New("Bot " + cfg.DiscordToken)
    if err != nil {
        return nil, fmt.Errorf("create session: %w", err)
    }

    session.Identify.Intents = discordgo.IntentsGuilds |
        discordgo.IntentsGuildMessages |
        discordgo.IntentMessageContent

    return &Bot{
        session: session,
        config:  cfg,
    }, nil
}

func (b *Bot) Start() error {
    n8nClient := services.NewN8nClient(b.config.N8nWebhookURL, b.config.N8nWebhookSecret)

    b.session.AddHandler(handlers.NewMessageHandler(n8nClient, b.config.AllowedChannels))
    b.session.AddHandler(handlers.NewMentionHandler(n8nClient))

    b.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
        log.Printf("Logged in as %s", r.User.String())
    })

    if err := b.session.Open(); err != nil {
        return fmt.Errorf("open session: %w", err)
    }

    return nil
}

func (b *Bot) Stop() error {
    return b.session.Close()
}
