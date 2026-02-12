## MODIFIED Requirements

### Requirement: Log line formatting
The system SHALL format each accepted log event as a single human-readable text line using a consistent bracketed format, ensuring all content is sanitised and rendered on a single line.

#### Scenario: Formats base line with timestamp, level, optional app/user, and message
- **WHEN** a valid log event is processed
- **THEN** the system SHALL write a line with segments in the following order:
  1. Timestamp segment: `[<timestamp>]` (RFC3339 format)
  2. Level segment: ` [<level>]` (uppercase 4-5 character abbreviation with padding to 7 chars total)
  3. App segment (if `app` is present and non-empty): ` [<app>]`
  4. User segment (if `user` is present and non-empty): ` [<user>]`
  5. Message: space followed by `<message>`
  6. Optional fields segment (if `fields` is non-empty): ` | <key-value pairs>`

#### Scenario: Level abbreviation and padding format
- **WHEN** formatting a log event level
- **THEN** the system SHALL:
  - Convert the level to uppercase: `DEBUG`, `INFO`, `WARN`, `ERROR`
  - Wrap it in square brackets: `[DEBUG]`, `[INFO]`, `[WARN]`, `[ERROR]`
  - Pad the output to exactly 7 characters total (including brackets) by adding a space after the closing bracket for `INFO` and `WARN`: `[DEBUG]`, `[INFO] `, `[WARN] `, `[ERROR]`

#### Scenario: Sanitises newlines and special characters
- **WHEN** formatting the log line
- **THEN** the system SHALL replace any `\n` or `\r` characters found in `message`, `app`, `user` values, and in string values within `fields` with horizontal tab characters (`\t`) to preserve the one-line-per-event property.

#### Scenario: Appends structured fields as key=value pairs
- **WHEN** a valid log event includes a non-empty `fields` object
- **THEN** the system SHALL append to the base line a space, a vertical bar, and a space (` | `) followed by the event fields rendered as `key=value` pairs sorted lexicographically by key and separated by single spaces, with complex values (objects, arrays) encoded as JSON.
