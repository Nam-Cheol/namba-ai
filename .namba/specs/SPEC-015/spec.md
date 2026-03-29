# SPEC-015

## Goal

Make `namba sync` outputs deterministic for tracked artifacts, ensure every managed README bundle stays aligned with the current renderer content, and tighten install/update lifecycle UX so users can see the installed version, update to the latest release, and find uninstall guidance.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Verified local context as of 2026-03-29:
  - `namba sync` first replaces managed README outputs, then runs `namba project`, then rewrites synced project docs such as change summary and release notes.
  - `buildStructureDoc` currently walks the repository with a narrow skip list, so transient paths like `.gocache/` and `.tmp/` can leak into `.namba/project/structure.md`.
  - Existing regression coverage in `internal/namba/readme_sync_test.go` proves English README and workflow-guide content, but it does not assert Korean, Japanese, or Chinese bundle freshness.
  - `cmd/namba/main.go` forwards directly into the CLI app and does not currently expose a top-level `namba --version` path.
  - `runUpdate` already defaults to the GitHub Release `latest` download URL, but that behavior is not surfaced as a clear user contract through version visibility plus README guidance.
  - The release workflow currently builds archives with `go build -trimpath -ldflags="-s -w"` and does not inject a release version into the binary.
  - Current `namba update` success messaging is functional but thin, so users in a plain command prompt do not get a very clear sense of current version, target version, platform asset, or what to do next.
  - The generated README currently documents install and update flows, but it does not document how to uninstall NambaAI cleanly.

## Problem

`namba sync` can produce noisy tracked diffs that do not reflect meaningful project changes, localized README bundles can drift behind the latest renderer contract, and the CLI install lifecycle is under-documented because users cannot reliably inspect the installed NambaAI version from the command line, clearly understand what `namba update` is doing in a terminal session, or find uninstall guidance in the generated README set.

## Desired Outcome

- Running `namba sync` on a clean repository should not create tracked doc churn from transient cache or temp directories.
- When README management is enabled, `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`, and their generated guide docs should refresh together from the current renderer content.
- `namba --version` should report the installed CLI version in a user-visible way that works outside a repository.
- Release builds should embed a trustworthy version string so installed binaries can report the release tag instead of an ambiguous placeholder.
- `namba update` should clearly remain a latest-release updater by default, preserve `--version vX.Y.Z` for pinned updates, and present intuitive CLI messaging about what version is being installed plus any restart/follow-up guidance.
- Generated README bundles should document how to uninstall NambaAI on supported platforms in English, Korean, Japanese, and Chinese.
- Regression tests should lock these behaviors so future sync, renderer, or install lifecycle changes cannot silently reintroduce drift.

## Scope

- Define and implement a deterministic exclusion policy for structure/project doc generation so transient runtime files and directories do not affect tracked sync artifacts.
- Tighten README bundle generation and replacement so all configured languages refresh consistently during `namba sync`.
- Add a version-reporting CLI path for `namba --version`, including an explicit release/build-version injection strategy that works for installed binaries and locally built development binaries.
- Improve `namba update` terminal UX so plain command-prompt users get understandable progress, target-version, success, restart, and failure guidance without needing to read source code.
- Make the default `namba update` latest-release behavior explicit in tests and user-facing docs.
- Add uninstall guidance to the generated README bundles and any directly related generated guides where that lifecycle information belongs, keeping the multilingual set aligned.
- Add regression tests for structure doc skip rules, multilingual README/guide synchronization, version reporting, update defaults, and update UX copy where stable assertions are appropriate.
- Preserve the current responsibility split where `namba sync` remains a local artifact refresh command rather than PR or merge automation.

## Non-Goals

- Do not redesign `namba sync` into PR, land, or release automation.
- Do not introduce external translation services or machine-generated localization beyond the current repo-owned renderer.
- Do not redesign the release packaging format or move distribution away from GitHub Release assets.
- Do not broaden the work into unrelated codemap redesign or documentation-system refactors.

## Design Constraints

- Keep `.namba/` as the source of truth for synced project docs and generated workflow assets.
- Ignore transient runtime state instead of snapshotting it into tracked docs.
- Preserve the existing English README contract while extending explicit guarantees to configured additional languages.
- Keep `namba update` aligned with the existing GitHub Release asset model rather than inventing a second update channel.
- Prefer a simple, deterministic version injection path that works in CI and is easy to reason about during local builds.
- Keep the implementation testable in local CI without network access.
