package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/robgyiv/avail/pkg/engine"
)

// newShowCmd creates the show command.
func newShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Display your availability",
		Long:  "Shows your availability for the next 5 days based on your calendar.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow()
		},
	}

	return cmd
}

func runShow() error {
	// Load availability data (next 5 days)
	data, err := LoadAvailabilityData(5)
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

	// Group by day
	availability := engine.GroupBlocksByDay(blocks)

	// Display
	fmt.Printf("Your availability (next 5 days):\n\n")
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
