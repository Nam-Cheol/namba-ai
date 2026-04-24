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
