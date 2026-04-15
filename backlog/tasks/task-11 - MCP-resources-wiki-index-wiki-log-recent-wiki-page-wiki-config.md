---
id: TASK-11
title: 'MCP resources (wiki://index, wiki://log/recent, wiki://page/*, wiki://config)'
status: Done
assignee: []
created_date: '2026-04-13 21:09'
updated_date: '2026-04-15 16:57'
labels: []
milestone: m-1
dependencies:
  - TASK-5
  - TASK-7
  - TASK-8
documentation:
  - doc-3
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Expose the four MCP resources from doc-3. Clients that prefer URI-based reads (some MCP hosts list resources in the sidebar) get a clean entry surface.

- `wiki://index` — rendered `index.md` contents
- `wiki://log/recent` — last 20 log entries concatenated
- `wiki://page/<relative-path>` — page body (with frontmatter retained)
- `wiki://config` — active config, `[safety]` section redacted if it contains anything sensitive (currently doesn't, but guard for future additions); read-only

`list_resources` advertises the four URIs plus one `wiki://page/<path>` per page (lazy-generated from disk scan; cap at ~500 entries for large wikis).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 All four resource URIs resolve with correct MIME types (text/markdown for pages, application/json for config)
- [x] #2 `list_resources` includes the page-per-URI entries up to the cap
- [x] #3 Reads enforce `confine_to_wiki_path`
- [x] #4 Integration test: an MCP test client reads all four resource shapes against `existing/wiki/`
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

1. Create `internal/wiki/resources.go`:
   - `RegisterResources(srv *server.Server)` — called from main.go
   - Register static resources: `wiki://index` (text/markdown), `wiki://log/recent` (text/markdown), `wiki://config` (application/json)
   - Register template: `wiki://page/{+path}` (text/markdown) — handles reads for any path
   - Scan disk at startup, register individual `wiki://page/<path>` resources up to 500 (for list_resources enumeration)
   - Handlers: index reads index.md raw; log/recent calls LogTail(20) and formats as markdown; config serializes cfg with safety section replaced by "[redacted]"; page strips URI prefix, calls cfg.ResolveWikiPath (enforces confine_to_wiki_path), reads raw file

2. Update `cmd/wiki-mcp/main.go` — add `wiki.RegisterResources(mcpSrv)` call

3. Create `internal/wiki/resources_test.go`:
   - Calls each handler directly against existing/wiki/ fixture
   - Verifies MIME type, content shape, path-escape rejection
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Implemented four MCP resources in `internal/wiki/resources.go` with `RegisterResources(srv)` called from `main.go`.

### Resources
- `wiki://index` (text/markdown) — raw `index.md`; falls back to placeholder if missing
- `wiki://log/recent` (text/markdown) — last 20 log entries formatted as markdown via `LogTail`
- `wiki://config` (application/json) — config JSON with `safety` section replaced by `"[redacted]"`
- `wiki://page/{+path}` (text/markdown) — raw page read (frontmatter retained); enforces `confine_to_wiki_path` via `cfg.ResolveWikiPath`

`list_resources` enumerates individual `wiki://page/<path>` URIs by scanning disk at startup (capped at 500). The `{+path}` template serves reads for any page including those beyond the cap.

### Tests
9 tests in `internal/wiki/resources_test.go` covering all four resource shapes against `existing/wiki/`, MIME types, path-escape rejection, and config redaction.

### Quality improvements (from simplify pass)
- Shared `textResource` helper eliminates repeated 3-field struct literal
- `mimeMarkdown`, `mimeJSON`, `maxPageResources` named constants
- Config JSON marshalled once at handler construction (not per-request)
- `index.md` path resolved once at handler construction
- `PageList` errors surfaced via `slog.Warn` instead of silently discarded
- Single `handleResourcePage` closure shared between template and static registrations
<!-- SECTION:FINAL_SUMMARY:END -->
