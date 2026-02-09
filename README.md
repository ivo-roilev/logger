## REST logging service

This project implements a small HTTP logging service in Go. It accepts JSON log events over a REST API and appends them to a local log file using a deterministic, human-readable format.

### API

- **Endpoint**: `POST /logs`
- **Content-Type**: `application/json`
- **Request body**:

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

### Log line format

Each accepted event is written as exactly one line:

- Base format:

```text
[timestamp] [level] message
```

- If `fields` is present and non-empty:

```text
[timestamp] [level] message | key1=value1 key2=value2
```

Newlines in `message` and string field values are replaced with tabs to keep one-event-per-line.

### Configuration

Environment variables:

- `PORT` — HTTP listen port (default: `8080`).
- `LOG_FILE_PATH` — path to the log file (default: `./logs/app.log`).

### Running locally

Ensure you have Go 1.22 or newer installed.

```bash
go run ./cmd/logger-server
```

This will start the server on `http://localhost:8080` and write logs to `./logs/app.log`.

### Example request

```bash
curl -X POST http://localhost:8080/logs \
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

