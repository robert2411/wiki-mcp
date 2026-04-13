---
id: doc-4
title: Wiki MCP - Config Schema
type: other
created_date: '2026-04-13 21:03'
---
# Config Schema

Config drives per-install behavior. Must be single-file, checkable into dotfiles, and overridable by env vars / CLI flags for one-off runs.

## Discovery Order

1. `--config <path>` CLI flag (explicit wins)
2. `WIKI_MCP_CONFIG` env var
3. `$XDG_CONFIG_HOME/wiki-mcp/config.toml` (falls back to `~/.config/wiki-mcp/config.toml`)
4. `./wiki-mcp.toml` in CWD
5. Built-in defaults (wiki path required; server errors out if unset)

Individual fields overridable by env: `WIKI_MCP_WIKI_PATH`, `WIKI_MCP_WEB_PORT`, etc.

## Schema (TOML)

```toml
# Required. Absolute path to the wiki root (dir containing index.md, log.md, subdirs).
wiki_path = "/home/robert/Documents/wiki"

# Optional. Where source files live. Defaults to <wiki_path>/../sources.
sources_path = "/home/robert/Documents/sources"

[web]
# Built-in read-only web server.
enabled = true            # default false. If false, server runs MCP-only.
port = 9000               # default 9000
bind = "127.0.0.1"        # default localhost-only
theme = "default"         # "default" | "minimal" | path to custom template dir
auto_rebuild = true       # default true. Re-renders on file change.

[index]
# index.md section order. Custom sections appended in first-seen order.
sections = [
  { key = "research",       title = "🔬 Research" },
  { key = "entities",       title = "🏷️ Entities" },
  { key = "concepts",       title = "💡 Concepts" },
  { key = "infrastructure", title = "🏗️ Infrastructure" },
]

[log]
# Log entry date format. Default ISO.
date_format = "%Y-%m-%d"

[links]
# How to write links from tools. Reading always accepts both.
# "obsidian" -> [[Title]]  "markdown" -> [Title](path)  "preserve" -> match existing page style
style = "preserve"

[safety]
# Hard guardrails.
read_only = false             # If true, all mutating tools refuse.
confine_to_wiki_path = true   # Reject paths outside wiki_path / sources_path.
max_page_bytes = 1048576      # 1 MiB default.
```

## Server Startup

```
wiki-mcp --config ./wiki-mcp.toml
wiki-mcp --wiki-path ~/Documents/wiki          # minimal override
wiki-mcp --wiki-path ~/Documents/wiki --serve  # MCP + web UI
wiki-mcp --serve-only --port 9000              # web UI only, no stdio MCP
```

## MCP Client Config Examples

### Claude Desktop / Claude Code (`claude_desktop_config.json` or `.mcp.json`)

```json
{
  "mcpServers": {
    "wiki": {
      "command": "uvx",
      "args": ["wiki-mcp"],
      "env": { "WIKI_MCP_WIKI_PATH": "/Users/robert/Documents/wiki" }
    }
  }
}
```

### Cursor / Cline / Continue / any other MCP client

Same shape — MCP stdio is standard. Command + env vars is the portable contract.

## Why TOML

- Comments allowed (YAML works too, but TOML's type story is clearer for the safety section).
- Claude Code's own `.mcp.json` is JSON, but that's wire format between host and server — the server's *own* config is a user file and benefits from comments.
