---
id: TASK-12
title: 'MCP prompt: ingest'
status: To Do
assignee: []
created_date: '2026-04-13 21:09'
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
- [ ] #1 `ingest` prompt listed via `list_prompts` with declared args
- [ ] #2 Filled prompt names the exact tool calls the LLM should make (page_write, index_upsert_entry, log_append, etc.)
- [ ] #3 Prompt preserves the 'one source at a time, fully finish' discipline from existing/skill.md
- [ ] #4 Integration test: call `get_prompt('ingest', {source: 'https://example.com'})` returns the expected instruction text
- [ ] #5 Prompt text kept in a versioned template file inside the package so changes are reviewable in diffs
<!-- AC:END -->
