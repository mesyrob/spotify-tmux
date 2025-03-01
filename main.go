// main.go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"github.com/yourusername/spotify-tmux/auth"
	"github.com/yourusername/spotify-tmux/config"
	"github.com/yourusername/spotify-tmux/player"
	"github.com/yourusername/spotify-tmux/ui"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize auth service
	authService := auth.NewAuthService(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURI)
	
	// Check if we need to authenticate
	if !authService.HasValidToken() {
		fmt.Println("No valid token found. Starting authentication flow...")
		if err := authService.Authenticate(); err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}
	}
	
	// Get the token
	token, err := authService.GetToken()
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}
	
	// Initialize player service
	playerService := player.NewPlayerService(token, authService)
	
	// Initialize UI
	userInterface := ui.NewUI(playerService)
	
	// Start the UI
	go userInterface.Start()
	
	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigChan
	fmt.Println("\nShutting down...")
	userInterface.Stop()
}