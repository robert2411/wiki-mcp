---
id: TASK-6
title: 'Page CRUD tools (read, write, delete, move, list)'
status: Done
assignee: []
created_date: '2026-04-13 21:08'
updated_date: '2026-04-14 19:16'
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
Implement the five page tools from doc-3 in Go: `page_read`, `page_write`, `page_delete`, `page_move`, `page_list`. All paths relative to `wiki_path`. Enforce `ConfineToWikiPath` and `ReadOnly` via the helpers from task-3.

Frontmatter: use `github.com/adrg/frontmatter` (or hand-split with `gopkg.in/yaml.v3`). `page_read` returns `{path, frontmatter: map[string]any, body: string}`. `page_write` accepts both; writes YAML frontmatter only when non-empty — never emit an empty `---\n---\n` block.

`page_move` rewrites incoming links (`[[Old Title]]` and `[text](old/path.md)` across the wiki). Use the shared link-graph package from task-9 — if that task hasn't landed, scaffold a minimal link scanner here and refactor later.

`page_list` supports filters: directory prefix, glob (stdlib `path/filepath.Match` or `doublestar` for `**`), `tag` from frontmatter, `updated_since` date.

All tools return MCP-shaped errors with a stable error-code field (e.g. `wiki.NotFound`, `wiki.ReadOnly`, `wiki.PathEscape`).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `page_read` returns frontmatter + body; missing file returns a structured not-found error (not a panic)
- [x] #2 `page_write` creates parent dirs as needed; refuses paths outside `wiki_path`; refuses when `ReadOnly=true`
- [x] #3 `page_write` does not emit an empty frontmatter block when no metadata is provided
- [x] #4 `page_delete` rejects `index.md` and `log.md` with a clear reason
- [x] #5 `page_move` rewrites `[[Title]]` and `[text](rel/path)` across the wiki and fixes the moved page's own outgoing links
- [x] #6 `page_list` supports dir/glob/tag/updated_since filters (table-driven test against `existing/wiki/` as fixture)
- [x] #7 All tools return structured errors with stable error codes
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

1. **page_read** — Parse frontmatter (hand-split `---`) + body, return structured result. Missing file → `wiki.NotFound` error.
2. **page_write** — Create/overwrite page. `MustMutate()` + `ResolveWikiPath()` guards. `MkdirAll` for parent dirs. Skip empty frontmatter block.
3. **page_delete** — Remove page. Reject `index.md` and `log.md` (by basename after resolve). `MustMutate()` guard.
4. **page_list** — Walk wiki dir with filters: directory prefix, glob, tag (from frontmatter), updated_since (from frontmatter `updated` field).
5. **page_move** — Rename/move file + rewrite incoming links (`[[Title]]` and `[text](path)`) across all wiki pages. Compute relative paths per-file.
6. **MCP tool registration** — Define `mcp.Tool` schemas, wire handlers via `server.RegisterTool()` in main.go.
7. **Error codes** — Define constants `wiki.NotFound`, `wiki.ReadOnly`, `wiki.PathEscape` etc. Return structured errors in tool responses.
8. **Tests** — Table-driven tests using `existing/wiki/` as fixture for reads, `t.TempDir()` for writes.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented all five page CRUD tools (page_read, page_write, page_delete, page_move, page_list) as MCP tools.\n\n**Files added:**\n- `internal/wiki/wiki.go` — Core page operations, frontmatter parsing (hand-split YAML), link rewriting for page_move\n- `internal/wiki/tools.go` — MCP tool definitions (schemas + handlers) using mcp-go\n- `internal/wiki/wiki_test.go` — 20 tests including table-driven tests against existing/wiki/ fixture\n\n**Files modified:**\n- `cmd/wiki-mcp/main.go` — Wire wiki.RegisterTools(srv)\n- `go.mod` / `go.sum` — Added gopkg.in/yaml.v3\n\n**Key decisions:**\n- Hand-split frontmatter (no external library) — simpler, handles empty blocks correctly\n- Structured ToolError with stable error codes (wiki.NotFound, wiki.ReadOnly, wiki.PathEscape, wiki.Forbidden, wiki.BadRequest, wiki.Internal)\n- page_delete protects index.md and log.md by basename (catches subdir/../index.md)\n- page_move rewrites both [[Title]] wiki links and [text](path) markdown links, adjusting relative paths per-file\n- page_list filters: dir prefix, glob (filepath.Match), tag, updated_since (from frontmatter)\n- Minimal link scanner scaffolded inline (no task-9 dependency yet)"
<!-- SECTION:FINAL_SUMMARY:END -->
