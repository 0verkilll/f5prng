package f5prng

import (
	"testing"

	"github.com/0verkilll/logger"
	logtesting "github.com/0verkilll/logger/testing"
	"github.com/0verkilll/sha1"
)

func TestSetLogger(t *testing.T) {
	// Save original and restore after test
	original := GetLogger()
	defer SetLogger(original)

	mock := logtesting.NewMockLogger()
	SetLogger(mock)

	got := GetLogger()
	if got != mock {
		t.Error("GetLogger did not return the logger set by SetLogger")
	}
}

func TestSetLoggerNil(t *testing.T) {
	// Save original and restore after test
	original := GetLogger()
	defer SetLogger(original)

	// First set a non-nil logger
	mock := logtesting.NewMockLogger()
	SetLogger(mock)

	// Then set nil
	SetLogger(nil)

	got := GetLogger()
	if _, ok := got.(logger.NopLogger); !ok {
		t.Errorf("SetLogger(nil) should reset to NopLogger, got %T", got)
	}
}

func TestDefaultLoggerIsNop(t *testing.T) {
	// Create fresh package state by setting nil
	SetLogger(nil)

	got := GetLogger()
	if _, ok := got.(logger.NopLogger); !ok {
		t.Errorf("default logger should be NopLogger, got %T", got)
	}
}

func TestSeedLogsDebugMessages(t *testing.T) {
	// Save original and restore after test
	original := GetLogger()
	defer SetLogger(original)

	mock := logtesting.NewMockLogger()
	SetLogger(mock)

	hasher := sha1.NewSHA1(sha1.NewBigEndian())
	sr := NewSecureRandom(hasher)
	defer sr.Clear()

	err := sr.Seed([]byte("test seed"))
	if err != nil {
		t.Fatalf("Seed failed: %v", err)
	}

	// Should have logged debug messages
	entries := mock.Entries()
	if len(entries) < 2 {
		t.Errorf("expected at least 2 log entries (start and end), got %d", len(entries))
	}

	// First entry should mention Seed called
	logtesting.AssertLogContains(t, mock, logger.LevelDebug, "Seed called")
	logtesting.AssertLogContains(t, mock, logger.LevelDebug, "Seed completed")
}

func TestNextBytesNoLogging(t *testing.T) {
	// Save original and restore after test
	original := GetLogger()
	defer SetLogger(original)

	mock := logtesting.NewMockLogger()
	SetLogger(mock)

	hasher := sha1.NewSHA1(sha1.NewBigEndian())
	sr := NewSecureRandom(hasher)
	defer sr.Clear()

	if err := sr.Seed([]byte("test")); err != nil {
		t.Fatalf("Seed failed: %v", err)
	}

	// Clear mock to only capture NextBytes logs
	mock.Reset()

	_ = sr.NextBytes(20)

	// Hot-path methods do NOT log to avoid variadic arg allocation overhead.
	// Verify no debug messages were logged for NextBytes.
	if mock.Len() > 0 {
		t.Errorf("expected no log messages from NextBytes (hot-path optimization), got %d", mock.Len())
	}
}

func TestNextIntNoLogging(t *testing.T) {
	// Save original and restore after test
	original := GetLogger()
	defer SetLogger(original)

	mock := logtesting.NewMockLogger()
	SetLogger(mock)

	hasher := sha1.NewSHA1(sha1.NewBigEndian())
	sr := NewSecureRandom(hasher)
	defer sr.Clear()

	if err := sr.Seed([]byte("test")); err != nil {
		t.Fatalf("Seed failed: %v", err)
	}

	// Clear mock to only capture NextInt logs
	mock.Reset()

	_ = sr.NextInt()

	// Hot-path methods do NOT log to avoid variadic arg allocation overhead.
	// Verify no debug messages were logged for NextInt.
	if mock.Len() > 0 {
		t.Errorf("expected no log messages from NextInt (hot-path optimization), got %d", mock.Len())
	}
}

func TestClearLogsDebugMessages(t *testing.T) {
	// Save original and restore after test
	original := GetLogger()
	defer SetLogger(original)

	mock := logtesting.NewMockLogger()
	SetLogger(mock)

	hasher := sha1.NewSHA1(sha1.NewBigEndian())
	sr := NewSecureRandom(hasher)

	if err := sr.Seed([]byte("test")); err != nil {
		t.Fatalf("Seed failed: %v", err)
	}

	// Clear mock to only capture Clear logs
	mock.Reset()

	sr.Clear()

	// Should have logged debug messages
	logtesting.AssertLogContains(t, mock, logger.LevelDebug, "Clear called")
	logtesting.AssertLogContains(t, mock, logger.LevelDebug, "Clear completed")
}

func TestNopLoggerDoesNotPanic(t *testing.T) {
	// Ensure default NopLogger doesn't cause issues
	SetLogger(nil)

	hasher := sha1.NewSHA1(sha1.NewBigEndian())
	sr := NewSecureRandom(hasher)
	defer sr.Clear()

	// Should not panic with NopLogger
	err := sr.Seed([]byte("test"))
	if err != nil {
		t.Fatalf("Seed failed: %v", err)
	}

	_ = sr.NextBytes(20)
	_ = sr.NextInt()
	sr.Clear()
}

func TestSilentByDefault(t *testing.T) {
	// Reset to default state
	SetLogger(nil)

	// Get the logger and verify it's a NopLogger
	lg := GetLogger()
	if _, ok := lg.(logger.NopLogger); !ok {
		t.Errorf("expected NopLogger by default, got %T", lg)
	}

	// Verify operations work silently (no output, no panic)
	hasher := sha1.NewSHA1(sha1.NewBigEndian())
	sr := NewSecureRandom(hasher)
	defer sr.Clear()

	if err := sr.Seed([]byte("silent test")); err != nil {
		t.Fatalf("Seed failed: %v", err)
	}

	_ = sr.NextBytes(20)
	_ = sr.NextInt()
	sr.Clear()
}

func TestSeedWithEmptySeedLogs(t *testing.T) {
	// Save original and restore after test
	original := GetLogger()
	defer SetLogger(original)

	mock := logtesting.NewMockLogger()
	SetLogger(mock)

	hasher := sha1.NewSHA1(sha1.NewBigEndian())
	sr := NewSecureRandom(hasher)
	defer sr.Clear()

	err := sr.Seed([]byte{})
	if err != nil {
		t.Fatalf("Seed with empty seed failed: %v", err)
	}

	// Should log Seed called and Seed completed
	logtesting.AssertLogContains(t, mock, logger.LevelDebug, "Seed called")
	logtesting.AssertLogContains(t, mock, logger.LevelDebug, "Seed completed")
}

func TestNextBytesZeroLengthNoLogging(t *testing.T) {
	// Save original and restore after test
	original := GetLogger()
	defer SetLogger(original)

	mock := logtesting.NewMockLogger()
	SetLogger(mock)

	hasher := sha1.NewSHA1(sha1.NewBigEndian())
	sr := NewSecureRandom(hasher)
	defer sr.Clear()

	if err := sr.Seed([]byte("test")); err != nil {
		t.Fatalf("Seed failed: %v", err)
	}

	// Clear mock to only capture NextBytes logs
	mock.Reset()

	_ = sr.NextBytes(0)

	// Hot-path methods do NOT log to avoid variadic arg allocation overhead.
	// Verify no debug messages were logged for NextBytes(0).
	if mock.Len() > 0 {
		t.Errorf("expected no log messages from NextBytes(0) (hot-path optimization), got %d", mock.Len())
	}
}

func TestLoggerThreadSafety(t *testing.T) {
	// Save original and restore after test
	original := GetLogger()
	defer SetLogger(original)

	// Test concurrent access to logger
	done := make(chan bool)

	// Writer goroutines
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				mock := logtesting.NewMockLogger()
				SetLogger(mock)
			}
			done <- true
		}()
	}

	// Reader goroutines
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = GetLogger()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}
