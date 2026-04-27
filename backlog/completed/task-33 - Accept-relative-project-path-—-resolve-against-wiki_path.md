---
id: TASK-33
title: Accept relative project path — resolve against wiki_path
status: Done
assignee: []
created_date: '2026-04-27 05:38'
updated_date: '2026-04-27 05:41'
labels:
  - bug
  - ux
  - config
dependencies: []
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Currently `project_path` (and `--project`) must be an absolute path that includes the full `wiki_path` prefix (e.g. `/home/user/wiki/my-project`). The UX should allow passing just the subpath relative to the wiki root (e.g. `my-project` or `research/2026`), which is the natural way to think about a project.

The server already knows `wiki_path`, so it can resolve the relative path itself.

**Examples of desired input:**

```toml
wiki_path    = "/home/user/wiki"
project_path = "my-project"          # resolved to /home/user/wiki/my-project
```

```bash
wiki-mcp --wiki-path /home/user/wiki --project research/2026
```

```json
{ "WIKI_MCP_WIKI_PATH": "/wiki", "WIKI_MCP_PROJECT_PATH": "work" }
```

Absolute paths should still be accepted unchanged (for backwards compatibility and for cases where the project lives at an explicit absolute path that happens to be within the wiki root).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 If project_path is a relative path, it is resolved relative to wiki_path during validate()
- [ ] #2 If project_path is already absolute and within wiki_path, it is accepted unchanged
- [ ] #3 Relative paths that would escape wiki_path (e.g. ../outside) are rejected with a clear error
- [ ] #4 --project flag and WIKI_MCP_PROJECT_PATH env var both accept relative paths
- [ ] #5 docs/config.md and docs/clients/claude-code.md updated to show relative path examples
- [ ] #6 Existing validation tests updated; new tests cover relative resolution and escape rejection
<!-- AC:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
In validate(), when ProjectPath is not absolute, join it onto WikiPath before resolving. Absolute paths are unchanged. The existing escape check rejects ../outside naturally. 3 new tests: relative subpath, nested relative path, relative escape rejection. docs/config.md and docs/clients/claude-code.md updated to show relative path examples throughout.
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linting is done: (golangci-lint run)
- [ ] #2 The code is committed
<!-- DOD:END -->
