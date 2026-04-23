package namba

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveCodexSubcommandDefinitions(t *testing.T) {
	t.Parallel()

	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	definition, ok := app.resolveCodexSubcommand("access")
	if !ok || definition.Run == nil {
		t.Fatalf("expected codex subcommand %q to resolve with runner, got %#v ok=%v", "access", definition, ok)
	}
	if strings.TrimSpace(definition.UsageSummary) == "" {
		t.Fatalf("expected codex access subcommand to expose usage summary, got %#v", definition)
	}
	if definition.UsageText == nil {
		t.Fatalf("expected codex access subcommand to expose usage text")
	}

	if _, ok := app.resolveCodexSubcommand("missing"); ok {
		t.Fatal("did not expect unknown codex subcommand to resolve")
	}
}

func TestCodexUsageTextMatchesSubcommandUsageSummaries(t *testing.T) {
	t.Parallel()

	got := codexUsageText()
	if !strings.HasPrefix(got, "namba codex\n\nUsage:\n") {
		t.Fatalf("unexpected codex usage header: %q", got)
	}
	if !strings.Contains(got, "\n\nBehavior:\n  Inspect or update repo-owned Codex access defaults from the project root.\n") {
		t.Fatalf("expected codex behavior block, got %q", got)
	}
	for _, definition := range codexSubcommandDefinitions() {
		if !strings.Contains(got, definition.UsageSummary) {
			t.Fatalf("expected codex usage to contain %q, got %q", definition.UsageSummary, got)
		}
	}
}

func TestResolveCodexAccessChoicePreservesPresetsAndCustomFallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		approvalPolicy string
		sandboxMode    string
		wantPresetID   string
		wantLabel      string
	}{
		{
			name:           "balanced preset",
			approvalPolicy: "on-request",
			sandboxMode:    "workspace-write",
			wantPresetID:   "balanced",
			wantLabel:      "Balanced workspace",
		},
		{
			name:           "full access preset",
			approvalPolicy: "never",
			sandboxMode:    "danger-full-access",
			wantPresetID:   "full-access",
			wantLabel:      "Full access",
		},
		{
			name:           "custom fallback",
			approvalPolicy: "never",
			sandboxMode:    "read-only",
			wantPresetID:   codexAccessPresetCustom,
			wantLabel:      "Custom access",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := resolveCodexAccessChoice(tt.approvalPolicy, tt.sandboxMode)
			if err != nil {
				t.Fatalf("resolveCodexAccessChoice returned error: %v", err)
			}
			if got.PresetID != tt.wantPresetID || got.Label != tt.wantLabel {
				t.Fatalf("unexpected access choice: %+v", got)
			}
		})
	}
}

func TestRunCodexAccessHelpIsReadOnly(t *testing.T) {
	t.Parallel()

	tmp := canonicalTempDir(t)
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	readmePath := filepath.Join(tmp, "README.md")
	writeTestFile(t, readmePath, "sentinel\n")

	if err := app.Run(context.Background(), []string{"codex", "access", "--help"}); err != nil {
		t.Fatalf("codex access help failed: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{"namba codex access", "Usage:", "--approval-policy", "--sandbox-mode"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected help output to contain %q, got %q", want, got)
		}
	}
	if gotReadme := mustReadFile(t, readmePath); gotReadme != "sentinel\n" {
		t.Fatalf("expected help path to stay read-only, got %q", gotReadme)
	}
}

func TestRunCodexAccessInspectsCurrentStateWithoutMutation(t *testing.T) {
	t.Parallel()

	tmp := canonicalTempDir(t)
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	manifestBefore := mustReadFile(t, filepath.Join(tmp, ".namba", "manifest.json"))

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"codex", "access"}); err != nil {
		t.Fatalf("codex access inspect failed: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{"Balanced workspace", "approval_policy: on-request", "sandbox_mode: workspace-write"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected inspect output to contain %q, got %q", want, got)
		}
	}
	if manifestAfter := mustReadFile(t, filepath.Join(tmp, ".namba", "manifest.json")); manifestAfter != manifestBefore {
		t.Fatalf("expected inspect path to avoid manifest churn\nbefore=%q\nafter=%q", manifestBefore, manifestAfter)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".namba", "logs", "session-refresh-required.json")); !os.IsNotExist(err) {
		t.Fatalf("expected inspect path to suppress session refresh notice, stat err=%v", err)
	}
}

func TestRunCodexAccessNoOpSuppressesSessionRefresh(t *testing.T) {
	t.Parallel()

	tmp := canonicalTempDir(t)
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	manifestBefore := mustReadFile(t, filepath.Join(tmp, ".namba", "manifest.json"))

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"codex", "access", "--approval-policy", "on-request", "--sandbox-mode", "workspace-write"}); err != nil {
		t.Fatalf("codex access no-op failed: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, "No change") {
		t.Fatalf("expected no-op output to confirm no change, got %q", got)
	}
	if manifestAfter := mustReadFile(t, filepath.Join(tmp, ".namba", "manifest.json")); manifestAfter != manifestBefore {
		t.Fatalf("expected no-op path to avoid manifest churn\nbefore=%q\nafter=%q", manifestBefore, manifestAfter)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".namba", "logs", "session-refresh-required.json")); !os.IsNotExist(err) {
		t.Fatalf("expected no-op path to suppress session refresh notice, stat err=%v", err)
	}
}

func TestRunCodexAccessUpdatesManagedConfigWithoutClobberingOtherFiles(t *testing.T) {
	t.Parallel()

	tmp := canonicalTempDir(t)
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	readmePath := filepath.Join(tmp, "README.md")
	writeTestFile(t, readmePath, "user-authored readme\n")

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"codex", "access", "--approval-policy", "never", "--sandbox-mode", "read-only"}); err != nil {
		t.Fatalf("codex access update failed: %v", err)
	}

	system := mustReadFile(t, filepath.Join(tmp, ".namba", "config", "sections", "system.yaml"))
	for _, want := range []string{"approval_policy: never", "sandbox_mode: read-only"} {
		if !strings.Contains(system, want) {
			t.Fatalf("expected updated system config to contain %q, got %q", want, system)
		}
	}

	codexConfig := mustReadFile(t, filepath.Join(tmp, ".codex", "config.toml"))
	for _, want := range []string{`approval_policy = "never"`, `sandbox_mode = "read-only"`} {
		if !strings.Contains(codexConfig, want) {
			t.Fatalf("expected updated codex config to contain %q, got %q", want, codexConfig)
		}
	}

	if gotReadme := mustReadFile(t, readmePath); gotReadme != "user-authored readme\n" {
		t.Fatalf("expected unrelated README content to stay untouched, got %q", gotReadme)
	}

	got := stdout.String()
	for _, want := range []string{"Updated Codex access defaults", "Session refresh required", ".codex/config.toml"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected update output to contain %q, got %q", want, got)
		}
	}

	notice := mustReadFile(t, filepath.Join(tmp, ".namba", "logs", "session-refresh-required.json"))
	if !strings.Contains(notice, ".codex/config.toml") {
		t.Fatalf("expected refresh notice to mention codex config, got %q", notice)
	}
}

func TestRunCodexAccessRejectsInvalidUsageAndBrokenManagedConfig(t *testing.T) {
	t.Parallel()

	t.Run("outside repo", func(t *testing.T) {
		t.Parallel()

		tmp := canonicalTempDir(t)
		app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
		restore := chdirExecution(t, tmp)
		defer restore()

		err := app.Run(context.Background(), []string{"codex", "access"})
		if err == nil || !strings.Contains(err.Error(), "no NambaAI project found") {
			t.Fatalf("expected missing-repo error, got %v", err)
		}
	})

	t.Run("half update rejected", func(t *testing.T) {
		t.Parallel()

		tmp := canonicalTempDir(t)
		app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
		if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		restore := chdirExecution(t, tmp)
		defer restore()

		err := app.Run(context.Background(), []string{"codex", "access", "--approval-policy", "never"})
		if err == nil || !strings.Contains(err.Error(), "pass both --approval-policy and --sandbox-mode") {
			t.Fatalf("expected half-update error, got %v", err)
		}
	})

	t.Run("invalid current config rejected", func(t *testing.T) {
		t.Parallel()

		tmp := canonicalTempDir(t)
		app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
		if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "system.yaml"), "runner: codex\napproval_policy: maybe\nsandbox_mode: workspace-write\n")

		restore := chdirExecution(t, tmp)
		defer restore()

		err := app.Run(context.Background(), []string{"codex", "access", "--approval-policy", "never", "--sandbox-mode", "read-only"})
		if err == nil || !strings.Contains(err.Error(), "current managed config") {
			t.Fatalf("expected invalid managed config error, got %v", err)
		}
	})
}

func TestRunInitWizardIncludesGuidedCodexAccessStep(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.stdin = strings.NewReader(strings.Repeat("\n", 16))

	defaults := initProfile{
		ProjectName:           "demo",
		ProjectType:           "new",
		Language:              "go",
		Framework:             "none",
		DevelopmentMode:       "tdd",
		ConversationLanguage:  "en",
		DocumentationLanguage: "en",
		CommentLanguage:       "en",
		ApprovalPolicy:        "on-request",
		SandboxMode:           "workspace-write",
		GitMode:               "manual",
		GitProvider:           "github",
		GitLabInstanceURL:     "https://gitlab.com",
		BranchPerWork:         true,
		BranchBase:            "main",
		SpecBranchPrefix:      "spec/",
		TaskBranchPrefix:      "task/",
		PRBaseBranch:          "main",
		PRLanguage:            "en",
		CodexReviewComment:    "@codex review",
		AutoCodexReview:       true,
		AgentMode:             "single",
		StatusLinePreset:      "namba",
		UserName:              "Developer",
		CreatedAt:             "2026-04-23T10:00:00Z",
	}

	if _, err := app.runInitWizard(defaults); err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{"Codex access preset", "Balanced workspace", "approval_policy=on-request", "sandbox_mode=workspace-write"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected wizard output to contain %q, got %q", want, got)
		}
	}
}
