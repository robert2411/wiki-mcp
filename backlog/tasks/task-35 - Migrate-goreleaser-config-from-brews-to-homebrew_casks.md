---
id: TASK-35
title: Migrate goreleaser config from brews to homebrew_casks
status: To Do
assignee: []
created_date: '2026-04-27 20:15'
labels:
  - chore
  - goreleaser
  - homebrew
dependencies: []
references:
  - 'https://goreleaser.com/deprecations#brews'
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
GoReleaser is phasing out `brews` in favor of `homebrew_casks`. The current `.goreleaser.yml` uses the deprecated `brews` key and will break when support is removed.

Migrate the `brews` config block to `homebrew_casks` following the official deprecation guide.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 All `brews` blocks replaced with `homebrew_casks` equivalents
- [ ] #2 goreleaser check passes with no deprecation warnings
- [ ] #3 CI release pipeline publishes Homebrew cask successfully
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linting is done: (golangci-lint run)
- [ ] #2 The code is committed
<!-- DOD:END -->
