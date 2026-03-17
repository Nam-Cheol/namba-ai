package namba

import (
	"fmt"
	"strings"
)

func renderAgents(profile initProfile) string {
	collab := renderCollaborationPolicy(profile)
	return fmt.Sprintf("# NambaAI\n\n"+
		"You are the NambaAI orchestrator for this repository.\n\n"+
		"## Codex-Native Mode\n\n"+
		"When the user references `namba`, `namba project`, `namba regen`, `namba update`, `namba plan`, `namba fix`, `namba run SPEC-XXX`, or `namba sync`, treat those as Namba workflow commands inside the current Codex session.\n\n"+
		"- Prefer direct Codex-native execution for `namba run SPEC-XXX`: read the SPEC package, implement the work in-session, run validation, and sync artifacts.\n"+
		"- Use the installed `namba` CLI for `init`, `doctor`, `project`, `regen`, `update`, `plan`, `fix`, and `sync` when it is available and the command should mutate repo state or maintain the installed CLI directly.\n"+
		"- If the `namba` CLI is unavailable, perform the equivalent workflow manually with `.namba/` as the source of truth.\n"+
		"- Use repo skills under `.agents/skills/` as the single skill surface. Command-entry skills such as `$namba-run` and `$namba-plan` replace provider-specific custom command wrappers.\n"+
		"- When delegating work with Codex multi-agent features, use custom agents under `.codex/agents/*.toml` and keep `.md` role cards as readable mirrors.\n\n"+
		"## Workflow\n\n"+
		"1. Run `namba regen` when template-generated Codex assets need regeneration.\n"+
		"2. Run `namba project` to refresh project docs and codemaps.\n"+
		"3. Run `namba plan \"<description>\"` for feature work or `namba fix \"<description>\"` for bug fixes.\n"+
		"4. Run `namba run SPEC-XXX` to execute the SPEC with Codex-native workflow.\n"+
		"5. Run `namba sync` to refresh artifacts and PR-ready documents.\n\n"+
		"## Collaboration Policy\n\n"+
		"%s\n"+
		"## Rules\n\n"+
		"- Prefer `.namba/` as the source of truth.\n"+
		"- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before implementation.\n"+
		"- Use `$namba` for general routing, or command-entry skills such as `$namba-run`, `$namba-plan`, `$namba-project`, and `$namba-sync` when the user invokes one command directly.\n"+
		"- Do not bypass validation. Run the configured quality commands after changes.\n"+
		"- Use worktrees for parallel execution; do not modify multiple branches in one workspace.\n\n"+
		"Project: %s\n"+
		"Methodology: %s\n"+
		"Agent mode: %s\n",
		collab,
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
		"Use this skill whenever the user mentions `namba`, `namba project`, `namba regen`, `namba update`, `namba plan`, `namba fix`, `namba run`, `namba sync`, or asks to use the Namba workflow.",
		"",
		"Command mapping:",
		"- `namba project`: refresh repository docs and codemaps.",
		"- `namba regen`: regenerate AGENTS, repo-local skills, command-entry skills, Codex custom agents, readable role cards, and repo-local Codex config from `.namba/config/sections/*.yaml`.",
		"- `namba update [--version vX.Y.Z]`: self-update the installed `namba` binary from GitHub Release assets.",
		"- `namba plan \"<description>\"`: create the next feature SPEC package under `.namba/specs/`.",
		"- `namba fix \"<description>\"`: create the next bugfix SPEC package under `.namba/specs/`.",

		"- `namba run SPEC-XXX`: execute the SPEC in the current Codex session. Read `spec.md`, `plan.md`, and `acceptance.md`, implement directly, validate, and sync artifacts.",
		"- `namba sync`: refresh change summary, PR checklist, and codemaps after implementation.",
		"- `namba doctor`: verify that AGENTS, repo skills, `.namba` config, Codex CLI, and the global `namba` command are available.",
		"",
		"Execution rules:",
		"1. Treat `.namba/` as the source of truth.",
		"2. Prefer repo-local skills in `.agents/skills/`.",
		"3. Prefer command-entry skills such as `$namba-run`, `$namba-plan`, `$namba-project`, and `$namba-sync` when the user is invoking one Namba command directly.",
		"4. Use the installed `namba` CLI for `project`, `regen`, `update`, `plan`, `fix`, and `sync` when it will update repo state more reliably or self-update the installed CLI directly.",
		"5. For `namba run` in an interactive Codex session, prefer Codex-native in-session execution over recursively calling `namba run`.",
		"6. Run validation commands from `.namba/config/sections/quality.yaml` before finishing.",
		"7. Start each new SPEC or task on a dedicated work branch when `.namba/config/sections/git-strategy.yaml` enables branch-per-work collaboration.",
		fmt.Sprintf("8. Prepare PRs against `%s`, write the title/body in %s, and request GitHub Codex review with `%s` when the review flow is enabled.", prBaseBranch(profile), humanLanguageName(profile.PRLanguage), codexReviewComment(profile)),
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
			"- Create the next sequential `SPEC-XXX` package under `.namba/specs/`.",
			"- Keep the scope concrete and implementation-ready.",
		},
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
			"- In an interactive Codex session, prefer Codex-native in-session execution over recursively calling `namba run`.",
			"- Only use the standalone CLI runner for `--parallel`, `--dry-run`, or when the user explicitly wants the non-interactive runner path.",
			"- Run validation commands from `.namba/config/sections/quality.yaml` and finish with `namba sync`.",
			fmt.Sprintf("- Collaboration defaults: branch from `%s`, open the PR into `%s`, write the PR in %s, and request `%s` on GitHub after the PR is open.", branchBase(profile), prBaseBranch(profile), humanLanguageName(profile.PRLanguage), codexReviewComment(profile)),
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
			"- Refresh `.namba/project/*` docs, release notes/checklists, and codemaps after implementation.",
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
		"- Claude custom slash-command workflows -> built-in Codex slash commands plus repo skills such as `$namba-run`, `$namba-plan`, `$namba-sync`, and the `namba` CLI",
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
		"4. Implement the work directly in the current Codex session",
		"5. Run configured validation commands",
		"6. Summarize results in `.namba/logs` and sync artifacts",
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
		"- Creates repo-local skills under `.agents/skills/`, including command-entry skills such as `namba-run`, `namba-plan`, and `namba-sync`.",
		"- Creates Codex custom agents under `.codex/agents/*.toml` and readable `.md` role-card mirrors.",
		"- Creates repo-local Codex config under `.codex/config.toml`, including the selected `approval_policy` and `sandbox_mode`.",
		"- Creates `.namba/` project state, configs, docs, and SPEC storage.",
		"",
		"## How Codex Uses Namba After Init",
		"",
		"1. Open Codex in the initialized project directory.",
		"2. Codex loads `AGENTS.md` and repo skills.",
		"3. Invoke `$namba` for routing or command-entry skills such as `$namba-run`, `$namba-plan`, and `$namba-sync` for direct command-style execution.",
		"4. Use built-in Codex delegation with `.codex/agents/*.toml` custom agents when multi-agent work is appropriate. The matching `.md` files remain readable mirrors.",
		"5. Use `namba project`, `namba regen`, `namba update`, `namba plan`, `namba fix`, `namba run SPEC-XXX`, and `namba sync` as workflow commands.",
		"",
		"## Workflow Command Semantics",
		"",
		"- `namba regen` regenerates `AGENTS.md`, repo skills under `.agents/skills/`, `.codex/agents/*.toml` custom agents, readable `.md` role-card mirrors, `.namba/codex/*`, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.",
		"- `namba update` self-updates the installed `namba` binary from GitHub Release assets. Use `--version vX.Y.Z` for a specific release.",
		"- `namba sync` refreshes `.namba/project/*` docs, release notes/checklists, and codemaps.",
		"- `namba release` requires a clean `main` branch and passing validators before it creates a tag. `--push` pushes both `main` and the new tag.",
		"- `namba run SPEC-XXX --parallel` refers to the standalone runner path. It uses git worktrees, merges only after every worker passes execution and validation, and preserves failed worktrees and branches for inspection.",
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
		"- Claude command wrappers become command-entry skills such as `$namba-run`, `$namba-plan`, and `$namba-sync`.",
		"- Claude subagents become explicit `.toml` custom agents used with Codex multi-agent delegation, with `.md` mirrors kept for readability.",
		"- Claude hooks become explicit validator and sync steps in Namba.",
		"- Claude custom workflow commands become `$namba`, command-entry repo skills, built-in Codex slash commands, and the `namba` CLI.",
		"",
		"## Important Distinction",
		"",
		"- In interactive Codex sessions, `namba run SPEC-XXX` means Codex should execute the SPEC directly in-session.",
		"- The standalone `namba run` CLI remains available for non-interactive runner-based execution.",
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
		"- `.claude/hooks/*` -> explicit validation commands, structured run logs, and `namba sync`",
		"- Claude slash-command-centric workflows -> built-in Codex slash commands plus `$namba` and `namba`",
		"",
		"Why this is different:",
		"- Claude Code has first-class hooks, subagents, and project slash-command workflows.",
		"- Codex has AGENTS, repo-local skills, command-entry skills, repo-local config, built-in slash commands, and experimental multi-agent delegation.",
		"- NambaAI keeps the workflow semantics but ports the control surface into Codex-compatible assets.",
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderRepoCodexConfig(profile initProfile) string {
	threads := 1
	if profile.AgentMode == "multi" {
		threads = 3
	}

	lines := []string{
		fmt.Sprintf("approval_policy = %q", approvalPolicy(profile)),
		fmt.Sprintf("sandbox_mode = %q", sandboxMode(profile)),
		"",
		"[agents]",
		fmt.Sprintf("max_threads = %d", threads),
	}
	if profile.StatusLinePreset == "namba" {
		lines = append(lines, "", strings.TrimSpace(renderCodexStatusLineExample()))
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderPlannerRoleCard() string {
	lines := []string{
		"# Namba Planner",
		"",
		"Use this role when breaking down a SPEC package before implementation.",
		"",
		"Responsibilities:",
		"- Read `spec.md`, `plan.md`, and `acceptance.md`.",
		"- Identify target files, risks, and validation commands.",
		"- Produce a concise execution plan for the main session.",
		"- Do not edit files directly.",
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderPlannerCustomAgent() string {
	return renderCustomAgent(
		"namba-planner",
		"Break down a SPEC package into an execution plan without editing files.",
		[]string{
			"You are Namba Planner.",
			"",
			"Use this custom agent when breaking down a SPEC package before implementation.",
			"",
			"Responsibilities:",
			"- Read `spec.md`, `plan.md`, and `acceptance.md`.",
			"- Identify target files, risks, and validation commands.",
			"- Produce a concise execution plan for the main session.",
			"- Do not edit files directly.",
		},
	)
}

func renderImplementerRoleCard() string {
	lines := []string{
		"# Namba Implementer",
		"",
		"Use this role when implementing an approved portion of a SPEC package.",
		"",
		"Responsibilities:",
		"- Change only the files assigned by the main session.",
		"- Preserve methodology rules from `.namba/config/sections/quality.yaml`.",
		"- Leave notes about validation status and residual risk.",
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderImplementerCustomAgent() string {
	return renderCustomAgent(
		"namba-implementer",
		"Implement approved SPEC work while preserving Namba quality rules.",
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
	lines := []string{
		"# Namba Reviewer",
		"",
		"Use this role for acceptance and quality review before sync.",
		"",
		"Responsibilities:",
		"- Compare the implementation with `acceptance.md`.",
		"- Check that validation output and artifacts exist.",
		"- Call out regressions, missing tests, or documentation drift.",
		"- Do not rewrite the implementation unless asked.",
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderReviewerCustomAgent() string {
	return renderCustomAgent(
		"namba-reviewer",
		"Review implementation quality and acceptance coverage before sync.",
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

func renderCustomAgent(name, description string, developerInstructions []string) string {
	lines := []string{
		fmt.Sprintf(`name = "%s"`, name),
		fmt.Sprintf(`description = "%s"`, description),
		`developer_instructions = """`,
	}
	lines = append(lines, developerInstructions...)
	lines = append(lines, `"""`)
	return strings.Join(lines, "\n") + "\n"
}

func renderProjectConfig(profile initProfile) string {
	return fmt.Sprintf("name: %s\nproject_type: %s\nlanguage: %s\nframework: %s\ncreated_at: %s\n", profile.ProjectName, profile.ProjectType, profile.Language, normalizeFramework(profile.Framework), profile.CreatedAt)
}

func renderQualityConfig(mode, testCmd, lintCmd, typecheckCmd string) string {
	return fmt.Sprintf(
		"development_mode: %s\ntest_command: %s\nlint_command: %s\ntypecheck_command: %s\n",
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
	return fmt.Sprintf(
		"agent_mode: %s\nstatus_line_preset: %s\nrepo_skills_path: %s\nrepo_agents_path: %s\n",
		profile.AgentMode,
		profile.StatusLinePreset,
		repoSkillsDir,
		repoCodexAgentsDir,
	)
}
