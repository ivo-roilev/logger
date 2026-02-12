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

// levelAbbreviation converts a LogLevel to its uppercase abbreviated form.
// Returns the bracketed abbreviation: [DEBUG], [INFO], [WARN], [ERROR]
func levelAbbreviation(level model.LogLevel) string {
	switch level {
	case model.LevelDebug:
		return "[DEBUG]"
	case model.LevelInfo:
		return "[INFO]"
	case model.LevelWarn:
		return "[WARN]"
	case model.LevelError:
		return "[ERROR]"
	default:
		return "[" + strings.ToUpper(string(level)) + "]"
	}
}

// levelFormatted returns the level with 7-character padding (including brackets).
// [DEBUG]  → [DEBUG] (7 chars, no padding needed)
// [INFO]   → [INFO]  (7 chars with space after)
// [WARN]   → [WARN]  (7 chars with space after)
// [ERROR]  → [ERROR] (7 chars, no padding needed)
func levelFormatted(level model.LogLevel) string {
	abbr := levelAbbreviation(level)
	// Pad to 7 chars total by adding space after bracket for shorter levels
	if len(abbr) < 7 {
		abbr = abbr + " "
	}
	return abbr
}

// FormatEvent renders an Event into a single log line according to the spec.
func FormatEvent(e model.Event) (string, error) {
	timestamp := e.Timestamp.Format(time.RFC3339)
	level := levelFormatted(e.Level)
	message := sanitizeString(e.Message)

	var b strings.Builder
	// [timestamp] [LEVEL] [app] [user] message
	// Level field is 7 chars total (with padding after bracket for shorter levels)
	fmt.Fprintf(&b, "[%s] %s", timestamp, level)

	if e.App != "" {
		fmt.Fprintf(&b, " [%s]", sanitizeString(e.App))
	}

	if e.User != "" {
		fmt.Fprintf(&b, " [%s]", sanitizeString(e.User))
	}

	fmt.Fprintf(&b, " %s", message)

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

