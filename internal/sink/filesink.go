package sink

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Sink is an abstraction over a destination that can accept log lines.
type Sink interface {
	WriteLine(ctx context.Context, line string, timestamp time.Time) error
}

// FileSink maintains date-based log files with current-day file handle optimization.
type FileSink struct {
	mu              sync.Mutex
	logDir          string
	currentDayStr   string   // Format: YYYY-MM-DD
	currentDayFile  *os.File // Always-open file handle for the current day
}

// NewFileSink creates a FileSink for the given log directory.
// It'll manage date-based log files (YYYY-MM-DD.log) and keep only today's file open.
func NewFileSink(logDir string) (*FileSink, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	fs := &FileSink{
		logDir: logDir,
	}

	// Initialise the current day file handle
	today := todayDateString()
	todayPath := dateFilePath(logDir, today)
	f, err := os.OpenFile(todayPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open today's log file: %w", err)
	}

	fs.currentDayStr = today
	fs.currentDayFile = f
	return fs, nil
}

// WriteLine writes a log line to the appropriate dated file based on the message's timestamp.
// For the current day, the always-open file handle is used.
// For adjacent days, the file is opened, written to, and immediately closed (no caching).
func (fs *FileSink) WriteLine(ctx context.Context, line string, timestamp time.Time) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.currentDayFile == nil {
		return fmt.Errorf("file sink is closed")
	}

	// Extract the UTC date from the timestamp
	targetDate := extractDateString(timestamp)

	// Check if server date has changed
	today := todayDateString()
	if today != fs.currentDayStr {
		// Close the old current-day file
		_ = fs.currentDayFile.Close()

		// Open a new file for the new current day
		newPath := dateFilePath(fs.logDir, today)
		newFile, err := os.OpenFile(newPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			fs.currentDayFile = nil
			return fmt.Errorf("open new current day file: %w", err)
		}

		fs.currentDayStr = today
		fs.currentDayFile = newFile
	}

	// Route to the appropriate file
	if targetDate == fs.currentDayStr {
		// Use the always-open current-day file handle
		if _, err := fs.currentDayFile.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("write to current day file: %w", err)
		}
		if err := fs.currentDayFile.Sync(); err != nil {
			return fmt.Errorf("sync current day file: %w", err)
		}
	} else {
		// Open-write-close for adjacent days (no caching, expected to be rare)
		targetPath := dateFilePath(fs.logDir, targetDate)
		f, err := os.OpenFile(targetPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("open dated file %s: %w", targetPath, err)
		}
		defer f.Close()

		if _, err := f.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("write to dated file %s: %w", targetPath, err)
		}
		if err := f.Sync(); err != nil {
			return fmt.Errorf("sync dated file %s: %w", targetPath, err)
		}
	}

	return nil
}

// Close closes the current-day file handle.
func (fs *FileSink) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.currentDayFile == nil {
		return nil
	}
	err := fs.currentDayFile.Close()
	fs.currentDayFile = nil
	return err
}

// todayDateString returns the current UTC date in YYYY-MM-DD format.
func todayDateString() string {
	return time.Now().UTC().Format("2006-01-02")
}

// extractDateString extracts the UTC date from a timestamp in YYYY-MM-DD format.
func extractDateString(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

// dateFilePath constructs the full path to a dated log file.
func dateFilePath(dir string, dateStr string) string {
	return filepath.Join(dir, dateStr+".log")
}


