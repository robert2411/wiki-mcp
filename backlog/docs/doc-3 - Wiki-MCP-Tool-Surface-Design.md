---
id: doc-3
title: Wiki MCP - Tool Surface Design
type: other
created_date: '2026-04-13 21:02'
---
# MCP Tool Surface Design

What the server exposes to MCP clients. Three layers: **Tools** (actions), **Resources** (readable URIs), **Prompts** (workflow templates).

## Design Principle

Skill-level workflows (Ingest/Query/Lint) are *prompts*, not tools. The LLM drives them. Tools are the low-level primitives the LLM calls during those workflows. Keep the tool surface small and composable — fewer, sharper tools beat many narrow ones.

## Tools

### Page CRUD

| Tool | Purpose |
|---|---|
| `page_read` | Read a page by relative path. Returns frontmatter + body separately. |
| `page_write` | Create or overwrite a page. Takes path, body, optional frontmatter map. |
| `page_delete` | Remove a page. Emits a log entry. |
| `page_move` | Rename/move a page. Rewrites incoming links. |
| `page_list` | List pages by directory / glob / tag / updated-since. |

### Index & Log

| Tool | Purpose |
|---|---|
| `index_read` | Structured read of `index.md` (sections + entries). |
| `index_upsert_entry` | Add/update a single entry under a named section. Keeps emoji headers intact. |
| `index_refresh_stats` | Recompute page count + last-updated from disk. |
| `log_append` | Append a dated entry. Enforces `## [YYYY-MM-DD] <op> \| <title>` header shape. |
| `log_tail` | Last N entries (for context without re-reading full log). |

### Search & Traversal

| Tool | Purpose |
|---|---|
| `wiki_search` | Full-text search across pages. Returns path + snippets. |
| `links_outgoing` | Links a page points at. Handles `[[Title]]` and `[](path)`. |
| `links_incoming` | Pages that link to a given page (backlinks). |
| `orphans` | Pages with no incoming links. |

### Sources

| Tool | Purpose |
|---|---|
| `source_fetch_url` | `curl -sL` a URL into `sources/<slug>.md` and return path. |
| `source_pdf_text` | Run pdftotext on a local PDF, return extracted text. |
| `source_list` | List stored sources. |

(Source tools are thin wrappers — if a client already has native web-fetch, it can skip these. They exist so non-Claude MCP clients without web tools still work.)

### Web UI Control (optional)

| Tool | Purpose |
|---|---|
| `web_status` | Is the built-in web server running, on what port, what URL. |
| `web_start` / `web_stop` | Toggle at runtime (if not already running from `--serve`). |

## Resources

MCP resources for clients that prefer URI reads over tool calls:

- `wiki://index` → rendered `index.md`
- `wiki://log/recent` → last 20 log entries
- `wiki://page/<relative-path>` → page body
- `wiki://config` → active config (read-only)

## Prompts

Ship the Ingest / Query / Lint workflows as MCP prompts so any client can invoke them by name:

- `ingest` — args: `source` (URL or path). Produces the full ingest instructions with the tool inventory pre-listed.
- `query` — args: `question`. Same, for the query workflow.
- `lint` — no args.

This is how the server stays AI-agnostic: the skill semantics ride inside the server, not inside one host's system prompt.

## What NOT to build as tools

- No `ingest_source` megatool. That conflates LLM reasoning with file ops. The *prompt* tells the LLM to call `source_fetch_url` → read → call `page_write` / `index_upsert_entry` / `log_append`.
- No `auto_lint` that "fixes" things. Lint returns findings; fixes are the LLM's decisions calling CRUD.
- No "summarize page" tool. That's the LLM's job.
