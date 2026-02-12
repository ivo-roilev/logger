## Why

The current log format places level information after optional fields (app, user), making it harder to quickly identify the severity of a log entry. Moving level to the second position (right after timestamp) and using uppercase abbreviated level names with consistent padding improves readability and scanning efficiency in log files.

## What Changes

- **Reorder fields** in log output: `[timestamp] [level] [app] [user] message` (was: `[timestamp] [app] [user] [level] message`)
- **Uppercase abbreviations** for log levels: `DEBUG`, `INFO`, `WARN`, `ERROR` (was: lowercase `debug`, `info`, `warn`, `error`)
- **Consistent padding** for log levels: Each level abbreviation is right-padded to 5 characters total width, with the closing bracket followed by a space for `INFO` and `WARN` only, resulting in 7-character-wide fields including brackets
- Formatting: `[DEBUG]`, `[INFO] `, `[WARN] `, `[ERROR]`

## Capabilities

### New Capabilities
<!-- None - this is a formatting change to existing functionality -->

### Modified Capabilities
- `rest-log-ingest-to-file`: Log output format has changed (field order, level representation, and padding adjustments)

## Impact

- **Code changes**: Internal format package (`internal/format/line.go`)
- **Tests**: Update tests in `internal/format/line_test.go` to reflect new format expectations
- **Documentation**: Update README and any log format specifications to reflect the new format
- **Breaking change**: Existing log parsers and monitoring systems that expect the old format will need to be updated
