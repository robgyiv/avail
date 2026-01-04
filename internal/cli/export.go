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
		Long: `Shows instructions for exporting your calendar to a local .ics file for use with local mode.

To export your calendar:
  1. Open your calendar application (Calendar.app on macOS, Google Calendar, etc.)
  2. Find the export option (typically File > Export or Settings)
  3. Export as .ics format
  4. Save the file
  5. Use: avail auth --provider local --file <path-to-exported-file.ics>

Alternatively, you can use a public calendar URL with:
  avail auth --provider url --url <your-public-calendar-url>`,
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
	fmt.Printf("3. After exporting, configure avail:\n")
	fmt.Printf("   avail auth --provider local --file %s\n\n", outputPath)
	fmt.Printf("Or configure manually:\n")
	fmt.Printf("   calendar_mode = \"local\"\n")
	fmt.Printf("   local_calendar_path = \"%s\"\n\n", outputPath)
	fmt.Printf("Alternative: Use a public calendar URL instead:\n")
	fmt.Printf("   avail auth --provider url --url <your-public-calendar-url>\n")

	return nil
}
