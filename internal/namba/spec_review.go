package namba

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	specReviewsDirName          = "reviews"
	specReviewReadinessFileName = "readiness.md"
	reviewStatusPending         = "pending"
)

type specReviewTemplate struct {
	Slug         string
	Title        string
	Skill        string
	ReviewerRole string
	Focus        string
}

type specReviewState struct {
	Template     specReviewTemplate
	Status       string
	LastReviewed string
	Reviewer     string
}

func specReviewTemplates() []specReviewTemplate {
	return []specReviewTemplate{
		{
			Slug:         "product",
			Title:        "Product Review",
			Skill:        "$namba-plan-pm-review",
			ReviewerRole: "namba-product-manager",
			Focus:        "Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.",
		},
		{
			Slug:         "engineering",
			Title:        "Engineering Review",
			Skill:        "$namba-plan-eng-review",
			ReviewerRole: "namba-planner",
			Focus:        "Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.",
		},
		{
			Slug:         "design",
			Title:        "Design Review",
			Skill:        "$namba-plan-design-review",
			ReviewerRole: "namba-designer",
			Focus:        "Clarify interaction quality, responsive states, accessibility, and visual direction before implementation starts.",
		},
	}
}

func specReviewOutputs(specID string) map[string]string {
	outputs := make(map[string]string)
	states := make([]specReviewState, 0, len(specReviewTemplates()))
	for _, template := range specReviewTemplates() {
		state := defaultSpecReviewState(template)
		states = append(states, state)
		outputs[specReviewPath(specID, template.Slug)] = buildSpecReviewDoc(state)
	}
	outputs[specReviewReadinessPath(specID)] = buildSpecReviewReadinessDoc("", specID, states)
	return outputs
}

func defaultSpecReviewState(template specReviewTemplate) specReviewState {
	return specReviewState{
		Template:     template,
		Status:       reviewStatusPending,
		LastReviewed: reviewStatusPending,
		Reviewer:     reviewStatusPending,
	}
}

func specReviewPath(specID, slug string) string {
	return filepath.ToSlash(filepath.Join(specsDir, specID, specReviewsDirName, slug+".md"))
}

func specReviewReadinessPath(specID string) string {
	return filepath.ToSlash(filepath.Join(specsDir, specID, specReviewsDirName, specReviewReadinessFileName))
}

func buildSpecReviewDoc(state specReviewState) string {
	return strings.Join([]string{
		fmt.Sprintf("# %s", state.Template.Title),
		"",
		fmt.Sprintf("- Status: %s", state.Status),
		fmt.Sprintf("- Last Reviewed: %s", state.LastReviewed),
		fmt.Sprintf("- Reviewer: %s", state.Reviewer),
		fmt.Sprintf("- Command Skill: `%s`", state.Template.Skill),
		fmt.Sprintf("- Recommended Role: `%s`", state.Template.ReviewerRole),
		"",
		"## Focus",
		"",
		fmt.Sprintf("- %s", state.Template.Focus),
		"",
		"## Findings",
		"",
		"- Pending.",
		"",
		"## Decisions",
		"",
		"- Pending.",
		"",
		"## Follow-ups",
		"",
		"- Pending.",
		"",
		"## Recommendation",
		"",
		"- Pending.",
		"",
	}, "\n")
}

func buildSpecReviewReadinessDoc(root, specID string, states []specReviewState) string {
	var (
		clearCount int
		blockers   []string
	)
	lines := []string{
		"# Review Readiness",
		"",
		fmt.Sprintf("SPEC: %s", specID),
		"",
		"Advisory only: use this summary to decide whether the SPEC has enough pre-implementation review depth before `namba run` or `namba pr`. Missing reviews should be visible, not silently blocking.",
		"",
		"## Review Tracks",
		"",
	}
	for _, state := range states {
		lines = append(lines,
			fmt.Sprintf("- %s", state.Template.Title),
			fmt.Sprintf("  Status: %s", state.Status),
			fmt.Sprintf("  Last Reviewed: %s", state.LastReviewed),
			fmt.Sprintf("  Reviewer: %s", state.Reviewer),
			fmt.Sprintf("  Skill: `%s`", state.Template.Skill),
			fmt.Sprintf("  Artifact: `%s`", specReviewPath(specID, state.Template.Slug)),
		)
		if isClearReviewStatus(state.Status) {
			clearCount++
		} else {
			blockers = append(blockers, fmt.Sprintf("%s=%s", state.Template.Slug, state.Status))
		}
	}
	lines = append(lines,
		"",
		"## Summary",
		"",
		fmt.Sprintf("- Cleared reviews: %d/%d", clearCount, len(states)),
	)
	if len(blockers) == 0 {
		lines = append(lines, "- Advisory status: all current review tracks are marked clear.")
	} else {
		lines = append(lines, fmt.Sprintf("- Advisory status: follow up on %s before execution or GitHub handoff if the risk profile justifies it.", strings.Join(blockers, ", ")))
	}
	if evidence := specReadinessEvidencePaths(root, specID); len(evidence) > 0 {
		lines = append(lines,
			"",
			"## Phase-1 Evidence",
			"",
		)
		for _, entry := range evidence {
			lines = append(lines, "- "+entry)
		}
	}
	lines = append(lines,
		"",
		"## Suggested Order",
		"",
		"1. Run product review when the user/problem framing or scope is still moving.",
		"2. Run engineering review before implementation starts on anything with architecture or failure-mode risk.",
		"3. Run design review when UX, interaction quality, or visual direction matters to acceptance.",
		"",
	)
	return strings.Join(lines, "\n")
}

func specReadinessEvidencePaths(root, specID string) []string {
	if strings.TrimSpace(root) == "" {
		return nil
	}

	type evidenceFile struct {
		label string
		path  string
	}
	candidates := []evidenceFile{
		{label: "Runtime contract anchor", path: filepath.Join(specsDir, specID, "contract.md")},
		{label: "Baseline evidence", path: filepath.Join(specsDir, specID, "baseline.md")},
		{label: "Extraction map", path: filepath.Join(specsDir, specID, "extraction-map.md")},
	}

	var lines []string
	for _, candidate := range candidates {
		abs := filepath.Join(root, filepath.FromSlash(candidate.path))
		if !exists(abs) {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: `%s`", candidate.label, filepath.ToSlash(candidate.path)))
	}
	return lines
}

func isClearReviewStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "clear", "cleared", "approved", "pass", "passed":
		return true
	default:
		return false
	}
}

func loadSpecReviewStates(specPath string) []specReviewState {
	states := make([]specReviewState, 0, len(specReviewTemplates()))
	for _, template := range specReviewTemplates() {
		state := defaultSpecReviewState(template)
		body, err := os.ReadFile(filepath.Join(specPath, specReviewsDirName, template.Slug+".md"))
		if err != nil {
			if os.IsNotExist(err) {
				state.Status = "missing"
				state.LastReviewed = "missing"
				state.Reviewer = "missing"
				states = append(states, state)
				continue
			}
			state.Status = "unknown"
			state.LastReviewed = "unknown"
			state.Reviewer = "unknown"
			states = append(states, state)
			continue
		}
		text := string(body)
		state.Status = parseSpecReviewField(text, "- Status:", state.Status)
		state.LastReviewed = parseSpecReviewField(text, "- Last Reviewed:", state.LastReviewed)
		state.Reviewer = parseSpecReviewField(text, "- Reviewer:", state.Reviewer)
		states = append(states, state)
	}
	return states
}

func parseSpecReviewField(text, prefix, fallback string) string {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
			if value != "" {
				return value
			}
			return fallback
		}
	}
	return fallback
}

func specReviewReadinessExists(root, specID string) bool {
	if strings.TrimSpace(specID) == "" {
		return false
	}
	return exists(filepath.Join(root, filepath.FromSlash(specReviewReadinessPath(specID))))
}

func specReviewAdvisorySummary(root, specID string) string {
	if !specReviewReadinessExists(root, specID) {
		return ""
	}
	states := loadSpecReviewStates(filepath.Join(root, specsDir, specID))
	var pending []string
	for _, state := range states {
		if !isClearReviewStatus(state.Status) {
			pending = append(pending, fmt.Sprintf("%s=%s", state.Template.Slug, state.Status))
		}
	}
	if len(pending) == 0 {
		return "all review tracks clear"
	}
	return strings.Join(pending, ", ")
}

func (a *App) refreshSpecReviewReadiness(root, specID string) (string, error) {
	if strings.TrimSpace(specID) == "" {
		return "", nil
	}
	if !exists(filepath.Join(root, specsDir, specID, specReviewsDirName)) {
		return "", nil
	}
	states := loadSpecReviewStates(filepath.Join(root, specsDir, specID))
	outputs := map[string]string{
		specReviewReadinessPath(specID): buildSpecReviewReadinessDoc(root, specID, states),
	}
	if _, err := a.writeOutputs(root, outputs); err != nil {
		return "", err
	}
	return specReviewAdvisorySummary(root, specID), nil
}

func (a *App) refreshAllSpecReviewReadiness(root string) error {
	entries, err := os.ReadDir(filepath.Join(root, specsDir))
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "SPEC-") {
			continue
		}
		if !exists(filepath.Join(root, specsDir, entry.Name(), specReviewsDirName)) {
			continue
		}
		if _, err := a.refreshSpecReviewReadiness(root, entry.Name()); err != nil {
			return err
		}
	}
	return nil
}
