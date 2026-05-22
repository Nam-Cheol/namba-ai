---
name: namba-run
description: Command-style entry point for executing a SPEC package with the Namba workflow.
---

State effect: mutating workflow entry point. Use help/probe paths read-only, and otherwise expect repository state or GitHub state to change.

Generated instruction contract for this command skill:
- Purpose: keep the role or command scope explicit, bounded, and testable.
- Boundary: honor read-only versus mutating state effects, configured sandbox mode, and assigned file or workflow ownership.
- Required output: report concrete actions, changed paths or artifacts, validation evidence, and pass/fail status or blockers.
- Pass/fail criteria: claim success only when acceptance criteria and configured validation are satisfied; otherwise name the exact blocker and impact.
- Evidence expectations: cite source artifacts such as SPEC files, `.namba/` configs, diffs, test output, PR/check links, or generated manifests instead of relying on unsupported assertions.
- Security responsibilities: never expose or commit secrets; treat auth, privacy, destructive commands, permission changes, and external network or credential use as security-sensitive.
- Destructive command and escalation policy: do not run destructive commands unless explicitly requested; request approval for privileged, networked, or sandbox-blocked actions only when the active approval mode allows it, and otherwise report the blocker or use a safe non-escalating path.
- Fallback implementer boundary: if a specialist path is unavailable and the main/default implementer takes over, stay within the assigned scope and preserve the same evidence and validation duties.
- Portability: keep durable guidance non-project-specific unless the current repository config or SPEC explicitly provides the project detail.

Use this skill when the user explicitly says `$namba-run`, `namba run SPEC-XXX`, or asks to execute a SPEC through Namba.

Behavior:
- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before implementation.
- Read `.namba/specs/<SPEC>/reviews/readiness.md` when it exists so advisory review depth is visible before coding starts.
- Read `.namba/specs/<SPEC>/frontend-brief.md` when it exists; it is the canonical source for frontend task classification and gate state.
- In an interactive Codex session, prefer Codex-native in-session execution over recursively calling `namba run`.
- Only use the standalone CLI runner for `--solo`, `--team`, `--parallel`, `--dry-run`, or when the user explicitly wants the non-interactive runner path.
- For `--solo`, stay inside one runner unless one domain clearly dominates and a single specialist would materially reduce risk.
- For `--team`, prefer one specialist when one domain dominates, expand to two or three only when acceptance spans multiple domains, and keep one integrator plus final validation owner in the workspace.
- For `--team`, honor each selected role's `model` and `model_reasoning_effort` metadata from `.codex/agents/*.toml` so planner/reviewer/security roles can think harder without making every delivery role heavy.
- Route art direction, palette/tone logic, composition, motion intent, Figma critique, and generic-section redesign work to `namba-designer`; route component boundaries, state ownership, and UI delivery planning to `namba-frontend-architect`; route approved UI implementation to `namba-frontend-implementer`; route mobile-specific UI delivery to `namba-mobile-engineer`; route API, schema, and pipeline work to backend/data; route auth, secrets, and compliance work to security; route deployment and runtime work to devops.
- `frontend-major` work must not move into architecture or implementation until `frontend-brief.md` shows coherent problem, reference, critique, decision, prototype evidence, a complete Do-Not Design Contract with reference-driven asset manifest, generated-image decision fields, generation plan, and aligned design clearance; `frontend-minor` keeps the lightweight advisory path.
- Treat review readiness as advisory by default for non-frontend and `frontend-minor` work, but block explicit `frontend-major` execution when the frontend brief is missing required evidence, internally contradictory, mismatched with design-review summaries, or missing/insufficient negative-first contract evidence.
- For `frontend-major`, first frontend and first major screen phases default to `Asset mode: generated-images` and `Imagegen requirement: required`; accept `existing-assets` or `not-applicable` only with validator-readable Asset decision proof.
- For `frontend-major` implementation results, require a `Do-Not Design Violation Check` that cites changed files, names any banned pattern found, and cites the exception path when one is used; when `Imagegen requirement: required` is present, require `Generated Asset Evidence` with manifest path and repeatable Asset ID blocks containing generated file, saved asset path, prompt summary, intended UI usage, and rendered usage evidence.
- For browser-rendered frontend work, use managed server lifecycle, wait for rendered DOM state, capture screenshots, inspect console errors, and prefer Playwright checks when the surface runs in a browser.
- Run validation commands from `.namba/config/sections/quality.yaml` and finish with `namba sync`. Use `namba pr` and `namba land` for the GitHub handoff and merge cycle instead of overloading `sync`.
- Collaboration defaults: branch from `main`, open the PR into `main`, write the PR in Korean, and request Codex review only when `namba pr --review` or queue `--review` is explicit; use `@codex review` as the request command.
