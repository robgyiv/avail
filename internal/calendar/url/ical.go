package url

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/robgyiv/avail/pkg/availability"
)

// parseICalendar parses iCalendar (ICS) format data and extracts events.
//
// Values that carry no zone of their own (floating date-times and DATE values)
// are interpreted in startRange's location, which callers set from the user's
// configured timezone.
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

// icalProperty is a parsed iCalendar content line: its parameters and its value.
type icalProperty struct {
	params map[string]string
	value  string
}

// Property patterns are compiled once rather than per event.
var (
	dtstartPattern  = propertyPattern("DTSTART")
	dtendPattern    = propertyPattern("DTEND")
	durationPattern = propertyPattern("DURATION")
	summaryPattern  = propertyPattern("SUMMARY")
)

// propertyPattern builds a pattern matching one iCalendar property line, capturing
// its parameters and its value separately. Property names are case-insensitive.
func propertyPattern(name string) *regexp.Regexp {
	return regexp.MustCompile(`(?mi)^` + regexp.QuoteMeta(name) + `((?:;[^:\r\n]*)?):(.*)$`)
}

// findProperty extracts a property from a VEVENT block, keeping its parameters
// (TZID, VALUE) instead of discarding them. Returns nil if the property is absent.
func findProperty(veventData string, pattern *regexp.Regexp) *icalProperty {
	match := pattern.FindStringSubmatch(veventData)
	if match == nil {
		return nil
	}

	prop := &icalProperty{
		params: make(map[string]string),
		value:  strings.TrimSpace(match[2]),
	}

	for _, param := range strings.Split(strings.TrimPrefix(match[1], ";"), ";") {
		key, value, found := strings.Cut(param, "=")
		if !found {
			continue
		}
		key = strings.ToUpper(strings.TrimSpace(key))
		prop.params[key] = strings.Trim(strings.TrimSpace(value), `"`)
	}

	return prop
}

// resolveLocation returns the location named by a TZID parameter, falling back to
// defaultLoc when the parameter is absent or names a zone that cannot be loaded.
// Outlook, for example, emits Windows zone names such as "AUS Eastern Standard
// Time"; falling back to the user's own zone keeps the event (and the busy time it
// represents) rather than dropping it.
func resolveLocation(tzid string, defaultLoc *time.Location) *time.Location {
	if defaultLoc == nil {
		defaultLoc = time.UTC
	}
	if tzid == "" {
		return defaultLoc
	}

	if loc, err := time.LoadLocation(tzid); err == nil {
		return loc
	}

	// Some clients prefix the IANA name, either with a leading "/" for global ids
	// or with a vendor path like "/mozilla.org/20050126_1/Europe/Berlin".
	trimmed := strings.Trim(tzid, "/")
	if loc, err := time.LoadLocation(trimmed); err == nil {
		return loc
	}
	if segments := strings.Split(trimmed, "/"); len(segments) > 2 {
		suffix := strings.Join(segments[len(segments)-2:], "/")
		if loc, err := time.LoadLocation(suffix); err == nil {
			return loc
		}
	}

	return defaultLoc
}

// parseVEvent parses a single VEVENT block.
func parseVEvent(veventData string, startRange, endRange time.Time) (*availability.Event, error) {
	event := &availability.Event{}

	// Per RFC 5545 a value without a zone is local to whoever reads the calendar,
	// which here means the timezone the user configured.
	defaultLoc := startRange.Location()

	dtstart := findProperty(veventData, dtstartPattern)
	if dtstart == nil {
		return nil, fmt.Errorf("missing DTSTART")
	}

	startTime, err := parseICalDateTime(dtstart.value, resolveLocation(dtstart.params["TZID"], defaultLoc))
	if err != nil {
		return nil, fmt.Errorf("invalid DTSTART: %w (value: %s)", err, dtstart.value)
	}
	event.Start = startTime

	// All-day events carry a DATE value rather than a DATE-TIME.
	if dtstart.params["VALUE"] == "DATE" || len(dtstart.value) == 8 {
		event.AllDay = true
	}

	// Extract DTEND or DURATION
	if dtend := findProperty(veventData, dtendPattern); dtend != nil {
		endTime, err := parseICalDateTime(dtend.value, resolveLocation(dtend.params["TZID"], defaultLoc))
		if err == nil {
			event.End = endTime
		}
	} else if duration := findProperty(veventData, durationPattern); duration != nil {
		d, err := parseICalDuration(duration.value)
		if err == nil {
			event.End = event.Start.Add(d)
		}
	}

	if event.End.IsZero() {
		if event.AllDay {
			// For all-day events, end is start of next day
			event.End = event.Start.AddDate(0, 0, 1)
		} else {
			// Default to 1 hour if no end time
			event.End = event.Start.Add(time.Hour)
		}
	}

	// Extract SUMMARY (title)
	if summary := findProperty(veventData, summaryPattern); summary != nil {
		event.Title = unescapeICalText(summary.value)
	}

	// Filter by time range
	if event.Start.After(endRange) || event.End.Before(startRange) {
		return nil, nil // Event outside range
	}

	return event, nil
}

// parseICalDateTime parses an iCalendar date-time value. Values ending in Z or
// carrying an explicit UTC offset describe an instant on their own; values without
// one are floating (RFC 5545 3.3.5) and are interpreted in loc, as are DATE values,
// whose midnight is local midnight.
func parseICalDateTime(value string, loc *time.Location) (time.Time, error) {
	value = strings.TrimSpace(value)
	if loc == nil {
		loc = time.UTC
	}

	// Date-only value: YYYYMMDD
	if len(value) == 8 && !strings.Contains(value, "T") {
		return time.ParseInLocation("20060102", value, loc)
	}

	// Try RFC3339 format first (common in many calendar systems)
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}

	// Formats that describe their own zone.
	zonedFormats := []string{
		"20060102T150405Z",     // UTC with Z
		"20060102T1504Z",       // UTC without seconds
		"20060102T150405-0700", // With timezone offset
		"20060102T1504-0700",   // Without seconds, with timezone offset
	}
	for _, format := range zonedFormats {
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}

	// Floating formats, interpreted in loc.
	floatingFormats := []string{
		"20060102T150405", // Local time
		"20060102T1504",   // Local time without seconds
	}
	for _, format := range floatingFormats {
		if t, err := time.ParseInLocation(format, value, loc); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date-time: %s", value)
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
