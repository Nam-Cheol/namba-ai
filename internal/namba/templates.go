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
		"When the user references `namba`, `namba help`, `namba project`, `namba regen`, `namba update`, `namba codex access`, `namba plan`, `namba harness`, `namba fix`, `namba run SPEC-XXX`, `namba queue`, `namba sync`, `namba pr`, `namba land`, `namba release`, `$namba-coach`, `$namba-create`, `$namba-queue`, `$namba-review-resolve`, or `$namba-release`, treat those as Namba workflow commands or guidance entry points inside the current Codex session.\n\n"+
		"- Prefer direct Codex-native execution for `namba run SPEC-XXX`: read the SPEC package, implement the work in-session, run validation, and sync artifacts.\n"+
		"- Use the installed `namba` CLI for `init`, `doctor`, `project`, `regen`, `update`, `codex access`, `plan`, `harness`, `fix`, `queue`, `pr`, `land`, `release`, and `sync` when it is available and the command should mutate repo state or maintain the installed CLI directly.\n"+
		"- If the `namba` CLI is unavailable, perform the equivalent workflow manually with `.namba/` as the source of truth.\n"+
		"- Use repo skills under `.agents/skills/` as the single skill surface. Command-entry and guidance skills such as `$namba-help`, `$namba-coach`, `$namba-create`, `$namba-run`, `$namba-queue`, `$namba-pr`, `$namba-land`, `$namba-release`, `$namba-plan`, `$namba-plan-review`, `$namba-harness`, `$namba-fix`, `$namba-review-resolve`, and the plan-review skills under `.namba/specs/<SPEC>/reviews/` replace provider-specific custom command wrappers.\n"+
		"- When delegating work with Codex multi-agent features, use built-in subagents such as `default`, `worker`, and `explorer`, plus project-scoped custom agents under `.codex/agents/*.toml`; keep `.md` role cards as readable mirrors.\n\n"+
		"## Workflow\n\n"+
		"1. Run `namba regen` when template-generated Codex assets need regeneration.\n"+
		"2. Run `namba project` to refresh project docs and codemaps.\n"+
		"3. Use `$namba-coach` when the user's current goal is vague or a command choice may need correction before handing off to one workflow.\n"+
		"4. Use `$namba-create` when you need to create a repo-local skill, a project-scoped custom agent, or both through the preview-first Codex-native creation flow.\n"+
		"5. Run `namba plan \"<description>\"` for feature planning and, unless `--no-review` is present, continue immediately with `$namba-plan-review SPEC-XXX` after the SPEC is created. Use `namba harness \"<description>\"` for harness-oriented planning, `namba fix --command plan \"<issue description>\"` for bugfix SPEC planning, or `namba fix \"<issue description>\"` for direct repair.\n"+
		"6. Run the relevant plan-review skills when review is not automatic, or use `$namba-plan-review` when you want the create-plus-review loop bundled, and keep `.namba/specs/<SPEC>/reviews/readiness.md` current when the SPEC needs product, engineering, or design sign-off.\n"+
		"7. Run `namba run SPEC-XXX` to execute the SPEC with Codex-native workflow.\n"+
		"8. Use `$namba-queue` or `namba queue start <SPEC-RANGE|SPEC-LIST>` when existing SPEC packages should move through the conveyor one active SPEC at a time.\n"+
		"9. Run `namba sync` to refresh artifacts and PR-ready documents.\n"+
		"10. Run `namba pr \"<title>\"` to prepare the GitHub review handoff.\n"+
		"11. Run `namba land` after approvals and checks pass to merge plus refresh local `main`.\n\n"+
		"12. Use `$namba-review-resolve` after GitHub review feedback arrives: inspect thread-aware GitHub state, fix meaningful comments, reply on the original threads with validation and CI/check evidence when relevant, resolve only addressed threads, and request review again without duplicating the configured marker.\n"+
		"13. Use `$namba-release` for NambaAI release-note handoff and release orchestration: start from clean `main`, generate commit-based notes, write `.namba/releases/<version>.md`, validate, then use the guarded `namba release` path.\n\n"+
		"## Collaboration Policy\n\n"+
		"%s\n"+
		"## Rules\n\n"+
		"- Prefer `.namba/` as the source of truth.\n"+
		"- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before implementation.\n"+
		"- Use `$namba` for general routing, `$namba-coach` for read-only current-goal command coaching, `$namba-help` for read-only onboarding and command semantics, or command-entry skills such as `$namba-create`, `$namba-run`, `$namba-queue`, `$namba-pr`, `$namba-land`, `$namba-release`, `$namba-plan`, `$namba-plan-review`, `$namba-harness`, `$namba-fix`, `$namba-review-resolve`, `$namba-plan-pm-review`, `$namba-plan-eng-review`, `$namba-plan-design-review`, `$namba-project`, and `$namba-sync` when the user invokes one command directly.\n"+
		"%s"+
		"- Keep the Namba report frame concise and high-signal. The response should feel like an engineering field report, not a rigid template dump.\n"+
		"- Keep `.namba/codex/validate-output-contract.py` as the fallback validator for this contract unless Namba explicitly adopts a documented upstream hook surface.\n"+
		"- Codex 0.131 boundary: Codex may display approval mode, permissions, service tier, and effective workspace roots; repo `.codex/config.toml` stores deterministic hook, agent, and MCP defaults only; `.codex/hooks.json` is the interactive guardrail; `.namba/hooks.toml` remains the `namba run` evidence and validation policy. Git helper commands may ignore repo hooks, so do not rely on Git hooks for validation.\n"+
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
		"Use this skill whenever the user mentions `namba`, `namba help`, `namba project`, `namba regen`, `namba update`, `namba codex access`, `namba plan`, `namba harness`, `namba fix`, `namba run`, `namba queue`, `namba sync`, `namba pr`, `namba land`, `namba release`, `$namba-help`, `$namba-coach`, `$namba-create`, `$namba-queue`, `$namba-plan-review`, `$namba-review-resolve`, `$namba-release`, or asks to use the Namba workflow.",
		"",
	}
	lines = append(lines, generatedInstructionContractSection("repo skill")...)
	lines = append(lines, "")
	lines = append(lines, renderNambaSkillCommandMappingSection()...)
	lines = append(lines, renderNambaSkillExecutionRulesSection(profile)...)
	return strings.Join(lines, "\n") + "\n"
}

func renderNambaSkillCommandMappingSection() []string {
	return []string{
		"Command mapping:",
		"- `$namba-help`: explain how to use NambaAI in this repository, which command or skill to use next, and where the authoritative docs live, without mutating repository state.",
		"- `$namba-coach`: restate the user's current goal, ask only essential routing questions when needed, correct clearly wrong command choices, and hand off to exactly one primary Namba workflow invocation without mutating repository state.",
		"- `$namba-create`: run the skill-first creation workflow for repo-local skills, project-scoped custom agents, or both. Keep the user-facing surface inside Codex and do not add a public `namba create` CLI command in this slice.",
		"- `namba project`: refresh repository docs and codemaps.",
		"- `namba codex access`: inspect or update Namba runner `codex exec` access defaults with explicit `approval_policy` / `sandbox_mode` flags; repo `.codex/config.toml` does not persist interactive approval or sandbox choices.",
		"- `namba regen`: regenerate AGENTS, repo-local skills, command-entry skills, Codex custom agents, readable role cards, and repo-local Codex config from `.namba/config/sections/*.yaml`.",
		"- `namba update [--version vX.Y.Z]`: self-update the installed `namba` binary from GitHub Release assets only after explicit user approval.",
		"- `namba plan \"<description>\"`: create the next feature SPEC package under `.namba/specs/`, and when branch-per-work is enabled create or switch to the dedicated `spec/...` branch in the current workspace; unless `--no-review` is present, continue immediately with `$namba-plan-review SPEC-XXX` after creation.",
		"- `$namba-plan-review`: create or resolve a SPEC, run the three plan-review tracks in parallel when possible, and drive the advisory readiness loop before implementation starts.",
		"- `namba harness \"<description>\"`: create the next harness-oriented SPEC package under `.namba/specs/` through the same dedicated-branch planning contract for reusable agent, skill, workflow, or orchestration work.",
		"- `$namba-plan-pm-review` / `$namba-plan-eng-review` / `$namba-plan-design-review`: update product, engineering, or design review artifacts under `.namba/specs/<SPEC>/reviews/` and refresh advisory readiness.",
		"- `namba fix --command plan \"<issue description>\"`: create the next bugfix SPEC package under `.namba/specs/` through the same dedicated-branch planning contract.",
		"- `namba fix \"<issue description>\"` or `namba fix --command run \"<issue description>\"`: perform direct repair in the current workspace without creating a SPEC package.",
		"- `$namba-review-resolve`: resolve the target PR from the current branch when possible, inspect unresolved review threads with thread-aware GitHub state, record thread identity and outcome, validate before replying or resolving, include CI/check evidence when relevant, and avoid duplicating the configured review marker.",
		"- `$namba-release`: draft release notes from commits since the previous semver tag, write the notes to a durable per-version artifact, then hand off to the guarded `namba release --version <version> --push` path with a GitHub Release body that uses the generated notes.",
		"- `namba run SPEC-XXX`: execute the SPEC in the current Codex session. Read `spec.md`, `plan.md`, and `acceptance.md`, implement directly, validate, and sync artifacts.",
		"- `namba run SPEC-XXX --solo|--team|--parallel`: use the standalone CLI runner when you need explicit single-subagent, multi-subagent, or worktree-parallel execution semantics.",
		"- `namba queue start <SPEC-RANGE|SPEC-LIST>`: process already-existing SPEC packages one at a time through review, run, PR, checks, optional land, and local main refresh. Use `status`, `resume`, `pause`, and `stop` to operate the durable queue.",
		"- `namba sync`: refresh change summary, PR checklist, codemaps, advisory review readiness, and PR-ready docs after implementation.",
		"- `namba pr \"<title>\"`: run sync plus validation by default, inspect PR checks, summarize bounded GitHub Actions failure snippets when checks fail, commit and push the current branch, and create or reuse a PR without requesting Codex review.",
		"- `namba pr --review \"<title>\"`: perform the same PR handoff and explicitly request Codex review with the configured marker exactly once.",
		"- `namba land`: resolve the current branch PR, optionally wait for checks, merge when the PR is clean, and update local `main` safely.",
		"- `namba doctor`: verify that AGENTS, repo skills, `.namba` config, Codex CLI, and the global `namba` command are available.",
		"",
	}
}

func renderNambaSkillExecutionRulesSection(profile initProfile) []string {
	return []string{
		"Execution rules:",
		"1. Treat `.namba/` as the source of truth.",
		"2. Prefer repo-local skills in `.agents/skills/`.",
		"3. Prefer `$namba-coach` for read-only current-goal command coaching, prefer `$namba-help` for read-only usage guidance, and prefer command-entry skills such as `$namba-create`, `$namba-run`, `$namba-queue`, `$namba-pr`, `$namba-land`, `$namba-release`, `$namba-plan`, `$namba-plan-review`, `$namba-harness`, `$namba-fix`, `$namba-review-resolve`, `$namba-plan-pm-review`, `$namba-plan-eng-review`, `$namba-plan-design-review`, `$namba-project`, and `$namba-sync` when the user is invoking one Namba command directly.",
		"4. Use the installed `namba` CLI for `project`, `regen`, `update`, `codex access`, `plan`, `harness`, `fix`, `queue`, `pr`, `land`, `release`, and `sync` when it will update repo state more reliably or self-update the installed CLI directly.",
		"5. Keep `.namba/specs/<SPEC>/reviews/*.md` and `readiness.md` current when you use the plan-review workflow; the readiness summary is advisory unless the user explicitly asks for a gate.",
		"6. For `namba run` in an interactive Codex session, prefer Codex-native in-session execution over recursively calling `namba run`, unless the user explicitly asks for standalone `--solo`, `--team`, `--parallel`, or `--dry-run` behavior.",
		"7. Run validation commands from `.namba/config/sections/quality.yaml` before finishing.",
		"8. Start each new SPEC or task on a dedicated work branch when `.namba/config/sections/git-strategy.yaml` enables branch-per-work collaboration.",
		fmt.Sprintf("9. Prepare PRs against `%s`, write the title/body in %s, and request Codex review only when `namba pr --review` or queue `--review` is explicit; use `%s` as the request command.", prBaseBranch(profile), humanLanguageName(profile.PRLanguage), codexReviewComment(profile)),
	}
}

func renderCommandSkill(name, description string, body []string) string {
	lines := []string{
		"---",
		fmt.Sprintf("name: %s", name),
		fmt.Sprintf("description: %s", description),
		"---",
		"",
		stateEffectForCommandSkill(name),
		"",
	}
	lines = append(lines, generatedInstructionContractSection("command skill")...)
	lines = append(lines, "")
	lines = append(lines, body...)
	return strings.Join(lines, "\n") + "\n"
}

func stateEffectForCommandSkill(name string) string {
	switch name {
	case "namba-help", "namba-coach":
		return "State effect: read-only guidance. Do not mutate repository state."
	case "namba-plan-pm-review", "namba-plan-eng-review", "namba-plan-design-review":
		return "State effect: mutating review update. Writes SPEC review artifacts and readiness summaries."
	case "namba-create":
		return "State effect: preview-first mutation. Do not write repo-local skills or agents until the preview is confirmed."
	case "namba-fix":
		return "State effect: mixed. Help/probe paths are read-only; direct repair mutates the current workspace; `--command plan` creates a bugfix SPEC."
	case "namba-codex-access":
		return "State effect: mixed. Inspect mode is read-only; explicit access flags update Namba runner defaults."
	default:
		return "State effect: mutating workflow entry point. Use help/probe paths read-only, and otherwise expect repository state or GitHub state to change."
	}
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
			"- Treat the init wizard as a repository-state-first flow: existing code should use detected language/framework defaults, while empty repositories should not ask for a starter app stack and should leave stack choice to the first clarified planning request. Use approachable emoji cues for step categories, echo each answer with a clear success marker before advancing, support `b`/`back` to revise the previous step, and do not ask the user for a GitHub username during onboarding.",
			"- After init, direct the user to open interactive Codex, run `/hooks` when Codex reports `6 hooks need review`, inspect the generated `.codex/hooks/namba_codex_guard.py` commands, and approve them before expecting prompt-refinement hooks to run.",
			"- Explain that repo skills live under `.agents/skills/` and Codex subagents live under `.codex/agents/*.toml`.",
			"- Keep the selected human language aligned across Codex conversation, docs, PR content, and code comments.",
		},
	)
}

func renderHelpCommandSkill() string {
	return renderCommandSkill(
		"namba-help",
		"Read-only entry point for explaining how to use NambaAI in the current repository.",
		[]string{
			"Use this skill when the user explicitly says `$namba-help`, asks how to use NambaAI, wants to know which command or skill to use next, or needs a read-only walkthrough of the current repository workflow.",
			"",
			"Behavior:",
			"- Stay read-only. Do not mutate repository state, create a SPEC, run validators, or invoke workflow commands just to answer a usage question.",
			"- Prefer `.namba/` and generated repository docs as the primary source of truth: `README*.md`, `docs/getting-started*.md`, `docs/workflow-guide*.md`, `.namba/codex/README.md`, and relevant repo skills under `.agents/skills/`.",
			"- Explain the practical differences between `$namba-create`, `namba project`, `namba codex access`, `namba plan`, `namba harness`, `namba fix`, `namba run`, `namba queue`, `namba sync`, `namba pr`, `namba land`, `namba release`, `namba regen`, `namba update`, `$namba-review-resolve`, `$namba-release`, and `namba doctor` when those distinctions matter to the user's goal.",
			"- When the user describes an outcome, recommend the next Namba command or skill concretely instead of giving a vague taxonomy.",
			"- If the user asks about pre-implementation review flow, explain when to use `$namba-plan-review` versus the individual review skills.",
			"- Distinguish `$namba-create` from `namba plan` and `namba harness`: use create when the user wants repo-local skills or custom agents directly, and use plan or harness when the user needs a SPEC package first.",
			"- If the user asks about a specific command, include concrete invocation examples and mention whether the path is read-only guidance, planning, execution, or GitHub handoff.",
			"- If repository docs and executable/config evidence disagree, call out the mismatch instead of smoothing it over.",
		},
	)
}

func renderCoachCommandSkill() string {
	return renderCommandSkill(
		"namba-coach",
		"Read-only advisory entry point for choosing the right Namba workflow handoff.",
		[]string{
			"Use this skill when the user explicitly says `$namba-coach`, arrives with a vague current goal, asks what to do next, or appears to have selected the wrong Namba command.",
			"",
			"Behavior:",
			"- Stay read-only. Do not create SPEC packages, edit repository files, generate skill, agent, source, or review artifacts, run implementation, update `.namba/specs/<SPEC>/reviews/readiness.md`, or invoke a public `namba coach` CLI command.",
			"- Restate the user's current goal briefly before choosing a path.",
			"- Follow this response order: brief restatement, up to three essential clarification questions when required, one primary executable handoff, optional single alternative when there is a real tradeoff, and a one- or two-sentence reason.",
			"- Treat essential clarification as information needed to choose the correct Namba workflow or make the handoff command usable, not information that would fully specify implementation.",
			"- Ask only 1-3 essential clarification questions when the request is underspecified.",
			"- Once the request is concrete enough, recommend exactly one primary executable invocation and at most one alternative.",
			"- Correct a clearly wrong command choice first instead of running it as-is.",
			"",
			"Boundary with `$namba-help`:",
			"- `$namba-help` explains how NambaAI works, what commands mean, and where docs live.",
			"- `$namba-coach` uses the user's current idea or question to choose the next workflow handoff.",
			"",
			"Routing rules:",
			"- New feature or product change: `namba plan \"<description>\"`",
			"- Reusable skill, agent, workflow, or orchestration SPEC: `namba harness \"<description>\"`",
			"- Direct repo-local skill or custom-agent creation: `$namba-create`",
			"- Bug repair: `namba fix \"<issue>\"`",
			"- Reviewable bugfix SPEC: `namba fix --command plan \"<issue>\"`",
			"- Review-thread cleanup and reply/resolve loop: `$namba-review-resolve`",
			"- NambaAI release orchestration and release-note handoff: `$namba-release`",
			"- Existing SPEC execution: `namba run SPEC-XXX`",
			"- Sequential execution of multiple existing SPEC packages: `$namba-queue` or `namba queue start SPEC-001..SPEC-003`",
			"- Usage or onboarding explanation: `$namba-help`",
			"- Implementation finished and artifacts need refresh: `namba sync`",
			"- Review handoff is ready: `namba pr \"<Korean title>\"`",
			"- Approved PR is ready to merge: `namba land`",
			"",
			"Wrong-command correction:",
			"- If the user asks for `$namba-plan` but the intent is reusable skill, agent, workflow, or orchestration planning, recommend `namba harness \"<description>\"` unless the request touches Namba core managed surfaces, where `namba plan \"<description>\"` remains appropriate.",
			"- If the user asks to directly create a repo-local skill or custom agent artifact, recommend `$namba-create` instead of `namba harness`.",
			"- If the user asks how Namba commands work or where docs live, recommend `$namba-help` instead of turning the answer into a planning handoff.",
			"",
			"Example:",
			"- For `todo 리스트를 만들고 싶은데 뭘 해야돼?`, ask only essential questions first: target environment, UI surface, and whether tasks should be local-only or persisted somewhere.",
			"- After those answers are concrete, hand off with `namba plan \"Build a todo list feature for <environment> with <UI surface> and <storage approach>.\"`",
		},
	)
}

func renderCreateCommandSkill() string {
	return renderCommandSkill(
		"namba-create",
		"Skill-first entry point for creating a repo-local skill, a project-scoped custom agent, or both.",
		[]string{
			"Use this skill when the user explicitly says `$namba-create` or asks to create a repo-local skill, a project-scoped custom agent, or both through Namba.",
			"",
			"Behavior:",
			"- Keep this flow skill-first. Do not introduce a documented public `namba create` Go CLI command as part of this slice.",
			"- Use progressive disclosure for generated guidance: keep `SKILL.md` bodies lean and place long procedures in references, assets, or deterministic helper candidates. Do not add `$CODEX_HOME/skills`, third-party install flows, or app-automation dependencies.",
			"- When the installed `namba` CLI is available, use the internal adapter `namba __create preview` and `namba __create apply` with JSON stdin/stdout as the durable generation path. Treat `__create` as wrapper-only plumbing, not a public command.",
			"- Run the interaction as a staged generator: `unresolved` -> `narrowed` -> `confirmed`.",
			"- Keep each turn stateful and visible: summarize the current candidate target and the remaining unresolved items before asking the next clarifying question.",
			"- If the user explicitly says `skill`, `agent`, or `both`, treat that directive as authoritative over any heuristic classification.",
			"- Use `sequential-thinking` when decomposition or clarification planning is non-trivial, use `context7` only when targeted external library or framework guidance materially helps the generated instructions, and use `playwright` only when browser verification is actually relevant.",
			"- Before any write, present a non-mutating preview that includes the chosen output type, slug or name, intended file paths, validation plan, whether a fresh Codex session will likely be required, and the planned five-role analysis or verification record.",
			"- Do not write files until the target, slug, paths, and overwrite decisions are explicit and the user has confirmed the preview.",
			"- Normalize names into a safe slug, constrain writes to `.agents/skills/<slug>/SKILL.md`, `.codex/agents/<slug>.toml`, and `.codex/agents/<slug>.md`, and reject path traversal, invalid slugs, silent overwrites, or incomplete agent mirror pairs.",
			"- Reject durable instructions that preserve raw unnormalized user prose, stale Claude-only primitives, or repository-policy violations.",
			"- Record at least five independent role outputs across planning or verification when the flow advances to generation, while degrading safely if the effective same-workspace thread limit is lower than the repo default.",
			"- Keep user-authored outputs distinct from Namba-managed built-ins so `namba regen` preserves them, and surface fresh-session guidance clearly when instruction surfaces change.",
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
			"- Treat `product.md` as the landing document, `tech.md` as the technical hub, and `structure.md` as appendix material.",
			"- Surface system boundaries, evidence/confidence, mismatch reporting, and quality warnings instead of flattening the repository into a shallow tree dump.",
			"- Summarize entry points, per-system artifacts, and any drift or thin-output warnings after the refresh.",
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
			"- Ask the user before running `namba update`; version advisory messages are not consent to update.",
			"- Use `namba update --version vX.Y.Z` when the user requests a specific version.",
			"- Do not confuse this command with scaffold regeneration; that belongs to `namba regen`.",
		},
	)
}

func renderReviewResolveCommandSkill() string {
	return renderCommandSkill(
		"namba-review-resolve",
		"Command-style entry point for resolving GitHub PR review feedback with thread-aware state.",
		[]string{
			"Use this skill when the user explicitly says `$namba-review-resolve`, asks to resolve review comments, or uses Korean wording equivalent to `리뷰 확인하고 의미있는 리뷰면 수정 후에 해당 리뷰에 답변을 달고, resolve한 다음 다시 리뷰 요청해`.",
			"",
			"Behavior:",
			"- Resolve the target PR from the current branch when the user does not provide a PR number.",
			"- Discover unresolved review threads with thread identity, path, comments, authors, resolved state, and outdated state, not just flat PR comments. Prefer a thread-aware GitHub path such as `gh api graphql` when the connector surface cannot expose review-thread state directly.",
			"- Classify each thread as meaningful/actionable or non-actionable. Meaningful threads cover correctness, tests, security, UX, docs, release, regression, and maintainability concerns that require a concrete change or precise answer.",
			"- Assign an explicit outcome to every reviewed thread: `fixed-and-resolved`, `answered-open`, or `skipped-with-rationale`.",
			"- Implement meaningful fixes in the smallest coherent patch, then run the configured validation commands before replying or resolving.",
			"- Record changed paths, commit or diff summary when available, validation evidence, and CI/check evidence when the review feedback or PR health depends on failing checks.",
			"- Inspect PR check status before re-requesting review; for failing GitHub Actions checks, include run URLs and bounded failure snippets, and for external checks report status plus details URL only.",
			"- Reply on the original thread with the concrete change made plus validation and relevant CI/check evidence, then resolve only the threads that were fixed or conclusively answered.",
			"- Leave non-actionable threads open unless they were genuinely answered. Never silently resolve a thread that was not addressed.",
			"- Re-request review only after all meaningful threads are addressed, validation passes, PR check state is understood, and the configured `@codex review` marker is present exactly once.",
			"- Treat flat PR comments as PR-level context only; use thread-aware review data for the actual resolution loop.",
		},
	)
}

func renderReleaseCommandSkill() string {
	return renderCommandSkill(
		"namba-release",
		"Command-style entry point for NambaAI release orchestration and release-note handoff.",
		[]string{
			"Use this skill when the user explicitly says `$namba-release`, `namba release` in a Codex workflow context, or Korean wording such as `릴리즈 진행해`.",
			"",
			"Behavior:",
			"- Treat this as NambaAI-specific release orchestration, not a generic release helper.",
			"- Start from `main` and require a clean working tree before the final tagging step.",
			"- If generated templates or docs changed during release prep, run `namba regen` and/or `namba sync`, validate, and commit those changes before tagging.",
			"- Determine the target version from explicit input or the next semver bump.",
			"- Collect commits since the previous semver tag, ignoring merge noise and excluding any release-note prep commit when the notes artifact is committed separately.",
			"- Draft release notes from that commit range and group them into user-visible changes, fixes, docs/workflow, and internal maintenance while preserving SPEC IDs, PR numbers, and short commit hashes when useful.",
			"- Release notes must describe actual changes and include evidence sources such as SPEC IDs, PR numbers, short commit hashes, source artifacts, and validation results when available.",
			"- Write the notes to a durable per-version artifact such as `.namba/releases/<version>.md`, then use that file as the handoff for the guarded `namba release --version <version> --push` path.",
			"- Write release-facing prose in the explicit user-requested language when one is given, such as `--language en`; otherwise use the init or project-configured language.",
			"- Do not tag until the notes exist and validation has passed.",
			"- Make sure the GitHub Release body uses the generated notes rather than an empty or generic body.",
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
			"- Before reading project docs, checking Git state, or running the CLI, run the Namba clarification gate on the user's raw request.",
			"- If the request is short, broad, or missing target surface, user flow, scope boundaries, constraints, acceptance criteria, or validation, do not run `namba plan` yet. Ask 1-3 concise questions first; `게시판 만들어줘` is the canonical example that must ask questions before planning.",
			"- Treat Codex Plan mode as the preferred clarification surface and `namba plan` as the executor: when native Plan mode choice UI is available, use it to ask the clarification questions before any CLI command.",
			"- The Plan mode output must be converted into a refined description shaped as Goal, Scope, Constraints, and Acceptance, then passed to `namba plan \"<refined description>\"`.",
			"- Repo hooks and skills cannot force Codex to switch modes by themselves; if native Plan mode UI is unavailable in the current session, fall back to concise text questions in chat instead of inventing a CLI wizard.",
			"- Continue the clarification loop across turns until you can restate the request as Goal, Scope, Constraints, and Acceptance. Only then invoke the CLI with the refined description.",
			"- Prefer the installed `namba plan` CLI when available.",
			"- Keep `namba plan` for feature-oriented SPEC work; use `namba harness` when the request is about reusable agent, skill, workflow, or orchestration scaffolding; use `$namba-create` when the user wants the repo-local skill or custom-agent artifact itself instead of another SPEC.",
			"- When repo-managed MCP presets are configured, prefer them for planning context before broader web search; for example, use `context7` for library and framework docs, `sequential-thinking` for deeper decomposition, and `playwright` for browser-verified flows.",
			"- Read `.namba/project/product.md`, `.namba/project/tech.md`, `.namba/project/mismatch-report.md`, `.namba/project/quality-report.md`, and any relevant `.namba/project/systems/*.md` artifacts before drafting the SPEC.",
			"- Treat executable code and authoritative config as stronger planning evidence than docs, and preserve code-vs-doc conflicts instead of smoothing them out.",
			"- Keep planning in the current workspace. When branch-per-work is enabled, create or switch to the dedicated `spec/...` branch before writing `.namba/specs/<SPEC>/`.",
			"- Use `--current-workspace` only when the user intentionally wants to scaffold on the current branch without creating a dedicated SPEC branch.",
			"- Do not create a planning worktree. Reserve temporary worktrees for overlapping `namba run SPEC-XXX --parallel` execution, and expect them to disappear after the run finishes cleanly.",
			"- Create the next sequential `SPEC-XXX` package under `.namba/specs/` after that branch decision is explicit.",
			"- Seed `.namba/specs/<SPEC>/reviews/` with product, engineering, design, and aggregate readiness artifacts.",
			"- Unless the invocation includes `--no-review`, treat the created SPEC as an automatic handoff and immediately continue with `$namba-plan-review SPEC-XXX` after the CLI prints the new SPEC ID.",
			"- After the automatic handoff, read `.namba/specs/<SPEC>/reviews/readiness.md`; if it does not say `Cleared reviews: 3/3`, do not treat planning as complete. Continue with `$namba-plan-review SPEC-XXX` again, or run the missing `$namba-plan-pm-review`, `$namba-plan-eng-review`, and `$namba-plan-design-review` tracks before implementation.",
			"- When `--no-review` is present, stop after scaffold creation and point any later review work to `$namba-plan-pm-review`, `$namba-plan-eng-review`, and `$namba-plan-design-review`, or `$namba-plan-review SPEC-XXX`.",
			"- Keep the scope concrete and implementation-ready.",
		},
	)
}

func renderPlanReviewLoopCommandSkill() string {
	return renderCommandSkill(
		"namba-plan-review",
		"Command-style entry point for bundling SPEC creation, parallel plan reviews, and advisory readiness validation.",
		[]string{
			"Use this skill when the user explicitly says `$namba-plan-review`, asks to create a SPEC and run the full pre-implementation review loop, or wants `namba plan` plus the review flow bundled into one skill.",
			"",
			"Behavior:",
			"- Resolve the target SPEC from an explicit `SPEC-XXX`; otherwise create the next SPEC with `namba plan` for feature work, `namba harness` for reusable agent/skill/workflow/orchestration work, or `namba fix --command plan` for bugfix planning.",
			"- When creating a new SPEC from a user request, inherit the same clarification gate as `$namba-plan`: ask first and do not run the CLI while the raw request is still vague, short, or missing Goal/Scope/Constraints/Acceptance.",
			"- Prefer the installed `namba` CLI for the SPEC-creation step when it is available; keep `.namba/` as the source of truth if you need to do the setup manually.",
			"- Inherit the same safe-by-default planning branch contract as `namba plan`: create or switch to the dedicated `spec/...` branch in the current workspace by default, treat `--current-workspace` as the explicit current-branch escape hatch, and do not create planning worktrees.",
			"- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before launching reviews or revising the planning artifacts.",
			"- Launch product, engineering, and design review passes in parallel when subagent routing is available, using `$namba-plan-pm-review`, `$namba-plan-eng-review`, and `$namba-plan-design-review` as the three authoritative review tracks.",
			"- Prefer `namba-product-manager`, `namba-planner`, and `namba-designer` for the three review tracks, and use `namba-plan-reviewer` as the aggregate validator when custom-agent routing is available.",
			"- After the three review tracks finish, run an aggregate validation pass over `spec.md`, `plan.md`, `acceptance.md`, and `.namba/specs/<SPEC>/reviews/*.md` to check coverage gaps, contradictions, and whether the advisory readiness state is credible.",
			"- If the aggregate validator finds issues, revise the SPEC or review artifacts directly, rerun only the affected review tracks, and repeat the validation loop instead of restarting every pass blindly.",
			"- Refresh `.namba/specs/<SPEC>/reviews/readiness.md` after each review and validation cycle so the advisory summary stays current.",
			"- Before ending, inspect `.namba/specs/<SPEC>/reviews/readiness.md`; if it does not say `Cleared reviews: 3/3`, do not report the review loop as complete. Continue the review loop when useful, or make the final next-work item explicitly name `$namba-plan-review SPEC-XXX` or the missing individual review skills to run next.",
			"- Keep the loop bounded and explicit: stop when the readiness state is clear enough to proceed or when the remaining blockers are concrete enough that another loop would be redundant.",
			"- Keep the whole flow advisory by default; missing depth or blockers should be visible, not silently converted into a hard gate.",
		},
	)
}

func renderHarnessCommandSkill() string {
	return renderCommandSkill(
		"namba-harness",
		"Command-style entry point for creating the next harness-oriented SPEC package.",
		[]string{
			"Use this skill when the user explicitly says `$namba-harness`, `namba harness`, or asks to create a harness-oriented SPEC package.",
			"",
			"Behavior:",
			"- Before reading project docs, checking Git state, or running the CLI, run the same Namba clarification gate used by `namba plan` on the user's raw request.",
			"- If the request is short, broad, or missing target surface, user flow, scope boundaries, constraints, acceptance criteria, or validation, do not run `namba harness` yet. Ask 1-3 concise questions first.",
			"- Treat Codex Plan mode as the preferred clarification surface, then pass only the refined Goal/Scope/Constraints/Acceptance description to `namba harness \"<refined description>\"`.",
			"- Prefer the installed `namba harness` CLI when available.",
			"- Use this path for reusable agent, skill, workflow, orchestration, or evaluation scaffolding when the user wants a reviewable SPEC first instead of generating the repo-local skill or agent artifact directly through `$namba-create`.",
			"- Start with the same dedicated-branch planning contract as `namba plan`, and use `--current-workspace` only when the user intentionally wants to scaffold on the current branch without creating a dedicated SPEC branch.",
			"- Do not create planning worktrees here either; temporary worktrees belong to overlapping `namba run SPEC-XXX --parallel` execution only.",
			"- Create the next sequential `SPEC-XXX` package under `.namba/specs/` without inventing a second artifact model.",
			"- Seed `.namba/specs/<SPEC>/reviews/` with product, engineering, design, and aggregate readiness artifacts so the review flow stays aligned with `namba plan`.",
			"- Keep command-entry skill guidance lean; move long PR-thread, CI-log, frontend, or MCP recipes into existing references or deterministic helper-script candidates instead of creating new standalone skills.",
			"- Evaluate deterministic helper-script candidates before implementation: they need `--help`, read-only defaults, bounded output, explicit network/auth assumptions, fixture or local-server tests, and no destructive or third-party app coupling.",
			"- For large managed-skill changes, inventory affected source and generated surfaces, classify mechanical versus behavioral edits, update templates first, regenerate, review generated diffs, and validate.",
			"- For harness/MCP quality, prefer workflow-first designs over raw endpoint wrappers, require context-budgeted outputs with pagination or truncation expectations, produce actionable errors, and define evaluation scenarios that are independent, read-only, realistic, verifiable, and stable.",
			"- Keep the output Codex-native and avoid Claude-only runtime primitives in the planned contract.",
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
	return renderCommandSkill(
		"namba-plan-design-review",
		"Command-style entry point for design review of a SPEC before implementation starts.",
		[]string{
			"Use this skill when the user explicitly says `$namba-plan-design-review` or asks for a design review on a SPEC package.",
			"",
			"Behavior:",
			"- Resolve the target SPEC from an explicit `SPEC-XXX`; otherwise use the latest SPEC under `.namba/specs/`.",
			"- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before writing review notes.",
			"- Update `.namba/specs/<SPEC>/reviews/design.md` with status, reviewer, findings, decisions, follow-ups, and recommendation.",
			"- Prefer `namba-designer` as the review role when subagent routing is appropriate.",
			"- Check art-direction clarity, palette temperature and undertone logic, restrained saturation, semantic component choice, anti-generic composition, meaningful motion, redesign of the most generic section when applicable, and anti-overcorrection safeguards.",
			"- Treat card, border, and grid usage as valid when they are justified; the regression target is default reliance or meaningless fallback use, not a categorical ban.",
			"- Reject both washed-out gray minimalism and novelty-for-novelty. Preserve accessibility, design-system fit, and implementation realism.",
			"- Refresh `.namba/specs/<SPEC>/reviews/readiness.md` so the advisory summary reflects the latest review state.",
			"- Keep the review advisory by default; surface missing depth or blockers clearly without silently turning the workflow into a hard gate.",
		},
	)
}

func renderFixCommandSkill() string {
	return renderCommandSkill(
		"namba-fix",
		"Command-style entry point for direct bug repair or bugfix SPEC planning.",
		[]string{
			"Use this skill when the user explicitly says `$namba-fix`, `namba fix`, or asks to repair a bug through Namba.",
			"",
			"Behavior:",
			"- Prefer the installed `namba fix` CLI when available.",
			"- Treat `namba fix \"<issue description>\"` as the default direct-repair path in the current workspace.",
			"- Use `namba fix --command run \"<issue description>\"` when the user wants the explicit direct-repair form.",
			"- Use `namba fix --command plan \"<issue description>\"` when the user wants a reviewable bugfix SPEC package under `.namba/specs/` via the same dedicated-branch planning contract.",
			"- For `namba fix --command plan`, run the same clarification gate as `namba plan` and `namba harness` before reading project docs, checking Git state, or running the CLI. Ask 1-3 concise questions first when the issue, target surface, scope, constraints, acceptance criteria, or validation are unclear.",
			"- When clarification is needed for the planning path, prefer Codex Plan mode and pass only the refined Goal/Scope/Constraints/Acceptance description to `namba fix --command plan \"<refined issue description>\"`.",
			"- Use `--current-workspace` only with `namba fix --command plan` when the user intentionally wants to scaffold on the current branch without creating a dedicated SPEC branch.",
			"- Do not create planning worktrees for this path; worktrees are reserved for temporary overlapping `namba run SPEC-XXX --parallel` execution.",
			"- Keep CLI help and flag probing read-only; `namba <command> --help`, `namba <command> -h`, and `namba help <command>` must not mutate repository state.",
			"- Keep direct repairs small, add targeted regression coverage, run validation, and finish with `namba sync`.",
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
			"- Read `.namba/specs/<SPEC>/frontend-brief.md` when it exists; it is the canonical source for frontend task classification and gate state.",
			"- In an interactive Codex session, prefer Codex-native in-session execution over recursively calling `namba run`.",
			"- Only use the standalone CLI runner for `--solo`, `--team`, `--parallel`, `--dry-run`, or when the user explicitly wants the non-interactive runner path.",
			"- For `--solo`, stay inside one runner unless one domain clearly dominates and a single specialist would materially reduce risk.",
			"- For `--team`, prefer one specialist when one domain dominates, expand to two or three only when acceptance spans multiple domains, and keep one integrator plus final validation owner in the workspace.",
			"- For `--team`, honor each selected role's `model` and `model_reasoning_effort` metadata from `.codex/agents/*.toml` so planner/reviewer/security roles can think harder without making every delivery role heavy.",
			"- Route art direction, palette/tone logic, composition, motion intent, Figma critique, and generic-section redesign work to `namba-designer`; route component boundaries, state ownership, and UI delivery planning to `namba-frontend-architect`; route approved UI implementation to `namba-frontend-implementer`; route mobile-specific UI delivery to `namba-mobile-engineer`; route API, schema, and pipeline work to backend/data; route auth, secrets, and compliance work to security; route deployment and runtime work to devops.",
			"- `frontend-major` work must not move into architecture or implementation until `frontend-brief.md` shows coherent problem, reference, critique, decision, prototype evidence, a complete Do-Not Design Contract with reference-driven asset manifest, generated-image decision fields, generation plan, and aligned design clearance; `frontend-minor` keeps the lightweight advisory path.",
			"- Treat review readiness as advisory by default for non-frontend and `frontend-minor` work, but block explicit `frontend-major` execution when the frontend brief is missing required evidence, internally contradictory, mismatched with design-review summaries, or missing/insufficient negative-first contract evidence.",
			"- For `frontend-major`, first frontend and first major screen phases default to `Asset mode: generated-images` and `Imagegen requirement: required`; accept `existing-assets` or `not-applicable` only with validator-readable Asset decision proof.",
			"- For `frontend-major` implementation results, require a `Do-Not Design Violation Check` that cites changed files, names any banned pattern found, and cites the exception path when one is used; when `Imagegen requirement: required` is present, require `Generated Asset Evidence` with manifest path and repeatable Asset ID blocks containing generated file, saved asset path, prompt summary, intended UI usage, and rendered usage evidence.",
			"- For browser-rendered frontend work, use managed server lifecycle, wait for rendered DOM state, capture screenshots, inspect console errors, and prefer Playwright checks when the surface runs in a browser.",
			"- Run validation commands from `.namba/config/sections/quality.yaml` and finish with `namba sync`. Use `namba pr` and `namba land` for the GitHub handoff and merge cycle instead of overloading `sync`.",
			fmt.Sprintf("- Collaboration defaults: branch from `%s`, open the PR into `%s`, write the PR in %s, and request Codex review only when `namba pr --review` or queue `--review` is explicit; use `%s` as the request command.", branchBase(profile), prBaseBranch(profile), humanLanguageName(profile.PRLanguage), codexReviewComment(profile)),
		},
	)
}

func renderQueueCommandSkill() string {
	return renderCommandSkill(
		"namba-queue",
		"Command-style entry point for operating the existing-SPEC queue conveyor.",
		[]string{
			"Use this skill when the user explicitly says `$namba-queue`, `namba queue`, or asks to process multiple existing SPEC packages in order.",
			"",
			"Behavior:",
			"- Prefer the installed `namba queue` CLI when available because queue state and Git/GitHub evidence are durable CLI-owned outputs.",
			"- Only consume already-existing SPEC packages. Do not create new SPEC packages from this command surface.",
			"- `namba queue start <SPEC-RANGE|SPEC-LIST>` accepts ranges such as `SPEC-001..SPEC-003` and explicit lists such as `SPEC-001 SPEC-004`, plus `--auto-land`, `--review`, deprecated `--skip-codex-review`, and `--remote origin`.",
			"- Use `namba queue status [--verbose]` to report active SPEC, durable state, blocker or wait reason, evidence path, PR link, and next safe command before deciding how to resume.",
			"- Use `namba queue resume` only after checking or resolving the current wait/blocker state; use `pause` and `stop` as cooperative controls that preserve branches, PRs, and evidence.",
			"- Treat `.namba/logs/queue/` as the durable queue state and report surface.",
			"- Continue one active SPEC at a time through review, implementation, validation, active-SPEC-aware sync/PR, checks, optional land, and local main refresh.",
			"- Block instead of skipping on failed validation, failed checks, non-mergeable PRs, dirty queue branches, GitHub auth failures, missing `gh`, diverged branches, ambiguous PR/check state, or unclear review readiness.",
			"- Without `--auto-land`, stop in `waiting_for_land` after green and mergeable PR evidence so the operator can land intentionally.",
			"- Queue PR handoff does not request Codex review unless `--review` is present; keep `--skip-codex-review` as deprecated compatibility syntax only, and never combine it with `--review`.",
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
			"- Use the configured PR base branch, PR language, and Codex review command from `.namba/config/sections/git-strategy.yaml`; let an explicit `--language <en|ko|ja|zh>` request override the configured PR language for that handoff.",
			"- Run `namba sync` and validation by default before creating review artifacts.",
			"- Build the PR body from concrete completed work, changed areas or files, evidence sources, validation results, and the latest SPEC review-readiness artifact when `.namba/specs/<SPEC>/reviews/readiness.md` exists.",
			"- PR bodies must avoid empty placeholder prose; cite source artifacts such as `.namba/project/change-summary.md`, `.namba/project/pr-checklist.md`, SPEC files, validation output, commits, or PR metadata when available.",
			"- Inspect current PR check status before review handoff; for failing GitHub Actions checks, capture run URLs and bounded GitHub Actions failure snippets, and report external checks by status and details URL only.",
			"- Commit and push the current work branch, create or reuse the GitHub PR, and do not request Codex review unless the user explicitly asked for `--review`.",
			"- When `--review` is present, ensure the configured Codex review marker exists exactly once.",
			fmt.Sprintf("- Collaboration defaults: PRs target `%s`, PR content is written in %s, and `%s` is the explicit review request command.", prBaseBranch(profile), humanLanguageName(profile.PRLanguage), codexReviewComment(profile)),
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
		"State effect: guidance plus execution support. Read-only when used for orientation; mutating only when the active Namba workflow is already executing.",
		"",
	}
	lines = append(lines, generatedInstructionContractSection("repo skill")...)
	lines = append(lines,
		"",
		"Key ideas:",
		"- SPEC-first execution",
		"- Codex-native implementation for `namba run` requests inside an interactive session",
		"- TDD for greenfield or sufficiently tested projects",
		"- DDD-style preserve/improve flow for brownfield projects with weak test coverage",
		"- TRUST gates after each execution phase",
		"- Worktree-based isolation for parallel work",
		"- Namba report frame for substantial user-facing responses: scoped definition, judgment, work completed, current issues, potential risks, next steps",
	)
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
		"State effect: mutating scaffold workflow when applied. Read-only when used only to explain migration or init behavior.",
		"",
	}
	lines = append(lines, generatedInstructionContractSection("repo skill")...)
	lines = append(lines,
		"",
		"Core mapping:",
		"- `CLAUDE.md` -> `AGENTS.md`",
		"- `.claude/skills/*` -> `.agents/skills/*`",
		"- `.claude/commands/*` -> command-entry repo skills such as `.agents/skills/namba-create/SKILL.md` and `.agents/skills/namba-run/SKILL.md`",
		"- `.claude/agents/*` -> `.codex/agents/*.toml` custom agents with `.md` role-card mirrors",
		"- `.claude/hooks/*` -> explicit validation pipeline and `namba` orchestration",
		"- Claude custom slash-command workflows -> built-in Codex slash commands plus repo skills such as `$namba-create`, `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, `$namba-sync`, and the `namba` CLI",
		"",
		"When implementing init changes:",
		"1. Keep `.namba/config/sections/*.yaml` as the durable source of truth.",
		"2. Never write tokens or secrets into generated config files.",
		"3. Prefer repo-local skills and `.toml` custom agents while keeping `.md` files as readable mirrors.",
		"4. Keep one selected human language aligned across Codex conversation, docs, PR content, and code comments unless the user explicitly overrides it.",
		"5. Keep generated assets readable so users can understand what `namba init .` changed.",
	)
	return strings.Join(lines, "\n") + "\n"
}

func renderProjectSkill() string {
	lines := []string{
		"---",
		"name: namba-workflow-project",
		"description: Project analysis, codemap refresh, and documentation generation for NambaAI.",
		"---",
		"",
		"State effect: mutating project documentation workflow. Read-only only when explaining current project-analysis behavior.",
		"",
	}
	lines = append(lines, generatedInstructionContractSection("repo skill")...)
	lines = append(lines,
		"",
		"Use this skill to:",
		"- refresh project docs",
		"- summarize structure and entry points",
		"- rebuild codemap artifacts under `.namba/project/codemaps`",
		"- explain how the repository is organized before planning or execution",
		"- implement the `namba project` command inside Codex when the CLI is not used directly",
	)
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
		"State effect: mutating SPEC execution workflow. Read-only only while inspecting the SPEC package before implementation.",
		"",
	}
	lines = append(lines, generatedInstructionContractSection("repo skill")...)
	lines = append(lines,
		"",
		"Execution pattern:",
		"1. Read `.namba/specs/<SPEC>/spec.md`",
		"2. Read `.namba/specs/<SPEC>/plan.md`",
		"3. Read `.namba/specs/<SPEC>/acceptance.md`",
		"4. Read `.namba/specs/<SPEC>/reviews/readiness.md` when present so advisory review status informs execution",
		"5. Read `.namba/specs/<SPEC>/frontend-brief.md` when present and treat it as the canonical frontend gate contract",
		"6. If the SPEC is explicitly `frontend-major`, stop and route back to research/synthesis when the frontend gate is incomplete, invalid, or contradicted by design-review summaries",
		"7. Implement the work directly in the current Codex session",
		"8. For browser-rendered frontend changes, use managed server lifecycle, inspect rendered DOM state, capture screenshots, record console errors, and prefer Playwright checks when practical",
		"9. Run configured validation commands",
		"10. Summarize results in `.namba/logs` and sync artifacts",
		"",
		fmt.Sprintf("Collaboration defaults: use a dedicated branch from `%s` for the SPEC, open the PR into `%s`, write the PR in %s, and request Codex review only when `namba pr --review` or queue `--review` is explicit; use `%s` as the request command.", branchBase(profile), prBaseBranch(profile), humanLanguageName(profile.PRLanguage), codexReviewComment(profile)),
		"",
		"Do not call `namba run` from inside Codex unless the user explicitly requests the non-interactive CLI runner.",
	)
	return strings.Join(lines, "\n") + "\n"
}

func renderCodexUsage(profile initProfile) string {
	lines := []string{
		"# Codex Integration",
		"",
		fmt.Sprintf("`%s` is configured for Codex-native Namba workflow.", profile.ProjectName),
		"",
		"NambaAI's differentiator is prompt refinement before execution: ambiguous ideas should become clearer goals, scope, constraints, and acceptance criteria before Codex starts coding.",
		"",
	}
	lines = append(lines, renderCodexUsageInitEnablesSection()...)
	lines = append(lines, renderCodexUsageHowCodexUsesNambaSection()...)
	lines = append(lines, renderCodexUsageWorkflowCommandSemanticsSection()...)
	lines = append(lines, renderCodexUsagePlatformReadinessSection()...)
	lines = append(lines, renderCodexUsageCompatibilityBoundarySection()...)
	lines = append(lines, renderCodexUsageAgentRosterSection()...)
	lines = append(lines, renderCodexUsageDelegationHeuristicsSection()...)
	lines = append(lines, renderCodexUsagePlanReviewReadinessSection()...)
	lines = append(lines, renderCodexUsageOutputContractSection(profile)...)
	lines = append(lines, renderCodexUsageGitCollaborationSection(profile)...)
	lines = append(lines, renderCodexUsageClaudeMappingSection()...)
	lines = append(lines, renderCodexUsageImportantDistinctionSection()...)
	return strings.Join(lines, "\n") + "\n"
}

func renderCodexUsagePlatformReadinessSection() []string {
	return []string{
		"## Optional Platform Readiness",
		"",
		"- Unified `@` mentions across files, directories, plugins, and skills are Codex platform search. Namba command routing still follows explicit `$namba-*` skills and `namba ...` command intent.",
		"- Plugin packaging, marketplace CLI commands, version-aware sharing, share checkout, shared-workspace plugin buckets, and default-enabled plugin hooks are readiness paths. They do not make plugin installation or marketplace publication required for local NambaAI use.",
		"- Remote-control and configured remote environments are optional diagnostics. When visible, they appear as additive `codex_diagnostics.remote_control` and `codex_diagnostics.remote_environments` evidence with normalized statuses such as `unavailable`, `disabled`, `enabled`, `configured`, or `local_fallback`.",
		"- Project diagnostics may run stable local Codex probes; run, queue, hook, and parallel evidence use non-blocking snapshots or neutral fallback status.",
		"- Multi-environment `apply_patch` selection remains a Codex platform capability; Namba's default editing path is the local workspace.",
		"- Codex Python SDK references should use the `openai-codex` distribution and `openai_codex` import package. NambaAI does not add a Python runtime dependency for documentation-only references.",
		"",
		"Field-level schema map:",
		"- `codex_diagnostics.remote_control.status`: owner `codexDiagnosticsEvidence`; allowed `unavailable`, `disabled`, `enabled`, `local_fallback`; producer `buildCodexDiagnosticsEvidence`; consumers project/run/queue/hook/parallel evidence and docs.",
		"- `codex_diagnostics.remote_control.source`: owner `codexRemoteControl`; allowed local source labels such as `stable_cli_status`, `stable_cli_help`, `explicit_config`, `environment_snapshot`, `codex_cli_missing`, and `non_blocking_snapshot`; producer diagnostics builder; consumers evidence readers and docs.",
		"- `codex_diagnostics.remote_environments.status`: owner `codexDiagnosticsEvidence`; allowed `unavailable` or `configured`; producer diagnostics builder from explicit options, environment snapshots, or `CODEX_HOME`; consumers project/run/queue/hook/parallel evidence and docs.",
		"- `codex_diagnostics.remote_environments.names`: owner `codexRemoteEnvironments`; allowed normalized configured environment labels; producer diagnostics builder; consumers project diagnostics and evidence readers.",
		"",
	}
}

func renderCodexUsageCompatibilityBoundarySection() []string {
	return []string{
		"",
		"## Codex 0.131 Boundary",
		"",
		"Read this boundary in order: Codex displays runtime state, Namba generates repo assets, Namba validates run evidence, and some runtime behavior remains outside Namba control.",
		"",
		"| Term | Boundary |",
		"| --- | --- |",
		"| `approval mode` | Codex may display the active session approval mode; repo `.codex/config.toml` does not set or persist it. |",
		"| `permissions` | Codex may display current permission state; Namba docs describe the state but do not own it. |",
		"| `effective workspace roots` | Codex may show the roots considered for the session; this is visibility, not a Namba promise that every root is writable. |",
		"| `scoped write roots` | Codex sandbox policy may narrow writable paths; Namba documents the distinction and keeps generated config out of that choice. |",
		"| `.codex/hooks.json` | Interactive Codex guardrails: context, prompt refinement, command denial, duplicate guard suppression, and generated-surface reminders. |",
		"| `.namba/hooks.toml` | `namba run` evidence and validation hooks; it is not used to dedupe interactive Codex hooks. |",
		"",
		"- Codex 0.131 also reports service tier and richer status-line state. Namba-generated docs can explain those fields, but repo config must not persist a user's service tier or approval/session choices.",
		"- Project, run, and queue evidence may include a `codex_diagnostics` payload with neutral statuses such as `detected`, `not_detected`, `unavailable`, `timed_out`, `advisory_mismatch`, and `redacted`.",
		"- Evidence payloads are covered by append-only schema contracts including `codex-diagnostics-evidence/v1`, `project-codex-diagnostics-evidence/v1`, `execution-evidence/v1`, `queue-runner-evidence/v1`, `namba-eval-scorecard/v1`, and `namba-report/v1`.",
		"- Inspect project diagnostics at `.namba/logs/project/codex-diagnostics-evidence.json`, run diagnostics at `.namba/logs/runs/<log-id>-evidence.json`, queue handoff diagnostics at `.namba/logs/runs/<spec-id-lower>-queue-evidence.json`, and doctor logs beside the artifact that triggered them.",
		"- Git helper commands may ignore configured repository hooks, so Namba cannot rely on Git hooks for safety or validation evidence.",
		"- On Windows, prefer WSL for Codex workspaces when possible; otherwise expect stricter deny-read behavior, scoped write roots, ineffective firewall policy handling, and PowerShell launcher constraints.",
	}
}

func renderCodexUsageWorkflowCommandSemanticsSection() []string {
	return []string{
		"## Workflow Command Semantics",
		"",
		"- `$namba-help` explains how to use NambaAI, which command to choose next, and where the authoritative docs live. It should stay read-only.",
		"- `$namba-coach` clarifies the user's current goal, corrects clearly wrong command choices, and hands off to exactly one primary Namba workflow invocation. It should stay read-only.",
		"- `$namba-create` is the preview-first creation path for repo-local skills and custom agents. Use it when the user wants `.agents/skills/*` or `.codex/agents/*` outputs directly instead of a SPEC package.",
		"- `$namba-queue` operates the existing-SPEC queue conveyor: start from a SPEC range or list, inspect durable status, resume after waits or blockers, and pause or stop without deleting branches, PRs, or evidence.",
		"- `$namba-review-resolve` resolves GitHub review threads one by one: discover unresolved thread state with a thread-aware GitHub path, classify meaningful feedback versus non-actionable remarks, reply on original threads with validation plus relevant CI/check evidence, and request review again only after the meaningful items are handled.",
		"- `$namba-release` handles NambaAI release orchestration: collect commits since the previous semver tag, draft release notes into a durable per-version artifact, and hand the release off through the guarded `namba release --version <version> --push` path.",
		"- `namba project` refreshes current repository docs and codemaps without creating a SPEC package.",
		"- `namba project` also writes `.namba/logs/project/codex-diagnostics-evidence.json` plus redacted `codex doctor` stdout/stderr logs when Codex diagnostics are available.",
		"- `namba codex access` inspects Namba runner `codex exec` access defaults and mutates `.namba/config/sections/system.yaml` only when explicit approval_policy / sandbox_mode flags are present.",
		"- Interactive Codex approval mode, permissions, service tier, effective workspace roots, models, auth, apps, web search, and platform sandbox choices stay user/session-owned; repo `.codex/config.toml` does not persist them.",
		"- Codex 0.131 may display approval mode, permissions, service tier, and effective workspace roots in the TUI. Treat those as observed runtime state, not Namba-owned policy.",
		"- Generated Codex lifecycle hooks are repo-local guardrails, not a complete security boundary: they add Namba context, guide ambiguous Namba prompts toward clarification questions without blocking submission, add approval-risk notes, check final-report format, deny destructive shell commands, and remind Codex about managed-surface changes.",
		"- Codex requires repo-local hooks to be reviewed before they run. In the first interactive session after init or regen, open `/hooks`, inspect the generated commands, and approve them only if they resolve to the current repository's `.codex/hooks/namba_codex_guard.sh` or Windows `.codex/hooks/namba_codex_guard.ps1` launcher for `.codex/hooks/namba_codex_guard.py`.",
		"- `namba regen` regenerates `AGENTS.md`, repo skills under `.agents/skills/`, `.codex/agents/*.toml` custom agents, readable `.md` role-card mirrors, `.namba/codex/*`, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.",
		"- `namba update` self-updates the installed `namba` binary from GitHub Release assets after explicit user approval. Use `--version vX.Y.Z` for a specific release.",
		"- `namba update` may print Codex 0.131 baseline advice, but it never installs, updates, or manages Codex.",
		"- `codex update` updates the upstream Codex CLI itself. Keep it separate from `namba update`.",
		"- `namba plan \"<description>\"` creates the next feature SPEC package plus review scaffolds.",
		"- `namba harness \"<description>\"` creates the next harness-oriented SPEC package plus review scaffolds while staying inside the standard `SPEC-XXX` model; harness/MCP plans should favor workflow-first design, bounded outputs, actionable errors, and stable read-only evaluations.",
		"- `namba fix --command plan \"<issue description>\"` creates the next bugfix SPEC package plus review scaffolds.",
		"- `namba fix \"<issue description>\"` and `namba fix --command run \"<issue description>\"` are the direct-repair paths in the current workspace. They should stay read-only for help/probe flows, avoid implicit SPEC creation, and finish with validation plus `namba sync`.",
		"- `namba sync` refreshes `.namba/project/*` docs, release notes/checklists, codemaps, and advisory review readiness summaries.",
		"- `namba queue start <SPEC-RANGE|SPEC-LIST>` processes existing SPEC packages in order through review, run, PR, checks, and optional land, while durable state under `.namba/logs/queue/` makes waits, blockers, and resume decisions explicit.",
		"- `namba pr` prepares the current branch for GitHub review by syncing, validating, inspecting PR checks, summarizing bounded GitHub Actions failure snippets when checks fail, committing, pushing, and opening or reusing the PR. Use `namba pr --review` only when Codex review should be requested.",
		"- `namba land` waits for checks when requested, merges a clean PR, and updates local `main` safely.",
		"- `namba release` requires a clean `main` branch and passing validators before it creates a tag. `--push` pushes both `main` and the new tag.",
		"- `namba run SPEC-XXX` keeps the standard standalone Codex flow when you use the CLI runner without extra mode flags, but explicit `frontend-major` work now reads `frontend-brief.md` as a canonical gate before coding.",
		"- `namba run SPEC-XXX --solo` requests a standalone Codex run that explicitly targets a single-subagent workflow inside one workspace.",
		"- `namba run SPEC-XXX --team` requests a standalone Codex run that explicitly coordinates multiple subagents inside one workspace.",
		"- `namba run SPEC-XXX --parallel` still refers to the standalone worktree runner path. It uses git worktrees, merges only after every worker passes execution and validation, and preserves failed worktrees and branches for inspection.",
		"- Codex `/goal` workflows are tracked as a future orchestration candidate, not a required Namba runtime dependency.",
	}
}

func renderCodexUsageInitEnablesSection() []string {
	return []string{
		"## What `namba init .` Enables",
		"",
		"- Creates `AGENTS.md` with Namba orchestration rules.",
		"- Creates repo-local skills under `.agents/skills/`, including read-only guidance plus command-entry skills such as `namba-help`, `namba-coach`, `namba-create`, `namba-run`, `namba-queue`, `namba-pr`, `namba-land`, `namba-release`, `namba-plan`, `namba-plan-review`, `namba-harness`, `namba-plan-pm-review`, `namba-plan-eng-review`, `namba-plan-design-review`, `namba-review-resolve`, and `namba-sync`.",
		"- Creates task-oriented Codex custom agents under `.codex/agents/*.toml` and readable `.md` role-card mirrors.",
		"- Creates repo-local Codex config under `.codex/config.toml`, keeping only deterministic repo-safe defaults such as lifecycle hook enablement, agent thread limits, status-line preference, and allow-listed repo-managed MCP presets when configured. It deliberately omits interactive approval mode, permissions, service tier, workspace roots, model, auth, apps, web search, and platform sandbox choices.",
		"- Creates repo-local Codex lifecycle hooks under `.codex/hooks.json` and `.codex/hooks/` with `features.hooks = true`, providing Namba context, non-blocking prompt-refinement guidance for ambiguous prompts, approval-risk notes, final-report format checks, destructive-command guardrails, duplicate Namba guard suppression, and generated-surface reminders. Codex will ask you to review these hooks in `/hooks` before they run.",
		"- Creates `.namba/codex/output-contract.md` plus `.namba/codex/validate-output-contract.py` for NambaAI response-shape guidance and fallback validation.",
		"- Creates `.namba/` project state, configs, docs, and SPEC storage.",
		"",
	}
}

func renderCodexUsageHowCodexUsesNambaSection() []string {
	return []string{
		"## How Codex Uses Namba After Init",
		"",
		"1. Open Codex in the initialized project directory.",
		"   On Windows, the current official Codex docs recommend using a WSL workspace for the best CLI experience.",
		"2. If Codex reports `6 hooks need review`, open `/hooks`, inspect that each generated command resolves to this repository's `.codex/hooks/namba_codex_guard.sh` or the Windows `.codex/hooks/namba_codex_guard.ps1` launcher for `.codex/hooks/namba_codex_guard.py`, and approve before expecting prompt-refinement hooks to run.",
		"3. Codex loads `AGENTS.md` and repo skills.",
		"4. Let the repo-local Codex lifecycle hook refine ambiguous Namba prompts before execution. When the goal, target surface, constraints, or acceptance criteria are underspecified, Codex should ask concise clarifying questions and restate the improved prompt before planning or editing.",
		"5. Invoke `$namba` for routing, `$namba-coach` for read-only current-goal command coaching, `$namba-help` for read-only Namba usage guidance, or command-entry skills such as `$namba-create`, `$namba-run`, `$namba-queue`, `$namba-pr`, `$namba-land`, `$namba-release`, `$namba-plan`, `$namba-plan-review`, `$namba-harness`, `$namba-fix`, `$namba-review-resolve`, `$namba-plan-pm-review`, `$namba-plan-eng-review`, `$namba-plan-design-review`, and `$namba-sync` for direct command-style execution.",
		"6. Use built-in Codex subagents such as `default`, `worker`, and `explorer`, plus project-scoped custom agents under `.codex/agents/*.toml`, when multi-agent work is appropriate. The matching `.md` files remain readable mirrors.",
		"7. Use the plan-review skills to update `.namba/specs/<SPEC>/reviews/*.md` and keep `.namba/specs/<SPEC>/reviews/readiness.md` current when a SPEC needs product, engineering, or design critique before implementation, or use `$namba-plan-review` when you want the create-plus-review loop bundled into one Codex entry point.",
		"8. Use `namba project`, `namba regen`, `namba update`, `namba codex access`, `namba plan`, `namba harness`, `namba fix`, `namba run SPEC-XXX`, `namba queue`, `namba sync`, `namba pr`, `namba land`, and `namba release` as workflow commands.",
	}
}

func renderCodexUsageAgentRosterSection() []string {
	return []string{
		"",
		"## Namba Custom Agent Roster",
		"",
		"- Strategy and readiness: `namba-product-manager` shapes scope and acceptance, `namba-planner` turns a SPEC into an execution plan, and `namba-plan-reviewer` validates whether the product/engineering/design review set is coherent enough to start implementation.",
		"- UI split: `namba-designer` owns art direction plus reference collection and synthesis, `namba-frontend-architect` plans hierarchy and state only after the frontend gate is satisfied, `namba-frontend-implementer` ships approved UI work only after synthesis plus design clearance, and `namba-mobile-engineer` handles mobile-specific constraints.",
		"- Routing examples: `Redesign this landing page hero so it stops looking generic` -> `namba-designer`; `Plan the component/state split for this dashboard` -> `namba-frontend-architect`; `Implement the approved dashboard filters and responsive states` -> `namba-frontend-implementer`.",
		"- Backend and data: `namba-backend-architect` plans service boundaries, `namba-backend-implementer` ships server-side changes, and `namba-data-engineer` owns data pipelines, transformations, migrations, and analytics-facing changes.",
		"- Security and delivery: `namba-security-engineer` handles hardening work, `namba-test-engineer` adds targeted regression coverage, `namba-devops-engineer` handles CI/CD and runtime changes, and `namba-reviewer` checks implementation acceptance before sync.",
		"- General delivery: `namba-implementer` remains the generalist execution agent for mixed-scope implementation slices.",
		"- Built-in Codex subagents such as `explorer` and `worker` still matter; use the Namba custom roster when responsibility and output expectations need tighter framing.",
	}
}

func renderCodexUsageDelegationHeuristicsSection() []string {
	return []string{
		"",
		"## Delegation Heuristics",
		"",
		"- Default `namba run` stays inside the standalone runner unless specialist signals are strong enough to justify delegation.",
		"- `--solo` uses at most one specialist when one domain clearly dominates the request.",
		"- `--team` prefers one specialist when one domain dominates and expands to two or three only when acceptance spans multiple domains.",
		"- Repo-managed same-workspace defaults set `.codex/config.toml [agents].max_threads = 5` when `agent_mode: multi`; Namba worktree workers remain separately controlled by `.namba/config/sections/workflow.yaml max_parallel_workers: 3` unless a later SPEC changes that fan-out.",
		"- Team mode honors each selected role's `model` and `model_reasoning_effort` metadata from `.codex/agents/*.toml`, keeping planner/reviewer/security roles stronger and delivery roles lighter.",
		"- Route art direction plus reference synthesis to `namba-designer`; route component, state, and delivery planning only after the frontend gate is satisfied to `namba-frontend-architect`; route approved UI implementation only after synthesis plus design clearance to `namba-frontend-implementer`; route mobile-specific delivery to `namba-mobile-engineer`; route API, schema, and pipeline work to backend/data; route auth, secrets, and compliance work to security; route deployment and runtime work to devops.",
		"- Keep the standalone runner as the integrator and final validation owner, and use `namba-reviewer` last when multiple specialists contribute.",
	}
}

func renderCodexUsagePlanReviewReadinessSection() []string {
	return []string{
		"",
		"## Plan Review Readiness",
		"",
		"- `namba plan`, `namba harness`, and `namba fix --command plan` seed `.namba/specs/<SPEC>/reviews/product.md`, `engineering.md`, `design.md`, and `readiness.md`.",
		"- Frontend-touching planning also seeds `.namba/specs/<SPEC>/frontend-brief.md`, and explicit `frontend-major` work uses that brief as the canonical gate contract.",
		"- `$namba-plan-review` bundles SPEC creation or resolution, the three review tracks, and an aggregate validation loop when you want that whole pre-implementation pass handled through one skill.",
		"- `$namba-plan-pm-review`, `$namba-plan-eng-review`, and `$namba-plan-design-review` update those review artifacts directly in the repository.",
		"- `namba run`, `namba sync`, and `namba pr` surface the latest readiness summary as advisory context for non-frontend and `frontend-minor` work, while explicit `frontend-major` runs can block on missing, insufficient, invalid, or mismatched frontend evidence.",
	}
}

func renderCodexUsageOutputContractSection(profile initProfile) []string {
	return []string{
		"",
		"## Output Contract",
		"",
		fmt.Sprintf("- `AGENTS.md` defines a Namba report header such as `%s` for substantial responses.", outputContractHeaderExample(profile)),
		fmt.Sprintf("- The report sections follow this semantic order: %s.", outputContractSequence(profile)),
		"- The semantic order stays fixed, but the exact labels can vary within the selected language palette so the writing does not become robotic.",
		"- The final next-work section should name the concrete command, review, validation, or handoff the operator should do next.",
		"- Hook-triggered rewrites should preserve concrete details from the prior answer, including changed files, file paths, commands, validation results, artifact paths, blockers, risks, and next steps.",
		"- `.namba/codex/validate-output-contract.py` checks this contract from a saved response file or stdin.",
		"- Namba keeps the validator script as the explicit repository enforcement path even as Codex's documented config and hook surface evolves.",
	}
}

func renderCodexUsageGitCollaborationSection(profile initProfile) []string {
	return []string{
		"",
		"## Git Collaboration Defaults",
		"",
		fmt.Sprintf("- Each SPEC or new task uses a dedicated branch from `%s`.", branchBase(profile)),
		fmt.Sprintf("- Recommended branch names: `%s<SPEC-ID>-<slug>` for SPEC work and `%s<slug>` for non-SPEC work.", specBranchPrefix(profile), taskBranchPrefix(profile)),
		fmt.Sprintf("- PRs target `%s`.", prBaseBranch(profile)),
		fmt.Sprintf("- PR titles and bodies should be written in %s.", humanLanguageName(profile.PRLanguage)),
		fmt.Sprintf("- Codex review requests are opt-in: use `namba pr --review` or queue `--review`, then confirm the `%s` request is present.", codexReviewComment(profile)),
	}
}

func renderCodexUsageClaudeMappingSection() []string {
	return []string{
		"",
		"## Claude to Codex Mapping",
		"",
		"- `CLAUDE.md` becomes `AGENTS.md`.",
		"- Claude skills become repo-local Codex skills under `.agents/skills/`.",
		"- Claude command wrappers become command-entry skills such as `$namba-create`, `$namba-run`, `$namba-queue`, `$namba-pr`, `$namba-land`, `$namba-plan`, `$namba-sync`, `$namba-review-resolve`, and `$namba-release`.",
		"- Claude subagents map to Codex built-in subagents plus project-scoped `.toml` custom agents, with `.md` mirrors kept for readability.",
		"- Claude hooks split into repo-local Codex lifecycle hooks under `.codex/hooks.json` plus Namba run hooks under `.namba/hooks.toml`, explicit validator scripts, documented response contracts, and sync steps.",
		"- Claude custom workflow commands become `$namba`, command-entry repo skills, built-in Codex slash commands, and the `namba` CLI.",
	}
}

func renderCodexUsageImportantDistinctionSection() []string {
	return []string{
		"",
		"## Important Distinction",
		"",
		"- In interactive Codex sessions, `namba run SPEC-XXX` means Codex should execute the SPEC directly in-session.",
		"- The standalone `namba run` CLI supports the default runner flow plus explicit `--solo`, `--team`, and worktree-based `--parallel` modes.",
		"- `.codex/hooks.json` configures Codex lifecycle hooks for interactive guardrails; `.namba/hooks.toml` configures Namba runner lifecycle hooks for `namba run` evidence and validation boundaries.",
		"- Codex 0.131 plugin hooks can also be enabled by the runtime. Namba-generated hooks dedupe identical Namba guard payloads at the interactive boundary, while `.namba/hooks.toml` remains separate runner evidence and is not used to dedupe interactive hooks.",
		"- Git helper commands may ignore configured repository hooks, so Namba validation must come from configured quality commands and run evidence rather than Git hooks alone.",
		"- Tokens and PATs are intentionally excluded from generated config. Use `gh auth login` or `glab auth login` instead.",
	}
}

func renderCodexStatusLineExample() string {
	return "[tui]\nstatus_line = [\"project-root\", \"git-branch\", \"current-dir\", \"model-with-reasoning\", \"context-remaining\", \"context-used\", \"used-tokens\", \"session-id\"]\n"
}

func renderModelRoutingDoc() string {
	return strings.Join([]string{
		"# Model Routing",
		"",
		"Policy: `gpt-5.6-cost-balanced-v1`.",
		"",
		"- `efficient` → `gpt-5.6-luna`",
		"- `standard` → `gpt-5.6-terra`",
		"- `deep` → `gpt-5.6-sol`",
		"",
		"Namba does not provide per-turn model or reasoning overrides. `requested` is the deterministic policy choice; `effective` is recorded only when Codex reports it; `external_unobserved` means that observation was unavailable and was not inferred.",
		"",
		"`namba run --dry-run` reports planned turns. Every planned explicit routed model is probed before execution. Unavailable Luna promotes to Terra only after Terra also passes its probe. Unavailable Terra or required Sol blocks preflight with `blocked_model_unavailable`; install or authorize the exact model and rerun the command after the capability probe succeeds.",
		"",
		"Sol is read-only for design, architecture, or high-risk checkpoints. Terra or Luna receives its checkpoint and performs implementation.",
	}, "\n") + "\n"
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
		"- `.claude/hooks/*` -> `.codex/hooks.json` lifecycle guardrails, `.namba/hooks.toml` runner hooks, explicit validation commands, output-contract validator scripts, structured run logs, and `namba sync`",
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
				{Emoji: "➡", Primary: "다음에 해야 할 작업", Aliases: []string{"다음에 해야 할 작업", "다음 작업", "다음 스텝", "다음 단계", "권장 작업", "권장 흐름"}},
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
				{Emoji: "➡", Primary: "次にやる作業", Aliases: []string{"次にやる作業", "次の作業", "次のステップ", "次の一手", "推奨フロー", "次段階"}},
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
				{Emoji: "➡", Primary: "下一步工作", Aliases: []string{"下一步工作", "下一步", "建议步骤", "推荐动作", "后续步骤"}},
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
				{Emoji: "➡", Primary: "Next Work", Aliases: []string{"Next Work", "Next Action", "Next Steps", "Recommended Next Steps", "Recommendations", "Next Move"}},
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
		"- The final next-work section should name the concrete command, review, validation, or handoff the operator should do next.",
		"- When a hook asks for a rewrite into this frame, preserve substantive details from the prior answer inside the sections, including changed files, file paths, commands, validation results, artifact paths, blockers, risks, and next steps.",
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
		threads = 5
	}

	lines := []string{
		"#:schema https://developers.openai.com/codex/config-schema.json",
		"# Generated by NambaAI from `.namba/config/sections/*.yaml`.",
		"# This file intentionally avoids session-owned Codex choices.",
		"# Keep models, auth, apps, web search, service tier, approval mode, permission profiles,",
		"# and platform-specific sandbox choices in your user-level Codex config or active session.",
		"# `.namba/config/sections/system.yaml` stores Namba runner defaults for `codex exec`,",
		"# not interactive Codex approval or permission policy.",
		"# Repo-managed MCP presets are the narrow exception when `.namba/config/sections/codex.yaml` opts in.",
		"# Reference: https://developers.openai.com/codex/config-reference/",
		"",
		"[features]",
		"hooks = true",
		"goals = true",
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

func renderNambaCodexHooksJSON() string {
	return renderNambaCodexHooksJSONForOS("")
}

func renderNambaCodexHooksJSONForOS(goos string) string {
	return strings.ReplaceAll(`{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup|resume|clear",
        "hooks": [
          {
            "type": "command",
            "command": "__NAMBA_CODEX_HOOK_COMMAND__",
            "timeout": 10,
            "statusMessage": "Loading NambaAI context"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "^Bash$",
        "hooks": [
          {
            "type": "command",
            "command": "__NAMBA_CODEX_HOOK_COMMAND__",
            "timeout": 10,
            "statusMessage": "Checking NambaAI shell safety"
          }
        ]
      }
    ],
    "PermissionRequest": [
      {
        "matcher": "^Bash$",
        "hooks": [
          {
            "type": "command",
            "command": "__NAMBA_CODEX_HOOK_COMMAND__",
            "timeout": 10,
            "statusMessage": "Checking NambaAI approval policy"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "__NAMBA_CODEX_HOOK_COMMAND__",
            "timeout": 10,
            "statusMessage": "Refining NambaAI prompt"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "^(Bash|apply_patch|Edit|Write)$",
        "hooks": [
          {
            "type": "command",
            "command": "__NAMBA_CODEX_HOOK_COMMAND__",
            "timeout": 10,
            "statusMessage": "Reviewing NambaAI surface changes"
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "__NAMBA_CODEX_HOOK_COMMAND__",
            "timeout": 10,
            "statusMessage": "Checking NambaAI report format"
          }
        ]
      }
    ]
  }
}
`, "__NAMBA_CODEX_HOOK_COMMAND__", codexHookCommandForOS(goos))
}

func codexHookCommandForOS(goos string) string {
	if goos == "windows" {
		return `powershell.exe -NoProfile -ExecutionPolicy Bypass -File .codex/hooks/namba_codex_guard.ps1`
	}
	return `sh .codex/hooks/namba_codex_guard.sh`
}

func renderNambaCodexHookGuardShellWrapper() string {
	return `#!/bin/sh
# NambaAI Codex lifecycle hook launcher for POSIX shells.

script_dir=${0%/*}
if [ "$script_dir" = "$0" ]; then
    script_dir="."
fi
script_path="$script_dir/namba_codex_guard.py"
raw_input=""

while IFS= read -r line || [ -n "$line" ]; do
    raw_input="${raw_input}${line}
"
done

failure_message="NambaAI hook guard failed: Python 3 was not found or could not run .codex/hooks/namba_codex_guard.py. Install Python 3, then run namba regen and review /hooks again. Hooks are guardrails, so continue only with explicit validation."
failed_candidates=""

extract_hook_event_name() {
    case "$raw_input" in
        *\"hook_event_name\"*)
            event_part=${raw_input#*\"hook_event_name\"}
            event_part=${event_part#*:}
            event_part=${event_part#*\"}
            event=${event_part%%\"*}
            ;;
        *)
            event=""
            ;;
    esac
    if [ -n "$event" ]; then
        printf '%s' "$event"
    else
        printf 'Unknown'
    fi
}

json_escape() {
    printf '%s' "$1"
}

emit_failure() {
    message=$1
    event_name=$2
    if [ -z "$event_name" ]; then
        event_name="Unknown"
    fi
    escaped_message=$(json_escape "$message")
    escaped_event=$(json_escape "$event_name")
    printf '%s\n' "$message" >&2
    printf '{"hookSpecificOutput":{"hookEventName":"%s","additionalContext":"%s"}}\n' "$escaped_event" "$escaped_message"
}

run_candidate() {
    executable=$1
    shift
    if ! command -v "$executable" >/dev/null 2>&1; then
        return 127
    fi

    stderr_path="${TMPDIR:-/tmp}/namba_codex_hook_$$.stderr"
    : > "$stderr_path" 2>/dev/null || stderr_path=""

    if [ -n "$stderr_path" ]; then
        output=$(printf '%s' "$raw_input" | "$executable" "$@" 2>"$stderr_path")
        status=$?
        stderr_text=""
        while IFS= read -r err_line || [ -n "$err_line" ]; do
            stderr_text="${stderr_text}${err_line} "
        done < "$stderr_path"
        : > "$stderr_path" 2>/dev/null || true
    else
        output=$(printf '%s' "$raw_input" | "$executable" "$@" 2>&1)
        status=$?
        stderr_text=""
    fi

    if [ "$status" -eq 0 ]; then
        if [ -n "$stderr_text" ]; then
            printf '%s\n' "$stderr_text" >&2
        fi
        if [ -n "$output" ]; then
            printf '%s\n' "$output"
        fi
        exit 0
    fi

    failed_candidates="${failed_candidates}${executable} exited ${status}; "
    if [ -n "$stderr_text" ]; then
        printf '%s failed: %s\n' "$executable" "$stderr_text" >&2
    fi
    return 1
}

run_candidate python3 "$script_path"
run_candidate python "$script_path"
run_candidate py -3 "$script_path"

if [ -n "$failed_candidates" ]; then
    printf 'NambaAI hook launcher candidates failed: %s\n' "$failed_candidates" >&2
fi
emit_failure "$failure_message" "$(extract_hook_event_name)"
exit 1
`
}

func renderNambaCodexHookGuardPowerShellWrapper() string {
	return `# NambaAI Codex lifecycle hook guard wrapper for Windows PowerShell 5.
$ErrorActionPreference = "Stop"

$scriptPath = Join-Path $PSScriptRoot "namba_codex_guard.py"
$rawInput = [Console]::In.ReadToEnd()
$eventName = "Unknown"
$failureMessage = "NambaAI hook guard failed: Python 3 was not found or could not run .codex/hooks/namba_codex_guard.py. Install Python 3 or the Python launcher, then run namba regen and review /hooks again. Hooks are guardrails, so continue only with explicit validation."

function Write-NambaHookFailure {
    param(
        [string]$Message,
        [string]$HookEventName
    )

    [Console]::Error.WriteLine($Message)
    $output = [ordered]@{
        hookSpecificOutput = [ordered]@{
            hookEventName = $HookEventName
            additionalContext = $Message
        }
    }
    $output | ConvertTo-Json -Compress -Depth 5
}

try {
    $trimmedInput = $rawInput.Trim()
    if ($trimmedInput.Length -gt 0) {
        $payload = $trimmedInput | ConvertFrom-Json -ErrorAction Stop
        if ($payload.hook_event_name) {
            $eventName = [string]$payload.hook_event_name
        }
    }
} catch {
    $eventName = "Unknown"
}

try {
    function Quote-NambaArgument {
        param([string]$Value)
        return '"' + ($Value -replace '"', '\"') + '"'
    }

    function Invoke-NambaPythonHook {
        param(
            [string]$Executable,
            [string[]]$Arguments,
            [string]$InputText
        )

        $startInfo = New-Object System.Diagnostics.ProcessStartInfo
        $startInfo.FileName = $Executable
        $startInfo.Arguments = ($Arguments | ForEach-Object { Quote-NambaArgument $_ }) -join " "
        $startInfo.UseShellExecute = $false
        $startInfo.RedirectStandardInput = $true
        $startInfo.RedirectStandardOutput = $true
        $startInfo.RedirectStandardError = $true
        $startInfo.CreateNoWindow = $true

        $process = New-Object System.Diagnostics.Process
        $process.StartInfo = $startInfo

        try {
            [void]$process.Start()
            $process.StandardInput.Write($InputText)
            $process.StandardInput.Close()
            $stdout = $process.StandardOutput.ReadToEnd()
            $stderr = $process.StandardError.ReadToEnd()
            $process.WaitForExit()
            return @{
                ExitCode = $process.ExitCode
                Stdout = $stdout
                Stderr = $stderr
            }
        } catch {
            return @{
                ExitCode = -1
                Stdout = ""
                Stderr = $_.Exception.Message
            }
        } finally {
            if ($process -ne $null) {
                $process.Dispose()
            }
        }
    }

    $candidates = @(
        @{ Name = "py.exe"; Args = @("-3", $scriptPath) },
        @{ Name = "py"; Args = @("-3", $scriptPath) },
        @{ Name = "python.exe"; Args = @($scriptPath) },
        @{ Name = "python"; Args = @($scriptPath) },
        @{ Name = "python3.exe"; Args = @($scriptPath) },
        @{ Name = "python3"; Args = @($scriptPath) }
    )

    $failedCandidates = @()
    foreach ($candidate in $candidates) {
        $resolved = Get-Command $candidate.Name -ErrorAction SilentlyContinue
        if (-not $resolved) {
            continue
        }

        $executable = $candidate.Name
        if ($resolved.Path) {
            $executable = $resolved.Path
        } elseif ($resolved.Source) {
            $executable = $resolved.Source
        }

        $result = Invoke-NambaPythonHook -Executable $executable -Arguments $candidate.Args -InputText $rawInput
        if ($result.Stderr.Trim().Length -gt 0) {
            [Console]::Error.Write($result.Stderr)
        }
        if ($result.ExitCode -eq 0) {
            if ($result.Stdout.Trim().Length -gt 0) {
                [Console]::Out.Write($result.Stdout)
            }
            exit 0
        }

        $summary = $candidate.Name + " exited " + $result.ExitCode
        if ($result.Stderr.Trim().Length -gt 0) {
            $summary = $summary + ": " + $result.Stderr.Trim()
        }
        $failedCandidates += $summary
    }

    if ($failedCandidates.Count -gt 0) {
        [Console]::Error.WriteLine("NambaAI hook launcher candidates failed: " + ($failedCandidates -join "; "))
    }

    Write-NambaHookFailure -Message $failureMessage -HookEventName $eventName
    exit 1
} catch {
    Write-NambaHookFailure -Message ("NambaAI hook launcher failed with an unhandled exception: " + $_.Exception.Message) -HookEventName $eventName
    exit 1
}
`
}

func renderNambaCodexHookGuardScript() string {
	return `#!/usr/bin/env python3
"""NambaAI Codex lifecycle hook guard.

This script is intentionally conservative. It adds lightweight NambaAI context,
blocks only clearly dangerous shell patterns, and reminds Codex to refresh
generated Namba surfaces when those files are touched.
"""

import json
import hashlib
import os
import re
import subprocess
import sys
import tempfile
import time


def configure_stdio():
    for stream in (sys.stdout, sys.stderr):
        reconfigure = getattr(stream, "reconfigure", None)
        if reconfigure:
            try:
                reconfigure(encoding="utf-8", errors="backslashreplace")
            except Exception:
                pass


DANGEROUS_PATTERNS = [
    (re.compile(r"\bgit\s+reset\s+--hard\b", re.IGNORECASE), "git reset --hard discards repository changes."),
    (re.compile(r"\bgit\s+clean\s+-[^\n;&|]*[fd][^\n;&|]*", re.IGNORECASE), "git clean can delete untracked files."),
    (re.compile(r"\bgit\s+push\b[^\n;&|]*(?:\s-f\b|\s--force(?:-with-lease)?\b)", re.IGNORECASE), "force-pushing requires an explicit human decision."),
    (re.compile(r"\brm\s+-[^\n;&|]*r[^\n;&|]*f[^\n;&|]*\s+(?:/|\*|\.|~|\$HOME)(?:\s|$|/)", re.IGNORECASE), "broad rm -rf targets are blocked by the NambaAI hook guard."),
    (re.compile(r"\brm\s+-[^\n;&|]*f[^\n;&|]*r[^\n;&|]*\s+(?:/|\*|\.|~|\$HOME)(?:\s|$|/)", re.IGNORECASE), "broad rm -rf targets are blocked by the NambaAI hook guard."),
    (re.compile(r"\bchmod\s+-R\s+777\b", re.IGNORECASE), "recursive chmod 777 is too broad for an automated turn."),
    (re.compile(r"\b(?:curl|wget)\b[^\n|;&]*(?:\||>)\s*(?:sudo\s+)?(?:sh|bash)\b", re.IGNORECASE), "piping downloaded content directly into a shell is blocked."),
]

MANAGED_EXACT = {
    "AGENTS.md",
    ".codex/config.toml",
    ".codex/hooks.json",
    ".namba/codex/README.md",
    ".namba/codex/output-contract.md",
    ".namba/codex/validate-output-contract.py",
}

MANAGED_PREFIXES = (
    ".agents/skills/",
    ".codex/agents/",
    ".codex/hooks/",
    ".namba/config/",
)

PROMPT_REFINEMENT_PATTERN = re.compile(
    r"(\bnamba\s+(?:plan|harness|fix|run)\b|\$namba-(?:plan|harness|fix|run|create|queue|review-resolve)|"
    r"\b(?:build|make|fix|improve|update|create|implement)\b|"
    r"(?:만들어|고쳐|수정|개선|구현|추가|해줘|해봐|처리|진행))",
    re.IGNORECASE,
)

VAGUE_MARKERS = (
    "뭔가",
    "대충",
    "적당히",
    "알아서",
    "좋게",
    "이거",
    "그거",
    "애매",
    "문제",
    "버그",
    "개선",
    "만들어줘",
    "해줘",
    "fix it",
    "make it better",
    "improve",
    "something",
    "stuff",
    "thing",
)

REPORT_HEADER = "NAMBA-AI 작업 결과 보고"
REPORT_SECTIONS = ("작업 정의", "판단", "수행한 작업", "현재 이슈", "잠재 문제", "다음에 해야 할 작업")
REPORT_SECTION_ALIASES = (
    ("작업 정의", "정의", "정의한 범위", "문제 정의"),
    ("판단", "내린 판단", "핵심 판단", "결정"),
    ("수행한 작업", "진행한 작업", "작업 내용", "적용한 작업"),
    ("현재 이슈", "이슈", "남은 이슈", "현재 문제"),
    ("잠재 문제", "잠재 리스크", "위험 요소", "잠재 이슈"),
    ("다음에 해야 할 작업", "다음 작업", "다음 스텝", "다음 단계", "권장 작업", "권장 흐름"),
)
CURRENT_PAYLOAD = {}
configure_stdio()


def event_name_from_raw(raw):
    match = re.search(r'"hook_event_name"\s*:\s*"([^"]+)"', raw)
    if match:
        return match.group(1)
    return "Unknown"


def read_payload():
    raw = sys.stdin.read()
    if not raw.strip():
        return {}, ""
    try:
        return json.loads(raw), ""
    except json.JSONDecodeError as exc:
        return {
            "hook_event_name": event_name_from_raw(raw),
            "_namba_malformed_json": True,
        }, "NambaAI hook guard received malformed JSON payload; continuing without policy-specific action: " + str(exc)


def emit(value):
    trace(CURRENT_PAYLOAD, value)
    print(json.dumps(value, ensure_ascii=True, separators=(",", ":")))


def emit_continue(message="", suppress_output=False):
    value = {"continue": True}
    if message:
        value["systemMessage"] = message
    if suppress_output:
        value["suppressOutput"] = True
    emit(value)


def trace(payload, output=None):
    path = os.environ.get("NAMBA_HOOK_TRACE_PATH")
    if not path:
        return
    try:
        event = payload.get("hook_event_name")
        record = {"event": event}
        session_id = session_id_from(payload)
        if session_id:
            record["session_id"] = session_id
        if event == "UserPromptSubmit":
            record["prompt_refinement"] = bool(prompt_refinement_context(prompt_from(payload)))
        if event in ("PreToolUse", "PermissionRequest"):
            command = command_from(payload)
            record["dangerous"] = bool(dangerous_reason(command))
            record["approval_risk_note"] = bool(approval_risk_note(command))
        if event == "Stop":
            message = payload.get("last_assistant_message")
            record["would_check_report"] = isinstance(message, str) and len(message.strip()) >= 500
        if output is not None:
            record["emitted"] = True
            record["output_keys"] = sorted(output.keys()) if isinstance(output, dict) else []
            if isinstance(output, dict):
                record["decision"] = output.get("decision")
                hook_output = output.get("hookSpecificOutput")
                if isinstance(hook_output, dict):
                    record["hook_event"] = hook_output.get("hookEventName")
        directory = os.path.dirname(path)
        if directory:
            os.makedirs(directory, exist_ok=True)
        with open(path, "a", encoding="utf-8") as handle:
            handle.write(json.dumps(record, ensure_ascii=False, separators=(",", ":")) + "\n")
    except Exception:
        return


def command_from(payload):
    tool_input = payload.get("tool_input")
    if isinstance(tool_input, dict):
        for key in ("command", "cmd"):
            value = tool_input.get(key)
            if isinstance(value, str):
                return value
    return ""


def session_id_from(payload):
    for key in ("session_id", "session-id", "thread_id", "thread-id"):
        value = payload.get(key)
        if isinstance(value, str) and value.strip():
            return value.strip()
    return ""


def safe_updated_input(payload):
    tool_input = payload.get("tool_input")
    if not isinstance(tool_input, dict):
        return None
    if "cmd" not in tool_input or "command" in tool_input:
        return None
    cmd = tool_input.get("cmd")
    if not isinstance(cmd, str) or not cmd.strip():
        return None
    updated = dict(tool_input)
    updated["command"] = updated.pop("cmd")
    return updated


def dedupe_key(payload):
    try:
        raw = json.dumps(payload, ensure_ascii=True, sort_keys=True, separators=(",", ":"))
    except Exception:
        raw = str(payload)
    return hashlib.sha256(raw.encode("utf-8", errors="replace")).hexdigest()


def should_suppress_duplicate(payload):
    flag = os.environ.get("NAMBA_HOOK_DEDUPE", "1").strip().lower()
    if flag in ("0", "false", "no", "off"):
        return False
    event = payload.get("hook_event_name")
    if not isinstance(event, str) or not event:
        return False
    root = os.environ.get("NAMBA_HOOK_DEDUPE_DIR")
    if not root:
        root = os.path.join(tempfile.gettempdir(), "namba_codex_hook_guard")
    try:
        os.makedirs(root, exist_ok=True)
        path = os.path.join(root, dedupe_key(payload) + ".seen")
        now = time.time()
        if os.path.exists(path) and now - os.path.getmtime(path) <= 5:
            return True
        with open(path, "w", encoding="utf-8") as handle:
            handle.write(str(now))
    except Exception:
        return False
    return False


def dangerous_reason(command):
    for pattern, reason in DANGEROUS_PATTERNS:
        if pattern.search(command):
            return reason
    return ""


def approval_risk_note(command):
    command = command.strip()
    if not command:
        return ""
    lowered = command.lower()
    risks = []
    if "sudo" in lowered:
        risks.append("uses sudo and may modify system-level state")
    if re.search(r"\b(?:curl|wget|npm|pnpm|yarn|pip|uv|go)\b", lowered):
        risks.append("may access the network or install executable dependencies")
    if re.search(r"\b(?:git\s+push|gh\s+pr\s+merge|gh\s+release|namba\s+land|namba\s+release)\b", lowered):
        risks.append("may publish, merge, or release remote state")
    if re.search(r"(^|\s)(?:/usr|/opt|/etc|/var|~|\$HOME|/tmp)(?:/|\s|$)", command):
        risks.append("mentions paths outside the repository workspace")
    if re.search(r"\b(?:chmod|chown|launchctl|systemctl|killall)\b", lowered):
        risks.append("may alter permissions or local runtime processes")
    if not risks:
        return ""
    return "NambaAI approval note: " + "; ".join(risks) + ". Approve only if this matches the user's latest intent and the command is scoped."


def prompt_from(payload):
    prompt = payload.get("prompt")
    return prompt if isinstance(prompt, str) else ""


def prompt_refinement_context(prompt):
    normalized = " ".join(prompt.split())
    if not normalized or not PROMPT_REFINEMENT_PATTERN.search(normalized):
        return ""
    lower = normalized.lower()
    namba_workflow = bool(re.search(r"\bnamba\s+(?:plan|harness|fix|run)\b|\$namba-", lower))
    too_short = len(normalized) < 80
    vague = any(marker in lower for marker in VAGUE_MARKERS)
    has_acceptance = bool(re.search(r"acceptance|criteria|success|완료|성공|검증|테스트|제약|constraint|scope|범위", lower))
    if not namba_workflow and not too_short and not vague:
        return ""
    if has_acceptance and len(normalized) >= 120 and not vague:
        return ""
    return (
        "NambaAI prompt-refinement gate: treat this as a spec-first request before implementation. "
        "If the goal, target surface, constraints, acceptance criteria, or existing context are unclear, "
        "stop the current workflow and ask 1-3 concise clarifying questions before running any Namba workflow or editing files. "
        "If native Codex Plan mode choice UI is available, use it as the clarification surface and feed its output into the matching SPEC creation command (namba plan, namba harness, or namba fix --command plan) only after it is restated as Goal, Scope, Constraints, Acceptance. "
        "Continue the question loop across turns until the improved prompt can be restated as Goal, Scope, Constraints, Acceptance. "
        "This is inspired by a Socratic/Ontology workflow: reduce ambiguity before code."
    )


def prompt_refinement_guidance(prompt, configured_language=""):
    context = prompt_refinement_context(prompt)
    if not context:
        return ""
    normalized = " ".join(prompt.split())
    language = (configured_language or "").strip().lower()
    if not language:
        language = "ko" if re.search(r"[가-힣]", normalized) else "en"
    if language == "ko":
        questions = [
            "1. 이 작업의 대상 surface는 무엇인가요? 예: CLI, 웹 앱, API, 특정 모듈.",
            "2. 원하는 사용자 흐름과 제외할 범위는 무엇인가요?",
            "3. 완료 기준과 검증 방법은 무엇인가요?",
        ]
        return (
            "NambaAI 프롬프트 교정 게이트: 지금은 SPEC을 만들지 말고 모호함을 먼저 줄여야 합니다.\n"
            "가능하면 Codex Plan mode 선택 UI로 아래 질문을 먼저 처리한 뒤, 정리된 Goal/Scope/Constraints/Acceptance만 해당 SPEC 생성 명령(namba plan, namba harness, namba fix --command plan)에 넘기세요.\n"
            + "\n".join(questions)
            + "\n\n답변은 가능하면 다음 형식으로 주세요:\n"
            "Goal: ...\nScope: ...\nConstraints: ...\nAcceptance: ..."
        )
    if language == "ja":
        questions = [
            "1. この作業の対象 surface は何ですか？例: CLI、Web app、API、特定モジュール。",
            "2. 望むユーザーフローと、今回の範囲外にするものは何ですか？",
            "3. 完了基準と検証方法は何ですか？",
        ]
        return (
            "NambaAI clarification gate: SPEC を作る前に、依頼がまだ広すぎるか曖昧です。\n"
            "可能なら Codex Plan mode の選択 UI で以下の質問を先に整理し、整理された Goal/Scope/Constraints/Acceptance だけを該当する SPEC 作成コマンド(namba plan, namba harness, namba fix --command plan)に渡してください。\n"
            + "\n".join(questions)
            + "\n\n可能なら次の形式で答えてください:\n"
            "Goal: ...\nScope: ...\nConstraints: ...\nAcceptance: ..."
        )
    if language == "zh":
        questions = [
            "1. 这项变更影响哪个目标 surface？例如 CLI、Web app、API 或特定模块。",
            "2. 期望的用户流程是什么？哪些内容应排除在范围外？",
            "3. 完成标准和验证方法是什么？",
        ]
        return (
            "NambaAI clarification gate: 在创建 SPEC 前，请先降低需求的模糊度。\n"
            "如果可以，请先用 Codex Plan mode 选择 UI 整理以下问题，然后只把整理后的 Goal/Scope/Constraints/Acceptance 传给对应的 SPEC 创建命令(namba plan, namba harness, namba fix --command plan)。\n"
            + "\n".join(questions)
            + "\n\n可以的话，请使用以下格式回答:\n"
            "Goal: ...\nScope: ...\nConstraints: ...\nAcceptance: ..."
        )
    questions = [
        "1. What target surface should this change affect, such as CLI, web app, API, or a specific module?",
        "2. What user flow is desired, and what should stay out of scope?",
        "3. What acceptance criteria and validation should define done?",
    ]
    return (
        "NambaAI prompt-refinement gate: do not create a SPEC yet; reduce ambiguity before execution.\n"
        "When native Codex Plan mode choice UI is available, use it first, then feed the refined Goal/Scope/Constraints/Acceptance output into the matching SPEC creation command: namba plan, namba harness, or namba fix --command plan.\n"
        + "\n".join(questions)
        + "\n\nPlease answer in this shape when possible:\n"
        "Goal: ...\nScope: ...\nConstraints: ...\nAcceptance: ..."
    )


def handle_user_prompt_submit(payload):
    guidance = prompt_refinement_guidance(prompt_from(payload), configured_prompt_language(cwd_from_payload(payload)))
    if not guidance:
        emit_continue()
        return
    emit({
        "hookSpecificOutput": {
            "hookEventName": "UserPromptSubmit",
            "additionalContext": guidance,
        }
    })


def handle_session_start(payload):
    session_id = session_id_from(payload)
    session_note = " Session metadata is optional and tolerated."
    if session_id:
        session_note = " Session id accepted: " + session_id + "."
    emit({
        "hookSpecificOutput": {
            "hookEventName": "SessionStart",
            "additionalContext": (
                "NambaAI lifecycle hook is active. Treat .namba/ as the source "
                "of truth, use AGENTS.md and .agents/skills/ for workflow routing, "
                "and run configured validation after changes. Codex hooks are "
                "guardrails, not a complete security boundary." + session_note
            ),
        }
    })


def handle_pre_tool_use(payload):
    reason = dangerous_reason(command_from(payload))
    if reason:
        emit({
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "deny",
                "permissionDecisionReason": reason,
            }
        })
        return
    updated = safe_updated_input(payload)
    if updated is not None:
        emit({
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "allow",
                "updatedInput": updated,
            }
        })
        return
    emit_continue()


def handle_permission_request(payload):
    reason = dangerous_reason(command_from(payload))
    if not reason:
        note = approval_risk_note(command_from(payload))
        if note:
            emit({"continue": True, "systemMessage": note})
        else:
            emit_continue()
        return
    emit({
        "hookSpecificOutput": {
            "hookEventName": "PermissionRequest",
            "decision": {
                "behavior": "deny",
                "message": reason,
            },
        }
    })


def cwd_from_payload(payload):
    value = payload.get("cwd")
    if isinstance(value, str) and value.strip():
        return value
    try:
        return os.getcwd()
    except Exception:
        return ""


def changed_paths(cwd):
    if not cwd:
        return []
    try:
        result = subprocess.run(
            ["git", "status", "--short"],
            cwd=cwd,
            text=True,
            capture_output=True,
            timeout=5,
            check=False,
        )
    except Exception:
        return []
    if result.returncode != 0:
        return []
    paths = []
    for line in result.stdout.splitlines():
        if len(line) < 4:
            continue
        path = line[3:].strip()
        if " -> " in path:
            path = path.split(" -> ", 1)[1].strip()
        path = path.strip('"')
        if path:
            paths.append(path)
    return paths


def is_namba_surface(path):
    if path in MANAGED_EXACT:
        return True
    return any(path.startswith(prefix) for prefix in MANAGED_PREFIXES)


def handle_post_tool_use(payload):
    cwd = cwd_from_payload(payload)
    touched = [path for path in changed_paths(cwd) if is_namba_surface(path)]
    if not touched:
        emit_continue()
        return
    shown = ", ".join(touched[:6])
    if len(touched) > 6:
        shown += ", ..."
    emit({
        "hookSpecificOutput": {
            "hookEventName": "PostToolUse",
            "additionalContext": (
                "Namba-managed instruction or config surfaces are changed: "
                + shown
                + ". Update source templates first, run namba regen when scaffold "
                "outputs changed, then run the configured validation commands."
            ),
        }
    })


def namba_repo(cwd):
    if not cwd:
        return False
    return os.path.exists(os.path.join(cwd, ".namba")) or os.path.exists(os.path.join(cwd, "AGENTS.md"))


def namba_root(cwd):
    if not cwd:
        return ""
    root = os.path.abspath(cwd)
    while True:
        if os.path.exists(os.path.join(root, ".namba")) or os.path.exists(os.path.join(root, "AGENTS.md")):
            return root
        parent = os.path.dirname(root)
        if parent == root:
            return ""
        root = parent


def read_simple_config(path):
    values = {}
    try:
        with open(path, "r", encoding="utf-8") as handle:
            for line in handle:
                stripped = line.strip()
                if not stripped or stripped.startswith("#") or ":" not in stripped:
                    continue
                key, value = stripped.split(":", 1)
                values[key.strip()] = value.strip().strip("\"'")
    except OSError:
        return values
    return values


def configured_prompt_language(cwd):
    root = namba_root(cwd)
    if not root:
        return ""
    language_cfg = read_simple_config(os.path.join(root, ".namba", "config", "sections", "language.yaml"))
    for key in ("conversation_language", "documentation_language", "comment_language"):
        value = language_cfg.get(key, "").strip().lower()
        if value:
            return value
    git_cfg = read_simple_config(os.path.join(root, ".namba", "config", "sections", "git-strategy.yaml"))
    return git_cfg.get("pr_language", "").strip().lower()


def report_missing_sections(message):
    if REPORT_HEADER not in message:
        return list(REPORT_SECTIONS)
    lines = message.splitlines()
    positions = []
    for index, aliases in enumerate(REPORT_SECTION_ALIASES):
        section_index = -1
        pattern = re.compile(r"^\s*(?:#{1,6}\s*|[-*]\s+)?(?:\*\*)?[\W_]*(?:" + "|".join(re.escape(alias) for alias in aliases) + r")(?:\*\*)?\s*(?:[:：-].*)?$")
        for line_index, line in enumerate(lines):
            if pattern.match(line.strip()):
                section_index = line_index
                break
        if section_index < 0:
            return [REPORT_SECTIONS[index]]
        positions.append(section_index)
    if positions != sorted(positions):
        return ["section order"]
    return []


def handle_stop(payload):
    if payload.get("stop_hook_active") is True:
        emit_continue()
        return
    message = payload.get("last_assistant_message")
    if not isinstance(message, str) or len(message.strip()) < 500:
        emit_continue()
        return
    cwd = cwd_from_payload(payload)
    if not namba_repo(cwd):
        emit_continue()
        return
    nambaish = any(token in message.lower() for token in ("namba", ".namba", "codex", "spec-", "검증", "브랜치"))
    if not nambaish:
        emit_continue()
        return
    missing = report_missing_sections(message)
    if not missing:
        emit_continue()
        return
    emit({
        "decision": "block",
        "reason": (
            "Before ending, rewrite the final response using the Namba report frame: "
            "# NAMBA-AI 작업 결과 보고, then 🧭 작업 정의, 🧠 판단, 🛠 수행한 작업, "
            "🚧 현재 이슈, ⚠ 잠재 문제, ➡ 다음에 해야 할 작업. "
            "Preserve substantive details from the prior answer inside that frame, including changed files, "
            "file paths, commands, validation results, artifact paths, blockers, risks, and next steps. "
            "In the final section, name the concrete command, review, validation, or handoff to do next. "
            "Keep it concise and high-signal."
        ),
    })


def emit_hook_error(event, message):
    print(message, file=sys.stderr)
    emit_continue(message)


def main():
    global CURRENT_PAYLOAD
    try:
        payload, parse_error = read_payload()
        CURRENT_PAYLOAD = payload
        trace(payload)
        if parse_error:
            emit_continue(parse_error)
            return 0
        if should_suppress_duplicate(payload):
            emit_continue(suppress_output=True)
            return 0
        event = payload.get("hook_event_name")
        if event == "SessionStart":
            handle_session_start(payload)
        elif event == "PreToolUse":
            handle_pre_tool_use(payload)
        elif event == "PermissionRequest":
            handle_permission_request(payload)
        elif event == "UserPromptSubmit":
            handle_user_prompt_submit(payload)
        elif event == "PostToolUse":
            handle_post_tool_use(payload)
        elif event == "Stop":
            handle_stop(payload)
        else:
            emit_continue()
        return 0
    except Exception as exc:
        event = CURRENT_PAYLOAD.get("hook_event_name") if isinstance(CURRENT_PAYLOAD, dict) else "Unknown"
        emit_hook_error(event, "NambaAI hook guard failed with an unhandled exception: " + str(exc))
        return 0


if __name__ == "__main__":
    sys.exit(main())
`
}

func renderRoleCard(title, useWhen string, responsibilities []string) string {
	lines := []string{
		fmt.Sprintf("# %s", title),
		"",
		useWhen,
		"",
	}
	lines = append(lines, generatedInstructionContractSection("role card")...)
	lines = append(lines,
		"",
		"Responsibilities:",
	)
	for _, responsibility := range responsibilities {
		lines = append(lines, "- "+responsibility)
	}
	return strings.Join(lines, "\n") + "\n"
}

func generatedInstructionContractSection(surface string) []string {
	return []string{
		fmt.Sprintf("Generated instruction contract for this %s:", surface),
		"- Purpose: keep the role or command scope explicit, bounded, and testable.",
		"- Boundary: honor read-only versus mutating state effects, configured sandbox mode, and assigned file or workflow ownership.",
		"- Required output: report concrete actions, changed paths or artifacts, validation evidence, and pass/fail status or blockers.",
		"- Pass/fail criteria: claim success only when acceptance criteria and configured validation are satisfied; otherwise name the exact blocker and impact.",
		"- Evidence expectations: cite source artifacts such as SPEC files, `.namba/` configs, diffs, test output, PR/check links, or generated manifests instead of relying on unsupported assertions.",
		"- Security responsibilities: never expose or commit secrets; treat auth, privacy, destructive commands, permission changes, and external network or credential use as security-sensitive.",
		"- Destructive command and escalation policy: do not run destructive commands unless explicitly requested; request approval for privileged, networked, or sandbox-blocked actions only when the active approval mode allows it, and otherwise report the blocker or use a safe non-escalating path.",
		"- Fallback implementer boundary: if a specialist path is unavailable and the main/default implementer takes over, stay within the assigned scope and preserve the same evidence and validation duties.",
		"- Portability: keep durable guidance non-project-specific unless the current repository config or SPEC explicitly provides the project detail.",
	}
}

type codexAgentTemplate struct {
	roleTitle                   string
	roleUseWhen                 string
	roleResponsibilities        []string
	customAgentName             string
	customAgentDescription      string
	customAgentSandboxMode      string
	customAgentUseWhen          string
	customAgentResponsibilities []string
}

func renderTemplatedRoleCard(template codexAgentTemplate) string {
	return renderRoleCard(template.roleTitle, template.roleUseWhen, template.roleResponsibilities)
}

func renderTemplatedCustomAgent(template codexAgentTemplate) string {
	lines := []string{
		fmt.Sprintf("You are %s.", template.roleTitle),
		"",
		template.customAgentUseWhen,
		"",
	}
	lines = append(lines, generatedInstructionContractSection("custom agent")...)
	lines = append(lines,
		"",
		"Responsibilities:",
	)
	for _, responsibility := range template.customAgentResponsibilities {
		lines = append(lines, "- "+responsibility)
	}
	return renderCustomAgentWithOptions(
		template.customAgentName,
		template.customAgentDescription,
		template.customAgentSandboxMode,
		lines,
	)
}

func plannerAgentTemplate() codexAgentTemplate {
	return codexAgentTemplate{
		roleTitle:              "Namba Planner",
		roleUseWhen:            "Use this role when breaking down a SPEC package before implementation.",
		roleResponsibilities:   []string{"Read `spec.md`, `plan.md`, and `acceptance.md`.", "Identify target files, risks, and validation commands.", "Produce a concise execution plan for the main session.", "Do not edit files directly."},
		customAgentName:        "namba-planner",
		customAgentDescription: "Break down a SPEC package into an execution plan without editing files.",
		customAgentSandboxMode: "read-only",
		customAgentUseWhen:     "Use this custom agent when breaking down a SPEC package before implementation.",
		customAgentResponsibilities: []string{
			"Read `spec.md`, `plan.md`, and `acceptance.md`.",
			"When repo-managed MCP presets are configured, consult them first when they can ground planning decisions with better source material or verification signals.",
			"Identify target files, risks, and validation commands.",
			"Produce a concise execution plan for the main session.",
			"Do not edit files directly.",
		},
	}
}

func renderPlannerRoleCard() string {
	return renderTemplatedRoleCard(plannerAgentTemplate())
}

func renderPlannerCustomAgent() string {
	return renderTemplatedCustomAgent(plannerAgentTemplate())
}

func renderHighRiskAdvisorRoleCard() string {
	return renderRoleCard("Namba High-Risk Advisor", "Use this role only for a critical, ambiguous, cross-system or irreversible decision checkpoint.", []string{
		"Read the relevant contract and source evidence.",
		"Report decision, risks, evidence, and the next Terra/Luna implementation handoff.",
		"Do not edit files, run mutating commands, or implement the change.",
	})
}

func renderHighRiskAdvisorCustomAgent() string {
	return renderCustomAgentWithOptions("namba-high-risk-advisor", "Read-only advisor for one high-risk Namba model-routing checkpoint.", "read-only", []string{
		"You are Namba High-Risk Advisor.",
		"Use this role only for a critical, ambiguous, cross-system or irreversible decision checkpoint.",
		"Read the relevant contract and source evidence.",
		"Do not edit files or run mutating commands. Do not implement the change.",
		"Return only the decision, risks, evidence, open questions, and an explicit Terra/Luna implementation handoff.",
	})
}

func planReviewerAgentTemplate() codexAgentTemplate {
	return codexAgentTemplate{
		roleTitle:   "Namba Plan Reviewer",
		roleUseWhen: "Use this role for aggregate validation of plan-review artifacts before implementation starts.",
		roleResponsibilities: []string{
			"Read `spec.md`, `plan.md`, `acceptance.md`, and the review artifacts under `.namba/specs/<SPEC>/reviews/`.",
			"Check whether the product, engineering, and design review set is coherent, sufficiently deep, and reflected correctly in `readiness.md`.",
			"Call out contradictions, missing review depth, or weak acceptance coverage, and identify which review tracks need to rerun.",
			"Do not implement code or quietly turn the advisory review flow into a hidden hard gate.",
		},
		customAgentName:        "namba-plan-reviewer",
		customAgentDescription: "Validate plan-review coherence and advisory readiness before implementation starts.",
		customAgentSandboxMode: "read-only",
		customAgentUseWhen:     "Use this custom agent for aggregate validation of plan-review artifacts before implementation starts.",
		customAgentResponsibilities: []string{
			"Read `spec.md`, `plan.md`, `acceptance.md`, and the review artifacts under `.namba/specs/<SPEC>/reviews/`.",
			"Check whether the product, engineering, and design review set is coherent, sufficiently deep, and reflected correctly in `readiness.md`.",
			"Call out contradictions, missing review depth, or weak acceptance coverage, and identify which review tracks need to rerun.",
			"Keep the review flow advisory unless the main session explicitly asks for a hard gate.",
			"Do not implement code or rewrite unrelated files.",
		},
	}
}

func renderPlanReviewerRoleCard() string {
	return renderTemplatedRoleCard(planReviewerAgentTemplate())
}

func renderPlanReviewerCustomAgent() string {
	return renderTemplatedCustomAgent(planReviewerAgentTemplate())
}

func productManagerAgentTemplate() codexAgentTemplate {
	return codexAgentTemplate{
		roleTitle:   "Namba Product Manager",
		roleUseWhen: "Use this role when shaping scope, acceptance, and delivery slicing before implementation.",
		roleResponsibilities: []string{
			"Translate user goals into concrete scope, constraints, and success criteria.",
			"Tighten acceptance criteria, non-goals, and rollout boundaries.",
			"Break large ideas into deliverable slices the main session can schedule.",
			"Call out UX, data, and operational implications early.",
		},
		customAgentName:        "namba-product-manager",
		customAgentDescription: "Shape product scope, acceptance, and delivery slicing before implementation.",
		customAgentSandboxMode: "read-only",
		customAgentUseWhen:     "Use this custom agent when a request needs stronger product framing before implementation starts.",
		customAgentResponsibilities: []string{
			"Translate user goals into concrete scope, constraints, and success criteria.",
			"Tighten acceptance criteria, non-goals, and rollout boundaries.",
			"Break large ideas into deliverable slices the main session can schedule.",
			"Call out UX, data, and operational implications early.",
			"Do not implement code directly.",
		},
	}
}

func renderProductManagerRoleCard() string {
	return renderTemplatedRoleCard(productManagerAgentTemplate())
}

func renderProductManagerCustomAgent() string {
	return renderTemplatedCustomAgent(productManagerAgentTemplate())
}

func frontendArchitectAgentTemplate() codexAgentTemplate {
	roleResponsibilities := []string{
		"Identify component boundaries, state ownership, and data flow.",
		"Map UI changes to file targets, design-system constraints, and accessibility impact.",
		"Highlight responsive, performance, and browser-risk considerations.",
		"Recommend the smallest coherent UI implementation slice.",
		"Verify that `frontend-major` synthesis and design clearance exist before planning hierarchy, file structure, or state ownership.",
		"Use the Do-Not Design Contract architecture handoff before planning component boundaries, state ownership, responsive behavior, or file structure.",
		"Plan file/module boundaries around the approved asset manifest so generated or existing imagery is treated as implementation input, not decoration added after generic layout work.",
		"Refuse generic primitives that the contract bans unless an explicit exception path is cited.",
	}
	return codexAgentTemplate{
		roleTitle:              "Namba Frontend Architect",
		roleUseWhen:            "Use this role when frontend structure, state flow, or UI delivery planning needs to be clarified before editing.",
		roleResponsibilities:   roleResponsibilities,
		customAgentName:        "namba-frontend-architect",
		customAgentDescription: "Plan frontend structure, component boundaries, state flow, and UI delivery risks.",
		customAgentSandboxMode: "read-only",
		customAgentUseWhen:     "Use this custom agent when a task needs frontend planning before implementation starts.",
		customAgentResponsibilities: append(
			append([]string{}, roleResponsibilities...),
			"Do not edit files directly.",
		),
	}
}

func renderFrontendArchitectRoleCard() string {
	return renderTemplatedRoleCard(frontendArchitectAgentTemplate())
}

func renderFrontendArchitectCustomAgent() string {
	return renderTemplatedCustomAgent(frontendArchitectAgentTemplate())
}

func renderFrontendImplementerRoleCard() string {
	return renderRoleCard(
		"Namba Frontend Implementer",
		"Use this role when implementing approved UI work after frontend synthesis is cleared.",
		[]string{
			"Change only the frontend files assigned by the main session.",
			"Preserve design-system conventions, accessibility, and responsive behavior.",
			"Keep loading, empty, and error states coherent with the surrounding UI.",
			"Run or report the relevant UI validation steps when feasible.",
			"Do not start `frontend-major` implementation until `frontend-brief.md` and design review agree on an approved direction.",
			"Implement within the approved visual grammar and include Do-Not Design Violation Check evidence before claiming completion.",
			"When the frontend brief says `Imagegen requirement: required`, generate the required bitmap assets before final layout, save them under the planned project asset paths, wire them into their intended UI elements, and report per-asset Generated Asset Evidence.",
		},
	)
}

func renderFrontendImplementerCustomAgent() string {
	instructions := []string{
		"You are Namba Frontend Implementer.",
		"",
		"Use this custom agent when implementing approved UI work after frontend synthesis is cleared.",
		"",
	}
	instructions = append(instructions, generatedInstructionContractSection("custom agent")...)
	instructions = append(instructions,
		"",
		"Responsibilities:",
		"- Change only the frontend files assigned by the main session.",
		"- Preserve design-system conventions, accessibility, and responsive behavior.",
		"- Keep loading, empty, and error states coherent with the surrounding UI.",
		"- Run or report the relevant UI validation steps when feasible.",
		"- Do not start `frontend-major` implementation until `frontend-brief.md` and design review agree on an approved direction.",
		"- Implement within the approved visual grammar and include Do-Not Design Violation Check evidence before claiming completion.",
		"- When the frontend brief says `Imagegen requirement: required`, generate the required bitmap assets before final layout, save them under the planned project asset paths, wire them into their intended UI elements, and report per-asset Generated Asset Evidence.",
	)
	return renderCustomAgentWithOptions(
		"namba-frontend-implementer",
		"Implement approved frontend work with design-system, accessibility, and responsive discipline.",
		"workspace-write",
		instructions,
	)
}

func designerAgentTemplate() codexAgentTemplate {
	roleResponsibilities := []string{
		"Lead with art direction: define the visual concept, hierarchy, and composition before defaulting to components or spacing tokens.",
		"Set palette logic with explicit temperature and undertone discipline, restrained saturation, and deliberate accent use instead of trend-chasing or washed-out minimalism.",
		"Choose semantic components and layout primitives that fit the content; do not default to interchangeable cards, border-heavy framing, or generic bento/grid patterns as the primary identity.",
		"Own the Do-Not Design Contract: apply the default anti-pattern library, define context-specific bans, allowed replacements, visual grammar, generated-image decision fields, repeatable reference-driven asset manifest, generated-image plan readiness, architecture handoff, and generic-section proof.",
		"When first frontend or first major screen work is asset-led, default to `Asset mode: generated-images` and define each generated imagegen asset, including icons, thumbnails, sprites, cutouts, textures, badges, and object images, instead of allowing brand-color-only imitation.",
		"Keep motion purposeful: use it only when it clarifies hierarchy, attention, or state change.",
		"For screen-, page-, or section-scale work, identify the most generic-looking section and propose a concrete redesign; for component-scale work, call out the risk without forcing gratuitous scope creep.",
		"Guard against overcorrection: do not flatten everything into gray minimalism, do not add novelty without payoff, and do not sacrifice accessibility, design-system fit, or implementation realism.",
		"Collect or critique references, synthesize adopt/avoid/why decisions, and define banned patterns before implementation begins.",
	}
	return codexAgentTemplate{
		roleTitle:              "Namba Designer",
		roleUseWhen:            "Use this role when art direction, palette/tone logic, composition, motion intent, or generic-looking UI surfaces need to be clarified before implementation or review.",
		roleResponsibilities:   roleResponsibilities,
		customAgentName:        "namba-designer",
		customAgentDescription: "Clarify art direction, palette logic, composition, motion intent, and anti-generic redesign guidance before implementation.",
		customAgentSandboxMode: "read-only",
		customAgentUseWhen:     "Use this custom agent when a task needs design direction, taste correction, or visual review before implementation starts.",
		customAgentResponsibilities: append(
			append([]string{}, roleResponsibilities...),
			"Do not edit files directly.",
		),
	}
}

func renderDesignerRoleCard() string {
	return renderTemplatedRoleCard(designerAgentTemplate())
}

func renderDesignerCustomAgent() string {
	return renderTemplatedCustomAgent(designerAgentTemplate())
}

func mobileEngineerAgentTemplate() codexAgentTemplate {
	roleResponsibilities := []string{
		"Define mobile component boundaries, platform-specific constraints, and ownership of shared versus native behavior.",
		"Map requested changes to navigation, lifecycle, offline, and responsive considerations.",
		"Highlight gesture, performance, and device-compatibility risks.",
		"Recommend the smallest mobile delivery slice the main session can delegate safely.",
	}
	return codexAgentTemplate{
		roleTitle:              "Namba Mobile Engineer",
		roleUseWhen:            "Use this role when mobile-specific constraints, navigation, lifecycle, or platform behavior need to be clarified before editing.",
		roleResponsibilities:   roleResponsibilities,
		customAgentName:        "namba-mobile-engineer",
		customAgentDescription: "Plan mobile-specific structure, navigation, lifecycle, and platform risks before implementation.",
		customAgentSandboxMode: "read-only",
		customAgentUseWhen:     "Use this custom agent when a task needs mobile-specific planning before implementation starts.",
		customAgentResponsibilities: append(
			append([]string{}, roleResponsibilities...),
			"Do not edit files directly.",
		),
	}
}

func renderMobileEngineerRoleCard() string {
	return renderTemplatedRoleCard(mobileEngineerAgentTemplate())
}

func renderMobileEngineerCustomAgent() string {
	return renderTemplatedCustomAgent(mobileEngineerAgentTemplate())
}

func backendArchitectAgentTemplate() codexAgentTemplate {
	roleResponsibilities := []string{
		"Define API, service, and persistence boundaries for the requested change.",
		"Call out schema, transaction, idempotency, and rollback risks.",
		"Identify security, observability, and migration implications.",
		"Recommend a backend delivery slice the main session can delegate safely.",
	}
	return codexAgentTemplate{
		roleTitle:              "Namba Backend Architect",
		roleUseWhen:            "Use this role when backend contracts, service boundaries, or persistence changes need to be clarified before implementation.",
		roleResponsibilities:   roleResponsibilities,
		customAgentName:        "namba-backend-architect",
		customAgentDescription: "Plan backend contracts, service boundaries, persistence changes, and delivery risks.",
		customAgentSandboxMode: "read-only",
		customAgentUseWhen:     "Use this custom agent when a task needs backend planning before implementation starts.",
		customAgentResponsibilities: append(
			append([]string{}, roleResponsibilities...),
			"Do not edit files directly.",
		),
	}
}

func renderBackendArchitectRoleCard() string {
	return renderTemplatedRoleCard(backendArchitectAgentTemplate())
}

func renderBackendArchitectCustomAgent() string {
	return renderTemplatedCustomAgent(backendArchitectAgentTemplate())
}

func backendImplementerAgentTemplate() codexAgentTemplate {
	roleResponsibilities := []string{
		"Change only the backend files assigned by the main session.",
		"Keep API contracts, validation, and persistence logic internally consistent.",
		"Add or update targeted backend tests when the change affects behavior.",
		"Report migration, rollout, or compatibility risks with the patch.",
	}
	return codexAgentTemplate{
		roleTitle:                   "Namba Backend Implementer",
		roleUseWhen:                 "Use this role when implementing approved server-side work.",
		roleResponsibilities:        roleResponsibilities,
		customAgentName:             "namba-backend-implementer",
		customAgentDescription:      "Implement approved server-side work across APIs, services, persistence, and backend tests.",
		customAgentSandboxMode:      "workspace-write",
		customAgentUseWhen:          "Use this custom agent when implementing approved server-side work.",
		customAgentResponsibilities: append([]string{}, roleResponsibilities...),
	}
}

func renderBackendImplementerRoleCard() string {
	return renderTemplatedRoleCard(backendImplementerAgentTemplate())
}

func renderBackendImplementerCustomAgent() string {
	return renderTemplatedCustomAgent(backendImplementerAgentTemplate())
}

func dataEngineerAgentTemplate() codexAgentTemplate {
	roleResponsibilities := []string{
		"Own data-model, migration, ETL, query, and analytics-facing code assigned by the main session.",
		"Keep schema changes, backfills, and data contracts internally consistent.",
		"Call out rollout sequencing, data quality risks, and irreversible migration concerns.",
		"Add or update focused validation for the changed data behavior when feasible.",
	}
	return codexAgentTemplate{
		roleTitle:                   "Namba Data Engineer",
		roleUseWhen:                 "Use this role when schema, migration, pipeline, analytics, or transformation work is part of the change.",
		roleResponsibilities:        roleResponsibilities,
		customAgentName:             "namba-data-engineer",
		customAgentDescription:      "Handle schema, migration, pipeline, and analytics-facing changes with explicit data-quality and rollout discipline.",
		customAgentSandboxMode:      "workspace-write",
		customAgentUseWhen:          "Use this custom agent when schema, migration, pipeline, analytics, or transformation work is part of the change.",
		customAgentResponsibilities: append([]string{}, roleResponsibilities...),
	}
}

func renderDataEngineerRoleCard() string {
	return renderTemplatedRoleCard(dataEngineerAgentTemplate())
}

func renderDataEngineerCustomAgent() string {
	return renderTemplatedCustomAgent(dataEngineerAgentTemplate())
}

func securityEngineerAgentTemplate() codexAgentTemplate {
	roleResponsibilities := []string{
		"Own security-sensitive code paths assigned by the main session.",
		"Tighten auth, permission, secret-handling, validation, and privacy boundaries without widening scope.",
		"Call out exploitability, compliance, rollback, and incident-response implications.",
		"Prefer the smallest defensible hardening patch plus explicit regression notes.",
	}
	return codexAgentTemplate{
		roleTitle:                   "Namba Security Engineer",
		roleUseWhen:                 "Use this role when authentication, authorization, secrets, privacy, or hardening work is part of the change.",
		roleResponsibilities:        roleResponsibilities,
		customAgentName:             "namba-security-engineer",
		customAgentDescription:      "Handle authentication, authorization, secrets, privacy, and hardening work with explicit security discipline.",
		customAgentSandboxMode:      "workspace-write",
		customAgentUseWhen:          "Use this custom agent when authentication, authorization, secrets, privacy, or hardening work is part of the change.",
		customAgentResponsibilities: append([]string{}, roleResponsibilities...),
	}
}

func renderSecurityEngineerRoleCard() string {
	return renderTemplatedRoleCard(securityEngineerAgentTemplate())
}

func renderSecurityEngineerCustomAgent() string {
	return renderTemplatedCustomAgent(securityEngineerAgentTemplate())
}

func testEngineerAgentTemplate() codexAgentTemplate {
	return codexAgentTemplate{
		roleTitle:   "Namba Test Engineer",
		roleUseWhen: "Use this role when acceptance coverage or regression protection needs to be strengthened.",
		roleResponsibilities: []string{
			"Turn acceptance criteria into concrete test scenarios and edge cases.",
			"Add the smallest high-value automated coverage for the changed behavior.",
			"Focus on regression detection rather than broad refactors.",
			"Report residual gaps when full automation is not practical.",
		},
		customAgentName:        "namba-test-engineer",
		customAgentDescription: "Design and add targeted regression coverage that tightens SPEC acceptance confidence.",
		customAgentSandboxMode: "workspace-write",
		customAgentUseWhen:     "Use this custom agent when acceptance coverage or regression protection needs to be strengthened.",
		customAgentResponsibilities: []string{
			"Turn acceptance criteria into concrete test scenarios and edge cases.",
			"Add the smallest high-value automated coverage for the changed behavior.",
			"Focus on regression detection rather than broad test refactors.",
			"Report residual gaps when full automation is not practical.",
		},
	}
}

func renderTestEngineerRoleCard() string {
	return renderTemplatedRoleCard(testEngineerAgentTemplate())
}

func renderTestEngineerCustomAgent() string {
	return renderTemplatedCustomAgent(testEngineerAgentTemplate())
}

func devOpsEngineerAgentTemplate() codexAgentTemplate {
	roleResponsibilities := []string{
		"Own pipeline, environment, container, and deployment-file changes assigned by the main session.",
		"Preserve release safety, rollback clarity, and secret-handling boundaries.",
		"Call out observability, operational risk, and environment drift.",
		"Keep infrastructure edits tightly scoped to the requested outcome.",
	}
	return codexAgentTemplate{
		roleTitle:                   "Namba DevOps Engineer",
		roleUseWhen:                 "Use this role when CI, runtime config, deployment, or operational automation is part of the change.",
		roleResponsibilities:        roleResponsibilities,
		customAgentName:             "namba-devops-engineer",
		customAgentDescription:      "Handle CI/CD, runtime config, deployment, and operational automation with explicit release safety.",
		customAgentSandboxMode:      "workspace-write",
		customAgentUseWhen:          "Use this custom agent when CI, runtime config, deployment, or operational automation is part of the change.",
		customAgentResponsibilities: append([]string{}, roleResponsibilities...),
	}
}

func renderDevOpsEngineerRoleCard() string {
	return renderTemplatedRoleCard(devOpsEngineerAgentTemplate())
}

func renderDevOpsEngineerCustomAgent() string {
	return renderTemplatedCustomAgent(devOpsEngineerAgentTemplate())
}

func implementerAgentTemplate() codexAgentTemplate {
	roleResponsibilities := []string{
		"Change only the files assigned by the main session.",
		"Preserve methodology rules from `.namba/config/sections/quality.yaml`.",
		"Run or report the relevant validation steps when feasible.",
		"Leave notes about validation status and residual risk.",
	}
	return codexAgentTemplate{
		roleTitle:                   "Namba Implementer",
		roleUseWhen:                 "Use this role when implementing an approved portion of a SPEC package.",
		roleResponsibilities:        roleResponsibilities,
		customAgentName:             "namba-implementer",
		customAgentDescription:      "Implement approved SPEC work while preserving Namba quality rules.",
		customAgentSandboxMode:      "workspace-write",
		customAgentUseWhen:          "Use this custom agent when implementing an approved portion of a SPEC package.",
		customAgentResponsibilities: append([]string{}, roleResponsibilities...),
	}
}

func renderImplementerRoleCard() string {
	return renderTemplatedRoleCard(implementerAgentTemplate())
}

func renderImplementerCustomAgent() string {
	return renderTemplatedCustomAgent(implementerAgentTemplate())
}

func reviewerAgentTemplate() codexAgentTemplate {
	return codexAgentTemplate{
		roleTitle:   "Namba Reviewer",
		roleUseWhen: "Use this role for acceptance and quality review before sync.",
		roleResponsibilities: []string{
			"Compare the implementation with `acceptance.md`.",
			"Check that validation output and artifacts exist.",
			"Call out regressions, missing tests, or documentation drift.",
			"Do not rewrite the implementation unless asked.",
		},
		customAgentName:        "namba-reviewer",
		customAgentDescription: "Review implementation quality and acceptance coverage before sync.",
		customAgentSandboxMode: "read-only",
		customAgentUseWhen:     "Use this custom agent for acceptance and quality review before sync.",
		customAgentResponsibilities: []string{
			"Compare the implementation with `acceptance.md`.",
			"Check that validation output and expected artifacts exist.",
			"Call out regressions, missing tests, and documentation drift.",
			"Do not rewrite the implementation unless explicitly asked.",
		},
	}
}

func renderReviewerRoleCard() string {
	return renderTemplatedRoleCard(reviewerAgentTemplate())
}

func renderReviewerCustomAgent() string {
	return renderTemplatedCustomAgent(reviewerAgentTemplate())
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
	lines = append(lines, fmt.Sprintf("- Codex review requests are explicit opt-in only: use `namba pr --review` or queue `--review`, then confirm the `%s` request exists without duplicating it.", codexReviewComment(profile)))
	return strings.Join(lines, "\n") + "\n"
}

func renderCodexProfileConfig(profile initProfile) string {
	lines := []string{
		fmt.Sprintf("agent_mode: %s", profile.AgentMode),
		fmt.Sprintf("status_line_preset: %s", profile.StatusLinePreset),
		fmt.Sprintf("repo_skills_path: %s", repoSkillsDir),
		fmt.Sprintf("repo_agents_path: %s", repoCodexAgentsDir),
		"# Namba-managed per-turn model routing. Do not combine this with legacy `model:`.",
		"model_routing_policy: gpt-5.6-cost-balanced-v1",
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
