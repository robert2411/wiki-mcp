---
id: TASK-14
title: 'MCP prompt: lint'
status: Done
assignee:
  - Robert Stevens
created_date: '2026-04-13 21:09'
updated_date: '2026-04-15 19:28'
labels: []
milestone: m-2
dependencies:
  - TASK-5
references:
  - existing/skill.md
documentation:
  - doc-2
  - doc-3
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Lint workflow as MCP prompt. Base on `existing/skill.md` § Lint and the existing successful pass recorded at the end of `existing/wiki/log.md` (§ "## [2026-04-11] lint | pass 1").

Flow: read index → scan pages for contradictions, orphans (via `orphans` tool), staleness, gaps, missing cross-refs → report findings → offer fixes → on completion append `## [YYYY-MM-DD] lint | pass N` via `log_append`.

No args. Pass number comes from scanning `log.md` for existing `lint | pass N` entries and incrementing.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `lint` prompt listed via `list_prompts` with no args
- [x] #2 Prompt references the `orphans` and `links_incoming` tools explicitly
- [x] #3 Prompt instructs computing next pass number from log history
- [x] #4 Prompt sets the log-append header format exactly: `## [YYYY-MM-DD] lint | pass N`
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

### Files to create/modify
1. `internal/wiki/prompts/lint.md` — new prompt template (no args, static text)
2. `internal/wiki/prompts.go` — add embed, promptDef, handler, register
3. `internal/wiki/prompts_test.go` — add tests for lint prompt

### Approach

**lint.md template**: Static (no template vars needed — no args). Must:
- Instruct reading `index_read()` first
- Instruct calling `orphans()` to find orphan pages
- Instruct calling `links_incoming(path)` for each candidate page to verify cross-ref gaps
- Instruct scanning for contradictions, staleness, gaps
- Instruct computing next pass number: call `log_tail(n=50)`, scan for entries where title matches `lint | pass N`, find max N, use N+1 (default 1 if none found)
- Instruct appending log with `log_append(operation="lint", title="lint | pass N", body="...")`
- The header format in log will be `## [YYYY-MM-DD] lint | pass N` (handled by log_append)

**prompts.go changes**:
- Add `//go:embed prompts/lint.md` + `var lintPromptTemplate string`
- Add `lintPromptDef()` — name "lint", no args
- Add `handleLintPrompt()` — simple static handler (no template data needed, just return the template text directly)
- Register in `RegisterPrompts`

**Tests**: Mirror pattern from existing prompt tests:
- `TestLintPromptDef_NoArgs` — verify name is "lint", 0 args
- `TestLintPrompt_Text` — verify `orphans`, `links_incoming`, pass number logic, log format present
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented `lint` MCP prompt in `internal/wiki/prompts/lint.md` with no args. Prompt instructs: read `index_read()`, compute pass number via `log_tail(n=50)` scanning for `lint | pass N` entries, call `orphans()` for orphan detection, call `links_incoming(path)` for cross-ref gaps, scan all pages for contradictions/staleness/gaps, report findings, then `log_append(operation="lint", title="lint | pass N", body="...")`. Registered in `prompts.go` alongside ingest/query. Tests in `prompts_test.go` verify no args, tool references, and log format.
<!-- SECTION:FINAL_SUMMARY:END -->
