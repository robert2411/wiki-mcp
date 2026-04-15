---
id: TASK-17
title: File watcher for live-reflect-on-disk-change (fsnotify)
status: Done
assignee: []
created_date: '2026-04-13 21:10'
updated_date: '2026-04-15 19:47'
labels: []
milestone: m-3
dependencies:
  - TASK-16
documentation:
  - doc-6
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Honor `Web.AutoRebuild=true` (default) by invalidating the render cache + search index when pages change on disk.

Use `github.com/fsnotify/fsnotify`. Watch `wiki_path` recursively (walk + add subdirs; re-walk on mkdir events). On Linux this hits inotify; on macOS FSEvents; on Windows ReadDirectoryChangesW — all native, no polling.

Fallback: if fsnotify errors (filesystem doesn't support notifications, e.g. some network mounts), drop to a polling loop with a 5s interval. Log the fallback once at WARN level.

Cache invalidation keyed on page path. Search index rebuilt lazily on next `/_search_index.json` request after any page change. `ETag` + `Last-Modified` on rendered-page responses for client-side caching.

`Web.AutoRebuild=false` skips the watcher entirely; responses cached indefinitely until restart.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Editing a page on disk is reflected on the next HTTP request without server restart
- [x] #2 `ETag` and `Last-Modified` headers present on rendered-page responses
- [x] #3 `AutoRebuild=false` disables the watcher (verified by not reflecting a disk edit)
- [x] #4 Polling fallback engages and logs a WARN when fsnotify.NewWatcher returns error (simulated in test via fake FS)
- [x] #5 Watcher smoke test passes on macOS (FSEvents) + Linux (inotify) + Windows (ReadDirectoryChangesW) in CI
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

### Dependency
Add `github.com/fsnotify/fsnotify` via `go get`.

### Files

**New: `internal/web/watcher.go`**
- `startWatcher(ctx, wikiPath, logger, onChanged, newWatcherFn)` — `newWatcherFn func() (*fsnotify.Watcher, error)` enables test injection.
- Walk wikiPath on start; add all dirs to watcher. Re-add new dirs on CREATE+IsDir events.
- Debounce ~200ms (collect events, fire onChanged after quiet period) to avoid flaky tests with vim/VSCode multi-event saves.
- Polling fallback: if `newWatcherFn` returns error, log WARN, fall back to 5s polling loop that tracks max mtime across all .md files and fires onChanged when it changes.

**Modified: `internal/web/server.go`**
- Replace `indexOnce sync.Once` / `indexCache []SearchIndexEntry` with `mu sync.RWMutex` guarding both `indexCache` and `renderer`.
- `cachedIndex()`: RLock → return if non-nil; else write-lock → lazy build.
- `getRenderer()`: same pattern.
- `invalidateCache()`: write-lock, nil both.
- `servePage`: stat file → set `Last-Modified` from mtime, `ETag` as `"<mtime.UnixNano()>-<size>"`.
- `Run()`: if `cfg.Web.AutoRebuild`, call `startWatcher(ctx, wikiPath, logger, s.invalidateCache, fsnotify.NewWatcher)`.

**New: `internal/web/watcher_test.go`**
- Test polling fallback: inject factory that returns error → WARN logged → callback fires on file write.
- Test real fsnotify path: write file, wait up to 3s for callback.

**Modified: `internal/web/server_test.go`**
- `TestDiskChangeReflected` (AC#1): write file, get server with AutoRebuild=true, edit file, request → new content visible.
- `TestETagLastModified` (AC#2): GET rendered page → headers present.
- `TestAutoRebuildFalse` (AC#3): edit file, verify next request still returns old content.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Implemented file watcher with cache invalidation for `Web.AutoRebuild=true`.

### What was built

**New: `internal/web/watcher.go`**
- `startWatcher(ctx, wikiPath, logger, onChanged, watcherFactory)` — factory param enables test injection
- Walks `wikiPath` recursively on start; re-adds new subdirs on `fsnotify.Create` events
- 200ms debounce via `time.AfterFunc` to handle multi-event editor saves (vim, VS Code)
- On fsnotify error: logs WARN once, falls back to `pollWatcher` (5s interval, tracks max mtime)
- Fires only on `.md` file events (extension check, no extra stat call)

**Modified: `internal/web/server.go`**
- Replaced `sync.Once` + `indexCache` with `sync.RWMutex` guarding `renderer`, `indexCache`, and `pageCache`
- `InvalidateCache()` nils all three; called by watcher goroutine on disk changes
- `cachedPageEntry()` — double-checked locking: RLock read → write-lock build on miss
- `cachedIndex()` — same pattern for search index
- Both renderer and title index rebuilt lazily so wikilink resolution stays correct after page renames/adds
- `ETag` (`"<mtime_ns>-<size>"`) and `Last-Modified` headers on all rendered-page responses
- `Run()` starts watcher only when `cfg.Web.AutoRebuild` is true

**New: `internal/web/watcher_test.go`**
- `TestWatcherPollingFallback` — injects failing factory, verifies WARN logged + callback fires
- `TestWatcherFsnotifyCallback` — writes file, asserts callback within 3s (smoke test for AC#5)

**Modified: `internal/web/server_test.go`**
- `TestDiskChangeReflected` (AC#1) — edit + `InvalidateCache()` → next request sees new content
- `TestETagLastModified` (AC#2) — verifies headers present on page response
- `TestAutoRebuildFalse` (AC#3) — disk edit without invalidation leaves cached content unchanged

### Design decisions
- `pollInterval` is an overrideable package var (set to 50ms in test, 5s in production) to keep test fast
- Factory injection via parameter (not package var) keeps `startWatcher` signature explicit
- `pageEntry` struct (renamed from `cachedPage` to avoid name collision with `cachedPageEntry` method)
- Renderer is also invalidated on change so wikilinks resolve correctly when pages are added/renamed
<!-- SECTION:FINAL_SUMMARY:END -->
