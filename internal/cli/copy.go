package cli

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/robgyiv/avail/pkg/engine"
)

// newCopyCmd creates the copy command.
func newCopyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy availability to clipboard",
		Long:  "Copies formatted availability text to the clipboard for pasting.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCopy()
		},
	}

	return cmd
}

func runCopy() error {
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

	// Format for clipboard
	var lines []string
	lines = append(lines, "I'm free:")
	for _, day := range availability {
		for _, block := range day.Blocks {
			dateStr := engine.FormatDate(day.Date, data.Location)
			blockStr := engine.FormatTimeBlock(block, data.Location)
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
