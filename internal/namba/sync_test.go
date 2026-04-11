package namba

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadSyncContextCollectsProfileAndLatestSpec(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	syncCtx, err := app.loadSyncContext(tmp)
	if err != nil {
		t.Fatalf("loadSyncContext failed: %v", err)
	}

	if syncCtx.Root != tmp || syncCtx.LatestSpec != "SPEC-001" {
		t.Fatalf("unexpected sync context identity: %+v", syncCtx)
	}
	if syncCtx.ProjectCfg.Name == "" || syncCtx.Profile.ProjectName == "" || syncCtx.DocsCfg.ReadmeProfile == "" {
		t.Fatalf("expected sync context to load project/profile/docs config, got %+v", syncCtx)
	}
}

func TestMaterializeSyncReadmeWritesManagedDocs(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	syncCtx, err := app.loadSyncContext(tmp)
	if err != nil {
		t.Fatalf("loadSyncContext failed: %v", err)
	}
	if err := app.materializeSyncReadme(syncCtx); err != nil {
		t.Fatalf("materializeSyncReadme failed: %v", err)
	}

	readme := mustReadFile(t, filepath.Join(tmp, "README.md"))
	if !strings.Contains(readme, "`$namba-run`") || !strings.Contains(readme, "`namba fix --command plan \"issue description\"`") {
		t.Fatalf("expected managed README content, got %q", readme)
	}
}

func TestRefreshSyncProjectArtifactsRefreshesAnalysisAndReadiness(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	syncCtx, err := app.loadSyncContext(tmp)
	if err != nil {
		t.Fatalf("loadSyncContext failed: %v", err)
	}
	if err := app.refreshSyncProjectArtifacts(context.Background(), syncCtx); err != nil {
		t.Fatalf("refreshSyncProjectArtifacts failed: %v", err)
	}

	structure := mustReadFile(t, filepath.Join(tmp, ".namba", "project", "structure.md"))
	if !strings.Contains(structure, "# Structure") {
		t.Fatalf("expected structure doc after project refresh, got %q", structure)
	}
	readiness := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "reviews", "readiness.md"))
	if !strings.Contains(readiness, "Advisory status:") {
		t.Fatalf("expected readiness summary after refresh, got %q", readiness)
	}
}

func TestWriteSyncProjectSupportDocsWritesProjectSummaries(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	syncCtx, err := app.loadSyncContext(tmp)
	if err != nil {
		t.Fatalf("loadSyncContext failed: %v", err)
	}
	if err := app.writeSyncProjectSupportDocs(syncCtx); err != nil {
		t.Fatalf("writeSyncProjectSupportDocs failed: %v", err)
	}

	changeSummary := mustReadFile(t, filepath.Join(tmp, ".namba", "project", "change-summary.md"))
	if !strings.Contains(changeSummary, "SPEC-001") {
		t.Fatalf("expected change summary to mention latest spec, got %q", changeSummary)
	}
	releaseChecklist := mustReadFile(t, filepath.Join(tmp, ".namba", "project", "release-checklist.md"))
	if !strings.Contains(releaseChecklist, "# Release Checklist") {
		t.Fatalf("expected release checklist output, got %q", releaseChecklist)
	}
}

func TestBuildSyncProjectSupportOutputsIncludesManagedDocSet(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	syncCtx, err := app.loadSyncContext(tmp)
	if err != nil {
		t.Fatalf("loadSyncContext failed: %v", err)
	}

	outputs := buildSyncProjectSupportOutputs(syncCtx)
	expectedPaths := []string{
		filepath.ToSlash(filepath.Join(projectDir, "change-summary.md")),
		filepath.ToSlash(filepath.Join(projectDir, "pr-checklist.md")),
		filepath.ToSlash(filepath.Join(projectDir, "release-notes.md")),
		filepath.ToSlash(filepath.Join(projectDir, "release-checklist.md")),
	}
	if len(outputs) != len(expectedPaths) {
		t.Fatalf("expected %d sync support outputs, got %d: %+v", len(expectedPaths), len(outputs), outputs)
	}
	for _, path := range expectedPaths {
		if _, ok := outputs[path]; !ok {
			t.Fatalf("expected sync support outputs to include %s, got %+v", path, outputs)
		}
	}
	if !strings.Contains(outputs[filepath.ToSlash(filepath.Join(projectDir, "change-summary.md"))], "SPEC-001") {
		t.Fatalf("expected change-summary output to mention latest spec, got %q", outputs[filepath.ToSlash(filepath.Join(projectDir, "change-summary.md"))])
	}
	if !strings.Contains(outputs[filepath.ToSlash(filepath.Join(projectDir, "pr-checklist.md"))], "`@codex review` review request is present on GitHub") {
		t.Fatalf("expected pr-checklist output to include Codex review marker, got %q", outputs[filepath.ToSlash(filepath.Join(projectDir, "pr-checklist.md"))])
	}
	if !strings.Contains(outputs[filepath.ToSlash(filepath.Join(projectDir, "release-notes.md"))], "## Release Commands") || !strings.Contains(outputs[filepath.ToSlash(filepath.Join(projectDir, "release-notes.md"))], "`checksums.txt`") {
		t.Fatalf("expected release-notes output to include command and asset sections, got %q", outputs[filepath.ToSlash(filepath.Join(projectDir, "release-notes.md"))])
	}
	if !strings.Contains(outputs[filepath.ToSlash(filepath.Join(projectDir, "pr-checklist.md"))], "Latest SPEC review readiness checked") {
		t.Fatalf("expected pr-checklist output to include readiness item, got %q", outputs[filepath.ToSlash(filepath.Join(projectDir, "pr-checklist.md"))])
	}
	if !strings.Contains(outputs[filepath.ToSlash(filepath.Join(projectDir, "release-checklist.md"))], "# Release Checklist") {
		t.Fatalf("expected release-checklist output content, got %q", outputs[filepath.ToSlash(filepath.Join(projectDir, "release-checklist.md"))])
	}
	if !strings.Contains(outputs[filepath.ToSlash(filepath.Join(projectDir, "release-checklist.md"))], "`checksums.txt`") {
		t.Fatalf("expected release-checklist output to include published asset checklist, got %q", outputs[filepath.ToSlash(filepath.Join(projectDir, "release-checklist.md"))])
	}
}

func TestMaterializeSyncProjectSupportOutputsWritesProvidedDocs(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	outputs := map[string]string{
		filepath.ToSlash(filepath.Join(projectDir, "change-summary.md")):    "summary",
		filepath.ToSlash(filepath.Join(projectDir, "release-checklist.md")): "checklist",
	}
	if err := app.materializeSyncProjectSupportOutputs(tmp, outputs); err != nil {
		t.Fatalf("materializeSyncProjectSupportOutputs failed: %v", err)
	}

	if got := mustReadFile(t, filepath.Join(tmp, ".namba", "project", "change-summary.md")); got != "summary" {
		t.Fatalf("expected materialized change summary, got %q", got)
	}
	if got := mustReadFile(t, filepath.Join(tmp, ".namba", "project", "release-checklist.md")); got != "checklist" {
		t.Fatalf("expected materialized release checklist, got %q", got)
	}
}

func TestChangeSummaryHeaderLinesUseFallbacks(t *testing.T) {
	lines := changeSummaryHeaderLines(projectConfig{Name: "demo"}, "")
	got := strings.Join(lines, "\n")

	for _, want := range []string{
		"# Change Summary",
		"Project: demo",
		"Project type: existing",
		"Latest SPEC: none",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected change-summary header to contain %q, got %q", want, got)
		}
	}
}

func TestChangeSummaryLatestReviewReadinessSectionReflectsArtifact(t *testing.T) {
	tmp, _, restore := prepareExecutionProject(t)
	defer restore()

	lines := changeSummaryLatestReviewReadinessSection(tmp, "SPEC-001")
	got := strings.Join(lines, "\n")
	if !strings.Contains(got, "## Latest Review Readiness") || !strings.Contains(got, specReviewReadinessPath("SPEC-001")) {
		t.Fatalf("expected readiness section to mention SPEC-001 artifact, got %q", got)
	}
	if !strings.Contains(got, "Advisory summary:") {
		t.Fatalf("expected readiness section to include advisory summary, got %q", got)
	}
}

func TestReleaseNotesHeaderLinesUseFallbacks(t *testing.T) {
	lines := releaseNotesHeaderLines(projectConfig{Name: "demo"}, "")
	got := strings.Join(lines, "\n")

	for _, want := range []string{
		"# Release Notes Draft",
		"Project: demo",
		"Project type: existing",
		"Reference SPEC: none",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected release-notes header to contain %q, got %q", want, got)
		}
	}
}

func TestReleaseNotesExpectedAssetsSectionIncludesChecksums(t *testing.T) {
	got := strings.Join(releaseNotesExpectedAssetsSection(), "\n")

	for _, want := range []string{
		"## Expected Assets",
		"`namba_Windows_x86_64.zip`",
		"`namba_macOS_arm64.tar.gz`",
		"`checksums.txt`",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected expected-assets section to contain %q, got %q", want, got)
		}
	}
}

func TestPRChecklistCoreItemsReflectProfile(t *testing.T) {
	got := strings.Join(prChecklistCoreItems(initProfile{
		BranchBase:         "main",
		PRBaseBranch:       "main",
		PRLanguage:         "ko",
		CodexReviewComment: "@codex review",
	}), "\n")

	for _, want := range []string{
		"- [ ] Dedicated work branch created from `main` for this SPEC/task",
		"- [ ] PR targets `main`",
		"- [ ] PR title and body are written in Korean",
		"- [ ] `@codex review` review request is present on GitHub",
		"- [ ] `namba sync` artifacts refreshed",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected PR checklist core items to contain %q, got %q", want, got)
		}
	}
}

func TestPRChecklistLatestReviewReadinessItemReflectsArtifact(t *testing.T) {
	tmp, _, restore := prepareExecutionProject(t)
	defer restore()

	lines := prChecklistLatestReviewReadinessItem(tmp, "SPEC-001")
	got := strings.Join(lines, "\n")
	if !strings.Contains(got, "Latest SPEC review readiness checked") || !strings.Contains(got, specReviewReadinessPath("SPEC-001")) {
		t.Fatalf("expected PR checklist readiness item to mention SPEC-001 readiness, got %q", got)
	}
}

func TestReleaseChecklistItemsIncludeReleaseFlow(t *testing.T) {
	got := strings.Join(releaseChecklistItems(), "\n")

	for _, want := range []string{
		"- [ ] `namba pr` used for the GitHub review handoff when the branch is ready",
		"- [ ] Working tree is clean and the current branch is `main`",
		"- [ ] `namba release --version vX.Y.Z` or `namba release --bump patch` executed",
		"- [ ] GitHub Release workflow completed and published assets plus `checksums.txt`",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected release checklist items to contain %q, got %q", want, got)
		}
	}
}
