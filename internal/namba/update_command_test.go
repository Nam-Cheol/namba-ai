package namba

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunUpdateRegeneratesCodexAssetsFromConfig(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmp, ".namba", "config", "sections", "codex.yaml"), []byte("agent_mode: multi\nstatus_line_preset: off\n"), 0o644); err != nil {
		t.Fatalf("write codex config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "AGENTS.md"), []byte("stale\n"), 0o644); err != nil {
		t.Fatalf("write AGENTS: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"update"}); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	agents := mustReadFile(t, filepath.Join(tmp, "AGENTS.md"))
	if !strings.Contains(agents, "Agent mode: multi") {
		t.Fatalf("expected regenerated AGENTS to reflect config, got %q", agents)
	}

	config := mustReadFile(t, filepath.Join(tmp, ".codex", "config.toml"))
	if !strings.Contains(config, "max_threads = 3") {
		t.Fatalf("expected multi-agent Codex config, got %q", config)
	}
	if strings.Contains(config, "status_line =") {
		t.Fatalf("expected status line preset off to omit status line, got %q", config)
	}
}
