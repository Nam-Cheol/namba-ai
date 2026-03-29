# SPEC-014

## Goal

Adopt the highest-fit workflow ideas from `garrytan/gstack` into NambaAI by adding a structured pre-implementation plan-review pipeline and a review-readiness summary for each SPEC.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Verified external context as of 2026-03-29:
  - `garrytan/gstack` positions its workflow as `Think -> Plan -> Build -> Review -> Test -> Ship -> Reflect`.
  - The highest-fit gstack capabilities for NambaAI are the plan-stage review passes (`/plan-ceo-review`, `/plan-eng-review`, `/plan-design-review`), the persisted review artifacts they write, and the `Review Readiness Dashboard` that `/ship` checks before PR handoff.
  - gstack's heavier browser and QA features are built around a persistent Chromium daemon and session tooling. Those are real product capabilities, but they are not the closest near-term fit for NambaAI's current core, which is SPEC orchestration and Codex-native workflow scaffolding.
  - NambaAI already covers the downstream flow well with `namba plan`, `namba run`, `namba sync`, `namba pr`, `namba land`, generated docs, and explicit planner/reviewer roles. The clearest gap is the plan-stage review surface between `SPEC created` and `implementation started`.

## Problem

NambaAI can create a SPEC package and execute it, but it does not currently give teams a first-class way to record whether a SPEC has received product, engineering, or design review before work begins. The repo has the right agent roster, but not the explicit artifact surface or readiness summary that would make pre-implementation review visible and reusable.

## Desired Outcome

- Each new SPEC package includes a review workspace under `.namba/specs/<SPEC>/reviews/`.
- NambaAI exposes explicit command-entry skills for product, engineering, and design plan reviews so the user can run those passes intentionally instead of burying them inside ad hoc prompts.
- NambaAI maintains a review-readiness summary for each SPEC that makes it obvious which review passes exist, when they were last run, and whether anything is still missing before execution or PR handoff.
- The readiness summary should inform `namba run` and `namba pr` workflows, but it should remain advisory by default instead of becoming a surprising hard gate.

## Scope

- Add generated review artifact files for each SPEC, including product, engineering, design, and aggregate readiness views.
- Add repo-local command-entry skills for plan review passes, likely centered on:
  - `$namba-plan-pm-review`
  - `$namba-plan-eng-review`
  - `$namba-plan-design-review`
- Update generated docs so maintainers understand when to run these review passes and how the readiness summary affects execution and PR handoff.
- Surface the readiness summary in synced project artifacts and/or PR-preparation flow so reviewers can see whether the SPEC received the expected pre-implementation review.
- Add regression tests for scaffold generation, generated skills, and readiness-summary output.

## Non-Goals

- Do not attempt full gstack skill parity.
- Do not add the browser daemon, cookie import, `/browse`, `/qa`, or `/qa-only` in this SPEC.
- Do not add deploy automation such as `/land-and-deploy`, canary monitoring, or platform-specific production health checks.
- Do not add telemetry, retro analytics, or Greptile-specific integrations in this SPEC.
- Do not redesign `namba run --solo|--team|--parallel`, PR, or land semantics beyond surfacing review readiness.

## Design Constraints

- Keep `.namba/` as the source of truth for persisted review artifacts.
- Prefer generated repo-local skills and synced docs over bespoke runtime state kept outside the repository.
- Keep the feature Codex-native and provider-neutral rather than hard-coding a Claude-only control surface.
- Bias toward lightweight Markdown artifacts and generated summaries rather than large new runtime subsystems.
