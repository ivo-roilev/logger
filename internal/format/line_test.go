package format

import (
	"testing"
	"time"

	"logger/internal/model"
)

func TestFormatEvent_NoFields(t *testing.T) {
	ev := model.Event{
		Timestamp: time.Date(2026, 2, 9, 12, 34, 56, 0, time.UTC),
		Level:     model.LevelInfo,
		Message:   "User logged in",
		Fields:    map[string]any{},
	}

	line, err := FormatEvent(ev)
	if err != nil {
		t.Fatalf("FormatEvent returned error: %v", err)
	}

	expected := "[2026-02-09T12:34:56Z] [info] User logged in"
	if line != expected {
		t.Fatalf("expected %q, got %q", expected, line)
	}
}

func TestFormatEvent_WithFieldsAndNewlines(t *testing.T) {
	ev := model.Event{
		Timestamp: time.Date(2026, 2, 9, 12, 34, 56, 0, time.UTC),
		Level:     model.LevelError,
		Message:   "Line1\nLine2",
		Fields: map[string]any{
			"user_id": "123",
			"ip":      "203.0.113.42",
		},
	}

	line, err := FormatEvent(ev)
	if err != nil {
		t.Fatalf("FormatEvent returned error: %v", err)
	}

	// Keys should be sorted: ip then user_id.
	expected := "[2026-02-09T12:34:56Z] [error] Line1\tLine2 | ip=203.0.113.42 user_id=123"
	if line != expected {
		t.Fatalf("expected %q, got %q", expected, line)
	}
}

