package cli

import (
	"context"
	"fmt"
	"time"

	cal "github.com/robgyiv/avail/internal/calendar"
	"github.com/robgyiv/avail/internal/calendar/aggregate"
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

	// Create aggregate provider from all configured calendars
	aggProvider, err := aggregate.NewProvider(cfg.Calendars)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize calendars: %w", err)
	}

	// Print any warnings about calendars that failed to load
	aggProvider.PrintWarnings()

	var provider cal.Provider = aggProvider

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

func FilterAvailabilityBlocks(blocks []availability.TimeBlock, includeWeekends bool, location *time.Location) []availability.TimeBlock {
	if includeWeekends {
		return blocks
	}

	filtered := make([]availability.TimeBlock, 0, len(blocks))
	for _, block := range blocks {
		weekday := block.Start.In(location).Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			continue
		}
		filtered = append(filtered, block)
	}

	return filtered
}
