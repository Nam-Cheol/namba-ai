package namba

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNextPlanningSpecIDScansAllActiveWorktrees(t *testing.T) {
	t.Parallel()

	sharedRoot := preparePlanningGitProject(t)
	externalRoot := canonicalTempDir(t)
	prepareAttachedPlanningWorkspace(t, sharedRoot, externalRoot)

	writeTestFile(t, filepath.Join(sharedRoot, ".namba", "specs", "SPEC-002", "spec.md"), "# SPEC-002\n")
	writeTestFile(t, filepath.Join(externalRoot, ".namba", "specs", "SPEC-007", "spec.md"), "# SPEC-007\n")

	specID, err := nextPlanningSpecID([]gitWorktree{
		{Path: sharedRoot, Branch: "main"},
		{Path: externalRoot, Branch: "feature/outside"},
	})
	if err != nil {
		t.Fatalf("nextPlanningSpecID failed: %v", err)
	}
	if specID != "SPEC-008" {
		t.Fatalf("expected SPEC-008, got %s", specID)
	}
}

func TestNextPlanningSpecIDIgnoresPermissionRestrictedWorktrees(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("permission semantics differ on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("permission semantics differ when tests run as root")
	}

	sharedRoot := preparePlanningGitProject(t)
	restrictedRoot := canonicalTempDir(t)
	prepareAttachedPlanningWorkspace(t, sharedRoot, restrictedRoot)

	writeTestFile(t, filepath.Join(sharedRoot, ".namba", "specs", "SPEC-002", "spec.md"), "# SPEC-002\n")
	writeTestFile(t, filepath.Join(restrictedRoot, ".namba", "specs", "SPEC-009", "spec.md"), "# SPEC-009\n")

	restrictedSpecsDir := filepath.Join(restrictedRoot, ".namba", "specs")
	if err := os.Chmod(restrictedSpecsDir, 0); err != nil {
		t.Fatalf("chmod restricted specs dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(restrictedSpecsDir, 0o755)
	})

	specID, err := nextPlanningSpecID([]gitWorktree{
		{Path: sharedRoot, Branch: "main"},
		{Path: restrictedRoot, Branch: "feature/restricted"},
	})
	if err != nil {
		t.Fatalf("nextPlanningSpecID failed: %v", err)
	}
	if specID != "SPEC-003" {
		t.Fatalf("expected SPEC-003 when restricted worktree is skipped, got %s", specID)
	}
}

func TestRunPlanCreatesDedicatedSpecBranchInCurrentWorkspace(t *testing.T) {
	t.Parallel()

	root := preparePlanningGitProject(t)
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})

	restore := chdirExecution(t, root)
	defer restore()

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && strings.Join(args, " ") == "worktree list --porcelain":
			return renderPlanningWorktreeList(
				gitWorktree{Path: root, Branch: "main"},
			), nil
		case name == "git" && strings.Join(args, " ") == "branch --show-current":
			if dir != root {
				t.Fatalf("expected branch lookup in %s, got %s", root, dir)
			}
			return "main", nil
		case name == "git" && strings.Join(args, " ") == "status --porcelain":
			if dir != root {
				t.Fatalf("expected status lookup in %s, got %s", root, dir)
			}
			return "", nil
		case name == "git" && strings.Join(args, " ") == "branch --list spec/SPEC-001-ship-safer-planning-flow":
			return "", nil
		case name == "git" && len(args) == 4 && args[0] == "checkout" && args[1] == "-b":
			if dir != root {
				t.Fatalf("expected checkout in %s, got %s", root, dir)
			}
			if args[2] != "spec/SPEC-001-ship-safer-planning-flow" {
				t.Fatalf("unexpected branch: %v", args)
			}
			if args[3] != "main" {
				t.Fatalf("unexpected checkout args: %v", args)
			}
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v dir=%s", name, args, dir)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"plan", "ship safer planning flow"}); err != nil {
		t.Fatalf("plan failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, ".namba", "specs", "SPEC-001", "spec.md")); err != nil {
		t.Fatalf("expected current workspace spec scaffold, stat err=%v", err)
	}

	got := stdout.String()
	for _, want := range []string{
		"Workspace action: created dedicated SPEC branch in current workspace",
		"Branch: spec/SPEC-001-ship-safer-planning-flow",
		"Workspace: " + root,
		"Next step: continue in the current workspace on branch spec/SPEC-001-ship-safer-planning-flow.",
		"Created SPEC-001",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected output to contain %q, got %q", want, got)
		}
	}
}

func TestRunPlanReusesCurrentMatchingSpecBranch(t *testing.T) {
	t.Parallel()

	root := preparePlanningGitProject(t)
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})

	restore := chdirExecution(t, root)
	defer restore()

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && strings.Join(args, " ") == "worktree list --porcelain":
			return renderPlanningWorktreeList(
				gitWorktree{Path: root, Branch: "spec/SPEC-001-tighten-reuse-behavior"},
			), nil
		case name == "git" && strings.Join(args, " ") == "branch --show-current":
			if dir != root {
				t.Fatalf("expected branch lookup in %s, got %s", root, dir)
			}
			return "spec/SPEC-001-tighten-reuse-behavior", nil
		case name == "git" && strings.Join(args, " ") == "status --porcelain":
			if dir != root {
				t.Fatalf("expected status lookup in %s, got %s", root, dir)
			}
			return "", nil
		case name == "git" && len(args) > 1 && args[0] == "checkout":
			t.Fatal("did not expect branch checkout when already on the target SPEC branch")
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v dir=%s", name, args, dir)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"plan", "tighten reuse behavior"}); err != nil {
		t.Fatalf("plan failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, ".namba", "specs", "SPEC-001", "spec.md")); err != nil {
		t.Fatalf("expected scaffold in current workspace, stat err=%v", err)
	}
	if got := stdout.String(); !strings.Contains(got, "Workspace action: reused current SPEC branch") {
		t.Fatalf("expected reuse summary, got %q", got)
	}
}

func TestRunPlanRefusesDirtyWorkspaceBeforeBranchSwitch(t *testing.T) {
	t.Parallel()

	root := preparePlanningGitProject(t)
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	restore := chdirExecution(t, root)
	defer restore()

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && strings.Join(args, " ") == "worktree list --porcelain":
			return renderPlanningWorktreeList(
				gitWorktree{Path: root, Branch: "feature/other-work"},
			), nil
		case name == "git" && strings.Join(args, " ") == "branch --show-current":
			return "feature/other-work", nil
		case name == "git" && strings.Join(args, " ") == "status --porcelain":
			return " M README.md", nil
		case name == "git" && len(args) > 1 && args[0] == "checkout":
			t.Fatal("did not expect branch switch for dirty workspace")
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v dir=%s", name, args, dir)
			return "", nil
		}
	}

	err := app.Run(context.Background(), []string{"plan", "unsafe dirty context"})
	if err == nil {
		t.Fatal("expected plan to refuse dirty workspace")
	}
	for _, want := range []string{
		"Workspace action: refused due to dirty workspace",
		"Branch: feature/other-work",
		currentWorkspacePlanningFlag,
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to contain %q, got %v", want, err)
		}
	}
}

func TestRunPlanCurrentWorkspaceOverrideAllowsDirtyWorkspace(t *testing.T) {
	t.Parallel()

	sharedRoot := preparePlanningGitProject(t)
	currentRoot := canonicalTempDir(t)
	prepareAttachedPlanningWorkspace(t, sharedRoot, currentRoot)

	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	restore := chdirExecution(t, currentRoot)
	defer restore()

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && strings.Join(args, " ") == "worktree list --porcelain":
			return renderPlanningWorktreeList(
				gitWorktree{Path: sharedRoot, Branch: "main"},
				gitWorktree{Path: currentRoot, Branch: "feature/other-work"},
			), nil
		case name == "git" && strings.Join(args, " ") == "branch --show-current":
			return "feature/other-work", nil
		case name == "git" && strings.Join(args, " ") == "status --porcelain":
			return " M README.md", nil
		case name == "git" && len(args) > 1 && args[0] == "worktree" && args[1] == "add":
			t.Fatal("did not expect worktree creation when override is explicit")
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v dir=%s", name, args, dir)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"plan", currentWorkspacePlanningFlag, "local escape hatch"}); err != nil {
		t.Fatalf("plan override failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(currentRoot, ".namba", "specs", "SPEC-001", "spec.md")); err != nil {
		t.Fatalf("expected scaffold in overridden workspace, stat err=%v", err)
	}
	if got := stdout.String(); !strings.Contains(got, "Workspace action: explicit current-workspace override") {
		t.Fatalf("expected override summary, got %q", got)
	}
}

func TestRunPlanPreservesCreatedBranchWhenScaffoldFails(t *testing.T) {
	t.Parallel()

	root := preparePlanningGitProject(t)
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	restore := chdirExecution(t, root)
	defer restore()

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && strings.Join(args, " ") == "worktree list --porcelain":
			return renderPlanningWorktreeList(
				gitWorktree{Path: root, Branch: "main"},
			), nil
		case name == "git" && strings.Join(args, " ") == "branch --show-current":
			return "main", nil
		case name == "git" && strings.Join(args, " ") == "status --porcelain":
			return "", nil
		case name == "git" && strings.Join(args, " ") == "branch --list spec/SPEC-001-fail-after-add":
			return "", nil
		case name == "git" && len(args) == 4 && args[0] == "checkout" && args[1] == "-b":
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v dir=%s", name, args, dir)
			return "", nil
		}
	}
	app.mkdirAll = func(path string, perm os.FileMode) error {
		if strings.Contains(path, filepath.Join(specsDir, "SPEC-001")) {
			return errors.New("mkdir denied")
		}
		return os.MkdirAll(path, perm)
	}

	err := app.Run(context.Background(), []string{"plan", "fail after add"})
	if err == nil {
		t.Fatal("expected plan to fail after branch creation")
	}
	for _, want := range []string{
		"Workspace action: created dedicated SPEC branch in current workspace",
		"scaffolding failed after creating the SPEC branch",
		"spec/SPEC-001-fail-after-add",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to contain %q, got %v", want, err)
		}
	}
}

func preparePlanningGitProject(t *testing.T) string {
	t.Helper()

	root := canonicalTempDir(t)
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", root, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	return root
}

func prepareAttachedPlanningWorkspace(t *testing.T, sharedRoot, targetRoot string) {
	t.Helper()

	if err := copyDirContents(filepath.Join(sharedRoot, ".namba", "config", "sections"), filepath.Join(targetRoot, ".namba", "config", "sections")); err != nil {
		t.Fatalf("copy config sections: %v", err)
	}
	writeTestFile(t, filepath.Join(targetRoot, ".namba", "specs", ".gitkeep"), "")
	writeTestFile(t, filepath.Join(targetRoot, ".git"), "gitdir: /tmp/fake\n")
}

func renderPlanningWorktreeList(worktrees ...gitWorktree) string {
	lines := make([]string, 0, len(worktrees)*4)
	for _, worktree := range worktrees {
		lines = append(lines,
			"worktree "+filepath.Clean(worktree.Path),
			"HEAD deadbeef",
			"branch refs/heads/"+worktree.Branch,
			"",
		)
	}
	return strings.Join(lines, "\n")
}
