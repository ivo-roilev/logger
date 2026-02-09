package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeSink struct {
	lines []string
	err   error
}

func (f *fakeSink) WriteLine(ctx context.Context, line string) error {
	if f.err != nil {
		return f.err
	}
	f.lines = append(f.lines, line)
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

