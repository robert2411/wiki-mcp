You are maintaining a personal technical wiki using the wiki-mcp MCP server.

Run a lint pass over the entire wiki: find issues, report them, offer fixes, then log the pass.

---

## Lint Workflow

### Step 1 — Read the index

Call `index_read()` to get the full structured index. This gives you every registered page, its path, and its one-line summary.

### Step 2 — Determine pass number

Call `log_tail(n=50)` and scan the returned entries for any whose title matches the pattern `lint | pass N` (where N is an integer). Find the highest N; the new pass number is N+1. If no lint entries exist, use pass 1.

### Step 3 — Find orphan pages

Call `orphans()` to get all pages with zero incoming links (excluding index.md and log.md). These are candidates for either linking into the wiki or deletion.

### Step 4 — Scan all pages

For each page listed in the index, call `page_read(path)` to load the full content. As you read, check for:

- **Contradictions** — conflicting claims across pages (e.g. two pages disagree on a version number, a benchmark result, or a capability). Flag with the conflicting pages and the specific claims.
- **Staleness** — claims that a more recently ingested source may have superseded. Flag pages whose `updated` frontmatter date is significantly older than related pages.
- **Gaps** — concepts or tools mentioned frequently across multiple pages but lacking a dedicated page.
- **Missing cross-refs** — pages that clearly relate to each other but have no link between them. For each suspected gap, call `links_incoming(path)` on the target page to confirm no link exists from the source page.

### Step 5 — Report findings

Present a structured report with sections for each category:

```
## Lint Report — pass N

### Orphans (N)
- path/to/page.md — no pages link here

### Contradictions (N)
- page-a.md vs page-b.md: [description of conflict]

### Staleness (N)
- path/to/page.md — last updated YYYY-MM-DD, possibly superseded by [newer page]

### Gaps (N)
- "concept name" — mentioned in [list of pages] but no dedicated page

### Missing cross-refs (N)
- page-a.md should link to page-b.md — [reason]
```

For each finding, offer to fix it. Wait for the user to confirm before making any changes.

### Step 6 — Log the pass

After reporting (and applying any fixes the user requested), call:

```
log_append(operation="lint", title="lint | pass N", body="<summary of findings and fixes>")
```

This appends the entry `## [YYYY-MM-DD] lint | pass N` to log.md.
