package namba

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunParallelSkipsMergeWhenAnyWorkerFails(t *testing.T) {
	root, app, commands := newParallelTestApp(t, func(name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "main", nil
		case name == "git" && len(args) >= 2 && args[0] == "worktree" && args[1] == "add":
			return "", nil
		case isCodexExec(name, args):
			if strings.HasSuffix(dir, "spec-001-p2") {
				return "", errors.New("worker failed")
			}
			return "ok", nil
		case isShellCommand(name):
			return "ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	})

	err := app.runParallel(context.Background(), root, specPackage{ID: "SPEC-001"}, []string{"task one", "task two"}, "prompt", qualityConfig{TestCommand: "echo ok", LintCommand: "none", TypecheckCommand: "none"}, systemConfig{Runner: "codex", ApprovalPolicy: "on-request", SandboxMode: "workspace-write"}, false)
	if err == nil {
		t.Fatal("expected parallel failure")
	}
	if !strings.Contains(err.Error(), "parallel execution blocked merge") || !strings.Contains(err.Error(), "spec-001-p2 execution failed: worker failed") {
		t.Fatalf("unexpected error: %v", err)
	}
	if anyCommandHasPrefix(*commands, "git merge") {
		t.Fatalf("expected no merges when a worker fails, got %v", *commands)
	}
	if anyCommandHasPrefix(*commands, "git worktree remove --force") || anyCommandHasPrefix(*commands, "git branch -D") {
		t.Fatalf("expected failed workers to be preserved for inspection, got %v", *commands)
	}
}

func TestRunParallelMergesOnlyAfterAllWorkersPass(t *testing.T) {
	root, app, commands := newParallelTestApp(t, func(name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "main", nil
		case name == "git" && len(args) >= 2 && args[0] == "worktree" && args[1] == "add":
			return "", nil
		case isCodexExec(name, args):
			return "ok", nil
		case isShellCommand(name):
			return "ok", nil
		case name == "git" && len(args) >= 1 && args[0] == "merge":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "worktree" && args[1] == "remove":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "-D":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "worktree" && args[1] == "prune":
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	})

	err := app.runParallel(context.Background(), root, specPackage{ID: "SPEC-002"}, []string{"task one", "task two"}, "prompt", qualityConfig{TestCommand: "echo ok", LintCommand: "none", TypecheckCommand: "none"}, systemConfig{Runner: "codex", ApprovalPolicy: "on-request", SandboxMode: "workspace-write"}, false)
	if err != nil {
		t.Fatalf("parallel run failed: %v", err)
	}

	firstMerge := firstCommandIndex(*commands, "git merge")
	lastValidation := lastCommandIndex(*commands, validationCommandPrefix("echo ok"))
	if firstMerge == -1 || lastValidation == -1 {
		t.Fatalf("expected validation and merge commands, got %v", *commands)
	}
	if firstMerge <= lastValidation {
		t.Fatalf("expected merges after all validations, got %v", *commands)
	}
}

func TestRunParallelReportsCleanupFailures(t *testing.T) {
	root, app, _ := newParallelTestApp(t, func(name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "main", nil
		case name == "git" && len(args) >= 2 && args[0] == "worktree" && args[1] == "add":
			return "", nil
		case isCodexExec(name, args):
			return "ok", nil
		case isShellCommand(name):
			return "ok", nil
		case name == "git" && len(args) >= 1 && args[0] == "merge":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "worktree" && args[1] == "remove":
			return "", errors.New("cleanup remove failed")
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "-D":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "worktree" && args[1] == "prune":
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	})

	err := app.runParallel(context.Background(), root, specPackage{ID: "SPEC-003"}, []string{"task one"}, "prompt", qualityConfig{TestCommand: "echo ok", LintCommand: "none", TypecheckCommand: "none"}, systemConfig{Runner: "codex", ApprovalPolicy: "on-request", SandboxMode: "workspace-write"}, false)
	if err == nil {
		t.Fatal("expected cleanup failure")
	}
	if !strings.Contains(err.Error(), "parallel cleanup failed") || !strings.Contains(err.Error(), "cleanup remove failed") {
		t.Fatalf("expected cleanup failure details, got %v", err)
	}
}

func newParallelTestApp(t *testing.T, responder func(name string, args []string, dir string) (string, error)) (string, *App, *[]string) {
	t.Helper()

	root := t.TempDir()
	for _, rel := range []string{".git", ".namba/logs/runs", ".namba/worktrees"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(rel)), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
	}

	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	app.lookPath = func(name string) (string, error) {
		switch name {
		case "git", "codex":
			return name, nil
		default:
			return "", errors.New("missing dependency")
		}
	}

	commands := make([]string, 0, 16)
	app.runCmd = func(ctx context.Context, name string, args []string, dir string) (string, error) {
		commands = append(commands, name+" "+strings.Join(args, " "))
		return responder(name, args, dir)
	}
	return root, app, &commands
}

func anyCommandHasPrefix(commands []string, prefix string) bool {
	return firstCommandIndex(commands, prefix) != -1
}

func firstCommandIndex(commands []string, prefix string) int {
	for i, command := range commands {
		if strings.HasPrefix(command, prefix) {
			return i
		}
	}
	return -1
}

func lastCommandIndex(commands []string, prefix string) int {
	for i := len(commands) - 1; i >= 0; i-- {
		if strings.HasPrefix(commands[i], prefix) {
			return i
		}
	}
	return -1
}

func validationCommandPrefix(command string) string {
	if runtime.GOOS == "windows" {
		return "powershell -NoProfile -Command " + command
	}
	return "sh -lc " + command
}
