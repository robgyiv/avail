package config

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "avail"
	tokenKey    = "oauth_token"
)

// StoreToken stores an OAuth token in the system keyring.
func StoreToken(token string) error {
	return keyring.Set(serviceName, tokenKey, token)
}

// GetToken retrieves an OAuth token from the system keyring.
func GetToken() (string, error) {
	token, err := keyring.Get(serviceName, tokenKey)
	if err == keyring.ErrNotFound {
		return "", ErrTokenNotFound
	}
	if err != nil {
		return "", fmt.Errorf("failed to get token from keyring: %w", err)
	}
	return token, nil
}

// DeleteToken removes an OAuth token from the system keyring.
func DeleteToken() error {
	err := keyring.Delete(serviceName, tokenKey)
	if err == keyring.ErrNotFound {
		return nil // Already deleted, not an error
	}
	return err
}

// ErrTokenNotFound is returned when a token is not found in the keyring.
var ErrTokenNotFound = fmt.Errorf("token not found in keyring")

