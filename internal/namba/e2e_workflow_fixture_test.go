package namba

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestE2EWorkflowFixtureHappyPath(t *testing.T) {
	tmp, app, fake, restore := prepareE2EWorkflowFixture(t)
	defer restore()

	runE2EApp(t, app, "project")
	runE2EApp(t, app, "plan", e2ePlanDescription("CLI workflow fixture", "exercise init project plan run queue pr land status"))
	markE2ESpecReviewsClear(t, tmp, "SPEC-001")
	runE2EApp(t, app, "harness", e2ePlanDescription("Harness fixture", "exercise deterministic harness planning artifacts"))
	runE2EApp(t, app, "run", "SPEC-001")
	runE2EApp(t, app, "queue", "start", "SPEC-001", "--auto-land")

	app.stdout = &bytes.Buffer{}
	runE2EApp(t, app, "status")
	statusOut := app.stdout.(*bytes.Buffer).String()
	for _, want := range []string{"Project: e2e-workflow", "SPEC packages: 2", "State dir: " + filepath.Join(tmp, ".namba")} {
		if !strings.Contains(statusOut, want) {
			t.Fatalf("status output missing %q: %s", want, statusOut)
		}
	}

	for _, rel := range []string{
		".namba/project/product.md",
		".namba/project/tech.md",
		".namba/project/structure.md",
		".namba/logs/project/codex-diagnostics-evidence.json",
		".namba/specs/SPEC-001/spec.md",
		".namba/specs/SPEC-001/plan.md",
		".namba/specs/SPEC-001/acceptance.md",
		".namba/specs/SPEC-002/harness-request.json",
		".namba/logs/runs/spec-001-execution.json",
		".namba/logs/runs/spec-001-validation.json",
		".namba/logs/runs/spec-001-queue-evidence.json",
		".namba/logs/queue/state.json",
		".namba/logs/queue/report.md",
	} {
		requireE2EFile(t, tmp, rel)
	}

	result := mustReadExecutionResult(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-execution.json"))
	if !result.Succeeded || result.Runner != "codex" || result.SpecID != "SPEC-001" {
		t.Fatalf("expected successful run evidence, got %+v", result)
	}
	validation := mustReadValidationReport(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-validation.json"))
	if !validation.Passed || len(validation.Steps) == 0 {
		t.Fatalf("expected successful validation evidence, got %+v", validation)
	}
	queueEvidence := mustReadE2EQueueRunnerEvidence(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-queue-evidence.json"))
	if queueEvidence.Status != "completed" || !queueEvidence.Validation.Passed {
		t.Fatalf("expected completed queue evidence, got %+v", queueEvidence)
	}
	state := mustReadE2EQueueState(t, tmp)
	if state.Status != queueStateDone || state.CompletedSpecCount != 1 {
		t.Fatalf("expected done queue state, got %+v", state)
	}
	specState := state.Specs["SPEC-001"]
	if specState.Phase != queuePhaseLanded || specState.PRNumber != 17 || !strings.Contains(specState.LandEvidence, "PR #17 merged") {
		t.Fatalf("expected landed queue spec with PR evidence, got %+v", specState)
	}

	fake.requireCommandContaining(t, "codex exec")
	fake.requireCommandContaining(t, "gh pr create")
	fake.requireCommandContaining(t, "gh pr merge 17 --merge --match-head-commit")
}

func TestE2EWorkflowFixtureBlocksUnsafeQueueState(t *testing.T) {
	tmp, app, fake, restore := prepareE2EWorkflowFixture(t)
	defer restore()

	runE2EApp(t, app, "plan", e2ePlanDescription("Unsafe queue fixture", "create a SPEC before dirty queue state is simulated"))
	markE2ESpecReviewsClear(t, tmp, "SPEC-001")

	fake.dirtyStatus = " M unsafe.txt"
	runE2EApp(t, app, "queue", "start", "SPEC-001")

	state := mustReadE2EQueueState(t, tmp)
	if state.Status != queueStateBlocked || state.Detail != "branch_ready_failed" {
		t.Fatalf("expected blocked branch-ready queue state, got %+v", state)
	}
	if !strings.Contains(state.LastBlocker, "uncommitted changes") || state.LastRecoveryAction == "" {
		t.Fatalf("expected diagnosable unsafe-state blocker, got %+v", state)
	}
	if spec := state.Specs["SPEC-001"]; spec.Phase != queuePhaseBlocked || spec.RecoveryAction == "" {
		t.Fatalf("expected blocked spec state with recovery action, got %+v", spec)
	}
}

func TestE2EWorkflowFixtureRecordsValidationFailureEvidence(t *testing.T) {
	tmp, app, fake, restore := prepareE2EWorkflowFixture(t)
	defer restore()

	runE2EApp(t, app, "plan", e2ePlanDescription("Validation failure fixture", "create a SPEC that fails fake validation during queue execution"))
	markE2ESpecReviewsClear(t, tmp, "SPEC-001")

	fake.failShellCommands["go test ./..."] = "simulated validation failure"
	runE2EApp(t, app, "queue", "start", "SPEC-001")

	state := mustReadE2EQueueState(t, tmp)
	if state.Status != queueStateBlocked || state.Detail != "runner_failed" {
		t.Fatalf("expected runner_failed queue blocker, got %+v", state)
	}
	if !strings.Contains(state.LastBlocker, "simulated validation failure") || state.LastEvidencePath != queueRunnerEvidencePath("SPEC-001") {
		t.Fatalf("expected validation failure blocker with queue evidence path, got %+v", state)
	}
	queueEvidence := mustReadE2EQueueRunnerEvidence(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-queue-evidence.json"))
	if queueEvidence.Status != "runner_failed" || queueEvidence.Validation.Passed {
		t.Fatalf("expected failed queue runner evidence, got %+v", queueEvidence)
	}
	if !strings.Contains(queueEvidence.Validation.Error, "simulated validation failure") {
		t.Fatalf("expected validation error detail, got %+v", queueEvidence)
	}
	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "validation_failed" {
		t.Fatalf("expected validation_failed execution evidence manifest, got %+v", manifest)
	}
}

func TestE2EFakeRunnerRejectsUnexpectedExternalCommand(t *testing.T) {
	tmp := canonicalTempDir(t)
	fake := newE2EFakeRunner(tmp)

	_, err := fake.run(context.Background(), "curl", []string{"https://example.com"}, tmp)
	if err == nil || !strings.Contains(err.Error(), "unexpected external command") || !strings.Contains(err.Error(), "curl https://example.com") {
		t.Fatalf("expected clear unexpected-command diagnostic, got %v", err)
	}
}

func prepareE2EWorkflowFixture(t *testing.T) (string, *App, *e2eFakeRunner, func()) {
	t.Helper()

	tmp := canonicalTempDir(t)
	writeTestFile(t, filepath.Join(tmp, "go.mod"), "module example.com/e2e\n\ngo 1.22\n")
	writeTestFile(t, filepath.Join(tmp, "main.go"), "package main\n\nfunc main() {}\n")

	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.now = func() time.Time { return time.Date(2026, 5, 21, 9, 30, 0, 0, time.UTC) }
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes", "--name", "e2e-workflow", "--mode", "tdd", "--project-type", "existing"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "quality.yaml"), strings.Join([]string{
		"development_mode: tdd",
		"test_command: go test ./...",
		"lint_command: gofmt -l .",
		"typecheck_command: go vet ./...",
		"build_command: none",
		"migration_dry_run_command: none",
		"smoke_start_command: none",
		"output_contract_command: none",
		"",
	}, "\n"))

	fake := newE2EFakeRunner(tmp)
	fake.install(app)
	restore := chdirExecution(t, tmp)
	return tmp, app, fake, restore
}

type e2eExternalCall struct {
	Name string
	Args []string
	Dir  string
}

type e2eFakeRunner struct {
	root              string
	currentBranch     string
	branches          map[string]bool
	dirtyStatus       string
	prCreated         bool
	prMerged          bool
	failShellCommands map[string]string
	calls             []e2eExternalCall
}

func newE2EFakeRunner(root string) *e2eFakeRunner {
	return &e2eFakeRunner{
		root:              filepath.Clean(root),
		currentBranch:     "main",
		branches:          map[string]bool{"main": true},
		failShellCommands: map[string]string{},
	}
}

func (f *e2eFakeRunner) install(app *App) {
	app.lookPath = func(name string) (string, error) {
		switch name {
		case "codex", "git", "gh":
			return name, nil
		default:
			return "", fmt.Errorf("E2E fake runner: dependency %q is not available", name)
		}
	}
	app.detectCodexCapabilities = func(context.Context, string, executionRequest) (codexCapabilityMatrix, error) {
		return testCodexCapabilities(), nil
	}
	app.runCmd = f.run
	app.runCmdWithInput = f.runWithInput
	app.runCodexCmdWithInput = f.runWithInput
	app.getenv = func(key string) string {
		switch key {
		case "NAMBA_CODEX_EFFECTIVE_WORKSPACE_ROOTS":
			return f.root
		default:
			return ""
		}
	}
}

func (f *e2eFakeRunner) run(_ context.Context, name string, args []string, dir string) (string, error) {
	f.calls = append(f.calls, e2eExternalCall{Name: name, Args: append([]string(nil), args...), Dir: filepath.Clean(dir)})
	if isShellCommand(name) {
		return f.runShell(args)
	}
	if name == "cmd" && len(args) >= 3 && args[0] == "/c" && args[1] == "codex" {
		return f.runCodex(args[2:])
	}

	switch name {
	case "codex":
		return f.runCodex(args)
	case "git":
		return f.runGit(args)
	case "gh":
		return f.runGH(args)
	default:
		return "", f.unexpected(name, args)
	}
}

func (f *e2eFakeRunner) runWithInput(_ context.Context, name string, args []string, dir, input string) (string, string, error) {
	f.calls = append(f.calls, e2eExternalCall{Name: name, Args: append([]string(nil), args...), Dir: filepath.Clean(dir)})
	if name == "codex" && strings.Join(args, " ") == "doctor" && input == "" {
		return "codex doctor ok\n", "", nil
	}
	if name == "codex" && len(args) > 0 && args[0] == "exec" && input != "" {
		out, err := f.runCodex(args)
		return out, "", err
	}
	return "", "", f.unexpected(name, args)
}

func (f *e2eFakeRunner) runShell(args []string) (string, error) {
	if len(args) == 0 {
		return "", f.unexpected("shell", args)
	}
	command := strings.TrimSpace(args[len(args)-1])
	if message := strings.TrimSpace(f.failShellCommands[command]); message != "" {
		return message, errors.New(message)
	}
	return "validation ok: " + command, nil
}

func (f *e2eFakeRunner) runCodex(args []string) (string, error) {
	switch {
	case len(args) == 1 && args[0] == "--version":
		return "codex 0.131.0", nil
	case strings.Join(args, " ") == "remote-control --help":
		return "Usage: codex remote-control <status>", nil
	case strings.Join(args, " ") == "remote-control status":
		return "disabled", nil
	case len(args) > 0 && args[0] == "exec":
		return "E2E fake Codex completed the requested turn.", nil
	default:
		return "", f.unexpected("codex", args)
	}
}

func (f *e2eFakeRunner) runGit(args []string) (string, error) {
	joined := strings.Join(args, " ")
	switch {
	case joined == "worktree list --porcelain":
		return renderPlanningWorktreeList(gitWorktree{Path: f.root, Branch: f.currentBranch}), nil
	case joined == "for-each-ref --format=%(refname:short) refs/heads":
		return strings.Join(f.sortedBranches(), "\n"), nil
	case len(args) == 6 && args[0] == "ls-tree" && args[1] == "-r" && args[2] == "--name-only" && args[3] == "--full-tree" && args[5] == specsDir:
		return f.specTreeForBranch(args[4]), nil
	case joined == "branch --show-current":
		return f.currentBranch, nil
	case joined == "status --porcelain":
		return f.dirtyStatus, nil
	case len(args) == 3 && args[0] == "branch" && args[1] == "--list":
		if f.branches[args[2]] {
			return "  " + args[2], nil
		}
		return "", nil
	case len(args) == 4 && args[0] == "checkout" && args[1] == "-b":
		f.branches[args[2]] = true
		f.currentBranch = args[2]
		return "", nil
	case len(args) == 2 && args[0] == "checkout":
		if !f.branches[args[1]] {
			return "", fmt.Errorf("branch %s not found", args[1])
		}
		f.currentBranch = args[1]
		return "", nil
	case len(args) == 4 && args[0] == "merge-base" && args[1] == "--is-ancestor":
		return "", fmt.Errorf("%s is not an ancestor of %s", args[2], args[3])
	case joined == "rev-parse HEAD":
		return "head-001", nil
	case len(args) == 4 && args[0] == "push" && args[1] == "--set-upstream":
		return "", nil
	case len(args) >= 1 && args[0] == "add":
		return "", nil
	case len(args) >= 3 && args[0] == "commit" && args[1] == "-m":
		f.dirtyStatus = ""
		return "", nil
	case len(args) == 3 && args[0] == "fetch":
		return "", nil
	case len(args) == 4 && args[0] == "branch" && args[1] == "-f":
		f.branches[args[2]] = true
		return "", nil
	case len(args) == 4 && args[0] == "log" && args[1] == "-1" && args[2] == "--format=%H":
		return "base-spec-commit", nil
	default:
		return "", f.unexpected("git", args)
	}
}

func (f *e2eFakeRunner) runGH(args []string) (string, error) {
	joined := strings.Join(args, " ")
	switch {
	case joined == "auth status":
		return "", nil
	case len(args) >= 2 && args[0] == "pr" && args[1] == "list":
		return "[]", nil
	case len(args) >= 2 && args[0] == "pr" && args[1] == "create":
		f.prCreated = true
		return "https://github.com/example/e2e/pull/17", nil
	case len(args) >= 2 && args[0] == "pr" && args[1] == "view":
		return f.renderPullRequest(args[2], strings.Join(args, " ")), nil
	case len(args) >= 2 && args[0] == "pr" && args[1] == "comment":
		return "", nil
	case len(args) >= 4 && args[0] == "pr" && args[1] == "merge" && args[2] == "17":
		f.prMerged = true
		return "", nil
	default:
		return "", f.unexpected("gh", args)
	}
}

func (f *e2eFakeRunner) renderPullRequest(selector, joinedArgs string) string {
	pr := githubPullRequest{
		Number:           17,
		URL:              "https://github.com/example/e2e/pull/17",
		Title:            "SPEC-001 CLI workflow fixture",
		HeadRefName:      f.currentBranch,
		HeadRefOID:       "head-001",
		BaseRefName:      "main",
		ReviewDecision:   "APPROVED",
		MergeStateStatus: "CLEAN",
		StatusChecks: []githubStatusCheck{
			{WorkflowName: "go test", Status: "COMPLETED", Conclusion: "SUCCESS"},
		},
	}
	if selector == "17" && strings.Contains(joinedArgs, "comments") && !strings.Contains(joinedArgs, "statusCheckRollup") {
		pr.Comments = []githubPRComment{}
	}
	data, err := json.Marshal(pr)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (f *e2eFakeRunner) sortedBranches() []string {
	branches := make([]string, 0, len(f.branches))
	for branch := range f.branches {
		branches = append(branches, branch)
	}
	sort.Strings(branches)
	return branches
}

func (f *e2eFakeRunner) specTreeForBranch(branch string) string {
	if branch == "main" {
		return ".namba/specs/.gitkeep"
	}
	parts := strings.Split(branch, "/")
	tail := parts[len(parts)-1]
	if strings.HasPrefix(tail, "SPEC-") && len(tail) >= len("SPEC-001") {
		return ".namba/specs/" + tail[:len("SPEC-001")] + "/spec.md"
	}
	return ".namba/specs/.gitkeep"
}

func (f *e2eFakeRunner) unexpected(name string, args []string) error {
	return fmt.Errorf("E2E fake runner: unexpected external command: %s %s", name, strings.Join(args, " "))
}

func (f *e2eFakeRunner) requireCommandContaining(t *testing.T, needle string) {
	t.Helper()
	for _, call := range f.calls {
		command := call.Name + " " + strings.Join(call.Args, " ")
		if strings.Contains(command, needle) {
			return
		}
	}
	t.Fatalf("expected fake runner command containing %q, got %+v", needle, f.calls)
}

func e2ePlanDescription(goal, scope string) string {
	return strings.Join([]string{
		"Goal: " + goal,
		"Scope: " + scope,
		"Constraints: offline deterministic CLI E2E only; no live Codex, GitHub, network, tokens, or real pull requests.",
		"Acceptance: generated artifacts, structured evidence, queue state, PR simulation, and land readiness are verified through App.Run.",
	}, "\n")
}

func runE2EApp(t *testing.T, app *App, args ...string) {
	t.Helper()
	if err := app.Run(context.Background(), args); err != nil {
		t.Fatalf("namba %s failed: %v", strings.Join(args, " "), err)
	}
}

func markE2ESpecReviewsClear(t *testing.T, root, specID string) {
	t.Helper()
	reviewsDir := filepath.Join(root, ".namba", "specs", specID, "reviews")
	for _, template := range specReviewTemplates() {
		writeTestFile(t, filepath.Join(reviewsDir, template.Slug+".md"), strings.Join([]string{
			"# " + template.Title,
			"",
			"- Status: clear",
			"- Last Reviewed: 2026-05-21",
			"- Reviewer: e2e-fixture",
			"",
			"## Findings",
			"",
			"- Clear for deterministic E2E fixture coverage.",
			"",
			"## Follow-ups",
			"",
			"- [non-blocking] None.",
			"",
		}, "\n"))
	}
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if _, err := app.refreshSpecReviewReadiness(root, specID); err != nil {
		t.Fatalf("refresh review readiness: %v", err)
	}
}

func requireE2EFile(t *testing.T, root, rel string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
		t.Fatalf("expected %s: %v", rel, err)
	}
}

func mustReadE2EQueueState(t *testing.T, root string) queueState {
	t.Helper()
	state, err := readQueueState(root)
	if err != nil {
		t.Fatalf("read queue state: %v", err)
	}
	return state
}

func mustReadE2EQueueRunnerEvidence(t *testing.T, path string) queueRunnerEvidence {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read queue runner evidence: %v", err)
	}
	var evidence queueRunnerEvidence
	if err := json.Unmarshal(data, &evidence); err != nil {
		t.Fatalf("unmarshal queue runner evidence: %v", err)
	}
	return evidence
}
