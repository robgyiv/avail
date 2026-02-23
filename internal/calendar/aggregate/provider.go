package aggregate

import (
	"context"
	"fmt"
	"os"
	"time"

	cal "github.com/robgyiv/avail/internal/calendar"
	googlecal "github.com/robgyiv/avail/internal/calendar/google"
	localcal "github.com/robgyiv/avail/internal/calendar/local"
	urlcal "github.com/robgyiv/avail/internal/calendar/url"
	"github.com/robgyiv/avail/internal/config"
	"github.com/robgyiv/avail/pkg/availability"
)

// Provider aggregates multiple calendar providers and merges their events.
// When multiple calendars have overlapping busy times, the merged result treats
// the union of all busy times as busy (conservative approach).
type Provider struct {
	providers []cal.Provider
	warnings  []string
}

// NewProvider creates a new aggregate provider from calendar configurations.
// It initializes all configured calendars and loads their credentials if needed.
// If a calendar fails to load, a warning is recorded but execution continues.
func NewProvider(calendars []config.Calendar) (*Provider, error) {
	if len(calendars) == 0 {
		return nil, fmt.Errorf("no calendars configured")
	}

	p := &Provider{
		providers: make([]cal.Provider, 0, len(calendars)),
		warnings:  make([]string, 0),
	}

	for i, calCfg := range calendars {
		provider, err := createProvider(calCfg)
		if err != nil {
			// Record warning but continue processing other calendars
			p.warnings = append(p.warnings, fmt.Sprintf("Calendar %d (%s): %v", i+1, calCfg.Provider, err))
			continue
		}

		// For providers that need authentication (google, network), load credentials
		if calCfg.Provider != "local" {
			if !provider.IsAuthenticated() {
				// Try to load token
				if loadable, ok := provider.(interface{ LoadToken(context.Context) error }); ok {
					ctx := context.Background()
					if err := loadable.LoadToken(ctx); err != nil {
						p.warnings = append(p.warnings, fmt.Sprintf("Calendar %d (%s): authentication failed: %v", i+1, calCfg.Provider, err))
						continue
					}
				} else {
					p.warnings = append(p.warnings, fmt.Sprintf("Calendar %d (%s): not authenticated", i+1, calCfg.Provider))
					continue
				}
			}
		}

		p.providers = append(p.providers, provider)
	}

	if len(p.providers) == 0 {
		return nil, fmt.Errorf("no calendars could be loaded")
	}

	return p, nil
}

// createProvider creates a single calendar provider based on configuration.
func createProvider(calCfg config.Calendar) (cal.Provider, error) {
	switch calCfg.Provider {
	case "google":
		provider := googlecal.NewProvider()
		// If calendar_id is specified, set it on the provider
		if calCfg.CalendarID != "" {
			provider.SetCalendarID(calCfg.CalendarID)
		}
		return provider, nil
	case "network":
		if calCfg.URL == "" {
			return nil, fmt.Errorf("url is required for network provider")
		}
		return urlcal.NewProviderFromURL(calCfg.URL), nil
	case "local":
		if calCfg.Path == "" {
			return nil, fmt.Errorf("path is required for local provider")
		}
		return localcal.NewProviderFromPath(calCfg.Path), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", calCfg.Provider)
	}
}

// ListEvents fetches events from all configured calendars and merges them.
// Returns the union of all events from all calendars. If any calendar fails,
// it's skipped but other calendars are still processed.
func (p *Provider) ListEvents(ctx context.Context, start, end time.Time) ([]availability.Event, error) {
	allEvents := make([]availability.Event, 0)

	for i, provider := range p.providers {
		events, err := provider.ListEvents(ctx, start, end)
		if err != nil {
			p.warnings = append(p.warnings, fmt.Sprintf("Failed to fetch events from calendar %d: %v", i+1, err))
			continue
		}
		allEvents = append(allEvents, events...)
	}

	return allEvents, nil
}

// Authenticate authenticates all providers that support it.
// Returns an error if authentication fails for any provider.
func (p *Provider) Authenticate(ctx context.Context) error {
	for i, provider := range p.providers {
		if err := provider.Authenticate(ctx); err != nil {
			return fmt.Errorf("failed to authenticate calendar %d: %w", i+1, err)
		}
	}
	return nil
}

// IsAuthenticated returns true if at least one provider is authenticated.
// In practice, since we skip unauthenticated providers in NewProvider,
// this should always return true for loaded providers.
func (p *Provider) IsAuthenticated() bool {
	for _, provider := range p.providers {
		if provider.IsAuthenticated() {
			return true
		}
	}
	return false
}

// Warnings returns any non-fatal warnings that occurred during provider initialization.
// These should be displayed to the user so they can verify all expected calendars are loaded.
func (p *Provider) Warnings() []string {
	return p.warnings
}

// PrintWarnings outputs any warnings to stderr.
func (p *Provider) PrintWarnings() {
	for _, warning := range p.warnings {
		fmt.Fprintf(os.Stderr, "⚠️  %s\n", warning)
	}
}
