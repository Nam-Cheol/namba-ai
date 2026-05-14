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

	frontendNegativeContractStatusComplete      = "complete"
	frontendNegativeContractStatusMissing       = "missing"
	frontendNegativeContractStatusInsufficient  = "insufficient"
	frontendNegativeContractStatusNotApplicable = "not-applicable"

	frontendAssetModeGeneratedImages = "generated-images"
	frontendAssetModeExistingAssets  = "existing-assets"
	frontendAssetModeNotApplicable   = "not-applicable"

	frontendImplementationPhaseFirstFrontend    = "first-frontend"
	frontendImplementationPhaseFirstMajorScreen = "first-major-screen"
	frontendImplementationPhaseIncremental      = "incremental"
	frontendImplementationPhaseNotApplicable    = "not-applicable"

	frontendImagegenRequirementRequired                = "required"
	frontendImagegenRequirementCoveredByExistingAssets = "covered-by-existing-assets"
	frontendImagegenRequirementNotApplicable           = "not-applicable"

	frontendGenerationPlanStatusReady = "ready"
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
	"Task Classification":           true,
	"Classification Rationale":      true,
	"Frontend Gate Status":          true,
	"Problem Gate":                  true,
	"Reference Gate":                true,
	"Critique Gate":                 true,
	"Decision Gate":                 true,
	"Prototype Gate":                true,
	"Prototype Evidence":            true,
	"Frontend implementation phase": true,
	"Asset mode":                    true,
	"Imagegen requirement":          true,
	"Asset decision proof":          true,
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
	Exists                      bool
	Path                        string
	Header                      frontendBriefHeader
	Valid                       bool
	ContractStatus              string
	EvidenceStatus              string
	NegativeContractStatus      string
	AssetMode                   string
	GeneratedImagePlan          string
	FrontendImplementationPhase string
	ImagegenRequirement         string
	AssetDecisionProof          string
	MissingLabels               []string
	ContractIssues              []string
	NegativeContractIssues      []string
	MissingGates                []string
	InsufficientGates           []string
	Mismatches                  []string
	DesignReview                frontendDesignReviewSummary
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
	if isBackendOnlyAmbiguousFrontendReference(text, touchHits, majorHits, minorHits) {
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

func isBackendOnlyAmbiguousFrontendReference(text string, touchHits, majorHits, minorHits []string) bool {
	if !hasBackendImplementationSignal(text) {
		return false
	}
	if !hasOnlyBackendAmbiguousMajorHits(majorHits) {
		return false
	}
	if !hasOnlyBackendAmbiguousTouchHits(touchHits) {
		return false
	}
	if !hasOnlyBackendAmbiguousMinorHits(minorHits) {
		return false
	}
	return len(touchHits) > 0 || len(majorHits) > 0 || len(minorHits) > 0
}

func hasOnlyBackendAmbiguousMajorHits(hits []string) bool {
	for _, hit := range hits {
		switch hit {
		case "dashboard", "hierarchy":
		default:
			return false
		}
	}
	return true
}

func hasOnlyBackendAmbiguousTouchHits(hits []string) bool {
	for _, hit := range hits {
		switch hit {
		case "dashboard", "settings", "form", "component", "page":
		default:
			return false
		}
	}
	return true
}

func hasOnlyBackendAmbiguousMinorHits(hits []string) bool {
	for _, hit := range hits {
		switch hit {
		case "text", "copy":
		default:
			return false
		}
	}
	return true
}

func hasBackendImplementationSignal(text string) bool {
	return len(findFrontendKeywordHits(text, []string{
		"api",
		"endpoint",
		"backend",
		"server",
		"service",
		"database",
		"storage",
		"schema",
		"handler",
		"resolver",
		"migration",
	})) > 0
}

func hasExplicitFrontendTouchSignal(hits []string) bool {
	explicit := map[string]bool{
		"frontend":      true,
		"front end":     true,
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
	if strings.TrimSpace(kind) != "fix" {
		return false
	}
	for _, hit := range majorHits {
		if isStructuralFrontendMajorHit(hit) {
			return false
		}
	}
	return len(majorHits) > 0 || len(minorHits) > 0
}

func isStructuralFrontendMajorHit(hit string) bool {
	switch hit {
	case "landing page", "redesign", "restructure", "new screen", "new page", "new section", "primary workflow", "interaction model", "visual tone", "hierarchy":
		return true
	default:
		return false
	}
}

func frontendTouchKeywords() []string {
	return []string{
		"frontend",
		"front end",
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
		fmt.Sprintf("Frontend implementation phase: %s", frontendImplementationPhaseFirstFrontend),
		fmt.Sprintf("Asset mode: %s", frontendAssetModeGeneratedImages),
		fmt.Sprintf("Imagegen requirement: %s", frontendImagegenRequirementRequired),
		"Asset decision proof: n/a; imagegen is required by default for first frontend and first major screen implementation.",
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
		"- Generated image assets required: Pending.",
		"- Asset manifest path: Pending.",
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
		"## Do-Not Design Contract",
		"",
		fmt.Sprintf("Contract Status: %s", frontendNegativeContractStatusMissing),
		"",
		"### Default Anti-Pattern Library",
		"",
		"- Generic card grids as the primary page grammar when the content is homogeneous or workflow-driven.",
		"- Bento grids used as an identity substitute rather than true information architecture.",
		"- Card-inside-card nesting, double backgrounds, and stacked border/shadow/tint/radius treatments without interaction semantics.",
		"- Decorative glassmorphism, blurred blobs, gradient-orb backgrounds, or unearned gradients.",
		"- Stock SaaS hero formulas: nav, oversized headline, vague CTA, fake metrics/logos, and abstract dashboard preview.",
		"- Interchangeable feature-card walls.",
		"- Generic testimonial, FAQ, pricing, or stats sections without product-specific trust evidence.",
		"- Dashboard KPI-card rows when the primary task is comparison, triage, workflow progress, or dense operational scanning.",
		"- Low-information empty states without next-step guidance.",
		"- Weak typography defaults such as tiny body text, too many text styles, or light weights on small text.",
		"- One-note palettes and decorative color that does not map to state, priority, or brand.",
		"- Dark, blurred, cropped, stock-like, or atmospheric imagery where users need to inspect the real product, place, object, gameplay, or state.",
		"",
		"### Context-Specific Banned Patterns",
		"",
		"- Banned pattern: Pending.",
		"  Why banned here: Pending.",
		"  Detection hints: Pending.",
		"  Allowed replacement: Pending.",
		"  Exception path: Pending.",
		"",
		"### Allowed Replacement Patterns",
		"",
		"- Pending.",
		"",
		"### Brand, Category, And Trust Reasoning",
		"",
		"- Brand: Pending.",
		"- Category: Pending.",
		"- Trust: Pending.",
		"- Audience and workflow: Pending.",
		"",
		"### Visual Grammar Contract",
		"",
		"- Layout primitives: Pending.",
		"- Hierarchy: Pending.",
		"- Typography: Pending.",
		"- Color and palette roles: Pending.",
		"- Density and spacing: Pending.",
		"- Depth and containers: Pending.",
		"- Imagery and icons: Pending.",
		"- Motion: Pending.",
		"",
		"### Reference-Driven Asset Manifest",
		"",
		fmt.Sprintf("- Frontend implementation phase: %s", frontendImplementationPhaseFirstFrontend),
		fmt.Sprintf("- Asset mode: %s", frontendAssetModeGeneratedImages),
		fmt.Sprintf("- Imagegen requirement: %s", frontendImagegenRequirementRequired),
		"- Asset decision proof: n/a; imagegen is required unless existing assets fully cover all concrete visual roles or imagery would harm the UX.",
		"- Asset ID: Pending.",
		"  Asset type: Pending.",
		"  Asset source: generated",
		"  Required status: required",
		"  Role in UI: Pending.",
		"  Reference signal: Pending.",
		"  Generation prompt/spec: Pending.",
		"  Source input path: n/a",
		"  Saved asset path: Pending.",
		"  Intended UI usage: Pending.",
		"  Rendered usage evidence: Pending.",
		"- Not-applicable proof: Pending.",
		"",
		"### Generated Image Execution Plan",
		"",
		fmt.Sprintf("- Generation plan status: %s", frontendGenerationPlanStatusReady),
		"- Tool path: Pending.",
		"- Execution order: Pending.",
		"- Prompt coverage: Pending.",
		"- Saved asset evidence plan: Pending.",
		"- Rendered usage evidence plan: Pending.",
		"- Not-applicable proof: Pending.",
		"",
		"### Most-Generic-Section Redesign Proof",
		"",
		"- Section: Pending.",
		"- Obvious generic fallback: Pending.",
		"- Why weak: Pending.",
		"- Replacement structure: Pending.",
		"- Evidence source: Pending.",
		"- Implementation implication: Pending.",
		"",
		"### Frontend Architecture Handoff",
		"",
		"- Allowed planning: Pending.",
		"- Out of scope: Pending.",
		"- Encouraged primitives/components: Pending.",
		"- Banned primitives/components: Pending.",
		"- Required evidence: Pending.",
		"- Responsive and accessibility constraints: Pending.",
		"- File/module planning notes: Pending.",
		"",
		"### Post-Implementation Violation Checks",
		"",
		"- Check: Pending.",
		"- Evidence to cite: Pending.",
		"- Blocking result: Pending.",
		"- Exception evidence: Pending.",
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
	report.FrontendImplementationPhase = normalizeFrontendBriefEnum(firstNonBlank(headerValues["Frontend implementation phase"], parseLooseLabel(text, "Frontend implementation phase")))
	report.AssetMode = normalizeFrontendBriefEnum(firstNonBlank(headerValues["Asset mode"], parseLooseLabel(text, "Asset mode")))
	report.ImagegenRequirement = normalizeFrontendBriefEnum(firstNonBlank(headerValues["Imagegen requirement"], parseLooseLabel(text, "Imagegen requirement")))
	report.AssetDecisionProof = strings.TrimSpace(firstNonBlank(headerValues["Asset decision proof"], parseLooseLabel(text, "Asset decision proof")))

	if report.Header.ClassificationRationale == "" {
		report.ContractIssues = append(report.ContractIssues, "Classification Rationale must not be blank.")
	}

	validateFrontendBriefEnums(&report)
	validateFrontendBriefConsistency(&report)
	validateFrontendNegativeContract(&report, text)
	report.ContractIssues = uniqueStrings(report.ContractIssues)
	report.NegativeContractIssues = uniqueStrings(report.NegativeContractIssues)
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

func validateFrontendNegativeContract(report *frontendBriefReport, text string) {
	if report.Header.TaskClassification != frontendTaskClassificationMajor {
		report.NegativeContractStatus = frontendNegativeContractStatusNotApplicable
		return
	}

	section, ok := markdownSection(text, "Do-Not Design Contract", 2)
	if !ok {
		report.NegativeContractStatus = frontendNegativeContractStatusMissing
		report.NegativeContractIssues = append(report.NegativeContractIssues, "Do-Not Design Contract section is missing.")
		return
	}

	status := normalizeFrontendBriefEnum(parseLooseLabel(section, "Contract Status"))
	switch status {
	case frontendNegativeContractStatusComplete, frontendNegativeContractStatusMissing, frontendNegativeContractStatusInsufficient:
		report.NegativeContractStatus = status
	case "":
		report.NegativeContractStatus = frontendNegativeContractStatusInsufficient
		report.NegativeContractIssues = append(report.NegativeContractIssues, "Do-Not Design Contract requires Contract Status.")
	default:
		report.NegativeContractStatus = frontendNegativeContractStatusInsufficient
		report.NegativeContractIssues = append(report.NegativeContractIssues, fmt.Sprintf("Do-Not Design Contract has unsupported Contract Status %q.", status))
	}

	requiredSections := []string{
		"Default Anti-Pattern Library",
		"Context-Specific Banned Patterns",
		"Allowed Replacement Patterns",
		"Brand, Category, And Trust Reasoning",
		"Visual Grammar Contract",
		"Reference-Driven Asset Manifest",
		"Generated Image Execution Plan",
		"Most-Generic-Section Redesign Proof",
		"Frontend Architecture Handoff",
		"Post-Implementation Violation Checks",
	}
	for _, heading := range requiredSections {
		content, exists := markdownSection(section, heading, 3)
		if !exists {
			report.NegativeContractIssues = append(report.NegativeContractIssues, fmt.Sprintf("Do-Not Design Contract missing `%s` section.", heading))
			continue
		}
		if !hasActionableMarkdownContent(content) {
			report.NegativeContractIssues = append(report.NegativeContractIssues, fmt.Sprintf("Do-Not Design Contract `%s` section is pending or empty.", heading))
		}
	}

	defaultLibrary, _ := markdownSection(section, "Default Anti-Pattern Library", 3)
	for _, item := range frontendDefaultAntiPatternNeedles() {
		if !strings.Contains(strings.ToLower(defaultLibrary), item) {
			report.NegativeContractIssues = append(report.NegativeContractIssues, fmt.Sprintf("Default anti-pattern library missing %q coverage.", item))
		}
	}

	contextBans, _ := markdownSection(section, "Context-Specific Banned Patterns", 3)
	for _, label := range []string{"Banned pattern", "Why banned here", "Detection hints", "Allowed replacement", "Exception path"} {
		value := parseLooseLabel(contextBans, label)
		if isPendingMarkdownValue(value) {
			report.NegativeContractIssues = append(report.NegativeContractIssues, fmt.Sprintf("Context-specific banned patterns require non-pending %s.", strings.ToLower(label)))
		}
	}

	validateFrontendReferenceDrivenAssets(report, section)

	if report.NegativeContractStatus == frontendNegativeContractStatusComplete && len(report.NegativeContractIssues) > 0 {
		report.NegativeContractStatus = frontendNegativeContractStatusInsufficient
	}
	if report.NegativeContractStatus == "" {
		report.NegativeContractStatus = frontendNegativeContractStatusInsufficient
	}
}

func validateFrontendReferenceDrivenAssets(report *frontendBriefReport, contractSection string) {
	manifestSection, manifestExists := markdownSection(contractSection, "Reference-Driven Asset Manifest", 3)
	if !manifestExists {
		report.NegativeContractIssues = append(report.NegativeContractIssues, "Reference-driven asset manifest is missing.")
		return
	}

	if value := parseLooseLabel(manifestSection, "Frontend implementation phase"); value != "" {
		report.FrontendImplementationPhase = normalizeFrontendBriefEnum(value)
	}
	if value := parseLooseLabel(manifestSection, "Asset mode"); value != "" {
		report.AssetMode = normalizeFrontendBriefEnum(value)
	}
	if value := parseLooseLabel(manifestSection, "Imagegen requirement"); value != "" {
		report.ImagegenRequirement = normalizeFrontendBriefEnum(value)
	}
	if value := parseLooseLabel(manifestSection, "Asset decision proof"); value != "" {
		report.AssetDecisionProof = strings.TrimSpace(value)
	}

	validateFrontendAssetDecisionFields(report, manifestSection)
	assetBlocks := parseFrontendAssetBlocks(manifestSection)
	switch report.AssetMode {
	case frontendAssetModeGeneratedImages:
		validateFrontendAssetBlocks(report, assetBlocks, true)
	case frontendAssetModeExistingAssets:
		validateFrontendAssetBlocks(report, assetBlocks, false)
	case frontendAssetModeNotApplicable:
		if isPendingMarkdownValue(parseLooseLabel(manifestSection, "Not-applicable proof")) {
			report.NegativeContractIssues = append(report.NegativeContractIssues, "Reference-driven asset manifest requires a non-pending not-applicable proof.")
		}
	case "":
		report.NegativeContractIssues = append(report.NegativeContractIssues, "Reference-driven asset manifest requires Asset mode.")
	default:
		report.NegativeContractIssues = append(report.NegativeContractIssues, fmt.Sprintf("Reference-driven asset manifest has unsupported Asset mode %q.", report.AssetMode))
	}

	planSection, planExists := markdownSection(contractSection, "Generated Image Execution Plan", 3)
	if !planExists {
		report.NegativeContractIssues = append(report.NegativeContractIssues, "Generated image execution plan is missing.")
		return
	}
	report.GeneratedImagePlan = normalizeFrontendBriefEnum(parseLooseLabel(planSection, "Generation plan status"))
	switch report.AssetMode {
	case frontendAssetModeGeneratedImages:
		if report.GeneratedImagePlan != frontendGenerationPlanStatusReady {
			report.NegativeContractIssues = append(report.NegativeContractIssues, "Generated image execution plan requires Generation plan status: ready when Asset mode is generated-images.")
		}
		for _, label := range []string{"Tool path", "Execution order", "Prompt coverage", "Saved asset evidence plan", "Rendered usage evidence plan"} {
			if isPendingMarkdownValue(parseLooseLabel(planSection, label)) {
				report.NegativeContractIssues = append(report.NegativeContractIssues, fmt.Sprintf("Generated image execution plan requires non-pending %s when Asset mode is generated-images.", strings.ToLower(label)))
			}
		}
	case frontendAssetModeExistingAssets, frontendAssetModeNotApplicable:
		if report.GeneratedImagePlan == "" {
			report.NegativeContractIssues = append(report.NegativeContractIssues, "Generated image execution plan requires Generation plan status.")
		}
		if report.GeneratedImagePlan != frontendNegativeContractStatusNotApplicable && report.GeneratedImagePlan != frontendGenerationPlanStatusReady {
			report.NegativeContractIssues = append(report.NegativeContractIssues, fmt.Sprintf("Generated image execution plan has unsupported Generation plan status %q for Asset mode %s.", report.GeneratedImagePlan, report.AssetMode))
		}
		if report.GeneratedImagePlan == frontendNegativeContractStatusNotApplicable && isPendingMarkdownValue(parseLooseLabel(planSection, "Not-applicable proof")) {
			report.NegativeContractIssues = append(report.NegativeContractIssues, "Generated image execution plan requires a non-pending not-applicable proof when Generation plan status is not-applicable.")
		}
	default:
		if report.GeneratedImagePlan == "" {
			report.NegativeContractIssues = append(report.NegativeContractIssues, "Generated image execution plan requires Generation plan status.")
		}
	}
}

type frontendAssetBlock struct {
	ID     string
	Fields map[string]string
}

func validateFrontendAssetDecisionFields(report *frontendBriefReport, manifestSection string) {
	if report.FrontendImplementationPhase == "" {
		report.NegativeContractIssues = append(report.NegativeContractIssues, "Reference-driven asset manifest requires Frontend implementation phase.")
	} else if !isAllowedFrontendImplementationPhase(report.FrontendImplementationPhase) {
		report.NegativeContractIssues = append(report.NegativeContractIssues, fmt.Sprintf("Reference-driven asset manifest has unsupported Frontend implementation phase %q.", report.FrontendImplementationPhase))
	}
	if report.ImagegenRequirement == "" {
		report.NegativeContractIssues = append(report.NegativeContractIssues, "Reference-driven asset manifest requires Imagegen requirement.")
	} else if !isAllowedFrontendImagegenRequirement(report.ImagegenRequirement) {
		report.NegativeContractIssues = append(report.NegativeContractIssues, fmt.Sprintf("Reference-driven asset manifest has unsupported Imagegen requirement %q.", report.ImagegenRequirement))
	}

	if report.ImagegenRequirement == frontendImagegenRequirementRequired && report.AssetMode != "" && report.AssetMode != frontendAssetModeGeneratedImages {
		report.NegativeContractIssues = append(report.NegativeContractIssues, "Reference-driven asset manifest with Imagegen requirement: required must use Asset mode: generated-images.")
	}
	if report.ImagegenRequirement == frontendImagegenRequirementCoveredByExistingAssets && report.AssetMode != "" && report.AssetMode != frontendAssetModeExistingAssets {
		report.NegativeContractIssues = append(report.NegativeContractIssues, "Reference-driven asset manifest with Imagegen requirement: covered-by-existing-assets must use Asset mode: existing-assets.")
	}
	if report.ImagegenRequirement == frontendImagegenRequirementNotApplicable && report.AssetMode != "" && report.AssetMode != frontendAssetModeNotApplicable {
		report.NegativeContractIssues = append(report.NegativeContractIssues, "Reference-driven asset manifest with Imagegen requirement: not-applicable must use Asset mode: not-applicable.")
	}

	if report.ImagegenRequirement == frontendImagegenRequirementCoveredByExistingAssets || report.ImagegenRequirement == frontendImagegenRequirementNotApplicable {
		if isPendingMarkdownValue(report.AssetDecisionProof) {
			report.NegativeContractIssues = append(report.NegativeContractIssues, "Reference-driven asset manifest requires non-pending Asset decision proof when imagegen is not required.")
		}
		if rejected := firstRejectedAssetSubstitute(report.AssetDecisionProof); rejected != "" {
			report.NegativeContractIssues = append(report.NegativeContractIssues, fmt.Sprintf("Reference-driven asset manifest rejects %s as generated image substitutes.", rejected))
		}
	}
	if report.ImagegenRequirement == frontendImagegenRequirementNotApplicable {
		proof := firstNonBlank(parseLooseLabel(manifestSection, "Not-applicable proof"), report.AssetDecisionProof)
		if isPendingMarkdownValue(proof) {
			report.NegativeContractIssues = append(report.NegativeContractIssues, "Reference-driven asset manifest requires explicit Not-applicable proof when Imagegen requirement is not-applicable.")
		}
	}
}

func isAllowedFrontendImplementationPhase(value string) bool {
	switch value {
	case frontendImplementationPhaseFirstFrontend, frontendImplementationPhaseFirstMajorScreen, frontendImplementationPhaseIncremental, frontendImplementationPhaseNotApplicable:
		return true
	default:
		return false
	}
}

func isAllowedFrontendImagegenRequirement(value string) bool {
	switch value {
	case frontendImagegenRequirementRequired, frontendImagegenRequirementCoveredByExistingAssets, frontendImagegenRequirementNotApplicable:
		return true
	default:
		return false
	}
}

func validateFrontendAssetBlocks(report *frontendBriefReport, blocks []frontendAssetBlock, generatedMode bool) {
	if len(blocks) == 0 {
		report.NegativeContractIssues = append(report.NegativeContractIssues, "Reference-driven asset manifest requires at least one repeatable Asset ID block.")
		return
	}
	hasGenerated := false
	for _, block := range blocks {
		source := normalizeFrontendBriefEnum(block.Fields["Asset source"])
		switch source {
		case "generated":
			hasGenerated = true
			validateFrontendAssetBlockFields(report, block, []string{
				"Asset ID",
				"Asset type",
				"Asset source",
				"Required status",
				"Role in UI",
				"Reference signal",
				"Generation prompt/spec",
				"Saved asset path",
				"Intended UI usage",
				"Rendered usage evidence",
			})
		case "existing-input", "existing-covered":
			if generatedMode && source == "existing-covered" {
				validateFrontendAssetBlockFields(report, block, []string{
					"Asset ID",
					"Asset type",
					"Asset source",
					"Required status",
					"Role in UI",
					"Reference signal",
					"Source input path",
					"Intended UI usage",
				})
			} else {
				validateFrontendAssetBlockFields(report, block, []string{
					"Asset ID",
					"Asset type",
					"Asset source",
					"Required status",
					"Role in UI",
					"Reference signal",
					"Source input path",
					"Intended UI usage",
					"Rendered usage evidence",
				})
			}
		case "":
			report.NegativeContractIssues = append(report.NegativeContractIssues, fmt.Sprintf("Reference-driven asset manifest Asset ID %q requires Asset source.", block.ID))
		default:
			report.NegativeContractIssues = append(report.NegativeContractIssues, fmt.Sprintf("Reference-driven asset manifest Asset ID %q has unsupported Asset source %q.", block.ID, source))
		}

		status := normalizeFrontendBriefEnum(block.Fields["Required status"])
		if status != "" && status != "required" && status != "supporting" && status != "covered" {
			report.NegativeContractIssues = append(report.NegativeContractIssues, fmt.Sprintf("Reference-driven asset manifest Asset ID %q has unsupported Required status %q.", block.ID, status))
		}
	}
	if generatedMode && !hasGenerated {
		report.NegativeContractIssues = append(report.NegativeContractIssues, "Reference-driven asset manifest with Asset mode: generated-images requires at least one generated Asset ID block.")
	}
	if !generatedMode && hasGenerated {
		report.NegativeContractIssues = append(report.NegativeContractIssues, "Reference-driven asset manifest with Asset mode: existing-assets must not contain generated Asset ID blocks.")
	}
}

func validateFrontendAssetBlockFields(report *frontendBriefReport, block frontendAssetBlock, labels []string) {
	for _, label := range labels {
		value := block.Fields[label]
		if isPendingMarkdownValue(value) {
			report.NegativeContractIssues = append(report.NegativeContractIssues, fmt.Sprintf("Reference-driven asset manifest Asset ID %q requires non-pending %s.", block.ID, strings.ToLower(label)))
		}
	}
}

func parseFrontendAssetBlocks(section string) []frontendAssetBlock {
	var blocks []frontendAssetBlock
	var current *frontendAssetBlock
	for _, rawLine := range strings.Split(section, "\n") {
		line := strings.TrimSpace(trimMarkdownListMarker(rawLine))
		if line == "" {
			continue
		}
		label, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		label = strings.TrimSpace(label)
		value = strings.TrimSpace(value)
		if label == "Asset ID" {
			blocks = append(blocks, frontendAssetBlock{
				ID:     value,
				Fields: map[string]string{"Asset ID": value},
			})
			current = &blocks[len(blocks)-1]
			continue
		}
		if current == nil {
			continue
		}
		switch label {
		case "Asset type", "Asset source", "Required status", "Role in UI", "Reference signal", "Generation prompt/spec", "Source input path", "Saved asset path", "Intended UI usage", "Rendered usage evidence":
			current.Fields[label] = value
		}
	}
	return blocks
}

func firstRejectedAssetSubstitute(value string) string {
	normalized := strings.ToLower(value)
	for _, needle := range frontendRejectedAssetSubstituteNeedles() {
		if strings.Contains(normalized, needle) {
			return needle
		}
	}
	return ""
}

func frontendRejectedAssetSubstituteNeedles() []string {
	return []string{
		"css gradient",
		"gradients",
		"abstract shape",
		"abstract decoration",
		"empty placeholder",
		"placeholder",
		"generic saas card",
		"card wall",
		"manually drawn decorative",
		"token-only",
		"color tokens",
	}
}

func frontendDefaultAntiPatternNeedles() []string {
	return []string{
		"generic card grids",
		"bento",
		"card-inside-card",
		"gradient",
		"stock saas hero",
		"feature-card",
		"testimonial",
		"kpi-card",
		"empty states",
		"typography",
		"one-note palettes",
		"stock-like imagery",
	}
}

func markdownSection(text, heading string, level int) (string, bool) {
	prefix := strings.Repeat("#", level) + " "
	stopPrefix := strings.Repeat("#", level) + " "
	var lines []string
	inSection := false
	for _, rawLine := range strings.Split(text, "\n") {
		line := strings.TrimSpace(rawLine)
		if strings.HasPrefix(line, prefix) {
			current := strings.TrimSpace(strings.TrimPrefix(line, prefix))
			if strings.EqualFold(current, heading) {
				inSection = true
				lines = nil
				continue
			}
			if inSection {
				break
			}
		}
		if inSection && level > 1 && strings.HasPrefix(line, strings.Repeat("#", level-1)+" ") {
			break
		}
		if inSection && strings.HasPrefix(line, stopPrefix) {
			break
		}
		if inSection {
			lines = append(lines, rawLine)
		}
	}
	if !inSection {
		return "", false
	}
	return strings.TrimSpace(strings.Join(lines, "\n")), true
}

func parseLooseLabel(text, label string) string {
	needle := strings.ToLower(strings.TrimSpace(label)) + ":"
	lines := strings.Split(text, "\n")
	for i, rawLine := range lines {
		line := strings.TrimSpace(trimMarkdownListMarker(rawLine))
		if !strings.HasPrefix(strings.ToLower(line), needle) {
			continue
		}
		value := strings.TrimSpace(line[len(needle):])
		if value != "" {
			return value
		}
		var continuation []string
		for _, nextRaw := range lines[i+1:] {
			next := strings.TrimSpace(nextRaw)
			if next == "" {
				break
			}
			if strings.Contains(strings.TrimSpace(trimMarkdownListMarker(next)), ":") && !strings.HasPrefix(nextRaw, " ") && !strings.HasPrefix(nextRaw, "\t") {
				break
			}
			continuation = append(continuation, next)
		}
		return strings.TrimSpace(strings.Join(continuation, "\n"))
	}
	return ""
}

func hasActionableMarkdownContent(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		value := strings.TrimSpace(trimMarkdownListMarker(line))
		if value == "" {
			continue
		}
		if strings.HasPrefix(value, "#") {
			continue
		}
		if !isPendingMarkdownValue(value) {
			return true
		}
	}
	return false
}

func isPendingMarkdownValue(value string) bool {
	normalized := normalizeDesignReviewPendingLine(value)
	switch normalized {
	case "", "pending", "n/a", "na", "tbd", "todo":
		return true
	default:
		return false
	}
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
			if !frontendBriefMissingLabel(report, label) {
				report.ContractIssues = append(report.ContractIssues, fmt.Sprintf("%s must not be blank.", label))
			}
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

func frontendBriefMissingLabel(report *frontendBriefReport, label string) bool {
	for _, missing := range report.MissingLabels {
		if missing == label {
			return true
		}
	}
	return false
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
	normalized := normalizeDesignReviewPendingField(value)
	return normalized == "" || normalized == "pending"
}

func normalizeDesignReviewPendingField(value string) string {
	sawValue := false
	for _, line := range strings.Split(value, "\n") {
		normalized := normalizeDesignReviewPendingLine(line)
		if normalized == "" {
			continue
		}
		sawValue = true
		if normalized != "pending" {
			return normalized
		}
	}
	if !sawValue {
		return ""
	}
	return "pending"
}

func normalizeDesignReviewPendingLine(line string) string {
	value := normalizeFrontendBriefEnum(line)
	value = trimMarkdownListMarker(value)
	value = strings.Trim(value, " \t\r\n.:;!?`\"'()[]")
	return strings.TrimSpace(value)
}

func trimMarkdownListMarker(value string) string {
	value = strings.TrimSpace(value)
	for _, marker := range []string{"- ", "* ", "+ "} {
		if strings.HasPrefix(value, marker) {
			return strings.TrimSpace(strings.TrimPrefix(value, marker))
		}
	}
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			if i > 0 && i+1 < len(value) && (value[i] == '.' || value[i] == ')') && value[i+1] == ' ' {
				return strings.TrimSpace(value[i+2:])
			}
			break
		}
	}
	return value
}

func frontendGateReadinessLines(root, specID string) []string {
	report := loadFrontendBriefReport(root, specID)
	return frontendGateReadinessLinesFromReport(report, specID)
}

func frontendGateReadinessLinesFromReport(report frontendBriefReport, specID string) []string {
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
				fmt.Sprintf("- Negative-First Contract Status: `%s`", report.NegativeContractStatus),
				fmt.Sprintf("- Frontend Implementation Phase: `%s`", firstNonBlank(report.FrontendImplementationPhase, "unknown")),
				fmt.Sprintf("- Reference Asset Mode: `%s`", firstNonBlank(report.AssetMode, "unknown")),
				fmt.Sprintf("- Imagegen Requirement: `%s`", firstNonBlank(report.ImagegenRequirement, "unknown")),
				fmt.Sprintf("- Generated Image Plan: `%s`", firstNonBlank(report.GeneratedImagePlan, "unknown")),
			)
			if len(report.MissingGates) > 0 {
				lines = append(lines, fmt.Sprintf("- Missing gates: %s", quoteList(report.MissingGates)))
			}
			if len(report.InsufficientGates) > 0 {
				lines = append(lines, fmt.Sprintf("- Insufficient gates: %s", quoteList(report.InsufficientGates)))
			}
			if len(report.NegativeContractIssues) > 0 {
				lines = append(lines, fmt.Sprintf("- Negative-first issues: %s", strings.Join(report.NegativeContractIssues, "; ")))
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
	return frontendGateAdvisorySummaryFromReport(report)
}

func frontendGateAdvisorySummaryFromReport(report frontendBriefReport) string {
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
	if report.NegativeContractStatus != frontendNegativeContractStatusComplete {
		return "frontend=negative-first-" + firstNonBlank(report.NegativeContractStatus, frontendNegativeContractStatusInsufficient)
	}
	if report.Header.FrontendGateStatus == frontendGateStatusApproved && report.EvidenceStatus == frontendEvidenceStatusComplete {
		return "frontend=approved"
	}
	return "frontend=" + firstNonBlank(report.Header.FrontendGateStatus, report.EvidenceStatus)
}

func frontendInvalidContractBlocksExecution(report frontendBriefReport) bool {
	return report.Header.TaskClassification != frontendTaskClassificationMinor
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
	if report.Header.TaskClassification == frontendTaskClassificationMajor {
		lines = append(lines, fmt.Sprintf("Negative-First Contract Status: `%s`", report.NegativeContractStatus))
		lines = append(lines, fmt.Sprintf("Frontend Implementation Phase: `%s`", firstNonBlank(report.FrontendImplementationPhase, "unknown")))
		lines = append(lines, fmt.Sprintf("Reference Asset Mode: `%s`", firstNonBlank(report.AssetMode, "unknown")))
		lines = append(lines, fmt.Sprintf("Imagegen Requirement: `%s`", firstNonBlank(report.ImagegenRequirement, "unknown")))
		lines = append(lines, fmt.Sprintf("Generated Image Plan: `%s`", firstNonBlank(report.GeneratedImagePlan, "unknown")))
	}
	if len(report.MissingGates) > 0 {
		lines = append(lines, fmt.Sprintf("Missing gates: %s.", quoteList(report.MissingGates)))
	}
	if len(report.InsufficientGates) > 0 {
		lines = append(lines, fmt.Sprintf("Insufficient gates: %s.", quoteList(report.InsufficientGates)))
	}
	for _, issue := range report.NegativeContractIssues {
		lines = append(lines, "- "+issue)
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
	appendIf(report.NegativeContractStatus == frontendNegativeContractStatusMissing, "Add the Do-Not Design Contract to `frontend-brief.md`, including default anti-patterns, context-specific bans, replacements, visual grammar, reference-driven asset manifest, generated-image decision fields, generation plan, architecture handoff, and violation checks.")
	appendIf(report.NegativeContractStatus == frontendNegativeContractStatusInsufficient, "Complete the Do-Not Design Contract with non-pending banned patterns, allowed replacements, brand/category/trust reasoning, visual grammar, repeatable asset manifest blocks, generated-image decision proof, generation plan, architecture handoff, and post-implementation checks.")
	switch report.Header.FrontendGateStatus {
	case frontendGateStatusBlocked:
		steps = append(steps, "Resolve the blocked frontend decision in `reviews/design.md`, update `frontend-brief.md` with the accepted direction and banned patterns, then mark the frontend gate approved only after design review clears.")
	case frontendGateStatusNeedsResearch:
		steps = append(steps, "Complete the requested frontend research in `frontend-brief.md`, refresh reference, critique, and decision evidence, then rerun design review before implementation.")
	}
	if len(report.Mismatches) > 0 {
		steps = append(steps, "Reconcile `frontend-brief.md` with `reviews/design.md` so the canonical gate state and summaries agree.")
	}
	return uniqueStrings(steps)
}
