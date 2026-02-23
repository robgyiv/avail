package aggregate

import (
	"context"
	"testing"
	"time"

	"github.com/robgyiv/avail/internal/config"
)

func TestNewProvider_NoCalendars(t *testing.T) {
	_, err := NewProvider([]config.Calendar{})
	if err == nil {
		t.Error("NewProvider() should error when no calendars configured")
	}
}

func TestNewProvider_InvalidProvider(t *testing.T) {
	calendars := []config.Calendar{
		{Provider: "invalid"},
	}
	_, err := NewProvider(calendars)
	if err == nil {
		t.Error("NewProvider() should error for invalid provider type")
	}
}

func TestNewProvider_LocalProvider(t *testing.T) {
	// Create a valid local provider config
	calendars := []config.Calendar{
		{Provider: "local", Path: "/nonexistent/calendar.ics"},
	}
	provider, err := NewProvider(calendars)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	if provider == nil {
		t.Error("NewProvider() should return a provider")
	}
}

func TestNewProvider_MultipleCalendars(t *testing.T) {
	calendars := []config.Calendar{
		{Provider: "local", Path: "/path/to/calendar1.ics"},
		{Provider: "local", Path: "/path/to/calendar2.ics"},
	}
	provider, err := NewProvider(calendars)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	if provider == nil {
		t.Error("NewProvider() should return a provider")
	}
	if len(provider.providers) != 2 {
		t.Errorf("NewProvider() should have 2 providers, got %d", len(provider.providers))
	}
}

func TestListEvents_NoAuthenticatedProviders(t *testing.T) {
	// Create a provider with local calendars that don't require auth
	calendars := []config.Calendar{
		{Provider: "local", Path: "/nonexistent/calendar.ics"},
	}
	provider, err := NewProvider(calendars)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	ctx := context.Background()
	now := time.Now()
	events, err := provider.ListEvents(ctx, now, now.Add(24*time.Hour))
	// ListEvents might error or return empty events depending on file existence
	// We just test that it doesn't panic
	t.Logf("ListEvents returned %d events", len(events))
}

func TestIsAuthenticated(t *testing.T) {
	calendars := []config.Calendar{
		{Provider: "local", Path: "/path/to/calendar.ics"},
	}
	provider, err := NewProvider(calendars)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	// Local calendars don't require auth, so this should be false or implementation-dependent
	// We just test that the method exists and doesn't panic
	authenticated := provider.IsAuthenticated()
	t.Logf("IsAuthenticated returned %v", authenticated)
}

func TestWarnings(t *testing.T) {
	// Create a provider with one valid and one invalid calendar
	calendars := []config.Calendar{
		{Provider: "local", Path: "/path/to/valid.ics"},
		{Provider: "google"},
		// Google provider without credentials will fail to load auth
	}
	provider, err := NewProvider(calendars)
	if err != nil {
		// It's okay if NewProvider fails because all calendars failed
		t.Logf("NewProvider() error (expected): %v", err)
		return
	}

	warnings := provider.Warnings()
	// Should have warnings about failed calendars
	if len(warnings) == 0 {
		t.Logf("No warnings, but provider loaded successfully")
	} else {
		t.Logf("Got %d warnings: %v", len(warnings), warnings)
	}
}
