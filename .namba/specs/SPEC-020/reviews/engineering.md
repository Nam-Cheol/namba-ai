# Engineering Review

- Status: approved
- Last Reviewed: 2026-04-01
- Reviewer: Codex
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- Separating `namba harness` into a top-level command is technically sound. The command represents a distinct user intent from ordinary feature planning, so keeping it out of `namba plan` reduces parser ambiguity and keeps the default plan path stable.
- The command shape is viable if implementation reuses existing SPEC package scaffolding rather than inventing a second planning artifact format. The current acceptance contract correctly keeps output under the existing `SPEC-XXX` workflow instead of introducing a parallel package system.
- Reusing the read-only help contract from `SPEC-019` is the right engineering boundary. Help parsing and accidental write prevention should live in one shared command-parsing layer; duplicating that logic inside `namba harness` would create drift quickly.
- The non-portable primitive exclusion is explicit enough for implementation. Banning `.claude/*`, `TeamCreate`, `SendMessage`, `TaskCreate`, and a hardcoded `model: "opus"` keeps the feature aligned with current Codex-facing assets and avoids recreating the Claude adapter inside the Codex adapter.

## Decisions

- Proceed with a top-level `namba harness` command rather than a `namba plan --template harness` flag.
- Keep `namba harness` on the existing `SPEC-XXX` storage model and review scaffold flow so downstream Namba commands can continue to reason about one planning artifact shape.
- Treat `SPEC-019` as a sequencing dependency for help/read-only safety. `SPEC-020` can define the contract, but parser hardening should be implemented once in the shared planning command path.

## Follow-ups

- During implementation, centralize command parsing and help handling so `plan`, `fix`, and `harness` do not grow separate flag validators.
- Decide early whether `namba harness` needs only a description argument in v1 or whether it also needs a mode/template flag surface. The current SPEC is cleaner if v1 stays minimal.
- Add regression tests that assert `namba harness` emits Codex-native guidance only and that generated docs explain the boundary between `namba plan` and `namba harness` with concrete examples.

## Recommendation

- Advisory recommendation: approved. Proceed after `SPEC-019` lands or implement both in sequence on a shared parser foundation so help/read-only safety remains centralized.
