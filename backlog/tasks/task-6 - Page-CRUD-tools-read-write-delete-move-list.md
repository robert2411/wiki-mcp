---
id: TASK-6
title: 'Page CRUD tools (read, write, delete, move, list)'
status: To Do
assignee: []
created_date: '2026-04-13 21:08'
updated_date: '2026-04-13 21:17'
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
- [ ] #1 `page_read` returns frontmatter + body; missing file returns a structured not-found error (not a panic)
- [ ] #2 `page_write` creates parent dirs as needed; refuses paths outside `wiki_path`; refuses when `ReadOnly=true`
- [ ] #3 `page_write` does not emit an empty frontmatter block when no metadata is provided
- [ ] #4 `page_delete` rejects `index.md` and `log.md` with a clear reason
- [ ] #5 `page_move` rewrites `[[Title]]` and `[text](rel/path)` across the wiki and fixes the moved page's own outgoing links
- [ ] #6 `page_list` supports dir/glob/tag/updated_since filters (table-driven test against `existing/wiki/` as fixture)
- [ ] #7 All tools return structured errors with stable error codes
<!-- AC:END -->
