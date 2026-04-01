# Engineering Review

- Status: approved
- Last Reviewed: 2026-04-01
- Reviewer: Codex
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- No blocking engineering gaps remain. The revised SPEC now gives the parser one explicit shape for bug work: `namba fix --command plan` for bugfix SPEC creation, `namba fix --command run` for direct repair, and plain `namba fix` as the default alias for the run path.
- The runtime contract is concrete enough to implement safely. Direct `namba fix` requires repo context plus an issue description, works in the current workspace, adds targeted regression coverage, runs configured validation, syncs artifacts, and avoids implicit `.namba/specs/` mutation.
- The documentation scope is now technically well framed. The authoritative change surface is the renderer and template layer such as `internal/namba/templates.go` and `internal/namba/readme.go`, not just the generated files, which is the right constraint to prevent contract drift.
- The main implementation risk is migration consistency, not architecture ambiguity. Help text, error text, skill descriptions, README bundles, and generated docs all need to land in the same change so old and new meanings do not coexist.

## Decisions

- Implement `fix --command` through one explicit parser path and keep the default `namba fix "..."` semantics equivalent to `namba fix --command run "..."`; do not add further implicit branching beyond that default.
- Preserve the existing bugfix SPEC scaffolding logic under `namba fix --command plan` instead of re-inventing a second planning mechanism.
- Land CLI parsing, regression coverage, renderer-source updates, and generated artifact refresh in one implementation slice so the public contract changes atomically.
- Keep unknown `--command` values, missing issue descriptions, and missing repo-context failures actionable and deterministic.

## Follow-ups

- Add parser tests for default `fix` run behavior, explicit `--command run`, explicit `--command plan`, unknown `--command`, and read-only help flows.
- Update generated guidance from source renderers first, then run `namba regen` and `namba sync` so repo-managed artifacts stay aligned with the new parser contract.
- Keep regression coverage on the historical failure mode: help or probe invocations must never create a `SPEC-XXX` package.

## Recommendation

- Advisory recommendation: approved. Proceed with implementation using the `fix --command` contract captured in `SPEC-019`.
