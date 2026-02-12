# Changelog

All notable changes to this project are documented in this file.

## Unreleased

### Added
- Implement file-based logging ingest service (rest-log-ingest-to-file) — 2026-02-10
  - Message-timestamp-based routing to daily files (YYYY-MM-DD.log)
  - Validation: accept events whose UTC date is within ±1 day of server date (3-day window)
  - File sink: always-open handle for current day; open-write-close for adjacent days
  - Single-line bracketed log format with deterministic fields and newline-to-tab sanitisation
  - HTTP API: `POST /logs` accepting JSON (timestamp, level, message, optional user/app/fields); `app` may be provided via `?app=` (body takes precedence)
  - Mutex-guarded writes and file `Sync` for durability

### Changed
- **BREAKING**: Updated log output format — 2026-02-12
  - Field order changed from `[timestamp] [app] [user] [level] message` to `[timestamp] [level] [app] [user] message`
  - Log levels now uppercase with abbreviations: `[DEBUG]`, `[INFO] `, `[WARN] `, `[ERROR]` (padded to 7 characters total)
  - Previously levels were lowercase: `[debug]`, `[info]`, `[warn]`, `[error]`
  - This improves log readability by placing severity immediately after the timestamp

