---
id: TASK-2
title: Scaffold the wiki-mcp Go module and binary
status: Done
assignee: []
created_date: '2026-04-13 21:08'
updated_date: '2026-04-14 18:51'
labels: []
milestone: m-0
dependencies:
  - TASK-1
documentation:
  - doc-1
  - doc-5
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Initialize the Go module and lay out the project skeleton. Decision: Go (doc-7).

Layout:

```
go.mod                       # module github.com/<user>/wiki-mcp
go.sum
cmd/wiki-mcp/main.go         # entry point, argparse via `flag` or `spf13/cobra`
internal/config/             # config loader (see task-3)
internal/server/             # MCP server (see task-5)
internal/wiki/               # page/index/log/linkgraph (M2 tasks)
internal/web/                # HTTP UI (M4 tasks)
internal/sources/            # source helpers (task-10)
web/theme/default/           # embedded via //go:embed
.golangci.yml
.goreleaser.yaml             # stub, filled in task-20
Makefile                     # common dev commands
README.md                    # stub pointing at backlog docs
```

Use `go:embed` for theme assets so the single binary is truly self-contained.

CLI framework: start with stdlib `flag`; switch to `cobra` only if subcommands emerge. Version info via `-ldflags "-X main.version=..."` set by goreleaser.

Deliver: `go build ./cmd/wiki-mcp && ./wiki-mcp --help` prints usage. No MCP, tools, or web UI implemented yet.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `go.mod` declares module `github.com/<user>/wiki-mcp` with a Go version floor of 1.22+
- [x] #2 `cmd/wiki-mcp/main.go` compiles and `./wiki-mcp --help` prints usage
- [x] #3 Directory layout matches the plan: `cmd/`, `internal/{config,server,wiki,web,sources}`, `web/theme/default/`
- [x] #4 `.golangci.yml` configured; `golangci-lint run` passes on the scaffold
- [x] #5 `go test ./...` runs green with at least one package-level smoke test
- [x] #6 `.gitignore` covers build artifacts (`wiki-mcp`, `dist/`)
- [x] #7 Makefile has targets: `build`, `test`, `lint`, `run`
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

1. Init Go module (`go mod init github.com/robert2411/wiki-mcp`, Go 1.22+)
2. Create directory layout: `cmd/wiki-mcp/`, `internal/{config,server,wiki,web,sources}/`, `web/theme/default/`
3. Write `cmd/wiki-mcp/main.go` with `flag` stdlib, version via ldflags, `--help` usage
4. Add placeholder `.go` files in each internal package + embedded theme stub
5. Write smoke test (at least one `_test.go`)
6. Create `.golangci.yml`
7. Update `.gitignore` for Go build artifacts
8. Create `Makefile` with `build`, `test`, `lint`, `run` targets
9. Verify: `go build`, `go test ./...`, `golangci-lint run`
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Scaffolded Go module and project skeleton.\n\n- `go.mod` declares `github.com/robert2411/wiki-mcp` with Go 1.22 floor\n- `cmd/wiki-mcp/main.go` uses stdlib `flag`; `--help`, `--version`, `--wiki`, `--http` flags; version injected via ldflags\n- Directory layout: `cmd/wiki-mcp/`, `internal/{config,server,wiki,web,sources}/`, `web/theme/default/`\n- `.golangci.yml` configured for golangci-lint v2; passes clean\n- Smoke test in `cmd/wiki-mcp/main_test.go` builds binary and verifies `--help` output\n- `.gitignore` covers `wiki-mcp`, `dist/`, `*.exe`\n- `Makefile` with `build`, `test`, `lint`, `run` targets\n- `.goreleaser.yaml` stub for TASK-20
<!-- SECTION:FINAL_SUMMARY:END -->
