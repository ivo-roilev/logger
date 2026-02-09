package format

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"logger/internal/model"
)

// sanitizeString replaces newlines with tabs to keep one-event-per-line.
func sanitizeString(s string) string {
	s = strings.ReplaceAll(s, "\n", "\t")
	s = strings.ReplaceAll(s, "\r", "\t")
	return s
}

// FormatEvent renders an Event into a single log line according to the spec.
func FormatEvent(e model.Event) (string, error) {
	timestamp := e.Timestamp.Format(time.RFC3339)
	level := strings.ToLower(string(e.Level))
	message := sanitizeString(e.Message)

	var b strings.Builder
	// [timestamp] [level] message
	fmt.Fprintf(&b, "[%s] [%s] %s", timestamp, level, message)

	if len(e.Fields) == 0 {
		return b.String(), nil
	}

	// Append extra fields in deterministic (sorted) order.
	keys := make([]string, 0, len(e.Fields))
	for k := range e.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b.WriteString(" | ")
	first := true
	for _, k := range keys {
		if !first {
			b.WriteByte(' ')
		}
		first = false

		v := e.Fields[k]
		valueStr := formatValue(v)
		valueStr = sanitizeString(valueStr)

		fmt.Fprintf(&b, "%s=%s", k, valueStr)
	}

	return b.String(), nil
}

func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	default:
		// For simple primitives or complex types, try JSON first for stability.
		if data, err := json.Marshal(v); err == nil {
			return string(data)
		}
		return fmt.Sprintf("%v", v)
	}
}

