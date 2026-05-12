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
	for _, want := range []string{"감지된 코드베이스", "감지값을 사용"} {
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
	for _, want := range []string{"앱 스택을 묻지 않습니다", "첫 `namba plan`"} {
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
	if got := stdout.String(); !strings.Contains(got, "이전 단계") {
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

func TestOutputContractSpecLocalizedByLanguage(t *testing.T) {
	t.Parallel()

	ko := outputContractSpecFor(initProfile{ConversationLanguage: "ko"})
	if ko.Header != "NAMBA-AI 작업 결과 보고" || ko.Sections[0].Primary != "작업 정의" || ko.Sections[5].Primary != "다음 스텝" {
		t.Fatalf("unexpected Korean output contract spec: %+v", ko)
	}

	en := outputContractSpecFor(initProfile{ConversationLanguage: "en"})
	if en.Header != "NAMBA-AI Work Report" || en.Sections[0].Primary != "Scope" || en.Sections[5].Primary != "Next Steps" {
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
	if got := selectedOut.String(); !strings.Contains(got, "✅") || !strings.Contains(got, "\U0001f1fa\U0001f1f8 \uc601\uc5b4") || !strings.Contains(got, "b. \uc774\uc804") {
		t.Fatalf("expected echoed selection and back option, got %q", got)
	}

	var backOut bytes.Buffer
	reader = bufio.NewReader(strings.NewReader("back\n"))
	value, back = promptSelectWizardResult(strings.NewReader(""), reader, &backOut, "\U0001f310 \uc791\uc5c5 \uc5b8\uc5b4", languageOptions(), "ko", true)
	if !back || value != "ko" {
		t.Fatalf("promptSelectWizardResult(back) = (%q, %v), want ko,true", value, back)
	}
	if got := backOut.String(); !strings.Contains(got, "\uc774\uc804 \ub2e8\uacc4") {
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
	if !strings.Contains(output, "\u2191/\u2193 \uc774\ub3d9 \u00b7 Enter \uc120\ud0dd") {
		t.Fatalf("expected localized interactive hint, got %q", output)
	}
	if !strings.Contains(output, "b \uc774\uc804") {
		t.Fatalf("expected back navigation hint, got %q", output)
	}
	if strings.Contains(output, "Use arrow keys") {
		t.Fatalf("expected long English hint to be removed, got %q", output)
	}
	if !strings.Contains(output, "\U0001f449 2. DDD - \uae30\uc874 \ucf54\ub4dc \ubd84\uc11d/\uac1c\uc120") {
		t.Fatalf("expected selected marker output, got %q", output)
	}
}
