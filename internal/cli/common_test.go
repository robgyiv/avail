package cli

import (
	"testing"
	"time"

	"github.com/robgyiv/avail/pkg/availability"
)

func TestCalculateEndDate_ThursdayMonToFriShowsFiveWorkingDays(t *testing.T) {
	// Regression test: running on a Thursday with the default Mon-Fri week
	// should look far enough ahead to cover 5 working days (Thu, Fri, Mon,
	// Tue, Wed), not just 5 calendar days (which would cut off after Mon).
	location, _ := time.LoadLocation("UTC")
	thursday := time.Date(2026, 9, 3, 0, 0, 0, 0, location) // a Thursday
	if thursday.Weekday() != time.Thursday {
		t.Fatalf("test setup error: %v is not a Thursday", thursday)
	}

	end := calculateEndDate(thursday, 5, 1, 5)

	// Walk the [thursday, end) range and count working days.
	var workingDays []time.Weekday
	for d := thursday; d.Before(end); d = d.Add(24 * time.Hour) {
		wd := d.Weekday()
		if wd != time.Saturday && wd != time.Sunday {
			workingDays = append(workingDays, wd)
		}
	}

	if len(workingDays) != 5 {
		t.Fatalf("calculateEndDate() covered %d working days, want 5 (got %v)", len(workingDays), workingDays)
	}

	want := []time.Weekday{time.Thursday, time.Friday, time.Monday, time.Tuesday, time.Wednesday}
	for i, wd := range want {
		if workingDays[i] != wd {
			t.Errorf("workingDays[%d] = %v, want %v", i, workingDays[i], wd)
		}
	}
}

func TestCalculateEndDate_SkipsWeekends(t *testing.T) {
	location, _ := time.LoadLocation("UTC")
	monday := time.Date(2026, 8, 31, 0, 0, 0, 0, location)
	if monday.Weekday() != time.Monday {
		t.Fatalf("test setup error: %v is not a Monday", monday)
	}

	end := calculateEndDate(monday, 5, 1, 5)
	want := monday.Add(5 * 24 * time.Hour) // Mon-Fri, no weekend in this span
	if !end.Equal(want) {
		t.Errorf("calculateEndDate() = %v, want %v", end, want)
	}

	// A 6th working day (Monday of the following week) should skip the weekend.
	end = calculateEndDate(monday, 6, 1, 5)
	want = monday.Add(8 * 24 * time.Hour) // Mon-Fri + Sat/Sun skipped + next Mon
	if !end.Equal(want) {
		t.Errorf("calculateEndDate() = %v, want %v", end, want)
	}
}

func TestCalculateEndDate_FullWeekIncludesWeekends(t *testing.T) {
	location, _ := time.LoadLocation("UTC")
	monday := time.Date(2026, 8, 31, 0, 0, 0, 0, location)

	end := calculateEndDate(monday, 5, 1, 7)
	want := monday.Add(5 * 24 * time.Hour)
	if !end.Equal(want) {
		t.Errorf("calculateEndDate() = %v, want %v", end, want)
	}
}

func TestInWeekRange(t *testing.T) {
	tests := []struct {
		name       string
		d          int
		start, end int
		want       bool
	}{
		{"monday in mon-fri", 1, 1, 5, true},
		{"friday in mon-fri", 5, 1, 5, true},
		{"saturday not in mon-fri", 6, 1, 5, false},
		{"sunday not in mon-fri", 7, 1, 5, false},
		{"monday in full week", 1, 1, 7, true},
		{"sunday in full week", 7, 1, 7, true},
		{"wraparound saturday in sat-tue", 6, 6, 2, true},
		{"wraparound monday in sat-tue", 1, 6, 2, true},
		{"wraparound wednesday not in sat-tue", 3, 6, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inWeekRange(tt.d, tt.start, tt.end); got != tt.want {
				t.Errorf("inWeekRange(%d, %d, %d) = %v, want %v", tt.d, tt.start, tt.end, got, tt.want)
			}
		})
	}
}

func TestFilterAvailabilityBlocks_MonToFri(t *testing.T) {
	location, _ := time.LoadLocation("UTC")

	blocks := []availability.TimeBlock{
		{Start: time.Date(2026, 8, 31, 9, 0, 0, 0, location), End: time.Date(2026, 8, 31, 10, 0, 0, 0, location)}, // Monday
		{Start: time.Date(2026, 9, 5, 9, 0, 0, 0, location), End: time.Date(2026, 9, 5, 10, 0, 0, 0, location)},   // Saturday
		{Start: time.Date(2026, 9, 6, 9, 0, 0, 0, location), End: time.Date(2026, 9, 6, 10, 0, 0, 0, location)},   // Sunday
	}

	filtered := FilterAvailabilityBlocks(blocks, 1, 5, location)

	if len(filtered) != 1 {
		t.Fatalf("FilterAvailabilityBlocks() returned %d blocks, want 1", len(filtered))
	}
	if filtered[0].Start.Weekday() != time.Monday {
		t.Errorf("FilterAvailabilityBlocks() kept %v, want Monday block", filtered[0].Start.Weekday())
	}
}

func TestFilterAvailabilityBlocks_FullWeekKeepsWeekends(t *testing.T) {
	location, _ := time.LoadLocation("UTC")

	blocks := []availability.TimeBlock{
		{Start: time.Date(2026, 9, 5, 9, 0, 0, 0, location), End: time.Date(2026, 9, 5, 10, 0, 0, 0, location)}, // Saturday
	}

	filtered := FilterAvailabilityBlocks(blocks, 1, 7, location)

	if len(filtered) != 1 {
		t.Fatalf("FilterAvailabilityBlocks() returned %d blocks, want 1", len(filtered))
	}
}
