package commands

import (
    "fmt"
    "strings"

    "github.com/bwmarrin/discordgo"
    "github.com/marshall/zero-ops-bot/internal/metadata"
)

var RepoCommand = &discordgo.ApplicationCommand{
    Name:        "repo",
    Description: "Manage repository metadata",
    Options: []*discordgo.ApplicationCommandOption{
        {
            Name:        "add",
            Description: "Add or update a repository",
            Type:        discordgo.ApplicationCommandOptionSubCommand,
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Name:        "name",
                    Description: "Repository name",
                    Type:        discordgo.ApplicationCommandOptionString,
                    Required:    true,
                },
                {
                    Name:        "description",
                    Description: "What this repo is for",
                    Type:        discordgo.ApplicationCommandOptionString,
                    Required:    true,
                },
            },
        },
        {
            Name:        "list",
            Description: "List all repositories",
            Type:        discordgo.ApplicationCommandOptionSubCommand,
        },
        {
            Name:        "remove",
            Description: "Remove a repository",
            Type:        discordgo.ApplicationCommandOptionSubCommand,
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Name:        "name",
                    Description: "Repository name to remove",
                    Type:        discordgo.ApplicationCommandOptionString,
                    Required:    true,
                },
            },
        },
    },
}

func HandleRepoCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
    options := i.ApplicationCommandData().Options
    if len(options) == 0 {
        respond(s, i, "No subcommand provided")
        return
    }

    switch options[0].Name {
    case "add":
        handleRepoAdd(s, i, options[0].Options)
    case "list":
        handleRepoList(s, i)
    case "remove":
        handleRepoRemove(s, i, options[0].Options)
    }
}

func handleRepoAdd(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
    var name, description string
    for _, opt := range opts {
        switch opt.Name {
        case "name":
            name = opt.StringValue()
        case "description":
            description = opt.StringValue()
        }
    }

    err := metadata.AddRepo(metadata.Repo{
        Name:        name,
        Description: description,
    })

    if err != nil {
        respond(s, i, "Failed to add repo: "+err.Error())
        return
    }

    respond(s, i, fmt.Sprintf("Added repo **%s**", name))
}

func handleRepoList(s *discordgo.Session, i *discordgo.InteractionCreate) {
    repos := metadata.ListRepos()

    if len(repos) == 0 {
        respond(s, i, "No repositories configured")
        return
    }

    var sb strings.Builder
    sb.WriteString("**Repositories:**\n")
    for _, repo := range repos {
        sb.WriteString(fmt.Sprintf("- **%s**: %s\n", repo.Name, repo.Description))
    }

    respond(s, i, sb.String())
}

func handleRepoRemove(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
    name := opts[0].StringValue()

    if metadata.RemoveRepo(name) {
        respond(s, i, fmt.Sprintf("Removed repo **%s**", name))
    } else {
        respond(s, i, fmt.Sprintf("Repo **%s** not found", name))
    }
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: content,
        },
    })
}
