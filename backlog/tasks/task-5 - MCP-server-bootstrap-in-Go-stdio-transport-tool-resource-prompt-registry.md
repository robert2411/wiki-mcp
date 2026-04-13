---
id: TASK-5
title: 'MCP server bootstrap in Go (stdio transport, tool/resource/prompt registry)'
status: To Do
assignee: []
created_date: '2026-04-13 21:08'
updated_date: '2026-04-13 21:17'
labels: []
milestone: m-1
dependencies:
  - TASK-3
documentation:
  - doc-3
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Stand up the MCP server using `github.com/modelcontextprotocol/go-sdk` (primary) or `github.com/mark3labs/mcp-go` (fallback). Evaluate both on these axes before picking:

- Does it support tools + resources + prompts?
- Is stdio transport stable?
- Does it expose a hook for streamable-http / SSE (task-22)?
- Is the surface ergonomic for ~25 tools without boilerplate blowup?

Pick one and record the pick in an `implementationNotes` update on this task.

Deliver a registration pattern (function registry keyed by tool name) so subsequent M2 tasks add tools by calling e.g. `srv.RegisterTool(PageRead)`. Register zero tools in this task; just prove the handshake works.

At startup: load config via task-3's loader, then log (via `log/slog`) the resolved wiki path + read-only state + transport mode. Support `--transport {stdio,sse}` — `sse` returns `errors.New("not implemented: see task-22")` until M5 lands.

Use a cancellable `context.Context` threaded through all tool handlers so shutdown on SIGINT/SIGTERM is clean.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 MCP SDK choice (official or mark3labs) documented in task implementation notes with rationale
- [ ] #2 `wiki-mcp` runs as stdio MCP server and passes a handshake with a reference MCP client
- [ ] #3 Tool/resource/prompt registration is a registry that M2/M3 tasks extend without touching `server.go`
- [ ] #4 Config loaded on startup; missing `wiki_path` aborts with the loader's error message
- [ ] #5 `--transport sse` returns a clear NotImplemented error pointing at task-22
- [ ] #6 Context cancellation on SIGINT closes the stdio loop within 1s (verified by test)
<!-- AC:END -->
