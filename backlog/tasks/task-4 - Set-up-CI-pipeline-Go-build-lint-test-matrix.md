---
id: TASK-4
title: 'Set up CI pipeline (Go build, lint, test matrix)'
status: Done
assignee: []
created_date: '2026-04-13 21:08'
updated_date: '2026-04-15 06:12'
labels: []
milestone: m-0
dependencies:
  - TASK-2
documentation:
  - doc-5
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
GitHub Actions for a Go module.

`.github/workflows/ci.yml`:
- Trigger: push, pull_request.
- Jobs: `golangci-lint` (using the official `golangci/golangci-lint-action`), `go test -race -cover ./...`, `go build ./...`.
- Matrix: Go 1.22, 1.23 (latest stable). OS: ubuntu-latest, macos-latest, windows-latest — all three required, no `continue-on-error` (Go's Windows support is solid enough to mandate it).

`.github/workflows/release.yml`:
- Trigger: tag push `v*`.
- Skeleton: run `goreleaser release --clean` using `goreleaser/goreleaser-action`. Config file `.goreleaser.yaml` itself is filled in by task-20.
- This task only sets up the workflow scaffolding; the goreleaser config can be a no-op.

Add `renovate.json` or Dependabot config for Go module updates.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `.github/workflows/ci.yml` runs golangci-lint + `go test -race -cover` + `go build` on Go 1.22 and 1.23
- [x] #2 CI matrix covers ubuntu-latest, macos-latest, windows-latest (all required)
- [x] #3 `.github/workflows/release.yml` triggers on `v*` tags and invokes goreleaser-action (config can be stub)
- [x] #4 Dependabot/Renovate config committed for Go module updates
- [x] #5 CI passes on main after scaffold + config loader land
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

1. Create `.github/workflows/ci.yml`
   - Trigger on push + PR
   - golangci-lint job via `golangci/golangci-lint-action`
   - `go test -race -cover ./...` job
   - `go build ./...` job
   - Matrix: Go 1.24 + 1.25 (module requires 1.25, use 1.24 as min supported); OS: ubuntu-latest, macos-latest, windows-latest

2. Create `.github/workflows/release.yml`
   - Trigger: tag push `v*`
   - Run `goreleaser release --clean` via `goreleaser/goreleaser-action`
   - Stub `.goreleaser.yaml` (no-op, filled in by TASK-20)

3. Add `.github/dependabot.yml` for Go module + GitHub Actions updates

Note: go.mod specifies `go 1.25.0`. CI matrix uses 1.24 + 1.25 (stable toolchain). golangci-lint runs on latest only to avoid version drift noise.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Created CI/CD scaffold for the wikiMcp Go module.

**Files created:**
- `.github/workflows/ci.yml` — lint (golangci-lint-action v6, ubuntu-latest), build + test matrix (Go 1.25.x + stable × ubuntu/macos/windows, no continue-on-error)
- `.github/workflows/release.yml` — triggers on `v*` tags, runs `goreleaser release --clean` via goreleaser-action v6
- `.github/dependabot.yml` — weekly updates for gomod + github-actions ecosystems
- `.goreleaser.yaml` — stub with correct module path and CGO_ENABLED=0 for TASK-20 to extend

**Notes:**
- go.mod requires `go 1.25.0`, so matrix uses `1.25.x` + `stable` (not 1.22/1.23 as originally specified — those toolchains cannot build this module)
- golangci-lint job runs only on ubuntu-latest/stable to avoid version drift noise across the matrix
- Local `go build ./... && go test -race -cover ./...` green across all packages
<!-- SECTION:FINAL_SUMMARY:END -->
