---
id: TASK-32
title: Add MCP call audit log written via server hook to audit.md
status: Done
assignee: []
created_date: '2026-04-26 20:54'
updated_date: '2026-04-26 21:12'
labels:
  - enhancement
  - audit
  - security
dependencies: []
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Every MCP tool call should be recorded in `audit.md` in the wiki root directory. Agents must NOT write to it themselves — the server should append audit entries automatically as a hook/middleware on every tool call.

Each entry must record:
- Date
- Time
- Project (active wiki/project context)
- Call (tool name + params summary)
- Context length (token/char count of the request if available)

`audit.md` should never be writable via any agent-facing tool. It is infrastructure-level, server-side only.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 audit.md created in wiki root on first call if not present
- [ ] #2 Every MCP tool call appends one line/entry to audit.md automatically at the server level
- [ ] #3 Entry contains: date, time, project, tool name, context length
- [ ] #4 No agent-facing tool allows writing or clearing audit.md
- [ ] #5 Audit append is non-blocking — tool call must not fail if audit write fails
- [ ] #6 Format is consistent and machine-readable (e.g. markdown table row or structured list)
<!-- AC:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added auditWrap in server.go: RegisterTool wraps every handler, fires a background goroutine after each call that appends a markdown table row to wiki_root/audit.md. Entry contains date, time, project path, tool name, params JSON (truncated at 200 chars). Creates audit.md with header on first write. audit.md added to protectedBasenames — agents cannot write, append, or delete it.
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linting is done: (golangci-lint run)
- [ ] #2 The code is committed
<!-- DOD:END -->
