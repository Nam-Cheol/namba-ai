package namba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
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
	scorecardOut     string
	summaryOut       string
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
		case "--scorecard-out":
			i++
			if i >= len(args) {
				return evalOptions{}, errors.New("--scorecard-out requires a path")
			}
			options.scorecardOut = strings.TrimSpace(args[i])
		case "--summary-out":
			i++
			if i >= len(args) {
				return evalOptions{}, errors.New("--summary-out requires a path")
			}
			options.summaryOut = strings.TrimSpace(args[i])
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
	if options.scorecardOut == "" && options.summaryOut == "" {
		return options, nil
	}
	if strings.TrimSpace(options.caseID) != "" {
		return evalOptions{}, errors.New("--scorecard-out and --summary-out require the full suite, not --case")
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
	if err := a.writeEvalArtifacts(root, result, options); err != nil {
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

func (a *App) writeEvalArtifacts(root string, result evalRunResult, options evalOptions) error {
	if options.scorecardOut != "" {
		if result.Scorecard == nil {
			return errors.New("eval scorecard was not generated")
		}
		data, err := json.MarshalIndent(result.Scorecard, "", "  ")
		if err != nil {
			return fmt.Errorf("render eval scorecard: %w", err)
		}
		if err := a.writeEvalArtifactFile(root, options.scorecardOut, append(data, '\n')); err != nil {
			return err
		}
	}
	if options.summaryOut != "" {
		if err := a.writeEvalArtifactFile(root, options.summaryOut, []byte(renderEvalScorecardMarkdown(result)+"\n")); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) writeEvalArtifactFile(root, path string, data []byte) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("eval artifact path cannot be empty")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	if err := a.mkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return a.writeFile(path, data, 0o644)
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
		"  namba eval --scorecard-out PATH --summary-out PATH",
		"",
		"Behavior:",
		"  Runs deterministic, local Namba harness-quality scenarios without live Codex, network, GitHub API, browser, telemetry, or LLM judging.",
	}, "\n") + "\n"
}
