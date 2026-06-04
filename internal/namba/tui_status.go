package namba

import (
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
)

type commandTUIState struct {
	Command string
	Title   string
	Detail  string
	Status  string
}

func (a *App) printCommandTUIState(state commandTUIState) bool {
	if !a.isInteractiveTerminal() {
		return false
	}
	printCommandTUIState(a.stdout, state)
	return true
}

func printCommandTUIState(out io.Writer, state commandTUIState) {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00A7A5")).Render(state.Title)
	status := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(commandTUIStatusColor(state.Status))).Render(strings.ToUpper(state.Status))
	detail := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C737F")).Render(state.Detail)
	body := strings.Join([]string{
		fmt.Sprintf("%s  %s", title, status),
		detail,
	}, "\n")
	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#5C6670")).
		Padding(0, 1).
		Render(body)
	fmt.Fprintln(out, box)
}

func commandTUIStatusColor(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "done", "ready", "updated":
		return "#1D7F45"
	case "warning":
		return "#A05A00"
	case "error":
		return "#B00020"
	default:
		return "#3A5FCD"
	}
}
