---
id: TASK-14
title: 'MCP prompt: lint'
status: To Do
assignee: []
created_date: '2026-04-13 21:09'
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
- [ ] #1 `lint` prompt listed via `list_prompts` with no args
- [ ] #2 Prompt references the `orphans` and `links_incoming` tools explicitly
- [ ] #3 Prompt instructs computing next pass number from log history
- [ ] #4 Prompt sets the log-append header format exactly: `## [YYYY-MM-DD] lint | pass N`
<!-- AC:END -->
