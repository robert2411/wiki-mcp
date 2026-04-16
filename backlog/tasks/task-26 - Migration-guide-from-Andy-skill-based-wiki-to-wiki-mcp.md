---
id: TASK-26
title: Migration guide from Andy-skill-based wiki to wiki-mcp
status: Done
assignee: []
created_date: '2026-04-13 21:11'
updated_date: '2026-04-16 06:47'
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
- [x] #1 `docs/migration-from-skill.md` covers all 5 steps with exact commands
- [x] #2 Rollback section explains how to re-enable the old systemd stack and restore `CLAUDE.md`
- [ ] #3 Instructions tested by actually performing the migration on the user's own machine and verifying the wiki still works
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

1. Set status In Progress
2. Write docs/migration-from-skill.md covering all 5 steps with exact commands:
   - Step 1: Point wiki-mcp at existing wiki directory (config + env var)
   - Step 2: Disable old MkDocs/systemd stack (exact systemctl commands)
   - Step 3: Remove wiki sections from Andy CLAUDE.md, leave pointer note
   - Step 4: Register wiki-mcp in nanoclaw container via .mcp.json
   - Step 5: Rollback instructions
3. Check AC#1 and AC#2 — both are documentation coverage, satisfied by the doc
4. AC#3 is "tested by actually performing the migration" — document as manual verification step; mark pending user confirmation
5. Run /simplify
6. Commit
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Created `docs/migration-from-skill.md` covering all five migration steps with exact commands:

1. **Step 1** — config creation + smoke-test verify (`2>/dev/null` for clean output)
2. **Step 2** — systemd teardown with corrected disable logic (timer drives service, not independently enabled) and three-unit status check
3. **Step 3** — CLAUDE.md cleanup with pointer note template
4. **Step 4** — container binary mount (with `containerPath` → `/workspace/extra/` resolution explained) + `jq`-based `~/.claude.json` merge to avoid clobbering other MCP servers + verify steps
5. **Step 5** — rollback for systemd units and CLAUDE.md

AC#1 and AC#2 satisfied by the document. AC#3 (live migration test) requires manual user verification.
<!-- SECTION:FINAL_SUMMARY:END -->
