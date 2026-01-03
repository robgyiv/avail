package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/robgyiv/availability/internal/config"
	applecal "github.com/robgyiv/availability/internal/calendar/apple"
	cal "github.com/robgyiv/availability/internal/calendar"
	googlecal "github.com/robgyiv/availability/internal/calendar/google"
)

// newAuthCmd creates the auth command.
func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with calendar provider",
		Long: `Authenticate with your calendar provider to enable calendar access.

Supported providers:
  - google: Google Calendar (OAuth2)
  - apple: Apple/iCloud Calendar (public calendar URL)

The authentication token/URL will be stored securely in your system keyring.

For Google Calendar:
  1. Go to https://console.cloud.google.com/apis/credentials
  2. Create OAuth 2.0 Client ID (Application type: Desktop app)
  3. Set environment variables:
     export GOOGLE_CLIENT_ID="your-client-id"
     export GOOGLE_CLIENT_SECRET="your-client-secret"

For Apple/iCloud Calendar:
  1. Open Calendar app on your iPhone/Mac
  2. Tap/click the "Calendars" button
  3. Tap/click the info icon (ℹ️) next to the calendar you want to share
  4. Toggle on "Public Calendar"
  5. Tap/click "Share Link" to copy the public calendar URL
  6. Provide this URL when prompted`,
		RunE: runAuth,
	}

	cmd.Flags().StringP("provider", "p", "", "Calendar provider (google, apple)")
	cmd.Flags().StringP("url", "u", "", "Public calendar URL (for Apple provider)")

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

	case "apple", "icloud":
		// Get public calendar URL
		calendarURL, _ := cmd.Flags().GetString("url")
		if calendarURL == "" {
			// Prompt for URL
			fmt.Print("Enter your public iCloud calendar URL (webcal:// or https://): ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			calendarURL = strings.TrimSpace(input)
		}

		if calendarURL == "" {
			fmt.Fprintf(os.Stderr, "Error: Public calendar URL required.\n\n")
			fmt.Fprintf(os.Stderr, "To get your public calendar URL:\n")
			fmt.Fprintf(os.Stderr, "1. Open Calendar app on your iPhone/Mac\n")
			fmt.Fprintf(os.Stderr, "2. Tap/click the 'Calendars' button\n")
			fmt.Fprintf(os.Stderr, "3. Tap/click the info icon (ℹ️) next to the calendar you want to share\n")
			fmt.Fprintf(os.Stderr, "4. Toggle on 'Public Calendar'\n")
			fmt.Fprintf(os.Stderr, "5. Tap/click 'Share Link' to copy the URL\n\n")
			fmt.Fprintf(os.Stderr, "Or use: avail auth --provider apple --url <your-calendar-url>\n")
			return fmt.Errorf("public calendar URL required")
		}

		provider = applecal.NewProviderFromURL(calendarURL)

	default:
		return fmt.Errorf("unknown provider: %s (supported: google, apple)", providerName)
	}

	// Authenticate
	if err := provider.Authenticate(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Printf("✓ Calendar connected (read-only) - Provider: %s\n", providerName)
	return nil
}

