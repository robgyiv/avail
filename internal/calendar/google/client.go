package google

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"

	cal "github.com/robgyiv/avail/internal/calendar"
	"github.com/robgyiv/avail/internal/config"
	"github.com/robgyiv/avail/pkg/availability"
)

// Provider implements the calendar.Provider interface for Google Calendar.
type Provider struct {
	token      *oauth2.Token
	config     *oauth2.Config
	client     *calendar.Service
	calendarID string // Calendar ID: "primary", email address, or other calendar ID
}

// NewProvider creates a new Google Calendar provider.
// Defaults to "primary" calendar. Use SetCalendarID to change it.
func NewProvider() *Provider {
	// For MVP, we'll need OAuth credentials
	// In production, these would come from config or environment variables
	return &Provider{
		calendarID: "primary", // Default to primary calendar
	}
}

// SetCalendarID sets the calendar ID to fetch events from.
// Can be "primary" for the main calendar or an email address for other calendars.
func (p *Provider) SetCalendarID(calendarID string) {
	p.calendarID = calendarID
}

// Authenticate performs OAuth2 authentication and stores the token.
func (p *Provider) Authenticate(ctx context.Context) error {
	// Get OAuth credentials from environment
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("OAuth credentials required (set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET environment variables)")
	}

	// OAuthConfig will set redirect URL in Authenticate function
	oauthConfig := OAuthConfig(clientID, clientSecret, "")

	token, err := Authenticate(ctx, oauthConfig)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	p.token = token

	// Store token in keyring
	tokenJSON, err := TokenToJSON(token)
	if err != nil {
		return fmt.Errorf("failed to serialize token: %w", err)
	}

	if err := config.StoreToken(tokenJSON); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	// Create calendar service
	service, err := NewCalendarService(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to create calendar service: %w", err)
	}

	p.client = service
	p.config = oauthConfig
	return nil
}

// IsAuthenticated checks if the provider has a valid token.
func (p *Provider) IsAuthenticated() bool {
	if p.token == nil {
		return false
	}
	return p.token.Valid()
}

// LoadToken loads a token from the keyring.
func (p *Provider) LoadToken(ctx context.Context) error {
	tokenJSON, err := config.GetToken()
	if err != nil {
		if err == config.ErrTokenNotFound {
			return fmt.Errorf("not authenticated: %w", err)
		}
		return fmt.Errorf("failed to get token: %w", err)
	}

	token, err := TokenFromJSON(tokenJSON)
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	// Refresh token if needed
	// For refresh, we don't need the full config, just the endpoint
	oauthConfig := &oauth2.Config{
		Endpoint: google.Endpoint,
	}
	token, err = RefreshToken(ctx, oauthConfig, token)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	p.token = token

	// Create calendar service
	service, err := NewCalendarService(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to create calendar service: %w", err)
	}

	p.client = service

	// Update stored token if it was refreshed
	tokenJSON, err = TokenToJSON(token)
	if err == nil {
		config.StoreToken(tokenJSON)
	}

	return nil
}

// ListEvents fetches events from the specified Google Calendar.
func (p *Provider) ListEvents(ctx context.Context, start, end time.Time) ([]availability.Event, error) {
	if p.client == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	// Fetch events from the configured calendar
	events, err := p.client.Events.List(p.calendarID).
		TimeMin(start.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()

	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	result := make([]availability.Event, 0, len(events.Items))
	for _, item := range events.Items {
		if item.Start.DateTime == "" {
			// Skip all-day events for now (they're handled differently)
			continue
		}

		startTime, err := time.Parse(time.RFC3339, item.Start.DateTime)
		if err != nil {
			continue
		}

		endTime, err := time.Parse(time.RFC3339, item.End.DateTime)
		if err != nil {
			continue
		}

		result = append(result, availability.Event{
			Start: startTime,
			End:   endTime,
			Title: item.Summary,
		})
	}

	return result, nil
}

// Ensure Provider implements the calendar.Provider interface.
var _ cal.Provider = (*Provider)(nil)
