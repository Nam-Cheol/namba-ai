---
name: namba
description: Codex-native Namba command surface for SPEC orchestration inside a repository.
---

Use this skill whenever the user mentions `namba`, `namba project`, `namba regen`, `namba update`, `namba plan`, `namba fix`, `namba run`, `namba sync`, `namba pr`, `namba land`, or asks to use the Namba workflow.

Command mapping:
- `namba project`: refresh repository docs and codemaps.
- `namba regen`: regenerate AGENTS, repo-local skills, command-entry skills, Codex custom agents, readable role cards, and repo-local Codex config from `.namba/config/sections/*.yaml`.
- `namba update [--version vX.Y.Z]`: self-update the installed `namba` binary from GitHub Release assets.
- `namba plan "<description>"`: create the next feature SPEC package under `.namba/specs/`.
- `$namba-plan-pm-review` / `$namba-plan-eng-review` / `$namba-plan-design-review`: update product, engineering, or design review artifacts under `.namba/specs/<SPEC>/reviews/` and refresh advisory readiness.
- `namba fix "<description>"`: create the next bugfix SPEC package under `.namba/specs/`.
- `namba run SPEC-XXX`: execute the SPEC in the current Codex session. Read `spec.md`, `plan.md`, and `acceptance.md`, implement directly, validate, and sync artifacts.
- `namba run SPEC-XXX --solo|--team|--parallel`: use the standalone CLI runner when you need explicit single-subagent, multi-subagent, or worktree-parallel execution semantics.
- `namba sync`: refresh change summary, PR checklist, codemaps, advisory review readiness, and PR-ready docs after implementation.
- `namba pr "<title>"`: run sync plus validation by default, commit and push the current branch, create or reuse a PR, and ensure the Codex review marker exists.
- `namba land`: resolve the current branch PR, optionally wait for checks, merge when the PR is clean, and update local `main` safely.
- `namba doctor`: verify that AGENTS, repo skills, `.namba` config, Codex CLI, and the global `namba` command are available.

Execution rules:
1. Treat `.namba/` as the source of truth.
2. Prefer repo-local skills in `.agents/skills/`.
3. Prefer command-entry skills such as `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, `$namba-plan-pm-review`, `$namba-plan-eng-review`, `$namba-plan-design-review`, `$namba-project`, and `$namba-sync` when the user is invoking one Namba command directly.
4. Use the installed `namba` CLI for `project`, `regen`, `update`, `plan`, `fix`, `pr`, `land`, and `sync` when it will update repo state more reliably or self-update the installed CLI directly.
5. Keep `.namba/specs/<SPEC>/reviews/*.md` and `readiness.md` current when you use the plan-review workflow; the readiness summary is advisory unless the user explicitly asks for a gate.
6. For `namba run` in an interactive Codex session, prefer Codex-native in-session execution over recursively calling `namba run`, unless the user explicitly asks for standalone `--solo`, `--team`, `--parallel`, or `--dry-run` behavior.
7. Run validation commands from `.namba/config/sections/quality.yaml` before finishing.
8. Start each new SPEC or task on a dedicated work branch when `.namba/config/sections/git-strategy.yaml` enables branch-per-work collaboration.
9. Prepare PRs against `main`, write the title/body in Korean, and request GitHub Codex review with `@codex review` when the review flow is enabled.
