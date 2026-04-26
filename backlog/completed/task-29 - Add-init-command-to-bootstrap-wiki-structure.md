---
id: TASK-29
title: Add init command to bootstrap wiki structure
status: Done
assignee: []
created_date: '2026-04-26 20:30'
updated_date: '2026-04-26 21:04'
labels:
  - feature
  - cli
dependencies: []
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Add `wiki init` command that scaffolds a new wiki in the configured wiki folder. Users starting fresh have no structure to work with — this command creates the baseline so other commands (add, search, list) work immediately.

The command should create the wiki directory if it doesn't exist, lay down a standard folder structure, and ensure an index file is present so navigation and linking work out of the box.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Running `wiki init` creates the wiki directory if it does not already exist
- [ ] #2 Command creates a basic folder structure (e.g. top-level sections or categories)
- [ ] #3 An index file is created (e.g. index.md or similar) that serves as the wiki entry point
- [ ] #4 Running `wiki init` on an already-initialised wiki does not overwrite existing content
- [ ] #5 Command exits with a clear success message listing what was created
- [ ] #6 Command exits with a clear error message if the wiki directory cannot be created (e.g. permission denied)
- [ ] #7 Init behaviour is covered by unit/integration tests
- [ ] #8 README or docs updated to describe the init command and expected output structure
<!-- AC:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented in previous session alongside TASK-30. wiki_init MCP tool, WikiInit function, init_test.go — all shipped in commit 26992df.
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linting is done: (golangci-lint run)
- [ ] #2 The code is committed
<!-- DOD:END -->
