package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/robgyiv/availability/internal/config"
	urlcal "github.com/robgyiv/availability/internal/calendar/url"
	cal "github.com/robgyiv/availability/internal/calendar"
	googlecal "github.com/robgyiv/availability/internal/calendar/google"
	localcal "github.com/robgyiv/availability/internal/calendar/local"
)

// newAuthCmd creates the auth command.
func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with calendar provider",
		Long: `Authenticate with your calendar provider to enable calendar access.

Supported providers:
  - google: Google Calendar (OAuth2)
  - url: Public calendar URL (any service serving iCalendar format)
  - local: Local .ics file (read from local file system)

The authentication token/URL/path will be stored securely in your system keyring.`,
		RunE: runAuth,
	}

	cmd.Flags().StringP("provider", "p", "", "Calendar provider (google, url, local)")
	cmd.Flags().StringP("url", "u", "", "Public calendar URL (for url provider)")
	cmd.Flags().StringP("file", "f", "", "Path to .ics file (for local provider)")

	return cmd
}

func runAuth(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load config to get default provider
	cfg, err := config.LoadOrCreate()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get provider from flag or config
	providerName, _ := cmd.Flags().GetString("provider")
	if providerName == "" {
		providerName = cfg.CalendarProvider
		if providerName == "" {
			providerName = "google" // Default
		}
	}

	// Create provider and authenticate based on provider
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

	case "url":
		// Get public calendar URL
		calendarURL, _ := cmd.Flags().GetString("url")
		if calendarURL == "" {
			// Check config
			calendarURL = cfg.CalendarURL
			if calendarURL == "" {
				// Prompt for URL
				fmt.Print("Enter public calendar URL (webcal:// or https://): ")
				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read input: %w", err)
				}
				calendarURL = strings.TrimSpace(input)
			}
		}

		if calendarURL == "" {
			fmt.Fprintf(os.Stderr, "Error: Public calendar URL required.\n\n")
			fmt.Fprintf(os.Stderr, "Supported sources:\n")
			fmt.Fprintf(os.Stderr, "  - Apple/iCloud: Calendar app > Calendars > Info icon > Public Calendar > Share Link\n")
			fmt.Fprintf(os.Stderr, "  - Google Calendar: Settings > Integrate calendar > Public URL\n")
			fmt.Fprintf(os.Stderr, "  - Other services: Check your calendar provider's documentation\n\n")
			fmt.Fprintf(os.Stderr, "Or use: avail auth --provider url --url <your-calendar-url>\n")
			return fmt.Errorf("public calendar URL required")
		}

		provider = urlcal.NewProviderFromURL(calendarURL)

	case "local":
		// Get local .ics file path
		icsPath, _ := cmd.Flags().GetString("file")
		if icsPath == "" {
			// Prompt for file path
			fmt.Print("Enter path to your .ics calendar file: ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			icsPath = strings.TrimSpace(input)
		}

		if icsPath == "" {
			fmt.Fprintf(os.Stderr, "Error: Calendar file path required.\n\n")
			fmt.Fprintf(os.Stderr, "Provide the path to a local .ics file.\n")
			fmt.Fprintf(os.Stderr, "You can export your calendar from Apple Calendar, Google Calendar, or any calendar app.\n\n")
			fmt.Fprintf(os.Stderr, "Or use: avail auth --provider local --file <path-to-calendar.ics>\n")
			return fmt.Errorf("calendar file path required")
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
		return fmt.Errorf("unknown provider: %s (supported: google, url, local)", providerName)
	}

	// Authenticate
	if err := provider.Authenticate(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Printf("✓ Calendar connected (read-only) - Provider: %s\n", providerName)
	return nil
}

