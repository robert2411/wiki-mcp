---
id: TASK-27
title: Project README and user-facing documentation polish
status: To Do
assignee: []
created_date: '2026-04-13 21:11'
labels: []
milestone: m-5
dependencies:
  - TASK-23
  - TASK-24
  - TASK-25
  - TASK-26
documentation:
  - doc-1
  - doc-5
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Finalize the top-level README and user docs so the project is presentable on GitHub.

README sections:
- What + why (2 paragraphs). Lean on doc-1.
- Quickstart: `uvx wiki-mcp --wiki-path ~/Documents/wiki --serve` + MCP client config snippet.
- Feature list with links to deep-dive docs.
- Links to per-host config docs from M6 verification tasks.
- Screenshot of the web UI.
- Contributing section pointing at backlog.

Verify all links in the docs tree resolve. Prune dead backlinks from the design docs now that implementation has landed.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 README has quickstart that a new user can copy-paste and get a working wiki in under 5 minutes
- [ ] #2 All intra-doc links resolve (link-check runs in CI)
- [ ] #3 Screenshot of the web UI included
- [ ] #4 Feature list reflects the shipped surface (tools/resources/prompts from M2+M3)
- [ ] #5 Docs directory structured: `docs/install.md`, `docs/clients/<host>.md`, `docs/migration-from-skill.md`, `docs/config.md`
<!-- AC:END -->
