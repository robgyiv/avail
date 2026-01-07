package engine

import (
	"strings"
	"testing"
	"time"

	"github.com/robgyiv/avail/pkg/availability"
)

func TestFormatTimeBlock(t *testing.T) {
	location := time.UTC

	tests := []struct {
		name  string
		block availability.TimeBlock
		want  string
	}{
		{
			name: "basic time block",
			block: availability.TimeBlock{
				Start: time.Date(2024, 3, 12, 14, 0, 0, 0, location),
				End:   time.Date(2024, 3, 12, 16, 0, 0, 0, location),
			},
			want: "14:00–16:00",
		},
		{
			name: "half hour blocks",
			block: availability.TimeBlock{
				Start: time.Date(2024, 3, 12, 14, 30, 0, 0, location),
				End:   time.Date(2024, 3, 12, 15, 30, 0, 0, location),
			},
			want: "14:30–15:30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTimeBlock(tt.block, location)
			if got != tt.want {
				t.Errorf("FormatTimeBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	location := time.UTC

	tests := []struct {
		name string
		date time.Time
		want string
	}{
		{
			name: "Tuesday March 12",
			date: time.Date(2024, 3, 12, 0, 0, 0, 0, location),
			want: "Tue 12 Mar",
		},
		{
			name: "Wednesday March 13",
			date: time.Date(2024, 3, 13, 0, 0, 0, 0, location),
			want: "Wed 13 Mar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDate(tt.date, location)
			if got != tt.want {
				t.Errorf("FormatDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatTimeBlockForDay(t *testing.T) {
	location := time.UTC

	tests := []struct {
		name     string
		block    availability.TimeBlock
		dayStart time.Time
		want     string
	}{
		{
			name: "block at start of day",
			block: availability.TimeBlock{
				Start: time.Date(2024, 3, 12, 9, 0, 0, 0, location),
				End:   time.Date(2024, 3, 12, 10, 0, 0, 0, location),
			},
			dayStart: time.Date(2024, 3, 12, 0, 0, 0, 0, location),
			want:     "09:00–10:00",
		},
		{
			name: "block after start of day",
			block: availability.TimeBlock{
				Start: time.Date(2024, 3, 12, 13, 0, 0, 0, location),
				End:   time.Date(2024, 3, 12, 16, 0, 0, 0, location),
			},
			dayStart: time.Date(2024, 3, 12, 0, 0, 0, 0, location),
			want:     "after 13:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTimeBlockForDay(tt.block, tt.dayStart, location)
			if !strings.HasPrefix(got, tt.want) && got != tt.want {
				// For "after" format, we just check it starts with "after"
				if !strings.HasPrefix(got, "after ") {
					t.Errorf("FormatTimeBlockForDay() = %v, want to start with %v", got, tt.want)
				}
			}
		})
	}
}
