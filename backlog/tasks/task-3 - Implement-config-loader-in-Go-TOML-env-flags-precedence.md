---
id: TASK-3
title: Implement config loader in Go (TOML + env + flags precedence)
status: Done
assignee: []
created_date: '2026-04-13 21:08'
updated_date: '2026-04-14 18:56'
labels: []
milestone: m-0
dependencies:
  - TASK-2
documentation:
  - doc-4
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Configuration subsystem per doc-4. Implementation in Go.

Approach: hand-rolled resolver using `github.com/BurntSushi/toml` for parsing, stdlib `flag` for CLI, `os.LookupEnv` for env vars. No `viper` — too much magic for a fixed schema.

Precedence (later wins):

1. Built-in defaults (struct literal in `internal/config`).
2. `./wiki-mcp.toml` in CWD.
3. `$XDG_CONFIG_HOME/wiki-mcp/config.toml` (fallback `$HOME/.config/wiki-mcp/config.toml`).
4. `WIKI_MCP_CONFIG` env pointing at a file, plus scalar env overrides (`WIKI_MCP_WIKI_PATH`, `WIKI_MCP_WEB_PORT`, etc.).
5. CLI flags (`--config`, `--wiki-path`, `--port`, `--bind`, ...).

Return a validated `Config` struct. `WikiPath` required; missing → startup aborts with a clear error pointing at the env var / flag / file option.

Expose safety helpers on the Config:
- `func (c *Config) ResolveWikiPath(rel string) (string, error)` — joins rel to wiki root and returns a cleaned absolute path, returning error if it escapes when `Safety.ConfineToWikiPath` is true.
- `func (c *Config) MustMutate() error` — returns a sentinel error when `Safety.ReadOnly` is set; every mutating tool calls this first.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Config parses from TOML; missing file falls back to next source without error
- [x] #2 Env vars override file values; CLI flags override env vars
- [x] #3 `WikiPath` required; missing aborts with helpful message naming env var + flag + file option
- [x] #4 `ResolveWikiPath` rejects traversal (`../etc/passwd`) when `ConfineToWikiPath=true` (table-driven test)
- [x] #5 `MustMutate` returns a sentinel error when `ReadOnly=true`
- [x] #6 Unit tests cover precedence, env parsing, invalid TOML, missing required fields, traversal rejection
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

1. Add `github.com/BurntSushi/toml` dependency
2. Define `Config` struct with all fields from doc-4 schema
3. Implement `Load()` function with 5-layer precedence:
   - Struct literal defaults
   - CWD `wiki-mcp.toml`
   - XDG config `$XDG_CONFIG_HOME/wiki-mcp/config.toml`
   - Env vars (`WIKI_MCP_CONFIG` file + scalar overrides)
   - CLI flags (`--config`, `--wiki-path`, `--port`, `--bind`)
4. Implement `ResolveWikiPath(rel)` with path traversal protection
5. Implement `MustMutate()` sentinel error
6. Validate WikiPath required, abort with helpful message
7. Write comprehensive unit tests (precedence, env, invalid TOML, missing fields, traversal)
8. Update `cmd/wiki-mcp/main.go` to use new config loader
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented full config subsystem in `internal/config/config.go`:\n\n- **5-layer precedence**: defaults → CWD TOML → XDG TOML → env vars (WIKI_MCP_CONFIG file + scalar overrides) → CLI flags\n- **Flags struct**: accepts only explicitly-set flags via `flag.Visit` pattern in main.go\n- **ResolveWikiPath**: path traversal protection with proper separator-aware prefix check\n- **MustMutate**: returns `ErrReadOnly` sentinel when read-only mode active\n- **Validation**: WikiPath required with helpful error naming all three config methods\n- **WikiPath normalized**: resolved to absolute path at load time\n- **SourcesPath default**: `<wiki_path>/../sources` when not explicitly set\n- **13 unit tests**: cover precedence, env parsing, invalid TOML, missing required fields, traversal rejection (table-driven), XDG discovery, scalar env overrides\n- **main.go updated**: uses `flag.Visit` to detect explicitly-set flags, passes `config.Flags` to `config.Load`
<!-- SECTION:FINAL_SUMMARY:END -->
