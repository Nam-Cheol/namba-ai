package namba

import (
	"encoding/json"
	"fmt"
	"strings"
)

func renderAgents(profile initProfile) string {
	collab := renderCollaborationPolicy(profile)
	reportRule := fmt.Sprintf(
		"- For substantial task responses, use a decorated report header such as `%s`, then keep the Namba report frame in this semantic order: %s. Use simple emoji section markers when they improve scanability. Keep the order stable, but vary the exact labels inside the language-specific palette so the tone does not become mechanical.\n",
		outputContractHeaderExample(profile),
		outputContractSequence(profile),
	)
	return fmt.Sprintf("# NambaAI\n\n"+
		"You are the NambaAI orchestrator for this repository.\n\n"+
		"## Codex-Native Mode\n\n"+
		"When the user references `namba`, `namba project`, `namba regen`, `namba update`, `namba plan`, `namba fix`, `namba run SPEC-XXX`, `namba sync`, `namba pr`, or `namba land`, treat those as Namba workflow commands inside the current Codex session.\n\n"+
		"- Prefer direct Codex-native execution for `namba run SPEC-XXX`: read the SPEC package, implement the work in-session, run validation, and sync artifacts.\n"+
		"- Use the installed `namba` CLI for `init`, `doctor`, `project`, `regen`, `update`, `plan`, `fix`, `pr`, `land`, and `sync` when it is available and the command should mutate repo state or maintain the installed CLI directly.\n"+
		"- If the `namba` CLI is unavailable, perform the equivalent workflow manually with `.namba/` as the source of truth.\n"+
		"- Use repo skills under `.agents/skills/` as the single skill surface. Command-entry skills such as `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, and the plan-review skills under `.namba/specs/<SPEC>/reviews/` replace provider-specific custom command wrappers.\n"+
		"- When delegating work with Codex multi-agent features, use built-in subagents such as `default`, `worker`, and `explorer`, plus project-scoped custom agents under `.codex/agents/*.toml`; keep `.md` role cards as readable mirrors.\n\n"+
		"## Workflow\n\n"+
		"1. Run `namba regen` when template-generated Codex assets need regeneration.\n"+
		"2. Run `namba project` to refresh project docs and codemaps.\n"+
		"3. Run `namba plan \"<description>\"` for feature work or `namba fix \"<description>\"` for bug fixes.\n"+
		"4. Run the relevant plan-review skills and keep `.namba/specs/<SPEC>/reviews/readiness.md` current when the SPEC needs product, engineering, or design sign-off.\n"+
		"5. Run `namba run SPEC-XXX` to execute the SPEC with Codex-native workflow.\n"+
		"6. Run `namba sync` to refresh artifacts and PR-ready documents.\n"+
		"7. Run `namba pr \"<title>\"` to prepare the GitHub review handoff.\n"+
		"8. Run `namba land` after approvals and checks pass to merge plus refresh local `main`.\n\n"+
		"## Collaboration Policy\n\n"+
		"%s\n"+
		"## Rules\n\n"+
		"- Prefer `.namba/` as the source of truth.\n"+
		"- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before implementation.\n"+
		"- Use `$namba` for general routing, or command-entry skills such as `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, `$namba-plan-pm-review`, `$namba-plan-eng-review`, `$namba-plan-design-review`, `$namba-project`, and `$namba-sync` when the user invokes one command directly.\n"+
		"%s"+
		"- Keep the Namba report frame concise and high-signal. The response should feel like an engineering field report, not a rigid template dump.\n"+
		"- Keep `.namba/codex/validate-output-contract.py` as the fallback validator for this contract unless Namba explicitly adopts a documented upstream hook surface.\n"+
		"- Do not bypass validation. Run the configured quality commands after changes.\n"+
		"- Use worktrees for parallel execution; do not modify multiple branches in one workspace.\n\n"+
		"Project: %s\n"+
		"Methodology: %s\n"+
		"Agent mode: %s\n",
		collab,
		reportRule,
		profile.ProjectName,
		profile.DevelopmentMode,
		profile.AgentMode,
	)
}

func renderNambaSkill(profile initProfile) string {
	lines := []string{
		"---",
		"name: namba",
		"description: Codex-native Namba command surface for SPEC orchestration inside a repository.",
		"---",
		"",
		"Use this skill whenever the user mentions `namba`, `namba project`, `namba regen`, `namba update`, `namba plan`, `namba fix`, `namba run`, `namba sync`, `namba pr`, `namba land`, or asks to use the Namba workflow.",
		"",
		"Command mapping:",
		"- `namba project`: refresh repository docs and codemaps.",
		"- `namba regen`: regenerate AGENTS, repo-local skills, command-entry skills, Codex custom agents, readable role cards, and repo-local Codex config from `.namba/config/sections/*.yaml`.",
		"- `namba update [--version vX.Y.Z]`: self-update the installed `namba` binary from GitHub Release assets.",
		"- `namba plan \"<description>\"`: create the next feature SPEC package under `.namba/specs/`.",
		"- `$namba-plan-pm-review` / `$namba-plan-eng-review` / `$namba-plan-design-review`: update product, engineering, or design review artifacts under `.namba/specs/<SPEC>/reviews/` and refresh advisory readiness.",
		"- `namba fix \"<description>\"`: create the next bugfix SPEC package under `.namba/specs/`.",

		"- `namba run SPEC-XXX`: execute the SPEC in the current Codex session. Read `spec.md`, `plan.md`, and `acceptance.md`, implement directly, validate, and sync artifacts.",
		"- `namba run SPEC-XXX --solo|--team|--parallel`: use the standalone CLI runner when you need explicit single-subagent, multi-subagent, or worktree-parallel execution semantics.",
		"- `namba sync`: refresh change summary, PR checklist, codemaps, advisory review readiness, and PR-ready docs after implementation.",
		"- `namba pr \"<title>\"`: run sync plus validation by default, commit and push the current branch, create or reuse a PR, and ensure the Codex review marker exists.",
		"- `namba land`: resolve the current branch PR, optionally wait for checks, merge when the PR is clean, and update local `main` safely.",
		"- `namba doctor`: verify that AGENTS, repo skills, `.namba` config, Codex CLI, and the global `namba` command are available.",
		"",
		"Execution rules:",
		"1. Treat `.namba/` as the source of truth.",
		"2. Prefer repo-local skills in `.agents/skills/`.",
		"3. Prefer command-entry skills such as `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, `$namba-plan-pm-review`, `$namba-plan-eng-review`, `$namba-plan-design-review`, `$namba-project`, and `$namba-sync` when the user is invoking one Namba command directly.",
		"4. Use the installed `namba` CLI for `project`, `regen`, `update`, `plan`, `fix`, `pr`, `land`, and `sync` when it will update repo state more reliably or self-update the installed CLI directly.",
		"5. Keep `.namba/specs/<SPEC>/reviews/*.md` and `readiness.md` current when you use the plan-review workflow; the readiness summary is advisory unless the user explicitly asks for a gate.",
		"6. For `namba run` in an interactive Codex session, prefer Codex-native in-session execution over recursively calling `namba run`, unless the user explicitly asks for standalone `--solo`, `--team`, `--parallel`, or `--dry-run` behavior.",
		"7. Run validation commands from `.namba/config/sections/quality.yaml` before finishing.",
		"8. Start each new SPEC or task on a dedicated work branch when `.namba/config/sections/git-strategy.yaml` enables branch-per-work collaboration.",
		fmt.Sprintf("9. Prepare PRs against `%s`, write the title/body in %s, and request GitHub Codex review with `%s` when the review flow is enabled.", prBaseBranch(profile), humanLanguageName(profile.PRLanguage), codexReviewComment(profile)),
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderCommandSkill(name, description string, body []string) string {
	lines := []string{
		"---",
		fmt.Sprintf("name: %s", name),
		fmt.Sprintf("description: %s", description),
		"---",
		"",
	}
	lines = append(lines, body...)
	return strings.Join(lines, "\n") + "\n"
}

func renderInitCommandSkill() string {
	return renderCommandSkill(
		"namba-init",
		"Command-style entry point for project bootstrap with NambaAI.",
		[]string{
			"Use this skill when the user explicitly says `$namba-init`, `namba init`, or asks to bootstrap a repository with NambaAI.",
			"",
			"Behavior:",
			"- Prefer running the installed `namba init` CLI when available because it writes the scaffold deterministically.",
			"- Keep `.namba/config/sections/*.yaml` as the durable source of truth.",
			"- Explain that repo skills live under `.agents/skills/` and Codex subagents live under `.codex/agents/*.toml`.",
			"- Keep the selected human language aligned across Codex conversation, docs, PR content, and code comments.",
		},
	)
}

func renderProjectCommandSkill() string {
	return renderCommandSkill(
		"namba-project",
		"Command-style entry point for refreshing project docs and codemaps.",
		[]string{
			"Use this skill when the user explicitly says `$namba-project`, `namba project`, or asks to analyze the current repository before implementation.",
			"",
			"Behavior:",
			"- Prefer the installed `namba project` CLI when available.",
			"- Refresh `.namba/project/*` docs and codemaps before planning or execution.",
			"- Summarize entry points, structure, and generated artifacts after the refresh.",
		},
	)
}

func renderRegenCommandSkill() string {
	return renderCommandSkill(
		"namba-regen",
		"Command-style entry point for regenerating Namba scaffold assets from config.",
		[]string{
			"Use this skill when the user explicitly says `$namba-regen`, `namba regen`, or asks to re-render generated Namba assets from configuration.",
			"",
			"Behavior:",
			"- Regenerate `AGENTS.md`, repo skills under `.agents/skills/`, `.codex/agents/*.toml`, readable `.md` role cards, `.namba/codex/*`, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.",
			"- Do not recreate `.codex/skills/`; that mirror causes duplicate skill discovery in Codex.",
			"- Remove obsolete generated skill files when the template set changes.",
		},
	)
}

func renderUpdateCommandSkill() string {
	return renderCommandSkill(
		"namba-update",
		"Command-style entry point for self-updating the installed NambaAI CLI.",
		[]string{
			"Use this skill when the user explicitly says `$namba-update`, `namba update`, or asks to update the installed NambaAI version.",
			"",
			"Behavior:",
			"- Treat `namba update` as CLI self-update from GitHub Release assets.",
			"- Use `namba update --version vX.Y.Z` when the user requests a specific version.",
			"- Do not confuse this command with scaffold regeneration; that belongs to `namba regen`.",
		},
	)
}

func renderPlanCommandSkill() string {
	return renderCommandSkill(
		"namba-plan",
		"Command-style entry point for creating the next feature SPEC package.",
		[]string{
			"Use this skill when the user explicitly says `$namba-plan`, `namba plan`, or asks to create a new feature SPEC package.",
			"",
			"Behavior:",
			"- Prefer the installed `namba plan` CLI when available.",
			"- When repo-managed MCP presets are configured, prefer them for planning context before broader web search; for example, use `context7` for library and framework docs, `sequential-thinking` for deeper decomposition, and `playwright` for browser-verified flows.",
			"- Create the next sequential `SPEC-XXX` package under `.namba/specs/`.",
			"- Seed `.namba/specs/<SPEC>/reviews/` with product, engineering, design, and aggregate readiness artifacts.",
			"- Point follow-up review work to `$namba-plan-pm-review`, `$namba-plan-eng-review`, and `$namba-plan-design-review` when the SPEC needs pre-implementation critique.",
			"- Keep the scope concrete and implementation-ready.",
		},
	)
}

func renderPlanReviewCommandSkill(name, description, role, slug, title string) string {
	return renderCommandSkill(
		name,
		description,
		[]string{
			fmt.Sprintf("Use this skill when the user explicitly says `$%s` or asks for a %s review on a SPEC package.", name, strings.ToLower(title)),
			"",
			"Behavior:",
			"- Resolve the target SPEC from an explicit `SPEC-XXX`; otherwise use the latest SPEC under `.namba/specs/`.",
			"- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before writing review notes.",
			fmt.Sprintf("- Update `.namba/specs/<SPEC>/reviews/%s.md` with status, reviewer, findings, decisions, follow-ups, and recommendation.", slug),
			fmt.Sprintf("- Prefer `%s` as the review role when subagent routing is appropriate.", role),
			"- Refresh `.namba/specs/<SPEC>/reviews/readiness.md` so the advisory summary reflects the latest review state.",
			"- Keep the review advisory by default; surface missing depth or blockers clearly without silently turning the workflow into a hard gate.",
		},
	)
}

func renderPlanPMReviewCommandSkill() string {
	return renderPlanReviewCommandSkill(
		"namba-plan-pm-review",
		"Command-style entry point for product review of a SPEC before implementation starts.",
		"namba-product-manager",
		"product",
		"product",
	)
}

func renderPlanEngReviewCommandSkill() string {
	return renderPlanReviewCommandSkill(
		"namba-plan-eng-review",
		"Command-style entry point for engineering review of a SPEC before implementation starts.",
		"namba-planner",
		"engineering",
		"engineering",
	)
}

func renderPlanDesignReviewCommandSkill() string {
	return renderPlanReviewCommandSkill(
		"namba-plan-design-review",
		"Command-style entry point for design review of a SPEC before implementation starts.",
		"namba-designer",
		"design",
		"design",
	)
}

func renderFixCommandSkill() string {
	return renderCommandSkill(
		"namba-fix",
		"Command-style entry point for creating the next bug-fix SPEC package.",
		[]string{
			"Use this skill when the user explicitly says `$namba-fix`, `namba fix`, or asks to prepare a bug-fix SPEC package.",
			"",
			"Behavior:",
			"- Prefer the installed `namba fix` CLI when available.",
			"- Create the next sequential `SPEC-XXX` fix package under `.namba/specs/`.",
			"- Bias toward the smallest safe fix and explicit regression coverage.",
		},
	)
}

func renderRunCommandSkill(profile initProfile) string {
	return renderCommandSkill(
		"namba-run",
		"Command-style entry point for executing a SPEC package with the Namba workflow.",
		[]string{
			"Use this skill when the user explicitly says `$namba-run`, `namba run SPEC-XXX`, or asks to execute a SPEC through Namba.",
			"",
			"Behavior:",
			"- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before implementation.",
			"- Read `.namba/specs/<SPEC>/reviews/readiness.md` when it exists so advisory review depth is visible before coding starts.",
			"- In an interactive Codex session, prefer Codex-native in-session execution over recursively calling `namba run`.",
			"- Only use the standalone CLI runner for `--solo`, `--team`, `--parallel`, `--dry-run`, or when the user explicitly wants the non-interactive runner path.",
			"- For `--solo`, stay inside one runner unless one domain clearly dominates and a single specialist would materially reduce risk.",
			"- For `--team`, prefer one specialist when one domain dominates, expand to two or three only when acceptance spans multiple domains, and keep one integrator plus final validation owner in the workspace.",
			"- For `--team`, honor each selected role's `model` and `model_reasoning_effort` metadata from `.codex/agents/*.toml` so planner/reviewer/security roles can think harder without making every delivery role heavy.",
			"- Route UI, responsive, mobile, and design work to frontend/mobile/designer roles; API, schema, and pipeline work to backend/data; auth, secrets, and compliance work to security; deployment and runtime work to devops.",
			"- Treat review readiness as advisory by default: missing plan-review artifacts should be surfaced clearly, not silently block execution.",
			"- Run validation commands from `.namba/config/sections/quality.yaml` and finish with `namba sync`. Use `namba pr` and `namba land` for the GitHub handoff and merge cycle instead of overloading `sync`.",
			fmt.Sprintf("- Collaboration defaults: branch from `%s`, open the PR into `%s`, write the PR in %s, and request `%s` on GitHub after the PR is open.", branchBase(profile), prBaseBranch(profile), humanLanguageName(profile.PRLanguage), codexReviewComment(profile)),
		},
	)
}

func renderPRCommandSkill(profile initProfile) string {
	return renderCommandSkill(
		"namba-pr",
		"Command-style entry point for preparing the current branch for GitHub review.",
		[]string{
			"Use this skill when the user explicitly says `$namba-pr`, `namba pr`, or asks to hand off the current branch for review.",
			"",
			"Behavior:",
			"- Use the configured PR base branch, PR language, and Codex review marker from `.namba/config/sections/git-strategy.yaml`.",
			"- Run `namba sync` and validation by default before creating review artifacts.",
			"- Include the latest SPEC review-readiness artifact in the PR summary/checklist when `.namba/specs/<SPEC>/reviews/readiness.md` exists.",
			"- Commit and push the current work branch, create or reuse the GitHub PR, and ensure the Codex review marker exists without duplication.",
			fmt.Sprintf("- Collaboration defaults: PRs target `%s`, PR content is written in %s, and `%s` is the review marker.", prBaseBranch(profile), humanLanguageName(profile.PRLanguage), codexReviewComment(profile)),
		},
	)
}

func renderLandCommandSkill(profile initProfile) string {
	return renderCommandSkill(
		"namba-land",
		"Command-style entry point for merging a prepared GitHub PR and updating local main safely.",
		[]string{
			"Use this skill when the user explicitly says `$namba-land`, `namba land`, or asks to merge a prepared PR.",
			"",
			"Behavior:",
			"- Resolve the PR from the current branch when a PR number is not provided.",
			"- Wait for required checks when requested, merge only when review and merge state are clean, and report blockers clearly.",
			fmt.Sprintf("- After merge, update local `%s` safely without clobbering unrelated working tree changes.", prBaseBranch(profile)),
		},
	)
}

func renderSyncCommandSkill() string {
	return renderCommandSkill(
		"namba-sync",
		"Command-style entry point for refreshing Namba project artifacts after implementation.",
		[]string{
			"Use this skill when the user explicitly says `$namba-sync`, `namba sync`, or asks to refresh PR-ready Namba artifacts after changes.",
			"",
			"Behavior:",
			"- Refresh `.namba/project/*` docs, release notes/checklists, codemaps, advisory review readiness under `.namba/specs/<SPEC>/reviews/`, and any README bundles enabled by `.namba/config/sections/docs.yaml` after implementation.",
			"- Keep PR creation and merge automation in `namba pr` and `namba land` so `sync` stays a local artifact refresh command.",
			"- Use `namba regen` separately when template-generated scaffold assets changed.",
			"- Run validation first when code changed and the quality config requires it.",
		},
	)
}

func renderFoundationSkill() string {
	lines := []string{
		"---",
		"name: namba-foundation-core",
		"description: Core NambaAI workflow, SPEC discipline, TRUST quality gates, and Codex-native execution rules.",
		"---",
		"",
		"Use this skill when the task involves NambaAI workflow orchestration, SPEC handling, quality gates, or phased delivery.",
		"",
		"Key ideas:",
		"- SPEC-first execution",
		"- Codex-native implementation for `namba run` requests inside an interactive session",
		"- TDD for greenfield or sufficiently tested projects",
		"- DDD-style preserve/improve flow for brownfield projects with weak test coverage",
		"- TRUST gates after each execution phase",
		"- Worktree-based isolation for parallel work",
		"- Namba report frame for substantial user-facing responses: scoped definition, judgment, work completed, current issues, potential risks, next steps",
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderInitSkill() string {
	lines := []string{
		"---",
		"name: namba-workflow-init",
		"description: Codex-adapted init workflow that maps MoAI and Claude assets into NambaAI scaffold.",
		"---",
		"",
		"Use this skill when the user asks about `namba init`, project bootstrap, or Claude-to-Codex migration.",
		"",
		"Core mapping:",
		"- `CLAUDE.md` -> `AGENTS.md`",
		"- `.claude/skills/*` -> `.agents/skills/*`",
		"- `.claude/commands/*` -> command-entry repo skills such as `.agents/skills/namba-run/SKILL.md`",
		"- `.claude/agents/*` -> `.codex/agents/*.toml` custom agents with `.md` role-card mirrors",
		"- `.claude/hooks/*` -> explicit validation pipeline and `namba` orchestration",
		"- Claude custom slash-command workflows -> built-in Codex slash commands plus repo skills such as `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, `$namba-sync`, and the `namba` CLI",
		"",
		"When implementing init changes:",
		"1. Keep `.namba/config/sections/*.yaml` as the durable source of truth.",
		"2. Never write tokens or secrets into generated config files.",
		"3. Prefer repo-local skills and `.toml` custom agents while keeping `.md` files as readable mirrors.",
		"4. Keep one selected human language aligned across Codex conversation, docs, PR content, and code comments unless the user explicitly overrides it.",
		"5. Keep generated assets readable so users can understand what `namba init .` changed.",
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderProjectSkill() string {
	lines := []string{
		"---",
		"name: namba-workflow-project",
		"description: Project analysis, codemap refresh, and documentation generation for NambaAI.",
		"---",
		"",
		"Use this skill to:",
		"- refresh project docs",
		"- summarize structure and entry points",
		"- rebuild codemap artifacts under `.namba/project/codemaps`",
		"- explain how the repository is organized before planning or execution",
		"- implement the `namba project` command inside Codex when the CLI is not used directly",
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderExecutionSkill(profile initProfile) string {
	lines := []string{
		"---",
		"name: namba-workflow-execution",
		"description: Execute NambaAI SPEC packages with Codex-native workflow and explicit validation.",
		"---",
		"",
		"Use this skill when implementing a SPEC package.",
		"",
		"Execution pattern:",
		"1. Read `.namba/specs/<SPEC>/spec.md`",
		"2. Read `.namba/specs/<SPEC>/plan.md`",
		"3. Read `.namba/specs/<SPEC>/acceptance.md`",
		"4. Read `.namba/specs/<SPEC>/reviews/readiness.md` when present so advisory review status informs execution",
		"5. Implement the work directly in the current Codex session",
		"6. Run configured validation commands",
		"7. Summarize results in `.namba/logs` and sync artifacts",
		"",
		fmt.Sprintf("Collaboration defaults: use a dedicated branch from `%s` for the SPEC, open the PR into `%s`, write the PR in %s, and request `%s` on GitHub after the PR is open.", branchBase(profile), prBaseBranch(profile), humanLanguageName(profile.PRLanguage), codexReviewComment(profile)),
		"",
		"Do not call `namba run` from inside Codex unless the user explicitly requests the non-interactive CLI runner.",
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderCodexUsage(profile initProfile) string {
	lines := []string{
		"# Codex Integration",
		"",
		fmt.Sprintf("`%s` is configured for Codex-native Namba workflow.", profile.ProjectName),
		"",
		"## What `namba init .` Enables",
		"",
		"- Creates `AGENTS.md` with Namba orchestration rules.",
		"- Creates repo-local skills under `.agents/skills/`, including command-entry skills such as `namba-run`, `namba-pr`, `namba-land`, `namba-plan`, `namba-plan-pm-review`, `namba-plan-eng-review`, `namba-plan-design-review`, and `namba-sync`.",
		"- Creates task-oriented Codex custom agents under `.codex/agents/*.toml` and readable `.md` role-card mirrors.",
		"- Creates repo-local Codex config under `.codex/config.toml`, keeping a narrow repo-safe baseline such as `approval_policy`, `sandbox_mode`, and agent thread limits, plus an allow-listed set of repo-managed MCP presets when configured.",
		"- Creates `.namba/codex/output-contract.md` plus `.namba/codex/validate-output-contract.py` for NambaAI response-shape guidance and fallback validation.",
		"- Creates `.namba/` project state, configs, docs, and SPEC storage.",
		"",
		"## How Codex Uses Namba After Init",
		"",
		"1. Open Codex in the initialized project directory.",
		"   On Windows, the current official Codex docs recommend using a WSL workspace for the best CLI experience.",
		"2. Codex loads `AGENTS.md` and repo skills.",
		"3. Invoke `$namba` for routing or command-entry skills such as `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, `$namba-plan-pm-review`, `$namba-plan-eng-review`, `$namba-plan-design-review`, and `$namba-sync` for direct command-style execution.",
		"4. Use built-in Codex subagents such as `default`, `worker`, and `explorer`, plus project-scoped custom agents under `.codex/agents/*.toml`, when multi-agent work is appropriate. The matching `.md` files remain readable mirrors.",
		"5. Use the plan-review skills to update `.namba/specs/<SPEC>/reviews/*.md` and keep `.namba/specs/<SPEC>/reviews/readiness.md` current when a SPEC needs product, engineering, or design critique before implementation.",
		"6. Use `namba project`, `namba regen`, `namba update`, `namba plan`, `namba fix`, `namba run SPEC-XXX`, `namba sync`, `namba pr`, and `namba land` as workflow commands.",
		"",
		"## Namba Custom Agent Roster",
		"",
		"- Strategy: `namba-product-manager` shapes scope and acceptance, `namba-planner` turns a SPEC into an execution plan, and both are the default review roles for the product and engineering plan-review passes.",
		"- UI: `namba-frontend-architect` plans component boundaries and UI risks, `namba-frontend-implementer` ships approved UI work, `namba-mobile-engineer` handles mobile-specific constraints, and `namba-designer` clarifies visual direction and interaction intent.",
		"- Backend and data: `namba-backend-architect` plans service boundaries, `namba-backend-implementer` ships server-side changes, and `namba-data-engineer` owns data pipelines, transformations, migrations, and analytics-facing changes.",
		"- Security and delivery: `namba-security-engineer` handles hardening work, `namba-test-engineer` adds targeted regression coverage, `namba-devops-engineer` handles CI/CD and runtime changes, and `namba-reviewer` checks acceptance before sync.",
		"- General delivery: `namba-implementer` remains the generalist execution agent for mixed-scope implementation slices.",
		"- Built-in Codex subagents such as `explorer` and `worker` still matter; use the Namba custom roster when responsibility and output expectations need tighter framing.",
		"",
		"## Delegation Heuristics",
		"",
		"- Default `namba run` stays inside the standalone runner unless specialist signals are strong enough to justify delegation.",
		"- `--solo` uses at most one specialist when one domain clearly dominates the request.",
		"- `--team` prefers one specialist when one domain dominates and expands to two or three only when acceptance spans multiple domains.",
		"- Team mode honors each selected role's `model` and `model_reasoning_effort` metadata from `.codex/agents/*.toml`, keeping planner/reviewer/security roles stronger and delivery roles lighter.",
		"- Route UI, responsive, mobile, and Figma work to frontend/mobile/designer; API, schema, and pipeline work to backend/data; auth, secrets, and compliance work to security; deployment and runtime work to devops.",
		"- Keep the standalone runner as the integrator and final validation owner, and use `namba-reviewer` last when multiple specialists contribute.",
		"",
		"## Plan Review Readiness",
		"",
		"- `namba plan` and `namba fix` seed `.namba/specs/<SPEC>/reviews/product.md`, `engineering.md`, `design.md`, and `readiness.md`.",
		"- `$namba-plan-pm-review`, `$namba-plan-eng-review`, and `$namba-plan-design-review` update those review artifacts directly in the repository.",
		"- `namba run`, `namba sync`, and `namba pr` surface the latest readiness summary as advisory context so review depth is visible without silently hard-blocking delivery.",
		"",
		"## Workflow Command Semantics",
		"",
		"- `namba regen` regenerates `AGENTS.md`, repo skills under `.agents/skills/`, `.codex/agents/*.toml` custom agents, readable `.md` role-card mirrors, `.namba/codex/*`, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.",
		"- `namba update` self-updates the installed `namba` binary from GitHub Release assets. Use `--version vX.Y.Z` for a specific release.",
		"- `namba sync` refreshes `.namba/project/*` docs, release notes/checklists, codemaps, and advisory review readiness summaries.",
		"- `namba pr` prepares the current branch for GitHub review by syncing, validating, committing, pushing, opening or reusing the PR, and ensuring the Codex review marker is present.",
		"- `namba land` waits for checks when requested, merges a clean PR, and updates local `main` safely.",
		"- `namba release` requires a clean `main` branch and passing validators before it creates a tag. `--push` pushes both `main` and the new tag.",
		"- `namba run SPEC-XXX` keeps the standard standalone Codex flow when you use the CLI runner without extra mode flags.",
		"- `namba run SPEC-XXX --solo` requests a standalone Codex run that explicitly targets a single-subagent workflow inside one workspace.",
		"- `namba run SPEC-XXX --team` requests a standalone Codex run that explicitly coordinates multiple subagents inside one workspace.",
		"- `namba run SPEC-XXX --parallel` still refers to the standalone worktree runner path. It uses git worktrees, merges only after every worker passes execution and validation, and preserves failed worktrees and branches for inspection.",
		"",
		"## Output Contract",
		"",
		fmt.Sprintf("- `AGENTS.md` defines a Namba report header such as `%s` for substantial responses.", outputContractHeaderExample(profile)),
		fmt.Sprintf("- The report sections follow this semantic order: %s.", outputContractSequence(profile)),
		"- The semantic order stays fixed, but the exact labels can vary within the selected language palette so the writing does not become robotic.",
		"- `.namba/codex/validate-output-contract.py` checks this contract from a saved response file or stdin.",
		"- Namba keeps the validator script as the explicit repository enforcement path even as Codex's documented config and hook surface evolves.",
		"",
		"## Git Collaboration Defaults",
		"",
		fmt.Sprintf("- Each SPEC or new task uses a dedicated branch from `%s`.", branchBase(profile)),
		fmt.Sprintf("- Recommended branch names: `%s<SPEC-ID>-<slug>` for SPEC work and `%s<slug>` for non-SPEC work.", specBranchPrefix(profile), taskBranchPrefix(profile)),
		fmt.Sprintf("- PRs target `%s`.", prBaseBranch(profile)),
		fmt.Sprintf("- PR titles and bodies should be written in %s.", humanLanguageName(profile.PRLanguage)),
		fmt.Sprintf("- After the GitHub PR is open, confirm the `%s` review request is present.", codexReviewComment(profile)),
		"",
		"## Claude to Codex Mapping",
		"",
		"- `CLAUDE.md` becomes `AGENTS.md`.",
		"- Claude skills become repo-local Codex skills under `.agents/skills/`.",
		"- Claude command wrappers become command-entry skills such as `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, and `$namba-sync`.",
		"- Claude subagents map to Codex built-in subagents plus project-scoped `.toml` custom agents, with `.md` mirrors kept for readability.",
		"- Claude hooks become explicit validator scripts, documented response contracts, and sync steps in Namba.",
		"- Claude custom workflow commands become `$namba`, command-entry repo skills, built-in Codex slash commands, and the `namba` CLI.",
		"",
		"## Important Distinction",
		"",
		"- In interactive Codex sessions, `namba run SPEC-XXX` means Codex should execute the SPEC directly in-session.",
		"- The standalone `namba run` CLI supports the default runner flow plus explicit `--solo`, `--team`, and worktree-based `--parallel` modes.",
		"- Tokens and PATs are intentionally excluded from generated config. Use `gh auth login` or `glab auth login` instead.",
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderCodexStatusLineExample() string {
	return "[tui]\nstatus_line = [\"project-root\", \"git-branch\", \"current-dir\", \"model-with-reasoning\", \"context-remaining\", \"context-used\", \"used-tokens\", \"session-id\"]\n"
}

func renderClaudeCodexMapping() string {
	lines := []string{
		"# Claude Code to Codex Mapping",
		"",
		"This repository uses a Codex-adapted variant of the MoAI bootstrap model.",
		"",
		"- `CLAUDE.md` -> `AGENTS.md`",
		"- `.claude/skills/*` -> `.agents/skills/*`",
		"- `.claude/commands/*` -> `.agents/skills/namba-*/SKILL.md` command-entry skills",
		"- `.claude/agents/*.md` -> `.codex/agents/*.toml` custom agents with `.md` role-card mirrors",
		"- `.claude/hooks/*` -> explicit validation commands, output-contract validator scripts, structured run logs, and `namba sync`",
		"- Claude slash-command-centric workflows -> built-in Codex slash commands plus `$namba` and `namba`",
		"",
		"Why this is different:",
		"- Claude Code has first-class hooks, subagents, and project slash-command workflows.",
		"- Codex has AGENTS, repo-local skills, command-entry skills, repo-local config, built-in slash commands, and built-in subagent workflows.",
		"- NambaAI keeps the workflow semantics but ports the control surface into Codex-compatible assets.",
	}
	return strings.Join(lines, "\n") + "\n"
}

type outputContractSection struct {
	Emoji   string   `json:"emoji"`
	Primary string   `json:"primary"`
	Aliases []string `json:"aliases"`
}

type outputContractSpec struct {
	Header        string                  `json:"header"`
	HeaderAliases []string                `json:"header_aliases"`
	Sections      []outputContractSection `json:"sections"`
}

func outputContractLanguage(profile initProfile) string {
	return firstNonBlank(profile.ConversationLanguage, profile.DocumentationLanguage, profile.PRLanguage, "en")
}

func outputContractSpecFor(profile initProfile) outputContractSpec {
	switch outputContractLanguage(profile) {
	case "ko":
		return outputContractSpec{
			Header:        "NAMBA-AI 작업 결과 보고",
			HeaderAliases: []string{"NAMBA-AI 작업 결과 보고", "NAMBA-AI 작업 보고", "NAMBA-AI 엔지니어링 보고"},
			Sections: []outputContractSection{
				{Emoji: "🧭", Primary: "작업 정의", Aliases: []string{"작업 정의", "정의", "정의한 범위", "문제 정의"}},
				{Emoji: "🧠", Primary: "판단", Aliases: []string{"판단", "내린 판단", "핵심 판단", "결정"}},
				{Emoji: "🛠", Primary: "수행한 작업", Aliases: []string{"수행한 작업", "진행한 작업", "작업 내용", "적용한 작업"}},
				{Emoji: "🚧", Primary: "현재 이슈", Aliases: []string{"현재 이슈", "이슈", "남은 이슈", "현재 문제"}},
				{Emoji: "⚠", Primary: "잠재 문제", Aliases: []string{"잠재 문제", "잠재 리스크", "위험 요소", "잠재 이슈"}},
				{Emoji: "➡", Primary: "다음 스텝", Aliases: []string{"다음 스텝", "다음 단계", "추천", "권장 흐름"}},
			},
		}
	case "ja":
		return outputContractSpec{
			Header:        "NAMBA-AI 作業結果報告",
			HeaderAliases: []string{"NAMBA-AI 作業結果報告", "NAMBA-AI 作業報告", "NAMBA-AI エンジニアリング報告"},
			Sections: []outputContractSection{
				{Emoji: "🧭", Primary: "作業定義", Aliases: []string{"作業定義", "定義", "スコープ定義", "問題定義"}},
				{Emoji: "🧠", Primary: "判断", Aliases: []string{"判断", "判断内容", "見立て", "決定"}},
				{Emoji: "🛠", Primary: "実施した作業", Aliases: []string{"実施した作業", "対応内容", "実施内容", "作業内容"}},
				{Emoji: "🚧", Primary: "現在の課題", Aliases: []string{"現在の課題", "現在のイシュー", "残課題", "現状の問題"}},
				{Emoji: "⚠", Primary: "潜在リスク", Aliases: []string{"潜在リスク", "潜在課題", "想定リスク", "潜在問題"}},
				{Emoji: "➡", Primary: "次のステップ", Aliases: []string{"次のステップ", "次の一手", "推奨フロー", "次段階"}},
			},
		}
	case "zh":
		return outputContractSpec{
			Header:        "NAMBA-AI 工作结果报告",
			HeaderAliases: []string{"NAMBA-AI 工作结果报告", "NAMBA-AI 工作报告", "NAMBA-AI 工程报告"},
			Sections: []outputContractSection{
				{Emoji: "🧭", Primary: "工作定义", Aliases: []string{"工作定义", "定义", "范围定义", "问题定义"}},
				{Emoji: "🧠", Primary: "判断", Aliases: []string{"判断", "判断结论", "研判", "决定"}},
				{Emoji: "🛠", Primary: "已完成工作", Aliases: []string{"已完成工作", "执行工作", "已做事项", "工作内容"}},
				{Emoji: "🚧", Primary: "当前问题", Aliases: []string{"当前问题", "当前议题", "现有问题", "剩余问题"}},
				{Emoji: "⚠", Primary: "潜在风险", Aliases: []string{"潜在风险", "潜在问题", "风险点", "潜在议题"}},
				{Emoji: "➡", Primary: "下一步", Aliases: []string{"下一步", "建议步骤", "推荐动作", "后续步骤"}},
			},
		}
	default:
		return outputContractSpec{
			Header:        "NAMBA-AI Work Report",
			HeaderAliases: []string{"NAMBA-AI Work Report", "NAMBA-AI Engineering Report", "NAMBA-AI Task Report"},
			Sections: []outputContractSection{
				{Emoji: "🧭", Primary: "Scope", Aliases: []string{"Scope", "Framing", "Problem Framing", "Definition"}},
				{Emoji: "🧠", Primary: "Decision", Aliases: []string{"Decision", "Judgment", "Judgement", "Assessment"}},
				{Emoji: "🛠", Primary: "Work Completed", Aliases: []string{"Work Completed", "Work Done", "Actions Taken", "Completed Work"}},
				{Emoji: "🚧", Primary: "Current Issues", Aliases: []string{"Current Issues", "Open Issues", "Current Gaps", "Current Problems"}},
				{Emoji: "⚠", Primary: "Potential Risks", Aliases: []string{"Potential Risks", "Risks", "Potential Problems", "Risk Boundaries"}},
				{Emoji: "➡", Primary: "Next Steps", Aliases: []string{"Next Steps", "Recommended Next Steps", "Recommendations", "Next Move"}},
			},
		}
	}
}

func outputContractHeaderExample(profile initProfile) string {
	return "# " + outputContractSpecFor(profile).Header
}

func outputContractSequence(profile initProfile) string {
	spec := outputContractSpecFor(profile)
	parts := make([]string, 0, len(spec.Sections))
	for _, section := range spec.Sections {
		parts = append(parts, fmt.Sprintf("`%s %s`", section.Emoji, section.Primary))
	}
	return strings.Join(parts, " -> ")
}

func renderOutputContractDocLocalized(profile initProfile) string {
	spec := outputContractSpecFor(profile)
	lines := []string{
		"# Namba Output Contract",
		"",
		"This repository uses a NambaAI-specific output contract for substantial task responses.",
		"",
		"## Contract",
		"",
		fmt.Sprintf("- Use a decorated header such as `%s`.", outputContractHeaderExample(profile)),
		"- Keep the report sections in this order:",
	}
	for index, section := range spec.Sections {
		lines = append(lines, fmt.Sprintf("%d. `%s %s`", index+1, section.Emoji, section.Primary))
	}
	lines = append(lines,
		"",
		"## Namba Style",
		"",
		fmt.Sprintf("- The header and label palette should follow the init-selected language: %s.", humanLanguageName(outputContractLanguage(profile))),
		"- The semantic order is fixed, but the exact labels may vary within the selected language palette.",
		"- Light visual styling such as simple emoji section markers is encouraged when it improves scanability.",
		"- Recommended label palette:",
	)
	for _, section := range spec.Sections {
		aliases := make([]string, 0, len(section.Aliases))
		for _, alias := range section.Aliases {
			aliases = append(aliases, fmt.Sprintf("`%s`", alias))
		}
		lines = append(lines, fmt.Sprintf("  - `%s %s`: %s", section.Emoji, section.Primary, strings.Join(aliases, ", ")))
	}
	lines = append(lines,
		"- The answer should read like a concise engineering field report rather than a stiff checklist.",
		"",
		"## Scope",
		"",
		"- Apply the full contract to implementation summaries, design decisions, operational guidance, code reviews, and other substantial responses.",
		"- Very short acknowledgements or one-line factual replies may stay shorter, but substantial responses should keep the same semantic order.",
		"",
		"## Validation",
		"",
		"- Use `.namba/codex/validate-output-contract.py --file <response.md>` to validate a saved response.",
		"- Use `.namba/codex/validate-output-contract.py` and pipe UTF-8 text through stdin to validate ad hoc content.",
		"",
		"## Hook Status",
		"",
		"- Namba keeps the validator script as the explicit repository enforcement path even as the documented Codex config and hook surface evolves.",
		"- Treat the validator script as the fallback until Namba deliberately adopts any upstream hook-based enforcement.",
	)
	return strings.Join(lines, "\n") + "\n"
}

func renderOutputContractValidatorLocalized(profile initProfile) string {
	specJSON, _ := json.MarshalIndent(outputContractSpecFor(profile), "", "  ")
	return strings.Join([]string{
		"#!/usr/bin/env python3",
		`"""Validate the NambaAI output contract from stdin or a file."""`,
		"",
		"from __future__ import annotations",
		"",
		"import argparse",
		"import pathlib",
		"import re",
		"import sys",
		"",
		"SPEC = " + string(specJSON),
		"",
		"",
		"def build_pattern(aliases: list[str]) -> re.Pattern[str]:",
		`    escaped = "|".join(re.escape(alias) for alias in aliases)`,
		`    return re.compile(r"^\s*(?:#{1,6}\s*|[-*]\s+)?(?:\*\*)?[\W_]*(?P<label>(" + escaped + r"))(?:\*\*)?\s*(?:[:：-].*)?$")`,
		"",
		"",
		"def read_text(args: argparse.Namespace) -> str:",
		"    if args.file:",
		`        return pathlib.Path(args.file).read_text(encoding="utf-8")`,
		"    return sys.stdin.read().lstrip('\\ufeff')",
		"",
		"",
		"def find_first_match(lines: list[str], aliases: list[str], start: int = 0) -> int:",
		"    pattern = build_pattern(aliases)",
		"    for index, line in enumerate(lines[start:], start=start):",
		"        if pattern.match(line.strip()):",
		"            return index",
		"    return -1",
		"",
		"",
		"def main() -> int:",
		`    parser = argparse.ArgumentParser(description="Validate the NambaAI output contract.")`,
		`    parser.add_argument("--file", help="Path to a saved response file.")`,
		"    args = parser.parse_args()",
		"",
		"    text = read_text(args)",
		"    if not text.strip():",
		`        print("output-contract: empty input", file=sys.stderr)`,
		"        return 1",
		"",
		"    lines = [line.lstrip('\\ufeff') for line in text.lstrip('\\ufeff').splitlines()]",
		"    header_index = find_first_match(lines, SPEC['header_aliases'])",
		"    if header_index < 0:",
		`        print(f"output-contract: missing header '{SPEC['header']}'", file=sys.stderr)`,
		"        return 1",
		"",
		"    previous = header_index",
		"    for section in SPEC['sections']:",
		"        found = find_first_match(lines, section['aliases'], start=previous + 1)",
		"        if found < 0:",
		`            print(f"output-contract: missing section '{section['primary']}'", file=sys.stderr)`,
		"            return 1",
		"        if found <= previous:",
		`            print(f"output-contract: section '{section['primary']}' is out of order", file=sys.stderr)`,
		"            return 1",
		"        previous = found",
		"",
		`    print("output-contract: ok")`,
		"    return 0",
		"",
		"",
		`if __name__ == "__main__":`,
		"    raise SystemExit(main())",
	}, "\n") + "\n"
}

func renderRepoCodexConfig(profile initProfile) string {
	threads := 1
	if profile.AgentMode == "multi" {
		threads = 3
	}

	lines := []string{
		"#:schema https://developers.openai.com/codex/config-schema.json",
		"# Generated by NambaAI from `.namba/config/sections/*.yaml`.",
		"# This file intentionally keeps only repo-safe Codex defaults under version control.",
		"# Keep user-specific settings such as models, auth, apps, web search, permissions profiles,",
		"# and platform-specific sandbox choices in your user-level Codex config.",
		"# Repo-managed MCP presets are the narrow exception when `.namba/config/sections/codex.yaml` opts in.",
		"# Reference: https://developers.openai.com/codex/config-reference/",
		"",
		fmt.Sprintf("approval_policy = %q", approvalPolicy(profile)),
		fmt.Sprintf("sandbox_mode = %q", sandboxMode(profile)),
		"",
		"[agents]",
		fmt.Sprintf("max_threads = %d", threads),
	}
	if profile.StatusLinePreset == "namba" {
		lines = append(lines, "", strings.TrimSpace(renderCodexStatusLineExample()))
	}
	for _, preset := range managedMCPServerPresetsForIDs(profile.DefaultMCPServers) {
		lines = append(lines,
			"",
			fmt.Sprintf("[mcp_servers.%s]", preset.ID),
			fmt.Sprintf("command = %q", preset.Command),
			fmt.Sprintf("args = [%s]", formatTOMLStringArray(preset.Args)),
		)
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderRoleCard(title, useWhen string, responsibilities []string) string {
	lines := []string{
		fmt.Sprintf("# %s", title),
		"",
		useWhen,
		"",
		"Responsibilities:",
	}
	for _, responsibility := range responsibilities {
		lines = append(lines, "- "+responsibility)
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderPlannerRoleCard() string {
	return renderRoleCard(
		"Namba Planner",
		"Use this role when breaking down a SPEC package before implementation.",
		[]string{
			"Read `spec.md`, `plan.md`, and `acceptance.md`.",
			"Identify target files, risks, and validation commands.",
			"Produce a concise execution plan for the main session.",
			"Do not edit files directly.",
		},
	)
}

func renderPlannerCustomAgent() string {
	return renderCustomAgentWithOptions(
		"namba-planner",
		"Break down a SPEC package into an execution plan without editing files.",
		"read-only",
		[]string{
			"You are Namba Planner.",
			"",
			"Use this custom agent when breaking down a SPEC package before implementation.",
			"",
			"Responsibilities:",
			"- Read `spec.md`, `plan.md`, and `acceptance.md`.",
			"- When repo-managed MCP presets are configured, consult them first when they can ground planning decisions with better source material or verification signals.",
			"- Identify target files, risks, and validation commands.",
			"- Produce a concise execution plan for the main session.",
			"- Do not edit files directly.",
		},
	)
}

func renderProductManagerRoleCard() string {
	return renderRoleCard(
		"Namba Product Manager",
		"Use this role when shaping scope, acceptance, and delivery slicing before implementation.",
		[]string{
			"Translate user goals into concrete scope, constraints, and success criteria.",
			"Tighten acceptance criteria, non-goals, and rollout boundaries.",
			"Break large ideas into deliverable slices the main session can schedule.",
			"Call out UX, data, and operational implications early.",
		},
	)
}

func renderProductManagerCustomAgent() string {
	return renderCustomAgentWithOptions(
		"namba-product-manager",
		"Shape product scope, acceptance, and delivery slicing before implementation.",
		"read-only",
		[]string{
			"You are Namba Product Manager.",
			"",
			"Use this custom agent when a request needs stronger product framing before implementation starts.",
			"",
			"Responsibilities:",
			"- Translate user goals into concrete scope, constraints, and success criteria.",
			"- Tighten acceptance criteria, non-goals, and rollout boundaries.",
			"- Break large ideas into deliverable slices the main session can schedule.",
			"- Call out UX, data, and operational implications early.",
			"- Do not implement code directly.",
		},
	)
}

func renderFrontendArchitectRoleCard() string {
	return renderRoleCard(
		"Namba Frontend Architect",
		"Use this role when frontend structure, state flow, or UI delivery planning needs to be clarified before editing.",
		[]string{
			"Identify component boundaries, state ownership, and data flow.",
			"Map UI changes to file targets, design-system constraints, and accessibility impact.",
			"Highlight responsive, performance, and browser-risk considerations.",
			"Recommend the smallest coherent UI implementation slice.",
		},
	)
}

func renderFrontendArchitectCustomAgent() string {
	return renderCustomAgentWithOptions(
		"namba-frontend-architect",
		"Plan frontend structure, component boundaries, state flow, and UI delivery risks.",
		"read-only",
		[]string{
			"You are Namba Frontend Architect.",
			"",
			"Use this custom agent when a task needs frontend planning before implementation starts.",
			"",
			"Responsibilities:",
			"- Identify component boundaries, state ownership, and data flow.",
			"- Map UI changes to file targets, design-system constraints, and accessibility impact.",
			"- Highlight responsive, performance, and browser-risk considerations.",
			"- Recommend the smallest coherent UI implementation slice.",
			"- Do not edit files directly.",
		},
	)
}

func renderFrontendImplementerRoleCard() string {
	return renderRoleCard(
		"Namba Frontend Implementer",
		"Use this role when implementing approved UI work.",
		[]string{
			"Change only the frontend files assigned by the main session.",
			"Preserve design-system conventions, accessibility, and responsive behavior.",
			"Keep loading, empty, and error states coherent with the surrounding UI.",
			"Run or report the relevant UI validation steps when feasible.",
		},
	)
}

func renderFrontendImplementerCustomAgent() string {
	return renderCustomAgentWithOptions(
		"namba-frontend-implementer",
		"Implement approved frontend work with design-system, accessibility, and responsive discipline.",
		"workspace-write",
		[]string{
			"You are Namba Frontend Implementer.",
			"",
			"Use this custom agent when implementing approved UI work.",
			"",
			"Responsibilities:",
			"- Change only the frontend files assigned by the main session.",
			"- Preserve design-system conventions, accessibility, and responsive behavior.",
			"- Keep loading, empty, and error states coherent with the surrounding UI.",
			"- Run or report the relevant UI validation steps when feasible.",
		},
	)
}

func renderDesignerRoleCard() string {
	return renderRoleCard(
		"Namba Designer",
		"Use this role when visual direction, spacing, typography, motion, or interaction intent need to be clarified before implementation.",
		[]string{
			"Define the visual hierarchy, spacing rhythm, typography, and interaction intent for the requested change.",
			"Align the work with design-system tokens, accessibility, and consistent component reuse.",
			"Call out motion, affordance, and layout risks early.",
			"Recommend the smallest design slice that can be implemented without a broader redesign.",
		},
	)
}

func renderDesignerCustomAgent() string {
	return renderCustomAgentWithOptions(
		"namba-designer",
		"Clarify visual direction, spacing, typography, motion, and interaction intent before implementation.",
		"read-only",
		[]string{
			"You are Namba Designer.",
			"",
			"Use this custom agent when a task needs design direction before implementation starts.",
			"",
			"Responsibilities:",
			"- Define the visual hierarchy, spacing rhythm, typography, and interaction intent for the requested change.",
			"- Align the work with design-system tokens, accessibility, and consistent component reuse.",
			"- Call out motion, affordance, and layout risks early.",
			"- Recommend the smallest design slice that can be implemented without a broader redesign.",
			"- Do not edit files directly.",
		},
	)
}

func renderMobileEngineerRoleCard() string {
	return renderRoleCard(
		"Namba Mobile Engineer",
		"Use this role when mobile-specific constraints, navigation, lifecycle, or platform behavior need to be clarified before editing.",
		[]string{
			"Define mobile component boundaries, platform-specific constraints, and ownership of shared versus native behavior.",
			"Map requested changes to navigation, lifecycle, offline, and responsive considerations.",
			"Highlight gesture, performance, and device-compatibility risks.",
			"Recommend the smallest mobile delivery slice the main session can delegate safely.",
		},
	)
}

func renderMobileEngineerCustomAgent() string {
	return renderCustomAgentWithOptions(
		"namba-mobile-engineer",
		"Plan mobile-specific structure, navigation, lifecycle, and platform risks before implementation.",
		"read-only",
		[]string{
			"You are Namba Mobile Engineer.",
			"",
			"Use this custom agent when a task needs mobile-specific planning before implementation starts.",
			"",
			"Responsibilities:",
			"- Define mobile component boundaries, platform-specific constraints, and ownership of shared versus native behavior.",
			"- Map requested changes to navigation, lifecycle, offline, and responsive considerations.",
			"- Highlight gesture, performance, and device-compatibility risks.",
			"- Recommend the smallest mobile delivery slice the main session can delegate safely.",
			"- Do not edit files directly.",
		},
	)
}

func renderBackendArchitectRoleCard() string {
	return renderRoleCard(
		"Namba Backend Architect",
		"Use this role when backend contracts, service boundaries, or persistence changes need to be clarified before implementation.",
		[]string{
			"Define API, service, and persistence boundaries for the requested change.",
			"Call out schema, transaction, idempotency, and rollback risks.",
			"Identify security, observability, and migration implications.",
			"Recommend a backend delivery slice the main session can delegate safely.",
		},
	)
}

func renderBackendArchitectCustomAgent() string {
	return renderCustomAgentWithOptions(
		"namba-backend-architect",
		"Plan backend contracts, service boundaries, persistence changes, and delivery risks.",
		"read-only",
		[]string{
			"You are Namba Backend Architect.",
			"",
			"Use this custom agent when a task needs backend planning before implementation starts.",
			"",
			"Responsibilities:",
			"- Define API, service, and persistence boundaries for the requested change.",
			"- Call out schema, transaction, idempotency, and rollback risks.",
			"- Identify security, observability, and migration implications.",
			"- Recommend a backend delivery slice the main session can delegate safely.",
			"- Do not edit files directly.",
		},
	)
}

func renderBackendImplementerRoleCard() string {
	return renderRoleCard(
		"Namba Backend Implementer",
		"Use this role when implementing approved server-side work.",
		[]string{
			"Change only the backend files assigned by the main session.",
			"Keep API contracts, validation, and persistence logic internally consistent.",
			"Add or update targeted backend tests when the change affects behavior.",
			"Report migration, rollout, or compatibility risks with the patch.",
		},
	)
}

func renderBackendImplementerCustomAgent() string {
	return renderCustomAgentWithOptions(
		"namba-backend-implementer",
		"Implement approved server-side work across APIs, services, persistence, and backend tests.",
		"workspace-write",
		[]string{
			"You are Namba Backend Implementer.",
			"",
			"Use this custom agent when implementing approved server-side work.",
			"",
			"Responsibilities:",
			"- Change only the backend files assigned by the main session.",
			"- Keep API contracts, validation, and persistence logic internally consistent.",
			"- Add or update targeted backend tests when the change affects behavior.",
			"- Report migration, rollout, or compatibility risks with the patch.",
		},
	)
}

func renderDataEngineerRoleCard() string {
	return renderRoleCard(
		"Namba Data Engineer",
		"Use this role when schema, migration, pipeline, analytics, or transformation work is part of the change.",
		[]string{
			"Own data-model, migration, ETL, query, and analytics-facing code assigned by the main session.",
			"Keep schema changes, backfills, and data contracts internally consistent.",
			"Call out rollout sequencing, data quality risks, and irreversible migration concerns.",
			"Add or update focused validation for the changed data behavior when feasible.",
		},
	)
}

func renderDataEngineerCustomAgent() string {
	return renderCustomAgentWithOptions(
		"namba-data-engineer",
		"Handle schema, migration, pipeline, and analytics-facing changes with explicit data-quality and rollout discipline.",
		"workspace-write",
		[]string{
			"You are Namba Data Engineer.",
			"",
			"Use this custom agent when schema, migration, pipeline, analytics, or transformation work is part of the change.",
			"",
			"Responsibilities:",
			"- Own data-model, migration, ETL, query, and analytics-facing code assigned by the main session.",
			"- Keep schema changes, backfills, and data contracts internally consistent.",
			"- Call out rollout sequencing, data quality risks, and irreversible migration concerns.",
			"- Add or update focused validation for the changed data behavior when feasible.",
		},
	)
}

func renderSecurityEngineerRoleCard() string {
	return renderRoleCard(
		"Namba Security Engineer",
		"Use this role when authentication, authorization, secrets, privacy, or hardening work is part of the change.",
		[]string{
			"Own security-sensitive code paths assigned by the main session.",
			"Tighten auth, permission, secret-handling, validation, and privacy boundaries without widening scope.",
			"Call out exploitability, compliance, rollback, and incident-response implications.",
			"Prefer the smallest defensible hardening patch plus explicit regression notes.",
		},
	)
}

func renderSecurityEngineerCustomAgent() string {
	return renderCustomAgentWithOptions(
		"namba-security-engineer",
		"Handle authentication, authorization, secrets, privacy, and hardening work with explicit security discipline.",
		"workspace-write",
		[]string{
			"You are Namba Security Engineer.",
			"",
			"Use this custom agent when authentication, authorization, secrets, privacy, or hardening work is part of the change.",
			"",
			"Responsibilities:",
			"- Own security-sensitive code paths assigned by the main session.",
			"- Tighten auth, permission, secret-handling, validation, and privacy boundaries without widening scope.",
			"- Call out exploitability, compliance, rollback, and incident-response implications.",
			"- Prefer the smallest defensible hardening patch plus explicit regression notes.",
		},
	)
}

func renderTestEngineerRoleCard() string {
	return renderRoleCard(
		"Namba Test Engineer",
		"Use this role when acceptance coverage or regression protection needs to be strengthened.",
		[]string{
			"Turn acceptance criteria into concrete test scenarios and edge cases.",
			"Add the smallest high-value automated coverage for the changed behavior.",
			"Focus on regression detection rather than broad refactors.",
			"Report residual gaps when full automation is not practical.",
		},
	)
}

func renderTestEngineerCustomAgent() string {
	return renderCustomAgentWithOptions(
		"namba-test-engineer",
		"Design and add targeted regression coverage that tightens SPEC acceptance confidence.",
		"workspace-write",
		[]string{
			"You are Namba Test Engineer.",
			"",
			"Use this custom agent when acceptance coverage or regression protection needs to be strengthened.",
			"",
			"Responsibilities:",
			"- Turn acceptance criteria into concrete test scenarios and edge cases.",
			"- Add the smallest high-value automated coverage for the changed behavior.",
			"- Focus on regression detection rather than broad test refactors.",
			"- Report residual gaps when full automation is not practical.",
		},
	)
}

func renderDevOpsEngineerRoleCard() string {
	return renderRoleCard(
		"Namba DevOps Engineer",
		"Use this role when CI, runtime config, deployment, or operational automation is part of the change.",
		[]string{
			"Own pipeline, environment, container, and deployment-file changes assigned by the main session.",
			"Preserve release safety, rollback clarity, and secret-handling boundaries.",
			"Call out observability, operational risk, and environment drift.",
			"Keep infrastructure edits tightly scoped to the requested outcome.",
		},
	)
}

func renderDevOpsEngineerCustomAgent() string {
	return renderCustomAgentWithOptions(
		"namba-devops-engineer",
		"Handle CI/CD, runtime config, deployment, and operational automation with explicit release safety.",
		"workspace-write",
		[]string{
			"You are Namba DevOps Engineer.",
			"",
			"Use this custom agent when CI, runtime config, deployment, or operational automation is part of the change.",
			"",
			"Responsibilities:",
			"- Own pipeline, environment, container, and deployment-file changes assigned by the main session.",
			"- Preserve release safety, rollback clarity, and secret-handling boundaries.",
			"- Call out observability, operational risk, and environment drift.",
			"- Keep infrastructure edits tightly scoped to the requested outcome.",
		},
	)
}

func renderImplementerRoleCard() string {
	return renderRoleCard(
		"Namba Implementer",
		"Use this role when implementing an approved portion of a SPEC package.",
		[]string{
			"Change only the files assigned by the main session.",
			"Preserve methodology rules from `.namba/config/sections/quality.yaml`.",
			"Run or report the relevant validation steps when feasible.",
			"Leave notes about validation status and residual risk.",
		},
	)
}

func renderImplementerCustomAgent() string {
	return renderCustomAgentWithOptions(
		"namba-implementer",
		"Implement approved SPEC work while preserving Namba quality rules.",
		"workspace-write",
		[]string{
			"You are Namba Implementer.",
			"",
			"Use this custom agent when implementing an approved portion of a SPEC package.",
			"",
			"Responsibilities:",
			"- Change only the files assigned by the main session.",
			"- Preserve methodology rules from `.namba/config/sections/quality.yaml`.",
			"- Run or report the relevant validation steps when feasible.",
			"- Leave notes about validation status and residual risk.",
		},
	)
}

func renderReviewerRoleCard() string {
	return renderRoleCard(
		"Namba Reviewer",
		"Use this role for acceptance and quality review before sync.",
		[]string{
			"Compare the implementation with `acceptance.md`.",
			"Check that validation output and artifacts exist.",
			"Call out regressions, missing tests, or documentation drift.",
			"Do not rewrite the implementation unless asked.",
		},
	)
}

func renderReviewerCustomAgent() string {
	return renderCustomAgentWithOptions(
		"namba-reviewer",
		"Review implementation quality and acceptance coverage before sync.",
		"read-only",
		[]string{
			"You are Namba Reviewer.",
			"",
			"Use this custom agent for acceptance and quality review before sync.",
			"",
			"Responsibilities:",
			"- Compare the implementation with `acceptance.md`.",
			"- Check that validation output and expected artifacts exist.",
			"- Call out regressions, missing tests, and documentation drift.",
			"- Do not rewrite the implementation unless explicitly asked.",
		},
	)
}

func renderCustomAgentWithOptions(name, description, sandboxMode string, developerInstructions []string) string {
	lines := []string{
		fmt.Sprintf(`name = "%s"`, name),
		fmt.Sprintf(`description = "%s"`, description),
	}
	if strings.TrimSpace(sandboxMode) != "" {
		lines = append(lines, fmt.Sprintf(`sandbox_mode = "%s"`, sandboxMode))
	}
	profile := runtimeProfileForAgent(name)
	if strings.TrimSpace(profile.Model) != "" {
		lines = append(lines, fmt.Sprintf(`model = "%s"`, profile.Model))
	}
	if strings.TrimSpace(profile.ModelReasoningEffort) != "" {
		lines = append(lines, fmt.Sprintf(`model_reasoning_effort = "%s"`, profile.ModelReasoningEffort))
	}
	lines = append(lines, `developer_instructions = """`)
	lines = append(lines, developerInstructions...)
	lines = append(lines, `"""`)
	return strings.Join(lines, "\n") + "\n"
}

func renderCustomAgent(name, description string, developerInstructions []string) string {
	return renderCustomAgentWithOptions(name, description, "", developerInstructions)
}

func renderProjectConfig(profile initProfile) string {
	return fmt.Sprintf("name: %s\nproject_type: %s\nlanguage: %s\nframework: %s\ncreated_at: %s\n", profile.ProjectName, profile.ProjectType, profile.Language, normalizeFramework(profile.Framework), profile.CreatedAt)
}

func renderQualityConfig(mode, testCmd, lintCmd, typecheckCmd string) string {
	return fmt.Sprintf(
		"development_mode: %s\ntest_command: %s\nlint_command: %s\ntypecheck_command: %s\nbuild_command: none\nmigration_dry_run_command: none\nsmoke_start_command: none\noutput_contract_command: none\n",
		mode,
		testCmd,
		lintCmd,
		typecheckCmd,
	)
}

func renderWorkflowConfig() string {
	return "default_parallel: false\nmax_parallel_workers: 3\nparallel_acceptance_threshold: 3\n"
}

func renderSystemConfig(profile initProfile) string {
	return fmt.Sprintf("runner: codex\napproval_policy: %s\nsandbox_mode: %s\n", approvalPolicy(profile), sandboxMode(profile))
}

func renderLanguageConfig(profile initProfile) string {
	return fmt.Sprintf(
		"conversation_language: %s\ndocumentation_language: %s\ncomment_language: %s\n",
		profile.ConversationLanguage,
		profile.DocumentationLanguage,
		profile.CommentLanguage,
	)
}

func renderDocsConfig(profile initProfile) string {
	cfg := defaultDocsConfig(profile.ProjectType)
	return fmt.Sprintf(
		"manage_readme: %t\nreadme_profile: %s\nreadme_default_language: %s\nreadme_additional_languages: %s\nreadme_hero_image: %s\n",
		cfg.ManageReadme,
		cfg.ReadmeProfile,
		cfg.DefaultLanguage,
		strings.Join(cfg.AdditionalLanguages, ","),
		cfg.HeroImage,
	)
}

func renderUserConfig(profile initProfile) string {
	return fmt.Sprintf("user_name: %s\n", profile.UserName)
}

func renderGitStrategyConfig(profile initProfile) string {
	return fmt.Sprintf(
		"git_mode: %s\ngit_provider: %s\ngit_username: %s\ngitlab_instance_url: %s\nstore_tokens: false\nbranch_per_work: %t\nbranch_base: %s\nspec_branch_prefix: %s\ntask_branch_prefix: %s\npr_base_branch: %s\npr_language: %s\ncodex_review_comment: %q\nauto_codex_review: %t\n",
		profile.GitMode,
		profile.GitProvider,
		profile.GitUsername,
		profile.GitLabInstanceURL,
		profile.BranchPerWork,
		branchBase(profile),
		specBranchPrefix(profile),
		taskBranchPrefix(profile),
		prBaseBranch(profile),
		firstNonBlank(profile.PRLanguage, profile.DocumentationLanguage, profile.ConversationLanguage, "en"),
		codexReviewComment(profile),
		profile.AutoCodexReview,
	)
}

func branchBase(profile initProfile) string {
	return firstNonBlank(profile.BranchBase, "main")
}

func specBranchPrefix(profile initProfile) string {
	return firstNonBlank(profile.SpecBranchPrefix, "spec/")
}

func taskBranchPrefix(profile initProfile) string {
	return firstNonBlank(profile.TaskBranchPrefix, "task/")
}

func prBaseBranch(profile initProfile) string {
	return firstNonBlank(profile.PRBaseBranch, branchBase(profile))
}

func codexReviewComment(profile initProfile) string {
	return firstNonBlank(profile.CodexReviewComment, "@codex review")
}

func approvalPolicy(profile initProfile) string {
	return firstNonBlank(profile.ApprovalPolicy, "on-request")
}

func sandboxMode(profile initProfile) string {
	return firstNonBlank(profile.SandboxMode, "workspace-write")
}

func humanLanguageName(code string) string {
	switch firstNonBlank(code, "en") {
	case "ko":
		return "Korean"
	case "ja":
		return "Japanese"
	case "zh":
		return "Chinese"
	default:
		return "English"
	}
}

func renderCollaborationPolicy(profile initProfile) string {
	lines := []string{
		fmt.Sprintf("- Start each new SPEC or task on a dedicated branch from `%s`.", branchBase(profile)),
		fmt.Sprintf("- Use `%s<SPEC-ID>-<slug>` for SPEC work and `%s<slug>` for other work when practical.", specBranchPrefix(profile), taskBranchPrefix(profile)),
		fmt.Sprintf("- Commit on the work branch and open PRs into `%s`.", prBaseBranch(profile)),
		fmt.Sprintf("- Write GitHub PR titles and bodies in %s.", humanLanguageName(profile.PRLanguage)),
	}
	if profile.AutoCodexReview {
		lines = append(lines, fmt.Sprintf("- After the PR is open on GitHub, confirm the `%s` review request comment exists instead of duplicating it.", codexReviewComment(profile)))
	} else {
		lines = append(lines, fmt.Sprintf("- After the PR is open on GitHub, request Codex review with `%s`.", codexReviewComment(profile)))
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderCodexProfileConfig(profile initProfile) string {
	lines := []string{
		fmt.Sprintf("agent_mode: %s", profile.AgentMode),
		fmt.Sprintf("status_line_preset: %s", profile.StatusLinePreset),
		fmt.Sprintf("repo_skills_path: %s", repoSkillsDir),
		fmt.Sprintf("repo_agents_path: %s", repoCodexAgentsDir),
		"# Optional: per-run Codex overrides resolved by `namba run`.",
		"# model:",
		"# profile:",
		"web_search: false",
		"add_dirs:",
		"session_mode: stateful",
		"repair_attempts: 1",
		"required_env:",
		"requires_network: false",
		"# Optional: comma-separated Namba-managed MCP presets to render into `.codex/config.toml`.",
		fmt.Sprintf("# Supported values: %s", strings.Join(supportedManagedMCPServerIDs(), ", ")),
	}
	if len(profile.DefaultMCPServers) == 0 {
		lines = append(lines, "# default_mcp_servers:")
	} else {
		lines = append(lines, fmt.Sprintf("default_mcp_servers: %s", strings.Join(profile.DefaultMCPServers, ",")))
	}
	return strings.Join(lines, "\n") + "\n"
}
