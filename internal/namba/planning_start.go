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
	Root            string
	SpecID          string
	Branch          string
	WorkspacePath   string
	WorkspaceAction string
	NextStep        string
	CreatedBranch   bool
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
			Branch:          "n/a (git branch automation unavailable)",
			WorkspacePath:   currentRoot,
			WorkspaceAction: "scaffolded in current workspace",
			NextStep:        "continue in the current workspace; git branch automation is unavailable here.",
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
			WorkspacePath:   currentRoot,
			WorkspaceAction: "scaffolded in current workspace",
			NextStep:        "continue in the current workspace; branch-per-work is disabled in git strategy.",
		}, nil
	}

	worktrees, err := a.planningWorktrees(ctx, currentRoot)
	if err != nil {
		return planningStartResolution{}, err
	}
	specID, err := a.nextPlanningSpecID(ctx, currentRoot, worktrees)
	if err != nil {
		return planningStartResolution{}, err
	}

	slug, err := normalizeCreateSlug(options.Description)
	if err != nil {
		return planningStartResolution{}, fmt.Errorf("normalize planning slug: %w", err)
	}
	baseBranch := branchBase(profile)
	specPrefix := specBranchPrefix(profile)
	targetBranch := specPrefix + specID + "-" + slug

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
			WorkspacePath:   currentRoot,
			WorkspaceAction: "explicit current-workspace override",
			NextStep:        fmt.Sprintf("continue in the current workspace on branch %s; %s was used intentionally.", firstNonBlank(currentBranch, "HEAD"), currentWorkspacePlanningFlag),
		}, nil
	}

	if dirty {
		return planningStartResolution{
			Root:            currentRoot,
			SpecID:          specID,
			Branch:          currentBranch,
			WorkspacePath:   currentRoot,
			WorkspaceAction: "refused due to dirty workspace",
			NextStep:        fmt.Sprintf("clean or commit existing changes, or rerun with %s only if you intentionally want to scaffold on the current branch without creating a dedicated SPEC branch.", currentWorkspacePlanningFlag),
		}, fmt.Errorf("unsafe planning context")
	}

	if currentBranch == targetBranch {
		return planningStartResolution{
			Root:            currentRoot,
			SpecID:          specID,
			Branch:          currentBranch,
			WorkspacePath:   currentRoot,
			WorkspaceAction: "reused current SPEC branch",
			NextStep:        fmt.Sprintf("continue in the current workspace on branch %s.", currentBranch),
		}, nil
	}

	exists, err := a.localBranchExists(ctx, currentRoot, targetBranch)
	if err != nil {
		return planningStartResolution{}, err
	}
	if exists {
		if _, err := a.runBinary(ctx, "git", []string{"checkout", targetBranch}, currentRoot); err != nil {
			return planningStartResolution{}, fmt.Errorf("checkout existing SPEC branch %s: %w", targetBranch, err)
		}
		return planningStartResolution{
			Root:            currentRoot,
			SpecID:          specID,
			Branch:          targetBranch,
			WorkspacePath:   currentRoot,
			WorkspaceAction: "checked out existing SPEC branch in current workspace",
			NextStep:        fmt.Sprintf("continue in the current workspace on branch %s.", targetBranch),
		}, nil
	}

	if _, err := a.runBinary(ctx, "git", []string{"checkout", "-b", targetBranch, baseBranch}, currentRoot); err != nil {
		return planningStartResolution{}, fmt.Errorf("create SPEC branch %s from %s: %w", targetBranch, baseBranch, err)
	}
	return planningStartResolution{
		Root:            currentRoot,
		SpecID:          specID,
		Branch:          targetBranch,
		WorkspacePath:   currentRoot,
		WorkspaceAction: "created dedicated SPEC branch in current workspace",
		NextStep:        fmt.Sprintf("continue in the current workspace on branch %s.", targetBranch),
		CreatedBranch:   true,
	}, nil
}

func (a *App) planningWorktrees(ctx context.Context, root string) ([]gitWorktree, error) {
	out, err := a.runBinary(ctx, "git", worktreeListArgs(), root)
	if err != nil {
		return nil, fmt.Errorf("list git worktrees: %w", err)
	}
	return parseGitWorktrees(out), nil
}

func (a *App) nextPlanningSpecID(ctx context.Context, root string, worktrees []gitWorktree) (string, error) {
	maxID, err := maxPlanningSpecIDFromWorktrees(worktrees)
	if err != nil {
		return "", err
	}

	branches, err := a.localBranches(ctx, root)
	if err != nil {
		return "", err
	}
	branchMaxID, err := a.maxPlanningSpecIDFromBranches(ctx, root, branches)
	if err != nil {
		return "", err
	}
	if branchMaxID > maxID {
		maxID = branchMaxID
	}
	return fmt.Sprintf("SPEC-%03d", maxID+1), nil
}

func nextPlanningSpecID(worktrees []gitWorktree) (string, error) {
	maxID, err := maxPlanningSpecIDFromWorktrees(worktrees)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("SPEC-%03d", maxID+1), nil
}

func maxPlanningSpecIDFromWorktrees(worktrees []gitWorktree) (int, error) {
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
			return 0, err
		}
		if !accessible {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(path, specsDir))
		if err != nil {
			if os.IsNotExist(err) || os.IsPermission(err) {
				continue
			}
			return 0, err
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
	return maxID, nil
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

func (a *App) localBranchExists(ctx context.Context, root, branch string) (bool, error) {
	out, err := a.runBinary(ctx, "git", []string{"branch", "--list", branch}, root)
	if err != nil {
		return false, fmt.Errorf("check existing branch %s: %w", branch, err)
	}
	return strings.TrimSpace(out) != "", nil
}

func (a *App) localBranches(ctx context.Context, root string) ([]string, error) {
	out, err := a.runBinary(ctx, "git", []string{"for-each-ref", "--format=%(refname:short)", "refs/heads"}, root)
	if err != nil {
		return nil, fmt.Errorf("list local branches: %w", err)
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

func (a *App) maxPlanningSpecIDFromBranches(ctx context.Context, root string, branches []string) (int, error) {
	maxID := 0
	seen := map[string]struct{}{}
	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if branch == "" {
			continue
		}
		if _, ok := seen[branch]; ok {
			continue
		}
		seen[branch] = struct{}{}

		out, err := a.runBinary(ctx, "git", []string{"ls-tree", "-r", "--name-only", "--full-tree", branch, specsDir}, root)
		if err != nil {
			return 0, fmt.Errorf("inspect specs on branch %s: %w", branch, err)
		}
		branchMaxID := maxPlanningSpecIDFromTreePaths(out)
		if branchMaxID > maxID {
			maxID = branchMaxID
		}
	}
	return maxID, nil
}

func maxPlanningSpecIDFromTreePaths(out string) int {
	maxID := 0
	for _, line := range strings.Split(out, "\n") {
		name := planningSpecNameFromTreePath(line)
		if !strings.HasPrefix(name, "SPEC-") {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(name, "SPEC-"))
		if err == nil && n > maxID {
			maxID = n
		}
	}
	return maxID
}

func planningSpecNameFromTreePath(path string) string {
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, specsDir+"/") {
		return ""
	}
	rest := strings.TrimPrefix(path, specsDir+"/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func formatPlanningStartSummary(start planningStartResolution) string {
	lines := []string{
		"Planning start:",
		fmt.Sprintf("- SPEC: %s", start.SpecID),
		fmt.Sprintf("- Workspace action: %s", start.WorkspaceAction),
		fmt.Sprintf("- Branch: %s", firstNonBlank(start.Branch, "n/a")),
		fmt.Sprintf("- Workspace: %s", firstNonBlank(start.WorkspacePath, start.Root)),
		fmt.Sprintf("- Next step: %s", start.NextStep),
	}
	return strings.Join(lines, "\n") + "\n"
}

func wrapPlanningScaffoldFailure(start planningStartResolution, err error) error {
	if !start.CreatedBranch {
		return err
	}
	return fmt.Errorf("%sscaffolding failed after creating the SPEC branch; continue in the current workspace on %s or delete the branch manually if you do not want to keep it: %w", formatPlanningStartSummary(start), start.Branch, err)
}
