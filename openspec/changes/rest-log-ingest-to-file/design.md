## Overview

This design describes the Go implementation of the logging service:

- Package layout.
- Data models and validation.
- HTTP routing and handlers.
- Formatting and file sink design for safe concurrent writes.

The implementation aims to stay small, explicit, and easy to extend.

## Package structure

Proposed layout under the repository root:

- `cmd/logger-server/main.go`
  - Application entrypoint.
  - Reads configuration (port, log file path) from environment variables.
  - Wires up dependencies (formatter, sink, HTTP handlers).
  - Starts the HTTP server.

- `internal/model/event.go`
  - Defines the request payload struct (log event).
  - Implements JSON binding tags.
  - Provides validation and normalisation functions:
    - Timestamp parsing and RFC3339 validation.
    - Level normalisation to a small enum-like type.
    - Message non-empty checks.

- `internal/format/line.go`
  - Exposes a function:
    - `FormatEvent(e Event) (string, error)`
  - Responsible for:
    - Applying sanitisation rules (newline → tab in strings).
    - Constructing the bracketed base format.
    - Rendering `fields` as sorted key/value pairs.

- `internal/sink/filesink.go`
  - Defines an interface:
    - `type Sink interface { WriteLine(ctx context.Context, line string) error }`
  - `FileSink` implementation:
    - Holds an `*os.File` and a `sync.Mutex`.
    - Provides `WriteLine` which:
      - Locks the mutex.
      - Writes `line + "\n"` using `WriteString`.
      - Unlocks the mutex.
    - Responsible for creating parent directories for the log file path at startup.

- `internal/httpapi/handlers.go`
  - Contains HTTP handler(s) for the `/logs` endpoint.
  - Responsibilities:
    - Enforce `Content-Type: application/json`.
    - Decode JSON body into `model.Event`.
    - Call validation/normalisation.
    - Pass the event to the formatter, then to the sink.
    - Return appropriate HTTP status codes and JSON error responses.

## Data model

In `internal/model/event.go`:

- Define `Event`:

  - Fields:
    - `Timestamp` (`time.Time`)
    - `Level` (custom type, e.g. `LogLevel`)
    - `Message` (`string`)
    - `Fields` (`map[string]any`)

- JSON binding:
  - Use a separate struct for decoding raw JSON:
    - `timestamp` as `string`.
    - `level` as `string`.
    - `message` as `string`.
    - `fields` as `map[string]any`.
  - Provide `FromRequestPayload` / `Normalize` helper that:
    - Parses the timestamp string into `time.Time` (RFC3339).
    - Normalises `level` to a known set (`debug`, `info`, `warn`, `error`).
    - Checks that `message` is non-empty after trimming.

- Level type:
  - Use:
    - `type LogLevel string`
  - Constants for allowed values.
  - Implement `String()` to use in formatting.

## HTTP layer

In `cmd/logger-server/main.go`:

- Initialise router using chi:
  - `r := chi.NewRouter()`
  - Mount `POST /logs` handler.

- Build handler dependencies:
  - Construct a `FileSink` with the configured file path.
  - Configure a formatter from `internal/format`.

In `internal/httpapi/handlers.go`:

- Define:
  - `type LoggerHandler struct { Sink Sink; Formatter Formatter }`
  - `Formatter` interface:
    - `FormatEvent(e model.Event) (string, error)`

- Implement:
  - `func (h *LoggerHandler) PostLog(w http.ResponseWriter, r *http.Request)`
    - Check method and content type.
    - Decode JSON body.
    - Normalise into `model.Event`.
    - Call formatter.
    - Write formatted line via sink.
    - Return `202` with `{ "status": "ok" }` on success.

## Formatting

In `internal/format/line.go`:

- `FormatEvent` logic:
  - Sanitise `Message`:
    - Replace `\n` and `\r` with `\t`.
  - Build base string:
    - `[<timestamp>] [<level>] <message>`
  - If `Fields` is non-empty:
    - Extract keys, sort lexicographically.
    - For each key:
      - Sanitise string values (newline → tab).
      - Convert non-string primitives using `%v`.
      - Convert complex values via `encoding/json` (`Marshal`) then sanitise.
    - Append: ` | key1=value1 key2=value2 ...`.

## File sink and concurrency

- `FileSink`:
  - Created at startup with a resolved file path:
    - Ensure parent directory exists using `os.MkdirAll`.
    - Open file with `os.OpenFile` in append mode (`O_APPEND|O_CREATE|O_WRONLY`).
  - Holds:
    - `file *os.File`
    - `mu   sync.Mutex`
  - `WriteLine`:
    - Lock `mu`.
    - Write `line + "\n"` to `file`.
    - Unlock `mu`.

- This ensures:
  - Only one goroutine writes to the file at a time.
  - Lines are not interleaved or truncated.

- Shutdown:
  - On process termination, the main function is responsible for closing the file.

## Configuration

- Environment variables:
  - `PORT`:
    - Read as string.
    - If empty, default to `":8080"`.
    - If present and numeric, normalise to `":" + value`.
  - `LOG_FILE_PATH`:
    - Default: `./logs/app.log`.
    - Passed to `FileSink` for directory creation and opening.

## Testing strategy

- Unit tests:
  - `internal/format/line_test.go`:
    - Verify formatting of events with:
      - All required fields.
      - Various `fields` combinations.
      - Newlines in `message` and field values.
  - `internal/model/event_test.go`:
    - Validate timestamp parsing and level normalisation.
    - Ensure invalid inputs are rejected as expected.

- HTTP handler tests:
  - `internal/httpapi/handlers_test.go`:
    - Use `httptest` to:
      - Send a well-formed request and assert `202` and sink call.
      - Send invalid JSON / missing fields / invalid timestamp and assert `400`.
    - Use a fake sink in tests (in-memory slice or buffer).

