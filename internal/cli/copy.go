package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/robgyiv/availability/internal/config"
	"github.com/robgyiv/availability/pkg/engine"
	urlcal "github.com/robgyiv/availability/internal/calendar/url"
	cal "github.com/robgyiv/availability/internal/calendar"
	googlecal "github.com/robgyiv/availability/internal/calendar/google"
	localcal "github.com/robgyiv/availability/internal/calendar/local"
)

// newCopyCmd creates the copy command.
func newCopyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy availability to clipboard",
		Long:  "Copies formatted availability text to the clipboard for pasting.",
		RunE:  runCopy,
	}

	return cmd
}

func runCopy(cmd *cobra.Command, args []string) error {
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

	// Create calendar provider based on config mode
	var provider cal.Provider
	calendarMode := cfg.CalendarMode
	if calendarMode == "" {
		calendarMode = "network" // Default
	}

	if calendarMode == "local" {
		// Local mode: read from .ics file
		if cfg.LocalCalendarPath == "" {
			return fmt.Errorf("local_calendar_path is required when calendar_mode is 'local'")
		}
		provider = localcal.NewProviderFromPath(cfg.LocalCalendarPath)
	} else {
		// Network mode: use HTTP-based providers
		providerName := cfg.CalendarProvider
		if providerName == "" {
			providerName = "google" // Default
		}

		switch providerName {
		case "google":
			provider = googlecal.NewProvider()
		case "url":
			if cfg.CalendarURL != "" {
				provider = urlcal.NewProviderFromURL(cfg.CalendarURL)
			} else {
				provider = urlcal.NewProvider()
			}
		default:
			return fmt.Errorf("unknown provider: %s (supported: google, url)", providerName)
		}
	}

	// Try to load existing token
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
		cfg.BufferDuration,
	)

	// Group by day
	availability := engine.GroupBlocksByDay(blocks)

	// Format for clipboard
	var lines []string
	lines = append(lines, "I'm free:")
	for _, day := range availability {
		for _, block := range day.Blocks {
			dateStr := engine.FormatDate(day.Date, location)
			blockStr := engine.FormatTimeBlock(block, location)
			lines = append(lines, fmt.Sprintf("• %s %s", dateStr, blockStr))
		}
	}

	output := strings.Join(lines, "\n")

	// Copy to clipboard
	if err := clipboard.WriteAll(output); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	fmt.Println("Availability copied to clipboard!")
	return nil
}

