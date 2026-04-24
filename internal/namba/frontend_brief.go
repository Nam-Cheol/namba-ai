package namba

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const (
	frontendBriefFileName = "frontend-brief.md"

	frontendTaskClassificationMajor = "frontend-major"
	frontendTaskClassificationMinor = "frontend-minor"

	frontendGateStatusApproved      = "approved"
	frontendGateStatusBlocked       = "blocked"
	frontendGateStatusNeedsResearch = "needs-research"
	frontendGateStatusNotApplicable = "not-applicable"

	frontendGateStateComplete      = "complete"
	frontendGateStateMissing       = "missing"
	frontendGateStateInsufficient  = "insufficient"
	frontendGateStateNotApplicable = "not-applicable"

	frontendEvidenceStatusComplete      = "complete"
	frontendEvidenceStatusMissing       = "missing"
	frontendEvidenceStatusInsufficient  = "insufficient"
	frontendEvidenceStatusNotApplicable = "not-applicable"
	frontendEvidenceStatusInvalid       = "invalid-contract"
)

var frontendBriefRequiredLabels = []string{
	"Task Classification",
	"Classification Rationale",
	"Frontend Gate Status",
	"Problem Gate",
	"Reference Gate",
	"Critique Gate",
	"Decision Gate",
	"Prototype Gate",
	"Prototype Evidence",
}

var frontendBriefAllowedLabels = map[string]bool{
	"Task Classification":      true,
	"Classification Rationale": true,
	"Frontend Gate Status":     true,
	"Problem Gate":             true,
	"Reference Gate":           true,
	"Critique Gate":            true,
	"Decision Gate":            true,
	"Prototype Gate":           true,
	"Prototype Evidence":       true,
}

type frontendBriefHeader struct {
	TaskClassification      string
	ClassificationRationale string
	FrontendGateStatus      string
	ProblemGate             string
	ReferenceGate           string
	CritiqueGate            string
	DecisionGate            string
	PrototypeGate           string
	PrototypeEvidence       string
}

type frontendDesignReviewSummary struct {
	EvidenceStatus      string
	GateDecision        string
	ApprovedDirection   string
	BannedPatterns      string
	OpenQuestions       string
	UnresolvedQuestions string
}

type frontendBriefReport struct {
	Exists            bool
	Path              string
	Header            frontendBriefHeader
	Valid             bool
	ContractStatus    string
	EvidenceStatus    string
	MissingLabels     []string
	ContractIssues    []string
	MissingGates      []string
	InsufficientGates []string
	Mismatches        []string
	DesignReview      frontendDesignReviewSummary
}

func frontendBriefPath(specID string) string {
	return filepath.ToSlash(filepath.Join(specsDir, specID, frontendBriefFileName))
}

func buildFrontendBriefDoc(kind, description string) (string, bool) {
	classification, rationale, ok := inferFrontendTaskClassification(kind, description)
	if !ok {
		return "", false
	}
	if classification == frontendTaskClassificationMinor {
		return buildFrontendMinorBriefDoc(rationale), true
	}
	return buildFrontendMajorBriefDoc(rationale), true
}

func inferFrontendTaskClassification(kind, description string) (string, string, bool) {
	text := strings.ToLower(strings.TrimSpace(description))
	if text == "" {
		return "", "", false
	}

	touchHits := findFrontendKeywordHits(text, frontendTouchKeywords())
	majorHits := findFrontendKeywordHits(text, frontendMajorKeywords())
	minorHits := findFrontendKeywordHits(text, frontendMinorKeywords())
	if len(touchHits) == 0 && len(majorHits) == 0 && len(minorHits) == 0 {
		return "", "", false
	}
	if len(majorHits) == 0 && len(minorHits) == 0 && !hasExplicitFrontendTouchSignal(touchHits) {
		return "", "", false
	}

	if isFixOnlyFrontendMinor(kind, majorHits, minorHits) {
		return frontendTaskClassificationMinor, fmt.Sprintf("Matched lightweight frontend fix signals: %s.", quoteList(minorHits)), true
	}
	if len(majorHits) > 0 {
		return frontendTaskClassificationMajor, fmt.Sprintf("Matched frontend-major signals: %s.", quoteList(majorHits)), true
	}
	if len(minorHits) > 0 {
		return frontendTaskClassificationMinor, fmt.Sprintf("Matched lightweight frontend signals: %s.", quoteList(minorHits)), true
	}
	if strings.TrimSpace(kind) == "fix" {
		return frontendTaskClassificationMinor, "Frontend-touching fix work defaults to `frontend-minor` when no major redesign signal is present.", true
	}
	return frontendTaskClassificationMajor, "Frontend-touching feature work defaults to `frontend-major` unless the change is clearly minor.", true
}

func hasExplicitFrontendTouchSignal(hits []string) bool {
	explicit := map[string]bool{
		"ui":            true,
		"screen":        true,
		"page":          true,
		"dashboard":     true,
		"landing":       true,
		"hero":          true,
		"layout":        true,
		"responsive":    true,
		"browser":       true,
		"css":           true,
		"typography":    true,
		"spacing":       true,
		"alignment":     true,
		"button":        true,
		"navigation":    true,
		"sidebar":       true,
		"header":        true,
		"footer":        true,
		"modal":         true,
		"dialog":        true,
		"visual":        true,
		"section":       true,
		"a11y":          true,
		"accessibility": true,
	}
	for _, hit := range hits {
		if explicit[hit] {
			return true
		}
	}
	return false
}

func isFixOnlyFrontendMinor(kind string, majorHits, minorHits []string) bool {
	if strings.TrimSpace(kind) != "fix" || len(minorHits) == 0 {
		return false
	}
	for _, hit := range majorHits {
		switch hit {
		case "landing page", "redesign", "restructure", "new screen", "new page", "new section", "primary workflow", "interaction model", "visual tone", "hierarchy":
			return false
		}
	}
	return true
}

func frontendTouchKeywords() []string {
	return []string{
		"ui",
		"screen",
		"page",
		"dashboard",
		"landing",
		"hero",
		"layout",
		"component",
		"responsive",
		"browser",
		"css",
		"typography",
		"spacing",
		"alignment",
		"button",
		"settings",
		"form",
		"navigation",
		"sidebar",
		"header",
		"footer",
		"modal",
		"dialog",
		"visual",
		"section",
		"a11y",
		"accessibility",
	}
}

func findFrontendKeywordHits(text string, keywords []string) []string {
	normalizedText := normalizeFrontendKeywordText(text)
	hits := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		normalizedKeyword := strings.TrimSpace(normalizeFrontendKeywordText(keyword))
		if normalizedKeyword == "" {
			continue
		}
		if strings.Contains(normalizedText, " "+normalizedKeyword+" ") {
			hits = append(hits, keyword)
		}
	}
	return uniqueStrings(hits)
}

func normalizeFrontendKeywordText(text string) string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	if len(fields) == 0 {
		return " "
	}
	return " " + strings.Join(fields, " ") + " "
}

func frontendMajorKeywords() []string {
	return []string{
		"landing page",
		"dashboard",
		"hero",
		"redesign",
		"restructure",
		"new screen",
		"new page",
		"new section",
		"primary workflow",
		"interaction model",
		"visual tone",
		"hierarchy",
	}
}

func frontendMinorKeywords() []string {
	return []string{
		"spacing",
		"alignment",
		"copy",
		"button",
		"existing screen",
		"existing component",
		"bug fix",
		"padding",
		"margin",
		"text",
	}
}

func buildFrontendMajorBriefDoc(rationale string) string {
	lines := []string{
		"# Frontend Brief",
		"",
		fmt.Sprintf("Task Classification: %s", frontendTaskClassificationMajor),
		fmt.Sprintf("Classification Rationale: %s", firstNonBlank(strings.TrimSpace(rationale), "Pending classification rationale.")),
		fmt.Sprintf("Frontend Gate Status: %s", frontendGateStatusNeedsResearch),
		fmt.Sprintf("Problem Gate: %s", frontendGateStateMissing),
		fmt.Sprintf("Reference Gate: %s", frontendGateStateMissing),
		fmt.Sprintf("Critique Gate: %s", frontendGateStateMissing),
		fmt.Sprintf("Decision Gate: %s", frontendGateStateMissing),
		fmt.Sprintf("Prototype Gate: %s", frontendGateStateMissing),
		"Prototype Evidence: n/a",
		"",
		"## Problem Frame",
		"",
		"- Problem statement: Pending.",
		"- User goal: Pending.",
		"- Target user: Pending.",
		"- Success metric: Pending.",
		"- Why now: Pending.",
		"- Scope boundary: Pending.",
		"",
		"## Asset Evidence",
		"",
		"- Brand assets: Pending.",
		"- Product or domain imagery: Pending.",
		"- Existing UI screenshots: Pending.",
		"- Asset constraints and gaps: Pending.",
		"",
		"## Reference Set",
		"",
		"- Reference 1: Pending.",
		"  Adopt: Pending.",
		"  Avoid: Pending.",
		"  Why: Pending.",
		"- Reference 2: Pending.",
		"  Adopt: Pending.",
		"  Avoid: Pending.",
		"  Why: Pending.",
		"- Reference 3: Pending.",
		"  Adopt: Pending.",
		"  Avoid: Pending.",
		"  Why: Pending.",
		"",
		"## Direction Alternatives",
		"",
		"- Direction A: Pending.",
		"  Tradeoff: Pending.",
		"- Direction B: Pending.",
		"  Tradeoff: Pending.",
		"- Direction C: Pending.",
		"  Tradeoff: Pending.",
		"- Selected direction rationale: Pending.",
		"",
		"## Synthesis",
		"",
		"- UX metaphor: Pending.",
		"- Section roles: Pending.",
		"- Hierarchy: Pending.",
		"- Reference synthesis: Pending.",
		"- Anti-generic bans: Pending.",
		"- Typography scale: Pending.",
		"- Spacing and density intent: Pending.",
		"- Depth and container budget: Pending.",
		"",
		"## Design Review Axes",
		"",
		"- Evidence fit: Pending.",
		"- Asset fidelity: Pending.",
		"- Alternative coverage: Pending.",
		"- Visual hierarchy: Pending.",
		"- Craft and detail: Pending.",
		"- Functionality and accessibility: Pending.",
		"- Differentiation without novelty drift: Pending.",
		"",
		"## Prototype Evidence",
		"",
		"- Artifact path or link: Pending.",
		"- Notes: Pending.",
		"",
		"## Open Decisions",
		"",
		"- Pending.",
		"",
	}
	return strings.Join(lines, "\n")
}

func buildFrontendMinorBriefDoc(rationale string) string {
	lines := []string{
		"# Frontend Brief",
		"",
		fmt.Sprintf("Task Classification: %s", frontendTaskClassificationMinor),
		fmt.Sprintf("Classification Rationale: %s", firstNonBlank(strings.TrimSpace(rationale), "Pending classification rationale.")),
		fmt.Sprintf("Frontend Gate Status: %s", frontendGateStatusNotApplicable),
		fmt.Sprintf("Problem Gate: %s", frontendGateStateNotApplicable),
		fmt.Sprintf("Reference Gate: %s", frontendGateStateNotApplicable),
		fmt.Sprintf("Critique Gate: %s", frontendGateStateNotApplicable),
		fmt.Sprintf("Decision Gate: %s", frontendGateStateNotApplicable),
		fmt.Sprintf("Prototype Gate: %s", frontendGateStateNotApplicable),
		"Prototype Evidence: n/a",
		"",
		"## Current Pattern",
		"",
		"- Pending.",
		"",
		"## Intended Change",
		"",
		"- Pending.",
		"",
		"## Notes",
		"",
		"- Pending.",
		"",
	}
	return strings.Join(lines, "\n")
}

func loadFrontendBriefReport(root, specID string) frontendBriefReport {
	report := frontendBriefReport{
		Path: filepath.ToSlash(filepath.Join(specsDir, specID, frontendBriefFileName)),
	}
	if strings.TrimSpace(root) == "" || strings.TrimSpace(specID) == "" {
		return report
	}

	path := filepath.Join(root, filepath.FromSlash(report.Path))
	if !exists(path) {
		return report
	}

	body, err := os.ReadFile(path)
	if err != nil {
		report.Exists = true
		report.ContractStatus = frontendEvidenceStatusInvalid
		report.ContractIssues = []string{fmt.Sprintf("unable to read frontend brief: %v", err)}
		return report
	}

	report = parseFrontendBrief(string(body))
	report.Exists = true
	report.Path = filepath.ToSlash(filepath.Join(specsDir, specID, frontendBriefFileName))

	reviewPath := filepath.Join(root, filepath.FromSlash(specReviewPath(specID, "design")))
	if body, err := os.ReadFile(reviewPath); err == nil {
		compareFrontendBriefAndDesignReview(&report, string(body))
	} else if report.Valid && report.Header.TaskClassification == frontendTaskClassificationMajor {
		report.Mismatches = append(report.Mismatches, fmt.Sprintf("Design review artifact missing: `%s`", specReviewPath(specID, "design")))
	}
	report.Mismatches = uniqueStrings(report.Mismatches)
	return report
}

func parseFrontendBrief(text string) frontendBriefReport {
	report := frontendBriefReport{
		ContractStatus: frontendEvidenceStatusInvalid,
	}

	headerValues, missingLabels, issues := parseFrontendBriefHeader(text)
	report.MissingLabels = missingLabels
	report.ContractIssues = append(report.ContractIssues, issues...)
	report.Header = frontendBriefHeader{
		TaskClassification:      normalizeFrontendBriefEnum(headerValues["Task Classification"]),
		ClassificationRationale: strings.TrimSpace(headerValues["Classification Rationale"]),
		FrontendGateStatus:      normalizeFrontendBriefEnum(headerValues["Frontend Gate Status"]),
		ProblemGate:             normalizeFrontendBriefEnum(headerValues["Problem Gate"]),
		ReferenceGate:           normalizeFrontendBriefEnum(headerValues["Reference Gate"]),
		CritiqueGate:            normalizeFrontendBriefEnum(headerValues["Critique Gate"]),
		DecisionGate:            normalizeFrontendBriefEnum(headerValues["Decision Gate"]),
		PrototypeGate:           normalizeFrontendBriefEnum(headerValues["Prototype Gate"]),
		PrototypeEvidence:       normalizeFrontendBriefEnum(headerValues["Prototype Evidence"]),
	}

	if report.Header.ClassificationRationale == "" {
		report.ContractIssues = append(report.ContractIssues, "Classification Rationale must not be blank.")
	}

	validateFrontendBriefEnums(&report)
	validateFrontendBriefConsistency(&report)
	report.ContractIssues = uniqueStrings(report.ContractIssues)
	report.MissingGates = uniqueStrings(report.MissingGates)
	report.InsufficientGates = uniqueStrings(report.InsufficientGates)
	report.Mismatches = uniqueStrings(report.Mismatches)
	report.Valid = len(report.MissingLabels) == 0 && len(report.ContractIssues) == 0
	if report.Valid {
		report.ContractStatus = "valid"
	} else {
		report.ContractStatus = frontendEvidenceStatusInvalid
	}
	report.EvidenceStatus = deriveFrontendEvidenceStatus(report)
	return report
}

func parseFrontendBriefHeader(text string) (map[string]string, []string, []string) {
	lines := strings.Split(text, "\n")
	headerValues := map[string]string{}
	var issues []string
	started := false
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if !started {
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "# ") {
				started = true
				continue
			}
			started = true
		}
		if line == "" {
			if len(headerValues) == 0 && len(issues) == 0 {
				continue
			}
			break
		}
		if strings.HasPrefix(line, "## ") {
			break
		}
		label, value, ok := strings.Cut(line, ":")
		if !ok {
			issues = append(issues, fmt.Sprintf("unexpected fixed-label line %q", line))
			continue
		}
		label = strings.TrimSpace(label)
		value = strings.TrimSpace(value)
		if !frontendBriefAllowedLabels[label] {
			issues = append(issues, fmt.Sprintf("unknown fixed-label %q", label))
			continue
		}
		if _, exists := headerValues[label]; exists {
			issues = append(issues, fmt.Sprintf("duplicate fixed-label %q", label))
			continue
		}
		headerValues[label] = value
	}

	var missing []string
	for _, label := range frontendBriefRequiredLabels {
		if _, ok := headerValues[label]; !ok {
			missing = append(missing, label)
		}
	}
	return headerValues, missing, issues
}

func normalizeFrontendBriefEnum(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validateFrontendBriefEnums(report *frontendBriefReport) {
	checkEnum := func(label, value string, allowed ...string) {
		if value == "" {
			return
		}
		for _, candidate := range allowed {
			if value == candidate {
				return
			}
		}
		report.ContractIssues = append(report.ContractIssues, fmt.Sprintf("%s has unsupported value %q", label, value))
	}

	checkEnum("Task Classification", report.Header.TaskClassification, frontendTaskClassificationMajor, frontendTaskClassificationMinor)
	checkEnum("Frontend Gate Status", report.Header.FrontendGateStatus, frontendGateStatusApproved, frontendGateStatusBlocked, frontendGateStatusNeedsResearch, frontendGateStatusNotApplicable)
	for label, value := range map[string]string{
		"Problem Gate":   report.Header.ProblemGate,
		"Reference Gate": report.Header.ReferenceGate,
		"Critique Gate":  report.Header.CritiqueGate,
		"Decision Gate":  report.Header.DecisionGate,
		"Prototype Gate": report.Header.PrototypeGate,
	} {
		checkEnum(label, value, frontendGateStateComplete, frontendGateStateMissing, frontendGateStateInsufficient, frontendGateStateNotApplicable)
	}
	checkEnum("Prototype Evidence", report.Header.PrototypeEvidence, "wireframe", "annotated-layout", "prototype", "equivalent", "n/a")
}

func validateFrontendBriefConsistency(report *frontendBriefReport) {
	gates := []struct {
		Name  string
		Value string
	}{
		{Name: "Problem Gate", Value: report.Header.ProblemGate},
		{Name: "Reference Gate", Value: report.Header.ReferenceGate},
		{Name: "Critique Gate", Value: report.Header.CritiqueGate},
		{Name: "Decision Gate", Value: report.Header.DecisionGate},
		{Name: "Prototype Gate", Value: report.Header.PrototypeGate},
	}

	for _, gate := range gates {
		switch gate.Value {
		case frontendGateStateMissing:
			report.MissingGates = append(report.MissingGates, gate.Name)
		case frontendGateStateInsufficient:
			report.InsufficientGates = append(report.InsufficientGates, gate.Name)
		}
	}

	switch report.Header.TaskClassification {
	case frontendTaskClassificationMinor:
		if report.Header.FrontendGateStatus != frontendGateStatusNotApplicable {
			report.ContractIssues = append(report.ContractIssues, "frontend-minor must use Frontend Gate Status: not-applicable.")
		}
		for _, gate := range gates {
			if gate.Value != frontendGateStateNotApplicable {
				report.ContractIssues = append(report.ContractIssues, fmt.Sprintf("frontend-minor must keep %s as not-applicable", gate.Name))
			}
		}
		if report.Header.PrototypeEvidence != "n/a" {
			report.ContractIssues = append(report.ContractIssues, "frontend-minor must use Prototype Evidence: n/a.")
		}
	case frontendTaskClassificationMajor:
		if report.Header.FrontendGateStatus == frontendGateStatusNotApplicable {
			report.ContractIssues = append(report.ContractIssues, "frontend-major cannot use Frontend Gate Status: not-applicable.")
		}
		for _, gate := range gates {
			if gate.Value == frontendGateStateNotApplicable {
				report.ContractIssues = append(report.ContractIssues, fmt.Sprintf("frontend-major cannot use %s: not-applicable", gate.Name))
			}
		}
		if report.Header.FrontendGateStatus == frontendGateStatusApproved && (len(report.MissingGates) > 0 || len(report.InsufficientGates) > 0) {
			report.ContractIssues = append(report.ContractIssues, "Frontend Gate Status: approved requires every major gate to be complete.")
		}
		if report.Header.PrototypeGate == frontendGateStateComplete && report.Header.PrototypeEvidence == "n/a" {
			report.ContractIssues = append(report.ContractIssues, "Prototype Gate: complete requires reviewable prototype evidence instead of n/a.")
		}
	}
}

func deriveFrontendEvidenceStatus(report frontendBriefReport) string {
	if !report.Valid {
		return frontendEvidenceStatusInvalid
	}
	switch report.Header.TaskClassification {
	case frontendTaskClassificationMinor:
		return frontendEvidenceStatusNotApplicable
	default:
		if len(report.InsufficientGates) > 0 {
			return frontendEvidenceStatusInsufficient
		}
		if len(report.MissingGates) > 0 {
			return frontendEvidenceStatusMissing
		}
		return frontendEvidenceStatusComplete
	}
}

func parseFrontendDesignReviewSummary(text string) frontendDesignReviewSummary {
	return frontendDesignReviewSummary{
		EvidenceStatus:      normalizeFrontendBriefEnum(parseSpecReviewField(text, "- Evidence Status:", "pending")),
		GateDecision:        normalizeFrontendBriefEnum(parseSpecReviewField(text, "- Gate Decision:", "pending")),
		ApprovedDirection:   strings.TrimSpace(parseSpecReviewField(text, "- Approved Direction:", "pending")),
		BannedPatterns:      strings.TrimSpace(parseSpecReviewField(text, "- Banned Patterns:", "pending")),
		OpenQuestions:       strings.TrimSpace(parseSpecReviewField(text, "- Open Questions:", "pending")),
		UnresolvedQuestions: strings.TrimSpace(parseSpecReviewField(text, "- Unresolved Questions:", "pending")),
	}
}

func compareFrontendBriefAndDesignReview(report *frontendBriefReport, reviewText string) {
	report.DesignReview = parseFrontendDesignReviewSummary(reviewText)
	if !report.Valid {
		return
	}
	gate := report.DesignReview.GateDecision
	evidence := report.DesignReview.EvidenceStatus
	if report.Header.TaskClassification == frontendTaskClassificationMajor {
		if gate == "" || gate == "pending" {
			report.Mismatches = append(report.Mismatches, "Design review gate decision is pending for frontend-major; design-review=pending")
		} else if gate != report.Header.FrontendGateStatus {
			report.Mismatches = append(report.Mismatches, fmt.Sprintf("Gate decision mismatch: frontend-brief=%s, design-review=%s", report.Header.FrontendGateStatus, gate))
		}
		if evidence == "" || evidence == "pending" {
			report.Mismatches = append(report.Mismatches, "Design review evidence status is pending for frontend-major; design-review=pending")
		} else if evidence != report.EvidenceStatus {
			report.Mismatches = append(report.Mismatches, fmt.Sprintf("Evidence status mismatch: frontend-brief=%s, design-review=%s", report.EvidenceStatus, evidence))
		}
		appendPendingDesignReviewDecisionFieldMismatches(report)
		return
	}
	if gate != "" && gate != "pending" && gate != report.Header.FrontendGateStatus {
		report.Mismatches = append(report.Mismatches, fmt.Sprintf("Gate decision mismatch: frontend-brief=%s, design-review=%s", report.Header.FrontendGateStatus, gate))
	}
	if evidence != "" && evidence != "pending" && evidence != report.EvidenceStatus {
		report.Mismatches = append(report.Mismatches, fmt.Sprintf("Evidence status mismatch: frontend-brief=%s, design-review=%s", report.EvidenceStatus, evidence))
	}
}

func appendPendingDesignReviewDecisionFieldMismatches(report *frontendBriefReport) {
	for _, field := range []struct {
		Message string
		Value   string
	}{
		{Message: "Design review approved direction is pending for frontend-major; design-review=pending", Value: report.DesignReview.ApprovedDirection},
		{Message: "Design review banned patterns are pending for frontend-major; design-review=pending", Value: report.DesignReview.BannedPatterns},
		{Message: "Design review open questions are pending for frontend-major; design-review=pending", Value: report.DesignReview.OpenQuestions},
		{Message: "Design review unresolved questions are pending for frontend-major; design-review=pending", Value: report.DesignReview.UnresolvedQuestions},
	} {
		if isPendingDesignReviewField(field.Value) {
			report.Mismatches = append(report.Mismatches, field.Message)
		}
	}
}

func isPendingDesignReviewField(value string) bool {
	normalized := normalizeFrontendBriefEnum(value)
	return normalized == "" || normalized == "pending"
}

func frontendGateReadinessLines(root, specID string) []string {
	report := loadFrontendBriefReport(root, specID)
	if !report.Exists {
		return nil
	}

	lines := []string{
		fmt.Sprintf("- Classification source: `%s`", frontendBriefPath(specID)),
	}
	if report.Valid {
		lines = append(lines,
			fmt.Sprintf("- Task Classification: `%s`", report.Header.TaskClassification),
			fmt.Sprintf("- Classification Rationale: %s", report.Header.ClassificationRationale),
			fmt.Sprintf("- Frontend Gate Status: `%s`", report.Header.FrontendGateStatus),
			fmt.Sprintf("- Evidence Status: `%s`", report.EvidenceStatus),
		)
		if report.Header.TaskClassification == frontendTaskClassificationMajor {
			lines = append(lines,
				fmt.Sprintf("- Problem Gate: `%s`", report.Header.ProblemGate),
				fmt.Sprintf("- Reference Gate: `%s`", report.Header.ReferenceGate),
				fmt.Sprintf("- Critique Gate: `%s`", report.Header.CritiqueGate),
				fmt.Sprintf("- Decision Gate: `%s`", report.Header.DecisionGate),
				fmt.Sprintf("- Prototype Gate: `%s`", report.Header.PrototypeGate),
				fmt.Sprintf("- Prototype Evidence: `%s`", report.Header.PrototypeEvidence),
			)
			if len(report.MissingGates) > 0 {
				lines = append(lines, fmt.Sprintf("- Missing gates: %s", quoteList(report.MissingGates)))
			}
			if len(report.InsufficientGates) > 0 {
				lines = append(lines, fmt.Sprintf("- Insufficient gates: %s", quoteList(report.InsufficientGates)))
			}
		} else {
			lines = append(lines, "- Gate mode: advisory passthrough for `frontend-minor`.")
		}
	} else {
		lines = append(lines,
			fmt.Sprintf("- Contract status: `%s`", report.ContractStatus),
			fmt.Sprintf("- Missing fixed labels: %s", quoteList(report.MissingLabels)),
			fmt.Sprintf("- Contract issues: %s", strings.Join(report.ContractIssues, "; ")),
		)
	}
	if len(report.Mismatches) == 0 {
		lines = append(lines, "- Cross-artifact mismatches: none.")
	} else {
		lines = append(lines, fmt.Sprintf("- Cross-artifact mismatches: %s", strings.Join(report.Mismatches, "; ")))
	}
	return lines
}

func frontendGateAdvisorySummary(root, specID string) string {
	report := loadFrontendBriefReport(root, specID)
	if !report.Exists {
		return ""
	}
	if !report.Valid {
		return "frontend=invalid-contract"
	}
	if len(report.Mismatches) > 0 {
		return "frontend=blocked"
	}
	if report.Header.TaskClassification == frontendTaskClassificationMinor {
		return "frontend=not-applicable"
	}
	if report.Header.FrontendGateStatus == frontendGateStatusApproved && report.EvidenceStatus == frontendEvidenceStatusComplete {
		return "frontend=approved"
	}
	return "frontend=" + firstNonBlank(report.Header.FrontendGateStatus, report.EvidenceStatus)
}

func frontendGateExecutionError(specID string, report frontendBriefReport) error {
	if !report.Valid {
		lines := []string{
			fmt.Sprintf("%s has an invalid frontend brief contract.", specID),
			fmt.Sprintf("Artifact: `%s`", report.Path),
		}
		if len(report.MissingLabels) > 0 {
			lines = append(lines, fmt.Sprintf("Missing fixed labels: %s.", quoteList(report.MissingLabels)))
		}
		for _, issue := range report.ContractIssues {
			lines = append(lines, "- "+issue)
		}
		lines = append(lines, "Fix the fixed-label header before running `namba run` again.")
		return errors.New(strings.Join(lines, "\n"))
	}

	lines := []string{
		fmt.Sprintf("%s is blocked for frontend synthesis.", specID),
		fmt.Sprintf("Artifact: `%s`", report.Path),
		fmt.Sprintf("Task Classification: `%s`", report.Header.TaskClassification),
		fmt.Sprintf("Frontend Gate Status: `%s`", report.Header.FrontendGateStatus),
		fmt.Sprintf("Evidence Status: `%s`", report.EvidenceStatus),
	}
	if len(report.MissingGates) > 0 {
		lines = append(lines, fmt.Sprintf("Missing gates: %s.", quoteList(report.MissingGates)))
	}
	if len(report.InsufficientGates) > 0 {
		lines = append(lines, fmt.Sprintf("Insufficient gates: %s.", quoteList(report.InsufficientGates)))
	}
	for _, mismatch := range report.Mismatches {
		lines = append(lines, "- "+mismatch)
	}
	lines = append(lines, "Next steps:")
	for _, step := range frontendGateRemediation(report) {
		lines = append(lines, "- "+step)
	}
	lines = append(lines, "- If independent non-frontend delivery matters, split this work into separate SPECs or explicit phases rather than expecting partial unblocking.")
	return errors.New(strings.Join(lines, "\n"))
}

func frontendGateRemediation(report frontendBriefReport) []string {
	var steps []string
	appendIf := func(condition bool, step string) {
		if condition {
			steps = append(steps, step)
		}
	}

	hasMissing := func(name string) bool {
		for _, gate := range report.MissingGates {
			if gate == name {
				return true
			}
		}
		return false
	}
	hasInsufficient := func(name string) bool {
		for _, gate := range report.InsufficientGates {
			if gate == name {
				return true
			}
		}
		return false
	}

	appendIf(hasMissing("Problem Gate"), "Clarify the problem frame, target user, and scope boundary in `frontend-brief.md`.")
	appendIf(hasMissing("Reference Gate"), "Gather or replace the reference set before implementation.")
	appendIf(hasInsufficient("Reference Gate"), "Strengthen weak reference synthesis with explicit adopt/avoid/why decisions.")
	appendIf(hasMissing("Critique Gate"), "Add critique notes that explain what works, what fails, and why.")
	appendIf(hasInsufficient("Critique Gate"), "Deepen the critique so weak observations become actionable design guidance.")
	appendIf(hasMissing("Decision Gate"), "Record the approved direction, banned patterns, and open decisions before coding.")
	appendIf(hasInsufficient("Decision Gate"), "Tighten the approved direction and banned-pattern guidance so implementation scope is explicit.")
	appendIf(hasMissing("Prototype Gate"), "Add reviewable prototype evidence such as a wireframe, annotated layout, or equivalent artifact.")
	appendIf(hasInsufficient("Prototype Gate"), "Replace weak prototype evidence with a clearer structure or interaction artifact.")
	if len(report.Mismatches) > 0 {
		steps = append(steps, "Reconcile `frontend-brief.md` with `reviews/design.md` so the canonical gate state and summaries agree.")
	}
	return uniqueStrings(steps)
}
