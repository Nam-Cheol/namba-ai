package namba

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type syncContext struct {
	Root       string
	ProjectCfg projectConfig
	LatestSpec string
	Profile    initProfile
	DocsCfg    docsConfig
	Support    syncSupportContext
}

func (a *App) runSync(_ context.Context, args []string) error {
	if handled, err := a.handleNoArgTopLevelCommand("sync", args); handled {
		return err
	}

	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}

	syncCtx, err := a.loadSyncContext(root)
	if err != nil {
		return err
	}
	qualityCfg, _ := a.loadQualityConfig(root)
	analysisCfg, err := a.loadAnalysisConfig(root)
	if err != nil {
		return err
	}

	readmeOutputs := buildReadmeOutputs(syncCtx.ProjectCfg, syncCtx.Profile, syncCtx.DocsCfg)
	analysis := analyzeProject(root, syncCtx.ProjectCfg, qualityCfg, analysisCfg)
	projectOutputs := analysis.renderOutputs()
	readinessBatch, err := buildSpecReviewReadinessBatch(root)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	syncCtx.Support = a.buildSyncSupportContext(root, syncCtx.LatestSpec, readinessBatch.Advisories)

	session, err := a.beginManagedOutputSession(root)
	if err != nil {
		return err
	}
	if err := session.replaceManagedOutputs(readmeOutputs, isReadmeManagedPath, nil); err != nil {
		return err
	}
	if err := session.replaceManagedOutputs(projectOutputs, isProjectAnalysisManagedPath, nil); err != nil {
		return err
	}
	for _, warning := range analysis.Quality.Warnings {
		fmt.Fprintf(a.stdout, "Project analysis warning: %s\n", warning)
	}
	if len(analysis.Quality.Errors) > 0 {
		if _, err := session.commit(); err != nil {
			return err
		}
		for _, item := range analysis.Quality.Errors {
			fmt.Fprintf(a.stdout, "Project analysis error: %s\n", item)
		}
		return errors.New("project analysis quality gate failed")
	}

	if err := session.replaceManagedOutputs(readinessBatch.Outputs, isSpecReviewReadinessManagedPath, nil); err != nil {
		return err
	}
	if err := session.replaceManagedOutputs(buildSyncProjectSupportOutputs(syncCtx), isSyncProjectSupportManagedPath, nil); err != nil {
		return err
	}
	if _, err := session.commit(); err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "Synced NambaAI artifacts.")
	a.printCachedVersionAdvisory()
	return nil
}

func (a *App) loadSyncContext(root string) (syncContext, error) {
	projectCfg, _ := a.loadProjectConfig(root)
	latestSpec, _ := latestSpecID(filepath.Join(root, specsDir))
	profile, err := a.loadInitProfileFromConfig(root)
	if err != nil {
		return syncContext{}, err
	}
	docsCfg, err := a.loadDocsConfig(root)
	if err != nil {
		return syncContext{}, err
	}
	support := a.buildSyncSupportContext(root, latestSpec, nil)
	return syncContext{
		Root:       root,
		ProjectCfg: projectCfg,
		LatestSpec: latestSpec,
		Profile:    profile,
		DocsCfg:    docsCfg,
		Support:    support,
	}, nil
}

func (a *App) materializeSyncReadme(syncCtx syncContext) error {
	if _, err := a.replaceManagedOutputs(syncCtx.Root, buildReadmeOutputs(syncCtx.ProjectCfg, syncCtx.Profile, syncCtx.DocsCfg), isReadmeManagedPath, nil); err != nil {
		return err
	}
	return nil
}

func (a *App) refreshSyncProjectArtifacts(ctx context.Context, syncCtx syncContext) error {
	if err := a.runProject(ctx, nil); err != nil {
		return err
	}
	if err := a.refreshAllSpecReviewReadiness(syncCtx.Root); err != nil {
		return err
	}
	return nil
}

func (a *App) writeSyncProjectSupportDocs(syncCtx syncContext) error {
	outputs := buildSyncProjectSupportOutputs(syncCtx)
	if err := a.materializeSyncProjectSupportOutputs(syncCtx.Root, outputs); err != nil {
		return err
	}
	return nil
}

func buildSyncProjectSupportOutputs(syncCtx syncContext) map[string]string {
	return map[string]string{
		filepath.ToSlash(filepath.Join(projectDir, "change-summary.md")):    buildChangeSummaryDocWithSupport(syncCtx.ProjectCfg, syncCtx.Profile, syncCtx.Support),
		filepath.ToSlash(filepath.Join(projectDir, "pr-checklist.md")):      buildPRChecklistDocWithSupport(syncCtx.Profile, syncCtx.Support),
		filepath.ToSlash(filepath.Join(projectDir, "release-notes.md")):     buildReleaseNotesDoc(syncCtx.ProjectCfg, syncCtx.LatestSpec, syncCtx.Profile),
		filepath.ToSlash(filepath.Join(projectDir, "release-checklist.md")): buildReleaseChecklistDoc(),
	}
}

func (a *App) materializeSyncProjectSupportOutputs(root string, outputs map[string]string) error {
	if _, err := a.writeOutputs(root, outputs); err != nil {
		return err
	}
	return nil
}

func buildChangeSummaryDoc(root string, projectCfg projectConfig, latestSpec string, profile initProfile) string {
	latestSpec = normalizedLatestSpec(latestSpec)
	lines := changeSummaryHeaderLines(projectCfg, latestSpec)
	lines = append(lines, "")
	lines = append(lines, changeSummaryWorkflowDocsSection(profile)...)
	lines = append(lines, "")
	lines = append(lines, changeSummaryRefreshCommandsSection()...)
	if readinessLines := changeSummaryLatestReviewReadinessSection(root, latestSpec); len(readinessLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, readinessLines...)
	}
	if proofLines := changeSummaryLatestExecutionProofSection(root); len(proofLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, proofLines...)
	}
	return strings.Join(lines, "\n") + "\n"
}

func normalizedProjectType(projectCfg projectConfig) string {
	projectType := projectCfg.ProjectType
	if strings.TrimSpace(projectType) == "" {
		projectType = "existing"
	}
	return projectType
}

func normalizedLatestSpec(latestSpec string) string {
	if strings.TrimSpace(latestSpec) == "" {
		return "none"
	}
	return latestSpec
}

func changeSummaryHeaderLines(projectCfg projectConfig, latestSpec string) []string {
	return []string{
		"# Change Summary",
		"",
		fmt.Sprintf("Project: %s", projectCfg.Name),
		fmt.Sprintf("Project type: %s", normalizedProjectType(projectCfg)),
		fmt.Sprintf("Latest SPEC: %s", normalizedLatestSpec(latestSpec)),
	}
}

func changeSummaryWorkflowDocsSection(profile initProfile) []string {
	return []string{
		"## Workflow Docs Synced",
		"",
		"- README bundles and product docs describe when to use `namba update`, `namba regen`, `namba sync`, `namba pr`, and `namba land`.",
		"- Release docs describe `namba release` guardrails on a clean `main` branch plus optional `--push` behavior.",
		"- Run docs separate the default standalone flow, `namba run SPEC-XXX --solo`, `namba run SPEC-XXX --team`, and the worktree fan-out policy for `namba run SPEC-XXX --parallel`.",
		"- AGENTS and Codex docs define the Namba output contract plus the fallback validator script at `.namba/codex/validate-output-contract.py`.",
		"- SPEC packages can keep advisory plan-review artifacts under `.namba/specs/<SPEC>/reviews/` so product, engineering, and design review state stays visible before execution and PR handoff.",
		fmt.Sprintf("- Collaboration docs require one branch per SPEC/task from `%s`, PRs into `%s`, %s PR content, and explicit Codex review requests via `namba pr --review` or queue `--review` using `%s`.", branchBase(profile), prBaseBranch(profile), strings.ToLower(humanLanguageName(profile.PRLanguage)), codexReviewComment(profile)),
	}
}

func changeSummaryRefreshCommandsSection() []string {
	return []string{
		"## Refresh Commands",
		"",
		"- `namba update` self-updates the installed `namba` binary from GitHub Release assets.",
		"- `namba regen` regenerates `AGENTS.md`, repo-local skills and command-entry skills under `.agents/skills`, `.codex/agents/*.toml` custom agents, readable `.md` role-card mirrors, `.namba/codex/*`, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.",
		"- `namba sync` refreshes `.namba/project/*` docs, release notes/checklists, codemaps, and any README bundles enabled in `.namba/config/sections/docs.yaml`.",
		"- `namba pr` prepares the current branch for GitHub review by running sync and validation by default, then committing, pushing, and opening or reusing the PR. Add `--review` to request Codex review explicitly.",
		"- `namba land` optionally waits for checks, merges only when the PR is clean, and updates local `main` safely.",
	}
}

func changeSummaryLatestReviewReadinessSection(root, latestSpec string) []string {
	if !specReviewReadinessExists(root, latestSpec) {
		return nil
	}
	return []string{
		"## Latest Review Readiness",
		"",
		fmt.Sprintf("- Latest readiness artifact: `%s`", specReviewReadinessPath(latestSpec)),
		fmt.Sprintf("- Advisory summary: %s", specReadinessAdvisorySummary(root, latestSpec)),
	}
}

func buildPRChecklistDoc(root, latestSpec string, profile initProfile) string {
	lines := prChecklistHeaderLines()
	lines = append(lines, prChecklistCoreItems(profile)...)
	lines = append(lines, prChecklistLatestReviewReadinessItem(root, latestSpec)...)
	lines = append(lines, prChecklistLatestExecutionProofItem(root)...)
	return strings.Join(lines, "\n") + "\n"
}

func prChecklistHeaderLines() []string {
	return []string{
		"# PR Checklist",
		"",
	}
}

func prChecklistCoreItems(profile initProfile) []string {
	return []string{
		fmt.Sprintf("- [ ] Dedicated work branch created from `%s` for this SPEC/task", branchBase(profile)),
		fmt.Sprintf("- [ ] PR targets `%s`", prBaseBranch(profile)),
		fmt.Sprintf("- [ ] PR title and body are written in %s", humanLanguageName(profile.PRLanguage)),
		fmt.Sprintf("- [ ] If Codex review was explicitly requested with `--review`, `%s` request is present on GitHub", codexReviewComment(profile)),
		"- [ ] README / user-facing docs refreshed",
		"- [ ] `namba regen` rerun if template-generated Codex assets changed",
		"- [ ] `namba sync` artifacts refreshed",
		"- [ ] `namba pr` used for the GitHub review handoff when the branch is ready",
		"- [ ] SPEC artifacts reviewed",
		"- [ ] Validation commands passed",
		"- [ ] Diff reviewed",
	}
}

func prChecklistLatestReviewReadinessItem(root, latestSpec string) []string {
	if !specReviewReadinessExists(root, latestSpec) {
		return nil
	}
	return []string{fmt.Sprintf("- [ ] Latest SPEC review readiness checked: `%s`", specReviewReadinessPath(latestSpec))}
}

func buildReleaseNotesDoc(projectCfg projectConfig, latestSpec string, profile initProfile) string {
	lines := releaseNotesHeaderLines(projectCfg, latestSpec)
	lines = append(lines, "")
	lines = append(lines, releaseNotesWorkflowChangesSection(profile)...)
	lines = append(lines, "")
	lines = append(lines, releaseNotesGuardrailsSection()...)
	lines = append(lines, "")
	lines = append(lines, releaseNotesCommandsSection()...)
	lines = append(lines, "")
	lines = append(lines, releaseNotesExpectedAssetsSection()...)
	return strings.Join(lines, "\n") + "\n"
}

func releaseNotesHeaderLines(projectCfg projectConfig, latestSpec string) []string {
	return []string{
		"# Release Notes Draft",
		"",
		fmt.Sprintf("Project: %s", projectCfg.Name),
		fmt.Sprintf("Project type: %s", normalizedProjectType(projectCfg)),
		fmt.Sprintf("Reference SPEC: %s", normalizedLatestSpec(latestSpec)),
	}
}

func releaseNotesWorkflowChangesSection(profile initProfile) []string {
	return []string{
		"## Workflow Changes",
		"",
		"- `namba update` self-updates the installed `namba` binary from GitHub Release assets.",
		"- `namba regen` regenerates `AGENTS.md`, repo-local skills and command-entry skills under `.agents/skills`, `.codex/agents/*.toml` custom agents, readable `.md` role-card mirrors, and repo-local Codex config from `.namba/config/sections/*.yaml`.",
		"- `namba sync` refreshes README bundles, product docs, codemaps, change summary, PR checklist, and release docs.",
		"- `namba pr` prepares the current branch for GitHub review by syncing, validating, committing, pushing, and opening or reusing the PR. Add `--review` to request Codex review explicitly.",
		"- `namba land` optionally waits for checks, merges only when the PR is clean, and updates local `main` safely.",
		"- `$namba-review-resolve` handles the meaningful GitHub review loop thread-by-thread: classify unresolved review threads, make scoped fixes, reply with validation evidence, resolve only addressed threads, and request review again without duplicating the configured marker.",
		"- `$namba-release` is the Codex-facing NambaAI release workflow: generate commit-based release notes, write `.namba/releases/<version>.md`, validate, then hand off to the guarded `namba release --version <version> --push` path.",
		"- `namba run SPEC-XXX` keeps the standard standalone Codex flow; `--solo` and `--team` request single-subagent or multi-subagent workflows inside one workspace; `--parallel` still fans out into up to three git worktrees and merges only after every worker passes execution and validation.",
		fmt.Sprintf("- Active collaboration defaults: one branch per SPEC/task from `%s`, PRs into `%s`, %s PR content, and explicit Codex review requests via `namba pr --review` or queue `--review` using `%s`.", branchBase(profile), prBaseBranch(profile), strings.ToLower(humanLanguageName(profile.PRLanguage)), codexReviewComment(profile)),
	}
}

func releaseNotesGuardrailsSection() []string {
	return []string{
		"## Release Guardrails",
		"",
		"- `namba release` requires a git repository, the `main` branch, and a clean working tree.",
		"- Validators from `.namba/config/sections/quality.yaml` run before the release tag is created.",
		"- With no explicit version, `namba release` defaults to the next `patch` tag. Use `--bump minor|major` or `--version vX.Y.Z` when needed.",
		"- Release notes must exist before tagging. The NambaAI release workflow writes `.namba/releases/<version>.md`, and the GitHub Release workflow uses that file as the release body.",
		"- `namba release --push` pushes both `main` and the new tag to the selected remote.",
	}
}

func releaseNotesCommandsSection() []string {
	return []string{
		"## Release Commands",
		"",
		"```text",
		"namba sync",
		"namba pr \"release review\"",
		"namba land",
		"namba release --bump patch",
		"# or",
		"namba release --version vX.Y.Z --push",
		"```",
	}
}

func releaseNotesExpectedAssetsSection() []string {
	return []string{
		"## Expected Assets",
		"",
		"- `namba_Windows_x86.zip`",
		"- `namba_Windows_x86_64.zip`",
		"- `namba_Windows_arm64.zip`",
		"- `namba_Linux_x86_64.tar.gz`",
		"- `namba_Linux_arm64.tar.gz`",
		"- `namba_macOS_x86_64.tar.gz`",
		"- `namba_macOS_arm64.tar.gz`",
		"- `checksums.txt`",
	}
}

func buildReleaseChecklistDoc() string {
	lines := releaseChecklistHeaderLines()
	lines = append(lines, releaseChecklistItems()...)
	return strings.Join(lines, "\n") + "\n"
}

func releaseChecklistHeaderLines() []string {
	return []string{
		"# Release Checklist",
		"",
	}
}

func releaseChecklistItems() []string {
	return []string{
		"- [ ] `namba regen` rerun if template-generated Codex assets changed",
		"- [ ] `namba sync` artifacts refreshed",
		"- [ ] `namba pr` used for the GitHub review handoff when the branch is ready",
		"- [ ] README and `.namba/codex/README.md` reflect update, release, and parallel workflow behavior",
		"- [ ] Working tree is clean and the current branch is `main`",
		"- [ ] Validation commands passed",
		"- [ ] `namba release --version vX.Y.Z` or `namba release --bump patch` executed",
		"- [ ] If `--push` was not used, `main` and the release tag were pushed manually",
		"- [ ] GitHub Release workflow completed and published assets plus `checksums.txt`",
	}
}
