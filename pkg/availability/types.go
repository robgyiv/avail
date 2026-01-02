package availability

import "time"

// TimeBlock represents a contiguous block of available time.
type TimeBlock struct {
	Start time.Time
	End   time.Time
}

// Availability represents available time blocks for a specific date.
type Availability struct {
	Date  time.Time
	Blocks []TimeBlock
}

// Event represents a calendar event that blocks availability.
type Event struct {
	Start time.Time
	End   time.Time
	// Title is optional, used for debugging/logging
	Title string
	// AllDay indicates if this is an all-day event
	AllDay bool
}

// WorkHours defines the working hours for availability calculation.
type WorkHours struct {
	Start time.Duration // e.g., 9*time.Hour for 09:00
	End   time.Duration // e.g., 17*time.Hour for 17:00
}

