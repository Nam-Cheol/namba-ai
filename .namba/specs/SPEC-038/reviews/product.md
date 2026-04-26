# Product Review

- Status: clear
- Last Reviewed: 2026-04-26
- Reviewer: Codex (`namba-product-manager`)
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Verify the revised SPEC artifacts now make `$namba-help` and `$namba-coach` distinct.
- Confirm coverage for both plain-language non-developer requests and wrong-command developer requests.
- Confirm `$namba-create` versus `namba harness` routing is explicit.
- Confirm post-implementation handoffs are stage-specific instead of bundled.

## Verification

- The `$namba-help` versus `$namba-coach` boundary is now explicit in `spec.md`, `contract.md`, and `acceptance.md`.
  - `$namba-help` is framed as usage, command semantics, and docs guidance.
  - `$namba-coach` is framed as current-goal clarification, wrong-command correction, and next-step handoff.
- Non-developer and developer coverage is now present in acceptance.
  - Non-developer coverage: a plain-language request with no Namba command must be clarified only enough to route.
  - Developer coverage: a clearly wrong command choice must be corrected rather than executed as-is.
- `$namba-create` versus `namba harness` is now separated cleanly.
  - Direct repo-local skill or custom-agent creation routes to `$namba-create`.
  - Reusable skill, agent, workflow, or orchestration planning routes to `namba harness "<description>"`.
  - Namba core managed-surface changes remain distinct and route to `namba plan "<description>"`.
- Stage-specific handoffs are now explicit in `spec.md`, `contract.md`, `acceptance.md`, and `eval-plan.md`.
  - Implementation finished -> `namba sync`
  - Review handoff ready -> `namba pr "<Korean title>"`
  - Approved PR ready to merge -> `namba land`

## Findings

- No blocking product issues remain in the revised source artifacts.
- The prior follow-ups are reflected in the source of truth and are specific enough to guide implementation and evaluation.
- The revised wording keeps `$namba-coach` product scope narrow and useful without collapsing into `$namba-help`, `$namba-create`, or planning commands.

## Residual Risk

- The remaining risk is execution quality rather than product definition: generated docs and skill text must preserve the same boundary language consistently across surfaces.
- That risk is already addressed appropriately by the updated eval plan and documentation/scaffold acceptance coverage.

## Recommendation

- Proceed.
- Product review is clear. The revised SPEC package now resolves the earlier ambiguity around command boundaries, user-shape coverage, and stage-specific workflow handoffs.
