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
	Checklist    []string
}

type specReviewState struct {
	Template     specReviewTemplate
	Status       string
	LastReviewed string
	Reviewer     string
}

type specReviewReadinessBatch struct {
	Outputs    map[string]string
	Advisories map[string]string
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
			Focus:        "Clarify art direction, palette discipline, anti-generic composition, purposeful motion, and visual risks before implementation starts.",
			Checklist: []string{
				"Art direction is clear and fits the task context.",
				"Palette temperature and undertone logic are coherent, saturation stays restrained, and the result does not collapse into washed-out gray minimalism.",
				"Semantic components and layout primitives match the content instead of defaulting to generic cards, border-heavy framing, or bento/grid fallback.",
				"Motion, if proposed, has a concrete hierarchy, attention, or state-change purpose.",
				"The most generic section is redesigned when the task is page-, screen-, or section-scale; component-scale tasks call out the risk without gratuitous scope creep.",
				"Anti-overcorrection guardrails hold: no novelty for novelty's sake, no decorative asymmetry without payoff, and no loss of accessibility, design-system fit, or implementation realism.",
			},
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
	lines := []string{
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
	}
	if len(state.Template.Checklist) > 0 {
		lines = append(lines, "## Review Checklist", "")
		for _, item := range state.Template.Checklist {
			lines = append(lines, "- "+item)
		}
		lines = append(lines, "")
	}
	lines = append(lines,
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
	)
	return strings.Join(lines, "\n")
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
	if harness := specHarnessAdvisorySummary(root, specID); harness != "" {
		blockers = append(blockers, "harness="+harness)
	}
	if len(blockers) == 0 {
		lines = append(lines, "- Advisory status: all current review tracks are marked clear.")
	} else {
		lines = append(lines, fmt.Sprintf("- Advisory status: follow up on %s before execution or GitHub handoff if the risk profile justifies it.", strings.Join(blockers, ", ")))
	}
	if harnessLines := buildHarnessReadinessLines(root, specID); len(harnessLines) > 0 {
		lines = append(lines,
			"",
			"## Harness Advisory",
			"",
		)
		lines = append(lines, harnessLines...)
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

func buildHarnessReadinessLines(root, specID string) []string {
	req, err := loadHarnessRequest(root, specID)
	if err != nil {
		return []string{fmt.Sprintf("- Harness request: invalid (%s)", err.Error())}
	}
	if req == nil {
		return nil
	}

	report := validateHarnessEvidence(root, specID, *req)
	lines := []string{
		fmt.Sprintf("- Route: `%s`", report.Route),
		fmt.Sprintf("- Request kind: `%s`", report.Request.RequestKind),
		fmt.Sprintf("- Delivery mode: `%s`", report.Request.DeliveryMode),
		fmt.Sprintf("- Adaptation mode: `%s`", report.Request.AdaptationMode),
		fmt.Sprintf("- Base contract ref: `%s`", fallbackOrValue(report.Request.BaseContractRef, "none")),
		fmt.Sprintf("- Touches Namba core: `%t`", report.Request.TouchesNambaCore),
		fmt.Sprintf("- Required evidence: `%s`", strings.Join(report.RequiredEvidence, ", ")),
	}
	if len(report.MissingEvidence) == 0 {
		lines = append(lines, "- Evidence status: complete")
	} else {
		lines = append(lines, fmt.Sprintf("- Missing evidence: `%s`", strings.Join(report.MissingEvidence, "`, `")))
	}
	lines = append(lines, fmt.Sprintf("- Required reviews: `%s`", strings.Join(report.RequiredReviews, ", ")))
	if len(report.MissingReviews) == 0 {
		lines = append(lines, "- Review artifact status: complete")
	} else {
		lines = append(lines, fmt.Sprintf("- Missing review artifacts: `%s`", strings.Join(report.MissingReviews, "`, `")))
	}
	if len(report.Problems) > 0 {
		lines = append(lines, fmt.Sprintf("- Problems: %s", strings.Join(report.Problems, "; ")))
	}
	return lines
}

func specHarnessAdvisorySummary(root, specID string) string {
	req, err := loadHarnessRequest(root, specID)
	if err != nil {
		return fmt.Sprintf("harness request invalid: %s", err.Error())
	}
	if req == nil {
		return ""
	}
	report := validateHarnessEvidence(root, specID, *req)
	if len(report.MissingEvidence) == 0 && len(report.MissingReviews) == 0 && len(report.Problems) == 0 {
		return ""
	}
	parts := []string{fmt.Sprintf("route=%s", report.Route)}
	if len(report.MissingEvidence) > 0 {
		parts = append(parts, fmt.Sprintf("missing evidence=%s", strings.Join(report.MissingEvidence, ",")))
	} else {
		parts = append(parts, "evidence=complete")
	}
	if len(report.MissingReviews) > 0 {
		parts = append(parts, fmt.Sprintf("missing review artifacts=%s", strings.Join(report.MissingReviews, ",")))
	}
	if len(report.Problems) > 0 {
		parts = append(parts, fmt.Sprintf("problems=%s", strings.Join(report.Problems, ",")))
	}
	return strings.Join(parts, ", ")
}

func fallbackOrValue(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
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
		{label: "Harness request", path: filepath.Join(specsDir, specID, harnessRequestFileName)},
		{label: "Eval plan", path: filepath.Join(specsDir, specID, "eval-plan.md")},
		{label: "Harness map", path: filepath.Join(specsDir, specID, "harness-map.md")},
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

func specReadinessAdvisorySummary(root, specID string) string {
	states := []specReviewState(nil)
	if specReviewReadinessExists(root, specID) || exists(filepath.Join(root, specsDir, specID, specReviewsDirName)) {
		states = loadSpecReviewStates(filepath.Join(root, specsDir, specID))
	}
	return specReadinessAdvisorySummaryFromStates(root, specID, states)
}

func specReadinessAdvisorySummaryFromStates(root, specID string, states []specReviewState) string {
	parts := make([]string, 0, 2)
	if review := specReviewAdvisorySummaryFromStates(states); review != "" {
		parts = append(parts, review)
	}
	if harness := specHarnessAdvisorySummary(root, specID); harness != "" {
		parts = append(parts, harness)
	}
	return strings.Join(parts, "; ")
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
	return loadSpecReviewStatesWithReadFile(os.ReadFile, specPath)
}

func loadSpecReviewStatesWithReadFile(readFile func(string) ([]byte, error), specPath string) []specReviewState {
	states := make([]specReviewState, 0, len(specReviewTemplates()))
	for _, template := range specReviewTemplates() {
		state := defaultSpecReviewState(template)
		body, err := readFile(filepath.Join(specPath, specReviewsDirName, template.Slug+".md"))
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
	return specReviewAdvisorySummaryFromStates(states)
}

func specReviewAdvisorySummaryFromStates(states []specReviewState) string {
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
	outputs, advisory, ok := buildSpecReviewReadinessOutput(root, specID)
	if !ok {
		return "", nil
	}
	if _, err := a.writeOutputs(root, outputs); err != nil {
		return "", err
	}
	return advisory, nil
}

func (a *App) refreshAllSpecReviewReadiness(root string) error {
	batch, err := buildSpecReviewReadinessBatch(root)
	if err != nil {
		return nil
	}
	if len(batch.Outputs) == 0 {
		return nil
	}
	_, err = a.writeOutputs(root, batch.Outputs)
	return err
}

func buildSpecReviewReadinessOutput(root, specID string) (map[string]string, string, bool) {
	if strings.TrimSpace(specID) == "" {
		return nil, "", false
	}
	if !exists(filepath.Join(root, specsDir, specID, specReviewsDirName)) {
		return nil, "", false
	}
	states := loadSpecReviewStates(filepath.Join(root, specsDir, specID))
	outputs := map[string]string{
		specReviewReadinessPath(specID): buildSpecReviewReadinessDoc(root, specID, states),
	}
	return outputs, specReadinessAdvisorySummaryFromStates(root, specID, states), true
}

func buildSpecReviewReadinessBatch(root string) (specReviewReadinessBatch, error) {
	entries, err := os.ReadDir(filepath.Join(root, specsDir))
	if err != nil {
		return specReviewReadinessBatch{}, err
	}

	batch := specReviewReadinessBatch{
		Outputs:    map[string]string{},
		Advisories: map[string]string{},
	}
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "SPEC-") {
			continue
		}
		outputs, advisory, ok := buildSpecReviewReadinessOutput(root, entry.Name())
		if !ok {
			continue
		}
		for path, body := range outputs {
			batch.Outputs[path] = body
		}
		batch.Advisories[entry.Name()] = advisory
	}
	return batch, nil
}

func isSpecReviewReadinessManagedPath(rel string) bool {
	return strings.HasPrefix(rel, filepath.ToSlash(specsDir)+"/") && strings.HasSuffix(rel, "/"+specReviewsDirName+"/"+specReviewReadinessFileName)
}
