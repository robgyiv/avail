package cli

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

// AvailabilityData contains all the data needed for availability calculations.
type AvailabilityData struct {
	Ctx       context.Context
	Cfg       *config.Config
	Location  *time.Location
	WorkHours availability.WorkHours
	Provider  cal.Provider
	Events    []availability.Event
	StartDate time.Time
	EndDate   time.Time
}

// LoadAvailabilityData loads configuration, creates a calendar provider, authenticates,
// and fetches events for the specified number of days.
func LoadAvailabilityData(days int) (*AvailabilityData, error) {
	ctx := context.Background()

	// Load config
	cfg, err := config.LoadOrCreate()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Load timezone
	location, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %w", err)
	}

	// Get work hours
	workHours, err := cfg.WorkHours()
	if err != nil {
		return nil, fmt.Errorf("invalid work hours: %w", err)
	}

	// Create calendar provider based on config
	var provider cal.Provider
	providerName := cfg.CalendarProvider
	if providerName == "" {
		// Migration: if LocalCalendarPath is set, assume local provider
		if cfg.LocalCalendarPath != "" {
			providerName = "local"
		} else {
			providerName = "google" // Default
		}
	}

	switch providerName {
	case "google":
		provider = googlecal.NewProvider()
	case "network":
		if cfg.CalendarURL != "" {
			provider = urlcal.NewProviderFromURL(cfg.CalendarURL)
		} else {
			provider = urlcal.NewProvider()
		}
	case "local":
		if cfg.LocalCalendarPath == "" {
			return nil, fmt.Errorf("local_calendar_path not set in config file\n\nSet local_calendar_path in ~/.config/avail/config.toml:\n  local_calendar_path = \"/path/to/calendar.ics\"")
		}
		provider = localcal.NewProviderFromPath(cfg.LocalCalendarPath)
	default:
		return nil, fmt.Errorf("unknown provider: %s (supported: google, network, local)", providerName)
	}

	// For providers that need authentication (google, network), load credentials from keyring
	if providerName != "local" {
		if !provider.IsAuthenticated() {
			// Try to load token (provider-specific)
			if loadable, ok := provider.(interface{ LoadToken(context.Context) error }); ok {
				if err := loadable.LoadToken(ctx); err != nil {
					fmt.Fprintf(os.Stderr, "Not authenticated. Please run 'avail auth' first.\n")
					return nil, err
				}
			} else {
				fmt.Fprintf(os.Stderr, "Not authenticated. Please run 'avail auth' first.\n")
				return nil, fmt.Errorf("not authenticated")
			}
		}
	}

	// Calculate date range
	now := time.Now().In(location)
	startDate := now.Truncate(24 * time.Hour)
	endDate := startDate.Add(time.Duration(days) * 24 * time.Hour)

	// Fetch events
	events, err := provider.ListEvents(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	return &AvailabilityData{
		Ctx:       ctx,
		Cfg:       cfg,
		Location:  location,
		WorkHours: workHours,
		Provider:  provider,
		Events:    events,
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}
