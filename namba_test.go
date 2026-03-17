package namba_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nam-Cheol/namba-ai/internal/namba"
)

func TestInitCreatesScaffold(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := app.Run(context.Background(), []string{"init", tmp}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	mustExist(t, filepath.Join(tmp, "AGENTS.md"))
	mustExist(t, filepath.Join(tmp, ".agents", "skills", "namba", "SKILL.md"))
	mustExist(t, filepath.Join(tmp, ".agents", "skills", "namba-foundation-core", "SKILL.md"))
	mustExist(t, filepath.Join(tmp, ".agents", "skills", "namba-workflow-init", "SKILL.md"))
	mustExist(t, filepath.Join(tmp, ".codex", "config.toml"))
	mustExist(t, filepath.Join(tmp, ".codex", "agents", "namba-planner.md"))
	mustExist(t, filepath.Join(tmp, ".codex", "agents", "namba-implementer.md"))
	mustExist(t, filepath.Join(tmp, ".codex", "agents", "namba-reviewer.md"))
	mustExist(t, filepath.Join(tmp, ".namba", "config", "sections", "project.yaml"))
	mustExist(t, filepath.Join(tmp, ".namba", "config", "sections", "language.yaml"))
	mustExist(t, filepath.Join(tmp, ".namba", "config", "sections", "git-strategy.yaml"))
	mustExist(t, filepath.Join(tmp, ".namba", "config", "sections", "codex.yaml"))
	mustExist(t, filepath.Join(tmp, ".namba", "codex", "claude-codex-mapping.md"))
	mustExist(t, filepath.Join(tmp, ".namba", "manifest.json"))
	mustNotExist(t, filepath.Join(tmp, ".codex", "skills", "namba", "SKILL.md"))

	codexProfile := mustRead(t, filepath.Join(tmp, ".namba", "config", "sections", "codex.yaml"))
	if !strings.Contains(codexProfile, "compat_skills_path: \n") {
		t.Fatalf("expected empty compat_skills_path for new project, got: %s", codexProfile)
	}
}

func TestInitCreatesCompatibilityMirrorForExistingProject(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "README.md"), []byte("# existing"), 0o644); err != nil {
		t.Fatalf("seed existing repo file: %v", err)
	}
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := app.Run(context.Background(), []string{"init", tmp}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	mustExist(t, filepath.Join(tmp, ".codex", "skills", "namba", "SKILL.md"))
}

func TestInitSupportsCodexProfileFlags(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	args := []string{
		"init",
		tmp,
		"--yes",
		"--name", "example",
		"--mode", "ddd",
		"--language", "go",
		"--framework", "cobra",
		"--conversation-language", "ko",
		"--documentation-language", "ko",
		"--comment-language", "en",
		"--git-mode", "team",
		"--git-provider", "github",
		"--git-username", "alice",
		"--agent-mode", "multi",
		"--statusline", "off",
		"--user-name", "Alice",
	}

	if err := app.Run(context.Background(), args); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	project := mustRead(t, filepath.Join(tmp, ".namba", "config", "sections", "project.yaml"))
	if !strings.Contains(project, "name: example") || !strings.Contains(project, "framework: cobra") {
		t.Fatalf("unexpected project config: %s", project)
	}

	language := mustRead(t, filepath.Join(tmp, ".namba", "config", "sections", "language.yaml"))
	if !strings.Contains(language, "conversation_language: ko") || !strings.Contains(language, "comment_language: en") {
		t.Fatalf("unexpected language config: %s", language)
	}

	gitStrategy := mustRead(t, filepath.Join(tmp, ".namba", "config", "sections", "git-strategy.yaml"))
	if !strings.Contains(gitStrategy, "git_mode: team") || !strings.Contains(gitStrategy, "git_username: alice") || !strings.Contains(gitStrategy, "store_tokens: false") {
		t.Fatalf("unexpected git strategy config: %s", gitStrategy)
	}

	codexProfile := mustRead(t, filepath.Join(tmp, ".namba", "config", "sections", "codex.yaml"))
	if !strings.Contains(codexProfile, "agent_mode: multi") || !strings.Contains(codexProfile, "status_line_preset: off") {
		t.Fatalf("unexpected codex config: %s", codexProfile)
	}

	codexConfig := mustRead(t, filepath.Join(tmp, ".codex", "config.toml"))
	if !strings.Contains(codexConfig, "max_threads = 3") {
		t.Fatalf("unexpected codex repo config: %s", codexConfig)
	}
	if strings.Contains(codexConfig, "status_line") {
		t.Fatalf("expected status line to be omitted when preset is off: %s", codexConfig)
	}
}

func TestInitRejectsUnsupportedMode(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	err := app.Run(context.Background(), []string{"init", tmp, "--yes", "--mode", "hybrid"})
	if err == nil {
		t.Fatal("expected invalid mode error")
	}
	if !strings.Contains(err.Error(), "development mode") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoctorReportsCodexNativeReadiness(t *testing.T) {
	tmp := t.TempDir()
	stdout := &bytes.Buffer{}
	app := namba.NewApp(stdout, &bytes.Buffer{})

	if err := app.Run(context.Background(), []string{"init", tmp}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"doctor"}); err != nil {
		t.Fatalf("doctor failed: %v", err)
	}

	text := stdout.String()
	if !strings.Contains(text, "Codex native repo: ready") {
		t.Fatalf("expected codex native readiness in doctor output: %s", text)
	}
	if !strings.Contains(text, "Codex compatibility mirror: ready") {
		t.Fatalf("expected codex compatibility readiness in doctor output: %s", text)
	}
}

func TestPlanCreatesSequentialSpecs(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := app.Run(context.Background(), []string{"init", tmp}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"plan", "build", "status", "page"}); err != nil {
		t.Fatalf("first plan failed: %v", err)
	}
	if err := app.Run(context.Background(), []string{"plan", "add", "sync", "report"}); err != nil {
		t.Fatalf("second plan failed: %v", err)
	}

	mustExist(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "spec.md"))
	mustExist(t, filepath.Join(tmp, ".namba", "specs", "SPEC-002", "spec.md"))
}

func TestProjectRunDryRunAndSync(t *testing.T) {
	tmp := t.TempDir()
	stdout := &bytes.Buffer{}
	app := namba.NewApp(stdout, &bytes.Buffer{})

	if err := app.Run(context.Background(), []string{"init", tmp}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "README.md"), []byte("# Demo"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"project"}); err != nil {
		t.Fatalf("project failed: %v", err)
	}
	if err := app.Run(context.Background(), []string{"plan", "implement", "healthcheck"}); err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	if err := app.Run(context.Background(), []string{"run", "SPEC-001", "--dry-run"}); err != nil {
		t.Fatalf("run dry-run failed: %v", err)
	}
	if err := app.Run(context.Background(), []string{"sync"}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	product, err := os.ReadFile(filepath.Join(tmp, ".namba", "project", "product.md"))
	if err != nil {
		t.Fatalf("read product doc: %v", err)
	}
	if !strings.Contains(string(product), "# Demo") {
		t.Fatalf("expected README content in product doc, got: %s", product)
	}

	mustExist(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-request.md"))
	mustExist(t, filepath.Join(tmp, ".namba", "project", "pr-checklist.md"))
}

func TestSyncRefreshesWorkflowDocs(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	stdout := &bytes.Buffer{}
	app := namba.NewApp(stdout, &bytes.Buffer{})

	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "README.md"), []byte("# Demo\n\nWorkflow docs.\n"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"plan", "sync", "workflow", "docs"}); err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	if err := app.Run(context.Background(), []string{"sync"}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	changeSummary := mustRead(t, filepath.Join(tmp, ".namba", "project", "change-summary.md"))
	if !strings.Contains(changeSummary, "`namba update`") || !strings.Contains(changeSummary, "`namba run SPEC-XXX --parallel`") {
		t.Fatalf("expected synced change summary to describe update and parallel workflow, got: %s", changeSummary)
	}

	releaseNotes := mustRead(t, filepath.Join(tmp, ".namba", "project", "release-notes.md"))
	if !strings.Contains(releaseNotes, "`namba release --push`") || !strings.Contains(releaseNotes, "`checksums.txt`") {
		t.Fatalf("expected synced release notes to describe release flow, got: %s", releaseNotes)
	}

	releaseChecklist := mustRead(t, filepath.Join(tmp, ".namba", "project", "release-checklist.md"))
	if !strings.Contains(releaseChecklist, "current branch is `main`") || !strings.Contains(releaseChecklist, "`namba update` rerun") {
		t.Fatalf("expected synced release checklist to describe release guardrails, got: %s", releaseChecklist)
	}
}

func TestInitConfiguresGoFormatTargets(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/demo\n\ngo 1.24.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "cmd", "demo"), 0o755); err != nil {
		t.Fatalf("mkdir cmd: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "cmd", "demo", "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "root_test.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write root_test.go: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "external", "ref"), 0o755); err != nil {
		t.Fatalf("mkdir external: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "external", "ref", "ignored.go"), []byte("package ref\n"), 0o644); err != nil {
		t.Fatalf("write ignored.go: %v", err)
	}

	if err := app.Run(context.Background(), []string{"init", tmp}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	quality, err := os.ReadFile(filepath.Join(tmp, ".namba", "config", "sections", "quality.yaml"))
	if err != nil {
		t.Fatalf("read quality config: %v", err)
	}

	text := string(quality)
	if !strings.Contains(text, `lint_command: gofmt -l "cmd" "root_test.go"`) {
		t.Fatalf("unexpected lint command: %s", text)
	}
	if strings.Contains(text, "external") {
		t.Fatalf("expected external paths to be excluded: %s", text)
	}
}

func mustExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func mustNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected %s not to exist", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func chdir(t *testing.T, dir string) func() {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to %s: %v", dir, err)
	}
	return func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}
}
