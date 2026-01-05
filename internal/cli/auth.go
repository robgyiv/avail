package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	cal "github.com/robgyiv/availability/internal/calendar"
	googlecal "github.com/robgyiv/availability/internal/calendar/google"
	localcal "github.com/robgyiv/availability/internal/calendar/local"
	urlcal "github.com/robgyiv/availability/internal/calendar/url"
	"github.com/robgyiv/availability/internal/config"
)

// newAuthCmd creates the auth command.
func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with calendar provider",
		Long: `Authenticate with your calendar provider based on your configuration.

The provider is determined by the calendar_provider setting in your config file.
Credentials are stored securely in your system keyring.

Before running this command, ensure your config file (~/.config/avail/config.toml) has:
  - calendar_provider set to "google", "network", or "local"
  - Provider-specific settings (calendar_url for network, local_calendar_path for local)`,
		RunE: runAuth,
	}

	return cmd
}

func runAuth(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load config
	cfg, err := config.LoadOrCreate()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get provider from config only
	providerName := cfg.CalendarProvider
	if providerName == "" {
		return fmt.Errorf("calendar_provider not set in config file\n\nSet calendar_provider in ~/.config/avail/config.toml:\n  calendar_provider = \"google\"  # or \"network\" or \"local\"")
	}

	// Create provider and authenticate based on provider from config
	var provider cal.Provider
	switch providerName {
	case "google":
		// Get OAuth credentials from environment
		clientID := os.Getenv("GOOGLE_CLIENT_ID")
		clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

		if clientID == "" || clientSecret == "" {
			fmt.Fprintf(os.Stderr, "Error: OAuth credentials not found.\n\n")
			fmt.Fprintf(os.Stderr, "Please set the following environment variables:\n")
			fmt.Fprintf(os.Stderr, "  export GOOGLE_CLIENT_ID=\"your-client-id\"\n")
			fmt.Fprintf(os.Stderr, "  export GOOGLE_CLIENT_SECRET=\"your-client-secret\"\n\n")
			fmt.Fprintf(os.Stderr, "To get OAuth credentials:\n")
			fmt.Fprintf(os.Stderr, "1. Go to https://console.cloud.google.com/apis/credentials\n")
			fmt.Fprintf(os.Stderr, "2. Create OAuth 2.0 Client ID (Application type: Desktop app)\n")
			fmt.Fprintf(os.Stderr, "3. Add http://localhost/callback as an authorized redirect URI\n\n")
			return fmt.Errorf("OAuth credentials required")
		}

		provider = googlecal.NewProvider()

	case "network":
		// Get public calendar URL from config
		calendarURL := cfg.CalendarURL
		if calendarURL == "" {
			return fmt.Errorf("calendar_url not set in config file\n\nSet calendar_url in ~/.config/avail/config.toml:\n  calendar_url = \"https://calendar.example.com/public.ics\"")
		}

		provider = urlcal.NewProviderFromURL(calendarURL)

	case "local":
		// Get local .ics file path from config
		icsPath := cfg.LocalCalendarPath
		if icsPath == "" {
			return fmt.Errorf("local_calendar_path not set in config file\n\nSet local_calendar_path in ~/.config/avail/config.toml:\n  local_calendar_path = \"/path/to/calendar.ics\"")
		}

		// Expand ~ to home directory
		if strings.HasPrefix(icsPath, "~/") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			icsPath = strings.Replace(icsPath, "~", homeDir, 1)
		}

		provider = localcal.NewProviderFromPath(icsPath)

	default:
		return fmt.Errorf("unknown provider: %s (supported: google, network, local)\n\nSet calendar_provider in ~/.config/avail/config.toml", providerName)
	}

	// Authenticate (stores credentials in keyring)
	if err := provider.Authenticate(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Printf("✓ Calendar connected (read-only) - Provider: %s\n", providerName)
	return nil
}
