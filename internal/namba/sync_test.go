package namba

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
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

func TestRefreshAllSpecReviewReadinessBatchesManifestWrites(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	for _, specID := range []string{"SPEC-001", "SPEC-002"} {
		reviewsDir := filepath.Join(tmp, ".namba", "specs", specID, "reviews")
		if err := os.MkdirAll(reviewsDir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", reviewsDir, err)
		}
		for rel, body := range specReviewOutputs(specID) {
			if strings.HasSuffix(rel, specReviewReadinessFileName) {
				continue
			}
			writeTestFile(t, filepath.Join(tmp, filepath.FromSlash(rel)), body)
		}
	}

	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	manifestWrites := 0
	app.writeManifestOverride = func(root string, manifest Manifest) error {
		if root != tmp {
			t.Fatalf("expected manifest write root %q, got %q", tmp, root)
		}
		manifestWrites++
		return nil
	}

	if err := app.refreshAllSpecReviewReadiness(tmp); err != nil {
		t.Fatalf("refreshAllSpecReviewReadiness failed: %v", err)
	}
	if manifestWrites != 1 {
		t.Fatalf("expected one manifest session for batched readiness refresh, got %d", manifestWrites)
	}

	for _, specID := range []string{"SPEC-001", "SPEC-002"} {
		readiness := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", specID, "reviews", "readiness.md"))
		if !strings.Contains(readiness, "Advisory status:") {
			t.Fatalf("expected refreshed readiness for %s, got %q", specID, readiness)
		}
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
	if !strings.Contains(outputs[filepath.ToSlash(filepath.Join(projectDir, "release-notes.md"))], "## Release Commands") || !strings.Contains(outputs[filepath.ToSlash(filepath.Join(projectDir, "release-notes.md"))], "`checksums.txt`") || !strings.Contains(outputs[filepath.ToSlash(filepath.Join(projectDir, "release-notes.md"))], "`$namba-release`") {
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

func TestWriteOutputsBatchesMultipleReadinessFilesInOneManifestSession(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	var (
		manifestWrites int
		manifest       Manifest
	)
	app.writeManifestOverride = func(root string, m Manifest) error {
		if root != tmp {
			t.Fatalf("expected manifest write root %q, got %q", tmp, root)
		}
		manifestWrites++
		manifest = m
		return nil
	}

	outputs := map[string]string{
		specReviewReadinessPath("SPEC-001"): "# Review Readiness\n\nSPEC: SPEC-001\n",
		specReviewReadinessPath("SPEC-002"): "# Review Readiness\n\nSPEC: SPEC-002\n",
	}
	report, err := app.writeOutputs(tmp, outputs)
	if err != nil {
		t.Fatalf("writeOutputs failed: %v", err)
	}
	if manifestWrites != 1 {
		t.Fatalf("expected one manifest session for batched readiness outputs, got %d", manifestWrites)
	}
	if len(report.ChangedPaths) != 2 {
		t.Fatalf("expected two changed readiness paths, got %+v", report.ChangedPaths)
	}
	wantEntries := map[string]bool{
		specReviewReadinessPath("SPEC-001"): true,
		specReviewReadinessPath("SPEC-002"): true,
	}
	if len(manifest.Entries) != len(wantEntries) {
		t.Fatalf("expected %d manifest entries, got %+v", len(wantEntries), manifest.Entries)
	}
	for _, entry := range manifest.Entries {
		if !wantEntries[entry.Path] {
			t.Fatalf("unexpected manifest entry path %q, got %+v", entry.Path, manifest.Entries)
		}
		if entry.Owner != manifestOwnerManaged {
			t.Fatalf("expected managed owner on manifest entry, got %+v", entry)
		}
		delete(wantEntries, entry.Path)
	}
	if len(wantEntries) != 0 {
		t.Fatalf("missing manifest entries for batched readiness outputs: %+v", wantEntries)
	}
}

func TestWriteOutputsRecoversFromMalformedManifest(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := app.mkdirAll(filepath.Join(tmp, nambaDir), 0o755); err != nil {
		t.Fatalf("mkdir .namba: %v", err)
	}
	if err := app.writeFile(filepath.Join(tmp, manifestPath), []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write malformed manifest: %v", err)
	}

	outputs := map[string]string{
		filepath.ToSlash(filepath.Join(projectDir, "change-summary.md")): "summary",
	}
	report, err := app.writeOutputs(tmp, outputs)
	if err != nil {
		t.Fatalf("writeOutputs failed: %v", err)
	}
	if got, want := report.ChangedPaths, []string{filepath.ToSlash(filepath.Join(projectDir, "change-summary.md"))}; !reflect.DeepEqual(got, want) {
		t.Fatalf("changed paths = %#v, want %#v", got, want)
	}
	if got := mustReadFile(t, filepath.Join(tmp, ".namba", "project", "change-summary.md")); got != "summary" {
		t.Fatalf("expected materialized output after manifest recovery, got %q", got)
	}

	manifest, err := app.readManifest(tmp)
	if err != nil {
		t.Fatalf("read healed manifest: %v", err)
	}
	if got, want := len(manifest.Entries), 1; got != want {
		t.Fatalf("manifest entry count = %d, want %d; manifest=%+v", got, want, manifest)
	}
	if manifest.Entries[0].Path != filepath.ToSlash(filepath.Join(projectDir, "change-summary.md")) {
		t.Fatalf("unexpected healed manifest entry: %+v", manifest.Entries[0])
	}
}

func TestWriteSyncProjectSupportDocsUsesOneManifestSession(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	syncCtx, err := app.loadSyncContext(tmp)
	if err != nil {
		t.Fatalf("loadSyncContext failed: %v", err)
	}

	manifestWrites := 0
	app.writeManifestOverride = func(root string, manifest Manifest) error {
		if root != tmp {
			t.Fatalf("expected manifest write root %q, got %q", tmp, root)
		}
		manifestWrites++
		return nil
	}

	outputs := buildSyncProjectSupportOutputs(syncCtx)
	if len(outputs) != 4 {
		t.Fatalf("expected four staged support outputs, got %d: %+v", len(outputs), outputs)
	}
	if err := app.writeSyncProjectSupportDocs(syncCtx); err != nil {
		t.Fatalf("writeSyncProjectSupportDocs failed: %v", err)
	}
	if manifestWrites != 1 {
		t.Fatalf("expected one manifest session for sync support docs, got %d", manifestWrites)
	}
}

func TestSyncDoesNotMutateOutputsWhenReadinessBatchFails(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	if err := app.Run(context.Background(), []string{"sync"}); err != nil {
		t.Fatalf("baseline sync failed: %v", err)
	}

	structurePath := filepath.Join(tmp, ".namba", "project", "structure.md")
	manifestPath := filepath.Join(tmp, ".namba", "manifest.json")
	beforeStructure := mustReadFile(t, structurePath)
	beforeManifest := mustReadFile(t, manifestPath)

	specsDirPath := filepath.Join(tmp, ".namba", "specs")
	if err := os.Rename(specsDirPath, specsDirPath+"-backup"); err != nil {
		t.Fatalf("rename specs dir: %v", err)
	}
	if err := os.WriteFile(specsDirPath, []byte("broken"), 0o644); err != nil {
		t.Fatalf("write broken specs path: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "internal", "spec034"), 0o755); err != nil {
		t.Fatalf("mkdir new source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "internal", "spec034", "service.go"), []byte("package spec034\n"), 0o644); err != nil {
		t.Fatalf("write new source file: %v", err)
	}

	if err := app.Run(context.Background(), []string{"sync"}); err == nil {
		t.Fatal("expected sync to fail when .namba/specs is not a directory")
	}

	if got := mustReadFile(t, structurePath); got != beforeStructure {
		t.Fatalf("expected structure doc to stay unchanged on readiness batch failure")
	}
	if got := mustReadFile(t, manifestPath); got != beforeManifest {
		t.Fatalf("expected manifest to stay unchanged on readiness batch failure")
	}
}

func TestSyncSupportDocsDoNotReferenceRemovedLatestReadiness(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	if err := app.Run(context.Background(), []string{"sync"}); err != nil {
		t.Fatalf("baseline sync failed: %v", err)
	}

	reviewsDir := filepath.Join(tmp, ".namba", "specs", "SPEC-001", "reviews")
	if err := os.RemoveAll(reviewsDir); err != nil {
		t.Fatalf("remove reviews dir: %v", err)
	}

	if err := app.Run(context.Background(), []string{"sync"}); err != nil {
		t.Fatalf("sync after review removal failed: %v", err)
	}

	readinessPath := filepath.Join(tmp, ".namba", "specs", "SPEC-001", "reviews", "readiness.md")
	if _, err := os.Stat(readinessPath); !os.IsNotExist(err) {
		t.Fatalf("expected readiness artifact to be removed, stat err=%v", err)
	}

	changeSummary := mustReadFile(t, filepath.Join(tmp, ".namba", "project", "change-summary.md"))
	if strings.Contains(changeSummary, "## Latest Review Readiness") || strings.Contains(changeSummary, specReviewReadinessPath("SPEC-001")) {
		t.Fatalf("expected change summary to drop removed readiness reference, got %q", changeSummary)
	}

	prChecklist := mustReadFile(t, filepath.Join(tmp, ".namba", "project", "pr-checklist.md"))
	if strings.Contains(prChecklist, "Latest SPEC review readiness checked") || strings.Contains(prChecklist, specReviewReadinessPath("SPEC-001")) {
		t.Fatalf("expected pr checklist to drop removed readiness reference, got %q", prChecklist)
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

func TestReleaseNotesGuardrailsIncludeGeneratedNotesHandoff(t *testing.T) {
	got := strings.Join(releaseNotesGuardrailsSection(), "\n")

	for _, want := range []string{
		"## Release Guardrails",
		"`.namba/releases/<version>.md`",
		"GitHub Release workflow uses that file as the release body",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected release guardrails to contain %q, got %q", want, got)
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
