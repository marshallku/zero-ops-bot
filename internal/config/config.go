package config

import (
    "errors"
    "os"
    "strings"
)

type Config struct {
    DiscordToken     string
    DiscordAppID     string
    DiscordGuildID   string
    N8nWebhookURL    string
    N8nWebhookSecret string
    AllowedChannels  []string
    MetadataPath     string
}

func Load() (*Config, error) {
    token := os.Getenv("DISCORD_TOKEN")
    if token == "" {
        return nil, errors.New("DISCORD_TOKEN is required")
    }

    appID := os.Getenv("DISCORD_APP_ID")
    if appID == "" {
        return nil, errors.New("DISCORD_APP_ID is required")
    }

    webhookURL := os.Getenv("N8N_WEBHOOK_URL")
    if webhookURL == "" {
        return nil, errors.New("N8N_WEBHOOK_URL is required")
    }

    var allowedChannels []string
    if channels := os.Getenv("ALLOWED_CHANNELS"); channels != "" {
        for _, ch := range strings.Split(channels, ",") {
            if trimmed := strings.TrimSpace(ch); trimmed != "" {
                allowedChannels = append(allowedChannels, trimmed)
            }
        }
    }

    metadataPath := os.Getenv("METADATA_PATH")
    if metadataPath == "" {
        metadataPath = "metadata.yaml"
    }

    return &Config{
        DiscordToken:     token,
        DiscordAppID:     appID,
        DiscordGuildID:   os.Getenv("DISCORD_GUILD_ID"),
        N8nWebhookURL:    webhookURL,
        N8nWebhookSecret: os.Getenv("N8N_WEBHOOK_SECRET"),
        AllowedChannels:  allowedChannels,
        MetadataPath:     metadataPath,
    }, nil
}
