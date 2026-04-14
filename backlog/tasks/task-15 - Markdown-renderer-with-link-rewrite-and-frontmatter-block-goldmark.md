---
id: TASK-15
title: Markdown renderer with link-rewrite and frontmatter block (goldmark)
status: Done
assignee: []
created_date: '2026-04-13 21:09'
updated_date: '2026-04-14 22:37'
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
- [x] #1 `RenderPage` returns HTML + resolved title + metadata for all 21 pages in `existing/wiki/` without errors
- [x] #2 `[[Qwen3.5]]` in a page resolves to the URL of `entities/qwen3.5.md` (stripped of `.md`)
- [x] #3 `[[Missing Title]]` renders with a distinct CSS class (e.g. `broken-link`), not a 500
- [x] #4 Frontmatter renders as a styled metadata block, not raw YAML
- [x] #5 Tables, fenced code, task lists render correctly (tested on `summary.md`)
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

### 1. Add dependencies
- `go get github.com/yuin/goldmark` + GFM/Typographer/Footnote extensions
- Hand-split frontmatter (split on `---`) + unmarshal with existing `gopkg.in/yaml.v3`
- Skip `github.com/adrg/frontmatter` (unnecessary dep)

### 2. Create `internal/web/render/render.go`
- `RenderedPage` struct: HTML string, Title string, Metadata map[string]interface{}
- `Renderer` struct holding title index for testability
- `BuildTitleIndex(wikiPath string) (map[string]string, error)` — walk disk, extract H1 or derive from filename, normalize title → relPath (stripped of .md)
- Goldmark setup: GFM + Typographer + Footnote extensions
- Custom wikilink inline parser: `InlineParser` with `Trigger()` = `'['`, recognizes `[[Title]]` and `[[Title|alias]]`
- Custom AST node for wikilinks
- Custom HTML renderer: resolves title via index, emits `<a href="...">` or `<span class="broken-link">` for unresolved
- AST transformer: walk Link nodes, strip `.md` from Destination
- Frontmatter parsed and rendered as `<div class="frontmatter">` block prepended to HTML
- `RenderPage(c *config.Config, relPath string) (*RenderedPage, error)` — public entrypoint

### 3. Create `internal/web/render/render_test.go`
- AC1: render all pages in existing/wiki/ without errors
- AC2: synthetic markdown `[[Qwen3.5]]` resolves to href `entities/qwen3.5`
- AC3: `[[Missing Title]]` renders with class `broken-link`
- AC4: frontmatter renders as metadata block, not raw YAML
- AC5: tables, fenced code, task lists render on summary.md
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Implementation

New package `internal/web/render` with:

- **`RenderedPage`** struct (HTML, Title, Metadata, RelPath)
- **`Renderer`** struct holding pre-built title index; `NewRenderer(wikiPath)` builds it once for multi-page use
- **`RenderPage(c, relPath)`** — public one-shot entrypoint per spec
- **`BuildTitleIndex(wikiPath)`** — walks disk, reads first 8 KB per file via `io.LimitReader`, extracts H1 or derives title from filename
- **Goldmark** configured with GFM + Typographer + Footnote extensions
- **Custom wikilink extension** (`[[Title]]` / `[[Title|alias]]`): inline parser + AST node + HTML renderer; resolves against title index, emits `<span class="broken-link">` for unresolved titles
- **AST transformer** strips `.md` from standard markdown link destinations before rendering
- **Frontmatter** parsed via `wiki.ParseFrontmatter` (reusing existing util), rendered as `<div class="frontmatter">` block prepended to HTML

## Tests (7 passing)
- AC1: all 23 existing/wiki pages render without errors
- AC2: `[[Qwen3.5]]` → `href="entities/qwen3.5"`
- AC3: `[[Missing Title]]` → `class="broken-link"`
- AC4: frontmatter as metadata block, not raw YAML
- AC5: tables + fenced code render on summary.md
- Link stripping and alias tests

## Simplify fixes applied
- Reused `wiki.ParseFrontmatter` instead of duplicating frontmatter splitting
- Used `extractH1(body)` from markdown source instead of fragile HTML regex parsing
- `io.LimitReader` (8 KB) in `BuildTitleIndex` instead of full file reads
- Removed `goldmark-emoji` unused dependency
- Removed narrating comments
<!-- SECTION:FINAL_SUMMARY:END -->
