---
id: TASK-11
title: 'MCP resources (wiki://index, wiki://log/recent, wiki://page/*, wiki://config)'
status: To Do
assignee: []
created_date: '2026-04-13 21:09'
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
- [ ] #1 All four resource URIs resolve with correct MIME types (text/markdown for pages, application/json for config)
- [ ] #2 `list_resources` includes the page-per-URI entries up to the cap
- [ ] #3 Reads enforce `confine_to_wiki_path`
- [ ] #4 Integration test: an MCP test client reads all four resource shapes against `existing/wiki/`
<!-- AC:END -->
