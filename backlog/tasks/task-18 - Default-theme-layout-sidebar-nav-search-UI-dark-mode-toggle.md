---
id: TASK-18
title: 'Default theme: layout, sidebar nav, search UI, dark-mode toggle'
status: Done
assignee: []
created_date: '2026-04-13 21:10'
updated_date: '2026-04-15 19:57'
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
- [x] #1 Default theme renders consistently across Safari/Chrome/Firefox without JS errors
- [x] #2 Sidebar nav reflects `index.md` sections in order; falls back to dir tree when index absent
- [x] #3 Client-side search works offline (no network after page load)
- [x] #4 Dark-mode toggle persists via localStorage and respects OS preference on first visit
- [x] #5 Total theme assets (CSS+JS, excluding per-wiki search index) stay under 50 KB
- [x] #6 Templates embedded via `//go:embed` — binary runs with no external theme files
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

### Gaps vs. existing TASK-16 work
- AC#2: nav uses flat search-index order; need index.md section parsing
- AC#3: search is server-side only; need client-side debounced fetch+filter
- AC#4: CSS has prefers-color-scheme but no localStorage toggle
- AC#1: no JS currently, so no JS errors to fix — just need correct JS added
- AC#5/6: already satisfied by existing embed setup

### Files to change
1. `internal/web/server.go` — add `navSection` struct, `navFromIndex()` parser, cache field, update data structs
2. `web/theme/default/style.css` — replace @media dark with [data-theme=dark], add toggle button styles
3. `web/theme/default/page.html` — section-aware nav, dark toggle button, inline dark-init script, search results div, theme.js link
4. `web/theme/default/search.html` — same header/nav updates
5. `web/theme/default/theme.js` (new) — dark toggle fn + client-side search (debounce, fetch, filter)

### Nav parsing logic (AC#2)
- Read `{wikiPath}/index.md`; iterate lines
- `### Heading` → new section
- `- [Title](path.md)` → nav link under current section (strip .md)
- If no index.md or no links → single unnamed section from flat search index
- Cache alongside indexCache; invalidated by InvalidateCache()

### Dark mode (AC#4)
- CSS: `[data-theme=dark]` block (replaces @media)
- Inline `<script>` in `<head>`: read localStorage, set `data-theme` on `<html>` before paint
- theme.js: `toggleDark()` fn flips attribute + writes localStorage

### Client-side search (AC#3)
- `theme.js` fetches `/_search_index.json` on page load, caches in variable
- Input event on search box → 200ms debounce → filter with includes()
- Title match scores 2, snippet match scores 1
- Results rendered in `#search-results` div below search input
- Form still submits to server on Enter (no-JS fallback kept)
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Implemented the full default theme per AC requirements. All 6 ACs verified.

### What changed

**`internal/web/server.go`**
- Added `navSection` struct grouping nav links under a heading
- Replaced flat `navLinks() []navEntry` with `navSections() []navSection` (cached)
- Added `navFromIndex(cfg, fallback)` — delegates to `wiki.ParseIndex` (canonical parser, eliminates duplicate hand-rolled parser and dangling-pointer bug)
- Added `cachedWikiTitle()` + `wikiTitle()` — extracts H1 from index.md via `wiki.ParseFrontmatter` + existing `h1Title()` (no duplicate logic)
- Added `titleCache string` and `navCache []navSection` fields, cleared by `InvalidateCache()`
- Added `WikiTitle` to both `pageData` and `searchData`

**`web/theme/default/page.html` / `search.html`**
- Show wiki title (from index.md H1) in `<title>` tag and sidebar header instead of hardcoded "wiki"
- Inline `<script>` in `<head>` sets `data-theme` attribute before paint (prevents flash of wrong theme)
- Dark-mode toggle button (`◑`) in sidebar header
- Search input has `id="search-input"`; `<div id="search-results">` dropdown below it
- `<script src="/_theme/theme.js">` at end of body
- Section-grouped nav: `{{range .Nav}}` iterates `navSection`, renders heading + links

**`web/theme/default/style.css`**
- Replaced `@media (prefers-color-scheme: dark)` with `[data-theme=dark]` so JS can control the theme
- Added `.sidebar-header`, `.dark-toggle`, `.nav-section` styles
- Added `.search-dropdown` styles for inline search results
- Removed redundant `<button type="submit">` from search form

**`web/theme/default/theme.js`** (new, 2.4 KB)
- `toggleDark()`: flips `data-theme` attribute, persists to `localStorage`
- Client-side search IIFE: fetches `/_search_index.json` once on load; debounces `input` events (200 ms); filters with `String.prototype.includes`; title matches score 2, snippet matches score 1; renders up to 12 results in dropdown; Enter still submits form (server-side fallback)

**`internal/web/server_test.go`**
- `TestNavFromIndexSections` — AC#2: section headings + links from index.md appear in nav
- `TestNavFallbackToDirTree` — AC#2 fallback: when index.md has no sections, flat dir listing used
- `TestThemeJS` — AC#6: theme.js served from embedded FS

### Asset sizes
- `style.css`: 3.8 KB
- `theme.js`: 2.4 KB
- **Total: 6.2 KB** (well under 50 KB AC#5 limit)

### Simplify fixes applied
- Replaced hand-rolled `parseMDLink()` + line-by-line parser with `wiki.ParseIndex` (fixes dangling-pointer-into-growing-slice bug and eliminates duplicate)
- `wikiTitle()` reuses `h1Title()` from search.go via `wiki.ParseFrontmatter` instead of duplicating the H1 extraction logic
<!-- SECTION:FINAL_SUMMARY:END -->
