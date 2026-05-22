package namba

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) loadProjectConfig(root string) (projectConfig, error) {
	values, err := readKeyValueFile(filepath.Join(root, configDir, "project.yaml"))
	if err != nil {
		return projectConfig{}, err
	}
	return projectConfig{
		Name:        values["name"],
		ProjectType: values["project_type"],
		Language:    values["language"],
		Framework:   values["framework"],
	}, nil
}

func (a *App) loadQualityConfig(root string) (qualityConfig, error) {
	values, err := readKeyValueFile(filepath.Join(root, configDir, "quality.yaml"))
	if err != nil {
		return qualityConfig{}, err
	}
	return qualityConfig{
		DevelopmentMode:        values["development_mode"],
		TestCommand:            values["test_command"],
		LintCommand:            values["lint_command"],
		TypecheckCommand:       values["typecheck_command"],
		BuildCommand:           values["build_command"],
		MigrationDryRunCommand: firstNonBlank(values["migration_dry_run_command"], values["migration_dry_run"]),
		SmokeStartCommand:      values["smoke_start_command"],
		OutputContractCommand:  firstNonBlank(values["output_contract_command"], values["contract_command"]),
	}, nil
}

func (a *App) loadDocsConfig(root string) (docsConfig, error) {
	projectCfg, err := a.loadProjectConfig(root)
	if err != nil {
		return docsConfig{}, err
	}
	cfg := defaultDocsConfig(projectCfg.ProjectType)
	values, err := readKeyValueFile(filepath.Join(root, configDir, "docs.yaml"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return docsConfig{}, err
	}
	cfg.ManageReadme = parseBoolValue(values["manage_readme"], cfg.ManageReadme)
	if value := strings.TrimSpace(values["readme_profile"]); value != "" {
		cfg.ReadmeProfile = value
	}
	if value := strings.TrimSpace(values["readme_default_language"]); value != "" {
		cfg.DefaultLanguage = value
	}
	if value, ok := values["readme_additional_languages"]; ok {
		cfg.AdditionalLanguages = parseCommaSeparatedList(value)
		cfg.AdditionalLanguagesSet = true
	}
	if value := strings.TrimSpace(values["readme_hero_image"]); value != "" {
		cfg.HeroImage = value
	}
	return normalizeDocsConfig(cfg, projectCfg.ProjectType), nil
}

func (a *App) loadInitProfileFromConfig(root string) (initProfile, error) {
	profile := a.detectInitProfile(root)

	projectValues, err := readKeyValueFile(filepath.Join(root, configDir, "project.yaml"))
	if err != nil {
		return initProfile{}, err
	}
	qualityValues, err := readKeyValueFile(filepath.Join(root, configDir, "quality.yaml"))
	if err != nil {
		return initProfile{}, err
	}
	languageValues, err := readKeyValueFile(filepath.Join(root, configDir, "language.yaml"))
	if err != nil {
		return initProfile{}, err
	}
	userValues, err := readKeyValueFile(filepath.Join(root, configDir, "user.yaml"))
	if err != nil {
		return initProfile{}, err
	}
	gitValues, err := readKeyValueFile(filepath.Join(root, configDir, "git-strategy.yaml"))
	if err != nil {
		return initProfile{}, err
	}
	codexValues, err := readKeyValueFile(filepath.Join(root, configDir, "codex.yaml"))
	if err != nil {
		return initProfile{}, err
	}
	systemValues, err := readKeyValueFile(filepath.Join(root, configDir, "system.yaml"))
	if err != nil {
		return initProfile{}, err
	}

	if value := strings.TrimSpace(projectValues["name"]); value != "" {
		profile.ProjectName = value
	}
	if value := strings.TrimSpace(projectValues["project_type"]); value != "" {
		profile.ProjectType = value
	}
	if value := strings.TrimSpace(projectValues["language"]); value != "" {
		profile.Language = value
	}
	if value := strings.TrimSpace(projectValues["framework"]); value != "" {
		profile.Framework = value
	}
	if value := strings.TrimSpace(projectValues["created_at"]); value != "" {
		profile.CreatedAt = value
	}
	if value := strings.TrimSpace(qualityValues["development_mode"]); value != "" {
		profile.DevelopmentMode = value
	}
	if value := strings.TrimSpace(languageValues["conversation_language"]); value != "" {
		profile.ConversationLanguage = value
	}
	if value := strings.TrimSpace(languageValues["documentation_language"]); value != "" {
		profile.DocumentationLanguage = value
	}
	if value := strings.TrimSpace(languageValues["comment_language"]); value != "" {
		profile.CommentLanguage = value
	}
	if value := strings.TrimSpace(userValues["user_name"]); value != "" {
		profile.UserName = value
	}
	if value := firstNonBlank(gitValues["git_mode"], gitValues["mode"]); value != "" {
		profile.GitMode = value
	}
	if value := firstNonBlank(gitValues["git_provider"], gitValues["provider"]); value != "" {
		profile.GitProvider = value
	}
	if value := firstNonBlank(gitValues["git_username"], gitValues["username"]); value != "" {
		profile.GitUsername = value
	}
	if value := firstNonBlank(gitValues["gitlab_instance_url"]); value != "" {
		profile.GitLabInstanceURL = value
	}
	profile.BranchPerWork = parseBoolValue(gitValues["branch_per_work"], profile.BranchPerWork)
	profile.AutoCodexReview = parseBoolValue(gitValues["auto_codex_review"], profile.AutoCodexReview)
	if value := firstNonBlank(gitValues["branch_base"]); value != "" {
		profile.BranchBase = value
	}
	if value := firstNonBlank(gitValues["spec_branch_prefix"]); value != "" {
		profile.SpecBranchPrefix = value
	}
	if value := firstNonBlank(gitValues["task_branch_prefix"]); value != "" {
		profile.TaskBranchPrefix = value
	}
	if value := firstNonBlank(gitValues["pr_base_branch"]); value != "" {
		profile.PRBaseBranch = value
	}
	if value := firstNonBlank(gitValues["pr_language"]); value != "" {
		profile.PRLanguage = value
	}
	if value := firstNonBlank(gitValues["codex_review_comment"]); value != "" {
		profile.CodexReviewComment = value
	}
	if value := strings.TrimSpace(codexValues["agent_mode"]); value != "" {
		profile.AgentMode = value
	}
	if value := strings.TrimSpace(codexValues["status_line_preset"]); value != "" {
		profile.StatusLinePreset = value
	}
	if value := strings.TrimSpace(codexValues["default_mcp_servers"]); value != "" {
		profile.DefaultMCPServers = parseCommaSeparatedValues(value)
	}
	if value := firstNonBlank(systemValues["approval_policy"], systemValues["approval_mode"]); value != "" {
		profile.ApprovalPolicy = value
	}
	if value := firstNonBlank(systemValues["sandbox_mode"]); value != "" {
		profile.SandboxMode = value
	}

	if err := validateInitProfile(profile); err != nil {
		return initProfile{}, err
	}
	return profile, nil
}

func readKeyValueFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		result[strings.TrimSpace(parts[0])] = trimConfigValue(strings.TrimSpace(parts[1]))
	}
	return result, nil
}

func trimConfigValue(value string) string {
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func parseBoolValue(raw string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return fallback
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func mustLoadQualityConfig(a *App, root string) qualityConfig {
	cfg, err := a.loadQualityConfig(root)
	if err != nil {
		return qualityConfig{}
	}
	return cfg
}

const timeLayoutDateTime = "2006-01-02T15:04:05Z07:00"
