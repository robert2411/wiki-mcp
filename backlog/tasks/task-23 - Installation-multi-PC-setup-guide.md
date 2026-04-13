---
id: TASK-23
title: Installation + multi-PC setup guide
status: To Do
assignee: []
created_date: '2026-04-13 21:11'
updated_date: '2026-04-13 21:19'
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
- [ ] #1 `docs/install.md` covers Homebrew / install.sh / direct-binary / Docker / `go install` with exact commands
- [ ] #2 Multi-PC setup section explains the git-sync workflow with a worked example
- [ ] #3 launchd + systemd-user + Windows service snippets included and tested on the user's actual machine
- [ ] #4 Troubleshooting table covers: missing wiki_path, port in use, client connection failure, Windows antivirus, Docker permissions
<!-- AC:END -->
