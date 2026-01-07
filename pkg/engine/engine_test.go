package engine

import (
	"testing"
	"time"

	"github.com/robgyiv/availability/pkg/availability"
)

func TestCalculateAvailability(t *testing.T) {
	location := time.UTC
	workHours := availability.WorkHours{
		Start: 9 * time.Hour,  // 09:00
		End:   17 * time.Hour, // 17:00
	}
	meetingDuration := 30 * time.Minute
	bufferDuration := 15 * time.Minute

	tests := []struct {
		name      string
		events    []availability.Event
		startDate time.Time
		endDate   time.Time
		wantCount int
	}{
		{
			name:      "no events",
			events:    []availability.Event{},
			startDate: time.Date(2024, 3, 12, 0, 0, 0, 0, location),
			endDate:   time.Date(2024, 3, 13, 0, 0, 0, 0, location),
			wantCount: 1, // One full day block
		},
		{
			name: "single event in middle of day",
			events: []availability.Event{
				{
					Start: time.Date(2024, 3, 12, 14, 0, 0, 0, location),
					End:   time.Date(2024, 3, 12, 15, 0, 0, 0, location),
				},
			},
			startDate: time.Date(2024, 3, 12, 0, 0, 0, 0, location),
			endDate:   time.Date(2024, 3, 13, 0, 0, 0, 0, location),
			wantCount: 2, // Morning block + afternoon block (with buffer)
		},
		{
			name: "overlapping events",
			events: []availability.Event{
				{
					Start: time.Date(2024, 3, 12, 14, 0, 0, 0, location),
					End:   time.Date(2024, 3, 12, 15, 0, 0, 0, location),
				},
				{
					Start: time.Date(2024, 3, 12, 14, 30, 0, 0, location),
					End:   time.Date(2024, 3, 12, 16, 0, 0, 0, location),
				},
			},
			startDate: time.Date(2024, 3, 12, 0, 0, 0, 0, location),
			endDate:   time.Date(2024, 3, 13, 0, 0, 0, 0, location),
			wantCount: 2, // Morning block + afternoon block (after merged events with buffer)
		},
		{
			name: "event blocks entire day",
			events: []availability.Event{
				{
					Start: time.Date(2024, 3, 12, 9, 0, 0, 0, location),
					End:   time.Date(2024, 3, 12, 17, 0, 0, 0, location),
				},
			},
			startDate: time.Date(2024, 3, 12, 0, 0, 0, 0, location),
			endDate:   time.Date(2024, 3, 13, 0, 0, 0, 0, location),
			wantCount: 0, // No free time
		},
		{
			name: "event too short for meeting",
			events: []availability.Event{
				{
					Start: time.Date(2024, 3, 12, 14, 0, 0, 0, location),
					End:   time.Date(2024, 3, 12, 16, 0, 0, 0, location),
				},
			},
			startDate: time.Date(2024, 3, 12, 0, 0, 0, 0, location),
			endDate:   time.Date(2024, 3, 13, 0, 0, 0, 0, location),
			wantCount: 2, // Morning + afternoon blocks
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := CalculateAvailability(tt.events, tt.startDate, tt.endDate, workHours, meetingDuration, bufferDuration)
			if len(blocks) != tt.wantCount {
				t.Errorf("CalculateAvailability() got %d blocks, want %d", len(blocks), tt.wantCount)
			}

			// Verify all blocks are valid (start < end, within work hours)
			for _, block := range blocks {
				if block.Start.After(block.End) || block.Start.Equal(block.End) {
					t.Errorf("Invalid block: start %v >= end %v", block.Start, block.End)
				}

				blockDuration := block.End.Sub(block.Start)
				if blockDuration < meetingDuration {
					t.Errorf("Block too short: %v < %v", blockDuration, meetingDuration)
				}
			}
		})
	}
}

func TestCalculateAvailability_RespectsWorkHours(t *testing.T) {
	location := time.UTC
	workHours := availability.WorkHours{
		Start: 9 * time.Hour,
		End:   17 * time.Hour,
	}
	meetingDuration := 30 * time.Minute
	bufferDuration := 15 * time.Minute

	startDate := time.Date(2024, 3, 12, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 3, 13, 0, 0, 0, 0, location)

	blocks := CalculateAvailability([]availability.Event{}, startDate, endDate, workHours, meetingDuration, bufferDuration)

	if len(blocks) == 0 {
		t.Fatal("Expected at least one block")
	}

	dayStart := startDate.Truncate(24 * time.Hour).Add(workHours.Start)
	dayEnd := startDate.Truncate(24 * time.Hour).Add(workHours.End)

	for _, block := range blocks {
		if block.Start.Before(dayStart) || block.Start.After(dayEnd) {
			t.Errorf("Block start %v outside work hours [%v, %v]", block.Start, dayStart, dayEnd)
		}
		if block.End.Before(dayStart) || block.End.After(dayEnd) {
			t.Errorf("Block end %v outside work hours [%v, %v]", block.End, dayStart, dayEnd)
		}
	}
}

func TestCalculateAvailability_WithBuffer(t *testing.T) {
	location := time.UTC
	workHours := availability.WorkHours{
		Start: 9 * time.Hour,
		End:   17 * time.Hour,
	}
	meetingDuration := 30 * time.Minute
	bufferDuration := 15 * time.Minute

	tests := []struct {
		name        string
		events      []availability.Event
		buffer      time.Duration
		wantStart   time.Time
		description string
	}{
		{
			name: "buffer applied after single event",
			events: []availability.Event{
				{
					Start: time.Date(2024, 3, 12, 14, 0, 0, 0, location),
					End:   time.Date(2024, 3, 12, 15, 0, 0, 0, location),
				},
			},
			buffer:      bufferDuration,
			wantStart:   time.Date(2024, 3, 12, 15, 15, 0, 0, location), // 15:00 + 15min buffer
			description: "Next availability should start 15 minutes after event ends",
		},
		{
			name: "zero buffer works like before",
			events: []availability.Event{
				{
					Start: time.Date(2024, 3, 12, 14, 0, 0, 0, location),
					End:   time.Date(2024, 3, 12, 15, 0, 0, 0, location),
				},
			},
			buffer:      0,
			wantStart:   time.Date(2024, 3, 12, 15, 0, 0, 0, location), // No buffer
			description: "Zero buffer should start immediately after event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startDate := time.Date(2024, 3, 12, 0, 0, 0, 0, location)
			endDate := time.Date(2024, 3, 13, 0, 0, 0, 0, location)

			blocks := CalculateAvailability(tt.events, startDate, endDate, workHours, meetingDuration, tt.buffer)

			// Find the block that starts at or after the event end
			var foundBlock *availability.TimeBlock
			eventEnd := tt.events[0].End
			for i := range blocks {
				// For zero buffer, block starts exactly at event end
				// For non-zero buffer, block starts after event end + buffer
				if blocks[i].Start.After(eventEnd) || (tt.buffer == 0 && blocks[i].Start.Equal(eventEnd)) {
					foundBlock = &blocks[i]
					break
				}
			}

			if foundBlock == nil {
				t.Fatalf("Expected to find a block after the event, got %d blocks total. Blocks: %v", len(blocks), blocks)
			}

			if !foundBlock.Start.Equal(tt.wantStart) {
				t.Errorf("Block start = %v, want %v (%s)", foundBlock.Start, tt.wantStart, tt.description)
			}
		})
	}
}

func TestCalculateAvailability_BackToBackEvents(t *testing.T) {
	location := time.UTC
	workHours := availability.WorkHours{
		Start: 9 * time.Hour,
		End:   17 * time.Hour,
	}
	meetingDuration := 30 * time.Minute
	bufferDuration := 15 * time.Minute

	// Two events back-to-back with buffer should eliminate the gap
	events := []availability.Event{
		{
			Start: time.Date(2024, 3, 12, 14, 0, 0, 0, location),
			End:   time.Date(2024, 3, 12, 15, 0, 0, 0, location),
		},
		{
			Start: time.Date(2024, 3, 12, 15, 0, 0, 0, location), // Starts exactly when first ends
			End:   time.Date(2024, 3, 12, 16, 0, 0, 0, location),
		},
	}

	startDate := time.Date(2024, 3, 12, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 3, 13, 0, 0, 0, 0, location)

	blocks := CalculateAvailability(events, startDate, endDate, workHours, meetingDuration, bufferDuration)

	// Should have morning block and afternoon block (after 16:15 with buffer)
	// But the gap between 15:00 and 16:00 should be eliminated due to buffer
	if len(blocks) < 2 {
		t.Errorf("Expected at least 2 blocks, got %d", len(blocks))
	}

	// Check that there's no block between 15:00 and 16:15
	for _, block := range blocks {
		if block.Start.After(time.Date(2024, 3, 12, 15, 0, 0, 0, location)) &&
			block.Start.Before(time.Date(2024, 3, 12, 16, 15, 0, 0, location)) {
			t.Errorf("Unexpected block found between events: %v-%v", block.Start, block.End)
		}
	}
}
