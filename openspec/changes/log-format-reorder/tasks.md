## 1. Update Format Package

- [x] 1.1 Update `internal/format/line.go` to reorder fields (timestamp, level, app, user, message)
- [x] 1.2 Add level abbreviation mapping function (debug→DEBUG, info→INFO, warn→WARN, error→ERROR)
- [x] 1.3 Implement 7-character padding for level fields ([DEBUG], [INFO] , [WARN] , [ERROR])
- [x] 1.4 Verify the format output matches the new spec

## 2. Update Format Tests

- [x] 2.1 Update `internal/format/line_test.go` TestFormatEvent_NoFields to expect new format
- [x] 2.2 Update TestFormatEvent_WithApp to expect new format
- [x] 2.3 Update TestFormatEvent_WithUser to expect new format
- [x] 2.4 Update TestFormatEvent_WithFields to expect new format (if exists)
- [x] 2.5 Add test case for correct level padding and alignment
- [x] 2.6 Verify all format tests pass

## 3. Update Model & Tests (if needed)

- [x] 3.1 Check `internal/model/event_test.go` for any level formatting tests
- [x] 3.2 Update any tests that check level representation to expect uppercase format
- [x] 3.3 Verify all model tests pass

## 4. Update Specifications

- [x] 4.1 Update `specs/rest-log-ingest-to-file/spec.md` to reflect new log format (field order and level representation)
- [x] 4.2 Verify the spec matches the implementation and design

## 5. Update Documentation

- [x] 5.1 Update README.md if it contains log format examples
- [x] 5.2 Update CHANGELOG.md with the log format change
- [ ] 5.3 Update RELEASE_NOTES.md documenting the breaking change in log format

## 6. Final Verification

- [x] 6.1 Run all tests (`go test ./...`)
- [x] 6.2 Build the project (`go build ./cmd/logger-server`)
- [x] 6.3 Manual test: Start the server and POST a test log event to verify format
- [x] 6.4 Verify output in the generated log file matches expected format
