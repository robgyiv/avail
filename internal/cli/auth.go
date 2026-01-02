package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	googlecal "github.com/robgyiv/availability/internal/calendar/google"
)

// newAuthCmd creates the auth command.
func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Google Calendar",
		Long: `Authenticate with Google Calendar to enable calendar access.

This command opens a browser window for Google OAuth authentication.
You'll need to create OAuth credentials in the Google Cloud Console:
1. Go to https://console.cloud.google.com/apis/credentials
2. Create OAuth 2.0 Client ID (Application type: Desktop app)
3. Set environment variables:
   export GOOGLE_CLIENT_ID="your-client-id"
   export GOOGLE_CLIENT_SECRET="your-client-secret"

The authentication token will be stored securely in your system keyring.`,
		RunE: runAuth,
	}

	return cmd
}

func runAuth(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

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

	// Create provider and authenticate
	provider := googlecal.NewProvider()
	if err := provider.Authenticate(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println("✓ Calendar connected (read-only)")
	return nil
}

