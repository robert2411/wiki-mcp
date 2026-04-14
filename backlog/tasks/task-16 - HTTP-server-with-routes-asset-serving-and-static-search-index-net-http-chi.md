---
id: TASK-16
title: >-
  HTTP server with routes, asset serving, and static search index (net/http +
  chi)
status: Done
assignee: []
created_date: '2026-04-13 21:10'
updated_date: '2026-04-14 22:49'
labels: []
milestone: m-3
dependencies:
  - TASK-15
  - TASK-9
documentation:
  - doc-6
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Wrap the renderer from task-15 in an HTTP server. Use stdlib `net/http` + `github.com/go-chi/chi/v5` router. Runs in a goroutine alongside stdio MCP via `errgroup`.

Routes per doc-6:
- `GET /` → render `index.md`
- `GET /{path}` → render `<path>.md`
- `GET /_log` → render `log.md`
- `GET /_search?q=...` → server-side search page (uses `wiki_search` from task-9)
- `GET /_assets/*` → static files from inside `wiki_path` only (confine!)
- `GET /_search_index.json` → pre-built client-side search index (titles + first 500 chars of body per page)

Bind to `Web.Bind` + `Web.Port`. 404 on paths that would escape `wiki_path` (use `filepath.Rel` + prefix check). Correct Content-Type headers.

Theme assets (HTML templates, CSS, JS) loaded via `//go:embed` from task-2's `web/theme/default/`. Templates rendered with stdlib `html/template`.

No file watcher in this task — that's task-17.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `wiki-mcp --serve` with `Web.Enabled=true` serves the wiki on the configured port
- [x] #2 All 21 pages in `existing/wiki/` render at their expected URLs (integration test walks index links)
- [x] #3 Path traversal attempts (`/_assets/../../etc/passwd`) return 404, not file content
- [x] #4 `/_search?q=qwen` returns a page listing matches
- [x] #5 `/_search_index.json` is well-formed JSON under 1 MiB for the existing wiki
- [x] #6 Graceful shutdown: server drains in-flight requests on context cancel
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

### Key constraint
`//go:embed` paths cannot use `..`. Create `web/embed.go` (package `webtheme`) to embed `web/theme/default/` files, then import from `internal/web/server.go`.

### New files
1. `web/embed.go` — embed.FS for theme assets
2. `web/theme/default/page.html` — Go html/template for rendered pages
3. `web/theme/default/search.html` — search results template  
4. `web/theme/default/style.css` — minimal CSS
5. `internal/web/server.go` — chi router, handlers, graceful shutdown
6. `internal/web/search.go` — BuildSearchIndex + Search (inline, no TASK-9 dependency)
7. `internal/web/server_test.go` — httptest integration tests

### Modified files
1. `cmd/wiki-mcp/main.go` — add `--serve` flag, errgroup for MCP + HTTP
2. `internal/web/render/render.go:303` — prepend `/` to wikilink hrefs
3. `web/theme/default/index.html` — replaced by page.html (can be removed)

### Routes
- `GET /` → render `index.md`
- `GET /_log` → render `log.md`
- `GET /_search?q=...` → inline keyword search (case-insensitive)
- `GET /_search_index.json` → JSON array of {path, title, snippet}
- `GET /_assets/{*}` → static files confined to wiki_path; reject .md; 404 on escape
- `GET /{*path}` → render `{path}.md`

### Dependencies to add
- `github.com/go-chi/chi/v5`
- `golang.org/x/sync` (errgroup)
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Implemented the read-only HTTP wiki UI alongside the existing stdio MCP server.

### What was built

**New files:**
- `web/embed.go` — `//go:embed` package for theme assets (Go embed can't use `..` paths, so this lives at `web/` to embed `theme/default/`)
- `web/theme/default/page.html` — Go `html/template` for rendered pages with sidebar nav + search bar
- `web/theme/default/search.html` — search results template
- `web/theme/default/style.css` — minimal CSS with dark-mode support
- `internal/web/server.go` — chi router, all route handlers, graceful shutdown, `sync.Once` index cache
- `internal/web/search.go` — `BuildSearchIndex` (builds once, cached) + `Search` (in-memory filter, no re-walk)
- `internal/web/server_test.go` — httptest integration tests covering: index render, nested-path render, 404, path traversal, search, search index JSON, CSS serving

**Modified files:**
- `cmd/wiki-mcp/main.go` — `--serve` flag sets `Web.Enabled=true`; `errgroup` runs MCP stdio + HTTP concurrently
- `internal/web/render/render.go` — fixed wikilink hrefs to use leading `/` (previously `entities/foo`, now `/entities/foo`)
- `internal/wiki/wiki.go` — exported `TitleFromPath` (was `titleFromPath`)

### Routes
| Route | Behaviour |
|---|---|
| `GET /` | Renders `index.md` |
| `GET /_log` | Renders `log.md` |
| `GET /_search?q=...` | In-memory substring search over cached index |
| `GET /_search_index.json` | Cached JSON index (path + title + first 500 chars) |
| `GET /_theme/*` | Embedded theme assets (CSS etc.) |
| `GET /_assets/*` | Wiki-dir static files; confined; `.md` rejected |
| `GET /*` | Renders `{path}.md` |

### Key design decisions
- Index built once with `sync.Once` and shared for nav, search, and `/_search_index.json`. task-17 (file watcher) will add cache invalidation.
- Path traversal prevented by `config.ResolveWikiPath` + unconditional wikiRoot prefix check in `handleAsset` (belt-and-suspenders: `ResolveWikiPath` only enforces when `ConfineToWikiPath` flag is set).
- TASK-9 (search MCP tools) not yet done, so search is implemented inline in `internal/web/search.go`. Adequate for AC#4.
<!-- SECTION:FINAL_SUMMARY:END -->
