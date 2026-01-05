package url

import (
	"testing"
	"time"
)

// Reuse the same parsing tests as local provider since they share the same parsing logic
// These tests verify the parsing functions work correctly for URL provider

func TestParseICalendar_EmptyInput(t *testing.T) {
	start := time.Date(2024, 3, 12, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 13, 0, 0, 0, 0, time.UTC)

	events, err := parseICalendar("", start, end)
	if err != nil {
		t.Errorf("parseICalendar() error = %v, expected no error for empty input", err)
	}
	// parseICalendar returns nil for empty input (no events found)
	if events != nil && len(events) != 0 {
		t.Errorf("parseICalendar() returned %d events, want 0 or nil", len(events))
	}
}

func TestParseVEvent_MissingDTSTART(t *testing.T) {
	start := time.Date(2024, 3, 12, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 13, 0, 0, 0, 0, time.UTC)

	vevent := `BEGIN:VEVENT
DTEND:20240312T150000Z
SUMMARY:Test Event
END:VEVENT`

	event, err := parseVEvent(vevent, start, end)
	if err == nil {
		t.Error("parseVEvent() should return error for missing DTSTART")
	}
	if event != nil {
		t.Error("parseVEvent() should return nil event when DTSTART is missing")
	}
}

func TestParseVEvent_InvalidDTSTART(t *testing.T) {
	start := time.Date(2024, 3, 12, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 13, 0, 0, 0, 0, time.UTC)

	vevent := `BEGIN:VEVENT
DTSTART:invalid-date-format
SUMMARY:Test Event
END:VEVENT`

	event, err := parseVEvent(vevent, start, end)
	if err == nil {
		t.Error("parseVEvent() should return error for invalid DTSTART")
	}
	if event != nil {
		t.Error("parseVEvent() should return nil event when DTSTART is invalid")
	}
}

func TestParseVEvent_NoSUMMARY(t *testing.T) {
	start := time.Date(2024, 3, 12, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 13, 0, 0, 0, 0, time.UTC)

	vevent := `BEGIN:VEVENT
DTSTART:20240312T140000Z
DTEND:20240312T150000Z
END:VEVENT`

	event, err := parseVEvent(vevent, start, end)
	if err != nil {
		t.Errorf("parseVEvent() error = %v, expected no error (SUMMARY is optional)", err)
	}
	if event == nil {
		t.Error("parseVEvent() should return event even without SUMMARY")
	}
	if event != nil && event.Title != "" {
		t.Errorf("parseVEvent() title = %q, want empty string", event.Title)
	}
}

func TestParseVEvent_OutsideDateRange(t *testing.T) {
	start := time.Date(2024, 3, 12, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 13, 0, 0, 0, 0, time.UTC)

	// Event is before the range
	vevent := `BEGIN:VEVENT
DTSTART:20240310T140000Z
DTEND:20240310T150000Z
SUMMARY:Past Event
END:VEVENT`

	event, err := parseVEvent(vevent, start, end)
	if err != nil {
		t.Errorf("parseVEvent() error = %v, expected no error (should filter silently)", err)
	}
	if event != nil {
		t.Error("parseVEvent() should return nil for events outside date range")
	}
}

