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
		"## ✅ Local Quality Gates",
		"## 🚢 Release Flow",
		"## 🧩 Command Skills In Codex",
		"## 🧱 Technical Snapshot",
		"`namba project`",
		"`namba plan`",
		"`namba harness \"description\"`",
		"`namba fix --command plan \"issue\"`",
		"`namba fix \"issue\"`",
		"swap `namba plan` for `namba harness \"description\"`",
		"`scripts/quality.sh`",
		"`toolchain go1.26.3`",
		"evidence schema validation",
		"report JSON validation",
		"`eval-scorecard.json`",
		"`schema-validation.txt`",
		"aggregate Go coverage threshold is 73.0%",
		"73.7% planning baseline",
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
	for _, bad := range []string{
		"| Plan harness work | `namba harness` /",
		"| Repair a bug | `namba fix`,",
		"`namba fix --command plan`, or",
	} {
		if strings.Contains(rootReadme, bad) {
			t.Fatalf("root README command chooser should only show runnable examples, found %q", bad)
		}
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
		"Do-Not Design Contract",
		"Do-Not Design Violation Check",
		"Generated Asset Evidence",
		"asset manifest",
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
			"`scripts/quality.sh`",
			"`toolchain go1.26.3`",
			"73.0%",
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

func TestGeneratedDocsUseVisualDocumentationGrammar(t *testing.T) {
	outputs := buildReadmeOutputs(projectConfig{Name: "NambaAI"}, initProfile{}, docsConfig{
		ManageReadme:        true,
		ReadmeProfile:       readmeProfileNambaCLI,
		DefaultLanguage:     "en",
		AdditionalLanguages: []string{"ko", "ja", "zh"},
		HeroImage:           "assets/images/namba-ai-hero.png",
	})

	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		root := outputs[readmePath(lang)]
		for _, want := range []string{
			renderGeneratedDocHeader(),
			"<img src=\"assets/images/namba-ai-hero.png\" alt=\"NambaAI\" width=\"100%\" />",
			"[![Release]",
			"[![CI]",
			"[![Security]",
			"[![License]",
			"[![Docs]",
			"| `namba project` |",
			"| `namba plan \"description\"` |",
			"`namba harness \"description\"`",
			"`namba fix --command plan",
			"`namba queue start SPEC-001..SPEC-003`",
			"`namba sync`",
			"`namba pr \"title\"`",
			"<details>",
			"<summary>",
			"</details>",
		} {
			assertContains(t, root, want, fmt.Sprintf("%s root README", lang))
		}
		for _, want := range []string{
			fmt.Sprintf("<a href=\"%s\"><img src=\"https://img.shields.io/static/v1?", guidePath("getting-started", lang)),
			fmt.Sprintf("<a href=\"%s\"><img src=\"https://img.shields.io/static/v1?", guidePath("workflow-guide", lang)),
			fmt.Sprintf("<a href=\"%s/releases/latest\"><img src=\"https://img.shields.io/static/v1?", nambaRepositoryURL),
			"<a href=\"SECURITY.md\"><img src=\"https://img.shields.io/static/v1?",
		} {
			assertContains(t, root, want, fmt.Sprintf("%s root README CTA row", lang))
		}

		firstTrust := strings.Index(root, "[![Release]")
		firstRun := strings.Index(root, "namba project")
		firstCommandChooser := strings.Index(root, readmeRootCommandChooserHeading(lang))
		firstReadNext := strings.Index(root, "## Read Next")
		if lang != "en" {
			firstReadNext = strings.Index(root, "## 다음에 읽을 문서")
			if firstReadNext < 0 {
				firstReadNext = strings.Index(root, "## 次に読む文書")
			}
			if firstReadNext < 0 {
				firstReadNext = strings.Index(root, "## 接下来阅读")
			}
		}
		if firstTrust < 0 || firstRun < 0 || firstCommandChooser < 0 || firstReadNext < 0 {
			t.Fatalf("%s root README missing expected orientation anchors", lang)
		}
		if !(firstTrust < firstRun && firstRun < firstCommandChooser && firstCommandChooser < firstReadNext) {
			t.Fatalf("%s root README orientation order is wrong: trust=%d run=%d chooser=%d next=%d", lang, firstTrust, firstRun, firstCommandChooser, firstReadNext)
		}

		for _, path := range []string{readmePath(lang), guidePath("getting-started", lang), guidePath("workflow-guide", lang)} {
			doc := outputs[path]
			for _, forbidden := range []string{"<script", "<iframe", "style=", "<button"} {
				if strings.Contains(strings.ToLower(doc), forbidden) {
					t.Fatalf("%s contains unsupported GitHub HTML %q: %q", path, forbidden, doc)
				}
			}
			for _, want := range []string{"<details>", "<summary>", "</details>"} {
				assertContains(t, doc, want, path)
			}
		}
	}
}

func TestGeneratedDocsIncludeGitHubSafeMermaidDiagrams(t *testing.T) {
	outputs := buildReadmeOutputs(projectConfig{Name: "NambaAI"}, initProfile{}, docsConfig{
		ManageReadme:        true,
		ReadmeProfile:       readmeProfileNambaCLI,
		DefaultLanguage:     "en",
		AdditionalLanguages: []string{"ko", "ja", "zh"},
	})
	repeatedOutputs := buildReadmeOutputs(projectConfig{Name: "NambaAI"}, initProfile{}, docsConfig{
		ManageReadme:        true,
		ReadmeProfile:       readmeProfileNambaCLI,
		DefaultLanguage:     "en",
		AdditionalLanguages: []string{"ko", "ja", "zh"},
	})
	for path, want := range outputs {
		if got := repeatedOutputs[path]; got != want {
			t.Fatalf("generated output is not deterministic for %s", path)
		}
	}

	commandFlowCommands := []string{
		"namba project",
		"namba plan",
		"namba harness",
		"namba fix",
		"namba run",
		"namba queue",
		"namba sync",
		"namba pr",
		"namba land",
	}
	lifecycleConcepts := []string{
		"SPEC",
		"validation",
		"docs sync",
		"PR",
		"blocked",
		"repair/retry",
	}

	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		root := outputs[readmePath(lang)]
		assertContains(t, root, localizedMermaidTitle(lang, "root-command-flow"), fmt.Sprintf("%s root README Mermaid title", lang))
		assertContains(t, root, "```mermaid", fmt.Sprintf("%s root README Mermaid fence", lang))
		for _, want := range commandFlowCommands {
			assertContains(t, root, want, fmt.Sprintf("%s root README command flow", lang))
		}

		gettingStarted := outputs[guidePath("getting-started", lang)]
		assertContains(t, gettingStarted, localizedMermaidTitle(lang, "getting-started-first-run"), fmt.Sprintf("%s getting-started Mermaid title", lang))
		assertContains(t, gettingStarted, "```mermaid", fmt.Sprintf("%s getting-started Mermaid fence", lang))
		for _, want := range []string{"install", "namba init .", "namba project", "namba plan", "namba run", "namba sync", "namba pr"} {
			assertContains(t, gettingStarted, want, fmt.Sprintf("%s getting-started first-run diagram", lang))
		}

		workflowGuide := outputs[guidePath("workflow-guide", lang)]
		assertContains(t, workflowGuide, localizedMermaidTitle(lang, "workflow-lifecycle"), fmt.Sprintf("%s workflow-guide Mermaid title", lang))
		assertContains(t, workflowGuide, "```mermaid", fmt.Sprintf("%s workflow-guide Mermaid fence", lang))
		for _, want := range lifecycleConcepts {
			assertContains(t, workflowGuide, want, fmt.Sprintf("%s workflow-guide lifecycle diagram", lang))
		}

		for _, doc := range []string{root, gettingStarted, workflowGuide} {
			assertContains(t, doc, "flowchart LR", fmt.Sprintf("%s Mermaid flowchart", lang))
			for _, forbidden := range []string{"<script", "<iframe", "style=", "<button"} {
				if strings.Contains(strings.ToLower(doc), forbidden) {
					t.Fatalf("%s generated diagram output contains unsupported GitHub HTML %q: %q", lang, forbidden, doc)
				}
			}
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
