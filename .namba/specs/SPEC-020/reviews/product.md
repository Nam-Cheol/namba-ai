# Product Review

- Status: approved
- Last Reviewed: 2026-04-01
- Reviewer: Codex
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- A dedicated `namba harness` surface is easier to understand than hiding the feature behind `namba plan --template harness`. Users think of reusable agent/skill-system design as a different job from ordinary feature planning, so the split improves discoverability.
- Keeping the output on the existing `SPEC-XXX` flow is product-coherent. Users do not need to learn a second planning artifact model just because the planned work is harness-oriented.
- Reusing `SPEC-019` for `namba plan --help` and shared read-only help behavior is the right product boundary. Users benefit more from one consistent “help never mutates state” rule than from separate command-specific caveats.
- The biggest product risk is ambiguity between `namba project`, `namba plan`, and `namba harness`. The docs and help text need concrete examples so users can tell “analyze the repo”, “plan a feature”, and “design a reusable harness” apart immediately.

## Decisions

- Proceed with `namba harness` as a top-level command.
- Keep v1 narrow: accept a description, create a reviewable SPEC package, and avoid adding secondary options until the base intent is proven.
- Treat `namba plan` as the default feature-planning path and `namba harness` as the reusable automation/planning path.

## Follow-ups

- Add one short example for each of `namba project`, `namba plan`, and `namba harness` in help text and generated docs so command intent is obvious at a glance.
- During implementation, keep the command naming plain. The term “harness” is acceptable only if the help text immediately explains it in user-facing language such as agent/skill system design.
- Verify that the generated review scaffolds and README sections explain why `SPEC-019` owns read-only help safety while `SPEC-020` owns the new command surface.

## Recommendation

- Advisory recommendation: approved. Proceed, but treat discoverability and example-led docs as part of the core product contract rather than optional polish.
