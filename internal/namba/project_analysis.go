package namba

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const systemDocsDir = ".namba/project/systems"

type analysisConfig struct {
	IncludePaths      []string
	ExcludePaths      []string
	SourcePriority    []string
	ImportantRoles    []string
	ImportantRuntimes []string
	OutputTemplate    string
}

type analysisFile struct {
	Path     string
	Category string
}

type analysisFinding struct {
	Claim      string
	Confidence string
	Evidence   []string
}

type analysisConflict struct {
	Claim    string
	Stronger string
	Weaker   string
	Reason   string
}

type runtimeMention struct {
	Label             string
	Patterns          []string
	CompatibleSignals []string
}

type analysisSystem struct {
	Name             string
	Slug             string
	Root             string
	Kind             string
	Purpose          []analysisFinding
	EntryPoints      []analysisFinding
	Modules          []analysisFinding
	DataState        []analysisFinding
	AuthIntegrations []analysisFinding
	Risks            []analysisFinding
}

type analysisQuality struct {
	Warnings []string
	Errors   []string
}

type projectAnalysis struct {
	Config       analysisConfig
	Project      projectConfig
	QualityCfg   qualityConfig
	ReadmePath   string
	ReadmeBody   string
	Files        []analysisFile
	Systems      []analysisSystem
	Conflicts    []analysisConflict
	Quality      analysisQuality
	StructureDoc string
	OverviewDoc  string
	EntryDoc     string
	DepsDoc      string
	FlowDoc      string
}

func renderAnalysisConfig(profile initProfile) string {
	runtimes := []string{"git", "codex"}
	if strings.TrimSpace(profile.Language) != "" {
		runtimes = append([]string{profile.Language}, runtimes...)
	}
	return strings.Join([]string{
		"include_paths: .",
		"exclude_paths: .git,.cache,.gocache,.idea,.vscode,.venv,venv,node_modules,dist,build,coverage,tmp,.tmp,logs,external,vendor,.namba/logs,.namba/project,.namba/worktrees",
		"source_priority: code,config,test,build,docs,planning,generated",
		"important_roles: operator,maintainer",
		fmt.Sprintf("important_runtimes: %s", strings.Join(runtimes, ",")),
		"output_template: foundation-v1",
		"",
	}, "\n")
}

func defaultAnalysisConfig() analysisConfig {
	return analysisConfig{
		IncludePaths:      []string{"."},
		ExcludePaths:      []string{".git", ".cache", ".gocache", ".idea", ".vscode", ".venv", "venv", "node_modules", "dist", "build", "coverage", "tmp", ".tmp", "logs", "external", "vendor", ".namba/logs", ".namba/project", ".namba/worktrees"},
		SourcePriority:    []string{"code", "config", "test", "build", "docs", "planning", "generated"},
		ImportantRoles:    []string{"operator", "maintainer"},
		ImportantRuntimes: []string{"git", "codex"},
		OutputTemplate:    "foundation-v1",
	}
}

func (a *App) loadAnalysisConfig(root string) (analysisConfig, error) {
	cfg := defaultAnalysisConfig()
	values, err := readKeyValueFile(filepath.Join(root, configDir, "analysis.yaml"))
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return analysisConfig{}, err
	}
	if raw := strings.TrimSpace(values["include_paths"]); raw != "" {
		cfg.IncludePaths = parseCommaSeparatedList(raw)
	}
	if raw := strings.TrimSpace(values["exclude_paths"]); raw != "" {
		cfg.ExcludePaths = append([]string{}, parseCommaSeparatedList(raw)...)
	}
	if raw := strings.TrimSpace(values["source_priority"]); raw != "" {
		cfg.SourcePriority = parseCommaSeparatedList(raw)
	}
	if raw := strings.TrimSpace(values["important_roles"]); raw != "" {
		cfg.ImportantRoles = parseCommaSeparatedList(raw)
	}
	if raw := strings.TrimSpace(values["important_runtimes"]); raw != "" {
		cfg.ImportantRuntimes = parseCommaSeparatedList(raw)
	}
	if raw := strings.TrimSpace(values["output_template"]); raw != "" {
		cfg.OutputTemplate = raw
	}
	return cfg, nil
}

func analyzeProject(root string, projectCfg projectConfig, qualityCfg qualityConfig, cfg analysisConfig) projectAnalysis {
	files := collectAnalysisFiles(root, cfg)
	readmePath := firstExisting(root, "README.md", "README.txt")
	readmeBody := ""
	if readmePath != "" {
		if data, err := os.ReadFile(filepath.Join(root, readmePath)); err == nil {
			readmeBody = string(data)
		}
	}

	systems := detectSystems(root, files)
	conflicts := detectAnalysisConflicts(projectCfg, readmePath, readmeBody, files, systems, cfg)

	analysis := projectAnalysis{
		Config:       cfg,
		Project:      projectCfg,
		QualityCfg:   qualityCfg,
		ReadmePath:   readmePath,
		ReadmeBody:   readmeBody,
		Files:        files,
		Systems:      systems,
		Conflicts:    conflicts,
		StructureDoc: renderAnalysisStructure(files),
	}
	analysis.EntryDoc = renderEntryPointsDoc(analysis.Systems)
	analysis.DepsDoc = renderDependenciesDoc(root, analysis.Systems)
	analysis.FlowDoc = renderDataFlowDoc(analysis.Systems, qualityCfg)
	analysis.OverviewDoc = renderOverviewDoc(analysis)
	analysis.Quality = evaluateAnalysisQuality(analysis)
	return analysis
}

func (p projectAnalysis) renderOutputs() map[string]string {
	outputs := map[string]string{
		filepath.ToSlash(filepath.Join(projectDir, "product.md")):         renderProductDoc(p),
		filepath.ToSlash(filepath.Join(projectDir, "tech.md")):            renderTechDoc(p),
		filepath.ToSlash(filepath.Join(projectDir, "structure.md")):       p.StructureDoc,
		filepath.ToSlash(filepath.Join(projectDir, "mismatch-report.md")): renderMismatchReportDoc(p),
		filepath.ToSlash(filepath.Join(projectDir, "quality-report.md")):  renderQualityReportDoc(p),
		filepath.ToSlash(filepath.Join(codemapsDir, "overview.md")):       p.OverviewDoc,
		filepath.ToSlash(filepath.Join(codemapsDir, "entry-points.md")):   p.EntryDoc,
		filepath.ToSlash(filepath.Join(codemapsDir, "dependencies.md")):   p.DepsDoc,
		filepath.ToSlash(filepath.Join(codemapsDir, "data-flow.md")):      p.FlowDoc,
	}
	for _, system := range p.Systems {
		outputs[filepath.ToSlash(filepath.Join(systemDocsDir, system.Slug+".md"))] = renderSystemDoc(system)
	}
	return outputs
}

func collectAnalysisFiles(root string, cfg analysisConfig) []analysisFile {
	var files []analysisFile
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if d.IsDir() {
			if shouldSkipAnalysisPath(rel, cfg) {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldSkipAnalysisPath(rel, cfg) {
			return nil
		}
		if !isIncludedAnalysisPath(rel, cfg) {
			return nil
		}
		files = append(files, analysisFile{
			Path:     rel,
			Category: classifyAnalysisFile(rel),
		})
		return nil
	})
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files
}

func shouldSkipAnalysisPath(rel string, cfg analysisConfig) bool {
	rel = filepath.ToSlash(rel)
	if rel == "" || rel == "." {
		return false
	}
	for _, prefix := range cfg.ExcludePaths {
		prefix = strings.Trim(strings.TrimSpace(filepath.ToSlash(prefix)), "/")
		if prefix == "" {
			continue
		}
		if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
			return true
		}
	}
	base := strings.ToLower(filepath.Base(rel))
	return strings.HasSuffix(base, ".log")
}

func isIncludedAnalysisPath(rel string, cfg analysisConfig) bool {
	if len(cfg.IncludePaths) == 0 {
		return true
	}
	for _, prefix := range cfg.IncludePaths {
		prefix = strings.Trim(strings.TrimSpace(filepath.ToSlash(prefix)), "/")
		if prefix == "" || prefix == "." {
			return true
		}
		if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
			return true
		}
	}
	return false
}

func classifyAnalysisFile(rel string) string {
	rel = filepath.ToSlash(rel)
	lower := strings.ToLower(rel)
	switch {
	case strings.Contains(lower, "/test/"), strings.Contains(lower, "__tests__"), strings.HasSuffix(lower, "_test.go"), strings.HasSuffix(lower, ".spec.ts"), strings.HasSuffix(lower, ".test.ts"), strings.HasSuffix(lower, ".spec.tsx"), strings.HasSuffix(lower, ".test.tsx"):
		return "test"
	case strings.HasPrefix(lower, ".namba/specs/"):
		return "planning"
	case strings.HasPrefix(lower, ".github/"), strings.Contains(lower, "docker-compose"), strings.HasSuffix(lower, "dockerfile"), strings.HasSuffix(lower, ".tf"), strings.Contains(lower, "/k8s/"), strings.Contains(lower, "/helm/"):
		return "build"
	case strings.HasSuffix(lower, ".md"), strings.HasSuffix(lower, ".adoc"), strings.HasSuffix(lower, ".rst"):
		return "docs"
	case strings.HasSuffix(lower, ".yaml"), strings.HasSuffix(lower, ".yml"), strings.HasSuffix(lower, ".json"), strings.HasSuffix(lower, ".toml"), strings.HasSuffix(lower, ".env"), strings.HasSuffix(lower, ".ini"):
		return "config"
	case strings.HasPrefix(lower, ".namba/"), strings.HasPrefix(lower, ".agents/"), strings.HasPrefix(lower, ".codex/"):
		return "generated"
	default:
		return "code"
	}
}

func detectSystems(root string, files []analysisFile) []analysisSystem {
	roots := collectSystemRoots(files)
	if len(roots) == 0 {
		roots = []string{"."}
	}

	var systems []analysisSystem
	for _, systemRoot := range roots {
		system := buildAnalysisSystem(root, systemRoot, roots, files)
		if totalSystemFindings(system) == 0 {
			continue
		}
		systems = append(systems, system)
	}
	if len(systems) == 0 {
		systems = append(systems, buildAnalysisSystem(root, ".", roots, files))
	}
	sort.Slice(systems, func(i, j int) bool { return systems[i].Root < systems[j].Root })
	assignUniqueSystemSlugs(systems)
	return systems
}

func assignUniqueSystemSlugs(systems []analysisSystem) {
	baseCounts := map[string]int{}
	for _, system := range systems {
		baseCounts[slugifySystemRoot(system.Root)]++
	}

	used := map[string]bool{}
	for i := range systems {
		base := slugifySystemRoot(systems[i].Root)
		if baseCounts[base] == 1 && !used[base] {
			systems[i].Slug = base
			used[base] = true
			continue
		}

		hash := checksum(systems[i].Root)
		slug := ""
		for width := 8; width <= len(hash); width += 4 {
			candidate := fmt.Sprintf("%s--%s", base, hash[:width])
			if used[candidate] {
				continue
			}
			slug = candidate
			break
		}
		if slug == "" {
			slug = fmt.Sprintf("%s--%s", base, hash)
		}
		systems[i].Slug = slug
		used[slug] = true
	}
}

func collectSystemRoots(files []analysisFile) []string {
	seen := map[string]bool{}
	var roots []string
	for _, file := range files {
		root := candidateSystemRoot(file.Path)
		if root == "" {
			continue
		}
		if !seen[root] {
			seen[root] = true
			roots = append(roots, root)
		}
	}
	sort.Strings(roots)
	return roots
}

func candidateSystemRoot(rel string) string {
	rel = filepath.ToSlash(rel)
	parts := strings.Split(rel, "/")
	base := strings.ToLower(filepath.Base(rel))
	switch {
	case base == "go.mod" || base == "package.json" || base == "pyproject.toml" || base == "requirements.txt" || base == "pom.xml" || base == "build.gradle" || base == "build.gradle.kts":
		return filepath.ToSlash(filepath.Dir(rel))
	case base == "docker-compose.yml" || base == "docker-compose.yaml" || base == "dockerfile" || strings.HasSuffix(base, ".tf") || base == "kustomization.yaml":
		if len(parts) >= 3 && (parts[0] == "apps" || parts[0] == "services" || parts[0] == "packages") {
			return filepath.ToSlash(filepath.Join(parts[0], parts[1]))
		}
		if len(parts) > 1 {
			return parts[0]
		}
		return "."
	case len(parts) >= 3 && (parts[0] == "apps" || parts[0] == "services" || parts[0] == "packages"):
		return filepath.ToSlash(filepath.Join(parts[0], parts[1]))
	case len(parts) >= 2 && isRecognizedSystemDir(parts[0]) && hasSignalFile(base):
		return parts[0]
	case strings.HasPrefix(rel, "cmd/") || base == "main.go":
		return "."
	default:
		return ""
	}
}

func isRecognizedSystemDir(name string) bool {
	switch strings.ToLower(name) {
	case "frontend", "web", "ui", "client", "backend", "server", "api", "worker", "infra":
		return true
	default:
		return false
	}
}

func hasSignalFile(base string) bool {
	switch base {
	case "go.mod", "package.json", "pyproject.toml", "requirements.txt", "pom.xml", "build.gradle", "build.gradle.kts", "docker-compose.yml", "docker-compose.yaml", "dockerfile", "main.go":
		return true
	default:
		return strings.HasSuffix(base, ".tf")
	}
}

func buildAnalysisSystem(root, systemRoot string, allRoots []string, files []analysisFile) analysisSystem {
	name := "workspace"
	if systemRoot != "." {
		name = filepath.Base(systemRoot)
	}
	systemFiles := filesForSystem(files, systemRoot, allRoots)
	kind := detectSystemKind(systemRoot, systemFiles)
	system := analysisSystem{
		Name: name,
		Slug: slugifySystemRoot(systemRoot),
		Root: systemRoot,
		Kind: kind,
	}
	system.Purpose = buildPurposeFindings(root, systemRoot, kind, systemFiles)
	system.EntryPoints = buildEntryPointFindings(root, systemRoot, kind)
	system.Modules = buildModuleFindings(systemRoot, systemFiles)
	system.DataState = buildDataStateFindings(systemRoot, systemFiles)
	system.AuthIntegrations = buildAuthIntegrationFindings(systemRoot, systemFiles)
	system.Risks = buildRiskFindings(systemRoot, systemFiles)
	return system
}

func detectSystemKind(systemRoot string, systemFiles []analysisFile) string {
	paths := make([]string, 0, len(systemFiles))
	for _, file := range systemFiles {
		paths = append(paths, strings.ToLower(file.Path))
	}
	switch {
	case anyPathMatches(paths, "package.json") && anyPathContains(paths, "src/main", "src/app", "react"):
		return "frontend"
	case anyPathMatches(paths, "go.mod"):
		if systemRoot == "." {
			return "go-service"
		}
		return "backend"
	case anyPathMatches(paths, "docker-compose.yml", "docker-compose.yaml") || anyPathContains(paths, "/helm/", "/k8s/") || anyPathSuffix(paths, ".tf"):
		return "infra"
	case anyPathMatches(paths, "pyproject.toml", "requirements.txt"):
		return "python-service"
	default:
		return "system"
	}
}

func filesForSystem(files []analysisFile, systemRoot string, allRoots []string) []analysisFile {
	var filtered []analysisFile
	for _, file := range files {
		if systemRoot == "." {
			if belongsToNestedSystem(file.Path, allRoots) {
				continue
			}
			filtered = append(filtered, file)
			continue
		}
		if file.Path == systemRoot || strings.HasPrefix(file.Path, systemRoot+"/") {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

func belongsToNestedSystem(path string, roots []string) bool {
	for _, root := range roots {
		if root == "." || root == "" {
			continue
		}
		if path == root || strings.HasPrefix(path, root+"/") {
			return true
		}
	}
	return false
}

func buildPurposeFindings(root, systemRoot, kind string, files []analysisFile) []analysisFinding {
	if summary, evidence := readSystemSummary(root, systemRoot, files); summary != "" {
		return []analysisFinding{newFinding(summary, confidenceForEvidence(evidence), evidence)}
	}
	switch kind {
	case "frontend":
		evidence := preferredEvidence(files, "package.json", "src/main.tsx", "src/main.jsx")
		return []analysisFinding{newFinding("A browser-facing frontend is present with a dedicated package manifest and client bootstrap.", "high", evidence)}
	case "backend", "go-service":
		entryPoints := systemEntryPointPaths(root, systemRoot, kind)
		evidence := append(preferredEvidence(files, "go.mod"), entryPoints...)
		return []analysisFinding{newFinding(fmt.Sprintf("A Go-based service surface is present under `%s`.", displaySystemRoot(systemRoot)), confidenceForEvidence(evidence), evidence)}
	case "infra":
		evidence := preferredEvidence(files, "docker-compose.yml", "docker-compose.yaml")
		if len(evidence) == 0 {
			evidence = suffixEvidence(files, ".tf")
		}
		return []analysisFinding{newFinding("Deployment or runtime infrastructure is tracked separately from application code.", confidenceForEvidence(evidence), evidence)}
	default:
		evidence := preferredEvidence(files, "README.md", "README.txt")
		if len(evidence) == 0 && len(files) > 0 {
			evidence = []string{files[0].Path}
		}
		return []analysisFinding{newFinding(fmt.Sprintf("System `%s` was inferred from repository structure and authoritative manifests.", displaySystemRoot(systemRoot)), confidenceForEvidence(evidence), evidence)}
	}
}

func buildEntryPointFindings(root, systemRoot, kind string) []analysisFinding {
	var bullets []string
	subroot := systemAbsRoot(root, systemRoot)
	switch kind {
	case "frontend":
		bullets = buildNodeEntryPoints(subroot)
	case "backend", "go-service":
		bullets = buildGoEntryPoints(subroot)
	case "python-service":
		bullets = buildPythonEntryPoints(subroot)
	default:
		if exists(filepath.Join(subroot, "package.json")) {
			bullets = buildNodeEntryPoints(subroot)
		}
	}
	if len(bullets) == 0 {
		return []analysisFinding{newFinding("No conventional application entry point was detected automatically; this may be a library, workspace root, or thin infrastructure slice.", "low", fallbackSystemEvidence(systemRoot))}
	}
	var findings []analysisFinding
	for _, bullet := range bullets {
		rel, summary := parseBulletFinding(bullet)
		path := prefixSystemPath(systemRoot, rel)
		claim := summary
		if rel != "" && summary != "" {
			claim = fmt.Sprintf("`%s`: %s", path, summary)
		}
		findings = append(findings, newFinding(claim, "high", []string{path}))
	}
	return findings
}

func buildModuleFindings(systemRoot string, files []analysisFile) []analysisFinding {
	dirs := map[string]bool{}
	for _, file := range files {
		if file.Category == "generated" {
			continue
		}
		if file.Path == systemRoot {
			continue
		}
		rel := file.Path
		if systemRoot != "." && systemRoot != "" {
			rel = strings.TrimPrefix(file.Path, systemRoot)
		}
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			continue
		}
		if !strings.Contains(rel, "/") {
			continue
		}
		segment := strings.Split(rel, "/")[0]
		switch segment {
		case ".git", ".namba", ".codex", ".agents", "node_modules", "dist", "vendor":
			continue
		}
		dirs[segment] = true
	}
	var modules []string
	for dir := range dirs {
		modules = append(modules, dir)
	}
	sort.Strings(modules)
	if len(modules) == 0 {
		return []analysisFinding{newFinding("Module boundaries are thin or rooted directly at the system entry files.", "low", fallbackSystemEvidence(systemRoot))}
	}
	limit := minInt(len(modules), 5)
	var findings []analysisFinding
	for _, dir := range modules[:limit] {
		path := prefixSystemPath(systemRoot, dir)
		findings = append(findings, newFinding(fmt.Sprintf("`%s` is a visible module boundary in this system.", path), "medium", []string{path}))
	}
	return findings
}

func buildDataStateFindings(systemRoot string, files []analysisFile) []analysisFinding {
	for _, file := range files {
		if file.Category == "generated" {
			continue
		}
		switch {
		case strings.HasPrefix(file.Path, ".namba/"):
			return []analysisFinding{newFinding("Generated project state is persisted under `.namba`, including manifests and project documents.", "high", []string{".namba/manifest.json"})}
		case strings.HasSuffix(file.Path, ".db"), strings.HasSuffix(file.Path, ".sqlite"), strings.HasSuffix(file.Path, ".sql"):
			return []analysisFinding{newFinding("Persistent data artifacts were detected directly in the repository.", "high", []string{file.Path})}
		case strings.Contains(strings.ToLower(file.Path), "routes"):
			return []analysisFinding{newFinding("Route state is encoded in dedicated routing modules rather than only in inline entrypoint code.", "medium", []string{file.Path})}
		}
	}
	return []analysisFinding{newFinding("No obvious durable state store was detected by the v1 heuristics; treat this as an inference gap rather than proof of statelessness.", "low", fallbackSystemEvidence(systemRoot))}
}

func buildAuthIntegrationFindings(systemRoot string, files []analysisFile) []analysisFinding {
	if evidence := preferredEvidence(files, "config.toml", "codex-review-request.yml", "ci.yml", "SECURITY.md"); len(evidence) > 0 {
		return []analysisFinding{newFinding("Codex, GitHub workflow, or security-policy integrations are explicitly represented in the repository surface.", confidenceForEvidence(evidence), evidence)}
	}
	for _, file := range files {
		if file.Category == "generated" {
			continue
		}
		lower := strings.ToLower(file.Path)
		switch {
		case strings.Contains(lower, "auth"), strings.Contains(lower, "oauth"), strings.Contains(lower, "jwt"), strings.Contains(lower, "session"):
			return []analysisFinding{newFinding("Authentication or authorization code paths are explicitly named in the repository.", "high", []string{file.Path})}
		case strings.Contains(lower, "github"), strings.Contains(lower, "gmail"), strings.Contains(lower, "calendar"), strings.Contains(lower, "codex"):
			return []analysisFinding{newFinding("External tool or platform integrations are part of the current system surface.", "medium", []string{file.Path})}
		}
	}
	return []analysisFinding{newFinding("No explicit auth boundary matched the current heuristics; absence here should be treated as low-confidence inference.", "low", fallbackSystemEvidence(systemRoot))}
}

func buildRiskFindings(systemRoot string, files []analysisFile) []analysisFinding {
	var testEvidence []string
	hasTests := false
	for _, file := range files {
		if file.Category == "test" {
			hasTests = true
			testEvidence = append(testEvidence, file.Path)
			if len(testEvidence) == 3 {
				break
			}
		}
	}
	if !hasTests {
		return []analysisFinding{newFinding("Automated regression coverage is not obvious inside this system slice, so behavior changes may rely on cross-system validation.", "medium", fallbackSystemEvidence(systemRoot))}
	}
	return []analysisFinding{newFinding("System-local regression coverage exists, but end-to-end drift across generated planning docs still needs command-level validation.", "medium", testEvidence)}
}

func detectAnalysisConflicts(projectCfg projectConfig, readmePath, readmeBody string, files []analysisFile, systems []analysisSystem, cfg analysisConfig) []analysisConflict {
	summary := docSummary(readmeBody)
	if summary == "" {
		return nil
	}
	lower := strings.ToLower(summary)
	var conflicts []analysisConflict

	if signals := runtimeSignals(projectCfg, files, systems); len(signals) > 0 {
		if contradictory := contradictoryRuntimeMention(lower, signals); contradictory != "" {
			conflicts = append(conflicts, analysisConflict{
				Claim:    fmt.Sprintf("README describes the repository as `%s`, but the strongest code/config signals point to `%s`.", contradictory, strings.Join(signals, ", ")),
				Stronger: runtimeSignalEvidence(signals, files, systems, cfg),
				Weaker:   readmePath,
				Reason:   "code and authoritative manifests outrank prose documentation in the configured source-priority contract",
			})
		}
	}

	return conflicts
}

func runtimeSignals(projectCfg projectConfig, files []analysisFile, systems []analysisSystem) []string {
	var signals []string
	seen := map[string]bool{}
	addSignal := func(signal string) {
		if signal == "" || seen[signal] {
			return
		}
		seen[signal] = true
		signals = append(signals, signal)
	}
	for _, system := range systems {
		addSignal(runtimeSignalForSystemKind(system.Kind))
	}
	if len(signals) > 0 {
		return signals
	}
	switch strings.ToLower(projectCfg.Language) {
	case "go":
		addSignal("go service")
	case "python":
		addSignal("python service")
	case "javascript", "typescript", "node":
		addSignal("node or frontend runtime")
	default:
		if anyFileWithBase(files, "go.mod") {
			addSignal("go service")
		}
		if anyFileWithBase(files, "package.json") {
			addSignal("node or frontend runtime")
		}
	}
	return signals
}

func runtimeSignalForSystemKind(kind string) string {
	switch kind {
	case "frontend":
		return "react frontend"
	case "backend", "go-service":
		return "go service"
	case "infra":
		return "infrastructure stack"
	case "python-service":
		return "python service"
	default:
		return ""
	}
}

func contradictoryRuntimeMention(readme string, signals []string) string {
	mentions := []runtimeMention{
		{Label: "rails", Patterns: []string{"rails"}},
		{Label: "django", Patterns: []string{"django"}, CompatibleSignals: []string{"python service"}},
		{Label: "spring", Patterns: []string{"spring"}},
		{Label: "fastapi", Patterns: []string{"fastapi"}, CompatibleSignals: []string{"python service"}},
		{Label: "python service", Patterns: []string{"python service", "python backend"}, CompatibleSignals: []string{"python service"}},
		{Label: "react frontend", Patterns: []string{"react frontend", "react app", "single-page app", "spa frontend"}, CompatibleSignals: []string{"react frontend", "node or frontend runtime"}},
		{Label: "go service", Patterns: []string{"go service", "golang service", "cobra cli"}, CompatibleSignals: []string{"go service"}},
		{Label: "infrastructure stack", Patterns: []string{"terraform", "docker compose", "infrastructure"}, CompatibleSignals: []string{"infrastructure stack"}},
	}

	for _, mention := range mentions {
		if !containsAny(readme, mention.Patterns...) {
			continue
		}
		if runtimeMentionSupported(signals, mention) {
			continue
		}
		return mention.Label
	}
	return ""
}

func runtimeMentionSupported(signals []string, mention runtimeMention) bool {
	compatible := mention.CompatibleSignals
	if len(compatible) == 0 {
		compatible = []string{mention.Label}
	}
	for _, signal := range signals {
		for _, candidate := range compatible {
			if signal == candidate {
				return true
			}
		}
	}
	return false
}

func containsAny(text string, patterns ...string) bool {
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

func runtimeSignalEvidence(signals []string, files []analysisFile, systems []analysisSystem, cfg analysisConfig) string {
	type candidate struct {
		path string
		rank int
	}
	var candidates []candidate
	allRoots := collectSystemRoots(files)
	appendEvidence := func(paths []string) {
		for _, evidence := range paths {
			candidates = append(candidates, candidate{
				path: evidence,
				rank: sourcePriorityRank(cfg, evidenceCategory(files, evidence)),
			})
		}
	}
	for _, signal := range signals {
		switch signal {
		case "go service":
			appendEvidence(preferredEvidence(files, "go.mod", "main.go"))
		case "react frontend", "node or frontend runtime":
			appendEvidence(preferredEvidence(files, "package.json", "src/main.tsx", "src/main.jsx", "src/app/App.tsx"))
		case "python service":
			appendEvidence(preferredEvidence(files, "pyproject.toml", "requirements.txt", "main.py"))
		case "infrastructure stack":
			appendEvidence(preferredEvidence(files, "docker-compose.yml", "docker-compose.yaml", "kustomization.yaml"))
		}
	}
	for _, system := range systems {
		signal := runtimeSignalForSystemKind(system.Kind)
		if signal == "" || !containsString(signals, signal) {
			continue
		}
		switch signal {
		case "react frontend":
			appendEvidence(preferredEvidence(filesForSystem(files, system.Root, allRoots), "package.json", "src/main.tsx", "src/main.jsx", "src/app/App.tsx"))
		case "go service":
			appendEvidence(preferredEvidence(filesForSystem(files, system.Root, allRoots), "go.mod", "main.go"))
		case "python service":
			appendEvidence(preferredEvidence(filesForSystem(files, system.Root, allRoots), "pyproject.toml", "requirements.txt", "main.py"))
		case "infrastructure stack":
			appendEvidence(preferredEvidence(filesForSystem(files, system.Root, allRoots), "docker-compose.yml", "docker-compose.yaml", "kustomization.yaml"))
		}
		if len(system.Purpose) > 0 && len(system.Purpose[0].Evidence) > 0 {
			appendEvidence(system.Purpose[0].Evidence)
		}
	}
	if len(candidates) == 0 {
		for _, file := range files {
			candidates = append(candidates, candidate{
				path: file.Path,
				rank: sourcePriorityRank(cfg, file.Category),
			})
		}
	}
	if len(candidates) == 0 {
		return "repository scan"
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].rank != candidates[j].rank {
			return candidates[i].rank < candidates[j].rank
		}
		return candidates[i].path < candidates[j].path
	})
	var evidence []string
	seen := map[string]bool{}
	for _, item := range candidates {
		if item.path == "." || seen[item.path] {
			continue
		}
		seen[item.path] = true
		evidence = append(evidence, item.path)
		if len(evidence) == 2 {
			break
		}
	}
	if len(evidence) == 0 {
		return "repository scan"
	}
	return strings.Join(evidence, ", ")
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func evaluateAnalysisQuality(analysis projectAnalysis) analysisQuality {
	var quality analysisQuality
	if len(analysis.Files) == 0 {
		quality.Errors = append(quality.Errors, "No analyzable files matched the configured scope. Check `.namba/config/sections/analysis.yaml` include/exclude paths.")
		return quality
	}
	if len(analysis.Systems) == 0 {
		quality.Errors = append(quality.Errors, "No system boundaries or analyzable repository surface were detected.")
		return quality
	}
	for _, system := range analysis.Systems {
		if totalSystemFindings(system) < 5 {
			quality.Warnings = append(quality.Warnings, fmt.Sprintf("System `%s` is thin; fewer than five evidence-backed findings were produced.", system.Name))
		}
		if lowConfidenceCount(system) > highMediumConfidenceCount(system) {
			quality.Warnings = append(quality.Warnings, fmt.Sprintf("System `%s` is dominated by low-confidence inference; add stronger code/config signals or tune analysis scope.", system.Name))
		}
	}
	if len(analysis.Conflicts) == 0 && analysis.ReadmePath != "" {
		quality.Warnings = append(quality.Warnings, "No code-vs-doc conflicts were detected. This may be correct, but the v1 conflict heuristics remain intentionally narrow.")
	}
	return quality
}

func renderProductDoc(analysis projectAnalysis) string {
	lines := []string{
		"# Product",
		"",
		"## Planning Context",
		"",
		fmt.Sprintf("- Primary landing doc for `%s`.", analysis.Project.Name),
		fmt.Sprintf("- Output template: `%s`.", analysis.Config.OutputTemplate),
		fmt.Sprintf("- Source priority: %s.", strings.Join(analysis.Config.SourcePriority, " > ")),
		"",
		"## System Landscape",
		"",
	}
	for _, system := range analysis.Systems {
		lines = append(lines, fmt.Sprintf("- `%s`: %s (`.namba/project/systems/%s.md`)", system.Name, firstClaim(system.Purpose), system.Slug))
	}
	lines = append(lines, "",
		"## Evidence Highlights",
		"",
	)
	for _, finding := range collectTopFindings(analysis.Systems, 5) {
		lines = append(lines, renderFindingBullet(finding))
	}
	lines = append(lines, "",
		"## Quality And Drift",
		"",
		fmt.Sprintf("- Mismatch report: `.namba/project/mismatch-report.md` (%d conflicts).", len(analysis.Conflicts)),
		fmt.Sprintf("- Quality report: `.namba/project/quality-report.md` (%d warnings, %d errors).", len(analysis.Quality.Warnings), len(analysis.Quality.Errors)),
	)
	if analysis.ReadmePath != "" {
		snapshot := docSummary(analysis.ReadmeBody)
		if snapshot == "" {
			snapshot = firstNonEmptyLine(analysis.ReadmeBody)
		}
		lines = append(lines, "",
			"## Source Snapshot",
			"",
			fmt.Sprintf("- Source: `%s`", analysis.ReadmePath),
			fmt.Sprintf("- Summary excerpt: %s", blankFallback(snapshot, "No maintained README summary was detected.")),
			"",
		)
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderTechDoc(analysis projectAnalysis) string {
	lines := []string{
		"# Tech",
		"",
		"## Runtime And Validation",
		"",
		fmt.Sprintf("- Language: %s", analysis.Project.Language),
		fmt.Sprintf("- Framework: %s", analysis.Project.Framework),
		fmt.Sprintf("- Important roles: %s", strings.Join(analysis.Config.ImportantRoles, ", ")),
		fmt.Sprintf("- Important runtimes/services: %s", strings.Join(analysis.Config.ImportantRuntimes, ", ")),
		fmt.Sprintf("- Validation commands: test=`%s`, lint=`%s`, typecheck=`%s`", blankFallback(analysis.QualityCfg.TestCommand, "none"), blankFallback(analysis.QualityCfg.LintCommand, "none"), blankFallback(analysis.QualityCfg.TypecheckCommand, "none")),
		"",
		"## Analysis Contract",
		"",
		"- Evidence: every generated finding cites one or more repository paths.",
		"- Confidence: `high` means direct code/config support, `medium` means multiple weaker signals or doc-backed interpretation, and `low` means heuristic inference.",
		"- Conflict handling: `mismatch-report.md` preserves stronger-vs-weaker source disagreements instead of flattening them into neutral prose.",
		"",
		"## Systems",
		"",
	}
	for _, system := range analysis.Systems {
		lines = append(lines, fmt.Sprintf("- `%s` (%s): `.namba/project/systems/%s.md`", system.Name, system.Kind, system.Slug))
	}
	lines = append(lines, "",
		"## Planning Signals",
		"",
		fmt.Sprintf("- Configured include paths: %s", strings.Join(analysis.Config.IncludePaths, ", ")),
		fmt.Sprintf("- Configured exclude paths: %s", strings.Join(analysis.Config.ExcludePaths, ", ")),
		"- `structure.md` is appendix material; use `product.md` and the per-system docs first.",
		"- `mismatch-report.md` preserves code-vs-doc conflicts instead of flattening them into prose.",
	)
	return strings.Join(lines, "\n") + "\n"
}

func renderMismatchReportDoc(analysis projectAnalysis) string {
	lines := []string{"# Mismatch Report", ""}
	if len(analysis.Conflicts) == 0 {
		lines = append(lines, "- No explicit code-vs-doc conflicts were detected by the current v1 heuristics.", "")
		return strings.Join(lines, "\n")
	}
	for _, conflict := range analysis.Conflicts {
		lines = append(lines,
			fmt.Sprintf("- Claim: %s", conflict.Claim),
			fmt.Sprintf("  Stronger source: `%s`", conflict.Stronger),
			fmt.Sprintf("  Weaker source: `%s`", conflict.Weaker),
			fmt.Sprintf("  Reason: %s", conflict.Reason),
			"",
		)
	}
	return strings.Join(lines, "\n")
}

func renderQualityReportDoc(analysis projectAnalysis) string {
	lines := []string{"# Quality Report", ""}
	if len(analysis.Quality.Errors) == 0 && len(analysis.Quality.Warnings) == 0 {
		lines = append(lines, "- No warnings or errors were produced by the current quality gate.")
		return strings.Join(lines, "\n") + "\n"
	}
	if len(analysis.Quality.Errors) > 0 {
		lines = append(lines, "## Errors", "")
		for _, item := range analysis.Quality.Errors {
			lines = append(lines, "- "+item)
		}
		lines = append(lines, "")
	}
	if len(analysis.Quality.Warnings) > 0 {
		lines = append(lines, "## Warnings", "")
		for _, item := range analysis.Quality.Warnings {
			lines = append(lines, "- "+item)
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderSystemDoc(system analysisSystem) string {
	lines := []string{
		fmt.Sprintf("# System: %s", system.Name),
		"",
		fmt.Sprintf("- Root: `%s`", displaySystemRoot(system.Root)),
		fmt.Sprintf("- Kind: %s", system.Kind),
		"",
		"## Purpose",
		"",
	}
	lines = appendFindings(lines, system.Purpose)
	lines = append(lines, "", "## Entry Points And Interfaces", "")
	lines = appendFindings(lines, system.EntryPoints)
	lines = append(lines, "", "## Module Boundaries", "")
	lines = appendFindings(lines, system.Modules)
	lines = append(lines, "", "## Data And State", "")
	lines = appendFindings(lines, system.DataState)
	lines = append(lines, "", "## Auth And Integrations", "")
	lines = appendFindings(lines, system.AuthIntegrations)
	lines = append(lines, "", "## Deploy Runtime And Test Risks", "")
	lines = appendFindings(lines, system.Risks)
	return strings.Join(lines, "\n") + "\n"
}

func isProjectAnalysisManagedPath(rel string) bool {
	switch rel {
	case filepath.ToSlash(filepath.Join(projectDir, "product.md")),
		filepath.ToSlash(filepath.Join(projectDir, "tech.md")),
		filepath.ToSlash(filepath.Join(projectDir, "structure.md")),
		filepath.ToSlash(filepath.Join(projectDir, "mismatch-report.md")),
		filepath.ToSlash(filepath.Join(projectDir, "quality-report.md")),
		filepath.ToSlash(filepath.Join(codemapsDir, "overview.md")),
		filepath.ToSlash(filepath.Join(codemapsDir, "entry-points.md")),
		filepath.ToSlash(filepath.Join(codemapsDir, "dependencies.md")),
		filepath.ToSlash(filepath.Join(codemapsDir, "data-flow.md")):
		return true
	default:
		return strings.HasPrefix(rel, filepath.ToSlash(systemDocsDir)+"/")
	}
}

func renderOverviewDoc(analysis projectAnalysis) string {
	lines := []string{"# Overview", ""}
	for _, system := range analysis.Systems {
		lines = append(lines, fmt.Sprintf("- `%s` (%s): %s", system.Name, system.Kind, firstClaim(system.Purpose)))
	}
	lines = append(lines, "",
		fmt.Sprintf("- Conflicts: %d (`.namba/project/mismatch-report.md`)", len(analysis.Conflicts)),
		fmt.Sprintf("- Quality warnings: %d (`.namba/project/quality-report.md`)", len(analysis.Quality.Warnings)),
	)
	return strings.Join(lines, "\n") + "\n"
}

func renderEntryPointsDoc(systems []analysisSystem) string {
	lines := []string{"# Entry Points", ""}
	for _, system := range systems {
		lines = append(lines, fmt.Sprintf("## %s", system.Name), "")
		lines = appendFindings(lines, system.EntryPoints)
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderDependenciesDoc(root string, systems []analysisSystem) string {
	lines := []string{"# Dependencies", ""}
	for _, system := range systems {
		lines = append(lines, fmt.Sprintf("## %s", system.Name), "")
		deps := systemDependencies(root, system)
		if len(deps) == 0 {
			lines = append(lines, "- No dependency manifest was detected automatically.", "")
			continue
		}
		lines = append(lines, deps...)
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderDataFlowDoc(systems []analysisSystem, qualityCfg qualityConfig) string {
	lines := []string{
		"# Data Flow",
		"",
		"1. The analyzer scopes repository inputs using `.namba/config/sections/analysis.yaml`.",
		"2. System boundaries are inferred before summarization so multi-system repos are not flattened into one tree dump.",
		"3. Evidence-backed findings are rendered into `product.md`, `tech.md`, and per-system docs.",
		"4. Mismatches and thin-output signals are emitted as first-class artifacts.",
		fmt.Sprintf("5. Validation still runs through the configured commands (`%s`, `%s`, `%s`).", blankFallback(qualityCfg.TestCommand, "none"), blankFallback(qualityCfg.LintCommand, "none"), blankFallback(qualityCfg.TypecheckCommand, "none")),
		"",
		"## System State Signals",
		"",
	}
	for _, system := range systems {
		lines = append(lines, fmt.Sprintf("### %s", system.Name), "")
		lines = appendFindings(lines, system.DataState)
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n") + "\n"
}

func readSystemSummary(root, systemRoot string, files []analysisFile) (string, []string) {
	for _, candidate := range []string{"README.md", "README.txt"} {
		path := candidate
		if systemRoot != "." {
			path = filepath.ToSlash(filepath.Join(systemRoot, candidate))
		}
		if !analysisFileExists(files, path) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			continue
		}
		if summary := docSummary(string(data)); summary != "" {
			return summary, []string{path}
		}
	}
	return "", nil
}

func renderAnalysisStructure(files []analysisFile) string {
	lines := []string{"# Structure", "", "Appendix output only. Use `product.md` and `tech.md` first.", "", "```"}
	for _, file := range files {
		lines = append(lines, file.Path)
	}
	lines = append(lines, "```", "")
	return strings.Join(lines, "\n")
}

func systemDependencies(root string, system analysisSystem) []string {
	subroot := systemAbsRoot(root, system.Root)
	switch system.Kind {
	case "frontend":
		return buildNodeDependencies(subroot)
	case "backend", "go-service":
		return buildStaticDependenciesBullets(subroot, projectConfig{Language: "go"})
	case "python-service":
		return buildStaticDependenciesBullets(subroot, projectConfig{Language: "python"})
	default:
		if exists(filepath.Join(subroot, "package.json")) {
			return buildNodeDependencies(subroot)
		}
		if exists(filepath.Join(subroot, "go.mod")) {
			return buildStaticDependenciesBullets(subroot, projectConfig{Language: "go"})
		}
		return nil
	}
}

func buildStaticDependenciesBullets(root string, cfg projectConfig) []string {
	switch cfg.Language {
	case "go":
		var bullets []string
		if module := readGoModule(root); module != "" {
			bullets = append(bullets, fmt.Sprintf("- Module: `%s`", module))
		}
		bullets = append(bullets, "- Runtime: Go standard library")
		bullets = append(bullets, "- External runtime: Codex CLI")
		bullets = append(bullets, "- External runtime: Git")
		return bullets
	case "python":
		var bullets []string
		if exists(filepath.Join(root, "pyproject.toml")) {
			bullets = append(bullets, "- Dependency manifest: `pyproject.toml`")
		}
		if exists(filepath.Join(root, "requirements.txt")) {
			bullets = append(bullets, "- Dependency manifest: `requirements.txt`")
		}
		return bullets
	default:
		return nil
	}
}

func appendFindings(lines []string, findings []analysisFinding) []string {
	for _, finding := range findings {
		lines = append(lines, renderFindingBullet(finding))
	}
	return lines
}

func renderFindingBullet(finding analysisFinding) string {
	return fmt.Sprintf("- %s Confidence: %s. Evidence: %s.", finding.Claim, finding.Confidence, renderEvidenceList(finding.Evidence))
}

func newFinding(claim, confidence string, evidence []string) analysisFinding {
	if len(evidence) == 0 {
		evidence = []string{"repository scan"}
	}
	return analysisFinding{
		Claim:      strings.TrimSpace(claim),
		Confidence: blankFallback(confidence, "low"),
		Evidence:   evidence,
	}
}

func parseBulletFinding(bullet string) (string, string) {
	text := strings.TrimSpace(strings.TrimPrefix(bullet, "- "))
	if strings.HasPrefix(text, "`") {
		parts := strings.SplitN(text, "`:", 2)
		if len(parts) == 2 {
			return strings.TrimPrefix(parts[0], "`"), strings.TrimSpace(parts[1])
		}
	}
	return text, text
}

func systemEntryPointPaths(root, systemRoot, kind string) []string {
	subroot := systemAbsRoot(root, systemRoot)
	var bullets []string
	switch kind {
	case "frontend":
		bullets = buildNodeEntryPoints(subroot)
	case "backend", "go-service":
		bullets = buildGoEntryPoints(subroot)
	}
	var paths []string
	for _, bullet := range bullets {
		rel, _ := parseBulletFinding(bullet)
		if rel != "" {
			paths = append(paths, prefixSystemPath(systemRoot, rel))
		}
	}
	return paths
}

func totalSystemFindings(system analysisSystem) int {
	return len(system.Purpose) + len(system.EntryPoints) + len(system.Modules) + len(system.DataState) + len(system.AuthIntegrations) + len(system.Risks)
}

func lowConfidenceCount(system analysisSystem) int {
	count := 0
	for _, bucket := range [][]analysisFinding{system.Purpose, system.EntryPoints, system.Modules, system.DataState, system.AuthIntegrations, system.Risks} {
		for _, finding := range bucket {
			if finding.Confidence == "low" {
				count++
			}
		}
	}
	return count
}

func highMediumConfidenceCount(system analysisSystem) int {
	count := 0
	for _, bucket := range [][]analysisFinding{system.Purpose, system.EntryPoints, system.Modules, system.DataState, system.AuthIntegrations, system.Risks} {
		for _, finding := range bucket {
			if finding.Confidence == "high" || finding.Confidence == "medium" {
				count++
			}
		}
	}
	return count
}

func preferredEvidence(files []analysisFile, names ...string) []string {
	var evidence []string
	for _, name := range names {
		for _, file := range files {
			if filepath.Base(file.Path) == name {
				evidence = append(evidence, file.Path)
				break
			}
		}
	}
	return evidence
}

func suffixEvidence(files []analysisFile, suffix string) []string {
	var evidence []string
	for _, file := range files {
		if strings.HasSuffix(file.Path, suffix) {
			evidence = append(evidence, file.Path)
		}
	}
	return evidence
}

func confidenceForEvidence(evidence []string) string {
	if len(evidence) >= 2 {
		return "high"
	}
	if len(evidence) == 1 {
		return "medium"
	}
	return "low"
}

func fallbackSystemEvidence(systemRoot string) []string {
	return []string{displaySystemRoot(systemRoot)}
}

func analysisFileExists(files []analysisFile, path string) bool {
	path = filepath.ToSlash(path)
	for _, file := range files {
		if file.Path == path {
			return true
		}
	}
	return false
}

func prefixSystemPath(systemRoot, rel string) string {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	if systemRoot == "." || systemRoot == "" {
		return rel
	}
	if rel == "" {
		return systemRoot
	}
	return filepath.ToSlash(filepath.Join(systemRoot, rel))
}

func displaySystemRoot(root string) string {
	if root == "." || root == "" {
		return "."
	}
	return filepath.ToSlash(root)
}

func slugifySystemRoot(root string) string {
	if root == "." || root == "" {
		return "workspace"
	}
	root = strings.ToLower(filepath.ToSlash(root))
	replacer := strings.NewReplacer("/", "-", "_", "-", ".", "-")
	root = replacer.Replace(root)
	root = strings.Trim(root, "-")
	if root == "" {
		return "system"
	}
	return root
}

func systemAbsRoot(root, systemRoot string) string {
	if systemRoot == "." || systemRoot == "" {
		return root
	}
	return filepath.Join(root, filepath.FromSlash(systemRoot))
}

func anyPathMatches(paths []string, candidates ...string) bool {
	for _, path := range paths {
		for _, candidate := range candidates {
			if filepath.Base(path) == candidate {
				return true
			}
		}
	}
	return false
}

func anyPathContains(paths []string, candidates ...string) bool {
	for _, path := range paths {
		for _, candidate := range candidates {
			if strings.Contains(path, candidate) {
				return true
			}
		}
	}
	return false
}

func anyPathSuffix(paths []string, suffix string) bool {
	for _, path := range paths {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}
	return false
}

func anyFileWithBase(files []analysisFile, base string) bool {
	for _, file := range files {
		if filepath.Base(file.Path) == base {
			return true
		}
	}
	return false
}

func firstClaim(findings []analysisFinding) string {
	if len(findings) == 0 {
		return "No summary available."
	}
	return findings[0].Claim
}

func collectTopFindings(systems []analysisSystem, limit int) []analysisFinding {
	var findings []analysisFinding
	for _, system := range systems {
		findings = append(findings, system.Purpose...)
		findings = append(findings, system.EntryPoints...)
	}
	if len(findings) > limit {
		return findings[:limit]
	}
	return findings
}

func blankFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func evidenceCategory(files []analysisFile, path string) string {
	path = filepath.ToSlash(strings.TrimSpace(path))
	for _, file := range files {
		if file.Path == path {
			return file.Category
		}
	}
	return classifyAnalysisFile(path)
}

func sourcePriorityRank(cfg analysisConfig, category string) int {
	for idx, item := range cfg.SourcePriority {
		if strings.EqualFold(strings.TrimSpace(item), category) {
			return idx
		}
	}
	return len(cfg.SourcePriority) + 1
}

func docSummary(body string) string {
	lines := strings.Split(body, "\n")
	var summary []string
	started := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		switch {
		case line == "":
			if started && len(summary) > 0 {
				return strings.Join(summary, " ")
			}
			continue
		case strings.HasPrefix(line, "<!--"), strings.HasPrefix(line, "<p"), strings.HasPrefix(line, "</p>"), strings.HasPrefix(line, "<img"), strings.HasPrefix(line, "[English]"), strings.HasPrefix(line, "[Latest Release]"):
			continue
		case strings.HasPrefix(line, "#"):
			continue
		case strings.HasPrefix(line, "```"):
			return strings.Join(summary, " ")
		default:
			started = true
			summary = append(summary, strings.TrimSpace(line))
			if len(summary) == 2 {
				return strings.Join(summary, " ")
			}
		}
	}
	return strings.Join(summary, " ")
}

func renderEvidenceList(evidence []string) string {
	if len(evidence) == 0 {
		return "`repository scan`"
	}
	formatted := make([]string, 0, len(evidence))
	for _, item := range evidence {
		formatted = append(formatted, fmt.Sprintf("`%s`", item))
	}
	return strings.Join(formatted, ", ")
}
