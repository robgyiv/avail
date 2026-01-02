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
	Timezone        string        `toml:"timezone"`
	MeetingDuration time.Duration `toml:"meeting_duration"`
	WorkHoursStart  string        `toml:"work_hours_start"` // e.g., "09:00"
	WorkHoursEnd    string        `toml:"work_hours_end"`   // e.g., "17:00"
	CalendarProvider string       `toml:"calendar_provider"` // "google", etc.
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

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
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

	if c.WorkHoursStart == "" || c.WorkHoursEnd == "" {
		return fmt.Errorf("work hours start and end are required")
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
		Timezone:        "UTC",
		MeetingDuration: 30 * time.Minute,
		WorkHoursStart:  "09:00",
		WorkHoursEnd:    "17:00",
		CalendarProvider: "google",
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

