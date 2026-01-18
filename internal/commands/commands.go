package commands

import (
    "github.com/bwmarrin/discordgo"
    "github.com/marshall/zero-ops-bot/internal/services"
)

type Command struct {
    Definition *discordgo.ApplicationCommand
    Handler    func(s *discordgo.Session, i *discordgo.InteractionCreate, n8n *services.N8nClient)
}

var Registry = make(map[string]*Command)

func Register(cmd *Command) {
    Registry[cmd.Definition.Name] = cmd
}

func GetDefinitions() []*discordgo.ApplicationCommand {
    defs := make([]*discordgo.ApplicationCommand, 0, len(Registry))
    for _, cmd := range Registry {
        defs = append(defs, cmd.Definition)
    }
    return defs
}

func init() {
    Register(CheckHealth)
}
