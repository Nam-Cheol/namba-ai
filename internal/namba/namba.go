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
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

	manifestOwnerManaged = "namba-managed"
)

type App struct {
	stdin                   io.Reader
	stdout                  io.Writer
	stderr                  io.Writer
	now                     func() time.Time
	getenv                  func(string) string
	getwd                   func() (string, error)
	readFile                func(string) ([]byte, error)
	writeFile               func(string, []byte, fs.FileMode) error
	mkdirAll                func(string, fs.FileMode) error
	lookPath                func(string) (string, error)
	detectCodexCapabilities func(context.Context, string, executionRequest) (codexCapabilityMatrix, error)
	runCmd                  func(context.Context, string, []string, string) (string, error)
	startCmd                func(string, []string, string) error
	downloadURL             func(context.Context, string) ([]byte, error)
	executablePath          func() (string, error)
	writeManifestOverride   func(string, Manifest) error
	newParallelProgressSink func(parallelProgressSinkConfig) (parallelProgressSink, error)
	goos                    string
	goarch                  string
}

type Manifest struct {
	GeneratedAt string          `json:"generated_at"`
	Entries     []ManifestEntry `json:"entries"`
}

type ManifestEntry struct {
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	Owner     string `json:"owner,omitempty"`
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
	DevelopmentMode        string
	TestCommand            string
	LintCommand            string
	TypecheckCommand       string
	BuildCommand           string
	MigrationDryRunCommand string
	SmokeStartCommand      string
	OutputContractCommand  string
}

type docsConfig struct {
	ManageReadme           bool
	ReadmeProfile          string
	DefaultLanguage        string
	AdditionalLanguages    []string
	AdditionalLanguagesSet bool
	HeroImage              string
}

type packageManifest struct {
	Name             string            `json:"name"`
	Dependencies     map[string]string `json:"dependencies"`
	DevDependencies  map[string]string `json:"devDependencies"`
	PeerDependencies map[string]string `json:"peerDependencies"`
}

type specPackage struct {
	ID          string
	Description string
	Path        string
}

type topLevelCommandDefinition struct {
	Name         string
	UsageSummary string
	UsageText    func() string
	Run          func(*App, context.Context, []string) error
}

type worktreeSubcommandDefinition struct {
	Name         string
	UsageSummary string
	Run          func(*App, context.Context, string, []string) error
}

type fixSubcommandDefinition struct {
	Name            string
	BehaviorSummary string
	Run             func(*App, context.Context, string, fixInvocation) error
}

type topLevelInvocation struct {
	UsageText string
	Command   topLevelCommandDefinition
	Args      []string
}

func NewApp(stdout, stderr io.Writer) *App {
	return &App{
		stdin:     os.Stdin,
		stdout:    stdout,
		stderr:    stderr,
		now:       time.Now,
		getenv:    os.Getenv,
		getwd:     os.Getwd,
		readFile:  os.ReadFile,
		writeFile: os.WriteFile,
		mkdirAll:  os.MkdirAll,
		lookPath:  exec.LookPath,
		newParallelProgressSink: func(cfg parallelProgressSinkConfig) (parallelProgressSink, error) {
			return newJSONLParallelProgressSink(cfg)
		},
		executablePath: os.Executable,
		goos:           runtime.GOOS,
		goarch:         runtime.GOARCH,
		runCmd: func(ctx context.Context, name string, args []string, dir string) (string, error) {
			cmd := exec.CommandContext(ctx, name, args...)
			cmd.Dir = dir
			output, err := cmd.CombinedOutput()
			return strings.TrimSpace(string(output)), err
		},
		startCmd: func(name string, args []string, dir string) error {
			cmd := exec.Command(name, args...)
			cmd.Dir = dir
			return cmd.Start()
		},
		downloadURL: func(ctx context.Context, url string) ([]byte, error) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("User-Agent", "NambaAI-Updater")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
				return nil, fmt.Errorf("download %s failed with status %d: %s", url, resp.StatusCode, strings.TrimSpace(string(body)))
			}
			return io.ReadAll(resp.Body)
		},
	}
}

func (a *App) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return a.printUsage()
	}
	invocation, err := a.resolveTopLevelInvocation(args)
	if err != nil {
		return err
	}
	return a.runTopLevelInvocation(ctx, invocation)
}

func (a *App) printUsage() error {
	_, err := fmt.Fprint(a.stdout, usageText())
	return err
}

func usageText() string {
	lines := []string{
		"NambaAI CLI",
		"",
		"Usage:",
		"  namba help [command]",
	}
	lines = append(lines, publicTopLevelCommandUsageSummaries()...)
	return strings.Join(lines, "\n") + "\n"
}

func parseTopLevelHelpTopic(args []string) (string, bool, error) {
	switch args[0] {
	case "help", "-h", "--help":
		if len(args) > 2 {
			return "", false, fmt.Errorf("help accepts at most one command\n\n%s", usageText())
		}
		if len(args) == 1 {
			return "", true, nil
		}
		return normalizeCommandName(args[1]), true, nil
	default:
		return "", false, nil
	}
}

func (a *App) resolveTopLevelInvocation(args []string) (topLevelInvocation, error) {
	if topic, ok, err := parseTopLevelHelpTopic(args); err != nil {
		return topLevelInvocation{}, err
	} else if ok {
		if topic == "" {
			return topLevelInvocation{UsageText: usageText()}, nil
		}
		text, ok := commandUsageText(topic)
		if !ok {
			return topLevelInvocation{}, unknownTopLevelCommandError(topic)
		}
		return topLevelInvocation{UsageText: text}, nil
	}

	command, ok := a.resolveTopLevelCommand(args[0])
	if !ok {
		return topLevelInvocation{}, unknownTopLevelCommandError(args[0])
	}
	return topLevelInvocation{Command: command, Args: args[1:]}, nil
}

func (a *App) runTopLevelInvocation(ctx context.Context, invocation topLevelInvocation) error {
	if invocation.UsageText != "" {
		_, err := fmt.Fprint(a.stdout, invocation.UsageText)
		return err
	}
	return invocation.Command.Run(a, ctx, invocation.Args)
}

func normalizeCommandName(name string) string {
	switch name {
	case "", "help", "-h", "--help":
		return ""
	default:
		return strings.TrimSpace(name)
	}
}

func (a *App) printCommandUsage(command string) error {
	if command == "" {
		return a.printUsage()
	}
	text, ok := commandUsageText(command)
	if !ok {
		return unknownTopLevelCommandError(command)
	}
	_, err := fmt.Fprint(a.stdout, text)
	return err
}

func unknownTopLevelCommandError(command string) error {
	return fmt.Errorf("unknown command %q\n\n%s", command, usageText())
}

func publicTopLevelCommandDefinitions() []topLevelCommandDefinition {
	return []topLevelCommandDefinition{
		{Name: "init", UsageSummary: "  namba init [path] [--yes] [--name NAME] [--mode tdd|ddd] [--project-type new|existing]", UsageText: initUsageText, Run: (*App).runInit},
		{Name: "doctor", UsageSummary: "  namba doctor", UsageText: doctorUsageText, Run: (*App).runDoctor},
		{Name: "status", UsageSummary: "  namba status", UsageText: statusUsageText, Run: (*App).runStatus},
		{Name: "project", UsageSummary: "  namba project", UsageText: projectUsageText, Run: (*App).runProject},
		{Name: "update", UsageSummary: "  namba update [--version vX.Y.Z]", UsageText: updateUsageText, Run: (*App).runUpdate},
		{Name: "regen", UsageSummary: "  namba regen", UsageText: regenUsageText, Run: (*App).runRegen},
		{Name: "codex", UsageSummary: "  namba codex access [--approval-policy POLICY --sandbox-mode MODE]", UsageText: codexUsageText, Run: (*App).runCodex},
		{Name: "plan", UsageSummary: "  namba plan \"<description>\"", UsageText: planUsageText, Run: (*App).runPlan},
		{Name: "harness", UsageSummary: "  namba harness \"<description>\"", UsageText: harnessUsageText, Run: (*App).runHarness},
		{Name: "fix", UsageSummary: "  namba fix [--command run|plan] \"<issue description>\"", UsageText: fixUsageText, Run: (*App).runFix},
		{Name: "run", UsageSummary: "  namba run SPEC-XXX [--solo|--team|--parallel] [--dry-run]", UsageText: runUsageText, Run: (*App).runExecute},
		{Name: "sync", UsageSummary: "  namba sync", UsageText: syncUsageText, Run: (*App).runSync},
		{Name: "pr", UsageSummary: "  namba pr \"<title>\" [--remote origin] [--no-sync] [--no-validate]", UsageText: prUsageText, Run: (*App).runPR},
		{Name: "land", UsageSummary: "  namba land [PR_NUMBER] [--wait] [--remote origin]", UsageText: landUsageText, Run: (*App).runLand},
		{Name: "release", UsageSummary: "  namba release [--bump patch|minor|major] [--version vX.Y.Z] [--push] [--remote origin]", UsageText: releaseUsageText, Run: (*App).runRelease},
		{Name: "worktree", UsageSummary: "  namba worktree <new|list|remove|clean>", UsageText: worktreeUsageText, Run: (*App).runWorktree},
	}
}

func publicTopLevelCommandUsageSummaries() []string {
	lines := make([]string, 0, len(publicTopLevelCommandDefinitions()))
	for _, definition := range publicTopLevelCommandDefinitions() {
		lines = append(lines, definition.UsageSummary)
	}
	return lines
}

func (a *App) topLevelCommandDefinitions() []topLevelCommandDefinition {
	definitions := append([]topLevelCommandDefinition{}, publicTopLevelCommandDefinitions()...)
	return append(definitions, topLevelCommandDefinition{
		Name: internalCreateCommandName,
		Run:  (*App).runInternalCreate,
	})
}

func (a *App) resolveTopLevelCommand(command string) (topLevelCommandDefinition, bool) {
	for _, definition := range a.topLevelCommandDefinitions() {
		if definition.Name == command {
			return definition, true
		}
	}
	return topLevelCommandDefinition{}, false
}

func commandUsageText(command string) (string, bool) {
	for _, definition := range publicTopLevelCommandDefinitions() {
		if definition.Name == command && definition.UsageText != nil {
			return definition.UsageText(), true
		}
	}
	return "", false
}

func wantsCommandHelp(args []string) bool {
	return len(args) == 1 && isHelpToken(args[0])
}

func (a *App) handleNoArgTopLevelCommand(command string, args []string) (bool, error) {
	switch {
	case wantsCommandHelp(args):
		return true, a.printCommandUsage(command)
	case len(args) != 0:
		return true, commandUsageError(command, fmt.Errorf("%s does not accept arguments", command))
	default:
		return false, nil
	}
}

func isHelpToken(arg string) bool {
	switch strings.TrimSpace(arg) {
	case "--help", "-h":
		return true
	default:
		return false
	}
}

func commandUsageError(command string, err error) error {
	if err == nil {
		return nil
	}
	text, ok := commandUsageText(command)
	if !ok {
		return err
	}
	return fmt.Errorf("%s\n\n%s", err.Error(), text)
}

func initUsageText() string {
	lines := []string{
		"namba init",
		"",
		"Usage:",
		"  namba init [path] [--yes] [--name NAME] [--mode tdd|ddd] [--project-type new|existing]",
		"  namba init [path] [--human-language LANG] [--approval-policy POLICY] [--sandbox-mode MODE]",
		"",
		"Behavior:",
		"  Initialize the NambaAI scaffold, config, and repo-local Codex assets in the target directory.",
		"  The wizard leads with Codex access presets and previews the resulting approval_policy / sandbox_mode pair.",
		"  After bootstrap, use `namba codex access` from the project root to inspect or change repo-owned Codex access defaults.",
	}
	return strings.Join(lines, "\n") + "\n"
}

func doctorUsageText() string {
	return singleUsageLineCommandUsageText(
		"doctor",
		"  namba doctor",
		"  Inspect the current repository and local toolchain readiness without mutating project files.",
	)
}

func statusUsageText() string {
	return singleUsageLineCommandUsageText(
		"status",
		"  namba status",
		"  Print a read-only summary of the current NambaAI repository state.",
	)
}

func projectUsageText() string {
	return singleUsageLineCommandUsageText(
		"project",
		"  namba project",
		"  Refresh .namba/project/* docs and codemaps for the current repository.",
	)
}

func regenUsageText() string {
	return singleUsageLineCommandUsageText(
		"regen",
		"  namba regen",
		"  Regenerate AGENTS, repo-local skills, Codex agents, and Codex config from .namba/config/sections/*.yaml.",
	)
}

func updateUsageText() string {
	return singleUsageLineCommandUsageText(
		"update",
		"  namba update [--version vX.Y.Z]",
		"  Download and install the requested NambaAI release for the current platform.",
	)
}

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
	for rel, scaffold := range codexScaffoldFiles(profile) {
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
	return nil
}

func (a *App) runDoctor(ctx context.Context, args []string) error {
	if handled, err := a.handleNoArgTopLevelCommand("doctor", args); handled {
		return err
	}

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

func (a *App) runStatus(_ context.Context, args []string) error {
	if handled, err := a.handleNoArgTopLevelCommand("status", args); handled {
		return err
	}

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

func (a *App) runProject(_ context.Context, args []string) error {
	if handled, err := a.handleNoArgTopLevelCommand("project", args); handled {
		return err
	}

	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}

	projectCfg, _ := a.loadProjectConfig(root)
	qualityCfg, _ := a.loadQualityConfig(root)
	analysisCfg, err := a.loadAnalysisConfig(root)
	if err != nil {
		return err
	}

	analysis := analyzeProject(root, projectCfg, qualityCfg, analysisCfg)
	outputs := analysis.renderOutputs()
	if _, err := a.replaceManagedOutputs(root, outputs, isProjectAnalysisManagedPath, nil); err != nil {
		return err
	}

	for _, warning := range analysis.Quality.Warnings {
		fmt.Fprintf(a.stdout, "Project analysis warning: %s\n", warning)
	}
	if len(analysis.Quality.Errors) > 0 {
		for _, item := range analysis.Quality.Errors {
			fmt.Fprintf(a.stdout, "Project analysis error: %s\n", item)
		}
		return errors.New("project analysis quality gate failed")
	}
	fmt.Fprintln(a.stdout, "Refreshed NambaAI project docs and codemaps.")
	return nil
}

func (a *App) runPlan(ctx context.Context, args []string) error {
	options, err := parseDescriptionCommandArgs("plan", "description", args)
	if err != nil {
		return commandUsageError("plan", err)
	}
	if options.help {
		return a.printPlanUsage()
	}
	return a.createSpecPackage(ctx, "plan", options.description, options.currentWorkspace)
}

func (a *App) runHarness(ctx context.Context, args []string) error {
	options, err := parseDescriptionCommandArgs("harness", "description", args)
	if err != nil {
		return commandUsageError("harness", err)
	}
	if options.help {
		return a.printHarnessUsage()
	}
	return a.createSpecPackage(ctx, "harness", options.description, options.currentWorkspace)
}

func (a *App) runFix(ctx context.Context, args []string) error {
	options, err := parseFixArgs(args)
	if err != nil {
		return commandUsageError("fix", err)
	}
	if options.help {
		return a.printFixUsage()
	}

	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}

	subcommand, ok := a.resolveFixSubcommand(options.command)
	if !ok {
		return fmt.Errorf("invalid fix command %q", options.command)
	}
	return subcommand.Run(a, ctx, root, options)
}

func (a *App) createSpecPackage(ctx context.Context, kind, description string, currentWorkspace bool) error {
	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}

	start, err := a.resolvePlanningStart(ctx, root, planningStartOptions{
		Kind:             kind,
		Description:      description,
		CurrentWorkspace: currentWorkspace,
	})
	if err != nil {
		if start.SpecID != "" {
			return errors.New(formatPlanningStartSummary(start))
		}
		return err
	}

	scaffoldCtx, err := a.loadResolvedSpecPackageScaffoldContext(start.Root, kind, description, start.SpecID)
	if err != nil {
		return wrapPlanningScaffoldFailure(start, err)
	}
	outputs := buildSpecPackageScaffoldOutputs(scaffoldCtx)
	if err := a.materializeSpecPackageScaffoldOutputs(scaffoldCtx, outputs); err != nil {
		return wrapPlanningScaffoldFailure(start, err)
	}

	fmt.Fprint(a.stdout, formatPlanningStartSummary(start))
	fmt.Fprintf(a.stdout, "Created %s\n", scaffoldCtx.SpecID)
	return nil
}

type planInvocation struct {
	help             bool
	currentWorkspace bool
	description      string
}

type fixInvocation struct {
	help             bool
	currentWorkspace bool
	command          string
	description      string
}

func parsePlanArgs(args []string) (planInvocation, error) {
	return parseDescriptionCommandArgs("plan", "description", args)
}

func parseDescriptionCommandArgs(command, field string, args []string) (planInvocation, error) {
	if len(args) == 0 {
		return planInvocation{}, fmt.Errorf("%s requires a %s", command, field)
	}
	invocation := planInvocation{}
	descriptionParts := make([]string, 0, len(args))
	afterDelimiter := false
	for _, arg := range args {
		if afterDelimiter {
			descriptionParts = append(descriptionParts, arg)
			continue
		}
		switch arg {
		case "--":
			afterDelimiter = true
		case "--help", "-h":
			return planInvocation{help: true}, nil
		case currentWorkspacePlanningFlag:
			invocation.currentWorkspace = true
		default:
			if isStandaloneFlagToken(arg) {
				return planInvocation{}, fmt.Errorf("unknown flag %q", arg)
			}
			descriptionParts = append(descriptionParts, arg)
		}
	}
	description := strings.TrimSpace(strings.Join(descriptionParts, " "))
	if description == "" {
		return planInvocation{}, fmt.Errorf("%s requires a %s", command, field)
	}
	invocation.description = description
	return invocation, nil
}

func parseFixArgs(args []string) (fixInvocation, error) {
	if len(args) == 0 {
		return fixInvocation{}, errors.New("fix requires an issue description")
	}

	invocation := fixInvocation{command: "run"}
	var descriptionParts []string
	afterDelimiter := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if afterDelimiter {
			descriptionParts = append(descriptionParts, arg)
			continue
		}
		switch arg {
		case "--":
			afterDelimiter = true
		case "--help", "-h":
			return fixInvocation{help: true}, nil
		case currentWorkspacePlanningFlag:
			invocation.currentWorkspace = true
		case "--command":
			if i+1 >= len(args) {
				return fixInvocation{}, errors.New("fix --command requires a value of run or plan")
			}
			value := strings.TrimSpace(args[i+1])
			if value != "run" && value != "plan" {
				return fixInvocation{}, fmt.Errorf("invalid fix --command %q: expected run or plan", value)
			}
			invocation.command = value
			i++
		case "--command=run", "--command=plan":
			invocation.command = strings.TrimPrefix(arg, "--command=")
		default:
			if isStandaloneFlagToken(arg) {
				return fixInvocation{}, fmt.Errorf("unknown flag %q", arg)
			}
			descriptionParts = append(descriptionParts, arg)
		}
	}

	invocation.description = strings.TrimSpace(strings.Join(descriptionParts, " "))
	if invocation.description == "" {
		return fixInvocation{}, errors.New("fix requires an issue description")
	}
	if invocation.currentWorkspace && invocation.command != "plan" {
		return fixInvocation{}, fmt.Errorf("%s is only valid with --command plan", currentWorkspacePlanningFlag)
	}
	return invocation, nil
}

func fixSubcommandDefinitions() []fixSubcommandDefinition {
	return []fixSubcommandDefinition{
		{Name: "plan", BehaviorSummary: "  Use --command plan to scaffold the next bugfix SPEC package under .namba/specs/.", Run: (*App).runFixPlanSubcommand},
		{Name: "run", BehaviorSummary: "  Use --command run, or omit --command, to repair the issue directly in the current workspace.", Run: (*App).runFixRunSubcommand},
	}
}

func fixSubcommandBehaviorSummaries() []string {
	lines := make([]string, 0, len(fixSubcommandDefinitions()))
	for _, definition := range fixSubcommandDefinitions() {
		lines = append(lines, definition.BehaviorSummary)
	}
	return lines
}

func (a *App) resolveFixSubcommand(name string) (fixSubcommandDefinition, bool) {
	for _, definition := range fixSubcommandDefinitions() {
		if definition.Name == name {
			return definition, true
		}
	}
	return fixSubcommandDefinition{}, false
}

func (a *App) runFixPlanSubcommand(ctx context.Context, _ string, options fixInvocation) error {
	return a.createSpecPackage(ctx, "fix", options.description, options.currentWorkspace)
}

func (a *App) runFixRunSubcommand(ctx context.Context, root string, options fixInvocation) error {
	return a.executeDirectFix(ctx, root, options.description)
}

func isStandaloneFlagToken(arg string) bool {
	trimmed := strings.TrimSpace(arg)
	if !strings.HasPrefix(trimmed, "-") {
		return false
	}
	return !strings.ContainsAny(trimmed, " \t\r\n")
}

func (a *App) printPlanUsage() error {
	_, err := fmt.Fprint(a.stdout, planUsageText())
	return err
}

func (a *App) printHarnessUsage() error {
	_, err := fmt.Fprint(a.stdout, harnessUsageText())
	return err
}

func (a *App) printFixUsage() error {
	_, err := fmt.Fprint(a.stdout, fixUsageText())
	return err
}

func planUsageText() string {
	return descriptionScaffoldUsageText(
		"plan",
		"  Create the next feature SPEC package under .namba/specs/ and seed review artifacts.",
		fmt.Sprintf("  Safe by default: create and switch to a dedicated SPEC branch in the current workspace unless you explicitly pass %s.", currentWorkspacePlanningFlag),
	)
}

func harnessUsageText() string {
	return descriptionScaffoldUsageText(
		"harness",
		"  Create the next harness-oriented SPEC package under .namba/specs/ and seed review artifacts.",
		fmt.Sprintf("  Safe by default: create and switch to a dedicated SPEC branch in the current workspace unless you explicitly pass %s.", currentWorkspacePlanningFlag),
	)
}

func descriptionScaffoldUsageText(command string, behaviorLines ...string) string {
	lines := []string{
		fmt.Sprintf("namba %s", command),
		"",
		"Usage:",
	}
	lines = append(lines, descriptionScaffoldUsageLines("namba "+command, "description")...)
	lines = append(lines, "", "Behavior:")
	lines = append(lines, behaviorLines...)
	return strings.Join(lines, "\n") + "\n"
}

func descriptionScaffoldUsageLines(invocation, subject string) []string {
	return []string{
		fmt.Sprintf("  %s \"<%s>\"", invocation, subject),
		fmt.Sprintf("  %s -- \"<%s with flag-like text>\"", invocation, subject),
		fmt.Sprintf("  %s %s \"<%s>\"", invocation, currentWorkspacePlanningFlag, subject),
	}
}

func fixUsageText() string {
	lines := []string{
		"namba fix",
		"",
		"Usage:",
	}
	lines = append(lines,
		"  namba fix [--command run|plan] \"<issue description>\"",
		"  namba fix [--command run|plan] -- \"<issue description with flag-like text>\"",
		fmt.Sprintf("  namba fix --command plan %s \"<issue description>\"", currentWorkspacePlanningFlag),
	)
	lines = append(lines, "", "Behavior:")
	lines = append(lines, fixSubcommandBehaviorSummaries()...)
	lines = append(lines, fmt.Sprintf("  Use %s with --command plan when you intentionally want to scaffold on the current branch without creating a dedicated SPEC branch.", currentWorkspacePlanningFlag))
	return strings.Join(lines, "\n") + "\n"
}

func singleUsageLineCommandUsageText(command, usageLine, behaviorLine string) string {
	lines := []string{
		fmt.Sprintf("namba %s", command),
		"",
		"Usage:",
		usageLine,
		"",
		"Behavior:",
		behaviorLine,
	}
	return strings.Join(lines, "\n") + "\n"
}

func runUsageText() string {
	return singleUsageLineCommandUsageText(
		"run",
		"  namba run SPEC-XXX [--solo|--team|--parallel] [--dry-run]",
		"  Execute the selected SPEC package with one runner, same-workspace team routing, or managed worktree fan-out.",
	)
}

func syncUsageText() string {
	return singleUsageLineCommandUsageText(
		"sync",
		"  namba sync",
		"  Refresh README bundles, project docs, review readiness summaries, and PR/release support artifacts.",
	)
}

func prUsageText() string {
	return singleUsageLineCommandUsageText(
		"pr",
		"  namba pr \"<title>\" [--remote origin] [--no-sync] [--no-validate]",
		"  Sync, validate, push the current work branch, and create or reuse a GitHub pull request into the base branch.",
	)
}

func landUsageText() string {
	return singleUsageLineCommandUsageText(
		"land",
		"  namba land [PR_NUMBER] [--wait] [--remote origin]",
		"  Merge an approved pull request into the base branch and refresh the local base branch checkout.",
	)
}

func releaseUsageText() string {
	return singleUsageLineCommandUsageText(
		"release",
		"  namba release [--bump patch|minor|major] [--version vX.Y.Z] [--push] [--remote origin]",
		"  Create a release tag from a clean main branch and optionally push main plus the tag.",
	)
}

func worktreeUsageText() string {
	lines := []string{
		"namba worktree",
		"",
		"Usage:",
	}
	lines = append(lines, worktreeSubcommandUsageSummaries()...)
	lines = append(lines,
		"",
		"Behavior:",
		"  Manage Namba-owned git worktrees under .namba/worktrees.",
	)
	return strings.Join(lines, "\n") + "\n"
}

func (a *App) executeDirectFix(ctx context.Context, root, description string) error {
	fixCtx, err := a.loadDirectFixExecutionContext(root, description)
	if err != nil {
		return err
	}
	if err := a.materializeDirectFixExecutionPrompt(fixCtx); err != nil {
		return err
	}
	return a.dispatchDirectFixExecution(ctx, fixCtx)
}

func buildDirectFixPrompt(root, description string, projectCfg projectConfig, qualityCfg qualityConfig) (string, delegationPlan) {
	delegation := suggestDelegationPlan(executionModeDefault, description, "", "- [ ] Add targeted regression coverage\n- [ ] Validation commands pass")
	lines := []string{
		"# NambaAI Direct Repair Request",
		"",
		"Repair the reported issue directly in the current workspace without creating a SPEC package.",
		"",
		"## Issue",
		description,
		"",
		"## Repair Contract",
	}
	lines = append(lines, directFixRepairContractLines()...)
	lines = append(lines, "")
	lines = append(lines, directFixProjectContextPromptLines(projectCfg, qualityCfg)...)
	lines = append(lines, "")
	lines = append(lines, formatDelegationPlanPrompt(delegation)...)
	lines = append(lines, "")
	lines = append(lines, directFixValidationPromptLines(qualityCfg)...)
	lines = append(lines, "", fmt.Sprintf("Project root: %s", root))
	return strings.Join(lines, "\n"), delegation
}

func directFixRepairContractLines() []string {
	return []string{
		"- Inspect the relevant repository files plus `.namba/config/sections/*.yaml` and `.namba/project/*` context before editing.",
		"- Implement the smallest safe fix that resolves the reported issue in the current workspace.",
		"- Add targeted regression coverage for the affected area.",
		"- Run the configured validation commands from `.namba/config/sections/quality.yaml`.",
		"- Finish by running `namba sync` in the same workspace after validation passes.",
		"- Do not create or mutate `.namba/specs/<SPEC>` as part of this direct repair flow.",
		"- For bugfix SPEC scaffolding, use `namba fix --command plan \"<issue description>\"`.",
	}
}

func directFixProjectContextPromptLines(projectCfg projectConfig, qualityCfg qualityConfig) []string {
	return []string{
		"## Project Context",
		fmt.Sprintf("- Project: %s", firstNonBlank(projectCfg.Name, "unknown")),
		fmt.Sprintf("- Project type: %s", firstNonBlank(projectCfg.ProjectType, "unknown")),
		fmt.Sprintf("- Language: %s", firstNonBlank(projectCfg.Language, "unknown")),
		fmt.Sprintf("- Framework: %s", firstNonBlank(projectCfg.Framework, "unknown")),
		fmt.Sprintf("- Development mode: %s", firstNonBlank(qualityCfg.DevelopmentMode, "unknown")),
	}
}

func directFixValidationPromptLines(qualityCfg qualityConfig) []string {
	lines := []string{"## Validation"}
	for _, step := range validationPipelineSteps(qualityCfg) {
		command := strings.TrimSpace(step.Command)
		if command == "" || command == "none" {
			continue
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", step.Name, command))
	}
	return lines
}

func buildSpecDoc(kind, specID, description string, projectCfg projectConfig, qualityCfg qualityConfig) string {
	switch kind {
	case "fix":
		return buildFixSpecDoc(specID, description, projectCfg, qualityCfg)
	case "harness":
		return buildHarnessSpecDoc(specID, description, projectCfg, qualityCfg)
	default:
		return buildFeatureSpecDoc(specID, description, projectCfg, qualityCfg)
	}
}

func buildSpecPlanDoc(kind, specID string) string {
	switch kind {
	case "fix":
		return buildFixSpecPlanDoc(specID)
	case "harness":
		return buildHarnessSpecPlanDoc(specID)
	default:
		return buildFeatureSpecPlanDoc(specID)
	}
}

func buildFixSpecDoc(specID, description string, projectCfg projectConfig, qualityCfg qualityConfig) string {
	return fmt.Sprintf("# %s\n\n## Problem\n\n%s\n\n## Goal\n\nApply the smallest safe fix that resolves the reported issue.\n\n## Context\n\n- Project: %s\n- Project type: %s\n- Language: %s\n- Mode: %s\n- Work type: fix\n", specID, description, projectCfg.Name, projectCfg.ProjectType, projectCfg.Language, qualityCfg.DevelopmentMode)
}

func buildHarnessSpecDoc(specID, description string, projectCfg projectConfig, qualityCfg qualityConfig) string {
	return fmt.Sprintf("# %s\n\n## Problem\n\nThe current repository needs a dedicated harness-oriented planning flow for the following request:\n\n%s\n\n## Goal\n\nDesign a Codex-native harness change under the existing `SPEC-XXX` artifact flow without inventing a second planning model or importing Claude-only runtime primitives.\n\n## Context\n\n- Project: %s\n- Project type: %s\n- Language: %s\n- Mode: %s\n- Work type: plan\n- Planning surface: `namba harness \"<description>\"`\n\n## Desired Outcome\n\n- `namba harness \"<description>\"` acts as a top-level planning command while `namba plan` keeps its current feature-planning behavior.\n- The scaffold captures Codex-native execution topology, agent/skill boundaries, progressive-disclosure guidance, trigger strategy, and evaluation strategy for reusable skills or agents.\n- Help and accidental-write safety stay aligned with the shared command-parsing contract instead of creating command-specific drift.\n- The planned output remains under `.namba/specs/<SPEC>` with the normal review artifacts.\n\n## Non-Goals\n\n- Do not create a second artifact model outside `.namba/specs/`.\n- Do not emit `.claude/*`, `TeamCreate`, `SendMessage`, `TaskCreate`, or a mandatory `model: \"opus\"` requirement as part of the Codex-facing contract.\n- Do not change the default behavior of `namba plan`.\n", specID, description, projectCfg.Name, projectCfg.ProjectType, projectCfg.Language, qualityCfg.DevelopmentMode)
}

func buildFeatureSpecDoc(specID, description string, projectCfg projectConfig, qualityCfg qualityConfig) string {
	return fmt.Sprintf("# %s\n\n## Problem\n\n%s\n\n## Goal\n\nImplement the requested change under the normal feature-planning workflow.\n\n## Context\n\n- Project: %s\n- Project type: %s\n- Language: %s\n- Mode: %s\n- Work type: plan\n", specID, description, projectCfg.Name, projectCfg.ProjectType, projectCfg.Language, qualityCfg.DevelopmentMode)
}

func buildFixSpecPlanDoc(specID string) string {
	return fmt.Sprintf("# %s Plan\n\n1. Refresh project context with `namba project`\n2. Reproduce or inspect the reported issue\n3. Run the relevant review passes under `.namba/specs/%s/reviews/` and refresh the readiness summary\n4. Implement the smallest safe fix\n5. Run validation commands and targeted regression checks\n6. Sync artifacts with `namba sync`\n", specID, specID)
}

func buildHarnessSpecPlanDoc(specID string) string {
	return fmt.Sprintf("# %s Plan\n\n1. Refresh project context with `namba project`\n2. Define the top-level `namba harness` command contract while keeping `namba plan` unchanged\n3. Capture Codex-native execution topology, agent/skill boundaries, progressive-disclosure layout, trigger guidance, and evaluation strategy in the scaffold\n4. Run the relevant review passes under `.namba/specs/%s/reviews/` and refresh the readiness summary\n5. Implement the requested command and scaffold changes\n6. Run validation commands\n7. Sync artifacts with `namba sync`\n", specID, specID)
}

func buildFeatureSpecPlanDoc(specID string) string {
	return fmt.Sprintf("# %s Plan\n\n1. Refresh project context with `namba project`\n2. Run the relevant review passes under `.namba/specs/%s/reviews/` and refresh the readiness summary\n3. Implement the requested change\n4. Run validation commands\n5. Sync artifacts with `namba sync`\n", specID, specID)
}

func buildSpecAcceptanceDoc(kind, description, mode string) string {
	if kind == "fix" {
		return buildFixAcceptanceDoc(description, mode)
	}
	if kind == "harness" {
		return buildHarnessAcceptanceDoc(description, mode)
	}
	return buildFeatureAcceptanceDoc(description, mode)
}

type runExecuteOptions struct {
	specID string
	mode   executionMode
	dryRun bool
}

type specPackageScaffoldContext struct {
	Root        string
	Kind        string
	Description string
	SpecID      string
	ProjectCfg  projectConfig
	QualityCfg  qualityConfig
}

type executionRuntimeConfig struct {
	QualityCfg qualityConfig
	SystemCfg  systemConfig
	CodexCfg   codexConfig
}

type directFixExecutionContext struct {
	Root        string
	Description string
	QualityCfg  qualityConfig
	SystemCfg   systemConfig
	CodexCfg    codexConfig
	Prompt      string
	PromptPath  string
	LogID       string
	Delegation  delegationPlan
}

type runExecutionContext struct {
	Root              string
	SpecPkg           specPackage
	ReadinessAdvisory string
	QualityCfg        qualityConfig
	SystemCfg         systemConfig
	CodexCfg          codexConfig
	WorkflowCfg       workflowConfig
	Prompt            string
	PromptPath        string
	Tasks             []string
	Delegation        delegationPlan
}

type syncContext struct {
	Root       string
	ProjectCfg projectConfig
	LatestSpec string
	Profile    initProfile
	DocsCfg    docsConfig
	Support    syncSupportContext
}

func parseRunExecuteOptions(args []string) (runExecuteOptions, error) {
	if len(args) == 0 {
		return runExecuteOptions{}, errors.New("run requires a SPEC id")
	}

	options := runExecuteOptions{specID: args[0], mode: executionModeDefault}
	var solo bool
	var team bool
	var parallel bool
	for _, arg := range args[1:] {
		switch arg {
		case "--solo":
			solo = true
		case "--team":
			team = true
		case "--parallel":
			parallel = true
		case "--dry-run":
			options.dryRun = true
		default:
			return runExecuteOptions{}, fmt.Errorf("unknown flag %q", arg)
		}
	}

	selectedModes := make([]string, 0, 3)
	if solo {
		selectedModes = append(selectedModes, "--solo")
	}
	if team {
		selectedModes = append(selectedModes, "--team")
	}
	if parallel {
		selectedModes = append(selectedModes, "--parallel")
	}
	if len(selectedModes) > 1 {
		return runExecuteOptions{}, fmt.Errorf("invalid flag combination: choose only one of --solo, --team, or --parallel (got %s)", strings.Join(selectedModes, ", "))
	}

	switch {
	case solo:
		options.mode = executionModeSolo
	case team:
		options.mode = executionModeTeam
	case parallel:
		options.mode = executionModeParallel
	}

	return options, nil
}

func (a *App) loadSpecPackageScaffoldContext(kind, description string) (specPackageScaffoldContext, error) {
	root, err := a.requireProjectRoot()
	if err != nil {
		return specPackageScaffoldContext{}, err
	}
	specID, err := nextSpecID(filepath.Join(root, specsDir))
	if err != nil {
		return specPackageScaffoldContext{}, err
	}
	return a.loadResolvedSpecPackageScaffoldContext(root, kind, description, specID)
}

func (a *App) loadResolvedSpecPackageScaffoldContext(root, kind, description, specID string) (specPackageScaffoldContext, error) {
	projectCfg, _ := a.loadProjectConfig(root)
	qualityCfg, _ := a.loadQualityConfig(root)
	return specPackageScaffoldContext{
		Root:        root,
		Kind:        kind,
		Description: description,
		SpecID:      specID,
		ProjectCfg:  projectCfg,
		QualityCfg:  qualityCfg,
	}, nil
}

func buildSpecPackageScaffoldOutputs(scaffoldCtx specPackageScaffoldContext) map[string]string {
	outputs := map[string]string{
		filepath.ToSlash(filepath.Join(specsDir, scaffoldCtx.SpecID, "spec.md")):       buildSpecDoc(scaffoldCtx.Kind, scaffoldCtx.SpecID, scaffoldCtx.Description, scaffoldCtx.ProjectCfg, scaffoldCtx.QualityCfg),
		filepath.ToSlash(filepath.Join(specsDir, scaffoldCtx.SpecID, "plan.md")):       buildSpecPlanDoc(scaffoldCtx.Kind, scaffoldCtx.SpecID),
		filepath.ToSlash(filepath.Join(specsDir, scaffoldCtx.SpecID, "acceptance.md")): buildSpecAcceptanceDoc(scaffoldCtx.Kind, scaffoldCtx.Description, scaffoldCtx.QualityCfg.DevelopmentMode),
	}
	frontendBriefBody := ""
	if body, ok := buildFrontendBriefDoc(scaffoldCtx.Kind, scaffoldCtx.Description); ok {
		outputs[filepath.ToSlash(filepath.Join(specsDir, scaffoldCtx.SpecID, frontendBriefFileName))] = body
		frontendBriefBody = body
	}
	if req := inferredPlanningHarnessRequest(scaffoldCtx.Kind, scaffoldCtx.Description); req != nil {
		if body, err := marshalHarnessRequest(*req); err == nil {
			outputs[specHarnessRequestPath(scaffoldCtx.SpecID)] = body
		}
	}
	for rel, body := range specReviewOutputsWithFrontendBrief(scaffoldCtx.SpecID, frontendBriefBody) {
		outputs[rel] = body
	}
	return outputs
}

func (a *App) materializeSpecPackageScaffoldOutputs(scaffoldCtx specPackageScaffoldContext, outputs map[string]string) error {
	if err := a.mkdirAll(filepath.Join(scaffoldCtx.Root, specsDir, scaffoldCtx.SpecID), 0o755); err != nil {
		return err
	}
	if _, err := a.writeOutputs(scaffoldCtx.Root, outputs); err != nil {
		return err
	}
	return nil
}

func (a *App) loadDirectFixExecutionContext(root, description string) (directFixExecutionContext, error) {
	projectCfg, err := a.loadProjectConfig(root)
	if err != nil {
		return directFixExecutionContext{}, err
	}
	runtimeCfg, err := a.loadExecutionRuntimeConfig(root)
	if err != nil {
		return directFixExecutionContext{}, err
	}

	prompt, delegation := buildDirectFixPrompt(root, description, projectCfg, runtimeCfg.QualityCfg)
	logID := "direct-fix"
	return directFixExecutionContext{
		Root:        root,
		Description: description,
		QualityCfg:  runtimeCfg.QualityCfg,
		SystemCfg:   runtimeCfg.SystemCfg,
		CodexCfg:    runtimeCfg.CodexCfg,
		Prompt:      prompt,
		PromptPath:  filepath.Join(root, logsDir, "runs", logID+"-request.md"),
		LogID:       logID,
		Delegation:  delegation,
	}, nil
}

func (a *App) materializeDirectFixExecutionPrompt(fixCtx directFixExecutionContext) error {
	return a.writeExecutionPrompt(fixCtx.PromptPath, fixCtx.Prompt)
}

func (a *App) dispatchDirectFixExecution(ctx context.Context, fixCtx directFixExecutionContext) error {
	request := a.newExecutionRequest("DIRECT-FIX", fixCtx.Root, fixCtx.Prompt, executionModeDefault, fixCtx.Delegation, fixCtx.SystemCfg, fixCtx.CodexCfg)
	request.TurnName = fixCtx.LogID
	request.TurnRole = fixCtx.Delegation.IntegratorRole
	if _, _, err := a.executeRun(ctx, fixCtx.Root, fixCtx.LogID, request, fixCtx.Root, fixCtx.QualityCfg, nil, ""); err != nil {
		return err
	}
	if err := a.runSync(ctx, nil); err != nil {
		return err
	}

	fmt.Fprintf(a.stdout, "Executed direct fix with %s\n", request.Runner)
	return nil
}

func (a *App) runExecute(ctx context.Context, args []string) error {
	if wantsCommandHelp(args) {
		return a.printCommandUsage("run")
	}
	options, err := parseRunExecuteOptions(args)
	if err != nil {
		return commandUsageError("run", err)
	}

	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}

	runCtx, err := a.loadRunExecutionContext(root, options)
	if err != nil {
		return err
	}
	if err := a.materializeRunExecutionPrompt(runCtx); err != nil {
		return err
	}
	return a.dispatchRunExecution(ctx, options, runCtx)
}

func (a *App) loadRunExecutionContext(root string, options runExecuteOptions) (runExecutionContext, error) {
	specPkg, err := a.loadSpec(root, options.specID)
	if err != nil {
		return runExecutionContext{}, err
	}
	readinessAdvisory, err := a.refreshSpecReviewReadiness(root, specPkg.ID)
	if err != nil {
		return runExecutionContext{}, err
	}
	if frontend := loadFrontendBriefReport(root, specPkg.ID); frontend.Exists {
		if !frontend.Valid {
			return runExecutionContext{}, frontendGateExecutionError(specPkg.ID, frontend)
		}
		if frontend.Header.TaskClassification == frontendTaskClassificationMajor {
			frontendReady := frontend.Header.FrontendGateStatus == frontendGateStatusApproved && frontend.EvidenceStatus == frontendEvidenceStatusComplete && len(frontend.Mismatches) == 0
			if !frontendReady {
				return runExecutionContext{}, frontendGateExecutionError(specPkg.ID, frontend)
			}
		}
	}
	runtimeCfg, err := a.loadExecutionRuntimeConfig(root)
	if err != nil {
		return runExecutionContext{}, err
	}
	workflowCfg, err := a.loadWorkflowConfig(root)
	if err != nil {
		return runExecutionContext{}, err
	}
	prompt, tasks, delegation, err := a.buildExecutionPrompt(root, specPkg, runtimeCfg.QualityCfg, options.mode)
	if err != nil {
		return runExecutionContext{}, err
	}

	return runExecutionContext{
		Root:              root,
		SpecPkg:           specPkg,
		ReadinessAdvisory: readinessAdvisory,
		QualityCfg:        runtimeCfg.QualityCfg,
		SystemCfg:         runtimeCfg.SystemCfg,
		CodexCfg:          runtimeCfg.CodexCfg,
		WorkflowCfg:       workflowCfg,
		Prompt:            prompt,
		PromptPath:        filepath.Join(root, logsDir, "runs", strings.ToLower(specPkg.ID)+"-request.md"),
		Tasks:             tasks,
		Delegation:        delegation,
	}, nil
}

func (a *App) materializeRunExecutionPrompt(runCtx runExecutionContext) error {
	return a.writeExecutionPrompt(runCtx.PromptPath, runCtx.Prompt)
}

func (a *App) loadExecutionRuntimeConfig(root string) (executionRuntimeConfig, error) {
	qualityCfg, err := a.loadQualityConfig(root)
	if err != nil {
		return executionRuntimeConfig{}, err
	}
	systemCfg, err := a.loadSystemConfig(root)
	if err != nil {
		return executionRuntimeConfig{}, err
	}
	codexCfg, err := a.loadCodexConfig(root)
	if err != nil {
		return executionRuntimeConfig{}, err
	}
	return executionRuntimeConfig{
		QualityCfg: qualityCfg,
		SystemCfg:  systemCfg,
		CodexCfg:   codexCfg,
	}, nil
}

func (a *App) writeExecutionPrompt(path, prompt string) error {
	if err := a.mkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return a.writeFile(path, []byte(prompt), 0o644)
}

func (a *App) dispatchRunExecution(ctx context.Context, options runExecuteOptions, runCtx runExecutionContext) error {
	if runCtx.ReadinessAdvisory != "" {
		fmt.Fprintf(a.stdout, "Review readiness for %s: %s (advisory only)\n", runCtx.SpecPkg.ID, runCtx.ReadinessAdvisory)
	}

	if options.mode == executionModeParallel {
		return a.runParallel(ctx, runCtx.Root, runCtx.SpecPkg, runCtx.Tasks, runCtx.Prompt, runCtx.QualityCfg, runCtx.SystemCfg, runCtx.CodexCfg, runCtx.WorkflowCfg, options.dryRun)
	}

	if options.dryRun {
		fmt.Fprintf(a.stdout, "Prepared execution request at %s\n", runCtx.PromptPath)
		return nil
	}

	request := a.newExecutionRequest(runCtx.SpecPkg.ID, runCtx.Root, runCtx.Prompt, options.mode, runCtx.Delegation, runCtx.SystemCfg, runCtx.CodexCfg)
	if _, _, err := a.executeRun(ctx, runCtx.Root, strings.ToLower(runCtx.SpecPkg.ID), request, runCtx.Root, runCtx.QualityCfg, nil, ""); err != nil {
		return err
	}

	fmt.Fprintf(a.stdout, "Executed %s with %s\n", runCtx.SpecPkg.ID, request.Runner)
	return nil
}

func (a *App) runSync(_ context.Context, args []string) error {
	if handled, err := a.handleNoArgTopLevelCommand("sync", args); handled {
		return err
	}

	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}

	syncCtx, err := a.loadSyncContext(root)
	if err != nil {
		return err
	}
	qualityCfg, _ := a.loadQualityConfig(root)
	analysisCfg, err := a.loadAnalysisConfig(root)
	if err != nil {
		return err
	}

	readmeOutputs := buildReadmeOutputs(syncCtx.ProjectCfg, syncCtx.Profile, syncCtx.DocsCfg)
	analysis := analyzeProject(root, syncCtx.ProjectCfg, qualityCfg, analysisCfg)
	projectOutputs := analysis.renderOutputs()
	readinessBatch, err := buildSpecReviewReadinessBatch(root)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	syncCtx.Support = a.buildSyncSupportContext(root, syncCtx.LatestSpec, readinessBatch.Advisories)

	session, err := a.beginManagedOutputSession(root)
	if err != nil {
		return err
	}
	if err := session.replaceManagedOutputs(readmeOutputs, isReadmeManagedPath, nil); err != nil {
		return err
	}
	if err := session.replaceManagedOutputs(projectOutputs, isProjectAnalysisManagedPath, nil); err != nil {
		return err
	}
	for _, warning := range analysis.Quality.Warnings {
		fmt.Fprintf(a.stdout, "Project analysis warning: %s\n", warning)
	}
	if len(analysis.Quality.Errors) > 0 {
		if _, err := session.commit(); err != nil {
			return err
		}
		for _, item := range analysis.Quality.Errors {
			fmt.Fprintf(a.stdout, "Project analysis error: %s\n", item)
		}
		return errors.New("project analysis quality gate failed")
	}

	if err := session.replaceManagedOutputs(readinessBatch.Outputs, isSpecReviewReadinessManagedPath, nil); err != nil {
		return err
	}
	if err := session.replaceManagedOutputs(buildSyncProjectSupportOutputs(syncCtx), isSyncProjectSupportManagedPath, nil); err != nil {
		return err
	}
	if _, err := session.commit(); err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "Synced NambaAI artifacts.")
	return nil
}

func (a *App) loadSyncContext(root string) (syncContext, error) {
	projectCfg, _ := a.loadProjectConfig(root)
	latestSpec, _ := latestSpecID(filepath.Join(root, specsDir))
	profile, err := a.loadInitProfileFromConfig(root)
	if err != nil {
		return syncContext{}, err
	}
	docsCfg, err := a.loadDocsConfig(root)
	if err != nil {
		return syncContext{}, err
	}
	support := a.buildSyncSupportContext(root, latestSpec, nil)
	return syncContext{
		Root:       root,
		ProjectCfg: projectCfg,
		LatestSpec: latestSpec,
		Profile:    profile,
		DocsCfg:    docsCfg,
		Support:    support,
	}, nil
}

func (a *App) materializeSyncReadme(syncCtx syncContext) error {
	if _, err := a.replaceManagedOutputs(syncCtx.Root, buildReadmeOutputs(syncCtx.ProjectCfg, syncCtx.Profile, syncCtx.DocsCfg), isReadmeManagedPath, nil); err != nil {
		return err
	}
	return nil
}

func (a *App) refreshSyncProjectArtifacts(ctx context.Context, syncCtx syncContext) error {
	if err := a.runProject(ctx, nil); err != nil {
		return err
	}
	if err := a.refreshAllSpecReviewReadiness(syncCtx.Root); err != nil {
		return err
	}
	return nil
}

func (a *App) writeSyncProjectSupportDocs(syncCtx syncContext) error {
	outputs := buildSyncProjectSupportOutputs(syncCtx)
	if err := a.materializeSyncProjectSupportOutputs(syncCtx.Root, outputs); err != nil {
		return err
	}
	return nil
}

func buildSyncProjectSupportOutputs(syncCtx syncContext) map[string]string {
	return map[string]string{
		filepath.ToSlash(filepath.Join(projectDir, "change-summary.md")):    buildChangeSummaryDocWithSupport(syncCtx.ProjectCfg, syncCtx.Profile, syncCtx.Support),
		filepath.ToSlash(filepath.Join(projectDir, "pr-checklist.md")):      buildPRChecklistDocWithSupport(syncCtx.Profile, syncCtx.Support),
		filepath.ToSlash(filepath.Join(projectDir, "release-notes.md")):     buildReleaseNotesDoc(syncCtx.ProjectCfg, syncCtx.LatestSpec, syncCtx.Profile),
		filepath.ToSlash(filepath.Join(projectDir, "release-checklist.md")): buildReleaseChecklistDoc(),
	}
}

func (a *App) materializeSyncProjectSupportOutputs(root string, outputs map[string]string) error {
	if _, err := a.writeOutputs(root, outputs); err != nil {
		return err
	}
	return nil
}

func (a *App) buildExecutionPrompt(root string, specPkg specPackage, qualityCfg qualityConfig, mode executionMode) (string, []string, delegationPlan, error) {
	specBytes, err := os.ReadFile(filepath.Join(specPkg.Path, "spec.md"))
	if err != nil {
		return "", nil, delegationPlan{}, err
	}
	planBytes, err := os.ReadFile(filepath.Join(specPkg.Path, "plan.md"))
	if err != nil {
		return "", nil, delegationPlan{}, err
	}
	acceptanceBytes, err := os.ReadFile(filepath.Join(specPkg.Path, "acceptance.md"))
	if err != nil {
		return "", nil, delegationPlan{}, err
	}

	tasks := extractAcceptanceTasks(string(acceptanceBytes))
	mode = normalizeExecutionMode(mode)
	modeGuidance := executionModePromptGuidance(mode)
	delegation := suggestDelegationPlan(mode, string(specBytes), string(planBytes), string(acceptanceBytes))
	promptLines := []string{
		"# NambaAI Execution Request",
		"",
		"Execute this SPEC package using the repository AGENTS.md and local Codex skills.",
		"",
		"## Run Mode",
		fmt.Sprintf("- Mode: %s", mode),
	}
	promptLines = append(promptLines, modeGuidance...)
	promptLines = append(promptLines, "")
	promptLines = append(promptLines, formatDelegationPlanPrompt(delegation)...)
	promptLines = append(promptLines,
		"",
		"## SPEC",
		string(specBytes),
		"",
		"## Plan",
		string(planBytes),
		"",
		"## Acceptance",
		string(acceptanceBytes),
	)
	if frontendBriefExists := exists(filepath.Join(specPkg.Path, frontendBriefFileName)); frontendBriefExists {
		frontendBytes, err := os.ReadFile(filepath.Join(specPkg.Path, frontendBriefFileName))
		if err != nil {
			return "", nil, delegationPlan{}, err
		}
		promptLines = append(promptLines,
			"",
			"## Frontend Brief",
			string(frontendBytes),
		)
	}
	if specReviewReadinessExists(root, specPkg.ID) {
		readinessBytes, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(specReviewReadinessPath(specPkg.ID))))
		if err != nil {
			return "", nil, delegationPlan{}, err
		}
		promptLines = append(promptLines,
			"",
			"## Review Readiness",
			string(readinessBytes),
		)
	}
	promptLines = append(promptLines, "", "## Validation")
	for _, step := range validationPipelineSteps(qualityCfg) {
		promptLines = append(promptLines, fmt.Sprintf("- %s: %s", step.Name, step.Command))
	}
	promptLines = append(promptLines, "", fmt.Sprintf("Project root: %s", root))
	prompt := strings.Join(promptLines, "\n")

	return prompt, tasks, delegation, nil
}

func executionModePromptGuidance(mode executionMode) []string {
	switch normalizeExecutionMode(mode) {
	case executionModeSolo:
		return []string{
			"- Execution style: one runner in one workspace.",
			"- Keep implementation, integration, and validation inside one runner rather than same-workspace team orchestration.",
			"- Do not reinterpret this mode as worktree parallelism.",
		}
	case executionModeTeam:
		return []string{
			"- Execution style: same-workspace multi-agent execution.",
			"- Keep work in one workspace while orchestrating specialist turns and a final reviewer inside one bounded runtime.",
			"- Role runtime profiles should materially affect the actual Codex turns, not only prompt wording.",
			"- Do not reinterpret this mode as worktree parallelism.",
		}
	case executionModeParallel:
		return []string{
			"- Execution style: Namba worktree parallel mode.",
			"- This mode means git worktree fan-out/fan-in managed by Namba, not same-workspace team orchestration.",
			"- Each worker request should stay within its assigned work package and merge only after all workers and validators pass.",
		}
	default:
		return []string{
			"- Execution style: standard standalone Codex run in one workspace.",
			"- Keep work inside the standalone runner unless the user explicitly picks `--team` or `--parallel`.",
			"- Do not reinterpret this mode as worktree parallelism.",
		}
	}
}

type delegationDomainConfig struct {
	Name             string
	PrimaryRole      string
	PlanningRole     string
	Keywords         []string
	PlanningKeywords []string
	ScoreBias        int
}

type delegationDomainMatch struct {
	Config        delegationDomainConfig
	Role          string
	Hits          []string
	Score         int
	WeightedScore int
}

func suggestDelegationPlan(mode executionMode, specText, planText, acceptanceText string) delegationPlan {
	combined := strings.ToLower(strings.Join([]string{specText, planText, acceptanceText}, "\n"))
	matches := make([]delegationDomainMatch, 0)
	for _, cfg := range delegationDomainConfigs() {
		hits := findKeywordHits(combined, cfg.Keywords)
		if len(hits) == 0 {
			continue
		}
		role := cfg.PrimaryRole
		if cfg.PlanningRole != "" {
			planningHits := findKeywordHits(combined, cfg.PlanningKeywords)
			if len(planningHits) > 0 {
				role = cfg.PlanningRole
				hits = uniqueStrings(append(hits, planningHits...))
			}
		}
		matches = append(matches, delegationDomainMatch{
			Config:        cfg,
			Role:          role,
			Hits:          hits,
			Score:         len(hits),
			WeightedScore: len(hits) + cfg.ScoreBias,
		})
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].WeightedScore == matches[j].WeightedScore {
			if matches[i].Score == matches[j].Score {
				return matches[i].Config.Name < matches[j].Config.Name
			}
			return matches[i].Score > matches[j].Score
		}
		return matches[i].WeightedScore > matches[j].WeightedScore
	})

	integratorRole := "standalone-runner"
	switch normalizeExecutionMode(mode) {
	case executionModeTeam:
		integratorRole = "same-workspace-integrator"
	case executionModeParallel:
		integratorRole = "parallel-orchestrator"
	}
	plan := delegationPlan{IntegratorRole: integratorRole}
	for _, match := range matches {
		plan.DominantDomains = append(plan.DominantDomains, match.Config.Name)
	}
	plan.SelectedRoles, plan.DelegationBudget, plan.ReviewerRole, plan.RoutingRationale = chooseDelegatedRoles(mode, matches)
	plan.SelectedRoleProfiles = runtimeProfilesForRoles(plan.SelectedRoles)
	return plan
}

func delegationDomainConfigs() []delegationDomainConfig {
	return []delegationDomainConfig{
		{
			Name:             "frontend",
			PrimaryRole:      "namba-frontend-implementer",
			PlanningRole:     "namba-frontend-architect",
			Keywords:         []string{"frontend", "ui", "component", "screen", "page", "responsive", "browser", "css", "accessibility", "a11y"},
			PlanningKeywords: []string{"component/state split", "plan the component", "plan the state", "component boundary", "component boundaries", "state ownership", "file planning", "delivery planning"},
		},
		{Name: "mobile", PrimaryRole: "namba-mobile-engineer", Keywords: []string{"mobile", "ios", "android", "swift", "kotlin", "react native", "flutter", "tablet", "touch"}, ScoreBias: 2},
		{Name: "backend", PrimaryRole: "namba-backend-implementer", Keywords: []string{"backend", "api", "endpoint", "server", "service", "controller", "handler", "webhook"}, ScoreBias: 1},
		{Name: "data", PrimaryRole: "namba-data-engineer", Keywords: []string{"schema", "migration", "sql", "query", "etl", "warehouse", "analytics", "dataset", "batch", "pipeline"}, ScoreBias: 2},
		{Name: "security", PrimaryRole: "namba-security-engineer", Keywords: []string{"security", "auth", "oauth", "permission", "secret", "token", "encryption", "vulnerability", "compliance", "privacy", "pii"}, ScoreBias: 2},
		{Name: "design", PrimaryRole: "namba-designer", Keywords: []string{"design", "figma", "art direction", "visual direction", "visual design", "palette", "tone logic", "composition", "look and feel", "redesign", "typography", "motion", "prototype", "brand"}, ScoreBias: 1},
		{Name: "devops", PrimaryRole: "namba-devops-engineer", Keywords: []string{"deploy", "deployment", "docker", "kubernetes", "helm", "terraform", "ci", "cd", "infra", "observability", "runtime", "environment"}, ScoreBias: 2},
		{Name: "quality", PrimaryRole: "namba-test-engineer", Keywords: []string{"test", "regression", "coverage", "qa", "e2e", "integration test", "acceptance test"}, ScoreBias: -1},
	}
}

func chooseDelegatedRoles(mode executionMode, matches []delegationDomainMatch) ([]string, int, string, []string) {
	switch normalizeExecutionMode(mode) {
	case executionModeParallel:
		return nil, 0, "", []string{
			"`--parallel` is reserved for Namba worktree fan-out, so do not route to Codex subagents in this mode.",
		}
	case executionModeSolo:
		if len(matches) == 0 || matches[0].Score < 2 {
			return nil, 0, "", []string{
				"No single specialist signal is strong enough, so stay inside one generalist runner.",
			}
		}
		return []string{matches[0].Role}, 1, "", []string{
			fmt.Sprintf("Highest-signal domain is %s via %s.", matches[0].Config.Name, quoteList(matches[0].Hits)),
			"Delegate to one bounded specialist only if it materially reduces risk, and keep integration plus validation in the standalone runner.",
		}
	case executionModeTeam:
		if len(matches) == 0 {
			return []string{"namba-implementer", "namba-reviewer"}, 2, "namba-reviewer", []string{
				"No domain clearly dominates, so keep team mode light with one general implementer plus a reviewer.",
				"Add more specialists only when acceptance criteria span multiple clearly independent domains.",
			}
		}
		maxDomains := 1
		if len(matches) > 1 {
			maxDomains = 2
		}
		roles := make([]string, 0, maxDomains+1)
		rationale := make([]string, 0, maxDomains+2)
		for i := 0; i < len(matches) && i < maxDomains; i++ {
			roles = append(roles, matches[i].Role)
			rationale = append(rationale, fmt.Sprintf("%s matched %s.", matches[i].Config.Name, quoteList(matches[i].Hits)))
		}
		if len(matches) > 2 {
			rationale = append(rationale, "More than two domains matched, but team mode stays light by using only the top two specialists before review.")
		} else if len(matches) == 1 {
			rationale = append(rationale, "One domain dominates, so start with one specialist and a reviewer rather than a larger swarm.")
		} else {
			rationale = append(rationale, "Multiple domains matched, so use one specialist per dominant domain before the final review pass.")
		}
		rationale = append(rationale, "Keep the standalone runner as the integrator and final validation owner.")
		roles = append(roles, "namba-reviewer")
		roles = uniqueStrings(roles)
		return roles, len(roles), "namba-reviewer", rationale
	default:
		return nil, 0, "", []string{
			"Default mode keeps work inside the standalone runner unless the user explicitly asks for specialist delegation.",
		}
	}
}

func formatDelegationPlanPrompt(plan delegationPlan) []string {
	lines := []string{"## Delegation Heuristics"}
	if len(plan.DominantDomains) == 0 {
		lines = append(lines, "- Dominant domains: none detected beyond general implementation.")
	} else {
		lines = append(lines, fmt.Sprintf("- Dominant domains: %s.", strings.Join(plan.DominantDomains, ", ")))
	}
	if len(plan.SelectedRoles) == 0 {
		lines = append(lines, "- Suggested roles: keep work inside the standalone runner without spawning specialists.")
	} else {
		lines = append(lines, fmt.Sprintf("- Suggested roles: %s.", quoteList(plan.SelectedRoles)))
	}
	for _, profile := range plan.SelectedRoleProfiles {
		if summary := formatAgentRuntimeProfile(profile); summary != "" {
			lines = append(lines, "- Role runtime: "+summary+".")
		}
	}
	lines = append(lines, fmt.Sprintf("- Delegation budget: %d.", plan.DelegationBudget))
	if plan.IntegratorRole != "" {
		lines = append(lines, fmt.Sprintf("- Integrator: `%s`.", plan.IntegratorRole))
	}
	if plan.ReviewerRole != "" {
		lines = append(lines, fmt.Sprintf("- Reviewer: `%s`.", plan.ReviewerRole))
	}
	for _, reason := range plan.RoutingRationale {
		lines = append(lines, "- "+reason)
	}
	return lines
}

func findKeywordHits(text string, keywords []string) []string {
	hits := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			hits = append(hits, keyword)
		}
	}
	return uniqueStrings(hits)
}

func quoteList(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	quoted := make([]string, 0, len(values))
	for _, value := range uniqueStrings(values) {
		quoted = append(quoted, fmt.Sprintf("`%s`", value))
	}
	return strings.Join(quoted, ", ")
}

func (a *App) runWorktree(ctx context.Context, args []string) error {
	if wantsCommandHelp(args) {
		return a.printCommandUsage("worktree")
	}
	if len(args) == 0 {
		return commandUsageError("worktree", errors.New("worktree requires a subcommand"))
	}

	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}
	subcommand, ok := a.resolveWorktreeSubcommand(args[0])
	if !ok {
		return commandUsageError("worktree", fmt.Errorf("unknown worktree subcommand %q", args[0]))
	}
	return subcommand.Run(a, ctx, root, args[1:])
}

func worktreeSubcommandDefinitions() []worktreeSubcommandDefinition {
	return []worktreeSubcommandDefinition{
		{Name: "new", UsageSummary: "  namba worktree new <name>", Run: (*App).runWorktreeNewSubcommand},
		{Name: "list", UsageSummary: "  namba worktree list", Run: (*App).runWorktreeListSubcommand},
		{Name: "remove", UsageSummary: "  namba worktree remove <name>", Run: (*App).runWorktreeRemoveSubcommand},
		{Name: "clean", UsageSummary: "  namba worktree clean", Run: (*App).runWorktreeCleanSubcommand},
	}
}

func worktreeSubcommandUsageSummaries() []string {
	lines := make([]string, 0, len(worktreeSubcommandDefinitions()))
	for _, definition := range worktreeSubcommandDefinitions() {
		lines = append(lines, definition.UsageSummary)
	}
	return lines
}

func (a *App) resolveWorktreeSubcommand(name string) (worktreeSubcommandDefinition, bool) {
	for _, definition := range worktreeSubcommandDefinitions() {
		if definition.Name == name {
			return definition, true
		}
	}
	return worktreeSubcommandDefinition{}, false
}

func (a *App) runWorktreeNewSubcommand(ctx context.Context, root string, args []string) error {
	if len(args) != 1 {
		if len(args) == 0 {
			return commandUsageError("worktree", errors.New("worktree new requires a name"))
		}
		return commandUsageError("worktree", errors.New("worktree new accepts exactly one name"))
	}
	name := args[0]
	path := filepath.Join(root, worktreesDir, name)
	_, err := a.runBinary(ctx, "git", []string{"worktree", "add", "-b", "namba/" + name, path, "HEAD"}, root)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.stdout, "Created worktree %s\n", path)
	return nil
}

func (a *App) runWorktreeListSubcommand(ctx context.Context, root string, args []string) error {
	if len(args) != 0 {
		return commandUsageError("worktree", errors.New("worktree list does not accept arguments"))
	}
	out, err := a.runBinary(ctx, "git", []string{"worktree", "list", "--porcelain"}, root)
	if err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, out)
	return nil
}

func (a *App) runWorktreeRemoveSubcommand(ctx context.Context, root string, args []string) error {
	if len(args) != 1 {
		if len(args) == 0 {
			return commandUsageError("worktree", errors.New("worktree remove requires a name"))
		}
		return commandUsageError("worktree", errors.New("worktree remove accepts exactly one name"))
	}
	path := filepath.Join(root, worktreesDir, args[0])
	_, err := a.runBinary(ctx, "git", []string{"worktree", "remove", "--force", path}, root)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.stdout, "Removed worktree %s\n", path)
	return nil
}

func (a *App) runWorktreeCleanSubcommand(ctx context.Context, root string, args []string) error {
	if len(args) != 0 {
		return commandUsageError("worktree", errors.New("worktree clean does not accept arguments"))
	}
	_, err := a.runBinary(ctx, "git", []string{"worktree", "prune"}, root)
	if err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "Pruned worktrees.")
	return nil
}

type parallelWorkerState struct {
	name   string
	path   string
	branch string
	err    error
}

func (a *App) runParallel(ctx context.Context, root string, specPkg specPackage, tasks []string, prompt string, qualityCfg qualityConfig, systemCfg systemConfig, codexCfg codexConfig, workflowCfg workflowConfig, dryRun bool) error {
	return a.executeParallelRun(ctx, root, specPkg, tasks, prompt, qualityCfg, systemCfg, codexCfg, workflowCfg, dryRun)
}

func (a *App) runCodexExec(ctx context.Context, dir, prompt string) (string, error) {
	if _, err := a.lookPath("codex"); err != nil {
		return "", errors.New("codex is not installed")
	}
	req := executionRequest{
		WorkDir:        dir,
		Prompt:         prompt,
		Runner:         "codex",
		ApprovalPolicy: "on-request",
		SandboxMode:    "workspace-write",
		SessionMode:    "stateful",
	}
	capabilities, err := a.codexCapabilities(ctx, dir, req)
	if err != nil {
		return "", err
	}
	args, err := buildCodexExecArgs(req, capabilities)
	if err != nil {
		return "", err
	}
	return a.runBinary(ctx, "codex", args, dir)
}

func (a *App) runValidators(ctx context.Context, root string, cfg qualityConfig) error {
	for _, step := range validationPipelineSteps(cfg) {
		command := strings.TrimSpace(step.Command)
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

func (a *App) writeOutputs(root string, outputs map[string]string) (outputWriteReport, error) {
	session, err := a.beginManagedOutputSessionAllowMalformedManifest(root)
	if err != nil {
		return outputWriteReport{}, err
	}
	if err := session.writeOutputs(outputs); err != nil {
		return outputWriteReport{}, err
	}
	return session.commit()
}

func (a *App) writeManifest(root string, manifest Manifest) error {
	if a.writeManifestOverride != nil {
		return a.writeManifestOverride(root, manifest)
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(root, manifestPath)
	if err := a.mkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	existing, err := a.readFile(path)
	if err == nil && string(existing) == string(data) {
		return nil
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return a.writeFile(path, data, 0o644)
}

func (a *App) readManifest(root string) (Manifest, error) {
	data, err := a.readFile(filepath.Join(root, manifestPath))
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
	case rel == ".gocache", strings.HasPrefix(rel, ".gocache/"):
		return true
	case rel == ".codex/skills", strings.HasPrefix(rel, ".codex/skills/"):
		return true
	case rel == ".tmp", strings.HasPrefix(rel, ".tmp/"):
		return true
	case rel == "dist", strings.HasPrefix(rel, "dist/"):
		return true
	case rel == "external", strings.HasPrefix(rel, "external/"):
		return true
	case rel == ".namba/logs", strings.HasPrefix(rel, ".namba/logs/"):
		return true
	case rel == ".namba/project/change-summary.md":
		return true
	case rel == ".namba/project/pr-checklist.md":
		return true
	case rel == ".namba/project/release-checklist.md":
		return true
	case rel == ".namba/project/release-notes.md":
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
	return fmt.Sprintf("# Tech\n\n- Language: %s\n- Framework: %s\n- Runtime adapter: Codex\n- Repo-local skills and command-entry skills: .agents/skills\n- Repo-local Codex config: .codex/config.toml\n- Built-in Codex subagents: default, worker, explorer\n- Project-scoped custom agents: .codex/agents/*.toml\n- Readable agent mirrors: .codex/agents/*.md\n- State directory: .namba\n", cfg.Language, cfg.Framework)
}

type jsImportInfo struct {
	Resolved string
	Bindings []string
}

var localJSImportPattern = regexp.MustCompile(`(?m)^\s*import\s+(?:(.+?)\s+from\s+)?["']([^"']+)["']`)
var renderJSXComponentPattern = regexp.MustCompile(`(?s)\.render\(\s*<([A-Z][A-Za-z0-9_]*)\b`)
var renderIdentifierPattern = regexp.MustCompile(`(?s)\.render\(\s*([A-Za-z_][A-Za-z0-9_]*)\s*\)`)

func buildCodemaps(root string, cfg projectConfig) (string, string, string, string) {
	overview := fmt.Sprintf("# Overview\n\n%s is managed by NambaAI.\n\n- Language: %s\n- Framework: %s\n", cfg.Name, cfg.Language, cfg.Framework)
	entries := buildEntryPointsDoc(root, cfg)
	deps := buildDependenciesDoc(root, cfg)
	flow := "# Data Flow\n\n1. `init` runs a Codex-adapted project wizard, writes `.namba/config/sections/*.yaml`, repo skills under `.agents/skills`, command-entry skills such as `$namba-run`, project-scoped custom agents under `.codex/agents/*.toml`, readable `.md` agent mirrors, and Codex repo config under `.codex/config.toml`\n2. `project` refreshes docs and codemaps\n3. `plan` creates a SPEC package\n4. `run` supports the default standalone flow, explicit `--solo` and `--team` subagent-oriented requests, and worktree-based `--parallel` execution\n5. `sync` emits PR-ready artifacts\n"
	return overview, entries, deps, flow
}

func buildEntryPointsDoc(root string, cfg projectConfig) string {
	lines := []string{"# Entry Points", ""}
	var bullets []string

	switch cfg.Language {
	case "go":
		bullets = buildGoEntryPoints(root)
	case "java":
		bullets = buildJavaEntryPoints(root)
	case "python":
		bullets = buildPythonEntryPoints(root)
	default:
		bullets = buildNodeEntryPoints(root)
	}

	if len(bullets) == 0 {
		bullets = append(bullets, "- No conventional application entry point was detected automatically.")
	}

	lines = append(lines, bullets...)
	return strings.Join(lines, "\n") + "\n"
}

func buildDependenciesDoc(root string, cfg projectConfig) string {
	lines := []string{"# Dependencies", ""}
	var bullets []string

	switch cfg.Language {
	case "go":
		if module := readGoModule(root); module != "" {
			bullets = append(bullets, fmt.Sprintf("- Module: `%s`", module))
		}
		bullets = append(bullets, "- Runtime: Go standard library")
		bullets = append(bullets, "- External runtime: Codex CLI")
		bullets = append(bullets, "- External runtime: Git")
	case "java":
		if exists(filepath.Join(root, "pom.xml")) {
			bullets = append(bullets, "- Build system: Maven (`pom.xml`)")
		}
		if exists(filepath.Join(root, "build.gradle")) || exists(filepath.Join(root, "build.gradle.kts")) || exists(filepath.Join(root, "gradlew")) || exists(filepath.Join(root, "gradlew.bat")) {
			bullets = append(bullets, "- Build system: Gradle")
		}
	case "python":
		if exists(filepath.Join(root, "pyproject.toml")) {
			bullets = append(bullets, "- Dependency manifest: `pyproject.toml`")
		}
		if exists(filepath.Join(root, "requirements.txt")) {
			bullets = append(bullets, "- Dependency manifest: `requirements.txt`")
		}
	default:
		bullets = buildNodeDependencies(root)
	}

	if len(bullets) == 0 {
		bullets = append(bullets, "- No dependency manifest was detected automatically.")
	}

	lines = append(lines, bullets...)
	return strings.Join(lines, "\n") + "\n"
}

func buildGoEntryPoints(root string) []string {
	seen := map[string]bool{}
	var bullets []string

	if mainFile := firstExisting(root, "main.go"); mainFile != "" {
		appendEntryPoint(&bullets, seen, mainFile, "application bootstrap")
	}

	cmdDir := filepath.Join(root, "cmd")
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return bullets
	}

	var candidates []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.ToSlash(filepath.Join("cmd", entry.Name(), "main.go"))
		if exists(filepath.Join(root, candidate)) {
			candidates = append(candidates, candidate)
		}
	}
	sort.Strings(candidates)
	for _, candidate := range candidates {
		appendEntryPoint(&bullets, seen, candidate, "Go command entry point")
	}

	return bullets
}

func buildJavaEntryPoints(root string) []string {
	var matches []string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if shouldSkipDiscoveryDir(root, path) {
				return filepath.SkipDir
			}
			return nil
		}
		base := strings.ToLower(filepath.Base(path))
		if strings.HasSuffix(base, "application.java") || base == "main.java" {
			rel, relErr := filepath.Rel(root, path)
			if relErr == nil {
				matches = append(matches, filepath.ToSlash(rel))
			}
		}
		return nil
	})

	sort.Strings(matches)
	seen := map[string]bool{}
	var bullets []string
	for _, match := range matches {
		appendEntryPoint(&bullets, seen, match, "Java application bootstrap")
	}
	return bullets
}

func buildPythonEntryPoints(root string) []string {
	seen := map[string]bool{}
	var bullets []string
	for _, candidate := range []string{"main.py", "app.py", "manage.py", "src/main.py"} {
		if exists(filepath.Join(root, candidate)) {
			appendEntryPoint(&bullets, seen, candidate, "Python application entry point")
		}
	}
	return bullets
}

func buildNodeEntryPoints(root string) []string {
	seen := map[string]bool{}
	var bullets []string

	bootstrap := firstExisting(root, "src/main.tsx", "src/main.jsx", "src/index.tsx", "src/index.jsx", "main.tsx", "main.jsx", "src/main.ts", "src/main.js")
	if bootstrap == "" {
		bootstrap = firstFileContaining(root, "createRoot(")
	}
	if bootstrap != "" {
		appendEntryPoint(&bullets, seen, bootstrap, summarizeEntryPoint(root, bootstrap))
		if appShell := resolveNodeAppShell(root, bootstrap); appShell != "" {
			appendEntryPoint(&bullets, seen, appShell, summarizeEntryPoint(root, appShell))
			if routerModule := firstRouterLikeJSImport(root, appShell); routerModule != "" {
				appendEntryPoint(&bullets, seen, routerModule, summarizeEntryPoint(root, routerModule))
			}
		}
	}

	for _, candidate := range []string{"src/app/routes.ts", "src/app/routes.tsx", "src/routes.ts", "src/routes.tsx", "src/router.ts", "src/router.tsx"} {
		if exists(filepath.Join(root, candidate)) {
			appendEntryPoint(&bullets, seen, candidate, summarizeEntryPoint(root, candidate))
		}
	}

	if routerModule := firstFileContaining(root, "createBrowserRouter("); routerModule != "" {
		appendEntryPoint(&bullets, seen, routerModule, summarizeEntryPoint(root, routerModule))
	}

	return bullets
}

func buildNodeDependencies(root string) []string {
	pkg, err := loadPackageManifest(root)
	if err != nil {
		return nil
	}

	var bullets []string

	if runtime := collectPackageVersions(pkg, "react", "react-dom", "react-router", "react-router-dom", "next"); len(runtime) > 0 {
		bullets = append(bullets, "- Runtime: "+strings.Join(runtime, ", "))
	}

	if build := collectPackageVersions(pkg, "vite", "@vitejs/plugin-react", "typescript", "tailwindcss", "@tailwindcss/vite"); len(build) > 0 {
		bullets = append(bullets, "- Build and styling: "+strings.Join(build, ", "))
	}

	var ui []string
	ui = append(ui, collectPackageVersions(pkg, "@mui/material", "@mui/icons-material", "@emotion/react", "@emotion/styled", "lucide-react")...)
	if radixCount := countPackagesWithPrefix(pkg, "@radix-ui/"); radixCount > 0 {
		ui = append(ui, fmt.Sprintf("%d Radix UI primitives", radixCount))
	}
	if len(ui) > 0 {
		bullets = append(bullets, "- UI system: "+strings.Join(ui, ", "))
	}

	if feature := collectPackageVersions(pkg, "motion", "recharts", "react-hook-form", "date-fns", "sonner", "embla-carousel-react"); len(feature) > 0 {
		bullets = append(bullets, "- Product features: "+strings.Join(feature, ", "))
	}

	if len(bullets) == 0 {
		if fallback := topPackageVersions(pkg, 8); len(fallback) > 0 {
			bullets = append(bullets, "- Packages: "+strings.Join(fallback, ", "))
		}
	}

	return bullets
}

func appendEntryPoint(lines *[]string, seen map[string]bool, rel, summary string) {
	rel = filepath.ToSlash(rel)
	if rel == "" || seen[rel] {
		return
	}
	*lines = append(*lines, fmt.Sprintf("- `%s`: %s", rel, summary))
	seen[rel] = true
}

func summarizeEntryPoint(root, rel string) string {
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		return "application module"
	}

	text := strings.ToLower(string(data))
	switch {
	case strings.Contains(text, "createroot("):
		return "React DOM bootstrap"
	case strings.Contains(text, "routerprovider"):
		return "application shell and router provider"
	case strings.Contains(text, "createbrowserrouter("):
		return "route table and page composition"
	case strings.Contains(text, "express(") || strings.Contains(text, "fastify(") || strings.Contains(text, "app.listen("):
		return "server bootstrap"
	case strings.Contains(text, "export default function app") || strings.Contains(text, "function app("):
		return "top-level application shell"
	default:
		return "application module"
	}
}

func firstResolvedLocalJSImport(root, rel string) string {
	imports := orderedResolvedLocalJSImports(root, rel)
	for _, info := range imports {
		if len(info.Bindings) > 0 {
			return info.Resolved
		}
	}
	if len(imports) > 0 {
		return imports[0].Resolved
	}
	return ""
}

func resolveNodeAppShell(root, bootstrap string) string {
	data, err := os.ReadFile(filepath.Join(root, bootstrap))
	if err != nil {
		return ""
	}

	imports := orderedResolvedLocalJSImports(root, bootstrap)
	if len(imports) == 0 {
		return ""
	}

	bindings := map[string]string{}
	for _, info := range imports {
		for _, binding := range info.Bindings {
			if binding == "" {
				continue
			}
			if _, exists := bindings[binding]; !exists {
				bindings[binding] = info.Resolved
			}
		}
	}

	for _, target := range renderTargetIdentifiers(string(data)) {
		if resolved := bindings[target]; resolved != "" {
			return resolved
		}
	}

	for _, info := range imports {
		for _, binding := range info.Bindings {
			lower := strings.ToLower(binding)
			if lower == "app" || strings.HasSuffix(lower, "app") || strings.Contains(lower, "shell") || strings.Contains(lower, "root") {
				return info.Resolved
			}
		}
	}

	return firstResolvedLocalJSImport(root, bootstrap)
}

func firstRouterLikeJSImport(root, rel string) string {
	for _, info := range orderedResolvedLocalJSImports(root, rel) {
		if looksLikeRouterModule(root, info.Resolved) {
			return info.Resolved
		}
	}
	return ""
}

func orderedResolvedLocalJSImports(root, rel string) []jsImportInfo {
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		return nil
	}

	var infos []jsImportInfo
	for _, match := range localJSImportPattern.FindAllStringSubmatch(string(data), -1) {
		specifier := strings.TrimSpace(match[2])
		if !strings.HasPrefix(specifier, ".") {
			continue
		}
		resolved := resolveJSImport(root, filepath.Dir(rel), specifier)
		if resolved == "" {
			continue
		}
		infos = append(infos, jsImportInfo{
			Resolved: resolved,
			Bindings: parseJSImportBindings(match[1]),
		})
	}
	return infos
}

func parseJSImportBindings(clause string) []string {
	clause = strings.TrimSpace(clause)
	if clause == "" || strings.HasPrefix(clause, "type ") {
		return nil
	}

	var bindings []string
	if brace := strings.Index(clause, "{"); brace >= 0 {
		prefix := strings.TrimSpace(strings.TrimSuffix(clause[:brace], ","))
		if prefix != "" {
			if namespace := parseJSImportNamespace(prefix); namespace != "" {
				bindings = append(bindings, namespace)
			} else {
				bindings = append(bindings, prefix)
			}
		}
		if end := strings.LastIndex(clause, "}"); end > brace {
			for _, part := range strings.Split(clause[brace+1:end], ",") {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				if strings.Contains(part, " as ") {
					parts := strings.Split(part, " as ")
					part = strings.TrimSpace(parts[len(parts)-1])
				}
				if part != "" && !strings.HasPrefix(part, "type ") {
					bindings = append(bindings, part)
				}
			}
		}
		return bindings
	}

	if comma := strings.Index(clause, ","); comma >= 0 {
		defaultBinding := strings.TrimSpace(clause[:comma])
		if defaultBinding != "" {
			bindings = append(bindings, defaultBinding)
		}
		if namespace := parseJSImportNamespace(strings.TrimSpace(clause[comma+1:])); namespace != "" {
			bindings = append(bindings, namespace)
		}
		return bindings
	}

	if namespace := parseJSImportNamespace(clause); namespace != "" {
		return []string{namespace}
	}

	return []string{clause}
}

func parseJSImportNamespace(clause string) string {
	if !strings.HasPrefix(clause, "* as ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(clause, "* as "))
}

func renderTargetIdentifiers(source string) []string {
	seen := map[string]bool{}
	var ids []string
	appendID := func(identifier string) {
		if identifier == "" || seen[identifier] {
			return
		}
		seen[identifier] = true
		ids = append(ids, identifier)
	}

	for _, match := range renderJSXComponentPattern.FindAllStringSubmatch(source, -1) {
		appendID(match[1])
	}
	for _, match := range renderIdentifierPattern.FindAllStringSubmatch(source, -1) {
		identifier := match[1]
		if target := resolveJSXAliasTarget(source, identifier); target != "" {
			appendID(target)
			continue
		}
		appendID(identifier)
	}
	return ids
}

func resolveJSXAliasTarget(source, identifier string) string {
	if identifier == "" {
		return ""
	}
	pattern := regexp.MustCompile(fmt.Sprintf(`(?m)^\s*(?:const|let|var)\s+%s\s*=\s*<([A-Z][A-Za-z0-9_]*)\b`, regexp.QuoteMeta(identifier)))
	match := pattern.FindStringSubmatch(source)
	if len(match) == 2 {
		return match[1]
	}
	return ""
}

func resolveJSImport(root, fromDir, specifier string) string {
	base := filepath.Clean(filepath.Join(fromDir, specifier))
	candidates := []string{base}
	for _, ext := range []string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"} {
		candidates = append(candidates, base+ext)
		candidates = append(candidates, filepath.Join(base, "index"+ext))
	}

	for _, candidate := range candidates {
		if !exists(filepath.Join(root, candidate)) {
			continue
		}
		switch strings.ToLower(filepath.Ext(candidate)) {
		case ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs":
			return filepath.ToSlash(candidate)
		}
	}
	return ""
}

func looksLikeRouterModule(root, rel string) bool {
	if rel == "" {
		return false
	}
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		return false
	}
	text := strings.ToLower(string(data))
	return strings.Contains(text, "createbrowserrouter(") || strings.Contains(text, "routerprovider") || strings.Contains(text, "routeobject")
}

func firstFileContaining(root, needle string) string {
	match := ""
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if shouldSkipDiscoveryDir(root, path) {
				return filepath.SkipDir
			}
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs":
		default:
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		if strings.Contains(string(data), needle) {
			rel, relErr := filepath.Rel(root, path)
			if relErr == nil {
				match = filepath.ToSlash(rel)
				return io.EOF
			}
		}
		return nil
	})
	return match
}

func shouldSkipDiscoveryDir(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		return false
	}
	for _, segment := range strings.Split(rel, "/") {
		switch segment {
		case ".git", ".namba", ".codex", ".agents", "dist", "external", "node_modules":
			return true
		}
	}
	return false
}

func loadPackageManifest(root string) (packageManifest, error) {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return packageManifest{}, err
	}
	var pkg packageManifest
	if err := json.Unmarshal(data, &pkg); err != nil {
		return packageManifest{}, err
	}
	return pkg, nil
}

func collectPackageVersions(pkg packageManifest, names ...string) []string {
	var labels []string
	for _, name := range names {
		if version := packageVersion(pkg, name); version != "" {
			labels = append(labels, fmt.Sprintf("%s@%s", name, version))
		}
	}
	return labels
}

func packageVersion(pkg packageManifest, name string) string {
	for _, group := range []map[string]string{pkg.Dependencies, pkg.DevDependencies, pkg.PeerDependencies} {
		if version, ok := group[name]; ok {
			return version
		}
	}
	return ""
}

func countPackagesWithPrefix(pkg packageManifest, prefix string) int {
	seen := map[string]bool{}
	for _, group := range []map[string]string{pkg.Dependencies, pkg.DevDependencies, pkg.PeerDependencies} {
		for name := range group {
			if strings.HasPrefix(name, prefix) {
				seen[name] = true
			}
		}
	}
	return len(seen)
}

func topPackageVersions(pkg packageManifest, limit int) []string {
	seen := map[string]bool{}
	var names []string
	for _, group := range []map[string]string{pkg.Dependencies, pkg.DevDependencies, pkg.PeerDependencies} {
		for name := range group {
			if seen[name] {
				continue
			}
			seen[name] = true
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if len(names) > limit {
		names = names[:limit]
	}

	var labels []string
	for _, name := range names {
		labels = append(labels, fmt.Sprintf("%s@%s", name, packageVersion(pkg, name)))
	}
	return labels
}

func readGoModule(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "module "))
		}
	}
	return ""
}

func buildFeatureAcceptanceDoc(description, mode string) string {
	bullets := featureAcceptanceCoreLines(description)
	bullets = append(bullets, featureAcceptanceModeLine(mode))
	return strings.Join(bullets, "\n")
}

func buildHarnessAcceptanceDoc(description, mode string) string {
	bullets := harnessAcceptanceCoreLines(description)
	bullets = append(bullets, harnessAcceptanceModeLine(mode))
	return strings.Join(bullets, "\n")
}

func featureAcceptanceCoreLines(description string) []string {
	return []string{
		"# Acceptance",
		"",
		"- [ ] The requested behavior described below is implemented:",
		"  " + description,
		"- [ ] Validation commands pass",
	}
}

func featureAcceptanceModeLine(mode string) string {
	if mode == "tdd" {
		return "- [ ] Tests covering the new behavior are present"
	}
	return "- [ ] Existing behavior is preserved while improving the target area"
}

func harnessAcceptanceCoreLines(description string) []string {
	return []string{
		"# Acceptance",
		"",
		"- [ ] `namba harness \"<description>\"` creates the next sequential `SPEC-XXX` package with a harness-oriented scaffold.",
		"- [ ] `namba plan \"<description>\"` keeps its current default feature-planning behavior.",
		"- [ ] `namba harness --help` is read-only and does not create or mutate `.namba/specs/<SPEC>`.",
		"- [ ] The generated scaffold captures Codex-native execution topology, agent/skill boundaries, progressive-disclosure guidance, trigger strategy, and evaluation strategy.",
		"- [ ] The generated scaffold stays on the existing `.namba/specs/<SPEC>` artifact model and does not invent a second planning package type.",
		"- [ ] The generated scaffold excludes Claude-only primitives such as `.claude/*`, `TeamCreate`, `SendMessage`, `TaskCreate`, and a mandatory `model: \"opus\"` requirement.",
		"- [ ] The requested harness-oriented behavior described below is represented in the scaffold:",
		"  " + description,
		"- [ ] Validation commands pass",
	}
}

func harnessAcceptanceModeLine(mode string) string {
	if mode == "tdd" {
		return "- [ ] Tests covering the new command/scaffold behavior are present"
	}
	return "- [ ] Existing planning behavior is preserved while adding the harness surface"
}

func buildChangeSummaryDoc(root string, projectCfg projectConfig, latestSpec string, profile initProfile) string {
	latestSpec = normalizedLatestSpec(latestSpec)
	lines := changeSummaryHeaderLines(projectCfg, latestSpec)
	lines = append(lines, "")
	lines = append(lines, changeSummaryWorkflowDocsSection(profile)...)
	lines = append(lines, "")
	lines = append(lines, changeSummaryRefreshCommandsSection()...)
	if readinessLines := changeSummaryLatestReviewReadinessSection(root, latestSpec); len(readinessLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, readinessLines...)
	}
	if proofLines := changeSummaryLatestExecutionProofSection(root); len(proofLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, proofLines...)
	}
	return strings.Join(lines, "\n") + "\n"
}

func normalizedProjectType(projectCfg projectConfig) string {
	projectType := projectCfg.ProjectType
	if strings.TrimSpace(projectType) == "" {
		projectType = "existing"
	}
	return projectType
}

func normalizedLatestSpec(latestSpec string) string {
	if strings.TrimSpace(latestSpec) == "" {
		return "none"
	}
	return latestSpec
}

func changeSummaryHeaderLines(projectCfg projectConfig, latestSpec string) []string {
	return []string{
		"# Change Summary",
		"",
		fmt.Sprintf("Project: %s", projectCfg.Name),
		fmt.Sprintf("Project type: %s", normalizedProjectType(projectCfg)),
		fmt.Sprintf("Latest SPEC: %s", normalizedLatestSpec(latestSpec)),
	}
}

func changeSummaryWorkflowDocsSection(profile initProfile) []string {
	return []string{
		"## Workflow Docs Synced",
		"",
		"- README bundles and product docs describe when to use `namba update`, `namba regen`, `namba sync`, `namba pr`, and `namba land`.",
		"- Release docs describe `namba release` guardrails on a clean `main` branch plus optional `--push` behavior.",
		"- Run docs separate the default standalone flow, `namba run SPEC-XXX --solo`, `namba run SPEC-XXX --team`, and the worktree fan-out policy for `namba run SPEC-XXX --parallel`.",
		"- AGENTS and Codex docs define the Namba output contract plus the fallback validator script at `.namba/codex/validate-output-contract.py`.",
		"- SPEC packages can keep advisory plan-review artifacts under `.namba/specs/<SPEC>/reviews/` so product, engineering, and design review state stays visible before execution and PR handoff.",
		fmt.Sprintf("- Collaboration docs require one branch per SPEC/task from `%s`, PRs into `%s`, %s PR content, and Codex review requests via `%s`.", branchBase(profile), prBaseBranch(profile), strings.ToLower(humanLanguageName(profile.PRLanguage)), codexReviewComment(profile)),
	}
}

func changeSummaryRefreshCommandsSection() []string {
	return []string{
		"## Refresh Commands",
		"",
		"- `namba update` self-updates the installed `namba` binary from GitHub Release assets.",
		"- `namba regen` regenerates `AGENTS.md`, repo-local skills and command-entry skills under `.agents/skills`, `.codex/agents/*.toml` custom agents, readable `.md` role-card mirrors, `.namba/codex/*`, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.",
		"- `namba sync` refreshes `.namba/project/*` docs, release notes/checklists, codemaps, and any README bundles enabled in `.namba/config/sections/docs.yaml`.",
		"- `namba pr` prepares the current branch for GitHub review by running sync and validation by default, then committing, pushing, opening or reusing the PR, and ensuring the Codex review marker exists.",
		"- `namba land` optionally waits for checks, merges only when the PR is clean, and updates local `main` safely.",
	}
}

func changeSummaryLatestReviewReadinessSection(root, latestSpec string) []string {
	if !specReviewReadinessExists(root, latestSpec) {
		return nil
	}
	return []string{
		"## Latest Review Readiness",
		"",
		fmt.Sprintf("- Latest readiness artifact: `%s`", specReviewReadinessPath(latestSpec)),
		fmt.Sprintf("- Advisory summary: %s", specReadinessAdvisorySummary(root, latestSpec)),
	}
}

func buildPRChecklistDoc(root, latestSpec string, profile initProfile) string {
	lines := prChecklistHeaderLines()
	lines = append(lines, prChecklistCoreItems(profile)...)
	lines = append(lines, prChecklistLatestReviewReadinessItem(root, latestSpec)...)
	lines = append(lines, prChecklistLatestExecutionProofItem(root)...)
	return strings.Join(lines, "\n") + "\n"
}

func prChecklistHeaderLines() []string {
	return []string{
		"# PR Checklist",
		"",
	}
}

func prChecklistCoreItems(profile initProfile) []string {
	return []string{
		fmt.Sprintf("- [ ] Dedicated work branch created from `%s` for this SPEC/task", branchBase(profile)),
		fmt.Sprintf("- [ ] PR targets `%s`", prBaseBranch(profile)),
		fmt.Sprintf("- [ ] PR title and body are written in %s", humanLanguageName(profile.PRLanguage)),
		fmt.Sprintf("- [ ] `%s` review request is present on GitHub", codexReviewComment(profile)),
		"- [ ] README / user-facing docs refreshed",
		"- [ ] `namba regen` rerun if template-generated Codex assets changed",
		"- [ ] `namba sync` artifacts refreshed",
		"- [ ] `namba pr` used for the GitHub review handoff when the branch is ready",
		"- [ ] SPEC artifacts reviewed",
		"- [ ] Validation commands passed",
		"- [ ] Diff reviewed",
	}
}

func prChecklistLatestReviewReadinessItem(root, latestSpec string) []string {
	if !specReviewReadinessExists(root, latestSpec) {
		return nil
	}
	return []string{fmt.Sprintf("- [ ] Latest SPEC review readiness checked: `%s`", specReviewReadinessPath(latestSpec))}
}

func buildReleaseNotesDoc(projectCfg projectConfig, latestSpec string, profile initProfile) string {
	lines := releaseNotesHeaderLines(projectCfg, latestSpec)
	lines = append(lines, "")
	lines = append(lines, releaseNotesWorkflowChangesSection(profile)...)
	lines = append(lines, "")
	lines = append(lines, releaseNotesGuardrailsSection()...)
	lines = append(lines, "")
	lines = append(lines, releaseNotesCommandsSection()...)
	lines = append(lines, "")
	lines = append(lines, releaseNotesExpectedAssetsSection()...)
	return strings.Join(lines, "\n") + "\n"
}

func releaseNotesHeaderLines(projectCfg projectConfig, latestSpec string) []string {
	return []string{
		"# Release Notes Draft",
		"",
		fmt.Sprintf("Project: %s", projectCfg.Name),
		fmt.Sprintf("Project type: %s", normalizedProjectType(projectCfg)),
		fmt.Sprintf("Reference SPEC: %s", normalizedLatestSpec(latestSpec)),
	}
}

func releaseNotesWorkflowChangesSection(profile initProfile) []string {
	return []string{
		"## Workflow Changes",
		"",
		"- `namba update` self-updates the installed `namba` binary from GitHub Release assets.",
		"- `namba regen` regenerates `AGENTS.md`, repo-local skills and command-entry skills under `.agents/skills`, `.codex/agents/*.toml` custom agents, readable `.md` role-card mirrors, and repo-local Codex config from `.namba/config/sections/*.yaml`.",
		"- `namba sync` refreshes README bundles, product docs, codemaps, change summary, PR checklist, and release docs.",
		"- `namba pr` prepares the current branch for GitHub review by syncing, validating, committing, pushing, opening or reusing the PR, and ensuring the Codex review marker exists.",
		"- `namba land` optionally waits for checks, merges only when the PR is clean, and updates local `main` safely.",
		"- `namba run SPEC-XXX` keeps the standard standalone Codex flow; `--solo` and `--team` request single-subagent or multi-subagent workflows inside one workspace; `--parallel` still fans out into up to three git worktrees and merges only after every worker passes execution and validation.",
		fmt.Sprintf("- Active collaboration defaults: one branch per SPEC/task from `%s`, PRs into `%s`, %s PR content, and Codex review requests via `%s`.", branchBase(profile), prBaseBranch(profile), strings.ToLower(humanLanguageName(profile.PRLanguage)), codexReviewComment(profile)),
	}
}

func releaseNotesGuardrailsSection() []string {
	return []string{
		"## Release Guardrails",
		"",
		"- `namba release` requires a git repository, the `main` branch, and a clean working tree.",
		"- Validators from `.namba/config/sections/quality.yaml` run before the release tag is created.",
		"- With no explicit version, `namba release` defaults to the next `patch` tag. Use `--bump minor|major` or `--version vX.Y.Z` when needed.",
		"- `namba release --push` pushes both `main` and the new tag to the selected remote.",
	}
}

func releaseNotesCommandsSection() []string {
	return []string{
		"## Release Commands",
		"",
		"```text",
		"namba sync",
		"namba pr \"release review\"",
		"namba land",
		"namba release --bump patch",
		"# or",
		"namba release --version vX.Y.Z --push",
		"```",
	}
}

func releaseNotesExpectedAssetsSection() []string {
	return []string{
		"## Expected Assets",
		"",
		"- `namba_Windows_x86.zip`",
		"- `namba_Windows_x86_64.zip`",
		"- `namba_Windows_arm64.zip`",
		"- `namba_Linux_x86_64.tar.gz`",
		"- `namba_Linux_arm64.tar.gz`",
		"- `namba_macOS_x86_64.tar.gz`",
		"- `namba_macOS_arm64.tar.gz`",
		"- `checksums.txt`",
	}
}

func buildReleaseChecklistDoc() string {
	lines := releaseChecklistHeaderLines()
	lines = append(lines, releaseChecklistItems()...)
	return strings.Join(lines, "\n") + "\n"
}

func releaseChecklistHeaderLines() []string {
	return []string{
		"# Release Checklist",
		"",
	}
}

func releaseChecklistItems() []string {
	return []string{
		"- [ ] `namba regen` rerun if template-generated Codex assets changed",
		"- [ ] `namba sync` artifacts refreshed",
		"- [ ] `namba pr` used for the GitHub review handoff when the branch is ready",
		"- [ ] README and `.namba/codex/README.md` reflect update, release, and parallel workflow behavior",
		"- [ ] Working tree is clean and the current branch is `main`",
		"- [ ] Validation commands passed",
		"- [ ] `namba release --version vX.Y.Z` or `namba release --bump patch` executed",
		"- [ ] If `--push` was not used, `main` and the release tag were pushed manually",
		"- [ ] GitHub Release workflow completed and published assets plus `checksums.txt`",
	}
}

func buildFixAcceptanceDoc(description, mode string) string {
	bullets := fixAcceptanceCoreLines(description)
	bullets = append(bullets, fixAcceptanceModeLine(mode))
	return strings.Join(bullets, "\n")
}

func fixAcceptanceCoreLines(description string) []string {
	return []string{
		"# Acceptance",
		"",
		"- [ ] The reported issue described below is resolved:",
		"  " + description,
		"- [ ] Validation commands pass",
		"- [ ] Existing behavior around the affected area is preserved",
	}
}

func fixAcceptanceModeLine(mode string) string {
	if mode == "tdd" {
		return "- [ ] A regression test covering the fix is present"
	}
	return "- [ ] A targeted reproduction or verification step is documented"
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

func findManifestEntry(manifest Manifest, path string) (ManifestEntry, bool) {
	for _, entry := range manifest.Entries {
		if entry.Path == path {
			return entry, true
		}
	}
	return ManifestEntry{}, false
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

func manifestEntryIsManaged(entry ManifestEntry, managed func(string) bool, ownedManaged func(ManifestEntry) bool) bool {
	if strings.TrimSpace(entry.Owner) != "" {
		if ownedManaged != nil {
			return ownedManaged(entry)
		}
		return entry.Owner == manifestOwnerManaged && managed(entry.Path)
	}
	return managed(entry.Path)
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
		AutoCodexReview:       true,
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
	applyHumanLanguage(&profile, promptSelect(a.stdin, a.stdout, "\U0001f310 \uc791\uc5c5 \uc5b8\uc5b4", languageOptions(), profile.ConversationLanguage))
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
	profile, err := a.promptCodexAccess(reader, profile)
	if err != nil {
		return initProfile{}, err
	}
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

func approvalPolicyOptions() []option {
	return []option{
		{Value: "on-request", Label: "on-request", Description: "\ud544\uc694\ud560 \ub54c Codex\uac00 \uc2b9\uc778 \uc694\uccad"},
		{Value: "untrusted", Label: "untrusted", Description: "\ubbff\uc744 \uc218 \uc5c6\ub294 \uc791\uc5c5\ub9cc \uc2b9\uc778 \ud655\uc778"},
		{Value: "never", Label: "never", Description: "\uc2b9\uc778 \uc5c6\uc774 \uacc4\uc18d \uc9c4\ud589"},
	}
}

func sandboxModeOptions() []option {
	return []option{
		{Value: "workspace-write", Label: "workspace-write", Description: "\ud604\uc7ac \uc791\uc5c5 \uacf5\uac04\ub9cc \uc4f0\uae30 \ud5c8\uc6a9"},
		{Value: "read-only", Label: "read-only", Description: "\ud30c\uc77c \uc4f0\uae30 \uc5c6\uc774 \uc77d\uae30 \uc804\uc6a9"},
		{Value: "danger-full-access", Label: "danger-full-access", Description: "\uc0cc\ub4dc\ubc15\uc2a4 \uc81c\ud55c \uc5c6\uc774 \uc804\uccb4 \uc811\uadfc"},
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
