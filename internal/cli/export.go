package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/robgyiv/availability/internal/config"
	applecal "github.com/robgyiv/availability/internal/calendar/apple"
)

// newExportCmd creates the export command for Apple Calendar.
func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export Apple Calendar to .ics file (macOS only)",
		Long: `Exports your Apple Calendar to a local .ics file for use with local mode.

This command is only available on macOS and requires Calendar.app access.

For manual export:
  1. Open Calendar.app
  2. Select the calendar you want to export
  3. File > Export > Export...
  4. Choose a location and save as .ics file
  5. Use: avail auth --provider local --file <path-to-exported-file.ics>

Alternatively, you can use the public calendar URL method with:
  avail auth --provider apple --url <your-public-calendar-url>`,
		RunE: runExport,
	}

	cmd.Flags().StringP("calendar", "c", "", "Calendar name to export (optional, exports all if not specified)")
	cmd.Flags().StringP("output", "o", "", "Output path for .ics file (default: ~/.config/avail/calendar.ics)")

	return cmd
}

func runExport(cmd *cobra.Command, args []string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("export command is only available on macOS")
	}


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

	calendarName, _ := cmd.Flags().GetString("calendar")

	fmt.Printf("Exporting Apple Calendar to: %s\n", outputPath)
	fmt.Printf("\nNote: Automatic export from Calendar.app is limited.\n")
	fmt.Printf("For best results, please export manually:\n")
	fmt.Printf("  1. Open Calendar.app\n")
	fmt.Printf("  2. File > Export > Export...\n")
	fmt.Printf("  3. Save to: %s\n", outputPath)
	fmt.Printf("\nOr use the public calendar URL method:\n")
	fmt.Printf("  avail auth --provider apple --url <your-public-calendar-url>\n\n")

	// Try to export (may not work fully, but provides framework)
	if calendarName != "" {
		if err := applecal.ExportCalendar(calendarName, outputPath); err != nil {
			// Don't fail - just warn
			fmt.Printf("Warning: %v\n", err)
			fmt.Printf("Please export manually as described above.\n")
		}
	} else {
		if err := applecal.ExportAllCalendars(outputPath); err != nil {
			fmt.Printf("Warning: %v\n", err)
			fmt.Printf("Please export manually as described above.\n")
		}
	}

	fmt.Printf("\nAfter exporting, you can use local mode:\n")
	fmt.Printf("  1. Update config: calendar_mode = \"local\"\n")
	fmt.Printf("  2. Update config: local_calendar_path = \"%s\"\n", outputPath)
	fmt.Printf("  3. Run: avail show\n")

	return nil
}

