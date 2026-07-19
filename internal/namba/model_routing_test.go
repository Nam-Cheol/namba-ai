package namba

import "testing"

func TestModelRoutingDecisionPolicyTable(t *testing.T) {
	tests := []struct {
		name     string
		input    modelRoutingInput
		model    string
		effort   string
		tier     string
		readOnly bool
		status   string
		reason   string
	}{
		{
			name:  "simple single subsystem implementation uses Luna",
			input: modelRoutingInput{Phase: routingPhaseImplement, Role: "namba-implementer", SimpleImplementation: true, SingleSubsystem: true, ExplicitTransformation: true, Reversible: true, DeterministicAcceptance: true},
			model: modelRoutingModelLuna, effort: "medium", tier: "efficient", status: modelRoutingStatusPlanned,
		},
		{
			name:  "ordinary implementation uses Terra",
			input: modelRoutingInput{Phase: routingPhaseImplement, Role: "namba-implementer"},
			model: modelRoutingModelTerra, effort: "medium", tier: "standard", status: modelRoutingStatusPlanned,
		},
		{
			name:  "build system changes stay on Terra",
			input: modelRoutingInput{Phase: routingPhaseImplement, Role: "namba-implementer", SimpleImplementation: true, SingleSubsystem: true, ExplicitTransformation: true, Reversible: true, DeterministicAcceptance: true, BuildSystemChange: true},
			model: modelRoutingModelTerra, effort: "medium", tier: "standard", status: modelRoutingStatusPlanned,
		},
		{
			name:  "critical simple implementation stays on Terra",
			input: modelRoutingInput{Phase: routingPhaseImplement, Role: "namba-security-engineer", SimpleImplementation: true, SingleSubsystem: true, ExplicitTransformation: true, Reversible: true, DeterministicAcceptance: true, CriticalRisk: true},
			model: modelRoutingModelTerra, effort: "medium", tier: "standard", status: modelRoutingStatusPlanned,
		},
		{
			name:  "cross system simple implementation stays on Terra",
			input: modelRoutingInput{Phase: routingPhaseImplement, Role: "namba-implementer", SimpleImplementation: true, SingleSubsystem: true, ExplicitTransformation: true, Reversible: true, DeterministicAcceptance: true, CrossSystem: true},
			model: modelRoutingModelTerra, effort: "medium", tier: "standard", status: modelRoutingStatusPlanned,
		},
		{
			name:  "irreversible simple implementation stays on Terra",
			input: modelRoutingInput{Phase: routingPhaseImplement, Role: "namba-implementer", SimpleImplementation: true, SingleSubsystem: true, ExplicitTransformation: true, Reversible: true, DeterministicAcceptance: true, Irreversible: true},
			model: modelRoutingModelTerra, effort: "medium", tier: "standard", status: modelRoutingStatusPlanned,
		},
		{
			name:  "cross system architecture uses Sol medium read only",
			input: modelRoutingInput{Phase: routingPhaseArchitecture, Role: "namba-backend-architect", CrossSystem: true},
			model: modelRoutingModelSol, effort: "medium", tier: "deep", readOnly: true, status: modelRoutingStatusPlanned,
		},
		{
			name:  "critical ambiguous irreversible risk uses Sol high",
			input: modelRoutingInput{Phase: routingPhaseReview, Role: "namba-security-engineer", CriticalRisk: true, HighAmbiguity: true, Irreversible: true, RemainingSolTurns: 1},
			model: modelRoutingModelSol, effort: "high", tier: "deep", readOnly: true, status: modelRoutingStatusPlanned,
		},
		{
			name:  "repeated failure without extra risk stays Terra high",
			input: modelRoutingInput{Phase: routingPhaseRepair, Role: "namba-implementer", RepairCount: 2},
			model: modelRoutingModelTerra, effort: "high", tier: "standard", status: modelRoutingStatusPlanned,
		},
		{
			name:  "Sol implementation is forbidden",
			input: modelRoutingInput{Phase: routingPhaseImplement, Role: "namba-backend-architect", CrossSystem: true},
			model: modelRoutingModelTerra, effort: "medium", tier: "standard", status: modelRoutingStatusPlanned,
		},
		{
			name:  "required Sol unavailable blocks",
			input: modelRoutingInput{Phase: routingPhaseArchitecture, Role: "namba-backend-architect", CrossSystem: true, SolAvailable: boolPtr(false)},
			model: modelRoutingModelSol, effort: "medium", tier: "deep", readOnly: true, status: modelRoutingStatusBlocked, reason: modelRoutingReasonBlockedModelUnavailable,
		},
		{
			name:  "optional Sol unavailable falls back to Terra high",
			input: modelRoutingInput{Phase: routingPhaseDesign, Role: "namba-designer", SolAvailable: boolPtr(false)},
			model: modelRoutingModelTerra, effort: "high", tier: "standard", status: modelRoutingStatusFallback, reason: modelRoutingReasonModelUnavailable,
		},
		{
			name:  "active zero Sol budget blocks required architecture decision",
			input: modelRoutingInput{Phase: routingPhaseArchitecture, Role: "namba-backend-architect", CrossSystem: true, RemainingSolTurns: 0, SolBudgetActive: true},
			model: modelRoutingModelSol, effort: "medium", tier: "deep", readOnly: true, status: modelRoutingStatusBlocked, reason: "sol_turn_budget_exhausted",
		},
		{
			name:  "active Sol budget decrements before planned decision returns",
			input: modelRoutingInput{Phase: routingPhaseArchitecture, Role: "namba-backend-architect", CrossSystem: true, RemainingSolTurns: 2, SolBudgetActive: true},
			model: modelRoutingModelSol, effort: "medium", tier: "deep", readOnly: true, status: modelRoutingStatusPlanned,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := modelRoutingDecision(tt.input)
			if got.Model != tt.model || got.ReasoningEffort != tt.effort || got.Tier != tt.tier || got.ReadOnly != tt.readOnly || got.Status != tt.status {
				t.Fatalf("decision = %+v", got)
			}
			if got.FallbackReason != tt.reason {
				t.Fatalf("fallback reason = %q, want %q", got.FallbackReason, tt.reason)
			}
			if got.ReasoningEffort == "xhigh" || got.ReasoningEffort == "max" {
				t.Fatalf("forbidden effort: %+v", got)
			}
			if tt.name == "active Sol budget decrements before planned decision returns" && got.RemainingSolTurns != 1 {
				t.Fatalf("remaining Sol turns = %d, want 1", got.RemainingSolTurns)
			}
		})
	}
}

func TestModelRoutingPolicyRegistryIsExplicitAndStable(t *testing.T) {
	for tier, want := range map[string]string{
		"efficient": modelRoutingModelLuna,
		"standard":  modelRoutingModelTerra,
		"deep":      modelRoutingModelSol,
	} {
		if got := modelRoutingPolicyRegistry[tier]; got != want {
			t.Fatalf("policy %q = %q, want %q", tier, got, want)
		}
	}
}

func boolPtr(value bool) *bool { return &value }
