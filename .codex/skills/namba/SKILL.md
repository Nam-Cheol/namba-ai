---
name: namba
description: Codex-native Namba command surface for SPEC orchestration inside a repository.
---

Use this skill whenever the user mentions `namba`, `namba project`, `namba plan`, `namba run`, `namba sync`, or asks to use the Namba workflow.

Command mapping:
- `namba project`: refresh repository docs and codemaps.
- `namba plan "<description>"`: create the next SPEC package under `.namba/specs/`.
- `namba run SPEC-XXX`: execute the SPEC in the current Codex session. Read `spec.md`, `plan.md`, and `acceptance.md`, implement directly, validate, and sync artifacts.
- `namba sync`: refresh change summary, PR checklist, and codemaps after implementation.
- `namba doctor`: verify that AGENTS, repo skills, `.namba` config, Codex CLI, and the global `namba` command are available.

Execution rules:
1. Treat `.namba/` as the source of truth.
2. Prefer repo-local skills in `.agents/skills/`.
3. Use the installed `namba` CLI for `project`, `plan`, and `sync` when it will update repo state more reliably.
4. For `namba run` in an interactive Codex session, prefer Codex-native in-session execution over recursively calling `namba run`.
5. Run validation commands from `.namba/config/sections/quality.yaml` before finishing.
