---
id: TASK-7
title: 'Index tools (structured read, upsert entry, refresh stats)'
status: Done
assignee: []
created_date: '2026-04-13 21:08'
updated_date: '2026-04-14 22:25'
labels: []
milestone: m-1
dependencies:
  - TASK-5
documentation:
  - doc-2
  - doc-3
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Implement `index_read`, `index_upsert_entry`, `index_refresh_stats` per doc-3. `index.md` has a strict shape (doc-2 § "index.md Structure") — emoji-prefixed section headers (🔬 Research, 🏷️ Entities, 💡 Concepts, 🏗️ Infrastructure), bullet entries `- [Title](path) — summary`, and a `## Stats` footer.

`index_read` returns `{sections: [{key, title, entries: [{title, path, summary}]}], stats: {...}}`.

`index_upsert_entry` takes `(section_key, title, path, summary)` — creates the entry if absent, updates the summary if title+path already present. Preserves section order and any free-form content between sections. If the target section key doesn't match an existing section but matches one in `[index].sections` config, insert the section in the configured order.

`index_refresh_stats` recomputes page count and last-updated date from disk and rewrites the `## Stats` block without touching the pages list.

Round-trip test: read `existing/wiki/index.md`, write it back unchanged. Must be byte-identical modulo trailing newline.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `index_read` parses `existing/wiki/index.md` into structured form without data loss
- [x] #2 Writing the parsed form back produces byte-identical output (round-trip test passes)
- [x] #3 `index_upsert_entry` updates summary when title+path match; otherwise appends within the correct section
- [x] #4 New sections are inserted per `[index].sections` config order
- [x] #5 `index_refresh_stats` rewrites only the stats block; pages list is untouched
- [x] #6 All mutations respect `read_only` config flag
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

1. **Create `internal/wiki/index.go`** — Core index parsing/rendering logic:
   - `IndexSection` / `IndexEntry` / `IndexStats` / `IndexDocument` types
   - `ParseIndex(data []byte) (*IndexDocument, error)` — parse index.md into structured form preserving all content
   - `RenderIndex(doc *IndexDocument) []byte` — render back to markdown (must round-trip byte-identical)
   - `IndexRead(cfg) (*IndexDocument, *ToolError)` — read and parse index.md
   - `IndexUpsertEntry(cfg, sectionKey, title, path, summary) *ToolError` — add/update entry
   - `IndexRefreshStats(cfg) *ToolError` — recompute stats from disk

2. **Create `internal/wiki/index_test.go`** — Tests:
   - Round-trip test with existing/wiki/index.md
   - Upsert existing entry (update summary)
   - Upsert new entry (append to section)
   - Upsert to new section (inserted per config order)
   - Refresh stats
   - Read-only guard tests

3. **Register tools in `tools.go`** — Add `index_read`, `index_upsert_entry`, `index_refresh_stats` tool definitions and handlers

4. **Verify** — `go build ./...` and `go test ./...`
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented `index_read`, `index_upsert_entry`, `index_refresh_stats` in `internal/wiki/index.go`.

Key fix: round-trip parser was storing the leading blank line after `## Stats` in `Extra`, causing a +1 byte on render. Fixed by skipping leading blanks before any stat key is parsed (since `RenderIndex` hardcodes that blank line).

All 8 index tests pass. Registered all three tools in `tools.go`. Full build and test suite green.
<!-- SECTION:FINAL_SUMMARY:END -->
