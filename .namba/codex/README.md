# Codex Integration

`namba-ai` is configured for Codex-native Namba workflow.

## What `namba init .` Enables

- Creates `AGENTS.md` with Namba orchestration rules.
- Creates repo-local skills under `.agents/skills/`, including read-only guidance plus command-entry skills such as `namba-help`, `namba-create`, `namba-run`, `namba-pr`, `namba-land`, `namba-plan`, `namba-plan-review`, `namba-harness`, `namba-plan-pm-review`, `namba-plan-eng-review`, `namba-plan-design-review`, and `namba-sync`.
- Creates task-oriented Codex custom agents under `.codex/agents/*.toml` and readable `.md` role-card mirrors.
- Creates repo-local Codex config under `.codex/config.toml`, keeping a narrow repo-safe baseline such as `approval_policy`, `sandbox_mode`, and agent thread limits, plus an allow-listed set of repo-managed MCP presets when configured.
- Creates `.namba/codex/output-contract.md` plus `.namba/codex/validate-output-contract.py` for NambaAI response-shape guidance and fallback validation.
- Creates `.namba/` project state, configs, docs, and SPEC storage.

## How Codex Uses Namba After Init

1. Open Codex in the initialized project directory.
   On Windows, the current official Codex docs recommend using a WSL workspace for the best CLI experience.
2. Codex loads `AGENTS.md` and repo skills.
3. Invoke `$namba` for routing, `$namba-help` for read-only Namba usage guidance, or command-entry skills such as `$namba-create`, `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, `$namba-plan-review`, `$namba-harness`, `$namba-fix`, `$namba-plan-pm-review`, `$namba-plan-eng-review`, `$namba-plan-design-review`, and `$namba-sync` for direct command-style execution.
4. Use built-in Codex subagents such as `default`, `worker`, and `explorer`, plus project-scoped custom agents under `.codex/agents/*.toml`, when multi-agent work is appropriate. The matching `.md` files remain readable mirrors.
5. Use the plan-review skills to update `.namba/specs/<SPEC>/reviews/*.md` and keep `.namba/specs/<SPEC>/reviews/readiness.md` current when a SPEC needs product, engineering, or design critique before implementation, or use `$namba-plan-review` when you want the create-plus-review loop bundled into one Codex entry point.
6. Use `namba project`, `namba regen`, `namba update`, `namba codex access`, `namba plan`, `namba harness`, `namba fix`, `namba run SPEC-XXX`, `namba sync`, `namba pr`, and `namba land` as workflow commands.
## Workflow Command Semantics

- `$namba-help` explains how to use NambaAI, which command to choose next, and where the authoritative docs live. It should stay read-only.
- `$namba-create` is the preview-first creation path for repo-local skills and custom agents. Use it when the user wants `.agents/skills/*` or `.codex/agents/*` outputs directly instead of a SPEC package.
- `namba project` refreshes current repository docs and codemaps without creating a SPEC package.
- `namba codex access` inspects the current repo-owned Codex access defaults and mutates them only when explicit approval_policy / sandbox_mode flags are present.
- `namba regen` regenerates `AGENTS.md`, repo skills under `.agents/skills/`, `.codex/agents/*.toml` custom agents, readable `.md` role-card mirrors, `.namba/codex/*`, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.
- `namba update` self-updates the installed `namba` binary from GitHub Release assets. Use `--version vX.Y.Z` for a specific release.
- `namba plan "<description>"` creates the next feature SPEC package plus review scaffolds.
- `namba harness "<description>"` creates the next harness-oriented SPEC package plus review scaffolds while staying inside the standard `SPEC-XXX` model.
- `namba fix --command plan "<issue description>"` creates the next bugfix SPEC package plus review scaffolds.
- `namba fix "<issue description>"` and `namba fix --command run "<issue description>"` are the direct-repair paths in the current workspace. They should stay read-only for help/probe flows, avoid implicit SPEC creation, and finish with validation plus `namba sync`.
- `namba sync` refreshes `.namba/project/*` docs, release notes/checklists, codemaps, and advisory review readiness summaries.
- `namba pr` prepares the current branch for GitHub review by syncing, validating, committing, pushing, opening or reusing the PR, and ensuring the Codex review marker is present.
- `namba land` waits for checks when requested, merges a clean PR, and updates local `main` safely.
- `namba release` requires a clean `main` branch and passing validators before it creates a tag. `--push` pushes both `main` and the new tag.
- `namba run SPEC-XXX` keeps the standard standalone Codex flow when you use the CLI runner without extra mode flags.
- `namba run SPEC-XXX --solo` requests a standalone Codex run that explicitly targets a single-subagent workflow inside one workspace.
- `namba run SPEC-XXX --team` requests a standalone Codex run that explicitly coordinates multiple subagents inside one workspace.
- `namba run SPEC-XXX --parallel` still refers to the standalone worktree runner path. It uses git worktrees, merges only after every worker passes execution and validation, and preserves failed worktrees and branches for inspection.

## Namba Custom Agent Roster

- Strategy and readiness: `namba-product-manager` shapes scope and acceptance, `namba-planner` turns a SPEC into an execution plan, and `namba-plan-reviewer` validates whether the product/engineering/design review set is coherent enough to start implementation.
- UI split: `namba-designer` owns art direction, palette/tone logic, composition, motion intent, and diagnosis of generic or bland sections; `namba-frontend-architect` owns component boundaries, state/file planning, and delivery slicing; `namba-frontend-implementer` ships approved UI work; `namba-mobile-engineer` handles mobile-specific constraints.
- Routing examples: `Redesign this landing page hero so it stops looking generic` -> `namba-designer`; `Plan the component/state split for this dashboard` -> `namba-frontend-architect`; `Implement the approved dashboard filters and responsive states` -> `namba-frontend-implementer`.
- Backend and data: `namba-backend-architect` plans service boundaries, `namba-backend-implementer` ships server-side changes, and `namba-data-engineer` owns data pipelines, transformations, migrations, and analytics-facing changes.
- Security and delivery: `namba-security-engineer` handles hardening work, `namba-test-engineer` adds targeted regression coverage, `namba-devops-engineer` handles CI/CD and runtime changes, and `namba-reviewer` checks implementation acceptance before sync.
- General delivery: `namba-implementer` remains the generalist execution agent for mixed-scope implementation slices.
- Built-in Codex subagents such as `explorer` and `worker` still matter; use the Namba custom roster when responsibility and output expectations need tighter framing.

## Delegation Heuristics

- Default `namba run` stays inside the standalone runner unless specialist signals are strong enough to justify delegation.
- `--solo` uses at most one specialist when one domain clearly dominates the request.
- `--team` prefers one specialist when one domain dominates and expands to two or three only when acceptance spans multiple domains.
- Repo-managed same-workspace defaults set `.codex/config.toml [agents].max_threads = 5` when `agent_mode: multi`; worktree fan-out remains separately controlled by `.namba/config/sections/workflow.yaml`.
- Team mode honors each selected role's `model` and `model_reasoning_effort` metadata from `.codex/agents/*.toml`, keeping planner/reviewer/security roles stronger and delivery roles lighter.
- Route art direction, palette/tone logic, composition, motion intent, redesign, and Figma critique to `namba-designer`; route component, state, and delivery planning to `namba-frontend-architect`; route approved UI implementation to `namba-frontend-implementer`; route mobile-specific delivery to `namba-mobile-engineer`; route API, schema, and pipeline work to backend/data; route auth, secrets, and compliance work to security; route deployment and runtime work to devops.
- Keep the standalone runner as the integrator and final validation owner, and use `namba-reviewer` last when multiple specialists contribute.

## Plan Review Readiness

- `namba plan`, `namba harness`, and `namba fix --command plan` seed `.namba/specs/<SPEC>/reviews/product.md`, `engineering.md`, `design.md`, and `readiness.md`.
- `$namba-plan-review` bundles SPEC creation or resolution, the three review tracks, and an aggregate validation loop when you want that whole pre-implementation pass handled through one skill.
- `$namba-plan-pm-review`, `$namba-plan-eng-review`, and `$namba-plan-design-review` update those review artifacts directly in the repository.
- `namba run`, `namba sync`, and `namba pr` surface the latest readiness summary as advisory context so review depth is visible without silently hard-blocking delivery.

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
- Claude command wrappers become command-entry skills such as `$namba-create`, `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, and `$namba-sync`.
- Claude subagents map to Codex built-in subagents plus project-scoped `.toml` custom agents, with `.md` mirrors kept for readability.
- Claude hooks become explicit validator scripts, documented response contracts, and sync steps in Namba.
- Claude custom workflow commands become `$namba`, command-entry repo skills, built-in Codex slash commands, and the `namba` CLI.

## Important Distinction

- In interactive Codex sessions, `namba run SPEC-XXX` means Codex should execute the SPEC directly in-session.
- The standalone `namba run` CLI supports the default runner flow plus explicit `--solo`, `--team`, and worktree-based `--parallel` modes.
- Tokens and PATs are intentionally excluded from generated config. Use `gh auth login` or `glab auth login` instead.
