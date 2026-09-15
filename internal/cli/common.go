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
	Ctx          context.Context
	Cfg          *config.Config
	Location     *time.Location
	WorkHours    availability.WorkHours
	Provider     cal.Provider
	Events       []availability.Event
	StartDate    time.Time
	EndDate      time.Time
	NumDaysAhead int
}

// LoadAvailabilityData loads configuration, creates a calendar provider, authenticates,
// and fetches events covering the next numDaysAhead working days (as defined by the
// config's week_start/week_end). If numDaysAhead is 0, the config's num_days_ahead is used.
func LoadAvailabilityData(numDaysAhead int) (*AvailabilityData, error) {
	ctx := context.Background()

	// Load config
	cfg, err := config.LoadOrCreate()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if numDaysAhead <= 0 {
		numDaysAhead = cfg.NumDaysAhead
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

	// Calculate date range. endDate is walked forward far enough to cover
	// numDaysAhead working days, not just numDaysAhead calendar days, so that
	// e.g. running on a Thursday with a Mon-Fri week still shows a full working week.
	now := time.Now().In(location)
	startDate := now.Truncate(24 * time.Hour)
	endDate := calculateEndDate(startDate, numDaysAhead, cfg.WeekStart, cfg.WeekEnd)

	// Fetch events
	events, err := provider.ListEvents(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	return &AvailabilityData{
		Ctx:          ctx,
		Cfg:          cfg,
		Location:     location,
		WorkHours:    workHours,
		Provider:     provider,
		Events:       events,
		StartDate:    startDate,
		EndDate:      endDate,
		NumDaysAhead: numDaysAhead,
	}, nil
}

// calculateEndDate walks forward from startDate (inclusive), counting only days whose
// ISO weekday falls within [weekStart, weekEnd], until numWorkingDays such days have
// been counted. It returns the (exclusive) end of that range.
func calculateEndDate(startDate time.Time, numWorkingDays, weekStart, weekEnd int) time.Time {
	current := startDate
	counted := 0
	for counted < numWorkingDays {
		if inWeekRange(isoWeekday(current.Weekday()), weekStart, weekEnd) {
			counted++
		}
		current = current.Add(24 * time.Hour)
	}
	return current
}

// isoWeekday converts a time.Weekday to ISO 8601 weekday numbering (Monday=1 .. Sunday=7).
func isoWeekday(w time.Weekday) int {
	if w == time.Sunday {
		return 7
	}
	return int(w)
}

// inWeekRange reports whether the ISO weekday d falls within [start, end]. The range
// wraps around the week when start > end (e.g. start=6, end=2 covers Sat, Sun, Mon).
func inWeekRange(d, start, end int) bool {
	if start <= end {
		return d >= start && d <= end
	}
	return d >= start || d <= end
}

// AvailabilityJSON is the JSON representation of a set of availability days.
type AvailabilityJSON struct {
	GeneratedAt string            `json:"generated_at"`
	Timezone    string            `json:"timezone"`
	Days        []AvailabilityDay `json:"days"`
}

// AvailabilityDay is the JSON representation of a single day's availability.
type AvailabilityDay struct {
	Date   string              `json:"date"`
	Blocks []AvailabilityBlock `json:"blocks"`
}

// AvailabilityBlock is the JSON representation of a single available time block.
type AvailabilityBlock struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// BuildAvailabilityJSON converts grouped availability into a JSON-serializable
// structure, formatting dates and times in the given location using RFC 3339.
func BuildAvailabilityJSON(days []availability.Availability, location *time.Location, timezone string) AvailabilityJSON {
	out := AvailabilityJSON{
		GeneratedAt: time.Now().In(location).Format(time.RFC3339),
		Timezone:    timezone,
		Days:        make([]AvailabilityDay, 0, len(days)),
	}

	for _, day := range days {
		blocks := make([]AvailabilityBlock, 0, len(day.Blocks))
		for _, block := range day.Blocks {
			blocks = append(blocks, AvailabilityBlock{
				Start: block.Start.In(location).Format(time.RFC3339),
				End:   block.End.In(location).Format(time.RFC3339),
			})
		}

		out.Days = append(out.Days, AvailabilityDay{
			Date:   day.Date.In(location).Format("2006-01-02"),
			Blocks: blocks,
		})
	}

	return out
}

// FilterAvailabilityBlocks keeps only the blocks whose start day falls within the
// [weekStart, weekEnd] ISO weekday range (1=Monday .. 7=Sunday).
func FilterAvailabilityBlocks(blocks []availability.TimeBlock, weekStart, weekEnd int, location *time.Location) []availability.TimeBlock {
	filtered := make([]availability.TimeBlock, 0, len(blocks))
	for _, block := range blocks {
		if !inWeekRange(isoWeekday(block.Start.In(location).Weekday()), weekStart, weekEnd) {
			continue
		}
		filtered = append(filtered, block)
	}

	return filtered
}
