package engine

import (
	"testing"
	"time"

	"github.com/robgyiv/availability/pkg/availability"
)

func TestCalculateAvailability(t *testing.T) {
	location := time.UTC
	workHours := availability.WorkHours{
		Start: 9 * time.Hour, // 09:00
		End:   17 * time.Hour, // 17:00
	}
	meetingDuration := 30 * time.Minute

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
			wantCount: 2, // Morning block + afternoon block
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
			wantCount: 2, // Morning block + afternoon block (after merged events)
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
			blocks := CalculateAvailability(tt.events, tt.startDate, tt.endDate, workHours, meetingDuration)
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

	startDate := time.Date(2024, 3, 12, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 3, 13, 0, 0, 0, 0, location)

	blocks := CalculateAvailability([]availability.Event{}, startDate, endDate, workHours, meetingDuration)

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

