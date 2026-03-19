---
name: namba-run
description: Command-style entry point for executing a SPEC package with the Namba workflow.
---

Use this skill when the user explicitly says `$namba-run`, `namba run SPEC-XXX`, or asks to execute a SPEC through Namba.

Behavior:
- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before implementation.
- In an interactive Codex session, prefer Codex-native in-session execution over recursively calling `namba run`.
- Only use the standalone CLI runner for `--solo`, `--team`, `--parallel`, `--dry-run`, or when the user explicitly wants the non-interactive runner path.
- For `--solo`, stay inside one runner unless one domain clearly dominates and a single specialist would materially reduce risk.
- For `--team`, prefer one specialist when one domain dominates, expand to two or three only when acceptance spans multiple domains, and keep one integrator plus final validation owner in the workspace.
- For `--team`, honor each selected role's `model` and `model_reasoning_effort` metadata from `.codex/agents/*.toml` so planner/reviewer/security roles can think harder without making every delivery role heavy.
- Route UI, responsive, mobile, and design work to frontend/mobile/designer roles; API, schema, and pipeline work to backend/data; auth, secrets, and compliance work to security; deployment and runtime work to devops.
- Run validation commands from `.namba/config/sections/quality.yaml` and finish with `namba sync`. Use `namba pr` and `namba land` for the GitHub handoff and merge cycle instead of overloading `sync`.
- Collaboration defaults: branch from `main`, open the PR into `main`, write the PR in Korean, and request `@codex review` on GitHub after the PR is open.
