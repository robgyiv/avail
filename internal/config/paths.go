package config

import (
	"os"
	"path/filepath"
)

// ConfigDir returns the XDG-compliant configuration directory.
// On Unix: ~/.config/avail
// On Windows: %APPDATA%/avail
func ConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Use XDG_CONFIG_HOME if set, otherwise use default
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(homeDir, ".config")
	}

	configDir := filepath.Join(configHome, "avail")
	return configDir, nil
}

// ConfigPath returns the path to the config file.
func ConfigPath() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.toml"), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func EnsureConfigDir() error {
	configDir, err := ConfigDir()
	if err != nil {
		return err
	}

	return os.MkdirAll(configDir, 0755)
}

