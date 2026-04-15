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

func TestResolveFixSubcommandDefinitions(t *testing.T) {
	t.Parallel()

	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	for _, name := range []string{"plan", "run"} {
		definition, ok := app.resolveFixSubcommand(name)
		if !ok || definition.Run == nil {
			t.Fatalf("expected fix subcommand %q to resolve with runner, got %#v ok=%v", name, definition, ok)
		}
		if strings.TrimSpace(definition.BehaviorSummary) == "" {
			t.Fatalf("expected fix subcommand %q to expose behavior summary, got %#v", name, definition)
		}
	}

	if _, ok := app.resolveFixSubcommand("missing"); ok {
		t.Fatalf("did not expect unknown fix subcommand to resolve")
	}
}

func TestFixUsageTextMatchesSubcommandBehaviorSummaries(t *testing.T) {
	t.Parallel()

	got := fixUsageText()
	if !strings.HasPrefix(got, "namba fix\n\nUsage:\n") {
		t.Fatalf("unexpected fix usage header: %q", got)
	}
	for _, want := range []string{
		"  namba fix [--command run|plan] \"<issue description>\"",
		"  namba fix [--command run|plan] -- \"<issue description with flag-like text>\"",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected fix usage to contain %q, got %q", want, got)
		}
	}
	firstExampleIndex := strings.Index(got, "  namba fix [--command run|plan] \"<issue description>\"")
	secondExampleIndex := strings.Index(got, "  namba fix [--command run|plan] -- \"<issue description with flag-like text>\"")
	if firstExampleIndex >= secondExampleIndex {
		t.Fatalf("expected fix usage example lines to stay ordered, got %q", got)
	}

	lastIndex := -1
	for _, definition := range fixSubcommandDefinitions() {
		index := strings.Index(got, definition.BehaviorSummary)
		if index < 0 {
			t.Fatalf("expected fix usage to contain %q, got %q", definition.BehaviorSummary, got)
		}
		if index <= lastIndex {
			t.Fatalf("expected behavior summary %q to appear after previous summary in %q", definition.BehaviorSummary, got)
		}
		lastIndex = index
	}
	if secondExampleIndex >= strings.Index(got, fixSubcommandDefinitions()[0].BehaviorSummary) {
		t.Fatalf("expected fix usage example lines to stay before behavior summaries, got %q", got)
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

func TestLoadSpecPackageScaffoldContextLoadsNextSpecAndConfigs(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	scaffoldCtx, err := app.loadSpecPackageScaffoldContext("plan", "improve review workflow")
	if err != nil {
		t.Fatalf("loadSpecPackageScaffoldContext failed: %v", err)
	}

	resolvedTmp, err := filepath.EvalSymlinks(tmp)
	if err != nil {
		t.Fatalf("eval temp root: %v", err)
	}
	resolvedRoot, err := filepath.EvalSymlinks(scaffoldCtx.Root)
	if err != nil {
		t.Fatalf("eval scaffold root: %v", err)
	}
	if resolvedRoot != resolvedTmp || scaffoldCtx.Kind != "plan" || scaffoldCtx.Description != "improve review workflow" {
		t.Fatalf("unexpected scaffold context identity: %+v", scaffoldCtx)
	}
	if scaffoldCtx.SpecID != "SPEC-001" {
		t.Fatalf("expected next spec id SPEC-001, got %+v", scaffoldCtx)
	}
	if scaffoldCtx.ProjectCfg.Name == "" || scaffoldCtx.QualityCfg.DevelopmentMode == "" {
		t.Fatalf("expected scaffold context to load project and quality config, got %+v", scaffoldCtx)
	}
}

func TestBuildSpecPackageScaffoldOutputsIncludesCoreAndReviewArtifacts(t *testing.T) {
	scaffoldCtx := specPackageScaffoldContext{
		Root:        "/repo",
		Kind:        "harness",
		Description: "design reusable agent/skill system",
		SpecID:      "SPEC-003",
		ProjectCfg: projectConfig{
			Name:        "namba-ai",
			ProjectType: "existing",
			Language:    "go",
		},
		QualityCfg: qualityConfig{DevelopmentMode: "tdd"},
	}

	outputs := buildSpecPackageScaffoldOutputs(scaffoldCtx)
	expectedPaths := []string{
		filepath.ToSlash(filepath.Join(specsDir, "SPEC-003", "spec.md")),
		filepath.ToSlash(filepath.Join(specsDir, "SPEC-003", "plan.md")),
		filepath.ToSlash(filepath.Join(specsDir, "SPEC-003", "acceptance.md")),
		filepath.ToSlash(filepath.Join(specsDir, "SPEC-003", "reviews", "product.md")),
		filepath.ToSlash(filepath.Join(specsDir, "SPEC-003", "reviews", "engineering.md")),
		filepath.ToSlash(filepath.Join(specsDir, "SPEC-003", "reviews", "design.md")),
		filepath.ToSlash(filepath.Join(specsDir, "SPEC-003", "reviews", "readiness.md")),
	}
	if len(outputs) != len(expectedPaths) {
		t.Fatalf("expected %d scaffold outputs, got %d: %+v", len(expectedPaths), len(outputs), outputs)
	}
	for _, path := range expectedPaths {
		if _, ok := outputs[path]; !ok {
			t.Fatalf("expected scaffold outputs to include %s, got %+v", path, outputs)
		}
	}
	if !strings.Contains(outputs[filepath.ToSlash(filepath.Join(specsDir, "SPEC-003", "spec.md"))], "`namba harness \"<description>\"`") {
		t.Fatalf("expected harness spec scaffold content, got %q", outputs[filepath.ToSlash(filepath.Join(specsDir, "SPEC-003", "spec.md"))])
	}
	if !strings.Contains(outputs[filepath.ToSlash(filepath.Join(specsDir, "SPEC-003", "reviews", "readiness.md"))], "Cleared reviews: 0/3") {
		t.Fatalf("expected readiness scaffold content, got %q", outputs[filepath.ToSlash(filepath.Join(specsDir, "SPEC-003", "reviews", "readiness.md"))])
	}
}

func TestBuildSpecDocRoutesByKind(t *testing.T) {
	projectCfg := projectConfig{Name: "namba-ai", ProjectType: "existing", Language: "go"}
	qualityCfg := qualityConfig{DevelopmentMode: "tdd"}

	fixDoc := buildSpecDoc("fix", "SPEC-001", "startup panic", projectCfg, qualityCfg)
	if !strings.Contains(fixDoc, "Apply the smallest safe fix") || !strings.Contains(fixDoc, "Work type: fix") {
		t.Fatalf("unexpected fix spec doc: %q", fixDoc)
	}

	harnessDoc := buildSpecDoc("harness", "SPEC-002", "design reusable agent/skill system", projectCfg, qualityCfg)
	for _, want := range []string{"`namba harness \"<description>\"`", "Do not create a second artifact model", "Codex-native harness change"} {
		if !strings.Contains(harnessDoc, want) {
			t.Fatalf("expected harness spec doc to contain %q, got %q", want, harnessDoc)
		}
	}

	featureDoc := buildSpecDoc("plan", "SPEC-003", "improve review workflow", projectCfg, qualityCfg)
	if !strings.Contains(featureDoc, "Implement the requested change under the normal feature-planning workflow") || !strings.Contains(featureDoc, "Work type: plan") {
		t.Fatalf("unexpected feature spec doc: %q", featureDoc)
	}
}

func TestBuildSpecPlanDocRoutesByKind(t *testing.T) {
	fixPlan := buildSpecPlanDoc("fix", "SPEC-001")
	if !strings.Contains(fixPlan, "Implement the smallest safe fix") || !strings.Contains(fixPlan, ".namba/specs/SPEC-001/reviews/") {
		t.Fatalf("unexpected fix spec plan: %q", fixPlan)
	}

	harnessPlan := buildSpecPlanDoc("harness", "SPEC-002")
	for _, want := range []string{"top-level `namba harness` command contract", "Codex-native execution topology", ".namba/specs/SPEC-002/reviews/"} {
		if !strings.Contains(harnessPlan, want) {
			t.Fatalf("expected harness spec plan to contain %q, got %q", want, harnessPlan)
		}
	}

	featurePlan := buildSpecPlanDoc("plan", "SPEC-003")
	if !strings.Contains(featurePlan, "Implement the requested change") || !strings.Contains(featurePlan, ".namba/specs/SPEC-003/reviews/") {
		t.Fatalf("unexpected feature spec plan: %q", featurePlan)
	}
}

func TestBuildSpecAcceptanceDocRoutesByKind(t *testing.T) {
	fixAcceptance := buildSpecAcceptanceDoc("fix", "startup panic", "tdd")
	for _, want := range []string{"The reported issue described below is resolved", "A regression test covering the fix is present"} {
		if !strings.Contains(fixAcceptance, want) {
			t.Fatalf("expected fix acceptance doc to contain %q, got %q", want, fixAcceptance)
		}
	}

	harnessAcceptance := buildSpecAcceptanceDoc("harness", "design reusable agent/skill system", "prod")
	for _, want := range []string{"`namba harness \"<description>\"` creates the next sequential `SPEC-XXX` package", "Existing planning behavior is preserved while adding the harness surface"} {
		if !strings.Contains(harnessAcceptance, want) {
			t.Fatalf("expected harness acceptance doc to contain %q, got %q", want, harnessAcceptance)
		}
	}

	featureAcceptance := buildSpecAcceptanceDoc("plan", "improve review workflow", "tdd")
	for _, want := range []string{"The requested behavior described below is implemented", "Tests covering the new behavior are present"} {
		if !strings.Contains(featureAcceptance, want) {
			t.Fatalf("expected feature acceptance doc to contain %q, got %q", want, featureAcceptance)
		}
	}
}

func TestAcceptanceDocModeLines(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "feature prod",
			got:  featureAcceptanceModeLine("prod"),
			want: "- [ ] Existing behavior is preserved while improving the target area",
		},
		{
			name: "harness tdd",
			got:  harnessAcceptanceModeLine("tdd"),
			want: "- [ ] Tests covering the new command/scaffold behavior are present",
		},
		{
			name: "fix prod",
			got:  fixAcceptanceModeLine("prod"),
			want: "- [ ] A targeted reproduction or verification step is documented",
		},
	}

	for _, tc := range cases {
		if tc.got != tc.want {
			t.Fatalf("%s mismatch: got %q want %q", tc.name, tc.got, tc.want)
		}
	}
}

func TestMaterializeSpecPackageScaffoldOutputsWritesArtifacts(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	scaffoldCtx := specPackageScaffoldContext{
		Root:   tmp,
		SpecID: "SPEC-001",
	}
	outputs := map[string]string{
		filepath.ToSlash(filepath.Join(specsDir, "SPEC-001", "spec.md")):                 "spec",
		filepath.ToSlash(filepath.Join(specsDir, "SPEC-001", "reviews", "readiness.md")): "ready",
	}
	if err := app.materializeSpecPackageScaffoldOutputs(scaffoldCtx, outputs); err != nil {
		t.Fatalf("materializeSpecPackageScaffoldOutputs failed: %v", err)
	}

	if got := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "spec.md")); got != "spec" {
		t.Fatalf("expected materialized spec file, got %q", got)
	}
	if got := mustReadFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "reviews", "readiness.md")); got != "ready" {
		t.Fatalf("expected materialized readiness file, got %q", got)
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

			err := app.Run(context.Background(), tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected malformed fix command failure containing %q, got %v", tc.wantErr, err)
			}
			if tc.name == "missing command value" {
				for _, want := range []string{
					"  namba fix [--command run|plan] \"<issue description>\"",
					"Use --command plan to scaffold the next bugfix SPEC package under .namba/specs/.",
					"Use --command run, or omit --command, to repair the issue directly in the current workspace.",
				} {
					if !strings.Contains(err.Error(), want) {
						t.Fatalf("expected malformed fix error to include %q, got %v", want, err)
					}
				}
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
