package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/robgyiv/avail/internal/api"
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
			return runPush(days)
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 5, "Number of days to calculate availability for")

	return cmd
}

func runPush(days int) error {
	// Load availability data
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
	blocks = FilterAvailabilityBlocks(blocks, data.Cfg.IncludeWeekends, data.Location)

	// Transform to API format
	apiReq := transformToAPIFormat(blocks, data.StartDate, data.EndDate, data.Cfg.Timezone)

	// Load API token
	token, err := config.LoadAPIToken()
	if err != nil {
		return err
	}

	// Push to API
	client := api.NewClient()
	if err := client.PushAvailability(data.Ctx, token, apiReq); err != nil {
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
