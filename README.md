# Logger Service

A Go HTTP logging service that accepts REST API requests and routes log events to date-based log files based on message timestamps.

## Features

- **REST API endpoint** for log ingestion: `POST /logs`
- **RFC3339 timestamp-based routing**: Events are written to log files named after the UTC date of the message timestamp (e.g., `2026-02-09.log`)
- **3-day validation window**: Only accepts log events with timestamps within ±1 day of the current server date (prevents unbounded file growth)
- **Current-day file handle optimization**: Always keeps today's log file open for efficient writes
- **Adjacent day handling**: For events from yesterday or tomorrow, files are opened, written to, and immediately closed
- **Flexible app identification**: App name can be provided in the JSON body or as a query parameter (JSON takes precedence)
- **Concurrent write safety**: Mutex-protected writes ensure no interleaving of log lines

## Setup & Running

### Prerequisites

- Go 1.22 or later
- `github.com/go-chi/chi/v5` dependency (included in go.mod)

### Build

```bash
go build ./cmd/logger-server
```

### Run

```bash
# Default port (9090) and log directory (./logs)
go run ./cmd/logger-server

# Custom configuration
PORT=3000 LOG_DIR=/var/log/app ./logger-server
```

## Configuration

### Environment Variables

- `PORT` (default: `9090`)
  - HTTP server port. If provided without a colon, will be prefixed with `:`.
  - Examples: `PORT=3000`, `PORT=:9000`.

- `LOG_DIR` (default: `./logs`)
  - Directory where dated log files will be created.
  - The service will create the directory and any parent directories as needed.

## API Usage

### Endpoint: POST /logs

Accepts JSON log events and writes them to dated log files.

#### Request

**Headers:**
- `Content-Type: application/json` (required)

**Query Parameters:**
- `app` (optional): Application name. Overridden by `app` field in JSON body.

**JSON Body:**
```json
{
  "timestamp": "2026-02-09T14:30:45Z",
  "level": "info",
  "message": "User login successful",
  "user": "alice",
  "app": "auth-service",
  "fields": {
    "user_id": "12345",
    "ip_address": "203.0.113.42"
  }
}
```

**Field Details:**
- `timestamp` (required): RFC3339 formatted UTC timestamp. Must be within ±1 day of current server date.
- `level` (required): Log level. One of: `debug`, `info`, `warn`, `error`.
- `message` (required): Log message. Cannot be empty.
- `user` (optional): User identifier.
- `app` (optional): Application identifier. Can be overridden by query parameter if not provided.
- `fields` (optional): Additional key-value fields. Values can be strings, numbers, booleans, or objects.

#### Response

**Success (202 Accepted):**
```json
{
  "status": "ok"
}
```

**Error (400 Bad Request):**
```json
{
  "error": "invalid JSON body"
}
```

Returns 400 for:
- Invalid JSON format
- Missing required fields
- Invalid timestamp format
- Timestamp outside 3-day window
- Invalid log level
- Empty message

### Example Requests

#### With app in JSON body:
```bash
curl -X POST http://localhost:9090/logs \
  -H "Content-Type: application/json" \
  -d '{
    "timestamp": "2026-02-09T14:30:00Z",
    "level": "info",
    "message": "Application started",
    "app": "myservice"
  }'
```

#### With app in query parameter:
```bash
curl -X POST "http://localhost:9090/logs?app=myservice" \
  -H "Content-Type: application/json" \
  -d '{
    "timestamp": "2026-02-09T14:30:00Z",
    "level": "warn",
    "message": "High memory usage detected"
  }'
```

#### With additional fields:
```bash
curl -X POST http://localhost:9090/logs \
  -H "Content-Type: application/json" \
  -d '{
    "timestamp": "2026-02-09T14:30:00Z",
    "level": "error",
    "message": "Database connection failed",
    "app": "api-service",
    "user": "system",
    "fields": {
      "host": "db.example.com",
      "port": 5432,
      "retry_count": 3,
      "error_code": "TIMEOUT"
    }
  }'
```

## Log File Format

Log lines are written in the following format:

```
[timestamp] [app] [user] [level] message | field1=value1 field2=value2 ...
```

### Example Log Files

**File: logs/2026-02-09.log**
```
[2026-02-09T14:30:00Z] [myservice] [alice] [info] User login successful | user_id=12345 ip_address=203.0.113.42
[2026-02-09T14:31:15Z] [api-service] [system] [error] Database connection failed | error_code=TIMEOUT host=db.example.com port=5432 retry_count=3
[2026-02-09T14:32:45Z] [web-frontend] [info] Page rendered | page=dashboard render_time_ms=245
```

**File: logs/2026-02-08.log**
```
[2026-02-08T23:45:30Z] [myservice] [warn] Deprecated endpoint accessed | user_id=67890
```

## Timestamp-Based File Organization and 3-Day Window

Events are routed to log files based on the UTC date of their `timestamp` field:

- Message with timestamp `2026-02-09T14:30:00Z` → written to `logs/2026-02-09.log`
- Message with timestamp `2026-02-08T10:00:00Z` → written to `logs/2026-02-08.log` (if within validation window)
- Message with timestamp `2026-02-10T08:15:00Z` → written to `logs/2026-02-10.log` (if within validation window)

**Validation Window:**
The server enforces a 3-day validation window (current date ± 1 day) to prevent:
- Acceptance of events with stale timestamps (e.g., from misclocked clients)
- Unbounded growth of log file names in the directory

If a timestamp falls outside this window, the request is rejected with `400 Bad Request`.

**File Handle Optimization:**

The service maintains an always-open file handle for today's log file to maximize write performance. When the server date changes:
1. The current-day file is closed
2. A new file handle is opened for the new current date
3. The date transition is detected automatically on the next incoming request

For log events from yesterday or tomorrow (rare, given the validation window), files are:
1. Opened
2. Written to
3. Immediately closed (no caching)

This design balances performance (frequent writes to today's file) with simplicity (no complex multi-file caching).

## Testing

Run unit tests:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test -v ./...
```

## Project Structure

```
.
├── cmd/
│   └── logger-server/
│       └── main.go              # Server entrypoint
├── internal/
│   ├── model/
│   │   ├── event.go             # Event model and validation
│   │   └── event_test.go        # Model tests
│   ├── format/
│   │   ├── line.go              # Log line formatting
│   │   └── line_test.go         # Formatter tests
│   ├── sink/
│   │   ├── filesink.go          # Date-based file sink implementation
│   │   └── filesink_test.go     # File sink tests
│   └── httpapi/
│       ├── handlers.go          # HTTP request handlers
│       └── handlers_test.go     # Handler tests
├── go.mod
├── go.sum
└── README.md                     # This file
```

## Implementation Notes

### Validation

The model package (`internal/model/event.go`) handles:
- RFC3339 timestamp parsing
- Log level normalisation (case-insensitive)
- 3-day window validation based on UTC date comparison
- Required field validation (timestamp, level, message)

### Formatting

The format package (`internal/format/line.go`) produces consistent, one-event-per-line output with:
- Sanitisation of newlines (replaced with tabs)
- Lexicographic sorting of extra fields
- Optional app and user segments
- Proper JSON marshalling of complex field values

### File Management

The file sink (`internal/sink/filesink.go`) handles:
- Automatic directory creation
- Date-based file routing
- Current-day file handle persistence
- Date change detection and file rotation
- Mutex-protected concurrent writes
- Open-write-close pattern for adjacent-day events

### HTTP API

The handler (`internal/httpapi/handlers.go`) provides:
- RFC3339 timestamp validation
- JSON payload parsing
- Query parameter extraction for app
- Proper HTTP status codes and error messages


```bash
go run ./cmd/logger-server
```

This will start the server on `http://localhost:9090` and write logs to `./logs/app.log`.

### Example request

```bash
curl -X POST http://localhost:9090/logs \
  -H "Content-Type: application/json" \
  -d '{
    "timestamp": "2026-02-09T12:34:56Z",
    "level": "info",
    "message": "User logged in",
    "fields": {
      "user_id": "123",
      "ip": "203.0.113.42"
    }
  }'
```

Example log line:

```text
[2026-02-09T12:34:56Z] [info] User logged in | ip=203.0.113.42 user_id=123
```

