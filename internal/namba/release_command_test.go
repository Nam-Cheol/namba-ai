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

func TestNextReleaseVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tags []string
		bump string
		want string
	}{
		{
			name: "first patch release",
			bump: "patch",
			want: "v0.1.0",
		},
		{
			name: "first major release",
			bump: "major",
			want: "v1.0.0",
		},
		{
			name: "next patch release",
			tags: []string{"v0.1.0", "v0.1.3", "v0.1.2"},
			bump: "patch",
			want: "v0.1.4",
		},
		{
			name: "next minor release",
			tags: []string{"v0.1.9", "v0.2.4"},
			bump: "minor",
			want: "v0.3.0",
		},
		{
			name: "next major release",
			tags: []string{"v0.9.0", "v1.4.2"},
			bump: "major",
			want: "v2.0.0",
		},
		{
			name: "ignore invalid tags",
			tags: []string{"demo", "v0.1.0", "nightly"},
			bump: "patch",
			want: "v0.1.1",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := nextReleaseVersion(tt.tags, tt.bump)
			if err != nil {
				t.Fatalf("nextReleaseVersion returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("nextReleaseVersion(%v, %q) = %q, want %q", tt.tags, tt.bump, got, tt.want)
			}
		})
	}
}

func TestParseReleaseArgsRejectsVersionAndBump(t *testing.T) {
	t.Parallel()

	_, err := parseReleaseArgs([]string{"--version", "v1.2.3", "--bump", "minor"})
	if err == nil {
		t.Fatal("expected conflicting flags error")
	}
}

func TestRunReleaseCreatesTagAndPrintsPushInstructions(t *testing.T) {
	tmp, stdout, app, restore := prepareReleaseProject(t)
	defer restore()

	var commands []string
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		commands = append(commands, name+" "+strings.Join(args, " "))
		if dir != tmp {
			t.Fatalf("expected workdir %s, got %s", tmp, dir)
		}

		switch {
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "main", nil
		case name == "git" && len(args) >= 2 && args[0] == "status" && args[1] == "--porcelain":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "tag" && args[1] == "--list":
			return "v0.1.0\nv0.1.1", nil
		case name == "git" && len(args) == 2 && args[0] == "tag" && args[1] == "v0.1.2":
			return "", nil
		case isShellCommand(name):
			return "ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"release"}); err != nil {
		t.Fatalf("release failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Created release tag v0.1.2") {
		t.Fatalf("expected release tag output, got %q", output)
	}
	if !strings.Contains(output, "git push origin main && git push origin v0.1.2") {
		t.Fatalf("expected push hint, got %q", output)
	}
	if !containsCommand(commands, "git tag v0.1.2") {
		t.Fatalf("expected git tag command, got %v", commands)
	}
}

func TestRunReleasePushesMainAndTag(t *testing.T) {
	tmp, stdout, app, restore := prepareReleaseProject(t)
	defer restore()

	var commands []string
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		commands = append(commands, name+" "+strings.Join(args, " "))
		if dir != tmp {
			t.Fatalf("expected workdir %s, got %s", tmp, dir)
		}

		switch {
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "main", nil
		case name == "git" && len(args) >= 2 && args[0] == "status" && args[1] == "--porcelain":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "tag" && args[1] == "--list":
			return "v0.1.0", nil
		case name == "git" && len(args) == 2 && args[0] == "tag" && args[1] == "v0.1.1":
			return "", nil
		case name == "git" && len(args) == 3 && args[0] == "push" && args[1] == "origin" && args[2] == "main":
			return "", nil
		case name == "git" && len(args) == 3 && args[0] == "push" && args[1] == "origin" && args[2] == "v0.1.1":
			return "", nil
		case isShellCommand(name):
			return "ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"release", "--push"}); err != nil {
		t.Fatalf("release failed: %v", err)
	}

	if !containsCommand(commands, "git push origin main") {
		t.Fatalf("expected main push, got %v", commands)
	}
	if !containsCommand(commands, "git push origin v0.1.1") {
		t.Fatalf("expected tag push, got %v", commands)
	}
	if !strings.Contains(stdout.String(), "Pushed main and v0.1.1 to origin") {
		t.Fatalf("expected push output, got %q", stdout.String())
	}
}

func TestRunReleaseRejectsDirtyTree(t *testing.T) {
	_, _, app, restore := prepareReleaseProject(t)
	defer restore()

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "main", nil
		case name == "git" && len(args) >= 2 && args[0] == "status" && args[1] == "--porcelain":
			return " M README.md", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	err := app.Run(context.Background(), []string{"release"})
	if err == nil {
		t.Fatal("expected dirty tree error")
	}
	if !strings.Contains(err.Error(), "clean working tree") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunReleaseRejectsNonMainBranch(t *testing.T) {
	_, _, app, restore := prepareReleaseProject(t)
	defer restore()

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current" {
			return "feature/release", nil
		}
		t.Fatalf("unexpected command: %s %v", name, args)
		return "", nil
	}

	err := app.Run(context.Background(), []string{"release"})
	if err == nil {
		t.Fatal("expected branch error")
	}
	if !strings.Contains(err.Error(), "main branch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func prepareReleaseProject(t *testing.T) (string, *bytes.Buffer, *App, func()) {
	t.Helper()
	tmp := canonicalTempDir(t)
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}

	restore := chdirExecution(t, tmp)
	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "quality.yaml"), "development_mode: tdd\ntest_command: go test ./...\nlint_command: gofmt -l .\ntypecheck_command: go vet ./...\n")

	app.lookPath = func(name string) (string, error) {
		if name == "git" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}

	return tmp, stdout, app, restore
}

func containsCommand(commands []string, target string) bool {
	for _, command := range commands {
		if command == target {
			return true
		}
	}
	return false
}
