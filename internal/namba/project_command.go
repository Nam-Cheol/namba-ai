package namba

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func (a *App) runDoctor(ctx context.Context, args []string) error {
	checkUpdate := false
	if wantsCommandHelp(args) {
		return a.printCommandUsage("doctor")
	}
	for _, arg := range args {
		switch arg {
		case "--check-update":
			checkUpdate = true
		default:
			if strings.HasPrefix(arg, "--") {
				return commandUsageError("doctor", fmt.Errorf("unknown flag %q", arg))
			}
			return commandUsageError("doctor", errors.New("doctor does not accept arguments"))
		}
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
	if checkUpdate {
		fmt.Fprint(a.stdout, formatVersionAdvisory(a.refreshVersionAdvisory(ctx), true))
	} else {
		a.printCachedVersionAdvisory()
	}
	return nil
}

func (a *App) runStatus(_ context.Context, args []string) error {
	if wantsCommandHelp(args) {
		return a.printCommandUsage("status")
	}
	statusJSON := false
	for _, arg := range args {
		switch arg {
		case "--json":
			statusJSON = true
		default:
			if strings.HasPrefix(arg, "--") {
				return commandUsageError("status", fmt.Errorf("unknown flag %q", arg))
			}
			return commandUsageError("status", errors.New("status does not accept arguments"))
		}
	}

	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}
	if statusJSON {
		report := collectNambaReport(root, a.now(), reportOptions{})
		output, err := renderStatusJSON(report)
		if err != nil {
			return commandExitError(2, err)
		}
		fmt.Fprint(a.stdout, output)
		return nil
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
	a.printCachedVersionAdvisory()
	return nil
}

func (a *App) runProject(ctx context.Context, args []string) error {
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
	systemCfg, _ := a.loadSystemConfig(root)
	codexCfg, _ := a.loadCodexConfig(root)
	diagnostics := a.buildCodexDiagnosticsEvidence(ctx, root, codexDiagnosticsOptions{
		LogDir:                filepath.ToSlash(filepath.Join(logsDir, "project")),
		LogPrefix:             "codex",
		RunCommands:           true,
		SystemConfig:          systemCfg,
		ConfiguredRoots:       append([]string{root}, codexCfg.AddDirs...),
		IncludeDoctorLogFiles: true,
	})
	projectEvidence := projectCodexDiagnosticsEvidence{
		SchemaVersion: projectCodexDiagnosticsSchema,
		GeneratedAt:   a.now().Format(time.RFC3339),
		ProjectRoot:   root,
		Diagnostics:   diagnostics,
	}
	if err := writeJSONFile(filepath.Join(root, logsDir, "project", "codex-diagnostics-evidence.json"), projectEvidence); err != nil {
		return err
	}
	if diagnostics.WorkspaceRootComparison.Status == "advisory_mismatch" {
		fmt.Fprintf(a.stdout, "Codex workspace advisory: %s Compared roots: %s\n", diagnostics.WorkspaceRootComparison.Message, strings.Join(diagnostics.WorkspaceRootComparison.Compared, ", "))
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
	a.printCachedVersionAdvisory()
	return nil
}
