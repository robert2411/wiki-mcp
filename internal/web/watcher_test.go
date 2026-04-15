package web

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

// TestWatcherPollingFallback verifies AC#4: when the watcher factory returns
// an error, the server falls back to polling and logs a WARN.
func TestWatcherPollingFallback(t *testing.T) {
	wikiPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(wikiPath, "index.md"), []byte("# Index\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	var called atomic.Bool
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	failFactory := func() (*fsnotify.Watcher, error) {
		return nil, errors.New("fsnotify not supported on this filesystem")
	}

	// Use a short poll interval so the test doesn't take 5s.
	orig := pollInterval
	pollInterval = 50 * time.Millisecond
	t.Cleanup(func() { pollInterval = orig })

	startWatcher(ctx, wikiPath, logger, func() { called.Store(true) }, failFactory)

	// WARN should have been logged synchronously before the goroutine starts.
	if !bytes.Contains(logBuf.Bytes(), []byte("polling")) {
		t.Errorf("expected WARN about polling fallback in log, got: %s", logBuf.String())
	}

	// Give the goroutine a moment to start, then write a file.
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(filepath.Join(wikiPath, "index.md"), []byte("# Changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if called.Load() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Error("polling fallback: onChanged was not called within 2s after file write")
}

// TestWatcherFsnotifyCallback verifies AC#5: the real fsnotify watcher fires
// the callback when a file is written.
func TestWatcherFsnotifyCallback(t *testing.T) {
	wikiPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(wikiPath, "index.md"), []byte("# Index\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	var called atomic.Bool
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startWatcher(ctx, wikiPath, logger, func() { called.Store(true) }, fsnotify.NewWatcher)

	// Give watcher time to initialise.
	time.Sleep(50 * time.Millisecond)

	// Write to a file — should trigger the callback within 2s (debounce + event).
	if err := os.WriteFile(filepath.Join(wikiPath, "index.md"), []byte("# Updated\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if called.Load() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Error("fsnotify watcher: onChanged was not called within 3s after file write")
}
