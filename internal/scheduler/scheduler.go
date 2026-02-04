package scheduler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/marshall/zero-ops-bot/internal/metadata"
	"github.com/marshall/zero-ops-bot/internal/notes"
	"github.com/marshall/zero-ops-bot/internal/services"
	"github.com/marshall/zero-ops-bot/internal/utils"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron    *cron.Cron
	session *discordgo.Session
	n8n     *services.N8nClient
	notes   *notes.Store
}

func New(session *discordgo.Session, n8n *services.N8nClient, notesStore *notes.Store, timezone string) *Scheduler {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("Invalid timezone %q, using Local: %v", timezone, err)
		loc = time.Local
	}

	return &Scheduler{
		cron:    cron.New(cron.WithLocation(loc)),
		session: session,
		n8n:     n8n,
		notes:   notesStore,
	}
}

func (s *Scheduler) Register(schedule metadata.Schedule) error {
	_, err := s.cron.AddFunc(schedule.Cron, func() {
		s.run(schedule)
	})
	if err != nil {
		return fmt.Errorf("register schedule %q: %w", schedule.Name, err)
	}

	log.Printf("Registered schedule: %s (cron: %s, command: %s, channel: %s)",
		schedule.Name, schedule.Cron, schedule.Command, schedule.ChannelID)
	return nil
}

func (s *Scheduler) Start() {
	s.cron.Start()
	log.Printf("Scheduler started with %d jobs", len(s.cron.Entries()))
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("Scheduler stopped")
}

func (s *Scheduler) Reload() error {
	ctx := s.cron.Stop()
	<-ctx.Done()

	loc := s.cron.Location()
	s.cron = cron.New(cron.WithLocation(loc))

	meta := metadata.Get()
	for _, schedule := range meta.Schedules {
		if err := s.Register(schedule); err != nil {
			log.Printf("Failed to register schedule %q on reload: %v", schedule.Name, err)
		}
	}

	s.cron.Start()
	log.Printf("Scheduler reloaded with %d jobs", len(s.cron.Entries()))
	return nil
}

func (s *Scheduler) run(schedule metadata.Schedule) {
	log.Printf("Running schedule: %s", schedule.Name)

	content := schedule.Prompt

	meta := metadata.Get()
	var repos []services.RepoMeta

	if schedule.IncludeRepos {
		repos = make([]services.RepoMeta, len(meta.Repos))
		for i, r := range meta.Repos {
			repos[i] = services.RepoMeta{
				Name:        r.Name,
				Description: r.Description,
				Path:        r.Path,
			}
		}

		var repoLines []string
		for _, r := range repos {
			repoLines = append(repoLines, fmt.Sprintf("- %s (%s): %s", r.Name, r.Path, r.Description))
		}
		if len(repoLines) > 0 {
			content += "\n\n## Registered Repositories\n" + strings.Join(repoLines, "\n")
		}
	}

	if schedule.IncludeNotes && s.notes != nil {
		today := time.Now().Format("2006-01-02")
		content += fmt.Sprintf("\n\n## Notes\nNotes directory: %s\nToday's notes: daily/%s.md", s.notes.BaseDir(), today)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := s.n8n.TriggerWebhook(ctx, services.WebhookPayload{
		Type:      "schedule",
		Command:   schedule.Command,
		ChannelID: schedule.ChannelID,
		Content:   content,
		Repos:     repos,
	})
	if err != nil {
		log.Printf("Schedule %s webhook failed: %v", schedule.Name, err)
		return
	}

	if result.Message == "" {
		return
	}

	for _, chunk := range utils.SplitMessage(result.Message) {
		if _, err := s.session.ChannelMessageSend(schedule.ChannelID, chunk); err != nil {
			log.Printf("Schedule %s message send failed: %v", schedule.Name, err)
			return
		}
	}
}
