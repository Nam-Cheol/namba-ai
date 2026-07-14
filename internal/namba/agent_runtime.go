package namba

import (
	"fmt"
	"strings"
)

type agentRuntimeProfile struct {
	Role                 string `json:"role,omitempty"`
	Model                string `json:"model,omitempty"`
	ModelReasoningEffort string `json:"model_reasoning_effort,omitempty"`
}

func runtimeProfileForAgent(role string) agentRuntimeProfile {
	profile := agentRuntimeProfile{Role: strings.TrimSpace(role)}
	switch profile.Role {
	case "namba-explorer", "namba-simple-implementer":
		profile.Model = modelRoutingModelLuna
		profile.ModelReasoningEffort = "low"
	case "namba-product-manager":
		profile.Model = modelRoutingModelTerra
		profile.ModelReasoningEffort = "medium"
	case "namba-planner", "namba-frontend-architect", "namba-designer", "namba-backend-architect":
		profile.Model = modelRoutingModelSol
		profile.ModelReasoningEffort = "medium"
	case "namba-high-risk-advisor":
		profile.Model = modelRoutingModelSol
		profile.ModelReasoningEffort = "high"
	case "namba-plan-reviewer", "namba-mobile-engineer", "namba-security-engineer", "namba-reviewer", "namba-frontend-implementer", "namba-backend-implementer", "namba-data-engineer", "namba-test-engineer", "namba-devops-engineer", "namba-implementer":
		profile.Model = modelRoutingModelTerra
		profile.ModelReasoningEffort = "medium"
	}
	return profile
}

func runtimeProfilesForRoles(roles []string) []agentRuntimeProfile {
	profiles := make([]agentRuntimeProfile, 0, len(roles))
	seen := make(map[string]bool, len(roles))
	for _, role := range roles {
		profile := runtimeProfileForAgent(role)
		if strings.TrimSpace(profile.Role) == "" || seen[profile.Role] {
			continue
		}
		seen[profile.Role] = true
		profiles = append(profiles, profile)
	}
	return profiles
}

func formatAgentRuntimeProfile(profile agentRuntimeProfile) string {
	role := strings.TrimSpace(profile.Role)
	if role == "" {
		return ""
	}

	details := make([]string, 0, 2)
	if strings.TrimSpace(profile.Model) != "" {
		details = append(details, fmt.Sprintf("model `%s`", profile.Model))
	}
	if strings.TrimSpace(profile.ModelReasoningEffort) != "" {
		details = append(details, fmt.Sprintf("model_reasoning_effort `%s`", profile.ModelReasoningEffort))
	}
	if len(details) == 0 {
		return fmt.Sprintf("`%s`", role)
	}
	return fmt.Sprintf("`%s` -> %s", role, strings.Join(details, ", "))
}
