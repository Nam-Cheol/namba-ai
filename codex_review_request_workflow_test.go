package namba_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexReviewRequestWorkflowUsesWritablePullRequestTarget(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(".github", "workflows", "codex-review-request.yml"))
	if err != nil {
		t.Fatalf("read codex review request workflow: %v", err)
	}

	text := string(data)
	for _, want := range []string{
		"pull_request_target:",
		"issues: write",
		"pull-requests: write",
		"ready_for_review",
		"github.event.pull_request.base.ref == 'main'",
		"github.event.pull_request.user.login != 'dependabot[bot]'",
		"const marker = \"<!-- namba-codex-review-request -->\";",
		"comment.body.includes(marker)",
		"github.rest.issues.createComment",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected workflow to contain %q, got:\n%s", want, text)
		}
	}
}
