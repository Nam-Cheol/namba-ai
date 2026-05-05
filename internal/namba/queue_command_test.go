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

func TestParseQueueInvocationStartRangeAndOptions(t *testing.T) {
	t.Parallel()

	inv, err := parseQueueInvocation([]string{"start", "SPEC-001..SPEC-003", "SPEC-005", "--auto-land", "--skip-codex-review", "--remote", "upstream"})
	if err != nil {
		t.Fatalf("parseQueueInvocation returned error: %v", err)
	}
	if inv.Subcommand != "start" || !inv.Options.AutoLand || !inv.Options.SkipCodexReview || inv.Options.Remote != "upstream" {
		t.Fatalf("unexpected invocation: %+v", inv)
	}
	if got := strings.Join(inv.Targets, ","); got != "SPEC-001..SPEC-003,SPEC-005" {
		t.Fatalf("targets = %s", got)
	}
}

func TestExpandQueueTargetsRejectsMissingAndDuplicateSpecs(t *testing.T) {
	t.Parallel()

	tmp := canonicalTempDir(t)
	writeQueueSpecFixture(t, tmp, "SPEC-001")

	if _, err := expandQueueTargets(tmp, []string{"SPEC-001", "SPEC-001"}); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate target error, got %v", err)
	}
	if _, err := expandQueueTargets(tmp, []string{"SPEC-001..SPEC-002"}); err == nil || !strings.Contains(err.Error(), "SPEC-002") {
		t.Fatalf("expected missing SPEC error, got %v", err)
	}
}

func TestQueueStartBlocksOnAmbiguousSpecBranches(t *testing.T) {
	tmp, _, app, restore := prepareQueueProject(t)
	defer restore()
	writeQueueSpecFixture(t, tmp, "SPEC-001")

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && strings.Join(args, " ") == "for-each-ref --format=%(refname:short) refs/heads":
			return "main\nspec/SPEC-001-alpha\nspec/SPEC-001-beta", nil
		default:
			t.Fatalf("unexpected command: %s %v in %s", name, args, dir)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"queue", "start", "SPEC-001", "--skip-codex-review"}); err != nil {
		t.Fatalf("queue start should persist blocked state without returning command failure: %v", err)
	}
	state, err := readQueueState(tmp)
	if err != nil {
		t.Fatalf("read queue state: %v", err)
	}
	if state.Status != queueStateBlocked || state.Detail != "branch_ambiguous" {
		t.Fatalf("unexpected state: %+v", state)
	}
	if !strings.Contains(state.LastBlocker, "multiple branches") {
		t.Fatalf("expected branch ambiguity blocker, got %q", state.LastBlocker)
	}
}

func TestEnsureQueueBranchBlocksDirtyCurrentQueueBranch(t *testing.T) {
	tmp, _, app, restore := prepareQueueProject(t)
	defer restore()

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && strings.Join(args, " ") == "branch --show-current":
			return "spec/SPEC-001-queue-fixture", nil
		case name == "git" && strings.Join(args, " ") == "status --porcelain":
			return " M unrelated.txt", nil
		default:
			t.Fatalf("unexpected command: %s %v in %s", name, args, dir)
			return "", nil
		}
	}

	err := app.ensureQueueBranch(context.Background(), tmp, queueState{}, "spec/SPEC-001-queue-fixture")
	if err == nil || !strings.Contains(err.Error(), "uncommitted changes") {
		t.Fatalf("expected dirty current queue branch to block, got %v", err)
	}
}

func TestQueueStartSkipsSpecAlreadyMergedIntoBase(t *testing.T) {
	tmp, _, app, restore := prepareQueueProject(t)
	defer restore()
	writeQueueSpecFixture(t, tmp, "SPEC-001")

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && strings.Join(args, " ") == "for-each-ref --format=%(refname:short) refs/heads":
			return "main\nspec/SPEC-001-queue-fixture", nil
		case name == "git" && strings.Join(args, " ") == "branch --list spec/SPEC-001-queue-fixture":
			return "  spec/SPEC-001-queue-fixture", nil
		case name == "git" && strings.Join(args, " ") == "merge-base --is-ancestor spec/SPEC-001-queue-fixture main":
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v in %s", name, args, dir)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"queue", "start", "SPEC-001", "--skip-codex-review"}); err != nil {
		t.Fatalf("queue start failed: %v", err)
	}
	state, err := readQueueState(tmp)
	if err != nil {
		t.Fatalf("read queue state: %v", err)
	}
	if state.Status != queueStateDone || state.SkippedSpecCount != 1 || len(state.SkippedSpecs) != 1 || state.SkippedSpecs[0] != "SPEC-001" {
		t.Fatalf("expected skipped done state, got %+v", state)
	}
	if got := state.Specs["SPEC-001"]; got.Phase != queuePhaseSkipped || !strings.Contains(got.SkipReason, "already merged") {
		t.Fatalf("expected skipped phase with landed evidence, got %+v", got)
	}
}

func TestQueueStartSkipsMergedPullRequestWhenLocalBranchIsMissing(t *testing.T) {
	tmp, _, app, restore := prepareQueueProject(t)
	defer restore()
	writeQueueSpecFixture(t, tmp, "SPEC-001")

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && strings.Join(args, " ") == "for-each-ref --format=%(refname:short) refs/heads":
			return "main", nil
		case name == "git" && len(args) == 3 && args[0] == "branch" && args[1] == "--list":
			return "", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list" && hasArg(args, "--state", "merged"):
			return mustMarshalJSON(t, []githubPullRequest{{Number: 23, URL: "https://github.com/example/repo/pull/23", State: "MERGED", MergedAt: "2026-05-05T10:00:00Z", MergeCommit: githubCommitRef{OID: "merge-23"}, HeadRefName: "spec/SPEC-001-queue-fixture", BaseRefName: "main"}}), nil
		case name == "git" && strings.Join(args, " ") == "log -1 --format=%H main -- .namba/specs/SPEC-001":
			return "spec-base", nil
		case name == "git" && strings.Join(args, " ") == "merge-base --is-ancestor spec-base merge-23":
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v in %s", name, args, dir)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"queue", "start", "SPEC-001"}); err != nil {
		t.Fatalf("queue start failed: %v", err)
	}
	state, err := readQueueState(tmp)
	if err != nil {
		t.Fatalf("read queue state: %v", err)
	}
	if state.Status != queueStateDone || state.SkippedSpecCount != 1 {
		t.Fatalf("expected merged PR skip, got %+v", state)
	}
	if got := state.Specs["SPEC-001"]; got.Phase != queuePhaseSkipped || !strings.Contains(got.SkipReason, "PR #23 merged") {
		t.Fatalf("expected merged PR skip evidence, got %+v", got)
	}
}

func TestQueueLandedEvidenceDoesNotUseMergedPRFallbackForLiveUnmergedBranch(t *testing.T) {
	tmp, _, app, restore := prepareQueueProject(t)
	defer restore()

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && strings.Join(args, " ") == "branch --list spec/SPEC-001-queue-fixture":
			return "  spec/SPEC-001-queue-fixture", nil
		case name == "git" && strings.Join(args, " ") == "merge-base --is-ancestor spec/SPEC-001-queue-fixture main":
			return "", errors.New("not ancestor")
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list" && hasArg(args, "--state", "merged"):
			return mustMarshalJSON(t, []githubPullRequest{{Number: 23, State: "MERGED", MergedAt: "2026-05-05T10:00:00Z", MergeCommit: githubCommitRef{OID: "old-merge"}}}), nil
		case name == "git" && strings.Join(args, " ") == "log -1 --format=%H main -- .namba/specs/SPEC-001":
			return "new-spec-base", nil
		case name == "git" && strings.Join(args, " ") == "merge-base --is-ancestor new-spec-base old-merge":
			return "", errors.New("stale merged PR does not contain latest SPEC package commit")
		default:
			t.Fatalf("unexpected command: %s %v in %s", name, args, dir)
			return "", nil
		}
	}

	landed, evidence := app.queueLandedEvidence(context.Background(), tmp, "SPEC-001", "spec/SPEC-001-queue-fixture")
	if landed || evidence != "" {
		t.Fatalf("expected live unmerged branch not to use stale merged PR fallback, landed=%v evidence=%q", landed, evidence)
	}
}

func TestQueueLandedEvidenceUsesMergedPRFallbackForSquashMergedBranch(t *testing.T) {
	tmp, _, app, restore := prepareQueueProject(t)
	defer restore()

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && strings.Join(args, " ") == "branch --list spec/SPEC-001-queue-fixture":
			return "  spec/SPEC-001-queue-fixture", nil
		case name == "git" && strings.Join(args, " ") == "merge-base --is-ancestor spec/SPEC-001-queue-fixture main":
			return "", errors.New("squash merge leaves branch tip outside base ancestry")
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list" && hasArg(args, "--state", "merged"):
			return mustMarshalJSON(t, []githubPullRequest{{Number: 23, State: "MERGED", MergedAt: "2026-05-05T10:00:00Z", MergeCommit: githubCommitRef{OID: "squash-merge"}}}), nil
		case name == "git" && strings.Join(args, " ") == "log -1 --format=%H main -- .namba/specs/SPEC-001":
			return "spec-base", nil
		case name == "git" && strings.Join(args, " ") == "merge-base --is-ancestor spec-base squash-merge":
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v in %s", name, args, dir)
			return "", nil
		}
	}

	landed, evidence := app.queueLandedEvidence(context.Background(), tmp, "SPEC-001", "spec/SPEC-001-queue-fixture")
	if !landed || !strings.Contains(evidence, "PR #23 merged") {
		t.Fatalf("expected squash-merged branch to use merged PR evidence, landed=%v evidence=%q", landed, evidence)
	}
}

func TestQueueMergedPullRequestEvidenceRequiresSpecBaseCommitInMergeCommit(t *testing.T) {
	tmp, _, app, restore := prepareQueueProject(t)
	defer restore()

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list" && hasArg(args, "--state", "merged"):
			if !hasArg(args, "--limit", "1000") {
				t.Fatalf("expected merged PR list to request full history, got %v", args)
			}
			return mustMarshalJSON(t, []githubPullRequest{{Number: 23, URL: "https://github.com/example/repo/pull/23", State: "MERGED", MergedAt: "2026-05-05T10:00:00Z", MergeCommit: githubCommitRef{OID: "old-merge"}}}), nil
		case name == "git" && strings.Join(args, " ") == "log -1 --format=%H main -- .namba/specs/SPEC-001":
			return "new-spec-base", nil
		case name == "git" && strings.Join(args, " ") == "merge-base --is-ancestor new-spec-base old-merge":
			return "", errors.New("stale merged PR does not contain latest SPEC package commit")
		default:
			t.Fatalf("unexpected command: %s %v in %s", name, args, dir)
			return "", nil
		}
	}

	landed, evidence := app.queueMergedPullRequestEvidence(context.Background(), tmp, "SPEC-001", "spec/SPEC-001-queue-fixture", "main")
	if landed || evidence != "" {
		t.Fatalf("expected stale merged PR history not to be landed evidence, landed=%v evidence=%q", landed, evidence)
	}
}

func TestQueueResumePostPRPhaseSkipsSyncAndPRPreparation(t *testing.T) {
	tmp, _, app, restore := prepareQueueProject(t)
	defer restore()
	writeQueueSpecFixture(t, tmp, "SPEC-001")

	branch := "spec/SPEC-001-queue-fixture"
	state := queueState{
		ID:             "queue-test",
		Status:         queueStateWaiting,
		OperatorState:  queueOperatorWaiting,
		Detail:         queuePhaseChecksPending,
		Targets:        []string{"SPEC-001"},
		ActiveSpecID:   "SPEC-001",
		ExpectedBranch: branch,
		Options:        queueOptions{Remote: defaultGitRemote},
		Specs: map[string]queueSpec{"SPEC-001": {
			SpecID:             "SPEC-001",
			Status:             queueStateActive,
			OperatorState:      queueOperatorWaiting,
			Phase:              queuePhaseChecksPending,
			Branch:             branch,
			PRNumber:           17,
			PRURL:              "https://github.com/example/repo/pull/17",
			ValidationEvidence: queueRunEvidencePath("SPEC-001"),
		}},
	}
	if err := app.writeQueueState(tmp, state); err != nil {
		t.Fatalf("write queue state: %v", err)
	}

	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && strings.Join(args, " ") == "branch --list "+branch:
			return "  " + branch, nil
		case name == "git" && strings.Join(args, " ") == "merge-base --is-ancestor "+branch+" main":
			return "", errors.New("not ancestor")
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list" && hasArg(args, "--state", "merged"):
			return "[]", nil
		case name == "git" && strings.Join(args, " ") == "branch --show-current":
			return branch, nil
		case name == "git" && strings.Join(args, " ") == "status --porcelain":
			return "", nil
		case name == "git" && strings.Join(args, " ") == "rev-parse HEAD":
			return "abc123", nil
		case name == "gh" && len(args) >= 3 && args[0] == "pr" && args[1] == "view" && args[2] == "17":
			return mustMarshalJSON(t, githubPullRequest{
				Number:           17,
				URL:              "https://github.com/example/repo/pull/17",
				Title:            "SPEC-001",
				HeadRefName:      branch,
				HeadRefOID:       "abc123",
				BaseRefName:      "main",
				ReviewDecision:   "APPROVED",
				MergeStateStatus: "CLEAN",
				StatusChecks:     []githubStatusCheck{{Name: "ci", Status: "COMPLETED", Conclusion: "SUCCESS"}},
			}), nil
		case name == "git" && strings.Join(args, " ") == "add -A":
			t.Fatalf("resume in checks_pending must not stage regenerated output")
		case name == "git" && len(args) >= 2 && args[0] == "push":
			t.Fatalf("resume in checks_pending must not push a new PR head")
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "create":
			t.Fatalf("resume in checks_pending must not recreate PR")
		default:
			t.Fatalf("unexpected command: %s %v in %s", name, args, dir)
		}
		return "", nil
	}

	if err := app.Run(context.Background(), []string{"queue", "resume"}); err != nil {
		t.Fatalf("queue resume failed: %v", err)
	}
	got, err := readQueueState(tmp)
	if err != nil {
		t.Fatalf("read queue state: %v", err)
	}
	if got.Status != queueStateWaiting || got.Detail != queuePhaseWaitingLand {
		t.Fatalf("expected resume to validate existing PR and wait for land, got %+v", got)
	}
}

func TestQueueStartPreparesActiveSpecPRAndSkipsReviewComment(t *testing.T) {
	tmp, stdout, app, restore := prepareQueueProject(t)
	defer restore()
	writeQueueSpecFixture(t, tmp, "SPEC-001")
	writeQueueSpecFixture(t, tmp, "SPEC-002")
	writeQueueRunEvidence(t, tmp, "SPEC-001")

	currentBranch := "main"
	var commands []string
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		commands = append(commands, name+" "+strings.Join(args, " "))
		switch {
		case name == "git" && strings.Join(args, " ") == "for-each-ref --format=%(refname:short) refs/heads":
			return "main", nil
		case name == "git" && strings.Join(args, " ") == "branch --show-current":
			return currentBranch, nil
		case name == "git" && len(args) == 3 && args[0] == "branch" && args[1] == "--list":
			return "", nil
		case name == "git" && len(args) == 4 && args[0] == "checkout" && args[1] == "-b":
			currentBranch = args[2]
			return "", nil
		case name == "git" && strings.Join(args, " ") == "rev-parse HEAD":
			return "abc123", nil
		case name == "git" && strings.Join(args, " ") == "status --porcelain":
			if currentBranch == "main" {
				return "", nil
			}
			return " M .namba/project/change-summary.md", nil
		case name == "git" && strings.Join(args, " ") == "add -A":
			return "", nil
		case name == "git" && len(args) == 3 && args[0] == "commit" && args[1] == "-m":
			return "", nil
		case name == "git" && len(args) == 4 && args[0] == "push" && args[1] == "--set-upstream":
			return "", nil
		case name == "gh" && strings.Join(args, " ") == "auth status":
			return "", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list":
			return "[]", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "create":
			bodyIndex := indexOfArg(args, "--body")
			if bodyIndex == -1 || !strings.Contains(args[bodyIndex+1], ".namba/specs/SPEC-001/reviews/readiness.md") || strings.Contains(args[bodyIndex+1], "SPEC-002/reviews") {
				t.Fatalf("expected active SPEC-001 PR body, got %v", args)
			}
			return "https://github.com/example/repo/pull/17", nil
		case name == "gh" && len(args) >= 3 && args[0] == "pr" && args[1] == "view" && args[2] == currentBranch:
			return mustMarshalJSON(t, githubPullRequest{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "SPEC-001", HeadRefName: currentBranch, BaseRefName: "main"}), nil
		case name == "gh" && len(args) >= 3 && args[0] == "pr" && args[1] == "view" && args[2] == "17":
			return mustMarshalJSON(t, githubPullRequest{
				Number:           17,
				URL:              "https://github.com/example/repo/pull/17",
				Title:            "SPEC-001",
				HeadRefName:      currentBranch,
				BaseRefName:      "main",
				ReviewDecision:   "APPROVED",
				MergeStateStatus: "CLEAN",
				StatusChecks:     []githubStatusCheck{{Name: "ci", Status: "COMPLETED", Conclusion: "SUCCESS"}},
			}), nil
		default:
			t.Fatalf("unexpected command: %s %v in %s", name, args, dir)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"queue", "start", "SPEC-001", "--skip-codex-review"}); err != nil {
		t.Fatalf("queue start failed: %v", err)
	}
	if hasCommandContaining(commands, "gh pr comment") {
		t.Fatalf("expected queue-scoped review skip to avoid PR comment, got %v", commands)
	}
	state, err := readQueueState(tmp)
	if err != nil {
		t.Fatalf("read queue state: %v", err)
	}
	if state.Status != queueStateWaiting || state.Detail != queuePhaseWaitingLand || state.ActiveSpecID != "SPEC-001" {
		t.Fatalf("unexpected queue state: %+v", state)
	}
	if !strings.Contains(stdout.String(), "waiting_for_land") {
		t.Fatalf("expected waiting output, got %q", stdout.String())
	}
}

func TestQueueStartBlocksDraftPullRequest(t *testing.T) {
	tmp, _, app, restore := prepareQueueProject(t)
	defer restore()
	writeQueueSpecFixture(t, tmp, "SPEC-001")
	writeQueueRunEvidence(t, tmp, "SPEC-001")

	currentBranch := "main"
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "git" && strings.Join(args, " ") == "for-each-ref --format=%(refname:short) refs/heads":
			return "main", nil
		case name == "git" && strings.Join(args, " ") == "branch --show-current":
			return currentBranch, nil
		case name == "git" && len(args) == 3 && args[0] == "branch" && args[1] == "--list":
			return "", nil
		case name == "git" && len(args) == 4 && args[0] == "checkout" && args[1] == "-b":
			currentBranch = args[2]
			return "", nil
		case name == "git" && strings.Join(args, " ") == "rev-parse HEAD":
			return "abc123", nil
		case name == "git" && strings.Join(args, " ") == "status --porcelain":
			return "", nil
		case name == "git" && len(args) == 4 && args[0] == "push" && args[1] == "--set-upstream":
			return "", nil
		case name == "gh" && strings.Join(args, " ") == "auth status":
			return "", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list" && hasArg(args, "--state", "merged"):
			return "[]", nil
		case name == "gh" && len(args) >= 2 && args[0] == "pr" && args[1] == "list":
			return mustMarshalJSON(t, []githubPullRequest{{Number: 17}}), nil
		case name == "gh" && len(args) >= 3 && args[0] == "pr" && args[1] == "view" && args[2] == "17":
			fields := strings.Join(args, " ")
			if strings.Contains(fields, "statusCheckRollup") {
				return mustMarshalJSON(t, githubPullRequest{
					Number:           17,
					URL:              "https://github.com/example/repo/pull/17",
					Title:            "SPEC-001",
					HeadRefName:      currentBranch,
					BaseRefName:      "main",
					IsDraft:          true,
					ReviewDecision:   "APPROVED",
					MergeStateStatus: "CLEAN",
					StatusChecks:     []githubStatusCheck{{Name: "ci", Status: "COMPLETED", Conclusion: "SUCCESS"}},
				}), nil
			}
			return mustMarshalJSON(t, githubPullRequest{Number: 17, URL: "https://github.com/example/repo/pull/17", Title: "SPEC-001", HeadRefName: currentBranch, BaseRefName: "main"}), nil
		default:
			t.Fatalf("unexpected command: %s %v in %s", name, args, dir)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"queue", "start", "SPEC-001", "--skip-codex-review"}); err != nil {
		t.Fatalf("queue start should persist blocked state without returning command failure: %v", err)
	}
	state, err := readQueueState(tmp)
	if err != nil {
		t.Fatalf("read queue state: %v", err)
	}
	if state.Status != queueStateBlocked || state.Detail != "pr_draft" {
		t.Fatalf("expected draft PR block, got %+v", state)
	}
}

func TestBuildPullRequestBodyForSpecUsesActiveSpec(t *testing.T) {
	t.Parallel()

	tmp := canonicalTempDir(t)
	writeQueueSpecFixture(t, tmp, "SPEC-001")
	writeQueueSpecFixture(t, tmp, "SPEC-002")
	body := buildPullRequestBodyForSpec(tmp, initProfile{PRLanguage: "en"}, "SPEC-001")
	if !strings.Contains(body, ".namba/specs/SPEC-001/reviews/readiness.md") || strings.Contains(body, "SPEC-002") {
		t.Fatalf("expected active SPEC readiness in PR body, got %q", body)
	}
}

func TestClassifyQueueCheckProofRequiresSurfacedChecks(t *testing.T) {
	t.Parallel()

	if proof, err := classifyQueueCheckProof(githubPullRequest{}); err == nil || proof != "" {
		t.Fatalf("expected ambiguous check proof, got proof=%q err=%v", proof, err)
	}
	proof, err := classifyQueueCheckProof(githubPullRequest{StatusChecks: []githubStatusCheck{{Name: "ci", Status: "COMPLETED", Conclusion: "SUCCESS"}}})
	if err != nil || proof != "all_surfaced_checks_green" {
		t.Fatalf("unexpected check proof: proof=%q err=%v", proof, err)
	}
}

func TestQueueMergePullRequestArgsPinsHeadCommit(t *testing.T) {
	t.Parallel()

	args, err := queueMergePullRequestArgs(githubPullRequest{Number: 17, HeadRefOID: "abc123"})
	if err != nil {
		t.Fatalf("queueMergePullRequestArgs returned error: %v", err)
	}
	if got := strings.Join(args, " "); got != "pr merge 17 --merge --match-head-commit abc123" {
		t.Fatalf("unexpected merge args: %s", got)
	}

	if _, err := queueMergePullRequestArgs(githubPullRequest{Number: 17}); err == nil || !strings.Contains(err.Error(), "headRefOid") {
		t.Fatalf("expected missing headRefOid error, got %v", err)
	}
}

func TestQueueExecutionSucceededRequiresCurrentHead(t *testing.T) {
	t.Parallel()

	tmp := canonicalTempDir(t)
	writeQueueSpecFixture(t, tmp, "SPEC-001")
	writeQueueRunEvidenceWithHead(t, tmp, "SPEC-001", "abc123")

	ok, evidence := queueExecutionSucceeded(tmp, "SPEC-001", "abc123")
	if !ok || !strings.Contains(evidence, "spec-001-validation.json") {
		t.Fatalf("expected matching head evidence to pass, ok=%v evidence=%q", ok, evidence)
	}
	ok, evidence = queueExecutionSucceeded(tmp, "SPEC-001", "def456")
	if ok || !strings.Contains(evidence, "spec-001-execution.json") {
		t.Fatalf("expected mismatched head evidence to fail on execution evidence, ok=%v evidence=%q", ok, evidence)
	}
}

func TestQueueExecutionSatisfiedTrustsPostExecutionCheckpoint(t *testing.T) {
	t.Parallel()

	tmp := canonicalTempDir(t)
	writeQueueSpecFixture(t, tmp, "SPEC-001")
	writeQueueRunEvidenceWithHead(t, tmp, "SPEC-001", "run-head")

	specState := queueSpec{
		SpecID:             "SPEC-001",
		Phase:              queuePhaseChecksPending,
		PRNumber:           17,
		ValidationEvidence: queueRunEvidencePath("SPEC-001"),
	}
	ok, evidence := queueExecutionSatisfied(tmp, "SPEC-001", "post-pr-head", specState)
	if !ok || !strings.Contains(evidence, "spec-001-validation.json") {
		t.Fatalf("expected post-execution checkpoint to satisfy execution, ok=%v evidence=%q", ok, evidence)
	}

	specState.Phase = queuePhasePRReady
	specState.PRNumber = 0
	ok, evidence = queueExecutionSatisfied(tmp, "SPEC-001", "post-pr-head", specState)
	if ok || !strings.Contains(evidence, "spec-001-execution.json") {
		t.Fatalf("expected pre-PR checkpoint to keep enforcing current HEAD evidence, ok=%v evidence=%q", ok, evidence)
	}
}

func TestHonorQueueControlRequestsReloadsPauseFromStateFile(t *testing.T) {
	tmp, _, app, restore := prepareQueueProject(t)
	defer restore()
	state := queueState{
		ID:            "queue-test",
		Status:        queueStateActive,
		OperatorState: queueOperatorRunning,
		Detail:        queuePhaseRunning,
		Targets:       []string{"SPEC-001"},
		Specs:         map[string]queueSpec{"SPEC-001": {SpecID: "SPEC-001", Status: queueStateActive, Phase: queuePhaseRunning}},
	}
	if err := app.writeQueueState(tmp, state); err != nil {
		t.Fatalf("write queue state: %v", err)
	}
	latest := state
	latest.PauseRequested = true
	if err := app.writeQueueState(tmp, latest); err != nil {
		t.Fatalf("write pause request: %v", err)
	}

	got, done, err := app.honorQueueControlRequests(tmp, state)
	if err != nil {
		t.Fatalf("honor control requests: %v", err)
	}
	if !done || got.Status != queueStatePaused || got.Detail != queuePhasePaused {
		t.Fatalf("expected reloaded pause request, done=%v state=%+v", done, got)
	}
}

func TestQueuePauseAndStopPersistDistinctStates(t *testing.T) {
	tmp, _, app, restore := prepareQueueProject(t)
	defer restore()
	state := queueState{
		ID:            "queue-test",
		Status:        queueStateActive,
		OperatorState: queueOperatorRunning,
		Detail:        queuePhaseRunning,
		Targets:       []string{"SPEC-001"},
		Specs:         map[string]queueSpec{"SPEC-001": {SpecID: "SPEC-001", Status: queueStateActive, Phase: queuePhaseRunning}},
	}
	if err := app.writeQueueState(tmp, state); err != nil {
		t.Fatalf("write queue state: %v", err)
	}

	if err := app.Run(context.Background(), []string{"queue", "pause"}); err != nil {
		t.Fatalf("pause failed: %v", err)
	}
	paused, err := readQueueState(tmp)
	if err != nil {
		t.Fatalf("read paused state: %v", err)
	}
	if !paused.PauseRequested || paused.Status != queueStatePaused || paused.Detail != queuePhasePaused {
		t.Fatalf("unexpected paused state: %+v", paused)
	}

	if err := app.Run(context.Background(), []string{"queue", "stop"}); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	stopped, err := readQueueState(tmp)
	if err != nil {
		t.Fatalf("read stopped state: %v", err)
	}
	if !stopped.StopRequested || stopped.Status != queueStateStopped || stopped.Detail != queuePhaseStopped {
		t.Fatalf("unexpected stopped state: %+v", stopped)
	}
}

func prepareQueueProject(t *testing.T) (string, *bytes.Buffer, *App, func()) {
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
	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "quality.yaml"), "development_mode: tdd\ntest_command: none\nlint_command: none\ntypecheck_command: none\nbuild_command: none\nmigration_dry_run_command: none\nsmoke_start_command: none\noutput_contract_command: none\n")
	app.lookPath = func(name string) (string, error) {
		if name == "gh" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	return tmp, stdout, app, restore
}

func writeQueueSpecFixture(t *testing.T, root, specID string) {
	t.Helper()
	specDir := filepath.Join(root, ".namba", "specs", specID)
	writeTestFile(t, filepath.Join(specDir, "spec.md"), "# "+specID+"\n\n## Problem\n\nQueue fixture for "+specID+".\n")
	writeTestFile(t, filepath.Join(specDir, "plan.md"), "# "+specID+" Plan\n\n1. Implement.\n")
	writeTestFile(t, filepath.Join(specDir, "acceptance.md"), "# Acceptance\n\n- [ ] Validation passes.\n")
	reviewsDir := filepath.Join(specDir, "reviews")
	for _, template := range specReviewTemplates() {
		writeTestFile(t, filepath.Join(reviewsDir, template.Slug+".md"), strings.Join([]string{
			"# " + template.Title,
			"",
			"- Status: clear",
			"- Last Reviewed: 2026-05-05",
			"- Reviewer: test",
			"",
			"## Findings",
			"",
			"- Clear.",
			"",
			"## Follow-ups",
			"",
			"- [non-blocking] None.",
			"",
		}, "\n"))
	}
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if _, err := app.refreshSpecReviewReadiness(root, specID); err != nil {
		t.Fatalf("refresh readiness: %v", err)
	}
}

func writeQueueRunEvidence(t *testing.T, root, specID string) {
	t.Helper()
	writeQueueRunEvidenceWithHead(t, root, specID, "abc123")
}

func writeQueueRunEvidenceWithHead(t *testing.T, root, specID, headSHA string) {
	t.Helper()
	logID := strings.ToLower(specID)
	if err := writeJSONFile(filepath.Join(root, ".namba", "logs", "runs", logID+"-execution.json"), executionResult{SpecID: specID, HeadSHA: headSHA, Succeeded: true}); err != nil {
		t.Fatalf("write execution evidence: %v", err)
	}
	if err := writeJSONFile(filepath.Join(root, ".namba", "logs", "runs", logID+"-validation.json"), validationReport{SpecID: specID, HeadSHA: headSHA, Passed: true}); err != nil {
		t.Fatalf("write validation evidence: %v", err)
	}
}

func hasArg(args []string, key, value string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == key && args[i+1] == value {
			return true
		}
	}
	return false
}
