package engine

import (
	"sort"
	"time"

	"github.com/robgyiv/avail/pkg/availability"
)

// CalculateAvailability computes free time blocks from calendar events.
// It takes events, a date range, work hours, meeting duration, and buffer duration
// to determine available time slots.
func CalculateAvailability(
	events []availability.Event,
	startDate, endDate time.Time,
	workHours availability.WorkHours,
	meetingDuration time.Duration,
	bufferDuration time.Duration,
) []availability.TimeBlock {
	// Sort events by start time
	sortedEvents := make([]availability.Event, len(events))
	copy(sortedEvents, events)
	sort.Slice(sortedEvents, func(i, j int) bool {
		return sortedEvents[i].Start.Before(sortedEvents[j].Start)
	})

	var freeBlocks []availability.TimeBlock

	// Iterate through each day in the range
	currentDate := startDate.Truncate(24 * time.Hour)
	for currentDate.Before(endDate) {
		dayStart := currentDate.Add(workHours.Start)
		dayEnd := currentDate.Add(workHours.End)

		// Find events that overlap with this day
		dayEvents := filterEventsForDay(sortedEvents, currentDate)

		// Calculate free time blocks for this day
		dayBlocks := calculateFreeBlocksForDay(dayStart, dayEnd, dayEvents, meetingDuration, bufferDuration)
		freeBlocks = append(freeBlocks, dayBlocks...)

		// Move to next day
		currentDate = currentDate.Add(24 * time.Hour)
	}

	return freeBlocks
}

// filterEventsForDay returns events that overlap with the given day.
func filterEventsForDay(events []availability.Event, day time.Time) []availability.Event {
	dayStart := day.Truncate(24 * time.Hour)
	dayEnd := dayStart.Add(24 * time.Hour)

	var dayEvents []availability.Event
	for _, event := range events {
		// Skip all-day events for now (they block the entire day)
		if event.AllDay {
			continue
		}

		// Check if event overlaps with this day
		if event.End.After(dayStart) && event.Start.Before(dayEnd) {
			dayEvents = append(dayEvents, event)
		}
	}

	return dayEvents
}

// calculateFreeBlocksForDay calculates free time blocks for a single day.
func calculateFreeBlocksForDay(
	dayStart, dayEnd time.Time,
	events []availability.Event,
	meetingDuration time.Duration,
	bufferDuration time.Duration,
) []availability.TimeBlock {
	var blocks []availability.TimeBlock

	// Sort events by start time
	sortedEvents := make([]availability.Event, len(events))
	copy(sortedEvents, events)
	sort.Slice(sortedEvents, func(i, j int) bool {
		return sortedEvents[i].Start.Before(sortedEvents[j].Start)
	})

	// Start from the beginning of the day
	currentTime := dayStart

	for _, event := range sortedEvents {
		// Skip events that start after work hours end
		if event.Start.After(dayEnd) {
			continue
		}

		eventStart := event.Start
		if eventStart.Before(dayStart) {
			eventStart = dayStart
		}
		eventEnd := event.End
		if eventEnd.After(dayEnd) {
			eventEnd = dayEnd
		}

		// Calculate the buffer zone before the event
		// Only apply buffer before if the event's original start is within work hours
		bufferStart := currentTime
		if event.Start.Before(dayEnd) && event.Start.After(dayStart) {
			// Event starts within work hours, apply buffer before it
			bufferStart = eventStart.Add(-bufferDuration)
			if bufferStart.Before(dayStart) {
				bufferStart = dayStart
			}
		}

		// If there's a gap before the buffer zone, add it as a free block
		if currentTime.Before(bufferStart) {
			gapDuration := bufferStart.Sub(currentTime)
			if gapDuration >= meetingDuration {
				blocks = append(blocks, availability.TimeBlock{
					Start: currentTime,
					End:   bufferStart,
				})
			}
		}

		// Move current time to the end of the event plus buffer (or later if events overlap)
		// Ensure we don't go beyond dayEnd
		nextTime := eventEnd.Add(bufferDuration)
		if nextTime.After(dayEnd) {
			nextTime = dayEnd
		}
		if nextTime.After(currentTime) {
			currentTime = nextTime
		}
	}

	// Add remaining time after the last event
	if currentTime.Before(dayEnd) {
		remainingDuration := dayEnd.Sub(currentTime)
		if remainingDuration >= meetingDuration {
			blocks = append(blocks, availability.TimeBlock{
				Start: currentTime,
				End:   dayEnd,
			})
		}
	}

	return blocks
}
