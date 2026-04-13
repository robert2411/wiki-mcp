---
id: TASK-21
title: 'Docker image: Dockerfile, GHCR publish, multi-arch (amd64 + arm64)'
status: To Do
assignee: []
created_date: '2026-04-13 21:10'
updated_date: '2026-04-13 21:19'
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
- [ ] #1 Dockerfile builds a <20 MB runtime image on distroless/scratch base
- [ ] #2 Image runs as non-root user
- [ ] #3 Multi-arch manifests (amd64 + arm64) published to ghcr.io
- [ ] #4 `docker run -i ghcr.io/<user>/wiki-mcp --help` prints CLI help
- [ ] #5 Mounting a wiki at `/wiki` and running the default CMD serves the UI on 9000
- [ ] #6 Image build is driven by goreleaser in the release job (no separate Dockerfile CI)
<!-- AC:END -->
