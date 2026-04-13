---
id: doc-6
title: Wiki MCP - Read-only Web UI Design
type: other
created_date: '2026-04-13 21:03'
---
# Read-only Web UI

Replaces the current MkDocs + systemd timer + python http.server stack (documented in `existing/wiki/infrastructure/wiki-ui.md`).

## Why Rebuild, Not Reuse MkDocs

MkDocs works. But:

1. Adds a separate build process + systemd timer on every host. Breaks the "one-command install" goal.
2. Polls every 30s because container writes don't produce inotify events — that constraint disappears when the server runs on the host directly.
3. User must install `mkdocs + mkdocs-material` via `pip --break-system-packages` separately. Extra step per PC.
4. We already have the wiki dir in-process — an in-server renderer has zero cold-start lag and always reflects disk state.

## Functional Requirements

- Read-only. No edit UI. No forms. GET-only.
- Browse all wiki pages. Render markdown → HTML.
- Resolve `[[Obsidian Links]]` and relative `[text](path)` links to the correct page URL.
- Render frontmatter as a metadata block (tags, updated date).
- Full-text search box (client-side over a pre-built index; small wikis need nothing more).
- Optional sidebar nav reflecting `index.md` structure (or dir structure when index.md absent).
- Serve static assets (images embedded in pages) from wiki dir.
- Live reflect disk changes (file watcher invalidates render cache).

## Non-Goals

- Auth, login, comments, revision UI.
- Theme gallery.
- Offline PWA.
- Multi-wiki dashboard.

## Rendering Pipeline

```
wiki/page.md  ──►  parse frontmatter  ──►  markdown → HTML  ──►  link rewrite  ──►  template
                                                                                       │
                                                   watchdog invalidates cache          ▼
                                                                              HTTP response
```

Tech: `markdown-it-py` (Python) or `mistune`. Plugin for wikilinks + frontmatter. Link-rewrite resolves `[[Qwen3.5]]` → `/entities/qwen3.5.html` by lookup against the page index.

## URL Scheme

| URL | Content |
|---|---|
| `/` | Render `index.md` |
| `/<path-minus-.md>` | Render `<path>.md` (e.g. `/entities/ollama` → `entities/ollama.md`) |
| `/_log` | Render `log.md` |
| `/_search?q=...` | Client-side search result page (JS), or server-side fallback |
| `/_assets/*` | Static files (images) resolved inside wiki dir |

## Config Touchpoints

Driven entirely by the `[web]` block in config (see Config Schema doc):

```toml
[web]
enabled = true
port = 9000
bind = "127.0.0.1"
theme = "default"
auto_rebuild = true
```

## Theming

Ship one default theme. Stretch: a `[web].theme = "/path/to/dir"` override pointing at a user template dir. Don't add a theme-package mechanism — single-user wiki, single-theme default is enough.

## Migration Notes

Users on the old MkDocs setup: disable the three systemd units, point `wiki-mcp --serve` at the same wiki dir, same port (9000). No content change.
