---
id: TASK-31
title: Add page_append tool to append content to existing wiki pages
status: To Do
assignee: []
created_date: '2026-04-26 20:54'
labels:
  - enhancement
  - tools
dependencies: []
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Add a `page_append` tool that appends markdown content to the end of an existing wiki page without requiring a full read-then-write round-trip.

Currently agents must `page_read` → concatenate → `page_write` to append. A dedicated tool reduces token usage and race-condition risk.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 page_append tool registered in tools.go alongside page_write
- [ ] #2 Appends body content to end of existing page, preserving frontmatter
- [ ] #3 Returns error if page does not exist (no silent create)
- [ ] #4 Handler tested with unit tests
- [ ] #5 Tool description and hints set correctly (destructive: false, idempotent: false)
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linting is done: (golangci-lint run)
- [ ] #2 The code is committed
<!-- DOD:END -->
