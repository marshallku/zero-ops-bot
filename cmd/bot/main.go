package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/marshall/zero-ops-bot/internal/bot"
	"github.com/marshall/zero-ops-bot/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	b, err := bot.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	if err := b.Start(); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}

	log.Println("Bot is running. Press Ctrl+C to exit.")

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down...")
	if err := b.Stop(); err != nil {
		log.Printf("Error stopping bot: %v", err)
	}
}
