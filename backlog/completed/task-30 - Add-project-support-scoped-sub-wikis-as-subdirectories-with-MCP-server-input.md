---
id: TASK-30
title: 'Add project support: scoped sub-wikis as subdirectories with MCP server input'
status: Done
assignee: []
created_date: '2026-04-26 20:35'
updated_date: '2026-04-26 20:59'
labels:
  - feature
  - cli
  - architecture
dependencies: []
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Add a first-class concept of "projects" to the wiki. A project is a scoped sub-wiki that lives as a subdirectory inside the main wiki folder (e.g. `wiki/my-project/`). Projects can nest — a project subdirectory can itself contain further project subdirectories.

The MCP server should accept a project path as input so that tools (page_read, page_write, index_read, etc.) operate within the scope of that project rather than the full wiki root. This lets a single wiki-mcp instance serve multiple focused contexts.

Key design constraints the implementer needs to know:
- `cfg.WikiPath` is the root. A project path is a relative subdirectory within that root, e.g. `my-project` or `my-project/sub-project`.
- `cfg.ResolveWikiPath` already enforces confinement to `WikiPath` — project paths must go through the same guard.
- All existing tools (page_read, page_write, page_list, index_*, log_*, wiki_search, links_*) must continue to work when scoped to a project; the scoping should be transparent to callers.
- `wiki_init` (TASK-29) should be aware of projects: running it with a project path should bootstrap that project subdirectory.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 MCP server accepts an optional project path parameter (e.g. `--project` flag or config field)
- [ ] #2 When a project path is set, all wiki tools operate relative to that project subdirectory
- [ ] #3 Project paths are validated: must be relative, must not escape wiki root (same confinement rules as pages)
- [ ] #4 Projects can nest arbitrarily deep (e.g. `research/2026/q1` is a valid project path)
- [ ] #5 Running `wiki_init` with a project path bootstraps that subdirectory with its own index.md and log.md
- [ ] #6 Listing pages within a project only returns pages under that project subtree
- [ ] #7 A new `project_list` tool (or equivalent) returns all known project subdirectories (directories containing an index.md)
- [ ] #8 No regression in existing tool behaviour when no project path is set
- [ ] #9 Tests cover: project-scoped reads/writes, path confinement, nested projects, and project_list
- [ ] #10 README and docs updated to describe project support
<!-- AC:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added full project support (scoped sub-wikis) and project overview in the web UI.

**Config changes** (`internal/config/config.go`):
- Added `ProjectPath string` to `Config` struct (TOML: `project_path`, env: `WIKI_MCP_PROJECT_PATH`)
- Added `ProjectPath *string` to `Flags` struct
- Added `Root() string` method — returns `ProjectPath` if set, else `WikiPath`
- Updated `ResolveWikiPath` to confine relative to `Root()` instead of `WikiPath`
- Added `ProjectPath` validation in `validate()`: resolves to absolute, must be within `WikiPath`
- Added `--project` CLI flag in `cmd/wiki-mcp/main.go`

**Wiki tool walk-root fixes** (6 locations):
- `wiki.go`: `PageList`, `PageMove` — `cfg.WikiPath` → `cfg.Root()`
- `graph.go`: `LinksIncoming`, `Orphans` — `cfg.WikiPath` → `cfg.Root()`
- `search.go`: `WikiSearch` — `cfg.WikiPath` → `cfg.Root()`
- `index.go`: `IndexRefreshStats` — `cfg.WikiPath` → `cfg.Root()`
- `init.go`: `WikiInit` — all 4 usages → `cfg.Root()`

**New `project_list` tool** (`internal/wiki/init.go` + `tools.go`):
- `ProjectList(cfg)` scans `cfg.WikiPath` (not `Root()`) for subdirs containing `index.md`
- Returns `[]ProjectInfo{Name, Path}`
- Always scans wiki root regardless of active project scope

**Resources** (`internal/wiki/resources.go`):
- Added `ProjectPath` to `safeConfig` (omitempty)

**Web UI** (`internal/web/server.go` + `web/theme/default/page.html`):
- Added `Projects []wiki.ProjectInfo` to `pageData`
- Added `handleRoot` / `serveRootPage` for `/` — identical to `servePage` but also calls `ProjectList`
- Template shows a "Projects" `<section>` below content when projects exist

**Tests added**:
- `config_test.go`: `TestRoot_NoProject`, `TestRoot_WithProject`, `TestProjectPath_MustBeWithinWikiPath`, `TestProjectPath_ValidSubdir`
- `init_test.go`: `TestProjectList_Empty`, `TestProjectList_SubdirsWithIndex`, `TestWikiInit_ScopedToProject`

All tests pass.
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linting is done: (golangci-lint run)
- [ ] #2 The code is committed
<!-- DOD:END -->
