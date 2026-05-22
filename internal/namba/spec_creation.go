package namba

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (a *App) runPlan(ctx context.Context, args []string) error {
	options, err := parseDescriptionCommandArgs("plan", "description", args)
	if err != nil {
		return commandUsageError("plan", err)
	}
	if options.help {
		return a.printPlanUsage()
	}
	if clarification, ok := a.evaluateSpecCreationClarification("plan", options.description); ok {
		fmt.Fprint(a.stdout, clarification)
		return errors.New("namba plan requires clarification before creating a SPEC")
	}
	return a.createSpecPackage(ctx, "plan", options.description, options.currentWorkspace, !options.noReview)
}

func (a *App) runHarness(ctx context.Context, args []string) error {
	options, err := parseDescriptionCommandArgs("harness", "description", args)
	if err != nil {
		return commandUsageError("harness", err)
	}
	if options.help {
		return a.printHarnessUsage()
	}
	if clarification, ok := a.evaluateSpecCreationClarification("harness", options.description); ok {
		fmt.Fprint(a.stdout, clarification)
		return errors.New("namba harness requires clarification before creating a SPEC")
	}
	return a.createSpecPackage(ctx, "harness", options.description, options.currentWorkspace, false)
}

func (a *App) runFix(ctx context.Context, args []string) error {
	options, err := parseFixArgs(args)
	if err != nil {
		return commandUsageError("fix", err)
	}
	if options.help {
		return a.printFixUsage()
	}

	if options.command == "plan" {
		if clarification, ok := a.evaluateSpecCreationClarification("fix", options.description); ok {
			fmt.Fprint(a.stdout, clarification)
			return errors.New("namba fix --command plan requires clarification before creating a SPEC")
		}
	}

	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}

	subcommand, ok := a.resolveFixSubcommand(options.command)
	if !ok {
		return fmt.Errorf("invalid fix command %q", options.command)
	}
	return subcommand.Run(a, ctx, root, options)
}

func (a *App) createSpecPackage(ctx context.Context, kind, description string, currentWorkspace, autoReview bool) error {
	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}

	start, err := a.resolvePlanningStart(ctx, root, planningStartOptions{
		Kind:             kind,
		Description:      description,
		CurrentWorkspace: currentWorkspace,
	})
	if err != nil {
		if start.SpecID != "" {
			return errors.New(formatPlanningStartSummary(start))
		}
		return err
	}

	scaffoldCtx, err := a.loadResolvedSpecPackageScaffoldContext(start.Root, kind, description, start.SpecID)
	if err != nil {
		return wrapPlanningScaffoldFailure(start, err)
	}
	outputs := buildSpecPackageScaffoldOutputs(scaffoldCtx)
	if err := a.materializeSpecPackageScaffoldOutputs(scaffoldCtx, outputs); err != nil {
		return wrapPlanningScaffoldFailure(start, err)
	}

	fmt.Fprint(a.stdout, formatPlanningStartSummary(start))
	fmt.Fprintf(a.stdout, "Created %s\n", scaffoldCtx.SpecID)
	if kind == "plan" {
		if autoReview {
			fmt.Fprintf(a.stdout, "Auto review: `$namba-plan-review %s`\n", scaffoldCtx.SpecID)
			fmt.Fprintf(a.stdout, "Pass %s to scaffold only.\n", noReviewPlanningFlag)
		} else {
			fmt.Fprintf(a.stdout, "Auto review skipped by %s.\n", noReviewPlanningFlag)
		}
	}
	profile, err := a.loadInitProfileFromConfig(start.Root)
	if err != nil {
		profile = initProfile{}
	}
	fmt.Fprint(a.stdout, formatSpecCreationReport(scaffoldCtx, start, outputs, autoReview, outputContractLanguage(profile)))
	return nil
}

type planInvocation struct {
	help             bool
	currentWorkspace bool
	noReview         bool
	description      string
}

type fixInvocation struct {
	help             bool
	currentWorkspace bool
	command          string
	description      string
}

func parsePlanArgs(args []string) (planInvocation, error) {
	return parseDescriptionCommandArgs("plan", "description", args)
}

func parseDescriptionCommandArgs(command, field string, args []string) (planInvocation, error) {
	if len(args) == 0 {
		return planInvocation{}, fmt.Errorf("%s requires a %s", command, field)
	}
	invocation := planInvocation{}
	descriptionParts := make([]string, 0, len(args))
	afterDelimiter := false
	for _, arg := range args {
		if afterDelimiter {
			descriptionParts = append(descriptionParts, arg)
			continue
		}
		switch arg {
		case "--":
			afterDelimiter = true
		case "--help", "-h":
			return planInvocation{help: true}, nil
		case currentWorkspacePlanningFlag:
			invocation.currentWorkspace = true
		case noReviewPlanningFlag:
			if command != "plan" {
				return planInvocation{}, fmt.Errorf("unknown flag %q", arg)
			}
			invocation.noReview = true
		default:
			if isStandaloneFlagToken(arg) {
				return planInvocation{}, fmt.Errorf("unknown flag %q", arg)
			}
			descriptionParts = append(descriptionParts, arg)
		}
	}
	description := strings.TrimSpace(strings.Join(descriptionParts, " "))
	if description == "" {
		return planInvocation{}, fmt.Errorf("%s requires a %s", command, field)
	}
	invocation.description = description
	return invocation, nil
}

func evaluatePlanClarification(description string) (string, bool) {
	return evaluateSpecCreationClarification("plan", description)
}

func evaluateSpecCreationClarification(command, description string) (string, bool) {
	return evaluateSpecCreationClarificationForLanguage(command, description, fallbackSpecCreationClarificationLanguage(description))
}

func (a *App) evaluateSpecCreationClarification(command, description string) (string, bool) {
	return evaluateSpecCreationClarificationForLanguage(command, description, a.specCreationClarificationLanguage(description))
}

func (a *App) specCreationClarificationLanguage(description string) string {
	root, err := a.requireProjectRoot()
	if err == nil {
		profile, err := a.loadInitProfileFromConfig(root)
		if err == nil {
			if language := strings.TrimSpace(firstNonBlank(profile.ConversationLanguage, profile.DocumentationLanguage, profile.PRLanguage)); language != "" {
				return language
			}
		}
	}
	return fallbackSpecCreationClarificationLanguage(description)
}

func fallbackSpecCreationClarificationLanguage(description string) string {
	if hasKorean(description) {
		return "ko"
	}
	return "en"
}

func evaluateSpecCreationClarificationForLanguage(command, description, language string) (string, bool) {
	normalized := strings.Join(strings.Fields(description), " ")
	if normalized == "" {
		return "", false
	}
	clarificationLanguage := normalizeReadmeLanguage(language)
	lower := strings.ToLower(normalized)
	if planDescriptionHasClarifyingEvidence(lower) {
		return "", false
	}
	if planDescriptionHasPartialClarifyingEvidence(lower) {
		return formatSpecCreationClarificationQuestionsForLanguage(command, normalized, clarificationLanguage), true
	}

	runeCount := len([]rune(normalized))
	vague := containsAnyFolded(lower, []string{
		"만들어줘",
		"만들어 줘",
		"고쳐줘",
		"고쳐 줘",
		"수정해줘",
		"수정해 줘",
		"개선해줘",
		"개선해 줘",
		"처리해줘",
		"처리해 줘",
		"해줘",
		"해 줘",
		"알아서",
		"대충",
		"적당히",
		"좋게",
		"뭔가",
		"버그",
		"문제",
		"create",
		"build",
		"make",
		"implement",
		"bug",
		"problem",
		"issue",
		"fix it",
		"fix this",
		"something",
		"stuff",
		"thing",
	})
	genericKoreanSurface := containsAnyFolded(lower, []string{
		"게시판",
		"대시보드",
		"관리자",
		"페이지",
		"앱",
		"웹",
		"api",
	})
	if hasKorean(normalized) && runeCount < 70 && (vague || genericKoreanSurface) {
		return formatSpecCreationClarificationQuestionsForLanguage(command, normalized, clarificationLanguage), true
	}
	if runeCount < 45 && vague {
		return formatSpecCreationClarificationQuestionsForLanguage(command, normalized, clarificationLanguage), true
	}
	return "", false
}

func planDescriptionHasClarifyingEvidence(lower string) bool {
	for _, group := range planClarificationEvidenceGroups() {
		if !containsAnyFolded(lower, group) {
			return false
		}
	}
	return true
}

func planDescriptionHasPartialClarifyingEvidence(lower string) bool {
	for _, group := range planClarificationEvidenceGroups() {
		if containsAnyFolded(lower, group) {
			return true
		}
	}
	return false
}

func planClarificationEvidenceGroups() [][]string {
	return [][]string{
		{"goal:", "goal -", "목표", "目標", "目标"},
		{"scope:", "scope -", "범위", "範囲", "范围"},
		{"constraints:", "constraint:", "constraints -", "constraint -", "제약", "制約", "制约", "约束"},
		{"acceptance:", "validation:", "acceptance -", "validation -", "완료 기준", "성공 기준", "검증", "完了基準", "成功基準", "検証", "完成标准", "成功标准", "验证", "验收标准"},
	}
}

func containsAnyFolded(value string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(value, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func hasKorean(value string) bool {
	for _, r := range value {
		if r >= '\uAC00' && r <= '\uD7A3' {
			return true
		}
	}
	return false
}

func formatPlanClarificationQuestions(description string, korean bool) string {
	return formatSpecCreationClarificationQuestions("plan", description, korean)
}

func formatSpecCreationClarificationQuestions(command, description string, korean bool) string {
	language := "en"
	if korean {
		language = "ko"
	}
	return formatSpecCreationClarificationQuestionsForLanguage(command, description, language)
}

func formatSpecCreationClarificationQuestionsForLanguage(command, description, language string) string {
	invocation := specCreationInvocation(command)
	switch normalizeReadmeLanguage(language) {
	case "ko":
		return strings.Join([]string{
			"NambaAI clarification gate: SPEC을 만들기 전에 요구가 아직 너무 넓거나 모호합니다.",
			"입력: " + description,
			"",
			fmt.Sprintf("Codex에서는 가능하면 Plan mode 선택 UI로 답변을 먼저 정리한 뒤, 정리된 Goal/Scope/Constraints/Acceptance만 `%s`에 넘기세요.", invocation),
			"",
			"먼저 아래 질문에 답해주세요:",
			"1. 대상 사용자는 누구이고 대상 surface와 핵심 사용 흐름은 무엇인가요?",
			"2. 이번 SPEC에 포함할 범위와 제외할 범위는 무엇인가요?",
			"3. 완료 기준과 검증 방법은 무엇인가요?",
			"",
			"가능하면 다음 형식으로 답해주세요:",
			"Goal: ...",
			"Scope: ...",
			"Constraints: ...",
			"Acceptance: ...",
			"",
		}, "\n")
	case "ja":
		return strings.Join([]string{
			"NambaAI clarification gate: SPEC を作る前に、依頼がまだ広すぎるか曖昧です。",
			"入力: " + description,
			"",
			fmt.Sprintf("Codex では可能なら Plan mode の選択 UI で回答を先に整理し、整理された Goal/Scope/Constraints/Acceptance だけを `%s` に渡してください。", invocation),
			"",
			"まず次の質問に答えてください:",
			"1. この作業の対象 surface は何ですか？例: CLI、Web app、API、特定モジュール。",
			"2. 望むユーザーフローと、今回の範囲外にするものは何ですか？",
			"3. 完了基準と検証方法は何ですか？",
			"",
			"可能なら次の形式で答えてください:",
			"Goal: ...",
			"Scope: ...",
			"Constraints: ...",
			"Acceptance: ...",
			"",
		}, "\n")
	case "zh":
		return strings.Join([]string{
			"NambaAI clarification gate: 在创建 SPEC 前，请先降低需求的模糊度。",
			"输入: " + description,
			"",
			fmt.Sprintf("在 Codex 中，如果可以，请先用 Plan mode 选择 UI 整理回答，然后只把整理后的 Goal/Scope/Constraints/Acceptance 传给 `%s`。", invocation),
			"",
			"请先回答以下问题:",
			"1. 这项变更影响哪个目标 surface？例如 CLI、Web app、API 或特定模块。",
			"2. 期望的用户流程是什么？哪些内容应排除在范围外？",
			"3. 完成标准和验证方法是什么？",
			"",
			"可以的话，请使用以下格式回答:",
			"Goal: ...",
			"Scope: ...",
			"Constraints: ...",
			"Acceptance: ...",
			"",
		}, "\n")
	}
	return strings.Join([]string{
		"NambaAI clarification gate: the request is still too broad or ambiguous to turn into a SPEC.",
		"Input: " + description,
		"",
		fmt.Sprintf("In Codex, use native Plan mode choice UI first when it is available, then pass only the refined Goal/Scope/Constraints/Acceptance output to `%s`.", invocation),
		"",
		"Please answer these questions first:",
		"1. Who is the target user, and what target surface or core user flow should change?",
		"2. What is in scope and out of scope for this SPEC?",
		"3. What acceptance criteria and validation define done?",
		"",
		"Use this shape when possible:",
		"Goal: ...",
		"Scope: ...",
		"Constraints: ...",
		"Acceptance: ...",
		"",
	}, "\n")
}

func specCreationInvocation(command string) string {
	switch command {
	case "harness":
		return "namba harness"
	case "fix":
		return "namba fix --command plan"
	default:
		return "namba plan"
	}
}

func parseFixArgs(args []string) (fixInvocation, error) {
	if len(args) == 0 {
		return fixInvocation{}, errors.New("fix requires an issue description")
	}

	invocation := fixInvocation{command: "run"}
	var descriptionParts []string
	afterDelimiter := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if afterDelimiter {
			descriptionParts = append(descriptionParts, arg)
			continue
		}
		switch arg {
		case "--":
			afterDelimiter = true
		case "--help", "-h":
			return fixInvocation{help: true}, nil
		case currentWorkspacePlanningFlag:
			invocation.currentWorkspace = true
		case "--command":
			if i+1 >= len(args) {
				return fixInvocation{}, errors.New("fix --command requires a value of run or plan")
			}
			value := strings.TrimSpace(args[i+1])
			if value != "run" && value != "plan" {
				return fixInvocation{}, fmt.Errorf("invalid fix --command %q: expected run or plan", value)
			}
			invocation.command = value
			i++
		case "--command=run", "--command=plan":
			invocation.command = strings.TrimPrefix(arg, "--command=")
		default:
			if isStandaloneFlagToken(arg) {
				return fixInvocation{}, fmt.Errorf("unknown flag %q", arg)
			}
			descriptionParts = append(descriptionParts, arg)
		}
	}

	invocation.description = strings.TrimSpace(strings.Join(descriptionParts, " "))
	if invocation.description == "" {
		return fixInvocation{}, errors.New("fix requires an issue description")
	}
	if invocation.currentWorkspace && invocation.command != "plan" {
		return fixInvocation{}, fmt.Errorf("%s is only valid with --command plan", currentWorkspacePlanningFlag)
	}
	return invocation, nil
}

func fixSubcommandDefinitions() []fixSubcommandDefinition {
	return []fixSubcommandDefinition{
		{Name: "plan", BehaviorSummary: "  Use --command plan to scaffold the next bugfix SPEC package under .namba/specs/.", Run: (*App).runFixPlanSubcommand},
		{Name: "run", BehaviorSummary: "  Use --command run, or omit --command, to repair the issue directly in the current workspace.", Run: (*App).runFixRunSubcommand},
	}
}

func fixSubcommandBehaviorSummaries() []string {
	lines := make([]string, 0, len(fixSubcommandDefinitions()))
	for _, definition := range fixSubcommandDefinitions() {
		lines = append(lines, definition.BehaviorSummary)
	}
	return lines
}

func (a *App) resolveFixSubcommand(name string) (fixSubcommandDefinition, bool) {
	for _, definition := range fixSubcommandDefinitions() {
		if definition.Name == name {
			return definition, true
		}
	}
	return fixSubcommandDefinition{}, false
}

func (a *App) runFixPlanSubcommand(ctx context.Context, _ string, options fixInvocation) error {
	if clarification, ok := a.evaluateSpecCreationClarification("fix", options.description); ok {
		fmt.Fprint(a.stdout, clarification)
		return errors.New("namba fix --command plan requires clarification before creating a SPEC")
	}
	return a.createSpecPackage(ctx, "fix", options.description, options.currentWorkspace, false)
}

func (a *App) runFixRunSubcommand(ctx context.Context, root string, options fixInvocation) error {
	return a.executeDirectFix(ctx, root, options.description)
}

func isStandaloneFlagToken(arg string) bool {
	trimmed := strings.TrimSpace(arg)
	if !strings.HasPrefix(trimmed, "-") {
		return false
	}
	return !strings.ContainsAny(trimmed, " \t\r\n")
}

type specPackageScaffoldContext struct {
	Root        string
	Kind        string
	Description string
	SpecID      string
	ProjectCfg  projectConfig
	QualityCfg  qualityConfig
}

func (a *App) loadSpecPackageScaffoldContext(kind, description string) (specPackageScaffoldContext, error) {
	root, err := a.requireProjectRoot()
	if err != nil {
		return specPackageScaffoldContext{}, err
	}
	specID, err := nextSpecID(filepath.Join(root, specsDir))
	if err != nil {
		return specPackageScaffoldContext{}, err
	}
	return a.loadResolvedSpecPackageScaffoldContext(root, kind, description, specID)
}

func (a *App) loadResolvedSpecPackageScaffoldContext(root, kind, description, specID string) (specPackageScaffoldContext, error) {
	projectCfg, _ := a.loadProjectConfig(root)
	qualityCfg, _ := a.loadQualityConfig(root)
	return specPackageScaffoldContext{
		Root:        root,
		Kind:        kind,
		Description: description,
		SpecID:      specID,
		ProjectCfg:  projectCfg,
		QualityCfg:  qualityCfg,
	}, nil
}

func buildSpecPackageScaffoldOutputs(scaffoldCtx specPackageScaffoldContext) map[string]string {
	outputs := map[string]string{
		filepath.ToSlash(filepath.Join(specsDir, scaffoldCtx.SpecID, "spec.md")):       buildSpecDoc(scaffoldCtx.Kind, scaffoldCtx.SpecID, scaffoldCtx.Description, scaffoldCtx.ProjectCfg, scaffoldCtx.QualityCfg),
		filepath.ToSlash(filepath.Join(specsDir, scaffoldCtx.SpecID, "plan.md")):       buildSpecPlanDoc(scaffoldCtx.Kind, scaffoldCtx.SpecID),
		filepath.ToSlash(filepath.Join(specsDir, scaffoldCtx.SpecID, "acceptance.md")): buildSpecAcceptanceDoc(scaffoldCtx.Kind, scaffoldCtx.Description, scaffoldCtx.QualityCfg.DevelopmentMode),
	}
	frontendBriefBody := ""
	if body, ok := buildFrontendBriefDoc(scaffoldCtx.Kind, scaffoldCtx.Description); ok {
		outputs[filepath.ToSlash(filepath.Join(specsDir, scaffoldCtx.SpecID, frontendBriefFileName))] = body
		frontendBriefBody = body
	}
	if req := inferredPlanningHarnessRequest(scaffoldCtx.Kind, scaffoldCtx.Description); req != nil {
		if body, err := marshalHarnessRequest(*req); err == nil {
			outputs[specHarnessRequestPath(scaffoldCtx.SpecID)] = body
		}
	}
	for rel, body := range specReviewOutputsWithFrontendBrief(scaffoldCtx.SpecID, frontendBriefBody) {
		outputs[rel] = body
	}
	return outputs
}

func (a *App) materializeSpecPackageScaffoldOutputs(scaffoldCtx specPackageScaffoldContext, outputs map[string]string) error {
	if err := a.mkdirAll(filepath.Join(scaffoldCtx.Root, specsDir, scaffoldCtx.SpecID), 0o755); err != nil {
		return err
	}
	if _, err := a.writeOutputs(scaffoldCtx.Root, outputs); err != nil {
		return err
	}
	return nil
}

func nextSpecID(root string) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	maxID := 0
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "SPEC-") {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(entry.Name(), "SPEC-"))
		if err == nil && n > maxID {
			maxID = n
		}
	}
	return fmt.Sprintf("SPEC-%03d", maxID+1), nil
}

func latestSpecID(root string) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	maxID := ""
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "SPEC-") && entry.Name() > maxID {
			maxID = entry.Name()
		}
	}
	if maxID == "" {
		return "none", nil
	}
	return maxID, nil
}
