package namba_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoGitHubWorkflowAutomaticallyRequestsCodexReview(t *testing.T) {
	if _, err := os.Stat(filepath.Join(".github", "workflows", "codex-review-request.yml")); !os.IsNotExist(err) {
		t.Fatalf("codex review request workflow must be removed, stat err=%v", err)
	}

	entries, err := os.ReadDir(filepath.Join(".github", "workflows"))
	if err != nil {
		t.Fatalf("read workflows: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}
		path := filepath.Join(".github", "workflows", entry.Name())
		workflow := mustRead(t, path)
		if strings.Contains(workflow, "@codex review") || strings.Contains(workflow, "namba:codex-review-request") || strings.Contains(workflow, "github.rest.issues.createComment") {
			t.Fatalf("workflow %s must not automatically request Codex review, got:\n%s", path, workflow)
		}
	}
}
