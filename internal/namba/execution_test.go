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
			mustContainArgs(t, args, []string{"-c", `approval_policy="on-request"`, "-s", "workspace-write"})
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
	if result.SessionMode != "stateful" || result.SessionID == "" {
		t.Fatalf("expected session metadata in execution result: %+v", result)
	}

	requestJSON := mustReadExecutionRequest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-request.json"))
	if requestJSON.DelegationPlan.IntegratorRole != "standalone-runner" {
		t.Fatalf("expected request json to persist delegation plan: %+v", requestJSON)
	}
	if requestJSON.SessionMode != "stateful" {
		t.Fatalf("expected request json to persist session mode: %+v", requestJSON)
	}

	request := mustReadFile(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-request.md"))
	if !strings.Contains(request, "- Mode: default") || !strings.Contains(request, "## Delegation Heuristics") || !strings.Contains(request, "Default mode keeps work inside the standalone runner") {
		t.Fatalf("expected default mode prompt guidance with delegation heuristics, got %q", request)
	}

	report := mustReadValidationReport(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-validation.json"))
	if !report.Passed {
		t.Fatalf("expected successful validation report: %+v", report)
	}
	if len(report.Steps) != 7 {
		t.Fatalf("expected 7 validation pipeline steps, got %d", len(report.Steps))
	}
}

func TestBuildCodexExecArgsSupportsFallbacksAndResumeSurface(t *testing.T) {
	tests := []struct {
		name string
		req  executionRequest
		caps codexCapabilityMatrix
		want []string
	}{
		{
			name: "exec falls back to config overrides",
			req: executionRequest{
				ApprovalPolicy: "on-request",
				SandboxMode:    "workspace-write",
				Model:          "gpt-5.4",
				Profile:        "namba",
				WebSearch:      true,
				AddDirs:        []string{"extra"},
				SessionMode:    "stateful",
				Prompt:         "ship it",
			},
			caps: codexCapabilityMatrix{
				Exec: codexCommandCapabilities{Config: true, SandboxFlag: true, ModelFlag: true, ProfileFlag: true, AddDirFlag: true},
			},
			want: []string{"exec", "-c", `approval_policy="on-request"`, "-s", "workspace-write", "-m", "gpt-5.4", "-p", "namba", "-c", `web_search="live"`, "--add-dir", "extra", "ship it"},
		},
		{
			name: "resume allows exec-level flags before resume",
			req: executionRequest{
				ApprovalPolicy: "never",
				SandboxMode:    "workspace-write",
				Model:          "gpt-5.4",
				Profile:        "namba",
				WebSearch:      true,
				AddDirs:        []string{`C:\extra`},
				SessionMode:    "stateful",
				ResumeSession:  true,
				Prompt:         "continue",
			},
			caps: codexCapabilityMatrix{
				Exec:   codexCommandCapabilities{Config: true, SandboxFlag: true, ModelFlag: true, ProfileFlag: true, AddDirFlag: true},
				Resume: codexCommandCapabilities{Config: true, ModelFlag: true},
			},
			want: []string{"exec", "-s", "workspace-write", "-m", "gpt-5.4", "-p", "namba", "--add-dir", `C:\extra`, "resume", "--last", "-c", `approval_policy="never"`, "-c", `web_search="live"`, "continue"},
		},
		{
			name: "resume uses resume-specific config fallbacks",
			req: executionRequest{
				ApprovalPolicy: "never",
				SandboxMode:    "workspace-write",
				Model:          "gpt-5.4",
				WebSearch:      true,
				AddDirs:        []string{`C:\extra`},
				SessionMode:    "stateful",
				ResumeSession:  true,
				Prompt:         "continue",
			},
			caps: codexCapabilityMatrix{
				Resume: codexCommandCapabilities{Config: true, ModelFlag: true},
			},
			want: []string{"exec", "resume", "--last", "-c", `approval_policy="never"`, "-c", `sandbox_mode="workspace-write"`, "-m", "gpt-5.4", "-c", `web_search="live"`, "-c", `sandbox_workspace_write.writable_roots=["C:\\extra"]`, "continue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := buildCodexExecArgs(tt.req, tt.caps)
			if err != nil {
				t.Fatalf("buildCodexExecArgs failed: %v", err)
			}
			if strings.Join(args, "\x00") != strings.Join(tt.want, "\x00") {
				t.Fatalf("unexpected args: got %v want %v", args, tt.want)
			}
		})
	}
}

func TestProbeCodexCapabilitiesSkipsResumeHelpWhenResumeIsNotPlanned(t *testing.T) {
	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	app.lookPath = func(name string) (string, error) {
		if name == "codex" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}

	resumeHelpCalls := 0
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if dir != tmp {
			t.Fatalf("expected capability probe workdir %s, got %s", tmp, dir)
		}
		switch {
		case isCodexVersionCommand(name, args):
			return "codex-cli test", nil
		case isCodexHelpCommand(name, args, false):
			return "-c, --config\n-s, --sandbox\n-m, --model\n-p, --profile\n--add-dir", nil
		case isCodexHelpCommand(name, args, true):
			resumeHelpCalls++
			return "-c, --config\n-m, --model", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	caps, err := app.probeCodexCapabilities(context.Background(), tmp, executionRequest{
		WorkDir:        tmp,
		Prompt:         "ship it",
		ApprovalPolicy: "on-request",
		SandboxMode:    "workspace-write",
		SessionMode:    "stateful",
		RepairAttempts: 0,
		ResumeSession:  false,
		Mode:           executionModeDefault,
	})
	if err != nil {
		t.Fatalf("probeCodexCapabilities failed: %v", err)
	}
	if resumeHelpCalls != 0 {
		t.Fatalf("expected no resume help probe, got %d", resumeHelpCalls)
	}
	if caps.Resume != (codexCommandCapabilities{}) {
		t.Fatalf("expected empty resume capabilities when no resume is planned, got %+v", caps.Resume)
	}
}

func TestProbeCodexCapabilitiesIncludesResumeHelpWhenResumeIsPlanned(t *testing.T) {
	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	app.lookPath = func(name string) (string, error) {
		if name == "codex" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}

	resumeHelpCalls := 0
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if dir != tmp {
			t.Fatalf("expected capability probe workdir %s, got %s", tmp, dir)
		}
		switch {
		case isCodexVersionCommand(name, args):
			return "codex-cli test", nil
		case isCodexHelpCommand(name, args, false):
			return "-c, --config\n-s, --sandbox\n-m, --model\n-p, --profile\n--add-dir", nil
		case isCodexHelpCommand(name, args, true):
			resumeHelpCalls++
			return "-c, --config\n-m, --model", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	caps, err := app.probeCodexCapabilities(context.Background(), tmp, executionRequest{
		WorkDir:        tmp,
		Prompt:         "ship it",
		ApprovalPolicy: "on-request",
		SandboxMode:    "workspace-write",
		SessionMode:    "stateful",
		RepairAttempts: 1,
		Mode:           executionModeDefault,
	})
	if err != nil {
		t.Fatalf("probeCodexCapabilities failed: %v", err)
	}
	if resumeHelpCalls != 1 {
		t.Fatalf("expected one resume help probe, got %d", resumeHelpCalls)
	}
	if !caps.Resume.Config || !caps.Resume.ModelFlag {
		t.Fatalf("expected populated resume capabilities, got %+v", caps.Resume)
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
			want: []string{"- Mode: solo", "one runner in one workspace", "same-workspace team orchestration"},
		},
		{
			name: "team",
			mode: executionModeTeam,
			want: []string{"- Mode: team", "same-workspace multi-agent execution", "Role runtime profiles should materially affect the actual Codex turns"},
		},
		{
			name: "parallel",
			mode: executionModeParallel,
			want: []string{"- Mode: parallel", "Namba worktree parallel mode", "not same-workspace team orchestration"},
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
			if !strings.Contains(prompt, "## Delegation Heuristics") || plan.IntegratorRole == "" {
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

	architectPlan := suggestDelegationPlan(
		executionModeSolo,
		"Plan the component/state split for this dashboard.",
		"Clarify component boundaries and state ownership before editing files.",
		"- [ ] Produce the frontend plan",
	)
	if len(architectPlan.SelectedRoles) != 1 || architectPlan.SelectedRoles[0] != "namba-frontend-architect" {
		t.Fatalf("expected solo planning prompt to choose the frontend architect, got %+v", architectPlan)
	}

	designPlan := suggestDelegationPlan(
		executionModeSolo,
		"Redesign the landing page hero art direction and palette so it stops feeling generic.",
		"Clarify the visual direction, composition, and motion intent before implementation.",
		"- [ ] Produce the design direction",
	)
	if len(designPlan.SelectedRoles) != 1 || designPlan.SelectedRoles[0] != "namba-designer" {
		t.Fatalf("expected solo plan to choose the designer for art-direction work, got %+v", designPlan)
	}

	implementationPlan := suggestDelegationPlan(
		executionModeSolo,
		"Implement split-screen dashboard state updates and browser accessibility fixes.",
		"Wire the UI changes into the existing dashboard component.",
		"- [ ] Ship the UI updates",
	)
	if len(implementationPlan.SelectedRoles) != 1 || implementationPlan.SelectedRoles[0] != "namba-frontend-implementer" {
		t.Fatalf("expected implementation prompt to stay with the frontend implementer, got %+v", implementationPlan)
	}

	milestonePlan := suggestDelegationPlan(
		executionModeTeam,
		"Plan the page milestone rollout for responsive browser accessibility work.",
		"Keep the milestone focused on UI delivery sequencing.",
		"- [ ] Ship the responsive page",
	)
	if strings.Contains(strings.Join(milestonePlan.DominantDomains, ","), "design") {
		t.Fatalf("expected milestone-only prompt to avoid design routing noise, got %+v", milestonePlan)
	}
	if strings.Contains(strings.Join(milestonePlan.SelectedRoles, ","), "namba-designer") {
		t.Fatalf("expected milestone-only prompt to avoid selecting the designer, got %+v", milestonePlan)
	}

	backendStatePlan := suggestDelegationPlan(
		executionModeTeam,
		"Implement backend state transitions for the workflow engine.",
		"Update the server-side state machine and controller flow.",
		"- [ ] Ship the backend transition changes",
	)
	if strings.Contains(strings.Join(backendStatePlan.DominantDomains, ","), "frontend") {
		t.Fatalf("expected backend state-transition prompt to avoid frontend routing noise, got %+v", backendStatePlan)
	}
	if strings.Contains(strings.Join(backendStatePlan.SelectedRoles, ","), "namba-frontend-implementer") || strings.Contains(strings.Join(backendStatePlan.SelectedRoles, ","), "namba-frontend-architect") {
		t.Fatalf("expected backend state-transition prompt to avoid frontend specialists, got %+v", backendStatePlan)
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
				if name == "codex" || name == "git" {
					return name, nil
				}
				return "", errors.New("missing dependency")
			}
			app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
				switch {
				case isCodexExec(name, args):
					if promptArg == "" && strings.Contains(args[len(args)-1], "- Mode: ") {
						promptArg = args[len(args)-1]
					}
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

func TestLoadExecutionRuntimeConfigLoadsQualitySystemAndCodex(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "quality.yaml"), "development_mode: tdd\ntest_command: test\nlint_command: lint\ntypecheck_command: vet\n")
	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "system.yaml"), "runner: codex\napproval_policy: never\nsandbox_mode: read-only\n")
	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "codex.yaml"), "profile: namba\nsession_mode: stateful\nrepair_attempts: 2\n")

	runtimeCfg, err := app.loadExecutionRuntimeConfig(tmp)
	if err != nil {
		t.Fatalf("loadExecutionRuntimeConfig failed: %v", err)
	}

	if runtimeCfg.QualityCfg.TestCommand != "test" || runtimeCfg.QualityCfg.LintCommand != "lint" || runtimeCfg.QualityCfg.TypecheckCommand != "vet" {
		t.Fatalf("unexpected quality config: %+v", runtimeCfg.QualityCfg)
	}
	if runtimeCfg.SystemCfg.Runner != "codex" || runtimeCfg.SystemCfg.ApprovalPolicy != "never" || runtimeCfg.SystemCfg.SandboxMode != "read-only" {
		t.Fatalf("unexpected system config: %+v", runtimeCfg.SystemCfg)
	}
	if runtimeCfg.CodexCfg.Profile != "namba" || runtimeCfg.CodexCfg.SessionMode != "stateful" || runtimeCfg.CodexCfg.RepairAttempts != 2 {
		t.Fatalf("unexpected codex config: %+v", runtimeCfg.CodexCfg)
	}
}

func TestWriteExecutionPromptCreatesParentDirAndWritesPrompt(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	path := filepath.Join(tmp, ".namba", "logs", "runs", "custom-request.md")
	if err := app.writeExecutionPrompt(path, "hello prompt"); err != nil {
		t.Fatalf("writeExecutionPrompt failed: %v", err)
	}

	if got := mustReadFile(t, path); got != "hello prompt" {
		t.Fatalf("expected prompt file to be written, got %q", got)
	}
}

func TestDirectFixRepairContractLinesCaptureRepairRules(t *testing.T) {
	got := strings.Join(directFixRepairContractLines(), "\n")

	for _, want := range []string{
		"Inspect the relevant repository files plus `.namba/config/sections/*.yaml` and `.namba/project/*` context before editing.",
		"Finish by running `namba sync` in the same workspace after validation passes.",
		"Do not create or mutate `.namba/specs/<SPEC>` as part of this direct repair flow.",
		`For bugfix SPEC scaffolding, use ` + "`namba fix --command plan \"<issue description>\"`" + ".",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected repair-contract lines to contain %q, got %q", want, got)
		}
	}
}

func TestDirectFixProjectContextPromptLinesUseUnknownFallbacks(t *testing.T) {
	lines := directFixProjectContextPromptLines(projectConfig{}, qualityConfig{})
	got := strings.Join(lines, "\n")

	for _, want := range []string{
		"## Project Context",
		"- Project: unknown",
		"- Project type: unknown",
		"- Language: unknown",
		"- Framework: unknown",
		"- Development mode: unknown",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected project-context lines to contain %q, got %q", want, got)
		}
	}
}

func TestDirectFixValidationPromptLinesSkipBlankCommands(t *testing.T) {
	lines := directFixValidationPromptLines(qualityConfig{
		TestCommand:            "go test ./...",
		LintCommand:            "none",
		TypecheckCommand:       "",
		BuildCommand:           "go build ./...",
		MigrationDryRunCommand: "none",
		SmokeStartCommand:      "",
		OutputContractCommand:  "python validate.py",
	})
	got := strings.Join(lines, "\n")

	for _, want := range []string{
		"## Validation",
		"- test: go test ./...",
		"- build: go build ./...",
		"- output-contract: python validate.py",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected validation lines to contain %q, got %q", want, got)
		}
	}
	for _, unwanted := range []string{"- lint:", "- typecheck:", "- migration-dry-run:", "- smoke-start:"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("expected validation lines to skip %q, got %q", unwanted, got)
		}
	}
}

func TestBuildDirectFixPromptPreservesSectionOrder(t *testing.T) {
	prompt, delegation := buildDirectFixPrompt("/tmp/demo", "startup panic", projectConfig{Name: "demo"}, qualityConfig{
		DevelopmentMode: "tdd",
		TestCommand:     "go test ./...",
	})
	if delegation.IntegratorRole == "" {
		t.Fatalf("expected delegation plan, got %+v", delegation)
	}

	sections := []string{
		"## Issue",
		"## Repair Contract",
		"## Project Context",
		"## Delegation Heuristics",
		"## Validation",
	}
	last := -1
	for _, section := range sections {
		idx := strings.Index(prompt, section)
		if idx == -1 {
			t.Fatalf("expected prompt to contain %q, got %q", section, prompt)
		}
		if idx <= last {
			t.Fatalf("expected section %q to appear after prior sections in prompt %q", section, prompt)
		}
		last = idx
	}
	if !strings.Contains(prompt, "Project root: /tmp/demo") {
		t.Fatalf("expected prompt to include project root, got %q", prompt)
	}
}

func TestLoadDirectFixExecutionContextBuildsPromptAndPromptPath(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	fixCtx, err := app.loadDirectFixExecutionContext(tmp, "startup panic")
	if err != nil {
		t.Fatalf("loadDirectFixExecutionContext failed: %v", err)
	}

	if fixCtx.Root != tmp || fixCtx.Description != "startup panic" {
		t.Fatalf("unexpected direct-fix context: %+v", fixCtx)
	}
	if fixCtx.LogID != "direct-fix" || fixCtx.PromptPath != filepath.Join(tmp, ".namba", "logs", "runs", "direct-fix-request.md") {
		t.Fatalf("unexpected direct-fix logging fields: %+v", fixCtx)
	}
	if !strings.Contains(fixCtx.Prompt, "# NambaAI Direct Repair Request") || !strings.Contains(fixCtx.Prompt, "startup panic") {
		t.Fatalf("expected direct-fix prompt content, got %q", fixCtx.Prompt)
	}
	if fixCtx.Delegation.IntegratorRole != "standalone-runner" {
		t.Fatalf("expected default direct-fix delegation role, got %+v", fixCtx.Delegation)
	}
}

func TestMaterializeDirectFixExecutionPromptWritesMarkdownRequest(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	fixCtx, err := app.loadDirectFixExecutionContext(tmp, "startup panic")
	if err != nil {
		t.Fatalf("loadDirectFixExecutionContext failed: %v", err)
	}
	if err := app.materializeDirectFixExecutionPrompt(fixCtx); err != nil {
		t.Fatalf("materializeDirectFixExecutionPrompt failed: %v", err)
	}

	written := mustReadFile(t, fixCtx.PromptPath)
	if !strings.Contains(written, "# NambaAI Direct Repair Request") || !strings.Contains(written, "## Repair Contract") {
		t.Fatalf("expected direct-fix request markdown to be written, got %q", written)
	}
}

func TestDispatchDirectFixExecutionRunsAndSyncs(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	stdout := &bytes.Buffer{}
	app.stdout = stdout
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
			return "repair output", nil
		case isShellCommand(name):
			return "validation ok", nil
		default:
			t.Fatalf("unexpected command during direct-fix dispatch: %s %v", name, args)
			return "", nil
		}
	}

	fixCtx, err := app.loadDirectFixExecutionContext(tmp, "startup panic")
	if err != nil {
		t.Fatalf("loadDirectFixExecutionContext failed: %v", err)
	}
	if err := app.materializeDirectFixExecutionPrompt(fixCtx); err != nil {
		t.Fatalf("materializeDirectFixExecutionPrompt failed: %v", err)
	}
	if err := app.dispatchDirectFixExecution(context.Background(), fixCtx); err != nil {
		t.Fatalf("dispatchDirectFixExecution failed: %v", err)
	}

	request := mustReadExecutionRequest(t, filepath.Join(tmp, ".namba", "logs", "runs", "direct-fix-request.json"))
	if request.SpecID != "DIRECT-FIX" || request.Mode != executionModeDefault || request.TurnName != "direct-fix" {
		t.Fatalf("unexpected direct-fix request metadata: %+v", request)
	}
	if request.TurnRole != fixCtx.Delegation.IntegratorRole {
		t.Fatalf("expected direct-fix turn role to follow delegation, got %+v", request)
	}
	result := mustReadExecutionResult(t, filepath.Join(tmp, ".namba", "logs", "runs", "direct-fix-execution.json"))
	if !result.Succeeded || result.SpecID != "DIRECT-FIX" {
		t.Fatalf("unexpected direct-fix result: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".namba", "project", "change-summary.md")); err != nil {
		t.Fatalf("expected sync outputs after direct-fix dispatch, stat err=%v", err)
	}
	if !strings.Contains(stdout.String(), "Executed direct fix with codex") {
		t.Fatalf("expected direct-fix dispatch output, got %q", stdout.String())
	}
}

func TestLoadRunExecutionContextBuildsPromptTasksAndPromptPath(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	runCtx, err := app.loadRunExecutionContext(tmp, runExecuteOptions{specID: "SPEC-001", mode: executionModeTeam})
	if err != nil {
		t.Fatalf("loadRunExecutionContext failed: %v", err)
	}

	if runCtx.Root != tmp || runCtx.SpecPkg.ID != "SPEC-001" {
		t.Fatalf("expected run context to target SPEC-001 in %s, got %+v", tmp, runCtx)
	}
	if runCtx.PromptPath != filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-request.md") {
		t.Fatalf("unexpected prompt path: %+v", runCtx)
	}
	if len(runCtx.Tasks) == 0 {
		t.Fatalf("expected acceptance tasks to be loaded, got %+v", runCtx)
	}
	if !strings.Contains(runCtx.Prompt, "- Mode: team") {
		t.Fatalf("expected team mode prompt guidance, got %q", runCtx.Prompt)
	}
	if runCtx.Delegation.IntegratorRole != "same-workspace-integrator" {
		t.Fatalf("expected team-mode integrator role, got %+v", runCtx.Delegation)
	}
	if runCtx.SystemCfg.Runner != "codex" || runCtx.WorkflowCfg.MaxParallelWorkers < 1 {
		t.Fatalf("expected configs to be loaded into run context, got %+v", runCtx)
	}
}

func TestMaterializeRunExecutionPromptWritesMarkdownRequest(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	runCtx, err := app.loadRunExecutionContext(tmp, runExecuteOptions{specID: "SPEC-001", mode: executionModeDefault})
	if err != nil {
		t.Fatalf("loadRunExecutionContext failed: %v", err)
	}
	if err := app.materializeRunExecutionPrompt(runCtx); err != nil {
		t.Fatalf("materializeRunExecutionPrompt failed: %v", err)
	}

	written := mustReadFile(t, runCtx.PromptPath)
	if !strings.Contains(written, "# NambaAI Execution Request") || !strings.Contains(written, "## Validation") {
		t.Fatalf("expected prompt request markdown to be written, got %q", written)
	}
}

func TestDispatchRunExecutionDryRunSkipsRunnerAndPrintsPromptPath(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	stdout := &bytes.Buffer{}
	app.stdout = stdout
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		t.Fatalf("unexpected command during dry-run dispatch: %s %v", name, args)
		return "", nil
	}

	options := runExecuteOptions{specID: "SPEC-001", mode: executionModeDefault, dryRun: true}
	runCtx, err := app.loadRunExecutionContext(tmp, options)
	if err != nil {
		t.Fatalf("loadRunExecutionContext failed: %v", err)
	}
	if err := app.materializeRunExecutionPrompt(runCtx); err != nil {
		t.Fatalf("materializeRunExecutionPrompt failed: %v", err)
	}
	if err := app.dispatchRunExecution(context.Background(), options, runCtx); err != nil {
		t.Fatalf("dispatchRunExecution failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "Prepared execution request at "+runCtx.PromptPath) {
		t.Fatalf("expected dry-run dispatch message, got %q", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-execution.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected dry-run dispatch to avoid execution artifacts, got err=%v", err)
	}
}

func TestRunBlocksFrontendMajorWhenFrontendSynthesisIsIncomplete(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", frontendBriefFileName), strings.Join([]string{
		"# Frontend Brief",
		"",
		"Task Classification: frontend-major",
		"Classification Rationale: Dashboard restructure.",
		"Frontend Gate Status: needs-research",
		"Problem Gate: complete",
		"Reference Gate: missing",
		"Critique Gate: missing",
		"Decision Gate: missing",
		"Prototype Gate: missing",
		"Prototype Evidence: n/a",
	}, "\n"))

	app.lookPath = func(name string) (string, error) {
		if name == "codex" || name == "git" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		t.Fatalf("runner/validators should not execute for blocked frontend-major run: %s %v", name, args)
		return "", nil
	}

	err := app.Run(context.Background(), []string{"run", "SPEC-001"})
	if err == nil {
		t.Fatal("expected frontend synthesis block")
	}
	if !strings.Contains(err.Error(), "blocked for frontend synthesis") {
		t.Fatalf("expected frontend block message, got %v", err)
	}
	if !strings.Contains(err.Error(), "Reference Gate") || !strings.Contains(err.Error(), "split this work into separate SPECs or explicit phases") {
		t.Fatalf("expected remediation guidance in block message, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-execution.json")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected no execution artifact for blocked run, got stat err=%v", statErr)
	}
}

func TestRunAllowsFrontendMinorExecutionAndEmbedsFrontendBriefInPrompt(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", frontendBriefFileName), strings.Join([]string{
		"# Frontend Brief",
		"",
		"Task Classification: frontend-minor",
		"Classification Rationale: Existing settings spacing fix.",
		"Frontend Gate Status: not-applicable",
		"Problem Gate: not-applicable",
		"Reference Gate: not-applicable",
		"Critique Gate: not-applicable",
		"Decision Gate: not-applicable",
		"Prototype Gate: not-applicable",
		"Prototype Evidence: n/a",
	}, "\n"))

	var promptArg string
	app.lookPath = func(name string) (string, error) {
		if name == "codex" || name == "git" {
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

	if err := app.Run(context.Background(), []string{"run", "SPEC-001"}); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !strings.Contains(promptArg, "## Frontend Brief") || !strings.Contains(promptArg, "Task Classification: frontend-minor") {
		t.Fatalf("expected prompt to embed frontend brief, got %q", promptArg)
	}
}

func TestLoadRunExecutionContextRejectsInvalidFrontendBriefContract(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", frontendBriefFileName), strings.Join([]string{
		"# Frontend Brief",
		"",
		"Task Classification: frontend-major",
		"Classification Rationale: New landing page.",
		"Frontend Gate Status: approved",
		"Problem Gate: complete",
		"Reference Gate: missing",
		"Critique Gate: complete",
		"Decision Gate: complete",
		"Prototype Gate: complete",
		"Prototype Evidence: wireframe",
	}, "\n"))

	_, err := app.loadRunExecutionContext(tmp, runExecuteOptions{specID: "SPEC-001", mode: executionModeDefault})
	if err == nil {
		t.Fatal("expected invalid frontend brief contract error")
	}
	if !strings.Contains(err.Error(), "invalid frontend brief contract") {
		t.Fatalf("expected invalid-contract error, got %v", err)
	}
}

func TestLoadRunExecutionContextBlocksFrontendMajorWhenDesignReviewPending(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", frontendBriefFileName), strings.Join([]string{
		"# Frontend Brief",
		"",
		"Task Classification: frontend-major",
		"Classification Rationale: New dashboard hierarchy.",
		"Frontend Gate Status: approved",
		"Problem Gate: complete",
		"Reference Gate: complete",
		"Critique Gate: complete",
		"Decision Gate: complete",
		"Prototype Gate: complete",
		"Prototype Evidence: wireframe",
		"",
		"## Reference Set",
		"",
		"- Reference 1: Complete.",
	}, "\n"))
	writeTestFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "reviews", "design.md"), strings.Join([]string{
		"# Design Review",
		"",
		"- Status: pending",
		"- Evidence Status: pending",
		"- Gate Decision: pending",
		"- Approved Direction: pending",
		"- Banned Patterns: pending",
		"- Open Questions: pending",
		"- Unresolved Questions: pending",
		"",
	}, "\n"))

	_, err := app.loadRunExecutionContext(tmp, runExecuteOptions{specID: "SPEC-001", mode: executionModeDefault})
	if err == nil {
		t.Fatal("expected pending design review to block frontend-major execution")
	}
	if !strings.Contains(err.Error(), "Design review gate decision is pending") || !strings.Contains(err.Error(), "design-review=pending") {
		t.Fatalf("expected pending design-review mismatch, got %v", err)
	}
}

func TestRunUsesConfiguredApprovalAndSandbox(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "system.yaml"), "runner: codex\napproval_policy: never\nsandbox_mode: read-only\n")

	app.lookPath = func(name string) (string, error) {
		if name == "codex" || name == "git" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexExec(name, args):
			mustContainArgs(t, args, []string{"-c", `approval_policy="never"`, "-s", "read-only"})
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
		if name == "codex" || name == "git" {
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
		if name == "codex" || name == "git" {
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
		if name == "codex" || name == "git" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if isCodexExec(name, args) {
			mustContainArgs(t, args, []string{"-c", `approval_policy="on-request"`, "-s", "workspace-write"})
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
		if name == "codex" || name == "git" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexExec(name, args):
			mustContainArgs(t, args, []string{"-c", `approval_policy="on-request"`})
			if indexOfArg(args, "resume") != -1 {
				mustContainArgs(t, args, []string{"-s", "workspace-write"})
			} else {
				mustContainArgs(t, args, []string{"-s", "workspace-write"})
			}
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

func TestRunFailsPreflightForMissingAddDir(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "codex.yaml"), "agent_mode: multi\nstatus_line_preset: namba\nrepo_skills_path: .agents/skills\nrepo_agents_path: .codex/agents\nweb_search: false\nadd_dirs: missing-dir\nsession_mode: stateful\nrepair_attempts: 1\n")

	app.lookPath = func(name string) (string, error) {
		if name == "codex" || name == "git" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}

	err := app.Run(context.Background(), []string{"run", "SPEC-001"})
	if err == nil {
		t.Fatal("expected preflight failure")
	}
	if !strings.Contains(err.Error(), "add_dir") {
		t.Fatalf("expected add_dir preflight error, got %v", err)
	}

	report := mustReadPreflightReport(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-preflight.json"))
	if report.Passed {
		t.Fatalf("expected failed preflight report: %+v", report)
	}
}

func TestRunAllowsResumeProfileViaExecLevelFlags(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "codex.yaml"), "agent_mode: multi\nstatus_line_preset: namba\nrepo_skills_path: .agents/skills\nrepo_agents_path: .codex/agents\nprofile: namba\nsession_mode: stateful\nrepair_attempts: 1\n")

	var sawResumeWithExecProfile bool
	app.lookPath = func(name string) (string, error) {
		if name == "codex" || name == "git" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if isCodexExec(name, args) {
			resumeIndex := indexOfArg(args, "resume")
			profileIndex := indexOfArg(args, "-p")
			if resumeIndex != -1 && profileIndex != -1 && profileIndex < resumeIndex && profileIndex+1 < len(args) && args[profileIndex+1] == "namba" {
				sawResumeWithExecProfile = true
			}
			return "runner output", nil
		}
		if isShellCommand(name) {
			return "validation ok", nil
		}
		return "", nil
	}

	if err := app.Run(context.Background(), []string{"run", "SPEC-001", "--team"}); err != nil {
		t.Fatalf("expected team run to succeed, got %v", err)
	}
	if !sawResumeWithExecProfile {
		t.Fatal("expected resume turns to carry profile via exec-level flags before resume")
	}
}

func prepareExecutionProject(t *testing.T) (string, *App, func()) {
	t.Helper()
	tmp := canonicalTempDir(t)
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	app.detectCodexCapabilities = func(context.Context, string, executionRequest) (codexCapabilityMatrix, error) {
		return testCodexCapabilities(), nil
	}
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
		return name == "cmd" && len(args) >= 3 && args[0] == "/c" && args[1] == "codex" && indexOfArg(args, "exec") != -1
	}
	return name == "codex" && indexOfArg(args, "exec") != -1
}

func isShellCommand(name string) bool {
	return name == "powershell" || name == "sh"
}

func isCodexVersionCommand(name string, args []string) bool {
	if runtime.GOOS == "windows" {
		return name == "cmd" && len(args) >= 3 && args[0] == "/c" && args[1] == "codex" && len(args) == 3 && args[2] == "--version"
	}
	return name == "codex" && len(args) == 1 && args[0] == "--version"
}

func isCodexHelpCommand(name string, args []string, resume bool) bool {
	if runtime.GOOS == "windows" {
		if name != "cmd" || len(args) < 4 || args[0] != "/c" || args[1] != "codex" || args[2] != "exec" {
			return false
		}
		if resume {
			return len(args) == 5 && args[3] == "resume" && args[4] == "--help"
		}
		return len(args) == 4 && args[3] == "--help"
	}
	if name != "codex" || len(args) < 2 || args[0] != "exec" {
		return false
	}
	if resume {
		return len(args) == 3 && args[1] == "resume" && args[2] == "--help"
	}
	return len(args) == 2 && args[1] == "--help"
}

func testCodexCapabilities() codexCapabilityMatrix {
	return codexCapabilityMatrix{
		Version: "codex-cli test",
		Exec: codexCommandCapabilities{
			Config:      true,
			SandboxFlag: true,
			ModelFlag:   true,
			ProfileFlag: true,
			AddDirFlag:  true,
		},
		Resume: codexCommandCapabilities{
			Config:    true,
			ModelFlag: true,
		},
	}
}

func mustContainArgs(t *testing.T, args []string, expected []string) {
	t.Helper()
	for i := 0; i < len(expected); i += 2 {
		if !containsArgPair(args, expected[i], expected[i+1]) {
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

func containsArgPair(args []string, key, value string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == key && args[i+1] == value {
			return true
		}
	}
	return false
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

func mustReadPreflightReport(t *testing.T, path string) preflightReport {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read preflight report: %v", err)
	}
	var report preflightReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("unmarshal preflight report: %v", err)
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
