---
id: TASK-8
title: 'Log tools (append-only log_append, log_tail)'
status: Done
assignee: []
created_date: '2026-04-13 21:08'
updated_date: '2026-04-14 22:27'
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
- [x] #1 `log_append` refuses operations outside the allowed set with a clear error
- [x] #2 Header is always `## [YYYY-MM-DD] <op> | <title>` using configured date format
- [x] #3 Existing log content is never rewritten — only appended to
- [x] #4 `log_tail` parses `existing/wiki/log.md` entries correctly (unit test against fixture)
- [x] #5 If `log.md` is missing, `log_append` creates it with the template header
- [x] #6 Respects `read_only` flag
<!-- AC:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented `log_append` and `log_tail` in `internal/wiki/log.go`. Registered both tools in `tools.go`.

- `LogAppend`: validates operation against allowlist (ingest/query/lint), stamps date via strftime→Go format conversion, appends with blank-line separator, creates log.md from embedded template if missing
- `LogTail`: parses entries by scanning for `## [YYYY-MM-DD] op | title` headers, returns last N in chronological order
- 8 tests all pass, full build green
<!-- SECTION:FINAL_SUMMARY:END -->
