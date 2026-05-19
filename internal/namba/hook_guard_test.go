package namba

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNambaCodexGuardPromptRefinement(t *testing.T) {
	t.Parallel()

	ambiguous := runNambaCodexGuard(t, map[string]any{
		"hook_event_name": "UserPromptSubmit",
		"prompt":          "namba plan improve this",
	})
	if !strings.Contains(ambiguous, "prompt-refinement gate") {
		t.Fatalf("expected ambiguous prompt guidance, got %q", ambiguous)
	}

	clear := runNambaCodexGuard(t, map[string]any{
		"hook_event_name": "UserPromptSubmit",
		"prompt":          "Implement the CLI parser for --review in namba pr. Scope: PR command only. Acceptance: parser test passes and unknown flags still fail. Constraints: keep existing output contract.",
	})
	if strings.TrimSpace(clear) != "" {
		t.Fatalf("expected clear prompt to avoid extra guidance, got %q", clear)
	}
}

func TestNambaCodexGuardBashCommandPolicy(t *testing.T) {
	t.Parallel()

	dangerous := runNambaCodexGuard(t, map[string]any{
		"hook_event_name": "PreToolUse",
		"tool_input": map[string]any{
			"cmd": "git reset --hard HEAD",
		},
	})
	if !strings.Contains(dangerous, `"permissionDecision":"deny"`) || !strings.Contains(dangerous, "git reset --hard") {
		t.Fatalf("expected dangerous command denial, got %q", dangerous)
	}

	safe := runNambaCodexGuard(t, map[string]any{
		"hook_event_name": "PreToolUse",
		"tool_input": map[string]any{
			"cmd": "git status --short",
		},
	})
	if strings.TrimSpace(safe) != "" {
		t.Fatalf("expected safe command to be allowed silently, got %q", safe)
	}
}

func runNambaCodexGuard(t *testing.T, payload map[string]any) string {
	t.Helper()

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve hook guard test path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	cmd := exec.Command("python3", filepath.Join(repoRoot, ".codex", "hooks", "namba_codex_guard.py"))
	cmd.Stdin = bytes.NewReader(data)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run guard: %v\n%s", err, output)
	}
	return string(output)
}
