package namba

import (
	"fmt"
	"strings"
)

func renderAgents(profile initProfile) string {
	return fmt.Sprintf("# NambaAI\n\n"+
		"You are the NambaAI orchestrator for this repository.\n\n"+
		"## Codex-Native Mode\n\n"+
		"When the user references `namba`, `namba project`, `namba update`, `namba plan`, `namba fix`, `namba run SPEC-XXX`, or `namba sync`, treat those as Namba workflow commands inside the current Codex session.\n\n"+
		"- Prefer direct Codex-native execution for `namba run SPEC-XXX`: read the SPEC package, implement the work in-session, run validation, and sync artifacts.\n"+
		"- Use the installed `namba` CLI for `init`, `doctor`, `project`, `update`, `plan`, `fix`, and `sync` when it is available and the command will update repository state more reliably.\n"+
		"- If the `namba` CLI is unavailable, perform the equivalent workflow manually with `.namba/` as the source of truth.\n"+
		"- Use repo skills under `.agents/skills/` first. `.codex/skills/` exists as a compatibility mirror.\n"+
		"- When delegating work with Codex multi-agent features, use the role cards under `.codex/agents/` as the agent prompt source.\n\n"+
		"## Workflow\n\n"+
		"1. Run `namba update` when template-generated Codex assets need regeneration.\n"+
		"2. Run `namba project` to refresh project docs and codemaps.\n"+
		"3. Run `namba plan \"<description>\"` for feature work or `namba fix \"<description>\"` for bug fixes.\n"+
		"4. Run `namba run SPEC-XXX` to execute the SPEC with Codex-native workflow.\n"+
		"5. Run `namba sync` to refresh artifacts and PR-ready documents.\n\n"+
		"## Rules\n\n"+
		"- Prefer `.namba/` as the source of truth.\n"+
		"- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before implementation.\n"+
		"- Use the `$namba` skill as the primary command surface when the user explicitly invokes Namba inside Codex; treat aliases like `$namba-run` as triggers for the same workflow.\n"+
		"- Do not bypass validation. Run the configured quality commands after changes.\n"+
		"- Use worktrees for parallel execution; do not modify multiple branches in one workspace.\n\n"+
		"Project: %s\n"+
		"Methodology: %s\n"+
		"Agent mode: %s\n",
		profile.ProjectName,
		profile.DevelopmentMode,
		profile.AgentMode,
	)
}

func renderNambaSkill() string {
	lines := []string{
		"---",
		"name: namba",
		"description: Codex-native Namba command surface for SPEC orchestration inside a repository.",
		"---",
		"",
		"Use this skill whenever the user mentions `namba`, `namba project`, `namba update`, `namba plan`, `namba fix`, `namba run`, `namba sync`, aliases like `$namba-run`, or asks to use the Namba workflow.",
		"",
		"Command mapping:",
		"- `namba project`: refresh repository docs and codemaps.",
		"- `namba update`: regenerate AGENTS, repo-local skills, compatibility skills, role cards, and repo-local Codex config from `.namba/config/sections/*.yaml`.",
		"- `namba plan \"<description>\"`: create the next feature SPEC package under `.namba/specs/`.",
		"- `namba fix \"<description>\"`: create the next bugfix SPEC package under `.namba/specs/`.",

		"- `namba run SPEC-XXX`: execute the SPEC in the current Codex session. Read `spec.md`, `plan.md`, and `acceptance.md`, implement directly, validate, and sync artifacts.",
		"- `namba sync`: refresh change summary, PR checklist, and codemaps after implementation.",
		"- `namba doctor`: verify that AGENTS, repo skills, `.namba` config, Codex CLI, and the global `namba` command are available.",
		"",
		"Execution rules:",
		"1. Treat `.namba/` as the source of truth.",
		"2. Prefer repo-local skills in `.agents/skills/`.",
		"3. Use the installed `namba` CLI for `project`, `update`, `plan`, `fix`, and `sync` when it will update repo state more reliably.",
		"4. For `namba run` in an interactive Codex session, prefer Codex-native in-session execution over recursively calling `namba run`.",
		"5. Run validation commands from `.namba/config/sections/quality.yaml` before finishing.",
	}
	return strings.Join(lines, "\n") + "\n"
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
		"- `.claude/skills/*` -> `.agents/skills/*` with optional `.codex/skills/*` compatibility mirror",
		"- `.claude/agents/*` -> `.codex/agents/*.md` role cards for Codex delegation",
		"- `.claude/hooks/*` -> explicit validation pipeline and `namba` orchestration",
		"- Claude custom slash-command workflows -> built-in Codex slash commands plus the `$namba` skill and `namba` CLI",
		"",
		"When implementing init changes:",
		"1. Keep `.namba/config/sections/*.yaml` as the durable source of truth.",
		"2. Never write tokens or secrets into generated config files.",
		"3. Prefer repo-local skills and agent role cards over provider-specific hidden state.",
		"4. Keep generated assets readable so users can understand what `namba init .` changed.",
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

func renderExecutionSkill() string {
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
		"- Creates repo-local skills under `.agents/skills/`.",
		"- Optionally creates a compatibility mirror when `compat_skills_path` is configured.",
		"- Creates Codex delegation role cards under `.codex/agents/`.",
		"- Creates repo-local Codex config under `.codex/config.toml`.",
		"- Creates `.namba/` project state, configs, docs, and SPEC storage.",
		"",
		"## How Codex Uses Namba After Init",
		"",
		"1. Open Codex in the initialized project directory.",
		"2. Codex loads `AGENTS.md` and repo skills.",
		"3. Invoke `$namba` or ask Codex to use the Namba workflow.",
		"4. Use built-in Codex delegation with the role cards in `.codex/agents/` when multi-agent work is appropriate.",
		"5. Use `namba project`, `namba update`, `namba plan`, `namba fix`, `namba run SPEC-XXX`, and `namba sync` as workflow commands.",
		"",
		"## Workflow Command Semantics",
		"",
		"- `namba update` regenerates `AGENTS.md`, repo-local skills, optional compatibility mirror skills, role cards, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.",
		"- `namba sync` refreshes `.namba/project/*` docs, release notes/checklists, and codemaps.",
		"- `namba release` requires a clean `main` branch and passing validators before it creates a tag. `--push` pushes both `main` and the new tag.",
		"- `namba run SPEC-XXX --parallel` refers to the standalone runner path. It uses git worktrees, merges only after every worker passes execution and validation, and preserves failed worktrees and branches for inspection.",
		"",
		"## Claude to Codex Mapping",
		"",
		"- `CLAUDE.md` becomes `AGENTS.md`.",
		"- Claude skills become repo-local Codex skills.",
		"- Claude subagents become explicit role-card files used with Codex multi-agent delegation.",
		"- Claude hooks become explicit validator and sync steps in Namba.",
		"- Claude custom workflow commands become `$namba`, built-in Codex slash commands, and the `namba` CLI.",
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
		"- `.claude/skills/*` -> `.agents/skills/*` (optional compatibility mirror path supported)",
		"- `.claude/agents/*.md` -> `.codex/agents/*.md` role cards",
		"- `.claude/hooks/*` -> explicit validation commands, structured run logs, and `namba sync`",
		"- Claude slash-command-centric workflows -> built-in Codex slash commands plus `$namba` and `namba`",
		"",
		"Why this is different:",
		"- Claude Code has first-class hooks, subagents, and project slash-command workflows.",
		"- Codex has AGENTS, repo-local skills, repo-local config, built-in slash commands, and experimental multi-agent delegation.",
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

func renderSystemConfig() string {
	return "runner: codex\napproval_mode: on-request\nsandbox_mode: workspace-write\n"
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
		"git_mode: %s\ngit_provider: %s\ngit_username: %s\ngitlab_instance_url: %s\nstore_tokens: false\n",
		profile.GitMode,
		profile.GitProvider,
		profile.GitUsername,
		profile.GitLabInstanceURL,
	)
}

func renderCodexProfileConfig(profile initProfile) string {
	return fmt.Sprintf(
		"agent_mode: %s\nstatus_line_preset: %s\nrepo_skills_path: %s\ncompat_skills_path: %s\nrepo_agents_path: %s\n",
		profile.AgentMode,
		profile.StatusLinePreset,
		repoSkillsDir,
		strings.TrimSpace(profile.CompatSkillsPath),
		repoCodexAgentsDir,
	)
}
