---
id: TASK-4
title: 'Set up CI pipeline (Go build, lint, test matrix)'
status: To Do
assignee: []
created_date: '2026-04-13 21:08'
updated_date: '2026-04-13 21:17'
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
- [ ] #1 `.github/workflows/ci.yml` runs golangci-lint + `go test -race -cover` + `go build` on Go 1.22 and 1.23
- [ ] #2 CI matrix covers ubuntu-latest, macos-latest, windows-latest (all required)
- [ ] #3 `.github/workflows/release.yml` triggers on `v*` tags and invokes goreleaser-action (config can be stub)
- [ ] #4 Dependabot/Renovate config committed for Go module updates
- [ ] #5 CI passes on main after scaffold + config loader land
<!-- AC:END -->
