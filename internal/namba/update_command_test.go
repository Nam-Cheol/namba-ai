package namba

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRegenRegeneratesCodexAssetsFromConfig(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmp, ".namba", "config", "sections", "codex.yaml"), []byte("agent_mode: multi\nstatus_line_preset: off\n"), 0o644); err != nil {
		t.Fatalf("write codex config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".namba", "config", "sections", "git-strategy.yaml"), []byte("git_mode: team\ngit_provider: github\ngit_username: alice\ngitlab_instance_url: https://gitlab.com\nstore_tokens: false\nbranch_per_work: true\nbranch_base: develop\nspec_branch_prefix: spec/\ntask_branch_prefix: task/\npr_base_branch: develop\npr_language: ko\ncodex_review_comment: \"@codex review\"\nauto_codex_review: true\n"), 0o644); err != nil {
		t.Fatalf("write git strategy config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".namba", "config", "sections", "system.yaml"), []byte("runner: codex\napproval_policy: never\nsandbox_mode: read-only\n"), 0o644); err != nil {
		t.Fatalf("write system config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".codex", "skills", "namba"), 0o755); err != nil {
		t.Fatalf("mkdir stale compat skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".codex", "skills", "namba", "SKILL.md"), []byte("stale compat skill\n"), 0o644); err != nil {
		t.Fatalf("write stale compat skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "AGENTS.md"), []byte("stale\n"), 0o644); err != nil {
		t.Fatalf("write AGENTS: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"regen"}); err != nil {
		t.Fatalf("regen failed: %v", err)
	}

	agents := mustReadFile(t, filepath.Join(tmp, "AGENTS.md"))
	if !strings.Contains(agents, "Agent mode: multi") {
		t.Fatalf("expected regenerated AGENTS to reflect config, got %q", agents)
	}
	if !strings.Contains(agents, "dedicated branch from `develop`") || !strings.Contains(agents, "`@codex review`") {
		t.Fatalf("expected regenerated AGENTS to reflect git collaboration policy, got %q", agents)
	}

	runSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-run", "SKILL.md"))
	if !strings.Contains(runSkill, "$namba-run") || !strings.Contains(runSkill, "namba run SPEC-XXX") {
		t.Fatalf("expected command-entry run skill, got %q", runSkill)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".codex", "skills")); !os.IsNotExist(err) {
		t.Fatalf("expected deprecated codex skills mirror to be removed, stat err=%v", err)
	}

	config := mustReadFile(t, filepath.Join(tmp, ".codex", "config.toml"))
	if !strings.Contains(config, "max_threads = 3") || !strings.Contains(config, `approval_policy = "never"`) || !strings.Contains(config, `sandbox_mode = "read-only"`) {
		t.Fatalf("expected multi-agent Codex config, got %q", config)
	}
	if strings.Contains(config, "status_line =") {
		t.Fatalf("expected status line preset off to omit status line, got %q", config)
	}

	codexReadme := mustReadFile(t, filepath.Join(tmp, ".namba", "codex", "README.md"))
	if !strings.Contains(codexReadme, "`namba regen` regenerates") || !strings.Contains(codexReadme, "`namba update` self-updates") {
		t.Fatalf("expected codex README to describe regen/update semantics, got %q", codexReadme)
	}
	if !strings.Contains(codexReadme, ".codex/agents/*.toml") {
		t.Fatalf("expected codex README to describe custom agents, got %q", codexReadme)
	}
	if !strings.Contains(codexReadme, "$namba-run") || strings.Contains(codexReadme, ".codex/skills/") {
		t.Fatalf("expected codex README to describe command-entry skills without codex skill mirror, got %q", codexReadme)
	}
	if !strings.Contains(codexReadme, "Namba closing frame") || !strings.Contains(codexReadme, "validate-output-contract.py") {
		t.Fatalf("expected codex README to describe output contract fallback validation, got %q", codexReadme)
	}
	if !strings.Contains(codexReadme, "PR titles and bodies should be written in Korean") || !strings.Contains(codexReadme, "`@codex review`") {
		t.Fatalf("expected codex README to describe PR collaboration defaults, got %q", codexReadme)
	}
	if !strings.Contains(codexReadme, "`namba run SPEC-XXX --parallel`") {
		t.Fatalf("expected codex README to describe standalone parallel semantics, got %q", codexReadme)
	}

	outputContractDoc := mustReadFile(t, filepath.Join(tmp, ".namba", "codex", "output-contract.md"))
	if !strings.Contains(outputContractDoc, "오늘의 결정") || !strings.Contains(outputContractDoc, "무너지는 조건") {
		t.Fatalf("expected output contract doc, got %q", outputContractDoc)
	}

	validator := mustReadFile(t, filepath.Join(tmp, ".namba", "codex", "validate-output-contract.py"))
	if !strings.Contains(validator, "output-contract: ok") || !strings.Contains(validator, "판단 근거") {
		t.Fatalf("expected output contract validator, got %q", validator)
	}

	customAgent := mustReadFile(t, filepath.Join(tmp, ".codex", "agents", "namba-planner.toml"))
	if !strings.Contains(customAgent, `name = "namba-planner"`) || !strings.Contains(customAgent, `developer_instructions = """`) {
		t.Fatalf("expected regenerated custom agent TOML, got %q", customAgent)
	}

	roleCard := mustReadFile(t, filepath.Join(tmp, ".codex", "agents", "namba-planner.md"))
	if !strings.Contains(roleCard, "# Namba Planner") {
		t.Fatalf("expected readable role-card mirror, got %q", roleCard)
	}
}
