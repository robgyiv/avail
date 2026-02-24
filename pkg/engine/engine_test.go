package engine

import (
	"testing"
	"time"

	"github.com/robgyiv/avail/pkg/availability"
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

func TestCalculateAvailability_BufferBeforeEvent(t *testing.T) {
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
		wantEnd     time.Time
		description string
	}{
		{
			name: "buffer applied before single event",
			events: []availability.Event{
				{
					Start: time.Date(2024, 3, 12, 14, 0, 0, 0, location),
					End:   time.Date(2024, 3, 12, 15, 0, 0, 0, location),
				},
			},
			buffer:      bufferDuration,
			wantEnd:     time.Date(2024, 3, 12, 13, 45, 0, 0, location), // 14:00 - 15min buffer
			description: "Last availability should end 15 minutes before event starts",
		},
		{
			name: "zero buffer allows availability right before event",
			events: []availability.Event{
				{
					Start: time.Date(2024, 3, 12, 14, 0, 0, 0, location),
					End:   time.Date(2024, 3, 12, 15, 0, 0, 0, location),
				},
			},
			buffer:      0,
			wantEnd:     time.Date(2024, 3, 12, 14, 0, 0, 0, location), // No buffer
			description: "Zero buffer should allow availability right up to event start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startDate := time.Date(2024, 3, 12, 0, 0, 0, 0, location)
			endDate := time.Date(2024, 3, 13, 0, 0, 0, 0, location)

			blocks := CalculateAvailability(tt.events, startDate, endDate, workHours, meetingDuration, tt.buffer)

			// Find the block that ends before the event start
			var foundBlock *availability.TimeBlock
			eventStart := tt.events[0].Start
			for i := range blocks {
				if blocks[i].End.Before(eventStart) || (tt.buffer == 0 && blocks[i].End.Equal(eventStart)) {
					// Keep the last block that ends before/at the event
					foundBlock = &blocks[i]
				}
			}

			if foundBlock == nil {
				t.Fatalf("Expected to find a block before the event, got %d blocks total. Blocks: %v", len(blocks), blocks)
			}

			if !foundBlock.End.Equal(tt.wantEnd) {
				t.Errorf("Block end = %v, want %v (%s)", foundBlock.End, tt.wantEnd, tt.description)
			}
		})
	}
}

func TestCalculateAvailability_BufferBothBeforeAndAfter(t *testing.T) {
	location := time.UTC
	workHours := availability.WorkHours{
		Start: 9 * time.Hour,
		End:   17 * time.Hour,
	}
	meetingDuration := 30 * time.Minute
	bufferDuration := 15 * time.Minute

	// Single event with gaps large enough for meetings on both sides
	events := []availability.Event{
		{
			Start: time.Date(2024, 3, 12, 12, 0, 0, 0, location),
			End:   time.Date(2024, 3, 12, 13, 0, 0, 0, location),
		},
	}

	startDate := time.Date(2024, 3, 12, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 3, 13, 0, 0, 0, 0, location)

	blocks := CalculateAvailability(events, startDate, endDate, workHours, meetingDuration, bufferDuration)

	// Should have blocks:
	// 1. Morning: 09:00-11:45 (ends 15min before event at 12:00)
	// 2. Afternoon: 13:15-17:00 (starts 15min after event ends at 13:00)
	if len(blocks) != 2 {
		t.Errorf("Expected 2 blocks, got %d. Blocks: %v", len(blocks), blocks)
		return
	}

	// Check morning block ends before buffer
	expectedMorningEnd := time.Date(2024, 3, 12, 11, 45, 0, 0, location)
	if !blocks[0].End.Equal(expectedMorningEnd) {
		t.Errorf("Morning block end = %v, want %v", blocks[0].End, expectedMorningEnd)
	}

	// Check afternoon block starts after buffer
	expectedAfternoonStart := time.Date(2024, 3, 12, 13, 15, 0, 0, location)
	if !blocks[1].Start.Equal(expectedAfternoonStart) {
		t.Errorf("Afternoon block start = %v, want %v", blocks[1].Start, expectedAfternoonStart)
	}
}

func TestCalculateAvailability_BufferNotExtendingBeyondWorkHours(t *testing.T) {
	location := time.UTC
	workHours := availability.WorkHours{
		Start: 10 * time.Hour, // 10:00
		End:   16 * time.Hour, // 16:00
	}
	meetingDuration := 30 * time.Minute
	bufferDuration := 10 * time.Minute

	// Event ends at 15:50, buffer would extend to 16:00 (end of work day)
	events := []availability.Event{
		{
			Start: time.Date(2024, 3, 12, 15, 50, 0, 0, location),
			End:   time.Date(2024, 3, 12, 15, 50, 0, 0, location),
		},
	}

	startDate := time.Date(2024, 3, 12, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 3, 13, 0, 0, 0, 0, location)

	blocks := CalculateAvailability(events, startDate, endDate, workHours, meetingDuration, bufferDuration)

	// Verify no block extends beyond work hours
	dayEnd := startDate.Truncate(24 * time.Hour).Add(workHours.End)
	for _, block := range blocks {
		if block.End.After(dayEnd) {
			t.Errorf("Block extends beyond work hours: %v (end at %v, should be <= %v)", block, block.End, dayEnd)
		}
	}
}

func TestCalculateAvailability_MeetingAtEndOfDay(t *testing.T) {
	location := time.UTC
	workHours := availability.WorkHours{
		Start: 10 * time.Hour, // 10:00
		End:   16 * time.Hour, // 16:00
	}
	meetingDuration := 30 * time.Minute
	bufferDuration := 15 * time.Minute

	// Meeting ends at 15:50 with 15min buffer would extend to 16:05, but should be clamped to 16:00
	events := []availability.Event{
		{
			Start: time.Date(2024, 3, 12, 15, 35, 0, 0, location),
			End:   time.Date(2024, 3, 12, 15, 50, 0, 0, location),
		},
	}

	startDate := time.Date(2024, 3, 12, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 3, 13, 0, 0, 0, 0, location)

	blocks := CalculateAvailability(events, startDate, endDate, workHours, meetingDuration, bufferDuration)

	// Should have morning block ending before 15:35 minus 15min buffer = 15:20
	if len(blocks) < 1 {
		t.Fatalf("Expected at least 1 block, got %d", len(blocks))
	}

	// Last block should not extend past work hours
	dayEnd := startDate.Truncate(24 * time.Hour).Add(workHours.End)
	lastBlock := blocks[len(blocks)-1]
	if lastBlock.End.After(dayEnd) {
		t.Errorf("Last block end %v exceeds work hours end %v", lastBlock.End, dayEnd)
	}
}

func TestCalculateAvailability_MeetingExtendsAfterWorkHours(t *testing.T) {
	location := time.UTC
	workHours := availability.WorkHours{
		Start: 10 * time.Hour, // 10:00
		End:   16 * time.Hour, // 16:00
	}
	meetingDuration := 30 * time.Minute
	bufferDuration := 15 * time.Minute

	// Meeting extends past work hours: 15:45 to 16:15 (15 min past end)
	events := []availability.Event{
		{
			Start: time.Date(2024, 3, 12, 15, 45, 0, 0, location),
			End:   time.Date(2024, 3, 12, 16, 15, 0, 0, location),
		},
	}

	startDate := time.Date(2024, 3, 12, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 3, 13, 0, 0, 0, 0, location)

	blocks := CalculateAvailability(events, startDate, endDate, workHours, meetingDuration, bufferDuration)

	// Verify no block extends past work hours
	dayEnd := startDate.Truncate(24 * time.Hour).Add(workHours.End)
	for _, block := range blocks {
		if block.End.After(dayEnd) {
			t.Errorf("Block end %v exceeds work hours end %v", block.End, dayEnd)
		}
	}

	// Should have morning block ending at 15:30 (15min buffer before 15:45 event)
	if len(blocks) < 1 {
		t.Errorf("Expected at least 1 block, got %d", len(blocks))
	} else {
		expectedEnd := time.Date(2024, 3, 12, 15, 30, 0, 0, location)
		if !blocks[0].End.Equal(expectedEnd) {
			t.Errorf("Morning block end = %v, want %v", blocks[0].End, expectedEnd)
		}
	}
}

func TestCalculateAvailability_AllBlocksRespectWorkHours(t *testing.T) {
	location := time.UTC
	workHours := availability.WorkHours{
		Start: 10 * time.Hour, // 10:00
		End:   16 * time.Hour, // 16:00
	}
	meetingDuration := 30 * time.Minute
	bufferDuration := 10 * time.Minute

	// Multiple events throughout the day
	events := []availability.Event{
		{
			Start: time.Date(2024, 3, 12, 11, 0, 0, 0, location),
			End:   time.Date(2024, 3, 12, 12, 0, 0, 0, location),
		},
		{
			Start: time.Date(2024, 3, 12, 14, 0, 0, 0, location),
			End:   time.Date(2024, 3, 12, 15, 30, 0, 0, location),
		},
		{
			Start: time.Date(2024, 3, 12, 15, 45, 0, 0, location),
			End:   time.Date(2024, 3, 12, 16, 00, 0, 0, location),
		},
	}

	startDate := time.Date(2024, 3, 12, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 3, 13, 0, 0, 0, 0, location)

	blocks := CalculateAvailability(events, startDate, endDate, workHours, meetingDuration, bufferDuration)

	// Verify NO block starts before dayStart or ends after dayEnd
	dayStart := startDate.Truncate(24 * time.Hour).Add(workHours.Start)
	dayEnd := startDate.Truncate(24 * time.Hour).Add(workHours.End)

	for i, block := range blocks {
		if block.Start.Before(dayStart) {
			t.Errorf("Block %d start %v is before work hours start %v", i, block.Start, dayStart)
		}
		if block.End.After(dayEnd) {
			t.Errorf("Block %d end %v exceeds work hours end %v", i, block.End, dayEnd)
		}
	}
}

func TestCalculateAvailability_EventAfterWorkHours(t *testing.T) {
	location := time.UTC
	workHours := availability.WorkHours{
		Start: 10 * time.Hour, // 10:00
		End:   16 * time.Hour, // 16:00
	}
	meetingDuration := 30 * time.Minute
	bufferDuration := 10 * time.Minute

	// Event at 16:30 (after work hours end at 16:00)
	// Should not affect availability calculation within work hours
	events := []availability.Event{
		{
			Start: time.Date(2024, 3, 12, 16, 30, 0, 0, location),
			End:   time.Date(2024, 3, 12, 17, 30, 0, 0, location),
		},
	}

	startDate := time.Date(2024, 3, 12, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 3, 13, 0, 0, 0, 0, location)

	blocks := CalculateAvailability(events, startDate, endDate, workHours, meetingDuration, bufferDuration)

	// Should have one full-day availability block (10:00 to 16:00)
	if len(blocks) != 1 {
		t.Errorf("Expected 1 block, got %d", len(blocks))
		return
	}

	dayStart := startDate.Truncate(24 * time.Hour).Add(workHours.Start)
	dayEnd := startDate.Truncate(24 * time.Hour).Add(workHours.End)

	if !blocks[0].Start.Equal(dayStart) {
		t.Errorf("Block start = %v, want %v", blocks[0].Start, dayStart)
	}
	if !blocks[0].End.Equal(dayEnd) {
		t.Errorf("Block end = %v, want %v", blocks[0].End, dayEnd)
	}
}
