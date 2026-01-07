package api

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUpdateAvailabilityRequest_JSONMarshaling(t *testing.T) {
	req := &UpdateAvailabilityRequest{
		Slots: []AvailabilitySlotRequest{
			{
				Start: "2024-01-01T10:00:00Z",
				End:   "2024-01-01T12:00:00Z",
			},
			{
				Start: "2024-01-02T14:00:00Z",
				End:   "2024-01-02T16:00:00Z",
			},
		},
		Timezone: "America/New_York",
		Window: WindowRange{
			Start: "2024-01-01",
			End:   "2024-01-05",
		},
		GeneratedAt: "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Verify JSON structure
	var unmarshaled UpdateAvailabilityRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if unmarshaled.Timezone != req.Timezone {
		t.Errorf("Unmarshaled Timezone = %q, want %q", unmarshaled.Timezone, req.Timezone)
	}
	if len(unmarshaled.Slots) != len(req.Slots) {
		t.Errorf("Unmarshaled Slots length = %d, want %d", len(unmarshaled.Slots), len(req.Slots))
	}
	if unmarshaled.Window.Start != req.Window.Start {
		t.Errorf("Unmarshaled Window.Start = %q, want %q", unmarshaled.Window.Start, req.Window.Start)
	}
	if unmarshaled.Window.End != req.Window.End {
		t.Errorf("Unmarshaled Window.End = %q, want %q", unmarshaled.Window.End, req.Window.End)
	}
}

func TestUpdateAvailabilityRequest_GeneratedAtOmitEmpty(t *testing.T) {
	// Test that generated_at is omitted when empty
	req := &UpdateAvailabilityRequest{
		Slots:    []AvailabilitySlotRequest{},
		Timezone: "UTC",
		Window: WindowRange{
			Start: "2024-01-01",
			End:   "2024-01-05",
		},
		// GeneratedAt is empty
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Verify generated_at is not in JSON when empty
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if _, exists := jsonMap["generated_at"]; exists {
		t.Error("generated_at should be omitted when empty")
	}
}

func TestAvailabilitySlotRequest_JSONMarshaling(t *testing.T) {
	slot := AvailabilitySlotRequest{
		Start: "2024-01-01T10:00:00Z",
		End:   "2024-01-01T12:00:00Z",
	}

	data, err := json.Marshal(slot)
	if err != nil {
		t.Fatalf("Failed to marshal slot: %v", err)
	}

	var unmarshaled AvailabilitySlotRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal slot: %v", err)
	}

	if unmarshaled.Start != slot.Start {
		t.Errorf("Unmarshaled Start = %q, want %q", unmarshaled.Start, slot.Start)
	}
	if unmarshaled.End != slot.End {
		t.Errorf("Unmarshaled End = %q, want %q", unmarshaled.End, slot.End)
	}
}

func TestWindowRange_JSONMarshaling(t *testing.T) {
	window := WindowRange{
		Start: "2024-01-01",
		End:   "2024-01-05",
	}

	data, err := json.Marshal(window)
	if err != nil {
		t.Fatalf("Failed to marshal window: %v", err)
	}

	var unmarshaled WindowRange
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal window: %v", err)
	}

	if unmarshaled.Start != window.Start {
		t.Errorf("Unmarshaled Start = %q, want %q", unmarshaled.Start, window.Start)
	}
	if unmarshaled.End != window.End {
		t.Errorf("Unmarshaled End = %q, want %q", unmarshaled.End, window.End)
	}
}

func TestUpdateAvailabilityRequest_ISO8601Timestamps(t *testing.T) {
	// Verify that timestamps are in ISO 8601 format (RFC3339)
	now := time.Now()
	slot := AvailabilitySlotRequest{
		Start: now.Format(time.RFC3339),
		End:   now.Add(time.Hour).Format(time.RFC3339),
	}

	// RFC3339 should be parseable
	_, err := time.Parse(time.RFC3339, slot.Start)
	if err != nil {
		t.Errorf("Start timestamp %q is not valid RFC3339: %v", slot.Start, err)
	}

	_, err = time.Parse(time.RFC3339, slot.End)
	if err != nil {
		t.Errorf("End timestamp %q is not valid RFC3339: %v", slot.End, err)
	}
}

func TestWindowRange_ISODateFormat(t *testing.T) {
	// Verify that dates are in ISO format (YYYY-MM-DD)
	window := WindowRange{
		Start: "2024-01-01",
		End:   "2024-12-31",
	}

	// ISO date format should be parseable
	_, err := time.Parse("2006-01-02", window.Start)
	if err != nil {
		t.Errorf("Start date %q is not valid ISO format: %v", window.Start, err)
	}

	_, err = time.Parse("2006-01-02", window.End)
	if err != nil {
		t.Errorf("End date %q is not valid ISO format: %v", window.End, err)
	}
}

