package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/robgyiv/avail/internal/api"
	cal "github.com/robgyiv/avail/internal/calendar"
	googlecal "github.com/robgyiv/avail/internal/calendar/google"
	localcal "github.com/robgyiv/avail/internal/calendar/local"
	urlcal "github.com/robgyiv/avail/internal/calendar/url"
	"github.com/robgyiv/avail/internal/config"
	"github.com/robgyiv/avail/pkg/availability"
	"github.com/robgyiv/avail/pkg/engine"
)

// newPushCmd creates the push command.
func newPushCmd() *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push availability to avail.website",
		Long:  "Calculates your availability and pushes it to avail.website API.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPush(cmd, args, days)
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 5, "Number of days to calculate availability for")

	return cmd
}

func runPush(cmd *cobra.Command, args []string, days int) error {
	ctx := context.Background()

	// Load config
	cfg, err := config.LoadOrCreate()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Load timezone
	location, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone: %w", err)
	}

	// Get work hours
	workHours, err := cfg.WorkHours()
	if err != nil {
		return fmt.Errorf("invalid work hours: %w", err)
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
			return fmt.Errorf("local_calendar_path not set in config file\n\nSet local_calendar_path in ~/.config/avail/config.toml:\n  local_calendar_path = \"/path/to/calendar.ics\"")
		}
		provider = localcal.NewProviderFromPath(cfg.LocalCalendarPath)
	default:
		return fmt.Errorf("unknown provider: %s (supported: google, network, local)", providerName)
	}

	// For providers that need authentication (google, network), load credentials from keyring
	if providerName != "local" {
		if !provider.IsAuthenticated() {
			// Try to load token (provider-specific)
			if loadable, ok := provider.(interface{ LoadToken(context.Context) error }); ok {
				if err := loadable.LoadToken(ctx); err != nil {
					fmt.Fprintf(os.Stderr, "Not authenticated. Please run 'avail auth' first.\n")
					return err
				}
			} else {
				fmt.Fprintf(os.Stderr, "Not authenticated. Please run 'avail auth' first.\n")
				return fmt.Errorf("not authenticated")
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
		return fmt.Errorf("failed to fetch events: %w", err)
	}

	// Calculate availability
	blocks := engine.CalculateAvailability(
		events,
		startDate,
		endDate,
		workHours,
		cfg.MeetingDuration,
		cfg.BufferDuration,
	)

	// Transform to API format
	apiReq := transformToAPIFormat(blocks, startDate, endDate, cfg.Timezone)

	// Load API token
	token, err := config.LoadAPIToken()
	if err != nil {
		return err
	}

	// Push to API
	client := api.NewClient()
	if err := client.PushAvailability(ctx, token, apiReq); err != nil {
		return fmt.Errorf("failed to push availability: %w", err)
	}

	fmt.Printf("✓ Availability pushed successfully (next %d days)\n", days)
	return nil
}

// transformToAPIFormat converts availability blocks to the API request format.
func transformToAPIFormat(blocks []availability.TimeBlock, startDate, endDate time.Time, timezone string) *api.UpdateAvailabilityRequest {
	// Convert time blocks to API slots
	slots := make([]api.AvailabilitySlotRequest, 0, len(blocks))
	for _, block := range blocks {
		slots = append(slots, api.AvailabilitySlotRequest{
			Start: block.Start.Format(time.RFC3339),
			End:   block.End.Format(time.RFC3339),
		})
	}

	// Create window range (ISO date format)
	window := api.WindowRange{
		Start: startDate.Format("2006-01-02"),
		End:   endDate.Format("2006-01-02"),
	}

	// Set generated_at to current time
	generatedAt := time.Now().Format(time.RFC3339)

	return &api.UpdateAvailabilityRequest{
		Slots:       slots,
		Timezone:    timezone,
		Window:      window,
		GeneratedAt: generatedAt,
	}
}
