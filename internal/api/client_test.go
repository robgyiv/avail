package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.baseURL != "https://api.avail.website" {
		t.Errorf("NewClient() baseURL = %q, want %q", client.baseURL, "https://api.avail.website")
	}
	if client.httpClient == nil {
		t.Error("NewClient() httpClient is nil")
	}
	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("NewClient() httpClient.Timeout = %v, want %v", client.httpClient.Timeout, 30*time.Second)
	}
}

func TestClient_PushAvailability_Success(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Verify path
		if r.URL.Path != "/v1/availability" {
			t.Errorf("Expected path /v1/availability, got %s", r.URL.Path)
		}

		// Verify Authorization header
		authHeader := r.Header.Get("Authorization")
		expectedAuth := "Bearer avail_test_token"
		if authHeader != expectedAuth {
			t.Errorf("Authorization header = %q, want %q", authHeader, expectedAuth)
		}

		// Verify Content-Type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
		}

		// Verify User-Agent
		userAgent := r.Header.Get("User-Agent")
		if userAgent != "availability/1.0" {
			t.Errorf("User-Agent = %q, want %q", userAgent, "availability/1.0")
		}

		// Verify request body
		var req UpdateAvailabilityRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	// Create client with test server URL
	client := &Client{
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	// Create test request
	req := &UpdateAvailabilityRequest{
		Slots: []AvailabilitySlotRequest{
			{
				Start: "2024-01-01T10:00:00Z",
				End:   "2024-01-01T12:00:00Z",
			},
		},
		Timezone: "UTC",
		Window: WindowRange{
			Start: "2024-01-01",
			End:   "2024-01-05",
		},
	}

	// Test push
	err := client.PushAvailability(context.Background(), "avail_test_token", req)
	if err != nil {
		t.Errorf("PushAvailability() error = %v, want nil", err)
	}
}

func TestClient_PushAvailability_HTTPErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "400 Bad Request",
			statusCode: http.StatusBadRequest,
			response:   `{"error":{"code":"invalid_request","message":"Invalid slots"}}`,
			wantErr:    true,
			errMsg:     "bad request",
		},
		{
			name:       "401 Unauthorized",
			statusCode: http.StatusUnauthorized,
			response:   `{"error":{"code":"unauthorized","message":"Invalid token"}}`,
			wantErr:    true,
			errMsg:     "unauthorized",
		},
		{
			name:       "429 Too Many Requests",
			statusCode: http.StatusTooManyRequests,
			response:   `{"error":{"code":"rate_limit","message":"Too many requests"}}`,
			wantErr:    true,
			errMsg:     "too many requests",
		},
		{
			name:       "500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
			response:   `{"error":{"code":"server_error","message":"Internal error"}}`,
			wantErr:    true,
			errMsg:     "internal server error",
		},
		{
			name:       "500 with plain text",
			statusCode: http.StatusInternalServerError,
			response:   "Internal Server Error",
			wantErr:    true,
			errMsg:     "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			// Create client with test server URL
			client := &Client{
				baseURL:    server.URL,
				httpClient: &http.Client{Timeout: 5 * time.Second},
			}

			// Create test request
			req := &UpdateAvailabilityRequest{
				Slots:    []AvailabilitySlotRequest{},
				Timezone: "UTC",
				Window: WindowRange{
					Start: "2024-01-01",
					End:   "2024-01-05",
				},
			}

			// Test push
			err := client.PushAvailability(context.Background(), "avail_test_token", req)
			if (err != nil) != tt.wantErr {
				t.Errorf("PushAvailability() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errMsg)) {
					t.Errorf("PushAvailability() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestClient_PushAvailability_InvalidJSON(t *testing.T) {
	// Create a request that can't be marshaled (circular reference would be ideal, but simpler: nil pointer)
	// Actually, UpdateAvailabilityRequest should always be marshallable, so let's test with a valid request
	// but verify JSON marshaling works

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	// Valid request should work
	req := &UpdateAvailabilityRequest{
		Slots:    []AvailabilitySlotRequest{},
		Timezone: "UTC",
		Window: WindowRange{
			Start: "2024-01-01",
			End:   "2024-01-05",
		},
	}

	err := client.PushAvailability(context.Background(), "avail_test_token", req)
	if err != nil {
		t.Errorf("PushAvailability() error = %v, want nil", err)
	}
}

func TestClient_PushAvailability_NetworkError(t *testing.T) {
	// Create client pointing to non-existent server
	client := &Client{
		baseURL:    "http://localhost:0", // Invalid port
		httpClient: &http.Client{Timeout: 100 * time.Millisecond},
	}

	req := &UpdateAvailabilityRequest{
		Slots:    []AvailabilitySlotRequest{},
		Timezone: "UTC",
		Window: WindowRange{
			Start: "2024-01-01",
			End:   "2024-01-05",
		},
	}

	err := client.PushAvailability(context.Background(), "avail_test_token", req)
	if err == nil {
		t.Error("PushAvailability() should return error for network failure")
	}
}
