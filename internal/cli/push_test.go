package cli

import (
	"testing"
	"time"

	"github.com/robgyiv/availability/internal/api"
	"github.com/robgyiv/availability/pkg/availability"
)

func TestTransformToAPIFormat(t *testing.T) {
	// Create test time blocks
	location, _ := time.LoadLocation("UTC")
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 1, 6, 0, 0, 0, 0, location)

	blocks := []availability.TimeBlock{
		{
			Start: time.Date(2024, 1, 1, 10, 0, 0, 0, location),
			End:   time.Date(2024, 1, 1, 12, 0, 0, 0, location),
		},
		{
			Start: time.Date(2024, 1, 2, 14, 0, 0, 0, location),
			End:   time.Date(2024, 1, 2, 16, 0, 0, 0, location),
		},
	}

	req := transformToAPIFormat(blocks, startDate, endDate, "UTC")

	// Verify structure
	if req.Timezone != "UTC" {
		t.Errorf("transformToAPIFormat() Timezone = %q, want %q", req.Timezone, "UTC")
	}

	if req.Window.Start != "2024-01-01" {
		t.Errorf("transformToAPIFormat() Window.Start = %q, want %q", req.Window.Start, "2024-01-01")
	}

	if req.Window.End != "2024-01-06" {
		t.Errorf("transformToAPIFormat() Window.End = %q, want %q", req.Window.End, "2024-01-06")
	}

	if len(req.Slots) != len(blocks) {
		t.Errorf("transformToAPIFormat() Slots length = %d, want %d", len(req.Slots), len(blocks))
	}

	// Verify first slot
	if req.Slots[0].Start != "2024-01-01T10:00:00Z" {
		t.Errorf("transformToAPIFormat() Slots[0].Start = %q, want %q", req.Slots[0].Start, "2024-01-01T10:00:00Z")
	}

	if req.Slots[0].End != "2024-01-01T12:00:00Z" {
		t.Errorf("transformToAPIFormat() Slots[0].End = %q, want %q", req.Slots[0].End, "2024-01-01T12:00:00Z")
	}

	// Verify second slot
	if req.Slots[1].Start != "2024-01-02T14:00:00Z" {
		t.Errorf("transformToAPIFormat() Slots[1].Start = %q, want %q", req.Slots[1].Start, "2024-01-02T14:00:00Z")
	}

	if req.Slots[1].End != "2024-01-02T16:00:00Z" {
		t.Errorf("transformToAPIFormat() Slots[1].End = %q, want %q", req.Slots[1].End, "2024-01-02T16:00:00Z")
	}

	// Verify generated_at is set
	if req.GeneratedAt == "" {
		t.Error("transformToAPIFormat() GeneratedAt should be set")
	}

	// Verify generated_at is valid RFC3339
	_, err := time.Parse(time.RFC3339, req.GeneratedAt)
	if err != nil {
		t.Errorf("transformToAPIFormat() GeneratedAt %q is not valid RFC3339: %v", req.GeneratedAt, err)
	}
}

func TestTransformToAPIFormat_EmptyBlocks(t *testing.T) {
	location, _ := time.LoadLocation("UTC")
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 1, 6, 0, 0, 0, 0, location)

	blocks := []availability.TimeBlock{}

	req := transformToAPIFormat(blocks, startDate, endDate, "America/New_York")

	if len(req.Slots) != 0 {
		t.Errorf("transformToAPIFormat() Slots length = %d, want 0", len(req.Slots))
	}

	if req.Timezone != "America/New_York" {
		t.Errorf("transformToAPIFormat() Timezone = %q, want %q", req.Timezone, "America/New_York")
	}

	if req.Window.Start != "2024-01-01" {
		t.Errorf("transformToAPIFormat() Window.Start = %q, want %q", req.Window.Start, "2024-01-01")
	}

	if req.Window.End != "2024-01-06" {
		t.Errorf("transformToAPIFormat() Window.End = %q, want %q", req.Window.End, "2024-01-06")
	}
}

func TestTransformToAPIFormat_DateFormats(t *testing.T) {
	location, _ := time.LoadLocation("UTC")
	startDate := time.Date(2024, 12, 31, 0, 0, 0, 0, location)
	endDate := time.Date(2025, 1, 5, 0, 0, 0, 0, location)

	blocks := []availability.TimeBlock{}

	req := transformToAPIFormat(blocks, startDate, endDate, "UTC")

	// Verify ISO date format (YYYY-MM-DD)
	if req.Window.Start != "2024-12-31" {
		t.Errorf("transformToAPIFormat() Window.Start = %q, want %q", req.Window.Start, "2024-12-31")
	}

	if req.Window.End != "2025-01-05" {
		t.Errorf("transformToAPIFormat() Window.End = %q, want %q", req.Window.End, "2025-01-05")
	}
}

func TestTransformToAPIFormat_TimestampFormats(t *testing.T) {
	location, _ := time.LoadLocation("America/New_York")
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 1, 6, 0, 0, 0, 0, location)

	blocks := []availability.TimeBlock{
		{
			Start: time.Date(2024, 1, 1, 9, 30, 0, 0, location),
			End:   time.Date(2024, 1, 1, 11, 45, 0, 0, location),
		},
	}

	req := transformToAPIFormat(blocks, startDate, endDate, "America/New_York")

	// Verify timestamps are in RFC3339 format (ISO 8601)
	startTime, err := time.Parse(time.RFC3339, req.Slots[0].Start)
	if err != nil {
		t.Errorf("Slots[0].Start %q is not valid RFC3339: %v", req.Slots[0].Start, err)
	}

	endTime, err := time.Parse(time.RFC3339, req.Slots[0].End)
	if err != nil {
		t.Errorf("Slots[0].End %q is not valid RFC3339: %v", req.Slots[0].End, err)
	}

	// Verify times are correct (accounting for timezone)
	expectedStart := time.Date(2024, 1, 1, 9, 30, 0, 0, location)
	expectedEnd := time.Date(2024, 1, 1, 11, 45, 0, 0, location)

	if !startTime.Equal(expectedStart) {
		t.Errorf("Parsed Start time = %v, want %v", startTime, expectedStart)
	}

	if !endTime.Equal(expectedEnd) {
		t.Errorf("Parsed End time = %v, want %v", endTime, expectedEnd)
	}
}

func TestTransformToAPIFormat_MatchesAPISpec(t *testing.T) {
	// Verify the transformed data matches the API spec structure
	location, _ := time.LoadLocation("UTC")
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 1, 5, 0, 0, 0, 0, location)

	blocks := []availability.TimeBlock{
		{
			Start: time.Date(2024, 1, 1, 10, 0, 0, 0, location),
			End:   time.Date(2024, 1, 1, 12, 0, 0, 0, location),
		},
	}

	req := transformToAPIFormat(blocks, startDate, endDate, "UTC")

	// Verify all required fields are present
	if req.Timezone == "" {
		t.Error("Timezone should not be empty")
	}

	if req.Window.Start == "" {
		t.Error("Window.Start should not be empty")
	}

	if req.Window.End == "" {
		t.Error("Window.End should not be empty")
	}

	// Verify slots structure
	for i, slot := range req.Slots {
		if slot.Start == "" {
			t.Errorf("Slots[%d].Start should not be empty", i)
		}
		if slot.End == "" {
			t.Errorf("Slots[%d].End should not be empty", i)
		}
	}
}

// Test that the function can handle the API models correctly
func TestTransformToAPIFormat_APIModelsCompatibility(t *testing.T) {
	location, _ := time.LoadLocation("UTC")
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 1, 5, 0, 0, 0, 0, location)

	blocks := []availability.TimeBlock{
		{
			Start: time.Date(2024, 1, 1, 10, 0, 0, 0, location),
			End:   time.Date(2024, 1, 1, 12, 0, 0, 0, location),
		},
	}

	req := transformToAPIFormat(blocks, startDate, endDate, "UTC")

	// Verify it's a valid UpdateAvailabilityRequest
	var _ *api.UpdateAvailabilityRequest = req

	// Verify slots are valid AvailabilitySlotRequest
	for _, slot := range req.Slots {
		var _ api.AvailabilitySlotRequest = slot
	}

	// Verify window is valid WindowRange
	var _ api.WindowRange = req.Window
}

