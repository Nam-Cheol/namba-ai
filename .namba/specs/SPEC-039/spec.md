# SPEC-039

## Problem

NambaAI already generates README bundles, repo-local Codex skills, GitHub PR handoff guidance, and a `namba release` CLI primitive. The current surfaces still leave three workflow gaps:

- The generated NambaAI root README is useful but comparatively plain outside the hook section. It does not use visual markers consistently to make command selection, workflow stages, and Namba-specific concepts easy to scan.
- There is no Namba-managed skill for the common PR review loop: inspect review feedback, decide whether comments are meaningful, implement meaningful fixes, reply on the exact review thread, resolve the thread, then request review again.
- Release handling is exposed mainly as `namba release`, which validates and tags, but the Codex-facing "릴리즈 진행해" workflow should be a NambaAI-specific skill that also produces release notes from commit history before publishing the release.

## Goal

Deliver a managed NambaAI UX and workflow upgrade that:

1. Enhances the generated NambaAI README experience with purposeful emoji and stronger visual structure while preserving generated-source discipline.
2. Adds a first-class Namba review-addressing skill for meaningful GitHub review feedback.
3. Adds a first-class Namba release skill for "릴리즈 진행해" that drafts release notes from commits and then performs the release through the Namba release path.

## Scope

## Delivery Slicing

- Must ship: `$namba-review-resolve` and `$namba-release` as managed NambaAI workflow skills with generated-surface coverage, tests, and validation.
- Must ship unless explicitly descoped by a later SPEC decision: README visual upgrade through the generated renderer.
- Implement README polish after the workflow-skill contracts are locked so visual changes do not hide unresolved review-thread or release-note behavior.
- The SPEC is not complete until all acceptance criteria pass, unless a later planning update explicitly removes the README visual upgrade from scope.

### README Visual Upgrade

- Update the generated README renderer in `internal/namba/readme.go` rather than editing generated `README*.md` files by hand.
- Apply the upgrade to the Namba CLI README profile configured by `.namba/config/sections/docs.yaml`.
- Use emoji as semantic scan markers for major sections, command groups, lifecycle stages, release/update paths, and Namba-specific workflow concepts.
- Keep the README professional and readable: emoji should clarify structure, not replace precise text or overload every bullet.
- Preserve the generated header, hero image, language links, release/CI/security links, command snippets, and multilingual output model.
- Regenerate the managed README outputs with `namba sync` after implementation.

### Review-Addressing Skill

- Add a managed repo-local skill, proposed name: `$namba-review-resolve`.
- Trigger it when the user says `$namba-review-resolve`, asks to resolve review comments, or uses wording equivalent to "리뷰 확인하고 의미있는 리뷰면 수정 후에 해당 리뷰에 답변을 달고, resolve한 다음 다시 리뷰 요청해".
- The skill should be Namba-owned and generated from `internal/namba/templates.go`, registered through `internal/namba/codex.go`, and surfaced in AGENTS, `$namba`, Codex usage docs, README command skill sections, and skill-to-command mapping.
- The skill should resolve the active PR from the current branch when possible, inspect unresolved review threads, classify comments as meaningful/actionable or non-actionable, implement the meaningful fixes, run validation, reply on the original review threads with concrete evidence, resolve only addressed threads, and request review again without duplicating the configured `@codex review` marker.
- Non-actionable comments should be documented or answered when useful, but they must not be silently resolved if they were not actually addressed.

### Namba Release Skill

- Add a managed repo-local skill, proposed name: `$namba-release`.
- Treat release as a NambaAI-specific skill surface, not a generic release helper.
- Trigger it when the user says `$namba-release`, `namba release` in a Codex workflow context, or Korean equivalents such as "릴리즈 진행해".
- The skill should use the existing `namba release` CLI as the guarded tag/push primitive, but it must add the missing Codex-native orchestration:
  - ensure the release starts from clean `main`;
  - run sync/regen only when needed;
  - run configured validation;
  - determine the target version from explicit input or the next semver bump;
  - collect commits since the previous semver tag;
  - draft release notes from those commits, grouped by meaningful work area and SPEC/PR references when available;
  - publish the release with those notes, not an empty or generic body.
- Add a durable release-note handoff path so the GitHub release workflow can publish the generated notes, for example `.namba/releases/<version>.md` committed before tagging or an equivalent notes-file path passed to the publishing step.
- Update `.github/workflows/release.yml` only if needed so the GitHub Release body uses the Namba-generated release notes while still publishing the existing assets and `checksums.txt`.

## Implementation Targets

- `internal/namba/templates.go`
  - Add `renderReviewResolveCommandSkill()`.
  - Add `renderReleaseCommandSkill()`.
  - Update `renderAgents()`, `renderNambaSkill()`, command mapping, execution rules, and Codex usage sections.
- `internal/namba/codex.go`
  - Register `namba-review-resolve` and `namba-release` as managed skill names.
  - Add both skills to `codexSkillTemplates()`.
  - Extend doctor/scaffold checks only if the expected managed skill set intentionally grows there.
- `internal/namba/readme.go`
  - Add semantic emoji structure to the generated Namba CLI root README sections.
  - Surface `$namba-review-resolve` and `$namba-release` consistently in command skills, skill mapping, workflow guide, and technical snapshot sections.
  - Keep multilingual generated outputs coherent for English, Korean, Japanese, and Chinese.
- `internal/namba/release_command.go`, `internal/namba/release.go`, and `.github/workflows/release.yml`
  - Add release-note generation or handoff support if the release skill needs executable support beyond skill instructions.
  - Preserve existing validation, clean-main, semver, asset, checksum, and push guardrails.
- Tests
  - Extend template, scaffold, README sync, release command, and workflow tests around the new generated surfaces and release notes behavior.

## Skill Contracts

### `$namba-review-resolve`

The skill owns PR review feedback handling for NambaAI repositories.

Behavior:

- Resolve the target PR from the current branch unless the user provides a PR number.
- Read unresolved review threads with thread-level resolution state when available.
- Classify each thread:
  - meaningful/actionable: correctness, tests, security, UX, docs, release, regression, or maintainability concerns that require a concrete change or a precise answer;
  - non-actionable: duplicates, stale comments, preference-only remarks without requested change, or comments already answered by current code.
- Implement meaningful fixes in the smallest coherent patch.
- Run the configured validation commands before replying as done.
- Reply on each original thread with the change made and validation evidence.
- Resolve only the threads that were fixed or conclusively answered.
- Re-request review after all meaningful threads are addressed and validation passes.
- Do not duplicate `@codex review`; confirm the configured marker exists or add it once.

### `$namba-release`

The skill owns NambaAI release orchestration.

Behavior:

- Start from `main` and require a clean working tree before final tagging.
- If generated templates or docs changed as part of release prep, run `namba regen` and/or `namba sync`, validate, and commit those changes before tagging.
- Determine release version from `--version`, `--bump`, or next patch default.
- Build release notes from `git log <previous-tag>..HEAD`, excluding merge noise and the release-note prep commit when applicable.
- Group notes by user-visible changes, fixes, docs/workflow, and internal maintenance; preserve SPEC IDs, PR numbers, and commit hashes when useful.
- Write the release notes in the configured release language for this repo, Korean by default via `pr_language: ko` / documentation language.
- Use `namba release --version <version> --push` or equivalent guarded path to create and push the tag.
- Ensure the GitHub Release body uses the generated notes and that asset publication still includes all platform archives plus `checksums.txt`.

## Non-Goals

- Do not hand-edit generated README outputs as the source of truth.
- Do not add emoji for decoration where it weakens readability, accessibility, or command clarity.
- Do not add a public `namba review-resolve` CLI command unless implementation evidence shows the skill alone is insufficient.
- Do not turn `$namba-review-resolve` into a generic GitHub plugin wrapper; it must preserve Namba branch, validation, PR language, and review marker rules.
- Do not resolve review threads that were not addressed or conclusively answered.
- Do not create a release tag before release notes are generated and the configured validators pass.
- Do not replace the existing release asset matrix, checksum generation, update installer contract, or clean-main release guardrails.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Current README source: `internal/namba/readme.go` plus `.namba/config/sections/docs.yaml`
- Current managed skill registry: `internal/namba/codex.go`
- Current release primitive: `internal/namba/release_command.go`
- Current release publish workflow: `.github/workflows/release.yml`
