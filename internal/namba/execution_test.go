package namba

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunWritesStructuredLogs(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		switch name {
		case "codex", "git":
			return name, nil
		default:
			return "", errors.New("missing dependency")
		}
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexExec(name, args):
			if dir != tmp {
				t.Fatalf("expected codex workdir %s, got %s", tmp, dir)
			}
			return "runner output", nil
		case isShellCommand(name):
			return "validation ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"run", "SPEC-001"}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	result := mustReadExecutionResult(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-execution.json"))
	if !result.Succeeded {
		t.Fatalf("expected successful execution result: %+v", result)
	}
	if result.Runner != "codex" {
		t.Fatalf("expected codex runner, got %s", result.Runner)
	}

	report := mustReadValidationReport(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-validation.json"))
	if !report.Passed {
		t.Fatalf("expected successful validation report: %+v", report)
	}
	if len(report.Steps) != 3 {
		t.Fatalf("expected 3 validation steps, got %d", len(report.Steps))
	}
}

func TestRunRejectsUnsupportedRunner(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "system.yaml"), "runner: unsupported\napproval_mode: on-request\nsandbox_mode: workspace-write\n")

	err := app.Run(context.Background(), []string{"run", "SPEC-001"})
	if err == nil {
		t.Fatal("expected unsupported runner error")
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunWritesExecutionLogOnRunnerFailure(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		if name == "codex" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if isCodexExec(name, args) {
			return "partial output", errors.New("runner failed")
		}
		if isShellCommand(name) {
			t.Fatal("validators should not run after runner failure")
		}
		return "", nil
	}

	err := app.Run(context.Background(), []string{"run", "SPEC-001"})
	if err == nil {
		t.Fatal("expected runner failure")
	}

	result := mustReadExecutionResult(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-execution.json"))
	if result.Succeeded {
		t.Fatalf("expected failed execution result: %+v", result)
	}
	if !strings.Contains(result.Error, "runner failed") {
		t.Fatalf("expected runner failure in result: %+v", result)
	}

	raw := mustReadFile(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-result.txt"))
	if raw != "partial output" {
		t.Fatalf("unexpected raw output: %q", raw)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-validation.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected validation report to be absent, got err=%v", err)
	}
}

func TestRunWritesValidationReportOnValidationFailure(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		if name == "codex" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexExec(name, args):
			return "runner output", nil
		case isShellCommand(name):
			command := args[len(args)-1]
			if strings.Contains(command, "gofmt") {
				return "formatting failed", errors.New("lint failed")
			}
			return "ok", nil
		default:
			return "", nil
		}
	}

	err := app.Run(context.Background(), []string{"run", "SPEC-001"})
	if err == nil {
		t.Fatal("expected validation failure")
	}

	report := mustReadValidationReport(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-validation.json"))
	if report.Passed {
		t.Fatalf("expected failed validation report: %+v", report)
	}
	if len(report.Steps) < 2 {
		t.Fatalf("expected at least two steps, got %+v", report)
	}
	if report.Steps[1].Name != "lint" || !strings.Contains(report.Steps[1].Error, "lint failed") {
		t.Fatalf("expected lint failure step, got %+v", report.Steps[1])
	}
}

func prepareExecutionProject(t *testing.T) (string, *App, func()) {
	t.Helper()
	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	if err := app.Run(context.Background(), []string{"plan", "runner", "core"}); err != nil {
		restore()
		t.Fatalf("plan failed: %v", err)
	}
	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "quality.yaml"), "development_mode: tdd\ntest_command: go test ./...\nlint_command: gofmt -l .\ntypecheck_command: go vet ./...\n")

	return tmp, app, restore
}

func chdirExecution(t *testing.T, dir string) func() {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	return func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}
}

func isCodexExec(name string, args []string) bool {
	if runtime.GOOS == "windows" {
		return name == "cmd" && len(args) >= 3 && args[0] == "/c" && args[1] == "codex"
	}
	return name == "codex" && len(args) >= 1 && args[0] == "exec"
}

func isShellCommand(name string) bool {
	return name == "powershell" || name == "sh"
}

func mustReadExecutionResult(t *testing.T, path string) executionResult {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read execution result: %v", err)
	}
	var result executionResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal execution result: %v", err)
	}
	return result
}

func mustReadValidationReport(t *testing.T, path string) validationReport {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read validation report: %v", err)
	}
	var report validationReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("unmarshal validation report: %v", err)
	}
	return report
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	return string(data)
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
