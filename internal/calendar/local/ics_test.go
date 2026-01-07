package local

import (
	"testing"
	"time"
)

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

func TestParseICalendar_InvalidFormat(t *testing.T) {
	start := time.Date(2024, 3, 12, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 13, 0, 0, 0, 0, time.UTC)

	invalidICS := `This is not a valid iCalendar format`
	events, err := parseICalendar(invalidICS, start, end)
	if err != nil {
		// Error is acceptable
		return
	}
	// parseICalendar returns nil when no VEVENT blocks are found
	if events != nil && len(events) != 0 {
		t.Errorf("parseICalendar() returned %d events for invalid format, want 0 or nil", len(events))
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

func TestParseVEvent_NoDTENDOrDURATION(t *testing.T) {
	start := time.Date(2024, 3, 12, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 13, 0, 0, 0, 0, time.UTC)

	vevent := `BEGIN:VEVENT
DTSTART:20240312T140000Z
SUMMARY:Test Event
END:VEVENT`

	event, err := parseVEvent(vevent, start, end)
	if err != nil {
		t.Errorf("parseVEvent() error = %v, expected no error (should default to 1 hour)", err)
	}
	if event == nil {
		t.Error("parseVEvent() should return event even without DTEND")
	}
	if event != nil {
		expectedEnd := event.Start.Add(time.Hour)
		if !event.End.Equal(expectedEnd) {
			t.Errorf("parseVEvent() default end time = %v, want %v", event.End, expectedEnd)
		}
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

func TestParseICalDateTime_InvalidFormats(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "not-a-date",
			wantErr: true,
		},
		{
			name:    "incomplete date",
			input:   "202403",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseICalDateTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseICalDateTime() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseICalDateTime_ValidFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "RFC3339 format",
			input: "2024-03-12T14:00:00Z",
		},
		{
			name:  "iCalendar format with Z",
			input: "20240312T140000Z",
		},
		{
			name:  "date only",
			input: "20240312",
		},
		{
			name:  "with timezone offset",
			input: "20240312T140000-0500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseICalDateTime(tt.input)
			if err != nil {
				t.Errorf("parseICalDateTime() error = %v, expected valid format", err)
			}
		})
	}
}

func TestParseICalDuration_InvalidFormats(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "no P prefix",
			input:   "1H30M",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "Pinvalid",
			wantErr: false, // Currently doesn't error, just returns 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseICalDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseICalDuration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseICalDuration_ValidFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantMins int
	}{
		{
			name:     "1 hour",
			input:    "PT1H",
			wantMins: 60,
		},
		{
			name:     "30 minutes",
			input:    "PT30M",
			wantMins: 30,
		},
		{
			name:     "1 hour 30 minutes",
			input:    "PT1H30M",
			wantMins: 90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := parseICalDuration(tt.input)
			if err != nil {
				t.Errorf("parseICalDuration() error = %v", err)
				return
			}
			gotMins := int(duration.Minutes())
			if gotMins != tt.wantMins {
				t.Errorf("parseICalDuration() = %v minutes, want %v", gotMins, tt.wantMins)
			}
		})
	}
}

func TestUnescapeICalText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "newline escape",
			input: "Line1\\nLine2",
			want:  "Line1\nLine2",
		},
		{
			name:  "comma escape",
			input: "Item1\\,Item2",
			want:  "Item1,Item2",
		},
		{
			name:  "semicolon escape",
			input: "Item1\\;Item2",
			want:  "Item1;Item2",
		},
		{
			name:  "backslash escape",
			input: "Path\\\\to\\\\file",
			want:  "Path\\to\\file",
		},
		{
			name:  "multiple escapes",
			input: "Title\\, with\\nnewline",
			want:  "Title, with\nnewline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unescapeICalText(tt.input)
			if got != tt.want {
				t.Errorf("unescapeICalText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseVEvent_AllDayEvent(t *testing.T) {
	start := time.Date(2024, 3, 12, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 13, 0, 0, 0, 0, time.UTC)

	vevent := `BEGIN:VEVENT
DTSTART;VALUE=DATE:20240312
DTEND;VALUE=DATE:20240313
SUMMARY:All Day Event
END:VEVENT`

	event, err := parseVEvent(vevent, start, end)
	if err != nil {
		t.Errorf("parseVEvent() error = %v, expected no error", err)
	}
	if event == nil {
		t.Fatal("parseVEvent() should return event for all-day event")
	}
	if !event.AllDay {
		t.Error("parseVEvent() AllDay = false, want true")
	}
}
