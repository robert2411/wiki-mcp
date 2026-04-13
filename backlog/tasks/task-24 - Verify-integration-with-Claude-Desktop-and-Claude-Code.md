---
id: TASK-24
title: Verify integration with Claude Desktop and Claude Code
status: To Do
assignee: []
created_date: '2026-04-13 21:11'
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
- [ ] #2 Claude Code successfully lists and invokes the same surface
- [ ] #3 A full ingest workflow completes end-to-end in each host using `existing/wiki/` as the target
- [ ] #4 Per-host config snippets filed under `docs/clients/` with observed quirks noted
- [ ] #5 Screenshots or terminal transcripts captured in the docs
<!-- AC:END -->
