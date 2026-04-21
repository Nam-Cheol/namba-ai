package namba

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpFlowsAreReadOnlyAndCommandSpecific(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		args      []string
		setupRepo bool
		sentinel  string
		want      []string
		stub      func(*App, *testing.T)
	}{
		{
			name: "init --help",
			args: []string{"init", "--help"},
			want: []string{"namba init", "Usage:"},
		},
		{
			name: "doctor --help",
			args: []string{"doctor", "--help"},
			want: []string{"namba doctor", "Usage:"},
		},
		{
			name: "status --help",
			args: []string{"status", "--help"},
			want: []string{"namba status", "Usage:"},
		},
		{
			name: "project --help outside repo",
			args: []string{"project", "--help"},
			want: []string{"namba project", "Usage:"},
		},
		{
			name:      "project --help",
			args:      []string{"project", "--help"},
			setupRepo: true,
			sentinel:  filepath.Join(".namba", "project", "change-summary.md"),
			want:      []string{"namba project", "Usage:"},
		},
		{
			name:      "regen --help",
			args:      []string{"regen", "--help"},
			setupRepo: true,
			sentinel:  "AGENTS.md",
			want:      []string{"namba regen", "Usage:"},
		},
		{
			name: "update --help",
			args: []string{"update", "--help"},
			want: []string{"namba update", "Usage:"},
			stub: func(app *App, t *testing.T) {
				app.downloadURL = func(context.Context, string) ([]byte, error) {
					t.Fatal("unexpected download during update help")
					return nil, nil
				}
			},
		},
		{
			name: "plan --help",
			args: []string{"plan", "--help"},
			want: []string{"namba plan", "Usage:", "Create the next feature SPEC package"},
		},
		{
			name: "harness --help",
			args: []string{"harness", "--help"},
			want: []string{"namba harness", "Usage:", "Create the next harness-oriented SPEC package"},
		},
		{
			name: "fix --help",
			args: []string{"fix", "--help"},
			want: []string{"namba fix", "Usage:", "Use --command run, or omit --command, to repair the issue directly in the current workspace."},
		},
		{
			name: "run --help",
			args: []string{"run", "--help"},
			want: []string{"namba run", "Usage:"},
		},
		{
			name: "sync --help outside repo",
			args: []string{"sync", "--help"},
			want: []string{"namba sync", "Usage:"},
		},
		{
			name:      "sync --help",
			args:      []string{"sync", "--help"},
			setupRepo: true,
			sentinel:  filepath.Join(".namba", "project", "change-summary.md"),
			want:      []string{"namba sync", "Usage:"},
		},
		{
			name: "pr --help",
			args: []string{"pr", "--help"},
			want: []string{"namba pr", "Usage:"},
		},
		{
			name: "land --help",
			args: []string{"land", "--help"},
			want: []string{"namba land", "Usage:"},
		},
		{
			name: "release --help",
			args: []string{"release", "--help"},
			want: []string{"namba release", "Usage:"},
		},
		{
			name: "worktree --help",
			args: []string{"worktree", "--help"},
			want: []string{"namba worktree", "Usage:"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := ""
			if tc.setupRepo {
				root = canonicalTempDir(t)
				stdout := &bytes.Buffer{}
				app := NewApp(stdout, &bytes.Buffer{})
				if err := app.Run(context.Background(), []string{"init", root, "--yes"}); err != nil {
					t.Fatalf("init failed: %v", err)
				}
				stdout.Reset()
				if tc.sentinel != "" {
					writeTestFile(t, filepath.Join(root, tc.sentinel), "sentinel\n")
				}
				restore := chdirExecution(t, root)
				defer restore()
				if tc.stub != nil {
					tc.stub(app, t)
				}
				if err := app.Run(context.Background(), tc.args); err != nil {
					t.Fatalf("help invocation failed: %v", err)
				}
				got := stdout.String()
				for _, want := range tc.want {
					if !strings.Contains(got, want) {
						t.Fatalf("expected help output to contain %q, got %q", want, got)
					}
				}
				if tc.sentinel != "" {
					gotSentinel := mustReadFile(t, filepath.Join(root, tc.sentinel))
					if gotSentinel != "sentinel\n" {
						t.Fatalf("expected %s to stay unchanged, got %q", tc.sentinel, gotSentinel)
					}
				}
				return
			}

			stdout := &bytes.Buffer{}
			app := NewApp(stdout, &bytes.Buffer{})
			if tc.stub != nil {
				tc.stub(app, t)
			}
			if err := app.Run(context.Background(), tc.args); err != nil {
				t.Fatalf("help invocation failed: %v", err)
			}
			got := stdout.String()
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Fatalf("expected help output to contain %q, got %q", want, got)
				}
			}
		})
	}
}

func TestHelpSubcommandMatchesCommandHelpForPlanningCommands(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "plan",
			args: []string{"plan", "--help"},
			want: []string{"namba plan", "Create the next feature SPEC package"},
		},
		{
			name: "harness",
			args: []string{"harness", "--help"},
			want: []string{"namba harness", "Create the next harness-oriented SPEC package"},
		},
		{
			name: "fix",
			args: []string{"fix", "--help"},
			want: []string{"namba fix", "Use --command run, or omit --command, to repair the issue directly in the current workspace."},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			root := canonicalTempDir(t)
			stdout := &bytes.Buffer{}
			app := NewApp(stdout, &bytes.Buffer{})
			if err := app.Run(context.Background(), []string{"init", root, "--yes"}); err != nil {
				t.Fatalf("init failed: %v", err)
			}

			restore := chdirExecution(t, root)
			defer restore()

			stdout.Reset()
			if err := app.Run(context.Background(), tc.args); err != nil {
				t.Fatalf("flag help failed: %v", err)
			}
			flagHelp := stdout.String()
			for _, want := range tc.want {
				if !strings.Contains(flagHelp, want) {
					t.Fatalf("expected flag help to contain %q, got %q", want, flagHelp)
				}
			}

			stdout.Reset()
			if err := app.Run(context.Background(), []string{"help", tc.name}); err != nil {
				t.Fatalf("subcommand help failed: %v", err)
			}
			subcommandHelp := stdout.String()
			for _, want := range tc.want {
				if !strings.Contains(subcommandHelp, want) {
					t.Fatalf("expected subcommand help to contain %q, got %q", want, subcommandHelp)
				}
			}
		})
	}
}

func TestCommandUsageTextMatchesPublicTopLevelCommandDefinitions(t *testing.T) {
	t.Parallel()

	for _, definition := range publicTopLevelCommandDefinitions() {
		text, ok := commandUsageText(definition.Name)
		if !ok {
			t.Fatalf("expected usage text for public command %q", definition.Name)
		}
		if definition.UsageText == nil {
			t.Fatalf("expected public command %q to provide usage text", definition.Name)
		}
		if text != definition.UsageText() {
			t.Fatalf("unexpected usage text for %q", definition.Name)
		}
	}
}

func TestResolveTopLevelCommandKeepsInternalCreateHiddenFromUsage(t *testing.T) {
	t.Parallel()

	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	publicCommand, ok := app.resolveTopLevelCommand("sync")
	if !ok || publicCommand.Run == nil {
		t.Fatalf("expected sync command to resolve with runner, got %#v ok=%v", publicCommand, ok)
	}

	internalCommand, ok := app.resolveTopLevelCommand(internalCreateCommandName)
	if !ok || internalCommand.Run == nil {
		t.Fatalf("expected internal create command to resolve with runner, got %#v ok=%v", internalCommand, ok)
	}
	if _, ok := commandUsageText(internalCreateCommandName); ok {
		t.Fatalf("did not expect hidden internal create command to expose usage text")
	}

	if _, ok := app.resolveTopLevelCommand("missing"); ok {
		t.Fatalf("did not expect unknown command to resolve")
	}
}

func TestResolveTopLevelInvocationPreservesHelpAndUnknownCommandContracts(t *testing.T) {
	t.Parallel()

	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	fullUsageInvocation, err := app.resolveTopLevelInvocation([]string{"help"})
	if err != nil {
		t.Fatalf("expected help invocation to resolve, got %v", err)
	}
	if fullUsageInvocation.UsageText != usageText() {
		t.Fatalf("expected bare help to resolve full usage, got %q", fullUsageInvocation.UsageText)
	}

	syncUsage, ok := commandUsageText("sync")
	if !ok {
		t.Fatal("expected sync usage text to resolve")
	}
	syncUsageInvocation, err := app.resolveTopLevelInvocation([]string{"help", "sync"})
	if err != nil {
		t.Fatalf("expected help sync invocation to resolve, got %v", err)
	}
	if syncUsageInvocation.UsageText != syncUsage {
		t.Fatalf("expected help sync to resolve sync usage, got %q", syncUsageInvocation.UsageText)
	}

	helpHelpInvocation, err := app.resolveTopLevelInvocation([]string{"help", "--help"})
	if err != nil {
		t.Fatalf("expected help --help invocation to resolve, got %v", err)
	}
	if helpHelpInvocation.UsageText != usageText() {
		t.Fatalf("expected help --help to resolve full usage, got %q", helpHelpInvocation.UsageText)
	}

	commandInvocation, err := app.resolveTopLevelInvocation([]string{"sync"})
	if err != nil {
		t.Fatalf("expected sync invocation to resolve, got %v", err)
	}
	if commandInvocation.Command.Name != "sync" || commandInvocation.Command.Run == nil {
		t.Fatalf("expected sync command invocation to resolve runner, got %#v", commandInvocation)
	}
	if len(commandInvocation.Args) != 0 {
		t.Fatalf("expected sync invocation to preserve zero args, got %#v", commandInvocation.Args)
	}

	err = unknownTopLevelCommandError("missing")
	if !strings.Contains(err.Error(), `unknown command "missing"`) || !strings.Contains(err.Error(), usageText()) {
		t.Fatalf("expected unknown command error to preserve message shape, got %v", err)
	}

	if _, err := app.resolveTopLevelInvocation([]string{"help", "missing"}); err == nil || !strings.Contains(err.Error(), `unknown command "missing"`) {
		t.Fatalf("expected help missing to preserve unknown command error, got %v", err)
	}

	if _, err := app.resolveTopLevelInvocation([]string{"help", "run", "extra"}); err == nil || !strings.Contains(err.Error(), "help accepts at most one command") {
		t.Fatalf("expected malformed help invocation to preserve parse error, got %v", err)
	}
}

func TestUsageTextMatchesPublicTopLevelCommandUsageSummaries(t *testing.T) {
	t.Parallel()

	got := usageText()
	if !strings.HasPrefix(got, "NambaAI CLI\n\nUsage:\n  namba help [command]\n") {
		t.Fatalf("unexpected usage header: %q", got)
	}

	lastIndex := -1
	for _, definition := range publicTopLevelCommandDefinitions() {
		if strings.TrimSpace(definition.UsageSummary) == "" {
			t.Fatalf("expected usage summary for public command %q", definition.Name)
		}
		index := strings.Index(got, definition.UsageSummary)
		if index < 0 {
			t.Fatalf("expected usage text to contain %q, got %q", definition.UsageSummary, got)
		}
		if index <= lastIndex {
			t.Fatalf("expected usage summary %q to appear after previous command line in %q", definition.UsageSummary, got)
		}
		lastIndex = index
	}
}

func TestDescriptionScaffoldUsageTextPreservesPlanAndHarnessShape(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		got          string
		commandLine1 string
		commandLine2 string
		commandLine3 string
		behaviorLine string
		safetyLine   string
	}{
		{
			name:         "plan",
			got:          planUsageText(),
			commandLine1: "  namba plan \"<description>\"",
			commandLine2: "  namba plan -- \"<description with flag-like text>\"",
			commandLine3: "  namba plan --current-workspace \"<description>\"",
			behaviorLine: "  Create the next feature SPEC package under .namba/specs/ and seed review artifacts.",
			safetyLine:   "  Safe by default: resolve or create an isolated worktree/branch unless you explicitly pass --current-workspace.",
		},
		{
			name:         "harness",
			got:          harnessUsageText(),
			commandLine1: "  namba harness \"<description>\"",
			commandLine2: "  namba harness -- \"<description with flag-like text>\"",
			commandLine3: "  namba harness --current-workspace \"<description>\"",
			behaviorLine: "  Create the next harness-oriented SPEC package under .namba/specs/ and seed review artifacts.",
			safetyLine:   "  Safe by default: resolve or create an isolated worktree/branch unless you explicitly pass --current-workspace.",
		},
	}

	for _, tc := range cases {
		if !strings.Contains(tc.got, "Usage:\n") || !strings.Contains(tc.got, "\n\nBehavior:\n") {
			t.Fatalf("%s usage text missing section shape: %q", tc.name, tc.got)
		}
		for _, want := range []string{tc.commandLine1, tc.commandLine2, tc.commandLine3, tc.behaviorLine, tc.safetyLine} {
			if !strings.Contains(tc.got, want) {
				t.Fatalf("%s usage text missing %q in %q", tc.name, want, tc.got)
			}
		}
	}
}

func TestSingleUsageLineCommandUsageTextPreservesSimpleCommandShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		got          string
		usageLine    string
		behaviorLine string
	}{
		{
			name:         "init",
			got:          initUsageText(),
			usageLine:    "  namba init [path] [--yes] [--name NAME] [--mode tdd|ddd] [--project-type new|existing]",
			behaviorLine: "  Initialize the NambaAI scaffold, config, and repo-local Codex assets in the target directory.",
		},
		{
			name:         "doctor",
			got:          doctorUsageText(),
			usageLine:    "  namba doctor",
			behaviorLine: "  Inspect the current repository and local toolchain readiness without mutating project files.",
		},
		{
			name:         "status",
			got:          statusUsageText(),
			usageLine:    "  namba status",
			behaviorLine: "  Print a read-only summary of the current NambaAI repository state.",
		},
		{
			name:         "project",
			got:          projectUsageText(),
			usageLine:    "  namba project",
			behaviorLine: "  Refresh .namba/project/* docs and codemaps for the current repository.",
		},
		{
			name:         "regen",
			got:          regenUsageText(),
			usageLine:    "  namba regen",
			behaviorLine: "  Regenerate AGENTS, repo-local skills, Codex agents, and Codex config from .namba/config/sections/*.yaml.",
		},
		{
			name:         "update",
			got:          updateUsageText(),
			usageLine:    "  namba update [--version vX.Y.Z]",
			behaviorLine: "  Download and install the requested NambaAI release for the current platform.",
		},
		{
			name:         "run",
			got:          runUsageText(),
			usageLine:    "  namba run SPEC-XXX [--solo|--team|--parallel] [--dry-run]",
			behaviorLine: "  Execute the selected SPEC package with one runner, same-workspace team routing, or managed worktree fan-out.",
		},
		{
			name:         "sync",
			got:          syncUsageText(),
			usageLine:    "  namba sync",
			behaviorLine: "  Refresh README bundles, project docs, review readiness summaries, and PR/release support artifacts.",
		},
		{
			name:         "pr",
			got:          prUsageText(),
			usageLine:    "  namba pr \"<title>\" [--remote origin] [--no-sync] [--no-validate]",
			behaviorLine: "  Sync, validate, push the current work branch, and create or reuse a GitHub pull request into the base branch.",
		},
		{
			name:         "land",
			got:          landUsageText(),
			usageLine:    "  namba land [PR_NUMBER] [--wait] [--remote origin]",
			behaviorLine: "  Merge an approved pull request into the base branch and refresh the local base branch checkout.",
		},
		{
			name:         "release",
			got:          releaseUsageText(),
			usageLine:    "  namba release [--bump patch|minor|major] [--version vX.Y.Z] [--push] [--remote origin]",
			behaviorLine: "  Create a release tag from a clean main branch and optionally push main plus the tag.",
		},
	}

	for _, tc := range cases {
		if !strings.HasPrefix(tc.got, "namba "+tc.name+"\n\nUsage:\n") {
			t.Fatalf("%s usage text has unexpected header: %q", tc.name, tc.got)
		}
		usageIndex := strings.Index(tc.got, tc.usageLine)
		if usageIndex < 0 {
			t.Fatalf("%s usage text missing usage line %q in %q", tc.name, tc.usageLine, tc.got)
		}
		behaviorIndex := strings.Index(tc.got, tc.behaviorLine)
		if behaviorIndex < 0 {
			t.Fatalf("%s usage text missing behavior line %q in %q", tc.name, tc.behaviorLine, tc.got)
		}
		if usageIndex >= behaviorIndex {
			t.Fatalf("%s usage text expected usage line before behavior line in %q", tc.name, tc.got)
		}
	}
}

func TestHandleNoArgTopLevelCommandKeepsHelpAndUsageErrorContracts(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})

	handled, err := app.handleNoArgTopLevelCommand("sync", []string{"--help"})
	if !handled || err != nil {
		t.Fatalf("expected sync help to be handled without error, got handled=%v err=%v", handled, err)
	}
	if got := stdout.String(); got != syncUsageText() {
		t.Fatalf("unexpected sync help output: %q", got)
	}

	stdout.Reset()
	handled, err = app.handleNoArgTopLevelCommand("doctor", []string{"extra"})
	if !handled || err == nil {
		t.Fatalf("expected doctor extra arg to be handled with error, got handled=%v err=%v", handled, err)
	}
	for _, want := range []string{"doctor does not accept arguments", doctorUsageText()} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected doctor usage error to contain %q, got %q", want, err.Error())
		}
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("expected no stdout on usage error, got %q", got)
	}

	handled, err = app.handleNoArgTopLevelCommand("project", nil)
	if handled || err != nil {
		t.Fatalf("expected project without args to fall through, got handled=%v err=%v", handled, err)
	}
}

func TestDelimiterPreservesLiteralFlagLikeDescriptionText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		args       []string
		specPath   string
		wantPieces []string
	}{
		{
			name:       "plan",
			args:       []string{"plan", "--", "--help"},
			specPath:   filepath.Join(".namba", "specs", "SPEC-001", "spec.md"),
			wantPieces: []string{"## Goal", "--help"},
		},
		{
			name:       "harness",
			args:       []string{"harness", "--", "--command"},
			specPath:   filepath.Join(".namba", "specs", "SPEC-001", "spec.md"),
			wantPieces: []string{"## Problem", "--command"},
		},
		{
			name:       "fix plan",
			args:       []string{"fix", "--command", "plan", "--", "--help"},
			specPath:   filepath.Join(".namba", "specs", "SPEC-001", "spec.md"),
			wantPieces: []string{"## Problem", "--help", "Work type: fix"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := canonicalTempDir(t)
			stdout := &bytes.Buffer{}
			app := NewApp(stdout, &bytes.Buffer{})
			if err := app.Run(context.Background(), []string{"init", root, "--yes"}); err != nil {
				t.Fatalf("init failed: %v", err)
			}

			restore := chdirExecution(t, root)
			defer restore()

			if err := app.Run(context.Background(), tc.args); err != nil {
				t.Fatalf("delimiter invocation failed: %v", err)
			}

			spec := mustReadFile(t, filepath.Join(root, tc.specPath))
			for _, want := range tc.wantPieces {
				if !strings.Contains(spec, want) {
					t.Fatalf("expected spec to contain %q, got %q", want, spec)
				}
			}
		})
	}
}

func TestMalformedInvocationsDoNotMutateManagedFiles(t *testing.T) {
	t.Parallel()

	root := canonicalTempDir(t)
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", root, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	writeTestFile(t, filepath.Join(root, "AGENTS.md"), "sentinel-agents\n")
	writeTestFile(t, filepath.Join(root, ".namba", "project", "change-summary.md"), "sentinel-change-summary\n")

	restore := chdirExecution(t, root)
	defer restore()

	cases := []struct {
		name           string
		args           []string
		wantErr        string
		sentinelPath   string
		sentinelBefore string
	}{
		{
			name:           "project extra arg",
			args:           []string{"project", "extra"},
			wantErr:        "project does not accept arguments",
			sentinelPath:   filepath.Join(root, ".namba", "project", "change-summary.md"),
			sentinelBefore: "sentinel-change-summary\n",
		},
		{
			name:           "sync extra arg",
			args:           []string{"sync", "extra"},
			wantErr:        "sync does not accept arguments",
			sentinelPath:   filepath.Join(root, ".namba", "project", "change-summary.md"),
			sentinelBefore: "sentinel-change-summary\n",
		},
		{
			name:           "regen extra arg",
			args:           []string{"regen", "extra"},
			wantErr:        "regen does not accept arguments",
			sentinelPath:   filepath.Join(root, "AGENTS.md"),
			sentinelBefore: "sentinel-agents\n",
		},
		{
			name:    "doctor extra arg",
			args:    []string{"doctor", "extra"},
			wantErr: "doctor does not accept arguments",
		},
		{
			name:    "status extra arg",
			args:    []string{"status", "extra"},
			wantErr: "status does not accept arguments",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.sentinelPath != "" {
				before := mustReadFile(t, tc.sentinelPath)
				if before != tc.sentinelBefore {
					t.Fatalf("unexpected sentinel precondition for %s: got %q", tc.sentinelPath, before)
				}
			}

			if err := app.Run(context.Background(), tc.args); err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected malformed invocation error containing %q, got %v", tc.wantErr, err)
			}

			if tc.sentinelPath != "" {
				after := mustReadFile(t, tc.sentinelPath)
				if after != tc.sentinelBefore {
					t.Fatalf("expected %s to remain unchanged, got %q", tc.sentinelPath, after)
				}
			}
		})
	}
}
