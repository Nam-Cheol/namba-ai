package namba

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBuildAnalysisInventoryCapturesFilesAndSystemRoots(t *testing.T) {
	t.Parallel()

	root := canonicalTempDir(t)
	writeSpec027TestFile(t, filepath.Join(root, "go.mod"), "module example.com/spec027\n\ngo 1.22\n")
	writeSpec027TestFile(t, filepath.Join(root, "frontend", "package.json"), `{"name":"frontend","dependencies":{"react":"18.3.1"}}`)
	writeSpec027TestFile(t, filepath.Join(root, "services", "api", "go.mod"), "module example.com/api\n\ngo 1.22\n")
	writeSpec027TestFile(t, filepath.Join(root, ".namba", "project", "product.md"), "# generated\n")

	inventory := buildAnalysisInventory(root, defaultAnalysisConfig())
	if len(inventory.Files) == 0 {
		t.Fatal("expected analysis inventory to capture files")
	}
	if got, want := inventory.SystemRoots, []string{".", "frontend", "services/api"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("system roots = %#v, want %#v", got, want)
	}
	for _, file := range inventory.Files {
		if file.Path == ".namba/project/product.md" {
			t.Fatalf("expected generated project outputs to be excluded from inventory, got %+v", inventory.Files)
		}
	}
}

func TestAnalyzeProjectPreservesInventoryForRenderAndQuality(t *testing.T) {
	t.Parallel()

	root := canonicalTempDir(t)
	writeSpec027TestFile(t, filepath.Join(root, "go.mod"), "module example.com/spec027\n\ngo 1.22\n")
	writeSpec027TestFile(t, filepath.Join(root, "cmd", "app", "main.go"), "package main\n\nfunc main() {}\n")
	writeSpec027TestFile(t, filepath.Join(root, "README.md"), "This repository is a Go service.\n")

	analysis := analyzeProject(
		root,
		projectConfig{Name: "spec027", Language: "go", Framework: "none"},
		qualityConfig{TestCommand: "go test ./...", LintCommand: "gofmt -l .", TypecheckCommand: "go vet ./..."},
		defaultAnalysisConfig(),
	)

	if len(analysis.Inventory.Files) == 0 {
		t.Fatal("expected project analysis to retain inventory files")
	}
	if got, want := analysis.Inventory.SystemRoots, []string{"."}; !reflect.DeepEqual(got, want) {
		t.Fatalf("inventory system roots = %#v, want %#v", got, want)
	}
	if len(analysis.Systems) == 0 {
		t.Fatal("expected project analysis to build systems from inventory")
	}
	if len(analysis.Conflicts) != 0 {
		t.Fatalf("expected matching README/runtime signals to avoid conflicts, got %+v", analysis.Conflicts)
	}
	if len(analysis.Quality.Warnings) == 0 {
		t.Fatal("expected quality evaluation to surface the narrow conflict-heuristic warning")
	}
}

func TestBuildAnalysisQualityInputsSummarizesSystemsAndConflicts(t *testing.T) {
	t.Parallel()

	inputs := buildAnalysisQualityInputs(
		analysisHeuristicInputs{
			ReadmePath: "README.md",
			Inventory: analysisInventory{
				Files: []analysisFile{
					{Path: "go.mod", Category: "code"},
					{Path: "cmd/app/main.go", Category: "code"},
				},
			},
			Systems: []analysisSystem{
				{
					Name:        "api",
					Purpose:     []analysisFinding{{Confidence: "high"}},
					EntryPoints: []analysisFinding{{Confidence: "medium"}},
					Modules:     []analysisFinding{{Confidence: "low"}},
					Risks:       []analysisFinding{{Confidence: "low"}},
				},
			},
		},
		[]analysisConflict{{Claim: "README mismatch"}},
	)

	if inputs.FileCount != 2 {
		t.Fatalf("file count = %d, want 2", inputs.FileCount)
	}
	if inputs.ReadmePath != "README.md" {
		t.Fatalf("readme path = %q, want README.md", inputs.ReadmePath)
	}
	if inputs.ConflictCount != 1 {
		t.Fatalf("conflict count = %d, want 1", inputs.ConflictCount)
	}
	if got, want := len(inputs.Systems), 1; got != want {
		t.Fatalf("system summaries = %d, want %d", got, want)
	}
	if got, want := inputs.Systems[0], (analysisSystemQualitySummary{
		Name:                  "api",
		TotalFindings:         4,
		LowConfidenceCount:    2,
		StrongConfidenceCount: 2,
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("quality summary = %#v, want %#v", got, want)
	}
}

func TestEvaluateAnalysisQualityUsesSummaryInputs(t *testing.T) {
	t.Parallel()

	quality := evaluateAnalysisQuality(analysisQualityInputs{
		FileCount:     2,
		ReadmePath:    "README.md",
		ConflictCount: 0,
		Systems: []analysisSystemQualitySummary{
			{
				Name:                  "api",
				TotalFindings:         4,
				LowConfidenceCount:    3,
				StrongConfidenceCount: 1,
			},
		},
	})

	if len(quality.Errors) != 0 {
		t.Fatalf("expected no quality errors, got %#v", quality.Errors)
	}
	for _, want := range []string{
		"System `api` is thin; fewer than five evidence-backed findings were produced.",
		"System `api` is dominated by low-confidence inference; add stronger code/config signals or tune analysis scope.",
		"No code-vs-doc conflicts were detected. This may be correct, but the v1 conflict heuristics remain intentionally narrow.",
	} {
		if !containsString(quality.Warnings, want) {
			t.Fatalf("expected quality warnings to contain %q, got %#v", want, quality.Warnings)
		}
	}
}

func TestBuildAnalysisConflictInputsSummarizesRuntimeSignalsAndEvidence(t *testing.T) {
	t.Parallel()

	inputs := buildAnalysisConflictInputs(analysisHeuristicInputs{
		Config:     defaultAnalysisConfig(),
		Project:    projectConfig{Name: "spec027", Language: "go", Framework: "none"},
		ReadmePath: "README.md",
		ReadmeBody: "This repository is a React frontend.\n",
		Inventory: analysisInventory{
			Files: []analysisFile{
				{Path: "go.mod", Category: "code"},
				{Path: "cmd/app/main.go", Category: "code"},
			},
			SystemRoots: []string{"."},
		},
		Systems: []analysisSystem{
			{
				Name: "workspace",
				Root: ".",
				Kind: "go-service",
				Purpose: []analysisFinding{
					{Claim: "A Go-based service surface is present.", Confidence: "high", Evidence: []string{"go.mod"}},
				},
			},
		},
	})

	if inputs.ReadmePath != "README.md" {
		t.Fatalf("readme path = %q, want README.md", inputs.ReadmePath)
	}
	if got, want := inputs.RuntimeSignals, []string{"go service"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("runtime signals = %#v, want %#v", got, want)
	}
	if !strings.Contains(inputs.RuntimeEvidence, "go.mod") {
		t.Fatalf("expected runtime evidence to cite go.mod, got %q", inputs.RuntimeEvidence)
	}
}

func TestBuildAnalysisConflictRuntimeSupportSummarizesSignalsAndEvidence(t *testing.T) {
	t.Parallel()

	support := buildAnalysisConflictRuntimeSupport(
		projectConfig{Name: "spec027", Language: "go", Framework: "none"},
		analysisInventory{
			Files: []analysisFile{
				{Path: "go.mod", Category: "code"},
				{Path: "cmd/app/main.go", Category: "code"},
			},
			SystemRoots: []string{"."},
		},
		[]analysisSystem{
			{
				Name: "workspace",
				Root: ".",
				Kind: "go-service",
				Purpose: []analysisFinding{
					{Claim: "A Go-based service surface is present.", Confidence: "high", Evidence: []string{"go.mod"}},
				},
			},
		},
		defaultAnalysisConfig(),
	)

	if got, want := support.Signals, []string{"go service"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("runtime support signals = %#v, want %#v", got, want)
	}
	if !strings.Contains(support.Evidence, "go.mod") {
		t.Fatalf("expected runtime support evidence to cite go.mod, got %q", support.Evidence)
	}
}

func TestBuildAnalysisConflictRuntimeSupportLeavesEvidenceBlankWithoutSignals(t *testing.T) {
	t.Parallel()

	support := buildAnalysisConflictRuntimeSupport(
		projectConfig{Name: "spec027", Language: "", Framework: "none"},
		analysisInventory{
			Files:       []analysisFile{{Path: "README.md", Category: "docs"}},
			SystemRoots: []string{"."},
		},
		nil,
		defaultAnalysisConfig(),
	)

	if len(support.Signals) != 0 {
		t.Fatalf("expected no runtime support signals, got %#v", support.Signals)
	}
	if support.Evidence != "" {
		t.Fatalf("expected blank runtime evidence without signals, got %q", support.Evidence)
	}
}

func TestDetectAnalysisConflictsUsesPrecomputedRuntimeSupport(t *testing.T) {
	t.Parallel()

	conflicts := detectAnalysisConflicts(analysisConflictInputs{
		ReadmePath:      "README.md",
		ReadmeBody:      "This repository is a React frontend.\n",
		RuntimeSignals:  []string{"go service"},
		RuntimeEvidence: "go.mod, cmd/app/main.go",
	})

	if got, want := len(conflicts), 1; got != want {
		t.Fatalf("conflict count = %d, want %d; conflicts=%+v", got, want, conflicts)
	}
	if conflicts[0].Stronger != "go.mod, cmd/app/main.go" {
		t.Fatalf("expected stronger evidence to come from precomputed runtime support, got %+v", conflicts[0])
	}
	if conflicts[0].Weaker != "README.md" {
		t.Fatalf("expected weaker source to stay on readme path, got %+v", conflicts[0])
	}
}

func TestEvaluateAnalysisHeuristicsSeparatesConflictsFromQualityWarnings(t *testing.T) {
	t.Parallel()

	inputs := analysisHeuristicInputs{
		Config:     defaultAnalysisConfig(),
		Project:    projectConfig{Name: "spec027", Language: "go", Framework: "none"},
		ReadmePath: "README.md",
		ReadmeBody: "This repository is a React frontend.\n",
		Inventory: analysisInventory{
			Files: []analysisFile{
				{Path: "go.mod", Category: "code"},
				{Path: "cmd/app/main.go", Category: "code"},
			},
			SystemRoots: []string{"."},
		},
		Systems: []analysisSystem{
			{
				Name: "workspace",
				Root: ".",
				Kind: "go-service",
				Purpose: []analysisFinding{
					{Claim: "A Go-based service surface is present.", Confidence: "high", Evidence: []string{"go.mod"}},
				},
				EntryPoints: []analysisFinding{
					{Claim: "`cmd/app/main.go`: entrypoint", Confidence: "high", Evidence: []string{"cmd/app/main.go"}},
				},
				Modules: []analysisFinding{
					{Claim: "`cmd` is a visible module boundary.", Confidence: "medium", Evidence: []string{"cmd"}},
				},
				DataState: []analysisFinding{
					{Claim: "Generated project state is persisted under `.namba`.", Confidence: "high", Evidence: []string{".namba/manifest.json"}},
				},
				AuthIntegrations: []analysisFinding{
					{Claim: "External tool integrations are part of the system surface.", Confidence: "medium", Evidence: []string{".github/workflows/ci.yml"}},
				},
				Risks: []analysisFinding{
					{Claim: "System-local regression coverage exists.", Confidence: "medium", Evidence: []string{"internal/namba/project_analysis_inventory_test.go"}},
				},
			},
		},
	}

	heuristics := evaluateAnalysisHeuristics(inputs)
	if got, want := len(heuristics.Conflicts), 1; got != want {
		t.Fatalf("conflict count = %d, want %d; conflicts=%+v", got, want, heuristics.Conflicts)
	}
	if !strings.Contains(heuristics.Conflicts[0].Stronger, "go.mod") {
		t.Fatalf("expected conflict evidence to cite go.mod, got %+v", heuristics.Conflicts[0])
	}
	for _, warning := range heuristics.Quality.Warnings {
		if strings.Contains(warning, "No code-vs-doc conflicts were detected") {
			t.Fatalf("did not expect no-conflict warning when conflicts exist, got %+v", heuristics.Quality.Warnings)
		}
	}
	if len(heuristics.Quality.Errors) != 0 {
		t.Fatalf("expected no quality errors, got %+v", heuristics.Quality.Errors)
	}
}

func TestAnalysisInventoryLookupContractUsesNormalizedPaths(t *testing.T) {
	t.Parallel()

	root := canonicalTempDir(t)
	writeSpec027TestFile(t, filepath.Join(root, "README.md"), "Root summary\n")
	writeSpec027TestFile(t, filepath.Join(root, "go.mod"), "module example.com/spec027\n\ngo 1.22\n")
	writeSpec027TestFile(t, filepath.Join(root, "cmd", "app", "main.go"), "package main\n\nfunc main() {}\n")
	writeSpec027TestFile(t, filepath.Join(root, "services", "api", "main.go"), "package main\n\nfunc main() {}\n")
	writeSpec027TestFile(t, filepath.Join(root, "frontend", "package.json"), `{"name":"frontend","dependencies":{"react":"18.3.1"}}`)
	writeSpec027TestFile(t, filepath.Join(root, ".namba", "project", "product.md"), "# generated\n")

	inventory := buildAnalysisInventory(root, defaultAnalysisConfig())
	if !analysisFileExists(inventory.Files, "go.mod") {
		t.Fatal("expected analysis inventory to include go.mod by normalized path")
	}
	if analysisFileExists(inventory.Files, ".namba/project/product.md") {
		t.Fatalf("expected analysis inventory to exclude generated project outputs, got %+v", inventory.Files)
	}
	if got, want := preferredEvidence(inventory.Files, "go.mod", "package.json"), []string{"go.mod", "frontend/package.json"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("preferred evidence = %#v, want %#v", got, want)
	}
	if got, want := suffixEvidence(inventory.Files, ".go"), []string{"cmd/app/main.go", "services/api/main.go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("suffix evidence = %#v, want %#v", got, want)
	}

	rootFiles := filesForSystem(inventory.Files, ".", inventory.SystemRoots)
	for _, blocked := range []string{"frontend/package.json", "services/api/main.go"} {
		for _, file := range rootFiles {
			if file.Path == blocked {
				t.Fatalf("expected root system scope to exclude nested system file %q, got %+v", blocked, rootFiles)
			}
		}
	}
}

func TestAnalysisIndexCachesRepeatedReads(t *testing.T) {
	t.Parallel()

	root := canonicalTempDir(t)
	writeSpec027TestFile(t, filepath.Join(root, "README.md"), "Root summary\n")
	writeSpec027TestFile(t, filepath.Join(root, "go.mod"), "module example.com/spec027\n\ngo 1.22\n")
	writeSpec027TestFile(t, filepath.Join(root, "cmd", "app", "main.go"), "package main\n\nfunc main() {}\n")
	writeSpec027TestFile(t, filepath.Join(root, "frontend", "package.json"), `{"name":"frontend","dependencies":{"react":"18.3.1"}}`)

	inventory := buildAnalysisInventory(root, defaultAnalysisConfig())
	index := buildAnalysisIndex(root, inventory.Files)

	summary, evidence := analysisReadSystemSummary(index, ".")
	if summary == "" || !reflect.DeepEqual(evidence, []string{"README.md"}) {
		t.Fatalf("expected root README summary from cached index, got summary=%q evidence=%#v", summary, evidence)
	}
	if got, want := len(index.textLoaded), 1; got != want {
		t.Fatalf("expected one cached text read after README summary, got %d", got)
	}
	summary, evidence = analysisReadSystemSummary(index, ".")
	if summary == "" || !reflect.DeepEqual(evidence, []string{"README.md"}) {
		t.Fatalf("expected repeated README summary lookup to stay stable, got summary=%q evidence=%#v", summary, evidence)
	}
	if got, want := len(index.textLoaded), 1; got != want {
		t.Fatalf("expected cached README summary lookup not to increase text reads, got %d", got)
	}

	module := analysisReadGoModule("go.mod", index)
	if module != "example.com/spec027" {
		t.Fatalf("expected go module lookup from cached index, got %q", module)
	}
	if got, want := len(index.textLoaded), 2; got != want {
		t.Fatalf("expected go.mod lookup to add one cached read, got %d", got)
	}
	module = analysisReadGoModule("go.mod", index)
	if module != "example.com/spec027" {
		t.Fatalf("expected repeated go.mod lookup to stay stable, got %q", module)
	}
	if got, want := len(index.textLoaded), 2; got != want {
		t.Fatalf("expected repeated go.mod lookup not to increase text reads, got %d", got)
	}

	bootstrap := analysisFirstFileContaining(index, inventory.Files, "createRoot(")
	if bootstrap != "" {
		t.Fatalf("expected createRoot lookup to be absent in the minimal fixture, got %q", bootstrap)
	}
}
