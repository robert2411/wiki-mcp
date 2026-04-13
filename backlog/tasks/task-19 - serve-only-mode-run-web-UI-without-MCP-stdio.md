---
id: TASK-19
title: '--serve-only mode: run web UI without MCP stdio'
status: To Do
assignee: []
created_date: '2026-04-13 21:10'
updated_date: '2026-04-13 21:18'
labels: []
milestone: m-3
dependencies:
  - TASK-16
documentation:
  - doc-5
  - doc-6
priority: low
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
CLI flag `--serve-only` runs the HTTP server standalone — no MCP stdio transport, no tool registry loaded. Lets users run the web UI as a daemon (launchd / systemd / Docker) without tying its lifetime to an MCP client session.

Also add `--serve` — combines stdio MCP + web UI in one process via `errgroup.Group`. Equivalent to `Web.Enabled=true` but as an explicit flag override that works even if the config file disables it.

Shutdown: install signal handlers (SIGINT, SIGTERM) that cancel the root context. `http.Server.Shutdown` drains in-flight requests with a 10s deadline; the MCP stdio loop exits on context cancel. No leaked goroutines.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `wiki-mcp --serve-only --port 9000` serves the wiki over HTTP without opening stdio MCP
- [ ] #2 `wiki-mcp --serve` starts both stdio MCP and web UI concurrently regardless of `Web.Enabled`
- [ ] #3 SIGINT/SIGTERM shut down cleanly in both modes with no leaked goroutines (verified via `goleak`)
- [ ] #4 `http.Server.Shutdown` drains in-flight requests within a 10s deadline
<!-- AC:END -->
