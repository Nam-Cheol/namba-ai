package namba

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	nambaDir     = ".namba"
	specsDir     = ".namba/specs"
	projectDir   = ".namba/project"
	codemapsDir  = ".namba/project/codemaps"
	configDir    = ".namba/config/sections"
	logsDir      = ".namba/logs"
	worktreesDir = ".namba/worktrees"
	manifestPath = ".namba/manifest.json"
)

type App struct {
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
	now      func() time.Time
	getenv   func(string) string
	getwd    func() (string, error)
	lookPath func(string) (string, error)
	runCmd   func(context.Context, string, []string, string) (string, error)
}

type Manifest struct {
	GeneratedAt string          `json:"generated_at"`
	Entries     []ManifestEntry `json:"entries"`
}

type ManifestEntry struct {
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	Checksum  string `json:"checksum"`
	UpdatedAt string `json:"updated_at"`
}

type projectConfig struct {
	Name        string
	ProjectType string
	Language    string
	Framework   string
}

type qualityConfig struct {
	DevelopmentMode  string
	TestCommand      string
	LintCommand      string
	TypecheckCommand string
}

type specPackage struct {
	ID          string
	Description string
	Path        string
}

func NewApp(stdout, stderr io.Writer) *App {
	return &App{
		stdin:    os.Stdin,
		stdout:   stdout,
		stderr:   stderr,
		now:      time.Now,
		getenv:   os.Getenv,
		getwd:    os.Getwd,
		lookPath: exec.LookPath,
		runCmd: func(ctx context.Context, name string, args []string, dir string) (string, error) {
			cmd := exec.CommandContext(ctx, name, args...)
			cmd.Dir = dir
			output, err := cmd.CombinedOutput()
			return strings.TrimSpace(string(output)), err
		},
	}
}

func (a *App) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return a.printUsage()
	}

	switch args[0] {
	case "init":
		return a.runInit(ctx, args[1:])
	case "doctor":
		return a.runDoctor(ctx, args[1:])
	case "status":
		return a.runStatus(ctx, args[1:])
	case "project":
		return a.runProject(ctx, args[1:])
	case "update":
		return a.runUpdate(ctx, args[1:])
	case "plan":
		return a.runPlan(ctx, args[1:])
	case "fix":
		return a.runFix(ctx, args[1:])
	case "run":
		return a.runExecute(ctx, args[1:])
	case "sync":
		return a.runSync(ctx, args[1:])
	case "release":
		return a.runRelease(ctx, args[1:])
	case "worktree":
		return a.runWorktree(ctx, args[1:])
	case "help", "-h", "--help":
		return a.printUsage()
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usageText())
	}
}

func (a *App) printUsage() error {
	_, err := fmt.Fprint(a.stdout, usageText())
	return err
}

func usageText() string {
	return `NambaAI CLI

Usage:
  namba init [path] [--yes] [--name NAME] [--mode tdd|ddd] [--project-type new|existing]
  namba doctor
  namba status
  namba project
  namba update
  namba plan "<description>"
  namba fix "<description>"
  namba run SPEC-XXX [--parallel] [--dry-run]
  namba sync
  namba release [--bump patch|minor|major] [--version vX.Y.Z] [--push] [--remote origin]
  namba worktree <new|list|remove|clean>
`
}

func (a *App) runInit(_ context.Context, args []string) error {
	opts, err := parseInitArgs(args)
	if err != nil {
		return err
	}

	root, err := filepath.Abs(opts.Path)
	if err != nil {
		return fmt.Errorf("resolve target: %w", err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("create target: %w", err)
	}

	profile, err := a.resolveInitProfile(root, opts)
	if err != nil {
		return err
	}

	testCmd, lintCmd, typecheckCmd := defaultQualityCommands(root, profile.Language, profile.Framework)
	files := map[string]string{
		"AGENTS.md": renderAgents(profile),
		filepath.ToSlash(filepath.Join(configDir, "project.yaml")):      renderProjectConfig(profile),
		filepath.ToSlash(filepath.Join(configDir, "quality.yaml")):      renderQualityConfig(profile.DevelopmentMode, testCmd, lintCmd, typecheckCmd),
		filepath.ToSlash(filepath.Join(configDir, "workflow.yaml")):     renderWorkflowConfig(),
		filepath.ToSlash(filepath.Join(configDir, "system.yaml")):       renderSystemConfig(),
		filepath.ToSlash(filepath.Join(configDir, "language.yaml")):     renderLanguageConfig(profile),
		filepath.ToSlash(filepath.Join(configDir, "user.yaml")):         renderUserConfig(profile),
		filepath.ToSlash(filepath.Join(configDir, "git-strategy.yaml")): renderGitStrategyConfig(profile),
		filepath.ToSlash(filepath.Join(configDir, "codex.yaml")):        renderCodexProfileConfig(profile),
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
	for rel, scaffold := range codexScaffoldFiles(profile) {
		files[rel] = scaffold
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
	fmt.Fprintln(a.stdout, "Codex-native mode is ready. Open Codex in this directory and invoke `$namba` or ask to use the Namba workflow.")
	return nil
}

func (a *App) runDoctor(ctx context.Context, _ []string) error {
	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}
	projectCfg, _ := a.loadProjectConfig(root)
	qualityCfg, _ := a.loadQualityConfig(root)

	codexPath, codexErr := a.lookPath("codex")
	gitPath, gitErr := a.lookPath("git")
	nambaPath, nambaErr := a.lookPath("namba")

	fmt.Fprintf(a.stdout, "Project: %s\n", projectCfg.Name)
	fmt.Fprintf(a.stdout, "Project type: %s\n", projectCfg.ProjectType)
	fmt.Fprintf(a.stdout, "Language: %s\n", projectCfg.Language)
	fmt.Fprintf(a.stdout, "Framework: %s\n", projectCfg.Framework)
	fmt.Fprintf(a.stdout, "Mode: %s\n", qualityCfg.DevelopmentMode)
	fmt.Fprintf(a.stdout, "Codex native repo: %s\n", formatDoctorStatus(codexNativeIssues(root)))
	fmt.Fprintf(a.stdout, "Codex compatibility marker: %s\n", formatDoctorStatus(codexCompatibilityIssues(root)))
	if codexErr != nil {
		fmt.Fprintln(a.stdout, "Codex: missing")
	} else {
		fmt.Fprintf(a.stdout, "Codex: %s\n", codexPath)
	}
	if gitErr != nil {
		fmt.Fprintln(a.stdout, "Git: missing")
	} else {
		fmt.Fprintf(a.stdout, "Git: %s\n", gitPath)
	}
	if nambaErr != nil {
		fmt.Fprintln(a.stdout, "Namba CLI: missing from PATH")
	} else {
		fmt.Fprintf(a.stdout, "Namba CLI: %s\n", nambaPath)
	}
	if codexErr == nil {
		out, err := a.runBinary(ctx, "codex", []string{"--version"}, root)
		if err == nil && out != "" {
			fmt.Fprintf(a.stdout, "Codex version: %s\n", out)
		}
	}
	return nil
}

func (a *App) runStatus(_ context.Context, _ []string) error {
	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}

	projectCfg, _ := a.loadProjectConfig(root)
	qualityCfg, _ := a.loadQualityConfig(root)
	specCount := countDirectories(filepath.Join(root, specsDir), "SPEC-")

	fmt.Fprintf(a.stdout, "Project: %s\n", projectCfg.Name)
	fmt.Fprintf(a.stdout, "Project type: %s\n", projectCfg.ProjectType)
	fmt.Fprintf(a.stdout, "Language: %s\n", projectCfg.Language)
	fmt.Fprintf(a.stdout, "Framework: %s\n", projectCfg.Framework)
	fmt.Fprintf(a.stdout, "Development mode: %s\n", qualityCfg.DevelopmentMode)
	fmt.Fprintf(a.stdout, "SPEC packages: %d\n", specCount)
	fmt.Fprintf(a.stdout, "State dir: %s\n", filepath.Join(root, nambaDir))
	return nil
}
func (a *App) runProject(_ context.Context, _ []string) error {
	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}

	projectCfg, _ := a.loadProjectConfig(root)
	readme := firstExisting(root, "README.md", "README.txt")
	product := "# Product\n\n"
	if readme != "" {
		content, _ := os.ReadFile(filepath.Join(root, readme))
		product += "Source: " + readme + "\n\n" + string(content)
	} else {
		product += fmt.Sprintf("%s is managed by NambaAI.\n", projectCfg.Name)
	}

	structure := buildStructureDoc(root)
	tech := buildTechDoc(projectCfg)
	overview, entries, deps, flow := buildCodemaps(root, projectCfg)

	outputs := map[string]string{
		filepath.ToSlash(filepath.Join(projectDir, "product.md")):       product,
		filepath.ToSlash(filepath.Join(projectDir, "structure.md")):     structure,
		filepath.ToSlash(filepath.Join(projectDir, "tech.md")):          tech,
		filepath.ToSlash(filepath.Join(codemapsDir, "overview.md")):     overview,
		filepath.ToSlash(filepath.Join(codemapsDir, "entry-points.md")): entries,
		filepath.ToSlash(filepath.Join(codemapsDir, "dependencies.md")): deps,
		filepath.ToSlash(filepath.Join(codemapsDir, "data-flow.md")):    flow,
	}

	if err := a.writeOutputs(root, outputs); err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "Refreshed NambaAI project docs and codemaps.")
	return nil
}

func (a *App) runPlan(_ context.Context, args []string) error {
	return a.createSpecPackage("plan", args)
}

func (a *App) runFix(_ context.Context, args []string) error {
	return a.createSpecPackage("fix", args)
}

func (a *App) createSpecPackage(kind string, args []string) error {
	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New(kind + " requires a description")
	}

	desc := strings.TrimSpace(strings.Join(args, " "))
	projectCfg, _ := a.loadProjectConfig(root)
	qualityCfg, _ := a.loadQualityConfig(root)

	specID, err := nextSpecID(filepath.Join(root, specsDir))
	if err != nil {
		return err
	}
	specPath := filepath.Join(root, specsDir, specID)
	if err := os.MkdirAll(specPath, 0o755); err != nil {
		return err
	}

	spec := buildSpecDoc(kind, specID, desc, projectCfg, qualityCfg)
	plan := buildSpecPlanDoc(kind, specID)
	acceptance := buildSpecAcceptanceDoc(kind, desc, qualityCfg.DevelopmentMode)

	outputs := map[string]string{
		filepath.ToSlash(filepath.Join(specsDir, specID, "spec.md")):       spec,
		filepath.ToSlash(filepath.Join(specsDir, specID, "plan.md")):       plan,
		filepath.ToSlash(filepath.Join(specsDir, specID, "acceptance.md")): acceptance,
	}
	if err := a.writeOutputs(root, outputs); err != nil {
		return err
	}

	fmt.Fprintf(a.stdout, "Created %s\n", specID)
	return nil
}

func buildSpecDoc(kind, specID, description string, projectCfg projectConfig, qualityCfg qualityConfig) string {
	switch kind {
	case "fix":
		return fmt.Sprintf("# %s\n\n## Problem\n\n%s\n\n## Goal\n\nApply the smallest safe fix that resolves the reported issue.\n\n## Context\n\n- Project: %s\n- Project type: %s\n- Language: %s\n- Mode: %s\n- Work type: fix\n", specID, description, projectCfg.Name, projectCfg.ProjectType, projectCfg.Language, qualityCfg.DevelopmentMode)
	default:
		return fmt.Sprintf("# %s\n\n## Goal\n\n%s\n\n## Context\n\n- Project: %s\n- Project type: %s\n- Language: %s\n- Mode: %s\n- Work type: plan\n", specID, description, projectCfg.Name, projectCfg.ProjectType, projectCfg.Language, qualityCfg.DevelopmentMode)
	}
}

func buildSpecPlanDoc(kind, specID string) string {
	switch kind {
	case "fix":
		return fmt.Sprintf("# %s Plan\n\n1. Refresh project context with `namba project`\n2. Reproduce or inspect the reported issue\n3. Implement the smallest safe fix\n4. Run validation commands and targeted regression checks\n5. Sync artifacts with `namba sync`\n", specID)
	default:
		return fmt.Sprintf("# %s Plan\n\n1. Refresh project context with `namba project`\n2. Implement the requested change\n3. Run validation commands\n4. Sync artifacts with `namba sync`\n", specID)
	}
}

func buildSpecAcceptanceDoc(kind, description, mode string) string {
	if kind == "fix" {
		return buildFixAcceptanceDoc(description, mode)
	}
	return buildAcceptanceDoc(description, mode)
}

func (a *App) runExecute(ctx context.Context, args []string) error {
	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New("run requires a SPEC id")
	}

	specID := args[0]
	parallel := false
	dryRun := false
	for _, arg := range args[1:] {
		switch arg {
		case "--parallel":
			parallel = true
		case "--dry-run":
			dryRun = true
		default:
			return fmt.Errorf("unknown flag %q", arg)
		}
	}

	specPkg, err := a.loadSpec(root, specID)
	if err != nil {
		return err
	}
	qualityCfg, err := a.loadQualityConfig(root)
	if err != nil {
		return err
	}
	systemCfg, err := a.loadSystemConfig(root)
	if err != nil {
		return err
	}
	prompt, tasks, err := a.buildExecutionPrompt(root, specPkg, qualityCfg)
	if err != nil {
		return err
	}

	logDir := filepath.Join(root, logsDir, "runs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return err
	}
	promptPath := filepath.Join(logDir, strings.ToLower(specID)+"-request.md")
	if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
		return err
	}

	if parallel {
		return a.runParallel(ctx, root, specPkg, tasks, prompt, qualityCfg, systemCfg, dryRun)
	}

	if dryRun {
		fmt.Fprintf(a.stdout, "Prepared execution request at %s\n", promptPath)
		return nil
	}

	request := a.newExecutionRequest(specPkg.ID, root, prompt, systemCfg)
	if _, _, err := a.executeRun(ctx, root, strings.ToLower(specID), request, root, qualityCfg); err != nil {
		return err
	}

	fmt.Fprintf(a.stdout, "Executed %s with %s\n", specID, request.Runner)
	return nil
}

func (a *App) runSync(ctx context.Context, _ []string) error {
	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}
	projectCfg, _ := a.loadProjectConfig(root)

	latestSpec, _ := latestSpecID(filepath.Join(root, specsDir))
	if err := a.runProject(ctx, nil); err != nil {
		return err
	}

	generatedAt := a.now().Format(time.RFC3339)
	summary := buildChangeSummaryDoc(projectCfg, latestSpec, generatedAt)
	checklist := buildPRChecklistDoc()
	releaseNotes := buildReleaseNotesDoc(projectCfg, latestSpec, generatedAt)
	releaseChecklist := buildReleaseChecklistDoc()
	outputs := map[string]string{
		filepath.ToSlash(filepath.Join(projectDir, "change-summary.md")):    summary,
		filepath.ToSlash(filepath.Join(projectDir, "pr-checklist.md")):      checklist,
		filepath.ToSlash(filepath.Join(projectDir, "release-notes.md")):     releaseNotes,
		filepath.ToSlash(filepath.Join(projectDir, "release-checklist.md")): releaseChecklist,
	}
	if err := a.writeOutputs(root, outputs); err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "Synced NambaAI artifacts.")
	return nil
}

func (a *App) buildExecutionPrompt(root string, specPkg specPackage, qualityCfg qualityConfig) (string, []string, error) {
	specBytes, err := os.ReadFile(filepath.Join(specPkg.Path, "spec.md"))
	if err != nil {
		return "", nil, err
	}
	planBytes, err := os.ReadFile(filepath.Join(specPkg.Path, "plan.md"))
	if err != nil {
		return "", nil, err
	}
	acceptanceBytes, err := os.ReadFile(filepath.Join(specPkg.Path, "acceptance.md"))
	if err != nil {
		return "", nil, err
	}

	tasks := extractAcceptanceTasks(string(acceptanceBytes))
	prompt := strings.Join([]string{
		"# NambaAI Execution Request",
		"",
		"Execute this SPEC package using the repository AGENTS.md and local Codex skills.",
		"",
		"## SPEC",
		string(specBytes),
		"",
		"## Plan",
		string(planBytes),
		"",
		"## Acceptance",
		string(acceptanceBytes),
		"",
		"## Validation",
		fmt.Sprintf("- test: %s", qualityCfg.TestCommand),
		fmt.Sprintf("- lint: %s", qualityCfg.LintCommand),
		fmt.Sprintf("- typecheck: %s", qualityCfg.TypecheckCommand),
		"",
		fmt.Sprintf("Project root: %s", root),
	}, "\n")

	return prompt, tasks, nil
}
func (a *App) runWorktree(ctx context.Context, args []string) error {
	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New("worktree requires a subcommand")
	}
	switch args[0] {
	case "new":
		if len(args) < 2 {
			return errors.New("worktree new requires a name")
		}
		name := args[1]
		path := filepath.Join(root, worktreesDir, name)
		_, err := a.runBinary(ctx, "git", []string{"worktree", "add", "-b", "namba/" + name, path, "HEAD"}, root)
		if err != nil {
			return err
		}
		fmt.Fprintf(a.stdout, "Created worktree %s\n", path)
		return nil
	case "list":
		out, err := a.runBinary(ctx, "git", []string{"worktree", "list", "--porcelain"}, root)
		if err != nil {
			return err
		}
		fmt.Fprintln(a.stdout, out)
		return nil
	case "remove":
		if len(args) < 2 {
			return errors.New("worktree remove requires a name")
		}
		path := filepath.Join(root, worktreesDir, args[1])
		_, err := a.runBinary(ctx, "git", []string{"worktree", "remove", "--force", path}, root)
		if err != nil {
			return err
		}
		fmt.Fprintf(a.stdout, "Removed worktree %s\n", path)
		return nil
	case "clean":
		_, err := a.runBinary(ctx, "git", []string{"worktree", "prune"}, root)
		if err != nil {
			return err
		}
		fmt.Fprintln(a.stdout, "Pruned worktrees.")
		return nil
	default:
		return fmt.Errorf("unknown worktree subcommand %q", args[0])
	}
}

type parallelWorkerState struct {
	name   string
	path   string
	branch string
	err    error
}

func (a *App) runParallel(ctx context.Context, root string, specPkg specPackage, tasks []string, prompt string, qualityCfg qualityConfig, systemCfg systemConfig, dryRun bool) error {
	return a.executeParallelRun(ctx, root, specPkg, tasks, prompt, qualityCfg, systemCfg, dryRun)
}

func (a *App) runCodexExec(ctx context.Context, dir, prompt string) (string, error) {
	if _, err := a.lookPath("codex"); err != nil {
		return "", errors.New("codex is not installed")
	}
	args := []string{"exec", "--full-auto", prompt}
	return a.runBinary(ctx, "codex", args, dir)
}

func (a *App) runValidators(ctx context.Context, root string, cfg qualityConfig) error {
	commands := []string{cfg.TestCommand, cfg.LintCommand, cfg.TypecheckCommand}
	for _, command := range commands {
		command = strings.TrimSpace(command)
		if command == "" || command == "none" {
			continue
		}
		if _, err := runShellCommand(ctx, a.runCmd, command, root); err != nil {
			return fmt.Errorf("validation failed for %q: %w", command, err)
		}
	}
	return nil
}

func (a *App) requireProjectRoot() (string, error) {
	cwd, err := a.getwd()
	if err != nil {
		return "", err
	}
	root := cwd
	for {
		if _, err := os.Stat(filepath.Join(root, nambaDir)); err == nil {
			return root, nil
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", errors.New("no NambaAI project found in current directory")
		}
		root = parent
	}
}

func (a *App) writeOutputs(root string, outputs map[string]string) error {
	manifest, _ := a.readManifest(root)
	if manifest.GeneratedAt == "" {
		manifest.GeneratedAt = a.now().Format(time.RFC3339)
	}
	for rel, content := range outputs {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			return err
		}
		manifest = upsertManifest(manifest, ManifestEntry{
			Path:      rel,
			Kind:      manifestKind(rel),
			Checksum:  checksum(content),
			UpdatedAt: a.now().Format(time.RFC3339),
		})
	}
	return a.writeManifest(root, manifest)
}

func (a *App) writeManifest(root string, manifest Manifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(root, manifestPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (a *App) readManifest(root string) (Manifest, error) {
	data, err := os.ReadFile(filepath.Join(root, manifestPath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Manifest{}, nil
		}
		return Manifest{}, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

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
		DevelopmentMode:  values["development_mode"],
		TestCommand:      values["test_command"],
		LintCommand:      values["lint_command"],
		TypecheckCommand: values["typecheck_command"],
	}, nil
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
	if value := strings.TrimSpace(gitValues["mode"]); value != "" {
		profile.GitMode = value
	}
	if value := strings.TrimSpace(gitValues["provider"]); value != "" {
		profile.GitProvider = value
	}
	if value := strings.TrimSpace(gitValues["username"]); value != "" {
		profile.GitUsername = value
	}
	if value := strings.TrimSpace(gitValues["gitlab_instance_url"]); value != "" {
		profile.GitLabInstanceURL = value
	}
	if value := strings.TrimSpace(codexValues["agent_mode"]); value != "" {
		profile.AgentMode = value
	}
	if value := strings.TrimSpace(codexValues["status_line_preset"]); value != "" {
		profile.StatusLinePreset = value
	}

	if err := validateInitProfile(profile); err != nil {
		return initProfile{}, err
	}
	return profile, nil
}
func (a *App) loadSpec(root, specID string) (specPackage, error) {
	specPath := filepath.Join(root, specsDir, specID)
	if _, err := os.Stat(specPath); err != nil {
		return specPackage{}, fmt.Errorf("spec %s not found", specID)
	}
	specBytes, _ := os.ReadFile(filepath.Join(specPath, "spec.md"))
	return specPackage{
		ID:          specID,
		Path:        specPath,
		Description: firstNonEmptyLine(string(specBytes)),
	}, nil
}

func (a *App) runBinary(ctx context.Context, name string, args []string, dir string) (string, error) {
	if runtime.GOOS == "windows" && name == "codex" {
		return a.runCmd(ctx, "cmd", append([]string{"/c", "codex"}, args...), dir)
	}
	return a.runCmd(ctx, name, args, dir)
}

func (a *App) currentBranch(ctx context.Context, root string) (string, error) {
	out, err := a.runBinary(ctx, "git", []string{"branch", "--show-current"}, root)
	if err != nil {
		return "", err
	}
	if out == "" {
		return "HEAD", nil
	}
	return out, nil
}

func runShellCommand(ctx context.Context, runner func(context.Context, string, []string, string) (string, error), command, dir string) (string, error) {
	if runtime.GOOS == "windows" {
		return runner(ctx, "powershell", []string{"-NoProfile", "-Command", command}, dir)
	}
	return runner(ctx, "sh", []string{"-lc", command}, dir)
}
func detectLanguageFramework(root string) (string, string) {
	switch {
	case exists(filepath.Join(root, "go.mod")):
		return "go", "none"
	case exists(filepath.Join(root, "pom.xml")) || exists(filepath.Join(root, "build.gradle")) || exists(filepath.Join(root, "build.gradle.kts")) || exists(filepath.Join(root, "gradlew")) || exists(filepath.Join(root, "gradlew.bat")) || treeContainsExtension(root, ".java"):
		return "java", detectJavaFramework(root)
	case exists(filepath.Join(root, "package.json")):
		return "typescript", detectNodeFramework(root)
	case exists(filepath.Join(root, "pyproject.toml")) || exists(filepath.Join(root, "requirements.txt")):
		return "python", "none"
	default:
		return "unknown", "none"
	}
}

func detectJavaFramework(root string) string {
	if exists(filepath.Join(root, "pom.xml")) {
		data, err := os.ReadFile(filepath.Join(root, "pom.xml"))
		if err == nil && strings.Contains(strings.ToLower(string(data)), "spring-boot") {
			return "spring-boot"
		}
		return "maven"
	}

	for _, name := range []string{"build.gradle", "build.gradle.kts"} {
		if !exists(filepath.Join(root, name)) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, name))
		if err == nil && strings.Contains(strings.ToLower(string(data)), "spring-boot") {
			return "spring-boot"
		}
		return "gradle"
	}

	if exists(filepath.Join(root, "gradlew")) || exists(filepath.Join(root, "gradlew.bat")) {
		return "gradle"
	}

	return "none"
}
func detectNodeFramework(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return "none"
	}
	text := string(data)
	switch {
	case strings.Contains(text, "\"next\""):
		return "nextjs"
	case strings.Contains(text, "\"react\""):
		return "react"
	case strings.Contains(text, "\"vue\""):
		return "vue"
	default:
		return "none"
	}
}

func detectMethodology(root string) string {
	source := 0
	tests := 0
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.Contains(path, string(filepath.Separator)+".namba"+string(filepath.Separator)) {
			return nil
		}
		switch filepath.Ext(path) {
		case ".go", ".java", ".js", ".ts", ".tsx", ".py", ".rs":
			source++
			if strings.Contains(strings.ToLower(filepath.Base(path)), "test") {
				tests++
			}
		}
		return nil
	})
	if source == 0 {
		return "tdd"
	}
	if float64(tests)/float64(source) >= 0.10 {
		return "tdd"
	}
	return "ddd"
}

func detectProjectType(root string) string {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "new"
	}
	if len(entries) == 0 {
		return "new"
	}
	return "existing"
}

func defaultQualityCommands(root, language, framework string) (string, string, string) {
	switch language {
	case "go":
		return "go test ./...", defaultGoFormatCommand(root), "go vet ./..."
	case "java":
		return defaultJavaQualityCommands(root, framework)
	case "typescript":
		return "npm test", "npm run lint", "npm run typecheck"
	case "python":
		return "pytest", "ruff check .", "none"
	default:
		return "none", "none", "none"
	}
}

func defaultJavaQualityCommands(root, framework string) (string, string, string) {
	switch normalizeFramework(framework) {
	case "spring-boot", "maven":
		return "mvn -q test", "mvn -q spotless:check", "mvn -q -DskipTests compile"
	case "gradle":
		gradle := defaultGradleCommand(root)
		return gradle + " test", gradle + " check", gradle + " compileJava"
	default:
		detected := detectJavaFramework(root)
		if detected != "none" {
			return defaultJavaQualityCommands(root, detected)
		}
		return "none", "none", "none"
	}
}

func defaultGradleCommand(root string) string {
	switch {
	case exists(filepath.Join(root, "gradlew")):
		return "./gradlew"
	case exists(filepath.Join(root, "gradlew.bat")):
		return ".\\gradlew.bat"
	default:
		return "gradle"
	}
}

func treeContainsExtension(root, ext string) bool {
	found := false
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ext {
			found = true
		}
		return nil
	})
	return found
}
func defaultGoFormatCommand(root string) string {
	skipDirs := map[string]bool{
		".git":     true,
		".namba":   true,
		".codex":   true,
		"external": true,
		"vendor":   true,
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return "none"
	}

	var targets []string
	for _, entry := range entries {
		name := entry.Name()
		switch {
		case entry.IsDir() && skipDirs[name]:
			continue
		case entry.IsDir() && directoryContainsGo(filepath.Join(root, name)):
			targets = append(targets, strconv.Quote(name))
		case !entry.IsDir() && filepath.Ext(name) == ".go":
			targets = append(targets, strconv.Quote(name))
		}
	}

	if len(targets) == 0 {
		return "none"
	}

	sort.Strings(targets)
	return "gofmt -l " + strings.Join(targets, " ")
}

func directoryContainsGo(root string) bool {
	found := false
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".go" {
			found = true
		}
		return nil
	})
	return found
}

func buildStructureDoc(root string) string {
	lines := []string{"# Structure", "", "```"}
	appendStructureEntries(&lines, root, "", 0, 2)
	lines = append(lines, "```", "")
	return strings.Join(lines, "\n")
}

func appendStructureEntries(lines *[]string, root, rel string, depth, maxDepth int) {
	entries, err := os.ReadDir(filepath.Join(root, rel))
	if err != nil {
		return
	}
	for _, entry := range entries {
		childRel := entry.Name()
		if rel != "" {
			childRel = filepath.Join(rel, entry.Name())
		}
		if shouldSkipStructureEntry(filepath.ToSlash(childRel)) {
			continue
		}
		*lines = append(*lines, childRel)
		if entry.IsDir() && depth < maxDepth {
			appendStructureEntries(lines, root, childRel, depth+1, maxDepth)
		}
	}
}

func shouldSkipStructureEntry(rel string) bool {
	switch {
	case rel == ".git", strings.HasPrefix(rel, ".git/"):
		return true
	case rel == ".cache", strings.HasPrefix(rel, ".cache/"):
		return true
	case rel == ".codex/skills", strings.HasPrefix(rel, ".codex/skills/"):
		return true
	case rel == "dist", strings.HasPrefix(rel, "dist/"):
		return true
	case rel == "external", strings.HasPrefix(rel, "external/"):
		return true
	case rel == ".namba/logs", strings.HasPrefix(rel, ".namba/logs/"):
		return true
	case rel == ".namba/worktrees", strings.HasPrefix(rel, ".namba/worktrees/"):
		return true
	case strings.HasSuffix(rel, ".exe"):
		return true
	case rel == "namba":
		return true
	default:
		return false
	}
}

func buildTechDoc(cfg projectConfig) string {
	return fmt.Sprintf("# Tech\n\n- Language: %s\n- Framework: %s\n- Runtime adapter: Codex\n- Repo-local skills: .agents/skills\n- Repo-local agent role cards: .codex/agents\n- State directory: .namba\n", cfg.Language, cfg.Framework)
}

func buildCodemaps(root string, cfg projectConfig) (string, string, string, string) {
	overview := fmt.Sprintf("# Overview\n\n%s is managed by NambaAI.\n\n- Language: %s\n- Framework: %s\n", cfg.Name, cfg.Language, cfg.Framework)
	entries := "# Entry Points\n\n- `cmd/namba/main.go`: CLI entry point\n- `internal/namba/namba.go`: command orchestration\n"
	deps := "# Dependencies\n\n- Go standard library\n- External runtime: Codex CLI\n- External runtime: Git\n"
	flow := "# Data Flow\n\n1. `init` runs a Codex-adapted project wizard, writes `.namba/config/sections/*.yaml`, repo skills under `.agents/skills`, role cards under `.codex/agents`, `.codex/skills/README.md` compatibility marker (no duplicated skills), and Codex repo config under `.codex/config.toml`\n2. `project` refreshes docs and codemaps\n3. `plan` creates a SPEC package\n4. `run` either builds a non-interactive Codex execution request or is interpreted as Codex-native in-session execution\n5. `sync` emits PR-ready artifacts\n"
	if exists(filepath.Join(root, "go.mod")) {
		deps += "- Project module detected via `go.mod`\n"
	}
	return overview, entries, deps, flow
}

func buildAcceptanceDoc(description, mode string) string {
	bullets := []string{"# Acceptance", "", "- [ ] The requested behavior described below is implemented:", "  " + description, "- [ ] Validation commands pass"}
	if mode == "tdd" {
		bullets = append(bullets, "- [ ] Tests covering the new behavior are present")
	} else {
		bullets = append(bullets, "- [ ] Existing behavior is preserved while improving the target area")
	}
	return strings.Join(bullets, "\n")
}

func buildChangeSummaryDoc(projectCfg projectConfig, latestSpec, generatedAt string) string {
	projectType := projectCfg.ProjectType
	if strings.TrimSpace(projectType) == "" {
		projectType = "existing"
	}
	if strings.TrimSpace(latestSpec) == "" {
		latestSpec = "none"
	}
	lines := []string{
		"# Change Summary",
		"",
		fmt.Sprintf("Project: %s", projectCfg.Name),
		fmt.Sprintf("Project type: %s", projectType),
		fmt.Sprintf("Latest SPEC: %s", latestSpec),
		fmt.Sprintf("Generated: %s", generatedAt),
		"",
		"## Workflow Docs Synced",
		"",
		"- README and product docs describe when to use `namba update` versus `namba sync`.",
		"- Release docs describe `namba release` guardrails on a clean `main` branch plus optional `--push` behavior.",
		"- Parallel run docs describe the worktree fan-out and merge-blocking policy for `namba run SPEC-XXX --parallel`.",
		"",
		"## Refresh Commands",
		"",
		"- `namba update` regenerates `AGENTS.md`, repo-local skills, role cards, `.codex/skills/README.md`, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.",
		"- `namba sync` refreshes `.namba/project/*` docs, release notes/checklists, and codemaps.",
	}
	return strings.Join(lines, "\n") + "\n"
}

func buildPRChecklistDoc() string {
	return strings.Join([]string{
		"# PR Checklist",
		"",
		"- [ ] README / user-facing docs refreshed",
		"- [ ] `namba update` rerun if template-generated Codex assets changed",
		"- [ ] `namba sync` artifacts refreshed",
		"- [ ] SPEC artifacts reviewed",
		"- [ ] Validation commands passed",
		"- [ ] Diff reviewed",
	}, "\n") + "\n"
}

func buildReleaseNotesDoc(projectCfg projectConfig, latestSpec, generatedAt string) string {
	projectType := projectCfg.ProjectType
	if strings.TrimSpace(projectType) == "" {
		projectType = "existing"
	}
	if strings.TrimSpace(latestSpec) == "" {
		latestSpec = "none"
	}
	lines := []string{
		"# Release Notes Draft",
		"",
		fmt.Sprintf("Project: %s", projectCfg.Name),
		fmt.Sprintf("Project type: %s", projectType),
		fmt.Sprintf("Reference SPEC: %s", latestSpec),
		fmt.Sprintf("Generated: %s", generatedAt),
		"",
		"## Workflow Changes",
		"",
		"- `namba update` regenerates `AGENTS.md`, repo-local skills, role cards, `.codex/skills/README.md`, and repo-local Codex config from `.namba/config/sections/*.yaml`.",
		"- `namba sync` refreshes product docs, codemaps, change summary, PR checklist, and release docs.",
		"- `namba run SPEC-XXX --parallel` fans out into up to three git worktrees, merges only after every worker passes execution and validation, and preserves failing worktrees and branches for inspection.",
		"",
		"## Release Guardrails",
		"",
		"- `namba release` requires a git repository, the `main` branch, and a clean working tree.",
		"- Validators from `.namba/config/sections/quality.yaml` run before the release tag is created.",
		"- With no explicit version, `namba release` defaults to the next `patch` tag. Use `--bump minor|major` or `--version vX.Y.Z` when needed.",
		"- `namba release --push` pushes both `main` and the new tag to the selected remote.",
		"",
		"## Release Commands",
		"",
		"```text",
		"namba sync",
		"namba release --bump patch",
		"# or",
		"namba release --version vX.Y.Z --push",
		"```",
		"",
		"## Expected Assets",
		"",
		"- `namba_Windows_x86_64.zip`",
		"- `namba_Windows_arm64.zip`",
		"- `namba_Linux_x86_64.tar.gz`",
		"- `namba_Linux_arm64.tar.gz`",
		"- `namba_macOS_x86_64.tar.gz`",
		"- `namba_macOS_arm64.tar.gz`",
		"- `checksums.txt`",
	}
	return strings.Join(lines, "\n") + "\n"
}

func buildReleaseChecklistDoc() string {
	return strings.Join([]string{
		"# Release Checklist",
		"",
		"- [ ] `namba update` rerun if template-generated Codex assets changed",
		"- [ ] `namba sync` artifacts refreshed",
		"- [ ] README and `.namba/codex/README.md` reflect update, release, and parallel workflow behavior",
		"- [ ] Working tree is clean and the current branch is `main`",
		"- [ ] Validation commands passed",
		"- [ ] `namba release --version vX.Y.Z` or `namba release --bump patch` executed",
		"- [ ] If `--push` was not used, `main` and the release tag were pushed manually",
		"- [ ] GitHub Release workflow completed and published assets plus `checksums.txt`",
	}, "\n") + "\n"
}

func buildFixAcceptanceDoc(description, mode string) string {
	bullets := []string{
		"# Acceptance",
		"",
		"- [ ] The reported issue described below is resolved:",
		"  " + description,
		"- [ ] Validation commands pass",
		"- [ ] Existing behavior around the affected area is preserved",
	}
	if mode == "tdd" {
		bullets = append(bullets, "- [ ] A regression test covering the fix is present")
	} else {
		bullets = append(bullets, "- [ ] A targeted reproduction or verification step is documented")
	}
	return strings.Join(bullets, "\n")
}

func nextSpecID(root string) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	maxID := 0
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "SPEC-") {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(entry.Name(), "SPEC-"))
		if err == nil && n > maxID {
			maxID = n
		}
	}
	return fmt.Sprintf("SPEC-%03d", maxID+1), nil
}

func latestSpecID(root string) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	maxID := ""
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "SPEC-") && entry.Name() > maxID {
			maxID = entry.Name()
		}
	}
	if maxID == "" {
		return "none", nil
	}
	return maxID, nil
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
		result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return result, nil
}

func extractAcceptanceTasks(text string) []string {
	var tasks []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- [ ]") {
			tasks = append(tasks, strings.TrimSpace(strings.TrimPrefix(line, "- [ ]")))
		}
	}
	return tasks
}

func chunkTasks(tasks []string, workers int) [][]string {
	if workers <= 1 || len(tasks) <= 1 {
		return [][]string{tasks}
	}
	chunks := make([][]string, workers)
	for i, task := range tasks {
		idx := i % workers
		chunks[idx] = append(chunks[idx], task)
	}
	var filtered [][]string
	for _, chunk := range chunks {
		if len(chunk) > 0 {
			filtered = append(filtered, chunk)
		}
	}
	return filtered
}

func countDirectories(root, prefix string) int {
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			count++
		}
	}
	return count
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func firstExisting(root string, candidates ...string) string {
	for _, candidate := range candidates {
		if exists(filepath.Join(root, candidate)) {
			return candidate
		}
	}
	return ""
}

func checksum(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func manifestKind(rel string) string {
	switch {
	case strings.HasPrefix(rel, ".agents/skills/"), strings.HasPrefix(rel, ".codex/skills/"):
		return "skill"
	case strings.HasPrefix(rel, ".codex/agents/"):
		return "agent"
	case strings.HasPrefix(rel, ".namba/specs/"):
		return "spec"
	case strings.HasPrefix(rel, ".namba/project/"):
		return "project-doc"
	case strings.HasSuffix(rel, ".yaml"):
		return "config"
	default:
		return "asset"
	}
}

func upsertManifest(manifest Manifest, entry ManifestEntry) Manifest {
	found := false
	for i := range manifest.Entries {
		if manifest.Entries[i].Path == entry.Path {
			manifest.Entries[i] = entry
			found = true
			break
		}
	}
	if !found {
		manifest.Entries = append(manifest.Entries, entry)
	}
	sort.Slice(manifest.Entries, func(i, j int) bool { return manifest.Entries[i].Path < manifest.Entries[j].Path })
	manifest.GeneratedAt = entry.UpdatedAt
	return manifest
}

func firstNonEmptyLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func mustLoadQualityConfig(a *App, root string) qualityConfig {
	cfg, err := a.loadQualityConfig(root)
	if err != nil {
		return qualityConfig{}
	}
	return cfg
}

func isGitRepository(root string) bool {
	return exists(filepath.Join(root, ".git"))
}

const timeLayoutDateTime = "2006-01-02T15:04:05Z07:00"

func (a *App) resolveInitProfile(root string, opts initOptions) (initProfile, error) {
	profile := a.detectInitProfile(root)
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
	language, framework := detectLanguageFramework(root)
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
		DevelopmentMode:       detectMethodology(root),
		ConversationLanguage:  locale,
		DocumentationLanguage: locale,
		CommentLanguage:       locale,
		GitMode:               "manual",
		GitProvider:           "github",
		GitLabInstanceURL:     "https://gitlab.com",
		AgentMode:             "single",
		StatusLinePreset:      "namba",
		UserName:              detectUserName(a.getenv),
		CreatedAt:             a.now().Format(timeLayoutDateTime),
	}
}

func applyInitOverrides(profile *initProfile, opts initOptions) {
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

func (a *App) runInitWizard(defaults initProfile) (initProfile, error) {
	reader := bufio.NewReader(a.stdin)
	profile := defaults

	renderInitBanner(a.stdout)
	fmt.Fprintln(a.stdout, wizardHeading(a.stdout, "\U0001f680 NambaAI \ucd08\uae30\ud654 \ub9c8\ubc95\uc0ac"))
	fmt.Fprintln(a.stdout, wizardHint(a.stdout, "MoAI init \ud750\ub984\uc744 Codex \ub124\uc774\ud2f0\ube0c \uc790\uc0b0\uc5d0 \ub9de\uac8c \uad6c\uc131\ud569\ub2c8\ub2e4."))
	fmt.Fprintln(a.stdout)

	profile.DevelopmentMode = promptSelect(
		a.stdin,
		a.stdout,
		"\U0001f9ea \uac1c\ubc1c \ubc29\ubc95\ub860",
		[]option{
			{Value: "tdd", Label: "TDD", Description: "\uc0c8 \uae30\ub2a5 RED-GREEN"},
			{Value: "ddd", Label: "DDD", Description: "\uae30\uc874 \ucf54\ub4dc \ubd84\uc11d/\uac1c\uc120"},
		},
		profile.DevelopmentMode,
	)
	profile.ProjectType = promptSelect(a.stdin, a.stdout, "\U0001f4e6 \ud504\ub85c\uc81d\ud2b8 \uc720\ud615", projectTypeOptions(), profile.ProjectType)
	profile = a.promptProjectScaffold(reader, profile)
	profile.ConversationLanguage = promptSelect(a.stdin, a.stdout, "\U0001f4ac \ub300\ud654 \uc5b8\uc5b4", languageOptions(), profile.ConversationLanguage)
	profile.DocumentationLanguage = promptSelect(a.stdin, a.stdout, "\U0001f4dd \ubb38\uc11c \uc5b8\uc5b4", languageOptions(), profile.DocumentationLanguage)
	profile.CommentLanguage = promptSelect(a.stdin, a.stdout, "\U0001f4bb \ucf54\ub4dc \uc8fc\uc11d \uc5b8\uc5b4", languageOptions(), profile.CommentLanguage)
	profile.AgentMode = promptSelect(
		a.stdin,
		a.stdout,
		"\U0001f916 Codex \uc5d0\uc774\uc804\ud2b8 \ubaa8\ub4dc",
		[]option{
			{Value: "single", Label: "\uc2f1\uae00", Description: "\uc548\uc815\uc801\uc778 \ub2e8\uc77c \ud750\ub984"},
			{Value: "multi", Label: "\uba40\ud2f0", Description: "\ubcd1\ub82c \uc791\uc5c5 \uc900\ube44"},
		},
		profile.AgentMode,
	)
	profile.StatusLinePreset = promptSelect(
		a.stdin,
		a.stdout,
		"\U0001f39b\ufe0f \uc0c1\ud0dc\uc904 \ud504\ub9ac\uc14b",
		[]option{
			{Value: "namba", Label: "Namba", Description: "\ud504\ub85c\uc81d\ud2b8 \uc911\uc2ec \ud45c\uc2dc"},
			{Value: "off", Label: "\ub044\uae30", Description: "\ucd94\ucc9c \uc124\uc815 \uc0dd\uc131 \uc548 \ud568"},
		},
		profile.StatusLinePreset,
	)
	profile.GitMode = promptSelect(
		a.stdin,
		a.stdout,
		"\U0001f33f Git \uc790\ub3d9\ud654 \ubaa8\ub4dc",
		[]option{
			{Value: "manual", Label: "\uc218\ub3d9", Description: "push/PR \uc790\ub3d9\ud654 \uc5c6\uc74c"},
			{Value: "personal", Label: "\uac1c\uc778", Description: "\ube0c\ub79c\uce58/\ucee4\ubc0b \ud5c8\uc6a9"},
			{Value: "team", Label: "\ud300", Description: "PR \uc900\ube44 \uc0b0\ucd9c\ubb3c \uc0dd\uc131"},
		},
		profile.GitMode,
	)
	if profile.GitMode != "manual" {
		profile.GitProvider = promptSelect(
			a.stdin,
			a.stdout,
			"\u2601\ufe0f Git \uc81c\uacf5\uc790",
			[]option{
				{Value: "github", Label: "GitHub", Description: "gh CLI \ub610\ub294 \uae30\uc874 \uc778\uc99d"},
				{Value: "gitlab", Label: "GitLab", Description: "glab CLI \ub610\ub294 \uae30\uc874 \uc778\uc99d"},
			},
			profile.GitProvider,
		)
		if profile.GitProvider == "gitlab" {
			profile.GitLabInstanceURL = promptInput(reader, a.stdout, "\U0001f517 GitLab \uc778\uc2a4\ud134\uc2a4 URL", profile.GitLabInstanceURL)
		}
		profile.GitUsername = promptInput(reader, a.stdout, "\U0001f464 Git \uc0ac\uc6a9\uc790\uba85", profile.GitUsername)
	}
	profile.UserName = promptInput(reader, a.stdout, "\U0001f64b \ud45c\uc2dc \uc774\ub984", profile.UserName)

	fmt.Fprintln(a.stdout)
	fmt.Fprintln(a.stdout, wizardHint(a.stdout, "\U0001f510 \ud1a0\ud070\uacfc \ube44\ubc00\uac12\uc740 \uc124\uc815 \ud30c\uc77c\uc5d0 \uc800\uc7a5\ud558\uc9c0 \uc54a\uc2b5\ub2c8\ub2e4. gh/glab login\uc744 \uc0ac\uc6a9\ud558\uc138\uc694."))
	return profile, nil
}

func (a *App) promptProjectScaffold(reader *bufio.Reader, profile initProfile) initProfile {
	profile.ProjectName = promptInput(reader, a.stdout, "\U0001f4db \ud504\ub85c\uc81d\ud2b8 \uc774\ub984", profile.ProjectName)

	if profile.ProjectType == "existing" {
		fmt.Fprintln(a.stdout, wizardHint(a.stdout, fmt.Sprintf("\U0001f50e \uac10\uc9c0\ub41c \uae30\ubcf8\uac12: \uc5b8\uc5b4=%s, \ud504\ub808\uc784\uc6cc\ud06c=%s", profile.Language, normalizeFramework(profile.Framework))))
		keepDetected := promptSelect(
			a.stdin,
			a.stdout,
			"\U0001f9ed \uc5b8\uc5b4/\ud504\ub808\uc784\uc6cc\ud06c \uc124\uc815",
			[]option{
				{Value: "keep", Label: "\uac10\uc9c0\uac12 \uc720\uc9c0", Description: "\ud604\uc7ac \uc800\uc7a5\uc18c \uae30\uc900 \uc0ac\uc6a9"},
				{Value: "override", Label: "\uc9c1\uc811 \uc120\ud0dd", Description: "\uc5b8\uc5b4\uc640 \ud504\ub808\uc784\uc6cc\ud06c \ub2e4\uc2dc \uace0\ub984"},
			},
			"keep",
		)
		if keepDetected == "keep" {
			return profile
		}
	}

	profile.Language = promptSelect(
		a.stdin,
		a.stdout,
		"\U0001f4a1 \uc8fc \uc0ac\uc6a9 \uc5b8\uc5b4",
		[]option{
			{Value: "go", Label: "Go", Description: "CLI/\uc11c\ube44\uc2a4"},
			{Value: "java", Label: "Java", Description: "JVM \uc571"},
			{Value: "typescript", Label: "TypeScript", Description: "Node/\ud504\ub860\ud2b8"},
			{Value: "python", Label: "Python", Description: "\uc2a4\ud06c\ub9bd\ud2b8/API"},
			{Value: "unknown", Label: "\ubbf8\uc815", Description: "\uc77c\ubc18 \ud504\ub85c\uc81d\ud2b8"},
		},
		profile.Language,
	)
	profile.Framework = promptSelect(a.stdin, a.stdout, "\U0001f9e9 \ud504\ub808\uc784\uc6cc\ud06c", frameworkOptions(profile.Language), normalizeFramework(profile.Framework))
	return profile
}
func promptInput(reader *bufio.Reader, out io.Writer, label, defaultValue string) string {
	prompt := wizardPrompt(out, label)
	if strings.TrimSpace(defaultValue) == "" {
		fmt.Fprintf(out, "%s: ", prompt)
	} else {
		fmt.Fprintf(out, "%s [%s]: ", prompt, defaultValue)
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		return strings.TrimSpace(defaultValue)
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return strings.TrimSpace(defaultValue)
	}
	return line
}

func renderInitBanner(out io.Writer) {
	fmt.Fprintln(out, " _   _    _    __  __ ____    _      ___ ")
	fmt.Fprintln(out, "| \\ | |  / \\  |  \\/  | __ )  / \\    |_ _|")
	fmt.Fprintln(out, "|  \\| | / _ \\ | |\\/| |  _ \\ / _ \\    | | ")
	fmt.Fprintln(out, "| |\\  |/ ___ \\| |  | | |_) / ___ \\   | | ")
	fmt.Fprintln(out, "|_| \\_/_/   \\_\\_|  |_|____/_/   \\_\\ |___|")
	fmt.Fprintln(out, wizardHint(out, "\u2728 \ud504\ub85c\uc81d\ud2b8 \uc124\uc815\uc744 \uc2dc\uc791\ud569\ub2c8\ub2e4"))
	fmt.Fprintln(out)
}

func wizardHeading(out io.Writer, text string) string {
	return styleWizardText(out, "1;36", text)
}

func wizardHint(out io.Writer, text string) string {
	return styleWizardText(out, "2;37", text)
}

func wizardPrompt(out io.Writer, text string) string {
	return styleWizardText(out, "1;33", text)
}

func wizardSelected(out io.Writer, text string) string {
	return styleWizardText(out, "1;32", text)
}

func styleWizardText(out io.Writer, code, text string) string {
	if !isTerminalWriter(out) {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func formatWizardChoice(choice option) string {
	if strings.TrimSpace(choice.Description) == "" {
		return choice.Label
	}
	return fmt.Sprintf("%s - %s", choice.Label, choice.Description)
}

func promptSelect(in io.Reader, out io.Writer, label string, choices []option, defaultValue string) string {
	if file, ok := in.(*os.File); ok && isTerminalReader(in) && isTerminalWriter(out) {
		if value, ok := promptSelectInteractive(file, out, label, choices, defaultValue); ok {
			return value
		}
	}

	reader := bufio.NewReader(in)
	return promptSelectLine(reader, out, label, choices, defaultValue)
}

func promptSelectLine(reader *bufio.Reader, out io.Writer, label string, choices []option, defaultValue string) string {
	fmt.Fprintln(out, wizardHeading(out, label))
	defaultIndex := 0
	for i, choice := range choices {
		if choice.Value == defaultValue {
			defaultIndex = i
		}
		fmt.Fprintf(out, "  %d. %s\n", i+1, formatWizardChoice(choice))
	}
	fmt.Fprintf(out, "%s [%d]: ", wizardPrompt(out, "\uc120\ud0dd"), defaultIndex+1)

	line, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultValue
	}
	index, err := strconv.Atoi(line)
	if err == nil && index >= 1 && index <= len(choices) {
		return choices[index-1].Value
	}
	for _, choice := range choices {
		if strings.EqualFold(choice.Value, line) || strings.EqualFold(choice.Label, line) {
			return choice.Value
		}
	}
	return defaultValue
}

type menuAction int

const (
	menuActionUnknown menuAction = iota
	menuActionUp
	menuActionDown
	menuActionSubmit
)

func promptSelectInteractive(in *os.File, out io.Writer, label string, choices []option, defaultValue string) (string, bool) {
	restoreInput, err := enableRawConsoleInput(in)
	if err != nil {
		return "", false
	}
	defer restoreInput()

	restoreOutput := enableVirtualTerminalOutput(out)
	defer restoreOutput()

	selected := 0
	for i, choice := range choices {
		if choice.Value == defaultValue {
			selected = i
			break
		}
	}

	reader := bufio.NewReader(in)
	lines := 0
	fmt.Fprint(out, "\x1b[?25l")
	defer fmt.Fprint(out, "\x1b[?25h")

	for {
		if lines > 0 {
			fmt.Fprintf(out, "\x1b[%dA", lines)
		}
		lines = renderInteractiveSelect(out, label, choices, selected)

		action, err := readMenuAction(reader)
		if err != nil {
			fmt.Fprintln(out)
			return defaultValue, true
		}

		switch action {
		case menuActionUp:
			selected = (selected - 1 + len(choices)) % len(choices)
		case menuActionDown:
			selected = (selected + 1) % len(choices)
		case menuActionSubmit:
			fmt.Fprintln(out)
			return choices[selected].Value, true
		}
	}
}

func renderInteractiveSelect(out io.Writer, label string, choices []option, selected int) int {
	lines := 0
	fmt.Fprintf(out, "\r\x1b[2K%s\n", wizardHeading(out, label))
	lines++
	fmt.Fprintf(out, "\r\x1b[2K%s\n", wizardHint(out, "\u2191/\u2193 \uc774\ub3d9 \u00b7 Enter \uc120\ud0dd"))
	lines++
	for i, choice := range choices {
		line := fmt.Sprintf("%d. %s", i+1, formatWizardChoice(choice))
		if i == selected {
			fmt.Fprintf(out, "\r\x1b[2K%s\n", wizardSelected(out, "\u276f "+line))
		} else {
			fmt.Fprintf(out, "\r\x1b[2K  %s\n", line)
		}
		lines++
	}
	return lines
}

func readMenuAction(reader *bufio.Reader) (menuAction, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return menuActionUnknown, err
	}

	switch b {
	case '\r', '\n':
		return menuActionSubmit, nil
	case 0x1b:
		next, err := reader.ReadByte()
		if err != nil {
			return menuActionUnknown, err
		}
		if next != '[' {
			return menuActionUnknown, nil
		}
		code, err := reader.ReadByte()
		if err != nil {
			return menuActionUnknown, err
		}
		switch code {
		case 'A':
			return menuActionUp, nil
		case 'B':
			return menuActionDown, nil
		default:
			return menuActionUnknown, nil
		}
	case 0x00, 0xe0:
		code, err := reader.ReadByte()
		if err != nil {
			return menuActionUnknown, err
		}
		switch code {
		case 72:
			return menuActionUp, nil
		case 80:
			return menuActionDown, nil
		default:
			return menuActionUnknown, nil
		}
	default:
		return menuActionUnknown, nil
	}
}

func projectTypeOptions() []option {
	return []option{
		{Value: "new", Label: "\uc0c8 \ud504\ub85c\uc81d\ud2b8", Description: "\ube48 \uc800\uc7a5\uc18c/\uc0c8 \ud3f4\ub354"},
		{Value: "existing", Label: "\uae30\uc874 \ud504\ub85c\uc81d\ud2b8", Description: "\ucf54\ub4dc\uac00 \uc788\ub294 \uc800\uc7a5\uc18c"},
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
		{Value: "ko", Label: "\ud55c\uad6d\uc5b4", Description: "ko"},
		{Value: "en", Label: "\uc601\uc5b4", Description: "en"},
		{Value: "ja", Label: "\uc77c\ubcf8\uc5b4", Description: "ja"},
		{Value: "zh", Label: "\uc911\uad6d\uc5b4", Description: "zh"},
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
	for _, value := range []string{profile.ConversationLanguage, profile.DocumentationLanguage, profile.CommentLanguage} {
		if !containsValue([]string{"en", "ko", "ja", "zh"}, value) {
			return fmt.Errorf("language preference %q is not supported", value)
		}
	}
	return nil
}

func containsValue(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func normalizeProjectName(name string) string {
	name = strings.TrimSpace(name)
	if name == "." || name == string(filepath.Separator) {
		return ""
	}
	return name
}

func normalizeFramework(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "none"
	}
	return value
}

func (a *App) isInteractiveTerminal() bool {
	file, ok := a.stdin.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil || (info.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	return isTerminalWriter(a.stdout)
}

func isTerminalReader(r io.Reader) bool {
	file, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func isTerminalWriter(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
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
	GitMode               string
	GitProvider           string
	GitUsername           string
	GitLabInstanceURL     string
	AgentMode             string
	StatusLinePreset      string
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
