package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CredentialsPath returns the path to the credentials file.
func CredentialsPath() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "credentials"), nil
}

// LoadAPIToken reads the API token from the credentials file.
func LoadAPIToken() (string, error) {
	credentialsPath, err := CredentialsPath()
	if err != nil {
		return "", fmt.Errorf("failed to get credentials path: %w", err)
	}

	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("credentials file not found at %s\n\nTo set up your API token:\n1. Sign up at https://avail.website/\n2. Generate a token\n3. Create %s with your token", credentialsPath, credentialsPath)
		}
		return "", fmt.Errorf("failed to read credentials file: %w", err)
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("credentials file is empty")
	}

	// Validate token format (should start with "avail_")
	if !strings.HasPrefix(token, "avail_") {
		return "", fmt.Errorf("invalid token format: token should start with 'avail_'")
	}

	return token, nil
}

// StoreAPIToken writes the API token to the credentials file.
func StoreAPIToken(token string) error {
	// Validate token format
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	if !strings.HasPrefix(token, "avail_") {
		return fmt.Errorf("invalid token format: token should start with 'avail_'")
	}

	// Ensure config directory exists
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	credentialsPath, err := CredentialsPath()
	if err != nil {
		return fmt.Errorf("failed to get credentials path: %w", err)
	}

	// Write token to file with 0600 permissions (read/write for owner only)
	if err := os.WriteFile(credentialsPath, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}
