package engine

import (
	"testing"
	"time"

	"github.com/robgyiv/avail/pkg/availability"
)

func TestGroupBlocksByDay(t *testing.T) {
	location := time.UTC

	tests := []struct {
		name   string
		blocks []availability.TimeBlock
		want   int // Number of days
	}{
		{
			name:   "empty blocks",
			blocks: []availability.TimeBlock{},
			want:   0,
		},
		{
			name: "single day",
			blocks: []availability.TimeBlock{
				{
					Start: time.Date(2024, 3, 12, 9, 0, 0, 0, location),
					End:   time.Date(2024, 3, 12, 10, 0, 0, 0, location),
				},
				{
					Start: time.Date(2024, 3, 12, 14, 0, 0, 0, location),
					End:   time.Date(2024, 3, 12, 15, 0, 0, 0, location),
				},
			},
			want: 1,
		},
		{
			name: "multiple days",
			blocks: []availability.TimeBlock{
				{
					Start: time.Date(2024, 3, 12, 9, 0, 0, 0, location),
					End:   time.Date(2024, 3, 12, 10, 0, 0, 0, location),
				},
				{
					Start: time.Date(2024, 3, 13, 14, 0, 0, 0, location),
					End:   time.Date(2024, 3, 13, 15, 0, 0, 0, location),
				},
				{
					Start: time.Date(2024, 3, 15, 10, 0, 0, 0, location),
					End:   time.Date(2024, 3, 15, 11, 0, 0, 0, location),
				},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GroupBlocksByDay(tt.blocks, location)
			if len(result) != tt.want {
				t.Errorf("GroupBlocksByDay() got %d days, want %d", len(result), tt.want)
			}

			// Verify blocks are grouped correctly
			totalBlocks := 0
			for _, day := range result {
				totalBlocks += len(day.Blocks)
				for _, block := range day.Blocks {
					blockDate := block.Start.Truncate(24 * time.Hour)
					if !blockDate.Equal(day.Date.Truncate(24 * time.Hour)) {
						t.Errorf("Block on wrong day: block date %v, day date %v", blockDate, day.Date)
					}
				}
			}

			if totalBlocks != len(tt.blocks) {
				t.Errorf("Lost blocks: got %d total blocks, want %d", totalBlocks, len(tt.blocks))
			}
		})
	}
}

// TestGroupBlocksByDay_MixedLocations covers blocks from one day whose boundaries
// carry different locations, including two separately loaded copies of the same
// zone. They belong to a single day, so grouping must key on the calendar date in
// the display location rather than compare time.Time values directly.
func TestGroupBlocksByDay_MixedLocations(t *testing.T) {
	melbourne, err := time.LoadLocation("Australia/Melbourne")
	if err != nil {
		t.Skipf("Australia/Melbourne unavailable: %v", err)
	}
	// A second load yields an equivalent zone with a different *Location pointer,
	// which is what the calendar parser produces for a TZID-qualified event.
	otherMelbourne, err := time.LoadLocation("Australia/Melbourne")
	if err != nil {
		t.Skipf("Australia/Melbourne unavailable: %v", err)
	}

	blocks := []availability.TimeBlock{
		{
			// 09:00-09:50 Melbourne, morning edge from the configured work hours.
			Start: time.Date(2026, 9, 25, 9, 0, 0, 0, melbourne),
			End:   time.Date(2026, 9, 25, 9, 50, 0, 0, otherMelbourne),
		},
		{
			// 11:40-17:00 the same day, expressed as UTC instants.
			Start: time.Date(2026, 9, 25, 1, 40, 0, 0, time.UTC),
			End:   time.Date(2026, 9, 25, 7, 0, 0, 0, time.UTC),
		},
	}

	result := GroupBlocksByDay(blocks, melbourne)

	if len(result) != 1 {
		t.Fatalf("GroupBlocksByDay() got %d days, want 1 (same Melbourne day)", len(result))
	}
	if got := len(result[0].Blocks); got != 2 {
		t.Errorf("GroupBlocksByDay() got %d blocks in the day, want 2", got)
	}
	if got, want := result[0].Date.In(melbourne).Format("2006-01-02"), "2026-09-25"; got != want {
		t.Errorf("GroupBlocksByDay() day = %s, want %s", got, want)
	}
}

// TestGroupBlocksByDay_SplitsAcrossDaysInLocation checks that grouping follows the
// display zone: two instants 6 hours apart can be the same UTC day but different
// Melbourne days.
func TestGroupBlocksByDay_SplitsAcrossDaysInLocation(t *testing.T) {
	melbourne, err := time.LoadLocation("Australia/Melbourne")
	if err != nil {
		t.Skipf("Australia/Melbourne unavailable: %v", err)
	}

	blocks := []availability.TimeBlock{
		// 2026-09-24T22:00Z is 08:00 on the 25th in Melbourne.
		{
			Start: time.Date(2026, 9, 24, 22, 0, 0, 0, time.UTC),
			End:   time.Date(2026, 9, 24, 23, 0, 0, 0, time.UTC),
		},
		// 2026-09-24T04:00Z is 14:00 on the 24th in Melbourne.
		{
			Start: time.Date(2026, 9, 24, 4, 0, 0, 0, time.UTC),
			End:   time.Date(2026, 9, 24, 5, 0, 0, 0, time.UTC),
		},
	}

	if got := len(GroupBlocksByDay(blocks, melbourne)); got != 2 {
		t.Errorf("GroupBlocksByDay(melbourne) got %d days, want 2", got)
	}
	if got := len(GroupBlocksByDay(blocks, time.UTC)); got != 1 {
		t.Errorf("GroupBlocksByDay(UTC) got %d days, want 1", got)
	}
}
