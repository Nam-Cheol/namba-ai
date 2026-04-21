package namba

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const currentWorkspacePlanningFlag = "--current-workspace"

type planningStartOptions struct {
	Kind             string
	Description      string
	CurrentWorkspace bool
}

type planningStartResolution struct {
	Root                  string
	SpecID                string
	Branch                string
	WorktreePath          string
	WorkspaceAction       string
	NextStep              string
	CreatedIsolatedTarget bool
}

func (a *App) resolvePlanningStart(ctx context.Context, currentRoot string, options planningStartOptions) (planningStartResolution, error) {
	if !isGitRepository(currentRoot) {
		specID, err := nextSpecID(filepath.Join(currentRoot, specsDir))
		if err != nil {
			return planningStartResolution{}, err
		}
		return planningStartResolution{
			Root:            currentRoot,
			SpecID:          specID,
			Branch:          "n/a (git worktree isolation unavailable)",
			WorktreePath:    currentRoot,
			WorkspaceAction: "scaffolded in current workspace",
			NextStep:        "continue in the current workspace; git worktree isolation is unavailable here.",
		}, nil
	}

	profile, err := a.loadInitProfileFromConfig(currentRoot)
	if err != nil {
		return planningStartResolution{}, err
	}
	if !profile.BranchPerWork {
		specID, err := nextSpecID(filepath.Join(currentRoot, specsDir))
		if err != nil {
			return planningStartResolution{}, err
		}
		return planningStartResolution{
			Root:            currentRoot,
			SpecID:          specID,
			Branch:          "n/a (branch-per-work disabled)",
			WorktreePath:    currentRoot,
			WorkspaceAction: "scaffolded in current workspace",
			NextStep:        "continue in the current workspace; branch-per-work is disabled in git strategy.",
		}, nil
	}

	worktrees, err := a.planningWorktrees(ctx, currentRoot)
	if err != nil {
		return planningStartResolution{}, err
	}
	specID, err := nextPlanningSpecID(worktrees)
	if err != nil {
		return planningStartResolution{}, err
	}

	slug, err := normalizeCreateSlug(options.Description)
	if err != nil {
		return planningStartResolution{}, fmt.Errorf("normalize planning slug: %w", err)
	}
	baseBranch := branchBase(profile)
	specPrefix := specBranchPrefix(profile)
	sharedRoot := sharedWorktreeRoot(worktrees, currentRoot, baseBranch)
	targetBranch := specPrefix + specID + "-" + slug
	targetPath := filepath.Join(sharedRoot, worktreesDir, strings.ToLower(specID+"-"+slug))

	currentBranch, err := a.currentBranch(ctx, currentRoot)
	if err != nil {
		return planningStartResolution{}, fmt.Errorf("detect current branch: %w", err)
	}
	dirty, err := a.hasWorkingTreeChanges(ctx, currentRoot)
	if err != nil {
		return planningStartResolution{}, err
	}

	if options.CurrentWorkspace {
		return planningStartResolution{
			Root:            currentRoot,
			SpecID:          specID,
			Branch:          currentBranch,
			WorktreePath:    currentRoot,
			WorkspaceAction: "explicit current-workspace override",
			NextStep:        fmt.Sprintf("continue in the current workspace; %s was used intentionally.", currentWorkspacePlanningFlag),
		}, nil
	}

	currentIsShared := samePlanningPath(currentRoot, sharedRoot)
	currentIsDedicated := !currentIsShared &&
		containsPlanningWorktreePath(worktrees, currentRoot) &&
		!dirty &&
		currentBranch != baseBranch &&
		strings.HasPrefix(currentBranch, specPrefix)
	if currentIsDedicated {
		return planningStartResolution{
			Root:            currentRoot,
			SpecID:          specID,
			Branch:          currentBranch,
			WorktreePath:    currentRoot,
			WorkspaceAction: "reused current isolated workspace",
			NextStep:        "continue in the current workspace.",
		}, nil
	}

	if dirty {
		action := "refused due to dirty ambiguous workspace"
		nextStep := fmt.Sprintf("clean or commit existing changes, or rerun with %s if this workspace is the intended scaffold target.", currentWorkspacePlanningFlag)
		if currentIsShared || currentBranch == baseBranch {
			action = "refused due to dirty shared/base workspace"
			nextStep = fmt.Sprintf("clean or commit the shared workspace first, or rerun with %s only if in-place scaffolding is intentional.", currentWorkspacePlanningFlag)
		}
		return planningStartResolution{
			Root:            currentRoot,
			SpecID:          specID,
			Branch:          currentBranch,
			WorktreePath:    currentRoot,
			WorkspaceAction: action,
			NextStep:        nextStep,
		}, fmt.Errorf("unsafe planning context")
	}

	if _, err := a.runBinary(ctx, "git", []string{"worktree", "add", "-b", targetBranch, targetPath, baseBranch}, sharedRoot); err != nil {
		return planningStartResolution{}, fmt.Errorf("create isolated worktree %s on %s from %s: %w", targetPath, targetBranch, baseBranch, err)
	}
	return planningStartResolution{
		Root:                  targetPath,
		SpecID:                specID,
		Branch:                targetBranch,
		WorktreePath:          targetPath,
		WorkspaceAction:       "created isolated workspace",
		NextStep:              fmt.Sprintf("cd %s", targetPath),
		CreatedIsolatedTarget: true,
	}, nil
}

func (a *App) planningWorktrees(ctx context.Context, root string) ([]gitWorktree, error) {
	out, err := a.runBinary(ctx, "git", worktreeListArgs(), root)
	if err != nil {
		return nil, fmt.Errorf("list git worktrees: %w", err)
	}
	return parseGitWorktrees(out), nil
}

func nextPlanningSpecID(worktrees []gitWorktree) (string, error) {
	maxID := 0
	seen := map[string]struct{}{}
	for _, worktree := range worktrees {
		path := filepath.Clean(worktree.Path)
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		accessible, err := isAccessiblePlanningWorkspace(path)
		if err != nil {
			return "", err
		}
		if !accessible {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(path, specsDir))
		if err != nil {
			if os.IsNotExist(err) || os.IsPermission(err) {
				continue
			}
			return "", err
		}
		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "SPEC-") {
				continue
			}
			n, err := strconv.Atoi(strings.TrimPrefix(entry.Name(), "SPEC-"))
			if err == nil && n > maxID {
				maxID = n
			}
		}
	}
	return fmt.Sprintf("SPEC-%03d", maxID+1), nil
}

func isAccessiblePlanningWorkspace(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) || os.IsPermission(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func sharedWorktreeRoot(worktrees []gitWorktree, fallback, baseBranch string) string {
	for _, worktree := range worktrees {
		info, err := os.Stat(filepath.Join(worktree.Path, ".git"))
		if err == nil && info.IsDir() {
			return filepath.Clean(worktree.Path)
		}
	}

	for _, worktree := range worktrees {
		if worktree.Branch != baseBranch {
			continue
		}
		if accessiblePlanningPath(worktree.Path) {
			return filepath.Clean(worktree.Path)
		}
	}

	if accessiblePlanningPath(fallback) {
		return filepath.Clean(fallback)
	}

	for _, worktree := range worktrees {
		if accessiblePlanningPath(worktree.Path) {
			return filepath.Clean(worktree.Path)
		}
	}
	return filepath.Clean(fallback)
}

func accessiblePlanningPath(path string) bool {
	accessible, err := isAccessiblePlanningWorkspace(path)
	return err == nil && accessible
}

func containsPlanningWorktreePath(worktrees []gitWorktree, path string) bool {
	for _, worktree := range worktrees {
		if samePlanningPath(worktree.Path, path) {
			return true
		}
	}
	return false
}

func samePlanningPath(left, right string) bool {
	return normalizePlanningPath(left) == normalizePlanningPath(right)
}

func normalizePlanningPath(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil && resolved != "" {
		return filepath.Clean(resolved)
	}
	if abs, err := filepath.Abs(path); err == nil && abs != "" {
		return filepath.Clean(abs)
	}
	return filepath.Clean(path)
}

func formatPlanningStartSummary(start planningStartResolution) string {
	lines := []string{
		"Planning start:",
		fmt.Sprintf("- SPEC: %s", start.SpecID),
		fmt.Sprintf("- Workspace action: %s", start.WorkspaceAction),
		fmt.Sprintf("- Branch: %s", firstNonBlank(start.Branch, "n/a")),
		fmt.Sprintf("- Worktree: %s", firstNonBlank(start.WorktreePath, start.Root)),
		fmt.Sprintf("- Next step: %s", start.NextStep),
	}
	return strings.Join(lines, "\n") + "\n"
}

func wrapPlanningScaffoldFailure(start planningStartResolution, err error) error {
	if !start.CreatedIsolatedTarget {
		return err
	}
	return fmt.Errorf("%sscaffolding failed after creating the isolated workspace; resume in %s or remove it manually if you do not want to keep it: %w", formatPlanningStartSummary(start), start.Root, err)
}
