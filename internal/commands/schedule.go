package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/marshall/zero-ops-bot/internal/metadata"
)

type ScheduleReloader interface {
	Reload() error
}

var ScheduleCommand = &discordgo.ApplicationCommand{
	Name:        "schedule",
	Description: "Manage scheduled tasks",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "list",
			Description: "List all schedules",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
		},
		{
			Name:        "add",
			Description: "Add a scheduled task",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "name",
					Description: "Schedule name",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "cron",
					Description: "Cron expression (e.g. '30 7 * * 1-5')",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "command",
					Description: "Webhook command to trigger",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "channel",
					Description: "Channel ID to post results (default: current channel)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
				},
			},
		},
		{
			Name:        "remove",
			Description: "Remove a scheduled task",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "name",
					Description: "Schedule name to remove",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
	},
}

func NewScheduleHandler(reloader ScheduleReloader) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		if len(options) == 0 {
			respond(s, i, "No subcommand provided")
			return
		}

		switch options[0].Name {
		case "list":
			handleScheduleList(s, i)
		case "add":
			handleScheduleAdd(s, i, options[0].Options, reloader)
		case "remove":
			handleScheduleRemove(s, i, options[0].Options, reloader)
		}
	}
}

func handleScheduleList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	schedules := metadata.ListSchedules()

	if len(schedules) == 0 {
		respond(s, i, "No schedules configured")
		return
	}

	var sb strings.Builder
	sb.WriteString("**Schedules:**\n")
	for _, sched := range schedules {
		flags := ""
		if sched.IncludeNotes {
			flags += " [notes]"
		}
		if sched.IncludeRepos {
			flags += " [repos]"
		}
		sb.WriteString(fmt.Sprintf("- **%s** `%s` → `%s`%s\n", sched.Name, sched.Cron, sched.Command, flags))
	}

	respond(s, i, sb.String())
}

func handleScheduleAdd(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption, reloader ScheduleReloader) {
	var name, cronExpr, command, channelID string
	for _, opt := range opts {
		switch opt.Name {
		case "name":
			name = opt.StringValue()
		case "cron":
			cronExpr = opt.StringValue()
		case "command":
			command = opt.StringValue()
		case "channel":
			channelID = opt.StringValue()
		}
	}

	if channelID == "" {
		channelID = i.ChannelID
	}

	schedule := metadata.Schedule{
		Name:      name,
		Cron:      cronExpr,
		ChannelID: channelID,
		Command:   command,
	}

	if err := metadata.AddSchedule(schedule); err != nil {
		respond(s, i, "Failed to add schedule: "+err.Error())
		return
	}

	if reloader != nil {
		if err := reloader.Reload(); err != nil {
			respond(s, i, fmt.Sprintf("Schedule saved but reload failed: %v", err))
			return
		}
	}

	respond(s, i, fmt.Sprintf("Added schedule **%s** (`%s` → `%s`)", name, cronExpr, command))
}

func handleScheduleRemove(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption, reloader ScheduleReloader) {
	name := opts[0].StringValue()

	if metadata.RemoveSchedule(name) {
		if reloader != nil {
			reloader.Reload()
		}
		respond(s, i, fmt.Sprintf("Removed schedule **%s**", name))
	} else {
		respond(s, i, fmt.Sprintf("Schedule **%s** not found", name))
	}
}
