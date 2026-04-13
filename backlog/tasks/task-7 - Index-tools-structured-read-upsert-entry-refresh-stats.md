---
id: TASK-7
title: 'Index tools (structured read, upsert entry, refresh stats)'
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
Implement `index_read`, `index_upsert_entry`, `index_refresh_stats` per doc-3. `index.md` has a strict shape (doc-2 § "index.md Structure") — emoji-prefixed section headers (🔬 Research, 🏷️ Entities, 💡 Concepts, 🏗️ Infrastructure), bullet entries `- [Title](path) — summary`, and a `## Stats` footer.

`index_read` returns `{sections: [{key, title, entries: [{title, path, summary}]}], stats: {...}}`.

`index_upsert_entry` takes `(section_key, title, path, summary)` — creates the entry if absent, updates the summary if title+path already present. Preserves section order and any free-form content between sections. If the target section key doesn't match an existing section but matches one in `[index].sections` config, insert the section in the configured order.

`index_refresh_stats` recomputes page count and last-updated date from disk and rewrites the `## Stats` block without touching the pages list.

Round-trip test: read `existing/wiki/index.md`, write it back unchanged. Must be byte-identical modulo trailing newline.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `index_read` parses `existing/wiki/index.md` into structured form without data loss
- [ ] #2 Writing the parsed form back produces byte-identical output (round-trip test passes)
- [ ] #3 `index_upsert_entry` updates summary when title+path match; otherwise appends within the correct section
- [ ] #4 New sections are inserted per `[index].sections` config order
- [ ] #5 `index_refresh_stats` rewrites only the stats block; pages list is untouched
- [ ] #6 All mutations respect `read_only` config flag
<!-- AC:END -->
