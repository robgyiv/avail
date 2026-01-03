package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/robgyiv/availability/internal/config"
	"github.com/robgyiv/availability/internal/engine"
	applecal "github.com/robgyiv/availability/internal/calendar/apple"
	cal "github.com/robgyiv/availability/internal/calendar"
	googlecal "github.com/robgyiv/availability/internal/calendar/google"
)

// newShowCmd creates the show command.
func newShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Display your availability",
		Long:  "Shows your availability for the next 5 days based on your calendar.",
		RunE:  runShow,
	}

	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
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
	providerName := cfg.CalendarProvider
	if providerName == "" {
		providerName = "google" // Default
	}

	var provider cal.Provider
	switch providerName {
	case "google":
		provider = googlecal.NewProvider()
	case "apple", "icloud":
		provider = applecal.NewProvider()
	default:
		return fmt.Errorf("unknown provider: %s (supported: google, apple)", providerName)
	}

	// Try to load existing token, otherwise authenticate
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

	// Calculate date range (next 5 days)
	now := time.Now().In(location)
	startDate := now.Truncate(24 * time.Hour)
	endDate := startDate.Add(5 * 24 * time.Hour)

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
	)

	// Group by day
	availability := engine.GroupBlocksByDay(blocks)

	// Display
	fmt.Printf("Your availability (next 5 days):\n\n")
	for _, day := range availability {
		dateStr := engine.FormatDate(day.Date, location)
		fmt.Printf("%s\n", dateStr)
		for _, block := range day.Blocks {
			blockStr := engine.FormatTimeBlock(block, location)
			fmt.Printf("  • %s\n", blockStr)
		}
		fmt.Println()
	}

	fmt.Printf("Time zone: %s\n", cfg.Timezone)

	return nil
}

