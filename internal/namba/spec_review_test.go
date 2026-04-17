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
