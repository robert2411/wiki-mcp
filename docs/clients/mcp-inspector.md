# wiki-mcp + MCP Inspector / mcp-dev-harness

> **See also:** [Claude Code](claude-code.md) | [Claude Desktop](claude-desktop.md)

This document covers connecting wiki-mcp to the official [MCP Inspector](https://github.com/modelcontextprotocol/inspector) and using a `@modelcontextprotocol/sdk` script as a scriptable dev harness. Both approaches use stdio transport and require no Claude-specific client.

## MCP Inspector (web UI)

The MCP Inspector is the official debug/explore tool for MCP servers. It connects via stdio and exposes a web UI to browse tools, resources, and prompts and make interactive tool calls.

### Install & run

```bash
WIKI_MCP_WIKI_PATH=/path/to/your/wiki \
  npx @modelcontextprotocol/inspector \
  /path/to/wiki-mcp
```

The inspector starts a proxy on `localhost:6277` (default) and opens a browser tab. The terminal prints a session token required for authentication:

```
Starting MCP inspector...
⚙️ Proxy server listening on localhost:6277
🔑 Session token: <token>
   Open http://localhost:6277?token=<token> to use the inspector
```

### What you get

- Full tool browser with schema and inline call forms
- Prompts tab (all three prompts visible: `ingest`, `lint`, `query`)
- Resources tab (returns empty array — resources not yet implemented, see TASK-11)
- Structured request/response log panel

### Quirks

- **Auth required:** The inspector generates a session token on every start. Append `?token=<token>` to the URL.
- **`DANGEROUSLY_OMIT_AUTH=true`** disables the token requirement (useful for automated testing pipelines):
  ```bash
  DANGEROUSLY_OMIT_AUTH=true WIKI_MCP_WIKI_PATH=... npx @modelcontextprotocol/inspector /path/to/wiki-mcp
  ```
- **Prompts tab may not render arg forms:** Some inspector versions do not expose prompt argument input fields. In that case use the Tools tab to call tools directly (`page_write`, `log_append`, etc.) — the prompts guide an LLM through that sequence, but the tools work independently.
- **Resources always empty:** `resources/list` returns `[]` until TASK-11 lands.

---

## Scriptable dev harness (`@modelcontextprotocol/sdk`)

For CI or reproducible transcripts, use the `StdioClientTransport` from the official SDK directly — no browser required.

### Setup

```bash
mkdir mcp-verify && cd mcp-verify
npm init -y
npm install @modelcontextprotocol/sdk
```

### Client script

```js
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";

const transport = new StdioClientTransport({
  command: "/path/to/wiki-mcp",
  args:    [],
  env:     { ...process.env, WIKI_MCP_WIKI_PATH: "/path/to/your/wiki" },
});

const client = new Client(
  { name: "mcp-dev-harness", version: "1.0.0" },
  { capabilities: {} }
);

await client.connect(transport);

try {
  const { tools } = await client.listTools();
  console.log("Tools:", tools.map(t => t.name));

  await Promise.all([
    client.callTool({ name: "page_write", arguments: { path: "test.md", body: "# Test" } }),
    client.callTool({ name: "log_append", arguments: { operation: "ingest", title: "Test", body: "Wrote test.md" } }),
  ]);
} finally {
  await client.close();
}
```

### Run

```bash
WIKI_MCP_WIKI_PATH=/path/to/your/wiki node verify.mjs /path/to/wiki-mcp
```

---

## Verified surface (2026-04-16)

Run against `existing/wiki/` using Node.js 25 + `@modelcontextprotocol/sdk` `StdioClientTransport`.

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

---

## End-to-end transcript (mcp-dev-harness)

Full output from `node verify.mjs` against `existing/wiki/`:

```
Connecting to wiki-mcp via stdio...
Connected.

=== Server version ===
{ "name": "wiki-mcp", "version": "dev" }

=== Server capabilities ===
{
  "logging": {},
  "prompts":   { "listChanged": true },
  "resources": { "listChanged": true },
  "tools":     { "listChanged": true }
}

=== Tools list ===
[
  "index_read", "index_refresh_stats", "index_upsert_entry",
  "links_incoming", "links_outgoing",
  "log_append", "log_tail",
  "orphans",
  "page_delete", "page_list", "page_move", "page_read", "page_write",
  "source_fetch_url", "source_list", "source_pdf_text",
  "wiki_search"
]

Total tools: 17

=== Prompts list ===
[ "ingest", "lint", "query" ]

=== page_write result ===
{
  "content": [{ "type": "text", "text": "page \"verification/mcp-inspector-test.md\" written successfully" }]
}

=== log_append result ===
{
  "content": [{ "type": "text", "text": "log entry appended: [2026-04-16] ingest | MCP Inspector verification page" }]
}

✓ Verification complete — all steps passed.
```

**Verified on disk:** `existing/wiki/verification/mcp-inspector-test.md` created; `log.md` updated.

---

## Quirks (non-Claude hosts in general)

- **No prompt argument UI:** Unlike Claude hosts, generic MCP clients do not render prompt argument forms. Use direct tool calls (`page_write`, `log_append`, etc.) instead of the `ingest` / `query` prompts.
- **Capabilities negotiation:** wiki-mcp advertises `logging`, `prompts.listChanged`, `resources.listChanged`, and `tools.listChanged`. Clients that ignore `listChanged` still work — wiki-mcp never pushes unsolicited notifications unless the client subscribes.
- **Concurrent request dispatch:** wiki-mcp handles requests concurrently. In scripted clients sending multiple requests in a burst, responses may arrive out of order; match by JSON-RPC `id`. The SDK client handles this automatically.
- **stderr logging:** wiki-mcp writes structured logs to stderr. The SDK `StdioClientTransport` inherits the parent process's stderr by default, so log lines appear in the terminal. Redirect with `2>/dev/null` or `2>/tmp/wiki-mcp.log` to suppress.
