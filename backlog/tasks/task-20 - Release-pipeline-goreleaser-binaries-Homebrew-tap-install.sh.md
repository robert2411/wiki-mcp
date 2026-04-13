---
id: TASK-20
title: 'Release pipeline: goreleaser binaries + Homebrew tap + install.sh'
status: To Do
assignee: []
created_date: '2026-04-13 21:10'
updated_date: '2026-04-13 21:19'
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
- [ ] #1 `goreleaser release --snapshot` produces binaries for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- [ ] #2 Binaries are static (CGO_ENABLED=0) and run without dynamic link deps
- [ ] #3 `brew install <user>/tap/wiki-mcp` works after the first release publishes the formula
- [ ] #4 `scripts/install.sh` detects OS/arch and installs the right binary on macOS + Linux (tested)
- [ ] #5 Tagging `v0.1.0` publishes a GitHub Release with all archive assets + checksums + SBOM
- [ ] #6 `go install github.com/<user>/wiki-mcp/cmd/wiki-mcp@latest` works
- [ ] #7 LICENSE file committed
<!-- AC:END -->
