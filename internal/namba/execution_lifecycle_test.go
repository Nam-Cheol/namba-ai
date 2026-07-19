package namba

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExecutionLifecycleStateOwnsRoutingSessionAndEvidenceBoundaries(t *testing.T) {
	req := executionRequest{
		SpecID:             "SPEC-069",
		Prompt:             "simple mechanical rename with deterministic acceptance tests",
		Mode:               executionModeDefault,
		ModelRoutingPolicy: modelRoutingPolicyCostBalancedV1,
		SessionMode:        "stateful",
		DelegationPlan: delegationPlan{
			IntegratorRole:  "namba-implementer",
			DominantDomains: []string{"core"},
		},
	}
	state := newExecutionLifecycleState(req)
	state.recordPreflight(codexCapabilityMatrix{SolAvailable: boolPtr(true)})

	primary, reached := state.modelRoutingEvidence()
	if primary == nil || primary.RequestedModel != modelRoutingModelLuna || len(reached) != 0 {
		t.Fatalf("preflight planning must provide primary evidence without claiming reached turns, primary=%+v reached=%+v", primary, reached)
	}
	planned := state.plannedExecutionTurns()
	if len(planned) != 1 {
		t.Fatalf("expected one planned execution turn, got %+v", planned)
	}

	turn := planned[0]
	state.replaceReachedRoutingTurns([]executionRequest{turn})
	primary, reached = state.modelRoutingEvidence()
	if primary == nil || primary.RequestedModel != modelRoutingModelLuna || len(reached) != 1 {
		t.Fatalf("reached routing must override planned-only evidence, primary=%+v reached=%+v", primary, reached)
	}

	threadID := "019f5f13-3132-76c3-b9c7-ac521e89355e"
	if !state.observeWritableThread(turn, threadID) || state.latestWritableThreadID(turn.Model) != threadID {
		t.Fatalf("valid writable UUID must become the only resume authority")
	}
	readOnly := turn
	readOnly.Model = modelRoutingModelSol
	readOnly.RoutingDecision.ReadOnly = true
	state.observeWritableThread(readOnly, "")
	if state.latestWritableThreadID(turn.Model) != threadID {
		t.Fatalf("read-only turns must not revoke writable resume authority")
	}
	if state.observeWritableThread(turn, "") || state.latestWritableThreadID(turn.Model) != "" {
		t.Fatalf("missing writable UUID must revoke stale resume authority")
	}
}

func TestExecuteRunBoundsAndPersistsSolProbeTimeout(t *testing.T) {
	tmp := t.TempDir()
	app := NewApp(nil, nil)
	app.capabilityProbeTimeout = 20 * time.Millisecond
	app.lookPath = func(name string) (string, error) {
		if name == "codex" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if name != "codex" || dir != tmp {
			t.Fatalf("unexpected capability command: %s %v dir=%s", name, args, dir)
		}
		switch {
		case isCodexVersionCommand(name, args):
			return "codex-cli test", nil
		case isCodexHelpCommand(name, args, false):
			return "-c, --config\n-a, --ask-for-approval\n-s, --sandbox\n-m, --model\n--ephemeral\n--json", nil
		default:
			t.Fatalf("unexpected capability command: %s %v", name, args)
			return "", nil
		}
	}
	app.runCodexCmdWithInput = func(ctx context.Context, _ string, _ []string, _ string, _ string) (string, string, error) {
		<-ctx.Done()
		return "", "", ctx.Err()
	}

	req := executionRequest{
		SpecID:             "SPEC-069",
		WorkDir:            tmp,
		Prompt:             "Cross-system architecture with deterministic acceptance tests.",
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

	started := time.Now()
	_, _, err := app.executeRun(context.Background(), tmp, "spec-069", req, tmp, qualityConfig{}, nil, "")
	if err == nil || !strings.Contains(err.Error(), modelRoutingReasonBlockedModelUnavailable) {
		t.Fatalf("timed-out required Sol probe must become a bounded routing block, got %v", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("Sol probe exceeded bounded lifecycle window: %s", elapsed)
	}

	report := mustReadPreflightReport(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-069-preflight.json"))
	if len(report.Probes) != 3 {
		t.Fatalf("expected version, exec-help, and Sol probe outcomes, got %+v", report.Probes)
	}
	solProbe := report.Probes[2]
	if solProbe.Name != lifecycleProbeSolAvailability || solProbe.Status != lifecycleProbeStatusTimedOut || solProbe.TimeoutMS != 20 {
		t.Fatalf("preflight artifact must persist the bounded Sol timeout, got %+v", solProbe)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-069-evidence.json"))
	if manifest.Status != "preflight_failed" || manifest.Preflight.State != executionEvidenceStatePresent || manifest.ModelRouting == nil || manifest.ModelRouting.State != modelRoutingStatusBlocked {
		t.Fatalf("bounded probe failure must finalize blocked preflight evidence, got %+v", manifest)
	}
}
