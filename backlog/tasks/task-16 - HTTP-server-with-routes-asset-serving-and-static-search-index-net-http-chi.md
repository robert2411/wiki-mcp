---
id: TASK-16
title: >-
  HTTP server with routes, asset serving, and static search index (net/http +
  chi)
status: To Do
assignee: []
created_date: '2026-04-13 21:10'
updated_date: '2026-04-13 21:18'
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
- [ ] #1 `wiki-mcp --serve` with `Web.Enabled=true` serves the wiki on the configured port
- [ ] #2 All 21 pages in `existing/wiki/` render at their expected URLs (integration test walks index links)
- [ ] #3 Path traversal attempts (`/_assets/../../etc/passwd`) return 404, not file content
- [ ] #4 `/_search?q=qwen` returns a page listing matches
- [ ] #5 `/_search_index.json` is well-formed JSON under 1 MiB for the existing wiki
- [ ] #6 Graceful shutdown: server drains in-flight requests on context cancel
<!-- AC:END -->
