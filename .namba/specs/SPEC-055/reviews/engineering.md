# Engineering Review

- Status: cleared
- Last Reviewed: 2026-05-20
- Reviewer: Codex
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The target workflow currently uses `windows-latest` in all three Windows build matrix rows, so the runner-notice fix is concrete and localized.
- The release workflow already has a manual `workflow_dispatch` trigger and guards the publish job with `if: startsWith(github.ref, 'refs/tags/')`, which gives the implementation a non-release execution path for build validation.
- Node.js 20 warning remediation must be evidence-driven. The implementation should identify the action emitting the warning from release-run output before changing action versions, then update only the needed `uses:` references in `.github/workflows/release.yml`.
- Existing release behavior has a regression foothold in `internal/namba/release_test.go`, including `TestReleaseWorkflowUsesNotesBodyPath`. Add or extend targeted coverage there for `windows-2022`, absence of `windows-latest`, and any action-version contract needed by the fix.
- `actionlint` is not currently available in the local shell, so static workflow validation may need to be satisfied through an available project path, installing/using actionlint in CI, or GitHub-side workflow validation evidence.

## Decisions

- Pin all Windows matrix runners in `.github/workflows/release.yml` to `windows-2022`.
- Preserve the existing build and publish job split, artifact names, archive packaging, checksums, and release body path behavior.
- Do not use tag-push validation for this SPEC.
- Keep implementation file scope to `.github/workflows/release.yml`, with tests allowed only where needed to prove the workflow contract.

## Follow-ups

- Confirm the exact Node.js 20 warning source from GitHub Actions logs before selecting action updates.
- Record the non-release workflow execution evidence in the implementation handoff.
- If static validation cannot run locally, name the unavailable tool and provide the PR check or `workflow_dispatch` evidence that covered workflow parsing.

## Recommendation

- Clear to proceed. The implementation path is small and testable, with the main risk concentrated in proving the Node.js warning source and avoiding accidental release publication.
