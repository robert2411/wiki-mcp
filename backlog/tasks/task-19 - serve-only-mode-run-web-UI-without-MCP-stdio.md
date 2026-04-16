---
id: TASK-19
title: '--serve-only mode: run web UI without MCP stdio'
status: Done
assignee:
  - Robert Stevens
created_date: '2026-04-13 21:10'
updated_date: '2026-04-16 07:00'
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
- [x] #1 `wiki-mcp --serve-only --port 9000` serves the wiki over HTTP without opening stdio MCP
- [x] #2 `wiki-mcp --serve` starts both stdio MCP and web UI concurrently regardless of `Web.Enabled`
- [x] #3 SIGINT/SIGTERM shut down cleanly in both modes with no leaked goroutines (verified via `goleak`)
- [x] #4 `http.Server.Shutdown` drains in-flight requests within a 10s deadline
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

### Status
AC#1, AC#2, AC#4 already implemented in main.go and internal/web/server.go:
- `--serve-only` flag skips MCP stdio (main.go:105-115)
- `--serve` flag sets Web.Enabled=true (main.go:88-89)
- Graceful shutdown with 10s drain (server.go:141-145)
- Signal handling via signal.NotifyContext (main.go:94)

### Remaining work: AC#3 goleak verification
1. Add `go.uber.org/goleak` to go.mod via `go get`
2. Add `TestMain` with `goleak.VerifyTestMain` in `internal/web/server_test.go`
3. Add `TestGracefulShutdownNoLeaks` that starts `webSrv.Run(ctx)` in a goroutine, cancels the context, waits for Run to return, then verifies no leaked goroutines

### Key files
- `internal/web/server_test.go` — add goleak TestMain + shutdown test
- `go.mod` / `go.sum` — add goleak dependency
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

AC#1, AC#2, and AC#4 were already implemented from prior work (TASK-16):
- `--serve-only` flag skips MCP stdio, runs web UI only (`main.go:105-115`)
- `--serve` flag forces `Web.Enabled=true`, runs both concurrently (`main.go:88-89`)
- `http.Server.Shutdown` with 10s drain on context cancel (`server.go:141-145`)
- Signal handling via `signal.NotifyContext` for SIGINT/SIGTERM (`main.go:94`)

### AC#3 — goroutine-leak verification via goleak

**New files / changes:**
- `go.mod` / `go.sum` — added `go.uber.org/goleak v1.3.0`
- `internal/web/server_test.go` — added `TestMain` with `goleak.VerifyTestMain` (covers all tests in the package) and `TestGracefulShutdownNoLeaks`
- `internal/web/server.go` — extended `Run` to accept `ready chan<- string`; sends the bound listening address once the port is allocated, enabling tests to confirm the server is up before triggering shutdown
- `cmd/wiki-mcp/main.go` — updated caller to pass `nil` for the ready channel

### Design decisions
- `ready chan<- string` is nil-safe; production code passes nil with no overhead.
- Test uses a no-keepalive `http.Client` (`DisableKeepAlives: true`) so transport goroutines don't linger and trigger goleak false positives.
- goleak `TestMain` covers the entire `web` test package, not just the shutdown test.
<!-- SECTION:FINAL_SUMMARY:END -->
