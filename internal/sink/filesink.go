package sink

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Sink is an abstraction over a destination that can accept log lines.
type Sink interface {
	WriteLine(ctx context.Context, line string) error
}

// FileSink appends log lines to a single file, guarding writes with a mutex.
type FileSink struct {
	mu   sync.Mutex
	file *os.File
}

// NewFileSink creates or opens the file at path and returns a FileSink.
func NewFileSink(path string) (*FileSink, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	return &FileSink{file: f}, nil
}

// WriteLine writes a single line followed by a newline to the file.
func (s *FileSink) WriteLine(ctx context.Context, line string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file == nil {
		return fmt.Errorf("file sink is closed")
	}

	if _, err := s.file.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("write log line: %w", err)
	}
	return nil
}

// Close closes the underlying file.
func (s *FileSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file == nil {
		return nil
	}
	err := s.file.Close()
	s.file = nil
	return err
}

