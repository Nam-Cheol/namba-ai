# Engineering Review

- Status: clear
- Last Reviewed: 2026-04-22
- Reviewer: codex
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The previous architecture blocker is resolved. The SPEC now defines an explicit analysis-index boundary with normalized membership, candidate lookup scope, and bounded lazy reads, which is specific enough to keep helper refactors on one reusable contract instead of ad hoc caching.
- The previous sync-boundary blocker is resolved. The revised package now states that sync stages README, analysis, readiness, and support outputs before mutation and keeps one sync-owned manifest session, while still requiring explicit cleanup ownership for README-managed and project-analysis-managed paths.
- The previous measurement blocker is resolved. Deterministic proof for repeated-read collapse, readiness batching, manifest-session churn reduction, and shared support-context reuse is now part of the acceptance contract, with benchmarks correctly positioned as secondary evidence.

## Decisions

- Keep this as a command-local optimization only: no persistent cache and no whole-repo content buffering.
- Require staged output assembly to preserve current cleanup and no-op guarantees before any batching refactor is considered complete.
- Treat measurement as a correctness surface: implementation should prove reduced I/O through deterministic tests or counters, then use benchmarks as secondary evidence.

## Follow-ups

- During implementation, keep the manifest-session abstraction narrow so batching does not accidentally weaken stale-output removal or no-op modtime guarantees.
- Add the deterministic I/O assertions early in the TDD sequence; otherwise the refactor can still appear green while leaving hidden duplicate discovery paths behind.
- Prefer proving improvement on both axes the SPEC names: large inventory reuse and many-SPEC readiness churn.

## Recommendation

- Advisory recommendation: clear. The implementation boundaries are now specific enough to execute safely.
