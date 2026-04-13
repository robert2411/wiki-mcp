---
id: TASK-3
title: Implement config loader in Go (TOML + env + flags precedence)
status: To Do
assignee: []
created_date: '2026-04-13 21:08'
updated_date: '2026-04-13 21:17'
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
- [ ] #1 Config parses from TOML; missing file falls back to next source without error
- [ ] #2 Env vars override file values; CLI flags override env vars
- [ ] #3 `WikiPath` required; missing aborts with helpful message naming env var + flag + file option
- [ ] #4 `ResolveWikiPath` rejects traversal (`../etc/passwd`) when `ConfineToWikiPath=true` (table-driven test)
- [ ] #5 `MustMutate` returns a sentinel error when `ReadOnly=true`
- [ ] #6 Unit tests cover precedence, env parsing, invalid TOML, missing required fields, traversal rejection
<!-- AC:END -->
