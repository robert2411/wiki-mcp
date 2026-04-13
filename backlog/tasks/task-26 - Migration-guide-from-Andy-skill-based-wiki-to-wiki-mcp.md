---
id: TASK-26
title: Migration guide from Andy-skill-based wiki to wiki-mcp
status: To Do
assignee: []
created_date: '2026-04-13 21:11'
labels: []
milestone: m-5
dependencies:
  - TASK-24
references:
  - existing/skill.md
  - existing/CLAUDE.md.back.md
  - existing/wiki/infrastructure/wiki-ui.md
documentation:
  - doc-1
  - doc-5
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Document the transition from the current skill-based setup (described in `existing/skill.md`, `existing/CLAUDE.md.back.md`, and `existing/wiki/infrastructure/wiki-ui.md`) to the MCP server.

Cover:

1. Point wiki-mcp at the existing wiki directory — no data migration needed.
2. Disable the old MkDocs/systemd stack (`systemctl --user disable --now nanoclaw-wiki nanoclaw-wiki-rebuild.timer nanoclaw-wiki-rebuild.service`). Reclaim port 9000.
3. Remove the wiki-related sections from Andy's `CLAUDE.md` — the MCP server's prompts now carry that semantics. Leave a pointer note.
4. Register the MCP server inside the nanoclaw container so Andy continues to have wiki access through the unified surface.
5. Rollback instructions in case something breaks.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `docs/migration-from-skill.md` covers all 5 steps with exact commands
- [ ] #2 Rollback section explains how to re-enable the old systemd stack and restore `CLAUDE.md`
- [ ] #3 Instructions tested by actually performing the migration on the user's own machine and verifying the wiki still works
<!-- AC:END -->
