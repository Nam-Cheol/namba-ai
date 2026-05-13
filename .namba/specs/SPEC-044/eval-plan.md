# SPEC-044 Evaluation Plan

## Evaluation Goals

- Prove that `frontend-major` work cannot pass with only the original five-gate evidence and approved design review.
- Prove that the Do-Not Design Contract is present, parseable enough for readiness, and actionable enough for architecture and implementation.
- Prove that default anti-patterns are adapted through context-specific bans, replacements, and exceptions rather than applied as blanket taste rules.
- Prove that post-implementation violation checks can block banned generic UI fallback reliance.
- Prove that `frontend-minor`, non-frontend, and legacy no-brief work remain intentionally lightweight.

## Regression Tests

- Scaffold generation:
  - `frontend-major` plan, harness, and fix-plan scaffolds include the Do-Not Design Contract.
  - `frontend-minor` scaffolds do not require the full negative-first contract.
  - Existing non-frontend planning still avoids frontend gate scaffolding unless classification explicitly applies.

- Parser and readiness:
  - missing contract fields produce missing or insufficient negative-first state
  - a complete five-gate header plus approved design review still blocks when negative-first evidence is absent
  - context-specific bans render with allowed replacements and detection hints
  - readiness reports failed or unresolved negative-first state separately from product, engineering, and design review status

- Execution:
  - pre-execution blocks before runner dispatch when the Do-Not Design Contract is incomplete
  - runner output that reports a banned generic fallback fails the run
  - runner output that uses an exception path passes only when it cites the contract evidence
  - blocked output names the banned pattern and remediation path

- Templates and generated surfaces:
  - `$namba-run` explains the negative-first gate and post-implementation check
  - `$namba-plan-design-review` checks the default library, replacements, visual grammar, and generic-section proof
  - `namba-designer`, `namba-frontend-architect`, and `namba-frontend-implementer` carry the same handoff contract
  - README and workflow guides mention the new gate without becoming long procedural manuals

## Fixture Scenarios

- Major dashboard restructure with complete old gates but no Do-Not Design Contract: blocked.
- Landing page direction that bans stock SaaS hero fallback and provides a workflow-led replacement: allowed after all contract fields are complete.
- Dashboard implementation result that admits it used the banned KPI-card row fallback: failed post-implementation violation check.
- Component-scale redesign where most-generic-section proof is marked not applicable with a valid scope rationale: allowed.
- Frontend-minor spacing fix with not-applicable gate fields: advisory passthrough.
- Historical SPEC without `frontend-brief.md`: no speculative frontend block.

## Validation Commands

Run after implementation:

```bash
namba regen
namba sync
gofmt -l "cmd" "internal" "namba_test.go"
GOCACHE=/tmp/namba-go-cache go test ./...
GOCACHE=/tmp/namba-go-cache go vet ./...
git diff --check
```

## Manual Review Checks

- Verify the default anti-pattern library is specific enough to prevent generic fallback but not so broad that it bans valid primitives.
- Verify replacement patterns are actionable for architecture and implementation.
- Verify generated role guidance stays compact.
- Verify post-implementation violation evidence is visible in run/sync/PR artifacts.

