package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/robgyiv/avail/pkg/engine"
)

// newShowCmd creates the show command.
func newShowCmd() *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Display your availability",
		Long:  "Shows your availability for the next N working days (num_days_ahead in config) based on your calendar.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(days)
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 0, "Number of working days to calculate availability for (default: num_days_ahead from config)")

	return cmd
}

func runShow(days int) error {
	// Load availability data (next num_days_ahead working days, per config)
	data, err := LoadAvailabilityData(days)
	if err != nil {
		return err
	}

	// Calculate availability
	blocks := engine.CalculateAvailability(
		data.Events,
		data.StartDate,
		data.EndDate,
		data.WorkHours,
		data.Cfg.MeetingDuration,
		data.Cfg.BufferDuration,
	)
	blocks = FilterAvailabilityBlocks(blocks, data.Cfg.WeekStart, data.Cfg.WeekEnd, data.Location)

	// Group by day
	availability := engine.GroupBlocksByDay(blocks)

	// Display
	fmt.Printf("Your availability (next %d working days):\n\n", data.NumDaysAhead)
	for _, day := range availability {
		dateStr := engine.FormatDate(day.Date, data.Location)
		fmt.Printf("%s\n", dateStr)
		for _, block := range day.Blocks {
			blockStr := engine.FormatTimeBlock(block, data.Location)
			fmt.Printf("  • %s\n", blockStr)
		}
		fmt.Println()
	}

	fmt.Printf("Time zone: %s\n", data.Cfg.Timezone)

	return nil
}
