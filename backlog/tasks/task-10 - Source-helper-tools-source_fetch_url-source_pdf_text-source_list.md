---
id: TASK-10
title: 'Source helper tools (source_fetch_url, source_pdf_text, source_list)'
status: To Do
assignee: []
created_date: '2026-04-13 21:09'
updated_date: '2026-04-13 21:18'
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
- [ ] #1 `source_fetch_url` saves a URL to `sources_path` with a reasonable default slug
- [ ] #2 Fetcher needs no external `curl` — pure Go `net/http`
- [ ] #3 `source_pdf_text` extracts text from a fixture PDF without requiring `pdftotext` on the system
- [ ] #4 `source_list` reports size + mtime; empty list when dir absent
- [ ] #5 All tools reject paths outside configured roots
<!-- AC:END -->
