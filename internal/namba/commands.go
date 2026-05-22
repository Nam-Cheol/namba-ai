package namba

import (
	"context"
	"fmt"
	"strings"
)

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
		{Name: "report", UsageSummary: "  namba report [--format markdown|text|json] [--json]", UsageText: reportUsageText, Run: (*App).runReport},
		{Name: "project", UsageSummary: "  namba project", UsageText: projectUsageText, Run: (*App).runProject},
		{Name: "update", UsageSummary: "  namba update [--version vX.Y.Z]", UsageText: updateUsageText, Run: (*App).runUpdate},
		{Name: "regen", UsageSummary: "  namba regen", UsageText: regenUsageText, Run: (*App).runRegen},
		{Name: "codex", UsageSummary: "  namba codex access [--approval-policy POLICY --sandbox-mode MODE]", UsageText: codexUsageText, Run: (*App).runCodex},
		{Name: "plan", UsageSummary: "  namba plan \"<description>\"", UsageText: planUsageText, Run: (*App).runPlan},
		{Name: "harness", UsageSummary: "  namba harness \"<description>\"", UsageText: harnessUsageText, Run: (*App).runHarness},
		{Name: "fix", UsageSummary: "  namba fix [--command run|plan] \"<issue description>\"", UsageText: fixUsageText, Run: (*App).runFix},
		{Name: "eval", UsageSummary: "  namba eval [--suite harness] [--format markdown|json]", UsageText: evalUsageText, Run: (*App).runEval},
		{Name: "run", UsageSummary: "  namba run SPEC-XXX [--solo|--team|--parallel] [--dry-run]", UsageText: runUsageText, Run: (*App).runExecute},
		{Name: "queue", UsageSummary: "  namba queue <start|status|resume|pause|stop>", UsageText: queueUsageText, Run: (*App).runQueue},
		{Name: "sync", UsageSummary: "  namba sync", UsageText: syncUsageText, Run: (*App).runSync},
		{Name: "pr", UsageSummary: "  namba pr \"<title>\" [--review] [--language en|ko|ja|zh] [--remote origin] [--no-sync] [--no-validate]", UsageText: prUsageText, Run: (*App).runPR},
		{Name: "land", UsageSummary: "  namba land [PR_NUMBER] [--wait] [--remote origin]", UsageText: landUsageText, Run: (*App).runLand},
		{Name: "release", UsageSummary: "  namba release [--bump patch|minor|major] [--version vX.Y.Z] [--language en|ko|ja|zh] [--push] [--remote origin]", UsageText: releaseUsageText, Run: (*App).runRelease},
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
