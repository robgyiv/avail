package google

import (
	"context"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestProvider_IsAuthenticated(t *testing.T) {
	p := NewProvider()
	if p.IsAuthenticated() {
		t.Error("Expected provider to not be authenticated initially")
	}
}

func TestProvider_ListEvents_NotAuthenticated(t *testing.T) {
	p := NewProvider()

	_, err := p.ListEvents(context.Background(), time.Now(), time.Now().Add(24*time.Hour))
	if err == nil {
		t.Error("ListEvents() should return error when not authenticated")
	}
}

func TestTokenFromJSON_InvalidJSON(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		wantErr bool
	}{
		{
			name:    "empty string",
			jsonStr: "",
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			jsonStr: "not json",
			wantErr: true,
		},
		{
			name:    "missing fields",
			jsonStr: `{}`,
			wantErr: false, // oauth2.Token can be created with empty fields
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := TokenFromJSON(tt.jsonStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenFromJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTokenToJSON_ValidToken(t *testing.T) {
	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	}

	jsonStr, err := TokenToJSON(token)
	if err != nil {
		t.Errorf("TokenToJSON() error = %v", err)
	}
	if jsonStr == "" {
		t.Error("TokenToJSON() returned empty string")
	}

	// Verify we can parse it back
	parsedToken, err := TokenFromJSON(jsonStr)
	if err != nil {
		t.Errorf("TokenFromJSON() error = %v", err)
	}
	if parsedToken.AccessToken != token.AccessToken {
		t.Errorf("TokenFromJSON() access token = %q, want %q", parsedToken.AccessToken, token.AccessToken)
	}
}

func TestOAuthConfig(t *testing.T) {
	config := OAuthConfig("test-client-id", "test-client-secret", "http://localhost:8080/callback")

	if config.ClientID != "test-client-id" {
		t.Errorf("OAuthConfig() ClientID = %q, want %q", config.ClientID, "test-client-id")
	}
	if config.ClientSecret != "test-client-secret" {
		t.Errorf("OAuthConfig() ClientSecret = %q, want %q", config.ClientSecret, "test-client-secret")
	}
	if config.RedirectURL != "http://localhost:8080/callback" {
		t.Errorf("OAuthConfig() RedirectURL = %q, want %q", config.RedirectURL, "http://localhost:8080/callback")
	}
}

// Note: Full integration tests would require OAuth setup and API credentials.
// For MVP, we're testing the structure and interface compliance.
// Full integration tests would be added later with proper OAuth test credentials.
