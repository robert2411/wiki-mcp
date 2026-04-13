---
id: TASK-2
title: Scaffold the wiki-mcp Go module and binary
status: To Do
assignee: []
created_date: '2026-04-13 21:08'
updated_date: '2026-04-13 21:17'
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
- [ ] #1 `go.mod` declares module `github.com/<user>/wiki-mcp` with a Go version floor of 1.22+
- [ ] #2 `cmd/wiki-mcp/main.go` compiles and `./wiki-mcp --help` prints usage
- [ ] #3 Directory layout matches the plan: `cmd/`, `internal/{config,server,wiki,web,sources}`, `web/theme/default/`
- [ ] #4 `.golangci.yml` configured; `golangci-lint run` passes on the scaffold
- [ ] #5 `go test ./...` runs green with at least one package-level smoke test
- [ ] #6 `.gitignore` covers build artifacts (`wiki-mcp`, `dist/`)
- [ ] #7 Makefile has targets: `build`, `test`, `lint`, `run`
<!-- AC:END -->
