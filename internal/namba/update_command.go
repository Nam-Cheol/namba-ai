package namba

import (
	"context"
	"errors"
	"fmt"
)

func (a *App) runUpdate(_ context.Context, args []string) error {
	if len(args) != 0 {
		return errors.New("update does not accept arguments")
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
	if err := a.writeOutputs(root, outputs); err != nil {
		return err
	}

	fmt.Fprintln(a.stdout, "Refreshed NambaAI AGENTS, repo skills, Codex agents, and Codex config.")
	return nil
}
