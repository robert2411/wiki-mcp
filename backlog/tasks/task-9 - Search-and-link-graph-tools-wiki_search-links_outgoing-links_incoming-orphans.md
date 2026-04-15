---
id: TASK-9
title: >-
  Search and link-graph tools (wiki_search, links_outgoing, links_incoming,
  orphans)
status: Done
assignee: []
created_date: '2026-04-13 21:08'
updated_date: '2026-04-15 06:04'
labels: []
milestone: m-1
dependencies:
  - TASK-5
documentation:
  - doc-2
  - doc-3
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Implement full-text search and link-graph tools in Go per doc-3.

`wiki_search(query, limit=20)`: substring/regex scan across page bodies (not frontmatter). Return `[{path, title, snippets: [string], score}]`. Naive in-memory scan is fine for wikis up to a few thousand pages — no `bleve`/`sqlite-fts` dependency unless we hit real performance issues.

`links_outgoing(path)`: parse a page and return all links. Handle `[[Title]]`, `[[Title|alias]]`, `[text](relative/path.md)`, `[text](./dir/file.md)`. Return `{internal: [...], external: [url]}`.

`links_incoming(path)`: scan all pages, return backlinks.

`orphans()`: pages with zero incoming links. Exclude `index.md` and `log.md`.

Build a shared `internal/wiki/linkgraph` package so page_move (task-6) reuses link resolution.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `wiki_search` returns scored snippet results across `existing/wiki/` fixture
- [x] #2 `links_outgoing` correctly parses both Obsidian `[[Title]]` (including `|alias`) and markdown `[text](path)` forms
- [x] #3 `links_incoming` produces expected backlinks on a known fixture
- [x] #4 `orphans()` output on `existing/wiki/` matches the 2 orphans noted in the wiki's lint pass 1 (`concepts/spring-boot-maven-docker.md`, `infrastructure/wiki-ui.md`)
- [x] #5 `internal/wiki/linkgraph` package is importable and used by the page_move implementation
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

### 1. Create `internal/wiki/linkgraph/linkgraph.go`
Pure link parsing — no config/ToolError dependency.
- `Links` struct: `Internal []string`, `External []string`
- `ParseOutgoing(relPath, content string) Links`
  - `[[Title]]` → title string in Internal
  - `[[Title|alias]]` → title string in Internal
  - `[text](./path)` / `[text](../path)` → wiki-root-relative path in Internal
  - `[text](https://...)` → URL in External
- `linkgraph_test.go` covering all four link forms

### 2. Update `internal/wiki/wiki.go`
- Import linkgraph package
- Refactor `PageMove`'s `fixOutgoingLinks` helper to use `linkgraph.ParseOutgoing` for scanning outgoing links of the moved page (satisfies AC#5)

### 3. Create `internal/wiki/graph.go`
- `LinksOutgoingResult` struct
- `LinksOutgoing(cfg, relPath) (*LinksOutgoingResult, *ToolError)`
- `LinksIncoming(cfg, relPath) ([]string, *ToolError)` — scan all pages via `linkgraph.ParseOutgoing`; match both path-resolved internal links and title-derived wikilinks
- `Orphans(cfg) ([]string, *ToolError)` — exclude index.md + log.md as BOTH candidates and link sources
- `graph_test.go` against existing/wiki fixture

### 4. Create `internal/wiki/search.go`
- `WikiSearchResult` struct: `{Path, Title, Snippets, Score}`
- `WikiSearch(cfg, query string, limit int) ([]WikiSearchResult, *ToolError)`
  - Substring scan (case-insensitive); regex if query contains regex metacharacters
  - Score = match count; Snippets = up to 3 context excerpts
- `search_test.go` against existing/wiki fixture

### 5. Update `internal/wiki/tools.go`
Register 4 new tools:
- `wiki_search(query, limit=20)`
- `links_outgoing(path)`
- `links_incoming(path)`
- `orphans()`

### Orphan logic
- When building incoming link map, skip index.md and log.md as sources
- Exclude index.md and log.md from orphan candidates
- Result: `concepts/spring-boot-maven-docker.md` and `infrastructure/wiki-ui.md` confirmed orphans
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Implementation Summary

Implemented full-text search and link-graph tools for the wiki MCP server.

### New files
- **`internal/wiki/linkgraph/linkgraph.go`** — pure link parsing package (no config/ToolError deps). `ParseOutgoing(relPath, content)` handles `[[Title]]`, `[[Title|alias]]`, `[text](./path)`, `[text](../path)`, and external URLs. Deduplicates results.
- **`internal/wiki/linkgraph/linkgraph_test.go`** — unit tests for all link forms.
- **`internal/wiki/graph.go`** — `LinksOutgoing`, `LinksIncoming`, `Orphans` using linkgraph. Orphans excludes index.md/log.md as both candidates and link sources (confirmed: `concepts/spring-boot-maven-docker.md` and `infrastructure/wiki-ui.md` are the 2 known orphans).
- **`internal/wiki/graph_test.go`** — tests against temp fixtures and existing/wiki fixture.
- **`internal/wiki/search.go`** — `WikiSearch` with case-insensitive substring/regex scan, scored results, snippet extraction.
- **`internal/wiki/search_test.go`** — tests for scoring, ordering, limit, frontmatter exclusion.

### Modified files
- **`internal/wiki/wiki.go`** — `fixOutgoingLinks` refactored to use `linkgraph.ParseOutgoing` (satisfies AC#5). Now builds a rewrite map then does a single regex pass instead of one pass per link.
- **`internal/wiki/tools.go`** — registered 4 new tools: `wiki_search`, `links_outgoing`, `links_incoming`, `orphans`. Fixed pre-existing bug: hardcoded `"today"` literal → `time.Now().Format("2006-01-02")`.

### Simplify fixes applied
- Orphans() merged from two WalkDir passes into one
- fixOutgoingLinks() builds rewrite map then single ReplaceAllStringFunc pass
- extractSnippets() replaced per-byte map with lastEnd pointer
- existingWikiConfig() in graph_test.go delegates to existingWikiPath() (same package)
<!-- SECTION:FINAL_SUMMARY:END -->
