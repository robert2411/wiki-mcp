package web

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// startWatcher watches wikiPath for file changes and calls onChanged after
// a 200ms debounce window. If fsnotify is unavailable it falls back to
// polling every pollInterval, logging a WARN once.
//
// The watcherFactory parameter lets tests inject a failing factory to exercise
// the polling fallback path.
func startWatcher(ctx context.Context, wikiPath string, logger *slog.Logger, onChanged func(), watcherFactory func() (*fsnotify.Watcher, error)) {
	w, err := watcherFactory()
	if err != nil {
		logger.Warn("fsnotify unavailable, falling back to polling watcher", "err", err)
		go pollWatcher(ctx, wikiPath, logger, onChanged)
		return
	}

	if err := addDirsRecursive(w, wikiPath); err != nil {
		logger.Warn("watcher: failed to add dirs, falling back to polling", "err", err)
		w.Close()
		go pollWatcher(ctx, wikiPath, logger, onChanged)
		return
	}

	go runFsnotifyWatcher(ctx, w, logger, onChanged)
}

// runFsnotifyWatcher reads fsnotify events, debounces them, and calls
// onChanged. New directories are added to the watcher automatically.
func runFsnotifyWatcher(ctx context.Context, w *fsnotify.Watcher, logger *slog.Logger, onChanged func()) {
	defer w.Close()

	const debounce = 200 * time.Millisecond
	var timer *time.Timer
	fire := func() {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounce, onChanged)
	}

	for {
		select {
		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return

		case event, ok := <-w.Events:
			if !ok {
				return
			}
			// On Create, check if it's a new directory and watch it recursively.
			// Reuse the stat info to avoid a second syscall below.
			if event.Has(fsnotify.Create) {
				if info, err := os.Stat(event.Name); err == nil {
					if info.IsDir() {
						if err := addDirsRecursive(w, event.Name); err != nil {
							logger.Warn("watcher: add new dir", "dir", event.Name, "err", err)
						}
						continue // directory events don't invalidate the page cache
					}
				}
			}
			// Fire on any .md file change (write, remove, rename, create).
			if strings.HasSuffix(event.Name, ".md") {
				fire()
			}

		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			logger.Warn("watcher error", "err", err)
		}
	}
}

// pollInterval is the delay between poll checks. Overridden in tests.
var pollInterval = 5 * time.Second

// pollWatcher is the fallback when fsnotify is unavailable. It checks the
// wiki directory every pollInterval and fires onChanged when any file mtime
// increases.
func pollWatcher(ctx context.Context, wikiPath string, logger *slog.Logger, onChanged func()) {
	lastMax := maxMtime(wikiPath)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if current := maxMtime(wikiPath); current.After(lastMax) {
				lastMax = current
				onChanged()
			}
		}
	}
}

// addDirsRecursive walks root and adds every directory to w.
func addDirsRecursive(w *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable dirs
		}
		if d.IsDir() {
			return w.Add(path)
		}
		return nil
	})
}

// maxMtime returns the latest modification time of any file under wikiPath.
func maxMtime(wikiPath string) time.Time {
	var max time.Time
	_ = filepath.WalkDir(wikiPath, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if t := info.ModTime(); t.After(max) {
			max = t
		}
		return nil
	})
	return max
}
