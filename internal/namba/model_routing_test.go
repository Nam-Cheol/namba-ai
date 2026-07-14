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
			model: modelRoutingModelSol, effort: "medium", tier: "deep", readOnly: true, status: modelRoutingStatusBlocked,
		},
		{
			name:  "optional Sol unavailable falls back to Terra high",
			input: modelRoutingInput{Phase: routingPhaseDesign, Role: "namba-designer", SolAvailable: boolPtr(false)},
			model: modelRoutingModelTerra, effort: "high", tier: "standard", status: modelRoutingStatusFallback,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := modelRoutingDecision(tt.input)
			if got.Model != tt.model || got.ReasoningEffort != tt.effort || got.Tier != tt.tier || got.ReadOnly != tt.readOnly || got.Status != tt.status {
				t.Fatalf("decision = %+v", got)
			}
			if got.ReasoningEffort == "xhigh" || got.ReasoningEffort == "max" {
				t.Fatalf("forbidden effort: %+v", got)
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
