package namba

import (
	"fmt"
	"strings"
)

func usageText() string {
	lines := []string{
		"NambaAI CLI",
		"",
		"Usage:",
		"  namba help [command]",
	}
	lines = append(lines, publicTopLevelCommandUsageSummaries()...)
	return strings.Join(lines, "\n") + "\n"
}

func initUsageText() string {
	lines := []string{
		"namba init",
		"",
		"Usage:",
		"  namba init [path] [--yes] [--name NAME] [--mode tdd|ddd] [--project-type new|existing]",
		"  namba init [path] [--human-language LANG] [--approval-policy POLICY] [--sandbox-mode MODE]",
		"",
		"Behavior:",
		"  Initialize the NambaAI scaffold, config, and repo-local Codex assets in the target directory.",
		"  The interactive wizard starts with a language-first screen, then explains repository-state defaults in the selected language.",
		"  Existing code keeps detected stack defaults, while empty repositories leave the app stack unset until the first planned request.",
		"  Selections are echoed before moving on, `b`/`back` returns to the previous step, and the wizard does not ask for a GitHub username.",
		"  Plain terminals use numeric/code choices and text markers such as [default], [recommended], [current], and [next] without relying on emoji, ANSI styling, or raw-key movement.",
		"  Codex access presets preview the resulting approval_policy / sandbox_mode pair.",
		"  The scaffold includes Codex lifecycle hooks; first interactive Codex use must review them with `/hooks` before prompt-refinement guardrails run.",
		"  After bootstrap, use `namba codex access` from the project root to inspect or change Namba runner Codex access defaults.",
	}
	return strings.Join(lines, "\n") + "\n"
}

func doctorUsageText() string {
	return strings.Join([]string{
		"namba doctor",
		"",
		"Usage:",
		"  namba doctor [--check-update]",
		"",
		"Behavior:",
		"  Inspect the current repository and local toolchain readiness without mutating project files.",
		"  Use --check-update to explicitly refresh official GitHub Release metadata for an advisory-only NambaAI CLI version check.",
		"  The check never runs `namba update`; ask the user before running `namba update`.",
	}, "\n") + "\n"
}

func statusUsageText() string {
	return strings.Join([]string{
		"namba status",
		"",
		"Usage:",
		"  namba status [--json]",
		"",
		"Behavior:",
		"  Print a read-only summary of the current NambaAI repository state.",
	}, "\n") + "\n"
}

func projectUsageText() string {
	return singleUsageLineCommandUsageText(
		"project",
		"  namba project",
		"  Refresh .namba/project/* docs and codemaps for the current repository.",
	)
}

func regenUsageText() string {
	return singleUsageLineCommandUsageText(
		"regen",
		"  namba regen",
		"  Regenerate AGENTS, repo-local skills, Codex agents, and Codex config from .namba/config/sections/*.yaml.",
	)
}

func updateUsageText() string {
	return singleUsageLineCommandUsageText(
		"update",
		"  namba update [--version vX.Y.Z]",
		"  Download and install the requested NambaAI release for the current platform.",
	)
}

func (a *App) printPlanUsage() error {
	_, err := fmt.Fprint(a.stdout, planUsageText())
	return err
}

func (a *App) printHarnessUsage() error {
	_, err := fmt.Fprint(a.stdout, harnessUsageText())
	return err
}

func (a *App) printFixUsage() error {
	_, err := fmt.Fprint(a.stdout, fixUsageText())
	return err
}

func planUsageText() string {
	lines := []string{
		"namba plan",
		"",
		"Usage:",
	}
	lines = append(lines, descriptionScaffoldUsageLines("namba plan", "description")...)
	lines = append(lines,
		fmt.Sprintf("  namba plan %s \"<description>\"", noReviewPlanningFlag),
		"",
		"Behavior:",
		"  Create the next feature SPEC package under .namba/specs/ and seed review artifacts.",
		"  Clarification gate: vague requests such as \"게시판 만들어줘\" ask questions before any SPEC is created.",
		fmt.Sprintf("  Safe by default: create and switch to a dedicated SPEC branch in the current workspace unless you explicitly pass %s.", currentWorkspacePlanningFlag),
		fmt.Sprintf("  Auto review by default: Codex should continue with `$namba-plan-review SPEC-XXX` unless you pass %s.", noReviewPlanningFlag),
	)
	return strings.Join(lines, "\n") + "\n"
}

func harnessUsageText() string {
	return descriptionScaffoldUsageText(
		"harness",
		"  Create the next harness-oriented SPEC package under .namba/specs/ and seed review artifacts.",
		fmt.Sprintf("  Safe by default: create and switch to a dedicated SPEC branch in the current workspace unless you explicitly pass %s.", currentWorkspacePlanningFlag),
	)
}

func descriptionScaffoldUsageText(command string, behaviorLines ...string) string {
	lines := []string{
		fmt.Sprintf("namba %s", command),
		"",
		"Usage:",
	}
	lines = append(lines, descriptionScaffoldUsageLines("namba "+command, "description")...)
	lines = append(lines, "", "Behavior:")
	lines = append(lines, behaviorLines...)
	return strings.Join(lines, "\n") + "\n"
}

func descriptionScaffoldUsageLines(invocation, subject string) []string {
	return []string{
		fmt.Sprintf("  %s \"<%s>\"", invocation, subject),
		fmt.Sprintf("  %s -- \"<%s with flag-like text>\"", invocation, subject),
		fmt.Sprintf("  %s %s \"<%s>\"", invocation, currentWorkspacePlanningFlag, subject),
	}
}

func fixUsageText() string {
	lines := []string{
		"namba fix",
		"",
		"Usage:",
	}
	lines = append(lines,
		"  namba fix [--command run|plan] \"<issue description>\"",
		"  namba fix [--command run|plan] -- \"<issue description with flag-like text>\"",
		fmt.Sprintf("  namba fix --command plan %s \"<issue description>\"", currentWorkspacePlanningFlag),
	)
	lines = append(lines, "", "Behavior:")
	lines = append(lines, fixSubcommandBehaviorSummaries()...)
	lines = append(lines, fmt.Sprintf("  Use %s with --command plan when you intentionally want to scaffold on the current branch without creating a dedicated SPEC branch.", currentWorkspacePlanningFlag))
	return strings.Join(lines, "\n") + "\n"
}

func singleUsageLineCommandUsageText(command, usageLine, behaviorLine string) string {
	lines := []string{
		fmt.Sprintf("namba %s", command),
		"",
		"Usage:",
		usageLine,
		"",
		"Behavior:",
		behaviorLine,
	}
	return strings.Join(lines, "\n") + "\n"
}

func runUsageText() string {
	return singleUsageLineCommandUsageText(
		"run",
		"  namba run SPEC-XXX [--solo|--team|--parallel] [--dry-run]",
		"  Execute the selected SPEC package with one runner, same-workspace team routing, or managed worktree fan-out.",
	)
}

func syncUsageText() string {
	return singleUsageLineCommandUsageText(
		"sync",
		"  namba sync",
		"  Refresh README bundles, project docs, review readiness summaries, and PR/release support artifacts.",
	)
}

func prUsageText() string {
	return singleUsageLineCommandUsageText(
		"pr",
		"  namba pr \"<title>\" [--review] [--language en|ko|ja|zh] [--remote origin] [--no-sync] [--no-validate]",
		"  Sync, validate, push the current work branch, and create or reuse a GitHub pull request. Add --review to request Codex review.",
	)
}

func landUsageText() string {
	return singleUsageLineCommandUsageText(
		"land",
		"  namba land [PR_NUMBER] [--wait] [--remote origin]",
		"  Merge an approved pull request into the base branch and refresh the local base branch checkout.",
	)
}

func releaseUsageText() string {
	return singleUsageLineCommandUsageText(
		"release",
		"  namba release [--bump patch|minor|major] [--version vX.Y.Z] [--language en|ko|ja|zh] [--push] [--remote origin]",
		"  Create a release tag from a clean main branch and optionally push main plus the tag.",
	)
}

func worktreeUsageText() string {
	lines := []string{
		"namba worktree",
		"",
		"Usage:",
	}
	lines = append(lines, worktreeSubcommandUsageSummaries()...)
	lines = append(lines,
		"",
		"Behavior:",
		"  Manage Namba-owned git worktrees under .namba/worktrees.",
	)
	return strings.Join(lines, "\n") + "\n"
}
