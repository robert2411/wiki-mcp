---
id: TASK-9
title: >-
  Search and link-graph tools (wiki_search, links_outgoing, links_incoming,
  orphans)
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
- [ ] #1 `wiki_search` returns scored snippet results across `existing/wiki/` fixture
- [ ] #2 `links_outgoing` correctly parses both Obsidian `[[Title]]` (including `|alias`) and markdown `[text](path)` forms
- [ ] #3 `links_incoming` produces expected backlinks on a known fixture
- [ ] #4 `orphans()` output on `existing/wiki/` matches the 2 orphans noted in the wiki's lint pass 1 (`concepts/spring-boot-maven-docker.md`, `infrastructure/wiki-ui.md`)
- [ ] #5 `internal/wiki/linkgraph` package is importable and used by the page_move implementation
<!-- AC:END -->
