package namba

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type initTUIStepKind int

const (
	initTUIStepChoice initTUIStepKind = iota
	initTUIStepText
)

type initTUICodexPhase int

const (
	initTUICodexPreset initTUICodexPhase = iota
	initTUICodexApproval
	initTUICodexSandbox
)

type initTUIModel struct {
	profile    initProfile
	step       int
	selected   int
	input      string
	codexPhase initTUICodexPhase
	done       bool
	canceled   bool
}

type initTUIStepState struct {
	Title       string
	Prompt      string
	Kind        initTUIStepKind
	Choices     []option
	Value       string
	Description string
}

type initTUIStyles struct {
	title    lipgloss.Style
	subtle   lipgloss.Style
	active   lipgloss.Style
	selected lipgloss.Style
	status   lipgloss.Style
	warning  lipgloss.Style
}

func (a *App) runInitTUIWizard(defaults initProfile) (initProfile, error) {
	program := tea.NewProgram(
		newInitTUIModel(defaults),
		tea.WithInput(a.stdin),
		tea.WithOutput(a.stdout),
	)
	finalModel, err := program.Run()
	if err != nil {
		return initProfile{}, err
	}
	model, ok := finalModel.(initTUIModel)
	if !ok {
		return initProfile{}, fmt.Errorf("init TUI returned unexpected model %T", finalModel)
	}
	if model.canceled {
		return initProfile{}, fmt.Errorf("init canceled")
	}
	if err := validateInitProfile(model.profile); err != nil {
		return initProfile{}, err
	}
	return model.profile, nil
}

func newInitTUIModel(profile initProfile) initTUIModel {
	model := initTUIModel{profile: profile, step: wizardStepHumanLanguage}
	model.resetInput()
	return model
}

func (m initTUIModel) Init() tea.Cmd {
	return nil
}

func (m initTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "ctrl+c", "q":
		m.canceled = true
		return m, tea.Quit
	case "esc", "left":
		m.back()
		return m, nil
	case "up":
		m.moveSelection(-1)
		return m, nil
	case "down":
		m.moveSelection(1)
		return m, nil
	case "enter", "right":
		if err := m.commit(); err != nil {
			m.canceled = true
			return m, tea.Quit
		}
		if m.done {
			return m, tea.Quit
		}
		return m, nil
	case "backspace":
		if m.currentStep().Kind == initTUIStepText && len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil
	default:
		if m.currentStep().Kind == initTUIStepText && key.Key().Text != "" {
			m.input += key.Key().Text
		}
		return m, nil
	}
}

func (m initTUIModel) View() tea.View {
	styles := defaultInitTUIStyles()
	step := m.currentStep()
	lines := []string{
		styles.title.Render("NambaAI init wizard"),
		styles.subtle.Render(wizardProgressRail(m.step, m.profile)),
		"",
		styles.active.Render(step.Title),
		styles.subtle.Render(step.Prompt),
	}
	if strings.TrimSpace(step.Description) != "" {
		lines = append(lines, styles.subtle.Render(step.Description))
	}
	lines = append(lines, "")

	switch step.Kind {
	case initTUIStepChoice:
		for i, choice := range step.Choices {
			row := fmt.Sprintf("%s  %s", choice.Label, choice.Description)
			if i == m.selected {
				lines = append(lines, styles.selected.Render("> "+row))
			} else {
				lines = append(lines, "  "+stripWizardDecorations(row))
			}
		}
	case initTUIStepText:
		value := m.input
		if value == "" {
			value = step.Value
		}
		lines = append(lines, styles.selected.Render("> "+value))
	}

	lines = append(lines,
		"",
		styles.status.Render("↑/↓ move  ←/Esc back  →/Enter continue  q quit"),
		"",
		styles.subtle.Render(fmt.Sprintf("Project %s · %s · %s", m.profile.ProjectName, m.profile.ProjectType, formatInitStack(m.profile))),
		styles.subtle.Render(fmt.Sprintf("Codex approval_policy=%s · sandbox_mode=%s", m.profile.ApprovalPolicy, m.profile.SandboxMode)),
	)
	if m.step == wizardStepCodexAccess {
		if choice, err := resolveCodexAccessChoice(approvalPolicy(m.profile), sandboxMode(m.profile)); err == nil {
			lines = append(lines, styles.warning.Render(choice.Consequence))
		}
	}

	view := tea.NewView(strings.Join(lines, "\n"))
	view.AltScreen = true
	return view
}

func (m initTUIModel) currentStep() initTUIStepState {
	switch m.step {
	case wizardStepHumanLanguage:
		return initTUIStepState{
			Title:   "Language / 언어 / 言語 / 语言",
			Prompt:  wizardMessage(m.profile.ConversationLanguage, "languageStep"),
			Kind:    initTUIStepChoice,
			Choices: languageOptions(),
			Value:   m.profile.ConversationLanguage,
		}
	case wizardStepProjectType:
		return initTUIStepState{
			Title:   wizardMessage(m.profile.ConversationLanguage, "repoState"),
			Prompt:  wizardMessage(m.profile.ConversationLanguage, "setupPathPrompt"),
			Kind:    initTUIStepChoice,
			Choices: projectTypeOptionsForLanguage(m.profile.ConversationLanguage, m.profile.ProjectType),
			Value:   m.profile.ProjectType,
		}
	case wizardStepProjectDetails:
		description := wizardProjectScaffoldMessage(m.profile.ConversationLanguage, "newHint")
		if m.profile.ProjectType == "existing" {
			description = fmt.Sprintf("%s: %s. %s", wizardProjectScaffoldMessage(m.profile.ConversationLanguage, "detectedCodebase"), formatInitStack(m.profile), wizardProjectScaffoldMessage(m.profile.ConversationLanguage, "existingHint"))
		}
		return initTUIStepState{
			Title:       wizardMessage(m.profile.ConversationLanguage, "projectDefaults"),
			Prompt:      wizardMessage(m.profile.ConversationLanguage, "projectName"),
			Kind:        initTUIStepText,
			Value:       m.profile.ProjectName,
			Description: description,
		}
	case wizardStepDevelopmentMode:
		return initTUIStepState{
			Title:   wizardMessage(m.profile.ConversationLanguage, "workMode"),
			Prompt:  wizardMessage(m.profile.ConversationLanguage, "workModePrompt"),
			Kind:    initTUIStepChoice,
			Choices: developmentModeOptionsForLanguage(m.profile.ConversationLanguage),
			Value:   m.profile.DevelopmentMode,
		}
	case wizardStepAgentMode:
		return initTUIStepState{
			Title:   wizardMessage(m.profile.ConversationLanguage, "agentStep"),
			Prompt:  wizardMessage(m.profile.ConversationLanguage, "agentPrompt"),
			Kind:    initTUIStepChoice,
			Choices: agentModeOptionsForLanguage(m.profile.ConversationLanguage),
			Value:   m.profile.AgentMode,
		}
	case wizardStepStatusLine:
		return initTUIStepState{
			Title:   wizardMessage(m.profile.ConversationLanguage, "statusStep"),
			Prompt:  wizardMessage(m.profile.ConversationLanguage, "statusPrompt"),
			Kind:    initTUIStepChoice,
			Choices: statusLineOptionsForLanguage(m.profile.ConversationLanguage),
			Value:   m.profile.StatusLinePreset,
		}
	case wizardStepCodexAccess:
		return m.currentCodexAccessStep()
	case wizardStepGitMode:
		return initTUIStepState{
			Title:       "Git",
			Prompt:      wizardMessage(m.profile.ConversationLanguage, "gitModePrompt"),
			Kind:        initTUIStepChoice,
			Choices:     gitModeOptionsForLanguage(m.profile.ConversationLanguage),
			Value:       m.profile.GitMode,
			Description: wizardMessage(m.profile.ConversationLanguage, "gitGuide"),
		}
	case wizardStepGitProvider:
		return initTUIStepState{
			Title:   "Git provider",
			Prompt:  wizardMessage(m.profile.ConversationLanguage, "gitProviderPrompt"),
			Kind:    initTUIStepChoice,
			Choices: gitProviderOptionsForLanguage(m.profile.ConversationLanguage),
			Value:   m.profile.GitProvider,
		}
	case wizardStepGitLabURL:
		return initTUIStepState{
			Title:  "GitLab URL",
			Prompt: wizardMessage(m.profile.ConversationLanguage, "gitLabURL"),
			Kind:   initTUIStepText,
			Value:  m.profile.GitLabInstanceURL,
		}
	default:
		return initTUIStepState{
			Title:  wizardMessage(m.profile.ConversationLanguage, "displayName"),
			Prompt: wizardMessage(m.profile.ConversationLanguage, "displayName"),
			Kind:   initTUIStepText,
			Value:  m.profile.UserName,
		}
	}
}

func (m initTUIModel) currentCodexAccessStep() initTUIStepState {
	switch m.codexPhase {
	case initTUICodexApproval:
		return initTUIStepState{
			Title:       "Codex access",
			Prompt:      "approval_policy",
			Kind:        initTUIStepChoice,
			Choices:     approvalPolicyOptions(),
			Value:       approvalPolicy(m.profile),
			Description: wizardMessage(m.profile.ConversationLanguage, "codexAccessGuide"),
		}
	case initTUICodexSandbox:
		return initTUIStepState{
			Title:       "Codex access",
			Prompt:      "sandbox_mode",
			Kind:        initTUIStepChoice,
			Choices:     sandboxModeOptions(),
			Value:       sandboxMode(m.profile),
			Description: wizardMessage(m.profile.ConversationLanguage, "codexAccessGuide"),
		}
	default:
		return initTUIStepState{
			Title:       "Codex access",
			Prompt:      "Codex access preset",
			Kind:        initTUIStepChoice,
			Choices:     codexAccessPresetOptions(),
			Value:       defaultCodexAccessPreset(m.profile),
			Description: wizardMessage(m.profile.ConversationLanguage, "codexAccessGuide"),
		}
	}
}

func (m *initTUIModel) moveSelection(delta int) {
	step := m.currentStep()
	if step.Kind != initTUIStepChoice || len(step.Choices) == 0 {
		return
	}
	m.selected = (m.selected + delta + len(step.Choices)) % len(step.Choices)
}

func (m *initTUIModel) back() {
	if m.step == wizardStepHumanLanguage {
		return
	}
	if m.step == wizardStepCodexAccess {
		switch m.codexPhase {
		case initTUICodexApproval:
			m.codexPhase = initTUICodexPreset
			m.resetInput()
			return
		case initTUICodexSandbox:
			m.codexPhase = initTUICodexApproval
			m.resetInput()
			return
		}
	}
	m.step = previousWizardStep(m.step, m.profile)
	m.codexPhase = initTUICodexPreset
	m.resetInput()
}

func (m *initTUIModel) commit() error {
	step := m.currentStep()
	switch step.Kind {
	case initTUIStepChoice:
		if len(step.Choices) == 0 {
			return nil
		}
		return m.applyChoice(step.Choices[m.selected].Value)
	case initTUIStepText:
		value := strings.TrimSpace(m.input)
		if value == "" {
			value = step.Value
		}
		m.applyText(value)
		m.advance()
		return nil
	default:
		return nil
	}
}

func (m *initTUIModel) applyChoice(value string) error {
	switch m.step {
	case wizardStepHumanLanguage:
		applyHumanLanguage(&m.profile, value)
	case wizardStepProjectType:
		m.profile.ProjectType = value
	case wizardStepDevelopmentMode:
		m.profile.DevelopmentMode = value
	case wizardStepAgentMode:
		m.profile.AgentMode = value
	case wizardStepStatusLine:
		m.profile.StatusLinePreset = value
	case wizardStepCodexAccess:
		return m.applyCodexChoice(value)
	case wizardStepGitMode:
		m.profile.GitMode = value
	case wizardStepGitProvider:
		m.profile.GitProvider = value
	default:
		return nil
	}
	m.advance()
	return nil
}

func (m *initTUIModel) applyCodexChoice(value string) error {
	switch m.codexPhase {
	case initTUICodexPreset:
		if value == codexAccessPresetCustom {
			m.codexPhase = initTUICodexApproval
			m.resetInput()
			return nil
		}
		if err := applyCodexAccessPreset(&m.profile, value); err != nil {
			return err
		}
		m.advance()
	case initTUICodexApproval:
		m.profile.ApprovalPolicy = value
		m.codexPhase = initTUICodexSandbox
		m.resetInput()
	case initTUICodexSandbox:
		m.profile.SandboxMode = value
		if _, err := resolveCodexAccessChoice(approvalPolicy(m.profile), sandboxMode(m.profile)); err != nil {
			return err
		}
		m.advance()
	}
	return nil
}

func (m *initTUIModel) applyText(value string) {
	switch m.step {
	case wizardStepProjectDetails:
		m.profile.ProjectName = value
		m.profile.Framework = normalizeFramework(m.profile.Framework)
		if m.profile.ProjectType != "existing" {
			m.profile.Language = firstNonBlank(m.profile.Language, "unknown")
		}
	case wizardStepGitLabURL:
		m.profile.GitLabInstanceURL = value
	case wizardStepDisplayName:
		m.profile.UserName = value
	}
}

func (m *initTUIModel) advance() {
	m.step = nextWizardStep(m.step, m.profile)
	m.codexPhase = initTUICodexPreset
	if m.step == wizardStepDone {
		m.done = true
		return
	}
	m.resetInput()
}

func (m *initTUIModel) resetInput() {
	step := m.currentStep()
	m.input = step.Value
	m.selected = 0
	for i, choice := range step.Choices {
		if choice.Value == step.Value {
			m.selected = i
			return
		}
	}
}

func defaultInitTUIStyles() initTUIStyles {
	return initTUIStyles{
		title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00A7A5")),
		subtle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#7A7F87")),
		active:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F0A202")),
		selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1D7F45")),
		status:   lipgloss.NewStyle().Foreground(lipgloss.Color("#3A5FCD")),
		warning:  lipgloss.NewStyle().Foreground(lipgloss.Color("#A05A00")),
	}
}
