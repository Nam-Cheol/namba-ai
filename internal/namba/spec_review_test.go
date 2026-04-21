package namba

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefreshAllSpecReviewReadinessRecreatesMissingFiles(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"plan", "restore readiness"}); err != nil {
		t.Fatalf("plan failed: %v", err)
	}

	readinessPath := filepath.Join(tmp, ".namba", "specs", "SPEC-001", "reviews", "readiness.md")
	if err := os.Remove(readinessPath); err != nil {
		t.Fatalf("remove readiness: %v", err)
	}

	if err := app.refreshAllSpecReviewReadiness(tmp); err != nil {
		t.Fatalf("refreshAllSpecReviewReadiness failed: %v", err)
	}

	readiness := mustReadFile(t, readinessPath)
	for _, want := range []string{"# Review Readiness", "SPEC: SPEC-001", "Advisory only"} {
		if !strings.Contains(readiness, want) {
			t.Fatalf("expected recreated readiness to contain %q, got %q", want, readiness)
		}
	}
}

func TestDesignReviewTemplateIncludesExplicitChecklist(t *testing.T) {
	t.Parallel()

	outputs := specReviewOutputs("SPEC-999")
	designReview := outputs[specReviewPath("SPEC-999", "design")]
	for _, want := range []string{
		"## Review Checklist",
		"Art direction is clear and fits the task context.",
		"Palette temperature and undertone logic are coherent",
		"generic cards, border-heavy framing, or bento/grid fallback",
		"The most generic section is redesigned",
		"no novelty for novelty's sake",
	} {
		if !strings.Contains(designReview, want) {
			t.Fatalf("expected design review scaffold to contain %q, got %q", want, designReview)
		}
	}
}

func TestRefreshSpecReviewReadinessKeepsLegacyEvidenceAnchors(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	specDir := filepath.Join(tmp, ".namba", "specs", "SPEC-099")
	reviewsDir := filepath.Join(specDir, "reviews")
	if err := os.MkdirAll(reviewsDir, 0o755); err != nil {
		t.Fatalf("mkdir reviews dir: %v", err)
	}
	writeTestFile(t, filepath.Join(specDir, "contract.md"), "# Contract\n")
	writeTestFile(t, filepath.Join(specDir, "baseline.md"), "# Baseline\n")
	writeTestFile(t, filepath.Join(specDir, "extraction-map.md"), "# Extraction Map\n")
	for rel, body := range specReviewOutputs("SPEC-099") {
		if !strings.HasPrefix(rel, filepath.ToSlash(filepath.Join(specsDir, "SPEC-099", specReviewsDirName, ""))) || strings.HasSuffix(rel, specReviewReadinessFileName) {
			continue
		}
		writeTestFile(t, filepath.Join(tmp, filepath.FromSlash(rel)), body)
	}

	advisory, err := app.refreshSpecReviewReadiness(tmp, "SPEC-099")
	if err != nil {
		t.Fatalf("refreshSpecReviewReadiness failed: %v", err)
	}
	if advisory != "product=pending, engineering=pending, design=pending" {
		t.Fatalf("unexpected advisory summary: %q", advisory)
	}

	readiness := mustReadFile(t, filepath.Join(reviewsDir, "readiness.md"))
	for _, want := range []string{
		"## Phase-1 Evidence",
		"Runtime contract anchor: `.namba/specs/SPEC-099/contract.md`",
		"Baseline evidence: `.namba/specs/SPEC-099/baseline.md`",
		"Extraction map: `.namba/specs/SPEC-099/extraction-map.md`",
	} {
		if !strings.Contains(readiness, want) {
			t.Fatalf("expected legacy readiness to contain %q, got %q", want, readiness)
		}
	}
	if strings.Contains(readiness, "harness-request.json") {
		t.Fatalf("expected legacy readiness to stay off typed harness sidecar plumbing, got %q", readiness)
	}
}
