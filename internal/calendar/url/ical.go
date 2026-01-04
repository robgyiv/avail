package url

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/robgyiv/availability/pkg/availability"
)

// parseICalendar parses iCalendar (ICS) format data and extracts events.
func parseICalendar(icalData string, startRange, endRange time.Time) ([]availability.Event, error) {
	var events []availability.Event

	// Simple iCalendar parser - looks for VEVENT blocks
	// A more robust implementation would use a proper iCalendar library
	veventRegex := regexp.MustCompile(`BEGIN:VEVENT[\s\S]*?END:VEVENT`)
	matches := veventRegex.FindAllString(icalData, -1)

	for _, match := range matches {
		event, err := parseVEvent(match, startRange, endRange)
		if err != nil {
			continue // Skip invalid events
		}
		if event != nil {
			events = append(events, *event)
		}
	}

	return events, nil
}

// parseVEvent parses a single VEVENT block.
func parseVEvent(veventData string, startRange, endRange time.Time) (*availability.Event, error) {
	event := &availability.Event{}

	// Extract DTSTART - handle both DTSTART;VALUE=DATE: and DTSTART: formats
	dtstartRegex := regexp.MustCompile(`DTSTART(?:[^:]*)?:(.+)`)
	dtstartMatch := dtstartRegex.FindStringSubmatch(veventData)
	if len(dtstartMatch) < 2 {
		return nil, fmt.Errorf("missing DTSTART")
	}

	dtstartLine := dtstartRegex.FindString(veventData)
	isAllDay := strings.Contains(dtstartLine, "VALUE=DATE") || strings.Contains(dtstartLine, ";VALUE=DATE")

	startTime, err := parseICalDateTime(dtstartMatch[1])
	if err != nil {
		return nil, fmt.Errorf("invalid DTSTART: %w (value: %s)", err, dtstartMatch[1])
	}
	event.Start = startTime

	// Check if all-day event (DATE value type or 8-char date)
	if isAllDay || len(strings.TrimSpace(dtstartMatch[1])) == 8 {
		event.AllDay = true
	}

	// Extract DTEND or DURATION
	dtendRegex := regexp.MustCompile(`DTEND(?:[^:]*)?:(.+)`)
	dtendMatch := dtendRegex.FindStringSubmatch(veventData)
	if len(dtendMatch) >= 2 {
		endTime, err := parseICalDateTime(dtendMatch[1])
		if err == nil {
			event.End = endTime
		}
	} else {
		// Try DURATION
		durationRegex := regexp.MustCompile(`DURATION(?:[^:]*)?:(.+)`)
		durationMatch := durationRegex.FindStringSubmatch(veventData)
		if len(durationMatch) >= 2 {
			duration, err := parseICalDuration(durationMatch[1])
			if err == nil {
				event.End = event.Start.Add(duration)
			}
		}
	}

	if event.End.IsZero() {
		if event.AllDay {
			// For all-day events, end is start of next day
			event.End = event.Start.Add(24 * time.Hour)
		} else {
			// Default to 1 hour if no end time
			event.End = event.Start.Add(time.Hour)
		}
	}

	// Extract SUMMARY (title)
	summaryRegex := regexp.MustCompile(`SUMMARY(?:[^:]*)?:(.+)`)
	summaryMatch := summaryRegex.FindStringSubmatch(veventData)
	if len(summaryMatch) >= 2 {
		event.Title = strings.TrimSpace(summaryMatch[1])
		// Unescape iCalendar text
		event.Title = unescapeICalText(event.Title)
	}

	// Filter by time range
	if event.Start.After(endRange) || event.End.Before(startRange) {
		return nil, nil // Event outside range
	}

	return event, nil
}

// parseICalDateTime parses an iCalendar date-time value.
func parseICalDateTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	originalValue := value

	// Check if it's a date-only value (8 characters, no time)
	if len(value) == 8 && !strings.Contains(value, "T") {
		// Date only format: YYYYMMDD
		return time.Parse("20060102", value)
	}

	// Handle VALUE=DATE parameter (all-day events)
	if strings.Contains(value, "VALUE=DATE") {
		// Extract just the date part
		parts := strings.Split(value, ":")
		if len(parts) >= 2 {
			datePart := strings.TrimSpace(parts[len(parts)-1])
			if len(datePart) == 8 {
				return time.Parse("20060102", datePart)
			}
		}
	}

	// Try RFC3339 format first (common in many calendar systems)
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}

	// Try formats with T separator
	formats := []string{
		"20060102T150405Z",     // UTC with Z
		"20060102T150405",      // Local time (no timezone)
		"20060102T1504Z",       // UTC without seconds
		"20060102T1504",        // Local time without seconds
		"20060102T150405-0700", // With timezone offset
		"20060102T150405+0700", // With timezone offset
		"20060102T1504-0700",   // Without seconds, with timezone
		"20060102T1504+0700",   // Without seconds, with timezone
	}

	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			// If format doesn't include timezone and no Z, assume UTC
			if !strings.Contains(format, "Z") && !strings.Contains(format, "-") && !strings.Contains(format, "+") {
				return t.UTC(), nil
			}
			return t, nil
		}
	}

	// Try date-only format
	if len(value) == 8 {
		if t, err := time.Parse("20060102", value); err == nil {
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date-time: %s (original: %s)", value, originalValue)
}

// parseICalDuration parses an iCalendar duration value (e.g., PT1H30M).
func parseICalDuration(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "P") {
		return 0, fmt.Errorf("invalid duration format")
	}

	var duration time.Duration
	value = strings.TrimPrefix(value, "P")

	// Check for time component (T)
	if strings.Contains(value, "T") {
		parts := strings.Split(value, "T")
		if len(parts) == 2 {
			timePart := parts[1]
			// Parse hours
			if h := regexp.MustCompile(`(\d+)H`).FindStringSubmatch(timePart); len(h) > 1 {
				var hours int
				fmt.Sscanf(h[1], "%d", &hours)
				duration += time.Duration(hours) * time.Hour
			}
			// Parse minutes
			if m := regexp.MustCompile(`(\d+)M`).FindStringSubmatch(timePart); len(m) > 1 {
				var minutes int
				fmt.Sscanf(m[1], "%d", &minutes)
				duration += time.Duration(minutes) * time.Minute
			}
			// Parse seconds
			if s := regexp.MustCompile(`(\d+)S`).FindStringSubmatch(timePart); len(s) > 1 {
				var seconds int
				fmt.Sscanf(s[1], "%d", &seconds)
				duration += time.Duration(seconds) * time.Second
			}
		}
	} else {
		// Date component (days, weeks)
		if d := regexp.MustCompile(`(\d+)D`).FindStringSubmatch(value); len(d) > 1 {
			var days int
			fmt.Sscanf(d[1], "%d", &days)
			duration += time.Duration(days) * 24 * time.Hour
		}
		if w := regexp.MustCompile(`(\d+)W`).FindStringSubmatch(value); len(w) > 1 {
			var weeks int
			fmt.Sscanf(w[1], "%d", &weeks)
			duration += time.Duration(weeks) * 7 * 24 * time.Hour
		}
	}

	return duration, nil
}

// unescapeICalText unescapes iCalendar text (e.g., \\n -> \n, \\, -> ,).
func unescapeICalText(text string) string {
	text = strings.ReplaceAll(text, "\\n", "\n")
	text = strings.ReplaceAll(text, "\\,", ",")
	text = strings.ReplaceAll(text, "\\;", ";")
	text = strings.ReplaceAll(text, "\\\\", "\\")
	return text
}

