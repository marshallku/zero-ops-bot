package bot

import (
    "fmt"
    "log"

    "github.com/bwmarrin/discordgo"
    "github.com/marshall/zero-ops-bot/internal/commands"
    "github.com/marshall/zero-ops-bot/internal/config"
    "github.com/marshall/zero-ops-bot/internal/handlers"
    "github.com/marshall/zero-ops-bot/internal/services"
)

type Bot struct {
    session           *discordgo.Session
    config            *config.Config
    registeredCommands []*discordgo.ApplicationCommand
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

    b.session.AddHandler(handlers.NewInteractionHandler(n8nClient))
    b.session.AddHandler(handlers.NewMessageHandler(n8nClient, b.config.AllowedChannels))

    b.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
        log.Printf("Logged in as %s", r.User.String())
    })

    if err := b.session.Open(); err != nil {
        return fmt.Errorf("open session: %w", err)
    }

    if err := b.registerCommands(); err != nil {
        return fmt.Errorf("register commands: %w", err)
    }

    return nil
}

func (b *Bot) registerCommands() error {
    defs := commands.GetDefinitions()

    // Use guild ID if provided (instant updates), otherwise global
    guildID := b.config.DiscordGuildID

    registered, err := b.session.ApplicationCommandBulkOverwrite(b.config.DiscordAppID, guildID, defs)
    if err != nil {
        return err
    }

    b.registeredCommands = registered
    log.Printf("Registered %d commands", len(registered))

    return nil
}

func (b *Bot) Stop() error {
    // Optionally clean up commands on shutdown
    // for _, cmd := range b.registeredCommands {
    //     b.session.ApplicationCommandDelete(b.config.DiscordAppID, b.config.DiscordGuildID, cmd.ID)
    // }

    return b.session.Close()
}
