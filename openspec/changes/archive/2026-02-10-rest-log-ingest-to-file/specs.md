## Overview

This document specifies the external behaviour of the logging service:

- HTTP API surface.
- JSON payload schema.
- Log line format and encoding rules.
- Configuration and operational expectations.

## HTTP API

### Endpoint

- **Method**: `POST`
- **Path**: `/logs`
- **Request body**: JSON (UTF-8)
- **Response codes**:
  - `202 Accepted` on success.
  - `400 Bad Request` for invalid JSON or schema violations.
  - `500 Internal Server Error` for unexpected failures while processing or writing.

### Request headers

- `Content-Type: application/json` is required.

If `Content-Type` is missing or not `application/json` (case-insensitive), the server **must** treat the request as invalid and return `400 Bad Request`.

## JSON payload schema

The service accepts a single log event per request. The canonical schema is:

```json
{
  "timestamp": "2026-02-09T12:34:56Z",
  "level": "info",
  "message": "User logged in",
  "fields": {
    "user_id": "123",
    "ip": "203.0.113.42"
  }
}
```

### Fields

- **`timestamp`** (required)
  - Type: string.
  - Must be a valid RFC3339 timestamp (e.g., `2026-02-09T12:34:56Z`).
  - If parsing fails, the request is invalid (`400 Bad Request`).

- **`level`** (required)
  - Type: string.
  - Case-insensitive accepted values (normalised internally to lowercase for output):
    - `debug`, `info`, `warn`, `error`.
  - If an unsupported value is provided, the request is invalid (`400 Bad Request`).

- **`message`** (required)
  - Type: string.
  - Must be non-empty after trimming whitespace.
  - May contain arbitrary characters including newlines; these are handled during formatting (see below).

- **`fields`** (optional)
  - Type: object/dictionary of additional attributes.
  - Keys:
    - Type: string, non-empty.
    - Recommended: snake_case or lowerCamelCase, but not enforced.
  - Values:
    - Allowed types: string, number, boolean, null.
    - Complex values (arrays or nested objects) must be stringified using JSON encoding before formatting.

- **Additional top-level fields**
  - The server must ignore any additional top-level fields not defined above. They must not cause a validation error.

### Error responses

On error the response body should be JSON:

```json
{
  "error": "description of the problem"
}
```

Examples:

- Invalid JSON: `{"error": "invalid JSON body"}`
- Missing required field: `{"error": "missing field: message"}`
- Invalid timestamp: `{"error": "invalid timestamp: must be RFC3339"}`

The exact error strings are not part of the stable API but should be clear and human-readable.

## Log line format

Each accepted event is written as exactly one line to the configured log file.

### Base format

- Template:

  - `[<timestamp>] [<level>] <message>`

- `timestamp`:
  - The `timestamp` field, after successful RFC3339 parsing and normalisation, is emitted as-is (in RFC3339 format).

- `level`:
  - The `level` field is lowercased before output (e.g., `INFO` → `info`).

- `message`:
  - The `message` field is used as free text with sanitisation rules below.

### Extra fields

If `fields` is present and non-empty, the line is extended with a delimiter and key/value pairs:

- Suffix template:

  - ` | key1=value1 key2=value2 ...`

- Ordering:
  - Keys should be rendered in deterministic order (e.g., sorted lexicographically) to keep output stable between runs.

- Key formatting:
  - Keys are written as provided (no additional quoting), assuming they do not contain spaces.

- Value formatting:
  - Primitive types:
    - string: written as-is after sanitisation (see below).
    - number: written using Go's default `%v` formatting.
    - boolean: written as `true` or `false`.
    - null: written as `null`.
  - Non-primitive (arrays, objects):
    - Stringify using compact JSON (no spaces) and then sanitise.

### Sanitisation rules

To preserve the **one-event-per-line** property:

- In the `message` and all string field values:
  - Replace `\n` and `\r` characters with a single tab character (`\t`).
  - Optionally trim trailing whitespace.

No additional escaping is performed for spaces; messages and values may contain spaces freely.

### Examples

Given the payload:

```json
{
  "timestamp": "2026-02-09T12:34:56Z",
  "level": "INFO",
  "message": "User logged in",
  "fields": {
    "user_id": "123",
    "ip": "203.0.113.42"
  }
}
```

The log line must be:

```text
[2026-02-09T12:34:56Z] [info] User logged in | ip=203.0.113.42 user_id=123
```

(assuming keys are sorted alphabetically).

With no `fields`:

```json
{
  "timestamp": "2026-02-09T12:34:56Z",
  "level": "error",
  "message": "Unhandled exception"
}
```

The log line must be:

```text
[2026-02-09T12:34:56Z] [error] Unhandled exception
```

## Configuration

- **Environment variables**:
  - `PORT`
    - Default: `9090`.
    - Type: string representing a TCP port (e.g., `":9090"` or `9090` depending on implementation; the service will document the exact format).
  - `LOG_FILE_PATH`
    - Default: `./logs/app.log`.
    - If the parent directory does not exist, the service must attempt to create it on startup.

If `LOG_FILE_PATH` cannot be opened or created at startup, the service should log an error to stderr and fail fast rather than accepting requests it cannot persist.

## Concurrency and durability expectations

- The service must support concurrent HTTP requests. Behavioural guarantees:
  - Each accepted request results in **exactly one** new line appended to the log file.
  - Log lines must not be interleaved or truncated, even under concurrent load.
  - Order of lines in the file reflects the order in which write operations are performed, which typically approximates, but does not strictly guarantee, event timestamp order.

- Durability:
  - The service relies on the filesystem's semantics for durability.
  - Flushing (e.g., via `Sync`) may be performed periodically or deferred; strict on-disk durability after each request is not required for this change.

