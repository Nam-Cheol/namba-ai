package namba

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultJavaQualityCommands(t *testing.T) {
	t.Parallel()

	testCmd, lintCmd, typecheckCmd := defaultQualityCommands(t.TempDir(), "java", "maven")
	if testCmd != "mvn -q test" || lintCmd != "mvn -q spotless:check" || typecheckCmd != "mvn -q -DskipTests compile" {
		t.Fatalf("unexpected java quality commands: %q %q %q", testCmd, lintCmd, typecheckCmd)
	}
}

func TestDetectLanguageFrameworkJava(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "pom.xml"), []byte("<project/>"), 0o644); err != nil {
		t.Fatalf("write pom.xml: %v", err)
	}

	language, framework := detectLanguageFramework(root)
	if language != "java" || framework != "maven" {
		t.Fatalf("detectLanguageFramework() = (%q, %q), want (%q, %q)", language, framework, "java", "maven")
	}
}

func TestDetectProjectType(t *testing.T) {
	t.Parallel()

	newRoot := t.TempDir()
	if got := detectProjectType(newRoot); got != "new" {
		t.Fatalf("detectProjectType(empty) = %q, want %q", got, "new")
	}

	existingRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(existingRoot, "README.md"), []byte("# existing"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}
	if got := detectProjectType(existingRoot); got != "existing" {
		t.Fatalf("detectProjectType(non-empty) = %q, want %q", got, "existing")
	}
}

func TestFrameworkOptionsJava(t *testing.T) {
	t.Parallel()

	options := frameworkOptions("java")
	if len(options) != 4 {
		t.Fatalf("frameworkOptions(java) len = %d, want 4", len(options))
	}
	if options[1].Value != "maven" || options[2].Value != "gradle" || options[3].Value != "spring-boot" {
		t.Fatalf("unexpected java framework options: %+v", options)
	}
}

func TestProjectTypeOptions(t *testing.T) {
	t.Parallel()

	options := projectTypeOptions()
	if len(options) != 2 {
		t.Fatalf("projectTypeOptions len = %d, want 2", len(options))
	}
	if options[0].Value != "new" || options[1].Value != "existing" {
		t.Fatalf("unexpected project type options: %+v", options)
	}
}

func TestPromptProjectScaffoldUsesDetectedStackForExistingCode(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.stdin = strings.NewReader("\n")

	profile, back := app.promptProjectScaffold(bufio.NewReader(app.stdin), initProfile{
		ProjectName: "demo",
		ProjectType: "existing",
		Language:    "go",
		Framework:   "gin",
	}, true)

	if back {
		t.Fatal("did not expect back navigation")
	}
	if profile.Language != "go" || profile.Framework != "gin" {
		t.Fatalf("expected detected stack to be kept, got %+v", profile)
	}
	got := stdout.String()
	for _, want := range []string{"Detected codebase", "detected values are used"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected existing-code scaffold prompt to include %q, got %q", want, got)
		}
	}
	if strings.Contains(got, "직접 선택") || strings.Contains(got, "주 사용 언어") {
		t.Fatalf("existing-code flow should not ask for early language/framework override, got %q", got)
	}
}

func TestPromptProjectScaffoldNewProjectSkipsStarterStack(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.stdin = strings.NewReader("\n")

	profile, back := app.promptProjectScaffold(bufio.NewReader(app.stdin), initProfile{
		ProjectName: "demo",
		ProjectType: "new",
		Language:    "unknown",
		Framework:   "none",
	}, true)

	if back {
		t.Fatal("did not expect back navigation")
	}
	if profile.Language != "unknown" || profile.Framework != "none" {
		t.Fatalf("expected empty repository to keep stack unselected, got %+v", profile)
	}
	got := stdout.String()
	for _, want := range []string{"do not choose an app stack", "first `namba plan`"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected new-project scaffold prompt to include %q, got %q", want, got)
		}
	}
	if strings.Contains(got, "시작할 앱 스택") || strings.Contains(got, "Next.js") {
		t.Fatalf("new-project flow should not ask for starter stack, got %q", got)
	}
}

func TestPromptProjectScaffoldSupportsBackFromProjectName(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.stdin = strings.NewReader("back\n")

	_, back := app.promptProjectScaffold(bufio.NewReader(app.stdin), initProfile{
		ProjectName: "demo",
		ProjectType: "new",
		Language:    "unknown",
		Framework:   "none",
	}, true)

	if !back {
		t.Fatal("expected back navigation from project name")
	}
	if got := stdout.String(); !strings.Contains(got, "previous step") {
		t.Fatalf("expected back hint in output, got %q", got)
	}
}

func TestParseInitArgsProjectType(t *testing.T) {
	t.Parallel()

	opts, err := parseInitArgs([]string{".", "--project-type", "existing"})
	if err != nil {
		t.Fatalf("parseInitArgs returned error: %v", err)
	}
	if opts.ProjectType != "existing" {
		t.Fatalf("opts.ProjectType = %q, want %q", opts.ProjectType, "existing")
	}
}

func TestParseInitArgsHumanLanguageAndSandbox(t *testing.T) {
	t.Parallel()

	opts, err := parseInitArgs([]string{".", "--human-language", "ko", "--approval-policy", "never", "--sandbox-mode", "read-only"})
	if err != nil {
		t.Fatalf("parseInitArgs returned error: %v", err)
	}
	if opts.HumanLanguage != "ko" || opts.ApprovalPolicy != "never" || opts.SandboxMode != "read-only" {
		t.Fatalf("unexpected init options: %+v", opts)
	}
}

func TestApplyHumanLanguageSyncsAllHumanFacingOutputs(t *testing.T) {
	t.Parallel()

	profile := initProfile{
		ConversationLanguage:  "en",
		DocumentationLanguage: "en",
		CommentLanguage:       "en",
		PRLanguage:            "en",
	}
	applyHumanLanguage(&profile, "ko")

	if profile.ConversationLanguage != "ko" || profile.DocumentationLanguage != "ko" || profile.CommentLanguage != "ko" || profile.PRLanguage != "ko" {
		t.Fatalf("expected human language sync, got %+v", profile)
	}
}

func TestDetectLocalePriorityForInitLanguageDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{name: "NAMBA_LANG wins", env: map[string]string{"NAMBA_LANG": "ja", "LC_ALL": "ko_KR.UTF-8", "LANG": "en_US.UTF-8"}, want: "ja"},
		{name: "LC_ALL wins over LANG", env: map[string]string{"LC_ALL": "zh_CN.UTF-8", "LANG": "ko_KR.UTF-8"}, want: "zh"},
		{name: "LANG fallback", env: map[string]string{"LANG": "ko_KR.UTF-8"}, want: "ko"},
		{name: "unknown fallback en", env: map[string]string{"LANG": "fr_FR.UTF-8"}, want: "en"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := detectLocale(func(key string) string { return tt.env[key] })
			if got != tt.want {
				t.Fatalf("detectLocale() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLanguageOptionsUseStableTextLabelsWithoutFlags(t *testing.T) {
	t.Parallel()

	got := strings.TrimSpace(func() string {
		var labels []string
		for _, option := range languageOptions() {
			labels = append(labels, option.Label)
		}
		return strings.Join(labels, "\n")
	}())

	for _, want := range []string{"[ko] Korean", "[en] English", "[ja] Japanese", "[zh] Simplified Chinese"} {
		if !strings.Contains(got, want) {
			t.Fatalf("language labels missing %q: %q", want, got)
		}
	}
	for _, flag := range []string{"🇰🇷", "🇺🇸", "🇯🇵", "🇨🇳"} {
		if strings.Contains(got, flag) {
			t.Fatalf("language labels must not contain country flags: %q", got)
		}
	}
}

func TestRunInitWizardStartsWithLanguageBeforeRepositoryState(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.stdin = strings.NewReader(strings.Repeat("\n", 16))

	defaults := initProfile{
		ProjectName:           "demo",
		ProjectType:           "existing",
		Language:              "go",
		Framework:             "cobra",
		DevelopmentMode:       "tdd",
		ConversationLanguage:  "ja",
		DocumentationLanguage: "ja",
		CommentLanguage:       "ja",
		ApprovalPolicy:        "on-request",
		SandboxMode:           "workspace-write",
		GitMode:               "manual",
		GitProvider:           "github",
		GitLabInstanceURL:     "https://gitlab.com",
		BranchPerWork:         true,
		BranchBase:            "main",
		SpecBranchPrefix:      "spec/",
		TaskBranchPrefix:      "task/",
		PRBaseBranch:          "main",
		PRLanguage:            "ja",
		CodexReviewComment:    "@codex review",
		AgentMode:             "single",
		StatusLinePreset:      "namba",
		UserName:              "Developer",
		CreatedAt:             "2026-04-23T10:00:00Z",
	}

	if _, err := app.runInitWizard(defaults); err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	got := stdout.String()
	languageIndex := strings.Index(got, "Step 01 ·")
	repoIndex := strings.Index(got, "Repository intelligence")
	if languageIndex < 0 || repoIndex < 0 || languageIndex > repoIndex {
		t.Fatalf("expected language step before repository intelligence, got %q", got)
	}
	for _, want := range []string{"[default] [ja] Japanese", "Language / 언어 / 言語 / 语言", "[current]", "[next]", "Target: demo", "Detected stack: go / cobra"} {
		if !strings.Contains(got, want) {
			t.Fatalf("wizard output missing %q: %q", want, got)
		}
	}
	for _, marker := range []string{"🇰🇷", "🇺🇸", "🇯🇵", "🇨🇳", "✅", "🧪", "🔐", "🙋", "🧭"} {
		if strings.Contains(got, marker) {
			t.Fatalf("plain wizard output must not contain fragile emoji marker %q: %q", marker, got)
		}
	}
}

func TestInitHelpAndGeneratedGettingStartedMentionLanguageFirstPlainFallback(t *testing.T) {
	t.Parallel()

	help := initUsageText()
	for _, want := range []string{"language-first screen", "Plain terminals", "[default]", "[recommended]"} {
		if !strings.Contains(help, want) {
			t.Fatalf("init help missing %q: %q", want, help)
		}
	}

	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		doc := strings.Join(renderNambaCLIGettingStartedBootstrapSection(lang), "\n")
		for _, want := range []string{"[ko] Korean", "[en] English", "[ja] Japanese", "[zh] Simplified Chinese", "[default]", "[recommended]", "[current]", "[next]"} {
			if !strings.Contains(doc, want) {
				t.Fatalf("%s getting-started bootstrap section missing %q: %q", lang, want, doc)
			}
		}
	}
}

func TestOutputContractSpecLocalizedByLanguage(t *testing.T) {
	t.Parallel()

	ko := outputContractSpecFor(initProfile{ConversationLanguage: "ko"})
	if ko.Header != "NAMBA-AI 작업 결과 보고" || ko.Sections[0].Primary != "작업 정의" || ko.Sections[5].Primary != "다음에 해야 할 작업" {
		t.Fatalf("unexpected Korean output contract spec: %+v", ko)
	}

	en := outputContractSpecFor(initProfile{ConversationLanguage: "en"})
	if en.Header != "NAMBA-AI Work Report" || en.Sections[0].Primary != "Scope" || en.Sections[5].Primary != "Next Work" {
		t.Fatalf("unexpected English output contract spec: %+v", en)
	}
}

func TestRenderProjectConfigIncludesProjectType(t *testing.T) {
	t.Parallel()

	body := renderProjectConfig(initProfile{
		ProjectName: "demo",
		ProjectType: "existing",
		Language:    "go",
		Framework:   "none",
		CreatedAt:   "2026-03-16 10:00:00",
	})
	if !strings.Contains(body, "project_type: existing\n") {
		t.Fatalf("renderProjectConfig() missing project_type: %q", body)
	}
}

func TestReadMenuAction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []byte
		want  menuAction
	}{
		{name: "ansi up", input: []byte{0x1b, '[', 'A'}, want: menuActionUp},
		{name: "ansi down", input: []byte{0x1b, '[', 'B'}, want: menuActionDown},
		{name: "windows up", input: []byte{0xe0, 72}, want: menuActionUp},
		{name: "windows down", input: []byte{0xe0, 80}, want: menuActionDown},
		{name: "enter", input: []byte{'\r'}, want: menuActionSubmit},
		{name: "back", input: []byte{'b'}, want: menuActionBack},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			action, err := readMenuAction(bufio.NewReaderSize(bytes.NewReader(tt.input), len(tt.input)))
			if err != nil {
				t.Fatalf("readMenuAction returned error: %v", err)
			}
			if action != tt.want {
				t.Fatalf("readMenuAction(%v) = %v, want %v", tt.input, action, tt.want)
			}
		})
	}
}

func TestPromptSelectLineUsesKoreanPrompt(t *testing.T) {
	t.Parallel()

	reader := bufio.NewReader(strings.NewReader("\n"))
	var out bytes.Buffer
	got := promptSelectLine(reader, &out, "\U0001f9ea \uac1c\ubc1c \ubc29\ubc95\ub860", []option{
		{Value: "tdd", Label: "TDD", Description: "\uc0c8 \uae30\ub2a5 RED-GREEN"},
		{Value: "ddd", Label: "DDD", Description: "\uae30\uc874 \ucf54\ub4dc \ubd84\uc11d/\uac1c\uc120"},
	}, "ddd")
	if got != "ddd" {
		t.Fatalf("promptSelectLine default = %q, want %q", got, "ddd")
	}
	output := out.String()
	if !strings.Contains(output, "\uc120\ud0dd [2]:") {
		t.Fatalf("expected Korean select prompt, got %q", output)
	}
	if strings.Contains(output, "Select [2]:") {
		t.Fatalf("expected English prompt to be removed, got %q", output)
	}
}

func TestPromptSelectWizardResultEchoesSelectionAndSupportsBack(t *testing.T) {
	t.Parallel()

	var selectedOut bytes.Buffer
	reader := bufio.NewReader(strings.NewReader("2\n"))
	value, back := promptSelectWizardResult(strings.NewReader(""), reader, &selectedOut, "\U0001f310 \uc791\uc5c5 \uc5b8\uc5b4", languageOptions(), "ko", true)
	if back || value != "en" {
		t.Fatalf("promptSelectWizardResult() = (%q, %v), want en,false", value, back)
	}
	if got := selectedOut.String(); !strings.Contains(got, "[en] English") || !strings.Contains(got, "b. back") {
		t.Fatalf("expected echoed selection and back option, got %q", got)
	}

	var backOut bytes.Buffer
	reader = bufio.NewReader(strings.NewReader("back\n"))
	value, back = promptSelectWizardResult(strings.NewReader(""), reader, &backOut, "\U0001f310 \uc791\uc5c5 \uc5b8\uc5b4", languageOptions(), "ko", true)
	if !back || value != "ko" {
		t.Fatalf("promptSelectWizardResult(back) = (%q, %v), want ko,true", value, back)
	}
	if got := backOut.String(); !strings.Contains(got, "previous step") {
		t.Fatalf("expected back navigation output, got %q", got)
	}
}

func TestWizardVisibleStepNumberSkipsConditionalGitSteps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		profile initProfile
		want    int
	}{
		{
			name:    "manual git skips provider steps",
			profile: initProfile{GitMode: "manual"},
			want:    9,
		},
		{
			name:    "github git skips gitlab url",
			profile: initProfile{GitMode: "personal", GitProvider: "github"},
			want:    10,
		},
		{
			name:    "gitlab git includes instance url",
			profile: initProfile{GitMode: "team", GitProvider: "gitlab"},
			want:    11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := wizardVisibleStepNumber(wizardStepDisplayName, tt.profile); got != tt.want {
				t.Fatalf("wizardVisibleStepNumber(display name) = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRenderInteractiveSelectUsesShortLocalizedHint(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	lines := renderInteractiveSelect(&out, "\U0001f9ea \uac1c\ubc1c \ubc29\ubc95\ub860", []option{
		{Value: "tdd", Label: "TDD", Description: "\uc0c8 \uae30\ub2a5 RED-GREEN"},
		{Value: "ddd", Label: "DDD", Description: "\uae30\uc874 \ucf54\ub4dc \ubd84\uc11d/\uac1c\uc120"},
	}, 1, true)
	if lines != 4 {
		t.Fatalf("renderInteractiveSelect lines = %d, want %d", lines, 4)
	}
	output := out.String()
	if !strings.Contains(output, "\u2191/\u2193 move \u00b7 Enter select") {
		t.Fatalf("expected concise interactive hint, got %q", output)
	}
	if !strings.Contains(output, "b back") {
		t.Fatalf("expected back navigation hint, got %q", output)
	}
	if strings.Contains(output, "Use arrow keys") {
		t.Fatalf("expected long English hint to be removed, got %q", output)
	}
	if !strings.Contains(output, "2. DDD - \uae30\uc874 \ucf54\ub4dc \ubd84\uc11d/\uac1c\uc120") {
		t.Fatalf("expected selected marker output, got %q", output)
	}
}
