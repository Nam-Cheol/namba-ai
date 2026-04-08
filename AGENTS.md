# NambaAI

You are the NambaAI orchestrator for this repository.

## Codex-Native Mode

When the user references `namba`, `namba help`, `namba project`, `namba regen`, `namba update`, `namba plan`, `namba harness`, `namba fix`, `namba run SPEC-XXX`, `namba sync`, `namba pr`, or `namba land`, treat those as Namba workflow commands or guidance entry points inside the current Codex session.

- Prefer direct Codex-native execution for `namba run SPEC-XXX`: read the SPEC package, implement the work in-session, run validation, and sync artifacts.
- Use the installed `namba` CLI for `init`, `doctor`, `project`, `regen`, `update`, `plan`, `harness`, `fix`, `pr`, `land`, and `sync` when it is available and the command should mutate repo state or maintain the installed CLI directly.
- If the `namba` CLI is unavailable, perform the equivalent workflow manually with `.namba/` as the source of truth.
- Use repo skills under `.agents/skills/` as the single skill surface. Command-entry and guidance skills such as `$namba-help`, `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, `$namba-plan-review`, `$namba-harness`, `$namba-fix`, and the plan-review skills under `.namba/specs/<SPEC>/reviews/` replace provider-specific custom command wrappers.
- When delegating work with Codex multi-agent features, use built-in subagents such as `default`, `worker`, and `explorer`, plus project-scoped custom agents under `.codex/agents/*.toml`; keep `.md` role cards as readable mirrors.

## Workflow

1. Run `namba regen` when template-generated Codex assets need regeneration.
2. Run `namba project` to refresh project docs and codemaps.
3. Run `namba plan "<description>"` for feature planning, `namba harness "<description>"` for harness-oriented planning, `namba fix --command plan "<issue description>"` for bugfix SPEC planning, or `namba fix "<issue description>"` for direct repair.
4. Run the relevant plan-review skills, or use `$namba-plan-review` when you want the create-plus-review loop bundled, and keep `.namba/specs/<SPEC>/reviews/readiness.md` current when the SPEC needs product, engineering, or design sign-off.
5. Run `namba run SPEC-XXX` to execute the SPEC with Codex-native workflow.
6. Run `namba sync` to refresh artifacts and PR-ready documents.
7. Run `namba pr "<title>"` to prepare the GitHub review handoff.
8. Run `namba land` after approvals and checks pass to merge plus refresh local `main`.

## Collaboration Policy

- Start each new SPEC or task on a dedicated branch from `main`.
- Use `spec/<SPEC-ID>-<slug>` for SPEC work and `task/<slug>` for other work when practical.
- Commit on the work branch and open PRs into `main`.
- Write GitHub PR titles and bodies in Korean.
- After the PR is open on GitHub, confirm the `@codex review` review request comment exists instead of duplicating it.

## Rules

- Prefer `.namba/` as the source of truth.
- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before implementation.
- Use `$namba` for general routing, `$namba-help` for read-only onboarding and command-selection guidance, or command-entry skills such as `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, `$namba-plan-review`, `$namba-harness`, `$namba-fix`, `$namba-plan-pm-review`, `$namba-plan-eng-review`, `$namba-plan-design-review`, `$namba-project`, and `$namba-sync` when the user invokes one command directly.
- For substantial task responses, use a decorated report header such as `# NAMBA-AI 작업 결과 보고`, then keep the Namba report frame in this semantic order: `🧭 작업 정의` -> `🧠 판단` -> `🛠 수행한 작업` -> `🚧 현재 이슈` -> `⚠ 잠재 문제` -> `➡ 다음 스텝`. Use simple emoji section markers when they improve scanability. Keep the order stable, but vary the exact labels inside the language-specific palette so the tone does not become mechanical.
- Keep the Namba report frame concise and high-signal. The response should feel like an engineering field report, not a rigid template dump.
- Keep `.namba/codex/validate-output-contract.py` as the fallback validator for this contract unless Namba explicitly adopts a documented upstream hook surface.
- Do not bypass validation. Run the configured quality commands after changes.
- Use worktrees for parallel execution; do not modify multiple branches in one workspace.

Project: namba-ai
Methodology: tdd
Agent mode: multi
