---
id: doc-2
title: Existing Wiki - Structure and Conventions
type: other
created_date: '2026-04-13 21:02'
---
# Existing Wiki Analysis

Captured from `existing/wiki/` and `existing/skill.md`. These conventions must keep working — the MCP server reads/writes this exact shape.

## Directory Layout

```
wiki/
  index.md                  # Master catalog. Curated. Updated every ingest.
  log.md                    # Append-only activity log. Chronological.
  <topic>/                  # Topical subdir (created on demand)
    summary.md              # Per-research-session summary
    <sub-page>.md
  entities/                 # Named things (tools, models, people, projects)
    <entity>.md             # lowercase-hyphen name
  concepts/                 # Ideas, patterns, techniques
    <concept>.md
  infrastructure/           # This repo's own infra notes
    <component>.md
sources/                    # Raw immutable inputs (optional sibling of wiki/)
  <slug>.md|.pdf|.png
```

Sibling `sources/` observed in `skill.md` but NOT present in the copied `existing/wiki/` — treat as optional.

## Page Conventions

- Filenames: lowercase, hyphens, `.md`. Example: `qwen2.5-coder.md`, `llm-quantization.md`.
- Entity names: exact casing as used in sources (file kept lowercase; title inside the page keeps original).
- Optional YAML frontmatter:
  ```yaml
  ---
  tags: [concept, architecture]
  sources: [sources/paper.pdf]
  updated: 2026-04-05
  ---
  ```
- Cross-refs: `[[Page Title]]` (Obsidian-style) OR `[text](../path/page.md)` (plain markdown). Both seen in wild.
- Contradiction marker: `> ⚠️ Contradicts: [other page]` inline.

## index.md Structure

Header + "How to Use" + `## Pages` section grouped by emoji-prefixed subheaders:

- `### 🔬 Research`
- `### 🏷️ Entities`
- `### 💡 Concepts`
- `### 🏗️ Infrastructure`

Each bullet: `- [Page Title](relative/path.md) — one-line summary`.

Footer `## Stats` block: sources ingested, wiki page count, last updated date.

## log.md Structure

- Frontmatter-less.
- Strict entry format: `## [YYYY-MM-DD] <operation> | <title>` where operation ∈ `ingest | query | lint`.
- Body: short bullets. Must list pages created/updated.
- Append-only. Never rewrite history.

## Three Operations (defined in skill.md)

1. **Ingest** — caller provides source (URL/PDF/image/text). Workflow: read full source → discuss → create/update summary page + relevant entity/concept pages + cross-refs → update `index.md` → append `log.md`. *One source at a time, fully finished.*
2. **Query** — caller asks question. Workflow: read `index.md` → read relevant pages in full → synthesize with citations. Offer to file substantial answers as new pages.
3. **Lint** — health check. Scan for: contradictions, orphans, staleness, gaps, missing cross-refs. Report + offer fixes. Append `## [YYYY-MM-DD] lint | pass N` to log.

## What the MCP Server Must Preserve

- Exact file layout (no renaming, no DB migration).
- Frontmatter parse/preserve round-trip.
- index.md section ordering + emoji headers.
- log.md append-only discipline.
- Support both Obsidian `[[links]]` and markdown `[text](path)` links (don't rewrite one to the other).

## Observed Stats (current real wiki, `existing/wiki/`)

- 21 pages
- 4 top-level categories + 1 topical subdir (`ollama-java-code-review/`)
- 10 entity pages, 2 concept pages, 1 infra page, 8 topical pages
- Most recent activity: 2026-04-11 (lint pass 1)
