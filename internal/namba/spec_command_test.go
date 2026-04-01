package namba

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPlanHelpIsReadOnly(t *testing.T) {
	t.Parallel()

	for _, helpFlag := range []string{"--help", "-h"} {
		helpFlag := helpFlag
		t.Run(helpFlag, func(t *testing.T) {
			t.Parallel()

			tmp := t.TempDir()
			stdout := &bytes.Buffer{}
			app := NewApp(stdout, &bytes.Buffer{})
			if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
				t.Fatalf("init failed: %v", err)
			}

			restore := chdirExecution(t, tmp)
			defer restore()

			if err := app.Run(context.Background(), []string{"plan", helpFlag}); err != nil {
				t.Fatalf("plan help failed: %v", err)
			}
			if got := stdout.String(); !strings.Contains(got, "namba plan") || !strings.Contains(got, "Create the next feature SPEC package") {
				t.Fatalf("unexpected plan help output: %q", got)
			}

			entries, err := os.ReadDir(filepath.Join(tmp, ".namba", "specs"))
			if err != nil {
				t.Fatalf("read specs dir: %v", err)
			}
			if got, want := len(entries), 1; got != want || entries[0].Name() != ".gitkeep" {
				t.Fatalf("expected no SPEC write on help, got entries=%v", entries)
			}
		})
	}
}

func TestRunHarnessHelpIsReadOnly(t *testing.T) {
	t.Parallel()

	for _, helpFlag := range []string{"--help", "-h"} {
		helpFlag := helpFlag
		t.Run(helpFlag, func(t *testing.T) {
			t.Parallel()

			tmp := t.TempDir()
			stdout := &bytes.Buffer{}
			app := NewApp(stdout, &bytes.Buffer{})
			if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
				t.Fatalf("init failed: %v", err)
			}

			restore := chdirExecution(t, tmp)
			defer restore()

			if err := app.Run(context.Background(), []string{"harness", helpFlag}); err != nil {
				t.Fatalf("harness help failed: %v", err)
			}
			if got := stdout.String(); !strings.Contains(got, "namba harness") || !strings.Contains(got, "Create the next harness-oriented SPEC package") {
				t.Fatalf("unexpected harness help output: %q", got)
			}

			entries, err := os.ReadDir(filepath.Join(tmp, ".namba", "specs"))
			if err != nil {
				t.Fatalf("read specs dir: %v", err)
			}
			if got, want := len(entries), 1; got != want || entries[0].Name() != ".gitkeep" {
				t.Fatalf("expected no SPEC write on harness help, got entries=%v", entries)
			}
		})
	}
}

func TestRunFixHelpIsReadOnlyAndDescribesCommands(t *testing.T) {
	t.Parallel()

	for _, helpFlag := range []string{"--help", "-h"} {
		helpFlag := helpFlag
		t.Run(helpFlag, func(t *testing.T) {
			t.Parallel()

			tmp := t.TempDir()
			stdout := &bytes.Buffer{}
			app := NewApp(stdout, &bytes.Buffer{})
			if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
				t.Fatalf("init failed: %v", err)
			}

			restore := chdirExecution(t, tmp)
			defer restore()

			if err := app.Run(context.Background(), []string{"fix", helpFlag}); err != nil {
				t.Fatalf("fix help failed: %v", err)
			}
			got := stdout.String()
			for _, want := range []string{"namba fix", "--command run|plan", "repair the issue directly in the current workspace"} {
				if !strings.Contains(got, want) {
					t.Fatalf("expected fix help to contain %q, got %q", want, got)
				}
			}

			entries, err := os.ReadDir(filepath.Join(tmp, ".namba", "specs"))
			if err != nil {
				t.Fatalf("read specs dir: %v", err)
			}
			if got, want := len(entries), 1; got != want || entries[0].Name() != ".gitkeep" {
				t.Fatalf("expected no SPEC write on help, got entries=%v", entries)
			}
		})
	}
}

func TestRunFixCommandPlanCreatesBugfixSpec(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"fix", "--command", "plan", "startup", "panic"}); err != nil {
		t.Fatalf("fix plan failed: %v", err)
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

func TestRunHarnessCreatesHarnessSpec(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"harness", "design reusable agent/skill system"}); err != nil {
		t.Fatalf("harness failed: %v", err)
	}

	spec := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "spec.md"))
	for _, want := range []string{"## Problem", "design reusable agent/skill system", "`namba harness", "Do not create a second artifact model"} {
		if !strings.Contains(spec, want) {
			t.Fatalf("expected harness spec to contain %q, got %q", want, spec)
		}
	}

	plan := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "plan.md"))
	for _, want := range []string{"top-level `namba harness` command contract", "Codex-native execution topology", ".namba/specs/SPEC-001/reviews/"} {
		if !strings.Contains(plan, want) {
			t.Fatalf("expected harness plan to contain %q, got %q", want, plan)
		}
	}

	acceptance := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "acceptance.md"))
	for _, want := range []string{"`namba harness \"<description>\"`", "does not invent a second planning package type", "Claude-only primitives"} {
		if !strings.Contains(acceptance, want) {
			t.Fatalf("expected harness acceptance to contain %q, got %q", want, acceptance)
		}
	}

	readiness := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "reviews", "readiness.md"))
	if !strings.Contains(readiness, "Cleared reviews: 0/3") {
		t.Fatalf("unexpected harness review readiness scaffold: %q", readiness)
	}
}

func TestRunFixDirectRepairExecutesCurrentWorkspaceAndSyncs(t *testing.T) {
	t.Parallel()

	tmp := canonicalTempDir(t)
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.detectCodexCapabilities = func(context.Context, string, executionRequest) (codexCapabilityMatrix, error) {
		return testCodexCapabilities(), nil
	}
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	writeTestFile(t, filepath.Join(tmp, "README.md"), "# Demo\n\nThis repository exercises the direct fix flow.\n")

	restore := chdirExecution(t, tmp)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		switch name {
		case "codex", "git":
			return name, nil
		default:
			return "", errors.New("missing dependency")
		}
	}
	cases := []struct {
		name string
		args []string
	}{
		{name: "default", args: []string{"fix", "startup", "panic"}},
		{name: "explicit run", args: []string{"fix", "--command", "run", "startup", "panic"}},
	}

	var prompts []string
	for i, tc := range cases {
		stdout.Reset()
		var promptArg string
		app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
			switch {
			case isCodexExec(name, args):
				promptArg = args[len(args)-1]
				return "repair output", nil
			case isShellCommand(name):
				return "validation ok", nil
			default:
				t.Fatalf("unexpected command: %s %v", name, args)
				return "", nil
			}
		}
		if err := app.Run(context.Background(), tc.args); err != nil {
			t.Fatalf("%s direct fix failed: %v", tc.name, err)
		}
		for _, want := range []string{"startup panic", "without creating a SPEC package", "Finish by running `namba sync`", "namba fix --command plan"} {
			if !strings.Contains(promptArg, want) {
				t.Fatalf("expected %s direct-fix prompt to contain %q, got %q", tc.name, want, promptArg)
			}
		}
		if request := mustReadExecutionRequest(t, filepath.Join(tmp, ".namba", "logs", "runs", "direct-fix-request.json")); request.SpecID != "DIRECT-FIX" || request.Mode != executionModeDefault {
			t.Fatalf("unexpected direct-fix request metadata: %+v", request)
		}
		result := mustReadExecutionResult(t, filepath.Join(tmp, ".namba", "logs", "runs", "direct-fix-execution.json"))
		if !result.Succeeded || result.SpecID != "DIRECT-FIX" || result.ExecutionMode != "default" {
			t.Fatalf("unexpected direct-fix result: %+v", result)
		}
		if !strings.Contains(stdout.String(), "Executed direct fix with codex") || !strings.Contains(stdout.String(), "Synced NambaAI artifacts.") {
			t.Fatalf("expected direct-fix output to report execution and sync, got %q", stdout.String())
		}
		prompts = append(prompts, promptArg)
		if i == 0 {
			continue
		}
		if prompts[i] != prompts[0] {
			t.Fatalf("expected %s prompt to match default direct-fix prompt\nfirst=%q\nsecond=%q", tc.name, prompts[0], prompts[i])
		}
	}

	entries, err := os.ReadDir(filepath.Join(tmp, ".namba", "specs"))
	if err != nil {
		t.Fatalf("read specs dir: %v", err)
	}
	if got, want := len(entries), 1; got != want || entries[0].Name() != ".gitkeep" {
		t.Fatalf("expected no SPEC write on direct fix, got entries=%v", entries)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".namba", "project", "change-summary.md")); err != nil {
		t.Fatalf("expected sync artifacts after direct fix, stat err=%v", err)
	}
}

func TestRunPlanRejectsFlagOnlyInvocation(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"plan", "--dry-run"}); err == nil || !strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("expected plan flag-only failure, got %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(tmp, ".namba", "specs"))
	if err != nil {
		t.Fatalf("read specs dir: %v", err)
	}
	if got, want := len(entries), 1; got != want || entries[0].Name() != ".gitkeep" {
		t.Fatalf("expected no SPEC write on malformed plan, got entries=%v", entries)
	}
}

func TestRunHarnessRejectsFlagOnlyInvocation(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"harness", "--dry-run"}); err == nil || !strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("expected harness flag-only failure, got %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(tmp, ".namba", "specs"))
	if err != nil {
		t.Fatalf("read specs dir: %v", err)
	}
	if got, want := len(entries), 1; got != want || entries[0].Name() != ".gitkeep" {
		t.Fatalf("expected no SPEC write on malformed harness, got entries=%v", entries)
	}
}

func TestRunFixRejectsMalformedCommandFlag(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing command value",
			args:    []string{"fix", "--command"},
			wantErr: "requires a value of run or plan",
		},
		{
			name:    "flag-only probe",
			args:    []string{"fix", "--dry-run"},
			wantErr: "unknown flag",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmp := t.TempDir()
			app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
			if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
				t.Fatalf("init failed: %v", err)
			}

			restore := chdirExecution(t, tmp)
			defer restore()

			if err := app.Run(context.Background(), tc.args); err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected malformed fix command failure containing %q, got %v", tc.wantErr, err)
			}

			entries, err := os.ReadDir(filepath.Join(tmp, ".namba", "specs"))
			if err != nil {
				t.Fatalf("read specs dir: %v", err)
			}
			if got, want := len(entries), 1; got != want || entries[0].Name() != ".gitkeep" {
				t.Fatalf("expected no SPEC write on malformed fix, got entries=%v", entries)
			}
		})
	}
}

func TestRunFixCommandPlanAllowsFlagLikeTextInIssueDescription(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"fix", "--command", "plan", "--dry-run crashes startup"}); err != nil {
		t.Fatalf("fix plan failed: %v", err)
	}

	spec := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "spec.md"))
	if !strings.Contains(spec, "--dry-run crashes startup") {
		t.Fatalf("expected fix spec to preserve flag-like issue description, got %q", spec)
	}
}

func TestRunFixDirectRepairAllowsFlagLikeTextInIssueDescription(t *testing.T) {
	t.Parallel()

	tmp := canonicalTempDir(t)
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	app.detectCodexCapabilities = func(context.Context, string, executionRequest) (codexCapabilityMatrix, error) {
		return testCodexCapabilities(), nil
	}
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	writeTestFile(t, filepath.Join(tmp, "README.md"), "# Demo\n\nThis repository exercises the direct fix flow.\n")

	restore := chdirExecution(t, tmp)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		switch name {
		case "codex", "git":
			return name, nil
		default:
			return "", errors.New("missing dependency")
		}
	}

	var promptArg string
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexExec(name, args):
			promptArg = args[len(args)-1]
			return "repair output", nil
		case isShellCommand(name):
			return "validation ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"fix", "--dry-run crashes startup"}); err != nil {
		t.Fatalf("direct fix failed: %v", err)
	}
	if !strings.Contains(promptArg, "--dry-run crashes startup") {
		t.Fatalf("expected direct-fix prompt to preserve flag-like issue description, got %q", promptArg)
	}
}

func TestRunFixRequiresIssueDescription(t *testing.T) {
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
	if !strings.Contains(err.Error(), "fix requires an issue description") {
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
