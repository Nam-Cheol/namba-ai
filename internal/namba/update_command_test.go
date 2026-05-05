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
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes", "--human-language", "ko"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmp, ".namba", "config", "sections", "codex.yaml"), []byte("agent_mode: multi\nstatus_line_preset: off\ndefault_mcp_servers: context7,sequential-thinking,playwright\n"), 0o644); err != nil {
		t.Fatalf("write codex config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".namba", "config", "sections", "git-strategy.yaml"), []byte("git_mode: team\ngit_provider: github\ngit_username: alice\ngitlab_instance_url: https://gitlab.com\nstore_tokens: false\nbranch_per_work: true\nbranch_base: develop\nspec_branch_prefix: spec/\ntask_branch_prefix: task/\npr_base_branch: develop\npr_language: ko\ncodex_review_comment: \"@codex review\"\nauto_codex_review: true\n"), 0o644); err != nil {
		t.Fatalf("write git strategy config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".namba", "config", "sections", "system.yaml"), []byte("runner: codex\napproval_policy: never\nsandbox_mode: read-only\n"), 0o644); err != nil {
		t.Fatalf("write system config: %v", err)
	}
	legacySkillPath := filepath.Join(tmp, ".codex", "skills", "namba", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(legacySkillPath), 0o755); err != nil {
		t.Fatalf("mkdir stale compat skill: %v", err)
	}
	if err := os.WriteFile(legacySkillPath, []byte("stale compat skill\n"), 0o644); err != nil {
		t.Fatalf("write stale compat skill: %v", err)
	}
	customSkillPath := filepath.Join(tmp, ".codex", "skills", "custom-skill", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(customSkillPath), 0o755); err != nil {
		t.Fatalf("mkdir custom skill: %v", err)
	}
	if err := os.WriteFile(customSkillPath, []byte("user custom skill\n"), 0o644); err != nil {
		t.Fatalf("write custom skill: %v", err)
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
	for _, want := range []string{"$namba-queue", "$namba-review-resolve", "$namba-release", "namba queue", "namba release"} {
		if !strings.Contains(agents, want) {
			t.Fatalf("expected regenerated AGENTS to contain %q, got %q", want, agents)
		}
	}

	runSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-run", "SKILL.md"))
	if !strings.Contains(runSkill, "$namba-run") || !strings.Contains(runSkill, "namba run SPEC-XXX") {
		t.Fatalf("expected command-entry run skill, got %q", runSkill)
	}
	if !strings.Contains(runSkill, "`--solo`, `--team`, `--parallel`, `--dry-run`") {
		t.Fatalf("expected run skill to describe standalone run modes, got %q", runSkill)
	}
	for _, want := range []string{"`namba-frontend-architect`", "generic-section redesign"} {
		if !strings.Contains(runSkill, want) {
			t.Fatalf("expected run skill to contain %q, got %q", want, runSkill)
		}
	}
	queueSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-queue", "SKILL.md"))
	for _, want := range []string{"$namba-queue", "namba queue start <SPEC-RANGE|SPEC-LIST>", "`SPEC-001..SPEC-003`", "`namba queue status [--verbose]`", ".namba/logs/queue/", "one active SPEC at a time", "Block instead of skipping", "waiting_for_land"} {
		if !strings.Contains(queueSkill, want) {
			t.Fatalf("expected queue skill to contain %q, got %q", want, queueSkill)
		}
	}
	pmReviewSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-plan-pm-review", "SKILL.md"))
	if !strings.Contains(pmReviewSkill, "$namba-plan-pm-review") || !strings.Contains(pmReviewSkill, "reviews/product.md") || !strings.Contains(pmReviewSkill, "readiness.md") {
		t.Fatalf("expected product review skill, got %q", pmReviewSkill)
	}
	engReviewSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-plan-eng-review", "SKILL.md"))
	if !strings.Contains(engReviewSkill, "$namba-plan-eng-review") || !strings.Contains(engReviewSkill, "`namba-planner`") || !strings.Contains(engReviewSkill, "advisory") {
		t.Fatalf("expected engineering review skill, got %q", engReviewSkill)
	}
	designReviewSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-plan-design-review", "SKILL.md"))
	if !strings.Contains(designReviewSkill, "$namba-plan-design-review") || !strings.Contains(designReviewSkill, "`namba-designer`") || !strings.Contains(designReviewSkill, "readiness.md") {
		t.Fatalf("expected design review skill, got %q", designReviewSkill)
	}
	for _, want := range []string{"art-direction clarity", "washed-out gray minimalism", "meaningless fallback use"} {
		if !strings.Contains(designReviewSkill, want) {
			t.Fatalf("expected design review skill to contain %q, got %q", want, designReviewSkill)
		}
	}
	planSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-plan", "SKILL.md"))
	if !strings.Contains(planSkill, "context7") || !strings.Contains(planSkill, "sequential-thinking") || !strings.Contains(planSkill, "playwright") || !strings.Contains(planSkill, "repo-managed MCP presets") {
		t.Fatalf("expected plan skill to describe managed MCP usage, got %q", planSkill)
	}
	if !strings.Contains(planSkill, "namba harness") {
		t.Fatalf("expected plan skill to distinguish harness-oriented planning, got %q", planSkill)
	}
	planReviewSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-plan-review", "SKILL.md"))
	for _, want := range []string{"$namba-plan-review", "parallel", "namba-plan-reviewer", "aggregate validation", "readiness.md"} {
		if !strings.Contains(planReviewSkill, want) {
			t.Fatalf("expected plan-review skill to contain %q, got %q", want, planReviewSkill)
		}
	}
	helpSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-help", "SKILL.md"))
	for _, want := range []string{"$namba-help", "read-only", "README*.md", "which command or skill to use next", "Do not mutate repository state"} {
		if !strings.Contains(helpSkill, want) {
			t.Fatalf("expected help skill to contain %q, got %q", want, helpSkill)
		}
	}
	createSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-create", "SKILL.md"))
	for _, want := range []string{"$namba-create", "skill-first", "`unresolved` -> `narrowed` -> `confirmed`", "sequential-thinking", "context7", "playwright", ".agents/skills/<slug>/SKILL.md", ".codex/agents/<slug>.toml", ".codex/agents/<slug>.md", "public `namba create` Go CLI command", "five independent role outputs", "namba __create preview", "namba __create apply"} {
		if !strings.Contains(createSkill, want) {
			t.Fatalf("expected create skill to contain %q, got %q", want, createSkill)
		}
	}
	reviewResolveSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-review-resolve", "SKILL.md"))
	for _, want := range []string{"$namba-review-resolve", "thread-aware GitHub path", "`fixed-and-resolved`, `answered-open`, or `skipped-with-rationale`", "CI/check evidence when the review feedback or PR health depends on failing checks", "configured `@codex review` marker is present exactly once"} {
		if !strings.Contains(reviewResolveSkill, want) {
			t.Fatalf("expected review-resolve skill to contain %q, got %q", want, reviewResolveSkill)
		}
	}
	releaseSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-release", "SKILL.md"))
	for _, want := range []string{"$namba-release", "clean working tree before the final tagging step", ".namba/releases/<version>.md", "guarded `namba release --version <version> --push` path"} {
		if !strings.Contains(releaseSkill, want) {
			t.Fatalf("expected release skill to contain %q, got %q", want, releaseSkill)
		}
	}
	harnessSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-harness", "SKILL.md"))
	for _, want := range []string{"$namba-harness", "namba harness", "without inventing a second artifact model", "deterministic helper-script candidates", "mechanical versus behavioral edits", "workflow-first designs", "Codex-native"} {
		if !strings.Contains(harnessSkill, want) {
			t.Fatalf("expected harness skill to contain %q, got %q", want, harnessSkill)
		}
	}
	fixSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-fix", "SKILL.md"))
	for _, want := range []string{"namba fix --command run", "namba fix --command plan", "read-only", "namba sync"} {
		if !strings.Contains(fixSkill, want) {
			t.Fatalf("expected fix skill to contain %q, got %q", want, fixSkill)
		}
	}
	prSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-pr", "SKILL.md"))
	if !strings.Contains(prSkill, "$namba-pr") || !strings.Contains(prSkill, "namba pr") {
		t.Fatalf("expected command-entry pr skill, got %q", prSkill)
	}
	for _, want := range []string{"Inspect current PR check status before review handoff", "bounded GitHub Actions failure snippets", "configured Codex review marker exists exactly once"} {
		if !strings.Contains(prSkill, want) {
			t.Fatalf("expected pr skill to contain %q, got %q", want, prSkill)
		}
	}
	landSkill := mustReadFile(t, filepath.Join(tmp, ".agents", "skills", "namba-land", "SKILL.md"))
	if !strings.Contains(landSkill, "$namba-land") || !strings.Contains(landSkill, "namba land") {
		t.Fatalf("expected command-entry land skill, got %q", landSkill)
	}
	if _, err := os.Stat(legacySkillPath); !os.IsNotExist(err) {
		t.Fatalf("expected legacy codex skill mirror to be removed, stat err=%v", err)
	}
	if _, err := os.Stat(customSkillPath); err != nil {
		t.Fatalf("expected custom codex skill to be preserved, stat err=%v", err)
	}

	config := mustReadFile(t, filepath.Join(tmp, ".codex", "config.toml"))
	if !strings.Contains(config, "#:schema https://developers.openai.com/codex/config-schema.json") || !strings.Contains(config, "repo-safe Codex defaults under version control") || !strings.Contains(config, "max_threads = 5") || !strings.Contains(config, `approval_policy = "never"`) || !strings.Contains(config, `sandbox_mode = "read-only"`) {
		t.Fatalf("expected multi-agent Codex config, got %q", config)
	}
	workflowConfig := mustReadFile(t, filepath.Join(tmp, ".namba", "config", "sections", "workflow.yaml"))
	if !strings.Contains(workflowConfig, "max_parallel_workers: 3") {
		t.Fatalf("expected worktree parallel config to stay separate from agent thread defaults, got %q", workflowConfig)
	}
	for _, want := range []string{"[mcp_servers.context7]", "@upstash/context7-mcp", "[mcp_servers.sequential-thinking]", "@modelcontextprotocol/server-sequential-thinking", "[mcp_servers.playwright]", "@playwright/mcp@latest"} {
		if !strings.Contains(config, want) {
			t.Fatalf("expected Codex config to include %q, got %q", want, config)
		}
	}
	if strings.Contains(config, "status_line =") {
		t.Fatalf("expected status line preset off to omit status line, got %q", config)
	}

	codexReadme := mustReadFile(t, filepath.Join(tmp, ".namba", "codex", "README.md"))
	if !strings.Contains(codexReadme, "`namba regen` regenerates") || !strings.Contains(codexReadme, "`namba update` self-updates") {
		t.Fatalf("expected codex README to describe regen/update semantics, got %q", codexReadme)
	}
	if !strings.Contains(codexReadme, ".codex/agents/*.toml") || !strings.Contains(codexReadme, "`default`, `worker`, and `explorer`") {
		t.Fatalf("expected codex README to describe built-in and custom agents, got %q", codexReadme)
	}
	if !strings.Contains(codexReadme, "repo-managed MCP presets") {
		t.Fatalf("expected codex README to describe managed MCP presets, got %q", codexReadme)
	}
	for _, want := range []string{"`namba fix --command plan \"<issue description>\"`", "`namba fix --command run \"<issue description>\"`", "direct-repair paths in the current workspace"} {
		if !strings.Contains(codexReadme, want) {
			t.Fatalf("expected codex README to contain %q, got %q", want, codexReadme)
		}
	}
	for _, want := range []string{"## Namba Custom Agent Roster", "## Delegation Heuristics", "## Plan Review Readiness", "`namba-product-manager`", "`namba-plan-reviewer`", "`namba-mobile-engineer`", "`namba-designer`", "`namba-data-engineer`", "`namba-security-engineer`", "`namba-test-engineer`", "`namba-devops-engineer`"} {
		if !strings.Contains(codexReadme, want) {
			t.Fatalf("expected codex README to contain %q, got %q", want, codexReadme)
		}
	}
	if !strings.Contains(codexReadme, "$namba-help") || !strings.Contains(codexReadme, "$namba-create") || !strings.Contains(codexReadme, "$namba-run") || !strings.Contains(codexReadme, "$namba-queue") || !strings.Contains(codexReadme, "$namba-harness") || !strings.Contains(codexReadme, "$namba-plan-review") || !strings.Contains(codexReadme, "$namba-plan-pm-review") || !strings.Contains(codexReadme, "$namba-review-resolve") || !strings.Contains(codexReadme, "$namba-release") || !strings.Contains(codexReadme, "reviews/readiness.md") || strings.Contains(codexReadme, ".codex/skills/") {
		t.Fatalf("expected codex README to describe command-entry skills without codex skill mirror, got %q", codexReadme)
	}
	if !strings.Contains(codexReadme, "`namba harness \"<description>\"`") {
		t.Fatalf("expected codex README to describe harness-oriented planning semantics, got %q", codexReadme)
	}
	if !strings.Contains(codexReadme, "NAMBA-AI 작업 결과 보고") || !strings.Contains(codexReadme, "validate-output-contract.py") || !strings.Contains(codexReadme, "selected language palette") {
		t.Fatalf("expected codex README to describe output contract fallback validation, got %q", codexReadme)
	}
	if !strings.Contains(codexReadme, "WSL workspace") || !strings.Contains(codexReadme, "documented config and hook surface evolves") || strings.Contains(codexReadme, "do not document a repository-configurable stop-hook surface") {
		t.Fatalf("expected codex README to describe current Windows guidance and explicit validator positioning, got %q", codexReadme)
	}
	if !strings.Contains(codexReadme, "PR titles and bodies should be written in Korean") || !strings.Contains(codexReadme, "`@codex review`") {
		t.Fatalf("expected codex README to describe PR collaboration defaults, got %q", codexReadme)
	}
	if !strings.Contains(codexReadme, "`namba pr`") || !strings.Contains(codexReadme, "`namba land`") {
		t.Fatalf("expected codex README to describe PR handoff commands, got %q", codexReadme)
	}
	if !strings.Contains(codexReadme, "`namba run SPEC-XXX --solo`") || !strings.Contains(codexReadme, "`namba run SPEC-XXX --team`") || !strings.Contains(codexReadme, "`namba run SPEC-XXX --parallel`") {
		t.Fatalf("expected codex README to describe standalone run modes, got %q", codexReadme)
	}

	outputContractDoc := mustReadFile(t, filepath.Join(tmp, ".namba", "codex", "output-contract.md"))
	if !strings.Contains(outputContractDoc, "NAMBA-AI 작업 결과 보고") || !strings.Contains(outputContractDoc, "작업 정의") || !strings.Contains(outputContractDoc, "잠재 문제") || !strings.Contains(outputContractDoc, "simple emoji section markers") {
		t.Fatalf("expected output contract doc, got %q", outputContractDoc)
	}
	if !strings.Contains(outputContractDoc, "documented Codex config and hook surface evolves") || !strings.Contains(outputContractDoc, "hook-based enforcement") {
		t.Fatalf("expected output contract doc to describe validator fallback against the current Codex hook surface, got %q", outputContractDoc)
	}

	claudeCodexMapping := mustReadFile(t, filepath.Join(tmp, ".namba", "codex", "claude-codex-mapping.md"))
	if !strings.Contains(claudeCodexMapping, "built-in subagent workflows") || strings.Contains(claudeCodexMapping, "experimental multi-agent delegation") {
		t.Fatalf("expected claude-codex mapping to describe current Codex subagent wording, got %q", claudeCodexMapping)
	}

	validator := mustReadFile(t, filepath.Join(tmp, ".namba", "codex", "validate-output-contract.py"))
	if !strings.Contains(validator, "output-contract: ok") || !strings.Contains(validator, "작업 정의") || !strings.Contains(validator, "header_aliases") || !strings.Contains(validator, "start=previous + 1") {
		t.Fatalf("expected output contract validator, got %q", validator)
	}

	agentFiles := []struct {
		path     string
		snippets []string
	}{
		{path: filepath.Join(tmp, ".codex", "agents", "namba-planner.toml"), snippets: []string{`name = "namba-planner"`, `sandbox_mode = "read-only"`, `model = "gpt-5.4"`, `model_reasoning_effort = "high"`, `developer_instructions = """`, `repo-managed MCP presets`}},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-plan-reviewer.toml"), snippets: []string{`name = "namba-plan-reviewer"`, `sandbox_mode = "read-only"`, `model = "gpt-5.4"`, `model_reasoning_effort = "high"`}},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-product-manager.toml"), snippets: []string{`name = "namba-product-manager"`, `sandbox_mode = "read-only"`, `model = "gpt-5.4"`, `model_reasoning_effort = "medium"`}},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-frontend-architect.toml"), snippets: []string{`name = "namba-frontend-architect"`, `sandbox_mode = "read-only"`, `model = "gpt-5.4"`, `model_reasoning_effort = "medium"`}},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-frontend-implementer.toml"), snippets: []string{`name = "namba-frontend-implementer"`, `sandbox_mode = "workspace-write"`, `model = "gpt-5.4-mini"`, `model_reasoning_effort = "medium"`}},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-mobile-engineer.toml"), snippets: []string{`name = "namba-mobile-engineer"`, `sandbox_mode = "read-only"`, `model = "gpt-5.4"`, `model_reasoning_effort = "medium"`}},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-designer.toml"), snippets: []string{`name = "namba-designer"`, `sandbox_mode = "read-only"`, `model = "gpt-5.4"`, `model_reasoning_effort = "medium"`, "art direction", "temperature and undertone discipline"}},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-backend-architect.toml"), snippets: []string{`name = "namba-backend-architect"`, `sandbox_mode = "read-only"`, `model = "gpt-5.4"`, `model_reasoning_effort = "medium"`}},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-backend-implementer.toml"), snippets: []string{`name = "namba-backend-implementer"`, `sandbox_mode = "workspace-write"`, `model = "gpt-5.4-mini"`, `model_reasoning_effort = "medium"`}},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-data-engineer.toml"), snippets: []string{`name = "namba-data-engineer"`, `sandbox_mode = "workspace-write"`, `model = "gpt-5.4-mini"`, `model_reasoning_effort = "medium"`}},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-security-engineer.toml"), snippets: []string{`name = "namba-security-engineer"`, `sandbox_mode = "workspace-write"`, `model = "gpt-5.4"`, `model_reasoning_effort = "high"`}},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-test-engineer.toml"), snippets: []string{`name = "namba-test-engineer"`, `sandbox_mode = "workspace-write"`, `model = "gpt-5.4-mini"`, `model_reasoning_effort = "medium"`}},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-devops-engineer.toml"), snippets: []string{`name = "namba-devops-engineer"`, `sandbox_mode = "workspace-write"`, `model = "gpt-5.4-mini"`, `model_reasoning_effort = "medium"`}},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-implementer.toml"), snippets: []string{`name = "namba-implementer"`, `sandbox_mode = "workspace-write"`, `model = "gpt-5.4-mini"`, `model_reasoning_effort = "medium"`}},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-reviewer.toml"), snippets: []string{`name = "namba-reviewer"`, `sandbox_mode = "read-only"`, `model = "gpt-5.4"`, `model_reasoning_effort = "high"`}},
	}
	for _, tc := range agentFiles {
		content := mustReadFile(t, tc.path)
		for _, snippet := range tc.snippets {
			if !strings.Contains(content, snippet) {
				t.Fatalf("expected %s to contain %q, got %q", tc.path, snippet, content)
			}
		}
	}

	roleCards := []struct {
		path    string
		heading string
	}{
		{path: filepath.Join(tmp, ".codex", "agents", "namba-planner.md"), heading: "# Namba Planner"},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-plan-reviewer.md"), heading: "# Namba Plan Reviewer"},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-product-manager.md"), heading: "# Namba Product Manager"},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-frontend-architect.md"), heading: "# Namba Frontend Architect"},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-frontend-implementer.md"), heading: "# Namba Frontend Implementer"},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-mobile-engineer.md"), heading: "# Namba Mobile Engineer"},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-designer.md"), heading: "# Namba Designer"},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-backend-architect.md"), heading: "# Namba Backend Architect"},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-backend-implementer.md"), heading: "# Namba Backend Implementer"},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-data-engineer.md"), heading: "# Namba Data Engineer"},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-security-engineer.md"), heading: "# Namba Security Engineer"},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-test-engineer.md"), heading: "# Namba Test Engineer"},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-devops-engineer.md"), heading: "# Namba DevOps Engineer"},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-implementer.md"), heading: "# Namba Implementer"},
		{path: filepath.Join(tmp, ".codex", "agents", "namba-reviewer.md"), heading: "# Namba Reviewer"},
	}
	for _, tc := range roleCards {
		content := mustReadFile(t, tc.path)
		if !strings.Contains(content, tc.heading) {
			t.Fatalf("expected readable role-card mirror %s to contain %q, got %q", tc.path, tc.heading, content)
		}
	}
}

func TestRunRegenSignalsSessionRefreshWhenInstructionSurfaceChanges(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	tmp := t.TempDir()
	app := NewApp(stdout, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmp, "AGENTS.md"), []byte("stale agents\n"), 0o644); err != nil {
		t.Fatalf("write stale AGENTS: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"regen"}); err != nil {
		t.Fatalf("regen failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "Session refresh required:") {
		t.Fatalf("expected regen output to signal session refresh, got %q", stdout.String())
	}

	notice := mustReadFile(t, filepath.Join(tmp, ".namba", "logs", "session-refresh-required.json"))
	if !strings.Contains(notice, "\"required\": true") || !strings.Contains(notice, "AGENTS.md") {
		t.Fatalf("expected session refresh notice log, got %q", notice)
	}
}

func TestRunRegenRejectsUnsupportedManagedMCPServer(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmp, ".namba", "config", "sections", "codex.yaml"), []byte("agent_mode: single\nstatus_line_preset: namba\ndefault_mcp_servers: does-not-exist\n"), 0o644); err != nil {
		t.Fatalf("write codex config: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	err := app.Run(context.Background(), []string{"regen"})
	if err == nil {
		t.Fatal("expected regen to reject an unsupported managed MCP server preset")
	}
	if !strings.Contains(err.Error(), `default MCP server "does-not-exist" is not supported`) {
		t.Fatalf("expected managed MCP validation error, got %v", err)
	}
}
