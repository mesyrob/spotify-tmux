// player/player.go
package player

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
//	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const (
	baseURL = "https://api.spotify.com/v1"
)

// Track represents a Spotify track
type Track struct {
	Name     string   `json:"name"`
	Artists  []Artist `json:"artists"`
	Album    Album    `json:"album"`
	Duration int      `json:"duration_ms"`
	URI      string   `json:"uri"`
}

// Artist represents a Spotify artist
type Artist struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

// Album represents a Spotify album
type Album struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

// CurrentlyPlaying represents the currently playing track
type CurrentlyPlaying struct {
	IsPlaying bool    `json:"is_playing"`
	Track     Track   `json:"item"`
	Progress  int     `json:"progress_ms"`
	Timestamp int64   `json:"timestamp"`
}

// TokenProvider is an interface for getting OAuth tokens
type TokenProvider interface {
	GetToken() (*oauth2.Token, error)
	GetClient() (*http.Client, error)
}

// PlayerService handles Spotify playback control
type PlayerService struct {
	token         *oauth2.Token
	tokenProvider TokenProvider
	client        *http.Client
}

// NewPlayerService creates a new player service
func NewPlayerService(token *oauth2.Token, tokenProvider TokenProvider) *PlayerService {
	return &PlayerService{
		token:         token,
		tokenProvider: tokenProvider,
	}
}

// getClient gets a valid HTTP client
func (p *PlayerService) getClient() (*http.Client, error) {
	if p.client != nil {
		return p.client, nil
	}
	
	client, err := p.tokenProvider.GetClient()
	if err != nil {
		return nil, err
	}
	
	p.client = client
	return client, nil
}

// GetCurrentlyPlaying gets the currently playing track
func (p *PlayerService) GetCurrentlyPlaying() (*CurrentlyPlaying, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	
	// Make the request
	resp, err := client.Get(baseURL + "/me/player/currently-playing")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Check if no content (no track playing)
	if resp.StatusCode == http.StatusNoContent {
		return &CurrentlyPlaying{IsPlaying: false}, nil
	}
	
	// Check for errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s, %s", resp.Status, string(body))
	}
	
	// Parse the response
	var current CurrentlyPlaying
	if err := json.NewDecoder(resp.Body).Decode(&current); err != nil {
		return nil, err
	}
	
	return &current, nil
}

// Play starts or resumes playback
func (p *PlayerService) Play() error {
	client, err := p.getClient()
	if err != nil {
		return err
	}
	
	// Create request
	req, err := http.NewRequest("PUT", baseURL+"/me/player/play", nil)
	if err != nil {
		return err
	}
	
	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// Check for errors
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s, %s", resp.Status, string(body))
	}
	
	return nil
}

// Pause pauses playback
func (p *PlayerService) Pause() error {
	client, err := p.getClient()
	if err != nil {
		return err
	}
	
	// Create request
	req, err := http.NewRequest("PUT", baseURL+"/me/player/pause", nil)
	if err != nil {
		return err
	}
	
	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// Check for errors
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s, %s", resp.Status, string(body))
	}
	
	return nil
}

// Next skips to the next track
func (p *PlayerService) Next() error {
	client, err := p.getClient()
	if err != nil {
		return err
	}
	
	// Create request
	req, err := http.NewRequest("POST", baseURL+"/me/player/next", nil)
	if err != nil {
		return err
	}
	
	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// Check for errors
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s, %s", resp.Status, string(body))
	}
	
	return nil
}

// Previous goes to the previous track
func (p *PlayerService) Previous() error {
	client, err := p.getClient()
	if err != nil {
		return err
	}
	
	// Create request
	req, err := http.NewRequest("POST", baseURL+"/me/player/previous", nil)
	if err != nil {
		return err
	}
	
	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// Check for errors
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s, %s", resp.Status, string(body))
	}
	
	return nil
}

// PlayPause toggles play/pause
func (p *PlayerService) PlayPause() error {
	// Get current state
	current, err := p.GetCurrentlyPlaying()
	if err != nil {
		return err
	}
	
	// Toggle based on current state
	if current.IsPlaying {
		return p.Pause()
	}
	return p.Play()
}

// FormatTrackInfo formats the current track information
func (p *PlayerService) FormatTrackInfo() (string, error) {
	current, err := p.GetCurrentlyPlaying()
	if err != nil {
		return "", err
	}
	
	if !current.IsPlaying || current.Track.Name == "" {
		return "No track currently playing", nil
	}
	
	// Format artists
	artistNames := make([]string, len(current.Track.Artists))
	for i, artist := range current.Track.Artists {
		artistNames[i] = artist.Name
	}
	artists := strings.Join(artistNames, ", ")
	
	// Format progress
	progress := time.Duration(current.Progress) * time.Millisecond
	duration := time.Duration(current.Track.Duration) * time.Millisecond
	progressStr := fmt.Sprintf("%d:%02d/%d:%02d", 
		int(progress.Minutes()), int(progress.Seconds())%60,
		int(duration.Minutes()), int(duration.Seconds())%60)
	
	return fmt.Sprintf("%s - %s (%s)", artists, current.Track.Name, progressStr), nil
}