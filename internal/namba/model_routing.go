package namba

import "strings"

// Model IDs are deliberately explicit: the gpt-5.6 alias can resolve to a
// different tier and is therefore not a stable execution contract.
const (
	modelRoutingPolicyCostBalancedV1 = "gpt-5.6-cost-balanced-v1"
	modelRoutingPolicyLegacyStaticV1 = "legacy-static-v1"
	modelRoutingModelLuna            = "gpt-5.6-luna"
	modelRoutingModelTerra           = "gpt-5.6-terra"
	modelRoutingModelSol             = "gpt-5.6-sol"
)

var modelRoutingPolicyRegistry = map[string]string{
	"efficient": modelRoutingModelLuna,
	"standard":  modelRoutingModelTerra,
	"deep":      modelRoutingModelSol,
}

type routingPhase string

const (
	routingPhaseIntake       routingPhase = "intake"
	routingPhasePlan         routingPhase = "plan"
	routingPhaseDesign       routingPhase = "design"
	routingPhaseArchitecture routingPhase = "architecture"
	routingPhaseImplement    routingPhase = "implement"
	routingPhaseTest         routingPhase = "test"
	routingPhaseIntegration  routingPhase = "integration"
	routingPhaseReview       routingPhase = "review"
	routingPhaseRepair       routingPhase = "repair"
)

const (
	modelRoutingStatusPlanned  = "planned"
	modelRoutingStatusFallback = "fallback"
	modelRoutingStatusBlocked  = "blocked"
)

type modelRoutingInput struct {
	Phase                   routingPhase
	Role                    string
	Domain                  string
	CriticalRisk            bool
	HighAmbiguity           bool
	CrossSystem             bool
	Irreversible            bool
	RepairCount             int
	RemainingSolTurns       int
	SimpleImplementation    bool
	SingleSubsystem         bool
	ExplicitTransformation  bool
	Reversible              bool
	DeterministicAcceptance bool
	PublicContractChange    bool
	UnresolvedReview        bool
	SolAvailable            *bool // nil means capability has not ruled Sol out.
}

type modelRoutingDecisionResult struct {
	Phase             routingPhase `json:"phase"`
	Tier              string       `json:"tier"`
	Model             string       `json:"model"`
	ReasoningEffort   string       `json:"reasoning_effort"`
	RuleID            string       `json:"rule_id"`
	ReasonCodes       []string     `json:"reason_codes,omitempty"`
	Status            string       `json:"status"`
	FallbackReason    string       `json:"fallback_reason,omitempty"`
	RequiredSol       bool         `json:"required_sol"`
	RemainingSolTurns int          `json:"remaining_sol_turns,omitempty"`
	ReadOnly          bool         `json:"read_only"`
}

func modelRoutingDecision(input modelRoutingInput) modelRoutingDecisionResult {
	decision := modelRoutingDecisionResult{
		Phase:             input.Phase,
		Tier:              "standard",
		Model:             modelRoutingModelTerra,
		ReasoningEffort:   "medium",
		RuleID:            "standard-default-v1",
		Status:            modelRoutingStatusPlanned,
		RemainingSolTurns: maxInt(input.RemainingSolTurns, 0),
	}

	if isSimpleLunaImplementation(input) {
		decision.Tier = "efficient"
		decision.Model = modelRoutingModelLuna
		decision.RuleID = "luna-simple-implementation-v1"
		decision.ReasonCodes = []string{"single_subsystem", "explicit_transformation", "reversible", "deterministic_acceptance"}
		return decision
	}

	// Sol is a decision/review tool, never an implementation model. Repair is
	// deliberately Terra unless the explicit repeated-failure risk predicate is
	// met below.
	if input.Phase != routingPhaseImplement && input.Phase != routingPhaseRepair {
		if input.CriticalRisk && input.HighAmbiguity && (input.CrossSystem || input.Irreversible) {
			return resolveSolDecision(input, "sol-high-risk-decision-v1", "high", true, []string{"critical_risk", "high_ambiguity", riskScopeReason(input)})
		}
		if shouldUseSolMedium(input) {
			required := input.CrossSystem && input.Phase == routingPhaseArchitecture
			return resolveSolDecision(input, "sol-medium-design-or-architecture-v1", "medium", required, []string{"initial_design_or_architecture", riskScopeReason(input)})
		}
	}

	if input.Phase == routingPhaseRepair && input.RepairCount > 0 {
		decision.ReasoningEffort = "high"
		decision.RuleID = "terra-repair-v1"
		decision.ReasonCodes = []string{"repair_attempt"}
		if input.CriticalRisk && input.HighAmbiguity && (input.CrossSystem || input.Irreversible) {
			return resolveSolDecision(input, "sol-high-repeated-risk-diagnosis-v1", "high", true, []string{"repeated_failure", "critical_risk", "high_ambiguity", riskScopeReason(input)})
		}
	}

	if input.Phase == routingPhaseImplement && (input.CrossSystem || input.Irreversible || input.CriticalRisk) {
		decision.ReasonCodes = []string{"implementation_requires_writer", "terra_writer"}
	}
	return decision
}

func isSimpleLunaImplementation(input modelRoutingInput) bool {
	return input.Phase == routingPhaseImplement && input.SimpleImplementation && input.SingleSubsystem && input.ExplicitTransformation && input.Reversible && input.DeterministicAcceptance && !input.PublicContractChange && !input.UnresolvedReview
}

func shouldUseSolMedium(input modelRoutingInput) bool {
	if input.Phase != routingPhasePlan && input.Phase != routingPhaseDesign && input.Phase != routingPhaseArchitecture && input.Phase != routingPhaseReview {
		return false
	}
	role := strings.TrimSpace(strings.ToLower(input.Role))
	return input.CrossSystem || input.Irreversible || input.Phase == routingPhaseArchitecture || input.Phase == routingPhaseDesign || strings.Contains(role, "planner") || strings.Contains(role, "architect") || strings.Contains(role, "designer")
}

func resolveSolDecision(input modelRoutingInput, rule, effort string, required bool, reasons []string) modelRoutingDecisionResult {
	decision := modelRoutingDecisionResult{
		Phase:             input.Phase,
		Tier:              "deep",
		Model:             modelRoutingModelSol,
		ReasoningEffort:   effort,
		RuleID:            rule,
		ReasonCodes:       reasons,
		Status:            modelRoutingStatusPlanned,
		RequiredSol:       required,
		RemainingSolTurns: maxInt(input.RemainingSolTurns, 0),
		ReadOnly:          true,
	}
	if input.RemainingSolTurns < 0 {
		// Keep the decision pure; callers that do not model a budget use zero as
		// an unlimited/unknown value. A negative value is the explicit exhausted
		// sentinel used by tests and turn planners.
		return fallbackOrBlockSol(decision, required, "sol_turn_budget_exhausted")
	}
	if input.SolAvailable != nil && !*input.SolAvailable {
		return fallbackOrBlockSol(decision, required, "model_unavailable")
	}
	return decision
}

func fallbackOrBlockSol(decision modelRoutingDecisionResult, required bool, reason string) modelRoutingDecisionResult {
	decision.FallbackReason = reason
	if required {
		decision.Status = modelRoutingStatusBlocked
		decision.RuleID += "-blocked"
		return decision
	}
	decision.Tier = "standard"
	decision.Model = modelRoutingModelTerra
	decision.ReasoningEffort = "high"
	decision.Status = modelRoutingStatusFallback
	decision.ReadOnly = false
	decision.RuleID += "-fallback-terra"
	return decision
}

func riskScopeReason(input modelRoutingInput) string {
	if input.CrossSystem {
		return "cross_system"
	}
	if input.Irreversible {
		return "irreversible"
	}
	return "architectural_judgment"
}
