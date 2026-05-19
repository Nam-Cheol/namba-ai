# Engineering Review

- Status: approved after revision
- Last Reviewed: 2026-05-19
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- `.codex/hooks/*` and `.codex/hooks.json` are generated surfaces owned by `internal/namba/templates.go` and tracked in `.namba/manifest.json`; implementation must update templates before generated outputs.
- The SPEC now fixes the Codex-facing output schemas for `PreToolUse`, `PermissionRequest`, non-deny approval notes, prompt guidance, and no-output cases.
- CI wiring is explicit: add a Python `unittest` hook-test step in addition to Go test, vet, and formatting.
- Validation sequencing is explicit: run `namba sync` after implementation changes and rerun validation afterward.

## Decisions

- Use black-box subprocess tests against `.codex/hooks/namba_codex_guard.py` with JSON stdin and zero-or-more JSON lines on stdout.
- Preserve current schemas: `permissionDecision` for `PreToolUse`, nested `decision.behavior` for `PermissionRequest`, and top-level `systemMessage` for non-deny risk notes.
- Keep dangerous-command policy limited to the fixed regression corpus in this SPEC.

## Follow-ups

- Later work may add command-parser edge coverage or branch slug ergonomics, but neither belongs in SPEC-047.
- During implementation, confirm generated hook output matches the template after `namba regen`.

## Recommendation

- Approved for implementation.
