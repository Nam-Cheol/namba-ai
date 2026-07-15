package namba

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunWritesExecutionEvidenceManifestOnSuccess(t *testing.T) {
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

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.SchemaVersion != executionEvidenceSchemaVersion || manifest.LogID != "spec-001" || manifest.SpecID != "SPEC-001" {
		t.Fatalf("unexpected execution evidence identity: %+v", manifest)
	}
	if manifest.Status != "completed" || manifest.ExecutionMode != string(executionModeDefault) {
		t.Fatalf("expected completed default manifest, got %+v", manifest)
	}
	for name, ref := range map[string]executionEvidenceRef{
		"request":    manifest.Request,
		"preflight":  manifest.Preflight,
		"execution":  manifest.Execution,
		"validation": manifest.Validation,
	} {
		if ref.State != executionEvidenceStatePresent {
			t.Fatalf("expected %s evidence to be present, got %+v", name, ref)
		}
	}
	if manifest.Progress.State != executionEvidenceStateNotApplicable {
		t.Fatalf("expected progress evidence to stay not_applicable, got %+v", manifest.Progress)
	}
	if manifest.Extensions.Browser.State != executionEvidenceStateNotApplicable {
		t.Fatalf("expected browser evidence to stay not_applicable, got %+v", manifest.Extensions.Browser)
	}
	if manifest.Extensions.Runtime.State != executionEvidenceStatePresent {
		t.Fatalf("expected runtime extension to be present, got %+v", manifest.Extensions.Runtime)
	}
	if manifest.CodexDiagnostics == nil || manifest.CodexDiagnostics.CodexAvailable != "unavailable" {
		t.Fatalf("expected non-blocking codex diagnostics in execution evidence, got %+v", manifest.CodexDiagnostics)
	}
	if manifest.CodexDiagnostics.Doctor.Status != "unavailable" {
		t.Fatalf("run evidence should avoid blocking doctor execution, got %+v", manifest.CodexDiagnostics.Doctor)
	}
	if manifest.ModelRouting == nil || manifest.ModelRouting.Version != "model-routing/v1" || manifest.ModelRouting.RequestedModel == "" {
		t.Fatalf("expected cost-balanced run evidence to include model routing, got %+v", manifest.ModelRouting)
	}
	if len(manifest.Extensions.Runtime.SignalBundles) != 1 || manifest.Extensions.Runtime.SignalBundles[0].Kind != "validation_attempts" {
		t.Fatalf("expected validation-attempt runtime bundle, got %+v", manifest.Extensions.Runtime.SignalBundles)
	}
	if !strings.Contains(strings.Join(manifest.Extensions.Runtime.SignalBundles[0].Paths, "\n"), "spec-001-validation-attempt-1.json") {
		t.Fatalf("expected validation-attempt path in runtime bundle, got %+v", manifest.Extensions.Runtime.SignalBundles[0])
	}
}

func TestRunEvidenceRecordsActualRoutedTurnsForReportAggregation(t *testing.T) {
	tmp := t.TempDir()
	app := NewApp(nil, nil)
	req := executionRequest{
		SpecID:             "SPEC-069",
		WorkDir:            tmp,
		Prompt:             "Cross-system security architecture design with acceptance tests.",
		Mode:               executionModeTeam,
		ModelRoutingPolicy: modelRoutingPolicyCostBalancedV1,
		DelegationPlan: delegationPlan{
			IntegratorRole:  "namba-implementer",
			DominantDomains: []string{"backend", "security"},
			SelectedRoleProfiles: []agentRuntimeProfile{
				runtimeProfileForAgent("namba-backend-architect"),
			},
		},
	}
	lifecycle := newHookLifecycle(app, tmp, "spec-069", req, "")
	lifecycle.recordModelRoutingTurns(buildExecutionTurnRequests(req))
	if err := lifecycle.writeRunEvidence(context.Background(), "completed", 0, false, ""); err != nil {
		t.Fatalf("write evidence: %v", err)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-069-evidence.json"))
	if len(manifest.ModelRoutingTurns) != 2 || manifest.ModelRoutingTurns[0].RequestedModel != modelRoutingModelSol || manifest.ModelRoutingTurns[1].RequestedModel != modelRoutingModelTerra || manifest.ModelRoutingTurns[0].RemainingSolHighTurns != 0 {
		t.Fatalf("expected phase-ordered Sol checkpoint then Terra implementation, got %+v", manifest.ModelRoutingTurns)
	}
	report := collectNambaReport(tmp, time.Now(), reportOptions{})
	if report.Runs.ModelRouting == nil || report.Runs.ModelRouting.TurnsByModel[modelRoutingModelTerra] != 1 || report.Runs.ModelRouting.TurnsByModel[modelRoutingModelSol] != 1 || report.Runs.ModelRouting.UnavailableUsageCount != 0 {
		t.Fatalf("expected report to aggregate routed turn models, got %+v", report.Runs.ModelRouting)
	}
}

func TestModelRoutingEvidenceDistinguishesUnobservedUsageFromUnavailableModel(t *testing.T) {
	planned := modelRoutingEvidenceForRequest(executionRequest{
		ModelRoutingPolicy:       modelRoutingPolicyCostBalancedV1,
		Model:                    modelRoutingModelTerra,
		RequestedReasoningEffort: "medium",
		RoutingDecision: modelRoutingDecisionResult{
			Phase:           routingPhaseImplement,
			Model:           modelRoutingModelTerra,
			ReasoningEffort: "medium",
			Status:          modelRoutingStatusPlanned,
		},
	})
	if planned == nil || planned.UsageState != modelRoutingUsageExternalUnobserved {
		t.Fatalf("successful routed turn must keep unobserved usage distinct, got %+v", planned)
	}

	blocked := modelRoutingEvidenceForRequest(executionRequest{
		ModelRoutingPolicy: modelRoutingPolicyCostBalancedV1,
		Model:              modelRoutingModelSol,
		RoutingDecision: modelRoutingDecisionResult{
			Phase:          routingPhaseArchitecture,
			Model:          modelRoutingModelSol,
			Status:         modelRoutingStatusBlocked,
			FallbackReason: "model_unavailable",
		},
	})
	if blocked == nil || blocked.UsageState != modelRoutingUsageUnavailable {
		t.Fatalf("blocked unavailable model must retain unavailable usage state, got %+v", blocked)
	}
}

func TestExecuteRunEvidenceRecordsResumeStateAfterItIsAssigned(t *testing.T) {
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
	threadIDs := []string{
		"019f5f13-3132-76c3-b9c7-ac521e89355e",
		"019f5f13-3132-76c3-b9c7-ac521e89355f",
	}
	codexCalls := 0
	app.runCmd = func(_ context.Context, name string, args []string, _ string) (string, error) {
		if !isCodexExec(name, args) {
			t.Fatalf("unexpected command: %s %v", name, args)
		}
		if codexCalls >= len(threadIDs) {
			t.Fatalf("unexpected extra Codex call: %v", args)
		}
		threadID := threadIDs[codexCalls]
		codexCalls++
		return `{"thread_id":"` + threadID + `"}`, nil
	}

	req := executionRequest{
		SpecID:             "SPEC-069",
		WorkDir:            tmp,
		Prompt:             "Implement a reversible local change with deterministic acceptance tests.",
		Mode:               executionModeTeam,
		Runner:             "codex",
		ApprovalPolicy:     "on-request",
		SandboxMode:        "workspace-write",
		ModelRoutingPolicy: modelRoutingPolicyCostBalancedV1,
		Model:              modelRoutingModelTerra,
		SessionMode:        "stateful",
		DelegationPlan: delegationPlan{
			IntegratorRole: "namba-implementer",
			SelectedRoleProfiles: []agentRuntimeProfile{
				{Role: "namba-reviewer", Model: modelRoutingModelTerra},
			},
		},
	}
	if _, _, err := app.executeRun(context.Background(), tmp, "spec-069", req, tmp, qualityConfig{TestCommand: "none", LintCommand: "none", TypecheckCommand: "none"}, nil, ""); err != nil {
		t.Fatalf("execute run: %v", err)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-069-evidence.json"))
	if len(manifest.ModelRoutingTurns) != 2 {
		t.Fatalf("expected two executed turns, got %+v", manifest.ModelRoutingTurns)
	}
	resumed := manifest.ModelRoutingTurns[1]
	if resumed.SessionStrategy != "explicit_thread_resume" || resumed.ThreadID != threadIDs[0] {
		t.Fatalf("expected evidence to retain the assigned resume state, got %+v", resumed)
	}
}

func TestExecuteRunPassesReadOnlyReviewCheckpointToTerraWriter(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		if name == "codex" || name == "git" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	threadIDs := []string{
		"019f5f13-3132-76c3-b9c7-ac521e89355e",
		"019f5f13-3132-76c3-b9c7-ac521e89355f",
		"019f5f13-3132-76c3-b9c7-ac521e893560",
	}
	var codexInputs []string
	app.runCodexCmdWithInput = func(_ context.Context, name string, args []string, _ string, input string) (string, string, error) {
		if !isCodexExec(name, args) || len(codexInputs) >= len(threadIDs) {
			t.Fatalf("unexpected Codex call: %s %v", name, args)
		}
		codexInputs = append(codexInputs, input)
		output := `{"thread_id":"` + threadIDs[len(codexInputs)-1] + `"}`
		if len(codexInputs) == 2 {
			output += "\n" + `{"type":"item.completed","item":{"text":"review checkpoint: tighten the cross-system boundary"}}`
		}
		return output, "", nil
	}
	app.runCmd = func(_ context.Context, name string, args []string, _ string) (string, error) {
		if isShellCommand(name) {
			return "validation ok", nil
		}
		t.Fatalf("unexpected command: %s %v", name, args)
		return "", nil
	}

	req := executionRequest{
		SpecID:             "SPEC-069",
		WorkDir:            tmp,
		Prompt:             "Cross-system feature with deterministic acceptance tests.",
		Mode:               executionModeTeam,
		Runner:             "codex",
		ApprovalPolicy:     "on-request",
		SandboxMode:        "workspace-write",
		ModelRoutingPolicy: modelRoutingPolicyCostBalancedV1,
		Model:              modelRoutingModelTerra,
		SessionMode:        "stateful",
		DelegationPlan: delegationPlan{
			IntegratorRole:  "namba-implementer",
			DominantDomains: []string{"backend", "frontend"},
			SelectedRoleProfiles: []agentRuntimeProfile{
				runtimeProfileForAgent("namba-reviewer"),
			},
		},
	}
	result, _, err := app.executeRun(context.Background(), tmp, "spec-069", req, tmp, qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"}, nil, "")
	if err != nil {
		t.Fatalf("execute run: %v", err)
	}
	if len(codexInputs) != 3 || len(result.Turns) != 3 {
		t.Fatalf("expected implement, review, and writer calls, inputs=%d turns=%+v", len(codexInputs), result.Turns)
	}
	if !strings.Contains(codexInputs[2], "## Read-only checkpoint handoff") || !strings.Contains(codexInputs[2], "review checkpoint: tighten the cross-system boundary") {
		t.Fatalf("Terra writer did not receive the Sol review handoff: %q", codexInputs[2])
	}
	if result.Turns[2].Name != "review-repair-writer" || result.Turns[2].Model != modelRoutingModelTerra {
		t.Fatalf("unexpected review handoff writer result: %+v", result.Turns[2])
	}
}

func TestExecuteRunRepairsFromLatestWritableModelThread(t *testing.T) {
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
	threadIDs := []string{
		"019f5f13-3132-76c3-b9c7-ac521e89355e",
		"019f5f13-3132-76c3-b9c7-ac521e89355f",
		"019f5f13-3132-76c3-b9c7-ac521e893560",
		"019f5f13-3132-76c3-b9c7-ac521e893561",
	}
	var codexArgs [][]string
	validationCalls := 0
	app.runCmd = func(_ context.Context, name string, args []string, _ string) (string, error) {
		switch {
		case isCodexExec(name, args):
			if len(codexArgs) >= len(threadIDs) {
				t.Fatalf("unexpected Codex call: %v", args)
			}
			codexArgs = append(codexArgs, append([]string(nil), args...))
			threadID := threadIDs[len(codexArgs)-1]
			return `{"thread_id":"` + threadID + `"}`, nil
		case isShellCommand(name):
			validationCalls++
			if validationCalls == 1 {
				return "validation failed", errors.New("simulated validation failure")
			}
			return "validation ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	req := executionRequest{
		SpecID:             "SPEC-069",
		WorkDir:            tmp,
		Prompt:             "Cross-system security architecture design with irreversible risk and acceptance tests.",
		Mode:               executionModeTeam,
		Runner:             "codex",
		ApprovalPolicy:     "on-request",
		SandboxMode:        "workspace-write",
		ModelRoutingPolicy: modelRoutingPolicyCostBalancedV1,
		Model:              modelRoutingModelTerra,
		SessionMode:        "stateful",
		RepairAttempts:     1,
		DelegationPlan: delegationPlan{
			IntegratorRole:  "namba-implementer",
			DominantDomains: []string{"backend", "security"},
			SelectedRoleProfiles: []agentRuntimeProfile{
				runtimeProfileForAgent("namba-backend-architect"),
				runtimeProfileForAgent("namba-reviewer"),
			},
		},
	}
	if _, _, err := app.executeRun(context.Background(), tmp, "spec-069", req, tmp, qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"}, nil, ""); err != nil {
		t.Fatalf("execute run: %v", err)
	}
	if len(codexArgs) != 4 {
		t.Fatalf("expected architecture, implementation, review, and repair calls, got %v", codexArgs)
	}
	repairArgs := codexArgs[3]
	resumeIndex := indexOfArg(repairArgs, "resume")
	if resumeIndex == -1 || resumeIndex+1 >= len(repairArgs) || repairArgs[resumeIndex+1] != threadIDs[2] {
		t.Fatalf("repair must resume the latest Terra writer thread %q, got %v", threadIDs[2], repairArgs)
	}
}

func TestExecuteRunBlocksRequiredSolWhenCapabilityProbeMarksItUnavailable(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		if name == "codex" || name == "git" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.detectCodexCapabilities = func(context.Context, string, executionRequest) (codexCapabilityMatrix, error) {
		capabilities := codexCapabilityMatrix{
			Version: "codex-cli without model surface",
			Exec: codexCommandCapabilities{
				ApprovalFlag: true,
				SandboxFlag:  true,
			},
		}
		capabilities.SolAvailable = boolPtr(false)
		return capabilities, nil
	}
	app.runCmd = func(_ context.Context, name string, args []string, _ string) (string, error) {
		if isCodexExec(name, args) {
			t.Fatalf("Sol-unavailable plan must block before Codex execution: %v", args)
		}
		t.Fatalf("unexpected command: %s %v", name, args)
		return "", nil
	}

	req := executionRequest{
		SpecID:             "SPEC-069",
		WorkDir:            tmp,
		Prompt:             "Cross-system architecture with acceptance tests.",
		Mode:               executionModeTeam,
		Runner:             "codex",
		ApprovalPolicy:     "on-request",
		SandboxMode:        "workspace-write",
		ModelRoutingPolicy: modelRoutingPolicyCostBalancedV1,
		SessionMode:        "stateful",
		DelegationPlan: delegationPlan{
			IntegratorRole:  "namba-implementer",
			DominantDomains: []string{"backend", "frontend"},
			SelectedRoleProfiles: []agentRuntimeProfile{
				runtimeProfileForAgent("namba-backend-architect"),
			},
		},
	}
	_, _, err := app.executeRun(context.Background(), tmp, "spec-069", req, tmp, qualityConfig{TestCommand: "none", LintCommand: "none", TypecheckCommand: "none"}, nil, "")
	if err == nil || !strings.Contains(err.Error(), "model_unavailable") {
		t.Fatalf("expected required Sol to block before execution, got %v", err)
	}
	if strings.Contains(err.Error(), "cannot be represented") {
		t.Fatalf("routing block must take precedence over generic invocation errors: %v", err)
	}
	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-069-evidence.json"))
	if len(manifest.ModelRoutingTurns) == 0 || manifest.ModelRoutingTurns[0].State != modelRoutingStatusBlocked || manifest.ModelRoutingTurns[0].FallbackReason != "model_unavailable" {
		t.Fatalf("expected blocked model-unavailable routing evidence, got %+v", manifest.ModelRoutingTurns)
	}
}

func TestExecuteRunBlocksUnavailableRepairBeforeCodex(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		if name == "codex" || name == "git" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.detectCodexCapabilities = func(context.Context, string, executionRequest) (codexCapabilityMatrix, error) {
		capabilities := testCodexCapabilities()
		capabilities.SolAvailable = boolPtr(false)
		return capabilities, nil
	}
	codexCalls := 0
	app.runCmd = func(_ context.Context, name string, args []string, _ string) (string, error) {
		switch {
		case isCodexExec(name, args):
			codexCalls++
			if codexCalls > 1 {
				t.Fatalf("blocked repair must not execute Codex: %v", args)
			}
			return `{"thread_id":"019f5f13-3132-76c3-b9c7-ac521e89355e"}`, nil
		case isShellCommand(name):
			return "validation failed", errors.New("simulated validation failure")
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	req := executionRequest{
		SpecID:             "SPEC-069",
		WorkDir:            tmp,
		Prompt:             "Cross-system security architecture with irreversible risk and acceptance tests.",
		Mode:               executionModeDefault,
		Runner:             "codex",
		ApprovalPolicy:     "on-request",
		SandboxMode:        "workspace-write",
		ModelRoutingPolicy: modelRoutingPolicyCostBalancedV1,
		Model:              modelRoutingModelTerra,
		SessionMode:        "stateful",
		RepairAttempts:     1,
		DelegationPlan: delegationPlan{
			IntegratorRole:  "namba-implementer",
			DominantDomains: []string{"backend", "security"},
		},
	}
	result, _, err := app.executeRun(context.Background(), tmp, "spec-069", req, tmp, qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"}, nil, "")
	if err == nil || !strings.Contains(err.Error(), "model_unavailable") {
		t.Fatalf("expected unavailable repair route to block, got %v", err)
	}
	if codexCalls != 1 || result.RetryCount != 0 {
		t.Fatalf("blocked repair must not spend an attempt or call Codex, calls=%d retries=%d", codexCalls, result.RetryCount)
	}
}

func TestExecuteRunUsesSolDiagnosticThenTerraWriterWithinOneRepairAttempt(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		if name == "codex" || name == "git" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.detectCodexCapabilities = func(context.Context, string, executionRequest) (codexCapabilityMatrix, error) {
		capabilities := testCodexCapabilities()
		capabilities.SolAvailable = boolPtr(true)
		return capabilities, nil
	}
	threadIDs := []string{
		"019f5f13-3132-76c3-b9c7-ac521e89355e",
		"019f5f13-3132-76c3-b9c7-ac521e89355f",
		"019f5f13-3132-76c3-b9c7-ac521e893560",
	}
	var codexArgs [][]string
	validationCalls := 0
	app.runCmd = func(_ context.Context, name string, args []string, _ string) (string, error) {
		switch {
		case isCodexExec(name, args):
			if len(codexArgs) >= len(threadIDs) {
				t.Fatalf("unexpected Codex call: %v", args)
			}
			codexArgs = append(codexArgs, append([]string(nil), args...))
			return `{"thread_id":"` + threadIDs[len(codexArgs)-1] + `"}`, nil
		case isShellCommand(name):
			validationCalls++
			if validationCalls == 1 {
				return "validation failed", errors.New("simulated validation failure")
			}
			return "validation ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	req := executionRequest{
		SpecID:             "SPEC-069",
		WorkDir:            tmp,
		Prompt:             "Cross-system security architecture with irreversible risk and acceptance tests.",
		Mode:               executionModeDefault,
		Runner:             "codex",
		ApprovalPolicy:     "on-request",
		SandboxMode:        "workspace-write",
		ModelRoutingPolicy: modelRoutingPolicyCostBalancedV1,
		Model:              modelRoutingModelTerra,
		SessionMode:        "stateful",
		RepairAttempts:     1,
		DelegationPlan: delegationPlan{
			IntegratorRole:  "namba-implementer",
			DominantDomains: []string{"backend", "security"},
		},
	}
	result, _, err := app.executeRun(context.Background(), tmp, "spec-069", req, tmp, qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"}, nil, "")
	if err != nil {
		t.Fatalf("execute run: %v", err)
	}
	if len(codexArgs) != 3 || result.RetryCount != 1 || len(result.Turns) != 3 {
		t.Fatalf("expected implement, diagnostic, and writer in one repair attempt, calls=%d retries=%d turns=%+v", len(codexArgs), result.RetryCount, result.Turns)
	}
	if result.Turns[0].Model != modelRoutingModelTerra || result.Turns[1].Model != modelRoutingModelSol || result.Turns[2].Model != modelRoutingModelTerra {
		t.Fatalf("unexpected repair model sequence: %+v", result.Turns)
	}
	if result.Turns[1].Name != "repair-1-diagnosis" || result.Turns[2].Name != "repair-1-writer" {
		t.Fatalf("unexpected repair turn names: %+v", result.Turns)
	}
	resumeIndex := indexOfArg(codexArgs[2], "resume")
	if resumeIndex == -1 || resumeIndex+1 >= len(codexArgs[2]) || codexArgs[2][resumeIndex+1] != threadIDs[0] {
		t.Fatalf("Terra writer must resume the latest writable Terra thread %q, got %v", threadIDs[0], codexArgs[2])
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-069-evidence.json"))
	if len(manifest.ModelRoutingTurns) != 3 || manifest.ModelRoutingTurns[2].RuleID != "terra-repair-after-sol-diagnostic-v1" {
		t.Fatalf("expected diagnostic and writer routing evidence, got %+v", manifest.ModelRoutingTurns)
	}
}

func TestCodexDiagnosticsEvidenceCoversVersionDoctorRedactionAndMismatch(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		if name == "codex" {
			return "codex", nil
		}
		return "", errors.New("missing")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexVersionCommand(name, args):
			return "codex 0.130.0", nil
		case name == "codex" && len(args) == 2 && args[0] == "remote-control" && args[1] == "--help":
			return "Usage: codex remote-control", nil
		case name == "codex" && len(args) == 2 && args[0] == "remote-control" && args[1] == "status":
			return "", errors.New("remote-control disabled")
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}
	app.runCmdWithInput = func(_ context.Context, name string, args []string, dir, input string) (string, string, error) {
		if name == "codex" && len(args) == 1 && args[0] == "doctor" {
			return "ok token=abc123", "Authorization: Bearer secret-token", nil
		}
		t.Fatalf("unexpected command with input: %s %v", name, args)
		return "", "", nil
	}

	diagnostics := app.buildCodexDiagnosticsEvidence(context.Background(), tmp, codexDiagnosticsOptions{
		LogDir:                filepath.ToSlash(filepath.Join(logsDir, "project")),
		LogPrefix:             "codex",
		RunCommands:           true,
		ConfiguredRoots:       []string{tmp, filepath.Join(tmp, "outside")},
		EffectiveRoots:        []string{tmp},
		IncludeDoctorLogFiles: true,
	})
	if diagnostics.Version.ParseStatus != "parsed" || diagnostics.Version.BaselineComparison != "older" {
		t.Fatalf("unexpected version evidence: %+v", diagnostics.Version)
	}
	if diagnostics.Doctor.Status != "detected" || diagnostics.Doctor.StdoutPath == "" || diagnostics.Doctor.StderrPath == "" {
		t.Fatalf("unexpected doctor evidence: %+v", diagnostics.Doctor)
	}
	if diagnostics.Redaction.Status != "redacted" || !diagnostics.Redaction.Applied {
		t.Fatalf("expected redaction evidence, got %+v", diagnostics.Redaction)
	}
	stdout := mustReadFile(t, filepath.Join(tmp, filepath.FromSlash(diagnostics.Doctor.StdoutPath)))
	stderr := mustReadFile(t, filepath.Join(tmp, filepath.FromSlash(diagnostics.Doctor.StderrPath)))
	if strings.Contains(stdout, "abc123") || strings.Contains(stderr, "secret-token") {
		t.Fatalf("doctor logs were not redacted: stdout=%q stderr=%q", stdout, stderr)
	}
	if diagnostics.WorkspaceRootComparison.Status != "advisory_mismatch" {
		t.Fatalf("expected advisory mismatch, got %+v", diagnostics.WorkspaceRootComparison)
	}
	if diagnostics.RemoteControl.Status != "disabled" || diagnostics.RemoteControl.Source != "stable_cli_help" {
		t.Fatalf("expected disabled remote-control readiness from help-only probe, got %+v", diagnostics.RemoteControl)
	}
	if diagnostics.RemoteEnvironments.Status != "unavailable" {
		t.Fatalf("expected absent remote environments to stay unavailable, got %+v", diagnostics.RemoteEnvironments)
	}
}

func TestCodexDiagnosticsEvidenceHandlesMissingAndUnparsableVersion(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) { return "", errors.New("not found") }
	missing := app.buildCodexDiagnosticsEvidence(context.Background(), tmp, codexDiagnosticsOptions{RunCommands: true})
	if missing.CodexAvailable != "not_detected" || missing.Doctor.Status != "local_fallback" {
		t.Fatalf("expected local fallback when codex is missing, got %+v", missing)
	}

	app.lookPath = func(name string) (string, error) { return "codex", nil }
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if isCodexVersionCommand(name, args) {
			return "codex nightly", nil
		}
		return "", nil
	}
	app.runCmdWithInput = func(_ context.Context, name string, args []string, dir, input string) (string, string, error) {
		return "", "", errors.New("doctor failed")
	}
	unparsable := app.buildCodexDiagnosticsEvidence(context.Background(), tmp, codexDiagnosticsOptions{RunCommands: true})
	if unparsable.Version.ParseStatus != "unparsable" || unparsable.Doctor.Status != "failed" {
		t.Fatalf("expected unparsable version and failed doctor, got %+v", unparsable)
	}
}

func TestCodexDiagnosticsRemoteControlAndEnvironmentReadiness(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		if name == "codex" {
			return "codex", nil
		}
		return "", errors.New("missing")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexVersionCommand(name, args):
			return "codex 0.131.0", nil
		case name == "codex" && len(args) == 2 && args[0] == "remote-control" && args[1] == "--help":
			return "Usage: codex remote-control", nil
		case name == "codex" && len(args) == 2 && args[0] == "remote-control" && args[1] == "status":
			return "remote-control enabled", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}
	app.runCmdWithInput = func(_ context.Context, name string, args []string, dir, input string) (string, string, error) {
		if name == "codex" && len(args) == 1 && args[0] == "doctor" {
			return "doctor ok", "", nil
		}
		t.Fatalf("unexpected command with input: %s %v", name, args)
		return "", "", nil
	}

	diagnostics := app.buildCodexDiagnosticsEvidence(context.Background(), tmp, codexDiagnosticsOptions{
		RunCommands:            true,
		RemoteEnvironmentNames: []string{"linux-large", "macos-arm64", "linux-large"},
	})
	if diagnostics.RemoteControl.Status != "enabled" || diagnostics.RemoteControl.Source != "stable_cli_status" {
		t.Fatalf("expected enabled remote-control status read, got %+v", diagnostics.RemoteControl)
	}
	if diagnostics.RemoteEnvironments.Status != "configured" || diagnostics.RemoteEnvironments.Source != "explicit_option" {
		t.Fatalf("expected configured remote environments, got %+v", diagnostics.RemoteEnvironments)
	}
	if got := strings.Join(diagnostics.RemoteEnvironments.Names, ","); got != "linux-large,macos-arm64" {
		t.Fatalf("expected normalized remote environment names, got %q", got)
	}
}

func TestRunEvidenceKeepsCodexRemoteReadinessNonBlocking(t *testing.T) {
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
			return "runner output", nil
		case isShellCommand(name):
			return "validation ok", nil
		case name == "codex" && len(args) > 0 && args[0] == "remote-control":
			t.Fatalf("run evidence must not probe remote-control: %s %v", name, args)
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"run", "SPEC-001"}); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.CodexDiagnostics == nil {
		t.Fatal("expected codex diagnostics in execution evidence")
	}
	if manifest.CodexDiagnostics.RemoteControl.Status != "local_fallback" {
		t.Fatalf("expected run evidence remote-control local_fallback, got %+v", manifest.CodexDiagnostics.RemoteControl)
	}
}

func TestCodexVersionBaselineComparisons(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want string
	}{
		{raw: "codex 0.130.9", want: "older"},
		{raw: "codex 0.131.0", want: "equal"},
		{raw: "codex 0.132.0", want: "newer"},
	} {
		got := buildCodexVersionEvidence(tc.raw, nil)
		if got.ParseStatus != "parsed" || got.BaselineComparison != tc.want {
			t.Fatalf("buildCodexVersionEvidence(%q) = %+v, want comparison %q", tc.raw, got, tc.want)
		}
	}
}

func TestCodexDiagnosticsDoctorTimeout(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) { return "codex", nil }
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if isCodexVersionCommand(name, args) {
			return "codex 0.131.0", nil
		}
		return "", nil
	}
	app.runCmdWithInput = func(ctx context.Context, name string, args []string, dir, input string) (string, string, error) {
		<-ctx.Done()
		return "", "", ctx.Err()
	}

	diagnostics := app.buildCodexDiagnosticsEvidence(context.Background(), tmp, codexDiagnosticsOptions{RunCommands: true, DoctorTimeout: time.Millisecond})
	if diagnostics.Doctor.Status != "timed_out" || !diagnostics.Doctor.TimedOut {
		t.Fatalf("expected timed_out doctor evidence, got %+v", diagnostics.Doctor)
	}
}

func TestRunProjectWritesCodexDiagnosticsEvidence(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		if name == "codex" {
			return "codex", nil
		}
		return "", errors.New("missing")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexVersionCommand(name, args):
			return "codex 0.131.0", nil
		case name == "codex" && len(args) == 2 && args[0] == "remote-control" && args[1] == "--help":
			return "Usage: codex remote-control", nil
		case name == "codex" && len(args) == 2 && args[0] == "remote-control" && args[1] == "status":
			return "remote-control disabled", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}
	app.runCmdWithInput = func(_ context.Context, name string, args []string, dir, input string) (string, string, error) {
		if name == "codex" && len(args) == 1 && args[0] == "doctor" {
			return "doctor ok", "", nil
		}
		t.Fatalf("unexpected command with input: %s %v", name, args)
		return "", "", nil
	}

	if err := app.Run(context.Background(), []string{"project"}); err != nil {
		t.Fatalf("project failed: %v", err)
	}
	var evidence projectCodexDiagnosticsEvidence
	data, err := os.ReadFile(filepath.Join(tmp, ".namba", "logs", "project", "codex-diagnostics-evidence.json"))
	if err != nil {
		t.Fatalf("read project diagnostics: %v", err)
	}
	if err := json.Unmarshal(data, &evidence); err != nil {
		t.Fatalf("unmarshal project diagnostics: %v", err)
	}
	if evidence.SchemaVersion != projectCodexDiagnosticsSchema || evidence.Diagnostics.Version.BaselineComparison != "equal" {
		t.Fatalf("unexpected project diagnostics evidence: %+v", evidence)
	}
}

func TestBuildExecutionEvidenceManifestExcludesStaleValidationAttempts(t *testing.T) {
	tmp := t.TempDir()
	writeTestFile(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-validation-attempt-1.json"), "{}")
	writeTestFile(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-validation-attempt-2.json"), "{}")

	manifest, err := buildExecutionEvidenceManifest(tmp, executionEvidenceOptions{
		ProjectRoot:        tmp,
		LogID:              "spec-001",
		SpecID:             "SPEC-001",
		ExecutionMode:      executionModeDefault,
		Status:             "completed",
		ValidationAttempts: 1,
	})
	if err != nil {
		t.Fatalf("buildExecutionEvidenceManifest failed: %v", err)
	}
	if len(manifest.Extensions.Runtime.SignalBundles) != 1 {
		t.Fatalf("expected one validation-attempt bundle, got %+v", manifest.Extensions.Runtime.SignalBundles)
	}
	if paths := manifest.Extensions.Runtime.SignalBundles[0].Paths; len(paths) != 1 || paths[0] != ".namba/logs/runs/spec-001-validation-attempt-1.json" {
		t.Fatalf("expected only current-run validation attempt path, got %+v", manifest.Extensions.Runtime.SignalBundles[0])
	}

	manifest, err = buildExecutionEvidenceManifest(tmp, executionEvidenceOptions{
		ProjectRoot:        tmp,
		LogID:              "spec-001",
		SpecID:             "SPEC-001",
		ExecutionMode:      executionModeDefault,
		Status:             "execution_failed",
		ValidationAttempts: 0,
	})
	if err != nil {
		t.Fatalf("buildExecutionEvidenceManifest failed without validation attempts: %v", err)
	}
	if len(manifest.Extensions.Runtime.SignalBundles) != 0 {
		t.Fatalf("expected stale validation attempts to be ignored when the current run never validated, got %+v", manifest.Extensions.Runtime.SignalBundles)
	}
	if manifest.Extensions.Runtime.State != executionEvidenceStateNotApplicable {
		t.Fatalf("expected runtime extension to stay not_applicable without current-run attempts, got %+v", manifest.Extensions.Runtime)
	}
}

func TestExecuteRunWritesProgressFailureEvidenceWhenFinalPublishFails(t *testing.T) {
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
		if isCodexExec(name, args) {
			return "runner output", nil
		}
		t.Fatalf("unexpected command: %s %v", name, args)
		return "", nil
	}

	req := app.newExecutionRequest(
		"SPEC-001",
		tmp,
		"prompt",
		executionModeParallel,
		suggestDelegationPlan(executionModeParallel, "prompt", "", ""),
		systemConfig{Runner: "codex", ApprovalPolicy: "on-request", SandboxMode: "workspace-write"},
		codexConfig{},
	)

	progress := &stubParallelProgressSink{
		path: filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-progress.events.jsonl"),
		failMatch: func(input parallelProgressEventInput, publishCount int) bool {
			return input.Phase == "merge_pending"
		},
		failPublishErr: errors.New("append denied"),
	}

	result, report, err := app.executeRun(
		context.Background(),
		tmp,
		"spec-001",
		req,
		tmp,
		qualityConfig{TestCommand: "none", LintCommand: "none", TypecheckCommand: "none"},
		progress,
		"spec-001-p1",
	)
	if err == nil || !strings.Contains(err.Error(), "append denied") {
		t.Fatalf("expected final publish failure, got err=%v result=%+v report=%+v", err, result, report)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "progress_log_failed" {
		t.Fatalf("expected progress_log_failed evidence after final publish failure, got %+v", manifest)
	}
	if manifest.Finalization.ProgressLogFailed != true {
		t.Fatalf("expected finalization to mark progress log failure, got %+v", manifest.Finalization)
	}
	if manifest.Progress.Path != ".namba/logs/runs/spec-001-progress.events.jsonl" || manifest.Progress.State != executionEvidenceStateMissing {
		t.Fatalf("expected progress reference to point at shared progress log path, got %+v", manifest.Progress)
	}
	if manifest.Validation.State != executionEvidenceStatePresent {
		t.Fatalf("expected validation artifact to remain present, got %+v", manifest.Validation)
	}
}

func TestExecuteRunPreservesPreviousValidationAttemptsWhenRetryPublishFails(t *testing.T) {
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

	validationCalls := 0
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexExec(name, args):
			return "runner output", nil
		case isShellCommand(name):
			validationCalls++
			if validationCalls == 1 {
				return "validation failed", errors.New("lint failed")
			}
			t.Fatalf("validation should not start a second attempt after publish failure: %s %v", name, args)
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	req := app.newExecutionRequest(
		"SPEC-001",
		tmp,
		"prompt",
		executionModeParallel,
		suggestDelegationPlan(executionModeParallel, "prompt", "", ""),
		systemConfig{Runner: "codex", ApprovalPolicy: "on-request", SandboxMode: "workspace-write"},
		codexConfig{},
	)
	req.RepairAttempts = 1

	progress := &stubParallelProgressSink{
		path: filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-progress.events.jsonl"),
		failMatch: func(input parallelProgressEventInput, publishCount int) bool {
			if input.Phase != "validating" {
				return false
			}
			attempt, _ := input.Metadata["attempt"].(int)
			return attempt == 2
		},
		failPublishErr: errors.New("append denied"),
	}

	result, report, err := app.executeRun(
		context.Background(),
		tmp,
		"spec-001",
		req,
		tmp,
		qualityConfig{TestCommand: "none", LintCommand: "lint", TypecheckCommand: "none"},
		progress,
		"spec-001-p1",
	)
	if err == nil || !strings.Contains(err.Error(), "append denied") {
		t.Fatalf("expected retry publish failure, got err=%v result=%+v report=%+v", err, result, report)
	}
	if validationCalls != 1 {
		t.Fatalf("expected exactly one completed validation attempt before publish failure, got %d", validationCalls)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "progress_log_failed" {
		t.Fatalf("expected progress_log_failed evidence after retry publish failure, got %+v", manifest)
	}
	if manifest.Finalization.ProgressLogFailed != true {
		t.Fatalf("expected finalization to mark progress log failure, got %+v", manifest.Finalization)
	}
	if manifest.Progress.Path != ".namba/logs/runs/spec-001-progress.events.jsonl" || manifest.Progress.State != executionEvidenceStateMissing {
		t.Fatalf("expected progress reference to point at shared progress log path, got %+v", manifest.Progress)
	}
	if len(manifest.Extensions.Runtime.SignalBundles) != 1 {
		t.Fatalf("expected validation-attempt bundle to remain present, got %+v", manifest.Extensions.Runtime.SignalBundles)
	}
	if paths := manifest.Extensions.Runtime.SignalBundles[0].Paths; len(paths) != 1 || paths[0] != ".namba/logs/runs/spec-001-validation-attempt-1.json" {
		t.Fatalf("expected first validation attempt to remain in evidence bundle, got %+v", manifest.Extensions.Runtime.SignalBundles[0])
	}
}

func TestExecuteRunExecutionFailureRecordsProgressFailureSeparately(t *testing.T) {
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
			return "partial output", errors.New("runner failed")
		case isShellCommand(name):
			t.Fatal("validators should not run after runner failure")
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	req := app.newExecutionRequest(
		"SPEC-001",
		tmp,
		"prompt",
		executionModeParallel,
		suggestDelegationPlan(executionModeParallel, "prompt", "", ""),
		systemConfig{Runner: "codex", ApprovalPolicy: "on-request", SandboxMode: "workspace-write"},
		codexConfig{},
	)

	progress := &stubParallelProgressSink{
		path: filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-progress.events.jsonl"),
		failMatch: func(input parallelProgressEventInput, publishCount int) bool {
			return input.Phase == "failed" && input.Status == "execution_failed"
		},
		failPublishErr: errors.New("append denied"),
	}

	result, report, err := app.executeRun(
		context.Background(),
		tmp,
		"spec-001",
		req,
		tmp,
		qualityConfig{TestCommand: "none", LintCommand: "none", TypecheckCommand: "none"},
		progress,
		"spec-001-p1",
	)
	if err == nil || !strings.Contains(err.Error(), "runner failed") || !strings.Contains(err.Error(), "append denied") {
		t.Fatalf("expected runner failure plus progress publish failure, got err=%v result=%+v report=%+v", err, result, report)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "execution_failed" {
		t.Fatalf("expected execution_failed evidence status to remain primary, got %+v", manifest)
	}
	if manifest.Finalization.ProgressLogFailed != true {
		t.Fatalf("expected finalization to record progress log failure separately, got %+v", manifest.Finalization)
	}
	if manifest.Progress.Path != ".namba/logs/runs/spec-001-progress.events.jsonl" || manifest.Progress.State != executionEvidenceStateMissing {
		t.Fatalf("expected progress reference to be preserved on execution failure, got %+v", manifest.Progress)
	}
}

func TestRunWritesExecutionEvidenceManifestOnPreflightFailure(t *testing.T) {
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

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "preflight_failed" {
		t.Fatalf("expected preflight_failed status, got %+v", manifest)
	}
	if manifest.Request.State != executionEvidenceStatePresent || manifest.Preflight.State != executionEvidenceStatePresent || manifest.Execution.State != executionEvidenceStatePresent {
		t.Fatalf("expected request/preflight/execution evidence on preflight failure, got %+v", manifest)
	}
	if manifest.Validation.State != executionEvidenceStateMissing {
		t.Fatalf("expected validation evidence to be missing on preflight failure, got %+v", manifest.Validation)
	}
}

func TestRunWritesExecutionEvidenceManifestOnExecutionFailure(t *testing.T) {
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
			return "partial output", errors.New("runner failed")
		case isShellCommand(name):
			t.Fatal("validators should not run after runner failure")
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	err := app.Run(context.Background(), []string{"run", "SPEC-001"})
	if err == nil {
		t.Fatal("expected execution failure")
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "execution_failed" {
		t.Fatalf("expected execution_failed status, got %+v", manifest)
	}
	if manifest.Execution.State != executionEvidenceStatePresent || manifest.Validation.State != executionEvidenceStateMissing {
		t.Fatalf("expected execution evidence without validation on runner failure, got %+v", manifest)
	}
}

func TestRunWritesExecutionEvidenceManifestOnValidationFailure(t *testing.T) {
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
			return "runner output", nil
		case isShellCommand(name):
			command := args[len(args)-1]
			if strings.Contains(command, "gofmt") {
				return "formatting failed", errors.New("lint failed")
			}
			return "ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	err := app.Run(context.Background(), []string{"run", "SPEC-001"})
	if err == nil {
		t.Fatal("expected validation failure")
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "validation_failed" {
		t.Fatalf("expected validation_failed status, got %+v", manifest)
	}
	if manifest.Validation.State != executionEvidenceStatePresent {
		t.Fatalf("expected validation artifact on validation failure, got %+v", manifest.Validation)
	}
	if manifest.Extensions.Runtime.State != executionEvidenceStatePresent {
		t.Fatalf("expected runtime extension on validation failure, got %+v", manifest.Extensions.Runtime)
	}
}

func TestRunParallelWritesExecutionEvidenceManifest(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-003"},
		[]string{"one", "two", "three"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{},
		workflowConfig{MaxParallelWorkers: 3},
		false,
	)
	if err != nil {
		t.Fatalf("runParallel failed: %v", err)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel-evidence.json"))
	if manifest.Status != "completed" || manifest.ExecutionMode != string(executionModeParallel) {
		t.Fatalf("expected completed parallel manifest, got %+v", manifest)
	}
	if manifest.Request.State != executionEvidenceStateNotApplicable || manifest.Validation.State != executionEvidenceStateNotApplicable {
		t.Fatalf("expected request/validation to stay not_applicable for aggregate parallel manifest, got %+v", manifest)
	}
	if manifest.Preflight.State != executionEvidenceStatePresent || manifest.Execution.State != executionEvidenceStatePresent || manifest.Progress.State != executionEvidenceStatePresent {
		t.Fatalf("expected preflight/execution/progress evidence on parallel run, got %+v", manifest)
	}
	if manifest.CodexDiagnostics == nil || manifest.CodexDiagnostics.RemoteControl.Status != "local_fallback" {
		t.Fatalf("expected non-blocking codex diagnostics on parallel aggregate evidence, got %+v", manifest.CodexDiagnostics)
	}
}

func TestRunParallelPreflightFailureWritesExecutionEvidenceManifest(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-003"},
		[]string{"one", "two", "three"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{AddDirs: []string{"missing-dir"}},
		workflowConfig{MaxParallelWorkers: 3},
		false,
	)
	if err == nil {
		t.Fatal("expected parallel preflight failure")
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel-evidence.json"))
	if manifest.Status != "preflight_failed" {
		t.Fatalf("expected preflight_failed parallel manifest, got %+v", manifest)
	}
	if manifest.Preflight.State != executionEvidenceStatePresent || manifest.Execution.State != executionEvidenceStatePresent || manifest.Progress.State != executionEvidenceStatePresent {
		t.Fatalf("expected aggregate parallel evidence on preflight failure, got %+v", manifest)
	}
}

func TestRunParallelCloseFailurePreservesCompletedStatusInExecutionEvidence(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	h.app.newParallelProgressSink = func(cfg parallelProgressSinkConfig) (parallelProgressSink, error) {
		return &stubParallelProgressSink{
			path:     cfg.Path,
			closeErr: errors.New("close denied"),
		}, nil
	}

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-003"},
		[]string{"one", "two", "three"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{},
		workflowConfig{MaxParallelWorkers: 3},
		false,
	)
	if err == nil || !strings.Contains(err.Error(), "close denied") {
		t.Fatalf("expected close failure to surface, got %v", err)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel-evidence.json"))
	if manifest.Status != "completed" {
		t.Fatalf("expected completed parallel evidence after close failure, got %+v", manifest)
	}
	if manifest.Finalization.ProgressLogFailed != true {
		t.Fatalf("expected finalization to preserve progress log failure separately, got %+v", manifest.Finalization)
	}
	if manifest.Execution.State != executionEvidenceStatePresent {
		t.Fatalf("expected aggregate execution artifact to remain present, got %+v", manifest.Execution)
	}
}

func TestRunParallelMergeFailureCloseFailurePreservesMergeFailedExecutionEvidence(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	h.mergeErr["namba/spec-003-p2"] = errors.New("merge conflict")
	h.app.newParallelProgressSink = func(cfg parallelProgressSinkConfig) (parallelProgressSink, error) {
		return &stubParallelProgressSink{
			path:     cfg.Path,
			closeErr: errors.New("close denied"),
		}, nil
	}

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-003"},
		[]string{"one", "two", "three"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{},
		workflowConfig{MaxParallelWorkers: 3},
		false,
	)
	if err == nil || !strings.Contains(err.Error(), "merge conflict") || !strings.Contains(err.Error(), "close denied") {
		t.Fatalf("expected merge and close failures to surface together, got %v", err)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel-evidence.json"))
	if manifest.Status != "merge_failed" {
		t.Fatalf("expected merge_failed parallel evidence after merge+close failure, got %+v", manifest)
	}
	if manifest.Finalization.ProgressLogFailed != true {
		t.Fatalf("expected finalization to preserve progress log failure separately, got %+v", manifest.Finalization)
	}
	if manifest.Progress.Path != ".namba/logs/runs/spec-003-parallel.events.jsonl" || manifest.Progress.State != executionEvidenceStateMissing {
		t.Fatalf("expected progress reference to keep the shared event log path, got %+v", manifest.Progress)
	}
}

func TestBuildExecutionEvidenceManifestAllowsTypedBrowserArtifacts(t *testing.T) {
	tmp := t.TempDir()
	browserPath := filepath.Join(tmp, ".namba", "logs", "runs", "spec-033-browser-trace.zip")
	writeTestFile(t, browserPath, "trace")

	manifest, err := buildExecutionEvidenceManifest(tmp, executionEvidenceOptions{
		ProjectRoot:   tmp,
		LogID:         "spec-033",
		RunID:         "spec-033",
		SpecID:        "SPEC-033",
		ExecutionMode: executionModeTeam,
		Status:        "completed",
		Request: executionEvidenceRefInput{
			Kind:          "request",
			NotApplicable: true,
		},
		Preflight: executionEvidenceRefInput{
			Kind:          "preflight",
			NotApplicable: true,
		},
		Execution: executionEvidenceRefInput{
			Kind:          "execution",
			NotApplicable: true,
		},
		Validation: executionEvidenceRefInput{
			Kind:          "validation",
			NotApplicable: true,
		},
		Progress: executionEvidenceRefInput{
			Kind:          "progress",
			NotApplicable: true,
		},
		BrowserArtifacts: []executionEvidenceRef{
			{
				Kind: "trace",
				Path: filepath.ToSlash(filepath.Join(logsDir, "runs", "spec-033-browser-trace.zip")),
			},
		},
	})
	if err != nil {
		t.Fatalf("buildExecutionEvidenceManifest failed: %v", err)
	}
	if manifest.Extensions.Browser.State != executionEvidenceStatePresent {
		t.Fatalf("expected browser extension to be present, got %+v", manifest.Extensions.Browser)
	}
	if len(manifest.Extensions.Browser.Artifacts) != 1 || manifest.Extensions.Browser.Artifacts[0].State != executionEvidenceStatePresent {
		t.Fatalf("expected present typed browser artifact, got %+v", manifest.Extensions.Browser.Artifacts)
	}
}

func TestChangeSummaryLatestExecutionProofSectionSkipsLegacyRuns(t *testing.T) {
	tmp, _, restore := prepareExecutionProject(t)
	defer restore()

	if lines := changeSummaryLatestExecutionProofSection(tmp); len(lines) != 0 {
		t.Fatalf("expected no execution-proof section without manifest, got %+v", lines)
	}
	if lines := prChecklistLatestExecutionProofItem(tmp); len(lines) != 0 {
		t.Fatalf("expected no execution-proof checklist item without manifest, got %+v", lines)
	}
}

func TestExecutionProofConsumersUseLatestAvailableManifest(t *testing.T) {
	tmp, _, restore := prepareExecutionProject(t)
	defer restore()

	olderPath := filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json")
	newerPath := filepath.Join(tmp, ".namba", "logs", "runs", "direct-fix-evidence.json")
	writeExecutionEvidenceFixture(t, olderPath, executionEvidenceManifest{
		SchemaVersion: executionEvidenceSchemaVersion,
		LogID:         "spec-001",
		RunID:         "spec-001",
		SpecID:        "SPEC-001",
		GeneratedAt:   "2026-04-22T09:00:00Z",
		ExecutionMode: string(executionModeTeam),
		Advisory:      true,
		Status:        "completed",
		Request:       executionEvidenceRef{Kind: "request", State: executionEvidenceStatePresent},
		Preflight:     executionEvidenceRef{Kind: "preflight", State: executionEvidenceStatePresent},
		Execution:     executionEvidenceRef{Kind: "execution", State: executionEvidenceStatePresent},
		Validation:    executionEvidenceRef{Kind: "validation", State: executionEvidenceStatePresent},
		Progress:      executionEvidenceRef{Kind: "progress", State: executionEvidenceStateNotApplicable},
		Extensions: executionEvidenceExtensions{
			Browser: executionEvidenceExtension{State: executionEvidenceStateNotApplicable},
			Runtime: executionEvidenceExtension{State: executionEvidenceStatePresent},
		},
	})
	writeExecutionEvidenceFixture(t, newerPath, executionEvidenceManifest{
		SchemaVersion: executionEvidenceSchemaVersion,
		LogID:         "direct-fix",
		RunID:         "direct-fix",
		SpecID:        "DIRECT-FIX",
		GeneratedAt:   "2026-04-22T10:00:00Z",
		ExecutionMode: string(executionModeDefault),
		Advisory:      true,
		Status:        "completed",
		Request:       executionEvidenceRef{Kind: "request", State: executionEvidenceStatePresent},
		Preflight:     executionEvidenceRef{Kind: "preflight", State: executionEvidenceStatePresent},
		Execution:     executionEvidenceRef{Kind: "execution", State: executionEvidenceStatePresent},
		Validation:    executionEvidenceRef{Kind: "validation", State: executionEvidenceStatePresent},
		Progress:      executionEvidenceRef{Kind: "progress", State: executionEvidenceStateNotApplicable},
		Extensions: executionEvidenceExtensions{
			Browser: executionEvidenceExtension{State: executionEvidenceStateNotApplicable},
			Runtime: executionEvidenceExtension{State: executionEvidenceStatePresent},
		},
	})

	changeSummaryLines := strings.Join(changeSummaryLatestExecutionProofSection(tmp), "\n")
	if !strings.Contains(changeSummaryLines, "direct-fix-evidence.json") || !strings.Contains(changeSummaryLines, "Proof target: `DIRECT-FIX`") {
		t.Fatalf("expected latest available proof in change summary section, got %q", changeSummaryLines)
	}
	checklistLines := strings.Join(prChecklistLatestExecutionProofItem(tmp), "\n")
	if !strings.Contains(checklistLines, "direct-fix-evidence.json") || !strings.Contains(checklistLines, "target `DIRECT-FIX`") {
		t.Fatalf("expected latest available proof in checklist item, got %q", checklistLines)
	}
}

func TestSyncOutputsSurfaceExecutionProofSeparatelyFromReadiness(t *testing.T) {
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

	syncCtx, err := app.loadSyncContext(tmp)
	if err != nil {
		t.Fatalf("loadSyncContext failed: %v", err)
	}
	if err := app.writeSyncProjectSupportDocs(syncCtx); err != nil {
		t.Fatalf("writeSyncProjectSupportDocs failed: %v", err)
	}

	changeSummary := mustReadFile(t, filepath.Join(tmp, ".namba", "project", "change-summary.md"))
	if !strings.Contains(changeSummary, "## Latest Review Readiness") || !strings.Contains(changeSummary, "## Latest Execution Proof") {
		t.Fatalf("expected change summary to separate readiness and execution proof, got %q", changeSummary)
	}
	if !strings.Contains(changeSummary, "Execution proof status: `completed`") {
		t.Fatalf("expected execution proof status in change summary, got %q", changeSummary)
	}

	prChecklist := mustReadFile(t, filepath.Join(tmp, ".namba", "project", "pr-checklist.md"))
	if !strings.Contains(prChecklist, "Latest SPEC review readiness checked") || !strings.Contains(prChecklist, "Latest execution proof checked") {
		t.Fatalf("expected PR checklist to include readiness and execution-proof items, got %q", prChecklist)
	}
}

func TestLatestExecutionProofBuildersAgreeOnArtifactSelection(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	olderManifest, err := buildExecutionEvidenceManifest(tmp, executionEvidenceOptions{
		ProjectRoot:   tmp,
		LogID:         "spec-000",
		SpecID:        "SPEC-000",
		ExecutionMode: executionModeDefault,
		Status:        "completed",
		GeneratedAt:   time.Date(2026, 4, 20, 8, 0, 0, 0, time.UTC),
		FinalizedBy:   "test",
		Request: executionEvidenceRefInput{
			Kind:          "request",
			NotApplicable: true,
		},
		Preflight: executionEvidenceRefInput{
			Kind:          "preflight",
			NotApplicable: true,
		},
		Execution: executionEvidenceRefInput{
			Kind:          "execution",
			NotApplicable: true,
		},
		Validation: executionEvidenceRefInput{
			Kind:          "validation",
			NotApplicable: true,
		},
		Progress: executionEvidenceRefInput{
			Kind:          "progress",
			NotApplicable: true,
		},
	})
	if err != nil {
		t.Fatalf("build older execution evidence manifest: %v", err)
	}
	newerManifest, err := buildExecutionEvidenceManifest(tmp, executionEvidenceOptions{
		ProjectRoot:   tmp,
		LogID:         "spec-001",
		SpecID:        "SPEC-001",
		ExecutionMode: executionModeDefault,
		Status:        "completed",
		GeneratedAt:   time.Date(2026, 4, 21, 8, 0, 0, 0, time.UTC),
		FinalizedBy:   "test",
		Request: executionEvidenceRefInput{
			Kind:          "request",
			NotApplicable: true,
		},
		Preflight: executionEvidenceRefInput{
			Kind:          "preflight",
			NotApplicable: true,
		},
		Execution: executionEvidenceRefInput{
			Kind:          "execution",
			NotApplicable: true,
		},
		Validation: executionEvidenceRefInput{
			Kind:          "validation",
			NotApplicable: true,
		},
		Progress: executionEvidenceRefInput{
			Kind:          "progress",
			NotApplicable: true,
		},
	})
	if err != nil {
		t.Fatalf("build newer execution evidence manifest: %v", err)
	}

	writeExecutionEvidenceFixture(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-000-evidence.json"), olderManifest)
	writeExecutionEvidenceFixture(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"), newerManifest)

	changeSummary := strings.Join(changeSummaryLatestExecutionProofSection(tmp), "\n")
	prChecklist := strings.Join(prChecklistLatestExecutionProofItem(tmp), "\n")

	for _, want := range []string{
		"spec-001-evidence.json",
		"Proof target: `SPEC-001`",
		"Execution proof status: `completed`",
	} {
		if !strings.Contains(changeSummary, want) {
			t.Fatalf("expected change summary to contain %q, got %q", want, changeSummary)
		}
	}
	for _, want := range []string{
		"spec-001-evidence.json",
		"target `SPEC-001`",
	} {
		if !strings.Contains(prChecklist, want) {
			t.Fatalf("expected PR checklist to contain %q, got %q", want, prChecklist)
		}
	}
	if strings.Contains(changeSummary, "spec-000-evidence.json") || strings.Contains(prChecklist, "spec-000-evidence.json") {
		t.Fatalf("expected both support builders to keep the latest execution proof selection aligned, got summary=%q checklist=%q", changeSummary, prChecklist)
	}
}

func mustReadExecutionEvidenceManifest(t *testing.T, path string) executionEvidenceManifest {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read execution evidence manifest: %v", err)
	}
	var manifest executionEvidenceManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("unmarshal execution evidence manifest: %v", err)
	}
	return manifest
}

func writeExecutionEvidenceFixture(t *testing.T, path string, manifest executionEvidenceManifest) {
	t.Helper()
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshal execution evidence fixture: %v", err)
	}
	writeTestFile(t, path, string(data))
}
