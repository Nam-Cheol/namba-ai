package namba

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	repoSkillsDir          = ".agents/skills"
	defaultCompatSkillsDir = ".codex/skills"
	repoCodexAgentsDir     = ".codex/agents"
	repoCodexConfigPath    = ".codex/config.toml"
	codexStateDir          = ".namba/codex"
)

func codexScaffoldFiles(profile initProfile) map[string]string {
	files := map[string]string{
		filepath.ToSlash(filepath.Join(codexStateDir, "README.md")):                 renderCodexUsage(profile),
		filepath.ToSlash(filepath.Join(codexStateDir, "statusline.example.toml")):   renderCodexStatusLineExample(),
		filepath.ToSlash(filepath.Join(codexStateDir, "claude-codex-mapping.md")):   renderClaudeCodexMapping(),
		filepath.ToSlash(repoCodexConfigPath):                                       renderRepoCodexConfig(profile),
		filepath.ToSlash(filepath.Join(repoCodexAgentsDir, "namba-planner.md")):     renderPlannerRoleCard(),
		filepath.ToSlash(filepath.Join(repoCodexAgentsDir, "namba-implementer.md")): renderImplementerRoleCard(),
		filepath.ToSlash(filepath.Join(repoCodexAgentsDir, "namba-reviewer.md")):    renderReviewerRoleCard(),
	}
	for rel, content := range codexSkillTemplates() {
		files[filepath.ToSlash(filepath.Join(repoSkillsDir, rel))] = content
	}
	if compatDir := strings.TrimSpace(profile.CompatSkillsPath); compatDir != "" {
		for rel, content := range codexSkillTemplates() {
			files[filepath.ToSlash(filepath.Join(compatDir, rel))] = content
		}
	}
	return files
}

func codexSkillTemplates() map[string]string {
	return map[string]string{
		filepath.ToSlash(filepath.Join("namba", "SKILL.md")):                    renderNambaSkill(),
		filepath.ToSlash(filepath.Join("namba-foundation-core", "SKILL.md")):    renderFoundationSkill(),
		filepath.ToSlash(filepath.Join("namba-workflow-init", "SKILL.md")):      renderInitSkill(),
		filepath.ToSlash(filepath.Join("namba-workflow-project", "SKILL.md")):   renderProjectSkill(),
		filepath.ToSlash(filepath.Join("namba-workflow-execution", "SKILL.md")): renderExecutionSkill(),
	}
}

func codexNativeIssues(root string) []string {
	checks := []struct {
		label string
		path  string
	}{
		{label: "AGENTS.md", path: filepath.Join(root, "AGENTS.md")},
		{label: ".agents/skills/namba/SKILL.md", path: filepath.Join(root, ".agents", "skills", "namba", "SKILL.md")},
		{label: ".codex/config.toml", path: filepath.Join(root, ".codex", "config.toml")},
		{label: ".codex/agents/namba-planner.md", path: filepath.Join(root, ".codex", "agents", "namba-planner.md")},
		{label: ".namba/config/sections/codex.yaml", path: filepath.Join(root, ".namba", "config", "sections", "codex.yaml")},
	}
	return missingChecks(checks)
}

func codexCompatibilityIssues(root string, compatPath string) []string {
	compatPath = strings.TrimSpace(compatPath)
	if compatPath == "" {
		return nil
	}
	checks := []struct {
		label string
		path  string
	}{
		{label: filepath.ToSlash(filepath.Join(compatPath, "namba", "SKILL.md")), path: filepath.Join(root, filepath.FromSlash(compatPath), "namba", "SKILL.md")},
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
