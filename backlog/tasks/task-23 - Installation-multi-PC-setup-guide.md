---
id: TASK-23
title: Installation + multi-PC setup guide
status: Done
assignee: []
created_date: '2026-04-13 21:11'
updated_date: '2026-04-16 00:21'
labels: []
milestone: m-4
dependencies:
  - TASK-20
  - TASK-21
documentation:
  - doc-4
  - doc-5
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Write `docs/install.md` (referenced from README). Covers:

- Four install paths: Homebrew (recommended, macOS/Linux), `install.sh` one-liner (Linux/macOS), direct binary download (any OS), Docker. `go install` as a dev/head-of-main option.
- Config discovery order + a default `wiki-mcp.toml` example, copy-pasteable (keep in sync with doc-4).
- Multi-PC workflow: "share the wiki dir via <git|Syncthing|...>; each PC installs the wiki-mcp binary locally and points at the shared dir via `WIKI_MCP_WIKI_PATH` or config file".
- Platform service snippets:
  - macOS: launchd `~/Library/LaunchAgents/com.wiki-mcp.serve.plist` (runs `wiki-mcp --serve-only`).
  - Linux: systemd user unit (`~/.config/systemd/user/wiki-mcp.service`).
  - Windows: NSSM or Task Scheduler snippet.
- Troubleshooting: missing wiki_path, port in use, MCP client can't see tools, Windows antivirus blocking the binary, Docker volume permissions.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `docs/install.md` covers Homebrew / install.sh / direct-binary / Docker / `go install` with exact commands
- [x] #2 Multi-PC setup section explains the git-sync workflow with a worked example
- [x] #3 launchd + systemd-user + Windows service snippets included and tested on the user's actual machine
- [x] #4 Troubleshooting table covers: missing wiki_path, port in use, client connection failure, Windows antivirus, Docker permissions
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

1. Create `docs/install.md` with sections:
   - Install methods: Homebrew, install.sh one-liner, direct binary, Docker, go install
   - Config discovery order + copy-pasteable `wiki-mcp.toml`
   - MCP client config examples (Claude Desktop, Claude Code, Cursor)
   - Multi-PC setup section (git-sync worked example)
   - Platform service snippets (launchd, systemd-user, Windows NSSM/Task Scheduler)
   - Troubleshooting table

2. Update `README.md` to reference `docs/install.md`

AC checklist:
- #1 All 5 install paths with exact commands
- #2 Multi-PC git-sync workflow with worked example
- #3 launchd + systemd + Windows snippets
- #4 Troubleshooting table (5 scenarios)
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Created `docs/install.md` — complete installation and setup guide covering all five install paths, configuration, multi-PC git-sync workflow, platform service snippets, and troubleshooting.

**Files changed:**
- `docs/install.md` — new file, ~480 lines
- `README.md` — added Docker snippet, updated link to `docs/install.md`

**Simplify fixes applied:**
- Replaced MCP client setup section with cross-references to `docs/clients/` (removed duplication with existing client docs)
- Fixed `<VERSION#v>` placeholder → `${VERSION#v}` with proper shell variable assignments
- Fixed version-pinning install.sh invocation: `export WIKI_MCP_VERSION=v1.2.3` before curl pipe so env var is inherited by `sh`
- Fixed `--serve` → `--serve-only` in Docker web UI example (consistent with rest of doc)
- Changed `robert` → `yourname` in full config reference (consistent username throughout)
- Removed XML plist narration comments (`<!-- Start at login -->` etc.)
- Trimmed redundant inline comments from git-sync bash blocks
- Added `--serve` vs `--serve-only` definition at top of background service section
- Added Apple Silicon note (`/opt/homebrew/bin`) to launchd plist section
- Added PowerShell `PATH` update command to Windows binary download section
- Added GOPATH/bin `$PATH` note to `go install` section
<!-- SECTION:FINAL_SUMMARY:END -->
