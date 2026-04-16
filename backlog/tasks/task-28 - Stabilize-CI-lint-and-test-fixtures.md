---
id: TASK-28
title: Stabilize CI lint and test fixtures
status: Done
assignee: []
created_date: '2026-04-16 09:10'
updated_date: '2026-04-16 09:10'
labels: []
milestone: m-0
dependencies:
  - TASK-4
  - TASK-6
documentation:
  - doc-5
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Fix CI failures introduced by stricter `golangci-lint` validation and by tests relying on ignored local fixture files.

Scope:
- Remove invalid/legacy lint config usage and make codebase pass current lint checks.
- Replace test dependencies on `existing/wiki/*.md` with tracked, deterministic fixtures under `internal/wiki/testdata/`.
- Prevent test panics caused by empty fixture reads in CI.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `golangci-lint run` passes with zero issues
- [x] #2 `go test ./...` passes after fixture and test hardening changes
- [x] #3 `internal/wiki` tests no longer read from ignored `existing/wiki/` paths
- [x] #4 Added tracked fixture files under `internal/wiki/testdata/` for index/log tests
- [x] #5 Lint findings fixed for `errcheck`, `gofmt`, `staticcheck`, and `unused`
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Move wiki test inputs to tracked `internal/wiki/testdata` fixtures.
2. Update `index_test.go` and `log_test.go` to load fixtures from `testdata` and use temp dirs for mutating tests.
3. Fix unchecked return values and formatting issues across the codebase.
4. Resolve staticcheck and unused symbol findings.
5. Re-run lint and full test suite to verify green state.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Stabilized CI by removing brittle local-file assumptions in tests and cleaning all lint findings.

### Changes

- **Fixtures and tests**
  - `internal/wiki/index_test.go`: added fixture helper, switched from `../../existing/wiki/index.md` to `testdata/index.md`, kept write tests isolated in `t.TempDir()`.
  - `internal/wiki/log_test.go`: switched to `testdata/log.md`, added safer assertions before indexing slices.
  - `internal/wiki/testdata/index.md`, `internal/wiki/testdata/log.md`: tracked fixture baselines for CI.

- **Lint and code-quality fixes**
  - `internal/config/config_test.go`, `internal/server/server_test.go`, `internal/web/server_test.go`, `internal/wiki/index_test.go`, `internal/wiki/log_test.go`: fixed unchecked return values.
  - `internal/sources/sources.go`, `internal/web/render/render.go`, `internal/web/watcher.go`, `internal/wiki/wiki.go`: handled deferred close/walk/encoder returns.
  - `internal/wiki/index.go`: replaced `WriteString(fmt.Sprintf(...))` with `fmt.Fprintf`, removed empty branch.
  - `internal/web/search.go`: removed unused `matchSnippet`.
  - `internal/wiki/resources.go`, `internal/wiki/graph_test.go` and other touched files: formatted with `gofmt`.

### Validation

- `golangci-lint run` → `0 issues`
- `go test ./...` → all packages pass
<!-- SECTION:FINAL_SUMMARY:END -->

