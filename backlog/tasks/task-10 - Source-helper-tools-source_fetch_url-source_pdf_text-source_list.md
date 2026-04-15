---
id: TASK-10
title: 'Source helper tools (source_fetch_url, source_pdf_text, source_list)'
status: Done
assignee:
  - Robert Stevens
created_date: '2026-04-13 21:09'
updated_date: '2026-04-15 19:28'
labels: []
milestone: m-1
dependencies:
  - TASK-5
documentation:
  - doc-3
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Thin tool wrappers for source ingestion, per doc-3 § Sources. Let non-Claude MCP clients (no native web fetch) drive the Ingest workflow.

`source_fetch_url(url, slug?)`: fetch via stdlib `net/http`. Save under `<sources_path>/<slug>.md` (default slug from URL host+path). Returns saved path. HTML→markdown via `github.com/JohannesKaufmann/html-to-markdown` (preserves structure); raw text fallback if HTML parse fails.

`source_pdf_text(path)`: extract via `github.com/ledongthuc/pdf`. If system `pdftotext` (poppler) is on PATH and env `WIKI_MCP_PREFER_PDFTOTEXT=1`, shell out for higher quality. Return `{text, page_count}`.

`source_list()`: list files in `sources_path` with size + mtime.

All honor `sources_path` (default `<wiki_path>/../sources`). Create dir on first write if missing. Reject any path outside `sources_path`.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `source_fetch_url` saves a URL to `sources_path` with a reasonable default slug
- [x] #2 Fetcher needs no external `curl` — pure Go `net/http`
- [x] #3 `source_pdf_text` extracts text from a fixture PDF without requiring `pdftotext` on the system
- [x] #4 `source_list` reports size + mtime; empty list when dir absent
- [x] #5 All tools reject paths outside configured roots
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

### Approach
Implement three MCP tool wrappers in `internal/sources/` following the same patterns as `internal/wiki/`.

### Files
1. `internal/sources/sources.go` — core logic: path resolution, FetchURL, PDFText, List
2. `internal/sources/tools.go` — MCP tool definitions, handlers, RegisterTools func
3. `internal/sources/sources_test.go` — unit tests covering all ACs
4. `cmd/wiki-mcp/main.go` — add `sources.RegisterTools(mcpSrv)` import+call
5. `go.mod` / `go.sum` — add `github.com/JohannesKaufmann/html-to-markdown/v2` and `github.com/ledongthuc/pdf`

### Key Design
- `resolvePath(sourcesPath, rel string) (string, error)` — cleans, joins, checks prefix (mirrors config.ResolveWikiPath but for sources_path)
- `source_fetch_url(url, slug?)`: net/http GET → html-to-markdown conversion (raw text fallback on HTML parse fail) → write `<sources_path>/<slug>.md` (mkdir on first write); default slug = sanitized URL host+path
- `source_pdf_text(path)`: path resolved within sources_path; pure Go ledongthuc/pdf reader; if env `WIKI_MCP_PREFER_PDFTOTEXT=1` AND `pdftotext` on PATH, shell out instead; return `{text, page_count}`
- `source_list()`: os.ReadDir sources_path → list {name, size, mtime}; return empty slice (not error) if dir absent
- All path operations reject paths that escape sources_path (AC #5)
- Error codes: reuse wiki ToolError pattern with `sources.` prefix codes

### Tests
- FetchURL: httptest.Server serving HTML → verify file saved, slug derived correctly
- FetchURL path escape: reject slug like `../../etc/passwd`
- PDFText: write a minimal valid PDF fixture → verify text extracted
- List: empty when dir absent; populated after fetch

### Registration
`sources.RegisterTools(srv *server.Server)` called from main.go after `wiki.RegisterTools`.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented `source_fetch_url`, `source_pdf_text`, and `source_list` in `internal/sources/sources.go` + `tools.go`. Pure-Go net/http fetch with html-to-markdown conversion, pure-Go PDF extraction via ledongthuc/pdf with optional pdftotext fallback, and directory listing with size+mtime. All path operations enforce sources_path confinement. Full test suite in `sources_test.go` with httptest server, PDF fixture in `testdata/sample.pdf`, and path-escape rejection tests. Registered via `sources.RegisterTools` in `cmd/wiki-mcp/main.go`.
<!-- SECTION:FINAL_SUMMARY:END -->
