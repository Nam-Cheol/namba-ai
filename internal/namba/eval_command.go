package namba

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type evalOptions struct {
	suite            string
	format           string
	fixture          string
	baseline         string
	failOnRegression bool
	updateBaseline   bool
	caseID           string
	help             bool
}

func parseEvalOptions(args []string) (evalOptions, error) {
	options := evalOptions{
		suite:    defaultEvalSuite,
		format:   defaultEvalFormat,
		fixture:  defaultHarnessEvalFixture,
		baseline: defaultHarnessEvalBase,
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			options.help = true
		case "--suite":
			i++
			if i >= len(args) {
				return evalOptions{}, errors.New("--suite requires a value")
			}
			options.suite = strings.TrimSpace(args[i])
		case "--format":
			i++
			if i >= len(args) {
				return evalOptions{}, errors.New("--format requires a value")
			}
			options.format = strings.TrimSpace(args[i])
		case "--fixture":
			i++
			if i >= len(args) {
				return evalOptions{}, errors.New("--fixture requires a path")
			}
			options.fixture = strings.TrimSpace(args[i])
		case "--baseline":
			i++
			if i >= len(args) {
				return evalOptions{}, errors.New("--baseline requires a path")
			}
			options.baseline = strings.TrimSpace(args[i])
		case "--case":
			i++
			if i >= len(args) {
				return evalOptions{}, errors.New("--case requires an id")
			}
			options.caseID = strings.TrimSpace(args[i])
		case "--fail-on-regression":
			options.failOnRegression = true
		case "--update-baseline":
			options.updateBaseline = true
		default:
			return evalOptions{}, fmt.Errorf("unknown flag %q", arg)
		}
	}
	if options.suite == "" {
		return evalOptions{}, errors.New("--suite cannot be empty")
	}
	if options.format == "" {
		return evalOptions{}, errors.New("--format cannot be empty")
	}
	return options, nil
}

func (a *App) runEval(ctx context.Context, args []string) error {
	options, err := parseEvalOptions(args)
	if err != nil {
		return commandUsageError("eval", err)
	}
	if options.help {
		return a.printCommandUsage("eval")
	}
	if err := validateEvalFormat(options.format); err != nil {
		return commandExitError(2, err)
	}
	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}
	result, err := a.runEvalSuite(ctx, root, options)
	if err != nil {
		return commandExitError(2, err)
	}
	var output string
	switch options.format {
	case "json":
		output, err = renderEvalJSON(result)
	case "markdown":
		output = renderEvalMarkdown(result)
	default:
		return commandExitError(2, fmt.Errorf("unsupported eval format %q", options.format))
	}
	if err != nil {
		return commandExitError(2, err)
	}
	fmt.Fprint(a.stdout, output)
	if result.Summary.Failed > 0 {
		return commandExitError(1, errors.New("namba eval scenarios failed"))
	}
	if options.failOnRegression && len(result.Regressions) > 0 {
		return commandExitError(1, errors.New("namba eval baseline regression detected"))
	}
	return nil
}

func validateEvalFormat(format string) error {
	switch format {
	case "json", "markdown":
		return nil
	default:
		return fmt.Errorf("unsupported eval format %q", format)
	}
}

func evalUsageText() string {
	return strings.Join([]string{
		"namba eval",
		"",
		"Usage:",
		"  namba eval [--suite harness] [--format markdown|json] [--fixture PATH] [--baseline PATH] [--fail-on-regression] [--update-baseline] [--case ID]",
		"",
		"Behavior:",
		"  Runs deterministic, local Namba harness-quality scenarios without live Codex, network, GitHub API, browser, telemetry, or LLM judging.",
	}, "\n") + "\n"
}
