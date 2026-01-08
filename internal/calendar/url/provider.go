package url

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	cal "github.com/robgyiv/avail/internal/calendar"
	"github.com/robgyiv/avail/internal/config"
	"github.com/robgyiv/avail/pkg/availability"
)

// Provider implements the calendar.Provider interface for public calendar URLs.
// Works with any public calendar that serves iCalendar (.ics) format:
// - Apple/iCloud public calendars
// - Google Calendar public feeds
// - CalDAV server public calendars
// - Any HTTP/HTTPS endpoint serving .ics format
type Provider struct {
	calendarURL string
	client      *http.Client
}

// NewProvider creates a new URL provider for public calendars.
func NewProvider() *Provider {
	return &Provider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewProviderFromURL creates a provider with a specific public calendar URL.
func NewProviderFromURL(calendarURL string) *Provider {
	return &Provider{
		calendarURL: normalizeCalendarURL(calendarURL),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// normalizeCalendarURL converts webcal:// URLs to https:// and ensures proper format.
func normalizeCalendarURL(url string) string {
	// Convert webcal:// to https://
	url = strings.Replace(url, "webcal://", "https://", 1)

	// Ensure it's a valid URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return ""
	}

	return url
}

// Authenticate stores the public calendar URL (no actual authentication needed).
func (p *Provider) Authenticate(ctx context.Context) error {
	// For public calendars, we just need the URL
	// It should be provided via NewProviderFromURL or loaded from keyring
	if p.calendarURL == "" {
		return fmt.Errorf("public calendar URL required")
	}

	// Test the URL by fetching it
	resp, err := p.client.Get(p.calendarURL)
	if err != nil {
		return fmt.Errorf("failed to fetch calendar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to fetch calendar: %d %s", resp.StatusCode, string(bodyBytes))
	}

	// Store the URL in keyring (non-fatal - if keyring is unavailable, we still consider auth successful)
	// This allows the provider to work in CI environments where keyring services may not be available
	if err := config.StoreToken(p.calendarURL); err != nil {
		// Log but don't fail - the URL was validated successfully
		// In environments without keyring (e.g., CI), authentication should still succeed
		// The URL is already set in the provider, so LoadToken won't be needed
	}

	return nil
}

// LoadToken loads the public calendar URL from the keyring.
func (p *Provider) LoadToken(ctx context.Context) error {
	calendarURL, err := config.GetToken()
	if err != nil {
		if err == config.ErrTokenNotFound {
			return fmt.Errorf("not authenticated: %w", err)
		}
		return fmt.Errorf("failed to get calendar URL: %w", err)
	}

	p.calendarURL = normalizeCalendarURL(calendarURL)
	if p.calendarURL == "" {
		return fmt.Errorf("invalid calendar URL format")
	}

	return nil
}

// IsAuthenticated checks if the provider has a calendar URL.
func (p *Provider) IsAuthenticated() bool {
	return p.calendarURL != ""
}

// ListEvents fetches events from the public calendar URL.
func (p *Provider) ListEvents(ctx context.Context, start, end time.Time) ([]availability.Event, error) {
	if !p.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	// Fetch the .ics file
	req, err := http.NewRequestWithContext(ctx, "GET", p.calendarURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "availability/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch calendar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch calendar: %d %s", resp.StatusCode, string(bodyBytes))
	}

	// Read the iCalendar data
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read calendar data: %w", err)
	}

	icalData := string(bodyBytes)

	// Parse iCalendar format
	events, err := parseICalendar(icalData, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to parse calendar: %w", err)
	}

	return events, nil
}

// Ensure Provider implements the calendar.Provider interface.
var _ cal.Provider = (*Provider)(nil)
