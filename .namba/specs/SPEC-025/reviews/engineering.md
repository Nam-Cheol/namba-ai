# Engineering Review

- Status: clear
- Last Reviewed: 2026-04-08
- Reviewer: codex
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The integration points are well-bounded. `internal/namba/codex.go`, `internal/namba/templates.go`, `internal/namba/update_command.go`, and the existing spec/update regression tests already define where discovery, generation, and regen behavior live.
- The main hidden risk is ownership. Because `namba regen` currently treats the full skill and agent trees as managed output, the implementation must add a durable distinction between built-in managed scaffolds and user-authored create outputs.
- The staged-generator contract is the right architecture: unresolved -> narrowed -> confirmed plus preview/confirm gating directly addresses the risk of premature file writes.
- The current repository config caps concurrent same-workspace agent threads at three, so the spec is right to include an explicit baseline change to `5` instead of only relying on batching.
- That change must stay scoped to `.codex/config.toml [agents].max_threads`; it should not implicitly alter `.namba/config/sections/workflow.yaml` worktree parallel caps.
- Deferring a real `namba create` CLI command to a follow-up keeps the first implementation inside the existing Codex-native skill surface instead of reopening command dispatch, help, and CLI parsing semantics.

## Decisions

- Keep phase 1 skill-first.
- Raise the generated repo-managed same-workspace `max_threads` default from `3` to `5` for `agent_mode: multi`.
- Add an explicit ownership model for user-authored create outputs so `namba regen` preserves them.
- Treat preview-before-write, path allowlisting, overwrite confirmation, and agent mirror consistency as first-class engineering requirements rather than optional polish.
- Reuse the existing test style from `internal/namba/spec_command_test.go` and `internal/namba/update_command_test.go` instead of inventing a separate test philosophy.

## Follow-ups

- Validate early whether ownership should be expressed through manifest metadata, a managed flag, or an equivalent durable mechanism; do not leave that decision implicit.
- Add at least one end-to-end scenario that proves no file writes occur before confirmation and that a later `namba regen` does not delete user-authored outputs.
- Update generation tests and docs together so `max_threads = 5` becomes an intentional repo contract rather than an undocumented template tweak.

## Recommendation

- Clear to proceed from an engineering perspective. The architecture, risk areas, and test strategy are concrete enough for implementation.
