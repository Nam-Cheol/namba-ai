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

func TestPreviousReleaseTag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tags    []string
		version string
		want    string
	}{
		{
			name:    "finds prior semver tag",
			tags:    []string{"v0.1.0", "v0.1.2", "v1.0.0"},
			version: "v1.0.1",
			want:    "v1.0.0",
		},
		{
			name:    "returns empty when release has no prior semver tag",
			tags:    []string{"demo", "v0.1.0"},
			version: "v0.1.0",
			want:    "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := previousReleaseTag(tt.tags, tt.version)
			if err != nil {
				t.Fatalf("previousReleaseTag returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("previousReleaseTag(%v, %q) = %q, want %q", tt.tags, tt.version, got, tt.want)
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

func TestParseReleaseCommitsSkipsPrepCommitAndRenderNotes(t *testing.T) {
	t.Parallel()

	output := strings.Join([]string{
		"1111111\x001111111\x00feat: add dashboard filters (#12)\x00SPEC-039\nFixes #12\n",
		"2222222\x002222222\x00fix: correct release handoff\x00PR #45\n",
		"3333333\x003333333\x00docs: update workflow guide\x00",
		"4444444\x004444444\x00chore: refresh dependencies\x00",
		"5555555\x005555555\x00chore(release): prepare release notes for v0.1.2 [namba-release-notes]\x00",
	}, "\x1e") + "\x1e"

	commits, err := parseReleaseCommits(output)
	if err != nil {
		t.Fatalf("parseReleaseCommits returned error: %v", err)
	}
	if got, want := len(commits), 4; got != want {
		t.Fatalf("parseReleaseCommits returned %d commits, want %d", got, want)
	}

	notes := renderReleaseNotes("v0.1.2", "v0.1.1", commits)
	for _, want := range []string{"# v0.1.2 릴리즈 노트", "v0.1.1 이후 변경 사항입니다.", "## 사용자에게 보이는 변경", "## 수정", "## 문서 및 워크플로", "## 내부 정비", "SPEC-039", "#12", "PR #45", "1111111", "4444444"} {
		if !strings.Contains(notes, want) {
			t.Fatalf("renderReleaseNotes missing %q: %q", want, notes)
		}
	}
	if strings.Contains(notes, "namba-release-notes") {
		t.Fatalf("renderReleaseNotes should exclude release prep commit: %q", notes)
	}
}

func TestRunReleaseCreatesNotesCommitAndPrintsPushInstructions(t *testing.T) {
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
		case name == "git" && len(args) >= 4 && args[0] == "log" && args[1] == "--no-merges" && args[2] == "--reverse":
			if !strings.Contains(strings.Join(args, " "), "v0.1.1..HEAD") {
				t.Fatalf("expected release log range to start from previous tag, got %v", args)
			}
			return strings.Join([]string{
				"aaaaaaa\x00aaaaaaa\x00feat: add dashboard filters (#12)\x00SPEC-039\nFixes #12\n",
				"bbbbbbb\x00bbbbbbb\x00fix: correct release handoff\x00PR #45\n",
				"ccccccc\x00ccccccc\x00docs: update workflow guide\x00",
				"ddddddd\x00ddddddd\x00chore: refresh dependencies\x00",
			}, "\x1e") + "\x1e", nil
		case name == "git" && len(args) == 2 && args[0] == "show":
			return "", errors.New("fixture has no spec artifact")
		case name == "git" && len(args) == 2 && args[0] == "add" && args[1] == ".namba/releases/v0.1.2.md":
			return "", nil
		case name == "git" && len(args) >= 5 && args[0] == "commit" && args[1] == "-m":
			return "", nil
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
	if !strings.Contains(output, "Wrote release notes to .namba/releases/v0.1.2.md") {
		t.Fatalf("expected release notes output, got %q", output)
	}
	if !strings.Contains(output, "git push origin main && git push origin v0.1.2") {
		t.Fatalf("expected push hint, got %q", output)
	}
	if !containsCommand(commands, "git add .namba/releases/v0.1.2.md") {
		t.Fatalf("expected release notes staging command, got %v", commands)
	}
	if !containsCommand(commands, "git commit -m chore(release): prepare release notes for v0.1.2 [namba-release-notes] -- .namba/releases/v0.1.2.md") {
		t.Fatalf("expected release notes commit command, got %v", commands)
	}
	if !containsCommand(commands, "git tag v0.1.2") {
		t.Fatalf("expected git tag command, got %v", commands)
	}
	if indexOfCommand(commands, "git commit -m chore(release): prepare release notes for v0.1.2 [namba-release-notes] -- .namba/releases/v0.1.2.md") > indexOfCommand(commands, "git tag v0.1.2") {
		t.Fatalf("expected release notes commit before tag, got %v", commands)
	}

	notes := mustReadFile(t, filepath.Join(tmp, ".namba", "releases", "v0.1.2.md"))
	for _, want := range []string{"# v0.1.2 릴리즈 노트", "## 사용자에게 보이는 변경", "## 수정", "## 문서 및 워크플로", "## 내부 정비"} {
		if !strings.Contains(notes, want) {
			t.Fatalf("release notes missing %q: %q", want, notes)
		}
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
		case name == "git" && len(args) >= 4 && args[0] == "log" && args[1] == "--no-merges" && args[2] == "--reverse":
			return strings.Join([]string{
				"aaaaaaa\x00aaaaaaa\x00feat: add release notes support\x00SPEC-039\n",
			}, "\x1e") + "\x1e", nil
		case name == "git" && len(args) == 2 && args[0] == "show":
			return "", errors.New("fixture has no spec artifact")
		case name == "git" && len(args) == 2 && args[0] == "add" && args[1] == ".namba/releases/v0.1.1.md":
			return "", nil
		case name == "git" && len(args) >= 5 && args[0] == "commit" && args[1] == "-m":
			return "", nil
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
	if !containsCommand(commands, "git commit -m chore(release): prepare release notes for v0.1.1 [namba-release-notes] -- .namba/releases/v0.1.1.md") {
		t.Fatalf("expected release notes commit, got %v", commands)
	}
	if indexOfCommand(commands, "git commit -m chore(release): prepare release notes for v0.1.1 [namba-release-notes] -- .namba/releases/v0.1.1.md") > indexOfCommand(commands, "git tag v0.1.1") {
		t.Fatalf("expected release notes commit before tag, got %v", commands)
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

func indexOfCommand(commands []string, target string) int {
	for i, command := range commands {
		if command == target {
			return i
		}
	}
	return -1
}
