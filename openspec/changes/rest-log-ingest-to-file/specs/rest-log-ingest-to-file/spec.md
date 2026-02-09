## ADDED Requirements

### Requirement: REST log ingestion endpoint
The system SHALL expose an HTTP endpoint that accepts JSON log events over HTTP and appends them to a local log file.

#### Scenario: Accepts valid log event
- **WHEN** a client sends an HTTP `POST` request to `/logs` with a `Content-Type: application/json` header and a well-formed JSON body containing `timestamp` (RFC3339 string), `level` (one of `debug`, `info`, `warn`, `error`), `message` (non-empty string), and optional `fields` (object)
- **THEN** the service SHALL respond with HTTP status `202 Accepted` and a JSON body containing `{ "status": "ok" }`, and SHALL append exactly one new line to the configured log file representing the event.

#### Scenario: Rejects invalid JSON payload
- **WHEN** a client sends a `POST /logs` request with a `Content-Type: application/json` header but the request body is not valid JSON or does not match the expected schema (e.g., missing `timestamp`, `level`, or `message`)
- **THEN** the service SHALL respond with an HTTP `4xx` status (e.g., `400 Bad Request`) and a JSON error body describing the validation problem, and SHALL NOT append any line to the log file.

### Requirement: Log line formatting
The system SHALL format each accepted log event as a single human-readable text line using a consistent bracketed format and field encoding.

#### Scenario: Formats base line with timestamp, level, and message
- **WHEN** a valid log event is processed
- **THEN** the system SHALL write a line in the form `[<timestamp>] [<level>] <message>` where `<timestamp>` is the RFC3339 timestamp string, `<level>` is the normalised log level, and `<message>` is the event message with any `\n` or `\r` characters replaced by tab characters.

#### Scenario: Appends structured fields as key=value pairs
- **WHEN** a valid log event includes a non-empty `fields` object
- **THEN** the system SHALL append to the base line a space, a vertical bar, and a space (` | `) followed by the event fields rendered as `key=value` pairs sorted lexicographically by key and separated by single spaces, sanitising any newline characters in string values to tabs.

### Requirement: Safe concurrent writes to log file
The system SHALL ensure that concurrent requests do not interleave partial log lines or corrupt the log file.

#### Scenario: Concurrent log events write whole lines
- **WHEN** multiple clients send valid `POST /logs` requests concurrently
- **THEN** the service SHALL serialise writes to the log file such that each event is written as a complete line terminated by a newline, with no interleaving or truncation of individual lines.

### Requirement: Configurable server and log file path
The system SHALL allow basic configuration of the listen port and log file location via environment variables, with sensible defaults.

#### Scenario: Uses default configuration when env vars are unset
- **WHEN** the service starts and neither `PORT` nor `LOG_FILE_PATH` environment variables are set
- **THEN** the service SHALL listen on TCP port `8080` and SHALL write log lines to `./logs/app.log`, creating the `logs` directory if it does not exist.

#### Scenario: Uses environment overrides for configuration
- **WHEN** the service starts with `PORT` set to a valid port number string and/or `LOG_FILE_PATH` set to a valid filesystem path
- **THEN** the service SHALL listen on the configured port and SHALL append log lines to the specified file path, creating parent directories as needed.

