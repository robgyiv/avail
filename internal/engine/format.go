package engine

import (
	"fmt"
	"time"

	"github.com/robgyiv/availability/pkg/availability"
)

// FormatTimeBlock formats a time block as a human-readable string.
// Examples: "14:00–16:00", "after 13:00"
func FormatTimeBlock(block availability.TimeBlock, location *time.Location) string {
	start := block.Start.In(location)
	end := block.End.In(location)

	startStr := formatTime(start)
	endStr := formatTime(end)

	return fmt.Sprintf("%s–%s", startStr, endStr)
}

// FormatTimeBlockForDay formats a time block relative to a day, using "after" notation
// when the block starts after the beginning of work hours.
func FormatTimeBlockForDay(block availability.TimeBlock, dayStart time.Time, location *time.Location) string {
	start := block.Start.In(location)

	dayStartTime := dayStart.In(location).Truncate(24 * time.Hour)
	blockStartTime := start.Truncate(24 * time.Hour)

	// If the block starts at the beginning of the day (or very close), use normal format
	if blockStartTime.Equal(dayStartTime) || start.Sub(dayStartTime) < time.Hour {
		return FormatTimeBlock(block, location)
	}

	// Otherwise, use "after" notation
	startStr := formatTime(start)
	return fmt.Sprintf("after %s", startStr)
}

// FormatDate formats a date in a human-readable way.
// Example: "Tue 12 Mar"
func FormatDate(date time.Time, location *time.Location) string {
	d := date.In(location)
	return d.Format("Mon 2 Jan")
}

// formatTime formats a time as HH:MM.
func formatTime(t time.Time) string {
	return t.Format("15:04")
}

