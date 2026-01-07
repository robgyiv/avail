package api

// UpdateAvailabilityRequest represents the request body for updating availability.
// Matches the OpenAPI spec definition models.UpdateAvailabilityRequest.
type UpdateAvailabilityRequest struct {
	Slots       []AvailabilitySlotRequest `json:"slots"`
	Timezone    string                    `json:"timezone"`
	Window      WindowRange               `json:"window"`
	GeneratedAt string                    `json:"generated_at,omitempty"`
}

// AvailabilitySlotRequest represents a single availability slot.
// Matches the OpenAPI spec definition models.AvailabilitySlotRequest.
type AvailabilitySlotRequest struct {
	Start string `json:"start"` // ISO 8601 timestamp
	End   string `json:"end"`   // ISO 8601 timestamp
}

// WindowRange represents the date range for availability.
// Matches the OpenAPI spec definition models.WindowRange.
type WindowRange struct {
	Start string `json:"start"` // ISO date format: YYYY-MM-DD
	End   string `json:"end"`   // ISO date format: YYYY-MM-DD
}
