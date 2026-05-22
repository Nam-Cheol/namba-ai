---
name: namba-release
description: Command-style entry point for NambaAI release orchestration and release-note handoff.
---

State effect: mutating workflow entry point. Use help/probe paths read-only, and otherwise expect repository state or GitHub state to change.

Generated instruction contract for this command skill:
- Purpose: keep the role or command scope explicit, bounded, and testable.
- Boundary: honor read-only versus mutating state effects, configured sandbox mode, and assigned file or workflow ownership.
- Required output: report concrete actions, changed paths or artifacts, validation evidence, and pass/fail status or blockers.
- Pass/fail criteria: claim success only when acceptance criteria and configured validation are satisfied; otherwise name the exact blocker and impact.
- Evidence expectations: cite source artifacts such as SPEC files, `.namba/` configs, diffs, test output, PR/check links, or generated manifests instead of relying on unsupported assertions.
- Security responsibilities: never expose or commit secrets; treat auth, privacy, destructive commands, permission changes, and external network or credential use as security-sensitive.
- Destructive command and escalation policy: do not run destructive commands unless explicitly requested, and request approval for privileged, networked, or sandbox-blocked actions.
- Fallback implementer boundary: if a specialist path is unavailable and the main/default implementer takes over, stay within the assigned scope and preserve the same evidence and validation duties.
- Portability: keep durable guidance non-project-specific unless the current repository config or SPEC explicitly provides the project detail.

Use this skill when the user explicitly says `$namba-release`, `namba release` in a Codex workflow context, or Korean wording such as `릴리즈 진행해`.

Behavior:
- Treat this as NambaAI-specific release orchestration, not a generic release helper.
- Start from `main` and require a clean working tree before the final tagging step.
- If generated templates or docs changed during release prep, run `namba regen` and/or `namba sync`, validate, and commit those changes before tagging.
- Determine the target version from explicit input or the next semver bump.
- Collect commits since the previous semver tag, ignoring merge noise and excluding any release-note prep commit when the notes artifact is committed separately.
- Draft release notes from that commit range and group them into user-visible changes, fixes, docs/workflow, and internal maintenance while preserving SPEC IDs, PR numbers, and short commit hashes when useful.
- Release notes must describe actual changes and include evidence sources such as SPEC IDs, PR numbers, short commit hashes, source artifacts, and validation results when available.
- Write the notes to a durable per-version artifact such as `.namba/releases/<version>.md`, then use that file as the handoff for the guarded `namba release --version <version> --push` path.
- Write release-facing prose in the explicit user-requested language when one is given, such as `--language en`; otherwise use the init or project-configured language.
- Do not tag until the notes exist and validation has passed.
- Make sure the GitHub Release body uses the generated notes rather than an empty or generic body.
