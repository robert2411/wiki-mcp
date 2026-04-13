---
id: TASK-17
title: File watcher for live-reflect-on-disk-change (fsnotify)
status: To Do
assignee: []
created_date: '2026-04-13 21:10'
updated_date: '2026-04-13 21:18'
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
- [ ] #1 Editing a page on disk is reflected on the next HTTP request without server restart
- [ ] #2 `ETag` and `Last-Modified` headers present on rendered-page responses
- [ ] #3 `AutoRebuild=false` disables the watcher (verified by not reflecting a disk edit)
- [ ] #4 Polling fallback engages and logs a WARN when fsnotify.NewWatcher returns error (simulated in test via fake FS)
- [ ] #5 Watcher smoke test passes on macOS (FSEvents) + Linux (inotify) + Windows (ReadDirectoryChangesW) in CI
<!-- AC:END -->
