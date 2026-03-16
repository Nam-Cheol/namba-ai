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
	if !strings.Contains(plan, "Implement the smallest safe fix") {
		t.Fatalf("unexpected fix plan: %q", plan)
	}

	acceptance := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "acceptance.md"))
	if !strings.Contains(acceptance, "reported issue") || !strings.Contains(acceptance, "regression test") {
		t.Fatalf("unexpected fix acceptance: %q", acceptance)
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
