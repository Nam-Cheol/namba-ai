package namba

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
)

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
