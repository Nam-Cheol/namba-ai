package namba

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type parallelHarness struct {
	t              *testing.T
	tmp            string
	stdout         *bytes.Buffer
	app            *App
	commands       []string
	codexCount     int
	lastCodexIdx   int
	mu             sync.Mutex
	execFailures   map[string]error
	validationErr  map[string]error
	mergeErr       map[string]error
	worktreeAddErr map[string]error
	removeErr      map[string]error
	branchErr      map[string]error
	pruneErr       error
}

func newParallelHarness(t *testing.T) (*parallelHarness, func()) {
	t.Helper()
	tmp := t.TempDir()
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "quality.yaml"), "development_mode: tdd\ntest_command: test\nlint_command: none\ntypecheck_command: none\n")

	h := &parallelHarness{
		t:              t,
		tmp:            tmp,
		stdout:         stdout,
		app:            app,
		lastCodexIdx:   -1,
		execFailures:   map[string]error{},
		validationErr:  map[string]error{},
		mergeErr:       map[string]error{},
		worktreeAddErr: map[string]error{},
		removeErr:      map[string]error{},
		branchErr:      map[string]error{},
	}
	app.lookPath = func(name string) (string, error) {
		switch name {
		case "git", "codex":
			return name, nil
		default:
			return "", fmt.Errorf("missing %s", name)
		}
	}
	app.runCmd = h.runCmd
	app.detectCodexCapabilities = func(context.Context, string, executionRequest) (codexCapabilityMatrix, error) {
		return testCodexCapabilities(), nil
	}

	restore := chdirExecution(t, tmp)
	return h, restore
}

func (h *parallelHarness) runCmd(_ context.Context, name string, args []string, dir string) (string, error) {
	h.mu.Lock()
	h.commands = append(h.commands, name+" "+strings.Join(args, " "))
	if isCodexExec(name, args) {
		h.codexCount++
		h.lastCodexIdx = len(h.commands) - 1
	}
	h.mu.Unlock()

	switch {
	case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
		return "main", nil
	case name == "git" && len(args) >= 5 && args[0] == "worktree" && args[1] == "add" && args[2] == "-b":
		path := args[4]
		if err := h.worktreeAddErr[path]; err != nil {
			return "", err
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			return "", err
		}
		return "", nil
	case isCodexExec(name, args):
		if err := h.execFailures[dir]; err != nil {
			return "runner failed", err
		}
		return "ok", nil
	case isShellCommand(name):
		if err := h.validationErr[dir]; err != nil {
			return "validation failed", err
		}
		return "ok", nil
	case name == "git" && len(args) >= 3 && args[0] == "merge" && args[1] == "--no-ff":
		branch := args[2]
		if err := h.mergeErr[branch]; err != nil {
			return "", err
		}
		return "", nil
	case name == "git" && len(args) == 4 && args[0] == "worktree" && args[1] == "remove" && args[2] == "--force":
		path := args[3]
		if err := h.removeErr[path]; err != nil {
			return "", err
		}
		return "", nil
	case name == "git" && len(args) == 3 && args[0] == "branch" && args[1] == "-D":
		branch := args[2]
		if err := h.branchErr[branch]; err != nil {
			return "", err
		}
		return "", nil
	case name == "git" && len(args) == 2 && args[0] == "worktree" && args[1] == "prune":
		if h.pruneErr != nil {
			return "", h.pruneErr
		}
		return "", nil
	default:
		h.t.Fatalf("unexpected command: %s %v", name, args)
		return "", nil
	}
}

func TestRunParallelAllSuccessCleansUpAndWritesReport(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	err := h.app.runParallel(context.Background(), h.tmp, specPackage{ID: "SPEC-003"}, []string{"one", "two", "three", "four"}, "prompt", qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"}, systemConfig{Runner: "codex"}, codexConfig{}, workflowConfig{MaxParallelWorkers: 3}, false)
	if err != nil {
		t.Fatalf("runParallel failed: %v", err)
	}

	firstMerge := indexCommandContaining(h.commands, "git merge --no-ff")
	lastExec := h.lastCodexIdx
	if firstMerge == -1 || lastExec == -1 || firstMerge <= lastExec {
		t.Fatalf("expected merges after all worker executions, got %v", h.commands)
	}
	if countCommandsContaining(h.commands, "git worktree remove --force") != 3 {
		t.Fatalf("expected cleanup worktree removals, got %v", h.commands)
	}
	if countCommandsContaining(h.commands, "git branch -D namba/spec-003-p") != 3 {
		t.Fatalf("expected cleanup branch deletions, got %v", h.commands)
	}

	report := mustReadParallelReport(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel.json"))
	if report.MergeBlocked {
		t.Fatalf("expected merge allowed report, got %+v", report)
	}
	for _, worker := range report.Workers {
		if !worker.MergeSucceeded || !worker.WorktreeRemoved || !worker.BranchRemoved || worker.Preserved {
			t.Fatalf("unexpected worker report: %+v", worker)
		}
	}
}

func TestRunParallelExecutionFailureBlocksMergeAndPreservesWorktrees(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	failedDir := filepath.Join(h.tmp, worktreesDir, "spec-003-p2")
	h.execFailures[failedDir] = errors.New("runner boom")

	err := h.app.runParallel(context.Background(), h.tmp, specPackage{ID: "SPEC-003"}, []string{"one", "two", "three"}, "prompt", qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"}, systemConfig{Runner: "codex"}, codexConfig{}, workflowConfig{MaxParallelWorkers: 3}, false)
	if err == nil {
		t.Fatal("expected execution failure")
	}
	if !strings.Contains(err.Error(), "blocked merge") {
		t.Fatalf("unexpected error: %v", err)
	}
	if countCommandsContaining(h.commands, "git merge --no-ff") != 0 {
		t.Fatalf("expected no merge attempts, got %v", h.commands)
	}
	if countCommandsContaining(h.commands, "git worktree remove --force") != 0 {
		t.Fatalf("expected no cleanup on failure, got %v", h.commands)
	}

	report := mustReadParallelReport(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel.json"))
	if !report.MergeBlocked {
		t.Fatalf("expected merge blocked report, got %+v", report)
	}
	failed := findWorker(report.Workers, "spec-003-p2")
	if failed.ExecutionError == "" || !failed.Preserved {
		t.Fatalf("expected execution failure to be recorded, got %+v", failed)
	}
}

func TestRunParallelValidationFailureBlocksMergeAndPreservesWorktrees(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	failedDir := filepath.Join(h.tmp, worktreesDir, "spec-003-p3")
	h.validationErr[failedDir] = errors.New("lint failed")

	err := h.app.runParallel(context.Background(), h.tmp, specPackage{ID: "SPEC-003"}, []string{"one", "two", "three"}, "prompt", qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"}, systemConfig{Runner: "codex"}, codexConfig{}, workflowConfig{MaxParallelWorkers: 3}, false)
	if err == nil {
		t.Fatal("expected validation failure")
	}
	if !strings.Contains(err.Error(), "blocked merge") {
		t.Fatalf("unexpected error: %v", err)
	}
	if countCommandsContaining(h.commands, "git merge --no-ff") != 0 {
		t.Fatalf("expected no merge attempts, got %v", h.commands)
	}

	report := mustReadParallelReport(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel.json"))
	failed := findWorker(report.Workers, "spec-003-p3")
	if failed.ValidationError == "" || !failed.Preserved {
		t.Fatalf("expected validation failure to be recorded, got %+v", failed)
	}
}

func TestRunParallelMergeFailureIsReportedExplicitly(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	h.mergeErr["namba/spec-003-p2"] = errors.New("merge conflict")

	err := h.app.runParallel(context.Background(), h.tmp, specPackage{ID: "SPEC-003"}, []string{"one", "two", "three"}, "prompt", qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"}, systemConfig{Runner: "codex"}, codexConfig{}, workflowConfig{MaxParallelWorkers: 3}, false)
	if err == nil {
		t.Fatal("expected merge failure")
	}
	if !strings.Contains(err.Error(), "merge namba/spec-003-p2") {
		t.Fatalf("unexpected error: %v", err)
	}
	if countCommandsContaining(h.commands, "git worktree remove --force") != 0 {
		t.Fatalf("expected no cleanup after merge failure, got %v", h.commands)
	}

	report := mustReadParallelReport(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel.json"))
	failed := findWorker(report.Workers, "spec-003-p2")
	if failed.MergeError == "" || !failed.Preserved {
		t.Fatalf("expected merge failure to be recorded, got %+v", failed)
	}
}

func TestRunParallelCleanupFailureIsReportedExplicitly(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	failedPath := filepath.Join(h.tmp, worktreesDir, "spec-003-p1")
	h.removeErr[failedPath] = errors.New("locked worktree")

	err := h.app.runParallel(context.Background(), h.tmp, specPackage{ID: "SPEC-003"}, []string{"one", "two", "three"}, "prompt", qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"}, systemConfig{Runner: "codex"}, codexConfig{}, workflowConfig{MaxParallelWorkers: 3}, false)
	if err == nil {
		t.Fatal("expected cleanup failure")
	}
	if !strings.Contains(err.Error(), "parallel cleanup failed") {
		t.Fatalf("unexpected error: %v", err)
	}

	report := mustReadParallelReport(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel.json"))
	failed := findWorker(report.Workers, "spec-003-p1")
	if failed.CleanupError == "" || failed.WorktreeRemoved {
		t.Fatalf("expected cleanup failure to be recorded, got %+v", failed)
	}
}

func TestRunParallelDryRunSkipsExecutionMergeAndCleanup(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	err := h.app.runParallel(context.Background(), h.tmp, specPackage{ID: "SPEC-003"}, []string{"one", "two", "three"}, "prompt", qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"}, systemConfig{Runner: "codex"}, codexConfig{}, workflowConfig{MaxParallelWorkers: 3}, true)
	if err != nil {
		t.Fatalf("runParallel failed: %v", err)
	}
	if h.codexCount != 0 {
		t.Fatalf("expected no codex execution in dry-run, got %v", h.commands)
	}
	if countCommandsContaining(h.commands, "git merge --no-ff") != 0 {
		t.Fatalf("expected no merge in dry-run, got %v", h.commands)
	}
	if countCommandsContaining(h.commands, "git worktree remove --force") != 0 {
		t.Fatalf("expected no cleanup in dry-run, got %v", h.commands)
	}
}

func TestRunParallelPreflightAllowsResumeProfileViaExecLevelFlags(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	writeTestFile(t, filepath.Join(h.tmp, ".namba", "config", "sections", "codex.yaml"), "agent_mode: multi\nstatus_line_preset: namba\nrepo_skills_path: .agents/skills\nrepo_agents_path: .codex/agents\nprofile: namba\nsession_mode: stateful\nrepair_attempts: 1\n")

	err := h.app.runParallel(context.Background(), h.tmp, specPackage{ID: "SPEC-003"}, []string{"one", "two", "three"}, "prompt", qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"}, systemConfig{Runner: "codex"}, codexConfig{Profile: "namba", SessionMode: "stateful", RepairAttempts: 1}, workflowConfig{MaxParallelWorkers: 3}, false)
	if err != nil {
		t.Fatalf("expected parallel preflight to allow exec-level resume profile flags, got %v", err)
	}
	if countCommandsContaining(h.commands, "git worktree add -b") != 3 {
		t.Fatalf("expected worktree creation after successful preflight, got %v", h.commands)
	}
}

func mustReadParallelReport(t *testing.T, path string) parallelRunReport {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read parallel report: %v", err)
	}
	var report parallelRunReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("unmarshal parallel report: %v", err)
	}
	return report
}

func findWorker(workers []parallelWorkerResult, name string) parallelWorkerResult {
	for _, worker := range workers {
		if worker.Name == name {
			return worker
		}
	}
	return parallelWorkerResult{}
}

func countCommandsContaining(commands []string, needle string) int {
	count := 0
	for _, command := range commands {
		if strings.Contains(command, needle) {
			count++
		}
	}
	return count
}

func indexCommandContaining(commands []string, needle string) int {
	for i, command := range commands {
		if strings.Contains(command, needle) {
			return i
		}
	}
	return -1
}
