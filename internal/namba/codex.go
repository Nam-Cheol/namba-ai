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
	}
	for rel, content := range codexAgentTemplates() {
		files[filepath.ToSlash(filepath.Join(repoCodexAgentsDir, rel))] = content
	}
	for rel, content := range codexSkillTemplates(profile) {
		files[filepath.ToSlash(filepath.Join(repoSkillsDir, rel))] = content
	}
	return files
}

func codexAgentTemplates() map[string]string {
	return map[string]string{
		"namba-planner.md":                renderPlannerRoleCard(),
		"namba-planner.toml":              renderPlannerCustomAgent(),
		"namba-plan-reviewer.md":          renderPlanReviewerRoleCard(),
		"namba-plan-reviewer.toml":        renderPlanReviewerCustomAgent(),
		"namba-product-manager.md":        renderProductManagerRoleCard(),
		"namba-product-manager.toml":      renderProductManagerCustomAgent(),
		"namba-frontend-architect.md":     renderFrontendArchitectRoleCard(),
		"namba-frontend-architect.toml":   renderFrontendArchitectCustomAgent(),
		"namba-frontend-implementer.md":   renderFrontendImplementerRoleCard(),
		"namba-frontend-implementer.toml": renderFrontendImplementerCustomAgent(),
		"namba-mobile-engineer.md":        renderMobileEngineerRoleCard(),
		"namba-mobile-engineer.toml":      renderMobileEngineerCustomAgent(),
		"namba-designer.md":               renderDesignerRoleCard(),
		"namba-designer.toml":             renderDesignerCustomAgent(),
		"namba-backend-architect.md":      renderBackendArchitectRoleCard(),
		"namba-backend-architect.toml":    renderBackendArchitectCustomAgent(),
		"namba-backend-implementer.md":    renderBackendImplementerRoleCard(),
		"namba-backend-implementer.toml":  renderBackendImplementerCustomAgent(),
		"namba-data-engineer.md":          renderDataEngineerRoleCard(),
		"namba-data-engineer.toml":        renderDataEngineerCustomAgent(),
		"namba-security-engineer.md":      renderSecurityEngineerRoleCard(),
		"namba-security-engineer.toml":    renderSecurityEngineerCustomAgent(),
		"namba-test-engineer.md":          renderTestEngineerRoleCard(),
		"namba-test-engineer.toml":        renderTestEngineerCustomAgent(),
		"namba-devops-engineer.md":        renderDevOpsEngineerRoleCard(),
		"namba-devops-engineer.toml":      renderDevOpsEngineerCustomAgent(),
		"namba-implementer.md":            renderImplementerRoleCard(),
		"namba-implementer.toml":          renderImplementerCustomAgent(),
		"namba-reviewer.md":               renderReviewerRoleCard(),
		"namba-reviewer.toml":             renderReviewerCustomAgent(),
	}
}

func requiredCodexAgentFiles() []string {
	return []string{
		"namba-planner.toml",
		"namba-plan-reviewer.toml",
		"namba-product-manager.toml",
		"namba-frontend-architect.toml",
		"namba-frontend-implementer.toml",
		"namba-mobile-engineer.toml",
		"namba-designer.toml",
		"namba-backend-architect.toml",
		"namba-backend-implementer.toml",
		"namba-data-engineer.toml",
		"namba-security-engineer.toml",
		"namba-test-engineer.toml",
		"namba-devops-engineer.toml",
		"namba-implementer.toml",
		"namba-reviewer.toml",
	}
}

func managedCodexSkillNames() []string {
	return []string{
		"namba",
		"namba-init",
		"namba-help",
		"namba-coach",
		"namba-create",
		"namba-project",
		"namba-regen",
		"namba-update",
		"namba-plan",
		"namba-plan-review",
		"namba-harness",
		"namba-plan-pm-review",
		"namba-plan-eng-review",
		"namba-plan-design-review",
		"namba-fix",
		"namba-run",
		"namba-pr",
		"namba-land",
		"namba-sync",
		"namba-foundation-core",
		"namba-workflow-init",
		"namba-workflow-project",
		"namba-workflow-execution",
	}
}

func codexSkillTemplates(profile initProfile) map[string]string {
	return map[string]string{
		filepath.ToSlash(filepath.Join("namba", "SKILL.md")):                    renderNambaSkill(profile),
		filepath.ToSlash(filepath.Join("namba-init", "SKILL.md")):               renderInitCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-help", "SKILL.md")):               renderHelpCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-coach", "SKILL.md")):              renderCoachCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-create", "SKILL.md")):             renderCreateCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-project", "SKILL.md")):            renderProjectCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-regen", "SKILL.md")):              renderRegenCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-update", "SKILL.md")):             renderUpdateCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-plan", "SKILL.md")):               renderPlanCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-plan-review", "SKILL.md")):        renderPlanReviewLoopCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-harness", "SKILL.md")):            renderHarnessCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-plan-pm-review", "SKILL.md")):     renderPlanPMReviewCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-plan-eng-review", "SKILL.md")):    renderPlanEngReviewCommandSkill(),
		filepath.ToSlash(filepath.Join("namba-plan-design-review", "SKILL.md")): renderPlanDesignReviewCommandSkill(),
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

func isManagedRepoSkillPath(rel string) bool {
	for _, name := range managedCodexSkillNames() {
		if rel == filepath.ToSlash(filepath.Join(repoSkillsDir, name, "SKILL.md")) {
			return true
		}
	}
	return false
}

func isManagedRepoCodexAgentPath(rel string) bool {
	if !strings.HasPrefix(rel, repoCodexAgentsDir+"/") {
		return false
	}
	rel = strings.TrimPrefix(rel, repoCodexAgentsDir+"/")
	_, ok := codexAgentTemplates()[rel]
	return ok
}

func codexNativeIssues(root string) []string {
	checks := []struct {
		label string
		path  string
	}{
		{label: "AGENTS.md", path: filepath.Join(root, "AGENTS.md")},
		{label: ".agents/skills/namba/SKILL.md", path: filepath.Join(root, ".agents", "skills", "namba", "SKILL.md")},
		{label: ".agents/skills/namba-coach/SKILL.md", path: filepath.Join(root, ".agents", "skills", "namba-coach", "SKILL.md")},
		{label: ".agents/skills/namba-create/SKILL.md", path: filepath.Join(root, ".agents", "skills", "namba-create", "SKILL.md")},
		{label: ".agents/skills/namba-run/SKILL.md", path: filepath.Join(root, ".agents", "skills", "namba-run", "SKILL.md")},
		{label: ".codex/config.toml", path: filepath.Join(root, ".codex", "config.toml")},
		{label: ".namba/codex/output-contract.md", path: filepath.Join(root, ".namba", "codex", "output-contract.md")},
		{label: ".namba/codex/validate-output-contract.py", path: filepath.Join(root, ".namba", "codex", "validate-output-contract.py")},
		{label: ".namba/config/sections/codex.yaml", path: filepath.Join(root, ".namba", "config", "sections", "codex.yaml")},
	}
	for _, rel := range requiredCodexAgentFiles() {
		checks = append(checks, struct {
			label string
			path  string
		}{
			label: filepath.ToSlash(filepath.Join(repoCodexAgentsDir, rel)),
			path:  filepath.Join(root, repoCodexAgentsDir, rel),
		})
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
