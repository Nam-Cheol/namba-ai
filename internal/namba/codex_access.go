package namba

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

const codexAccessPresetCustom = "custom"

type codexSubcommandDefinition struct {
	Name         string
	UsageSummary string
	UsageText    func() string
	Run          func(*App, context.Context, []string) error
}

type codexAccessPreset struct {
	ID                string
	Label             string
	WizardDescription string
	Consequence       string
	ApprovalPolicy    string
	SandboxMode       string
}

type codexAccessChoice struct {
	PresetID       string
	Label          string
	Consequence    string
	ApprovalPolicy string
	SandboxMode    string
}

type codexAccessArgs struct {
	ApprovalPolicy string
	SandboxMode    string
}

func codexSubcommandDefinitions() []codexSubcommandDefinition {
	return []codexSubcommandDefinition{
		{
			Name:         "access",
			UsageSummary: "  namba codex access [--approval-policy POLICY --sandbox-mode MODE]",
			UsageText:    codexAccessUsageText,
			Run:          (*App).runCodexAccessSubcommand,
		},
	}
}

func (a *App) resolveCodexSubcommand(name string) (codexSubcommandDefinition, bool) {
	for _, definition := range codexSubcommandDefinitions() {
		if definition.Name == name {
			return definition, true
		}
	}
	return codexSubcommandDefinition{}, false
}

func codexSubcommandUsageSummaries() []string {
	lines := make([]string, 0, len(codexSubcommandDefinitions()))
	for _, definition := range codexSubcommandDefinitions() {
		lines = append(lines, definition.UsageSummary)
	}
	return lines
}

func codexUsageText() string {
	lines := []string{
		"namba codex",
		"",
		"Usage:",
	}
	lines = append(lines, codexSubcommandUsageSummaries()...)
	lines = append(lines,
		"",
		"Behavior:",
		"  Inspect or update repo-owned Codex access defaults from the project root.",
	)
	return strings.Join(lines, "\n") + "\n"
}

func codexAccessUsageText() string {
	lines := []string{
		"namba codex access",
		"",
		"Usage:",
		"  namba codex access",
		"  namba codex access --approval-policy POLICY --sandbox-mode MODE",
		"",
		"Behavior:",
		"  Inspect the current repo-owned Codex access preset and effective approval_policy / sandbox_mode without mutating by default.",
		"  Apply a change only when both --approval-policy and --sandbox-mode are provided.",
		"  `namba init` and `namba codex access` share the same preset labels, consequence statements, and raw-value preview.",
	}
	return strings.Join(lines, "\n") + "\n"
}

func codexAccessUsageError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s\n\n%s", err.Error(), codexAccessUsageText())
}

func (a *App) runCodex(ctx context.Context, args []string) error {
	if wantsCommandHelp(args) {
		return a.printCommandUsage("codex")
	}
	if len(args) == 0 {
		return commandUsageError("codex", errors.New("codex requires a subcommand"))
	}

	subcommand, ok := a.resolveCodexSubcommand(args[0])
	if !ok {
		return commandUsageError("codex", fmt.Errorf("unknown codex subcommand %q", args[0]))
	}
	return subcommand.Run(a, ctx, args[1:])
}

func (a *App) runCodexAccessSubcommand(_ context.Context, args []string) error {
	if wantsCommandHelp(args) {
		_, err := fmt.Fprint(a.stdout, codexAccessUsageText())
		return err
	}

	options, err := parseCodexAccessArgs(args)
	if err != nil {
		return codexAccessUsageError(err)
	}

	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}

	profile, err := a.loadInitProfileFromConfig(root)
	if err != nil {
		return fmt.Errorf("current managed config is invalid: %w", err)
	}

	current, err := resolveCodexAccessChoice(approvalPolicy(profile), sandboxMode(profile))
	if err != nil {
		return fmt.Errorf("current managed config is invalid: %w", err)
	}

	if strings.TrimSpace(options.ApprovalPolicy) == "" && strings.TrimSpace(options.SandboxMode) == "" {
		writeCodexAccessSummary(a.stdout, "Current Codex access", current)
		return nil
	}

	desiredProfile := profile
	desiredProfile.ApprovalPolicy = options.ApprovalPolicy
	desiredProfile.SandboxMode = options.SandboxMode
	if err := validateInitProfile(desiredProfile); err != nil {
		return codexAccessUsageError(err)
	}

	desired, err := resolveCodexAccessChoice(approvalPolicy(desiredProfile), sandboxMode(desiredProfile))
	if err != nil {
		return codexAccessUsageError(err)
	}

	if current.ApprovalPolicy == desired.ApprovalPolicy && current.SandboxMode == desired.SandboxMode {
		writeCodexAccessSummary(a.stdout, "Current Codex access", current)
		fmt.Fprintln(a.stdout, "No change: requested access already matches the current repo defaults.")
		return nil
	}

	outputs := map[string]string{
		filepath.ToSlash(filepath.Join(configDir, "system.yaml")): renderSystemConfig(desiredProfile),
		filepath.ToSlash(repoCodexConfigPath):                     renderRepoCodexConfig(desiredProfile),
	}
	report, err := a.writeOutputs(root, outputs)
	if err != nil {
		return err
	}

	writeCodexAccessSummary(a.stdout, "Previous Codex access", current)
	fmt.Fprintln(a.stdout, "Updated Codex access defaults.")
	writeCodexAccessSummary(a.stdout, "New Codex access", desired)
	if len(report.InstructionSurfacePaths) > 0 {
		fmt.Fprintf(a.stdout, "Session refresh required: start a fresh Codex session before continuing long team or repair runs (%s)\n", strings.Join(report.InstructionSurfacePaths, ", "))
	}
	return nil
}

func parseCodexAccessArgs(args []string) (codexAccessArgs, error) {
	opts := codexAccessArgs{}

	consumeValue := func(args []string, index *int, flag string) (string, error) {
		*index++
		if *index >= len(args) {
			return "", fmt.Errorf("%s requires a value", flag)
		}
		return args[*index], nil
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--approval-policy":
			value, err := consumeValue(args, &i, args[i])
			if err != nil {
				return codexAccessArgs{}, err
			}
			opts.ApprovalPolicy = value
		case "--sandbox-mode":
			value, err := consumeValue(args, &i, args[i])
			if err != nil {
				return codexAccessArgs{}, err
			}
			opts.SandboxMode = value
		default:
			return codexAccessArgs{}, fmt.Errorf("unknown flag %q", args[i])
		}
	}

	hasApproval := strings.TrimSpace(opts.ApprovalPolicy) != ""
	hasSandbox := strings.TrimSpace(opts.SandboxMode) != ""
	if hasApproval != hasSandbox {
		return codexAccessArgs{}, errors.New("pass both --approval-policy and --sandbox-mode to apply a change, or run `namba codex access` without flags to inspect current settings")
	}
	return opts, nil
}

func codexAccessPresets() []codexAccessPreset {
	return []codexAccessPreset{
		{
			ID:                "cautious-read-only",
			Label:             "Cautious read-only",
			WizardDescription: "승인을 묻고 파일 쓰기는 막습니다",
			Consequence:       "Codex asks before risky actions and cannot write files.",
			ApprovalPolicy:    "on-request",
			SandboxMode:       "read-only",
		},
		{
			ID:                "balanced",
			Label:             "Balanced workspace",
			WizardDescription: "필요할 때만 승인하고 저장소 파일은 수정할 수 있습니다",
			Consequence:       "Codex asks when needed and can edit files in the repo workspace.",
			ApprovalPolicy:    "on-request",
			SandboxMode:       "workspace-write",
		},
		{
			ID:                "full-access",
			Label:             "Full access",
			WizardDescription: "승인 없이 진행하고 전체 파일시스템에 접근합니다",
			Consequence:       "Codex keeps moving without approval prompts and can access the full filesystem.",
			ApprovalPolicy:    "never",
			SandboxMode:       "danger-full-access",
		},
	}
}

func resolveCodexAccessChoice(approvalValue, sandboxValue string) (codexAccessChoice, error) {
	approvalValue = normalizeApprovalPolicy(approvalValue)
	sandboxValue = normalizeSandboxMode(sandboxValue)
	if err := validateCodexAccessPair(approvalValue, sandboxValue); err != nil {
		return codexAccessChoice{}, err
	}

	for _, preset := range codexAccessPresets() {
		if preset.ApprovalPolicy == approvalValue && preset.SandboxMode == sandboxValue {
			return codexAccessChoice{
				PresetID:       preset.ID,
				Label:          preset.Label,
				Consequence:    preset.Consequence,
				ApprovalPolicy: preset.ApprovalPolicy,
				SandboxMode:    preset.SandboxMode,
			}, nil
		}
	}

	return codexAccessChoice{
		PresetID:       codexAccessPresetCustom,
		Label:          "Custom access",
		Consequence:    buildCustomCodexAccessConsequence(approvalValue, sandboxValue),
		ApprovalPolicy: approvalValue,
		SandboxMode:    sandboxValue,
	}, nil
}

func validateCodexAccessPair(approvalValue, sandboxValue string) error {
	if !isAllowedApprovalPolicy(normalizeApprovalPolicy(approvalValue)) {
		return fmt.Errorf("approval policy %q is not supported", approvalValue)
	}
	if !isAllowedSandboxMode(normalizeSandboxMode(sandboxValue)) {
		return fmt.Errorf("sandbox mode %q is not supported", sandboxValue)
	}
	return nil
}

func buildCustomCodexAccessConsequence(approvalValue, sandboxValue string) string {
	return fmt.Sprintf("%s %s", codexAccessApprovalConsequence(approvalValue), codexAccessSandboxConsequence(sandboxValue))
}

func codexAccessApprovalConsequence(value string) string {
	switch normalizeApprovalPolicy(value) {
	case "untrusted":
		return "Codex pauses before untrusted operations."
	case "on-failure":
		return "Codex retries failed operations only after approval."
	case "never":
		return "Codex does not pause for approvals."
	default:
		return "Codex asks before escalations when it needs approval."
	}
}

func codexAccessSandboxConsequence(value string) string {
	switch normalizeSandboxMode(value) {
	case "read-only":
		return "Filesystem writes stay blocked."
	case "danger-full-access":
		return "Sandbox restrictions are lifted."
	default:
		return "Repo workspace writes are allowed."
	}
}

func codexAccessPresetOptions() []option {
	options := make([]option, 0, len(codexAccessPresets())+1)
	for _, preset := range codexAccessPresets() {
		options = append(options, option{
			Value:       preset.ID,
			Label:       preset.Label,
			Description: preset.WizardDescription,
		})
	}
	options = append(options, option{
		Value:       codexAccessPresetCustom,
		Label:       "Custom access",
		Description: "approval_policy와 sandbox_mode를 직접 조합합니다",
	})
	return options
}

func defaultCodexAccessPreset(profile initProfile) string {
	choice, err := resolveCodexAccessChoice(approvalPolicy(profile), sandboxMode(profile))
	if err != nil {
		return "balanced"
	}
	return choice.PresetID
}

func applyCodexAccessPreset(profile *initProfile, presetID string) error {
	if strings.TrimSpace(presetID) == codexAccessPresetCustom {
		return nil
	}
	for _, preset := range codexAccessPresets() {
		if preset.ID == presetID {
			profile.ApprovalPolicy = preset.ApprovalPolicy
			profile.SandboxMode = preset.SandboxMode
			return nil
		}
	}
	return fmt.Errorf("codex access preset %q is not supported", presetID)
}

func (a *App) promptCodexAccess(reader *bufio.Reader, profile initProfile) (initProfile, error) {
	presetID := promptSelect(
		a.stdin,
		a.stdout,
		"\U0001f510 Codex access preset",
		codexAccessPresetOptions(),
		defaultCodexAccessPreset(profile),
	)

	if presetID == codexAccessPresetCustom {
		profile.ApprovalPolicy = promptSelect(
			a.stdin,
			a.stdout,
			"\u2705 approval_policy",
			approvalPolicyOptions(),
			approvalPolicy(profile),
		)
		profile.SandboxMode = promptSelect(
			a.stdin,
			a.stdout,
			"\U0001f512 sandbox_mode",
			sandboxModeOptions(),
			sandboxMode(profile),
		)
	} else if err := applyCodexAccessPreset(&profile, presetID); err != nil {
		return initProfile{}, err
	}

	choice, err := resolveCodexAccessChoice(approvalPolicy(profile), sandboxMode(profile))
	if err != nil {
		return initProfile{}, err
	}
	writeWizardCodexAccessPreview(a.stdout, choice)
	return profile, nil
}

func writeWizardCodexAccessPreview(out io.Writer, choice codexAccessChoice) {
	fmt.Fprintln(out, wizardHint(out, fmt.Sprintf("-> %s", choice.Label)))
	fmt.Fprintln(out, wizardHint(out, "   "+choice.Consequence))
	fmt.Fprintln(out, wizardHint(out, fmt.Sprintf("   approval_policy=%s", choice.ApprovalPolicy)))
	fmt.Fprintln(out, wizardHint(out, fmt.Sprintf("   sandbox_mode=%s", choice.SandboxMode)))
}

func writeCodexAccessSummary(out io.Writer, heading string, choice codexAccessChoice) {
	if strings.TrimSpace(heading) != "" {
		fmt.Fprintln(out, heading)
	}
	fmt.Fprintf(out, "Codex access preset: %s\n", choice.Label)
	fmt.Fprintf(out, "Consequence: %s\n", choice.Consequence)
	fmt.Fprintf(out, "approval_policy: %s\n", choice.ApprovalPolicy)
	fmt.Fprintf(out, "sandbox_mode: %s\n", choice.SandboxMode)
}
