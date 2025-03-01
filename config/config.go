// config/config.go
package config

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"errors"
	"os"
	"path/filepath"
	"log"
)

func init() {
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }
}

// Config holds the application configuration
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	TokenFile    string `json:"token_file"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()
	
	return Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURI:  "http://localhost:8080/callback",
		TokenFile:    filepath.Join(homeDir, ".spotify-tmux", "token.json"),
	}
}

// Load loads the configuration from file or environment
func Load() (Config, error) {
	config := DefaultConfig()
	
	// Try to load from file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config, err
	}
	
	configDir := filepath.Join(homeDir, ".spotify-tmux")
	os.MkdirAll(configDir, 0755)
	
	configFile := filepath.Join(configDir, "config.json")
	
	// Check if config file exists
	if _, err := os.Stat(configFile); err == nil {
		// Load from file
		data, err := os.ReadFile(configFile)
		if err != nil {
			return config, err
		}
		
		if err := json.Unmarshal(data, &config); err != nil {
			return config, err
		}
	}
	
	// Check environment variables
	if os.Getenv("SPOTIFY_CLIENT_ID") != "" {
		config.ClientID = os.Getenv("SPOTIFY_CLIENT_ID")
	}
	
	if os.Getenv("SPOTIFY_CLIENT_SECRET") != "" {
		config.ClientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")
	}
	
	if os.Getenv("SPOTIFY_REDIRECT_URI") != "" {
		config.RedirectURI = os.Getenv("SPOTIFY_REDIRECT_URI")
	}
	
	// Validate configuration
	if config.ClientID == "" || config.ClientSecret == "" {
		return config, errors.New("client ID and secret must be provided")
	}
	
	// Ensure token directory exists
	tokenDir := filepath.Dir(config.TokenFile)
	if err := os.MkdirAll(tokenDir, 0755); err != nil {
		return config, err
	}
	
	return config, nil
}

// Save saves the configuration to file
func Save(config Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	configDir := filepath.Join(homeDir, ".spotify-tmux")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	
	configFile := filepath.Join(configDir, "config.json")
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(configFile, data, 0644)
}