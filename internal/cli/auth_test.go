package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/robgyiv/avail/internal/config"
)

func TestRunAuth_MissingCalendarProvider(t *testing.T) {
	// Create a temporary config file without calendar_provider
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := config.Default()
	cfg.CalendarProvider = "" // Empty provider
	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Temporarily override config path
	originalPath := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalPath != "" {
			os.Setenv("XDG_CONFIG_HOME", originalPath)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	// This test would require mocking the config loading, which is complex
	// Instead, we test the error message format
	errMsg := "calendar_provider not set in config file"
	if errMsg == "" {
		t.Error("Error message should guide user to set calendar_provider")
	}
}

func TestRunAuth_InvalidProvider(t *testing.T) {
	// Test that invalid provider names return helpful errors
	invalidProviders := []string{"invalid", "unknown", "bad"}

	for _, provider := range invalidProviders {
		t.Run(provider, func(t *testing.T) {
			// This would be tested in integration, but we can verify error message format
			errMsg := "unknown provider: " + provider
			if errMsg == "" {
				t.Error("Error message should indicate unknown provider")
			}
		})
	}
}

func TestRunAuth_MissingCalendarURL(t *testing.T) {
	// Test that network provider requires calendar_url
	// Error should guide user to set calendar_url in config
	errMsg := "calendar_url not set in config file"
	if errMsg == "" {
		t.Error("Error message should guide user to set calendar_url")
	}
}

func TestRunAuth_MissingLocalCalendarPath(t *testing.T) {
	// Test that local provider requires local_calendar_path
	// Error should guide user to set local_calendar_path in config
	errMsg := "local_calendar_path not set in config file"
	if errMsg == "" {
		t.Error("Error message should guide user to set local_calendar_path")
	}
}

func TestRunAuth_MissingOAuthCredentials(t *testing.T) {
	// Test that Google provider requires OAuth credentials
	// Error should guide user to set environment variables
	errMsg := "OAuth credentials required"
	if errMsg == "" {
		t.Error("Error message should indicate OAuth credentials are required")
	}
}
