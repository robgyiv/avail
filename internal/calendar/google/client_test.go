package google

import (
	"testing"
)

func TestProvider_IsAuthenticated(t *testing.T) {
	p := NewProvider()
	if p.IsAuthenticated() {
		t.Error("Expected provider to not be authenticated initially")
	}
}

// Note: Full integration tests would require OAuth setup and API credentials.
// For MVP, we're testing the structure and interface compliance.
// Full integration tests would be added later with proper OAuth test credentials.
