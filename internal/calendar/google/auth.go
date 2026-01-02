package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	// OAuth2 scopes for Google Calendar API (read-only)
	scope = calendar.CalendarReadonlyScope
)

var (
	oauth2Config *oauth2.Config
)

// OAuthConfig returns the OAuth2 configuration for Google Calendar.
func OAuthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	if oauth2Config != nil {
		return oauth2Config
	}

	oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{scope},
		Endpoint:     google.Endpoint,
	}

	return oauth2Config
}

// Authenticate performs OAuth2 authentication flow.
// For MVP, this is a simplified implementation that opens a browser.
func Authenticate(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	// Generate auth URL
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	// Open browser
	if err := openBrowser(authURL); err != nil {
		return nil, fmt.Errorf("failed to open browser: %w", err)
	}

	// For MVP, we'll need the user to copy the code manually
	// In a full implementation, this would use a local server to catch the callback
	fmt.Printf("Please visit: %s\n", authURL)
	fmt.Print("Enter the authorization code: ")

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, fmt.Errorf("failed to read authorization code: %w", err)
	}

	// Exchange code for token
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}

// openBrowser opens the specified URL in the default browser.
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)

	return exec.Command(cmd, args...).Start()
}

// TokenFromJSON converts a JSON token string to an oauth2.Token.
func TokenFromJSON(jsonToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{}
	if err := json.Unmarshal([]byte(jsonToken), token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}
	return token, nil
}

// TokenToJSON converts an oauth2.Token to a JSON string.
func TokenToJSON(token *oauth2.Token) (string, error) {
	data, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token: %w", err)
	}
	return string(data), nil
}

// RefreshToken refreshes an OAuth2 token if it has expired.
func RefreshToken(ctx context.Context, config *oauth2.Config, token *oauth2.Token) (*oauth2.Token, error) {
	if token.Valid() {
		return token, nil
	}

	newToken, err := config.TokenSource(ctx, token).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return newToken, nil
}

// NewCalendarService creates a new Google Calendar service from a token.
func NewCalendarService(ctx context.Context, token *oauth2.Token) (*calendar.Service, error) {
	config := &oauth2.Config{
		Endpoint: google.Endpoint,
	}

	client := config.Client(ctx, token)
	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	return service, nil
}

