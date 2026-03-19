# Codex Integration

`namba-ai` is configured for Codex-native Namba workflow.

## What `namba init .` Enables

- Creates `AGENTS.md` with Namba orchestration rules.
- Creates repo-local skills under `.agents/skills/`, including command-entry skills such as `namba-run`, `namba-pr`, `namba-land`, `namba-plan`, and `namba-sync`.
- Creates Codex custom agents under `.codex/agents/*.toml` and readable `.md` role-card mirrors.
- Creates repo-local Codex config under `.codex/config.toml`, including the selected `approval_policy` and `sandbox_mode`.
- Creates `.namba/codex/output-contract.md` plus `.namba/codex/validate-output-contract.py` for NambaAI response-shape guidance and fallback validation.
- Creates `.namba/` project state, configs, docs, and SPEC storage.

## How Codex Uses Namba After Init

1. Open Codex in the initialized project directory.
2. Codex loads `AGENTS.md` and repo skills.
3. Invoke `$namba` for routing or command-entry skills such as `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, and `$namba-sync` for direct command-style execution.
4. Use built-in Codex delegation with `.codex/agents/*.toml` custom agents when multi-agent work is appropriate. The matching `.md` files remain readable mirrors.
5. Use `namba project`, `namba regen`, `namba update`, `namba plan`, `namba fix`, `namba run SPEC-XXX`, `namba sync`, `namba pr`, and `namba land` as workflow commands.

## Workflow Command Semantics

- `namba regen` regenerates `AGENTS.md`, repo skills under `.agents/skills/`, `.codex/agents/*.toml` custom agents, readable `.md` role-card mirrors, `.namba/codex/*`, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.
- `namba update` self-updates the installed `namba` binary from GitHub Release assets. Use `--version vX.Y.Z` for a specific release.
- `namba sync` refreshes `.namba/project/*` docs, release notes/checklists, and codemaps.
- `namba pr` prepares the current branch for GitHub review by syncing, validating, committing, pushing, opening or reusing the PR, and ensuring the Codex review marker is present.
- `namba land` waits for checks when requested, merges a clean PR, and updates local `main` safely.
- `namba release` requires a clean `main` branch and passing validators before it creates a tag. `--push` pushes both `main` and the new tag.
- `namba run SPEC-XXX --parallel` refers to the standalone runner path. It uses git worktrees, merges only after every worker passes execution and validation, and preserves failed worktrees and branches for inspection.

## Output Contract

- `AGENTS.md` defines a Namba report header such as `# NAMBA-AI 작업 결과 보고` for substantial responses.
- The report sections follow this semantic order: `🧭 작업 정의` -> `🧠 판단` -> `🛠 수행한 작업` -> `🚧 현재 이슈` -> `⚠ 잠재 문제` -> `➡ 다음 스텝`.
- The semantic order stays fixed, but the exact labels can vary within the selected language palette so the writing does not become robotic.
- `.namba/codex/validate-output-contract.py` checks this contract from a saved response file or stdin.
- OpenAI Codex docs currently describe AGENTS, repo skills, and built-in slash commands, but they do not document a repository-configurable stop-hook surface. Treat the validator script as the fallback until upstream hook support is documented.

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
- Claude subagents become explicit `.toml` custom agents used with Codex multi-agent delegation, with `.md` mirrors kept for readability.
- Claude hooks become explicit validator scripts, documented response contracts, and sync steps in Namba.
- Claude custom workflow commands become `$namba`, command-entry repo skills, built-in Codex slash commands, and the `namba` CLI.

## Important Distinction

- In interactive Codex sessions, `namba run SPEC-XXX` means Codex should execute the SPEC directly in-session.
- The standalone `namba run` CLI remains available for non-interactive runner-based execution.
- Tokens and PATs are intentionally excluded from generated config. Use `gh auth login` or `glab auth login` instead.
