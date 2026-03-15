package namba

import "fmt"

func renderAgents(projectName string) string {
	return fmt.Sprintf("# NambaAI\n\n"+
		"You are the NambaAI orchestrator for this repository.\n\n"+
		"## Workflow\n\n"+
		"Use NambaAI's explicit workflow:\n\n"+
		"1. `namba project` to refresh project docs and codemaps\n"+
		"2. `namba plan \"<description>\"` to create a SPEC package\n"+
		"3. `namba run SPEC-XXX` to execute the SPEC with Codex\n"+
		"4. `namba sync` to refresh artifacts and PR-ready documents\n\n"+
		"## Rules\n\n"+
		"- Prefer the `.namba/` state as the source of truth.\n"+
		"- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before implementation.\n"+
		"- Use the `.codex/skills/` assets when relevant.\n"+
		"- Do not bypass validation. Run the configured quality commands after changes.\n"+
		"- Use worktrees for parallel execution; do not modify multiple branches in one workspace.\n\n"+
		"Project: %s\n", projectName)
}

func renderFoundationSkill() string {
	return `---
name: namba-foundation-core
description: Core NambaAI workflow, SPEC discipline, TRUST quality gates, and execution rules.
---

Use this skill when the task involves NambaAI workflow orchestration, SPEC handling,
quality gates, or phased delivery.

Key ideas:
- SPEC-first execution
- TDD for greenfield or sufficiently tested projects
- DDD-style preserve/improve flow for brownfield projects with weak test coverage
- TRUST gates after each execution phase
- Worktree-based isolation for parallel work
`
}

func renderProjectSkill() string {
	return `---
name: namba-workflow-project
description: Project analysis, codemap refresh, and documentation generation for NambaAI.
---

Use this skill to:
- refresh project docs
- summarize structure and entry points
- rebuild codemap artifacts under .namba/project/codemaps
- explain how the repository is organized before planning or execution
`
}

func renderExecutionSkill() string {
	return `---
name: namba-workflow-execution
description: Execute NambaAI SPEC packages with Codex and explicit validation.
---

Use this skill when implementing a SPEC package.

Execution pattern:
1. Read .namba/specs/<SPEC>/spec.md
2. Read .namba/specs/<SPEC>/plan.md
3. Read .namba/specs/<SPEC>/acceptance.md
4. Implement changes
5. Run configured validation commands
6. Summarize results in .namba/logs and sync artifacts
`
}

func renderProjectConfig(name, language, framework string) string {
	return fmt.Sprintf("name: %s\nlanguage: %s\nframework: %s\n", name, language, framework)
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
