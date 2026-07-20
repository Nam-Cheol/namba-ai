package namba

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type runExecuteOptions struct {
	specID string
	mode   executionMode
	dryRun bool
}

type executionRuntimeConfig struct {
	QualityCfg qualityConfig
	SystemCfg  systemConfig
	CodexCfg   codexConfig
}

type directFixExecutionContext struct {
	Root        string
	Description string
	QualityCfg  qualityConfig
	SystemCfg   systemConfig
	CodexCfg    codexConfig
	Prompt      string
	PromptPath  string
	LogID       string
	Delegation  delegationPlan
}

type runExecutionContext struct {
	Root              string
	SpecPkg           specPackage
	ReadinessAdvisory string
	QualityCfg        qualityConfig
	SystemCfg         systemConfig
	CodexCfg          codexConfig
	WorkflowCfg       workflowConfig
	Prompt            string
	PromptPath        string
	Tasks             []string
	Delegation        delegationPlan
}

func parseRunExecuteOptions(args []string) (runExecuteOptions, error) {
	if len(args) == 0 {
		return runExecuteOptions{}, errors.New("run requires a SPEC id")
	}

	options := runExecuteOptions{specID: args[0], mode: executionModeDefault}
	var solo bool
	var team bool
	var parallel bool
	for _, arg := range args[1:] {
		switch arg {
		case "--solo":
			solo = true
		case "--team":
			team = true
		case "--parallel":
			parallel = true
		case "--dry-run":
			options.dryRun = true
		default:
			return runExecuteOptions{}, fmt.Errorf("unknown flag %q", arg)
		}
	}

	selectedModes := make([]string, 0, 3)
	if solo {
		selectedModes = append(selectedModes, "--solo")
	}
	if team {
		selectedModes = append(selectedModes, "--team")
	}
	if parallel {
		selectedModes = append(selectedModes, "--parallel")
	}
	if len(selectedModes) > 1 {
		return runExecuteOptions{}, fmt.Errorf("invalid flag combination: choose only one of --solo, --team, or --parallel (got %s)", strings.Join(selectedModes, ", "))
	}

	switch {
	case solo:
		options.mode = executionModeSolo
	case team:
		options.mode = executionModeTeam
	case parallel:
		options.mode = executionModeParallel
	}

	return options, nil
}

func (a *App) runExecute(ctx context.Context, args []string) error {
	if wantsCommandHelp(args) {
		return a.printCommandUsage("run")
	}
	options, err := parseRunExecuteOptions(args)
	if err != nil {
		return commandUsageError("run", err)
	}

	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}

	runCtx, err := a.loadRunExecutionContext(root, options)
	if err != nil {
		return err
	}
	if err := a.materializeRunExecutionPrompt(runCtx); err != nil {
		return err
	}
	return a.dispatchRunExecution(ctx, options, runCtx)
}

func (a *App) loadRunExecutionContext(root string, options runExecuteOptions) (runExecutionContext, error) {
	specPkg, err := a.loadSpec(root, options.specID)
	if err != nil {
		return runExecutionContext{}, err
	}
	readinessAdvisory, err := a.refreshSpecReviewReadiness(root, specPkg.ID)
	if err != nil {
		return runExecutionContext{}, err
	}
	if frontend := loadFrontendBriefReport(root, specPkg.ID); frontend.Exists {
		if !frontend.Valid {
			if frontendInvalidContractBlocksExecution(frontend) {
				return runExecutionContext{}, frontendGateExecutionError(specPkg.ID, frontend)
			}
		} else if frontend.Header.TaskClassification == frontendTaskClassificationMajor {
			frontendReady := frontend.Header.FrontendGateStatus == frontendGateStatusApproved && frontend.EvidenceStatus == frontendEvidenceStatusComplete && frontend.NegativeContractStatus == frontendNegativeContractStatusComplete && len(frontend.Mismatches) == 0
			if !frontendReady {
				return runExecutionContext{}, frontendGateExecutionError(specPkg.ID, frontend)
			}
		}
	}
	runtimeCfg, err := a.loadExecutionRuntimeConfig(root)
	if err != nil {
		return runExecutionContext{}, err
	}
	workflowCfg, err := a.loadWorkflowConfig(root)
	if err != nil {
		return runExecutionContext{}, err
	}
	prompt, tasks, delegation, err := a.buildExecutionPrompt(root, specPkg, runtimeCfg.QualityCfg, options.mode)
	if err != nil {
		return runExecutionContext{}, err
	}

	return runExecutionContext{
		Root:              root,
		SpecPkg:           specPkg,
		ReadinessAdvisory: readinessAdvisory,
		QualityCfg:        runtimeCfg.QualityCfg,
		SystemCfg:         runtimeCfg.SystemCfg,
		CodexCfg:          runtimeCfg.CodexCfg,
		WorkflowCfg:       workflowCfg,
		Prompt:            prompt,
		PromptPath:        filepath.Join(root, logsDir, "runs", strings.ToLower(specPkg.ID)+"-request.md"),
		Tasks:             tasks,
		Delegation:        delegation,
	}, nil
}

func (a *App) materializeRunExecutionPrompt(runCtx runExecutionContext) error {
	return a.writeExecutionPrompt(runCtx.PromptPath, runCtx.Prompt)
}

func (a *App) loadExecutionRuntimeConfig(root string) (executionRuntimeConfig, error) {
	qualityCfg, err := a.loadQualityConfig(root)
	if err != nil {
		return executionRuntimeConfig{}, err
	}
	systemCfg, err := a.loadSystemConfig(root)
	if err != nil {
		return executionRuntimeConfig{}, err
	}
	codexCfg, err := a.loadCodexConfig(root)
	if err != nil {
		return executionRuntimeConfig{}, err
	}
	return executionRuntimeConfig{
		QualityCfg: qualityCfg,
		SystemCfg:  systemCfg,
		CodexCfg:   codexCfg,
	}, nil
}

func (a *App) writeExecutionPrompt(path, prompt string) error {
	if err := a.mkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return a.writeFile(path, []byte(prompt), 0o644)
}

func (a *App) dispatchRunExecution(ctx context.Context, options runExecuteOptions, runCtx runExecutionContext) error {
	if runCtx.ReadinessAdvisory != "" {
		fmt.Fprintf(a.stdout, "Review readiness for %s: %s (advisory only)\n", runCtx.SpecPkg.ID, runCtx.ReadinessAdvisory)
	}

	if options.mode == executionModeParallel {
		return a.runParallel(ctx, runCtx.Root, runCtx.SpecPkg, runCtx.Tasks, runCtx.Prompt, runCtx.QualityCfg, runCtx.SystemCfg, runCtx.CodexCfg, runCtx.WorkflowCfg, options.dryRun)
	}

	if options.dryRun {
		request := a.newExecutionRequest(runCtx.SpecPkg.ID, runCtx.Root, runCtx.Prompt, options.mode, runCtx.Delegation, runCtx.SystemCfg, runCtx.CodexCfg)
		fmt.Fprintf(a.stdout, "Model routing plan for %s (planned)\n", runCtx.SpecPkg.ID)
		for _, turn := range plannedExecutionTurnRequests(request) {
			decision := turn.RoutingDecision
			sessionStrategy := "fresh_session"
			if turn.ResumeSession {
				sessionStrategy = "explicit_thread_resume"
			}
			fmt.Fprintf(a.stdout, "- phase=%s role=%s tier=%s model=%s effort=%s rule=%s reasons=%s session=%s sol_remaining=%d state=%s\n",
				decision.Phase, firstNonBlank(turn.TurnRole, "integrator"), decision.Tier, turn.Model, turn.RequestedReasoningEffort, decision.RuleID, strings.Join(decision.ReasonCodes, ","), sessionStrategy, decision.RemainingSolTurns, decision.Status)
		}
		fmt.Fprintf(a.stdout, "Prepared execution request at %s\n", runCtx.PromptPath)
		return nil
	}

	request := a.newExecutionRequest(runCtx.SpecPkg.ID, runCtx.Root, runCtx.Prompt, options.mode, runCtx.Delegation, runCtx.SystemCfg, runCtx.CodexCfg)
	if _, _, err := a.executeRun(ctx, runCtx.Root, strings.ToLower(runCtx.SpecPkg.ID), request, runCtx.Root, runCtx.QualityCfg, nil, ""); err != nil {
		return err
	}

	fmt.Fprintf(a.stdout, "Executed %s with %s\n", runCtx.SpecPkg.ID, request.Runner)
	return nil
}

func (a *App) buildExecutionPrompt(root string, specPkg specPackage, qualityCfg qualityConfig, mode executionMode) (string, []string, delegationPlan, error) {
	specBytes, err := os.ReadFile(filepath.Join(specPkg.Path, "spec.md"))
	if err != nil {
		return "", nil, delegationPlan{}, err
	}
	planBytes, err := os.ReadFile(filepath.Join(specPkg.Path, "plan.md"))
	if err != nil {
		return "", nil, delegationPlan{}, err
	}
	acceptanceBytes, err := os.ReadFile(filepath.Join(specPkg.Path, "acceptance.md"))
	if err != nil {
		return "", nil, delegationPlan{}, err
	}

	tasks := extractAcceptanceTasks(string(acceptanceBytes))
	mode = normalizeExecutionMode(mode)
	modeGuidance := executionModePromptGuidance(mode)
	delegation := suggestDelegationPlan(mode, string(specBytes), string(planBytes), string(acceptanceBytes))
	promptLines := []string{
		"# NambaAI Execution Request",
		"",
		"Execute this SPEC package using the repository AGENTS.md and local Codex skills.",
		"",
		"## Run Mode",
		fmt.Sprintf("- Mode: %s", mode),
	}
	promptLines = append(promptLines, modeGuidance...)
	promptLines = append(promptLines, "")
	promptLines = append(promptLines, formatDelegationPlanPrompt(delegation)...)
	promptLines = append(promptLines,
		"",
		"## SPEC",
		string(specBytes),
		"",
		"## Plan",
		string(planBytes),
		"",
		"## Acceptance",
		string(acceptanceBytes),
	)
	if frontendBriefExists := exists(filepath.Join(specPkg.Path, frontendBriefFileName)); frontendBriefExists {
		frontendBytes, err := os.ReadFile(filepath.Join(specPkg.Path, frontendBriefFileName))
		if err != nil {
			return "", nil, delegationPlan{}, err
		}
		promptLines = append(promptLines,
			"",
			"## Frontend Brief",
			string(frontendBytes),
		)
		frontendReport := parseFrontendBrief(string(frontendBytes))
		if frontendReport.Header.TaskClassification == frontendTaskClassificationMajor {
			promptLines = append(promptLines,
				"",
				"## Frontend Major Negative-First Execution Contract",
				"- Implement only within the approved Do-Not Design Contract, visual grammar, and frontend architecture handoff.",
				"- Before first frontend or first major screen coding, read `Frontend implementation phase`, `Asset mode`, `Imagegen requirement`, and `Asset decision proof` from `frontend-brief.md`.",
				"- If `Imagegen requirement: required` is present, use available image-generation capability before coding the final screen, save each generated bitmap under the planned project asset path, and wire it into the intended UI element.",
				"- Treat `generated-images` as any imagegen-created UI asset: hero image, section image, icon, small object image, product thumbnail, cutout, texture, sprite, badge, illustration, or other concrete visual asset.",
				"- Do not satisfy asset-led references with CSS gradients, abstract shapes, empty placeholders, generic SaaS card walls, manually drawn decoration, token-only styling, or brand colors around a generic layout.",
				"- Include a `## Generated Asset Evidence` section when imagegen is required, with `Manifest path` and repeatable `Asset ID` blocks containing `Generated file`, `Saved asset path`, `Prompt summary`, `Intended UI usage`, and `Rendered usage evidence`.",
				"- Before claiming completion, include a `## Do-Not Design Violation Check` section in the runner result.",
				"- The section must cite changed files and screenshot, DOM, or local inspection evidence when available.",
				"- If a banned pattern appears, report `Status: failed`, name the banned pattern, and provide the remediation path.",
				"- If an exception path is used, report `Status: passed` only when the exception evidence cites the contract field that allows it.",
			)
		}
	}
	if specReviewReadinessExists(root, specPkg.ID) {
		readinessBytes, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(specReviewReadinessPath(specPkg.ID))))
		if err != nil {
			return "", nil, delegationPlan{}, err
		}
		promptLines = append(promptLines,
			"",
			"## Review Readiness",
			string(readinessBytes),
		)
	}
	promptLines = append(promptLines, "", "## Validation")
	for _, step := range validationPipelineSteps(qualityCfg) {
		promptLines = append(promptLines, fmt.Sprintf("- %s: %s", step.Name, step.Command))
	}
	promptLines = append(promptLines, "", fmt.Sprintf("Project root: %s", root))
	prompt := strings.Join(promptLines, "\n")

	return prompt, tasks, delegation, nil
}

func executionModePromptGuidance(mode executionMode) []string {
	switch normalizeExecutionMode(mode) {
	case executionModeSolo:
		return []string{
			"- Execution style: one runner in one workspace.",
			"- Keep implementation, integration, and validation inside one runner rather than same-workspace team orchestration.",
			"- Do not reinterpret this mode as worktree parallelism.",
		}
	case executionModeTeam:
		return []string{
			"- Execution style: same-workspace multi-agent execution.",
			"- Keep work in one workspace while orchestrating specialist turns and a final reviewer inside one bounded runtime.",
			"- Role runtime profiles should materially affect the actual Codex turns, not only prompt wording.",
			"- Do not reinterpret this mode as worktree parallelism.",
		}
	case executionModeParallel:
		return []string{
			"- Execution style: Namba worktree parallel mode.",
			"- This mode means git worktree fan-out/fan-in managed by Namba, not same-workspace team orchestration.",
			"- Each worker request should stay within its assigned work package and merge only after all workers and validators pass.",
		}
	default:
		return []string{
			"- Execution style: standard standalone Codex run in one workspace.",
			"- Keep work inside the standalone runner unless the user explicitly picks `--team` or `--parallel`.",
			"- Do not reinterpret this mode as worktree parallelism.",
		}
	}
}

type delegationDomainConfig struct {
	Name             string
	PrimaryRole      string
	PlanningRole     string
	Keywords         []string
	PlanningKeywords []string
	ScoreBias        int
}

type delegationDomainMatch struct {
	Config        delegationDomainConfig
	Role          string
	Hits          []string
	Score         int
	WeightedScore int
}

func suggestDelegationPlan(mode executionMode, specText, planText, acceptanceText string) delegationPlan {
	combined := strings.ToLower(strings.Join([]string{specText, planText, acceptanceText}, "\n"))
	matches := make([]delegationDomainMatch, 0)
	for _, cfg := range delegationDomainConfigs() {
		hits := findKeywordHits(combined, cfg.Keywords)
		if len(hits) == 0 {
			continue
		}
		role := cfg.PrimaryRole
		if cfg.PlanningRole != "" {
			planningHits := findKeywordHits(combined, cfg.PlanningKeywords)
			if len(planningHits) > 0 {
				role = cfg.PlanningRole
				hits = uniqueStrings(append(hits, planningHits...))
			}
		}
		matches = append(matches, delegationDomainMatch{
			Config:        cfg,
			Role:          role,
			Hits:          hits,
			Score:         len(hits),
			WeightedScore: len(hits) + cfg.ScoreBias,
		})
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].WeightedScore == matches[j].WeightedScore {
			if matches[i].Score == matches[j].Score {
				return matches[i].Config.Name < matches[j].Config.Name
			}
			return matches[i].Score > matches[j].Score
		}
		return matches[i].WeightedScore > matches[j].WeightedScore
	})

	integratorRole := "standalone-runner"
	switch normalizeExecutionMode(mode) {
	case executionModeTeam:
		integratorRole = "same-workspace-integrator"
	case executionModeParallel:
		integratorRole = "parallel-orchestrator"
	}
	plan := delegationPlan{IntegratorRole: integratorRole}
	for _, match := range matches {
		plan.DominantDomains = append(plan.DominantDomains, match.Config.Name)
	}
	plan.SelectedRoles, plan.DelegationBudget, plan.ReviewerRole, plan.RoutingRationale = chooseDelegatedRoles(mode, matches)
	plan.SelectedRoleProfiles = runtimeProfilesForRoles(plan.SelectedRoles)
	return plan
}

func delegationDomainConfigs() []delegationDomainConfig {
	return []delegationDomainConfig{
		{
			Name:             "frontend",
			PrimaryRole:      "namba-frontend-implementer",
			PlanningRole:     "namba-frontend-architect",
			Keywords:         []string{"frontend", "ui", "component", "screen", "page", "responsive", "browser", "css", "accessibility", "a11y"},
			PlanningKeywords: []string{"component/state split", "plan the component", "plan the state", "component boundary", "component boundaries", "state ownership", "file planning", "delivery planning"},
		},
		{Name: "mobile", PrimaryRole: "namba-mobile-engineer", Keywords: []string{"mobile", "ios", "android", "swift", "kotlin", "react native", "flutter", "tablet", "touch"}, ScoreBias: 2},
		{Name: "backend", PrimaryRole: "namba-backend-implementer", Keywords: []string{"backend", "api", "endpoint", "server", "service", "controller", "handler", "webhook"}, ScoreBias: 1},
		{Name: "data", PrimaryRole: "namba-data-engineer", Keywords: []string{"schema", "migration", "sql", "query", "etl", "warehouse", "analytics", "dataset", "batch", "pipeline"}, ScoreBias: 2},
		{Name: "security", PrimaryRole: "namba-security-engineer", Keywords: []string{"security", "auth", "oauth", "permission", "secret", "token", "encryption", "vulnerability", "compliance", "privacy", "pii"}, ScoreBias: 2},
		{Name: "design", PrimaryRole: "namba-designer", Keywords: []string{"design", "figma", "art direction", "visual direction", "visual design", "palette", "tone logic", "composition", "look and feel", "redesign", "typography", "motion", "prototype", "brand"}, ScoreBias: 1},
		{Name: "devops", PrimaryRole: "namba-devops-engineer", Keywords: []string{"deploy", "deployment", "docker", "kubernetes", "helm", "terraform", "ci", "cd", "infra", "observability", "runtime", "environment"}, ScoreBias: 2},
		{Name: "quality", PrimaryRole: "namba-test-engineer", Keywords: []string{"test", "regression", "coverage", "qa", "e2e", "integration test", "acceptance test"}, ScoreBias: -1},
	}
}

func chooseDelegatedRoles(mode executionMode, matches []delegationDomainMatch) ([]string, int, string, []string) {
	switch normalizeExecutionMode(mode) {
	case executionModeParallel:
		return nil, 0, "", []string{
			"`--parallel` is reserved for Namba worktree fan-out, so do not route to Codex subagents in this mode.",
		}
	case executionModeSolo:
		if len(matches) == 0 || matches[0].Score < 2 {
			return nil, 0, "", []string{
				"No single specialist signal is strong enough, so stay inside one generalist runner.",
			}
		}
		return []string{matches[0].Role}, 1, "", []string{
			fmt.Sprintf("Highest-signal domain is %s via %s.", matches[0].Config.Name, quoteList(matches[0].Hits)),
			"Delegate to one bounded specialist only if it materially reduces risk, and keep integration plus validation in the standalone runner.",
		}
	case executionModeTeam:
		if len(matches) == 0 {
			return []string{"namba-implementer", "namba-reviewer"}, 2, "namba-reviewer", []string{
				"No domain clearly dominates, so keep team mode light with one general implementer plus a reviewer.",
				"Add more specialists only when acceptance criteria span multiple clearly independent domains.",
			}
		}
		maxDomains := 1
		if len(matches) > 1 {
			maxDomains = 2
		}
		roles := make([]string, 0, maxDomains+1)
		rationale := make([]string, 0, maxDomains+2)
		for i := 0; i < len(matches) && i < maxDomains; i++ {
			roles = append(roles, matches[i].Role)
			rationale = append(rationale, fmt.Sprintf("%s matched %s.", matches[i].Config.Name, quoteList(matches[i].Hits)))
		}
		if len(matches) > 2 {
			rationale = append(rationale, "More than two domains matched, but team mode stays light by using only the top two specialists before review.")
		} else if len(matches) == 1 {
			rationale = append(rationale, "One domain dominates, so start with one specialist and a reviewer rather than a larger swarm.")
		} else {
			rationale = append(rationale, "Multiple domains matched, so use one specialist per dominant domain before the final review pass.")
		}
		rationale = append(rationale, "Keep the standalone runner as the integrator and final validation owner.")
		roles = append(roles, "namba-reviewer")
		roles = uniqueStrings(roles)
		return roles, len(roles), "namba-reviewer", rationale
	default:
		return nil, 0, "", []string{
			"Default mode keeps work inside the standalone runner unless the user explicitly asks for specialist delegation.",
		}
	}
}

func formatDelegationPlanPrompt(plan delegationPlan) []string {
	lines := []string{"## Delegation Heuristics"}
	if len(plan.DominantDomains) == 0 {
		lines = append(lines, "- Dominant domains: none detected beyond general implementation.")
	} else {
		lines = append(lines, fmt.Sprintf("- Dominant domains: %s.", strings.Join(plan.DominantDomains, ", ")))
	}
	if len(plan.SelectedRoles) == 0 {
		lines = append(lines, "- Suggested roles: keep work inside the standalone runner without spawning specialists.")
	} else {
		lines = append(lines, fmt.Sprintf("- Suggested roles: %s.", quoteList(plan.SelectedRoles)))
	}
	for _, profile := range plan.SelectedRoleProfiles {
		if summary := formatAgentRuntimeProfile(profile); summary != "" {
			lines = append(lines, "- Role runtime: "+summary+".")
		}
	}
	lines = append(lines, fmt.Sprintf("- Delegation budget: %d.", plan.DelegationBudget))
	if plan.IntegratorRole != "" {
		lines = append(lines, fmt.Sprintf("- Integrator: `%s`.", plan.IntegratorRole))
	}
	if plan.ReviewerRole != "" {
		lines = append(lines, fmt.Sprintf("- Reviewer: `%s`.", plan.ReviewerRole))
	}
	for _, reason := range plan.RoutingRationale {
		lines = append(lines, "- "+reason)
	}
	return lines
}

func findKeywordHits(text string, keywords []string) []string {
	hits := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			hits = append(hits, keyword)
		}
	}
	return uniqueStrings(hits)
}

func quoteList(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	quoted := make([]string, 0, len(values))
	for _, value := range uniqueStrings(values) {
		quoted = append(quoted, fmt.Sprintf("`%s`", value))
	}
	return strings.Join(quoted, ", ")
}
