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
	"sync"
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
			mustContainArgs(t, args, []string{"-a", "on-request", "-s", "workspace-write"})
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
	if result.ExecutionMode != "default" {
		t.Fatalf("expected default execution mode, got %+v", result)
	}
	if result.ApprovalPolicy != "on-request" || result.SandboxMode != "workspace-write" {
		t.Fatalf("unexpected runtime modes: %+v", result)
	}
	if result.DelegationObserved {
		t.Fatalf("expected standalone runner logs to record plan, not observed delegation: %+v", result)
	}
	if result.DelegationSummary == "" || result.DelegationPlan.IntegratorRole != "standalone-runner" {
		t.Fatalf("expected delegation summary and integrator role in execution result: %+v", result)
	}

	requestJSON := mustReadExecutionRequest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-request.json"))
	if requestJSON.DelegationPlan.IntegratorRole != "standalone-runner" {
		t.Fatalf("expected request json to persist delegation plan: %+v", requestJSON)
	}

	request := mustReadFile(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-request.md"))
	if !strings.Contains(request, "- Mode: default") || !strings.Contains(request, "## Delegation Heuristics") || !strings.Contains(request, "Default mode keeps work inside the standalone runner") {
		t.Fatalf("expected default mode prompt guidance with delegation heuristics, got %q", request)
	}

	report := mustReadValidationReport(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-validation.json"))
	if !report.Passed {
		t.Fatalf("expected successful validation report: %+v", report)
	}
	if len(report.Steps) != 3 {
		t.Fatalf("expected 3 validation steps, got %d", len(report.Steps))
	}
}

func TestBuildExecutionPromptIncludesModeGuidance(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	specPkg, err := app.loadSpec(tmp, "SPEC-001")
	if err != nil {
		t.Fatalf("load spec: %v", err)
	}
	qualityCfg, err := app.loadQualityConfig(tmp)
	if err != nil {
		t.Fatalf("load quality config: %v", err)
	}

	tests := []struct {
		name string
		mode executionMode
		want []string
	}{
		{
			name: "default",
			mode: executionModeDefault,
			want: []string{"- Mode: default", "standard standalone Codex run in one workspace", "Default mode keeps work inside the standalone runner"},
		},
		{
			name: "solo",
			mode: executionModeSolo,
			want: []string{"- Mode: solo", "explicit single-subagent workflow", "Use at most one delegated specialist"},
		},
		{
			name: "team",
			mode: executionModeTeam,
			want: []string{"- Mode: team", "explicit multi-subagent coordination", "Prefer one specialist when one domain dominates"},
		},
		{
			name: "parallel",
			mode: executionModeParallel,
			want: []string{"- Mode: parallel", "Namba worktree parallel mode", "not Codex subagent orchestration"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, _, plan, err := app.buildExecutionPrompt(tmp, specPkg, qualityCfg, tt.mode)
			if err != nil {
				t.Fatalf("buildExecutionPrompt failed: %v", err)
			}
			for _, want := range tt.want {
				if !strings.Contains(prompt, want) {
					t.Fatalf("expected prompt to contain %q, got %q", want, prompt)
				}
			}
			if !strings.Contains(prompt, "## Delegation Heuristics") || plan.IntegratorRole != "standalone-runner" {
				t.Fatalf("expected delegation heuristics and integrator role, got prompt=%q plan=%+v", prompt, plan)
			}
		})
	}
}

func TestSuggestDelegationPlanRoutesSpecialists(t *testing.T) {
	teamPlan := suggestDelegationPlan(
		executionModeTeam,
		"Implement a mobile settings screen that stores auth tokens securely.",
		"Update the mobile app layout, tighten token handling, and validate the acceptance path.",
		`- [ ] Ship the mobile UI
- [ ] Harden auth token storage
- [ ] Add regression coverage`,
	)
	for _, want := range []string{"mobile", "security", "quality"} {
		if !strings.Contains(strings.Join(teamPlan.DominantDomains, ","), want) {
			t.Fatalf("expected dominant domains to include %q, got %+v", want, teamPlan)
		}
	}
	for _, want := range []string{"namba-mobile-engineer", "namba-security-engineer", "namba-reviewer"} {
		if !strings.Contains(strings.Join(teamPlan.SelectedRoles, ","), want) {
			t.Fatalf("expected selected roles to include %q, got %+v", want, teamPlan)
		}
	}
	for role, want := range map[string]agentRuntimeProfile{
		"namba-mobile-engineer":   {Role: "namba-mobile-engineer", Model: "gpt-5.4", ModelReasoningEffort: "medium"},
		"namba-security-engineer": {Role: "namba-security-engineer", Model: "gpt-5.4", ModelReasoningEffort: "high"},
		"namba-reviewer":          {Role: "namba-reviewer", Model: "gpt-5.4", ModelReasoningEffort: "high"},
	} {
		found := false
		for _, profile := range teamPlan.SelectedRoleProfiles {
			if profile.Role == role {
				found = true
				if profile.Model != want.Model || profile.ModelReasoningEffort != want.ModelReasoningEffort {
					t.Fatalf("unexpected runtime profile for %s: %+v", role, profile)
				}
			}
		}
		if !found {
			t.Fatalf("expected runtime profile for %s, got %+v", role, teamPlan.SelectedRoleProfiles)
		}
	}
	if prompt := strings.Join(formatDelegationPlanPrompt(teamPlan), "\n"); !strings.Contains(prompt, "model_reasoning_effort `high`") || !strings.Contains(prompt, "`namba-mobile-engineer` -> model `gpt-5.4`") {
		t.Fatalf("expected team prompt to include role runtime metadata, got %q", prompt)
	}
	if teamPlan.DelegationBudget < 2 {
		t.Fatalf("expected team delegation budget to allow multiple specialists, got %+v", teamPlan)
	}

	soloPlan := suggestDelegationPlan(
		executionModeSolo,
		"Implement responsive UI filters and browser accessibility states.",
		"Update the screen component and responsive layout.",
		"- [ ] Ship the UI updates",
	)
	if len(soloPlan.SelectedRoles) != 1 || soloPlan.SelectedRoles[0] != "namba-frontend-implementer" {
		t.Fatalf("expected solo plan to choose one frontend specialist, got %+v", soloPlan)
	}
	if soloPlan.DelegationBudget != 1 {
		t.Fatalf("expected solo delegation budget 1, got %+v", soloPlan)
	}
}

func TestRunExecutesExplicitSubagentModes(t *testing.T) {
	tests := []struct {
		name string
		flag string
		mode string
	}{
		{name: "solo", flag: "--solo", mode: "solo"},
		{name: "team", flag: "--team", mode: "team"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp, app, restore := prepareExecutionProject(t)
			defer restore()

			var promptArg string
			app.lookPath = func(name string) (string, error) {
				if name == "codex" {
					return name, nil
				}
				return "", errors.New("missing dependency")
			}
			app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
				switch {
				case isCodexExec(name, args):
					promptArg = args[len(args)-1]
					return "runner output", nil
				case isShellCommand(name):
					return "validation ok", nil
				default:
					t.Fatalf("unexpected command: %s %v", name, args)
					return "", nil
				}
			}

			if err := app.Run(context.Background(), []string{"run", "SPEC-001", tt.flag}); err != nil {
				t.Fatalf("run failed: %v", err)
			}
			if !strings.Contains(promptArg, "- Mode: "+tt.mode) {
				t.Fatalf("expected prompt to include mode %s, got %q", tt.mode, promptArg)
			}

			result := mustReadExecutionResult(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-execution.json"))
			if result.ExecutionMode != tt.mode {
				t.Fatalf("expected execution mode %s, got %+v", tt.mode, result)
			}
		})
	}
}

func TestRunRejectsConflictingExecutionModes(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "solo and team", args: []string{"run", "SPEC-001", "--solo", "--team"}, want: "--solo, --team"},
		{name: "solo and parallel", args: []string{"run", "SPEC-001", "--solo", "--parallel"}, want: "--solo, --parallel"},
		{name: "team and parallel", args: []string{"run", "SPEC-001", "--team", "--parallel"}, want: "--team, --parallel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, app, restore := prepareExecutionProject(t)
			defer restore()

			err := app.Run(context.Background(), tt.args)
			if err == nil {
				t.Fatal("expected conflicting mode error")
			}
			if !strings.Contains(err.Error(), "invalid flag combination") || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRunUsesConfiguredApprovalAndSandbox(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "system.yaml"), "runner: codex\napproval_policy: never\nsandbox_mode: read-only\n")

	app.lookPath = func(name string) (string, error) {
		if name == "codex" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexExec(name, args):
			mustContainArgs(t, args, []string{"-a", "never", "-s", "read-only"})
			return "runner output", nil
		case isShellCommand(name):
			return "validation ok", nil
		default:
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"run", "SPEC-001"}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	result := mustReadExecutionResult(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-execution.json"))
	if result.ApprovalPolicy != "never" || result.SandboxMode != "read-only" {
		t.Fatalf("unexpected runtime modes: %+v", result)
	}
}

func TestRunRejectsUnsupportedRunner(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "system.yaml"), "runner: unsupported\napproval_policy: on-request\nsandbox_mode: workspace-write\n")

	err := app.Run(context.Background(), []string{"run", "SPEC-001"})
	if err == nil {
		t.Fatal("expected unsupported runner error")
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunRejectsInvalidApprovalPolicy(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "system.yaml"), "runner: codex\napproval_policy: maybe\nsandbox_mode: workspace-write\n")

	app.lookPath = func(name string) (string, error) {
		if name == "codex" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if isShellCommand(name) {
			t.Fatal("validators should not run when approval policy is invalid")
		}
		if isCodexExec(name, args) {
			t.Fatal("codex should not run when approval policy is invalid")
		}
		return "", nil
	}

	err := app.Run(context.Background(), []string{"run", "SPEC-001"})
	if err == nil {
		t.Fatal("expected invalid approval policy error")
	}
	if !strings.Contains(err.Error(), "approval_policy") {
		t.Fatalf("unexpected error: %v", err)
	}

	result := mustReadExecutionResult(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-execution.json"))
	if !strings.Contains(result.Error, "approval_policy") {
		t.Fatalf("expected approval policy error in execution result: %+v", result)
	}
}

func TestRunRejectsInvalidSandboxMode(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "system.yaml"), "runner: codex\napproval_policy: on-request\nsandbox_mode: moon-write\n")

	app.lookPath = func(name string) (string, error) {
		if name == "codex" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if isShellCommand(name) {
			t.Fatal("validators should not run when sandbox mode is invalid")
		}
		if isCodexExec(name, args) {
			t.Fatal("codex should not run when sandbox mode is invalid")
		}
		return "", nil
	}

	err := app.Run(context.Background(), []string{"run", "SPEC-001"})
	if err == nil {
		t.Fatal("expected invalid sandbox mode error")
	}
	if !strings.Contains(err.Error(), "sandbox_mode") {
		t.Fatalf("unexpected error: %v", err)
	}

	result := mustReadExecutionResult(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-execution.json"))
	if !strings.Contains(result.Error, "sandbox_mode") {
		t.Fatalf("expected sandbox mode error in execution result: %+v", result)
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
			mustContainArgs(t, args, []string{"-a", "on-request", "-s", "workspace-write"})
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
			mustContainArgs(t, args, []string{"-a", "on-request", "-s", "workspace-write"})
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

var cwdMu sync.Mutex

func chdirExecution(t *testing.T, dir string) func() {
	t.Helper()
	cwdMu.Lock()

	previous, err := os.Getwd()
	if err != nil {
		cwdMu.Unlock()
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		cwdMu.Unlock()
		t.Fatalf("chdir %s: %v", dir, err)
	}
	return func() {
		defer cwdMu.Unlock()
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}
}
func isCodexExec(name string, args []string) bool {
	if runtime.GOOS == "windows" {
		return name == "cmd" && len(args) >= 7 && args[0] == "/c" && args[1] == "codex" && args[2] == "exec"
	}
	return name == "codex" && len(args) >= 6 && args[0] == "exec"
}

func isShellCommand(name string) bool {
	return name == "powershell" || name == "sh"
}

func mustContainArgs(t *testing.T, args []string, expected []string) {
	t.Helper()
	for i := 0; i < len(expected); i += 2 {
		index := indexOfArg(args, expected[i])
		if index == -1 || index+1 >= len(args) || args[index+1] != expected[i+1] {
			t.Fatalf("expected args to contain %s %s, got %v", expected[i], expected[i+1], args)
		}
	}
}

func indexOfArg(args []string, needle string) int {
	for i, arg := range args {
		if arg == needle {
			return i
		}
	}
	return -1
}

func mustReadExecutionRequest(t *testing.T, path string) executionRequest {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read execution request: %v", err)
	}
	var request executionRequest
	if err := json.Unmarshal(data, &request); err != nil {
		t.Fatalf("unmarshal execution request: %v", err)
	}
	return request
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
