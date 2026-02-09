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

- **REST ingestion endpoint**: Provide a simple HTTP endpoint (`POST /logs`) that accepts a JSON representation of a single log event.
- **Structured input, plain-text output**: Accept structured JSON, but store events as single-line, human-readable text entries in a log file.
- **Deterministic line format**: Use a consistent bracketed format:
  - `[<timestamp>] [<level>] <message>`
  - If there are extra fields, append: ` | key=value key2=value2`
- **Safe concurrent writes**: Ensure concurrent requests cannot interleave partial lines or corrupt the file.
- **Configuration via environment**: Allow configuring the listen port and log file path via environment variables, with sensible defaults.

## Non-goals

- **Log rotation and retention**: Rotating files, enforcing retention policies, or compressing old logs are explicitly out of scope for this change.
- **Advanced querying / analytics**: This service only appends to a file. Searching, indexing, or aggregating logs are out of scope.
- **Authentication and authorisation**: For now, the endpoint is unauthenticated. Network-layer controls or auth can be introduced in a future change.
- **Guaranteed durability beyond the OS**: We rely on the OS and filesystem for durability. We do not implement additional journaling, replication, or queueing.

## High-level approach

- Implement the service in Go using a small HTTP router (e.g., chi) to keep dependencies light.
- Define a JSON payload schema with the following core fields:
  - `timestamp` (string, RFC3339)
  - `level` (string, e.g., `debug`, `info`, `warn`, `error`)
  - `message` (string)
  - `fields` (optional object with arbitrary key/value pairs)
- Normalise and validate incoming requests:
  - Parse the JSON body into a Go struct.
  - Validate required fields and timestamp format where appropriate.
  - Allow extra top-level fields but ignore them for now.
- Format each event into a single line using the bracketed format, replacing `\n` and `\r` in `message` with tabs to preserve the one-line-per-event property.
- Implement a file sink that:
  - Opens (or creates) the configured log file.
  - Serialises writes (e.g., using a mutex) so each event is written as an atomic line with a trailing newline.
  - Can be reused across HTTP requests without reopening the file each time.
- Expose configuration through environment variables:
  - `PORT` (default `8080`)
  - `LOG_FILE_PATH` (default `./logs/app.log`)

## Success criteria

- When the service is running locally and a client sends a valid JSON payload to `POST /logs`, exactly one new line appears in the configured log file in the expected format.
- Multiple concurrent clients can send log events without generating interleaved or corrupted log lines.
- The service is straightforward to run locally with a documented command and example `curl` usage.

