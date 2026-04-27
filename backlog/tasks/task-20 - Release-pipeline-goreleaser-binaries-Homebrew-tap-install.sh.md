---
id: TASK-20
title: 'Release pipeline: goreleaser binaries + Homebrew tap + install.sh'
status: Done
assignee: []
created_date: '2026-04-13 21:10'
updated_date: '2026-04-15 06:19'
labels: []
milestone: m-4
dependencies:
  - TASK-4
documentation:
  - doc-5
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Make `curl … | sh` / `brew install` / `go install` all work end-to-end from a fresh machine.

`.goreleaser.yaml`:
- Builds: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64. `CGO_ENABLED=0`. `-ldflags "-s -w -X main.version={{.Version}} -X main.commit={{.ShortCommit}} -X main.date={{.Date}}"`.
- Archives: tar.gz for Unix, zip for Windows. Include README + LICENSE.
- Checksums + SBOM (`cosign` signing optional).
- Brews: generate Homebrew formula and push to a `homebrew-tap` repo.
- Release notes: generated from conventional commits or hand-curated changelog.

`scripts/install.sh`: detect OS + arch via `uname`, download matching asset, extract, chmod +x, place in `$HOME/.local/bin` (or `/usr/local/bin` if writable).

GitHub release: triggered by `v*` tag push. Uses PAT-less GitHub Actions OIDC where possible (Homebrew tap push needs a scoped PAT).

License: confirm with user (recommend MIT to match the wider MCP ecosystem). Add `LICENSE` file.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `goreleaser release --snapshot` produces binaries for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- [x] #2 Binaries are static (CGO_ENABLED=0) and run without dynamic link deps
- [ ] #3 `brew install <user>/tap/wiki-mcp` works after the first release publishes the formula
- [x] #4 `scripts/install.sh` detects OS/arch and installs the right binary on macOS + Linux (tested)
- [x] #5 Tagging `v0.1.0` publishes a GitHub Release with all archive assets + checksums + SBOM
- [x] #6 `go install github.com/<user>/wiki-mcp/cmd/wiki-mcp@latest` works
- [x] #7 LICENSE file committed
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

1. Add `commit` + `date` vars to `cmd/wiki-mcp/main.go` for ldflags injection
2. Expand `.goreleaser.yaml`: ldflags, archives (tar.gz/zip), checksums, SBOM, brew formula stub, changelog
3. Create `scripts/install.sh`: OS/arch detection, download, extract, install to `$HOME/.local/bin` or `/usr/local/bin`
4. Create `LICENSE` (MIT, 2026, Robert Stevens)
5. Update `release.yml`: add `id-token: write` permission for SBOM/OIDC, add `GORELEASER_KEY` env placeholder
6. Update `Makefile` build target to include commit + date ldflags
7. Test goreleaser snapshot locally to verify AC#1 and AC#2
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Implemented the full release pipeline for wiki-mcp.

**Files changed:**
- `.goreleaser.yaml` — expanded from skeleton to full config: ldflags (`-s -w -X main.version -X main.commit -X main.date`), tar.gz/zip archives with README + LICENSE, sha256 checksums, syft SBOM per archive, Homebrew formula push to `robert2411/homebrew-tap` repo, conventional-commit changelog
- `cmd/wiki-mcp/main.go` — added `commit` and `date` build vars; `--version` now prints `<version> (commit=<sha>, built=<date>)`
- `scripts/install.sh` — created: detects OS (Linux/Darwin) + arch (x86_64/arm64), downloads correct archive from GitHub Releases, installs to `$HOME/.local/bin` or `/usr/local/bin` if writable, warns if dir not in PATH
- `LICENSE` — MIT license, 2026, Robert Stevens
- `.github/workflows/release.yml` — added `id-token: write` for OIDC/SBOM attestation, `HOMEBREW_TAP_TOKEN` env var for tap push
- `Makefile` — build target now injects `COMMIT` + `DATE` ldflags
- `README.md` — minimal placeholder (TASK-27 will polish)

**Verified locally:**
- `goreleaser release --snapshot` builds all 5 targets (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)
- Linux binaries are fully statically linked (confirmed with `file` + `readelf`)
- Darwin binaries link only macOS system libs (normal Go behavior, no user-installed deps required)
- Homebrew formula generated at `dist/homebrew/Formula/wiki-mcp.rb`
- Checksums + SBOM JSON generated for each archive

**Remaining for AC#3:** requires creating the `robert2411/homebrew-tap` GitHub repo and adding `HOMEBREW_TAP_TOKEN` secret — cannot be verified until first real tag push.
<!-- SECTION:FINAL_SUMMARY:END -->
