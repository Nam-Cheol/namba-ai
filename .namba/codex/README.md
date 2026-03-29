# Codex Integration

`namba-ai` is configured for Codex-native Namba workflow.

## What `namba init .` Enables

- Creates `AGENTS.md` with Namba orchestration rules.
- Creates repo-local skills under `.agents/skills/`, including command-entry skills such as `namba-run`, `namba-pr`, `namba-land`, `namba-plan`, `namba-plan-pm-review`, `namba-plan-eng-review`, `namba-plan-design-review`, and `namba-sync`.
- Creates task-oriented Codex custom agents under `.codex/agents/*.toml` and readable `.md` role-card mirrors.
- Creates repo-local Codex config under `.codex/config.toml`, keeping a narrow repo-safe baseline such as `approval_policy`, `sandbox_mode`, and agent thread limits, plus an allow-listed set of repo-managed MCP presets when configured.
- Creates `.namba/codex/output-contract.md` plus `.namba/codex/validate-output-contract.py` for NambaAI response-shape guidance and fallback validation.
- Creates `.namba/` project state, configs, docs, and SPEC storage.

## How Codex Uses Namba After Init

1. Open Codex in the initialized project directory.
   On Windows, the current official Codex docs recommend using a WSL workspace for the best CLI experience.
2. Codex loads `AGENTS.md` and repo skills.
3. Invoke `$namba` for routing or command-entry skills such as `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, `$namba-plan-pm-review`, `$namba-plan-eng-review`, `$namba-plan-design-review`, and `$namba-sync` for direct command-style execution.
4. Use built-in Codex subagents such as `default`, `worker`, and `explorer`, plus project-scoped custom agents under `.codex/agents/*.toml`, when multi-agent work is appropriate. The matching `.md` files remain readable mirrors.
5. Use the plan-review skills to update `.namba/specs/<SPEC>/reviews/*.md` and keep `.namba/specs/<SPEC>/reviews/readiness.md` current when a SPEC needs product, engineering, or design critique before implementation.
6. Use `namba project`, `namba regen`, `namba update`, `namba plan`, `namba fix`, `namba run SPEC-XXX`, `namba sync`, `namba pr`, and `namba land` as workflow commands.

## Namba Custom Agent Roster

- Strategy: `namba-product-manager` shapes scope and acceptance, `namba-planner` turns a SPEC into an execution plan, and both are the default review roles for the product and engineering plan-review passes.
- UI: `namba-frontend-architect` plans component boundaries and UI risks, `namba-frontend-implementer` ships approved UI work, `namba-mobile-engineer` handles mobile-specific constraints, and `namba-designer` clarifies visual direction and interaction intent.
- Backend and data: `namba-backend-architect` plans service boundaries, `namba-backend-implementer` ships server-side changes, and `namba-data-engineer` owns data pipelines, transformations, migrations, and analytics-facing changes.
- Security and delivery: `namba-security-engineer` handles hardening work, `namba-test-engineer` adds targeted regression coverage, `namba-devops-engineer` handles CI/CD and runtime changes, and `namba-reviewer` checks acceptance before sync.
- General delivery: `namba-implementer` remains the generalist execution agent for mixed-scope implementation slices.
- Built-in Codex subagents such as `explorer` and `worker` still matter; use the Namba custom roster when responsibility and output expectations need tighter framing.

## Delegation Heuristics

- Default `namba run` stays inside the standalone runner unless specialist signals are strong enough to justify delegation.
- `--solo` uses at most one specialist when one domain clearly dominates the request.
- `--team` prefers one specialist when one domain dominates and expands to two or three only when acceptance spans multiple domains.
- Team mode honors each selected role's `model` and `model_reasoning_effort` metadata from `.codex/agents/*.toml`, keeping planner/reviewer/security roles stronger and delivery roles lighter.
- Route UI, responsive, mobile, and Figma work to frontend/mobile/designer; API, schema, and pipeline work to backend/data; auth, secrets, and compliance work to security; deployment and runtime work to devops.
- Keep the standalone runner as the integrator and final validation owner, and use `namba-reviewer` last when multiple specialists contribute.

## Plan Review Readiness

- `namba plan` and `namba fix` seed `.namba/specs/<SPEC>/reviews/product.md`, `engineering.md`, `design.md`, and `readiness.md`.
- `$namba-plan-pm-review`, `$namba-plan-eng-review`, and `$namba-plan-design-review` update those review artifacts directly in the repository.
- `namba run`, `namba sync`, and `namba pr` surface the latest readiness summary as advisory context so review depth is visible without silently hard-blocking delivery.

## Workflow Command Semantics

- `namba regen` regenerates `AGENTS.md`, repo skills under `.agents/skills/`, `.codex/agents/*.toml` custom agents, readable `.md` role-card mirrors, `.namba/codex/*`, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.
- `namba update` self-updates the installed `namba` binary from GitHub Release assets. Use `--version vX.Y.Z` for a specific release.
- `namba sync` refreshes `.namba/project/*` docs, release notes/checklists, codemaps, and advisory review readiness summaries.
- `namba pr` prepares the current branch for GitHub review by syncing, validating, committing, pushing, opening or reusing the PR, and ensuring the Codex review marker is present.
- `namba land` waits for checks when requested, merges a clean PR, and updates local `main` safely.
- `namba release` requires a clean `main` branch and passing validators before it creates a tag. `--push` pushes both `main` and the new tag.
- `namba run SPEC-XXX` keeps the standard standalone Codex flow when you use the CLI runner without extra mode flags.
- `namba run SPEC-XXX --solo` requests a standalone Codex run that explicitly targets a single-subagent workflow inside one workspace.
- `namba run SPEC-XXX --team` requests a standalone Codex run that explicitly coordinates multiple subagents inside one workspace.
- `namba run SPEC-XXX --parallel` still refers to the standalone worktree runner path. It uses git worktrees, merges only after every worker passes execution and validation, and preserves failed worktrees and branches for inspection.

## Output Contract

- `AGENTS.md` defines a Namba report header such as `# NAMBA-AI 작업 결과 보고` for substantial responses.
- The report sections follow this semantic order: `🧭 작업 정의` -> `🧠 판단` -> `🛠 수행한 작업` -> `🚧 현재 이슈` -> `⚠ 잠재 문제` -> `➡ 다음 스텝`.
- The semantic order stays fixed, but the exact labels can vary within the selected language palette so the writing does not become robotic.
- `.namba/codex/validate-output-contract.py` checks this contract from a saved response file or stdin.
- Namba keeps the validator script as the explicit repository enforcement path even as Codex's documented config and hook surface evolves.

## Git Collaboration Defaults

- Each SPEC or new task uses a dedicated branch from `main`.
- Recommended branch names: `spec/<SPEC-ID>-<slug>` for SPEC work and `task/<slug>` for non-SPEC work.
- PRs target `main`.
- PR titles and bodies should be written in Korean.
- After the GitHub PR is open, confirm the `@codex review` review request is present.

## Claude to Codex Mapping

- `CLAUDE.md` becomes `AGENTS.md`.
- Claude skills become repo-local Codex skills under `.agents/skills/`.
- Claude command wrappers become command-entry skills such as `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, and `$namba-sync`.
- Claude subagents map to Codex built-in subagents plus project-scoped `.toml` custom agents, with `.md` mirrors kept for readability.
- Claude hooks become explicit validator scripts, documented response contracts, and sync steps in Namba.
- Claude custom workflow commands become `$namba`, command-entry repo skills, built-in Codex slash commands, and the `namba` CLI.

## Important Distinction

- In interactive Codex sessions, `namba run SPEC-XXX` means Codex should execute the SPEC directly in-session.
- The standalone `namba run` CLI supports the default runner flow plus explicit `--solo`, `--team`, and worktree-based `--parallel` modes.
- Tokens and PATs are intentionally excluded from generated config. Use `gh auth login` or `glab auth login` instead.
