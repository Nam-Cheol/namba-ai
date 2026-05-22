package namba

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func (a *App) runInit(_ context.Context, args []string) error {
	if wantsCommandHelp(args) {
		return a.printCommandUsage("init")
	}
	opts, err := parseInitArgs(args)
	if err != nil {
		return commandUsageError("init", err)
	}

	root, err := filepath.Abs(opts.Path)
	if err != nil {
		return fmt.Errorf("resolve target: %w", err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("create target: %w", err)
	}

	scan := scanInitRepository(root)
	profile, err := a.resolveInitProfileWithScan(root, opts, scan)
	if err != nil {
		return err
	}

	testCmd, lintCmd, typecheckCmd := defaultQualityCommandsWithScan(root, profile.Language, profile.Framework, scan)
	files := map[string]string{
		"AGENTS.md": renderAgents(profile),
		filepath.ToSlash(filepath.Join(configDir, "project.yaml")):      renderProjectConfig(profile),
		filepath.ToSlash(filepath.Join(configDir, "analysis.yaml")):     renderAnalysisConfig(profile),
		filepath.ToSlash(filepath.Join(configDir, "quality.yaml")):      renderQualityConfig(profile.DevelopmentMode, testCmd, lintCmd, typecheckCmd),
		filepath.ToSlash(filepath.Join(configDir, "workflow.yaml")):     renderWorkflowConfig(),
		filepath.ToSlash(filepath.Join(configDir, "system.yaml")):       renderSystemConfig(profile),
		filepath.ToSlash(filepath.Join(configDir, "language.yaml")):     renderLanguageConfig(profile),
		filepath.ToSlash(filepath.Join(configDir, "user.yaml")):         renderUserConfig(profile),
		filepath.ToSlash(filepath.Join(configDir, "git-strategy.yaml")): renderGitStrategyConfig(profile),
		filepath.ToSlash(filepath.Join(configDir, "codex.yaml")):        renderCodexProfileConfig(profile),
		filepath.ToSlash(filepath.Join(configDir, "docs.yaml")):         renderDocsConfig(profile),
		filepath.ToSlash(filepath.Join(projectDir, "product.md")):       "# Product\n\nDescribe the product goals here.\n",
		filepath.ToSlash(filepath.Join(projectDir, "structure.md")):     "# Structure\n\nRun `namba project` to refresh this document.\n",
		filepath.ToSlash(filepath.Join(projectDir, "tech.md")):          "# Tech\n\nRun `namba project` to refresh this document.\n",
		filepath.ToSlash(filepath.Join(codemapsDir, "overview.md")):     "# Overview\n\nRun `namba project` to refresh this document.\n",
		filepath.ToSlash(filepath.Join(codemapsDir, "entry-points.md")): "# Entry Points\n\nRun `namba project` to refresh this document.\n",
		filepath.ToSlash(filepath.Join(codemapsDir, "dependencies.md")): "# Dependencies\n\nRun `namba project` to refresh this document.\n",
		filepath.ToSlash(filepath.Join(codemapsDir, "data-flow.md")):    "# Data Flow\n\nRun `namba project` to refresh this document.\n",
		filepath.ToSlash(filepath.Join(logsDir, ".gitkeep")):            "",
		filepath.ToSlash(filepath.Join(specsDir, ".gitkeep")):           "",
		filepath.ToSlash(filepath.Join(worktreesDir, ".gitkeep")):       "",
	}
	for rel, scaffold := range codexScaffoldFilesForOS(profile, a.goos) {
		files[rel] = scaffold
	}
	projectCfg := projectConfig{
		Name:        profile.ProjectName,
		ProjectType: profile.ProjectType,
		Language:    profile.Language,
		Framework:   profile.Framework,
	}
	for rel, body := range buildReadmeOutputs(projectCfg, profile, defaultDocsConfig(profile.ProjectType)) {
		files[rel] = body
	}

	manifest := Manifest{GeneratedAt: a.now().Format(time.RFC3339)}
	for rel, body := range files {
		absPath := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			return fmt.Errorf("create parent for %s: %w", rel, err)
		}
		if err := os.WriteFile(absPath, []byte(body), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", rel, err)
		}
		manifest.Entries = append(manifest.Entries, ManifestEntry{
			Path:      rel,
			Kind:      manifestKind(rel),
			Checksum:  checksum(body),
			UpdatedAt: manifest.GeneratedAt,
		})
	}

	sort.Slice(manifest.Entries, func(i, j int) bool { return manifest.Entries[i].Path < manifest.Entries[j].Path })
	if err := a.writeManifest(root, manifest); err != nil {
		return err
	}

	fmt.Fprintf(a.stdout, "Initialized NambaAI in %s\n", root)
	fmt.Fprintf(a.stdout, "Project: %s | Type: %s | Mode: %s | Agent mode: %s\n", profile.ProjectName, profile.ProjectType, profile.DevelopmentMode, profile.AgentMode)
	fmt.Fprintln(a.stdout, "Codex-native mode is ready. Open Codex in this directory and invoke `$namba`, `$namba-run`, or ask to use the Namba workflow.")
	fmt.Fprintln(a.stdout, "Codex hook review:")
	fmt.Fprintln(a.stdout, "  1. Open an interactive Codex session in this directory.")
	fmt.Fprintln(a.stdout, "  2. If Codex shows `6 hooks need review`, run `/hooks`.")
	fmt.Fprintln(a.stdout, "  3. Approve only after confirming every command points to this repository's `.codex/hooks/namba_codex_guard.sh` or Windows `.codex/hooks/namba_codex_guard.ps1` launcher.")
	fmt.Fprintln(a.stdout, "  4. Re-run an ambiguous Namba prompt to confirm Codex asks clarification questions before planning.")
	return nil
}

func (a *App) resolveInitProfile(root string, opts initOptions) (initProfile, error) {
	return a.resolveInitProfileWithScan(root, opts, scanInitRepository(root))
}

func (a *App) resolveInitProfileWithScan(root string, opts initOptions, scan initRepositoryScan) (initProfile, error) {
	profile := a.detectInitProfileWithScan(root, scan)
	applyInitOverrides(&profile, opts)
	if !opts.Yes && a.isInteractiveTerminal() {
		var err error
		profile, err = a.runInitWizard(profile)
		if err != nil {
			return initProfile{}, err
		}
	}
	if err := validateInitProfile(profile); err != nil {
		return initProfile{}, err
	}
	return profile, nil
}

func (a *App) detectInitProfile(root string) initProfile {
	return a.detectInitProfileWithScan(root, scanInitRepository(root))
}

func (a *App) detectInitProfileWithScan(root string, scan initRepositoryScan) initProfile {
	language, framework := detectLanguageFrameworkWithScan(root, scan)
	locale := detectLocale(a.getenv)
	name := normalizeProjectName(filepath.Base(root))
	if name == "" {
		name = "my-project"
	}

	return initProfile{
		ProjectName:           name,
		ProjectType:           detectProjectType(root),
		Language:              language,
		Framework:             framework,
		DevelopmentMode:       detectMethodologyWithScan(scan),
		ConversationLanguage:  locale,
		DocumentationLanguage: locale,
		CommentLanguage:       locale,
		GitMode:               "manual",
		GitProvider:           "github",
		GitLabInstanceURL:     "https://gitlab.com",
		ApprovalPolicy:        "on-request",
		SandboxMode:           "workspace-write",
		BranchPerWork:         true,
		BranchBase:            "main",
		SpecBranchPrefix:      "spec/",
		TaskBranchPrefix:      "task/",
		PRBaseBranch:          "main",
		PRLanguage:            locale,
		CodexReviewComment:    "@codex review",
		AutoCodexReview:       false,
		AgentMode:             "single",
		StatusLinePreset:      "namba",
		UserName:              detectUserName(a.getenv),
		CreatedAt:             a.now().Format(timeLayoutDateTime),
	}
}

func applyInitOverrides(profile *initProfile, opts initOptions) {
	if value := strings.TrimSpace(opts.HumanLanguage); value != "" {
		applyHumanLanguage(profile, value)
	}
	if value := strings.TrimSpace(opts.ProjectName); value != "" {
		profile.ProjectName = value
	}
	if value := strings.TrimSpace(opts.ProjectType); value != "" {
		profile.ProjectType = value
	}
	if value := strings.TrimSpace(opts.Language); value != "" {
		profile.Language = value
	}
	if value := strings.TrimSpace(opts.Framework); value != "" {
		profile.Framework = value
	}
	if value := strings.TrimSpace(opts.DevelopmentMode); value != "" {
		profile.DevelopmentMode = value
	}
	if value := strings.TrimSpace(opts.ConversationLanguage); value != "" {
		profile.ConversationLanguage = value
	}
	if value := strings.TrimSpace(opts.DocumentationLanguage); value != "" {
		profile.DocumentationLanguage = value
	}
	if value := strings.TrimSpace(opts.CommentLanguage); value != "" {
		profile.CommentLanguage = value
	}
	if value := strings.TrimSpace(opts.DocumentationLanguage); value != "" {
		profile.PRLanguage = value
	}
	if value := strings.TrimSpace(opts.ApprovalPolicy); value != "" {
		profile.ApprovalPolicy = value
	}
	if value := strings.TrimSpace(opts.SandboxMode); value != "" {
		profile.SandboxMode = value
	}
	if value := strings.TrimSpace(opts.GitMode); value != "" {
		profile.GitMode = value
	}
	if value := strings.TrimSpace(opts.GitProvider); value != "" {
		profile.GitProvider = value
	}
	if value := strings.TrimSpace(opts.GitUsername); value != "" {
		profile.GitUsername = value
	}
	if value := strings.TrimSpace(opts.GitLabInstanceURL); value != "" {
		profile.GitLabInstanceURL = value
	}
	if value := strings.TrimSpace(opts.AgentMode); value != "" {
		profile.AgentMode = value
	}
	if value := strings.TrimSpace(opts.StatusLinePreset); value != "" {
		profile.StatusLinePreset = value
	}
	if value := strings.TrimSpace(opts.UserName); value != "" {
		profile.UserName = value
	}
}

func applyHumanLanguage(profile *initProfile, language string) {
	value := strings.TrimSpace(language)
	if value == "" {
		return
	}
	profile.ConversationLanguage = value
	profile.DocumentationLanguage = value
	profile.CommentLanguage = value
	profile.PRLanguage = value
}

func formatInitStack(profile initProfile) string {
	language := firstNonBlank(profile.Language, "unknown")
	framework := normalizeFramework(profile.Framework)
	if language == "unknown" && framework == "none" {
		return "not selected yet"
	}
	if framework == "none" {
		return language
	}
	return language + " / " + framework
}

func projectTypeOptions() []option {
	return projectTypeOptionsFor("")
}

func projectTypeOptionsFor(defaultValue string) []option {
	return projectTypeOptionsForLanguage("ko", defaultValue)
}

func projectTypeOptionsForLanguage(language, defaultValue string) []option {
	newMarker := ""
	existingMarker := ""
	switch defaultValue {
	case "new":
		newMarker = " [recommended]"
	case "existing":
		existingMarker = " [recommended]"
	}
	if normalizeReadmeLanguage(language) == "ko" {
		return []option{
			{Value: "new", Label: "\uc0c8 \ud504\ub85c\uc81d\ud2b8" + newMarker, Description: "\ube48 \uc800\uc7a5\uc18c/\uc0c8 \ud3f4\ub354"},
			{Value: "existing", Label: "\uae30\uc874 \ud504\ub85c\uc81d\ud2b8" + existingMarker, Description: "\ucf54\ub4dc\uac00 \uc788\ub294 \uc800\uc7a5\uc18c"},
		}
	}
	return []option{
		{Value: "new", Label: "New project" + newMarker, Description: "empty repository/new folder"},
		{Value: "existing", Label: "Existing project" + existingMarker, Description: "repository with code"},
	}
}

func developmentModeOptionsForLanguage(language string) []option {
	if normalizeReadmeLanguage(language) == "ko" {
		return []option{
			{Value: "tdd", Label: "\U0001f9ea TDD", Description: "\uc0c8 \uae30\ub2a5\uc744 \uc791\uc740 \uac80\uc99d \ub2e8\uc704\ub85c \uc9c4\ud589"},
			{Value: "ddd", Label: "\U0001f9ed DDD", Description: "\uae30\uc874 \ub3c4\uba54\uc778/\ucf54\ub4dc \ubd84\uc11d\uc744 \uba3c\uc800 \uc815\ub82c"},
		}
	}
	return []option{
		{Value: "tdd", Label: "\U0001f9ea TDD", Description: "ship new work through small verification steps"},
		{Value: "ddd", Label: "\U0001f9ed DDD", Description: "align domain and code understanding before changing behavior"},
	}
}

func agentModeOptionsForLanguage(language string) []option {
	if normalizeReadmeLanguage(language) == "ko" {
		return []option{
			{Value: "single", Label: "\U0001f464 \uc2f1\uae00", Description: "\uc548\uc815\uc801\uc778 \ub2e8\uc77c \ud750\ub984"},
			{Value: "multi", Label: "\U0001f465 \uba40\ud2f0", Description: "\ubcd1\ub82c \uc791\uc5c5 \uc900\ube44"},
		}
	}
	return []option{
		{Value: "single", Label: "\U0001f464 Single", Description: "stable single-workspace flow"},
		{Value: "multi", Label: "\U0001f465 Multi", Description: "prepare for parallel work"},
	}
}

func statusLineOptionsForLanguage(language string) []option {
	if normalizeReadmeLanguage(language) == "ko" {
		return []option{
			{Value: "namba", Label: "\U0001f39b\ufe0f Namba", Description: "\ud504\ub85c\uc81d\ud2b8 \uc911\uc2ec \ud45c\uc2dc"},
			{Value: "off", Label: "\U0001f515 \ub044\uae30", Description: "\ucd94\ucc9c \uc124\uc815 \uc0dd\uc131 \uc548 \ud568"},
		}
	}
	return []option{
		{Value: "namba", Label: "\U0001f39b\ufe0f Namba", Description: "project-centered status display"},
		{Value: "off", Label: "\U0001f515 Off", Description: "do not generate the recommended setting"},
	}
}

func gitModeOptionsForLanguage(language string) []option {
	if normalizeReadmeLanguage(language) == "ko" {
		return []option{
			{Value: "manual", Label: "\u270b \uc218\ub3d9", Description: "push/PR \uc790\ub3d9\ud654 \uc5c6\uc74c"},
			{Value: "personal", Label: "\U0001f464 \uac1c\uc778", Description: "\ube0c\ub79c\uce58/\ucee4\ubc0b \ud5c8\uc6a9"},
			{Value: "team", Label: "\U0001f465 \ud300", Description: "PR \uc900\ube44 \uc0b0\ucd9c\ubb3c \uc0dd\uc131"},
		}
	}
	return []option{
		{Value: "manual", Label: "\u270b Manual", Description: "no push/PR automation"},
		{Value: "personal", Label: "\U0001f464 Personal", Description: "allow branch and commit helpers"},
		{Value: "team", Label: "\U0001f465 Team", Description: "prepare PR handoff artifacts"},
	}
}

func gitProviderOptionsForLanguage(language string) []option {
	if normalizeReadmeLanguage(language) == "ko" {
		return []option{
			{Value: "github", Label: "\U0001f419 GitHub", Description: "gh CLI \ub610\ub294 \uae30\uc874 \uc778\uc99d"},
			{Value: "gitlab", Label: "\U0001f98a GitLab", Description: "glab CLI \ub610\ub294 \uae30\uc874 \uc778\uc99d"},
		}
	}
	return []option{
		{Value: "github", Label: "\U0001f419 GitHub", Description: "gh CLI or existing authentication"},
		{Value: "gitlab", Label: "\U0001f98a GitLab", Description: "glab CLI or existing authentication"},
	}
}

func frameworkOptions(language string) []option {
	switch language {
	case "go":
		return []option{
			{Value: "none", Label: "\uc5c6\uc74c", Description: "\uae30\ubcf8 Go \ud504\ub85c\uc81d\ud2b8"},
			{Value: "cobra", Label: "Cobra", Description: "CLI \uc571"},
			{Value: "gin", Label: "Gin", Description: "HTTP \uc11c\ube44\uc2a4"},
			{Value: "echo", Label: "Echo", Description: "HTTP \uc11c\ube44\uc2a4"},
		}
	case "java":
		return []option{
			{Value: "none", Label: "\uc5c6\uc74c", Description: "\uae30\ubcf8 Java \ud504\ub85c\uc81d\ud2b8"},
			{Value: "maven", Label: "Maven", Description: "pom.xml \uae30\ubc18"},
			{Value: "gradle", Label: "Gradle", Description: "Gradle \ube4c\ub4dc"},
			{Value: "spring-boot", Label: "Spring Boot", Description: "Boot \uc571"},
		}
	case "typescript":
		return []option{
			{Value: "none", Label: "\uc5c6\uc74c", Description: "\uae30\ubcf8 Node \ud504\ub85c\uc81d\ud2b8"},
			{Value: "nextjs", Label: "Next.js", Description: "React \ud480\uc2a4\ud0dd"},
			{Value: "react", Label: "React", Description: "\ud074\ub77c\uc774\uc5b8\ud2b8 \uc571"},
			{Value: "nest", Label: "NestJS", Description: "\ubc31\uc5d4\ub4dc \uc11c\ube44\uc2a4"},
		}
	case "python":
		return []option{
			{Value: "none", Label: "\uc5c6\uc74c", Description: "\uae30\ubcf8 Python \ud504\ub85c\uc81d\ud2b8"},
			{Value: "fastapi", Label: "FastAPI", Description: "API \uc11c\ube44\uc2a4"},
			{Value: "django", Label: "Django", Description: "\uc6f9 \uc571"},
		}
	default:
		return []option{
			{Value: "none", Label: "\uc5c6\uc74c", Description: "\ud504\ub808\uc784\uc6cc\ud06c \ubbf8\uc120\ud0dd"},
		}
	}
}

func languageOptions() []option {
	return []option{
		{Value: "ko", Label: "[ko] Korean", Description: "\ud55c\uad6d\uc5b4"},
		{Value: "en", Label: "[en] English", Description: "English"},
		{Value: "ja", Label: "[ja] Japanese", Description: "\u65e5\u672c\u8a9e"},
		{Value: "zh", Label: "[zh] Simplified Chinese", Description: "\u7b80\u4f53\u4e2d\u6587"},
	}
}

func humanLanguageLabel(value string) string {
	for _, choice := range languageOptions() {
		if choice.Value == normalizeReadmeLanguage(value) {
			return choice.Label
		}
	}
	return "[en] English"
}

func approvalPolicyOptions() []option {
	return []option{
		{Value: "on-request", Label: "\U0001f6ce\ufe0f on-request", Description: "Codex asks for approval when needed"},
		{Value: "untrusted", Label: "\U0001f6a7 untrusted", Description: "Codex asks only for untrusted work"},
		{Value: "never", Label: "\u26a1 never", Description: "Codex continues without approval prompts"},
	}
}

func sandboxModeOptions() []option {
	return []option{
		{Value: "workspace-write", Label: "\U0001f4dd workspace-write", Description: "allow writes only in the current workspace"},
		{Value: "read-only", Label: "\U0001f441\ufe0f read-only", Description: "read files without write access"},
		{Value: "danger-full-access", Label: "\U0001f525 danger-full-access", Description: "allow full access without sandbox restrictions"},
	}
}

func detectLocale(getenv func(string) string) string {
	for _, key := range []string{"NAMBA_LANG", "LC_ALL", "LANG"} {
		value := strings.ToLower(getenv(key))
		switch {
		case strings.Contains(value, "ko"):
			return "ko"
		case strings.Contains(value, "ja"):
			return "ja"
		case strings.Contains(value, "zh"):
			return "zh"
		case strings.Contains(value, "en"):
			return "en"
		}
	}
	return "en"
}

func detectUserName(getenv func(string) string) string {
	for _, key := range []string{"NAMBA_USER", "USERNAME", "USER"} {
		if value := strings.TrimSpace(getenv(key)); value != "" {
			return value
		}
	}
	return "Developer"
}

func validateInitProfile(profile initProfile) error {
	if normalizeProjectName(profile.ProjectName) == "" {
		return fmt.Errorf("project name is required")
	}
	if !containsValue([]string{"tdd", "ddd"}, profile.DevelopmentMode) {
		return fmt.Errorf("development mode %q is not supported", profile.DevelopmentMode)
	}
	if !containsValue([]string{"new", "existing"}, profile.ProjectType) {
		return fmt.Errorf("project type %q is not supported", profile.ProjectType)
	}
	if !containsValue([]string{"go", "java", "typescript", "python", "unknown"}, profile.Language) {
		return fmt.Errorf("language %q is not supported", profile.Language)
	}
	if !containsValue([]string{"manual", "personal", "team"}, profile.GitMode) {
		return fmt.Errorf("git mode %q is not supported", profile.GitMode)
	}
	if !containsValue([]string{"github", "gitlab"}, profile.GitProvider) {
		return fmt.Errorf("git provider %q is not supported", profile.GitProvider)
	}
	if !containsValue([]string{"single", "multi"}, profile.AgentMode) {
		return fmt.Errorf("agent mode %q is not supported", profile.AgentMode)
	}
	if !containsValue([]string{"namba", "off"}, profile.StatusLinePreset) {
		return fmt.Errorf("status line preset %q is not supported", profile.StatusLinePreset)
	}
	if err := validateManagedMCPServerIDs(profile.DefaultMCPServers); err != nil {
		return err
	}
	for _, value := range []string{profile.ConversationLanguage, profile.DocumentationLanguage, profile.CommentLanguage} {
		if !containsValue([]string{"en", "ko", "ja", "zh"}, value) {
			return fmt.Errorf("language preference %q is not supported", value)
		}
	}
	if profile.PRLanguage != "" && !containsValue([]string{"en", "ko", "ja", "zh"}, profile.PRLanguage) {
		return fmt.Errorf("PR language %q is not supported", profile.PRLanguage)
	}
	if err := validateCodexAccessPair(profile.ApprovalPolicy, profile.SandboxMode); err != nil {
		return err
	}
	for field, value := range map[string]string{
		"branch base":          profile.BranchBase,
		"spec branch prefix":   profile.SpecBranchPrefix,
		"task branch prefix":   profile.TaskBranchPrefix,
		"PR base branch":       profile.PRBaseBranch,
		"codex review comment": profile.CodexReviewComment,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", field)
		}
	}
	return nil
}

type initProfile struct {
	ProjectName           string
	ProjectType           string
	Language              string
	Framework             string
	DevelopmentMode       string
	ConversationLanguage  string
	DocumentationLanguage string
	CommentLanguage       string
	ApprovalPolicy        string
	SandboxMode           string
	GitMode               string
	GitProvider           string
	GitUsername           string
	GitLabInstanceURL     string
	BranchPerWork         bool
	BranchBase            string
	SpecBranchPrefix      string
	TaskBranchPrefix      string
	PRBaseBranch          string
	PRLanguage            string
	CodexReviewComment    string
	AutoCodexReview       bool
	AgentMode             string
	StatusLinePreset      string
	DefaultMCPServers     []string
	UserName              string
	CreatedAt             string
}

type initOptions struct {
	Path                  string
	Yes                   bool
	ProjectName           string
	ProjectType           string
	Language              string
	Framework             string
	DevelopmentMode       string
	ConversationLanguage  string
	DocumentationLanguage string
	CommentLanguage       string
	HumanLanguage         string
	ApprovalPolicy        string
	SandboxMode           string
	GitMode               string
	GitProvider           string
	GitUsername           string
	GitLabInstanceURL     string
	AgentMode             string
	StatusLinePreset      string
	UserName              string
}

type option struct {
	Value       string
	Label       string
	Description string
}

func parseInitArgs(args []string) (initOptions, error) {
	opts := initOptions{
		Path:              ".",
		GitLabInstanceURL: "https://gitlab.com",
	}

	consumeValue := func(args []string, index *int, flag string) (string, error) {
		*index = *index + 1
		if *index >= len(args) {
			return "", fmt.Errorf("%s requires a value", flag)
		}
		return args[*index], nil
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			if opts.Path != "." {
				return initOptions{}, fmt.Errorf("unexpected argument %q", arg)
			}
			opts.Path = arg
			continue
		}

		switch arg {
		case "--yes":
			opts.Yes = true
		case "--name":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.ProjectName = value
		case "--project-type":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.ProjectType = value
		case "--language":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.Language = value
		case "--framework":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.Framework = value
		case "--mode":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.DevelopmentMode = value
		case "--conversation-language":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.ConversationLanguage = value
		case "--documentation-language":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.DocumentationLanguage = value
		case "--comment-language":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.CommentLanguage = value
		case "--human-language":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.HumanLanguage = value
		case "--approval-policy":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.ApprovalPolicy = value
		case "--sandbox-mode":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.SandboxMode = value
		case "--git-mode":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.GitMode = value
		case "--git-provider":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.GitProvider = value
		case "--git-username":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.GitUsername = value
		case "--gitlab-instance-url":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.GitLabInstanceURL = value
		case "--agent-mode":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.AgentMode = value
		case "--statusline":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.StatusLinePreset = value
		case "--user-name":
			value, err := consumeValue(args, &i, arg)
			if err != nil {
				return initOptions{}, err
			}
			opts.UserName = value
		default:
			return initOptions{}, fmt.Errorf("unknown flag %q", arg)
		}
	}

	return opts, nil
}
