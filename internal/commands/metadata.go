package commands

import "github.com/marshall/zero-ops-bot/internal/services"

var AvailableCommands = []services.CommandMeta{
    {Name: "chat", Description: "General conversation, questions, help"},
    {Name: "check-health", Description: "Check system health status"},
    {Name: "infra", Description: "Infrastructure operations, homelab management"},
}
