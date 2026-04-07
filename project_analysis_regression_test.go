package namba_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nam-Cheol/namba-ai/internal/namba"
)

func TestProjectAnalysisWritesFoundationArtifacts(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := os.MkdirAll(filepath.Join(tmp, "cmd", "app"), 0o755); err != nil {
		t.Fatalf("mkdir cmd/app: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/foundation\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "cmd", "app", "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write cmd/app/main.go: %v", err)
	}

	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"project"}); err != nil {
		t.Fatalf("project failed: %v", err)
	}

	mustExist(t, filepath.Join(tmp, ".namba", "config", "sections", "analysis.yaml"))
	for _, rel := range []string{
		filepath.Join(".namba", "project", "product.md"),
		filepath.Join(".namba", "project", "tech.md"),
		filepath.Join(".namba", "project", "structure.md"),
		filepath.Join(".namba", "project", "mismatch-report.md"),
		filepath.Join(".namba", "project", "quality-report.md"),
		filepath.Join(".namba", "project", "codemaps", "overview.md"),
		filepath.Join(".namba", "project", "codemaps", "entry-points.md"),
		filepath.Join(".namba", "project", "codemaps", "dependencies.md"),
		filepath.Join(".namba", "project", "codemaps", "data-flow.md"),
		filepath.Join(".namba", "project", "systems", "workspace.md"),
	} {
		mustExist(t, filepath.Join(tmp, rel))
	}

	structure := mustRead(t, filepath.Join(tmp, ".namba", "project", "structure.md"))
	if !strings.Contains(structure, "Appendix output only. Use `product.md` and `tech.md` first.") {
		t.Fatalf("expected structure appendix notice, got: %s", structure)
	}

	product := mustRead(t, filepath.Join(tmp, ".namba", "project", "product.md"))
	if !strings.Contains(product, ".namba/project/systems/workspace.md") || !strings.Contains(product, "System Landscape") {
		t.Fatalf("expected product doc to reference system summaries, got: %s", product)
	}

	quality := mustRead(t, filepath.Join(tmp, ".namba", "project", "quality-report.md"))
	if !strings.Contains(quality, "No warnings or errors") && !strings.Contains(quality, "Warnings") {
		t.Fatalf("expected quality report to be emitted, got: %s", quality)
	}
}

func TestProjectAnalysisSeparatesBackendAndFrontendSystems(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := os.MkdirAll(filepath.Join(tmp, "src", "app"), 0o755); err != nil {
		t.Fatalf("mkdir src/app: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "services", "api", "cmd", "api"), 0o755); err != nil {
		t.Fatalf("mkdir services/api/cmd/api: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "package.json"), []byte(`{
  "name": "web",
  "dependencies": {
    "react": "18.3.1",
    "react-dom": "18.3.1",
    "react-router": "7.13.0"
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

export default function App() {
  return <RouterProvider router={{}} />;
}
`), 0o644); err != nil {
		t.Fatalf("write src/app/App.tsx: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "services", "api", "go.mod"), []byte("module example.com/api\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatalf("write services/api/go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "services", "api", "cmd", "api", "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write services/api/cmd/api/main.go: %v", err)
	}
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "README.md"), []byte("This repository is a Go service.\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"project"}); err != nil {
		t.Fatalf("project failed: %v", err)
	}

	frontend := mustRead(t, filepath.Join(tmp, ".namba", "project", "systems", "workspace.md"))
	if !strings.Contains(frontend, "Kind: frontend") || !strings.Contains(frontend, "src/main.tsx") {
		t.Fatalf("expected frontend system summary, got: %s", frontend)
	}

	backend := mustRead(t, filepath.Join(tmp, ".namba", "project", "systems", "services-api.md"))
	if !strings.Contains(backend, "Kind: backend") || !strings.Contains(backend, "services/api/cmd/api/main.go") {
		t.Fatalf("expected backend system summary, got: %s", backend)
	}

	mismatch := mustRead(t, filepath.Join(tmp, ".namba", "project", "mismatch-report.md"))
	if !strings.Contains(mismatch, "No explicit code-vs-doc conflicts were detected") {
		t.Fatalf("expected mixed-runtime repo to avoid false-positive README drift, got: %s", mismatch)
	}

	overview := mustRead(t, filepath.Join(tmp, ".namba", "project", "codemaps", "overview.md"))
	if !strings.Contains(overview, "workspace") || !strings.Contains(overview, "api") {
		t.Fatalf("expected overview to separate backend and frontend systems, got: %s", overview)
	}
}

func TestProjectAnalysisRemovesStaleSystemDocs(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := os.MkdirAll(filepath.Join(tmp, "frontend", "src"), 0o755); err != nil {
		t.Fatalf("mkdir frontend/src: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "backend", "cmd", "api"), 0o755); err != nil {
		t.Fatalf("mkdir backend/cmd/api: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "frontend", "package.json"), []byte(`{"name":"frontend","dependencies":{"react":"18.3.1","react-dom":"18.3.1"}}`), 0o644); err != nil {
		t.Fatalf("write frontend/package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "frontend", "src", "main.tsx"), []byte(`export function main() { return null; }`), 0o644); err != nil {
		t.Fatalf("write frontend/src/main.tsx: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "backend", "go.mod"), []byte("module example.com/backend\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatalf("write backend/go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "backend", "cmd", "api", "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write backend/cmd/api/main.go: %v", err)
	}
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"project"}); err != nil {
		t.Fatalf("first project failed: %v", err)
	}
	mustExist(t, filepath.Join(tmp, ".namba", "project", "systems", "frontend.md"))

	if err := os.RemoveAll(filepath.Join(tmp, "frontend")); err != nil {
		t.Fatalf("remove frontend system: %v", err)
	}
	if err := app.Run(context.Background(), []string{"project"}); err != nil {
		t.Fatalf("second project failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, ".namba", "project", "systems", "frontend.md")); !os.IsNotExist(err) {
		t.Fatalf("expected stale frontend system doc to be removed, stat err=%v", err)
	}

	product := mustRead(t, filepath.Join(tmp, ".namba", "project", "product.md"))
	if strings.Contains(product, ".namba/project/systems/frontend.md") {
		t.Fatalf("expected product doc to drop stale frontend reference, got: %s", product)
	}

	manifest := mustRead(t, filepath.Join(tmp, ".namba", "manifest.json"))
	if strings.Contains(manifest, ".namba/project/systems/frontend.md") {
		t.Fatalf("expected manifest to drop stale frontend system doc, got: %s", manifest)
	}
}

func TestProjectAnalysisGeneratesUniqueSystemDocSlugs(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := os.MkdirAll(filepath.Join(tmp, "services", "api", "cmd", "api"), 0o755); err != nil {
		t.Fatalf("mkdir services/api/cmd/api: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "services-api", "cmd", "alt"), 0o755); err != nil {
		t.Fatalf("mkdir services-api/cmd/alt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "services", "api", "go.mod"), []byte("module example.com/serviceapi\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatalf("write services/api/go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "services", "api", "cmd", "api", "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write services/api main: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "services-api", "go.mod"), []byte("module example.com/servicesdashapi\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatalf("write services-api/go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "services-api", "cmd", "alt", "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write services-api main: %v", err)
	}
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"project"}); err != nil {
		t.Fatalf("project failed: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(tmp, ".namba", "project", "systems"))
	if err != nil {
		t.Fatalf("read systems dir: %v", err)
	}
	if len(entries) < 2 {
		t.Fatalf("expected at least two distinct system docs, got %d", len(entries))
	}

	var sawNested bool
	var sawDashed bool
	for _, entry := range entries {
		body := mustRead(t, filepath.Join(tmp, ".namba", "project", "systems", entry.Name()))
		if strings.Contains(body, "- Root: `services/api`") {
			sawNested = true
		}
		if strings.Contains(body, "- Root: `services-api`") {
			sawDashed = true
		}
	}
	if !sawNested || !sawDashed {
		t.Fatalf("expected distinct system docs for both roots, nested=%v dashed=%v", sawNested, sawDashed)
	}
}

func TestProjectAnalysisKeepsNestedInfraRootsSeparate(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := os.MkdirAll(filepath.Join(tmp, "services", "api"), 0o755); err != nil {
		t.Fatalf("mkdir services/api: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "services", "web"), 0o755); err != nil {
		t.Fatalf("mkdir services/web: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "services", "api", "docker-compose.yml"), []byte("services:\n  api:\n    image: busybox\n"), 0o644); err != nil {
		t.Fatalf("write services/api/docker-compose.yml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "services", "web", "docker-compose.yml"), []byte("services:\n  web:\n    image: busybox\n"), 0o644); err != nil {
		t.Fatalf("write services/web/docker-compose.yml: %v", err)
	}
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"project"}); err != nil {
		t.Fatalf("project failed: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(tmp, ".namba", "project", "systems"))
	if err != nil {
		t.Fatalf("read systems dir: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected two infra system docs, got %d", len(entries))
	}

	var sawAPI bool
	var sawWeb bool
	for _, entry := range entries {
		body := mustRead(t, filepath.Join(tmp, ".namba", "project", "systems", entry.Name()))
		if strings.Contains(body, "- Root: `services/api`") {
			sawAPI = true
		}
		if strings.Contains(body, "- Root: `services/web`") {
			sawWeb = true
		}
		if strings.Contains(body, "- Root: `services`") {
			t.Fatalf("expected nested infra roots to remain separated, got merged doc: %s", body)
		}
	}
	if !sawAPI || !sawWeb {
		t.Fatalf("expected separate docs for nested infra roots, api=%v web=%v", sawAPI, sawWeb)
	}
}

func TestProjectAnalysisFailsWhenScopeExcludesEverything(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := os.MkdirAll(filepath.Join(tmp, "cmd", "app"), 0o755); err != nil {
		t.Fatalf("mkdir cmd/app: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/thin\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "cmd", "app", "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write cmd/app/main.go: %v", err)
	}

	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".namba", "config", "sections", "analysis.yaml"), []byte("include_paths: missing-scope\n"), 0o644); err != nil {
		t.Fatalf("write analysis config: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	err := app.Run(context.Background(), []string{"project"})
	if err == nil || !strings.Contains(err.Error(), "quality gate failed") {
		t.Fatalf("expected quality gate failure, got: %v", err)
	}

	for _, rel := range []string{
		filepath.Join(".namba", "project", "product.md"),
		filepath.Join(".namba", "project", "tech.md"),
		filepath.Join(".namba", "project", "structure.md"),
		filepath.Join(".namba", "project", "mismatch-report.md"),
		filepath.Join(".namba", "project", "quality-report.md"),
	} {
		mustExist(t, filepath.Join(tmp, rel))
	}

	quality := mustRead(t, filepath.Join(tmp, ".namba", "project", "quality-report.md"))
	if !strings.Contains(quality, "No analyzable files matched the configured scope") {
		t.Fatalf("expected quality report to explain the scope failure, got: %s", quality)
	}
}
