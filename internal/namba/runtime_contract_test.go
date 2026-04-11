package namba

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewExecutionRequestAppliesModeRuntimeContract(t *testing.T) {
	t.Parallel()

	app := NewApp(nil, nil)
	systemCfg := systemConfig{
		Runner:         "codex",
		ApprovalPolicy: "on-request",
		SandboxMode:    "workspace-write",
	}
	codexCfg := codexConfig{
		Model:          "gpt-5.4",
		Profile:        "namba",
		AddDirs:        []string{"extra"},
		SessionMode:    "stateful",
		RepairAttempts: 2,
		RequiredEnv:    []string{"OPENAI_API_KEY"},
	}

	tests := []struct {
		name string
		mode executionMode
		want string
	}{
		{name: "default", mode: executionModeDefault, want: "stateful"},
		{name: "solo", mode: executionModeSolo, want: "solo"},
		{name: "team", mode: executionModeTeam, want: "team"},
		{name: "parallel", mode: executionModeParallel, want: "parallel-worker"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := app.newExecutionRequest("SPEC-027", "/tmp/work", "prompt", tt.mode, delegationPlan{IntegratorRole: "standalone-runner"}, systemCfg, codexCfg)
			if req.Mode != tt.mode {
				t.Fatalf("mode = %q, want %q", req.Mode, tt.mode)
			}
			if req.SessionMode != tt.want {
				t.Fatalf("session mode = %q, want %q", req.SessionMode, tt.want)
			}
			if req.Model != "gpt-5.4" || req.Profile != "namba" {
				t.Fatalf("expected runtime config to propagate, got %+v", req)
			}
			if req.RepairAttempts != 2 || len(req.RequiredEnv) != 1 || req.RequiredEnv[0] != "OPENAI_API_KEY" {
				t.Fatalf("expected repair/runtime contract fields to propagate, got %+v", req)
			}
		})
	}
}

func TestResolveRuntimeAddDirsNormalizesAndDeduplicates(t *testing.T) {
	t.Parallel()

	root := canonicalTempDir(t)
	nested := filepath.Join(root, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	resolved, err := resolveRuntimeAddDirs(root, []string{"nested", "./nested", nested})
	if err != nil {
		t.Fatalf("resolveRuntimeAddDirs: %v", err)
	}
	if len(resolved) != 1 || resolved[0] != nested {
		t.Fatalf("unexpected resolved add dirs: %#v", resolved)
	}
}
