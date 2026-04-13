---
id: TASK-22
title: SSE / streamable-http transport for remote MCP access
status: To Do
assignee: []
created_date: '2026-04-13 21:10'
updated_date: '2026-04-13 21:19'
labels: []
milestone: m-4
dependencies:
  - TASK-5
documentation:
  - doc-5
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Implement the `--transport sse` placeholder from task-5. Enables remote AI hosts to connect.

Which transport: use the MCP SDK's current recommendation at implementation time — streamable HTTP (MCP 2025-06 spec) is the forward path; plain SSE (legacy) may still be supported by specific clients. Implement whichever the chosen SDK exposes stably; feature-flag if both are needed.

CLI: `--transport sse`, `--mcp-port` (separate from web UI port), `--bind` (default `127.0.0.1`). Binding to `0.0.0.0` is allowed but logs a warning: user accepts responsibility for firewall.

Optional `--auth-token <token>` — server rejects requests without matching `Authorization: Bearer <token>` header. Constant-time comparison (`crypto/subtle.ConstantTimeCompare`).

No other auth. Threat model: trusted home LAN.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `wiki-mcp --transport sse --mcp-port 8765` accepts a local MCP client connection
- [ ] #2 Default bind is `127.0.0.1`; `--bind 0.0.0.0` is explicit and logged with a warning
- [ ] #3 Works against the MCP SDK's reference client
- [ ] #4 `--auth-token` check uses constant-time compare and rejects missing/wrong tokens with 401
- [ ] #5 Documentation notes threat model: trusted LAN only, use a reverse proxy + TLS for anything beyond
<!-- AC:END -->
