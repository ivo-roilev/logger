# Release draft — rest-log-ingest-to-file (2026-02-10)

## Title
rest-log-ingest-to-file — File-based logging ingest service

## Summary
Implemented a lightweight HTTP ingest service that writes JSON log events to date-based files. The service routes events by the UTC date extracted from each event's timestamp and keeps an always-open handle for the current day while using open-write-close for adjacent-day events.

## Key features
- Message-timestamp-based routing to daily files named `YYYY-MM-DD.log`.
- 3-day validation window: accept events whose UTC date is current ±1 day (date-only comparison).
- File sink policy: always-open handle for today; open-write-close for prev/next days.
- Single-line, bracketed log format with sorted `key=value` fields and newline-to-tab sanitisation.
- `POST /logs` HTTP API accepting `application/json` payloads (`timestamp`, `level`, `message`, optional `user`, `app`, `fields`). `app` can also be passed via `?app=`; JSON body wins.
- Safe concurrent writes with mutexes and `Sync` after writes.

## Files changed / added
- internal/model/event.go
- internal/format/line.go
- internal/sink/filesink.go
- internal/httpapi/handlers.go
- cmd/logger-server/main.go
- specs/rest-log-ingest-to-file/spec.md

## Tests
Unit tests for model, format, sink, and handlers were added/updated. Run `go test ./...` — all tests passed at archive time.

## How to run
```
go build ./cmd/logger-server
./logger-server
```
or
```
go run ./cmd/logger-server
```

## Notes
- No breaking API changes. Clients must supply RFC3339 timestamps and valid levels.
- Because routing uses message timestamps, events may land in prior/next-day files depending on the supplied timestamp.

## Suggested tag
v0.1.0

## Author
Ivo Roylev
