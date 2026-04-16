# wiki-mcp + Claude Desktop

> **See also:** [Claude Code](claude-code.md) | [MCP Inspector / dev harness](mcp-inspector.md)

## Configuration

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or
`%APPDATA%\Claude\claude_desktop_config.json` (Windows).

Add a `mcpServers` key at the top level:

```json
{
  "mcpServers": {
    "wiki": {
      "command": "/path/to/wiki-mcp",
      "args": ["--wiki-path", "/path/to/your/wiki"]
    }
  }
}
```

Or use an environment variable instead of the CLI flag:

```json
{
  "mcpServers": {
    "wiki": {
      "command": "/path/to/wiki-mcp",
      "args": [],
      "env": {
        "WIKI_MCP_WIKI_PATH": "/path/to/your/wiki"
      }
    }
  }
}
```

**Restart Claude Desktop** after editing the config — it reads `claude_desktop_config.json` once at startup.

## Verified surface

The server exposes the same surface as in Claude Code (see [claude-code.md](claude-code.md)).

To verify after restart, open a new conversation and use the paperclip / MCP panel to confirm `wiki` appears in the tool list.

## End-to-end ingest workflow

Use the `ingest` prompt (available from the prompt picker in Claude Desktop):

1. Click the paperclip icon → **Use a prompt** → select `ingest`
2. Fill in `source` (URL, local PDF path, or paste raw text)
3. Optionally fill in `hint` (e.g. "focus on performance benchmarks")
4. Submit — Claude will call `page_write`, `index_upsert_entry`, `index_refresh_stats`, and `log_append` automatically

The full ingest workflow transcript is documented in [claude-code.md](claude-code.md) — the wire protocol is identical between hosts.

## Quirks

- **Restart required:** Claude Desktop does not hot-reload `claude_desktop_config.json`. Any config change requires a full restart.
- **Prompt arg UI:** Claude Desktop renders prompt arguments as a form dialog before submission. `ingest` has an optional `hint` field; `query` has an optional `file_answer` field; `lint` takes no arguments. Leave optional fields blank to omit.
- **Resources not yet implemented:** The `wiki` server will show zero resources in the UI until TASK-11 lands.
- **Binary must be in PATH or use absolute path:** Claude Desktop does not inherit the user's shell `PATH`. Always use an absolute path for `command`, or install the binary to `/usr/local/bin`.
- **Tool approval dialogs:** Claude Desktop prompts for approval on first use of each tool per session. Click "Allow for this session" to suppress further prompts for that specific tool for the duration of the session. Destructive tools (`page_write`, `page_delete`, `page_move`) will prompt again on each new session.
- **stderr suppressed:** Server logs written to stderr are not visible in the Claude Desktop UI. Run the binary manually in a terminal to capture logs: `WIKI_MCP_WIKI_PATH=/path/to/wiki wiki-mcp`.
