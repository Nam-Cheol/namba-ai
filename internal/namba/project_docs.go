package namba

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func detectProjectType(root string) string {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "new"
	}
	if len(entries) == 0 {
		return "new"
	}
	return "existing"
}

func buildStructureDoc(root string) string {
	lines := []string{"# Structure", "", "```"}
	appendStructureEntries(&lines, root, "", 0, 2)
	lines = append(lines, "```", "")
	return strings.Join(lines, "\n")
}

func appendStructureEntries(lines *[]string, root, rel string, depth, maxDepth int) {
	entries, err := os.ReadDir(filepath.Join(root, rel))
	if err != nil {
		return
	}
	for _, entry := range entries {
		childRel := entry.Name()
		if rel != "" {
			childRel = filepath.Join(rel, entry.Name())
		}
		if shouldSkipStructureEntry(filepath.ToSlash(childRel)) {
			continue
		}
		*lines = append(*lines, childRel)
		if entry.IsDir() && depth < maxDepth {
			appendStructureEntries(lines, root, childRel, depth+1, maxDepth)
		}
	}
}

func shouldSkipStructureEntry(rel string) bool {
	switch {
	case rel == ".git", strings.HasPrefix(rel, ".git/"):
		return true
	case rel == ".cache", strings.HasPrefix(rel, ".cache/"):
		return true
	case rel == ".gocache", strings.HasPrefix(rel, ".gocache/"):
		return true
	case rel == ".codex/skills", strings.HasPrefix(rel, ".codex/skills/"):
		return true
	case rel == ".tmp", strings.HasPrefix(rel, ".tmp/"):
		return true
	case rel == "dist", strings.HasPrefix(rel, "dist/"):
		return true
	case rel == "external", strings.HasPrefix(rel, "external/"):
		return true
	case rel == ".namba/logs", strings.HasPrefix(rel, ".namba/logs/"):
		return true
	case rel == ".namba/project/change-summary.md":
		return true
	case rel == ".namba/project/pr-checklist.md":
		return true
	case rel == ".namba/project/release-checklist.md":
		return true
	case rel == ".namba/project/release-notes.md":
		return true
	case rel == ".namba/worktrees", strings.HasPrefix(rel, ".namba/worktrees/"):
		return true
	case strings.HasSuffix(rel, ".exe"):
		return true
	case rel == "namba":
		return true
	default:
		return false
	}
}

func buildTechDoc(cfg projectConfig) string {
	return fmt.Sprintf("# Tech\n\n- Language: %s\n- Framework: %s\n- Runtime adapter: Codex\n- Repo-local skills and command-entry skills: .agents/skills\n- Repo-local Codex config: .codex/config.toml\n- Built-in Codex subagents: default, worker, explorer\n- Project-scoped custom agents: .codex/agents/*.toml\n- Readable agent mirrors: .codex/agents/*.md\n- State directory: .namba\n", cfg.Language, cfg.Framework)
}

type jsImportInfo struct {
	Resolved string
	Bindings []string
}

var localJSImportPattern = regexp.MustCompile(`(?m)^\s*import\s+(?:(.+?)\s+from\s+)?["']([^"']+)["']`)

var renderJSXComponentPattern = regexp.MustCompile(`(?s)\.render\(\s*<([A-Z][A-Za-z0-9_]*)\b`)

var renderIdentifierPattern = regexp.MustCompile(`(?s)\.render\(\s*([A-Za-z_][A-Za-z0-9_]*)\s*\)`)

func buildCodemaps(root string, cfg projectConfig) (string, string, string, string) {
	overview := fmt.Sprintf("# Overview\n\n%s is managed by NambaAI.\n\n- Language: %s\n- Framework: %s\n", cfg.Name, cfg.Language, cfg.Framework)
	entries := buildEntryPointsDoc(root, cfg)
	deps := buildDependenciesDoc(root, cfg)
	flow := "# Data Flow\n\n1. `init` runs a Codex-adapted project wizard, writes `.namba/config/sections/*.yaml`, repo skills under `.agents/skills`, command-entry skills such as `$namba-run`, project-scoped custom agents under `.codex/agents/*.toml`, readable `.md` agent mirrors, and Codex repo config under `.codex/config.toml`\n2. `project` refreshes docs and codemaps\n3. `plan` creates a SPEC package\n4. `run` supports the default standalone flow, explicit `--solo` and `--team` subagent-oriented requests, and worktree-based `--parallel` execution\n5. `sync` emits PR-ready artifacts\n"
	return overview, entries, deps, flow
}

func buildEntryPointsDoc(root string, cfg projectConfig) string {
	lines := []string{"# Entry Points", ""}
	var bullets []string

	switch cfg.Language {
	case "go":
		bullets = buildGoEntryPoints(root)
	case "java":
		bullets = buildJavaEntryPoints(root)
	case "python":
		bullets = buildPythonEntryPoints(root)
	default:
		bullets = buildNodeEntryPoints(root)
	}

	if len(bullets) == 0 {
		bullets = append(bullets, "- No conventional application entry point was detected automatically.")
	}

	lines = append(lines, bullets...)
	return strings.Join(lines, "\n") + "\n"
}

func buildDependenciesDoc(root string, cfg projectConfig) string {
	lines := []string{"# Dependencies", ""}
	var bullets []string

	switch cfg.Language {
	case "go":
		if module := readGoModule(root); module != "" {
			bullets = append(bullets, fmt.Sprintf("- Module: `%s`", module))
		}
		bullets = append(bullets, "- Runtime: Go standard library")
		bullets = append(bullets, "- External runtime: Codex CLI")
		bullets = append(bullets, "- External runtime: Git")
	case "java":
		if exists(filepath.Join(root, "pom.xml")) {
			bullets = append(bullets, "- Build system: Maven (`pom.xml`)")
		}
		if exists(filepath.Join(root, "build.gradle")) || exists(filepath.Join(root, "build.gradle.kts")) || exists(filepath.Join(root, "gradlew")) || exists(filepath.Join(root, "gradlew.bat")) {
			bullets = append(bullets, "- Build system: Gradle")
		}
	case "python":
		if exists(filepath.Join(root, "pyproject.toml")) {
			bullets = append(bullets, "- Dependency manifest: `pyproject.toml`")
		}
		if exists(filepath.Join(root, "requirements.txt")) {
			bullets = append(bullets, "- Dependency manifest: `requirements.txt`")
		}
	default:
		bullets = buildNodeDependencies(root)
	}

	if len(bullets) == 0 {
		bullets = append(bullets, "- No dependency manifest was detected automatically.")
	}

	lines = append(lines, bullets...)
	return strings.Join(lines, "\n") + "\n"
}

func buildGoEntryPoints(root string) []string {
	seen := map[string]bool{}
	var bullets []string

	if mainFile := firstExisting(root, "main.go"); mainFile != "" {
		appendEntryPoint(&bullets, seen, mainFile, "application bootstrap")
	}

	cmdDir := filepath.Join(root, "cmd")
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return bullets
	}

	var candidates []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.ToSlash(filepath.Join("cmd", entry.Name(), "main.go"))
		if exists(filepath.Join(root, candidate)) {
			candidates = append(candidates, candidate)
		}
	}
	sort.Strings(candidates)
	for _, candidate := range candidates {
		appendEntryPoint(&bullets, seen, candidate, "Go command entry point")
	}

	return bullets
}

func buildJavaEntryPoints(root string) []string {
	var matches []string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if shouldSkipDiscoveryDir(root, path) {
				return filepath.SkipDir
			}
			return nil
		}
		base := strings.ToLower(filepath.Base(path))
		if strings.HasSuffix(base, "application.java") || base == "main.java" {
			rel, relErr := filepath.Rel(root, path)
			if relErr == nil {
				matches = append(matches, filepath.ToSlash(rel))
			}
		}
		return nil
	})

	sort.Strings(matches)
	seen := map[string]bool{}
	var bullets []string
	for _, match := range matches {
		appendEntryPoint(&bullets, seen, match, "Java application bootstrap")
	}
	return bullets
}

func buildPythonEntryPoints(root string) []string {
	seen := map[string]bool{}
	var bullets []string
	for _, candidate := range []string{"main.py", "app.py", "manage.py", "src/main.py"} {
		if exists(filepath.Join(root, candidate)) {
			appendEntryPoint(&bullets, seen, candidate, "Python application entry point")
		}
	}
	return bullets
}

func buildNodeEntryPoints(root string) []string {
	seen := map[string]bool{}
	var bullets []string

	bootstrap := firstExisting(root, "src/main.tsx", "src/main.jsx", "src/index.tsx", "src/index.jsx", "main.tsx", "main.jsx", "src/main.ts", "src/main.js")
	if bootstrap == "" {
		bootstrap = firstFileContaining(root, "createRoot(")
	}
	if bootstrap != "" {
		appendEntryPoint(&bullets, seen, bootstrap, summarizeEntryPoint(root, bootstrap))
		if appShell := resolveNodeAppShell(root, bootstrap); appShell != "" {
			appendEntryPoint(&bullets, seen, appShell, summarizeEntryPoint(root, appShell))
			if routerModule := firstRouterLikeJSImport(root, appShell); routerModule != "" {
				appendEntryPoint(&bullets, seen, routerModule, summarizeEntryPoint(root, routerModule))
			}
		}
	}

	for _, candidate := range []string{"src/app/routes.ts", "src/app/routes.tsx", "src/routes.ts", "src/routes.tsx", "src/router.ts", "src/router.tsx"} {
		if exists(filepath.Join(root, candidate)) {
			appendEntryPoint(&bullets, seen, candidate, summarizeEntryPoint(root, candidate))
		}
	}

	if routerModule := firstFileContaining(root, "createBrowserRouter("); routerModule != "" {
		appendEntryPoint(&bullets, seen, routerModule, summarizeEntryPoint(root, routerModule))
	}

	return bullets
}

func buildNodeDependencies(root string) []string {
	pkg, err := loadPackageManifest(root)
	if err != nil {
		return nil
	}

	var bullets []string

	if runtime := collectPackageVersions(pkg, "react", "react-dom", "react-router", "react-router-dom", "next"); len(runtime) > 0 {
		bullets = append(bullets, "- Runtime: "+strings.Join(runtime, ", "))
	}

	if build := collectPackageVersions(pkg, "vite", "@vitejs/plugin-react", "typescript", "tailwindcss", "@tailwindcss/vite"); len(build) > 0 {
		bullets = append(bullets, "- Build and styling: "+strings.Join(build, ", "))
	}

	var ui []string
	ui = append(ui, collectPackageVersions(pkg, "@mui/material", "@mui/icons-material", "@emotion/react", "@emotion/styled", "lucide-react")...)
	if radixCount := countPackagesWithPrefix(pkg, "@radix-ui/"); radixCount > 0 {
		ui = append(ui, fmt.Sprintf("%d Radix UI primitives", radixCount))
	}
	if len(ui) > 0 {
		bullets = append(bullets, "- UI system: "+strings.Join(ui, ", "))
	}

	if feature := collectPackageVersions(pkg, "motion", "recharts", "react-hook-form", "date-fns", "sonner", "embla-carousel-react"); len(feature) > 0 {
		bullets = append(bullets, "- Product features: "+strings.Join(feature, ", "))
	}

	if len(bullets) == 0 {
		if fallback := topPackageVersions(pkg, 8); len(fallback) > 0 {
			bullets = append(bullets, "- Packages: "+strings.Join(fallback, ", "))
		}
	}

	return bullets
}

func appendEntryPoint(lines *[]string, seen map[string]bool, rel, summary string) {
	rel = filepath.ToSlash(rel)
	if rel == "" || seen[rel] {
		return
	}
	*lines = append(*lines, fmt.Sprintf("- `%s`: %s", rel, summary))
	seen[rel] = true
}

func summarizeEntryPoint(root, rel string) string {
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		return "application module"
	}

	text := strings.ToLower(string(data))
	switch {
	case strings.Contains(text, "createroot("):
		return "React DOM bootstrap"
	case strings.Contains(text, "routerprovider"):
		return "application shell and router provider"
	case strings.Contains(text, "createbrowserrouter("):
		return "route table and page composition"
	case strings.Contains(text, "express(") || strings.Contains(text, "fastify(") || strings.Contains(text, "app.listen("):
		return "server bootstrap"
	case strings.Contains(text, "export default function app") || strings.Contains(text, "function app("):
		return "top-level application shell"
	default:
		return "application module"
	}
}

func firstResolvedLocalJSImport(root, rel string) string {
	imports := orderedResolvedLocalJSImports(root, rel)
	for _, info := range imports {
		if len(info.Bindings) > 0 {
			return info.Resolved
		}
	}
	if len(imports) > 0 {
		return imports[0].Resolved
	}
	return ""
}

func resolveNodeAppShell(root, bootstrap string) string {
	data, err := os.ReadFile(filepath.Join(root, bootstrap))
	if err != nil {
		return ""
	}

	imports := orderedResolvedLocalJSImports(root, bootstrap)
	if len(imports) == 0 {
		return ""
	}

	bindings := map[string]string{}
	for _, info := range imports {
		for _, binding := range info.Bindings {
			if binding == "" {
				continue
			}
			if _, exists := bindings[binding]; !exists {
				bindings[binding] = info.Resolved
			}
		}
	}

	for _, target := range renderTargetIdentifiers(string(data)) {
		if resolved := bindings[target]; resolved != "" {
			return resolved
		}
	}

	for _, info := range imports {
		for _, binding := range info.Bindings {
			lower := strings.ToLower(binding)
			if lower == "app" || strings.HasSuffix(lower, "app") || strings.Contains(lower, "shell") || strings.Contains(lower, "root") {
				return info.Resolved
			}
		}
	}

	return firstResolvedLocalJSImport(root, bootstrap)
}

func firstRouterLikeJSImport(root, rel string) string {
	for _, info := range orderedResolvedLocalJSImports(root, rel) {
		if looksLikeRouterModule(root, info.Resolved) {
			return info.Resolved
		}
	}
	return ""
}

func orderedResolvedLocalJSImports(root, rel string) []jsImportInfo {
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		return nil
	}

	var infos []jsImportInfo
	for _, match := range localJSImportPattern.FindAllStringSubmatch(string(data), -1) {
		specifier := strings.TrimSpace(match[2])
		if !strings.HasPrefix(specifier, ".") {
			continue
		}
		resolved := resolveJSImport(root, filepath.Dir(rel), specifier)
		if resolved == "" {
			continue
		}
		infos = append(infos, jsImportInfo{
			Resolved: resolved,
			Bindings: parseJSImportBindings(match[1]),
		})
	}
	return infos
}

func parseJSImportBindings(clause string) []string {
	clause = strings.TrimSpace(clause)
	if clause == "" || strings.HasPrefix(clause, "type ") {
		return nil
	}

	var bindings []string
	if brace := strings.Index(clause, "{"); brace >= 0 {
		prefix := strings.TrimSpace(strings.TrimSuffix(clause[:brace], ","))
		if prefix != "" {
			if namespace := parseJSImportNamespace(prefix); namespace != "" {
				bindings = append(bindings, namespace)
			} else {
				bindings = append(bindings, prefix)
			}
		}
		if end := strings.LastIndex(clause, "}"); end > brace {
			for _, part := range strings.Split(clause[brace+1:end], ",") {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				if strings.Contains(part, " as ") {
					parts := strings.Split(part, " as ")
					part = strings.TrimSpace(parts[len(parts)-1])
				}
				if part != "" && !strings.HasPrefix(part, "type ") {
					bindings = append(bindings, part)
				}
			}
		}
		return bindings
	}

	if comma := strings.Index(clause, ","); comma >= 0 {
		defaultBinding := strings.TrimSpace(clause[:comma])
		if defaultBinding != "" {
			bindings = append(bindings, defaultBinding)
		}
		if namespace := parseJSImportNamespace(strings.TrimSpace(clause[comma+1:])); namespace != "" {
			bindings = append(bindings, namespace)
		}
		return bindings
	}

	if namespace := parseJSImportNamespace(clause); namespace != "" {
		return []string{namespace}
	}

	return []string{clause}
}

func parseJSImportNamespace(clause string) string {
	if !strings.HasPrefix(clause, "* as ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(clause, "* as "))
}

func renderTargetIdentifiers(source string) []string {
	seen := map[string]bool{}
	var ids []string
	appendID := func(identifier string) {
		if identifier == "" || seen[identifier] {
			return
		}
		seen[identifier] = true
		ids = append(ids, identifier)
	}

	for _, match := range renderJSXComponentPattern.FindAllStringSubmatch(source, -1) {
		appendID(match[1])
	}
	for _, match := range renderIdentifierPattern.FindAllStringSubmatch(source, -1) {
		identifier := match[1]
		if target := resolveJSXAliasTarget(source, identifier); target != "" {
			appendID(target)
			continue
		}
		appendID(identifier)
	}
	return ids
}

func resolveJSXAliasTarget(source, identifier string) string {
	if identifier == "" {
		return ""
	}
	pattern := regexp.MustCompile(fmt.Sprintf(`(?m)^\s*(?:const|let|var)\s+%s\s*=\s*<([A-Z][A-Za-z0-9_]*)\b`, regexp.QuoteMeta(identifier)))
	match := pattern.FindStringSubmatch(source)
	if len(match) == 2 {
		return match[1]
	}
	return ""
}

func resolveJSImport(root, fromDir, specifier string) string {
	base := filepath.Clean(filepath.Join(fromDir, specifier))
	candidates := []string{base}
	for _, ext := range []string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"} {
		candidates = append(candidates, base+ext)
		candidates = append(candidates, filepath.Join(base, "index"+ext))
	}

	for _, candidate := range candidates {
		if !exists(filepath.Join(root, candidate)) {
			continue
		}
		switch strings.ToLower(filepath.Ext(candidate)) {
		case ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs":
			return filepath.ToSlash(candidate)
		}
	}
	return ""
}

func looksLikeRouterModule(root, rel string) bool {
	if rel == "" {
		return false
	}
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		return false
	}
	text := strings.ToLower(string(data))
	return strings.Contains(text, "createbrowserrouter(") || strings.Contains(text, "routerprovider") || strings.Contains(text, "routeobject")
}

func firstFileContaining(root, needle string) string {
	match := ""
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if shouldSkipDiscoveryDir(root, path) {
				return filepath.SkipDir
			}
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs":
		default:
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		if strings.Contains(string(data), needle) {
			rel, relErr := filepath.Rel(root, path)
			if relErr == nil {
				match = filepath.ToSlash(rel)
				return io.EOF
			}
		}
		return nil
	})
	return match
}

func shouldSkipDiscoveryDir(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		return false
	}
	for _, segment := range strings.Split(rel, "/") {
		switch segment {
		case ".git", ".namba", ".codex", ".agents", "dist", "external", "node_modules", "vendor":
			return true
		}
	}
	return false
}

func loadPackageManifest(root string) (packageManifest, error) {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return packageManifest{}, err
	}
	var pkg packageManifest
	if err := json.Unmarshal(data, &pkg); err != nil {
		return packageManifest{}, err
	}
	return pkg, nil
}

func collectPackageVersions(pkg packageManifest, names ...string) []string {
	var labels []string
	for _, name := range names {
		if version := packageVersion(pkg, name); version != "" {
			labels = append(labels, fmt.Sprintf("%s@%s", name, version))
		}
	}
	return labels
}

func packageVersion(pkg packageManifest, name string) string {
	for _, group := range []map[string]string{pkg.Dependencies, pkg.DevDependencies, pkg.PeerDependencies} {
		if version, ok := group[name]; ok {
			return version
		}
	}
	return ""
}

func countPackagesWithPrefix(pkg packageManifest, prefix string) int {
	seen := map[string]bool{}
	for _, group := range []map[string]string{pkg.Dependencies, pkg.DevDependencies, pkg.PeerDependencies} {
		for name := range group {
			if strings.HasPrefix(name, prefix) {
				seen[name] = true
			}
		}
	}
	return len(seen)
}

func topPackageVersions(pkg packageManifest, limit int) []string {
	seen := map[string]bool{}
	var names []string
	for _, group := range []map[string]string{pkg.Dependencies, pkg.DevDependencies, pkg.PeerDependencies} {
		for name := range group {
			if seen[name] {
				continue
			}
			seen[name] = true
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if len(names) > limit {
		names = names[:limit]
	}

	var labels []string
	for _, name := range names {
		labels = append(labels, fmt.Sprintf("%s@%s", name, packageVersion(pkg, name)))
	}
	return labels
}

func readGoModule(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "module "))
		}
	}
	return ""
}
