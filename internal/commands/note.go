package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/marshall/zero-ops-bot/internal/notes"
)

var NoteCommand = &discordgo.ApplicationCommand{
	Name:        "note",
	Description: "Manage personal notes",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "add",
			Description: "Add a note",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "text",
					Description: "Note content",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "category",
					Description: "Category name (default: daily)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
				},
			},
		},
		{
			Name:        "today",
			Description: "Show today's notes",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
		},
		{
			Name:        "list",
			Description: "List notes by date or category",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "date",
					Description: "Date in YYYY-MM-DD format",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
				},
				{
					Name:        "category",
					Description: "Category name",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
				},
			},
		},
		{
			Name:        "remove",
			Description: "Remove a note by index",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "index",
					Description: "Note number to remove",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
				{
					Name:        "date",
					Description: "Date in YYYY-MM-DD format (default: today)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
				},
			},
		},
		{
			Name:        "search",
			Description: "Search across all notes",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "query",
					Description: "Search query",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
	},
}

func NewNoteHandler(store *notes.Store) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		if len(options) == 0 {
			respond(s, i, "No subcommand provided")
			return
		}

		switch options[0].Name {
		case "add":
			handleNoteAdd(s, i, options[0].Options, store)
		case "today":
			handleNoteToday(s, i, store)
		case "list":
			handleNoteList(s, i, options[0].Options, store)
		case "remove":
			handleNoteRemove(s, i, options[0].Options, store)
		case "search":
			handleNoteSearch(s, i, options[0].Options, store)
		}
	}
}

func handleNoteAdd(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption, store *notes.Store) {
	var text, category string
	for _, opt := range opts {
		switch opt.Name {
		case "text":
			text = opt.StringValue()
		case "category":
			category = opt.StringValue()
		}
	}

	if err := store.Add(text, category); err != nil {
		respond(s, i, "Failed to add note: "+err.Error())
		return
	}

	label := "daily"
	if category != "" {
		label = category
	}
	respond(s, i, fmt.Sprintf("Noted in **%s**: %s", label, text))
}

func handleNoteToday(s *discordgo.Session, i *discordgo.InteractionCreate, store *notes.Store) {
	content, err := store.GetToday()
	if err != nil {
		respond(s, i, "Failed to read notes: "+err.Error())
		return
	}
	if content == "" {
		respond(s, i, "No notes for today")
		return
	}
	respond(s, i, content)
}

func handleNoteList(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption, store *notes.Store) {
	var date, category string
	for _, opt := range opts {
		switch opt.Name {
		case "date":
			date = opt.StringValue()
		case "category":
			category = opt.StringValue()
		}
	}

	if category != "" {
		content, err := store.GetByCategory(category)
		if err != nil {
			respond(s, i, "Failed to read category: "+err.Error())
			return
		}
		if content == "" {
			respond(s, i, fmt.Sprintf("No notes in category **%s**", category))
			return
		}
		respond(s, i, content)
		return
	}

	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	content, err := store.GetByDate(date)
	if err != nil {
		respond(s, i, "Failed to read notes: "+err.Error())
		return
	}
	if content == "" {
		respond(s, i, fmt.Sprintf("No notes for %s", date))
		return
	}
	respond(s, i, content)
}

func handleNoteRemove(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption, store *notes.Store) {
	var index int64
	date := time.Now().Format("2006-01-02")

	for _, opt := range opts {
		switch opt.Name {
		case "index":
			index = opt.IntValue()
		case "date":
			date = opt.StringValue()
		}
	}

	if err := store.Remove(date, int(index)); err != nil {
		respond(s, i, "Failed to remove note: "+err.Error())
		return
	}

	respond(s, i, fmt.Sprintf("Removed note #%d from %s", index, date))
}

func handleNoteSearch(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption, store *notes.Store) {
	query := opts[0].StringValue()

	results, err := store.Search(query)
	if err != nil {
		respond(s, i, "Search failed: "+err.Error())
		return
	}

	if len(results) == 0 {
		respond(s, i, fmt.Sprintf("No notes matching **%s**", query))
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**Search results for \"%s\":**\n", query))
	for _, r := range results {
		sb.WriteString(r + "\n")
	}

	respond(s, i, sb.String())
}
