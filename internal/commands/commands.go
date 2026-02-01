package commands

import "github.com/bwmarrin/discordgo"

func GetDefinitions() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		RepoCommand,
	}
}
