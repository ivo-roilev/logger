## Overview

This design describes the Go implementation of the logging service:

- Package layout.
- Data models and validation (including 3-day timestamp window validation).
- HTTP routing and handlers.
- Formatting and file sink design for safe concurrent writes with current-day file handle optimisation.

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
    - `User` (`string`, optional)
    - `App` (`string`, optional)
    - `Fields` (`map[string]any`)

- JSON binding:
  - Use a separate struct for decoding raw JSON:
    - `timestamp` as `string`.
    - `level` as `string`.
    - `message` as `string`.
    - `user` as `string` (optional, omitempty).
    - `app` as `string` (optional, omitempty).
    - `fields` as `map[string]any`.
  - Provide `FromRequestPayload` / `Normalize` helper that:
    - Parses the timestamp string into `time.Time` (RFC3339).
    - Validates the timestamp is within a 3-day window (current UTC date \u00b1 1 day):
      - Extract the UTC date from the parsed timestamp.
      - Extract the current UTC date from the server.
      - Reject if the timestamp date is more than 1 day before or after the current date.
    - Normalises `level` to a known set (`debug`, `info`, `warn`, `error`).
    - Checks that `message` is non-empty after trimming.
    - Preserves `user` and `app` values if present.

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
  - `Sink` interface:
    - `WriteLine(ctx context.Context, line string, timestamp time.Time) error`
  - `Formatter` interface:
    - `FormatEvent(e model.Event) (string, error)`

- Implement:
  - `func (h *LoggerHandler) PostLog(w http.ResponseWriter, r *http.Request)`
    - Extract optional `app` query parameter from URL (e.g., `?app=checkout-service`).
    - Check method and content type.
    - Decode JSON body.
    - If `app` is not present in the JSON body and an `app` query parameter is provided and non-empty, use the query parameter value.
    - Normalise into `model.Event`.
    - Call formatter to get the formatted line.
    - Pass the formatted line along with the event's timestamp to the sink.
    - The sink will determine the correct dated file based on the timestamp.
    - Return `202` with `{ "status": "ok" }` on success.

## Formatting

In `internal/format/line.go`:

- `FormatEvent` logic:
  - Sanitise `Message`:
    - Replace `\n` and `\r` with `\t`.
  - Start with the timestamp segment:
    - `[<timestamp>]`
  - If `App` is present and non-empty:
    - Sanitise value (newline → tab).
    - Append: ` [<app>]`.
  - If `User` is present and non-empty:
    - Sanitise value (newline → tab).
    - Append: ` [<user>]`.
  - Append the level segment:
    - ` [<level>]`.
  - Append a space and the message:
    - ` <message>`.
  - If `Fields` is non-empty:
    - Extract keys, sort lexicographically.
    - For each key:
      - Sanitise string values (newline → tab).
      - Convert non-string primitives using `%v`.
      - Convert complex values via `encoding/json` (`Marshal`) then sanitise.
    - Append: ` | key1=value1 key2=value2 ...`.

## File sink and message-timestamp-based file organisation with current-day optimisation

- `FileSink`:
  - Created at startup with a log directory path:
    - Ensure parent directory exists using `os.MkdirAll`.
    - Determine the current server UTC date.
    - Open and store the file handle for today (the current-day file).
  - Holds:
    - `logDir string` - configured log directory
    - `currentDayStr string` - the current server date in format `YYYY-MM-DD`
    - `currentDayFile *os.File` - the always-open file handle for today
    - `mu sync.Mutex` - protects access to the file handle and currentDayStr
  - `WriteLine(ctx context.Context, line string, timestamp time.Time)`:
    - Lock `mu`.
    - Extract the UTC date from `timestamp` (format: `YYYY-MM-DD`).
    - Determine today's UTC date by checking the current system date.
    - If today's date differs from `currentDayStr`:
      - Close `currentDayFile`.
      - Update `currentDayStr` to today.
      - Open a new file handle for today and store in `currentDayFile`.
    - If the target date matches `currentDayStr`, use `currentDayFile` (write and keep open).
    - Otherwise (adjacent day or outside validation window - should not occur due to model validation):
      - Open the file for the target date, write `line + "\n"`, flush, and immediately close it (no caching).
      - Do not cache or keep the file handle open.
    - Unlock `mu`.

- This ensures:
  - The current day's file is always open for efficient reuse.
  - Date changes are detected automatically when events arrive.
  - Old file handles are properly closed when the date changes.
  - Events from adjacent days (which are rare due to the 3-day validation window) use open-write-close, avoiding unnecessary caching.
  - Only one goroutine writes to any particular file at a time (mutex-protected).
  - Lines are not interleaved or truncated.

- Helper function:
  - `func todayDateString() string`:
    - Returns the current UTC date in format `YYYY-MM-DD`.

- Helper function:
  - `func extractDateString(timestamp time.Time) string`:
    - Formats the UTC date as `YYYY-MM-DD`.
    - Example: `2026-02-09T12:00:00Z` → `2026-02-09`.

- Helper function:
  - `func dateFilePath(dir string, dateStr string) string`:
    - Returns the full path: `dir/dateStr.log`.

- Shutdown:
  - On process termination, the main function should close currentDayFile.

## Configuration

- Environment variables:
  - `PORT`:
    - Read as string.
    - If empty, default to `":8080"`.
    - If present and numeric, normalise to `":" + value`.
  - `LOG_DIR`:
    - Default: `./logs`.
    - Passed to `FileSink` for directory creation and daily log file management.

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
    - Test 3-day window validation:
      - Accept timestamps for current date, previous day, and next day.
      - Reject timestamps more than 1 day in the past or future (e.g., 2 days ago, 2 days from now).

- HTTP handler tests:
  - `internal/httpapi/handlers_test.go`:
    - Use `httptest` to:
      - Send a well-formed request with a valid timestamp and assert `202` and sink call.
      - Send invalid JSON / missing fields / invalid timestamp and assert `400`.
      - Send a request with a timestamp outside the 3-day window and assert `400` with an appropriate error message.
      - Send a request with `app` in query parameter when JSON body does not include `app`, and verify the query parameter value is used.
      - Send a request with `app` in both query parameter and JSON body, and verify JSON body value takes precedence.
      - Send a request with an empty or whitespace-only `app` query parameter, and verify it is ignored.
    - Use a fake sink in tests (in-memory slice or buffer).

- File sink tests:
  - `internal/sink/filesink_test.go`:
    - Verify current-day file handle management:
      - At startup, verify that today's file handle is opened and ready.
      - Write multiple events on the same day, verifying the same file handle is reused.
    - Verify 3-day window file handling:
      - Write events with timestamps from today, yesterday, and tomorrow.
      - Verify they go to the correct dated files.
      - Verify file handles for adjacent days are created and closed immediately (open-write-close pattern, no caching).
    - Verify date change detection and rotation:
      - Mock the system date or use a testable time package to simulate date changes.
      - Write events as day 1, verify they use the current-day file.
      - Simulate the date changing to day 2 (via mocking).
      - Write a new event, verify the old file is closed and a new file is opened.
      - Verify file integrity (no data loss during rotation).
    - Verify concurrent access safety:
      - Test concurrent writes to the same dated file (no interleaving).
      - Test concurrent writes to different dated files (isolation).
      - Test concurrent writes during simulated date changes.

