package model

import (
	"fmt"
	"strings"
	"time"
)

// LogLevel represents a normalised log level.
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// ParseLogLevel normalises a level string into a LogLevel.
func ParseLogLevel(s string) (LogLevel, error) {
	level := strings.ToLower(strings.TrimSpace(s))
	switch level {
	case string(LevelDebug):
		return LevelDebug, nil
	case string(LevelInfo):
		return LevelInfo, nil
	case string(LevelWarn):
		return LevelWarn, nil
	case string(LevelError):
		return LevelError, nil
	default:
		return "", fmt.Errorf("unsupported level: %q", s)
	}
}

// EventPayload is the JSON payload as received over HTTP.
type EventPayload struct {
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

// Event is the validated and normalised internal representation.
type Event struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Fields    map[string]any
}

// ToEvent validates and normalises the incoming payload into an Event.
func (p *EventPayload) ToEvent() (Event, error) {
	var ev Event

	ts := strings.TrimSpace(p.Timestamp)
	if ts == "" {
		return ev, fmt.Errorf("missing field: timestamp")
	}
	parsed, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ev, fmt.Errorf("invalid timestamp: must be RFC3339")
	}

	level, err := ParseLogLevel(p.Level)
	if err != nil {
		return ev, err
	}

	msg := strings.TrimSpace(p.Message)
	if msg == "" {
		return ev, fmt.Errorf("missing field: message")
	}

	fields := p.Fields
	if fields == nil {
		fields = make(map[string]any)
	}

	ev = Event{
		Timestamp: parsed,
		Level:     level,
		Message:   msg,
		Fields:    fields,
	}
	return ev, nil
}

