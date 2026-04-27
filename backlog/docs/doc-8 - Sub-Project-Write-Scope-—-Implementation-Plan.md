---
id: doc-8
title: Sub-Project Write Scope — Implementation Plan
type: other
created_date: '2026-04-27 16:43'
---
# Sub-Project Write Scope — Implementation Plan

## Goal

When wiki-mcp is configured with a sub-project (e.g. `msb-cb/mappers`), write operations must be scoped to:
- **Own dir**: `msb-cb/mappers/**` — any depth
- **Parent dir, top-level only**: files directly in `msb-cb/` (e.g. `msb-cb/foo.md`), not into sibling sub-projects
- **Blocked**: wiki root and sibling sub-projects (e.g. `msb-cb/other-subproject/**`)

---

## Design: Path Semantics

**Choice**: Relax `ConfineToWikiPath` boundary to `ProjectPath` (parent) when `SubProjectPath` is active, then enforce sibling exclusion via `MustAllowWrite`.

**How agents address parent files**: via `../parent-file.md` (relative to `Root()` = SubProjectPath). `filepath.Join` resolves `..`, and the boundary is now `ProjectPath` not `SubProjectPath`, so it reaches the parent.

**Tradeoff**: punches a controlled hole in path confinement. Mitigated by `MustAllowWrite` which enforces sibling exclusion. The old confinement check's goal (block escape attacks) is preserved — you can only traverse up to `ProjectPath`, not beyond.

---

## Config Changes (`internal/config/config.go`)

| Location | Change |
|---|---|
| `Config` struct | Add `SubProjectPath string` with `toml:"sub_project_path"` |
| `applyEnvOverrides` | Add `envStr("WIKI_MCP_SUB_PROJECT_PATH", &cfg.SubProjectPath)` |
| CLI flags | Add `--sub-project-path` flag in `cmd/wiki-mcp/main.go` |
| `validate()` | Resolve SubProjectPath (abs, must be within ProjectPath, must not equal it, requires ProjectPath set) |
| `Root()` | Return `SubProjectPath` if set, else existing logic |
| `ResolveWikiPath()` | When `SubProjectPath != ""`, confine to `ProjectPath` instead of `Root()` |
| New method | `MustAllowWrite(absPath string) error` — see permission rule below |

---

## Permission Rule (`MustAllowWrite`)

```
if SubProjectPath == "": return nil  // no sub-project restrictions

allow if absPath is within SubProjectPath (own scope)
allow if filepath.Dir(absPath) == ProjectPath (direct file in parent, no subdirs)
block all else → ErrCodeForbidden
```

---

## Write-Site Changes (`internal/wiki/`)

Every mutating function calls `cfg.MustAllowWrite(abs)` after resolving the path, before writing.

| File | Functions |
|---|---|
| `wiki.go` | `PageWrite`, `PageAppend`, `PageDelete` |
| `wiki.go` | `PageMove` — check both `oldAbs` and `newAbs` |
| `index.go` | `IndexUpsertEntry`, `IndexRefreshStats` |
| `log.go` | `LogAppend` |
| `init.go` | `WikiInit` — root-level meta bootstrap is already guarded by `cfg.ProjectPath == ""`; add guard: skip root bootstrap if `SubProjectPath != ""` too (no change needed in practice since the guard catches it, but verify) |

---

## Open Question

`index_upsert_entry` and `log_append` write to `Root()/index.md` and `Root()/log.md`. When sub-project is active, `Root()` = SubProjectPath, so those go to the sub-project's own index/log. **Do sub-projects also need to upsert entries into the parent's `index.md` and `log.md`?**

If yes: those tools need a `--target=parent` parameter or a separate "parent index upsert" tool.
If no: current behavior is fine — sub-project manages its own index/log only.

---

## Test Matrix

| Scenario | Expected |
|---|---|
| Write own sub-project file | OK |
| Write file directly in parent dir | OK |
| Write into sibling sub-project | `ErrCodeForbidden` |
| Write to wiki root | `ErrCodeForbidden` |
| `../../../escape.md` path | blocked by confinement to ProjectPath |
| `PageMove` old=own, new=parent-direct | OK |
| `PageMove` old=own, new=sibling | `ErrCodeForbidden` |
| No SubProjectPath set | no change in behavior |

---

## Out of Scope

- Nested sub-sub-projects (3+ levels)
- `project_list` listing sub-projects (currently scans `WikiPath` one level deep)
- Retroactive enforcement on existing wikis
- Per-user ACLs
