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
  - Accept and preserve optional `user` and `app` fields only if they are non-empty strings
  - Accept `fields` as an optional object with arbitrary key-value content
  - Normalise `level` to lowercase for consistent formatting

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
The system SHALL allow basic configuration of the listen port and log file location via environment variables, with sensible defaults.

#### Scenario: Uses default configuration when env vars are unset
- **WHEN** the service starts and neither `PORT` nor `LOG_FILE_PATH` environment variables are set
- **THEN** the service SHALL listen on TCP port `8080` (using address `:8080`) and SHALL write log lines to `./logs/app.log`, creating the `logs` directory if it does not exist.

#### Scenario: Uses environment overrides for configuration
- **WHEN** the service starts with `PORT` set to a valid port number string (e.g., `"9090"`) and/or `LOG_FILE_PATH` set to a valid filesystem path
- **THEN** the service SHALL parse `PORT` as a numeric string and normalise it by prepending `:` (e.g., `"9090"` → listen on `:9090`), listen on the configured port, append log lines to the specified `LOG_FILE_PATH` and create parent directories as needed.

### Requirement: File sink implementation and startup
The system SHALL initialise a file sink at startup that safely manages the configured log file and ensures atomic line writes.

#### Scenario: Creates log file and parent directories at startup
- **WHEN** the service starts with a `LOG_FILE_PATH`
- **THEN** the system SHALL ensure all parent directories exist (creating them with appropriate permissions if necessary) and open/create the log file in append mode, ready for writing log lines.

#### Scenario: Serialises file writes to avoid interleaving
- **WHEN** the HTTP handler writes a formatted log line to the file sink
- **THEN** the system SHALL use internal synchronisation (e.g., mutex) to ensure that only one goroutine writes to the file at a time, and each write operation appends the line followed by a newline character and flushes to ensure durability.

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

