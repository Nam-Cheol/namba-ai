# SPEC-025

## Problem

NambaAI currently gives users clear entry points for feature planning, harness planning, bug repair, execution, sync, and PR handoff, but it does not provide a first-class way to create a new repo-local skill, a new project-scoped custom agent, or both from inside the Codex-native workflow.

That gap is visible in the current repository surface:

- Repo-local skills are generated into `.agents/skills/*/SKILL.md`, and project-scoped custom agents are generated into `.codex/agents/*.toml` plus readable `.md` mirrors by `internal/namba/codex.go`.
- The generated routing guidance in `internal/namba/templates.go` currently exposes `$namba-help`, `$namba-plan`, `$namba-harness`, `$namba-fix`, `$namba-run`, and related review skills, but there is no `$namba-create` entry point for skill/agent creation.
- `internal/namba/update_command.go` treats the whole `.agents/skills/` and `.codex/agents/` trees as regen-managed output, so naive direct generation into those paths can be deleted or overwritten by `namba regen`.
- `internal/namba/namba.go` exposes `runPlan`, `runHarness`, and `runFix`, but there is no existing intent-to-artifact builder that keeps clarifying user intent until the output target is unambiguous.
- `.namba/config/sections/codex.yaml` opts into `context7`, `sequential-thinking`, and `playwright`, but there is no current planning or generation contract that requires those tools to be used meaningfully during skill/agent creation.
- `.codex/config.toml` currently sets `[agents] max_threads = 3`, which is too low for the requested five-role same-workspace creation and review flow under repo defaults.

The user problem is therefore not "add one more file generator." It is "add a Codex-native creation workflow that keeps asking until nothing important is ambiguous, routes the request to skill vs agent vs both, uses repo-managed MCPs and multi-agent analysis intentionally, and writes durable artifacts without violating Namba's source-of-truth or regen rules."

## Goal

Add a new skill-first `namba-create` workflow that stages clarification before generation, chooses `skill`, `agent`, or `both` deterministically, uses repo-managed MCPs and independent role analysis as part of the flow, raises the repo-managed same-workspace agent thread default from `3` to `5` for `agent_mode: multi`, and writes validated artifacts without conflicting with Namba's managed scaffold contract.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Primary user-facing entrypoint for phase 1: `$namba-create`
- Existing authoritative skill and agent surfaces:
  - `.agents/skills/*/SKILL.md`
  - `.codex/agents/*.toml`
  - `.codex/agents/*.md`
- Current scaffold source-of-truth:
  - `internal/namba/codex.go` defines generated skill and agent outputs.
  - `internal/namba/templates.go` defines the command-entry skill text and routing guidance.
  - `internal/namba/update_command.go` regenerates those surfaces and currently treats the full skill/agent trees as managed.
- Existing regression anchors:
  - `internal/namba/spec_command_test.go` already proves read-only help and SPEC scaffold behavior.
  - `internal/namba/update_command_test.go` already proves regen/update surface generation and session-refresh signaling.
- Existing repo-managed MCP presets:
  - `.namba/config/sections/codex.yaml` enables `context7,sequential-thinking,playwright`.
- Existing concurrency constraint:
  - `.codex/config.toml` sets `[agents] max_threads = 3`.
  - `.namba/config/sections/workflow.yaml` separately sets `max_parallel_workers = 3` for worktree fan-out and should remain a distinct control surface.
- Planning evidence gathered for this SPEC:
  - `namba-planner` analyzed integration points and recommended a skill-first scope.
  - `namba-security-engineer` defined ambiguity, path-safety, instruction-safety, and ownership guardrails.
  - `namba-test-engineer` defined a staged verification strategy centered on clarification-state and branch-matrix tests.
  - `namba-reviewer` identified scope risks around CLI expansion, regen ownership, and preview-before-write behavior.
  - `explorer` mapped the concrete files and renderer layers that must be touched.
- External evidence:
  - A `context7` review of Cobra guidance reinforced that a future `namba create` CLI would need its own explicit help and argument-validation contract, which supports keeping the first delivery skill-first instead of expanding CLI scope immediately.

## Desired Outcome

- `$namba-create` exists as a generated repo-local command-entry skill and is surfaced in the generated routing/help documentation alongside the other Namba entry points.
- Phase 1 is explicitly skill-first: the main delivery is the `$namba-create` skill surface, not a new `namba create` Go CLI command.
- The flow behaves as a staged generator rather than a direct file writer:
  - `unresolved`: ambiguity remains, so the system keeps asking and does not write files.
  - `narrowed`: the candidate output type is visible, but unresolved items are still tracked.
  - `confirmed`: all required decisions are explicit, previewed, and ready for generation.
- The interaction loop makes remaining unknowns visible each turn so the user can see convergence rather than an opaque repeated interview.
- If the user explicitly says `skill`, `agent`, or `both`, that directive outranks heuristic classification.
- If the user does not specify the target, the workflow explains why the request maps to `skill`, `agent`, or `both` before it writes anything.
- Before generation, the workflow presents a non-mutating resolution summary that includes:
  - chosen output type
  - slug/name
  - intended file paths
  - validation plan
  - whether session refresh messaging will likely be required
- Generated skill artifacts land only at `.agents/skills/<slug>/SKILL.md`.
- Generated agent artifacts land only at `.codex/agents/<slug>.toml` plus `.codex/agents/<slug>.md`.
- The repo-managed same-workspace agent thread default for `agent_mode: multi` is raised from `3` to `5`, so a five-role creation or review flow is supported under repo defaults.
- The implementation records at least five independent analysis or validation roles for the creation workflow. Under the repo default it can execute those roles directly in same-workspace parallel form, while still degrading safely if an external override lowers the effective limit.
- User-authored artifacts created through `namba-create` survive `namba regen`; regen must keep managing built-in Namba scaffolds without silently deleting user-created skill/agent outputs.
- Instruction-surface mutations continue to emit explicit session-refresh messaging when a fresh Codex session is needed.
- Validation and regression coverage make the workflow safe enough to implement later without re-litigating the interaction contract.

## Target User

- Users who want to create a repo-specific skill without manually drafting `SKILL.md` structure and trigger rules.
- Users who want to create a project-scoped custom agent without manually drafting both the `.toml` definition and the readable `.md` mirror.
- Users migrating MoAI-style "skill-builder" or "agent-builder" expectations into the Codex-native Namba model.
- Maintainers who need the generation flow to be deterministic, reviewable, safe to probe, and compatible with `namba regen`.

## Scope

- Add `$namba-create` as a generated repo-local command-entry skill.
- Update generated routing/help/readme/AGENTS guidance so users can discover when to use `$namba-create` versus `$namba-plan` or `$namba-harness`.
- Define a clarification state machine for the creation flow, including explicit unresolved-item tracking and a preview/confirm gate before writes.
- Define a routing matrix for `skill`, `agent`, and `both`, with explicit-user-intent precedence over classifier inference.
- Implement safe artifact writers for skill and agent outputs, including slug normalization, allowlisted paths, overwrite confirmation, and agent mirror consistency checks.
- Introduce a durable ownership model so built-in Namba-managed scaffolds and user-authored `namba-create` outputs can coexist without `namba regen` deleting user-created artifacts.
- Raise the generated repo-local Codex baseline from `[agents] max_threads = 3` to `5` when `agent_mode: multi`.
- Require meaningful use of repo-managed MCP presets where appropriate:
  - `sequential-thinking` for decomposition and clarification planning
  - `context7` for targeted external library or framework guidance when needed
  - `playwright` only when browser-verified flows are relevant
- Define the minimum multi-agent contract for the flow as five independent role outputs, with the repo-managed same-workspace default raised to support five concurrent roles under normal multi-agent settings.
- Add validation, session-refresh handling, and regression tests for the new workflow.

## Non-Goals

- Do not add a new `namba create` Go CLI command in this first delivery.
- Do not redesign `namba plan`, `namba harness`, `namba fix`, or the existing plan-review skill family beyond the routing/help updates needed to introduce `$namba-create`.
- Do not create a second artifact model outside the existing skill surface, custom-agent surface, and `.namba/` workflow state.
- Do not allow raw user prose to become durable instructions without normalization, policy checks, and artifact-shape validation.
- Do not automatically change worktree `max_parallel_workers` as part of this SPEC; same-workspace agent thread defaults and worktree fan-out limits remain separate controls.

## Design Constraints

- Keep `.namba/` as the source of truth for workflow state and generated planning artifacts.
- Treat `$namba-create` as a staged generator: no file writes before ambiguity is resolved and the plan is confirmed.
- Keep explicit user directives authoritative over heuristic classification.
- Keep skill outputs under `.agents/skills/<slug>/SKILL.md` and agent outputs under `.codex/agents/<slug>.toml` plus `.codex/agents/<slug>.md`.
- Constrain writes to repo-configured skill and agent roots; reject path traversal, invalid slugs, silent overwrites, and incomplete agent mirror pairs.
- Preserve a clear separation between Namba-managed built-in scaffolds and user-authored `namba-create` outputs so `namba regen` remains safe and predictable.
- Keep concurrency ownership separate: `.codex/config.toml [agents].max_threads` governs same-workspace Codex child-session or subagent concurrency, while `.namba/config/sections/workflow.yaml` governs worktree `--parallel` worker caps.
- Keep preview and validation steps non-optional; generation should be deterministic only after the target, paths, and checks are explicit.
- Record at least five independent role contributions across planning or verification, with the repo-managed multi-agent default raised to permit five same-workspace roles under normal settings.
- Surface session-refresh requirements explicitly when instruction-surface files change.
