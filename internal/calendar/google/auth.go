package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

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
	// Create a new config each time (redirectURL may vary)
	oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL, // Will be set in Authenticate if empty
		Scopes:       []string{scope},
		Endpoint:     google.Endpoint,
	}

	return oauth2Config
}

// Authenticate performs OAuth2 authentication flow using a local callback server.
func Authenticate(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://localhost:%d/callback", port)
	config.RedirectURL = redirectURL

	state := "state-token-" + fmt.Sprintf("%d", time.Now().Unix())
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	// Channel to receive the auth code
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Start local server to catch the callback
	server := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/callback" {
			http.NotFound(w, r)
			return
		}

		// Verify state
		if r.URL.Query().Get("state") != state {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			errChan <- fmt.Errorf("invalid state parameter")
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			errChan <- fmt.Errorf("missing authorization code")
			return
		}

		// Send success response
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
			<html>
				<body>
					<h1>Authentication successful!</h1>
					<p>You can close this window and return to the terminal.</p>
				</body>
			</html>
		`)

		codeChan <- code

		// Shutdown server after a short delay
		go func() {
			time.Sleep(500 * time.Millisecond)
			server.Shutdown(context.Background())
		}()
	})

	// Start server in goroutine
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Open browser
	fmt.Printf("Opening browser for authentication...\n")
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Please visit the following URL in your browser:\n%s\n\n", authURL)
	}

	// Wait for callback or timeout
	select {
	case code := <-codeChan:
		// Exchange code for token
		token, err := config.Exchange(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("failed to exchange code for token: %w", err)
		}
		fmt.Println("✓ Authentication successful!")
		return token, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("authentication timeout - please try again")
	}
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
