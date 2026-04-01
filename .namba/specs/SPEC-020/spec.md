# SPEC-020

## Problem

NambaAI already ships a Codex-native execution surface with repo-local skills, project-scoped custom agents, same-workspace multi-agent execution, and worktree fan-out/fan-in.

What it does not yet provide is a first-class way to plan the harness itself when a change request is really about agent topology, skill boundaries, trigger quality, or evaluation strategy.

Local evidence in the current repository:

- `docs/workflow-guide.md` and `.namba/project/product.md` define how Namba executes work once a SPEC exists, but they do not provide a dedicated planning shape for designing new skill/agent systems.
- `.agents/skills/namba-plan/SKILL.md` creates a normal feature SPEC package, yet it does not guide harness-heavy planning toward explicit agent-boundary, trigger, and evaluation decisions.
- `docs/moai-adk-codex-migration-analysis.md` already records that Claude Agent Teams semantics are not directly portable to Codex and should be redesigned as orchestrated fan-out/fan-in instead of copied literally.

External evidence from `revfactory/harness` shows portable value above the runtime layer:

- Harness is a meta-skill that designs domain-specific agent teams and skills rather than only executing one task at a time.
- It uses a reusable architecture vocabulary such as pipeline, fan-out/fan-in, expert pool, producer-reviewer, supervisor, and hierarchical delegation.
- It treats skill descriptions, progressive disclosure, and evaluation loops as first-class design concerns rather than afterthoughts.
- It includes concrete evaluation ideas such as should-trigger versus should-not-trigger queries and with-skill versus baseline comparisons.

However, direct code or workflow adoption is not safe because Harness is built around Claude-specific primitives:

- `.claude/agents/` and `.claude/skills/` as the generated output surface
- `TeamCreate`, `SendMessage`, and `TaskCreate` as the core live-team protocol
- a required `model: "opus"` assumption inside generated agent instructions

Current official Codex docs instead center customization on `AGENTS.md`, repo skills under `.agents/skills`, and project-scoped custom agents under `.codex/agents/*.toml`, with subagents spawned explicitly rather than via a Claude Team API.

## Goal

Bring the portable Harness concepts into NambaAI by adding a dedicated `namba harness` Codex-native planning surface that helps users design skill/agent-heavy work as a reviewable SPEC package, without importing Claude-only runtime primitives.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Proposed planning surface: `namba harness "<description>"`
- Affected area: `internal/namba/namba.go`, `internal/namba/templates.go`, `.agents/skills/namba-plan/SKILL.md`, `.agents/skills/namba/SKILL.md`, `internal/namba/readme.go`, generated README/workflow docs, and any new harness-planning references or template assets under `.namba/` or `.agents/skills/`
- Adjacent dependency: `namba plan --help` and other read-only help flows are already tracked by `SPEC-019`; this SPEC should reuse that contract rather than redefining it.

## Desired Outcome

- `namba plan "<description>"` keeps its current default feature-planning behavior.
- `namba harness "<description>"` creates the next sequential `SPEC-XXX` package with the normal review scaffolds plus a harness-oriented problem frame.
- `namba harness --help` and `namba plan --help` both behave as read-only help surfaces with no `SPEC-XXX` writes, with the shared help-safety contract sourced from `SPEC-019`.
- Harness-mode SPECs explicitly capture the intended Codex execution topology using Namba-native terms such as standalone, one specialist, same-workspace multi-agent, or worktree fan-out/fan-in.
- Harness-mode SPECs include an explicit agent and skill boundary plan that maps work to `.agents/skills/*`, built-in subagents, and `.codex/agents/*.toml` custom agents rather than `.claude/*`.
- Harness-mode SPECs include progressive-disclosure guidance for any new reusable skill package: metadata trigger, `SKILL.md` body, and optional `references/`, `scripts/`, or `assets/`.
- Harness-mode SPECs include a trigger and evaluation strategy covering should-trigger versus should-not-trigger checks, with-skill versus baseline comparisons where meaningful, and assertion/timing capture guidance for measurable workflows.
- Generated docs and skill guidance explain which Harness concepts were adopted and which Claude-only primitives were intentionally rejected for Codex compatibility.
- No generated Namba artifact in this flow emits `.claude/agents`, `.claude/skills`, `TeamCreate`, `SendMessage`, `TaskCreate`, or a mandatory `model: "opus"` requirement as if they were part of Namba's Codex contract.

## Scope

- Add a dedicated `namba harness` planning command surface.
- Define harness-template SPEC scaffolding for `spec.md`, `plan.md`, and `acceptance.md`.
- Teach the generated planning guidance to capture:
  - Codex-native execution topology
  - proposed agent and skill boundaries
  - progressive-disclosure asset layout
  - trigger-writing guidance
  - evaluation strategy for reusable skills and agents
- Add shared authoring guidance or reference material for harness-template planning so the concepts are reusable across sessions instead of living only in one SPEC.
- Update repo-local skill descriptions and generated README/workflow docs so users understand when to stay on default `namba plan` versus when to use `namba harness`.
- Add regression coverage for CLI parsing, scaffold generation, non-portable primitive exclusion, and generated guidance text where stable assertions are appropriate.

## Non-Goals

- Do not copy Claude Team lifecycle semantics into NambaAI.
- Do not generate `.claude/*` assets as part of this feature.
- Do not redesign `namba run --team` or `namba run --parallel` semantics beyond documenting how harness-mode plans should target them.
- Do not make every normal feature plan carry harness-specific overhead; harness planning lives on its own command surface.
- Do not duplicate the `namba plan --help` read-only parsing fix here; keep that contract in `SPEC-019` and consume it from this feature.

## Design Constraints

- Keep `.namba/` as the source of truth for authored planning artifacts.
- Keep the default `namba plan` path stable for normal feature work.
- Keep `namba harness` discoverable as a distinct user intent: designing a reusable agent/skill harness is not the same task as planning an ordinary feature.
- Treat official Codex customization primitives as authoritative: `AGENTS.md`, repo-local skills, MCP where needed, built-in subagents, and project-scoped `.codex/agents/*.toml`.
- Reframe portable Harness ideas into Codex-native language rather than translating Claude terms literally.
- Keep planning artifacts concise but implementation-ready; the harness template must sharpen scope, not become a vague research memo.
- Make the evaluation guidance practical enough to implement later in repo-local scripts or docs without assuming network access or external SaaS dependencies.
- Update renderer sources first and regenerate derived instruction surfaces rather than patching generated docs alone.
