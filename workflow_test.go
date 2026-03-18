package namba_test

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexReviewWorkflowSkipsDependabotPullRequests(t *testing.T) {
	workflow := mustRead(t, filepath.Join(".github", "workflows", "codex-review-request.yml"))

	if !strings.Contains(workflow, "pull-requests: write") {
		t.Fatalf("expected workflow to request pull-requests: write for PR comment creation, got: %s", workflow)
	}

	if !strings.Contains(workflow, "github.event.pull_request.user.login != 'dependabot[bot]'") {
		t.Fatalf("expected workflow to skip Dependabot PRs to avoid read-only token 403 failures, got: %s", workflow)
	}

	if !strings.Contains(workflow, "github.rest.issues.createComment") {
		t.Fatalf("expected workflow to keep creating Codex review comments for eligible PRs, got: %s", workflow)
	}
}
