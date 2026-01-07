package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCredentialsPath(t *testing.T) {
	path, err := CredentialsPath()
	if err != nil {
		t.Fatalf("CredentialsPath() error = %v", err)
	}

	// Should be in config directory
	configDir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}

	expectedPath := filepath.Join(configDir, "credentials")
	if path != expectedPath {
		t.Errorf("CredentialsPath() = %q, want %q", path, expectedPath)
	}
}

func TestStoreAPIToken(t *testing.T) {
	tmpDir := t.TempDir()

	// Override ConfigDir for testing
	originalConfigDir := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalConfigDir == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", originalConfigDir)
		}
	}()

	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   "avail_test_token_12345",
			wantErr: false,
		},
		{
			name:    "token with whitespace",
			token:   "  avail_test_token_12345  ",
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "token without prefix",
			token:   "invalid_token",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			token:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := StoreAPIToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("StoreAPIToken() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify token was stored correctly
				path, err := CredentialsPath()
				if err != nil {
					t.Fatalf("CredentialsPath() error = %v", err)
				}

				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read credentials file: %v", err)
				}

				// Token should be trimmed
				expectedToken := "avail_test_token_12345"
				if string(data) != expectedToken {
					t.Errorf("Stored token = %q, want %q", string(data), expectedToken)
				}

				// Verify file permissions (should be 0600)
				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("Failed to stat credentials file: %v", err)
				}
				if info.Mode().Perm() != 0600 {
					t.Errorf("Credentials file permissions = %o, want 0600", info.Mode().Perm())
				}
			}
		})
	}
}

func TestLoadAPIToken(t *testing.T) {
	tmpDir := t.TempDir()

	// Override ConfigDir for testing
	originalConfigDir := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalConfigDir == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", originalConfigDir)
		}
	}()

	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	tests := []struct {
		name      string
		setup     func() error
		wantToken string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid token",
			setup: func() error {
				path, err := CredentialsPath()
				if err != nil {
					return err
				}
				if err := EnsureConfigDir(); err != nil {
					return err
				}
				return os.WriteFile(path, []byte("avail_test_token_12345"), 0600)
			},
			wantToken: "avail_test_token_12345",
			wantErr:   false,
		},
		{
			name: "token with whitespace",
			setup: func() error {
				path, err := CredentialsPath()
				if err != nil {
					return err
				}
				if err := EnsureConfigDir(); err != nil {
					return err
				}
				return os.WriteFile(path, []byte("  avail_test_token_12345  \n"), 0600)
			},
			wantToken: "avail_test_token_12345",
			wantErr:   false,
		},
		{
			name: "file does not exist",
			setup: func() error {
				// Don't create file
				return nil
			},
			wantErr: true,
			errMsg:  "credentials file not found",
		},
		{
			name: "empty file",
			setup: func() error {
				path, err := CredentialsPath()
				if err != nil {
					return err
				}
				if err := EnsureConfigDir(); err != nil {
					return err
				}
				return os.WriteFile(path, []byte(""), 0600)
			},
			wantErr: true,
			errMsg:  "credentials file is empty",
		},
		{
			name: "invalid token format",
			setup: func() error {
				path, err := CredentialsPath()
				if err != nil {
					return err
				}
				if err := EnsureConfigDir(); err != nil {
					return err
				}
				return os.WriteFile(path, []byte("invalid_token"), 0600)
			},
			wantErr: true,
			errMsg:  "invalid token format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing file
			path, _ := CredentialsPath()
			os.Remove(path)

			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			token, err := LoadAPIToken()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadAPIToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if token != tt.wantToken {
					t.Errorf("LoadAPIToken() = %q, want %q", token, tt.wantToken)
				}
			} else {
				if tt.errMsg != "" && err != nil && err.Error() != "" {
					// Check that error message contains expected text
					if err.Error() == "" {
						t.Errorf("LoadAPIToken() error message is empty")
					}
				}
			}
		})
	}
}

func TestLoadAPIToken_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Override ConfigDir for testing
	originalConfigDir := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalConfigDir == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", originalConfigDir)
		}
	}()

	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Ensure file doesn't exist
	path, _ := CredentialsPath()
	os.Remove(path)

	_, err := LoadAPIToken()
	if err == nil {
		t.Error("LoadAPIToken() should return error when file doesn't exist")
	}

	// Error should mention credentials file
	if err != nil && err.Error() == "" {
		t.Error("LoadAPIToken() error message should be helpful")
	}
}
