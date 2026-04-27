---
name: namba-release
description: Command-style entry point for NambaAI release orchestration and release-note handoff.
---

Use this skill when the user explicitly says `$namba-release`, `namba release` in a Codex workflow context, or Korean wording such as `릴리즈 진행해`.

Behavior:
- Treat this as NambaAI-specific release orchestration, not a generic release helper.
- Start from `main` and require a clean working tree before the final tagging step.
- If generated templates or docs changed during release prep, run `namba regen` and/or `namba sync`, validate, and commit those changes before tagging.
- Determine the target version from explicit input or the next semver bump.
- Collect commits since the previous semver tag, ignoring merge noise and excluding any release-note prep commit when the notes artifact is committed separately.
- Draft release notes from that commit range and group them into user-visible changes, fixes, docs/workflow, and internal maintenance while preserving SPEC IDs, PR numbers, and short commit hashes when useful.
- Write the notes to a durable per-version artifact such as `.namba/releases/<version>.md`, then use that file as the handoff for the guarded `namba release --version <version> --push` path.
- Write release-facing prose in Korean by default for this repository unless configuration changes the language contract.
- Do not tag until the notes exist and validation has passed.
- Make sure the GitHub Release body uses the generated notes rather than an empty or generic body.
