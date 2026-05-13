# Engineering Review

- Status: clear
- Last Reviewed: 2026-05-13
- Reviewer: Codex acting as `namba-planner`
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The implementation surface is well-bounded. The change maps cleanly onto `internal/namba/frontend_brief.go` for canonical contract parsing/blocking, `internal/namba/spec_review.go` for design/readiness summaries, `internal/namba/templates.go` for generated guidance, and the existing frontend/execution/template test suites for regression coverage.
- The main architecture risk is parser coupling. The current frontend brief contract relies on a small fixed-label header plus deterministic body sections, so adding many new top-level labels would create unnecessary compatibility churn; the negative-first contract should stay structured and parser-visible without widening the header surface more than necessary.
- Canonical ownership is coherent as long as `frontend-brief.md` remains the only machine-authoritative source for negative-first state. `reviews/design.md` and `reviews/readiness.md` should summarize and cross-check it, not introduce parallel truth that can drift.
- The execution model needs two distinct enforcement phases: pre-dispatch blocking when the Do-Not Design Contract is missing or insufficient, and post-implementation blocking when the required violation-check evidence is missing or explicitly reports a banned fallback. Keeping those phases separate will make remediation messages and tests much more stable.
- False-positive risk is real if the implementation treats generic tokens like `card`, `grid`, or `hero` as global violations. The SPEC is on the right track by requiring context-specific banned patterns, detection hints, allowed replacements, and exception paths; enforcement should be anchored to that contract, not naive keyword bans.
- Compatibility remains the biggest open edge. Historical no-brief SPECs are already called out, but older `frontend-major` briefs created before this contract will need an explicit, test-backed policy so reruns and queue operations do not start failing by accident.
- `SPEC-044` itself is correctly scaffolded as `frontend-minor` in `frontend-brief.md`, which means this implementation must preserve the advisory path for the current harness package while hardening future `frontend-major` work.

## Decisions

- Keep `frontend-brief.md` as the canonical parser-visible contract, with design review and readiness acting as summaries plus mismatch detectors.
- Implement boundary-first in this order: failing tests, contract/parser model, scaffold generation, readiness/design-review rendering, run-time preflight blocking, post-run violation-check handling, then template regeneration and sync artifacts.
- Represent the negative-first contract in a deterministic structure that is explicit enough for readiness and run-time checks without turning the header parser into a brittle free-form document validator.
- Require the post-implementation "Do-Not Design Violation Check" to have a stable output shape that cites banned pattern, evidence, and exception-path support when applicable; do not rely on vague narrative success claims.
- Preserve intentional passthrough behavior for `frontend-minor`, non-frontend, and legacy no-brief scenarios, and add explicit coverage for any changed treatment of older `frontend-major` briefs.

## Follow-ups

- Decide and document the compatibility policy for pre-SPEC-044 `frontend-major` briefs: grandfather, migrate on rerun, or block intentionally with a deterministic remediation path.
- Define the exact parser-visible shape for the new contract and the exact runner-result shape for the post-implementation violation check before touching templates, so execution and template tests share one contract.
- Add regression cases that prove `SPEC-044` remains advisory as a harness-side `frontend-minor` package while future `frontend-major` packages take the full negative-first gate.
- Keep generated guidance compact and contract-oriented; test for behaviorally important strings, not incidental prose, to avoid brittle regen churn.

## Recommendation

- Clear for implementation. The SPEC is technically coherent, but historical `frontend-major` compatibility must be treated as an intentional, test-backed decision during execution.
