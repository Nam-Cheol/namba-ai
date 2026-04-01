package namba

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) runRegen(_ context.Context, args []string) error {
	if len(args) != 0 {
		return errors.New("regen does not accept arguments")
	}

	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}

	profile, err := a.loadInitProfileFromConfig(root)
	if err != nil {
		return err
	}

	outputs := map[string]string{
		"AGENTS.md": renderAgents(profile),
	}
	for rel, body := range codexScaffoldFiles(profile) {
		outputs[rel] = body
	}
	if err := removeLegacyCodexSkillMirror(root); err != nil {
		return err
	}
	report, err := a.replaceManagedOutputs(root, outputs, isRegenManagedPath)
	if err != nil {
		return err
	}

	fmt.Fprintln(a.stdout, "Regenerated NambaAI AGENTS, repo skills, command-entry skills, Codex agents, and Codex config.")
	if len(report.InstructionSurfacePaths) > 0 {
		fmt.Fprintf(a.stdout, "Session refresh required: start a fresh Codex session before continuing long team or repair runs (%s)\n", strings.Join(report.InstructionSurfacePaths, ", "))
	}
	return nil
}

func (a *App) replaceManagedOutputs(root string, outputs map[string]string, managed func(string) bool) (outputWriteReport, error) {
	manifest, err := a.readManifest(root)
	if err != nil {
		return outputWriteReport{}, err
	}

	filtered := manifest.Entries[:0]
	for _, entry := range manifest.Entries {
		if managed(entry.Path) {
			if _, keep := outputs[entry.Path]; keep {
				filtered = append(filtered, entry)
				continue
			}
			if err := os.RemoveAll(filepath.Join(root, filepath.FromSlash(entry.Path))); err != nil && !errors.Is(err, os.ErrNotExist) {
				return outputWriteReport{}, fmt.Errorf("remove obsolete generated file %s: %w", entry.Path, err)
			}
			continue
		}
		filtered = append(filtered, entry)
	}
	manifest.Entries = filtered
	if err := a.writeManifest(root, manifest); err != nil {
		return outputWriteReport{}, err
	}
	return a.writeOutputs(root, outputs)
}

func isRegenManagedPath(rel string) bool {
	switch {
	case rel == "AGENTS.md":
		return true
	case rel == repoCodexConfigPath:
		return true
	case strings.HasPrefix(rel, repoSkillsDir+"/"):
		return true
	case strings.HasPrefix(rel, ".codex/skills/"):
		return true
	case strings.HasPrefix(rel, repoCodexAgentsDir+"/"):
		return true
	case strings.HasPrefix(rel, codexStateDir+"/"):
		return true
	default:
		return false
	}
}

func removeLegacyCodexSkillMirror(root string) error {
	for _, rel := range legacyCodexSkillMirrorPaths() {
		if err := os.RemoveAll(filepath.Join(root, filepath.FromSlash(rel))); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove deprecated codex skill mirror %s: %w", rel, err)
		}
	}
	legacyRoot := filepath.Join(root, ".codex", "skills")
	entries, err := os.ReadDir(legacyRoot)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read deprecated codex skills dir: %w", err)
	}
	if len(entries) == 0 {
		if err := os.Remove(legacyRoot); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove empty deprecated codex skills dir: %w", err)
		}
	}
	return nil
}

func legacyCodexSkillMirrorPaths() []string {
	names := []string{
		"namba",
		"namba-init",
		"namba-project",
		"namba-regen",
		"namba-update",
		"namba-plan",
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
	paths := make([]string, 0, len(names))
	for _, name := range names {
		paths = append(paths, filepath.ToSlash(filepath.Join(".codex", "skills", name)))
	}
	return paths
}
