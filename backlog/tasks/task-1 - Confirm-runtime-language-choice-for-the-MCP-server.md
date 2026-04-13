---
id: TASK-1
title: Confirm runtime/language choice for the MCP server
status: Done
assignee: []
created_date: '2026-04-13 21:08'
updated_date: '2026-04-13 21:19'
labels: []
milestone: m-0
dependencies: []
documentation:
  - doc-5
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Lock the language + distribution mechanism before scaffolding. Doc-5 (Deployment Strategy) currently recommends Python packaged via PyPI with `uvx`/`pipx` as the primary install path, because `uvx wiki-mcp` gives zero-install isolated execution and the official MCP SDK is most mature in Python. Node + `npx` is a viable alternative; Go would mean a single binary but loses ecosystem maturity.

This decision blocks every downstream milestone, so it must be explicitly confirmed by the user (not chosen silently). Present the trade-off table from doc-5 and capture the chosen option here.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 User has explicitly confirmed the runtime (Python, Node, Go, or other)
- [x] #2 Decision + rationale recorded in a Backlog decision document
- [x] #3 Doc-5 (Deployment Strategy) updated if the chosen runtime differs from the current recommendation
- [x] #4 Any follow-on implementation tasks in M1/M2/M4/M5 updated if the choice changes required libraries
<!-- AC:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Runtime: **Go**. Decision captured in doc-7.

- Doc-5 (Deployment Strategy) rewritten: single static binary distribution, `goreleaser` + Homebrew tap + `install.sh` + Docker, Windows promoted from best-effort to required.
- Doc-7 (new) captures rationale, tradeoffs, and tentative library picks: `modelcontextprotocol/go-sdk` (fallback `mark3labs/mcp-go`), `goldmark`, `BurntSushi/toml`, `go-chi/chi`, `fsnotify`, `adrg/frontmatter`, `ledongthuc/pdf`, `goreleaser`.
- Downstream tasks rewritten for Go: task-2 (module scaffold), task-3 (hand-rolled TOML+env+flag loader), task-4 (Go CI matrix + windows required), task-5 (SDK pick + registration pattern), task-6 (Go page CRUD), task-9 (Go linkgraph package), task-10 (Go source helpers), task-15 (goldmark renderer), task-16 (net/http + chi), task-17 (fsnotify + polling fallback), task-18 (html/template + go:embed), task-19 (errgroup + goleak shutdown), task-20 (goreleaser + Homebrew + install.sh, replaces PyPI), task-21 (Dockerfile on distroless), task-22 (streamable-http with constant-time auth), task-23 (install guide updated to binary/brew/docker).

Ordering + dependencies unchanged. Task-2 unblocked.
<!-- SECTION:FINAL_SUMMARY:END -->
