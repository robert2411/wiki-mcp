---
id: TASK-5
title: 'MCP server bootstrap in Go (stdio transport, tool/resource/prompt registry)'
status: Done
assignee: []
created_date: '2026-04-13 21:08'
updated_date: '2026-04-14 19:05'
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
- [x] #1 MCP SDK choice (official or mark3labs) documented in task implementation notes with rationale
- [x] #2 `wiki-mcp` runs as stdio MCP server and passes a handshake with a reference MCP client
- [x] #3 Tool/resource/prompt registration is a registry that M2/M3 tasks extend without touching `server.go`
- [x] #4 Config loaded on startup; missing `wiki_path` aborts with the loader's error message
- [x] #5 `--transport sse` returns a clear NotImplemented error pointing at task-22
- [x] #6 Context cancellation on SIGINT closes the stdio loop within 1s (verified by test)
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. `go get github.com/mark3labs/mcp-go`\n2. Implement `internal/server/server.go` with Server struct, New(), registration methods, Run(ctx)\n3. Update `cmd/wiki-mcp/main.go` with --transport flag, signal handling, server wiring\n4. Write tests for handshake, signal shutdown, SSE error\n5. Verify all ACs pass
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
**SDK Choice: `github.com/mark3labs/mcp-go`**\n\nRationale:\n- Supports tools + resources + prompts with clean builder APIs\n- Stdio transport stable with `StdioServer.Listen(ctx, stdin, stdout)` accepting context.Context\n- SSE + streamable-HTTP transports built-in (`NewSSEServer`, `NewStreamableHTTPServer`) — ready for TASK-22\n- Ergonomic for ~25 tools: `s.AddTool(mcp.NewTool(...), handler)` pattern\n- Official go-sdk (`modelcontextprotocol/go-sdk`) has sparse docs on resources/prompts, no SSE transport confirmed\n\n**Architecture:**\n- `internal/server/server.go` — Server struct wrapping MCPServer, New() constructor, Run(ctx) method\n- Tool/resource/prompt registration via methods on Server struct; M2+ tasks call these from their own packages\n- `cmd/wiki-mcp/main.go` — add `--transport` flag, signal.NotifyContext for SIGINT/SIGTERM, wire config → server
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## MCP Server Bootstrap (TASK-5)\n\n### SDK Choice\nSelected `github.com/mark3labs/mcp-go` v0.48.0 over official `modelcontextprotocol/go-sdk`.\nRationale: full tools+resources+prompts support, stable stdio with context.Context, built-in SSE/streamable-HTTP for TASK-22.\n\n### Changes\n- **`internal/server/server.go`** — `Server` struct wrapping `MCPServer`. Exposes `RegisterTool`, `RegisterResource`, `RegisterResourceTemplate`, `RegisterPrompt` methods for M2+ tasks. `RunStdio(ctx, stdin, stdout)` runs stdio transport with slog error logging.\n- **`cmd/wiki-mcp/main.go`** — Added `--transport` flag (stdio|sse). `signal.NotifyContext` for clean SIGINT/SIGTERM shutdown. SSE rejected with clear error pointing to TASK-22.\n- **`internal/server/server_test.go`** — Tests: JSON-RPC initialize handshake verifying server name + capabilities, context cancellation shutdown <1s, SSE error message.\n\n### Test Results\nAll tests pass: `go test ./...` — server handshake, context shutdown, SSE error, config loading, binary build."
<!-- SECTION:FINAL_SUMMARY:END -->
