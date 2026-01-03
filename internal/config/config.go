package config

import (
	"fmt"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/robgyiv/availability/pkg/availability"
)

// Config represents the application configuration.
type Config struct {
	Timezone          string        `toml:"timezone"`
	MeetingDuration   time.Duration `toml:"meeting_duration"`
	BufferDuration    time.Duration `toml:"buffer_duration"`
	WorkHoursStart    string        `toml:"work_hours_start"`    // e.g., "09:00"
	WorkHoursEnd      string        `toml:"work_hours_end"`      // e.g., "17:00"
	CalendarProvider  string        `toml:"calendar_provider"`   // "google", "apple", etc.
	CalendarMode      string        `toml:"calendar_mode"`       // "network" (default) or "local"
	LocalCalendarPath string        `toml:"local_calendar_path"` // Path to .ics file for local mode
}

// WorkHours returns the WorkHours struct from config.
func (c *Config) WorkHours() (availability.WorkHours, error) {
	start, err := parseTime(c.WorkHoursStart)
	if err != nil {
		return availability.WorkHours{}, fmt.Errorf("invalid work_hours_start: %w", err)
	}

	end, err := parseTime(c.WorkHoursEnd)
	if err != nil {
		return availability.WorkHours{}, fmt.Errorf("invalid work_hours_end: %w", err)
	}

	return availability.WorkHours{
		Start: start,
		End:   end,
	}, nil
}

// parseTime parses a time string in HH:MM format and returns a Duration.
func parseTime(s string) (time.Duration, error) {
	var h, m int
	_, err := fmt.Sscanf(s, "%d:%d", &h, &m)
	if err != nil {
		return 0, err
	}

	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, fmt.Errorf("invalid time: %s", s)
	}

	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute, nil
}

// configTOML is an intermediate struct for TOML unmarshaling that handles duration strings.
type configTOML struct {
	Timezone          string `toml:"timezone"`
	MeetingDuration   string `toml:"meeting_duration"`
	BufferDuration    string `toml:"buffer_duration"`
	WorkHoursStart    string `toml:"work_hours_start"`
	WorkHoursEnd      string `toml:"work_hours_end"`
	CalendarProvider  string `toml:"calendar_provider"`
	CalendarMode      string `toml:"calendar_mode"`
	LocalCalendarPath string `toml:"local_calendar_path"`
}

// Load reads and parses the config file.
// If the file doesn't exist, it returns the default config.
func Load(path string) (*Config, error) {
	data, err := readFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}

	var cfgTOML configTOML
	if err := toml.Unmarshal(data, &cfgTOML); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg := &Config{
		Timezone:          cfgTOML.Timezone,
		WorkHoursStart:    cfgTOML.WorkHoursStart,
		WorkHoursEnd:      cfgTOML.WorkHoursEnd,
		CalendarProvider:  cfgTOML.CalendarProvider,
		CalendarMode:      cfgTOML.CalendarMode,
		LocalCalendarPath: cfgTOML.LocalCalendarPath,
	}

	// Parse durations
	if cfgTOML.MeetingDuration != "" {
		duration, err := time.ParseDuration(cfgTOML.MeetingDuration)
		if err != nil {
			return nil, fmt.Errorf("invalid meeting_duration: %w", err)
		}
		cfg.MeetingDuration = duration
	}

	if cfgTOML.BufferDuration != "" {
		duration, err := time.ParseDuration(cfgTOML.BufferDuration)
		if err != nil {
			return nil, fmt.Errorf("invalid buffer_duration: %w", err)
		}
		cfg.BufferDuration = duration
	}

	// Apply defaults if not set
	if cfg.MeetingDuration == 0 {
		cfg.MeetingDuration = 30 * time.Minute
	}
	if cfg.BufferDuration == 0 {
		cfg.BufferDuration = 15 * time.Minute
	}
	if cfg.CalendarMode == "" {
		cfg.CalendarMode = "network"
	}
	if cfg.Timezone == "" {
		cfg.Timezone = "UTC"
	}
	if cfg.CalendarProvider == "" {
		cfg.CalendarProvider = "google"
	}

	return cfg, nil
}

// LoadOrCreate loads the config from the default path, creating it if it doesn't exist.
func LoadOrCreate() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	cfg, err := Load(configPath)
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, create it with defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := EnsureConfigDir(); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}

		if err := cfg.Save(configPath); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
	}

	return cfg, nil
}

// Validate checks that the config values are valid.
func (c *Config) Validate() error {
	if c.Timezone == "" {
		return fmt.Errorf("timezone is required")
	}

	if c.MeetingDuration <= 0 {
		return fmt.Errorf("meeting_duration must be positive")
	}

	if c.BufferDuration < 0 {
		return fmt.Errorf("buffer_duration must be non-negative")
	}

	if c.WorkHoursStart == "" || c.WorkHoursEnd == "" {
		return fmt.Errorf("work hours start and end are required")
	}

	calendarMode := c.CalendarMode
	if calendarMode == "" {
		calendarMode = "network" // Default
	}

	if calendarMode != "network" && calendarMode != "local" {
		return fmt.Errorf("calendar_mode must be 'network' or 'local'")
	}

	if calendarMode == "local" && c.LocalCalendarPath == "" {
		return fmt.Errorf("local_calendar_path is required when calendar_mode is 'local'")
	}

	_, err := c.WorkHours()
	if err != nil {
		return err
	}

	return nil
}

// Save writes the config to a file.
func (c *Config) Save(path string) error {
	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return writeFile(path, data)
}

// Default returns a default configuration.
func Default() *Config {
	return &Config{
		Timezone:         "UTC",
		MeetingDuration:  30 * time.Minute,
		BufferDuration:   15 * time.Minute,
		WorkHoursStart:   "09:00",
		WorkHoursEnd:     "17:00",
		CalendarProvider: "google",
		CalendarMode:     "network",
	}
}

// readFile reads a file from disk.
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// writeFile writes data to a file.
func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0600)
}
