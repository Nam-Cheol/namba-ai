package namba

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParsePRArgs(t *testing.T) {
	t.Parallel()

	opts, err := parsePRArgs([]string{"review", "title", "--remote", "upstream", "--no-sync", "--no-validate"})
	if err != nil {
		t.Fatalf("parsePRArgs returned error: %v", err)
	}
	if opts.Title != "review title" || opts.Remote != "upstream" || !opts.SkipSync || !opts.SkipValidation {
		t.Fatalf("unexpected pr options: %+v", opts)
	}
}

func TestParseLandArgs(t *testing.T) {
	t.Parallel()

	opts, err := parseLandArgs([]string{"17", "--wait", "--remote", "upstream"})
	if err != nil {
		t.Fatalf("parseLandArgs returned error: %v", err)
	}
	if opts.Number != 17 || opts.Remote != "upstream" || !opts.Wait {
		t.Fatalf("unexpected land options: %+v", opts)
	}
}

func TestRunPRCreatesPullRequestAndAddsReviewComment(t *testing.T) {
	tmp, stdout, app, restore := preparePRLandProject(t)
	defer restore()

	var commands []string
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		commands = append(commands, name+" "+strings.Join(args, " "))
		if dir != tmp {
			t.Fatalf("expected workdir %s, got %s", tmp, dir)
		}

		switch {
		case name == "gh" && len(args) == 2 && args[0] == "auth" && args[1] == "status":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "feature/login-audit", nil
		case isShellCommand(name):
			return "ok", nil
		case name == "git" && len(args) >= 2 && args[0] == "status" && args[1] == "--porcelain":
			return " M README.md", nil
		case name == "git" && len(args) == 2 && args[0] == "add" && args[1] == "-A":
			return "", nil
		case name == "git" && len(args) == 3 && args[0] == "commit" && args[1] == "-m":
			if args[2] != "Add login audit logs" {
				t.Fatalf("unexpected commit title: %v", args)
			}
			return "", nil
		case name == "git" && len(args) == 4 && args[0] == "push" && args[1] == "--set-upstream":
			return "", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list":
			return "[]", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "create":
			mustContainArgs(t, args, []string{"--base", "main", "--head", "feature/login-audit", "--title", "Add login audit logs"})
			if indexOfArg(args, "--body") == -1 {
				t.Fatalf("expected --body in pr create args, got %v", args)
			}
			return "https://github.com/example/repo/pull/17", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "view" && args[2] == "feature/login-audit":
			return mustMarshalJSON(t, githubPullRequest{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Add login audit logs", HeadRefName: "feature/login-audit", BaseRefName: "main"}), nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "view" && args[2] == "17":
			return mustMarshalJSON(t, githubPullRequest{Comments: []githubPRComment{}}), nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "comment":
			mustContainArgs(t, args, []string{"--body", buildReviewRequestCommentBody("@codex review")})
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"pr", "Add", "login", "audit", "logs"}); err != nil {
		t.Fatalf("pr failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "Prepared PR #17") {
		t.Fatalf("expected PR output, got %q", stdout.String())
	}
	if !hasCommandContaining(commands, "gh pr create") {
		t.Fatalf("expected PR creation command, got %v", commands)
	}
	if !hasCommandContaining(commands, "gh pr comment 17 --body") {
		t.Fatalf("expected review comment command, got %v", commands)
	}
}

func TestRunPRAllowsSingleDashHelpLikeTitleWord(t *testing.T) {
	tmp, stdout, app, restore := preparePRLandProject(t)
	defer restore()

	var commands []string
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		commands = append(commands, name+" "+strings.Join(args, " "))
		if dir != tmp {
			t.Fatalf("expected workdir %s, got %s", tmp, dir)
		}

		switch {
		case name == "gh" && len(args) == 2 && args[0] == "auth" && args[1] == "status":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "feature/login-audit", nil
		case name == "git" && len(args) >= 2 && args[0] == "status" && args[1] == "--porcelain":
			return "", nil
		case name == "git" && len(args) == 4 && args[0] == "push" && args[1] == "--set-upstream":
			return "", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list":
			return "[]", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "create":
			mustContainArgs(t, args, []string{"--title", "fix -h parsing"})
			return "https://github.com/example/repo/pull/17", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "view" && args[2] == "feature/login-audit":
			return mustMarshalJSON(t, githubPullRequest{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "fix -h parsing", HeadRefName: "feature/login-audit", BaseRefName: "main"}), nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "view" && args[2] == "17":
			return mustMarshalJSON(t, githubPullRequest{Comments: []githubPRComment{}}), nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "comment":
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"pr", "fix", "-h", "parsing", "--no-sync", "--no-validate"}); err != nil {
		t.Fatalf("pr failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "Prepared PR #17") {
		t.Fatalf("expected PR output, got %q", stdout.String())
	}
	if !hasCommandContaining(commands, "gh pr create") {
		t.Fatalf("expected PR creation command, got %v", commands)
	}
}

func TestRunPRIncludesLatestReviewReadinessInBody(t *testing.T) {
	tmp, _, app, restore := preparePRLandProject(t)
	defer restore()

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && len(args) == 3 && args[0] == "worktree" && args[1] == "list" && args[2] == "--porcelain":
			return renderPlanningWorktreeList(gitWorktree{Path: tmp, Branch: "main"}), nil
		case name == "git" && len(args) == 3 && args[0] == "for-each-ref" && args[1] == "--format=%(refname:short)" && args[2] == "refs/heads":
			return "main", nil
		case name == "git" && len(args) == 6 && args[0] == "ls-tree" && args[1] == "-r" && args[2] == "--name-only" && args[3] == "--full-tree" && args[4] == "main" && args[5] == ".namba/specs":
			return ".namba/specs/.gitkeep", nil
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "main", nil
		case name == "git" && len(args) >= 2 && args[0] == "status" && args[1] == "--porcelain":
			return "", nil
		default:
			t.Fatalf("unexpected command during plan setup: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"plan", currentWorkspacePlanningFlag, "add", "review", "readiness"}); err != nil {
		t.Fatalf("plan failed: %v", err)
	}

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if dir != tmp {
			t.Fatalf("expected workdir %s, got %s", tmp, dir)
		}

		switch {
		case name == "gh" && len(args) == 2 && args[0] == "auth" && args[1] == "status":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "feature/login-audit", nil
		case isShellCommand(name):
			return "ok", nil
		case name == "git" && len(args) >= 2 && args[0] == "status" && args[1] == "--porcelain":
			return "", nil
		case name == "git" && len(args) == 4 && args[0] == "push" && args[1] == "--set-upstream":
			return "", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list":
			return "[]", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "create":
			bodyIndex := indexOfArg(args, "--body")
			if bodyIndex == -1 || bodyIndex+1 >= len(args) {
				t.Fatalf("expected --body argument, got %v", args)
			}
			if !strings.Contains(args[bodyIndex+1], ".namba/project/change-summary.md") || !strings.Contains(args[bodyIndex+1], ".namba/specs/SPEC-001/reviews/readiness.md") {
				t.Fatalf("expected PR body to reference latest readiness artifact, got %q", args[bodyIndex+1])
			}
			return "https://github.com/example/repo/pull/17", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "view" && args[2] == "feature/login-audit":
			return mustMarshalJSON(t, githubPullRequest{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Review readiness", HeadRefName: "feature/login-audit", BaseRefName: "main"}), nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "view" && args[2] == "17":
			return mustMarshalJSON(t, githubPullRequest{Comments: []githubPRComment{}}), nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "comment":
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"pr", "Review", "readiness"}); err != nil {
		t.Fatalf("pr failed: %v", err)
	}
}

func TestRunPRReusesExistingPullRequestWithoutDuplicateReviewComment(t *testing.T) {
	tmp, stdout, app, restore := preparePRLandProject(t)
	defer restore()

	var commands []string
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		commands = append(commands, name+" "+strings.Join(args, " "))
		if dir != tmp {
			t.Fatalf("expected workdir %s, got %s", tmp, dir)
		}

		switch {
		case name == "gh" && len(args) == 2 && args[0] == "auth" && args[1] == "status":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "feature/login-audit", nil
		case isShellCommand(name):
			return "ok", nil
		case name == "git" && len(args) >= 2 && args[0] == "status" && args[1] == "--porcelain":
			return "", nil
		case name == "git" && len(args) == 4 && args[0] == "push" && args[1] == "--set-upstream":
			return "", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list":
			return mustMarshalJSON(t, []githubPullRequest{{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Existing PR", HeadRefName: "feature/login-audit", BaseRefName: "main"}}), nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "view" && args[2] == "17":
			return mustMarshalJSON(t, githubPullRequest{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Existing PR", HeadRefName: "feature/login-audit", BaseRefName: "main", Comments: []githubPRComment{{Body: "@codex review"}}}), nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"pr", "Reuse", "existing", "pr"}); err != nil {
		t.Fatalf("pr failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "Reused PR #17") {
		t.Fatalf("expected reuse output, got %q", stdout.String())
	}
	if hasCommandContaining(commands, "gh pr create") {
		t.Fatalf("expected no PR creation, got %v", commands)
	}
	if hasCommandContaining(commands, "gh pr comment") {
		t.Fatalf("expected no duplicate review comment, got %v", commands)
	}
}

func TestRunPRSkipsReviewCommentWhenMarkerCommentAlreadyExists(t *testing.T) {
	tmp, stdout, app, restore := preparePRLandProject(t)
	defer restore()

	var commands []string
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		commands = append(commands, name+" "+strings.Join(args, " "))
		if dir != tmp {
			t.Fatalf("expected workdir %s, got %s", tmp, dir)
		}

		switch {
		case name == "gh" && len(args) == 2 && args[0] == "auth" && args[1] == "status":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "feature/login-audit", nil
		case isShellCommand(name):
			return "ok", nil
		case name == "git" && len(args) >= 2 && args[0] == "status" && args[1] == "--porcelain":
			return "", nil
		case name == "git" && len(args) == 4 && args[0] == "push" && args[1] == "--set-upstream":
			return "", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list":
			return "[]", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "create":
			return "https://github.com/example/repo/pull/17", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "view" && args[2] == "feature/login-audit":
			return mustMarshalJSON(t, githubPullRequest{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Existing PR", HeadRefName: "feature/login-audit", BaseRefName: "main"}), nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "view" && args[2] == "17":
			return mustMarshalJSON(t, githubPullRequest{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Existing PR", HeadRefName: "feature/login-audit", BaseRefName: "main", Comments: []githubPRComment{{Body: buildReviewRequestCommentBody("@codex review")}}}), nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"pr", "Reuse", "existing", "pr"}); err != nil {
		t.Fatalf("pr failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "Prepared PR #17") {
		t.Fatalf("expected prepared output, got %q", stdout.String())
	}
	if hasCommandContaining(commands, "gh pr comment") {
		t.Fatalf("expected no duplicate marker review comment, got %v", commands)
	}
}

func TestRunPRRejectsBaseBranch(t *testing.T) {
	tmp, _, app, restore := preparePRLandProject(t)
	defer restore()

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if dir != tmp {
			t.Fatalf("expected workdir %s, got %s", tmp, dir)
		}
		switch {
		case name == "gh" && len(args) == 2 && args[0] == "auth" && args[1] == "status":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "main", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	err := app.Run(context.Background(), []string{"pr", "Review"})
	if err == nil {
		t.Fatal("expected base branch error")
	}
	if !strings.Contains(err.Error(), "work branch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPRRequiresGitHubAuth(t *testing.T) {
	tmp, _, app, restore := preparePRLandProject(t)
	defer restore()

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if dir != tmp {
			t.Fatalf("expected workdir %s, got %s", tmp, dir)
		}
		if name == "gh" && len(args) == 2 && args[0] == "auth" && args[1] == "status" {
			return "", errors.New("not logged in")
		}
		t.Fatalf("unexpected command: %s %v", name, args)
		return "", nil
	}

	err := app.Run(context.Background(), []string{"pr", "Review"})
	if err == nil {
		t.Fatal("expected auth error")
	}
	if !strings.Contains(err.Error(), "gh auth login") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPRRequiresGitRepository(t *testing.T) {
	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	err := app.Run(context.Background(), []string{"pr", "Review"})
	if err == nil {
		t.Fatal("expected git repository error")
	}
	if !strings.Contains(err.Error(), "git repository") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLandRejectsPendingChecksWithoutWait(t *testing.T) {
	tmp, _, app, restore := preparePRLandProject(t)
	defer restore()

	var commands []string
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		commands = append(commands, name+" "+strings.Join(args, " "))
		if dir != tmp {
			t.Fatalf("expected workdir %s, got %s", tmp, dir)
		}

		switch {
		case name == "gh" && len(args) == 2 && args[0] == "auth" && args[1] == "status":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "feature/login-audit", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list":
			return mustMarshalJSON(t, []githubPullRequest{{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Existing PR", HeadRefName: "feature/login-audit", BaseRefName: "main"}}), nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "view":
			return mustMarshalJSON(t, githubPullRequest{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Existing PR", HeadRefName: "feature/login-audit", BaseRefName: "main", ReviewDecision: "APPROVED", MergeStateStatus: "CLEAN", StatusChecks: []githubStatusCheck{{Name: "ci", Status: "IN_PROGRESS"}}}), nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	err := app.Run(context.Background(), []string{"land"})
	if err == nil {
		t.Fatal("expected pending-check error")
	}
	if !strings.Contains(err.Error(), "pending checks") {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasCommandContaining(commands, "gh pr merge") {
		t.Fatalf("expected no merge command, got %v", commands)
	}
}

func TestRunLandRejectsChangesRequested(t *testing.T) {
	tmp, _, app, restore := preparePRLandProject(t)
	defer restore()

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if dir != tmp {
			t.Fatalf("expected workdir %s, got %s", tmp, dir)
		}
		switch {
		case name == "gh" && len(args) == 2 && args[0] == "auth" && args[1] == "status":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "feature/login-audit", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list":
			return mustMarshalJSON(t, []githubPullRequest{{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Existing PR", HeadRefName: "feature/login-audit", BaseRefName: "main"}}), nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "view":
			return mustMarshalJSON(t, githubPullRequest{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Existing PR", HeadRefName: "feature/login-audit", BaseRefName: "main", ReviewDecision: "CHANGES_REQUESTED", MergeStateStatus: "BLOCKED"}), nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	err := app.Run(context.Background(), []string{"land"})
	if err == nil {
		t.Fatal("expected review-blocked error")
	}
	if !strings.Contains(err.Error(), "requested changes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLandMergesAndUpdatesLocalMain(t *testing.T) {
	tmp, stdout, app, restore := preparePRLandProject(t)
	defer restore()

	var commands []string
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		commands = append(commands, name+" "+strings.Join(args, " "))
		if dir != tmp {
			t.Fatalf("expected workdir %s, got %s", tmp, dir)
		}

		switch {
		case name == "gh" && len(args) == 2 && args[0] == "auth" && args[1] == "status":
			return "", nil
		case name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "feature/login-audit", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list":
			return mustMarshalJSON(t, []githubPullRequest{{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Existing PR", HeadRefName: "feature/login-audit", BaseRefName: "main"}}), nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "checks":
			return "", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "view":
			return mustMarshalJSON(t, githubPullRequest{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Existing PR", HeadRefName: "feature/login-audit", BaseRefName: "main", ReviewDecision: "APPROVED", MergeStateStatus: "CLEAN"}), nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "merge":
			return "", nil
		case name == "git" && len(args) == 3 && args[0] == "fetch" && args[1] == "origin" && args[2] == "main":
			return "", nil
		case name == "git" && len(args) == 3 && args[0] == "worktree" && args[1] == "list" && args[2] == "--porcelain":
			return "", nil
		case name == "git" && len(args) == 4 && args[0] == "branch" && args[1] == "-f" && args[2] == "main" && args[3] == "origin/main":
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"land", "--wait"}); err != nil {
		t.Fatalf("land failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "Merged PR #17 and updated local main") {
		t.Fatalf("expected land output, got %q", stdout.String())
	}
	if !hasCommandContaining(commands, "gh pr merge 17 --merge") {
		t.Fatalf("expected merge command, got %v", commands)
	}
	if !hasCommandContaining(commands, "git worktree list --porcelain") {
		t.Fatalf("expected worktree inspection, got %v", commands)
	}
	if !hasCommandContaining(commands, "git branch -f main origin/main") {
		t.Fatalf("expected local main ref update, got %v", commands)
	}
}

func TestRunLandUpdatesCheckedOutBaseBranchThroughWorktree(t *testing.T) {
	tmp, stdout, app, restore := preparePRLandProject(t)
	defer restore()

	baseWorktree := filepath.Join(tmp, "..", "main-worktree")
	worktreeList := strings.Join([]string{
		"worktree " + filepath.Clean(tmp),
		"HEAD 1111111111111111111111111111111111111111",
		"branch refs/heads/feature/login-audit",
		"",
		"worktree " + filepath.Clean(baseWorktree),
		"HEAD 2222222222222222222222222222222222222222",
		"branch refs/heads/main",
		"",
	}, "\n")

	var commands []string
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		commands = append(commands, name+" "+strings.Join(args, " "))

		switch {
		case dir == tmp && name == "gh" && len(args) == 2 && args[0] == "auth" && args[1] == "status":
			return "", nil
		case dir == tmp && name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "feature/login-audit", nil
		case dir == tmp && name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list":
			return mustMarshalJSON(t, []githubPullRequest{{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Existing PR", HeadRefName: "feature/login-audit", BaseRefName: "main"}}), nil
		case dir == tmp && name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "view":
			return mustMarshalJSON(t, githubPullRequest{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Existing PR", HeadRefName: "feature/login-audit", BaseRefName: "main", ReviewDecision: "APPROVED", MergeStateStatus: "CLEAN"}), nil
		case dir == tmp && name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "merge":
			return "", nil
		case dir == tmp && name == "git" && len(args) == 3 && args[0] == "fetch" && args[1] == "origin" && args[2] == "main":
			return "", nil
		case dir == tmp && name == "git" && len(args) == 3 && args[0] == "worktree" && args[1] == "list" && args[2] == "--porcelain":
			return worktreeList, nil
		case dir == filepath.Clean(baseWorktree) && name == "git" && len(args) >= 2 && args[0] == "status" && args[1] == "--porcelain":
			return "", nil
		case dir == filepath.Clean(baseWorktree) && name == "git" && len(args) == 4 && args[0] == "pull" && args[1] == "--ff-only" && args[2] == "origin" && args[3] == "main":
			return "", nil
		default:
			t.Fatalf("unexpected command in %s: %s %v", dir, name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"land"}); err != nil {
		t.Fatalf("land failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "Merged PR #17 and updated local main") {
		t.Fatalf("expected land output, got %q", stdout.String())
	}
	if !hasCommandContaining(commands, "git worktree list --porcelain") {
		t.Fatalf("expected worktree inspection, got %v", commands)
	}
	if !hasCommandContaining(commands, "git pull --ff-only origin main") {
		t.Fatalf("expected base worktree fast-forward, got %v", commands)
	}
	if hasCommandContaining(commands, "git branch -f main origin/main") {
		t.Fatalf("expected no force-update of checked out base branch, got %v", commands)
	}
}

func TestRunLandRejectsDirtyBaseWorktreeUpdate(t *testing.T) {
	tmp, _, app, restore := preparePRLandProject(t)
	defer restore()

	baseWorktree := filepath.Join(tmp, "..", "main-worktree")
	worktreeList := strings.Join([]string{
		"worktree " + filepath.Clean(tmp),
		"HEAD 1111111111111111111111111111111111111111",
		"branch refs/heads/feature/login-audit",
		"",
		"worktree " + filepath.Clean(baseWorktree),
		"HEAD 2222222222222222222222222222222222222222",
		"branch refs/heads/main",
		"",
	}, "\n")

	var commands []string
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		commands = append(commands, name+" "+strings.Join(args, " "))

		switch {
		case dir == tmp && name == "gh" && len(args) == 2 && args[0] == "auth" && args[1] == "status":
			return "", nil
		case dir == tmp && name == "git" && len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
			return "feature/login-audit", nil
		case dir == tmp && name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list":
			return mustMarshalJSON(t, []githubPullRequest{{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Existing PR", HeadRefName: "feature/login-audit", BaseRefName: "main"}}), nil
		case dir == tmp && name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "view":
			return mustMarshalJSON(t, githubPullRequest{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "Existing PR", HeadRefName: "feature/login-audit", BaseRefName: "main", ReviewDecision: "APPROVED", MergeStateStatus: "CLEAN"}), nil
		case dir == tmp && name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "merge":
			return "", nil
		case dir == tmp && name == "git" && len(args) == 3 && args[0] == "fetch" && args[1] == "origin" && args[2] == "main":
			return "", nil
		case dir == tmp && name == "git" && len(args) == 3 && args[0] == "worktree" && args[1] == "list" && args[2] == "--porcelain":
			return worktreeList, nil
		case dir == filepath.Clean(baseWorktree) && name == "git" && len(args) >= 2 && args[0] == "status" && args[1] == "--porcelain":
			return " M README.md", nil
		default:
			t.Fatalf("unexpected command in %s: %s %v", dir, name, args)
			return "", nil
		}
	}

	err := app.Run(context.Background(), []string{"land"})
	if err == nil {
		t.Fatal("expected dirty worktree error")
	}
	if !strings.Contains(err.Error(), "uncommitted changes") {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasCommandContaining(commands, "git pull --ff-only origin main") {
		t.Fatalf("expected no base worktree fast-forward, got %v", commands)
	}
	if hasCommandContaining(commands, "git branch -f main origin/main") {
		t.Fatalf("expected no branch force update, got %v", commands)
	}
}
func preparePRLandProject(t *testing.T) (string, *bytes.Buffer, *App, func()) {
	t.Helper()
	tmp := canonicalTempDir(t)
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}

	restore := chdirExecution(t, tmp)
	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "quality.yaml"), "development_mode: tdd\ntest_command: go test ./...\nlint_command: gofmt -l .\ntypecheck_command: go vet ./...\n")

	app.lookPath = func(name string) (string, error) {
		if name == "gh" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}

	return tmp, stdout, app, restore
}

func mustMarshalJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return string(data)
}

func hasCommandContaining(commands []string, needle string) bool {
	for _, command := range commands {
		if strings.Contains(command, needle) {
			return true
		}
	}
	return false
}
