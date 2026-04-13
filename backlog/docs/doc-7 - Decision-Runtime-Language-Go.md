---
id: doc-7
title: 'Decision - Runtime Language: Go'
type: other
created_date: '2026-04-13 21:16'
---
# Decision: Go runtime for wiki-mcp

Date: 2026-04-13
Status: Accepted
Related: task-1, doc-5

## Decision

Implement wiki-mcp in **Go**. Distribute as static binaries (one per OS/arch) plus a Docker image. No Python/Node runtime required on the host.

## Rationale

1. **Single-binary distribution is the best deploy UX for multi-PC.** `curl -L <release-url>/wiki-mcp-<os>-<arch> -o ~/bin/wiki-mcp && chmod +x` works on any box, no package manager, no runtime pre-install.
2. **Cross-compile is trivial** (`GOOS=linux GOARCH=amd64`, etc.) — one CI job produces every artifact.
3. **Static link → no library drift** across PCs. A binary built today runs on a box upgraded two years from now.
4. **Small container image.** Scratch or distroless base + ~10 MB binary instead of ~150 MB Python image.
5. **Performance of the web UI renderer + file watcher is better out of the box** with goroutines than with Python asyncio.

## Tradeoffs accepted

- **MCP SDK less mature than Python.** Two options: the official `github.com/modelcontextprotocol/go-sdk` (co-announced by Anthropic + Google, Aug 2025), or community `github.com/mark3labs/mcp-go`. Picking the official SDK as the default, reverting to mark3labs if blocker surfaces. SDK selection is a micro-decision inside task-5.
- **Markdown ecosystem smaller than Python's.** `goldmark` is the standard choice (CommonMark + GFM + extensions including wikilinks via plugins). Frontmatter via `go-yaml/yaml` + manual split, or `adrg/frontmatter`.
- **No `uvx`-style zero-install runner.** Mitigated by `go install github.com/<user>/wiki-mcp@latest` (for Go users) and a one-line curl script + Homebrew tap (for everyone else).
- **Windows best-effort still applies**, but Go's Windows support is strong so the story actually improves vs Python.

## Library choices (tentative, confirmed during implementation)

| Concern | Pick | Alternative |
|---|---|---|
| MCP SDK | `github.com/modelcontextprotocol/go-sdk` | `github.com/mark3labs/mcp-go` |
| Config | `github.com/BurntSushi/toml` + env/flag wiring (hand-rolled) | `spf13/viper` if it gets complex |
| HTTP server | `net/http` + `go-chi/chi` router | stdlib only |
| Markdown | `github.com/yuin/goldmark` + extensions | `russross/blackfriday` |
| Frontmatter | `github.com/adrg/frontmatter` | hand split + `yaml.v3` |
| File watcher | `github.com/fsnotify/fsnotify` | polling fallback |
| PDF text | `github.com/ledongthuc/pdf` | shell out to `pdftotext` if present |
| HTML → markdown | `github.com/JohannesKaufmann/html-to-markdown` | `jaytaylor/html2text` |
| Structured logging | `log/slog` (stdlib) | - |
| Release tooling | `goreleaser` | manual cross-compile |

## Distribution

- GitHub Releases: static binaries for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64.
- Docker image on GHCR (scratch or distroless base, multi-arch).
- Homebrew tap: `brew install <tap>/wiki-mcp`.
- `go install github.com/<user>/wiki-mcp/cmd/wiki-mcp@latest` for Go devs.

## Impact on backlog

- Doc-5 rewritten to reflect Go-first distribution.
- Task-2 (scaffold), task-3 (config), task-4 (CI), task-5 (MCP server), task-15/16/17/18 (web UI), task-20 (release), task-21 (Docker) all updated to Go-specific tooling.
- Task-10 (source helpers) and task-22 (SSE) get library swaps.
- Ordering and dependencies unchanged.
