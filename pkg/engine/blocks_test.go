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
			result := GroupBlocksByDay(tt.blocks)
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
