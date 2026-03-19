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
	for _, want := range []string{"--solo", "--team", "--parallel", "## Command Skills In Codex", "## Skill To Command Mapping", "## Custom Agents In Codex", "`$namba-run`", "`$namba-sync`", "`$namba-pr`", "`$namba-regen`", "`namba-product-manager`", "`namba-mobile-engineer`", "`namba-designer`", "`namba-data-engineer`", "`namba-security-engineer`"} {
		if !strings.Contains(readme, want) {
			t.Fatalf("expected README to contain %q, got %q", want, readme)
		}
	}

	workflowGuide := mustReadFile(t, filepath.Join(tmp, "docs", "workflow-guide.md"))
	for _, want := range []string{"## Run modes", "## Role routing", "`namba run SPEC-XXX --solo`", "`namba run SPEC-XXX --team`", "`namba run SPEC-XXX --parallel`", "`namba-mobile-engineer`", "`namba-security-engineer`"} {
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
