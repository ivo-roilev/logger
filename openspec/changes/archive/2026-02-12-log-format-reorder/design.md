## Context

The logger service currently formats log lines with the order: `[timestamp] [app] [user] [level] message`. The level field is lowercase and positioned after optional app/user fields. We need to improve log readability by moving the level field to the second position (immediately after timestamp) and using uppercase abbreviated level identifiers with consistent padding.

Current code responsible:
- `internal/format/line.go:FormatEvent()` - builds log lines
- `internal/model/event.go` - defines LogLevel constants (`debug`, `info`, `warn`, `error`)
- Tests in `internal/format/line_test.go` and `internal/model/event_test.go`

## Goals / Non-Goals

**Goals:**
- Move level field to second position in the log line, right after timestamp
- Convert level to uppercase abbreviations: `DEBUG`, `INFO`, `WARN`, `ERROR`
- Implement consistent (7-character) padding: `[DEBUG]`, `[INFO] `, `[WARN] `, `[ERROR]` (space after bracket for INFO and WARN only)
- Update all tests to reflect the new output format
- Update specification to document the new format

**Non-Goals:**
- Change the timestamp format or the timestamp's position
- Modify app/user/message handling logic
- Change the HTTP API request/response behavior
- Implement log format versioning or migration tooling

## Decisions

1. **Padding implementation**: Right-pad the closing bracket with a space for `INFO` and `WARN` to achieve 7-character-wide fields. This makes alignment natural and readable in text editors without requiring internal spaces.
   - Rationale: External padding (after bracket) is cleaner than internal character substitution and avoids questions about what character to use for padding.

2. **Level abbreviation mapping**: Use 4-5 character abbreviations for brevity while maintaining clarity.
   - `debug` → `DEBUG` (5 chars)
   - `info` → `INFO` (4 chars, padded to 7 total with space)
   - `warn` → `WARN` (4 chars, padded to 7 total with space)
   - `error` → `ERROR` (5 chars)
   - Rationale: Maintains distinctiveness while improving log readability; abbreviations are standard logging conventions.

3. **Implementation location**: Modify `FormatEvent()` in `internal/format/line.go` to:
   - Build the new field order: `[timestamp] [level] [app] [user] message`
   - Apply uppercase and padding to the level before formatting

4. **Testing approach**: Update all existing tests to expect the new format, add additional tests to verify padding alignment is correct.

## Risks / Trade-offs

**Risk**: Breaking change to log format
- **Mitigation**: This project is still in development (no production consumers). Document in RELEASE_NOTES.md that the format has changed.

**Risk**: External log parsers expecting the old format will fail
- **Mitigation**: Not a concern for this stage of development, but worth noting in the spec update.

**Trade-off**: The new field order (`[timestamp] [level] ...`) changes the positional contract vs. the current spec
- **Rationale**: The spec will be updated as part of this change; this is an intentional improvement to usability, not a hidden breaking change.

## Migration Plan

1. Update `internal/format/line.go` to implement new formatting with uppercase abbreviations and padding
2. Update model constants or add a mapping function for level abbreviations
3. Update all tests in `internal/format/line_test.go` to expect new format
4. Update all tests in `internal/model/event_test.go` if they check level formatting
5. Update `specs/rest-log-ingest-to-file/spec.md` to document the new format
6. Update README.md if it mentions log format examples
7. Document the change in RELEASE_NOTES.md

No rollback strategy needed (development stage only).
