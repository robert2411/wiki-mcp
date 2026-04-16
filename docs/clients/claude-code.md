# wiki-mcp + Claude Code

> **See also:** [Claude Desktop](claude-desktop.md) | [MCP Inspector / dev harness](mcp-inspector.md)

## Configuration

### Project-level (`.mcp.json` in repo root)

```json
{
  "mcpServers": {
    "wiki": {
      "type": "stdio",
      "command": "/path/to/wiki-mcp",
      "args": [],
      "env": {
        "WIKI_MCP_WIKI_PATH": "/path/to/your/wiki"
      }
    }
  }
}
```

### Global config (`~/.claude.json`)

Add under the top-level `mcpServers` key:

```json
{
  "mcpServers": {
    "wiki": {
      "type": "stdio",
      "command": "/path/to/wiki-mcp",
      "args": [],
      "env": {
        "WIKI_MCP_WIKI_PATH": "/path/to/your/wiki"
      }
    }
  }
}
```

`WIKI_MCP_WIKI_PATH` is required. Alternatively pass it as a CLI flag:

```json
{
  "mcpServers": {
    "wiki": {
      "type": "stdio",
      "command": "/path/to/wiki-mcp",
      "args": ["--wiki-path", "/path/to/your/wiki"]
    }
  }
}
```

## Verified surface (2026-04-16)

**17 tools:**
`index_read`, `index_refresh_stats`, `index_upsert_entry`,
`links_incoming`, `links_outgoing`,
`log_append`, `log_tail`,
`orphans`,
`page_delete`, `page_list`, `page_move`, `page_read`, `page_write`,
`source_fetch_url`, `source_list`, `source_pdf_text`,
`wiki_search`

**3 prompts:** `ingest`, `lint`, `query`

**Resources:** none yet (TASK-11)

## End-to-end ingest transcript

The following transcript shows a full ingest workflow executed against `existing/wiki/` via the stdio transport. Requests were sent sequentially; responses include the JSON-RPC envelope.

### 1 — Initialize

```
→ {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"claude-code","version":"1.0"}}}

← {"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","capabilities":{"logging":{},"prompts":{"listChanged":true},"resources":{"listChanged":true},"tools":{"listChanged":true}},"serverInfo":{"name":"wiki-mcp","version":"dev"}}}
```

### 2 — Write page

```
→ {"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"page_write","arguments":{"path":"concepts/mcp-stdio-transport.md","body":"# MCP stdio Transport\n\n..."}}}

← {"jsonrpc":"2.0","id":2,"result":{"content":[{"type":"text","text":"page \"concepts/mcp-stdio-transport.md\" written successfully"}]}}
```

### 3 — Update index

```
→ {"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"index_upsert_entry","arguments":{"section_key":"concepts","title":"MCP stdio Transport","path":"concepts/mcp-stdio-transport.md","summary":"How the MCP stdio transport works and when to prefer it over SSE"}}}

← {"jsonrpc":"2.0","id":3,"result":{"content":[{"type":"text","text":"entry \"MCP stdio Transport\" upserted in section \"concepts\""}]}}
```

### 4 — Refresh stats

```
→ {"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"index_refresh_stats","arguments":{}}}

← {"jsonrpc":"2.0","id":4,"result":{"content":[{"type":"text","text":"index stats refreshed"}]}}
```

### 5 — Append log

```
→ {"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"log_append","arguments":{"operation":"ingest","title":"MCP stdio Transport concept page","body":"Created concepts/mcp-stdio-transport.md. Updated index stats."}}}

← {"jsonrpc":"2.0","id":5,"result":{"content":[{"type":"text","text":"log entry appended: [2026-04-15] ingest | MCP stdio Transport concept page"}]}}
```

**Verified:** `concepts/mcp-stdio-transport.md` created on disk; `log.md` updated with new entry; `index.md` stats block refreshed.

## Quirks

- **Out-of-order responses when pipelining:** The server dispatches requests concurrently. If you send multiple requests in a single pipe burst, responses may arrive out of order (matched by `id`). For the ingest workflow Claude handles this correctly; raw scripting should send requests sequentially.
- **Resources not yet implemented:** `resources/list` returns an empty array. MCP resource URIs (`wiki://index` etc.) are planned under TASK-11.
- **Prompt arg UI:** Claude Code renders prompt arguments as form fields. The `ingest` prompt's `hint` argument is optional and can be left blank.
- **stderr logging:** The server writes structured logs to stderr; Claude Code suppresses these by default. To capture debug output, wrap the invocation in a shell:
  ```json
  {
    "command": "sh",
    "args": ["-c", "/path/to/wiki-mcp 2>/tmp/wiki-mcp.log"],
    "env": { "WIKI_MCP_WIKI_PATH": "/path/to/your/wiki" }
  }
  ```
