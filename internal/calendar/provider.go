package calendar

import (
	"context"
	"time"

	"github.com/robgyiv/avail/pkg/availability"
)

// Provider defines the interface for calendar integrations.
type Provider interface {
	// ListEvents fetches calendar events for the given time range.
	ListEvents(ctx context.Context, start, end time.Time) ([]availability.Event, error)

	// Authenticate performs OAuth authentication.
	Authenticate(ctx context.Context) error

	// IsAuthenticated checks if the provider is currently authenticated.
	IsAuthenticated() bool
}
