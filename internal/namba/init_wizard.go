package namba

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func (a *App) runInitWizard(defaults initProfile) (initProfile, error) {
	if a.isInteractiveTerminal() {
		return a.runInitTUIWizard(defaults)
	}
	return a.runInitLineWizard(defaults)
}

func (a *App) runInitLineWizard(defaults initProfile) (initProfile, error) {
	reader := bufio.NewReader(a.stdin)
	profile := defaults

	renderInitBanner(a.stdout)
	renderLanguageFirstIntro(a.stdout, profile.ConversationLanguage)

	for step := wizardStepHumanLanguage; step < wizardStepDone; {
		allowBack := step != wizardStepHumanLanguage
		switch step {
		case wizardStepHumanLanguage:
			renderWizardStepHeader(a.stdout, step, profile, wizardMessage(profile.ConversationLanguage, "languageStep"))
			value, back := promptSelectWizardResult(a.stdin, reader, a.stdout, "Language / \uc5b8\uc5b4 / \u8a00\u8a9e / \u8bed\u8a00", languageOptions(), profile.ConversationLanguage, allowBack)
			if back {
				step = previousWizardStep(step, profile)
				continue
			}
			applyHumanLanguage(&profile, value)
			renderWizardWelcome(a.stdout, profile)
			renderRepoIntelligenceHeader(a.stdout, profile)
			step = nextWizardStep(step, profile)
		case wizardStepProjectType:
			renderWizardStepHeader(a.stdout, step, profile, wizardMessage(profile.ConversationLanguage, "repoState"))
			value, back := promptSelectWizardResult(a.stdin, reader, a.stdout, wizardMessage(profile.ConversationLanguage, "setupPathPrompt"), projectTypeOptionsForLanguage(profile.ConversationLanguage, profile.ProjectType), profile.ProjectType, allowBack)
			if back {
				step = previousWizardStep(step, profile)
				continue
			}
			profile.ProjectType = value
			step = nextWizardStep(step, profile)
		case wizardStepProjectDetails:
			renderWizardStepHeader(a.stdout, step, profile, wizardMessage(profile.ConversationLanguage, "projectDefaults"))
			next, back := a.promptProjectScaffold(reader, profile, allowBack)
			if back {
				step = previousWizardStep(step, profile)
				continue
			}
			profile = next
			step = nextWizardStep(step, profile)
		case wizardStepDevelopmentMode:
			renderWizardStepHeader(a.stdout, step, profile, wizardMessage(profile.ConversationLanguage, "workMode"))
			value, back := promptSelectWizardResult(
				a.stdin,
				reader,
				a.stdout,
				wizardMessage(profile.ConversationLanguage, "workModePrompt"),
				developmentModeOptionsForLanguage(profile.ConversationLanguage),
				profile.DevelopmentMode,
				allowBack,
			)
			if back {
				step = previousWizardStep(step, profile)
				continue
			}
			profile.DevelopmentMode = value
			step = nextWizardStep(step, profile)
		case wizardStepAgentMode:
			renderWizardStepHeader(a.stdout, step, profile, wizardMessage(profile.ConversationLanguage, "agentStep"))
			value, back := promptSelectWizardResult(
				a.stdin,
				reader,
				a.stdout,
				wizardMessage(profile.ConversationLanguage, "agentPrompt"),
				agentModeOptionsForLanguage(profile.ConversationLanguage),
				profile.AgentMode,
				allowBack,
			)
			if back {
				step = previousWizardStep(step, profile)
				continue
			}
			profile.AgentMode = value
			step = nextWizardStep(step, profile)
		case wizardStepStatusLine:
			renderWizardStepHeader(a.stdout, step, profile, wizardMessage(profile.ConversationLanguage, "statusStep"))
			value, back := promptSelectWizardResult(
				a.stdin,
				reader,
				a.stdout,
				wizardMessage(profile.ConversationLanguage, "statusPrompt"),
				statusLineOptionsForLanguage(profile.ConversationLanguage),
				profile.StatusLinePreset,
				allowBack,
			)
			if back {
				step = previousWizardStep(step, profile)
				continue
			}
			profile.StatusLinePreset = value
			step = nextWizardStep(step, profile)
		case wizardStepCodexAccess:
			renderWizardStepHeader(a.stdout, step, profile, "Codex access")
			fmt.Fprintln(a.stdout, wizardHint(a.stdout, "[next] "+wizardMessage(profile.ConversationLanguage, "codexAccessGuide")))
			next, back, err := a.promptCodexAccessStep(reader, profile, allowBack)
			if err != nil {
				return initProfile{}, err
			}
			if back {
				step = previousWizardStep(step, profile)
				continue
			}
			profile = next
			step = nextWizardStep(step, profile)
		case wizardStepGitMode:
			renderWizardStepHeader(a.stdout, step, profile, "Git")
			fmt.Fprintln(a.stdout, wizardHint(a.stdout, "[next] "+wizardMessage(profile.ConversationLanguage, "gitGuide")))
			value, back := promptSelectWizardResult(
				a.stdin,
				reader,
				a.stdout,
				wizardMessage(profile.ConversationLanguage, "gitModePrompt"),
				gitModeOptionsForLanguage(profile.ConversationLanguage),
				profile.GitMode,
				allowBack,
			)
			if back {
				step = previousWizardStep(step, profile)
				continue
			}
			profile.GitMode = value
			step = nextWizardStep(step, profile)
		case wizardStepGitProvider:
			renderWizardStepHeader(a.stdout, step, profile, "Git provider")
			value, back := promptSelectWizardResult(
				a.stdin,
				reader,
				a.stdout,
				wizardMessage(profile.ConversationLanguage, "gitProviderPrompt"),
				gitProviderOptionsForLanguage(profile.ConversationLanguage),
				profile.GitProvider,
				allowBack,
			)
			if back {
				step = previousWizardStep(step, profile)
				continue
			}
			profile.GitProvider = value
			step = nextWizardStep(step, profile)
		case wizardStepGitLabURL:
			renderWizardStepHeader(a.stdout, step, profile, "GitLab URL")
			value, back := promptInputResult(reader, a.stdout, wizardMessage(profile.ConversationLanguage, "gitLabURL"), profile.GitLabInstanceURL, allowBack)
			if back {
				step = previousWizardStep(step, profile)
				continue
			}
			profile.GitLabInstanceURL = value
			step = nextWizardStep(step, profile)
		case wizardStepDisplayName:
			renderWizardStepHeader(a.stdout, step, profile, wizardMessage(profile.ConversationLanguage, "displayName"))
			value, back := promptInputResult(reader, a.stdout, wizardMessage(profile.ConversationLanguage, "displayName"), profile.UserName, allowBack)
			if back {
				step = previousWizardStep(step, profile)
				continue
			}
			profile.UserName = value
			step = nextWizardStep(step, profile)
		}
	}

	fmt.Fprintln(a.stdout)
	renderInitWizardSummary(a.stdout, profile)
	fmt.Fprintln(a.stdout)
	fmt.Fprintln(a.stdout, wizardHint(a.stdout, "[next] "+wizardMessage(profile.ConversationLanguage, "secretHint")))
	return profile, nil
}

const (
	wizardStepHumanLanguage = iota
	wizardStepProjectType
	wizardStepProjectDetails
	wizardStepDevelopmentMode
	wizardStepAgentMode
	wizardStepStatusLine
	wizardStepCodexAccess
	wizardStepGitMode
	wizardStepGitProvider
	wizardStepGitLabURL
	wizardStepDisplayName
	wizardStepDone
)

func nextWizardStep(step int, profile initProfile) int {
	switch step {
	case wizardStepGitMode:
		if profile.GitMode == "manual" {
			return wizardStepDisplayName
		}
		return wizardStepGitProvider
	case wizardStepGitProvider:
		if profile.GitProvider == "gitlab" {
			return wizardStepGitLabURL
		}
		return wizardStepDisplayName
	case wizardStepGitLabURL:
		return wizardStepDisplayName
	case wizardStepDisplayName:
		return wizardStepDone
	default:
		return step + 1
	}
}

func previousWizardStep(step int, profile initProfile) int {
	switch step {
	case wizardStepDisplayName:
		if profile.GitMode == "manual" {
			return wizardStepGitMode
		}
		if profile.GitProvider == "gitlab" {
			return wizardStepGitLabURL
		}
		return wizardStepGitProvider
	case wizardStepGitLabURL:
		return wizardStepGitProvider
	case wizardStepGitProvider:
		return wizardStepGitMode
	default:
		if step <= wizardStepHumanLanguage {
			return wizardStepHumanLanguage
		}
		return step - 1
	}
}

func renderWizardStepHeader(out io.Writer, step int, profile initProfile, title string) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, wizardHint(out, wizardProgressRail(step, profile)))
	fmt.Fprintln(out, wizardHeading(out, fmt.Sprintf("%s Step %02d · %s", wizardStepMarker(out, step), wizardVisibleStepNumber(step, profile), title)))
}

func wizardVisibleStepNumber(target int, profile initProfile) int {
	if target <= wizardStepHumanLanguage {
		return 1
	}
	step := wizardStepHumanLanguage
	visible := 1
	for guard := 0; step < wizardStepDone && guard < wizardStepDone+1; guard++ {
		if step == target {
			return visible
		}
		step = nextWizardStep(step, profile)
		visible++
	}
	if target == wizardStepDone {
		return visible
	}
	return target + 1
}

func wizardProgressRail(step int, profile initProfile) string {
	total := wizardVisibleStepNumber(wizardStepDone, profile) - 1
	current := wizardVisibleStepNumber(step, profile)
	return fmt.Sprintf("[current] Step %02d/%02d [next] %s", current, total, wizardNextStepLabel(step, profile))
}

func wizardNextStepLabel(step int, profile initProfile) string {
	next := nextWizardStep(step, profile)
	if next == wizardStepDone {
		return "ready handoff"
	}
	switch next {
	case wizardStepHumanLanguage:
		return "language"
	case wizardStepProjectType:
		return "repository state"
	case wizardStepProjectDetails:
		return "project defaults"
	case wizardStepDevelopmentMode:
		return "work mode"
	case wizardStepAgentMode:
		return "agent mode"
	case wizardStepStatusLine:
		return "status line"
	case wizardStepCodexAccess:
		return "codex access"
	case wizardStepGitMode, wizardStepGitProvider, wizardStepGitLabURL:
		return "git setup"
	default:
		return "display name"
	}
}

func wizardStepEmoji(step int) string {
	switch step {
	case wizardStepProjectType:
		return "\U0001f9ed"
	case wizardStepProjectDetails:
		return "\U0001f4e6"
	case wizardStepHumanLanguage:
		return "\U0001f310"
	case wizardStepDevelopmentMode:
		return "\U0001f9ea"
	case wizardStepAgentMode:
		return "\U0001f916"
	case wizardStepStatusLine:
		return "\U0001f39b\ufe0f"
	case wizardStepCodexAccess:
		return "\U0001f510"
	case wizardStepGitMode, wizardStepGitProvider, wizardStepGitLabURL:
		return "\U0001f33f"
	case wizardStepDisplayName:
		return "\U0001f64b"
	default:
		return "\u2728"
	}
}

func wizardStepMarker(out io.Writer, step int) string {
	if !isTerminalWriter(out) {
		return ""
	}
	return wizardStepEmoji(step)
}

func renderLanguageFirstIntro(out io.Writer, defaultLanguage string) {
	fmt.Fprintln(out, wizardHeading(out, "NambaAI init language setup"))
	fmt.Fprintln(out, wizardHint(out, "Choose setup language first. Supported: [ko] Korean, [en] English, [ja] Japanese, [zh] Simplified Chinese."))
	fmt.Fprintf(out, "%s\n", wizardHint(out, fmt.Sprintf("[default] %s. Continue with number or code input.", humanLanguageLabel(defaultLanguage))))
	fmt.Fprintln(out)
}

func renderWizardWelcome(out io.Writer, profile initProfile) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, wizardHeading(out, wizardMessage(profile.ConversationLanguage, "welcomeTitle")))
	fmt.Fprintln(out, wizardHint(out, wizardMessage(profile.ConversationLanguage, "welcomeHint")))
	fmt.Fprintln(out, wizardHint(out, wizardMessage(profile.ConversationLanguage, "backHint")))
}

func renderRepoIntelligenceHeader(out io.Writer, profile initProfile) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, wizardHeading(out, wizardMessage(profile.ConversationLanguage, "repoIntelTitle")))
	fmt.Fprintf(out, "  [current] Target: %s\n", profile.ProjectName)
	fmt.Fprintf(out, "  [current] Repository state: %s\n", profile.ProjectType)
	fmt.Fprintf(out, "  [current] Detected stack: %s\n", formatInitStack(profile))
	fmt.Fprintf(out, "  [current] Methodology: %s\n", strings.ToUpper(firstNonBlank(profile.DevelopmentMode, "tdd")))
	fmt.Fprintf(out, "  [default] %s\n", wizardMessage(profile.ConversationLanguage, "repoIntelDefault"))
}

func wizardMessage(language, key string) string {
	lang := normalizeReadmeLanguage(language)
	messages := map[string]map[string]string{
		"ko": {
			"languageStep":      "사용 언어",
			"welcomeTitle":      "NambaAI 초기화 마법사",
			"welcomeHint":       "먼저 언어를 고른 뒤 저장소 상태를 확인하고, 스택은 감지값이나 첫 계획 요청으로 정합니다.",
			"backHint":          "선택 후에는 결과를 표시하고 `b` 또는 `back`으로 이전 단계를 수정할 수 있습니다.",
			"repoIntelTitle":    "저장소 인텔리전스",
			"repoIntelDefault":  "기존 repo는 감지된 stack을 유지하고, 빈 repo는 app stack을 비워 둡니다.",
			"repoState":         "저장소 상태",
			"setupPathPrompt":   "설정 경로",
			"projectDefaults":   "프로젝트 기본값",
			"projectName":       "프로젝트 이름",
			"workMode":          "작업 방식",
			"workModePrompt":    "기본 작업 방식",
			"agentStep":         "Codex 에이전트",
			"agentPrompt":       "Codex 에이전트 모드",
			"statusStep":        "상태줄",
			"statusPrompt":      "상태줄 프리셋",
			"gitModePrompt":     "Git 자동화 모드",
			"gitProviderPrompt": "Git 제공자",
			"gitLabURL":         "GitLab 인스턴스 URL",
			"displayName":       "표시 이름",
			"codexAccessGuide":  "Codex access preset은 Namba runner의 approval_policy / sandbox_mode 결과를 먼저 보여줍니다.",
			"gitGuide":          "Git은 manual, personal, team 중 고릅니다. GitHub 사용자명은 묻지 않고 기존 gh/glab 인증을 사용합니다.",
			"secretHint":        "토큰과 비밀값은 설정 파일에 저장하지 않습니다. gh/glab login을 사용하세요.",
			"summaryTitle":      "준비 완료 핸드오프",
		},
		"ja": {
			"languageStep":      "使用言語",
			"welcomeTitle":      "NambaAI 初期化ウィザード",
			"welcomeHint":       "言語を先に選び、その後で repository state を確認し、stack は検出値または最初の計画依頼で決めます。",
			"backHint":          "選択後は結果を表示し、`b` または `back` で前の step に戻れます。",
			"repoIntelTitle":    "Repository intelligence",
			"repoIntelDefault":  "既存 repo は検出 stack を維持し、空 repo は app stack を未設定のままにします。",
			"repoState":         "Repository state",
			"setupPathPrompt":   "Setup path",
			"projectDefaults":   "Project defaults",
			"projectName":       "Project name",
			"workMode":          "Work mode",
			"workModePrompt":    "Default work mode",
			"agentStep":         "Codex agent",
			"agentPrompt":       "Codex agent mode",
			"statusStep":        "Status line",
			"statusPrompt":      "Status line preset",
			"gitModePrompt":     "Git automation mode",
			"gitProviderPrompt": "Git provider",
			"gitLabURL":         "GitLab instance URL",
			"displayName":       "Display name",
			"codexAccessGuide":  "Codex access preset は Namba runner の approval_policy / sandbox_mode を事前表示します。",
			"gitGuide":          "Git は manual、personal、team から選びます。GitHub username は尋ねず既存の gh/glab 認証を使います。",
			"secretHint":        "Token と secret は設定ファイルに保存しません。gh/glab login を使ってください。",
			"summaryTitle":      "Ready handoff",
		},
		"zh": {
			"languageStep":      "使用语言",
			"welcomeTitle":      "NambaAI 初始化向导",
			"welcomeHint":       "先选择语言，然后确认仓库状态；stack 使用检测值，或在第一次计划请求中确定。",
			"backHint":          "选择后会显示结果，也可以用 `b` 或 `back` 回到上一步修改。",
			"repoIntelTitle":    "Repository intelligence",
			"repoIntelDefault":  "已有 repo 保留检测到的 stack；空 repo 保持 app stack 未设置。",
			"repoState":         "Repository state",
			"setupPathPrompt":   "Setup path",
			"projectDefaults":   "Project defaults",
			"projectName":       "Project name",
			"workMode":          "Work mode",
			"workModePrompt":    "Default work mode",
			"agentStep":         "Codex agent",
			"agentPrompt":       "Codex agent mode",
			"statusStep":        "Status line",
			"statusPrompt":      "Status line preset",
			"gitModePrompt":     "Git automation mode",
			"gitProviderPrompt": "Git provider",
			"gitLabURL":         "GitLab instance URL",
			"displayName":       "Display name",
			"codexAccessGuide":  "Codex access preset 会预览 Namba runner 的 approval_policy / sandbox_mode。",
			"gitGuide":          "Git 可选择 manual、personal 或 team。不会询问 GitHub 用户名，会使用已有 gh/glab 认证。",
			"secretHint":        "Token 和 secret 不会保存到配置文件。请使用 gh/glab login。",
			"summaryTitle":      "Ready handoff",
		},
		"en": {
			"languageStep":      "Working language",
			"welcomeTitle":      "NambaAI init wizard",
			"welcomeHint":       "Choose language first, then review repository state. Existing stacks come from detection; empty repos decide stack in the first plan.",
			"backHint":          "Each answer is echoed before moving on, and `b` or `back` returns to the previous step.",
			"repoIntelTitle":    "Repository intelligence",
			"repoIntelDefault":  "Existing repos keep detected stack values; empty repos leave the app stack unset.",
			"repoState":         "Repository state",
			"setupPathPrompt":   "Setup path",
			"projectDefaults":   "Project defaults",
			"projectName":       "Project name",
			"workMode":          "Work mode",
			"workModePrompt":    "Default work mode",
			"agentStep":         "Codex agent",
			"agentPrompt":       "Codex agent mode",
			"statusStep":        "Status line",
			"statusPrompt":      "Status line preset",
			"gitModePrompt":     "Git automation mode",
			"gitProviderPrompt": "Git provider",
			"gitLabURL":         "GitLab instance URL",
			"displayName":       "Display name",
			"codexAccessGuide":  "Codex access presets preview the resulting Namba runner approval_policy / sandbox_mode pair.",
			"gitGuide":          "Git setup uses manual, personal, or team mode. It does not ask for a GitHub username and relies on existing gh/glab auth.",
			"secretHint":        "Tokens and secrets are not stored in config files. Use gh/glab login.",
			"summaryTitle":      "Ready handoff",
		},
	}
	if value := messages[lang][key]; value != "" {
		return value
	}
	return messages["en"][key]
}

func (a *App) promptProjectScaffold(reader *bufio.Reader, profile initProfile, allowBack bool) (initProfile, bool) {
	projectName, back := promptInputResult(reader, a.stdout, wizardMessage(profile.ConversationLanguage, "projectName"), profile.ProjectName, allowBack)
	if back {
		return profile, true
	}
	profile.ProjectName = projectName

	if profile.ProjectType == "existing" {
		profile.Framework = normalizeFramework(profile.Framework)
		fmt.Fprintln(a.stdout, wizardHint(a.stdout, fmt.Sprintf("%s: %s", wizardProjectScaffoldMessage(profile.ConversationLanguage, "detectedCodebase"), formatInitStack(profile))))
		fmt.Fprintln(a.stdout, wizardHint(a.stdout, wizardProjectScaffoldMessage(profile.ConversationLanguage, "existingHint")))
		return profile, false
	}

	profile.Language = firstNonBlank(profile.Language, "unknown")
	profile.Framework = normalizeFramework(profile.Framework)
	fmt.Fprintln(a.stdout, wizardHint(a.stdout, wizardProjectScaffoldMessage(profile.ConversationLanguage, "newHint")))
	return profile, false
}

func wizardProjectScaffoldMessage(language, key string) string {
	lang := normalizeReadmeLanguage(language)
	messages := map[string]map[string]string{
		"ko": {
			"detectedCodebase": "감지된 코드베이스",
			"existingHint":     "기존 코드의 언어/프레임워크는 묻지 않고 감지값을 사용합니다. 필요하면 init flag로 override하세요.",
			"newHint":          "빈 저장소에서는 앱 스택을 묻지 않습니다. NambaAI만 준비하고, 첫 `namba plan`에서 목표/제약에 맞게 스택을 정합니다.",
		},
		"ja": {
			"detectedCodebase": "Detected codebase",
			"existingHint":     "The existing code language/framework is not asked again; detected values are used unless init flags override them.",
			"newHint":          "Empty repositories do not choose an app stack here. NambaAI is prepared now, and the first `namba plan` decides the stack from the goal and constraints.",
		},
		"zh": {
			"detectedCodebase": "Detected codebase",
			"existingHint":     "已有代码的 language/framework 不会再次询问；默认使用检测值，除非用 init flag 覆盖。",
			"newHint":          "空仓库不会在这里选择 app stack。现在只准备 NambaAI，第一次 `namba plan` 会根据目标和约束确定 stack。",
		},
		"en": {
			"detectedCodebase": "Detected codebase",
			"existingHint":     "The existing code language/framework is not asked again; detected values are used unless init flags override them.",
			"newHint":          "Empty repositories do not choose an app stack here. NambaAI is prepared now, and the first `namba plan` decides the stack from the goal and constraints.",
		},
	}
	if value := messages[lang][key]; value != "" {
		return value
	}
	return messages["en"][key]
}

func renderInitWizardSummary(out io.Writer, profile initProfile) {
	fmt.Fprintln(out, wizardHeading(out, wizardMessage(profile.ConversationLanguage, "summaryTitle")))
	fmt.Fprintf(out, "  [current] Project: %s (%s)\n", profile.ProjectName, profile.ProjectType)
	fmt.Fprintf(out, "  [current] Stack: %s\n", formatInitStack(profile))
	fmt.Fprintf(out, "  [current] Working language: %s\n", humanLanguageLabel(profile.ConversationLanguage))
	fmt.Fprintf(out, "  [current] Codex access: approval_policy=%s, sandbox_mode=%s\n", profile.ApprovalPolicy, profile.SandboxMode)
	fmt.Fprintf(out, "  [current] Git automation: %s", profile.GitMode)
	if profile.GitMode != "manual" {
		fmt.Fprintf(out, " via %s", profile.GitProvider)
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  [next] Run `namba project`, then plan or execute SPEC work from this repository.")
}

func promptInput(reader *bufio.Reader, out io.Writer, label, defaultValue string) string {
	value, _ := promptInputResult(reader, out, label, defaultValue, false)
	return value
}

func promptInputResult(reader *bufio.Reader, out io.Writer, label, defaultValue string, allowBack bool) (string, bool) {
	prompt := wizardPrompt(out, label)
	if strings.TrimSpace(defaultValue) == "" {
		fmt.Fprintf(out, "%s: ", prompt)
	} else {
		fmt.Fprintf(out, "%s [%s]: ", prompt, defaultValue)
	}
	if allowBack {
		fmt.Fprintf(out, "%s ", wizardHint(out, "(\u21a9\ufe0f b/back: back)"))
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		value := strings.TrimSpace(defaultValue)
		printWizardSelection(out, label, value)
		return value, false
	}
	line = strings.TrimSpace(line)
	if allowBack && isWizardBackInput(line) {
		printWizardBack(out)
		return strings.TrimSpace(defaultValue), true
	}
	if line == "" {
		value := strings.TrimSpace(defaultValue)
		printWizardSelection(out, label, value)
		return value, false
	}
	printWizardSelection(out, label, line)
	return line, false
}

func renderInitBanner(out io.Writer) {
	fmt.Fprintln(out, " _   _    _    __  __ ____    _      ___ ")
	fmt.Fprintln(out, "| \\ | |  / \\  |  \\/  | __ )  / \\    |_ _|")
	fmt.Fprintln(out, "|  \\| | / _ \\ | |\\/| |  _ \\ / _ \\    | | ")
	fmt.Fprintln(out, "| |\\  |/ ___ \\| |  | | |_) / ___ \\   | | ")
	fmt.Fprintln(out, "|_| \\_/_/   \\_\\_|  |_|____/_/   \\_\\ |___|")
	fmt.Fprintln(out, wizardHint(out, "\u2728 \ud504\ub85c\uc81d\ud2b8 \uc124\uc815\uc744 \uc2dc\uc791\ud569\ub2c8\ub2e4"))
	fmt.Fprintln(out)
}

func wizardHeading(out io.Writer, text string) string {
	return styleWizardText(out, "1;36", text)
}

func wizardHint(out io.Writer, text string) string {
	return styleWizardText(out, "2;37", text)
}

func wizardPrompt(out io.Writer, text string) string {
	return styleWizardText(out, "1;33", text)
}

func wizardSelected(out io.Writer, text string) string {
	return styleWizardText(out, "1;32", text)
}

func styleWizardText(out io.Writer, code, text string) string {
	if !isTerminalWriter(out) {
		return stripWizardDecorations(text)
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func formatWizardChoice(choice option) string {
	return formatWizardChoiceFor(nil, choice)
}

func formatWizardChoiceFor(out io.Writer, choice option) string {
	label := choice.Label
	description := choice.Description
	if out != nil && !isTerminalWriter(out) {
		label = stripWizardDecorations(label)
		description = stripWizardDecorations(description)
	}
	if strings.TrimSpace(choice.Description) == "" {
		return label
	}
	return fmt.Sprintf("%s - %s", label, description)
}

func choiceLabel(choices []option, value string) string {
	return choiceLabelFor(nil, choices, value)
}

func choiceLabelFor(out io.Writer, choices []option, value string) string {
	for _, choice := range choices {
		if choice.Value == value {
			if out != nil && !isTerminalWriter(out) {
				return stripWizardDecorations(choice.Label)
			}
			return choice.Label
		}
	}
	return value
}

func stripWizardDecorations(text string) string {
	replacer := strings.NewReplacer(
		"✨ ", "", "🚀 ", "", "🧭 ", "", "💡 ", "", "📦 ", "", "🌱 ", "", "🔎 ", "", "🛠️ ", "", "🧪 ", "", "🤖 ", "", "👤 ", "", "👥 ", "", "🎛️ ", "", "🔕 ", "", "🔐 ", "", "🛡️ ", "", "⚖️ ", "", "🔥 ", "", "🧩 ", "", "✅ ", "", "🔒 ", "", "🛎️ ", "", "🧱 ", "", "🌿 ", "", "✋ ", "", "☁️ ", "", "🐙 ", "", "🦊 ", "", "🔗 ", "", "🙋 ", "", "📋 ", "", "📛 ", "", "🌐 ", "", "👉 ", "", "↩️ ", "", "💬 ", "",
		"🇰🇷 ", "", "🇺🇸 ", "", "🇯🇵 ", "", "🇨🇳 ", "",
	)
	return strings.TrimSpace(replacer.Replace(text))
}

func isWizardBackInput(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "b", "back", "<", "\uc774\uc804":
		return true
	default:
		return false
	}
}

func printWizardSelection(out io.Writer, label, value string) {
	if strings.TrimSpace(value) == "" {
		value = "(empty)"
	}
	fmt.Fprintf(out, "%s\n", wizardSelected(out, fmt.Sprintf("\u2705 %s: %s", label, value)))
}

func printWizardBack(out io.Writer) {
	fmt.Fprintln(out, wizardHint(out, "\u21a9\ufe0f Back to the previous step."))
}

func promptSelect(in io.Reader, out io.Writer, label string, choices []option, defaultValue string) string {
	return promptSelectWizard(in, bufio.NewReader(in), out, label, choices, defaultValue)
}

func promptSelectWizard(in io.Reader, reader *bufio.Reader, out io.Writer, label string, choices []option, defaultValue string) string {
	value, _ := promptSelectWizardResult(in, reader, out, label, choices, defaultValue, false)
	return value
}

func promptSelectWizardResult(in io.Reader, reader *bufio.Reader, out io.Writer, label string, choices []option, defaultValue string, allowBack bool) (string, bool) {
	if file, ok := in.(*os.File); ok && isTerminalReader(in) && isTerminalWriter(out) {
		if value, back, ok := promptSelectInteractive(file, out, label, choices, defaultValue, allowBack); ok {
			if back {
				printWizardBack(out)
				return defaultValue, true
			}
			printWizardSelection(out, label, choiceLabelFor(out, choices, value))
			return value, false
		}
	}
	value, back := promptSelectLineResult(reader, out, label, choices, defaultValue, allowBack)
	if back {
		printWizardBack(out)
		return defaultValue, true
	}
	printWizardSelection(out, label, choiceLabelFor(out, choices, value))
	return value, false
}

func promptSelectLine(reader *bufio.Reader, out io.Writer, label string, choices []option, defaultValue string) string {
	value, _ := promptSelectLineResult(reader, out, label, choices, defaultValue, false)
	return value
}

func promptSelectLineResult(reader *bufio.Reader, out io.Writer, label string, choices []option, defaultValue string, allowBack bool) (string, bool) {
	fmt.Fprintln(out, wizardHeading(out, label))
	defaultIndex := 0
	for i, choice := range choices {
		if choice.Value == defaultValue {
			defaultIndex = i
		}
		fmt.Fprintf(out, "  %d. %s\n", i+1, formatWizardChoiceFor(out, choice))
	}
	if allowBack {
		if isTerminalWriter(out) {
			fmt.Fprintln(out, "  \u21a9\ufe0f b. back")
		} else {
			fmt.Fprintln(out, "  b. back")
		}
	}
	fmt.Fprintf(out, "%s [%d]: ", wizardPrompt(out, "\uc120\ud0dd"), defaultIndex+1)

	line, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue, false
	}
	line = strings.TrimSpace(line)
	if allowBack && isWizardBackInput(line) {
		return defaultValue, true
	}
	if line == "" {
		return defaultValue, false
	}
	index, err := strconv.Atoi(line)
	if err == nil && index >= 1 && index <= len(choices) {
		return choices[index-1].Value, false
	}
	for _, choice := range choices {
		if strings.EqualFold(choice.Value, line) || strings.EqualFold(choice.Label, line) {
			return choice.Value, false
		}
	}
	return defaultValue, false
}

type menuAction int

const (
	menuActionUnknown menuAction = iota
	menuActionUp
	menuActionDown
	menuActionSubmit
	menuActionBack
)

func promptSelectInteractive(in *os.File, out io.Writer, label string, choices []option, defaultValue string, allowBack bool) (string, bool, bool) {
	restoreInput, err := enableRawConsoleInput(in)
	if err != nil {
		return "", false, false
	}
	defer restoreInput()

	restoreOutput := enableVirtualTerminalOutput(out)
	defer restoreOutput()

	selected := 0
	for i, choice := range choices {
		if choice.Value == defaultValue {
			selected = i
			break
		}
	}

	reader := bufio.NewReader(in)
	lines := 0
	fmt.Fprint(out, "\x1b[?25l")
	defer fmt.Fprint(out, "\x1b[?25h")

	for {
		if lines > 0 {
			fmt.Fprintf(out, "\x1b[%dA", lines)
		}
		lines = renderInteractiveSelect(out, label, choices, selected, allowBack)

		action, err := readMenuAction(reader)
		if err != nil {
			fmt.Fprintln(out)
			return defaultValue, false, true
		}

		switch action {
		case menuActionUp:
			selected = (selected - 1 + len(choices)) % len(choices)
		case menuActionDown:
			selected = (selected + 1) % len(choices)
		case menuActionBack:
			if allowBack {
				fmt.Fprintln(out)
				return defaultValue, true, true
			}
		case menuActionSubmit:
			fmt.Fprintln(out)
			return choices[selected].Value, false, true
		}
	}
}

func renderInteractiveSelect(out io.Writer, label string, choices []option, selected int, allowBack bool) int {
	lines := 0
	fmt.Fprintf(out, "\r\x1b[2K%s\n", wizardHeading(out, label))
	lines++
	hint := "\u2191/\u2193 move \u00b7 Enter select"
	if allowBack {
		hint += " \u00b7 \u21a9\ufe0f b back"
	}
	fmt.Fprintf(out, "\r\x1b[2K%s\n", wizardHint(out, hint))
	lines++
	for i, choice := range choices {
		line := fmt.Sprintf("%d. %s", i+1, formatWizardChoiceFor(out, choice))
		if i == selected {
			fmt.Fprintf(out, "\r\x1b[2K%s\n", wizardSelected(out, "\U0001f449 "+line))
		} else {
			fmt.Fprintf(out, "\r\x1b[2K  %s\n", line)
		}
		lines++
	}
	return lines
}

func readMenuAction(reader *bufio.Reader) (menuAction, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return menuActionUnknown, err
	}

	switch b {
	case '\r', '\n':
		return menuActionSubmit, nil
	case 'b', 'B', 0x7f:
		return menuActionBack, nil
	case 0x1b:
		next, err := reader.ReadByte()
		if err != nil {
			return menuActionUnknown, err
		}
		if next != '[' {
			return menuActionUnknown, nil
		}
		code, err := reader.ReadByte()
		if err != nil {
			return menuActionUnknown, err
		}
		switch code {
		case 'A':
			return menuActionUp, nil
		case 'B':
			return menuActionDown, nil
		default:
			return menuActionUnknown, nil
		}
	case 0x00, 0xe0:
		code, err := reader.ReadByte()
		if err != nil {
			return menuActionUnknown, err
		}
		switch code {
		case 72:
			return menuActionUp, nil
		case 80:
			return menuActionDown, nil
		default:
			return menuActionUnknown, nil
		}
	default:
		return menuActionUnknown, nil
	}
}

func (a *App) isInteractiveTerminal() bool {
	file, ok := a.stdin.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil || (info.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	return isTerminalWriter(a.stdout)
}

func isTerminalReader(r io.Reader) bool {
	file, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func isTerminalWriter(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
