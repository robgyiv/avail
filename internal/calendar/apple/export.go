package apple

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ExportCalendar exports an Apple Calendar to a .ics file using AppleScript.
// This only works on macOS.
func ExportCalendar(calendarName string, outputPath string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("Apple Calendar export is only supported on macOS")
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// AppleScript to export calendar
	// Note: Apple Calendar doesn't have a direct export command in AppleScript
	// We'll use a workaround: export via Calendar.app's export functionality
	// or use the calendar's URL if it's a subscribed calendar

	// For now, we'll provide instructions to the user
	// A more advanced implementation could use Calendar.app's scripting
	// but that requires user interaction or more complex automation

	// Alternative: Use icalBuddy if available, or provide manual instructions
	return exportViaCalendarApp(calendarName, outputPath)
}

// exportViaCalendarApp attempts to export via Calendar.app
// This is a simplified version - full implementation would require
// more complex AppleScript or use of external tools
func exportViaCalendarApp(calendarName string, outputPath string) error {
	// Try using osascript to export
	// Note: This is a placeholder - actual implementation would need
	// to interact with Calendar.app more directly

	script := fmt.Sprintf(`
		tell application "Calendar"
			try
				set cal to calendar "%s"
				-- Calendar.app doesn't have direct export in AppleScript
				-- User needs to export manually or we use alternative method
				return "Calendar found: " & name of cal
			on error
				return "Calendar not found"
			end try
		end tell
	`, calendarName)

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to access Calendar.app: %w (output: %s)", err, string(output))
	}

	// For now, return instructions
	// A full implementation would need to use Calendar.app's export functionality
	// which may require user interaction or a different approach
	return fmt.Errorf("automatic export not yet fully implemented. Please export manually from Calendar.app or use a local .ics file")
}

// ExportAllCalendars attempts to export all calendars (placeholder)
func ExportAllCalendars(outputPath string) error {
	return ExportCalendar("", outputPath)
}
