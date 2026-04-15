---
id: TASK-13
title: 'MCP prompt: query'
status: Done
assignee: []
created_date: '2026-04-13 21:09'
updated_date: '2026-04-14 22:30'
labels: []
milestone: m-2
dependencies:
  - TASK-5
references:
  - existing/skill.md
documentation:
  - doc-2
  - doc-3
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Query workflow as MCP prompt. Base on `existing/skill.md` § Query.

Flow instructed by the prompt: read `wiki://index` → identify relevant pages → read them in full via `page_read` or `wiki://page/<path>` → synthesize with citations → offer to file the answer as a new page when the answer is substantial and reusable.

Prompt args:
- `question` (required)
- `file_answer` (optional bool): if true, skip the "offer" step and file the answer directly.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `query` prompt listed via `list_prompts`
- [x] #2 Filled prompt instructs index-first reading and citation discipline
- [x] #3 `file_answer=true` branch instructs the LLM to call `page_write` + `index_upsert_entry` + `log_append`
- [x] #4 Integration test against prompt template
<!-- AC:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented the `query` MCP prompt in `internal/wiki/prompts.go`.

Template at `internal/wiki/prompts/query.md`. Uses `{{.FileAnswer}}` conditional: false branch says offer; true branch instructs direct `page_write` + `index_upsert_entry` + `log_append`. Registered alongside `ingest` in `RegisterPrompts`. 3 query tests + full suite green.
<!-- SECTION:FINAL_SUMMARY:END -->
