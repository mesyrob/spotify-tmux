// auth/auth.go
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

// TokenInfo holds the OAuth token information
type TokenInfo struct {
	Token       *oauth2.Token `json:"token"`
	ClientID    string        `json:"client_id"`
	LastRefresh time.Time     `json:"last_refresh"`
}

// AuthService handles Spotify authentication
type AuthService struct {
	config    *oauth2.Config
	tokenFile string
	token     *oauth2.Token
}

// NewAuthService creates a new authentication service
func NewAuthService(clientID, clientSecret, redirectURI string) *AuthService {
	homeDir, _ := os.UserHomeDir()
	tokenFile := fmt.Sprintf("%s/.spotify-tmux/token.json", homeDir)
	
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes: []string{
			"user-read-playback-state",
			"user-modify-playback-state",
			"user-read-currently-playing",
		},
		Endpoint: spotify.Endpoint,
	}
	
	return &AuthService{
		config:    config,
		tokenFile: tokenFile,
	}
}

// generateRandomState generates a random state for OAuth security
func generateRandomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Authenticate starts the OAuth flow
func (a *AuthService) Authenticate() error {
	// Generate a random state for CSRF protection
	state, err := generateRandomState()
	if err != nil {
		return err
	}
	
	// Create a channel to receive the authorization code
	codeChan := make(chan string)
	errChan := make(chan error)
	
	// Create an HTTP server for the callback
	server := &http.Server{Addr: ":8080"}
	
	// Define the callback handler
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Verify state
		if r.URL.Query().Get("state") != state {
			errChan <- fmt.Errorf("state mismatch")
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}
		
		// Get the code
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no code in response")
			http.Error(w, "No code in response", http.StatusBadRequest)
			return
		}
		
		// Send success page
		fmt.Fprint(w, "Authentication successful! You can now close this window.")
		
		// Send the code to the channel
		codeChan <- code
	})
	
	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()
	
	// Generate the auth URL
	authURL := a.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	
	// Print the auth URL
	fmt.Printf("Please open the following URL in your browser:\n%s\n", authURL)
	
	// Wait for the code or error
	var code string
	select {
	case code = <-codeChan:
		// Got the code, proceed
	case err := <-errChan:
		// Shutdown the server
		server.Shutdown(context.Background())
		return err
	case <-time.After(5 * time.Minute):
		// Timeout
		server.Shutdown(context.Background())
		return fmt.Errorf("authentication timed out")
	}
	
	// Shutdown the server
	server.Shutdown(context.Background())
	
	// Exchange the code for a token
	token, err := a.config.Exchange(context.Background(), code)
	if err != nil {
		return err
	}
	
	// Save the token
	a.token = token
	return a.saveToken()
}

// HasValidToken checks if a valid token exists
func (a *AuthService) HasValidToken() bool {
	if a.token != nil && a.token.Valid() {
		return true
	}
	
	// Try to load token from file
	if err := a.loadToken(); err != nil {
		return false
	}
	
	return a.token != nil && a.token.Valid()
}

// GetToken returns the OAuth token
func (a *AuthService) GetToken() (*oauth2.Token, error) {
	if a.token == nil {
		if err := a.loadToken(); err != nil {
			return nil, err
		}
	}
	
	// Check if token needs refresh
	if a.token != nil && !a.token.Valid() {
		// Refresh the token
		newToken, err := a.config.TokenSource(context.Background(), a.token).Token()
		if err != nil {
			return nil, err
		}
		
		a.token = newToken
		if err := a.saveToken(); err != nil {
			return nil, err
		}
	}
	
	return a.token, nil
}

// GetClient returns an HTTP client with authentication
func (a *AuthService) GetClient() (*http.Client, error) {
	token, err := a.GetToken()
	if err != nil {
		return nil, err
	}
	
	return a.config.Client(context.Background(), token), nil
}

// loadToken loads the token from file
func (a *AuthService) loadToken() error {
	// Check if token file exists
	if _, err := os.Stat(a.tokenFile); os.IsNotExist(err) {
		return fmt.Errorf("token file does not exist")
	}
	
	// Read the token file
	data, err := os.ReadFile(a.tokenFile)
	if err != nil {
		return err
	}
	
	// Parse the token
	var tokenInfo TokenInfo
	if err := json.Unmarshal(data, &tokenInfo); err != nil {
		return err
	}
	
	a.token = tokenInfo.Token
	return nil
}

// saveToken saves the token to file
func (a *AuthService) saveToken() error {
	// Ensure directory exists
	dir := fmt.Sprintf("%s/.spotify-tmux", os.Getenv("HOME"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// Create token info
	tokenInfo := TokenInfo{
		Token:       a.token,
		ClientID:    a.config.ClientID,
		LastRefresh: time.Now(),
	}
	
	// Serialize the token
	data, err := json.MarshalIndent(tokenInfo, "", "  ")
	if err != nil {
		return err
	}
	
	// Write to file
	return os.WriteFile(a.tokenFile, data, 0600)
}