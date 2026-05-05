package namba

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReadmeRendererIncludesOnboardingAnchorsForRepoConfig(t *testing.T) {
	root := repoRoot(t)
	projectCfg, docsCfg, profile := loadRepoDocsConfig(t, root)

	outputs := buildReadmeOutputs(projectCfg, profile, docsCfg)
	if got, want := len(outputs), 12; got != want {
		t.Fatalf("buildReadmeOutputs() produced %d outputs, want %d", got, want)
	}

	rootReadme := outputs[readmePath("en")]
	for _, want := range []string{
		"## 🧭 Which Command Should I Use?",
		"## 🧰 What You Can Do With NambaAI",
		"## 🚀 Quick Start",
		"## 🚢 Release Flow",
		"## 🧩 Command Skills In Codex",
		"## 🧱 Technical Snapshot",
		"`namba project`",
		"`namba plan`",
		"`namba harness`",
		"`namba fix --command plan`",
		"`namba fix`",
		"swap `namba plan` for `namba harness \"description\"`",
		"`$namba`: general router",
		"`$namba-coach`",
		"`$namba-help`",
		"`$namba-create`: use when you want a preview-first flow",
		"`$namba-project`: use when you need project docs and codemaps refreshed",
		"`$namba-plan`: use when you want to create the next feature SPEC package",
		"`$namba-plan-review`: use when you want one Codex entry point",
		"`$namba-harness`: use when you want a harness-oriented SPEC package",
		"`$namba-fix`: use when you need direct repair in the current workspace",
		"`$namba-run`: use when you want to execute an existing SPEC package",
		"`$namba-queue`: use when you want to process multiple existing SPEC packages",
		"## 🪝 Hook Runtime",
		"`.namba/hooks.toml`",
		"`continue_on_failure = false`",
		"`before_validation`",
		"`$namba-sync`",
		"`$namba-pr`",
		"`$namba-land`",
		"`$namba-review-resolve`",
		"`$namba-release`",
		"`$namba-regen`",
		"`$namba-update`",
		"Emoji density rule",
		"`namba-frontend-architect`",
		"`Redesign this landing page hero so it stops looking generic`",
	} {
		assertContains(t, rootReadme, want, "root README")
	}
	if strings.Contains(rootReadme, "temperature and undertone discipline") {
		t.Fatalf("root README should stay lightweight and not inline the full designer manifesto: %q", rootReadme)
	}

	workflowGuide := outputs[guidePath("workflow-guide", "en")]
	for _, want := range []string{
		"## `update`, `regen`, `sync`, `pr`, and `land` are different commands",
		"## Planning commands",
		"## `namba run` modes",
		"## Role routing",
		"## Review readiness",
		"## PR and merge flow",
		"## 🚢 Release Flow",
		"`namba project`",
		"`$namba-coach`",
		"`$namba-create`: use the preview-first creation flow",
		"`namba harness \"description\"`",
		"`namba fix --command run \"issue description\"`",
		"`namba plan`, `namba harness`, and `namba fix --command plan`",
		"`namba regen`",
		"`namba-frontend-implementer`",
		"`namba-frontend-architect`",
		"`namba-mobile-engineer`",
		"`namba-designer`",
		"`namba-backend-implementer`",
		"`namba-data-engineer`",
		"`namba-security-engineer`",
		"`namba-devops-engineer`",
		"`namba-reviewer`",
		"`$namba-help`",
		"`$namba-plan-review`",
		"`$namba-release`",
		"`Redesign the hero so it stops looking generic`",
		"`Plan the component boundaries for this dashboard`",
		"`Implement the approved dashboard filters`",
	} {
		assertContains(t, workflowGuide, want, "workflow guide")
	}
	for _, unwanted := range []string{"temperature and undertone discipline", "washed-out minimalism"} {
		if strings.Contains(workflowGuide, unwanted) {
			t.Fatalf("workflow guide should keep the role split lightweight and not contain %q: %q", unwanted, workflowGuide)
		}
	}

	for _, lang := range []string{"ko", "ja", "zh"} {
		readme := outputs[readmePath(lang)]
		for _, want := range []string{
			"`$namba-help`",
			"`$namba-create`",
			"`$namba-run`",
			"`$namba-queue`",
			"`$namba-harness`",
			"`$namba-plan-review`",
			"`$namba-plan-pm-review`",
			"`$namba-plan-eng-review`",
			"`$namba-plan-design-review`",
			"`namba harness \"description\"`",
			"`namba fix --command plan \"issue description\"`",
			"`namba sync`",
			"`namba pr`",
			"`namba land`",
			"## 🪝 Hook Runtime",
			"`$namba-review-resolve`",
			"`$namba-release`",
			"`.namba/hooks.toml`",
		} {
			assertContains(t, readme, want, fmt.Sprintf("%s README", lang))
		}

		guide := outputs[guidePath("workflow-guide", lang)]
		for _, want := range []string{
			"`namba project`",
			"`$namba-create`",
			"`namba harness \"description\"`",
			"`namba run SPEC-XXX --team`",
			"`namba run SPEC-XXX --parallel`",
			"`namba fix --command plan \"issue description\"`",
			"`namba fix --command run \"issue description\"`",
			"`namba-frontend-architect`",
			"`namba-reviewer`",
			"`$namba-plan-review`",
			"`namba pr`",
			"`namba land`",
		} {
			assertContains(t, guide, want, fmt.Sprintf("%s workflow guide", lang))
		}
	}
}

func TestSyncedReadmeOutputsMatchRendererForRepoConfig(t *testing.T) {
	root := repoRoot(t)
	projectCfg, docsCfg, profile := loadRepoDocsConfig(t, root)
	expected := buildReadmeOutputs(projectCfg, profile, docsCfg)

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	if err := copyDirContents(filepath.Join(root, ".namba", "config", "sections"), filepath.Join(tmp, ".namba", "config", "sections")); err != nil {
		t.Fatalf("copy config sections: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"sync"}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	for path, want := range expected {
		got := mustReadFile(t, filepath.Join(tmp, path))
		if got != want {
			t.Fatalf("synced output mismatch for %s", path)
		}
	}
}

func TestCheckedInRepoDocsMatchRendererForRepoConfig(t *testing.T) {
	root := repoRoot(t)
	projectCfg, docsCfg, profile := loadRepoDocsConfig(t, root)
	expected := buildReadmeOutputs(projectCfg, profile, docsCfg)

	for path, want := range expected {
		got := mustReadFile(t, filepath.Join(root, path))
		if got != want {
			t.Fatalf("checked-in generated doc drift for %s; run `namba sync`", path)
		}
	}
}

func loadRepoDocsConfig(t *testing.T, root string) (projectConfig, docsConfig, initProfile) {
	t.Helper()

	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	projectCfg, err := app.loadProjectConfig(root)
	if err != nil {
		t.Fatalf("load project config: %v", err)
	}
	docsCfg, err := app.loadDocsConfig(root)
	if err != nil {
		t.Fatalf("load docs config: %v", err)
	}
	profile, err := app.loadInitProfileFromConfig(root)
	if err != nil {
		t.Fatalf("load init profile: %v", err)
	}
	return projectCfg, docsCfg, profile
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func copyDirContents(srcDir, dstDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dstPath, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func assertContains(t *testing.T, haystack, needle, label string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("%s missing %q", label, needle)
	}
}
