## Implementation tasks

- [ ] **Scaffold Go module**
  - Initialise `go.mod` for the project.
  - Add dependency on `github.com/go-chi/chi/v5`.

- [ ] **Implement event model and validation**
  - Create `internal/model/event.go`.
  - Define request payload structures and JSON tags.
  - Implement timestamp parsing (RFC3339) and level normalisation.
  - Add validation for required fields (`timestamp`, `level`, `message`).
  - Add 3-day window validation: reject timestamps more than 1 day in the past or future.
  - Include optional `user` and `app` string fields in the data model.

- [ ] **Implement line formatter**
  - Create `internal/format/line.go`.
  - Implement `FormatEvent` to:
    - Build the base `[timestamp] [app] [user] [level] message` string, omitting `[app]` and `[user]` segments when those fields are not present.
    - Append sorted `fields` as `key=value` pairs.
    - Apply sanitisation (newline → tab) for messages, user, app, and string values.

- [ ] **Implement file sink with current-day file handle optimisation**
  - Create `internal/sink/filesink.go`.
  - Define `Sink` interface with `WriteLine(ctx context.Context, line string, timestamp time.Time) error`.
  - Implement `FileSink`:
    - Maintain an always-open file handle for the current server date.
    - On each write, detect if the server date has changed:
      - If changed, close the old current-day file and open a new one.
    - Extract the message timestamp's UTC date and route to the appropriate file.
    - For the current date, use the always-open file handle.
    - For adjacent dates (previous/next day), open the file, write the line, and immediately close it (no caching).
    - For dates within the 3-day window, route correctly; validation ensures only valid dates arrive.
  - Ensure the log directory is created for `LOG_DIR`.
  - Implement mutex-guarded writes to prevent interleaved writes to the same file.

- [ ] **Implement HTTP handler**
  - Create `internal/httpapi/handlers.go`.
  - Implement handler for `POST /logs`:
    - Extract optional `app` query parameter from URL.
    - Enforce `Content-Type: application/json`.
    - Decode and validate the JSON payload.
    - If `app` is not in the JSON body and a valid `app` query parameter is provided, use the query parameter value.
    - Format the event into a single line.
    - Pass the formatted line and the event's timestamp to the sink so it can route to the correct dated file.
    - Return `202` with `{ "status": "ok" }` on success, appropriate errors otherwise.

- [ ] **Wire server entrypoint**
  - Create `cmd/logger-server/main.go`.
  - Read `PORT` and `LOG_DIR` from environment with defaults.
  - Initialise `FileSink` with the log directory, formatter, and HTTP handler.
  - Start HTTP server with chi router.

- [ ] **Add unit tests**
  - Add tests for formatter (`internal/format/line_test.go`).
  - Add tests for model validation (`internal/model/event_test.go`):
    - Test timestamp parsing and level normalisation.
    - Test 3-day window validation: accept current ± 1 day, reject beyond ±1 day.
  - Add file sink tests (`internal/sink/filesink_test.go`):
    - Test current-day file handle: verify it's opened at startup and reused for multiple events on the same day.
    - Test date change detection: simulate date change and verify file rotation.
    - Test 3-day window file routing (timestamps on current ± 1 day go to correct files).
    - Test file handle caching for previous/next day.
    - Test concurrent writes to the same dated file (no interleaving).
    - Test concurrent writes to different dated files (isolation).
    - Test concurrent access during date change.
  - Add handler tests using `httptest` (`internal/httpapi/handlers_test.go`) with a fake sink:
    - Test app from JSON payload.
    - Test app from query parameter when not in JSON.
    - Test JSON payload app takes precedence over query parameter.
    - Test empty/whitespace app query parameter is ignored.
    - Test that events with timestamps outside the 3-day window are rejected with `400`.
    - Test that valid events within the 3-day window reach the sink with correct timestamps.

- [ ] **Add README and usage examples**
  - Create `README.md` with:
    - Setup instructions (`go run ./cmd/logger-server`).
    - Configuration options (`PORT`, `LOG_DIR`) and explanation of 3-day window validation and message-timestamp-based file organisation.
    - Description of how events are routed: each event is written to a file named after the UTC date from its timestamp (e.g., timestamp `2026-02-09T12:00:00Z` goes to `2026-02-09.log`).
    - Note that only timestamps within current ± 1 day are accepted.
    - Description of current-day file handle optimisation and automatic date change handling.
    - Example `curl` commands:
      - With `app` in JSON body: `curl -X POST http://localhost:8080/logs -H "Content-Type: application/json" -d '{"timestamp":"2026-02-09T12:00:00Z","level":"info","message":"Hello","app":"myapp"}'`
      - With `app` in query parameter: `curl -X POST "http://localhost:8080/logs?app=myapp" -H "Content-Type: application/json" -d '{"timestamp":"2026-02-09T12:00:00Z","level":"info","message":"Hello"}'`
    - Sample log file output showing events in dated log files (`./logs/2026-02-09.log`, `./logs/2026-02-10.log`, etc.).

