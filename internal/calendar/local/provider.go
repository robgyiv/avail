package local

import (
	"context"
	"fmt"
	"os"
	"time"

	cal "github.com/robgyiv/avail/internal/calendar"
	"github.com/robgyiv/avail/pkg/availability"
)

// Provider implements the calendar.Provider interface for local .ics files.
type Provider struct {
	icsFilePath string
}

// NewProvider creates a new local calendar provider.
func NewProvider() *Provider {
	return &Provider{}
}

// NewProviderFromPath creates a provider with a specific .ics file path.
func NewProviderFromPath(icsFilePath string) *Provider {
	return &Provider{
		icsFilePath: icsFilePath,
	}
}

// Authenticate stores the .ics file path (no actual authentication needed).
func (p *Provider) Authenticate(ctx context.Context) error {
	// For local calendars, we just need the file path
	// It should be provided via NewProviderFromPath or loaded from keyring
	if p.icsFilePath == "" {
		return fmt.Errorf("local calendar file path required")
	}

	// Verify the file exists and is readable
	if _, err := os.Stat(p.icsFilePath); os.IsNotExist(err) {
		return fmt.Errorf("calendar file does not exist: %s", p.icsFilePath)
	}

	// Try to read a small portion to verify it's readable
	file, err := os.Open(p.icsFilePath)
	if err != nil {
		return fmt.Errorf("failed to open calendar file: %w", err)
	}
	file.Close()

	// For local provider, path is stored in config file, not keyring
	// Keyring is only used for sensitive credentials (OAuth tokens, etc.)
	return nil
}

// LoadToken is not used for local provider - path comes from config file.
// This method exists to satisfy the interface but should not be called.
func (p *Provider) LoadToken(ctx context.Context) error {
	// Local provider paths are stored in config, not keyring
	return fmt.Errorf("local provider does not use keyring - set local_calendar_path in config file")
}

// IsAuthenticated checks if the provider has a calendar file path.
func (p *Provider) IsAuthenticated() bool {
	return p.icsFilePath != ""
}

// ListEvents reads events from the local .ics file.
func (p *Provider) ListEvents(ctx context.Context, start, end time.Time) ([]availability.Event, error) {
	if !p.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	// Read the .ics file
	icalData, err := os.ReadFile(p.icsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read calendar file: %w", err)
	}

	// Parse iCalendar format
	events, err := parseICalendar(string(icalData), start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to parse calendar: %w", err)
	}

	return events, nil
}

// Ensure Provider implements the calendar.Provider interface.
var _ cal.Provider = (*Provider)(nil)
