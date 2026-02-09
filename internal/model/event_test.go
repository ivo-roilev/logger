package model

import (
	"testing"
	"time"
)

func TestToEvent_ValidPayload(t *testing.T) {
	payload := EventPayload{
		Timestamp: "2026-02-09T12:34:56Z",
		Level:     "INFO",
		Message:   "Hello",
		Fields: map[string]any{
			"k": "v",
		},
	}

	ev, err := payload.ToEvent()
	if err != nil {
		t.Fatalf("ToEvent returned error: %v", err)
	}

	if ev.Level != LevelInfo {
		t.Fatalf("expected level %q, got %q", LevelInfo, ev.Level)
	}

	if ev.Message != "Hello" {
		t.Fatalf("unexpected message: %q", ev.Message)
	}

	if ev.Timestamp.Format(time.RFC3339) != "2026-02-09T12:34:56Z" {
		t.Fatalf("unexpected timestamp: %s", ev.Timestamp.Format(time.RFC3339))
	}
}

func TestToEvent_InvalidTimestamp(t *testing.T) {
	payload := EventPayload{
		Timestamp: "invalid",
		Level:     "info",
		Message:   "Hello",
	}

	if _, err := payload.ToEvent(); err == nil {
		t.Fatalf("expected error for invalid timestamp")
	}
}

func TestToEvent_InvalidLevel(t *testing.T) {
	payload := EventPayload{
		Timestamp: "2026-02-09T12:34:56Z",
		Level:     "verbose",
		Message:   "Hello",
	}

	if _, err := payload.ToEvent(); err == nil {
		t.Fatalf("expected error for invalid level")
	}
}

func TestToEvent_EmptyMessage(t *testing.T) {
	payload := EventPayload{
		Timestamp: "2026-02-09T12:34:56Z",
		Level:     "info",
		Message:   "   ",
	}

	if _, err := payload.ToEvent(); err == nil {
		t.Fatalf("expected error for empty message")
	}
}

func TestToEvent_3DayWindowValidation_Accept_Today(t *testing.T) {
	// Today (2026-02-09) should be accepted
	payload := EventPayload{
		Timestamp: "2026-02-09T12:00:00Z",
		Level:     "info",
		Message:   "Hello",
	}

	if _, err := payload.ToEvent(); err != nil {
		t.Fatalf("expected valid timestamp for today, got error: %v", err)
	}
}

func TestToEvent_3DayWindowValidation_Accept_Yesterday(t *testing.T) {
	// Yesterday (2026-02-08) should be accepted
	payload := EventPayload{
		Timestamp: "2026-02-08T12:00:00Z",
		Level:     "info",
		Message:   "Hello",
	}

	if _, err := payload.ToEvent(); err != nil {
		t.Fatalf("expected valid timestamp for yesterday, got error: %v", err)
	}
}

func TestToEvent_3DayWindowValidation_Accept_Tomorrow(t *testing.T) {
	// Tomorrow (2026-02-10) should be accepted
	payload := EventPayload{
		Timestamp: "2026-02-10T12:00:00Z",
		Level:     "info",
		Message:   "Hello",
	}

	if _, err := payload.ToEvent(); err != nil {
		t.Fatalf("expected valid timestamp for tomorrow, got error: %v", err)
	}
}

func TestToEvent_3DayWindowValidation_Reject_TwoDaysAgo(t *testing.T) {
	// Two days ago (2026-02-07) should be rejected
	payload := EventPayload{
		Timestamp: "2026-02-07T12:00:00Z",
		Level:     "info",
		Message:   "Hello",
	}

	if _, err := payload.ToEvent(); err == nil {
		t.Fatalf("expected error for timestamp two days ago")
	}
}

func TestToEvent_3DayWindowValidation_Reject_InTwoDays(t *testing.T) {
	// In two days (2026-02-11) should be rejected
	payload := EventPayload{
		Timestamp: "2026-02-11T12:00:00Z",
		Level:     "info",
		Message:   "Hello",
	}

	if _, err := payload.ToEvent(); err == nil {
		t.Fatalf("expected error for timestamp in two days")
	}
}

