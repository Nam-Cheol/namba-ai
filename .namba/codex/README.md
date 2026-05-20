# Codex Integration

`namba-ai` is configured for Codex-native Namba workflow.

NambaAI's differentiator is prompt refinement before execution: ambiguous ideas should become clearer goals, scope, constraints, and acceptance criteria before Codex starts coding.

## What `namba init .` Enables

- Creates `AGENTS.md` with Namba orchestration rules.
- Creates repo-local skills under `.agents/skills/`, including read-only guidance plus command-entry skills such as `namba-help`, `namba-coach`, `namba-create`, `namba-run`, `namba-queue`, `namba-pr`, `namba-land`, `namba-release`, `namba-plan`, `namba-plan-review`, `namba-harness`, `namba-plan-pm-review`, `namba-plan-eng-review`, `namba-plan-design-review`, `namba-review-resolve`, and `namba-sync`.
- Creates task-oriented Codex custom agents under `.codex/agents/*.toml` and readable `.md` role-card mirrors.
- Creates repo-local Codex config under `.codex/config.toml`, keeping only deterministic repo-safe defaults such as lifecycle hook enablement, agent thread limits, status-line preference, and allow-listed repo-managed MCP presets when configured. It deliberately omits interactive approval mode, permissions, service tier, workspace roots, model, auth, apps, web search, and platform sandbox choices.
- Creates repo-local Codex lifecycle hooks under `.codex/hooks.json` and `.codex/hooks/` with `features.hooks = true`, providing Namba context, non-blocking prompt-refinement guidance for ambiguous prompts, approval-risk notes, final-report format checks, destructive-command guardrails, duplicate Namba guard suppression, and generated-surface reminders. Codex will ask you to review these hooks in `/hooks` before they run.
- Creates `.namba/codex/output-contract.md` plus `.namba/codex/validate-output-contract.py` for NambaAI response-shape guidance and fallback validation.
- Creates `.namba/` project state, configs, docs, and SPEC storage.

## How Codex Uses Namba After Init

1. Open Codex in the initialized project directory.
   On Windows, the current official Codex docs recommend using a WSL workspace for the best CLI experience.
2. If Codex reports `6 hooks need review`, open `/hooks`, inspect that each generated command resolves to this repository's `.codex/hooks/namba_codex_guard.sh` or the Windows `.codex/hooks/namba_codex_guard.ps1` launcher for `.codex/hooks/namba_codex_guard.py`, and approve before expecting prompt-refinement hooks to run.
3. Codex loads `AGENTS.md` and repo skills.
4. Let the repo-local Codex lifecycle hook refine ambiguous Namba prompts before execution. When the goal, target surface, constraints, or acceptance criteria are underspecified, Codex should ask concise clarifying questions and restate the improved prompt before planning or editing.
5. Invoke `$namba` for routing, `$namba-coach` for read-only current-goal command coaching, `$namba-help` for read-only Namba usage guidance, or command-entry skills such as `$namba-create`, `$namba-run`, `$namba-queue`, `$namba-pr`, `$namba-land`, `$namba-release`, `$namba-plan`, `$namba-plan-review`, `$namba-harness`, `$namba-fix`, `$namba-review-resolve`, `$namba-plan-pm-review`, `$namba-plan-eng-review`, `$namba-plan-design-review`, and `$namba-sync` for direct command-style execution.
6. Use built-in Codex subagents such as `default`, `worker`, and `explorer`, plus project-scoped custom agents under `.codex/agents/*.toml`, when multi-agent work is appropriate. The matching `.md` files remain readable mirrors.
7. Use the plan-review skills to update `.namba/specs/<SPEC>/reviews/*.md` and keep `.namba/specs/<SPEC>/reviews/readiness.md` current when a SPEC needs product, engineering, or design critique before implementation, or use `$namba-plan-review` when you want the create-plus-review loop bundled into one Codex entry point.
8. Use `namba project`, `namba regen`, `namba update`, `namba codex access`, `namba plan`, `namba harness`, `namba fix`, `namba run SPEC-XXX`, `namba queue`, `namba sync`, `namba pr`, `namba land`, and `namba release` as workflow commands.
## Workflow Command Semantics

- `$namba-help` explains how to use NambaAI, which command to choose next, and where the authoritative docs live. It should stay read-only.
- `$namba-coach` clarifies the user's current goal, corrects clearly wrong command choices, and hands off to exactly one primary Namba workflow invocation. It should stay read-only.
- `$namba-create` is the preview-first creation path for repo-local skills and custom agents. Use it when the user wants `.agents/skills/*` or `.codex/agents/*` outputs directly instead of a SPEC package.
- `$namba-queue` operates the existing-SPEC queue conveyor: start from a SPEC range or list, inspect durable status, resume after waits or blockers, and pause or stop without deleting branches, PRs, or evidence.
- `$namba-review-resolve` resolves GitHub review threads one by one: discover unresolved thread state with a thread-aware GitHub path, classify meaningful feedback versus non-actionable remarks, reply on original threads with validation plus relevant CI/check evidence, and request review again only after the meaningful items are handled.
- `$namba-release` handles NambaAI release orchestration: collect commits since the previous semver tag, draft release notes into a durable per-version artifact, and hand the release off through the guarded `namba release --version <version> --push` path.
- `namba project` refreshes current repository docs and codemaps without creating a SPEC package.
- `namba project` also writes `.namba/logs/project/codex-diagnostics-evidence.json` plus redacted `codex doctor` stdout/stderr logs when Codex diagnostics are available.
- `namba codex access` inspects Namba runner `codex exec` access defaults and mutates `.namba/config/sections/system.yaml` only when explicit approval_policy / sandbox_mode flags are present.
- Interactive Codex approval mode, permissions, service tier, effective workspace roots, models, auth, apps, web search, and platform sandbox choices stay user/session-owned; repo `.codex/config.toml` does not persist them.
- Codex 0.131 may display approval mode, permissions, service tier, and effective workspace roots in the TUI. Treat those as observed runtime state, not Namba-owned policy.
- Generated Codex lifecycle hooks are repo-local guardrails, not a complete security boundary: they add Namba context, guide ambiguous Namba prompts toward clarification questions without blocking submission, add approval-risk notes, check final-report format, deny destructive shell commands, and remind Codex about managed-surface changes.
- Codex requires repo-local hooks to be reviewed before they run. In the first interactive session after init or regen, open `/hooks`, inspect the generated commands, and approve them only if they resolve to the current repository's `.codex/hooks/namba_codex_guard.sh` or Windows `.codex/hooks/namba_codex_guard.ps1` launcher for `.codex/hooks/namba_codex_guard.py`.
- `namba regen` regenerates `AGENTS.md`, repo skills under `.agents/skills/`, `.codex/agents/*.toml` custom agents, readable `.md` role-card mirrors, `.namba/codex/*`, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.
- `namba update` self-updates the installed `namba` binary from GitHub Release assets. Use `--version vX.Y.Z` for a specific release.
- `namba update` may print Codex 0.131 baseline advice, but it never installs, updates, or manages Codex.
- `codex update` updates the upstream Codex CLI itself. Keep it separate from `namba update`.
- `namba plan "<description>"` creates the next feature SPEC package plus review scaffolds.
- `namba harness "<description>"` creates the next harness-oriented SPEC package plus review scaffolds while staying inside the standard `SPEC-XXX` model; harness/MCP plans should favor workflow-first design, bounded outputs, actionable errors, and stable read-only evaluations.
- `namba fix --command plan "<issue description>"` creates the next bugfix SPEC package plus review scaffolds.
- `namba fix "<issue description>"` and `namba fix --command run "<issue description>"` are the direct-repair paths in the current workspace. They should stay read-only for help/probe flows, avoid implicit SPEC creation, and finish with validation plus `namba sync`.
- `namba sync` refreshes `.namba/project/*` docs, release notes/checklists, codemaps, and advisory review readiness summaries.
- `namba queue start <SPEC-RANGE|SPEC-LIST>` processes existing SPEC packages in order through review, run, PR, checks, and optional land, while durable state under `.namba/logs/queue/` makes waits, blockers, and resume decisions explicit.
- `namba pr` prepares the current branch for GitHub review by syncing, validating, inspecting PR checks, summarizing bounded GitHub Actions failure snippets when checks fail, committing, pushing, and opening or reusing the PR. Use `namba pr --review` only when Codex review should be requested.
- `namba land` waits for checks when requested, merges a clean PR, and updates local `main` safely.
- `namba release` requires a clean `main` branch and passing validators before it creates a tag. `--push` pushes both `main` and the new tag.
- `namba run SPEC-XXX` keeps the standard standalone Codex flow when you use the CLI runner without extra mode flags, but explicit `frontend-major` work now reads `frontend-brief.md` as a canonical gate before coding.
- `namba run SPEC-XXX --solo` requests a standalone Codex run that explicitly targets a single-subagent workflow inside one workspace.
- `namba run SPEC-XXX --team` requests a standalone Codex run that explicitly coordinates multiple subagents inside one workspace.
- `namba run SPEC-XXX --parallel` still refers to the standalone worktree runner path. It uses git worktrees, merges only after every worker passes execution and validation, and preserves failed worktrees and branches for inspection.
- Codex `/goal` workflows are tracked as a future orchestration candidate, not a required Namba runtime dependency.
## Optional Platform Readiness

- Unified `@` mentions across files, directories, plugins, and skills are Codex platform search. Namba command routing still follows explicit `$namba-*` skills and `namba ...` command intent.
- Plugin packaging, marketplace CLI commands, version-aware sharing, share checkout, shared-workspace plugin buckets, and default-enabled plugin hooks are readiness paths. They do not make plugin installation or marketplace publication required for local NambaAI use.
- Remote-control and configured remote environments are optional diagnostics. When visible, they appear as additive `codex_diagnostics.remote_control` and `codex_diagnostics.remote_environments` evidence with normalized statuses such as `unavailable`, `disabled`, `enabled`, `configured`, or `local_fallback`.
- Project diagnostics may run stable local Codex probes; run, queue, hook, and parallel evidence use non-blocking snapshots or neutral fallback status.
- Multi-environment `apply_patch` selection remains a Codex platform capability; Namba's default editing path is the local workspace.
- Codex Python SDK references should use the `openai-codex` distribution and `openai_codex` import package. NambaAI does not add a Python runtime dependency for documentation-only references.

Field-level schema map:
- `codex_diagnostics.remote_control.status`: owner `codexDiagnosticsEvidence`; allowed `unavailable`, `disabled`, `enabled`, `local_fallback`; producer `buildCodexDiagnosticsEvidence`; consumers project/run/queue/hook/parallel evidence and docs.
- `codex_diagnostics.remote_control.source`: owner `codexRemoteControl`; allowed local source labels such as `stable_cli_status`, `stable_cli_help`, `explicit_config`, `environment_snapshot`, `codex_cli_missing`, and `non_blocking_snapshot`; producer diagnostics builder; consumers evidence readers and docs.
- `codex_diagnostics.remote_environments.status`: owner `codexDiagnosticsEvidence`; allowed `unavailable` or `configured`; producer diagnostics builder from explicit options, environment snapshots, or `CODEX_HOME`; consumers project/run/queue/hook/parallel evidence and docs.
- `codex_diagnostics.remote_environments.names`: owner `codexRemoteEnvironments`; allowed normalized configured environment labels; producer diagnostics builder; consumers project diagnostics and evidence readers.


## Codex 0.131 Boundary

Read this boundary in order: Codex displays runtime state, Namba generates repo assets, Namba validates run evidence, and some runtime behavior remains outside Namba control.

| Term | Boundary |
| --- | --- |
| `approval mode` | Codex may display the active session approval mode; repo `.codex/config.toml` does not set or persist it. |
| `permissions` | Codex may display current permission state; Namba docs describe the state but do not own it. |
| `effective workspace roots` | Codex may show the roots considered for the session; this is visibility, not a Namba promise that every root is writable. |
| `scoped write roots` | Codex sandbox policy may narrow writable paths; Namba documents the distinction and keeps generated config out of that choice. |
| `.codex/hooks.json` | Interactive Codex guardrails: context, prompt refinement, command denial, duplicate guard suppression, and generated-surface reminders. |
| `.namba/hooks.toml` | `namba run` evidence and validation hooks; it is not used to dedupe interactive Codex hooks. |

- Codex 0.131 also reports service tier and richer status-line state. Namba-generated docs can explain those fields, but repo config must not persist a user's service tier or approval/session choices.
- Project, run, and queue evidence may include a `codex_diagnostics` payload with neutral statuses such as `detected`, `not_detected`, `unavailable`, `timed_out`, `advisory_mismatch`, and `redacted`.
- Inspect project diagnostics at `.namba/logs/project/codex-diagnostics-evidence.json`, run diagnostics at `.namba/logs/runs/<log-id>-evidence.json`, queue handoff diagnostics at `.namba/logs/runs/<spec-id-lower>-queue-evidence.json`, and doctor logs beside the artifact that triggered them.
- Git helper commands may ignore configured repository hooks, so Namba cannot rely on Git hooks for safety or validation evidence.
- On Windows, prefer WSL for Codex workspaces when possible; otherwise expect stricter deny-read behavior, scoped write roots, ineffective firewall policy handling, and PowerShell launcher constraints.

## Namba Custom Agent Roster

- Strategy and readiness: `namba-product-manager` shapes scope and acceptance, `namba-planner` turns a SPEC into an execution plan, and `namba-plan-reviewer` validates whether the product/engineering/design review set is coherent enough to start implementation.
- UI split: `namba-designer` owns art direction plus reference collection and synthesis, `namba-frontend-architect` plans hierarchy and state only after the frontend gate is satisfied, `namba-frontend-implementer` ships approved UI work only after synthesis plus design clearance, and `namba-mobile-engineer` handles mobile-specific constraints.
- Routing examples: `Redesign this landing page hero so it stops looking generic` -> `namba-designer`; `Plan the component/state split for this dashboard` -> `namba-frontend-architect`; `Implement the approved dashboard filters and responsive states` -> `namba-frontend-implementer`.
- Backend and data: `namba-backend-architect` plans service boundaries, `namba-backend-implementer` ships server-side changes, and `namba-data-engineer` owns data pipelines, transformations, migrations, and analytics-facing changes.
- Security and delivery: `namba-security-engineer` handles hardening work, `namba-test-engineer` adds targeted regression coverage, `namba-devops-engineer` handles CI/CD and runtime changes, and `namba-reviewer` checks implementation acceptance before sync.
- General delivery: `namba-implementer` remains the generalist execution agent for mixed-scope implementation slices.
- Built-in Codex subagents such as `explorer` and `worker` still matter; use the Namba custom roster when responsibility and output expectations need tighter framing.

## Delegation Heuristics

- Default `namba run` stays inside the standalone runner unless specialist signals are strong enough to justify delegation.
- `--solo` uses at most one specialist when one domain clearly dominates the request.
- `--team` prefers one specialist when one domain dominates and expands to two or three only when acceptance spans multiple domains.
- Repo-managed same-workspace defaults set `.codex/config.toml [agents].max_threads = 5` when `agent_mode: multi`; Namba worktree workers remain separately controlled by `.namba/config/sections/workflow.yaml max_parallel_workers: 3` unless a later SPEC changes that fan-out.
- Team mode honors each selected role's `model` and `model_reasoning_effort` metadata from `.codex/agents/*.toml`, keeping planner/reviewer/security roles stronger and delivery roles lighter.
- Route art direction plus reference synthesis to `namba-designer`; route component, state, and delivery planning only after the frontend gate is satisfied to `namba-frontend-architect`; route approved UI implementation only after synthesis plus design clearance to `namba-frontend-implementer`; route mobile-specific delivery to `namba-mobile-engineer`; route API, schema, and pipeline work to backend/data; route auth, secrets, and compliance work to security; route deployment and runtime work to devops.
- Keep the standalone runner as the integrator and final validation owner, and use `namba-reviewer` last when multiple specialists contribute.

## Plan Review Readiness

- `namba plan`, `namba harness`, and `namba fix --command plan` seed `.namba/specs/<SPEC>/reviews/product.md`, `engineering.md`, `design.md`, and `readiness.md`.
- Frontend-touching planning also seeds `.namba/specs/<SPEC>/frontend-brief.md`, and explicit `frontend-major` work uses that brief as the canonical gate contract.
- `$namba-plan-review` bundles SPEC creation or resolution, the three review tracks, and an aggregate validation loop when you want that whole pre-implementation pass handled through one skill.
- `$namba-plan-pm-review`, `$namba-plan-eng-review`, and `$namba-plan-design-review` update those review artifacts directly in the repository.
- `namba run`, `namba sync`, and `namba pr` surface the latest readiness summary as advisory context for non-frontend and `frontend-minor` work, while explicit `frontend-major` runs can block on missing, insufficient, invalid, or mismatched frontend evidence.

## Output Contract

- `AGENTS.md` defines a Namba report header such as `# NAMBA-AI 작업 결과 보고` for substantial responses.
- The report sections follow this semantic order: `🧭 작업 정의` -> `🧠 판단` -> `🛠 수행한 작업` -> `🚧 현재 이슈` -> `⚠ 잠재 문제` -> `➡ 다음에 해야 할 작업`.
- The semantic order stays fixed, but the exact labels can vary within the selected language palette so the writing does not become robotic.
- The final next-work section should name the concrete command, review, validation, or handoff the operator should do next.
- `.namba/codex/validate-output-contract.py` checks this contract from a saved response file or stdin.
- Namba keeps the validator script as the explicit repository enforcement path even as Codex's documented config and hook surface evolves.

## Git Collaboration Defaults

- Each SPEC or new task uses a dedicated branch from `main`.
- Recommended branch names: `spec/<SPEC-ID>-<slug>` for SPEC work and `task/<slug>` for non-SPEC work.
- PRs target `main`.
- PR titles and bodies should be written in Korean.
- Codex review requests are opt-in: use `namba pr --review` or queue `--review`, then confirm the `@codex review` request is present.

## Claude to Codex Mapping

- `CLAUDE.md` becomes `AGENTS.md`.
- Claude skills become repo-local Codex skills under `.agents/skills/`.
- Claude command wrappers become command-entry skills such as `$namba-create`, `$namba-run`, `$namba-queue`, `$namba-pr`, `$namba-land`, `$namba-plan`, `$namba-sync`, `$namba-review-resolve`, and `$namba-release`.
- Claude subagents map to Codex built-in subagents plus project-scoped `.toml` custom agents, with `.md` mirrors kept for readability.
- Claude hooks split into repo-local Codex lifecycle hooks under `.codex/hooks.json` plus Namba run hooks under `.namba/hooks.toml`, explicit validator scripts, documented response contracts, and sync steps.
- Claude custom workflow commands become `$namba`, command-entry repo skills, built-in Codex slash commands, and the `namba` CLI.

## Important Distinction

- In interactive Codex sessions, `namba run SPEC-XXX` means Codex should execute the SPEC directly in-session.
- The standalone `namba run` CLI supports the default runner flow plus explicit `--solo`, `--team`, and worktree-based `--parallel` modes.
- `.codex/hooks.json` configures Codex lifecycle hooks for interactive guardrails; `.namba/hooks.toml` configures Namba runner lifecycle hooks for `namba run` evidence and validation boundaries.
- Codex 0.131 plugin hooks can also be enabled by the runtime. Namba-generated hooks dedupe identical Namba guard payloads at the interactive boundary, while `.namba/hooks.toml` remains separate runner evidence and is not used to dedupe interactive hooks.
- Git helper commands may ignore configured repository hooks, so Namba validation must come from configured quality commands and run evidence rather than Git hooks alone.
- Tokens and PATs are intentionally excluded from generated config. Use `gh auth login` or `glab auth login` instead.
