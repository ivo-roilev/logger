## Implementation tasks

- [ ] **Scaffold Go module**
  - Initialise `go.mod` for the project.
  - Add dependency on `github.com/go-chi/chi/v5`.

- [ ] **Implement event model and validation**
  - Create `internal/model/event.go`.
  - Define request payload structures and JSON tags.
  - Implement timestamp parsing (RFC3339) and level normalisation.
  - Add validation for required fields (`timestamp`, `level`, `message`).
  - Include optional `user` and `app` string fields in the data model.

- [ ] **Implement line formatter**
  - Create `internal/format/line.go`.
  - Implement `FormatEvent` to:
    - Build the base `[timestamp] [app] [user] [level] message` string, omitting `[app]` and `[user]` segments when those fields are not present.
    - Append sorted `fields` as `key=value` pairs.
    - Apply sanitisation (newline → tab) for messages, user, app, and string values.

- [ ] **Implement file sink**
  - Create `internal/sink/filesink.go`.
  - Define `Sink` interface and `FileSink` implementation.
  - Ensure directories are created for `LOG_FILE_PATH`.
  - Implement mutex-guarded `WriteLine` to prevent interleaved writes.

- [ ] **Implement HTTP handler**
  - Create `internal/httpapi/handlers.go`.
  - Implement handler for `POST /logs`:
    - Extract optional `app` query parameter from URL.
    - Enforce `Content-Type: application/json`.
    - Decode and validate the JSON payload.
    - If `app` is not in the JSON body and a valid `app` query parameter is provided, use the query parameter value.
    - Use formatter and sink to write the log line.
    - Return `202` with `{ "status": "ok" }` on success, appropriate errors otherwise.

- [ ] **Wire server entrypoint**
  - Create `cmd/logger-server/main.go`.
  - Read `PORT` and `LOG_FILE_PATH` from environment with defaults.
  - Initialise `FileSink`, formatter, and HTTP handler.
  - Start HTTP server with chi router.

- [ ] **Add unit tests**
  - Add tests for formatter (`internal/format/line_test.go`).
  - Add tests for model validation (`internal/model/event_test.go`).
  - Add handler tests using `httptest` (`internal/httpapi/handlers_test.go`) with a fake sink:
    - Test app from JSON payload.
    - Test app from query parameter when not in JSON.
    - Test JSON payload app takes precedence over query parameter.
    - Test empty/whitespace app query parameter is ignored.

- [ ] **Add README and usage examples**
  - Create `README.md` with:
    - Setup instructions (`go run ./cmd/logger-server`).
    - Configuration options (`PORT`, `LOG_FILE_PATH`).
    - Example `curl` commands:
      - With `app` in JSON body: `curl -X POST http://localhost:8080/logs -H "Content-Type: application/json" -d '{"timestamp":"2026-02-09T12:00:00Z","level":"info","message":"Hello","app":"myapp"}'`
      - With `app` in query parameter: `curl -X POST "http://localhost:8080/logs?app=myapp" -H "Content-Type: application/json" -d '{"timestamp":"2026-02-09T12:00:00Z","level":"info","message":"Hello"}'`
    - Sample log file output.

