---
id: doc-1
title: Wiki MCP - Project Overview and Goals
type: other
created_date: '2026-04-13 21:02'
---
# Wiki MCP - Project Overview

## Problem

User maintains a persistent technical knowledge-base wiki (markdown pages, index, log). Current implementation is a **skill** wired into a single assistant (Andy / nanoclaw). Behavior is defined in `skill.md` and `CLAUDE.md` prompts, executed by that one agent inside a container. The wiki lives on one machine; the MkDocs web UI is glued together with host-side systemd timers.

The user wants this capability to be **portable and AI-agnostic**: any MCP-capable assistant (Claude Desktop, Claude Code, Cursor, Cline, other hosts) on any of his PCs should be able to manage the same wiki through a standard tool surface, without re-implementing the skill prompts per host.

## Goal

Build an MCP server that:

1. Exposes wiki read/write as MCP tools (and resources/prompts where useful) so any MCP client can manage the wiki.
2. Takes wiki location from **its own config** (not hard-coded, not tied to a container mount).
3. Ships a built-in **read-only web server** that serves the wiki for browsing (no separate MkDocs+systemd stack required).
4. Is **trivially deployable on multiple PCs** (one command install, cross-platform).

## Non-Goals

- Not a wiki engine (no auth, no multi-user editing UI, no comments, no revisions beyond git).
- Not a replacement for the skill's *prompt guidance* (Ingest/Query/Lint workflow). Prompts stay as MCP prompts or docs — the server provides the tools the prompts call.
- No cloud sync. Wiki lives in a local directory; user syncs via git / Syncthing / Dropbox / etc if needed.
- Not an LLM runner. Ingestion-of-web-content steps (fetch URL, pdftotext, etc) can be tool-exposed, but the *reading and synthesis* is the client LLM's job.

## Inputs to Design (what exists in this repo)

- `existing/wiki/` — current real wiki: `index.md`, `log.md`, `entities/`, `concepts/`, `infrastructure/`, topical subdirs. Frontmatter optional.
- `existing/skill.md` — current skill definition, documents Ingest/Query/Lint ops and conventions.
- `existing/CLAUDE.md.back.md` — Andy's full system prompt; wiki section starts "You maintain a persistent technical/engineering knowledge base…"
- `existing/wiki/infrastructure/wiki-ui.md` — current MkDocs+systemd web UI architecture (the thing we're replacing).

## Success Criteria

- A second PC gets the wiki going in <5 minutes: install command + config file pointing at wiki path + MCP client wired up.
- Same wiki directory works from any MCP client unchanged.
- Web view is one `--serve` flag away; no extra daemons.
- Existing `existing/wiki/` content opens without migration.
