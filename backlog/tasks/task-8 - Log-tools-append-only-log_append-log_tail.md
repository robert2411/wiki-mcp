---
id: TASK-8
title: 'Log tools (append-only log_append, log_tail)'
status: To Do
assignee: []
created_date: '2026-04-13 21:08'
labels: []
milestone: m-1
dependencies:
  - TASK-5
documentation:
  - doc-2
  - doc-3
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Implement `log_append` and `log_tail` per doc-3 against `log.md`. Log is append-only (doc-2). Entry header format is strict: `## [YYYY-MM-DD] <op> | <title>` where `op` ∈ `ingest | query | lint`.

`log_append` takes `(operation, title, body)`. The tool stamps the date itself from the system clock (using `[log].date_format` config). Enforces the allowed operations set. Appends with one blank line separator before the new header.

`log_tail` returns the last N entries parsed as `{date, operation, title, body}` objects. Default N=10.

Never rewrite or truncate the log. If the file doesn't exist, create it with the template header from `existing/wiki/log.md` (title + format line + divider).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `log_append` refuses operations outside the allowed set with a clear error
- [ ] #2 Header is always `## [YYYY-MM-DD] <op> | <title>` using configured date format
- [ ] #3 Existing log content is never rewritten — only appended to
- [ ] #4 `log_tail` parses `existing/wiki/log.md` entries correctly (unit test against fixture)
- [ ] #5 If `log.md` is missing, `log_append` creates it with the template header
- [ ] #6 Respects `read_only` flag
<!-- AC:END -->
