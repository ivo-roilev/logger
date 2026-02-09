## ADDED Requirements

### Requirement: HTTP API and JSON payload schema
The system SHALL expose an HTTP `POST /logs` endpoint that accepts JSON log events conforming to a defined schema.

#### Scenario: JSON payload structure and validation
- **WHEN** the service processes a `POST /logs` request
- **THEN** the service SHALL expect and validate a JSON object with the following fields:
  - `timestamp` (string, required): RFC3339-formatted UTC timestamp
  - `level` (string, required): one of `debug`, `info`, `warn`, `error` (case-insensitive, normalised to lowercase)
  - `message` (string, required): non-empty after trimming whitespace
  - `user` (string, optional): if present, must be a non-empty string
  - `app` (string, optional): if present, must be a non-empty string
  - `fields` (object, optional): arbitrary key-value pairs; keys must be strings; values may be strings, numbers, booleans, null, nested objects, or arrays

#### Scenario: Accepts valid log event
- **WHEN** a client sends an HTTP `POST` request to `/logs` with a `Content-Type: application/json` header and a well-formed JSON body containing all required fields and optional fields as defined above
- **THEN** the service SHALL respond with HTTP status `202 Accepted` and a JSON body containing `{ "status": "ok" }`, and SHALL append exactly one new line to the configured log file representing the event.

#### Scenario: Rejects invalid JSON payload
- **WHEN** a client sends a `POST /logs` request with a `Content-Type: application/json` header but the request body is not valid JSON or does not match the expected schema
- **THEN** the service SHALL respond with an HTTP `400 Bad Request` status and a JSON error body describing the validation problem, and SHALL NOT append any line to the log file.

#### Scenario: Field validation and normalisation
- **WHEN** the service receives a JSON payload
- **THEN** the service SHALL:
  - Reject requests where `timestamp` is not a valid RFC3339 string
  - Reject requests where `level` is not one of the allowed values
  - Reject requests where `message` is missing, empty, or contains only whitespace
  - Reject requests where the UTC date extracted from `timestamp` is not within a 3-day window: the current server date, the previous day, or the next day
  - Accept and preserve optional `user` and `app` fields only if they are non-empty strings
  - Accept `fields` as an optional object with arbitrary key-value content
  - Normalise `level` to lowercase for consistent formatting

#### Scenario: Rejects out-of-window timestamps
- **WHEN** a client sends a `POST /logs` request with a timestamp whose UTC date is outside the 3-day window (e.g., more than 1 day in the past or more than 1 day in the future)
- **THEN** the service SHALL respond with an HTTP `400 Bad Request` status and a JSON error body indicating that the timestamp is outside the acceptable window, and SHALL NOT append any line to the log file.

#### Scenario: Accepts app value from URL query parameter
- **WHEN** a client sends a `POST /logs` request with an optional `app` query parameter (e.g., `POST /logs?app=checkout-service`)
- **THEN** the service SHALL accept the `app` value from the query parameter and use it in the log line if no `app` value is present in the JSON body; if both the query parameter and JSON body contain an `app` value, the JSON body value SHALL take precedence.

#### Scenario: Rejects invalid app query parameter
- **WHEN** a client sends a `POST /logs` request with an `app` query parameter that is empty or contains only whitespace
- **THEN** the service SHALL treat the query parameter as not provided and only use the `app` value from the JSON body if present.

### Requirement: Log line formatting
The system SHALL format each accepted log event as a single human-readable text line using a consistent bracketed format, ensuring all content is sanitised and rendered on a single line.

#### Scenario: Formats base line with timestamp, optional app/user, level, and message
- **WHEN** a valid log event is processed
- **THEN** the system SHALL write a line with segments in the following order:
  1. Timestamp segment: `[<timestamp>]` (RFC3339 format)
  2. App segment (if `app` is present and non-empty): ` [<app>]`
  3. User segment (if `user` is present and non-empty): ` [<user>]`
  4. Level segment: ` [<level>]` (normalised to lowercase)
  5. Message: space followed by `<message>`
  6. Optional fields segment (if `fields` is non-empty): ` | <key-value pairs>`

#### Scenario: Sanitises newlines and special characters
- **WHEN** formatting the log line
- **THEN** the system SHALL replace any `\n` or `\r` characters found in `message`, `app`, `user` values, and in string values within `fields` with horizontal tab characters (`\t`) to preserve the one-line-per-event property.

#### Scenario: Appends structured fields as key=value pairs
- **WHEN** a valid log event includes a non-empty `fields` object
- **THEN** the system SHALL append to the base line a space, a vertical bar, and a space (` | `) followed by the event fields rendered as `key=value` pairs sorted lexicographically by key and separated by single spaces, with complex values (objects, arrays) encoded as JSON.

### Requirement: Safe concurrent writes to log file
The system SHALL ensure that concurrent requests do not interleave partial log lines or corrupt the log file.

#### Scenario: Concurrent log events write whole lines
- **WHEN** multiple clients send valid `POST /logs` requests concurrently
- **THEN** the service SHALL serialise writes to the log file such that each event is written as a complete line terminated by a newline, with no interleaving or truncation of individual lines.

### Requirement: Configurable server and log file path
The system SHALL allow basic configuration of the listen port and log file directory via environment variables, with sensible defaults, and SHALL automatically manage daily log file rotation.

#### Scenario: Uses default configuration when env vars are unset
- **WHEN** the service starts and neither `PORT` nor `LOG_DIR` environment variables are set
- **THEN** the service SHALL listen on TCP port `8080` (using address `:8080`) and SHALL write log lines to the directory `./logs`, creating it if it does not exist, with daily log files named in the format `YYYY-MM-DD.log` (e.g., `./logs/2026-02-09.log` for February 9, 2026).

#### Scenario: Uses environment overrides for configuration
- **WHEN** the service starts with `PORT` set to a valid port number string (e.g., `"9090"`) and/or `LOG_DIR` set to a valid filesystem path
- **THEN** the service SHALL parse `PORT` as a numeric string and normalise it by prepending `:` (e.g., `"9090"` → listen on `:9090`), listen on the configured port, and write log lines to the specified `LOG_DIR` directory in daily files with names in the format `YYYY-MM-DD.log`, creating parent directories as needed.

### Requirement: Message-timestamp-based file organisation with current-day optimisation
The system SHALL automatically organise log entries into files based on the UTC date extracted from each message's timestamp. The system SHALL maintain an always-open file handle for the current server date to optimise file I/O, and detect date changes to update this handle.

#### Scenario: Routes events to correct dated file based on message timestamp
- **WHEN** a valid log event with timestamp `2026-02-09T12:00:00Z` is processed (on a server whose current date is 2026-02-09)
- **THEN** the system SHALL extract the UTC date (`2026-02-09`) from the timestamp and append the formatted log line to the file named `2026-02-09.log` in the configured log directory, using the always-open current-day file handle.

#### Scenario: Maintains optimised current-day file handle
- **WHEN** the service processes multiple log events on the same calendar day (e.g., all with timestamps on 2026-02-09)
- **THEN** the system SHALL reuse the same open file handle for all events, avoiding repeated open/close operations.

#### Scenario: Detects date changes and updates current-day file handle
- **WHEN** the server date changes from one UTC day to the next (e.g., from 2026-02-09 to 2026-02-10), detected when a new log event arrives
- **THEN** the system SHALL:
  - Close the file handle for the previous day (if any pending writes are complete).
  - Open a new file handle for the new current date.
  - Continue processing subsequent events with the new file handle.

#### Scenario: Handles events from adjacent days with open-write-close
- **WHEN** valid log events with timestamps from the previous day (2026-02-08) or next day (2026-02-10) are received while the server date is 2026-02-09
- **THEN** the system SHALL:
  - Open the file for the event's date (e.g., `2026-02-08.log` or `2026-02-10.log`).
  - Append the formatted log line to the file.
  - Close the file immediately after writing (no caching).
  - Keep the current-day file handle always open for reuse.

### Requirement: File sink implementation and multi-file management
The system SHALL implement a file sink that safely manages multiple date-based log files and ensures atomic line writes.

#### Scenario: Creates log directory and parent directories at startup
- **WHEN** the service starts with a `LOG_DIR`
- **THEN** the system SHALL ensure the log directory and any parent directories exist (creating them with appropriate permissions if necessary), and open the file handle for today's log file, ready for writing.

#### Scenario: Initializes and maintains current-day file handle
- **WHEN** the service starts
- **THEN** the system SHALL:
  - Determine the current server UTC date.
  - Open and maintain an always-open file handle for today's log file (format: `YYYY-MM-DD.log`).
  - Store the current date internally for future date-change detection.

#### Scenario: Routes events with timestamp validation
- **WHEN** a valid log event with a timestamp within the 3-day window is received
- **THEN** the system SHALL:
  - Compute the target file name from the message's UTC timestamp date (format: `YYYY-MM-DD.log`).
  - If the target date matches the current server date, use the always-open current-day file handle.
  - If the target date is the previous or next day, open the file, write the event, and close it immediately (no caching).
  - Append the formatted log line to the correct file (append mode).

#### Scenario: Detects server date changes and rotates current-day handle
- **WHEN** a log event is received and the current server UTC date differs from the stored current date
- **THEN** the system SHALL:
  - Close the file handle for the stored (previous) current date.
  - Update the stored current date to today.
  - Open a new file handle for the new current date.
  - Process the incoming event with the new file handle.

#### Scenario: Serialises writes and prevents interleaving
- **WHEN** the HTTP handler writes formatted log lines to the file sink
- **THEN** the system SHALL use internal synchronisation (e.g., mutex) to ensure that only one goroutine writes to any particular file at a time, and each write operation appends the line followed by a newline character and flushes to ensure durability.

### Requirement: Testing structure
The system SHALL include unit tests covering core functionality and integration tests verifying end-to-end behaviour.

#### Scenario: Unit test coverage for formatting
- **WHEN** the format module is tested
- **THEN** tests SHALL verify:
  - Correct formatting of events with all required and optional fields
  - Proper handling of newlines and special characters
  - Correct rendering of various field value types
  - Correct lexicographic sorting of field keys

#### Scenario: Unit test coverage for event model
- **WHEN** the model module is tested
- **THEN** tests SHALL verify:
  - Valid timestamp parsing and RFC3339 format validation
  - Level normalisation to lowercase
  - Rejection of invalid timestamps, levels, or empty messages
  - Preservation of optional fields when provided

#### Scenario: Integration test coverage for HTTP handlers
- **WHEN** the HTTP handler module is tested
- **THEN** tests SHALL verify:
  - Well-formed requests produce `202 Accepted` status and write to sink
  - Invalid JSON or missing required fields produce `400 Bad Request` status
  - Requests with non-JSON content type are rejected appropriately
  - Events with different message timestamps are routed to different dated files
  - Concurrent requests with timestamps from the same date write to the same file without interleaving
  - Concurrent requests with timestamps from different dates write to different files correctly

