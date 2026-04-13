---
id: doc-5
title: Wiki MCP - Deployment Strategy
type: other
created_date: '2026-04-13 21:03'
updated_date: '2026-04-13 21:17'
---
# Deployment Strategy

Primary requirement: **easy deployable on multiple PCs, across AI hosts.**

Runtime decision: **Go**. See doc-7 for rationale and tradeoffs.

## Target Install Experience

```bash
# On any new PC, any OS — pick one:

# Homebrew (macOS/Linux)
brew install <user>/tap/wiki-mcp

# One-line install (Linux/macOS)
curl -sSfL https://github.com/<user>/wiki-mcp/releases/latest/download/install.sh | sh

# Direct binary (any OS)
curl -L <release-url>/wiki-mcp-$(uname -s)-$(uname -m) -o ~/bin/wiki-mcp && chmod +x ~/bin/wiki-mcp

# Docker
docker pull ghcr.io/<user>/wiki-mcp:latest

# Go users
go install github.com/<user>/wiki-mcp/cmd/wiki-mcp@latest
```

No Python/Node runtime required. No compile step for end users. No systemd glue.

## Runtime: Go

Static single binary per OS/arch. MCP SDK: `github.com/modelcontextprotocol/go-sdk` (official) as default, fallback to `github.com/mark3labs/mcp-go` if blocker.

| Criterion | Go |
|---|---|
| MCP SDK maturity | Official SDK present (Anthropic + Google, Aug 2025); community `mark3labs/mcp-go` also viable |
| Install UX | Static binary — best-in-class for multi-PC |
| Cross-compile | Trivial (`GOOS`/`GOARCH`) |
| Container image size | ~10 MB on scratch/distroless |
| Markdown tooling | `goldmark` (CommonMark + GFM) |
| Web server | `net/http` + `go-chi/chi` |
| File watcher | `fsnotify` |

See doc-7 for the full library pick table and runner-up picks per concern.

## Packaging Deliverables

1. **GitHub Releases** — static binaries for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64. Built via `goreleaser`.
2. **Homebrew tap** — `brew install <user>/tap/wiki-mcp`.
3. **Docker image** `ghcr.io/<user>/wiki-mcp:latest` — scratch or distroless base, multi-arch (amd64 + arm64).
4. **`install.sh`** — one-liner script that detects OS/arch and downloads the right binary to `$HOME/.local/bin`.
5. **`go install`** target for Go devs who want head-of-main.

## Cross-PC Sync

Not solved by this server. User chooses:
- Git remote on the wiki directory (recommended).
- Syncthing / Dropbox / iCloud folder.
- `rsync` in a cron.

Server treats the wiki dir as the source of truth. If two servers write to the same dir concurrently through different sync tools, conflicts are the sync tool's problem.

## Transports

- **stdio** (default) — for local MCP clients (Claude Desktop, Claude Code, Cursor, Cline, etc).
- **SSE / streamable-http** (optional flag) — for remote access from an AI host running on a different machine. Binds to `127.0.0.1` by default.

## Web UI

Built into the server process. No external MkDocs, no systemd, no extra daemons.

- Option A: server process hosts both MCP (stdio) and HTTP (port) concurrently via goroutines.
- Option B: `--serve-only` mode runs just the web UI — useful for running as a background service without an attached MCP client.

Render markdown → HTML on request, cache in memory, invalidate on file change via `fsnotify`.

## Cross-Platform Constraints

- Linux: primary target. amd64 + arm64 binaries.
- macOS: dev machine. darwin/amd64 + darwin/arm64.
- Windows: amd64. Go's Windows support is solid — full parity expected, not "best-effort" as originally scoped.

## MCP Client Config Examples

### Claude Desktop / Claude Code

```json
{
  "mcpServers": {
    "wiki": {
      "command": "/usr/local/bin/wiki-mcp",
      "env": { "WIKI_MCP_WIKI_PATH": "/Users/robert/Documents/wiki" }
    }
  }
}
```

Same shape for Cursor / Cline / Continue / any MCP client. The binary path + env vars is the portable contract.

## Non-goals for Deployment

- No managed cloud hosting.
- No auto-update daemon.
- No installer GUIs.
