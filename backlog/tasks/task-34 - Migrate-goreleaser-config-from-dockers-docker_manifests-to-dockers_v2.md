---
id: TASK-34
title: Migrate goreleaser config from dockers/docker_manifests to dockers_v2
status: To Do
assignee: []
created_date: '2026-04-27 20:15'
labels:
  - chore
  - goreleaser
  - docker
dependencies: []
references:
  - 'https://goreleaser.com/deprecations#dockers'
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
GoReleaser is phasing out `dockers` and `docker_manifests` in favor of `dockers_v2`. The current `.goreleaser.yml` uses the deprecated keys and will break when support is removed.

Migrate all `dockers` and `docker_manifests` config blocks to the `dockers_v2` equivalent following the official deprecation guide.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 All `dockers` blocks replaced with `dockers_v2` equivalents
- [ ] #2 All `docker_manifests` blocks replaced with `dockers_v2` multi-arch manifest equivalents
- [ ] #3 goreleaser check passes with no deprecation warnings
- [ ] #4 CI release pipeline builds and pushes Docker images successfully
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linting is done: (golangci-lint run)
- [ ] #2 The code is committed
<!-- DOD:END -->
