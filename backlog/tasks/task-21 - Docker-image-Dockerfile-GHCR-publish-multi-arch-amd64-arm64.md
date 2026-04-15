---
id: TASK-21
title: 'Docker image: Dockerfile, GHCR publish, multi-arch (amd64 + arm64)'
status: Done
assignee: []
created_date: '2026-04-13 21:10'
updated_date: '2026-04-15 20:02'
labels: []
milestone: m-4
dependencies:
  - TASK-20
documentation:
  - doc-5
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Ship an official container image for users who want zero-install or to mount into an AI-host container.

Dockerfile: multi-stage. Build stage uses `golang:1.23-alpine` to produce a static binary; runtime stage is `gcr.io/distroless/static-debian12:nonroot` (or `scratch` if no CA bundle needed). Binary copied in, nothing else. Final image ~15 MB.

Non-root by default. `EXPOSE 9000`. Volume at `/wiki`. Default CMD: `["/wiki-mcp", "--serve-only", "--wiki-path=/wiki", "--bind=0.0.0.0", "--port=9000"]`.

For MCP stdio over Docker, document `docker run -i --rm -v $WIKI:/wiki ghcr.io/<user>/wiki-mcp --wiki-path /wiki` in the usage notes — this is how AI hosts plug the container into their MCP config.

Multi-arch (amd64 + arm64) via `docker/build-push-action` + QEMU. Tags: `latest` on main, `vX.Y.Z` + `vX.Y` + `vX` on release tags. Publish to `ghcr.io`.

Integrate into goreleaser: goreleaser can build + push Docker images in the same release job.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Dockerfile builds a <20 MB runtime image on distroless/scratch base
- [x] #2 Image runs as non-root user
- [x] #3 Multi-arch manifests (amd64 + arm64) published to ghcr.io
- [x] #4 `docker run -i ghcr.io/<user>/wiki-mcp --help` prints CLI help
- [x] #5 Mounting a wiki at `/wiki` and running the default CMD serves the UI on 9000
- [x] #6 Image build is driven by goreleaser in the release job (no separate Dockerfile CI)
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

1. **`cmd/wiki-mcp/main.go`** — add `--serve-only` flag: force `Web.Enabled=true`, skip stdio goroutine, exit after web server. (Partial AC#1 of TASK-19; scoped to what Docker needs.)

2. **`Dockerfile`** — single-stage runtime on `gcr.io/distroless/static-debian12:nonroot`. Goreleaser provides pre-built binary in build context — no builder stage. ENTRYPOINT + CMD split for AC#4 (`docker run image --help` replaces CMD, keeps ENTRYPOINT).

3. **`.goreleaser.yaml`** — add `dockers:` (two entries, one per arch, `use: buildx`, `--platform` flag, per-arch image tag) + `docker_manifests:` (four manifest tags: `vX.Y.Z`, `vX.Y`, `vX`, `latest`).

4. **`.github/workflows/release.yml`** — add `packages: write` permission; insert QEMU, buildx, and GHCR login steps before goreleaser step.

Note: Dockerfile is single-stage (not multi-stage) because goreleaser provides the binary; the task description assumed standalone use.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Shipped Docker image support for wiki-mcp.

**Files changed:**
- `Dockerfile` — single-stage runtime on `gcr.io/distroless/static-debian12:nonroot`. Goreleaser provides the pre-built static binary; no builder stage needed. `ENTRYPOINT ["/wiki-mcp"]` + `CMD ["--serve-only", ...]` split so `docker run image --help` overrides CMD only, keeping ENTRYPOINT (AC#4). Non-root by default (distroless nonroot variant). EXPOSE 9000, VOLUME /wiki (AC#1, AC#2).
- `.goreleaser.yaml` — added `dockers:` (two entries, linux/amd64 + linux/arm64, `use: buildx`, OCI labels) and `docker_manifests:` (four tags: `vX.Y.Z`, `vX.Y`, `vX`, `latest`). `latest` skips push on pre-release tags. Goreleaser drives the entire Docker build in the same release job (AC#3, AC#6).
- `.github/workflows/release.yml` — added `packages: write` permission; QEMU, buildx, and GHCR login steps inserted before goreleaser (AC#3).
- `cmd/wiki-mcp/main.go` — added `--serve-only` flag: forces `Web.Enabled=true`, skips stdio goroutine. Transport validation skipped in serve-only mode. Required for default CMD to serve the web UI on port 9000 without stdin (AC#5). Partially implements TASK-19 AC#1.

**Design notes:**
- Dockerfile is single-stage (not multi-stage as described) because goreleaser pre-builds the static binary and copies it into the Docker build context — no need for a Go builder stage.
- `latest` manifest skips pre-release tags via `skip_push: "{{ if .Prerelease }}true{{ end }}"`.
- `docker run -i --rm -v $WIKI:/wiki ghcr.io/robertstevens/wiki-mcp --wiki-path /wiki` overrides CMD for MCP stdio use.
<!-- SECTION:FINAL_SUMMARY:END -->
