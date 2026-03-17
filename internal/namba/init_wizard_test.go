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

func TestRenderInteractiveSelectUsesShortLocalizedHint(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	lines := renderInteractiveSelect(&out, "\U0001f9ea \uac1c\ubc1c \ubc29\ubc95\ub860", []option{
		{Value: "tdd", Label: "TDD", Description: "\uc0c8 \uae30\ub2a5 RED-GREEN"},
		{Value: "ddd", Label: "DDD", Description: "\uae30\uc874 \ucf54\ub4dc \ubd84\uc11d/\uac1c\uc120"},
	}, 1)
	if lines != 4 {
		t.Fatalf("renderInteractiveSelect lines = %d, want %d", lines, 4)
	}
	output := out.String()
	if !strings.Contains(output, "\u2191/\u2193 \uc774\ub3d9 \u00b7 Enter \uc120\ud0dd") {
		t.Fatalf("expected localized interactive hint, got %q", output)
	}
	if strings.Contains(output, "Use arrow keys") {
		t.Fatalf("expected long English hint to be removed, got %q", output)
	}
	if !strings.Contains(output, "\u276f 2. DDD - \uae30\uc874 \ucf54\ub4dc \ubd84\uc11d/\uac1c\uc120") {
		t.Fatalf("expected selected marker output, got %q", output)
	}
}
