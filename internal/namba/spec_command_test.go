package namba

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFixCreatesBugfixSpec(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"fix", "startup", "panic"}); err != nil {
		t.Fatalf("fix failed: %v", err)
	}

	spec := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "spec.md"))
	if !strings.Contains(spec, "## Problem") || !strings.Contains(spec, "startup panic") || !strings.Contains(spec, "Work type: fix") {
		t.Fatalf("unexpected fix spec: %q", spec)
	}

	plan := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "plan.md"))
	if !strings.Contains(plan, "Implement the smallest safe fix") || !strings.Contains(plan, ".namba/specs/SPEC-001/reviews/") {
		t.Fatalf("unexpected fix plan: %q", plan)
	}

	acceptance := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "acceptance.md"))
	if !strings.Contains(acceptance, "reported issue") || !strings.Contains(acceptance, "regression test") {
		t.Fatalf("unexpected fix acceptance: %q", acceptance)
	}

	productReview := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "reviews", "product.md"))
	if !strings.Contains(productReview, "# Product Review") || !strings.Contains(productReview, "$namba-plan-pm-review") || !strings.Contains(productReview, "- Status: pending") {
		t.Fatalf("unexpected product review scaffold: %q", productReview)
	}

	engineeringReview := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "reviews", "engineering.md"))
	if !strings.Contains(engineeringReview, "# Engineering Review") || !strings.Contains(engineeringReview, "$namba-plan-eng-review") || !strings.Contains(engineeringReview, "`namba-planner`") {
		t.Fatalf("unexpected engineering review scaffold: %q", engineeringReview)
	}

	designReview := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "reviews", "design.md"))
	if !strings.Contains(designReview, "# Design Review") || !strings.Contains(designReview, "$namba-plan-design-review") || !strings.Contains(designReview, "`namba-designer`") {
		t.Fatalf("unexpected design review scaffold: %q", designReview)
	}

	readiness := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "reviews", "readiness.md"))
	if !strings.Contains(readiness, "# Review Readiness") || !strings.Contains(readiness, "Advisory only") || !strings.Contains(readiness, "Cleared reviews: 0/3") {
		t.Fatalf("unexpected review readiness scaffold: %q", readiness)
	}
}

func TestRunFixRequiresDescription(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	err := app.Run(context.Background(), []string{"fix"})
	if err == nil {
		t.Fatal("expected fix description error")
	}
	if !strings.Contains(err.Error(), "fix requires a description") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPlanCreatesReviewArtifacts(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"plan", "improve", "review", "workflow"}); err != nil {
		t.Fatalf("plan failed: %v", err)
	}

	plan := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "plan.md"))
	if !strings.Contains(plan, "Run the relevant review passes under `.namba/specs/SPEC-001/reviews/` and refresh the readiness summary") {
		t.Fatalf("unexpected feature plan: %q", plan)
	}

	readiness := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "reviews", "readiness.md"))
	for _, want := range []string{"# Review Readiness", "$namba-plan-pm-review", "$namba-plan-eng-review", "$namba-plan-design-review", "follow up on product=pending, engineering=pending, design=pending"} {
		if !strings.Contains(readiness, want) {
			t.Fatalf("expected readiness scaffold to contain %q, got %q", want, readiness)
		}
	}
}
