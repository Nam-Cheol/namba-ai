# Engineering Review

- Status: cleared
- Last Reviewed: 2026-05-19
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation
  strategy before execution starts.

## Findings

- Route drift concern is cleared. `route_cases.json` must assert
  `expected_command`, `expected_harness_request_kind_or_none`, and
  `expected_sidecar_persisted` rather than a synthetic classifier.
- Evidence manifest ambiguity is cleared. Raw-schema validation and
  builder-normalization validation must be separate so struct defaults do not
  hide missing persisted fields.
- Hook subprocess ownership is clear. Prompt refinement and guardrail evals stay
  in Python only when they need the real `.codex/hooks/namba_codex_guard.py`
  subprocess.
- CI direction is appropriate: extend the existing unittest, Go test, vet, and
  formatting chain instead of creating a separate job unless necessary.

## Decisions

- Go owns route, evidence, PR review, fixture schema, and diagnostic formatting
  evals.
- Python owns prompt refinement and shell guardrail subprocess evals only when
  the hook is the source of truth.
- PR review fixtures should use runtime seams around `runPR` and
  `ensureReviewComment`, not a config-only simulation.

## Follow-ups

- Implement raw manifest validation at raw JSON/map level before struct
  normalization.
- Keep any route helper thin; it should wrap existing command-selection behavior
  rather than become a new classifier.
- Use common diagnostic formatting in Go and equivalent fixture context in
  Python so failures name the case, input/command/manifest, expected, actual,
  and rationale.

## Recommendation

- Proceed to implementation.
