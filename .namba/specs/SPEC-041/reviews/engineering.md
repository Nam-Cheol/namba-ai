# Engineering Review

- Status: clear with follow-ups
- Last Reviewed: 2026-05-01
- Reviewer: Codex as `namba-planner`
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

Re-evaluate the prior advisory gaps after aggregate clarifications and confirm the SPEC is implementation-ready without widening ownership or validation scope.

## Findings

- The prior `codex exec --json` ownership gap is closed. `spec.md`, `plan.md`, and `acceptance.md` now consistently limit JSON-field tolerance work to a named Namba-owned consumer if one is found during implementation; otherwise the requirement narrows to a compatibility note rather than speculative parser coverage.
- The prior `[agents] max_threads = 5` evidence gap is closed. The SPEC now requires a concrete proof path: validate the repo-managed `.codex/config.toml` shape against the target Codex schema or the local `0.128.0` config/help surface, instead of treating template rendering or prose as sufficient evidence.
- The prior hook/plugin/MCP scope gap is closed. Scope language now stays on Namba-owned boundaries: stable hook event names, observation payload shape, MCP approval persistence assumptions, repo-managed preset/cache/path boundaries, and plugin-bundled hook guidance only where Namba documents or integrates.
- The concurrency boundary remains correctly separated and explicitly named: Codex same-workspace subagent threads stay at `[agents] max_threads = 5`, while Namba worktree fan-out stays at `max_parallel_workers: 3` unless separately accepted.
- The remaining engineering work is execution depth, not SPEC ambiguity. Implementation still needs to turn the proof steps into fixtures, docs, and validation commands, especially for the lighter-documentation `0.126.0` and `0.127.0` range.

## Decisions

- Treat the three previously advisory items as resolved at the SPEC level.
- Keep the first implementation slice focused on capability fixtures/tests, config-schema or local-surface proof for `[agents]`, and bounded hook/MCP guidance updates.
- Keep worktree parallelism changes out of scope unless separately accepted during implementation.

## Follow-ups

- During implementation, explicitly record the `codex exec --json` ownership outcome: either name the consumer and test it, or document that no Namba-owned consumer exists.
- Make the `[agents] max_threads = 5` proof durable in code or test artifacts, not only in reviewer notes.
- Preserve the narrowed hook/plugin/MCP boundary in commit scope so this SPEC does not turn into upstream marketplace or external-agent certification work.

## Recommendation

Proceed. The earlier engineering advisory gaps are closed by the current clarifications, and SPEC-041 is now implementation-ready with normal follow-up discipline rather than additional pre-implementation rework.
