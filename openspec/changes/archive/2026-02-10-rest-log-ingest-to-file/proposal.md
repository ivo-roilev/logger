## Summary

Build a small HTTP logging service in Go that exposes a REST endpoint for ingesting JSON log events and appends them to a local file using a predictable, human-readable line format.

Clients will send structured JSON describing each log event. The service will validate and normalise this payload, then write a single formatted text line per event to a configured log file on disk.

## Problem

Today, services that want to centralise basic logs into a shared file often have to:

- Embed file handling logic directly into each application.
- Reimplement common concerns like JSON decoding, validation, and formatting.
- Coordinate conventions (timestamp format, log level naming, extra fields) in an ad-hoc way.

This leads to:

- Inconsistent log formats across services.
- Fragile ad-hoc scripts for parsing logs.
- Tight coupling between business logic and logging concerns.

We want a small, focused logging service that accepts log events over HTTP and is responsible for:

- Validating and normalising incoming events.
- Enforcing a consistent log line format.
- Handling concurrent writes to a log file safely.

## Goals

- **REST ingestion endpoint**: Provide a simple HTTP endpoint (`POST /logs`) that accepts a JSON representation of a single log event, with support for optional URL query parameters.
- **Flexible app identification**: Allow the `app` (application name) to be specified either in the JSON payload or as a query parameter (`?app=<app-name>`), with JSON payload taking precedence when both are provided.
- **Structured input, plain-text output**: Accept structured JSON, but store events as single-line, human-readable text entries in a log file.
- **Deterministic line format**: Use a consistent bracketed format:
  - `[<timestamp>] [<user>] [<app>] [<level>] <message>`, omitting any `[<user>]` or `[<app>]` segments when those fields are not present
  - If there are extra fields, append: ` | key=value key2=value2`
- **Safe concurrent writes**: Ensure concurrent requests cannot interleave partial lines or corrupt the file.
- **Message-timestamp-based file organisation**: Automatically organise log entries into files based on the UTC date of each message's timestamp, with file names formatted as `YYYY-MM-DD.log` (e.g., `2026-02-09.log`). Messages from different dates are written to different files.
- **Timestamp validation window**: Accept only log events whose UTC date is within a 3-day window: the current server date, the previous day, or the next day. Reject events outside this window.
- **Configuration via environment**: Allow configuring the listen port and log file directory via environment variables, with sensible defaults.

## Non-goals

- **Log retention and cleanup**: Deleting old log files, enforcing retention policies, or compressing old logs are explicitly out of scope.
- **Advanced querying / analytics**: This service only appends to files. Searching, indexing, or aggregating logs are out of scope.
- **Automatic time-based rotation**: The service does not rotate files based on system time. File organisation is driven by the timestamp in each message.
- **Authentication and authorisation**: For now, the endpoint is unauthenticated. Network-layer controls or auth can be introduced in a future change.
- **Guaranteed durability beyond the OS**: We rely on the OS and filesystem for durability. We do not implement additional journaling, replication, or queueing.

## High-level approach

- Implement the service in Go using a small HTTP router (e.g., chi) to keep dependencies light.
- Define a JSON payload schema with the following core fields:
  - `timestamp` (string, RFC3339)
  - `level` (string, e.g., `debug`, `info`, `warn`, `error`)
  - `message` (string)
  - `user` (optional string)
  - `app` (optional string)
  - `fields` (optional object with arbitrary key/value pairs)
- Normalise and validate incoming requests:
  - Parse the JSON body into a Go struct.
  - Validate required fields and timestamp format where appropriate.
  - Extract the UTC date from the timestamp and validate it is within the 3-day window (current date ± 1 day). Reject timestamps outside this window.
  - Allow extra top-level fields but ignore them for now.
- Format each event into a single line using the bracketed format, replacing `\n` and `\r` in `message` with tabs to preserve the one-line-per-event property.
- Implement a file sink that:
  - Tracks the current server date (in UTC) and maintains an always-open file handle for today's log file only.
  - When a new log event arrives, checks if the server date has changed since the last event. If it has:
    - Closes the old file handle.
    - Opens a new file handle for the new current date.
  - For events with timestamps from the previous or next day, opens the file, writes the event, and immediately closes the file (no caching). These events are expected to be rare.
  - Extracts the UTC date from each event's timestamp and routes it to the appropriate file.
  - Serialises writes (e.g., using a mutex) so each event is written as an atomic line with a trailing newline.
  - Minimises file handle overhead by keeping only today's file open continuously.
- Expose configuration through environment variables:
  - `PORT` (default `8080`)
  - `LOG_DIR` (default `./logs`) - directory where daily log files will be stored

## Success criteria

- When the service is running locally and a client sends a valid JSON payload with timestamp within the 3-day window (e.g., today's date) to `POST /logs`, exactly one new line appears in the corresponding dated log file in the expected format.
- When a client sends a request with a timestamp outside the 3-day window (e.g., 10 days in the future or past), the service rejects it with a `400 Bad Request` response.
- When the server date changes (e.g., from 2026-02-09 to 2026-02-10), the current-day file handle is automatically closed and a new one is opened for the new day, maintaining file integrity.
- Multiple concurrent clients can send log events without generating interleaved or corrupted log lines, even when writing to multiple date-based files.
- The service is straightforward to run locally with a documented command and example `curl` usage.

