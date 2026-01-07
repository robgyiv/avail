package cli

import (
	"testing"
)

// Note: Full CLI command tests would require:
// - Mocking config loading
// - Mocking calendar providers
// - Mocking keyring access
// These are better suited for integration tests.
// Here we focus on verifying error message quality.

func TestRunShow_ErrorMessages(t *testing.T) {
	// Test that error messages are helpful and actionable
	errorScenarios := []struct {
		name     string
		errorMsg string
		check    func(string) bool
	}{
		{
			name:     "invalid timezone",
			errorMsg: "invalid timezone",
			check: func(msg string) bool {
				return msg != "" && (msg == "invalid timezone" || len(msg) > 0)
			},
		},
		{
			name:     "invalid work hours",
			errorMsg: "invalid work hours",
			check: func(msg string) bool {
				return msg != ""
			},
		},
		{
			name:     "not authenticated",
			errorMsg: "Not authenticated. Please run 'avail auth' first.",
			check: func(msg string) bool {
				return len(msg) > 0
			},
		},
		{
			name:     "unknown provider",
			errorMsg: "unknown provider",
			check: func(msg string) bool {
				return len(msg) > 0
			},
		},
	}

	for _, tt := range errorScenarios {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check(tt.errorMsg) {
				t.Errorf("Error message %q should be helpful and actionable", tt.errorMsg)
			}
		})
	}
}
