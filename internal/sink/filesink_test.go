package sink

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileSink_CreatesDirectoryAndFile(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs", "subdir")

	fs, err := NewFileSink(logDir)
	if err != nil {
		t.Fatalf("NewFileSink failed: %v", err)
	}
	defer fs.Close()

	// Verify directory was created
	if info, err := os.Stat(logDir); err != nil || !info.IsDir() {
		t.Fatalf("log directory was not created")
	}

	// Verify today's file was created
	todayFile := dateFilePath(logDir, todayDateString())
	if info, err := os.Stat(todayFile); err != nil || info.IsDir() {
		t.Fatalf("today's log file was not created")
	}
}

func TestFileSink_WritesToCurrentDayFile(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileSink(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSink failed: %v", err)
	}
	defer fs.Close()

	ctx := context.Background()
	now := time.Now().UTC()
	err = fs.WriteLine(ctx, "test line 1", now)
	if err != nil {
		t.Fatalf("WriteLine failed: %v", err)
	}

	err = fs.WriteLine(ctx, "test line 2", now)
	if err != nil {
		t.Fatalf("WriteLine failed: %v", err)
	}

	// Verify lines were written to today's file
	todayFile := dateFilePath(tmpDir, todayDateString())
	content, err := os.ReadFile(todayFile)
	if err != nil {
		t.Fatalf("failed to read today's log file: %v", err)
	}

	lines := string(content)
	if !contains(lines, "test line 1") || !contains(lines, "test line 2") {
		t.Fatalf("expected lines not found in file. Got: %s", lines)
	}
}

func TestFileSink_WritesToAdjacentDayFiles(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileSink(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSink failed: %v", err)
	}
	defer fs.Close()

	ctx := context.Background()
	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	// Write to today
	if err := fs.WriteLine(ctx, "today's log", now); err != nil {
		t.Fatalf("WriteLine for today failed: %v", err)
	}

	// Write to yesterday
	if err := fs.WriteLine(ctx, "yesterday's log", yesterday); err != nil {
		t.Fatalf("WriteLine for yesterday failed: %v", err)
	}

	// Write to tomorrow
	if err := fs.WriteLine(ctx, "tomorrow's log", tomorrow); err != nil {
		t.Fatalf("WriteLine for tomorrow failed: %v", err)
	}

	// Verify each file has the correct content
	todayFile := dateFilePath(tmpDir, extractDateString(now))
	yesterdayFile := dateFilePath(tmpDir, extractDateString(yesterday))
	tomorrowFile := dateFilePath(tmpDir, extractDateString(tomorrow))

	todayContent, _ := os.ReadFile(todayFile)
	if !contains(string(todayContent), "today's log") {
		t.Fatalf("today's file should contain 'today's log'")
	}

	yesterdayContent, _ := os.ReadFile(yesterdayFile)
	if !contains(string(yesterdayContent), "yesterday's log") {
		t.Fatalf("yesterday's file should contain 'yesterday's log'")
	}

	tomorrowContent, _ := os.ReadFile(tomorrowFile)
	if !contains(string(tomorrowContent), "tomorrow's log") {
		t.Fatalf("tomorrow's file should contain 'tomorrow's log'")
	}
}

func TestFileSink_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileSink(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSink failed: %v", err)
	}
	defer fs.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	now := time.Now().UTC()
	err = fs.WriteLine(ctx, "test line", now)
	if err == nil {
		t.Fatalf("expected error for cancelled context")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
