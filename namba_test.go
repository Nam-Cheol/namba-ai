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
	mustExist(t, filepath.Join(tmp, ".codex", "skills", "namba-foundation-core", "SKILL.md"))
	mustExist(t, filepath.Join(tmp, ".namba", "config", "sections", "project.yaml"))
	mustExist(t, filepath.Join(tmp, ".namba", "manifest.json"))
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
