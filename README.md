# wiki-mcp

Personal wiki server with MCP (Model Context Protocol) integration. Any MCP-capable AI assistant — Claude Desktop, Claude Code, Cursor, Cline, and others — can read and write your wiki through a standard tool surface. A built-in web UI lets you browse the wiki from a browser with no extra daemons.

The project replaces a per-assistant skill + MkDocs/systemd stack with a single static binary: install once per machine, point at your wiki directory, wire up your MCP client, and every AI host you use gets the same consistent tool surface.

## Quickstart

Get a working wiki in under 5 minutes:

**1. Install**

```bash
# Homebrew (macOS / Linux)
brew install robertstevens/tap/wiki-mcp

# One-line installer (macOS / Linux)
curl -sSfL https://github.com/robertstevens/wiki-mcp/releases/latest/download/install.sh | sh
```

See [docs/install.md](docs/install.md) for Docker, Windows, and `go install` options.

**2. Configure**

```bash
mkdir -p ~/.config/wiki-mcp
cat > ~/.config/wiki-mcp/config.toml <<'EOF'
wiki_path = "/Users/yourname/Documents/wiki"
EOF
```

**3. Wire up your MCP client**

Add to your MCP client config (example for Claude Code / Claude Desktop):

```json
{
  "mcpServers": {
    "wiki": {
      "command": "/path/to/wiki-mcp",
      "env": {
        "WIKI_MCP_WIKI_PATH": "/Users/yourname/Documents/wiki"
      }
    }
  }
}
```

Use `which wiki-mcp` to get the correct binary path. See [docs/install.md § MCP client setup](docs/install.md#mcp-client-setup) for per-host guides.

**4. Browse the web UI (optional)**

```bash
wiki-mcp --serve-only
# Open http://localhost:9000
```

Done. Ask your AI assistant to list the wiki tools and start using them.

---

## Features

### Tools (20)

| Group | Tools |
|-------|-------|
| Setup | `wiki_init`, `project_list` |
| Pages | `page_read`, `page_write`, `page_append`, `page_delete`, `page_move`, `page_list` |
| Index | `index_read`, `index_upsert_entry`, `index_refresh_stats` |
| Log | `log_append`, `log_tail` |
| Search & links | `wiki_search`, `links_incoming`, `links_outgoing`, `orphans` |
| Sources | `source_fetch_url`, `source_pdf_text`, `source_list` |

### Prompts (3)

| Prompt | Purpose |
|--------|---------|
| `ingest` | Fetch a URL or local document, synthesise, write to wiki |
| `query` | Research a topic across existing wiki pages |
| `lint` | Check for stale pages, broken links, and index gaps |

### Transport

- **stdio** (default) — for local MCP clients
- **Streamable HTTP / SSE** (`--transport sse`) — for remote access from a different machine

### Web UI

Built-in read-only browser. Renders markdown, sidebar navigation, full-text search, dark mode toggle. No MkDocs, no systemd.

*Screenshot: browse your wiki at `http://localhost:9000` with `wiki-mcp --serve-only`.*

---

## Documentation

| Document | Description |
|----------|-------------|
| [docs/install.md](docs/install.md) | All install methods, full config reference, multi-PC setup, background service, troubleshooting |
| [docs/config.md](docs/config.md) | Standalone config reference (TOML + env vars) |
| [docs/clients/claude-desktop.md](docs/clients/claude-desktop.md) | Claude Desktop setup and verified tool surface |
| [docs/clients/claude-code.md](docs/clients/claude-code.md) | Claude Code setup and end-to-end ingest transcript |
| [docs/clients/mcp-inspector.md](docs/clients/mcp-inspector.md) | MCP Inspector and scriptable dev harness |
| [docs/migration-from-skill.md](docs/migration-from-skill.md) | Migrate from Andy skill-based wiki + MkDocs/systemd |

---

## Contributing

Tasks and design docs live in the `backlog/` directory. The project uses [Backlog.md](https://github.com/MrLesk/Backlog.md) for task management — pick up any `To Do` task from the backlog and follow the workflow in `CLAUDE.md`.

To report bugs or request features, open a GitHub issue.
