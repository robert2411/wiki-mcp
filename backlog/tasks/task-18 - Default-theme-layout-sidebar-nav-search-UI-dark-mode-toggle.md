---
id: TASK-18
title: 'Default theme: layout, sidebar nav, search UI, dark-mode toggle'
status: To Do
assignee: []
created_date: '2026-04-13 21:10'
updated_date: '2026-04-13 21:18'
labels: []
milestone: m-3
dependencies:
  - TASK-16
documentation:
  - doc-6
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Ship one bundled theme under `web/theme/default/`. Stdlib `html/template` + vanilla CSS/JS; NO build step. Embedded via `//go:embed` so the single binary carries it.

Aim <50 KB total CSS+JS so any PC serves it instantly.

Layout:
- Top bar: wiki title (from `Web.Title` config or `index.md` H1), search box, dark-mode toggle.
- Left sidebar: auto-generated nav reflecting `index.md` section structure; fallback to directory tree if absent.
- Main area: rendered page body with frontmatter metadata block at top.
- Right rail (stretch): outgoing links + backlinks for the current page from the task-9 link-graph package.

Client-side search: tiny debounced fetch-and-filter script against `/_search_index.json`. No Lunr/Fuse.js — native `String.prototype.includes` + fuzzy scoring is enough.

Dark-mode: CSS-only via `prefers-color-scheme`, localStorage override toggle.

Template inheritance: `base.html` → `page.html` / `search.html` / `log.html`.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Default theme renders consistently across Safari/Chrome/Firefox without JS errors
- [ ] #2 Sidebar nav reflects `index.md` sections in order; falls back to dir tree when index absent
- [ ] #3 Client-side search works offline (no network after page load)
- [ ] #4 Dark-mode toggle persists via localStorage and respects OS preference on first visit
- [ ] #5 Total theme assets (CSS+JS, excluding per-wiki search index) stay under 50 KB
- [ ] #6 Templates embedded via `//go:embed` — binary runs with no external theme files
<!-- AC:END -->
