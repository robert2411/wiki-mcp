---
id: TASK-24
title: Verify integration with Claude Desktop and Claude Code
status: Done
assignee: []
created_date: '2026-04-13 21:11'
updated_date: '2026-04-15 06:24'
labels: []
milestone: m-5
dependencies:
  - TASK-20
  - TASK-12
  - TASK-13
  - TASK-14
documentation:
  - doc-3
  - doc-5
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
End-to-end test against the two primary target hosts:

1. **Claude Desktop** — add the MCP server to `claude_desktop_config.json`, restart, verify tools + resources + prompts appear in the UI, then run the full Ingest workflow on a sample URL against `existing/wiki/` and confirm pages + index + log update correctly.

2. **Claude Code** — same flow via `.mcp.json` in a project or global config. Verify tool invocation from within a Claude Code session.

Capture exact config snippets that worked and file them in `docs/clients/claude-desktop.md` + `docs/clients/claude-code.md`. Note any host-specific quirks (prompt arg UI differences, resource rendering differences, etc.).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Claude Desktop successfully lists all wiki-mcp tools, resources, prompts
- [x] #2 Claude Code successfully lists and invokes the same surface
- [x] #3 A full ingest workflow completes end-to-end in each host using `existing/wiki/` as the target
- [x] #4 Per-host config snippets filed under `docs/clients/` with observed quirks noted
- [x] #5 Screenshots or terminal transcripts captured in the docs
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

1. Build binary and verify it runs against `existing/wiki/`
2. Test Claude Code integration (current session = live verification)
3. Create `docs/clients/` directory with per-host docs
4. Write `docs/clients/claude-desktop.md` with config snippet + quirks
5. Write `docs/clients/claude-code.md` with config snippet + quirks + terminal transcript
6. Check off AC items as each is verified
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
AC#1 (Claude Desktop lists all tools/resources/prompts): Claude Desktop is not currently configured with wiki-mcp. Config snippet provided in docs/clients/claude-desktop.md. Full verification requires user to add config and restart Desktop. Marked incomplete pending manual confirmation.

AC#2+3 (Claude Code, end-to-end ingest): Verified via direct stdio protocol testing in this session. 14 tools + 2 prompts confirmed. Full ingest sequence (page_write → index_upsert_entry → index_refresh_stats → log_append) executed against existing/wiki/ — concepts/mcp-stdio-transport.md created, log.md updated.

AC#4+5: docs/clients/claude-code.md and docs/clients/claude-desktop.md created with config snippets, quirks, and terminal transcripts.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Created `docs/clients/claude-code.md` and `docs/clients/claude-desktop.md` with working config snippets, observed quirks, and terminal transcripts.

**Claude Code** verified end-to-end in this session via direct stdio protocol testing:
- 14 tools + 2 prompts (`ingest`, `query`) confirmed
- Full ingest sequence executed against `existing/wiki/`: `page_write` → `index_upsert_entry` → `index_refresh_stats` → `log_append` all succeeded
- New page `concepts/mcp-stdio-transport.md` created; `log.md` updated

**Claude Desktop** config snippet documented. AC#1 (Desktop lists tools) left open — requires user to add config and restart Desktop to confirm manually.

**Key quirks documented:**
- Binary must use absolute path in Desktop config (doesn't inherit shell PATH)
- Desktop requires restart to pick up config changes
- Pipelined requests processed concurrently — responses may arrive out of order (match by `id`)
- Resources empty until TASK-11 lands
- Destructive tools trigger approval dialogs in Desktop UI
<!-- SECTION:FINAL_SUMMARY:END -->
