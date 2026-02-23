package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "valid time",
			input:   "09:00",
			want:    9 * time.Hour,
			wantErr: false,
		},
		{
			name:    "valid time with minutes",
			input:   "14:30",
			want:    14*time.Hour + 30*time.Minute,
			wantErr: false,
		},
		{
			name:    "midnight",
			input:   "00:00",
			want:    0,
			wantErr: false,
		},
		{
			name:    "end of day",
			input:   "23:59",
			want:    23*time.Hour + 59*time.Minute,
			wantErr: false,
		},
		{
			name:    "invalid format - no colon",
			input:   "0900",
			wantErr: true,
		},
		{
			name:    "invalid format - text",
			input:   "not-a-time",
			wantErr: true,
		},
		{
			name:    "invalid format - empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid hour - too high",
			input:   "25:00",
			wantErr: true,
		},
		{
			name:    "invalid hour - negative",
			input:   "-1:00",
			wantErr: true,
		},
		{
			name:    "invalid minute - too high",
			input:   "09:60",
			wantErr: true,
		},
		{
			name:    "invalid minute - negative",
			input:   "09:-1",
			wantErr: true,
		},
		{
			name:    "invalid format - extra colon (parses first two numbers)",
			input:   "09:00:00",
			want:    9 * time.Hour, // parseTime only reads first two numbers, so this actually works
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_WorkHours(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid work hours",
			config: &Config{
				WorkHoursStart: "09:00",
				WorkHoursEnd:   "17:00",
			},
			wantErr: false,
		},
		{
			name: "invalid start time",
			config: &Config{
				WorkHoursStart: "25:00",
				WorkHoursEnd:   "17:00",
			},
			wantErr: true,
		},
		{
			name: "invalid end time",
			config: &Config{
				WorkHoursStart: "09:00",
				WorkHoursEnd:   "25:00",
			},
			wantErr: true,
		},
		{
			name: "empty start time",
			config: &Config{
				WorkHoursStart: "",
				WorkHoursEnd:   "17:00",
			},
			wantErr: true,
		},
		{
			name: "empty end time",
			config: &Config{
				WorkHoursStart: "09:00",
				WorkHoursEnd:   "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.config.WorkHours()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.WorkHours() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with google calendar",
			config: &Config{
				Timezone:        "America/New_York",
				MeetingDuration: 30 * time.Minute,
				BufferDuration:  15 * time.Minute,
				WorkHoursStart:  "09:00",
				WorkHoursEnd:    "17:00",
				Calendars: []Calendar{
					{Provider: "google", CalendarID: "primary"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing timezone",
			config: &Config{
				Timezone:        "",
				MeetingDuration: 30 * time.Minute,
				BufferDuration:  15 * time.Minute,
				WorkHoursStart:  "09:00",
				WorkHoursEnd:    "17:00",
			},
			wantErr: true,
			errMsg:  "timezone is required",
		},
		{
			name: "zero meeting duration",
			config: &Config{
				Timezone:        "UTC",
				MeetingDuration: 0,
				BufferDuration:  15 * time.Minute,
				WorkHoursStart:  "09:00",
				WorkHoursEnd:    "17:00",
			},
			wantErr: true,
			errMsg:  "meeting_duration must be positive",
		},
		{
			name: "negative meeting duration",
			config: &Config{
				Timezone:        "UTC",
				MeetingDuration: -30 * time.Minute,
				BufferDuration:  15 * time.Minute,
				WorkHoursStart:  "09:00",
				WorkHoursEnd:    "17:00",
			},
			wantErr: true,
			errMsg:  "meeting_duration must be positive",
		},
		{
			name: "negative buffer duration",
			config: &Config{
				Timezone:        "UTC",
				MeetingDuration: 30 * time.Minute,
				BufferDuration:  -15 * time.Minute,
				WorkHoursStart:  "09:00",
				WorkHoursEnd:    "17:00",
			},
			wantErr: true,
			errMsg:  "buffer_duration must be non-negative",
		},
		{
			name: "missing work hours start",
			config: &Config{
				Timezone:        "UTC",
				MeetingDuration: 30 * time.Minute,
				BufferDuration:  15 * time.Minute,
				WorkHoursStart:  "",
				WorkHoursEnd:    "17:00",
			},
			wantErr: true,
			errMsg:  "work hours start and end are required",
		},
		{
			name: "missing work hours end",
			config: &Config{
				Timezone:        "UTC",
				MeetingDuration: 30 * time.Minute,
				BufferDuration:  15 * time.Minute,
				WorkHoursStart:  "09:00",
				WorkHoursEnd:    "",
			},
			wantErr: true,
			errMsg:  "work hours start and end are required",
		},
		{
			name: "local provider missing path",
			config: &Config{
				Timezone:        "UTC",
				MeetingDuration: 30 * time.Minute,
				BufferDuration:  15 * time.Minute,
				WorkHoursStart:  "09:00",
				WorkHoursEnd:    "17:00",
				Calendars: []Calendar{
					{Provider: "local", Path: ""},
				},
			},
			wantErr: true,
			errMsg:  "calendars[0]: path is required for local provider",
		},
		{
			name: "local provider with path",
			config: &Config{
				Timezone:        "UTC",
				MeetingDuration: 30 * time.Minute,
				BufferDuration:  15 * time.Minute,
				WorkHoursStart:  "09:00",
				WorkHoursEnd:    "17:00",
				Calendars: []Calendar{
					{Provider: "local", Path: "/path/to/calendar.ics"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid work hours format",
			config: &Config{
				Timezone:        "UTC",
				MeetingDuration: 30 * time.Minute,
				BufferDuration:  15 * time.Minute,
				WorkHoursStart:  "25:00",
				WorkHoursEnd:    "17:00",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() != tt.errMsg {
					t.Errorf("Config.Validate() error = %v, want error message containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	// Create a temporary file with invalid TOML
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	invalidTOML := `timezone = "UTC"
meeting_duration = "30m"
work_hours_start = "09:00"
work_hours_end = "17:00"
[invalid bracket`

	err := os.WriteFile(configPath, []byte(invalidTOML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err = Load(configPath)
	if err == nil {
		t.Error("Load() should return error for invalid TOML")
	}
}

func TestLoad_InvalidDurations(t *testing.T) {
	tests := []struct {
		name     string
		tomlData string
		wantErr  bool
	}{
		{
			name: "invalid meeting duration",
			tomlData: `meeting_duration = "30x"
timezone = "UTC"`,
			wantErr: true,
		},
		{
			name: "invalid buffer duration",
			tomlData: `buffer_duration = "abc"
timezone = "UTC"`,
			wantErr: true,
		},
		{
			name: "valid durations",
			tomlData: `meeting_duration = "30m"
buffer_duration = "15m"
timezone = "UTC"`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.toml")

			err := os.WriteFile(configPath, []byte(tt.tomlData), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			_, err = Load(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoad_IncludeWeekends(t *testing.T) {
	tests := []struct {
		name     string
		tomlData string
		want     bool
	}{
		{
			name:     "include weekends true",
			tomlData: "include_weekends = true\n",
			want:     true,
		},
		{
			name:     "include weekends false",
			tomlData: "include_weekends = false\n",
			want:     false,
		},
		{
			name:     "include weekends default",
			tomlData: "timezone = \"UTC\"\n",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.toml")

			err := os.WriteFile(configPath, []byte(tt.tomlData), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			cfg, err := Load(configPath)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.IncludeWeekends != tt.want {
				t.Errorf("Load() include weekends = %v, want %v", cfg.IncludeWeekends, tt.want)
			}
		})
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	// Load should return default config for nonexistent file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.toml")

	cfg, err := Load(configPath)
	if err != nil {
		t.Errorf("Load() error = %v, expected default config", err)
		return
	}

	// Should have defaults
	if cfg.Timezone != "UTC" {
		t.Errorf("Load() default timezone = %v, want UTC", cfg.Timezone)
	}
	if cfg.MeetingDuration != 30*time.Minute {
		t.Errorf("Load() default meeting duration = %v, want 30m", cfg.MeetingDuration)
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Timezone != "UTC" {
		t.Errorf("Default() timezone = %v, want UTC", cfg.Timezone)
	}
	if cfg.MeetingDuration != 30*time.Minute {
		t.Errorf("Default() meeting duration = %v, want 30m", cfg.MeetingDuration)
	}
	if cfg.BufferDuration != 15*time.Minute {
		t.Errorf("Default() buffer duration = %v, want 15m", cfg.BufferDuration)
	}
	if cfg.WorkHoursStart != "09:00" {
		t.Errorf("Default() work hours start = %v, want 09:00", cfg.WorkHoursStart)
	}
	if cfg.WorkHoursEnd != "17:00" {
		t.Errorf("Default() work hours end = %v, want 17:00", cfg.WorkHoursEnd)
	}
	if len(cfg.Calendars) != 1 || cfg.Calendars[0].Provider != "google" {
		t.Errorf("Default() calendars = %v, want one google calendar", cfg.Calendars)
	}
	if cfg.IncludeWeekends {
		t.Errorf("Default() include weekends = %v, want false", cfg.IncludeWeekends)
	}
}
