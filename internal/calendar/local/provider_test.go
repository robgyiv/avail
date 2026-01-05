package local

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProvider_Authenticate_FileNotFound(t *testing.T) {
	provider := NewProviderFromPath("/nonexistent/path/calendar.ics")

	err := provider.Authenticate(context.Background())
	if err == nil {
		t.Error("Authenticate() should return error for nonexistent file")
	}
	if err != nil && err.Error() == "" {
		t.Error("Authenticate() should return descriptive error message")
	}
}

func TestProvider_Authenticate_EmptyPath(t *testing.T) {
	provider := NewProvider()

	err := provider.Authenticate(context.Background())
	if err == nil {
		t.Error("Authenticate() should return error for empty path")
	}
}

func TestProvider_IsAuthenticated(t *testing.T) {
	tests := []struct {
		name     string
		provider *Provider
		want     bool
	}{
		{
			name:     "no path set",
			provider: NewProvider(),
			want:     false,
		},
		{
			name:     "path set",
			provider: NewProviderFromPath("/path/to/calendar.ics"),
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.provider.IsAuthenticated()
			if got != tt.want {
				t.Errorf("IsAuthenticated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_ListEvents_NotAuthenticated(t *testing.T) {
	provider := NewProvider()

	_, err := provider.ListEvents(context.Background(), time.Now(), time.Now().Add(24*time.Hour))
	if err == nil {
		t.Error("ListEvents() should return error when not authenticated")
	}
}

func TestProvider_ListEvents_FileNotFound(t *testing.T) {
	provider := NewProviderFromPath("/nonexistent/path/calendar.ics")

	// Authenticate will fail, but let's test ListEvents directly
	_, err := provider.ListEvents(context.Background(), time.Now(), time.Now().Add(24*time.Hour))
	if err == nil {
		t.Error("ListEvents() should return error for nonexistent file")
	}
}

func TestProvider_ListEvents_InvalidICalendar(t *testing.T) {
	tmpDir := t.TempDir()
	icsPath := filepath.Join(tmpDir, "invalid.ics")

	// Write invalid iCalendar data
	invalidICS := `This is not a valid iCalendar file`
	err := os.WriteFile(icsPath, []byte(invalidICS), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	provider := NewProviderFromPath(icsPath)
	
	// Authenticate should succeed (file exists)
	err = provider.Authenticate(context.Background())
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	// ListEvents should handle invalid format gracefully
	events, err := provider.ListEvents(context.Background(), time.Now(), time.Now().Add(24*time.Hour))
	if err != nil {
		// Error is acceptable for invalid format
		return
	}
	// ListEvents may return nil for invalid format (no valid events found)
	if events != nil && len(events) != 0 {
		t.Errorf("ListEvents() returned %d events for invalid format, want 0 or nil", len(events))
	}
}

func TestProvider_ListEvents_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	icsPath := filepath.Join(tmpDir, "empty.ics")

	// Write empty file
	err := os.WriteFile(icsPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	provider := NewProviderFromPath(icsPath)
	err = provider.Authenticate(context.Background())
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	events, err := provider.ListEvents(context.Background(), time.Now(), time.Now().Add(24*time.Hour))
	if err != nil {
		t.Errorf("ListEvents() error = %v, expected no error for empty file", err)
	}
	// ListEvents may return nil for empty file (no events found)
	if events != nil && len(events) != 0 {
		t.Errorf("ListEvents() returned %d events, want 0 or nil", len(events))
	}
}

func TestProvider_LoadToken(t *testing.T) {
	provider := NewProviderFromPath("/path/to/calendar.ics")

	err := provider.LoadToken(context.Background())
	if err == nil {
		t.Error("LoadToken() should return error for local provider")
	}
}

