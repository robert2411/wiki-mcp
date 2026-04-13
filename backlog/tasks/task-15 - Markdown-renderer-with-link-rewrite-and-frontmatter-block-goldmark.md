---
id: TASK-15
title: Markdown renderer with link-rewrite and frontmatter block (goldmark)
status: To Do
assignee: []
created_date: '2026-04-13 21:09'
updated_date: '2026-04-13 21:18'
labels: []
milestone: m-3
dependencies:
  - TASK-3
documentation:
  - doc-6
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Page-rendering pipeline per doc-6, implemented in Go with `github.com/yuin/goldmark`.

Goldmark setup: enable `extension.GFM` (tables, strikethrough, task lists, autolink), `extension.Typographer`, `extension.Footnote`. Add a custom wikilink parser extension for `[[Title]]` and `[[Title|alias]]` that resolves against a page-title index computed from disk (H1 title or filename-derived). Markdown `[text](path.md)` links get `.md` stripped in the emitted URL.

Frontmatter: parse via `github.com/adrg/frontmatter` or hand-split before goldmark sees it; render as a styled metadata block (tags, `updated`, etc.) prepended to the HTML.

Deliver a pure function:
```go
func RenderPage(c *config.Config, relPath string) (*RenderedPage, error)
```

No HTTP server in this task — just the rendering package in `internal/web/render`. Must be unit-testable without network or file watching.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `RenderPage` returns HTML + resolved title + metadata for all 21 pages in `existing/wiki/` without errors
- [ ] #2 `[[Qwen3.5]]` in a page resolves to the URL of `entities/qwen3.5.md` (stripped of `.md`)
- [ ] #3 `[[Missing Title]]` renders with a distinct CSS class (e.g. `broken-link`), not a 500
- [ ] #4 Frontmatter renders as a styled metadata block, not raw YAML
- [ ] #5 Tables, fenced code, task lists render correctly (tested on `summary.md`)
<!-- AC:END -->
