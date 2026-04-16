---
id: TASK-22
title: SSE / streamable-http transport for remote MCP access
status: Done
assignee: []
created_date: '2026-04-13 21:10'
updated_date: '2026-04-16 00:13'
labels: []
milestone: m-4
dependencies:
  - TASK-5
documentation:
  - doc-5
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Implement the `--transport sse` placeholder from task-5. Enables remote AI hosts to connect.

Which transport: use the MCP SDK's current recommendation at implementation time — streamable HTTP (MCP 2025-06 spec) is the forward path; plain SSE (legacy) may still be supported by specific clients. Implement whichever the chosen SDK exposes stably; feature-flag if both are needed.

CLI: `--transport sse`, `--mcp-port` (separate from web UI port), `--bind` (default `127.0.0.1`). Binding to `0.0.0.0` is allowed but logs a warning: user accepts responsibility for firewall.

Optional `--auth-token <token>` — server rejects requests without matching `Authorization: Bearer <token>` header. Constant-time comparison (`crypto/subtle.ConstantTimeCompare`).

No other auth. Threat model: trusted home LAN.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `wiki-mcp --transport sse --mcp-port 8765` accepts a local MCP client connection
- [x] #2 Default bind is `127.0.0.1`; `--bind 0.0.0.0` is explicit and logged with a warning
- [x] #3 Works against the MCP SDK's reference client
- [x] #4 `--auth-token` check uses constant-time compare and rejects missing/wrong tokens with 401
- [x] #5 Documentation notes threat model: trusted LAN only, use a reverse proxy + TLS for anything beyond
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

1. **Config**: Add `MCPConfig` struct (Port=8765, Bind=127.0.0.1, AuthToken) to `Config`. Wire env vars and CLI flags.
2. **CLI flags**: `--mcp-port` → `cfg.MCP.Port`; `--auth-token` → `cfg.MCP.AuthToken`; `--bind` also sets `cfg.MCP.Bind` (shared with web UI bind flag).
3. **Server**: Add `RunStreamableHTTP(ctx)` to `server.go` using `mcpserver.NewStreamableHTTPServer`. Wrap handler with auth middleware if AuthToken set. Log warning when Bind is `0.0.0.0`.
4. **main.go**: Replace the SSE-not-implemented error with actual startup of the streamable HTTP server.
5. **Tests**: Add basic tests for auth middleware.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented streamable-HTTP (MCP 2025-03 spec) transport via `mcp-go`'s `NewStreamableHTTPServer`.

**Changes:**
- `internal/config/config.go`: added `MCPConfig{Port,Bind,AuthToken}` with defaults (port 8765, bind 127.0.0.1), TOML tags, env vars (`WIKI_MCP_MCP_PORT/BIND/AUTH_TOKEN`), and CLI flag wiring via `Flags.MCPPort/AuthToken`. `--bind` now sets both Web and MCP bind addresses.
- `internal/server/server.go`: added `RunStreamableHTTP(ctx)` — starts streamable-HTTP server, warns on `0.0.0.0`, and wraps with `BearerAuthMiddleware` when `AuthToken` is set. `BearerAuthMiddleware` uses `crypto/subtle.ConstantTimeCompare` for timing-safe token comparison, returning 401 on mismatch. Retained `ErrSSENotImplemented` for backwards compat.
- `cmd/wiki-mcp/main.go`: added `--mcp-port` and `--auth-token` flags; when `--transport sse` is active, runs `RunStreamableHTTP` instead of stdio.
- `internal/server/server_test.go`: added `TestBearerAuthMiddleware` (4 sub-tests) and `TestRunStreamableHTTP_BindsAndResponds`.

All 9 packages pass tests.
<!-- SECTION:FINAL_SUMMARY:END -->
