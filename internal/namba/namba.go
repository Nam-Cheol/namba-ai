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
	Name      string
	Language  string
	Framework string
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
	case "plan":
		return a.runPlan(ctx, args[1:])
	case "run":
		return a.runExecute(ctx, args[1:])
	case "sync":
		return a.runSync(ctx, args[1:])
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
  namba init [path] [--yes] [--name NAME] [--mode tdd|ddd]
  namba doctor
  namba status
  namba project
  namba plan "<description>"
  namba run SPEC-XXX [--parallel] [--dry-run]
  namba sync
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

	testCmd, lintCmd, typecheckCmd := defaultQualityCommands(root, profile.Language)
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
	fmt.Fprintf(a.stdout, "Project: %s | Mode: %s | Agent mode: %s\n", profile.ProjectName, profile.DevelopmentMode, profile.AgentMode)
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
	fmt.Fprintf(a.stdout, "Language: %s\n", projectCfg.Language)
	fmt.Fprintf(a.stdout, "Framework: %s\n", projectCfg.Framework)
	fmt.Fprintf(a.stdout, "Mode: %s\n", qualityCfg.DevelopmentMode)
	fmt.Fprintf(a.stdout, "Codex native repo: %s\n", formatDoctorStatus(codexNativeIssues(root)))
	fmt.Fprintf(a.stdout, "Codex compatibility mirror: %s\n", formatDoctorStatus(codexCompatibilityIssues(root)))
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
	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New("plan requires a description")
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

	spec := fmt.Sprintf("# %s\n\n## Goal\n\n%s\n\n## Context\n\n- Project: %s\n- Language: %s\n- Mode: %s\n", specID, desc, projectCfg.Name, projectCfg.Language, qualityCfg.DevelopmentMode)
	plan := fmt.Sprintf("# %s Plan\n\n1. Refresh project context with `namba project`\n2. Implement the requested change\n3. Run validation commands\n4. Sync artifacts with `namba sync`\n", specID)
	acceptance := buildAcceptanceDoc(desc, qualityCfg.DevelopmentMode)

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

	summary := fmt.Sprintf("# Change Summary\n\nProject: %s\nLatest SPEC: %s\nGenerated: %s\n", projectCfg.Name, latestSpec, a.now().Format(time.RFC3339))
	checklist := "# PR Checklist\n\n- [ ] Project docs refreshed\n- [ ] SPEC artifacts reviewed\n- [ ] Validation commands passed\n- [ ] Diff reviewed\n"
	outputs := map[string]string{
		filepath.ToSlash(filepath.Join(projectDir, "change-summary.md")): summary,
		filepath.ToSlash(filepath.Join(projectDir, "pr-checklist.md")):   checklist,
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

func (a *App) runParallel(ctx context.Context, root string, specPkg specPackage, tasks []string, prompt string, qualityCfg qualityConfig, systemCfg systemConfig, dryRun bool) error {
	if _, err := a.lookPath("git"); err != nil {
		return errors.New("parallel execution requires git")
	}
	if !isGitRepository(root) {
		return errors.New("parallel execution requires a git repository")
	}

	baseBranch, err := a.currentBranch(ctx, root)
	if err != nil {
		return err
	}
	workers := minInt(len(tasks), 3)
	if workers == 0 {
		workers = 1
	}
	chunks := chunkTasks(tasks, workers)

	type workResult struct {
		name string
		err  error
	}
	results := make([]workResult, 0, len(chunks))

	for i, chunk := range chunks {
		name := strings.ToLower(specPkg.ID) + "-p" + strconv.Itoa(i+1)
		path := filepath.Join(root, worktreesDir, name)
		branch := "namba/" + name
		if _, err := a.runBinary(ctx, "git", []string{"worktree", "add", "-b", branch, path, baseBranch}, root); err != nil {
			return err
		}

		workerPrompt := prompt + "\n\n## Assigned work package\n\n" + strings.Join(chunk, "\n")
		logPath := filepath.Join(root, logsDir, "runs", name+"-request.md")
		if err := os.WriteFile(logPath, []byte(workerPrompt), 0o644); err != nil {
			return err
		}

		if dryRun {
			results = append(results, workResult{name: name})
			continue
		}

		request := a.newExecutionRequest(specPkg.ID, path, workerPrompt, systemCfg)
		_, _, err := a.executeRun(ctx, root, name, request, path, qualityCfg)
		results = append(results, workResult{name: name, err: err})
	}

	for _, result := range results {
		if result.err != nil {
			return fmt.Errorf("parallel worker %s failed: %w", result.name, result.err)
		}
		if dryRun {
			continue
		}
		branch := "namba/" + result.name
		if _, err := a.runBinary(ctx, "git", []string{"merge", "--no-ff", branch, "-m", "merge " + branch}, root); err != nil {
			return fmt.Errorf("merge %s: %w", branch, err)
		}
	}

	if dryRun {
		fmt.Fprintf(a.stdout, "Prepared %d parallel work packages for %s\n", len(results), specPkg.ID)
		return nil
	}
	fmt.Fprintf(a.stdout, "Executed %s in %d parallel worktrees with %s\n", specPkg.ID, len(results), normalizeRunner(systemCfg.Runner))
	return nil
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
		Name:      values["name"],
		Language:  values["language"],
		Framework: values["framework"],
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
	case exists(filepath.Join(root, "package.json")):
		return "typescript", detectNodeFramework(root)
	case exists(filepath.Join(root, "pyproject.toml")) || exists(filepath.Join(root, "requirements.txt")):
		return "python", "none"
	default:
		return "unknown", "none"
	}
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
		case ".go", ".js", ".ts", ".tsx", ".py", ".rs":
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

func defaultQualityCommands(root, language string) (string, string, string) {
	switch language {
	case "go":
		return "go test ./...", defaultGoFormatCommand(root), "go vet ./..."
	case "typescript":
		return "npm test", "npm run lint", "npm run typecheck"
	case "python":
		return "pytest", "ruff check .", "none"
	default:
		return "none", "none", "none"
	}
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
	flow := "# Data Flow\n\n1. `init` runs a Codex-adapted project wizard, writes `.namba/config/sections/*.yaml`, repo skills under `.agents/skills`, role cards under `.codex/agents`, a compatibility mirror under `.codex/skills`, and Codex repo config under `.codex/config.toml`\n2. `project` refreshes docs and codemaps\n3. `plan` creates a SPEC package\n4. `run` either builds a non-interactive Codex execution request or is interpreted as Codex-native in-session execution\n5. `sync` emits PR-ready artifacts\n"
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

	fmt.Fprintln(a.stdout, "NambaAI init wizard (Codex edition)")
	fmt.Fprintln(a.stdout, "This adapts the MoAI init flow to Codex-native assets.")

	profile.ProjectName = promptInput(reader, a.stdout, "Project name", profile.ProjectName)
	profile.DevelopmentMode = promptSelect(
		reader,
		a.stdout,
		"Development methodology",
		[]option{
			{Value: "tdd", Label: "TDD", Description: "RED-GREEN-REFACTOR for new work"},
			{Value: "ddd", Label: "DDD", Description: "ANALYZE-PRESERVE-IMPROVE for brownfield"},
		},
		profile.DevelopmentMode,
	)
	profile.Language = promptSelect(
		reader,
		a.stdout,
		"Primary language",
		[]option{
			{Value: "go", Label: "Go", Description: "Compiled CLI and services"},
			{Value: "typescript", Label: "TypeScript", Description: "Node.js and frontend projects"},
			{Value: "python", Label: "Python", Description: "Scripting and backend"},
			{Value: "unknown", Label: "Unknown", Description: "Leave as a generic project"},
		},
		profile.Language,
	)
	profile.Framework = promptSelect(reader, a.stdout, "Framework", frameworkOptions(profile.Language), normalizeFramework(profile.Framework))
	profile.ConversationLanguage = promptSelect(reader, a.stdout, "Conversation language", languageOptions(), profile.ConversationLanguage)
	profile.DocumentationLanguage = promptSelect(reader, a.stdout, "Documentation language", languageOptions(), profile.DocumentationLanguage)
	profile.CommentLanguage = promptSelect(reader, a.stdout, "Comment language", languageOptions(), profile.CommentLanguage)
	profile.AgentMode = promptSelect(
		reader,
		a.stdout,
		"Codex agent mode",
		[]option{
			{Value: "single", Label: "Single-agent", Description: "Recommended for stable serial execution"},
			{Value: "multi", Label: "Multi-agent", Description: "Prepare repo-local agent role cards and higher concurrency"},
		},
		profile.AgentMode,
	)
	profile.StatusLinePreset = promptSelect(
		reader,
		a.stdout,
		"Status line preset",
		[]option{
			{Value: "namba", Label: "Namba", Description: "Project-focused Codex status line"},
			{Value: "off", Label: "Off", Description: "Do not write a repo status line recommendation"},
		},
		profile.StatusLinePreset,
	)
	profile.GitMode = promptSelect(
		reader,
		a.stdout,
		"Git automation mode",
		[]option{
			{Value: "manual", Label: "Manual", Description: "AI never pushes or opens PRs"},
			{Value: "personal", Label: "Personal", Description: "AI may branch and commit locally"},
			{Value: "team", Label: "Team", Description: "AI prepares branch and PR-ready artifacts"},
		},
		profile.GitMode,
	)
	if profile.GitMode != "manual" {
		profile.GitProvider = promptSelect(
			reader,
			a.stdout,
			"Git provider",
			[]option{
				{Value: "github", Label: "GitHub", Description: "Use gh CLI or existing credentials"},
				{Value: "gitlab", Label: "GitLab", Description: "Use glab CLI or existing credentials"},
			},
			profile.GitProvider,
		)
		if profile.GitProvider == "gitlab" {
			profile.GitLabInstanceURL = promptInput(reader, a.stdout, "GitLab instance URL", profile.GitLabInstanceURL)
		}
		profile.GitUsername = promptInput(reader, a.stdout, "Git username", profile.GitUsername)
	}
	profile.UserName = promptInput(reader, a.stdout, "Display name", profile.UserName)

	fmt.Fprintln(a.stdout, "Note: tokens and secrets are not stored in NambaAI config. Use gh/glab login instead.")
	return profile, nil
}

func promptInput(reader *bufio.Reader, out io.Writer, label, defaultValue string) string {
	fmt.Fprintf(out, "%s [%s]: ", label, defaultValue)
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

func promptSelect(reader *bufio.Reader, out io.Writer, label string, choices []option, defaultValue string) string {
	fmt.Fprintln(out, label)
	defaultIndex := 0
	for i, choice := range choices {
		if choice.Value == defaultValue {
			defaultIndex = i
		}
		fmt.Fprintf(out, "  %d. %s - %s\n", i+1, choice.Label, choice.Description)
	}
	fmt.Fprintf(out, "Select [%d]: ", defaultIndex+1)

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

func frameworkOptions(language string) []option {
	switch language {
	case "go":
		return []option{
			{Value: "none", Label: "None", Description: "Keep a generic Go project"},
			{Value: "cobra", Label: "Cobra", Description: "CLI application"},
			{Value: "gin", Label: "Gin", Description: "HTTP service"},
			{Value: "echo", Label: "Echo", Description: "HTTP service"},
		}
	case "typescript":
		return []option{
			{Value: "none", Label: "None", Description: "Keep a generic Node project"},
			{Value: "nextjs", Label: "Next.js", Description: "React full-stack"},
			{Value: "react", Label: "React", Description: "Client application"},
			{Value: "nest", Label: "NestJS", Description: "Backend service"},
		}
	case "python":
		return []option{
			{Value: "none", Label: "None", Description: "Keep a generic Python project"},
			{Value: "fastapi", Label: "FastAPI", Description: "Modern API service"},
			{Value: "django", Label: "Django", Description: "Full-stack web framework"},
		}
	default:
		return []option{
			{Value: "none", Label: "None", Description: "No framework selected"},
		}
	}
}

func languageOptions() []option {
	return []option{
		{Value: "en", Label: "English", Description: "English"},
		{Value: "ko", Label: "Korean", Description: "Korean"},
		{Value: "ja", Label: "Japanese", Description: "Japanese"},
		{Value: "zh", Label: "Chinese", Description: "Chinese"},
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
	if !containsValue([]string{"go", "typescript", "python", "unknown"}, profile.Language) {
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
