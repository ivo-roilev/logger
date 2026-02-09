package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeSink struct {
	lines      []string
	timestamps []time.Time
	err        error
}

func (f *fakeSink) WriteLine(ctx context.Context, line string, timestamp time.Time) error {
	if f.err != nil {
		return f.err
	}
	f.lines = append(f.lines, line)
	f.timestamps = append(f.timestamps, timestamp)
	return nil
}

func TestPostLog_Success(t *testing.T) {
	fs := &fakeSink{}
	h := NewLoggerHandler(fs)

	body := []byte(`{
		"timestamp": "2026-02-09T12:34:56Z",
		"level": "info",
		"message": "ok"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.PostLog(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rr.Code)
	}

	if len(fs.lines) != 1 {
		t.Fatalf("expected 1 line written, got %d", len(fs.lines))
	}
}

func TestPostLog_InvalidContentType(t *testing.T) {
	fs := &fakeSink{}
	h := NewLoggerHandler(fs)

	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewReader([]byte(`{}`)))
	// No Content-Type set.

	rr := httptest.NewRecorder()
	h.PostLog(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestPostLog_AppFromQuery(t *testing.T) {
	fs := &fakeSink{}
	h := NewLoggerHandler(fs)

	body := []byte(`{
		"timestamp": "2026-02-09T12:34:56Z",
		"level": "info",
		"message": "ok"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/logs?app=myapp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.PostLog(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rr.Code)
	}

	if len(fs.lines) != 1 {
		t.Fatalf("expected 1 line written, got %d", len(fs.lines))
	}

	// Check that app from query parameter is included in the line
	if !bytes.Contains([]byte(fs.lines[0]), []byte("[myapp]")) {
		t.Fatalf("expected app [myapp] in line, got: %s", fs.lines[0])
	}
}

func TestPostLog_AppJSONTakesPrecedence(t *testing.T) {
	fs := &fakeSink{}
	h := NewLoggerHandler(fs)

	body := []byte(`{
		"timestamp": "2026-02-09T12:34:56Z",
		"level": "info",
		"message": "ok",
		"app": "jsonapp"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/logs?app=queryapp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.PostLog(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rr.Code)
	}

	// Check that JSON app takes precedence over query parameter
	if !bytes.Contains([]byte(fs.lines[0]), []byte("[jsonapp]")) {
		t.Fatalf("expected app [jsonapp] in line, got: %s", fs.lines[0])
	}
	if bytes.Contains([]byte(fs.lines[0]), []byte("[queryapp]")) {
		t.Fatalf("should not have query app [queryapp] in line: %s", fs.lines[0])
	}
}

func TestPostLog_OutsideWindow(t *testing.T) {
	fs := &fakeSink{}
	h := NewLoggerHandler(fs)

	body := []byte(`{
		"timestamp": "2026-02-07T12:34:56Z",
		"level": "info",
		"message": "ok"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.PostLog(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d (outside window), got %d", http.StatusBadRequest, rr.Code)
	}

	if len(fs.lines) != 0 {
		t.Fatalf("expected no lines written, got %d", len(fs.lines))
	}
}

