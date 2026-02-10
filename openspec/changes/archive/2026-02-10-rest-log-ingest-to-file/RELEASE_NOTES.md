# Release Notes — rest-log-ingest-to-file (2026-02-10)

**Summary:** Implemented a file-based logging ingest service and archived the OpenSpec change rest-log-ingest-to-file.

**Implemented features:**
- **Routing:** Message-timestamp-based routing to daily files named YYYY-MM-DD.log (UTC date extracted from event timestamp).
- **Validation window:** Accept events only when the UTC date of the event timestamp is within ±1 day of server date (3-day window). Date-only comparison is used.
- **File sink policy:** Always-open file handle for the current server date; open-write-close for previous/next-day events.
- **Format:** Single-line bracketed format: timestamp, optional app, optional user, level, message, then ` | ` followed by sorted key=value fields. Newlines in strings are replaced with tabs.
- **HTTP API:** POST /logs accepting application/json with fields `timestamp`, `level`, `message`, optional `user`, `app`, and `fields`; `app` may also be provided via `?app=` (JSON body wins).
- **Concurrency & durability:** Mutex-guarded writes and file `Sync` after writes to avoid interleaving and ensure durability.

**Files added/changed (implementation & specs):**
- [internal/model/event.go](internal/model/event.go)
- [internal/format/line.go](internal/format/line.go)
- [internal/sink/filesink.go](internal/sink/filesink.go)
- [internal/httpapi/handlers.go](internal/httpapi/handlers.go)
- [cmd/logger-server/main.go](cmd/logger-server/main.go)
- [specs/rest-log-ingest-to-file/spec.md](specs/rest-log-ingest-to-file/spec.md)

**Tests:**
- Unit tests added/updated for model, formatter, sink, and handlers. Run `go test ./...` — all tests pass in the repository at the time of archiving.

**Configuration & run:**
- `PORT` environment variable (default `:8080` when unset or `8080` numeric string normalized).
- `LOG_DIR` environment variable (default `./logs`).
- Build & run:
```
go build ./cmd/logger-server
./logger-server
```
or
```
go run ./cmd/logger-server
```

**Backward-compatibility / notes:**
- No breaking API changes to clients were introduced by this change; clients must provide RFC3339 timestamps and allowed `level` values.
- Because routing is determined by message timestamps, events may be written into prior/next-day files based on the supplied timestamp.

**Archive:**
- OpenSpec change directory moved to: [openspec/changes/archive/2026-02-10-rest-log-ingest-to-file](openspec/changes/archive/2026-02-10-rest-log-ingest-to-file)

**Contact / Author:** Ivo Roylev
