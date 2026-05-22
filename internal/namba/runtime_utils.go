package namba

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

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

func isGitRepository(root string) bool {
	return exists(filepath.Join(root, ".git"))
}
