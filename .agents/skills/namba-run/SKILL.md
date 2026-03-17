---
name: namba-run
description: Command-style entry point for executing a SPEC package with the Namba workflow.
---

Use this skill when the user explicitly says `$namba-run`, `namba run SPEC-XXX`, or asks to execute a SPEC through Namba.

Behavior:
- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before implementation.
- In an interactive Codex session, prefer Codex-native in-session execution over recursively calling `namba run`.
- Only use the standalone CLI runner for `--parallel`, `--dry-run`, or when the user explicitly wants the non-interactive runner path.
- Run validation commands from `.namba/config/sections/quality.yaml` and finish with `namba sync`.
- Collaboration defaults: branch from `main`, open the PR into `main`, write the PR in Korean, and request `@codex review` on GitHub after the PR is open.
