package local

import (
	"testing"
	"time"
)

// melbourne is the "user's configured timezone" for these tests: a zone with a
// large, non-zero offset, so a dropped TZID shows up as a wrong instant rather
// than an accidentally correct one.
func melbourne(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Australia/Melbourne")
	if err != nil {
		t.Skipf("Australia/Melbourne unavailable: %v", err)
	}
	return loc
}

// TestParseVEvent_Timezones checks that each way an .ics file can express a time
// resolves to the correct instant. The range is expressed in the user's zone, which
// is what zone-less (floating) values are interpreted in.
func TestParseVEvent_Timezones(t *testing.T) {
	loc := melbourne(t)
	start := time.Date(2026, 9, 21, 0, 0, 0, 0, loc)
	end := time.Date(2026, 9, 28, 0, 0, 0, 0, loc)

	tests := []struct {
		name      string
		dtstart   string
		dtend     string
		wantStart string // expected instant, as RFC3339 in UTC
		wantEnd   string
	}{
		{
			name:      "TZID names an IANA zone",
			dtstart:   "DTSTART;TZID=Australia/Melbourne:20260925T100000",
			dtend:     "DTEND;TZID=Australia/Melbourne:20260925T113000",
			wantStart: "2026-09-25T00:00:00Z",
			wantEnd:   "2026-09-25T01:30:00Z",
		},
		{
			name:      "quoted TZID",
			dtstart:   `DTSTART;TZID="Australia/Melbourne":20260925T100000`,
			dtend:     `DTEND;TZID="Australia/Melbourne":20260925T110000`,
			wantStart: "2026-09-25T00:00:00Z",
			wantEnd:   "2026-09-25T01:00:00Z",
		},
		{
			name:      "TZID in a different zone from the user",
			dtstart:   "DTSTART;TZID=Europe/Berlin:20260925T100000",
			dtend:     "DTEND;TZID=Europe/Berlin:20260925T110000",
			wantStart: "2026-09-25T08:00:00Z",
			wantEnd:   "2026-09-25T09:00:00Z",
		},
		{
			name:      "vendor-prefixed TZID",
			dtstart:   "DTSTART;TZID=/mozilla.org/20050126_1/Europe/Berlin:20260925T100000",
			dtend:     "DTEND;TZID=/mozilla.org/20050126_1/Europe/Berlin:20260925T110000",
			wantStart: "2026-09-25T08:00:00Z",
			wantEnd:   "2026-09-25T09:00:00Z",
		},
		{
			name:      "unloadable TZID falls back to the user's zone",
			dtstart:   "DTSTART;TZID=AUS Eastern Standard Time:20260925T100000",
			dtend:     "DTEND;TZID=AUS Eastern Standard Time:20260925T110000",
			wantStart: "2026-09-25T00:00:00Z",
			wantEnd:   "2026-09-25T01:00:00Z",
		},
		{
			name:      "floating time is the user's local time",
			dtstart:   "DTSTART:20260925T100000",
			dtend:     "DTEND:20260925T110000",
			wantStart: "2026-09-25T00:00:00Z",
			wantEnd:   "2026-09-25T01:00:00Z",
		},
		{
			name:      "UTC values keep their instant",
			dtstart:   "DTSTART:20260925T100000Z",
			dtend:     "DTEND:20260925T110000Z",
			wantStart: "2026-09-25T10:00:00Z",
			wantEnd:   "2026-09-25T11:00:00Z",
		},
		{
			name:      "explicit offset keeps its instant",
			dtstart:   "DTSTART:20260925T100000-0500",
			dtend:     "DTEND:20260925T110000-0500",
			wantStart: "2026-09-25T15:00:00Z",
			wantEnd:   "2026-09-25T16:00:00Z",
		},
		{
			name:      "TZID applies per property",
			dtstart:   "DTSTART;TZID=Australia/Melbourne:20260925T100000",
			dtend:     "DTEND;TZID=UTC:20260925T013000",
			wantStart: "2026-09-25T00:00:00Z",
			wantEnd:   "2026-09-25T01:30:00Z",
		},
		{
			name:      "DURATION is relative to a zoned start",
			dtstart:   "DTSTART;TZID=Australia/Melbourne:20260925T100000",
			dtend:     "DURATION:PT1H30M",
			wantStart: "2026-09-25T00:00:00Z",
			wantEnd:   "2026-09-25T01:30:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vevent := "BEGIN:VEVENT\nUID:1@example.com\n" + tt.dtstart + "\n" + tt.dtend + "\nSUMMARY:Standup\nEND:VEVENT"

			event, err := parseVEvent(vevent, start, end)
			if err != nil {
				t.Fatalf("parseVEvent() error = %v", err)
			}
			if event == nil {
				t.Fatal("parseVEvent() returned no event; it should fall inside the range")
			}

			if got := event.Start.UTC().Format(time.RFC3339); got != tt.wantStart {
				t.Errorf("start = %s, want %s", got, tt.wantStart)
			}
			if got := event.End.UTC().Format(time.RFC3339); got != tt.wantEnd {
				t.Errorf("end = %s, want %s", got, tt.wantEnd)
			}
		})
	}
}

// TestParseVEvent_AllDayUsesLocalMidnight checks that a DATE value starts at midnight
// in the user's zone, not at UTC midnight (which is mid-morning in Melbourne and so
// would place the event on the wrong day).
func TestParseVEvent_AllDayUsesLocalMidnight(t *testing.T) {
	loc := melbourne(t)
	start := time.Date(2026, 9, 21, 0, 0, 0, 0, loc)
	end := time.Date(2026, 9, 28, 0, 0, 0, 0, loc)

	vevent := "BEGIN:VEVENT\nUID:1@example.com\nDTSTART;VALUE=DATE:20260925\nDTEND;VALUE=DATE:20260926\nSUMMARY:Public holiday\nEND:VEVENT"

	event, err := parseVEvent(vevent, start, end)
	if err != nil {
		t.Fatalf("parseVEvent() error = %v", err)
	}
	if event == nil {
		t.Fatal("parseVEvent() returned no event")
	}
	if !event.AllDay {
		t.Error("AllDay = false, want true for a VALUE=DATE event")
	}

	wantStart := time.Date(2026, 9, 25, 0, 0, 0, 0, loc)
	if !event.Start.Equal(wantStart) {
		t.Errorf("start = %s, want %s", event.Start.Format(time.RFC3339), wantStart.Format(time.RFC3339))
	}
	if got, want := event.Start.In(loc).Hour(), 0; got != want {
		t.Errorf("start hour in user's zone = %d, want %d", got, want)
	}
}

// TestParseVEvent_DateTimeValueIsNotAllDay guards the VALUE parameter comparison:
// "DATE-TIME" contains "DATE" but describes a timed event.
func TestParseVEvent_DateTimeValueIsNotAllDay(t *testing.T) {
	loc := melbourne(t)
	start := time.Date(2026, 9, 21, 0, 0, 0, 0, loc)
	end := time.Date(2026, 9, 28, 0, 0, 0, 0, loc)

	vevent := "BEGIN:VEVENT\nUID:1@example.com\nDTSTART;VALUE=DATE-TIME;TZID=Australia/Melbourne:20260925T100000\nDTEND;VALUE=DATE-TIME;TZID=Australia/Melbourne:20260925T110000\nSUMMARY:Standup\nEND:VEVENT"

	event, err := parseVEvent(vevent, start, end)
	if err != nil {
		t.Fatalf("parseVEvent() error = %v", err)
	}
	if event == nil {
		t.Fatal("parseVEvent() returned no event")
	}
	if event.AllDay {
		t.Error("AllDay = true, want false for a VALUE=DATE-TIME event")
	}
	if got, want := event.Start.UTC().Format(time.RFC3339), "2026-09-25T00:00:00Z"; got != want {
		t.Errorf("start = %s, want %s", got, want)
	}
}

// TestResolveLocation covers the TZID spellings calendar clients emit.
func TestResolveLocation(t *testing.T) {
	loc := melbourne(t)

	tests := []struct {
		name     string
		tzid     string
		wantName string
	}{
		{name: "empty falls back", tzid: "", wantName: "Australia/Melbourne"},
		{name: "IANA name", tzid: "Europe/Berlin", wantName: "Europe/Berlin"},
		{name: "global id with leading slash", tzid: "/Europe/Berlin", wantName: "Europe/Berlin"},
		{name: "vendor prefix", tzid: "/mozilla.org/20050126_1/Europe/Berlin", wantName: "Europe/Berlin"},
		{name: "windows name falls back", tzid: "AUS Eastern Standard Time", wantName: "Australia/Melbourne"},
		{name: "nonsense falls back", tzid: "Not/A/Zone", wantName: "Australia/Melbourne"},
		{name: "UTC", tzid: "UTC", wantName: "UTC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveLocation(tt.tzid, loc)
			if got.String() != tt.wantName {
				t.Errorf("resolveLocation(%q) = %s, want %s", tt.tzid, got, tt.wantName)
			}
		})
	}
}

// TestFindProperty checks that parameters are kept rather than folded into the value.
func TestFindProperty(t *testing.T) {
	vevent := "BEGIN:VEVENT\nUID:1@example.com\nDTSTART;TZID=Australia/Melbourne;VALUE=DATE-TIME:20260925T100000\r\nEND:VEVENT"

	prop := findProperty(vevent, dtstartPattern)
	if prop == nil {
		t.Fatal("findProperty() = nil, want the DTSTART property")
	}
	if got, want := prop.value, "20260925T100000"; got != want {
		t.Errorf("value = %q, want %q", got, want)
	}
	if got, want := prop.params["TZID"], "Australia/Melbourne"; got != want {
		t.Errorf("TZID = %q, want %q", got, want)
	}
	if got, want := prop.params["VALUE"], "DATE-TIME"; got != want {
		t.Errorf("VALUE = %q, want %q", got, want)
	}

	if findProperty("BEGIN:VEVENT\nUID:1@example.com\nEND:VEVENT", dtstartPattern) != nil {
		t.Error("findProperty() found a DTSTART that is not there")
	}
}
