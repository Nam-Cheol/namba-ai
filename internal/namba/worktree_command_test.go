package namba

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveWorktreeSubcommandDefinitions(t *testing.T) {
	t.Parallel()

	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	for _, name := range []string{"new", "list", "remove", "clean"} {
		definition, ok := app.resolveWorktreeSubcommand(name)
		if !ok || definition.Run == nil {
			t.Fatalf("expected worktree subcommand %q to resolve with runner, got %#v ok=%v", name, definition, ok)
		}
		if strings.TrimSpace(definition.UsageSummary) == "" {
			t.Fatalf("expected worktree subcommand %q to provide usage summary, got %#v", name, definition)
		}
	}

	if _, ok := app.resolveWorktreeSubcommand("missing"); ok {
		t.Fatalf("did not expect unknown worktree subcommand to resolve")
	}
}

func TestWorktreeUsageTextMatchesSubcommandUsageSummaries(t *testing.T) {
	t.Parallel()

	got := worktreeUsageText()
	if !strings.HasPrefix(got, "namba worktree\n\nUsage:\n") {
		t.Fatalf("unexpected worktree usage header: %q", got)
	}
	if !strings.Contains(got, "\n\nBehavior:\n  Manage Namba-owned git worktrees under .namba/worktrees.\n") {
		t.Fatalf("expected worktree behavior block, got %q", got)
	}

	lastIndex := -1
	for _, definition := range worktreeSubcommandDefinitions() {
		index := strings.Index(got, definition.UsageSummary)
		if index < 0 {
			t.Fatalf("expected worktree usage to contain %q, got %q", definition.UsageSummary, got)
		}
		if index <= lastIndex {
			t.Fatalf("expected usage summary %q to appear after previous summary in %q", definition.UsageSummary, got)
		}
		lastIndex = index
	}
}

func TestRunWorktreeSubcommandsRouteThroughDefinitions(t *testing.T) {
	t.Parallel()

	root := canonicalTempDir(t)
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", root, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, root)
	defer restore()

	var commands []string
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		commands = append(commands, name+" "+strings.Join(args, " "))
		if dir != root {
			t.Fatalf("expected workdir %s, got %s", root, dir)
		}

		switch {
		case name == "git" && len(args) == 6 && args[0] == "worktree" && args[1] == "add" && args[2] == "-b":
			return "", nil
		case name == "git" && len(args) == 3 && args[0] == "worktree" && args[1] == "list" && args[2] == "--porcelain":
			return "worktree " + filepath.Clean(root), nil
		case name == "git" && len(args) == 4 && args[0] == "worktree" && args[1] == "remove" && args[2] == "--force":
			return "", nil
		case name == "git" && len(args) == 2 && args[0] == "worktree" && args[1] == "prune":
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	tests := []struct {
		name        string
		args        []string
		wantCommand string
		wantOutput  string
	}{
		{
			name:        "new",
			args:        []string{"worktree", "new", "scratch"},
			wantCommand: "git worktree add -b namba/scratch",
			wantOutput:  "Created worktree " + filepath.Join(root, worktreesDir, "scratch"),
		},
		{
			name:        "list",
			args:        []string{"worktree", "list"},
			wantCommand: "git worktree list --porcelain",
			wantOutput:  "worktree " + filepath.Clean(root),
		},
		{
			name:        "remove",
			args:        []string{"worktree", "remove", "scratch"},
			wantCommand: "git worktree remove --force " + filepath.Join(root, worktreesDir, "scratch"),
			wantOutput:  "Removed worktree " + filepath.Join(root, worktreesDir, "scratch"),
		},
		{
			name:        "clean",
			args:        []string{"worktree", "clean"},
			wantCommand: "git worktree prune",
			wantOutput:  "Pruned worktrees.",
		},
	}

	for _, tc := range tests {
		stdout.Reset()
		before := len(commands)

		if err := app.Run(context.Background(), tc.args); err != nil {
			t.Fatalf("%s worktree command failed: %v", tc.name, err)
		}

		gotOutput := stdout.String()
		if !strings.Contains(gotOutput, tc.wantOutput) {
			t.Fatalf("%s expected output to contain %q, got %q", tc.name, tc.wantOutput, gotOutput)
		}

		newCommands := append([]string{}, commands[before:]...)
		if !hasCommandContaining(newCommands, tc.wantCommand) {
			t.Fatalf("%s expected command containing %q, got %v", tc.name, tc.wantCommand, newCommands)
		}
	}
}

func TestRunWorktreeUnknownSubcommandUsesUsageError(t *testing.T) {
	t.Parallel()

	root := canonicalTempDir(t)
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", root, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, root)
	defer restore()

	app.runCmd = func(context.Context, string, []string, string) (string, error) {
		t.Fatal("did not expect git command for unknown worktree subcommand")
		return "", nil
	}

	err := app.Run(context.Background(), []string{"worktree", "unknown"})
	if err == nil || !strings.Contains(err.Error(), "unknown worktree subcommand \"unknown\"") {
		t.Fatalf("expected unknown worktree subcommand error, got %v", err)
	}
	for _, want := range []string{"namba worktree new <name>", "namba worktree clean"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected unknown worktree subcommand error to include %q, got %v", want, err)
		}
	}
}

func TestRunWorktreeNewRejectsExtraArguments(t *testing.T) {
	t.Parallel()

	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	err := app.runWorktreeNewSubcommand(context.Background(), "/repo", []string{"one", "two"})
	if err == nil || !strings.Contains(err.Error(), "worktree new accepts exactly one name") {
		t.Fatalf("expected arity error, got %v", err)
	}

	err = app.runWorktreeRemoveSubcommand(context.Background(), "/repo", nil)
	if err == nil || !strings.Contains(err.Error(), "worktree remove requires a name") {
		t.Fatalf("expected missing-name error, got %v", err)
	}

	err = app.runWorktreeListSubcommand(context.Background(), "/repo", []string{"extra"})
	if err == nil || !strings.Contains(err.Error(), "worktree list does not accept arguments") {
		t.Fatalf("expected list arity error, got %v", err)
	}

	err = app.runWorktreeCleanSubcommand(context.Background(), "/repo", []string{"extra"})
	if err == nil || !strings.Contains(err.Error(), "worktree clean does not accept arguments") {
		t.Fatalf("expected clean arity error, got %v", err)
	}
}

func TestRunWorktreeSubcommandPropagatesGitErrors(t *testing.T) {
	t.Parallel()

	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	app.runCmd = func(context.Context, string, []string, string) (string, error) {
		return "", errors.New("git failed")
	}

	err := app.runWorktreeCleanSubcommand(context.Background(), "/repo", nil)
	if err == nil || !strings.Contains(err.Error(), "git failed") {
		t.Fatalf("expected git error, got %v", err)
	}
}
