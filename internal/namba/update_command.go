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
	if handled, err := a.handleNoArgTopLevelCommand("regen", args); handled {
		return err
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
	report, err := a.replaceManagedOutputs(root, outputs, isRegenManagedPath, isOwnedRegenManagedPath)
	if err != nil {
		return err
	}

	fmt.Fprintln(a.stdout, "Regenerated NambaAI AGENTS, repo skills, command-entry skills, Codex agents, and Codex config.")
	if len(report.InstructionSurfacePaths) > 0 {
		fmt.Fprintf(a.stdout, "Session refresh required: start a fresh Codex session before continuing long team or repair runs (%s)\n", strings.Join(report.InstructionSurfacePaths, ", "))
	}
	return nil
}

func (a *App) replaceManagedOutputs(root string, outputs map[string]string, managed func(string) bool, ownedManaged func(ManifestEntry) bool) (outputWriteReport, error) {
	session, err := a.beginManagedOutputSession(root)
	if err != nil {
		return outputWriteReport{}, err
	}
	if err := session.replaceManagedOutputs(outputs, managed, ownedManaged); err != nil {
		return outputWriteReport{}, err
	}
	return session.commit()
}

func isOwnedRegenManagedPath(entry ManifestEntry) bool {
	if entry.Owner != manifestOwnerManaged {
		return false
	}

	rel := entry.Path
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

func isRegenManagedPath(rel string) bool {
	switch {
	case rel == "AGENTS.md":
		return true
	case rel == repoCodexConfigPath:
		return true
	case isManagedRepoSkillPath(rel):
		return true
	case strings.HasPrefix(rel, ".codex/skills/"):
		return true
	case isManagedRepoCodexAgentPath(rel):
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
	names := managedCodexSkillNames()
	paths := make([]string, 0, len(names))
	for _, name := range names {
		paths = append(paths, filepath.ToSlash(filepath.Join(".codex", "skills", name)))
	}
	return paths
}
