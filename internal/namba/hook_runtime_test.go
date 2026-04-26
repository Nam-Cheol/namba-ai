package namba

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLoadHookConfigSelectsEnabledHooksInNameOrder(t *testing.T) {
	tmp := canonicalTempDir(t)
	writeTestFile(t, filepath.Join(tmp, ".namba", "hooks.toml"), `
[hooks.z_after]
event = "before_validation"
command = "echo z"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = true

[hooks.disabled]
event = "before_validation"
command = "echo disabled"
cwd = "."
timeout = 5
enabled = false
continue_on_failure = true

[hooks.a_first]
event = "before_validation"
command = "echo a"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = true
`)

	cfg, err := loadHookConfig(tmp)
	if err != nil {
		t.Fatalf("loadHookConfig failed: %v", err)
	}
	hooks := cfg.enabledHooksForEvent(hookEventBeforeValidation)
	if len(hooks) != 2 {
		t.Fatalf("expected two enabled hooks, got %+v", hooks)
	}
	if hooks[0].Name != "a_first" || hooks[1].Name != "z_after" {
		t.Fatalf("expected deterministic hook_name order, got %+v", hooks)
	}
}

func TestRunShellCommandWithInputUsesNonLoginShell(t *testing.T) {
	var gotName string
	var gotArgs []string
	_, _, err := runShellCommandWithInput(
		context.Background(),
		func(_ context.Context, name string, args []string, dir, input string) (string, string, error) {
			gotName = name
			gotArgs = append([]string(nil), args...)
			return "", "", nil
		},
		"echo ok",
		"/tmp",
		"{}",
	)
	if err != nil {
		t.Fatalf("runShellCommandWithInput failed: %v", err)
	}
	if runtime.GOOS == "windows" {
		if gotName != "powershell" || strings.Join(gotArgs, "\x00") != strings.Join([]string{"-NoProfile", "-Command", "echo ok"}, "\x00") {
			t.Fatalf("expected powershell command, got %s %v", gotName, gotArgs)
		}
		return
	}
	if gotName != "sh" || strings.Join(gotArgs, "\x00") != strings.Join([]string{"-c", "echo ok"}, "\x00") {
		t.Fatalf("expected non-login sh -c command, got %s %v", gotName, gotArgs)
	}
}

func TestExecuteRunRecordsHookEvidenceAndArtifacts(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "hooks.toml"), `
[hooks.capture_context]
event = "before_validation"
command = "capture-context"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = true
`)

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
			return "runner output", nil
		case isShellCommand(name):
			return "validation ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}
	app.runCmdWithInput = func(_ context.Context, name string, args []string, dir, input string) (string, string, error) {
		if !isShellCommand(name) || !strings.Contains(strings.Join(args, " "), "capture-context") {
			t.Fatalf("unexpected hook command: %s %v", name, args)
		}
		if dir != tmp {
			t.Fatalf("expected hook cwd %s, got %s", tmp, dir)
		}
		var ctx hookExecutionContext
		if err := json.Unmarshal([]byte(input), &ctx); err != nil {
			t.Fatalf("hook context is not JSON: %v", err)
		}
		if ctx.SchemaVersion != hookContextSchemaVersion || ctx.Event != string(hookEventBeforeValidation) || ctx.SpecID != "SPEC-001" {
			t.Fatalf("unexpected hook context: %+v", ctx)
		}
		wantEvidencePath := filepath.ToSlash(filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
		if ctx.Artifacts["evidence"] != wantEvidencePath {
			t.Fatalf("expected evidence artifact path in context, got %+v", ctx.Artifacts)
		}
		return "hook stdout", "hook stderr", nil
	}

	if err := app.Run(context.Background(), []string{"run", "SPEC-001"}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if len(manifest.Hooks) != 1 {
		t.Fatalf("expected one hook result, got %+v", manifest.Hooks)
	}
	hook := manifest.Hooks[0]
	if hook.Event != string(hookEventBeforeValidation) || hook.HookName != "capture_context" || hook.Status != hookStatusSucceeded {
		t.Fatalf("unexpected hook result: %+v", hook)
	}
	if hook.Blocking || hook.FailureAction != hookFailureActionNotApplicable || hook.Scope != hookScopeWorker {
		t.Fatalf("unexpected hook operator fields: %+v", hook)
	}
	if stdout := mustReadFile(t, filepath.Join(tmp, filepath.FromSlash(hook.StdoutPath))); stdout != "hook stdout" {
		t.Fatalf("unexpected hook stdout artifact: %q", stdout)
	}
	if stderr := mustReadFile(t, filepath.Join(tmp, filepath.FromSlash(hook.StderrPath))); stderr != "hook stderr" {
		t.Fatalf("unexpected hook stderr artifact: %q", stderr)
	}
}

func TestExecuteRunRunsNambaOwnedLifecycleHooksInOrder(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "hooks.toml"), `
[hooks.01_before_preflight]
event = "before_preflight"
command = "before_preflight"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = true

[hooks.02_after_preflight]
event = "after_preflight"
command = "after_preflight"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = true

[hooks.03_before_execution]
event = "before_execution"
command = "before_execution"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = true

[hooks.04_after_execution]
event = "after_execution"
command = "after_execution"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = true

[hooks.05_before_validation]
event = "before_validation"
command = "before_validation"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = true

[hooks.06_after_validation]
event = "after_validation"
command = "after_validation"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = true
`)

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
			return "runner output", nil
		case isShellCommand(name):
			return "validation ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}
	var hookCommands []string
	app.runCmdWithInput = func(_ context.Context, name string, args []string, dir, input string) (string, string, error) {
		if !isShellCommand(name) {
			t.Fatalf("unexpected hook command: %s %v", name, args)
		}
		hookCommands = append(hookCommands, args[len(args)-1])
		return "", "", nil
	}

	if err := app.Run(context.Background(), []string{"run", "SPEC-001"}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	want := []string{"before_preflight", "after_preflight", "before_execution", "after_execution", "before_validation", "after_validation"}
	if strings.Join(hookCommands, ",") != strings.Join(want, ",") {
		t.Fatalf("unexpected lifecycle hook order: got %v want %v", hookCommands, want)
	}
}

func TestExecuteRunAdvisoryHookFailureContinues(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "hooks.toml"), `
[hooks.advisory_failure]
event = "before_validation"
command = "exit 7"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = true
`)

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
			return "runner output", nil
		case isShellCommand(name):
			return "validation ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"run", "SPEC-001"}); err != nil {
		t.Fatalf("advisory hook failure should not fail run: %v", err)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "completed" || len(manifest.Hooks) != 1 {
		t.Fatalf("expected completed manifest with one hook, got %+v", manifest)
	}
	hook := manifest.Hooks[0]
	if hook.Status != hookStatusFailed || hook.ExitCode != 7 || hook.Blocking {
		t.Fatalf("expected advisory failed hook result, got %+v", hook)
	}
	if hook.FailureAction != hookFailureActionContinued {
		t.Fatalf("expected continued failure action, got %+v", hook)
	}
}

func TestExecuteRunRecordsAdvisoryHookTimeout(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "hooks.toml"), `
[hooks.slow_hook]
event = "before_validation"
command = "slow"
cwd = "."
timeout = 1
enabled = true
continue_on_failure = true
`)

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
			return "runner output", nil
		case isShellCommand(name):
			return "validation ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}
	app.runCmdWithInput = func(ctx context.Context, name string, args []string, dir, input string) (string, string, error) {
		<-ctx.Done()
		return "", "", ctx.Err()
	}

	if err := app.Run(context.Background(), []string{"run", "SPEC-001"}); err != nil {
		t.Fatalf("timeout advisory hook should not fail run: %v", err)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if len(manifest.Hooks) != 1 {
		t.Fatalf("expected one timeout hook, got %+v", manifest.Hooks)
	}
	hook := manifest.Hooks[0]
	if hook.Status != hookStatusTimeout || hook.ExitCode != -1 || hook.FailureAction != hookFailureActionContinued {
		t.Fatalf("unexpected timeout hook evidence: %+v", hook)
	}
}

func TestExecuteRunBlockingHookStopsBeforeValidationAndRunsOnFailureOnce(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "hooks.toml"), `
[hooks.stop_validation]
event = "before_validation"
command = "exit 9"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = false

[hooks.failure_notice]
event = "on_failure"
command = "echo failure"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = true
`)

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
			return "runner output", nil
		case isShellCommand(name):
			t.Fatal("validation should not run after blocking before_validation hook")
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	err := app.Run(context.Background(), []string{"run", "SPEC-001"})
	if err == nil || !strings.Contains(err.Error(), "blocking hook stop_validation failed") {
		t.Fatalf("expected blocking hook failure, got %v", err)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "hook_failed" || manifest.Validation.State != executionEvidenceStateMissing {
		t.Fatalf("expected hook_failed before validation, got %+v", manifest)
	}
	if len(manifest.Hooks) != 2 {
		t.Fatalf("expected blocking hook and one on_failure hook, got %+v", manifest.Hooks)
	}
	if manifest.Hooks[0].HookName != "stop_validation" || !manifest.Hooks[0].Blocking || manifest.Hooks[0].FailureAction != hookFailureActionStopped {
		t.Fatalf("unexpected blocking hook result: %+v", manifest.Hooks[0])
	}
	if manifest.Hooks[1].Event != string(hookEventOnFailure) || manifest.Hooks[1].HookName != "failure_notice" {
		t.Fatalf("expected one on_failure hook result, got %+v", manifest.Hooks)
	}
}

func TestMalformedHooksConfigStopsBeforePreflightWithConfigErrorEvidence(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "hooks.toml"), `
[hooks.bad_event]
event = "codex_pre_tool_use"
command = "echo bad"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = false
`)

	app.lookPath = func(name string) (string, error) {
		t.Fatalf("preflight should not inspect %s after malformed hook config", name)
		return "", errors.New("unexpected preflight")
	}

	err := app.Run(context.Background(), []string{"run", "SPEC-001"})
	if err == nil || !strings.Contains(err.Error(), "invalid hook event") {
		t.Fatalf("expected hook config error, got %v", err)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "hook_failed" || len(manifest.Hooks) != 1 {
		t.Fatalf("expected hook_failed manifest with config error hook, got %+v", manifest)
	}
	hook := manifest.Hooks[0]
	if hook.HookName != hookConfigErrorName || hook.Status != hookStatusError || hook.ExitCode != -1 || !hook.Blocking {
		t.Fatalf("unexpected config error hook result: %+v", hook)
	}
	if hook.FailureAction != hookFailureActionStopped || !strings.Contains(hook.ErrorSummary, "invalid hook event") {
		t.Fatalf("expected concrete stopped config error, got %+v", hook)
	}
	if manifest.Preflight.State != executionEvidenceStateMissing {
		t.Fatalf("preflight should not have run, got %+v", manifest.Preflight)
	}
}

func TestUnsupportedRunnerObservationsProduceNoFakeHookResults(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "hooks.toml"), `
[hooks.after_bash_only]
event = "after_bash"
command = "echo bash"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = true
`)

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
			return "runner output", nil
		case isShellCommand(name):
			return "validation ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}
	app.runCmdWithInput = func(_ context.Context, name string, args []string, dir, input string) (string, string, error) {
		t.Fatalf("after_bash hook should not run without runner observations: %s %v", name, args)
		return "", "", nil
	}

	if err := app.Run(context.Background(), []string{"run", "SPEC-001"}); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if len(manifest.Hooks) != 0 {
		t.Fatalf("unsupported runner observations must not create fake hook results, got %+v", manifest.Hooks)
	}
}

func TestParallelModeRecordsHooksOnlyInWorkerEvidence(t *testing.T) {
	tmp := canonicalTempDir(t)
	worker := filepath.Join(tmp, ".namba", "worktrees", "spec-001-p1")
	writeTestFile(t, filepath.Join(worker, ".namba", "hooks.toml"), `
[hooks.worker_validation]
event = "before_validation"
command = "worker-hook"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = true
`)

	app := NewApp(&strings.Builder{}, &strings.Builder{})
	app.detectCodexCapabilities = func(context.Context, string, executionRequest) (codexCapabilityMatrix, error) {
		return testCodexCapabilities(), nil
	}
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
			if dir != worker {
				t.Fatalf("expected worker codex dir %s, got %s", worker, dir)
			}
			return "runner output", nil
		case isShellCommand(name):
			if dir != worker {
				t.Fatalf("expected worker validation dir %s, got %s", worker, dir)
			}
			return "validation ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}
	app.runCmdWithInput = func(_ context.Context, name string, args []string, dir, input string) (string, string, error) {
		if dir != worker {
			t.Fatalf("expected worker hook dir %s, got %s", worker, dir)
		}
		var ctx hookExecutionContext
		if err := json.Unmarshal([]byte(input), &ctx); err != nil {
			t.Fatalf("hook context is not JSON: %v", err)
		}
		if ctx.ProjectRoot != worker {
			t.Fatalf("expected hook project_root to remain worker root %s, got %s", worker, ctx.ProjectRoot)
		}
		wantEvidencePath := filepath.ToSlash(filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-p1-evidence.json"))
		if ctx.Artifacts["evidence"] != wantEvidencePath {
			t.Fatalf("expected absolute central evidence path %s, got %+v", wantEvidencePath, ctx.Artifacts)
		}
		return "worker hook", "", nil
	}

	req := app.newExecutionRequest(
		"SPEC-001",
		worker,
		"prompt",
		executionModeParallel,
		suggestDelegationPlan(executionModeParallel, "prompt", "", ""),
		systemConfig{Runner: "codex", ApprovalPolicy: "on-request", SandboxMode: "workspace-write"},
		codexConfig{},
	)
	if _, _, err := app.executeRun(
		context.Background(),
		tmp,
		"spec-001-p1",
		req,
		worker,
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		nil,
		"spec-001-p1",
	); err != nil {
		t.Fatalf("worker executeRun failed: %v", err)
	}

	workerManifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-p1-evidence.json"))
	if len(workerManifest.Hooks) != 1 || workerManifest.Hooks[0].Scope != hookScopeWorker {
		t.Fatalf("expected worker-scoped hook result in worker evidence, got %+v", workerManifest.Hooks)
	}
	if err := app.writeParallelExecutionEvidence(tmp, "SPEC-001", "parallel-run", "completed", false, false); err != nil {
		t.Fatalf("write aggregate evidence failed: %v", err)
	}
	aggregateManifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-parallel-evidence.json"))
	if len(aggregateManifest.Hooks) != 0 {
		t.Fatalf("aggregate parallel evidence must not duplicate worker hooks, got %+v", aggregateManifest.Hooks)
	}
}

func TestRunnerObservationSinkEmitsToolBoundaryHook(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell observation command uses portable App test seam elsewhere; direct sink keeps this focused on dispatch")
	}
	tmp := canonicalTempDir(t)
	app := NewApp(&strings.Builder{}, &strings.Builder{})
	writeTestFile(t, filepath.Join(tmp, ".namba", "hooks.toml"), `
[hooks.log_shell]
event = "after_bash"
command = "echo observed"
cwd = "."
timeout = 5
enabled = true
continue_on_failure = true
`)
	req := executionRequest{SpecID: "SPEC-001", WorkDir: tmp, Mode: executionModeDefault}
	lifecycle := newHookLifecycle(app, tmp, "spec-001", req, "")

	if err := lifecycle.ObserveRunnerTool(context.Background(), runnerObservation{
		ObservationType: runnerObservationBashCompleted,
		ToolName:        "Bash",
		Command:         "go test ./...",
		CWD:             tmp,
		ExitCode:        0,
		Status:          "succeeded",
	}); err != nil {
		t.Fatalf("ObserveRunnerTool failed: %v", err)
	}
	if err := lifecycle.writeRunEvidence(context.Background(), "completed", 0, false); err != nil {
		t.Fatalf("write evidence failed: %v", err)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if len(manifest.Hooks) != 1 || manifest.Hooks[0].Event != string(hookEventAfterBash) {
		t.Fatalf("expected after_bash hook evidence from observation, got %+v", manifest.Hooks)
	}
}
