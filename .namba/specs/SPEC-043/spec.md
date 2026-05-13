# SPEC-043

## Problem

The NambaAI manual surface under `docs/` is accurate but too linear. `docs/getting-started.*` and `docs/workflow-guide.*` give the right facts, yet they lack the visual scanning aids users expect when deciding how to install NambaAI, which command to use, how the Codex workflow proceeds, and where to continue reading. The problem repeats across English, Korean, Japanese, and Simplified Chinese, so improving only one locale would create drift.

The reference direction is the visible structure of the Ouroboros Korean README: language navigation, compact identity framing, badges or status links where useful, a short table of contents, quick-start blocks, feature or command tables, and section groupings that make a long manual feel approachable in GitHub Markdown.

## Goal

Redesign the NambaAI manual docs so every maintained locale is easier to scan, faster to start from, and more consistent across languages while preserving the existing technical truth of the CLI and workflow.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Primary product promise: NambaAI guides Codex work by turning vague requests into goal, scope, constraints, and acceptance criteria before implementation.
- Current manual targets:
  - `docs/getting-started.md`
  - `docs/getting-started.ko.md`
  - `docs/getting-started.ja.md`
  - `docs/getting-started.zh.md`
  - `docs/workflow-guide.md`
  - `docs/workflow-guide.ko.md`
  - `docs/workflow-guide.ja.md`
  - `docs/workflow-guide.zh.md`
- Supporting docs to evaluate for navigation polish without rewriting their historical or reference purpose:
  - `docs/codex-upstream-reference.md`
  - `docs/moai-adk-codex-migration-analysis.md`
- Root README files are sync-managed surfaces. Do not hand-edit generated README output unless the implementation identifies and updates the renderer or config source that owns those files.

## Scope

- Apply a reusable GitHub-compatible documentation structure to the getting-started and workflow-guide manuals in English, Korean, Japanese, and Simplified Chinese.
- Add or improve, per locale, language switch links, compact top navigation, quick-start path, command-selection tables, workflow phase summaries, and deeper-reference links.
- Use badges or status links only where they carry real information such as latest release, CI, security policy, or supported platform status.
- Preserve all existing guidance for install, update, uninstall, init, project, plan, harness, fix, run, queue, sync, pr, land, release, reviews, generated assets, and collaboration defaults.
- Keep reference docs navigable from the manuals and decide whether they need only a small top navigation block or a deeper visual pass.

## Out Of Scope

- Changing CLI behavior, command names, install scripts, release automation, or validation policy.
- Adding new documentation tooling unless a small helper is necessary for parity validation.
- Translating new product claims that are not backed by code, project docs, or existing manual content.
- Hand-editing generated README bundles without updating the owning renderer or config.

## Constraints

- Code, executable config, and current project docs outrank stale prose when content conflicts appear.
- All four locales must keep equivalent structure and equivalent meaning, even when wording is naturally localized.
- Markdown and HTML must render acceptably on GitHub without relying on external scripts or CSS.
- Visual improvements must not reduce maintainability; repeated patterns should be simple enough for future edits.
- Generated-file warnings and sync ownership comments must remain intact where they already exist.

## Reference Direction

- Adopt from the Ouroboros README reference: top language navigation, strong first-screen identity, compact link row, quick-start emphasis, command and feature tables, and clear section breaks.
- Adapt for NambaAI: the tone should be practical and workflow-oriented rather than purely manifesto-style. The manual should help a maintainer choose the next command quickly.
- Avoid: decorative-only badges, dense emoji noise, tables that duplicate paragraphs without making decisions easier, and one-locale-only polish.
