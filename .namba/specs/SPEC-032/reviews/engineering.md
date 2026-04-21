# Engineering Review

- Status: clear
- Last Reviewed: 2026-04-21
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The transport split is now implementation-aligned: SPEC routes persist `.namba/specs/<SPEC>/harness-request.json`, while direct routes reuse the existing JSON create preview/apply flow without inventing a SPEC package.
- Validator ownership is now explicit enough for v1. The harness-evidence validator is attached to readiness refresh and `namba sync`, remains advisory, and stays out of `namba run` execution validation.
- Legacy readiness compatibility is adequately bounded. Older SPECs without `harness-request.json` stay on the existing path, and prior evidence such as `extraction-map.md` remains valid.
- V1 review/runtime boundaries are correctly constrained to the existing `product`, `engineering`, and `design` tracks instead of speculating about new runtime surfaces.

## Decisions

- Proceed with one shared Go model and JSON transport across both planning and direct-create paths.
- Treat `harness-request.json` presence as the clean v1 opt-in boundary for harness-classified SPEC behavior.
- Keep the harness-evidence validator advisory in readiness and `namba sync` only for this slice.
- Keep `required_reviews` scoped to the current review runtime in v1.

## Follow-ups

- Keep the code and docs wording consistent that `harness-map.md` is conditionally required, not globally optional or globally mandatory.
- Add regression coverage for legacy SPEC readiness refresh so `namba sync` cannot reinterpret historical packages as typed harness SPECs.
- Keep classifier tests explicit around ambiguous artifact-versus-contract cases so `$namba-create` cannot win on unresolved core-contract changes.

## Recommendation

- Clear to proceed. The contract is now specific enough on JSON transport, validator hook placement, legacy readiness behavior, and v1 review/runtime scope to land safely.
