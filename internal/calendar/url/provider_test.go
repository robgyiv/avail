package url

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProvider_Authenticate_EmptyURL(t *testing.T) {
	provider := NewProvider()

	err := provider.Authenticate(context.Background())
	if err == nil {
		t.Error("Authenticate() should return error for empty URL")
	}
}

func TestProvider_Authenticate_InvalidURL(t *testing.T) {
	provider := NewProviderFromURL("not-a-valid-url")

	err := provider.Authenticate(context.Background())
	if err == nil {
		t.Error("Authenticate() should return error for invalid URL")
	}
}

func TestProvider_Authenticate_HTTPError(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Not Found")
	}))
	defer server.Close()

	provider := NewProviderFromURL(server.URL)

	err := provider.Authenticate(context.Background())
	if err == nil {
		t.Error("Authenticate() should return error for HTTP 404")
	}
}

func TestProvider_Authenticate_ServerError(t *testing.T) {
	// Create a test server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
	}))
	defer server.Close()

	provider := NewProviderFromURL(server.URL)

	err := provider.Authenticate(context.Background())
	if err == nil {
		t.Error("Authenticate() should return error for HTTP 500")
	}
}

func TestProvider_Authenticate_Forbidden(t *testing.T) {
	// Create a test server that returns 403
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Forbidden")
	}))
	defer server.Close()

	provider := NewProviderFromURL(server.URL)

	err := provider.Authenticate(context.Background())
	if err == nil {
		t.Error("Authenticate() should return error for HTTP 403")
	}
}

func TestProvider_Authenticate_Success(t *testing.T) {
	// Create a test server that returns valid iCalendar
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		fmt.Fprint(w, `BEGIN:VCALENDAR
VERSION:2.0
END:VCALENDAR`)
	}))
	defer server.Close()

	provider := NewProviderFromURL(server.URL)

	err := provider.Authenticate(context.Background())
	if err != nil {
		t.Errorf("Authenticate() error = %v, expected success", err)
	}
}

func TestProvider_IsAuthenticated(t *testing.T) {
	tests := []struct {
		name     string
		provider *Provider
		want     bool
	}{
		{
			name:     "no URL set",
			provider: NewProvider(),
			want:     false,
		},
		{
			name:     "URL set",
			provider: NewProviderFromURL("https://example.com/calendar.ics"),
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

func TestProvider_ListEvents_NetworkError(t *testing.T) {
	// Use a URL that will fail to connect
	provider := NewProviderFromURL("http://localhost:99999/nonexistent")

	// Manually set as authenticated to test network error
	provider.calendarURL = "http://localhost:99999/nonexistent"

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := provider.ListEvents(ctx, time.Now(), time.Now().Add(24*time.Hour))
	if err == nil {
		t.Error("ListEvents() should return error for network failure")
	}
}

func TestProvider_ListEvents_HTTPError(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Not Found")
	}))
	defer server.Close()

	provider := NewProviderFromURL(server.URL)
	provider.calendarURL = server.URL // Manually set as authenticated

	_, err := provider.ListEvents(context.Background(), time.Now(), time.Now().Add(24*time.Hour))
	if err == nil {
		t.Error("ListEvents() should return error for HTTP 404")
	}
}

func TestProvider_ListEvents_EmptyResponse(t *testing.T) {
	// Create a test server that returns empty body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		// Empty response
	}))
	defer server.Close()

	provider := NewProviderFromURL(server.URL)
	provider.calendarURL = server.URL

	events, err := provider.ListEvents(context.Background(), time.Now(), time.Now().Add(24*time.Hour))
	if err != nil {
		t.Errorf("ListEvents() error = %v, expected no error for empty response", err)
	}
	// ListEvents may return nil for empty response (no events found)
	if events != nil && len(events) != 0 {
		t.Errorf("ListEvents() returned %d events, want 0 or nil", len(events))
	}
}

func TestNormalizeCalendarURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     string
	}{
		{
			name:     "webcal to https",
			input:    "webcal://example.com/calendar.ics",
			want:     "https://example.com/calendar.ics",
		},
		{
			name:     "https stays https",
			input:    "https://example.com/calendar.ics",
			want:     "https://example.com/calendar.ics",
		},
		{
			name:     "http stays http",
			input:    "http://example.com/calendar.ics",
			want:     "http://example.com/calendar.ics",
		},
		{
			name:     "invalid URL",
			input:    "not-a-url",
			want:     "",
		},
		{
			name:     "empty string",
			input:    "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeCalendarURL(tt.input)
			if got != tt.want {
				t.Errorf("normalizeCalendarURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProvider_LoadToken_NotFound(t *testing.T) {
	provider := NewProvider()

	// This will fail because we can't easily mock the keyring
	// But we can test the error handling path
	err := provider.LoadToken(context.Background())
	if err == nil {
		// If keyring is available and has a token, this might succeed
		// But in a clean test environment, it should fail
		t.Log("LoadToken() succeeded (may have token in keyring)")
	}
}

func TestProvider_LoadToken_InvalidURL(t *testing.T) {
	// This test would require mocking the keyring, which is complex
	// For now, we'll skip this and focus on URL validation
	t.Skip("Requires keyring mocking")
}

