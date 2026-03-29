package namba

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCodexSmokeRun(t *testing.T) {
	if os.Getenv("CODEX_SMOKE") != "1" {
		t.Skip("set CODEX_SMOKE=1 to run the live Codex smoke suite")
	}
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skipf("codex not available: %v", err)
	}

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"plan", "create", "a", "smoke", "file"}); err != nil {
		t.Fatalf("plan failed: %v", err)
	}

	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "quality.yaml"), "development_mode: tdd\ntest_command: none\nlint_command: none\ntypecheck_command: none\nbuild_command: none\nmigration_dry_run_command: none\nsmoke_start_command: none\noutput_contract_command: none\n")
	writeTestFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "spec.md"), "# SPEC-001\n\n## Goal\n\nCreate a file named `CODEX_SMOKE.txt` containing `ok`.\n")
	writeTestFile(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "acceptance.md"), "# Acceptance\n\n- [ ] Create `CODEX_SMOKE.txt` with the exact content `ok`\n- [ ] Validation commands pass\n")

	if err := app.Run(context.Background(), []string{"run", "SPEC-001", "--solo"}); err != nil {
		t.Fatalf("smoke run failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "CODEX_SMOKE.txt")); err != nil {
		t.Fatalf("expected smoke output file: %v", err)
	}
}
