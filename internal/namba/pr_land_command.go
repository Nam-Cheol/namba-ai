package namba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultGitRemote         = "origin"
	codexReviewRequestMarker = "<!-- namba-codex-review-request -->"
)

type prOptions struct {
	Title          string
	Remote         string
	SkipSync       bool
	SkipValidation bool
}

type landOptions struct {
	Number int
	Remote string
	Wait   bool
}

type githubPullRequest struct {
	Number           int                 `json:"number"`
	URL              string              `json:"url"`
	Title            string              `json:"title"`
	BaseRefName      string              `json:"baseRefName"`
	HeadRefName      string              `json:"headRefName"`
	ReviewDecision   string              `json:"reviewDecision"`
	MergeStateStatus string              `json:"mergeStateStatus"`
	IsDraft          bool                `json:"isDraft"`
	Comments         []githubPRComment   `json:"comments"`
	StatusChecks     []githubStatusCheck `json:"statusCheckRollup"`
}

type githubPRComment struct {
	Body string `json:"body"`
}

type githubStatusCheck struct {
	Name         string `json:"name"`
	Context      string `json:"context"`
	WorkflowName string `json:"workflowName"`
	Status       string `json:"status"`
	State        string `json:"state"`
	Conclusion   string `json:"conclusion"`
}

type gitWorktree struct {
	Path   string
	Branch string
}

func (a *App) runPR(ctx context.Context, args []string) error {
	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}
	if !isGitRepository(root) {
		return errors.New("pr requires a git repository")
	}

	opts, err := parsePRArgs(args)
	if err != nil {
		return err
	}

	profile, err := a.loadInitProfileFromConfig(root)
	if err != nil {
		return err
	}
	if !strings.EqualFold(strings.TrimSpace(profile.GitProvider), "github") {
		return fmt.Errorf("pr currently supports only the GitHub provider, got %q", profile.GitProvider)
	}
	if err := a.requireGitHubCLI(ctx, root); err != nil {
		return err
	}

	baseBranch := prBaseBranch(profile)
	currentBranch, err := a.currentBranch(ctx, root)
	if err != nil {
		return fmt.Errorf("detect current branch: %w", err)
	}
	if currentBranch == baseBranch {
		return fmt.Errorf("pr requires a work branch; current branch is %q", baseBranch)
	}

	if !opts.SkipSync {
		if err := a.runSync(ctx, nil); err != nil {
			return fmt.Errorf("sync before pr: %w", err)
		}
	}
	if !opts.SkipValidation {
		qualityCfg, err := a.loadQualityConfig(root)
		if err != nil {
			return err
		}
		if err := a.runValidators(ctx, root, qualityCfg); err != nil {
			return err
		}
	}

	dirty, err := a.hasWorkingTreeChanges(ctx, root)
	if err != nil {
		return err
	}
	if dirty {
		if _, err := a.runBinary(ctx, "git", []string{"add", "-A"}, root); err != nil {
			return fmt.Errorf("stage changes: %w", err)
		}
		if _, err := a.runBinary(ctx, "git", []string{"commit", "-m", opts.Title}, root); err != nil {
			return fmt.Errorf("create commit: %w", err)
		}
	}

	if _, err := a.runBinary(ctx, "git", []string{"push", "--set-upstream", opts.Remote, currentBranch}, root); err != nil {
		return fmt.Errorf("push branch %s: %w", currentBranch, err)
	}

	pr, created, err := a.findOrCreatePullRequest(ctx, root, currentBranch, baseBranch, opts.Title, buildPullRequestBody(profile))
	if err != nil {
		return err
	}
	if profile.AutoCodexReview {
		if err := a.ensureReviewComment(ctx, root, pr.Number, codexReviewComment(profile)); err != nil {
			return err
		}
	}

	if created {
		fmt.Fprintf(a.stdout, "Prepared PR #%d %s\n", pr.Number, pr.URL)
		return nil
	}
	fmt.Fprintf(a.stdout, "Reused PR #%d %s\n", pr.Number, pr.URL)
	return nil
}

func (a *App) runLand(ctx context.Context, args []string) error {
	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}
	if !isGitRepository(root) {
		return errors.New("land requires a git repository")
	}

	opts, err := parseLandArgs(args)
	if err != nil {
		return err
	}

	profile, err := a.loadInitProfileFromConfig(root)
	if err != nil {
		return err
	}
	if !strings.EqualFold(strings.TrimSpace(profile.GitProvider), "github") {
		return fmt.Errorf("land currently supports only the GitHub provider, got %q", profile.GitProvider)
	}
	if err := a.requireGitHubCLI(ctx, root); err != nil {
		return err
	}

	baseBranch := prBaseBranch(profile)
	currentBranch, err := a.currentBranch(ctx, root)
	if err != nil {
		return fmt.Errorf("detect current branch: %w", err)
	}

	pr, err := a.resolvePullRequestForLand(ctx, root, opts, currentBranch, baseBranch)
	if err != nil {
		return err
	}
	if pr.BaseRefName != "" && pr.BaseRefName != baseBranch {
		return fmt.Errorf("pull request #%d targets %q, expected %q", pr.Number, pr.BaseRefName, baseBranch)
	}
	if pr.IsDraft {
		return fmt.Errorf("pull request #%d is still a draft", pr.Number)
	}

	if opts.Wait {
		if _, err := a.runBinary(ctx, "gh", []string{"pr", "checks", strconv.Itoa(pr.Number), "--watch"}, root); err != nil {
			return fmt.Errorf("wait for pull request #%d checks: %w", pr.Number, err)
		}
		pr, err = a.loadPullRequest(ctx, root, strconv.Itoa(pr.Number), landPullRequestFields()...)
		if err != nil {
			return err
		}
	}

	if err := validateLandPullRequest(pr, opts.Wait); err != nil {
		return err
	}

	if _, err := a.runBinary(ctx, "gh", []string{"pr", "merge", strconv.Itoa(pr.Number), "--merge"}, root); err != nil {
		return fmt.Errorf("merge pull request #%d: %w", pr.Number, err)
	}
	if err := a.updateLocalBaseBranch(ctx, root, currentBranch, baseBranch, opts.Remote); err != nil {
		return err
	}

	fmt.Fprintf(a.stdout, "Merged PR #%d and updated local %s\n", pr.Number, baseBranch)
	return nil
}

func parsePRArgs(args []string) (prOptions, error) {
	opts := prOptions{Remote: defaultGitRemote}
	titleParts := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--remote":
			value, err := consumeFlagValue(args, &i, args[i])
			if err != nil {
				return prOptions{}, err
			}
			opts.Remote = strings.TrimSpace(value)
		case "--no-sync":
			opts.SkipSync = true
		case "--no-validate":
			opts.SkipValidation = true
		default:
			if strings.HasPrefix(args[i], "--") {
				return prOptions{}, fmt.Errorf("unknown flag %q", args[i])
			}
			titleParts = append(titleParts, args[i])
		}
	}

	opts.Title = strings.TrimSpace(strings.Join(titleParts, " "))
	if opts.Title == "" {
		return prOptions{}, errors.New("pr requires a title")
	}
	if opts.Remote == "" {
		return prOptions{}, errors.New("pr remote is required")
	}
	return opts, nil
}

func parseLandArgs(args []string) (landOptions, error) {
	opts := landOptions{Remote: defaultGitRemote}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--wait":
			opts.Wait = true
		case "--remote":
			value, err := consumeFlagValue(args, &i, args[i])
			if err != nil {
				return landOptions{}, err
			}
			opts.Remote = strings.TrimSpace(value)
		default:
			if strings.HasPrefix(args[i], "--") {
				return landOptions{}, fmt.Errorf("unknown flag %q", args[i])
			}
			if opts.Number != 0 {
				return landOptions{}, errors.New("land accepts at most one pull request number")
			}
			number, err := strconv.Atoi(strings.TrimSpace(args[i]))
			if err != nil || number <= 0 {
				return landOptions{}, fmt.Errorf("invalid pull request number %q", args[i])
			}
			opts.Number = number
		}
	}

	if opts.Remote == "" {
		return landOptions{}, errors.New("land remote is required")
	}
	return opts, nil
}

func consumeFlagValue(args []string, index *int, flag string) (string, error) {
	*index = *index + 1
	if *index >= len(args) {
		return "", fmt.Errorf("%s requires a value", flag)
	}
	return args[*index], nil
}

func (a *App) requireGitHubCLI(ctx context.Context, root string) error {
	if _, err := a.lookPath("gh"); err != nil {
		return errors.New("GitHub CLI is required; install `gh` and run `gh auth login`")
	}
	if _, err := a.runBinary(ctx, "gh", []string{"auth", "status"}, root); err != nil {
		return fmt.Errorf("GitHub CLI authentication is required; run `gh auth login`: %w", err)
	}
	return nil
}

func (a *App) hasWorkingTreeChanges(ctx context.Context, root string) (bool, error) {
	status, err := a.runBinary(ctx, "git", []string{"status", "--porcelain"}, root)
	if err != nil {
		return false, fmt.Errorf("check working tree: %w", err)
	}
	return strings.TrimSpace(status) != "", nil
}

func (a *App) findOrCreatePullRequest(ctx context.Context, root, headBranch, baseBranch, title, body string) (githubPullRequest, bool, error) {
	pullRequests, err := a.listOpenPullRequests(ctx, root, headBranch, baseBranch)
	if err != nil {
		return githubPullRequest{}, false, err
	}
	if len(pullRequests) > 0 {
		pr, err := a.loadPullRequest(ctx, root, strconv.Itoa(pullRequests[0].Number), reviewPullRequestFields()...)
		if err != nil {
			return githubPullRequest{}, false, err
		}
		return pr, false, nil
	}

	if _, err := a.runBinary(ctx, "gh", []string{"pr", "create", "--base", baseBranch, "--head", headBranch, "--title", title, "--body", body}, root); err != nil {
		return githubPullRequest{}, false, fmt.Errorf("create pull request from %s into %s: %w", headBranch, baseBranch, err)
	}

	pr, err := a.loadPullRequest(ctx, root, headBranch, reviewPullRequestFields()...)
	if err != nil {
		return githubPullRequest{}, false, err
	}
	return pr, true, nil
}

func (a *App) ensureReviewComment(ctx context.Context, root string, prNumber int, reviewCommand string) error {
	pr, err := a.loadPullRequest(ctx, root, strconv.Itoa(prNumber), "comments")
	if err != nil {
		return err
	}
	for _, comment := range pr.Comments {
		if isReviewRequestComment(comment.Body, reviewCommand) {
			return nil
		}
	}
	commentBody := buildReviewRequestCommentBody(reviewCommand)
	if _, err := a.runBinary(ctx, "gh", []string{"pr", "comment", strconv.Itoa(prNumber), "--body", commentBody}, root); err != nil {
		return fmt.Errorf("request Codex review on pull request #%d: %w", prNumber, err)
	}
	return nil
}

func isReviewRequestComment(body, reviewCommand string) bool {
	trimmedBody := strings.TrimSpace(body)
	return trimmedBody == strings.TrimSpace(reviewCommand) || strings.Contains(body, codexReviewRequestMarker)
}

func buildReviewRequestCommentBody(reviewCommand string) string {
	return codexReviewRequestMarker + "\n" + strings.TrimSpace(reviewCommand)
}

func (a *App) resolvePullRequestForLand(ctx context.Context, root string, opts landOptions, currentBranch, baseBranch string) (githubPullRequest, error) {
	if opts.Number > 0 {
		return a.loadPullRequest(ctx, root, strconv.Itoa(opts.Number), landPullRequestFields()...)
	}
	if currentBranch == baseBranch {
		return githubPullRequest{}, fmt.Errorf("land requires a pull request number when current branch is %q", baseBranch)
	}
	pullRequests, err := a.listOpenPullRequests(ctx, root, currentBranch, baseBranch)
	if err != nil {
		return githubPullRequest{}, err
	}
	if len(pullRequests) == 0 {
		return githubPullRequest{}, fmt.Errorf("no open pull request found for branch %q", currentBranch)
	}
	return a.loadPullRequest(ctx, root, strconv.Itoa(pullRequests[0].Number), landPullRequestFields()...)
}

func validateLandPullRequest(pr githubPullRequest, waited bool) error {
	if len(pr.StatusChecks) > 0 {
		pending, failed := classifyStatusChecks(pr.StatusChecks)
		if len(failed) > 0 {
			return fmt.Errorf("pull request #%d has failing checks: %s", pr.Number, strings.Join(failed, ", "))
		}
		if len(pending) > 0 {
			message := fmt.Sprintf("pull request #%d has pending checks: %s", pr.Number, strings.Join(pending, ", "))
			if !waited {
				message += " (rerun with --wait to wait for completion)"
			}
			return errors.New(message)
		}
	}

	switch strings.ToUpper(strings.TrimSpace(pr.ReviewDecision)) {
	case "CHANGES_REQUESTED":
		return fmt.Errorf("pull request #%d is blocked by requested changes", pr.Number)
	case "REVIEW_REQUIRED":
		return fmt.Errorf("pull request #%d is waiting for approval", pr.Number)
	}

	if strings.ToUpper(strings.TrimSpace(pr.MergeStateStatus)) != "CLEAN" {
		return fmt.Errorf("pull request #%d is not mergeable: merge state is %s", pr.Number, strings.ToLower(strings.TrimSpace(pr.MergeStateStatus)))
	}
	return nil
}

func classifyStatusChecks(checks []githubStatusCheck) ([]string, []string) {
	pending := make([]string, 0, len(checks))
	failed := make([]string, 0, len(checks))
	for _, check := range checks {
		name := check.displayName()
		status := strings.ToUpper(strings.TrimSpace(check.Status))
		state := strings.ToUpper(strings.TrimSpace(check.State))
		conclusion := strings.ToUpper(strings.TrimSpace(check.Conclusion))

		switch {
		case isFailingCheckState(conclusion) || isFailingCheckState(state):
			failed = append(failed, name)
		case isPassingCheckState(conclusion) || isPassingCheckState(state):
		case status == "QUEUED" || status == "IN_PROGRESS" || state == "PENDING" || state == "EXPECTED":
			pending = append(pending, name)
		case status == "COMPLETED" && conclusion == "":
			pending = append(pending, name)
		}
	}
	return uniqueStrings(pending), uniqueStrings(failed)
}

func isPassingCheckState(value string) bool {
	switch value {
	case "SUCCESS", "NEUTRAL", "SKIPPED":
		return true
	default:
		return false
	}
}

func isFailingCheckState(value string) bool {
	switch value {
	case "FAILURE", "ERROR", "TIMED_OUT", "ACTION_REQUIRED", "CANCELLED":
		return true
	default:
		return false
	}
}

func (check githubStatusCheck) displayName() string {
	return firstNonBlank(check.WorkflowName, check.Name, check.Context, "unnamed-check")
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func (a *App) listOpenPullRequests(ctx context.Context, root, headBranch, baseBranch string) ([]githubPullRequest, error) {
	out, err := a.runBinary(ctx, "gh", []string{"pr", "list", "--head", headBranch, "--base", baseBranch, "--state", "open", "--json", strings.Join(reviewPullRequestFields(), ",")}, root)
	if err != nil {
		return nil, fmt.Errorf("list pull requests for branch %s: %w", headBranch, err)
	}
	var pullRequests []githubPullRequest
	if err := json.Unmarshal([]byte(firstNonBlank(out, "[]")), &pullRequests); err != nil {
		return nil, fmt.Errorf("parse pull request list: %w", err)
	}
	return pullRequests, nil
}

func (a *App) loadPullRequest(ctx context.Context, root, selector string, fields ...string) (githubPullRequest, error) {
	out, err := a.runBinary(ctx, "gh", []string{"pr", "view", selector, "--json", strings.Join(fields, ",")}, root)
	if err != nil {
		return githubPullRequest{}, fmt.Errorf("load pull request %s: %w", selector, err)
	}
	var pr githubPullRequest
	if err := json.Unmarshal([]byte(out), &pr); err != nil {
		return githubPullRequest{}, fmt.Errorf("parse pull request %s: %w", selector, err)
	}
	return pr, nil
}

func reviewPullRequestFields() []string {
	return []string{"number", "url", "title", "headRefName", "baseRefName"}
}

func landPullRequestFields() []string {
	return []string{"number", "url", "title", "headRefName", "baseRefName", "reviewDecision", "mergeStateStatus", "isDraft", "statusCheckRollup"}
}

func worktreeListArgs() []string {
	return []string{"worktree", "list", "--porcelain"}
}

func (a *App) worktreeForBranch(ctx context.Context, root, branch string) (gitWorktree, bool, error) {
	out, err := a.runBinary(ctx, "git", worktreeListArgs(), root)
	if err != nil {
		return gitWorktree{}, false, fmt.Errorf("list git worktrees: %w", err)
	}
	for _, worktree := range parseGitWorktrees(out) {
		if worktree.Branch == branch {
			return worktree, true, nil
		}
	}
	return gitWorktree{}, false, nil
}

func parseGitWorktrees(out string) []gitWorktree {
	lines := strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n")
	worktrees := make([]gitWorktree, 0)
	current := gitWorktree{}
	flush := func() {
		if current.Path == "" {
			current = gitWorktree{}
			return
		}
		current.Path = filepath.Clean(current.Path)
		worktrees = append(worktrees, current)
		current = gitWorktree{}
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		switch {
		case strings.HasPrefix(line, "worktree "):
			current.Path = strings.TrimSpace(strings.TrimPrefix(line, "worktree "))
		case strings.HasPrefix(line, "branch "):
			current.Branch = strings.TrimPrefix(strings.TrimSpace(strings.TrimPrefix(line, "branch ")), "refs/heads/")
		}
	}
	flush()
	return worktrees
}

func (a *App) updateLocalBaseBranch(ctx context.Context, root, currentBranch, baseBranch, remote string) error {
	if _, err := a.runBinary(ctx, "git", []string{"fetch", remote, baseBranch}, root); err != nil {
		return fmt.Errorf("fetch %s/%s: %w", remote, baseBranch, err)
	}
	if currentBranch == baseBranch {
		dirty, err := a.hasWorkingTreeChanges(ctx, root)
		if err != nil {
			return err
		}
		if dirty {
			return fmt.Errorf("cannot update local %s with uncommitted changes in the working tree", baseBranch)
		}
		if _, err := a.runBinary(ctx, "git", []string{"pull", "--ff-only", remote, baseBranch}, root); err != nil {
			return fmt.Errorf("fast-forward local %s: %w", baseBranch, err)
		}
		return nil
	}

	worktree, found, err := a.worktreeForBranch(ctx, root, baseBranch)
	if err != nil {
		return err
	}
	if found {
		dirty, err := a.hasWorkingTreeChanges(ctx, worktree.Path)
		if err != nil {
			return err
		}
		if dirty {
			return fmt.Errorf("cannot update local %s in worktree %s with uncommitted changes", baseBranch, worktree.Path)
		}
		if _, err := a.runBinary(ctx, "git", []string{"pull", "--ff-only", remote, baseBranch}, worktree.Path); err != nil {
			return fmt.Errorf("fast-forward local %s in worktree %s: %w", baseBranch, worktree.Path, err)
		}
		return nil
	}

	remoteRef := fmt.Sprintf("%s/%s", remote, baseBranch)
	if _, err := a.runBinary(ctx, "git", []string{"branch", "-f", baseBranch, remoteRef}, root); err != nil {
		return fmt.Errorf("update local %s from %s/%s: %w", baseBranch, remote, baseBranch, err)
	}
	return nil
}

func buildPullRequestBody(profile initProfile) string {
	summaryPath := filepath.ToSlash(filepath.Join(projectDir, "change-summary.md"))
	checklistPath := filepath.ToSlash(filepath.Join(projectDir, "pr-checklist.md"))

	switch normalizeReadmeLanguage(profile.PRLanguage) {
	case "ko":
		return strings.Join([]string{
			"## \uC791\uC5C5 \uC694\uC57D",
			fmt.Sprintf("- \uBCC0\uACBD \uC694\uC57D: `%s`", summaryPath),
			fmt.Sprintf("- \uAC80\uD1A0 \uCCB4\uD06C\uB9AC\uC2A4\uD2B8: `%s`", checklistPath),
			"",
			"## \uAC80\uD1A0 \uBA54\uBAA8",
			"- `namba pr`\uAC00 sync, validation, commit, push\uB97C \uB9C8\uCE5C \uC0C1\uD0DC\uC785\uB2C8\uB2E4.",
		}, "\n")
	case "ja":
		return strings.Join([]string{
			"## \u4F5C\u696D\u6982\u8981",
			fmt.Sprintf("- \u5909\u66F4\u30B5\u30DE\u30EA\u30FC: `%s`", summaryPath),
			fmt.Sprintf("- \u30EC\u30D3\u30E5\u30FC\u30C1\u30A7\u30C3\u30AF\u30EA\u30B9\u30C8: `%s`", checklistPath),
			"",
			"## \u30EC\u30D3\u30E5\u30FC\u7528\u30E1\u30E2",
			"- `namba pr` \u306F sync\u3001validation\u3001commit\u3001push \u307E\u3067\u5B8C\u4E86\u3057\u3066\u3044\u307E\u3059\u3002",
		}, "\n")
	case "zh":
		return strings.Join([]string{
			"## \u53D8\u66F4\u6458\u8981",
			fmt.Sprintf("- \u53D8\u66F4\u8BF4\u660E\uFF1A`%s`", summaryPath),
			fmt.Sprintf("- \u8BC4\u5BA1\u6E05\u5355\uFF1A`%s`", checklistPath),
			"",
			"## \u8BC4\u5BA1\u5907\u6CE8",
			"- `namba pr` \u5DF2\u5B8C\u6210 sync\u3001validation\u3001commit \u548C push\u3002",
		}, "\n")
	default:
		return strings.Join([]string{
			"## Summary",
			fmt.Sprintf("- Change summary: `%s`", summaryPath),
			fmt.Sprintf("- Review checklist: `%s`", checklistPath),
			"",
			"## Review Notes",
			"- `namba pr` has already completed sync, validation, commit, and push.",
		}, "\n")
	}
}
