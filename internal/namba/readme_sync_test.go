package namba

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSyncWritesRunModeDocs(t *testing.T) {
	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"sync"}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	readme := mustReadFile(t, filepath.Join(tmp, "README.md"))
	for _, want := range []string{"--solo", "--team", "--parallel", "## Command Skills In Codex", "## Skill To Command Mapping", "## Custom Agents In Codex", "`$namba-run`", "`$namba-sync`", "`$namba-pr`", "`$namba-regen`", "`$namba-plan-pm-review`", "`$namba-plan-eng-review`", "`$namba-plan-design-review`", "`namba-product-manager`", "`namba-mobile-engineer`", "`namba-designer`", "`namba-data-engineer`", "`namba-security-engineer`"} {
		if !strings.Contains(readme, want) {
			t.Fatalf("expected README to contain %q, got %q", want, readme)
		}
	}

	workflowGuide := mustReadFile(t, filepath.Join(tmp, "docs", "workflow-guide.md"))
	for _, want := range []string{"## Run modes", "## Role routing", "## Review readiness", "`namba run SPEC-XXX --solo`", "`namba run SPEC-XXX --team`", "`namba run SPEC-XXX --parallel`", "`namba-mobile-engineer`", "`namba-security-engineer`", "`$namba-plan-pm-review`"} {
		if !strings.Contains(workflowGuide, want) {
			t.Fatalf("expected workflow guide to contain %q, got %q", want, workflowGuide)
		}
	}
}

func TestRenderNambaCLIWorkflowGuideIncludesRoleRouting(t *testing.T) {
	guide := renderReadmeGuide("en", "workflow-guide", projectConfig{}, initProfile{}, docsConfig{ReadmeProfile: readmeProfileNambaCLI})
	for _, want := range []string{"## Role routing", "`namba run SPEC-XXX --team`", "`namba-mobile-engineer`", "`namba-security-engineer`", "`namba-reviewer`"} {
		if !strings.Contains(guide, want) {
			t.Fatalf("expected namba-cli workflow guide to contain %q, got %q", want, guide)
		}
	}
}

func TestBuildReadmeOutputsForNambaCLIIncludesLocalizedLifecycleDocs(t *testing.T) {
	outputs := buildReadmeOutputs(projectConfig{}, initProfile{}, docsConfig{
		ManageReadme:        true,
		ReadmeProfile:       readmeProfileNambaCLI,
		DefaultLanguage:     "en",
		AdditionalLanguages: []string{"ko", "ja", "zh"},
	})

	if got, want := len(outputs), 12; got != want {
		t.Fatalf("buildReadmeOutputs() produced %d outputs, want %d", got, want)
	}

	cases := []struct {
		lang                    string
		rootLifecycleHeading    string
		gettingStartedUpdate    string
		gettingStartedUninstall string
		workflowModesHeading    string
		workflowReviewHeading   string
	}{
		{
			lang:                    "en",
			rootLifecycleHeading:    "## Install, Update, and Uninstall",
			gettingStartedUpdate:    "## 2. Update",
			gettingStartedUninstall: "## 3. Uninstall",
			workflowModesHeading:    "## `namba run` modes",
			workflowReviewHeading:   "## Review readiness",
		},
		{
			lang:                    "ko",
			rootLifecycleHeading:    "## 설치, 업데이트, 제거",
			gettingStartedUpdate:    "## 2. 업데이트",
			gettingStartedUninstall: "## 3. 제거",
			workflowModesHeading:    "## `namba run` 모드",
			workflowReviewHeading:   "## 리뷰 준비도",
		},
		{
			lang:                    "ja",
			rootLifecycleHeading:    "## インストール、更新、アンインストール",
			gettingStartedUpdate:    "## 2. 更新",
			gettingStartedUninstall: "## 3. アンインストール",
			workflowModesHeading:    "## `namba run` モード",
			workflowReviewHeading:   "## レビュー準備度",
		},
		{
			lang:                    "zh",
			rootLifecycleHeading:    "## 安装、更新与卸载",
			gettingStartedUpdate:    "## 2. 更新",
			gettingStartedUninstall: "## 3. 卸载",
			workflowModesHeading:    "## `namba run` 模式",
			workflowReviewHeading:   "## 评审就绪度",
		},
	}

	for _, tc := range cases {
		root := outputs[readmePath(tc.lang)]
		for _, want := range []string{
			tc.rootLifecycleHeading,
			"`$namba-run`",
			"`$namba-update`",
			"`namba-mobile-engineer`",
			"`namba-security-engineer`",
			"`namba update --version vX.Y.Z`",
			"namba pr",
			"namba land",
			nambaWindowsBinaryPath,
			nambaUnixBinaryPath,
		} {
			if !strings.Contains(root, want) {
				t.Fatalf("%s README missing %q: %q", tc.lang, want, root)
			}
		}

		gettingStarted := outputs[guidePath("getting-started", tc.lang)]
		for _, want := range []string{
			tc.gettingStartedUpdate,
			tc.gettingStartedUninstall,
			"`namba update`",
			"`namba update --version vX.Y.Z`",
			"`NAMBA_INSTALL_DIR`",
			"namba pr",
			"namba land",
			nambaWindowsBinaryPath,
			nambaUnixBinaryPath,
		} {
			if !strings.Contains(gettingStarted, want) {
				t.Fatalf("%s getting-started guide missing %q: %q", tc.lang, want, gettingStarted)
			}
		}

		workflowGuide := outputs[guidePath("workflow-guide", tc.lang)]
		for _, want := range []string{
			tc.workflowModesHeading,
			tc.workflowReviewHeading,
			"`namba run SPEC-XXX --team`",
			"`namba-reviewer`",
			"`$namba-plan-eng-review`",
			"`namba pr`",
			"`namba land`",
		} {
			if !strings.Contains(workflowGuide, want) {
				t.Fatalf("%s workflow guide missing %q: %q", tc.lang, want, workflowGuide)
			}
		}
	}
}
