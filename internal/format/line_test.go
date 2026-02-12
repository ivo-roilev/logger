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

	expected := "[2026-02-09T12:34:56Z] [INFO] User logged in"
	if line != expected {
		t.Fatalf("expected %q, got %q", expected, line)
	}
}

func TestFormatEvent_WithApp(t *testing.T) {
	ev := model.Event{
		Timestamp: time.Date(2026, 2, 9, 12, 34, 56, 0, time.UTC),
		Level:     model.LevelInfo,
		Message:   "Something happened",
		App:       "myservice",
		Fields:    map[string]any{},
	}

	line, err := FormatEvent(ev)
	if err != nil {
		t.Fatalf("FormatEvent returned error: %v", err)
	}

	expected := "[2026-02-09T12:34:56Z] [INFO] [myservice] Something happened"
	if line != expected {
		t.Fatalf("expected %q, got %q", expected, line)
	}
}

func TestFormatEvent_WithUser(t *testing.T) {
	ev := model.Event{
		Timestamp: time.Date(2026, 2, 9, 12, 34, 56, 0, time.UTC),
		Level:     model.LevelInfo,
		Message:   "Something happened",
		User:      "alice",
		Fields:    map[string]any{},
	}

	line, err := FormatEvent(ev)
	if err != nil {
		t.Fatalf("FormatEvent returned error: %v", err)
	}

	expected := "[2026-02-09T12:34:56Z] [INFO] [alice] Something happened"
	if line != expected {
		t.Fatalf("expected %q, got %q", expected, line)
	}
}

func TestFormatEvent_WithAppAndUser(t *testing.T) {
	ev := model.Event{
		Timestamp: time.Date(2026, 2, 9, 12, 34, 56, 0, time.UTC),
		Level:     model.LevelInfo,
		Message:   "Something happened",
		App:       "myservice",
		User:      "alice",
		Fields:    map[string]any{},
	}

	line, err := FormatEvent(ev)
	if err != nil {
		t.Fatalf("FormatEvent returned error: %v", err)
	}

	expected := "[2026-02-09T12:34:56Z] [INFO] [myservice] [alice] Something happened"
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
	expected := "[2026-02-09T12:34:56Z] [ERROR] Line1\tLine2 | ip=203.0.113.42 user_id=123"
	if line != expected {
		t.Fatalf("expected %q, got %q", expected, line)
	}
}

func TestFormatEvent_LevelPaddingAlignment(t *testing.T) {
	// Verify all levels are properly padded to 7 characters (including brackets)
	tests := []struct {
		level    model.LogLevel
		expected string
	}{
		{model.LevelDebug, "[2026-02-09T12:34:56Z] [DEBUG]"},
		{model.LevelInfo, "[2026-02-09T12:34:56Z] [INFO] "},
		{model.LevelWarn, "[2026-02-09T12:34:56Z] [WARN] "},
		{model.LevelError, "[2026-02-09T12:34:56Z] [ERROR]"},
	}

	for _, tt := range tests {
		ev := model.Event{
			Timestamp: time.Date(2026, 2, 9, 12, 34, 56, 0, time.UTC),
			Level:     tt.level,
			Message:   "test",
			Fields:    map[string]any{},
		}

		line, err := FormatEvent(ev)
		if err != nil {
			t.Fatalf("FormatEvent returned error: %v", err)
		}

		// Extract just the timestamp and level part (first 38 chars for the expected prefix)
		if len(line) < 38 {
			t.Fatalf("line too short: %q", line)
		}
		prefix := line[:38]
		if prefix != tt.expected {
			t.Fatalf("level %s: expected %q, got %q", tt.level, tt.expected, prefix)
		}
	}
}

