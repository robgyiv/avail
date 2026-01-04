package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/robgyiv/availability/internal/config"
)

// newExportCmd creates the export command for calendar export instructions.
func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Show instructions for exporting calendar to .ics file",
		Long: `Shows instructions for exporting your calendar to a local .ics file for use with local provider.

To export your calendar:
  1. Open your calendar application (Calendar.app on macOS, Google Calendar, etc.)
  2. Find the export option (typically File > Export or Settings)
  3. Export as .ics format
  4. Save the file
  5. Configure avail to use the local file in ~/.config/avail/config.toml

Alternatively, you can use a public calendar URL by setting calendar_provider = "network" in your config.`,
		RunE: runExport,
	}

	cmd.Flags().StringP("output", "o", "", "Suggested output path for .ics file (default: ~/.config/avail/calendar.ics)")

	return cmd
}

func runExport(cmd *cobra.Command, args []string) error {
	// Get output path
	outputPath, _ := cmd.Flags().GetString("output")
	if outputPath == "" {
		// Default to config directory
		configDir, err := config.ConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get config directory: %w", err)
		}
		outputPath = filepath.Join(configDir, "calendar.ics")
	}

	// Expand ~ to home directory
	if strings.HasPrefix(outputPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		outputPath = filepath.Join(homeDir, outputPath[2:])
	} else if outputPath == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		outputPath = homeDir
	}

	fmt.Printf("Suggested export location: %s\n\n", outputPath)
	fmt.Printf("To export your calendar:\n\n")
	fmt.Printf("1. Open your calendar application:\n")
	fmt.Printf("   - macOS: Calendar.app (File > Export > Export...)\n")
	fmt.Printf("   - Google Calendar: Settings > Export calendar\n")
	fmt.Printf("   - Other: Check your calendar app's export options\n\n")
	fmt.Printf("2. Export as .ics format and save to: %s\n\n", outputPath)
	fmt.Printf("3. Configure avail to use the local file.\n")
	fmt.Printf("   Edit ~/.config/avail/config.toml:\n\n")
	fmt.Printf("   calendar_provider = \"local\"\n")
	fmt.Printf("   local_calendar_path = \"%s\"\n\n", outputPath)
	fmt.Printf("4. Use avail commands:\n")
	fmt.Printf("   avail show\n\n")
	fmt.Printf("Alternative: Use a public calendar URL by setting calendar_provider = \"network\"\n")
	fmt.Printf("and calendar_url in your config file.\n")

	return nil
}
