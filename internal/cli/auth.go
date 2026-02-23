package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	googlecal "github.com/robgyiv/avail/internal/calendar/google"
	"github.com/robgyiv/avail/internal/config"
)

// newAuthCmd creates the auth command.
func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with calendar provider(s)",
		Long: `Authenticate with your configured calendar provider(s).

Credentials are stored securely in your system keyring.

Configure your calendars in ~/.config/avail/config.toml:
  [[calendars]]
  provider = "google"
  calendar_id = "primary"  # or email address for other calendars
  
  [[calendars]]
  provider = "network"
  url = "https://example.com/calendar.ics"`,
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

	if len(cfg.Calendars) == 0 {
		return fmt.Errorf("no calendars configured\n\nConfigure calendars in ~/.config/avail/config.toml:\n  [[calendars]]\n  provider = \"google\"\n  calendar_id = \"primary\"")
	}

	// Authenticate each calendar that requires it
	googleCalsCount := 0
	authSuccess := 0
	authFailed := 0

	for i, calCfg := range cfg.Calendars {
		switch calCfg.Provider {
		case "google":
			googleCalsCount++
			provider := googlecal.NewProvider()
			if calCfg.CalendarID != "" {
				provider.SetCalendarID(calCfg.CalendarID)
			}

			// Only show Google OAuth prompt once per auth session
			if googleCalsCount == 1 {
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
					authFailed++
					continue
				}
			}

			if err := provider.Authenticate(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Calendar %d (%s): authentication failed: %v\n", i+1, calCfg.Provider, err)
				authFailed++
				continue
			}

			calIDDisplay := calCfg.CalendarID
			if calIDDisplay == "" {
				calIDDisplay = "primary"
			}
			fmt.Printf("✓ Google Calendar authenticated (calendar_id: %s)\n", calIDDisplay)
			authSuccess++

		case "network":
			// Network calendars don't require authentication
			fmt.Printf("✓ Network calendar configured: %s\n", calCfg.URL)
			authSuccess++

		case "local":
			// Local calendars don't require authentication
			fmt.Printf("✓ Local calendar configured: %s\n", calCfg.Path)
			authSuccess++

		default:
			fmt.Fprintf(os.Stderr, "⚠️  Calendar %d: unknown provider '%s'\n", i+1, calCfg.Provider)
			authFailed++
		}
	}

	if authSuccess == 0 {
		return fmt.Errorf("failed to authenticate any calendars")
	}

	if authFailed > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠️  %d calendar(s) failed to authenticate. Check your configuration.\n", authFailed)
	}

	return nil
}
