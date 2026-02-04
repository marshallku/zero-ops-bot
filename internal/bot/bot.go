package bot

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/marshall/zero-ops-bot/internal/commands"
	"github.com/marshall/zero-ops-bot/internal/config"
	"github.com/marshall/zero-ops-bot/internal/handlers"
	"github.com/marshall/zero-ops-bot/internal/metadata"
	"github.com/marshall/zero-ops-bot/internal/notes"
	"github.com/marshall/zero-ops-bot/internal/scheduler"
	"github.com/marshall/zero-ops-bot/internal/services"
)

const shutdownTimeout = 10 * time.Second

type Bot struct {
	session   *discordgo.Session
	config    *config.Config
	n8nClient *services.N8nClient
	scheduler *scheduler.Scheduler
	notes     *notes.Store
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
	if err := metadata.Load(b.config.MetadataPath); err != nil {
		return fmt.Errorf("load metadata: %w", err)
	}

	b.n8nClient = services.NewN8nClient(b.config.N8nWebhookURL, b.config.N8nWebhookSecret)

	noteStore, err := notes.NewStore(b.config.NotesDir)
	if err != nil {
		return fmt.Errorf("init notes: %w", err)
	}
	b.notes = noteStore

	b.scheduler = scheduler.New(b.session, b.n8nClient, b.notes, b.config.Timezone)

	noteHandler := commands.NewNoteHandler(b.notes)
	scheduleHandler := commands.NewScheduleHandler(b.scheduler)

	b.session.AddHandler(handlers.NewInteractionHandler(handlers.InteractionHandlers{
		Note:     noteHandler,
		Schedule: scheduleHandler,
	}))
	b.session.AddHandler(handlers.NewMessageHandler(b.n8nClient, b.config.AllowedChannels))
	b.session.AddHandler(handlers.NewMentionHandler(b.n8nClient, b.notes))

	b.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as %s", r.User.String())
	})

	if err := b.session.Open(); err != nil {
		return fmt.Errorf("open session: %w", err)
	}

	if err := b.registerCommands(); err != nil {
		return fmt.Errorf("register commands: %w", err)
	}

	meta := metadata.Get()
	for _, sched := range meta.Schedules {
		if err := b.scheduler.Register(sched); err != nil {
			log.Printf("Failed to register schedule %q: %v", sched.Name, err)
		}
	}
	b.scheduler.Start()

	return nil
}

func (b *Bot) registerCommands() error {
	defs := commands.GetDefinitions()
	guildID := b.config.DiscordGuildID

	registered, err := b.session.ApplicationCommandBulkOverwrite(b.config.DiscordAppID, guildID, defs)
	if err != nil {
		return err
	}

	log.Printf("Registered %d commands", len(registered))
	return nil
}

func (b *Bot) Stop() error {
	if b.scheduler != nil {
		b.scheduler.Stop()
	}

	if b.n8nClient != nil {
		b.n8nClient.Shutdown(nil)
	}

	return b.session.Close()
}
