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
