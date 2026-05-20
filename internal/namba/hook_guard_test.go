package namba

import (
	"bytes"
	"encoding/json"
	"os"
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
		"cwd":             repoRootForHookTest(t),
	})
	ambiguousOutput := hookSpecificOutput(t, parseGuardJSON(t, ambiguous))
	guidance, ok := ambiguousOutput["additionalContext"].(string)
	if !ok || !strings.Contains(guidance, "프롬프트 교정 게이트") || !strings.Contains(guidance, "대상 surface") {
		t.Fatalf("expected configured-language ambiguous prompt guidance, got %q", ambiguous)
	}

	clear := runNambaCodexGuard(t, map[string]any{
		"hook_event_name": "UserPromptSubmit",
		"prompt":          "Implement the CLI parser for --review in namba pr. Scope: PR command only. Acceptance: parser test passes and unknown flags still fail. Constraints: keep existing output contract.",
		"cwd":             repoRootForHookTest(t),
	})
	if parsed := parseGuardJSON(t, clear); parsed["continue"] != true {
		t.Fatalf("expected clear prompt to emit explicit continue JSON, got %q", clear)
	}
}

func TestNambaCodexGuardPromptRefinementUsesConfiguredJapaneseAndChinese(t *testing.T) {
	t.Parallel()

	cases := []struct {
		language string
		want     string
	}{
		{language: "ja", want: "SPEC を作る前に"},
		{language: "zh", want: "创建 SPEC 前"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.language, func(t *testing.T) {
			t.Parallel()
			tmp := t.TempDir()
			writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "language.yaml"), "conversation_language: "+tc.language+"\ndocumentation_language: "+tc.language+"\ncomment_language: "+tc.language+"\n")
			ambiguous := runNambaCodexGuard(t, map[string]any{
				"hook_event_name": "UserPromptSubmit",
				"prompt":          "namba plan improve this",
				"cwd":             tmp,
			})
			hookOutput := hookSpecificOutput(t, parseGuardJSON(t, ambiguous))
			guidance, ok := hookOutput["additionalContext"].(string)
			if !ok || !strings.Contains(guidance, tc.want) {
				t.Fatalf("expected %s prompt guidance to contain %q, got %q", tc.language, tc.want, ambiguous)
			}
		})
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
			"command": "git status --short",
		},
	})
	if parsed := parseGuardJSON(t, safe); parsed["continue"] != true {
		t.Fatalf("expected safe command to emit explicit continue JSON, got %q", safe)
	}
}

func TestNambaCodexGuardPreToolUseUpdatedInputPrecedence(t *testing.T) {
	t.Parallel()

	rewrite := runNambaCodexGuard(t, map[string]any{
		"hook_event_name": "PreToolUse",
		"tool_input": map[string]any{
			"cmd": "git status --short",
		},
	})
	hookOutput := hookSpecificOutput(t, parseGuardJSON(t, rewrite))
	if hookOutput["permissionDecision"] != "allow" {
		t.Fatalf("expected safe cmd alias rewrite to allow the command, got %q", rewrite)
	}
	updatedInput, ok := hookOutput["updatedInput"].(map[string]any)
	if !ok || updatedInput["command"] != "git status --short" {
		t.Fatalf("expected updatedInput to rewrite cmd into command, got %q", rewrite)
	}

	denied := runNambaCodexGuard(t, map[string]any{
		"hook_event_name": "PreToolUse",
		"tool_input": map[string]any{
			"cmd": "git reset --hard HEAD",
		},
	})
	deniedOutput := hookSpecificOutput(t, parseGuardJSON(t, denied))
	if deniedOutput["permissionDecision"] != "deny" {
		t.Fatalf("expected deny to win over rewrite, got %q", denied)
	}
	if _, ok := deniedOutput["updatedInput"]; ok {
		t.Fatalf("dangerous command denial must not include updatedInput, got %q", denied)
	}
}

func TestNambaCodexGuardSixHookEventsEmitValidJSON(t *testing.T) {
	t.Parallel()

	longBadReport := strings.Repeat("namba codex spec-050 missing frame ", 30)
	payloads := []map[string]any{
		{"hook_event_name": "SessionStart", "source": "startup"},
		{"hook_event_name": "SessionStart", "session_id": "session-123", "thread-id": "thread-legacy"},
		{"hook_event_name": "SessionStart"},
		{"hook_event_name": "UserPromptSubmit", "prompt": ""},
		{"hook_event_name": "UserPromptSubmit", "prompt": "namba plan \"quoted\"\nmultiline improve"},
		{"hook_event_name": "PermissionRequest", "tool_input": map[string]any{"command": "sudo true"}},
		{"hook_event_name": "PermissionRequest", "tool_input": map[string]any{"command": "git reset --hard HEAD"}},
		{"hook_event_name": "PostToolUse", "tool_response": map[string]any{"status": "ok"}},
		{"hook_event_name": "PostToolUse", "tool_response": map[string]any{"status": "failed", "error": "boom"}},
		{"hook_event_name": "PostToolUse"},
		{"hook_event_name": "Stop", "last_assistant_message": "short"},
		{"hook_event_name": "Stop", "last_assistant_message": longBadReport, "cwd": repoRootForHookTest(t)},
	}

	for _, payload := range payloads {
		payload := payload
		t.Run(payload["hook_event_name"].(string), func(t *testing.T) {
			t.Parallel()

			output := runNambaCodexGuard(t, payload)
			parseGuardJSON(t, output)
		})
	}
}

func TestNambaCodexGuardMalformedPayloadEmitsValidJSON(t *testing.T) {
	t.Parallel()

	output := runNambaCodexGuardRaw(t, `{"hook_event_name":"PreToolUse","tool_input":`)
	parsed := parseGuardJSON(t, output)
	if parsed["continue"] != true {
		t.Fatalf("expected malformed input to fail open with valid continue JSON, got %q", output)
	}
	if !strings.Contains(output, "malformed JSON") {
		t.Fatalf("expected malformed payload diagnostic, got %q", output)
	}
}

func TestNambaCodexGuardSuppressesDuplicateNambaHookPayloads(t *testing.T) {
	t.Parallel()

	dedupeDir := t.TempDir()
	payload := map[string]any{
		"hook_event_name": "PreToolUse",
		"session_id":      "session-duplicate",
		"tool_use_id":     "tool-1",
		"tool_input": map[string]any{
			"command": "git status --short",
		},
	}
	dedupeEnv := []string{"NAMBA_HOOK_DEDUPE=1", "NAMBA_HOOK_DEDUPE_DIR=" + dedupeDir}
	first := runNambaCodexGuardWithEnv(t, payload, dedupeEnv)
	if parsed := parseGuardJSON(t, first); parsed["continue"] != true || parsed["suppressOutput"] == true {
		t.Fatalf("expected first hook payload to continue normally, got %q", first)
	}

	second := runNambaCodexGuardWithEnv(t, payload, dedupeEnv)
	if parsed := parseGuardJSON(t, second); parsed["continue"] != true || parsed["suppressOutput"] != true {
		t.Fatalf("expected duplicate hook payload to be suppressed with valid JSON, got %q", second)
	}
}

func runNambaCodexGuard(t *testing.T, payload map[string]any) string {
	t.Helper()

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return runNambaCodexGuardRaw(t, string(data))
}

func runNambaCodexGuardWithEnv(t *testing.T, payload map[string]any, env []string) string {
	t.Helper()

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return runNambaCodexGuardRawWithEnv(t, string(data), env)
}

func runNambaCodexGuardRaw(t *testing.T, payload string) string {
	t.Helper()

	return runNambaCodexGuardRawWithEnv(t, payload, nil)
}

func runNambaCodexGuardRawWithEnv(t *testing.T, payload string, env []string) string {
	t.Helper()

	repoRoot := repoRootForHookTest(t)
	cmd := exec.Command("python3", filepath.Join(repoRoot, ".codex", "hooks", "namba_codex_guard.py"))
	cmd.Stdin = bytes.NewReader([]byte(payload))
	cmd.Env = append(os.Environ(), "NAMBA_HOOK_DEDUPE=0")
	cmd.Env = append(cmd.Env, env...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run guard: %v\n%s", err, output)
	}
	return string(output)
}

func repoRootForHookTest(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve hook guard test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func parseGuardJSON(t *testing.T, output string) map[string]any {
	t.Helper()

	var parsed map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); err != nil {
		t.Fatalf("hook output is not valid JSON: %v\n%s", err, output)
	}
	return parsed
}

func hookSpecificOutput(t *testing.T, parsed map[string]any) map[string]any {
	t.Helper()

	hookOutput, ok := parsed["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatalf("expected hookSpecificOutput object, got %+v", parsed)
	}
	return hookOutput
}
