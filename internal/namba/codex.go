package namba

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	repoSkillsDir       = ".agents/skills"
	repoCodexAgentsDir  = ".codex/agents"
	repoCodexConfigPath = ".codex/config.toml"
	codexStateDir       = ".namba/codex"
)

func codexScaffoldFiles(profile initProfile) map[string]string {
	files := map[string]string{
		filepath.ToSlash(filepath.Join(codexStateDir, "README.md")):                   renderCodexUsage(profile),
		filepath.ToSlash(filepath.Join(codexStateDir, "statusline.example.toml")):     renderCodexStatusLineExample(),
		filepath.ToSlash(filepath.Join(codexStateDir, "claude-codex-mapping.md")):     renderClaudeCodexMapping(),
		filepath.ToSlash(filepath.Join(codexStateDir, "output-contract.md")):          renderOutputContractDocLocalized(profile),
		filepath.ToSlash(filepath.Join(codexStateDir, "validate-output-contract.py")): renderOutputContractValidatorLocalized(profile),
		filepath.ToSlash(repoCodexConfigPath):                                         renderRepoCodexConfig(profile),
		filepath.ToSlash(filepath.Join(repoCodexAgentsDir, "namba-planner.md")):       renderPlannerRoleCard(),
		filepath.ToSlash(filepath.Join(repoCodexAgentsDir, "namba-planner.toml")):     renderPlannerCustomAgent(),
		filepath.ToSlash(filepath.Join(repoCodexAgentsDir, "namba-implementer.md")):   renderImplementerRoleCard(),
		filepath.ToSlash(filepath.Join(repoCodexAgentsDir, "namba-implementer.toml")): renderImplementerCustomAgent(),
		filepath.ToSlash(filepath.Join(repoCodexAgentsDir, "namba-reviewer.md")):      renderReviewerRoleCard(),
		filepath.ToSlash(filepath.Join(repoCodexAgentsDir, "namba-reviewer.toml")):    renderReviewerCustomAgent(),
	}
	for rel, content := range codexSkillTemplates(profile) {
		files[filepath.ToSlash(filepath.Join(repoSkillsDir, rel))] = content
	}
	return files
}

func codexSkillTemplates(profile initProfile) map[string]string {
	return map[string]string{
		filepath.ToSlash(filepath.Join("namba", "SKILL.md")):                    renderNambaSkill(profile),
		filepath.ToSlash(filepath.Join("namba-init", "SKILL.md")):               renderInitCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-project", "SKILL.md")):            renderProjectCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-regen", "SKILL.md")):              renderRegenCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-update", "SKILL.md")):             renderUpdateCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-plan", "SKILL.md")):               renderPlanCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-fix", "SKILL.md")):                renderFixCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-run", "SKILL.md")):                renderRunCommandSkill(profile),
		filepath.ToSlash(filepath.Join("namba-pr", "SKILL.md")):                 renderPRCommandSkill(profile),
		filepath.ToSlash(filepath.Join("namba-land", "SKILL.md")):               renderLandCommandSkill(profile),
		filepath.ToSlash(filepath.Join("namba-sync", "SKILL.md")):               renderSyncCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-foundation-core", "SKILL.md")):    renderFoundationSkill(),
		filepath.ToSlash(filepath.Join("namba-workflow-init", "SKILL.md")):      renderInitSkill(),
		filepath.ToSlash(filepath.Join("namba-workflow-project", "SKILL.md")):   renderProjectSkill(),
		filepath.ToSlash(filepath.Join("namba-workflow-execution", "SKILL.md")): renderExecutionSkill(profile),
	}
}

func codexNativeIssues(root string) []string {
	checks := []struct {
		label string
		path  string
	}{
		{label: "AGENTS.md", path: filepath.Join(root, "AGENTS.md")},
		{label: ".agents/skills/namba/SKILL.md", path: filepath.Join(root, ".agents", "skills", "namba", "SKILL.md")},
		{label: ".agents/skills/namba-run/SKILL.md", path: filepath.Join(root, ".agents", "skills", "namba-run", "SKILL.md")},
		{label: ".codex/config.toml", path: filepath.Join(root, ".codex", "config.toml")},
		{label: ".codex/agents/namba-planner.toml", path: filepath.Join(root, ".codex", "agents", "namba-planner.toml")},
		{label: ".namba/codex/output-contract.md", path: filepath.Join(root, ".namba", "codex", "output-contract.md")},
		{label: ".namba/codex/validate-output-contract.py", path: filepath.Join(root, ".namba", "codex", "validate-output-contract.py")},
		{label: ".namba/config/sections/codex.yaml", path: filepath.Join(root, ".namba", "config", "sections", "codex.yaml")},
	}
	return missingChecks(checks)
}

func missingChecks(checks []struct {
	label string
	path  string
}) []string {
	var missing []string
	for _, check := range checks {
		if !exists(check.path) {
			missing = append(missing, check.label)
		}
	}
	return missing
}

func formatDoctorStatus(issues []string) string {
	if len(issues) == 0 {
		return "ready"
	}
	return fmt.Sprintf("missing %s", strings.Join(issues, ", "))
}
