---
id: TASK-25
title: Verify integration with at least one non-Claude MCP host
status: To Do
assignee: []
created_date: '2026-04-13 21:11'
labels: []
milestone: m-5
dependencies:
  - TASK-24
documentation:
  - doc-3
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Prove AI-agnosticism. Pick one of: Cursor, Cline, Continue, Zed, Windsurf, or a plain `mcp-cli`/`mcp dev` harness. Run the same end-to-end ingest + query + lint flow used in the Claude hosts test.

Goal isn't feature parity across hosts (each host exposes MCP primitives differently) — it's proving the server's tool/resource/prompt surface works without Claude-specific assumptions. If a host doesn't support prompts, document that limitation and show how the Ingest flow still works via direct tool calls.

File per-host config snippet + caveats under `docs/clients/<host>.md`.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 At least one non-Claude MCP host successfully connects and lists tools
- [ ] #2 Basic page_write + log_append flow succeeds from that host
- [ ] #3 Host-specific doc filed under `docs/clients/`
- [ ] #4 Any host-specific workarounds (e.g. prompts-not-supported fallbacks) documented
<!-- AC:END -->
