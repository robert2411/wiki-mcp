# Configuration Reference

wiki-mcp is configured by layering sources — each later source overrides the earlier ones:

1. Built-in defaults
2. `./wiki-mcp.toml` in the working directory
3. `$XDG_CONFIG_HOME/wiki-mcp/config.toml` (falls back to `~/.config/wiki-mcp/config.toml`)
4. `WIKI_MCP_CONFIG` environment variable (path to a TOML file) + scalar env overrides
5. `--config <path>` CLI flag (path to a TOML file) + remaining CLI flags

**`wiki_path` has no default and must be set, or the server exits with an error.**

## Minimal config

```toml
wiki_path = "/home/yourname/Documents/wiki"
```

## Full config reference

```toml
# Required. Absolute path to the wiki root directory.
wiki_path = "/home/yourname/Documents/wiki"

# Optional. Where source files live. Defaults to <wiki_path>/../sources.
sources_path = "/home/yourname/Documents/sources"

[web]
# Built-in read-only web server.
enabled = true            # default false
port = 9000               # default 9000
bind = "127.0.0.1"        # default localhost-only
theme = "default"         # "default" | "minimal" | path to custom template dir
auto_rebuild = true       # default true — re-renders on file change

[mcp]
# HTTP transport for remote MCP access (streamable-HTTP, MCP 2025-03 spec).
# Only active when running with --transport sse.
port = 8765               # default 8765
bind = "127.0.0.1"        # default localhost-only
auth_token = ""           # bearer token; if set, all MCP HTTP requests must include it

[index]
# Section order in index.md. Custom sections appended in first-seen order.
sections = [
  { key = "research",       title = "🔬 Research" },
  { key = "entities",       title = "🏷️ Entities" },
  { key = "concepts",       title = "💡 Concepts" },
  { key = "infrastructure", title = "🏗️ Infrastructure" },
]

[log]
# Log entry date format. Default ISO 8601.
date_format = "%Y-%m-%d"

[links]
# How tools write links. Reading always accepts both styles.
# "obsidian" -> [[Title]]  |  "markdown" -> [Title](path)  |  "preserve" -> match existing page style
style = "preserve"

[safety]
read_only = false             # If true, all mutating tools refuse.
confine_to_wiki_path = true   # Reject paths outside wiki_path / sources_path.
max_page_bytes = 1048576      # 1 MiB default.
```

## Environment variables

| Env var                    | Config key                 | Description                                      |
|----------------------------|----------------------------|--------------------------------------------------|
| `WIKI_MCP_CONFIG`          | *(path override)*          | Path to a TOML config file                       |
| `WIKI_MCP_WIKI_PATH`       | `wiki_path`                | Path to the wiki root directory                  |
| `WIKI_MCP_SOURCES_PATH`    | `sources_path`             | Path to source files directory                   |
| `WIKI_MCP_WEB_ENABLED`     | `web.enabled`              | Enable the web UI (`true` or `1`)                |
| `WIKI_MCP_WEB_PORT`        | `web.port`                 | Web UI port (default 9000)                       |
| `WIKI_MCP_WEB_BIND`        | `web.bind`                 | Web UI bind address (default `127.0.0.1`)        |
| `WIKI_MCP_WEB_THEME`       | `web.theme`                | Web UI theme                                     |
| `WIKI_MCP_MCP_PORT`        | `mcp.port`                 | MCP HTTP transport port (default 8765)           |
| `WIKI_MCP_MCP_BIND`        | `mcp.bind`                 | MCP HTTP transport bind address                  |
| `WIKI_MCP_MCP_AUTH_TOKEN`  | `mcp.auth_token`           | Bearer token for MCP HTTP transport              |
| `WIKI_MCP_LINKS_STYLE`     | `links.style`              | Link write style (`preserve`/`obsidian`/`markdown`) |
| `WIKI_MCP_SAFETY_READ_ONLY`| `safety.read_only`         | Refuse all mutating tools (`true` or `1`)        |
| `WIKI_MCP_SAFETY_CONFINE`  | `safety.confine_to_wiki_path` | Reject paths outside wiki root (`true` or `1`) |

## CLI flags

| Flag              | Description                                                          |
|-------------------|----------------------------------------------------------------------|
| `--wiki-path`     | Path to wiki root (same as `WIKI_MCP_WIKI_PATH`)                    |
| `--config`        | Path to config file                                                  |
| `--port`          | Web UI port                                                          |
| `--bind`          | Bind address for both the web UI and MCP HTTP transport              |
| `--mcp-port`      | MCP HTTP transport port (used with `--transport sse`)                |
| `--auth-token`    | Bearer token required for MCP HTTP transport requests                |
| `--transport`     | Transport mode: `stdio` (default) or `sse`                           |
| `--serve`         | Enable the web UI alongside the default MCP transport                |
| `--serve-only`    | Run web UI only, no MCP transport                                    |
| `--version`       | Print version and exit                                               |

## See also

- [docs/install.md](install.md) — installation, multi-PC setup, background service, troubleshooting
- [docs/install.md § MCP client setup](install.md#mcp-client-setup) — per-host MCP client configuration
