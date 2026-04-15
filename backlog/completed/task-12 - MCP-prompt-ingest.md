---
id: TASK-12
title: 'MCP prompt: ingest'
status: Done
assignee: []
created_date: '2026-04-13 21:09'
updated_date: '2026-04-14 22:29'
labels: []
milestone: m-2
dependencies:
  - TASK-5
references:
  - existing/skill.md
  - existing/CLAUDE.md.back.md
documentation:
  - doc-2
  - doc-3
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Ship the Ingest workflow as an MCP prompt so any host can trigger it by name.

Base the prompt text on `existing/skill.md` § Ingest and `existing/CLAUDE.md.back.md` § Ingest Discipline. Keep the "one source at a time, fully finish before next" rule explicit. Inline a tool inventory so the client LLM knows exactly which MCP tools to call at each step (`source_fetch_url` / `source_pdf_text` → `page_read`/`page_write` → `index_upsert_entry` → `log_append`).

Prompt args:
- `source` (required): URL, local path to PDF/image, or raw text
- `hint` (optional): caller's note about what they care about in the source

Return the prompt as a MCP `GetPromptResult` with a single user message containing the filled-in instructions.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `ingest` prompt listed via `list_prompts` with declared args
- [x] #2 Filled prompt names the exact tool calls the LLM should make (page_write, index_upsert_entry, log_append, etc.)
- [x] #3 Prompt preserves the 'one source at a time, fully finish' discipline from existing/skill.md
- [x] #4 Integration test: call `get_prompt('ingest', {source: 'https://example.com'})` returns the expected instruction text
- [x] #5 Prompt text kept in a versioned template file inside the package so changes are reviewable in diffs
<!-- AC:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented the `ingest` MCP prompt in `internal/wiki/prompts.go`.

Template stored in `internal/wiki/prompts/ingest.md` (embedded via `//go:embed`), so changes are diff-reviewable. Uses `text/template` to fill `{{.Source}}` and optional `{{.Hint}}`.

Handler names all tool calls explicitly (source_fetch_url, source_pdf_text, page_read, page_write, index_upsert_entry, index_refresh_stats, log_append) and preserves the one-source-at-a-time discipline. Registered via `RegisterPrompts(srv)` called from main.

3 prompt tests pass, full suite green.
<!-- SECTION:FINAL_SUMMARY:END -->
