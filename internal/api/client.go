package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles API requests to the avail.website API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new API client with the default base URL.
func NewClient() *Client {
	return &Client{
		baseURL: "https://api.avail.website",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PushAvailability sends availability data to the API.
func (c *Client) PushAvailability(ctx context.Context, token string, req *UpdateAvailabilityRequest) error {
	// Marshal request body
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.baseURL + "/v1/availability"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("User-Agent", "availability/1.0")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for error messages
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Handle HTTP errors
	if resp.StatusCode != http.StatusOK {
		return handleHTTPError(resp.StatusCode, bodyBytes)
	}

	return nil
}

// handleHTTPError creates user-friendly error messages based on HTTP status codes.
func handleHTTPError(statusCode int, bodyBytes []byte) error {
	var errorMsg string

	switch statusCode {
	case http.StatusBadRequest:
		errorMsg = "bad request"
	case http.StatusUnauthorized:
		errorMsg = "unauthorized - invalid or expired API token"
	case http.StatusTooManyRequests:
		errorMsg = "too many requests - please try again later"
	case http.StatusInternalServerError:
		errorMsg = "internal server error"
	default:
		errorMsg = fmt.Sprintf("unexpected status code: %d", statusCode)
	}

	// Try to parse error response for more details
	var errorResp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &errorResp); err == nil && errorResp.Error.Message != "" {
		return fmt.Errorf("%s: %s", errorMsg, errorResp.Error.Message)
	}

	// If we can't parse the error, return the status code and raw body (truncated)
	bodyStr := string(bodyBytes)
	if len(bodyStr) > 200 {
		bodyStr = bodyStr[:200] + "..."
	}
	if bodyStr != "" {
		return fmt.Errorf("%s: %s", errorMsg, bodyStr)
	}

	return fmt.Errorf("%s", errorMsg)
}

