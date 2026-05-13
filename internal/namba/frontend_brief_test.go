package namba

import (
	"strings"
	"testing"
)

func TestParseFrontendBriefRejectsApprovedMajorWithMissingGate(t *testing.T) {
	t.Parallel()

	report := parseFrontendBrief(strings.Join([]string{
		"# Frontend Brief",
		"",
		"Task Classification: frontend-major",
		"Classification Rationale: Major dashboard restructure.",
		"Frontend Gate Status: approved",
		"Problem Gate: complete",
		"Reference Gate: missing",
		"Critique Gate: complete",
		"Decision Gate: complete",
		"Prototype Gate: complete",
		"Prototype Evidence: wireframe",
		"",
		"## Problem Frame",
		"",
		"- Pending.",
	}, "\n"))

	if report.Valid {
		t.Fatalf("expected invalid report, got %+v", report)
	}
	if report.ContractStatus != "invalid-contract" {
		t.Fatalf("expected invalid-contract status, got %+v", report)
	}
	if !strings.Contains(strings.Join(report.ContractIssues, "\n"), "approved") {
		t.Fatalf("expected approved/missing contradiction to be reported, got %+v", report)
	}
}

func TestParseFrontendBriefAcceptsFrontendMinorNotApplicableHeader(t *testing.T) {
	t.Parallel()

	report := parseFrontendBrief(strings.Join([]string{
		"# Frontend Brief",
		"",
		"Task Classification: frontend-minor",
		"Classification Rationale: Existing settings screen spacing fix.",
		"Frontend Gate Status: not-applicable",
		"Problem Gate: not-applicable",
		"Reference Gate: not-applicable",
		"Critique Gate: not-applicable",
		"Decision Gate: not-applicable",
		"Prototype Gate: not-applicable",
		"Prototype Evidence: n/a",
		"",
		"## Current Pattern",
		"",
		"- Existing inline settings rows.",
	}, "\n"))

	if !report.Valid {
		t.Fatalf("expected valid minor report, got %+v", report)
	}
	if report.Header.TaskClassification != "frontend-minor" {
		t.Fatalf("expected frontend-minor classification, got %+v", report)
	}
	if report.EvidenceStatus != "not-applicable" {
		t.Fatalf("expected not-applicable evidence status, got %+v", report)
	}
	if len(report.ContractIssues) != 0 {
		t.Fatalf("expected no contract issues, got %+v", report)
	}
}

func TestParseFrontendBriefRejectsBlankFixedLabelValues(t *testing.T) {
	t.Parallel()

	report := parseFrontendBrief(strings.Join([]string{
		"# Frontend Brief",
		"",
		"Task Classification:",
		"Classification Rationale: Blank enum labels should fail.",
		"Frontend Gate Status: approved",
		"Problem Gate: complete",
		"Reference Gate: complete",
		"Critique Gate: complete",
		"Decision Gate: complete",
		"Prototype Gate: complete",
		"Prototype Evidence: wireframe",
	}, "\n"))

	if report.Valid {
		t.Fatalf("expected invalid report, got %+v", report)
	}
	if !strings.Contains(strings.Join(report.ContractIssues, "\n"), "Task Classification must not be blank.") {
		t.Fatalf("expected blank classification to be reported, got %+v", report.ContractIssues)
	}
}

func TestCompareFrontendBriefAndDesignReviewAcceptsMultilineDecisionFields(t *testing.T) {
	t.Parallel()

	report := parseFrontendBrief(strings.Join([]string{
		"# Frontend Brief",
		"",
		"Task Classification: frontend-major",
		"Classification Rationale: Major dashboard restructure.",
		"Frontend Gate Status: approved",
		"Problem Gate: complete",
		"Reference Gate: complete",
		"Critique Gate: complete",
		"Decision Gate: complete",
		"Prototype Gate: complete",
		"Prototype Evidence: wireframe",
	}, "\n"))
	if !report.Valid {
		t.Fatalf("expected valid report before design review comparison, got %+v", report)
	}

	compareFrontendBriefAndDesignReview(&report, strings.Join([]string{
		"# Design Review",
		"",
		"- Evidence Status: complete",
		"- Gate Decision: approved",
		"- Approved Direction:",
		"  - Use the focused dashboard table direction.",
		"- Banned Patterns:",
		"  - Avoid generic card grids.",
		"- Open Questions:",
		"  - none",
		"- Unresolved Questions:",
		"  - none",
	}, "\n"))

	if len(report.Mismatches) > 0 {
		t.Fatalf("expected multiline decision fields to satisfy design review contract, got %+v", report.Mismatches)
	}
}

func TestParseFrontendBriefBlocksMajorWhenDoNotDesignContractMissing(t *testing.T) {
	t.Parallel()

	report := parseFrontendBrief(strings.Join([]string{
		"# Frontend Brief",
		"",
		"Task Classification: frontend-major",
		"Classification Rationale: Major dashboard restructure.",
		"Frontend Gate Status: approved",
		"Problem Gate: complete",
		"Reference Gate: complete",
		"Critique Gate: complete",
		"Decision Gate: complete",
		"Prototype Gate: complete",
		"Prototype Evidence: wireframe",
	}, "\n"))

	if !report.Valid {
		t.Fatalf("expected original fixed-label contract to remain valid, got %+v", report)
	}
	if report.NegativeContractStatus != frontendNegativeContractStatusMissing {
		t.Fatalf("expected missing negative-first contract, got %+v", report)
	}
	if !strings.Contains(strings.Join(report.NegativeContractIssues, "\n"), "Do-Not Design Contract section is missing") {
		t.Fatalf("expected missing contract issue, got %+v", report.NegativeContractIssues)
	}
}

func TestParseFrontendBriefAcceptsCompleteDoNotDesignContract(t *testing.T) {
	t.Parallel()

	report := parseFrontendBrief(validFrontendMajorBriefWithDoNotDesignContract())

	if !report.Valid {
		t.Fatalf("expected fixed-label contract to be valid, got %+v", report)
	}
	if report.NegativeContractStatus != frontendNegativeContractStatusComplete {
		t.Fatalf("expected complete negative-first contract, got %+v", report)
	}
	if report.EvidenceStatus != frontendEvidenceStatusComplete {
		t.Fatalf("expected original evidence status to stay complete, got %+v", report)
	}
	if report.AssetMode != frontendAssetModeGeneratedImages {
		t.Fatalf("expected generated image asset mode, got %+v", report)
	}
	if report.GeneratedImagePlan != frontendNegativeContractStatusComplete {
		t.Fatalf("expected complete generated image plan, got %+v", report)
	}
	if len(report.NegativeContractIssues) != 0 {
		t.Fatalf("expected no negative-first issues, got %+v", report.NegativeContractIssues)
	}
}

func TestParseFrontendBriefRejectsContextBanWithoutReplacement(t *testing.T) {
	t.Parallel()

	body := strings.Replace(validFrontendMajorBriefWithDoNotDesignContract(), "Allowed replacement: Workflow lane with state chips and inline proof.", "Allowed replacement: Pending.", 1)
	report := parseFrontendBrief(body)

	if report.NegativeContractStatus != frontendNegativeContractStatusInsufficient {
		t.Fatalf("expected insufficient negative-first contract, got %+v", report)
	}
	if !strings.Contains(strings.Join(report.NegativeContractIssues, "\n"), "allowed replacement") {
		t.Fatalf("expected replacement issue, got %+v", report.NegativeContractIssues)
	}
}

func TestParseFrontendBriefRejectsGeneratedImageModeWithoutAssetEvidence(t *testing.T) {
	t.Parallel()

	body := strings.Replace(validFrontendMajorBriefWithDoNotDesignContract(), "Output path: frontend/assets/workflow-state-hero.png", "Output path: Pending.", 1)
	report := parseFrontendBrief(body)

	if report.NegativeContractStatus != frontendNegativeContractStatusInsufficient {
		t.Fatalf("expected insufficient generated-image contract, got %+v", report)
	}
	if !strings.Contains(strings.Join(report.NegativeContractIssues, "\n"), "output path") {
		t.Fatalf("expected generated-image output path issue, got %+v", report.NegativeContractIssues)
	}
}

func TestCompareFrontendBriefAndDesignReviewBlocksPendingMarkdownMarkers(t *testing.T) {
	t.Parallel()

	report := parseFrontendBrief(strings.Join([]string{
		"# Frontend Brief",
		"",
		"Task Classification: frontend-major",
		"Classification Rationale: Major dashboard restructure.",
		"Frontend Gate Status: approved",
		"Problem Gate: complete",
		"Reference Gate: complete",
		"Critique Gate: complete",
		"Decision Gate: complete",
		"Prototype Gate: complete",
		"Prototype Evidence: wireframe",
	}, "\n"))
	if !report.Valid {
		t.Fatalf("expected valid report before design review comparison, got %+v", report)
	}

	compareFrontendBriefAndDesignReview(&report, strings.Join([]string{
		"# Design Review",
		"",
		"- Evidence Status: complete",
		"- Gate Decision: approved",
		"- Approved Direction:",
		"  - pending",
		"- Banned Patterns: Pending.",
		"- Open Questions:",
		"  - Pending.",
		"- Unresolved Questions: none",
	}, "\n"))

	mismatches := strings.Join(report.Mismatches, "\n")
	for _, want := range []string{
		"Design review approved direction is pending",
		"Design review banned patterns are pending",
		"Design review open questions are pending",
	} {
		if !strings.Contains(mismatches, want) {
			t.Fatalf("expected %q mismatch, got %+v", want, report.Mismatches)
		}
	}
	if strings.Contains(mismatches, "Design review unresolved questions are pending") {
		t.Fatalf("expected non-pending unresolved questions field to pass, got %+v", report.Mismatches)
	}
}

func TestFrontendGateRemediationIncludesStatusOnlyMajorBlocks(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		status string
		want   string
	}{
		{
			status: frontendGateStatusBlocked,
			want:   "Resolve the blocked frontend decision",
		},
		{
			status: frontendGateStatusNeedsResearch,
			want:   "Complete the requested frontend research",
		},
	} {
		t.Run(tt.status, func(t *testing.T) {
			report := parseFrontendBrief(strings.Join([]string{
				"# Frontend Brief",
				"",
				"Task Classification: frontend-major",
				"Classification Rationale: Major dashboard restructure.",
				"Frontend Gate Status: " + tt.status,
				"Problem Gate: complete",
				"Reference Gate: complete",
				"Critique Gate: complete",
				"Decision Gate: complete",
				"Prototype Gate: complete",
				"Prototype Evidence: wireframe",
			}, "\n"))
			if !report.Valid {
				t.Fatalf("expected valid report, got %+v", report)
			}

			steps := strings.Join(frontendGateRemediation(report), "\n")
			if !strings.Contains(steps, tt.want) {
				t.Fatalf("expected remediation to contain %q, got %q", tt.want, steps)
			}
		})
	}
}

func TestInferFrontendTaskClassificationMatchesUIAtBoundaries(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name               string
		kind               string
		description        string
		wantClassification string
	}{
		{
			name:               "start boundary",
			kind:               "fix",
			description:        "UI polish on the existing settings screen",
			wantClassification: frontendTaskClassificationMinor,
		},
		{
			name:               "end boundary",
			kind:               "plan",
			description:        "improve UI",
			wantClassification: frontendTaskClassificationMajor,
		},
		{
			name:               "fix-only dashboard maintenance",
			kind:               "fix",
			description:        "fix button spacing on dashboard",
			wantClassification: frontendTaskClassificationMinor,
		},
		{
			name:               "dashboard-only bugfix",
			kind:               "fix",
			description:        "fix dashboard filters not loading",
			wantClassification: frontendTaskClassificationMinor,
		},
		{
			name:               "explicit frontend token",
			kind:               "plan",
			description:        "frontend component refactor",
			wantClassification: frontendTaskClassificationMajor,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			classification, _, ok := inferFrontendTaskClassification(tt.kind, tt.description)
			if !ok {
				t.Fatalf("expected %q to be classified as frontend-touching", tt.description)
			}
			if classification != tt.wantClassification {
				t.Fatalf("classification = %q, want %q", classification, tt.wantClassification)
			}
		})
	}
}

func TestInferFrontendTaskClassificationIgnoresBackendOnlyAmbiguousTouchKeywords(t *testing.T) {
	t.Parallel()

	for _, description := range []string{
		"add settings API endpoint",
		"add form submission API",
		"refactor auth component service",
		"add dashboard metrics API endpoint",
		"add dashboard text API endpoint",
		"refactor database hierarchy for settings API endpoint",
		"add page token to API endpoint",
	} {
		classification, rationale, ok := inferFrontendTaskClassification("plan", description)
		if ok {
			t.Fatalf("expected backend-only description %q to avoid frontend classification, got classification=%q rationale=%q", description, classification, rationale)
		}
	}
}

func TestInferFrontendTaskClassificationIgnoresDocumentationSectionOnly(t *testing.T) {
	t.Parallel()

	classification, rationale, ok := inferFrontendTaskClassification("plan", "add a section to README")
	if ok {
		t.Fatalf("expected documentation-only section request to avoid frontend classification, got classification=%q rationale=%q", classification, rationale)
	}
}

func validFrontendMajorBriefWithDoNotDesignContract() string {
	return strings.Join([]string{
		"# Frontend Brief",
		"",
		"Task Classification: frontend-major",
		"Classification Rationale: Major dashboard restructure.",
		"Frontend Gate Status: approved",
		"Problem Gate: complete",
		"Reference Gate: complete",
		"Critique Gate: complete",
		"Decision Gate: complete",
		"Prototype Gate: complete",
		"Prototype Evidence: wireframe",
		"",
		"## Do-Not Design Contract",
		"",
		"Contract Status: complete",
		"",
		"### Default Anti-Pattern Library",
		"",
		"- Generic card grids as the primary page grammar.",
		"- Bento grids used as identity substitute.",
		"- Card-inside-card nesting and double backgrounds.",
		"- Decorative glass, blurred blobs, gradient orbs, or unearned gradients.",
		"- Stock SaaS hero formulas.",
		"- Interchangeable feature-card walls.",
		"- Generic testimonial, FAQ, pricing, and stats sections.",
		"- Dashboard KPI-card rows used against task needs.",
		"- Low-information empty states.",
		"- Weak typography defaults.",
		"- One-note palettes.",
		"- Stock-like imagery where inspection is needed.",
		"",
		"### Context-Specific Banned Patterns",
		"",
		"- Banned pattern: Three-column feature-card wall for workflow explanation.",
		"  Why banned here: The user compares state transitions rather than interchangeable claims.",
		"  Detection hints: feature-card grid, repeated icon/title/body cards, grid-cols-3 marketing modules.",
		"  Allowed replacement: Workflow lane with state chips and inline proof.",
		"  Exception path: Allowed only for independent feature categories with distinct actions.",
		"",
		"### Allowed Replacement Patterns",
		"",
		"- Workflow lane with state chips, decision points, and inline proof.",
		"",
		"### Brand, Category, And Trust Reasoning",
		"",
		"- Brand: Existing product typography and repository tone favor direct engineering clarity.",
		"- Category: Developer workflow tool with repeated operational scanning.",
		"- Trust: Users must believe state transitions and gates are explicit before acting.",
		"- Audience and workflow: Dense scanning, fast decisions, and accessible repeated use.",
		"",
		"### Visual Grammar Contract",
		"",
		"- Layout primitives: workflow lanes, tables, command surfaces, and inspectors.",
		"- Hierarchy: Current state, blocker, next action, then supporting evidence.",
		"- Typography: Stable scale with readable body-size floor.",
		"- Color and palette roles: brand, surface, state, priority, danger, success, selected, disabled, focus.",
		"- Density and spacing: Compact rhythm with responsive wrapping.",
		"- Depth and containers: Borders only for semantic grouping.",
		"- Imagery and icons: Functional icons only unless domain assets are inspectable.",
		"- Motion: State-change attention only.",
		"",
		"### Reference-Driven Asset Manifest",
		"",
		"- Asset mode: generated-images",
		"- Asset ID: workflow-state-hero",
		"  Role in screen: first viewport workflow-state illustration that replaces generic KPI cards.",
		"  Reference signal: reference synthesis requires visible workflow state proof instead of abstract SaaS cards.",
		"  Generation prompt/spec: create an original high-fidelity workflow-state product visual with no logos, no watermark, and no decorative dashboard placeholder.",
		"  Source asset path: n/a",
		"  Output path: frontend/assets/workflow-state-hero.png",
		"  Usage in implementation: hero media beside the workflow lane and above detailed evidence rows.",
		"  Validation evidence: rendered screenshot confirms the generated asset appears in the first viewport.",
		"- Not-applicable proof: n/a",
		"",
		"### Generated Image Execution Plan",
		"",
		"- Generation status: complete",
		"- Tool path: Codex built-in image generation.",
		"- Execution order: define asset manifest, generate bitmap, copy to frontend/assets, wire into layout, capture rendered screen.",
		"- Prompt coverage: prompt names subject, reference signal, style, composition, output constraints, and banned logo/watermark/text drift.",
		"- Output evidence: cite asset manifest and generated PNG paths.",
		"- Rendered usage evidence: cite browser screenshot or local render showing the generated asset in context.",
		"- Not-applicable proof: n/a",
		"",
		"### Most-Generic-Section Redesign Proof",
		"",
		"- Section: Dashboard overview.",
		"- Obvious generic fallback: KPI card row above bento feature cards.",
		"- Why weak: It hides workflow state behind interchangeable metrics.",
		"- Replacement structure: Triage lane with state chips and inline validation evidence.",
		"- Evidence source: Reference synthesis and product state model.",
		"- Implementation implication: Build workflow primitives instead of metric cards.",
		"",
		"### Frontend Architecture Handoff",
		"",
		"- Allowed planning: Workflow lane components, state chips, validation evidence rows, responsive inspector.",
		"- Out of scope: Marketing feature-card walls and decorative dashboard metrics.",
		"- Encouraged primitives/components: lane, state chip, evidence row, command surface.",
		"- Banned primitives/components: KPI card row, generic bento grid, card-inside-card.",
		"- Required evidence: changed files and screenshot or DOM inspection when browser app is available.",
		"- Responsive and accessibility constraints: no tiny body text; controls remain reachable on mobile.",
		"- File/module planning notes: Keep state ownership near workflow data, not presentational cards.",
		"",
		"### Post-Implementation Violation Checks",
		"",
		"- Check: Inspect changed files for KPI card rows, generic bento grids, and feature-card walls.",
		"- Evidence to cite: changed files and screenshot or browser evidence when available.",
		"- Blocking result: Any banned pattern appears without exception evidence.",
		"- Exception evidence: cite the exception path above before claiming pass.",
	}, "\n")
}
