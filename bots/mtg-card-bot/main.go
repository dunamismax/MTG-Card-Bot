package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dunamismax/go-discord-bots/pkg/config"
	"github.com/dunamismax/go-discord-bots/pkg/discord"
	"github.com/dunamismax/go-discord-bots/pkg/scryfall"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	
	log.Printf("Starting MTG Card Bot...")
	log.Printf("Bot Name: %s", cfg.BotName)
	log.Printf("Command Prefix: %s", cfg.CommandPrefix)
	log.Printf("Log Level: %s", cfg.LogLevel)
	
	// Create Scryfall client
	scryfallClient := scryfall.NewClient()
	log.Println("Scryfall client initialized")
	
	// Create Discord bot
	bot, err := discord.NewBot(cfg, scryfallClient)
	if err != nil {
		log.Fatalf("Failed to create Discord bot: %v", err)
	}
	
	// Start the bot
	if err := bot.Start(); err != nil {
		log.Fatalf("Failed to start Discord bot: %v", err)
	}
	
	// Print usage instructions
	printUsageInstructions(cfg.CommandPrefix)
	
	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	
	log.Println("Bot is running. Press Ctrl+C to stop.")
	<-stop
	
	log.Println("Shutting down bot...")
	
	// Stop the bot
	if err := bot.Stop(); err != nil {
		log.Printf("Error stopping bot: %v", err)
	}
	
	log.Println("Bot stopped successfully")
}

func printUsageInstructions(prefix string) {
	log.Println("")
	log.Println("=== MTG Card Bot Usage ===")
	log.Printf("%s<card-name> - Look up a Magic: The Gathering card by name", prefix)
	log.Printf("Example: %sthe-one-ring", prefix)
	log.Printf("Example: %sLightning Bolt", prefix)
	log.Printf("%srandom - Get a random Magic card", prefix)
	log.Println("")
	log.Println("The bot supports fuzzy matching, so partial names work!")
	log.Println("Examples of valid searches:")
	log.Printf("- %sjac bele (finds Jace Beleren)", prefix)
	log.Printf("- %sbol (finds Lightning Bolt)", prefix)
	log.Printf("- %sforce of will", prefix)
	log.Println("===========================")
	log.Println("")
}