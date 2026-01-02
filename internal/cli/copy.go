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
	"github.com/robgyiv/availability/internal/engine"
	googlecal "github.com/robgyiv/availability/internal/calendar/google"
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

	// Create calendar provider
	provider := googlecal.NewProvider()

	// Try to load existing token
	if !provider.IsAuthenticated() {
		if err := provider.LoadToken(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Not authenticated. Please run 'avail auth' first.\n")
			return err
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

