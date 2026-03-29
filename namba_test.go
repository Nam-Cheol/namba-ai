package namba_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Nam-Cheol/namba-ai/internal/namba"
)

func TestInitCreatesScaffold(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := app.Run(context.Background(), []string{"init", tmp}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	mustExist(t, filepath.Join(tmp, "AGENTS.md"))
	mustExist(t, filepath.Join(tmp, ".agents", "skills", "namba", "SKILL.md"))
	mustExist(t, filepath.Join(tmp, ".agents", "skills", "namba-run", "SKILL.md"))
	mustExist(t, filepath.Join(tmp, ".agents", "skills", "namba-pr", "SKILL.md"))
	mustExist(t, filepath.Join(tmp, ".agents", "skills", "namba-land", "SKILL.md"))
	mustExist(t, filepath.Join(tmp, ".agents", "skills", "namba-project", "SKILL.md"))
	mustExist(t, filepath.Join(tmp, ".agents", "skills", "namba-foundation-core", "SKILL.md"))
	mustExist(t, filepath.Join(tmp, ".agents", "skills", "namba-workflow-init", "SKILL.md"))
	mustExist(t, filepath.Join(tmp, ".codex", "config.toml"))
	mustExist(t, filepath.Join(tmp, ".codex", "agents", "namba-planner.md"))
	mustExist(t, filepath.Join(tmp, ".codex", "agents", "namba-planner.toml"))
	mustExist(t, filepath.Join(tmp, ".codex", "agents", "namba-implementer.md"))
	mustExist(t, filepath.Join(tmp, ".codex", "agents", "namba-implementer.toml"))
	mustExist(t, filepath.Join(tmp, ".codex", "agents", "namba-reviewer.md"))
	mustExist(t, filepath.Join(tmp, ".codex", "agents", "namba-reviewer.toml"))
	mustExist(t, filepath.Join(tmp, ".namba", "config", "sections", "project.yaml"))
	mustExist(t, filepath.Join(tmp, ".namba", "config", "sections", "language.yaml"))
	mustExist(t, filepath.Join(tmp, ".namba", "config", "sections", "git-strategy.yaml"))
	mustExist(t, filepath.Join(tmp, ".namba", "config", "sections", "codex.yaml"))
	mustExist(t, filepath.Join(tmp, ".namba", "config", "sections", "docs.yaml"))
	mustExist(t, filepath.Join(tmp, ".namba", "codex", "claude-codex-mapping.md"))
	mustExist(t, filepath.Join(tmp, ".namba", "codex", "output-contract.md"))
	mustExist(t, filepath.Join(tmp, ".namba", "codex", "validate-output-contract.py"))
	mustExist(t, filepath.Join(tmp, ".namba", "manifest.json"))
	mustExist(t, filepath.Join(tmp, "README.md"))
	mustExist(t, filepath.Join(tmp, "README.ko.md"))
	mustExist(t, filepath.Join(tmp, "docs", "getting-started.md"))
	mustExist(t, filepath.Join(tmp, "docs", "workflow-guide.md"))

	agents := mustRead(t, filepath.Join(tmp, "AGENTS.md"))
	if !strings.Contains(agents, "NAMBA-AI Work Report") || !strings.Contains(agents, "🧭 Scope") || !strings.Contains(agents, "validate-output-contract.py") {
		t.Fatalf("expected AGENTS to describe the Namba output contract, got: %s", agents)
	}
	if strings.Contains(agents, "Until Codex exposes a documented stop-hook surface") {
		t.Fatalf("expected AGENTS to avoid stale stop-hook wording, got: %s", agents)
	}

	gettingStarted := mustRead(t, filepath.Join(tmp, "docs", "getting-started.md"))
	if !strings.Contains(gettingStarted, "WSL workspace") {
		t.Fatalf("expected getting started guide to describe current Windows WSL guidance, got: %s", gettingStarted)
	}

	plannerAgent := mustRead(t, filepath.Join(tmp, ".codex", "agents", "namba-planner.toml"))
	if !strings.Contains(plannerAgent, `name = "namba-planner"`) || !strings.Contains(plannerAgent, `developer_instructions = """`) {
		t.Fatalf("unexpected planner custom agent: %s", plannerAgent)
	}
	if strings.Contains(plannerAgent, `prompt = "`) {
		t.Fatalf("expected planner custom agent to use developer_instructions, got: %s", plannerAgent)
	}
	mustNotExist(t, filepath.Join(tmp, ".codex", "skills"))

	validator := mustRead(t, filepath.Join(tmp, ".namba", "codex", "validate-output-contract.py"))
	if !strings.Contains(validator, "output-contract: ok") || !strings.Contains(validator, "Scope") || !strings.Contains(validator, "missing header") || !strings.Contains(validator, "start=previous + 1") {
		t.Fatalf("expected output contract validator script, got: %s", validator)
	}

	readme := mustRead(t, filepath.Join(tmp, "README.md"))
	if !strings.Contains(readme, "This repository is managed with NambaAI") || !strings.Contains(readme, "Workflow Guide") {
		t.Fatalf("expected generated README bundle, got: %s", readme)
	}
}

func TestInitSupportsCodexProfileFlags(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	args := []string{
		"init",
		tmp,
		"--yes",
		"--name", "example",
		"--mode", "ddd",
		"--language", "go",
		"--framework", "cobra",
		"--human-language", "ko",
		"--approval-policy", "never",
		"--sandbox-mode", "read-only",
		"--git-mode", "team",
		"--git-provider", "github",
		"--git-username", "alice",
		"--agent-mode", "multi",
		"--statusline", "off",
		"--user-name", "Alice",
	}

	if err := app.Run(context.Background(), args); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	project := mustRead(t, filepath.Join(tmp, ".namba", "config", "sections", "project.yaml"))
	if !strings.Contains(project, "name: example") || !strings.Contains(project, "framework: cobra") {
		t.Fatalf("unexpected project config: %s", project)
	}

	language := mustRead(t, filepath.Join(tmp, ".namba", "config", "sections", "language.yaml"))
	if !strings.Contains(language, "conversation_language: ko") || !strings.Contains(language, "documentation_language: ko") || !strings.Contains(language, "comment_language: ko") {
		t.Fatalf("unexpected language config: %s", language)
	}

	gitStrategy := mustRead(t, filepath.Join(tmp, ".namba", "config", "sections", "git-strategy.yaml"))
	if !strings.Contains(gitStrategy, "git_mode: team") || !strings.Contains(gitStrategy, "git_username: alice") || !strings.Contains(gitStrategy, "store_tokens: false") {
		t.Fatalf("unexpected git strategy config: %s", gitStrategy)
	}
	if !strings.Contains(gitStrategy, "branch_per_work: true") || !strings.Contains(gitStrategy, "branch_base: main") || !strings.Contains(gitStrategy, "pr_base_branch: main") || !strings.Contains(gitStrategy, `codex_review_comment: "@codex review"`) || !strings.Contains(gitStrategy, "pr_language: ko") {
		t.Fatalf("expected git collaboration defaults in git strategy config: %s", gitStrategy)
	}

	codexProfile := mustRead(t, filepath.Join(tmp, ".namba", "config", "sections", "codex.yaml"))
	if !strings.Contains(codexProfile, "agent_mode: multi") || !strings.Contains(codexProfile, "status_line_preset: off") {
		t.Fatalf("unexpected codex config: %s", codexProfile)
	}
	if strings.Contains(codexProfile, "compat_skills_path:") {
		t.Fatalf("expected deprecated compat skill path to be removed: %s", codexProfile)
	}

	system := mustRead(t, filepath.Join(tmp, ".namba", "config", "sections", "system.yaml"))
	if !strings.Contains(system, "approval_policy: never") || !strings.Contains(system, "sandbox_mode: read-only") {
		t.Fatalf("unexpected system config: %s", system)
	}

	codexConfig := mustRead(t, filepath.Join(tmp, ".codex", "config.toml"))
	if !strings.Contains(codexConfig, "#:schema https://developers.openai.com/codex/config-schema.json") || !strings.Contains(codexConfig, "repo-safe Codex defaults under version control") || !strings.Contains(codexConfig, "max_threads = 3") || !strings.Contains(codexConfig, `approval_policy = "never"`) || !strings.Contains(codexConfig, `sandbox_mode = "read-only"`) {
		t.Fatalf("unexpected codex repo config: %s", codexConfig)
	}
	if strings.Contains(codexConfig, "status_line") {
		t.Fatalf("expected status line to be omitted when preset is off: %s", codexConfig)
	}

	agents := mustRead(t, filepath.Join(tmp, "AGENTS.md"))
	if !strings.Contains(agents, "NAMBA-AI 작업 결과 보고") || !strings.Contains(agents, "🧭 작업 정의") || !strings.Contains(agents, "➡ 다음 스텝") {
		t.Fatalf("expected localized Korean output contract in AGENTS, got: %s", agents)
	}

	docsCfg := mustRead(t, filepath.Join(tmp, ".namba", "config", "sections", "docs.yaml"))
	if !strings.Contains(docsCfg, "manage_readme: true") || !strings.Contains(docsCfg, "readme_profile: managed-project") {
		t.Fatalf("unexpected docs config: %s", docsCfg)
	}
}

func TestRepoCodexUpstreamReferenceUsesLiveDocs(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	upstreamReference := mustRead(t, filepath.Join(filepath.Dir(file), "docs", "codex-upstream-reference.md"))
	if !strings.Contains(upstreamReference, "https://developers.openai.com/codex/") || !strings.Contains(upstreamReference, "https://github.com/openai/codex") || !strings.Contains(upstreamReference, "supplemental implementation reference") {
		t.Fatalf("expected codex upstream reference doc to use live docs as the primary baseline, got: %s", upstreamReference)
	}
}

func TestInitRejectsUnsupportedMode(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	err := app.Run(context.Background(), []string{"init", tmp, "--yes", "--mode", "hybrid"})
	if err == nil {
		t.Fatal("expected invalid mode error")
	}
	if !strings.Contains(err.Error(), "development mode") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoctorReportsCodexNativeReadiness(t *testing.T) {
	tmp := t.TempDir()
	stdout := &bytes.Buffer{}
	app := namba.NewApp(stdout, &bytes.Buffer{})

	if err := app.Run(context.Background(), []string{"init", tmp}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"doctor"}); err != nil {
		t.Fatalf("doctor failed: %v", err)
	}

	text := stdout.String()
	if !strings.Contains(text, "Codex native repo: ready") {
		t.Fatalf("expected codex native readiness in doctor output: %s", text)
	}
	if strings.Contains(text, "Codex compatibility mirror") {
		t.Fatalf("expected compatibility mirror line to be removed, got: %s", text)
	}
}

func TestPlanCreatesSequentialSpecs(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := app.Run(context.Background(), []string{"init", tmp}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"plan", "build", "status", "page"}); err != nil {
		t.Fatalf("first plan failed: %v", err)
	}
	if err := app.Run(context.Background(), []string{"plan", "add", "sync", "report"}); err != nil {
		t.Fatalf("second plan failed: %v", err)
	}

	mustExist(t, filepath.Join(tmp, ".namba", "specs", "SPEC-001", "spec.md"))
	mustExist(t, filepath.Join(tmp, ".namba", "specs", "SPEC-002", "spec.md"))
}

func TestProjectRunDryRunAndSync(t *testing.T) {
	tmp := t.TempDir()
	stdout := &bytes.Buffer{}
	app := namba.NewApp(stdout, &bytes.Buffer{})

	if err := app.Run(context.Background(), []string{"init", tmp}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"project"}); err != nil {
		t.Fatalf("project failed: %v", err)
	}
	if err := app.Run(context.Background(), []string{"plan", "implement", "healthcheck"}); err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	if err := app.Run(context.Background(), []string{"run", "SPEC-001", "--dry-run"}); err != nil {
		t.Fatalf("run dry-run failed: %v", err)
	}
	if err := app.Run(context.Background(), []string{"sync"}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	product, err := os.ReadFile(filepath.Join(tmp, ".namba", "project", "product.md"))
	if err != nil {
		t.Fatalf("read product doc: %v", err)
	}
	if !strings.Contains(string(product), "This repository is managed with NambaAI") {
		t.Fatalf("expected README content in product doc, got: %s", product)
	}

	mustExist(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-request.md"))
	mustExist(t, filepath.Join(tmp, ".namba", "project", "pr-checklist.md"))
}

func TestProjectGeneratesReactCodemaps(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := os.MkdirAll(filepath.Join(tmp, "src", "app"), 0o755); err != nil {
		t.Fatalf("mkdir src/app: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "package.json"), []byte(`{
  "name": "aura",
  "dependencies": {
    "react-router": "7.13.0",
    "@mui/material": "7.3.5",
    "@emotion/react": "11.14.0",
    "motion": "12.23.24",
    "sonner": "2.0.3",
    "@radix-ui/react-dialog": "1.1.6"
  },
  "devDependencies": {
    "vite": "6.3.5",
    "@vitejs/plugin-react": "4.7.0",
    "tailwindcss": "4.1.12"
  },
  "peerDependencies": {
    "react": "18.3.1",
    "react-dom": "18.3.1"
  }
}`), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "src", "main.tsx"), []byte(`import { createRoot } from "react-dom/client";
import App from "./app/App";

createRoot(document.getElementById("root")!).render(<App />);
`), 0o644); err != nil {
		t.Fatalf("write src/main.tsx: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "src", "app", "App.tsx"), []byte(`import { RouterProvider } from "react-router";
import { router } from "./routes";

export default function App() {
  return <RouterProvider router={router} />;
}
`), 0o644); err != nil {
		t.Fatalf("write src/app/App.tsx: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "src", "app", "routes.ts"), []byte(`import { createBrowserRouter } from "react-router";

export const router = createBrowserRouter([]);
`), 0o644); err != nil {
		t.Fatalf("write src/app/routes.ts: %v", err)
	}

	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"project"}); err != nil {
		t.Fatalf("project failed: %v", err)
	}

	entryPoints := mustRead(t, filepath.Join(tmp, ".namba", "project", "codemaps", "entry-points.md"))
	if !strings.Contains(entryPoints, "`src/main.tsx`") || !strings.Contains(entryPoints, "`src/app/App.tsx`") || !strings.Contains(entryPoints, "`src/app/routes.ts`") {
		t.Fatalf("expected React entry points, got: %s", entryPoints)
	}
	if strings.Contains(entryPoints, "cmd/namba/main.go") {
		t.Fatalf("expected project entry points instead of Namba CLI defaults, got: %s", entryPoints)
	}

	deps := mustRead(t, filepath.Join(tmp, ".namba", "project", "codemaps", "dependencies.md"))
	if !strings.Contains(deps, "react@18.3.1") || !strings.Contains(deps, "react-router@7.13.0") || !strings.Contains(deps, "vite@6.3.5") {
		t.Fatalf("expected package.json-driven dependency summary, got: %s", deps)
	}
	if !strings.Contains(deps, "Radix UI primitives") {
		t.Fatalf("expected grouped UI dependency summary, got: %s", deps)
	}
}

func TestProjectIgnoresNestedNodeModulesDuringCodemapDiscovery(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := os.MkdirAll(filepath.Join(tmp, "nested", "node_modules", "vendor"), 0o755); err != nil {
		t.Fatalf("mkdir nested node_modules: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "src", "router"), 0o755); err != nil {
		t.Fatalf("mkdir src/router: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "package.json"), []byte(`{"name":"aura","dependencies":{"react":"18.3.1","react-dom":"18.3.1","react-router":"7.13.0"}}`), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "src", "main.tsx"), []byte(`import { createRoot } from "react-dom/client";
import App from "./App";

createRoot(document.getElementById("root")!).render(<App />);
`), 0o644); err != nil {
		t.Fatalf("write src/main.tsx: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "src", "App.tsx"), []byte(`import { RouterProvider } from "react-router";
import { router } from "./router/app-router";

export default function App() {
  return <RouterProvider router={router} />;
}
`), 0o644); err != nil {
		t.Fatalf("write src/App.tsx: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "src", "router", "app-router.ts"), []byte(`import { createBrowserRouter } from "react-router";

export const router = createBrowserRouter([]);
`), 0o644); err != nil {
		t.Fatalf("write real router: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "nested", "node_modules", "vendor", "router.ts"), []byte(`import { createBrowserRouter } from "react-router";

export const vendorRouter = createBrowserRouter([]);
`), 0o644); err != nil {
		t.Fatalf("write vendor router: %v", err)
	}

	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"project"}); err != nil {
		t.Fatalf("project failed: %v", err)
	}

	entryPoints := mustRead(t, filepath.Join(tmp, ".namba", "project", "codemaps", "entry-points.md"))
	if !strings.Contains(entryPoints, "`src/router/app-router.ts`") {
		t.Fatalf("expected project router entry point, got: %s", entryPoints)
	}
	if strings.Contains(entryPoints, "`nested/node_modules/vendor/router.ts`") {
		t.Fatalf("expected nested node_modules to be skipped, got: %s", entryPoints)
	}
}

func TestProjectUsesRenderTargetForReactAppShell(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := os.MkdirAll(filepath.Join(tmp, "src", "app"), 0o755); err != nil {
		t.Fatalf("mkdir src/app: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "package.json"), []byte(`{"name":"aura","dependencies":{"react":"18.3.1","react-dom":"18.3.1","react-router":"7.13.0"}}`), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "src", "main.tsx"), []byte(`import "./telemetry";
import { createRoot } from "react-dom/client";
import App from "./app/App";

createRoot(document.getElementById("root")!).render(<App />);
`), 0o644); err != nil {
		t.Fatalf("write src/main.tsx: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "src", "telemetry.ts"), []byte(`console.log("boot telemetry");
`), 0o644); err != nil {
		t.Fatalf("write telemetry: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "src", "app", "App.tsx"), []byte(`import { RouterProvider } from "react-router";
import { router } from "./routes";

export default function App() {
  return <RouterProvider router={router} />;
}
`), 0o644); err != nil {
		t.Fatalf("write src/app/App.tsx: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "src", "app", "routes.ts"), []byte(`import { createBrowserRouter } from "react-router";

export const router = createBrowserRouter([]);
`), 0o644); err != nil {
		t.Fatalf("write src/app/routes.ts: %v", err)
	}

	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"project"}); err != nil {
		t.Fatalf("project failed: %v", err)
	}

	entryPoints := mustRead(t, filepath.Join(tmp, ".namba", "project", "codemaps", "entry-points.md"))
	if !strings.Contains(entryPoints, "`src/app/App.tsx`") || !strings.Contains(entryPoints, "`src/app/routes.ts`") {
		t.Fatalf("expected render target app shell and router module, got: %s", entryPoints)
	}
	if strings.Contains(entryPoints, "`src/telemetry.ts`") {
		t.Fatalf("expected side-effect bootstrap imports to be ignored, got: %s", entryPoints)
	}
}

func TestSyncRefreshesWorkflowDocs(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	stdout := &bytes.Buffer{}
	app := namba.NewApp(stdout, &bytes.Buffer{})

	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"plan", "sync", "workflow", "docs"}); err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	if err := app.Run(context.Background(), []string{"sync"}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	changeSummary := mustRead(t, filepath.Join(tmp, ".namba", "project", "change-summary.md"))
	if !strings.Contains(changeSummary, "`namba update`") || !strings.Contains(changeSummary, "`namba pr`") || !strings.Contains(changeSummary, "`namba land`") || !strings.Contains(changeSummary, "`namba run SPEC-XXX --parallel`") || !strings.Contains(changeSummary, "`@codex review`") {
		t.Fatalf("expected synced change summary to describe update and parallel workflow, got: %s", changeSummary)
	}

	prChecklist := mustRead(t, filepath.Join(tmp, ".namba", "project", "pr-checklist.md"))
	if !strings.Contains(prChecklist, "PR targets `main`") || !strings.Contains(prChecklist, "`@codex review` review request is present on GitHub") {
		t.Fatalf("expected synced PR checklist to describe branch and review policy, got: %s", prChecklist)
	}

	releaseNotes := mustRead(t, filepath.Join(tmp, ".namba", "project", "release-notes.md"))
	if !strings.Contains(releaseNotes, "`namba release --push`") || !strings.Contains(releaseNotes, "`checksums.txt`") || !strings.Contains(releaseNotes, "`@codex review`") || !strings.Contains(releaseNotes, "`namba_Windows_x86.zip`") {
		t.Fatalf("expected synced release notes to describe release flow, got: %s", releaseNotes)
	}

	releaseChecklist := mustRead(t, filepath.Join(tmp, ".namba", "project", "release-checklist.md"))
	if !strings.Contains(releaseChecklist, "current branch is `main`") || !strings.Contains(releaseChecklist, "`namba regen` rerun") {
		t.Fatalf("expected synced release checklist to describe release guardrails, got: %s", releaseChecklist)
	}

	readme := mustRead(t, filepath.Join(tmp, "README.md"))
	if !strings.Contains(readme, "What You Can Do In This Repository") || !strings.Contains(readme, "namba pr") || !strings.Contains(readme, "Workflow Guide") {
		t.Fatalf("expected synced README bundle, got: %s", readme)
	}

	localizedReadme := mustRead(t, filepath.Join(tmp, "README.ko.md"))
	if !strings.Contains(localizedReadme, "이 저장소에서 할 수 있는 일") {
		t.Fatalf("expected localized README bundle, got: %s", localizedReadme)
	}

	workflowGuide := mustRead(t, filepath.Join(tmp, "docs", "workflow-guide.md"))
	if !strings.Contains(workflowGuide, "Collaboration rules") || !strings.Contains(workflowGuide, "namba land") || !strings.Contains(workflowGuide, "multi-subagent workflow") {
		t.Fatalf("expected workflow guide doc, got: %s", workflowGuide)
	}

	gettingStarted := mustRead(t, filepath.Join(tmp, "docs", "getting-started.md"))
	if !strings.Contains(gettingStarted, "WSL workspace") {
		t.Fatalf("expected synced getting started doc to describe current Windows WSL guidance, got: %s", gettingStarted)
	}

}

func TestSyncRespectsExplicitEmptyReadmeLanguages(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".namba", "config", "sections", "docs.yaml"), []byte("manage_readme: true\nreadme_profile: managed-project\nreadme_default_language: en\nreadme_additional_languages:\nreadme_hero_image:\n"), 0o644); err != nil {
		t.Fatalf("write docs config: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"sync"}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	mustExist(t, filepath.Join(tmp, "README.md"))
	mustExist(t, filepath.Join(tmp, "docs", "workflow-guide.md"))
	mustNotExist(t, filepath.Join(tmp, "README.ko.md"))
	mustNotExist(t, filepath.Join(tmp, "docs", "workflow-guide.ko.md"))
}

func TestSyncRemovesManagedReadmesWhenDisabled(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".namba", "config", "sections", "docs.yaml"), []byte("manage_readme: false\nreadme_profile: managed-project\nreadme_default_language: en\nreadme_additional_languages:\nreadme_hero_image:\n"), 0o644); err != nil {
		t.Fatalf("write docs config: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"sync"}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	mustNotExist(t, filepath.Join(tmp, "README.md"))
	mustNotExist(t, filepath.Join(tmp, "docs", "getting-started.md"))
	mustNotExist(t, filepath.Join(tmp, "docs", "workflow-guide.md"))
}

func TestInitConfiguresGoFormatTargets(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/demo\n\ngo 1.24.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "cmd", "demo"), 0o755); err != nil {
		t.Fatalf("mkdir cmd: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "cmd", "demo", "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "root_test.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write root_test.go: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "external", "ref"), 0o755); err != nil {
		t.Fatalf("mkdir external: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "external", "ref", "ignored.go"), []byte("package ref\n"), 0o644); err != nil {
		t.Fatalf("write ignored.go: %v", err)
	}

	if err := app.Run(context.Background(), []string{"init", tmp}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	quality, err := os.ReadFile(filepath.Join(tmp, ".namba", "config", "sections", "quality.yaml"))
	if err != nil {
		t.Fatalf("read quality config: %v", err)
	}

	text := string(quality)
	if !strings.Contains(text, `lint_command: gofmt -l "cmd" "root_test.go"`) {
		t.Fatalf("unexpected lint command: %s", text)
	}
	if strings.Contains(text, "external") {
		t.Fatalf("expected external paths to be excluded: %s", text)
	}
}

func mustExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func mustNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected %s to be absent", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func chdir(t *testing.T, dir string) func() {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to %s: %v", dir, err)
	}
	return func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}
}
