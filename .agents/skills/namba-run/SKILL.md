---
name: namba-run
description: Command-style entry point for executing a SPEC package with the Namba workflow.
---

Use this skill when the user explicitly says `$namba-run`, `namba run SPEC-XXX`, or asks to execute a SPEC through Namba.

Behavior:
- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before implementation.
- Read `.namba/specs/<SPEC>/reviews/readiness.md` when it exists so advisory review depth is visible before coding starts.
- In an interactive Codex session, prefer Codex-native in-session execution over recursively calling `namba run`.
- Only use the standalone CLI runner for `--solo`, `--team`, `--parallel`, `--dry-run`, or when the user explicitly wants the non-interactive runner path.
- For `--solo`, stay inside one runner unless one domain clearly dominates and a single specialist would materially reduce risk.
- For `--team`, prefer one specialist when one domain dominates, expand to two or three only when acceptance spans multiple domains, and keep one integrator plus final validation owner in the workspace.
- For `--team`, honor each selected role's `model` and `model_reasoning_effort` metadata from `.codex/agents/*.toml` so planner/reviewer/security roles can think harder without making every delivery role heavy.
- Route art direction, palette/tone logic, composition, motion intent, Figma critique, and generic-section redesign work to `namba-designer`; route component boundaries, state ownership, and UI delivery planning to `namba-frontend-architect`; route approved UI implementation to `namba-frontend-implementer`; route mobile-specific UI delivery to `namba-mobile-engineer`; route API, schema, and pipeline work to backend/data; route auth, secrets, and compliance work to security; route deployment and runtime work to devops.
- Treat review readiness as advisory by default: missing plan-review artifacts should be surfaced clearly, not silently block execution.
- Run validation commands from `.namba/config/sections/quality.yaml` and finish with `namba sync`. Use `namba pr` and `namba land` for the GitHub handoff and merge cycle instead of overloading `sync`.
- Collaboration defaults: branch from `main`, open the PR into `main`, write the PR in Korean, and request `@codex review` on GitHub after the PR is open.
