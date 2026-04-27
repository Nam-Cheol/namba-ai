# Engineering Review

- Status: clear
- Last Reviewed: 2026-04-27
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock generated-surface registration and regen/sync ownership for the two new managed skills.
- Verify GitHub review-thread resolution feasibility with thread-aware state, not flat PR comments.
- Lock release-note handoff mechanics so generated notes actually become the GitHub Release body.
- Preserve release guardrails, validation order, and existing asset publication behavior.

## Findings

- Generated-surface ownership is dual-keyed in the current implementation. `internal/namba/codex.go` uses `managedCodexSkillNames()` to decide managed-skill ownership and cleanup, while `codexSkillTemplates()` is the materialization map that actually writes repo skills during regen. `SPEC-039` already points at both, which is correct, but implementation also has to update the repeated generated-doc surfaces in `internal/namba/templates.go` and `internal/namba/readme.go` rather than treating registry plus template creation as sufficient. `renderAgents()`, `renderNambaSkill()`, the command-mapping and execution-rules sections, `renderCodexUsage*`, and the Namba CLI README sections all repeat the command-skill surface today.
- Doctor coverage should stay intentionally narrow unless the contract is explicitly widened. The current `codexNativeIssues()` checks only a subset of mandatory repo skills and core Codex assets; it is not a complete enumeration of every managed skill. The SPEC's "extend doctor/scaffold checks only if the expected managed skill set intentionally grows there" is the right boundary and should remain a conscious decision rather than accidental scope creep.
- GitHub review-thread resolution is not safely expressible through the current repo-owned PR command path alone. `internal/namba/pr_land_command.go` handles PR creation, reuse, top-level PR comments, merge checks, and the `@codex review` marker, but it does not model review threads, thread resolution state, or inline review anchors. The connected GitHub review skill guidance also states that thread-aware state requires `gh api graphql` or the bundled thread-fetch script because flat connector comments do not preserve `reviewThreads`, `isResolved`, or inline context. That means `$namba-review-resolve` should stay a Namba-owned skill orchestration surface backed by GitHub app metadata plus `gh` GraphQL/thread-aware reads, not a hidden extension of the existing `namba pr`/`namba land` Go command flow.
- Release-note publication is the main architecture gap in the current implementation. `internal/namba/release_command.go` correctly enforces git-repo, `main`, clean-tree, validator, semver, duplicate-tag, and push guardrails before tag creation, but it has no release-note generation or handoff concept. `.github/workflows/release.yml` publishes assets and `checksums.txt` with `softprops/action-gh-release@v2`, yet it does not currently pass any generated body into the GitHub Release. `.namba/project/release-notes.md` is a sync-produced support document, not a per-version handoff artifact.
- The release-note handoff therefore needs a durable pre-tag boundary. If Namba chooses a committed notes file such as `.namba/releases/<version>.md`, the file must be generated and committed before the guarded `namba release --version <version> --push` step so the final release still starts from clean `main`. If Namba chooses a non-committed handoff file consumed directly by the publishing step, the implementation still needs an explicit executable boundary that produces a deterministic version-specific notes file before tag push. In both cases, the spec requirement to exclude the release-note prep commit from the commit range is important because otherwise a committed notes artifact will pollute `git log <previous-tag>..HEAD`.
- Validation sequencing needs to be explicit in two places. For `$namba-review-resolve`, validation must finish before any "done" reply, resolve action, or review re-request. For `$namba-release`, any `namba regen`/`namba sync` work and any release-note artifact creation or commit must happen before the final guarded `namba release` invocation, because `runRelease` already assumes a clean tree and reruns validators immediately before tagging.
- The test surface is broad but well-localized. Existing suites already cover release guardrails (`internal/namba/release_command_test.go`), command-skill anchors (`internal/namba/templates_test.go`), localized README section anchors (`internal/namba/readme_sync_test.go`), and release support docs (`internal/namba/sync_test.go`). `SPEC-039` should extend those adjacent tests instead of introducing brittle full-file snapshots or one-off integration-only coverage.

## Decisions

- Keep both additions as Namba-managed repo-local skills generated from `internal/namba/templates.go` and registered through `internal/namba/codex.go`; do not add a public `namba review-resolve` CLI command in this slice.
- Treat dual registration plus repeated generated-doc exposure as acceptance-critical behavior: implementation is incomplete if the skills exist on disk but are missing from AGENTS, `$namba`, Codex usage docs, README command-skill sections, or skill-to-command mapping.
- Lock review-thread operations to a thread-aware GitHub path. Flat PR comments are acceptable for confirming the existing `@codex review` marker, but unresolved-thread discovery, per-thread replies, and thread resolution must use a `gh`/GraphQL-capable path that preserves review-thread state.
- Lock release orchestration into a two-phase model: prepare any generated assets and version-specific release notes first, then run the existing guarded `namba release` path only after the tree is clean again. Do not weaken the current `main`/clean-tree/validator/tag guardrails to make room for release-note generation.
- Keep the GitHub Release workflow responsible for publishing the existing archive matrix plus `checksums.txt`, but require it to consume an explicit Namba-generated release-note body rather than a generic empty release description.

## Follow-ups

- Add targeted regressions for both skill registrations in `managedCodexSkillNames()` and `codexSkillTemplates()`, plus the repeated command-skill lists in `templates_test.go` and `readme_sync_test.go`.
- Add release-note rendering and handoff coverage near `internal/namba/release_command_test.go`, `internal/namba/sync_test.go`, and any workflow-facing test helper so commit-range grouping, previous-tag selection, prep-commit exclusion, and workflow body selection stay deterministic.
- Make the `$namba-review-resolve` skill contract explicit about `gh` auth, unresolved-thread reads, non-actionable-thread handling, validation-before-reply, resolve-only-when-addressed, and no duplicate `@codex review` requests.
- Choose and document one durable release-note transport before implementation starts: committed `.namba/releases/<version>.md` or an equivalent executable file handoff consumed by the publish workflow.

## Recommendation

- Clear with follow-ups. `SPEC-039` is implementable, but only if thread-aware GitHub review handling, dual registration plus repeated generated-surface rollout, release-note handoff, and pre-tag validation sequencing are treated as first-class contract items rather than cleanup details.
