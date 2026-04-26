---
id: TASK-31
title: Add page_append tool to append content to existing wiki pages
status: Done
assignee: []
created_date: '2026-04-26 20:54'
updated_date: '2026-04-26 21:12'
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

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added PageAppend function to wiki.go, pageAppendTool/handlePageAppend to tools.go. Appends to body preserving frontmatter, rejects non-existent pages (ErrCodeNotFound), rejects audit.md/index.md/log.md (ErrCodeForbidden). Added to protectedBasenames guard in PageWrite and PageAppend. 6 unit tests in wiki_test.go. README updated to 20 tools.
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linting is done: (golangci-lint run)
- [ ] #2 The code is committed
<!-- DOD:END -->
