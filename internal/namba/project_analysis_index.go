package namba

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type analysisIndex struct {
	root       string
	files      []analysisFile
	byPath     map[string]analysisFile
	textCache  map[string]string
	textLoaded map[string]bool
}

func buildAnalysisIndex(root string, files []analysisFile) analysisIndex {
	byPath := make(map[string]analysisFile, len(files))
	for _, file := range files {
		byPath[file.Path] = file
	}
	return analysisIndex{
		root:       root,
		files:      files,
		byPath:     byPath,
		textCache:  map[string]string{},
		textLoaded: map[string]bool{},
	}
}

func (idx analysisIndex) has(rel string) bool {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	if rel == "" {
		return false
	}
	_, ok := idx.byPath[rel]
	return ok
}

func (idx analysisIndex) firstExisting(candidates ...string) string {
	for _, candidate := range candidates {
		rel := filepath.ToSlash(strings.TrimSpace(candidate))
		if rel == "" {
			continue
		}
		if idx.has(rel) {
			return rel
		}
	}
	return ""
}

func (idx *analysisIndex) readText(rel string) (string, bool) {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	if rel == "" {
		return "", false
	}
	if idx.textLoaded[rel] {
		text, ok := idx.textCache[rel]
		return text, ok
	}

	path := filepath.Join(idx.root, filepath.FromSlash(rel))
	data, err := os.ReadFile(path)
	idx.textLoaded[rel] = true
	if err != nil {
		idx.textCache[rel] = ""
		return "", false
	}
	text := string(data)
	idx.textCache[rel] = text
	return text, true
}

func analysisReadSystemSummary(index analysisIndex, systemRoot string) (string, []string) {
	for _, candidate := range []string{"README.md", "README.txt"} {
		path := candidate
		if systemRoot != "." {
			path = filepath.ToSlash(filepath.Join(systemRoot, candidate))
		}
		if !index.has(path) {
			continue
		}
		if data, ok := index.readText(path); ok {
			if summary := docSummary(data); summary != "" {
				return summary, []string{path}
			}
		}
	}
	return "", nil
}

func analysisRelativePath(systemRoot, path string) string {
	path = filepath.ToSlash(strings.TrimSpace(path))
	systemRoot = filepath.ToSlash(strings.TrimSpace(systemRoot))
	if path == "" || systemRoot == "" || systemRoot == "." {
		return path
	}
	prefix := systemRoot + "/"
	if path == systemRoot {
		return "."
	}
	if strings.HasPrefix(path, prefix) {
		return strings.TrimPrefix(path, prefix)
	}
	return path
}

func analysisSystemEntryPointPaths(root, systemRoot, kind string, files []analysisFile, index analysisIndex) []string {
	var paths []string
	for _, bullet := range analysisBuildEntryPointBullets(root, systemRoot, kind, files, index) {
		rel, _ := parseBulletFinding(bullet)
		if rel != "" {
			paths = append(paths, prefixSystemPath(systemRoot, rel))
		}
	}
	return paths
}

func analysisBuildEntryPointBullets(root, systemRoot, kind string, files []analysisFile, index analysisIndex) []string {
	switch kind {
	case "frontend":
		return analysisBuildNodeEntryPoints(root, systemRoot, files, index)
	case "backend", "go-service":
		return analysisBuildGoEntryPoints(systemRoot, files)
	case "python-service":
		return analysisBuildPythonEntryPoints(systemRoot, files)
	default:
		if index.has(filepath.ToSlash(filepath.Join(systemRoot, "package.json"))) {
			return analysisBuildNodeEntryPoints(root, systemRoot, files, index)
		}
	}
	return nil
}

func analysisBuildGoEntryPoints(systemRoot string, files []analysisFile) []string {
	seen := map[string]bool{}
	var bullets []string
	rootMain := filepath.ToSlash(filepath.Join(systemRoot, "main.go"))
	for _, file := range files {
		if file.Path == rootMain {
			appendEntryPoint(&bullets, seen, analysisRelativePath(systemRoot, file.Path), "application bootstrap")
			break
		}
	}

	cmdPrefix := filepath.ToSlash(filepath.Join(systemRoot, "cmd")) + "/"
	var candidates []string
	for _, file := range files {
		if !strings.HasPrefix(file.Path, cmdPrefix) || filepath.Base(file.Path) != "main.go" {
			continue
		}
		relDir := filepath.Dir(file.Path)
		if relDir == filepath.ToSlash(filepath.Join(systemRoot, "cmd")) {
			continue
		}
		candidates = append(candidates, file.Path)
	}
	sort.Strings(candidates)
	for _, candidate := range candidates {
		appendEntryPoint(&bullets, seen, analysisRelativePath(systemRoot, candidate), "Go command entry point")
	}
	return bullets
}

func analysisBuildPythonEntryPoints(systemRoot string, files []analysisFile) []string {
	seen := map[string]bool{}
	var bullets []string
	for _, candidate := range []string{"main.py", "app.py", "manage.py", "src/main.py"} {
		rel := filepath.ToSlash(filepath.Join(systemRoot, candidate))
		for _, file := range files {
			if file.Path == rel {
				appendEntryPoint(&bullets, seen, analysisRelativePath(systemRoot, rel), "Python application entry point")
				break
			}
		}
	}
	return bullets
}

func analysisBuildNodeEntryPoints(root, systemRoot string, files []analysisFile, index analysisIndex) []string {
	seen := map[string]bool{}
	var bullets []string

	bootstrap := ""
	for _, candidate := range []string{"src/main.tsx", "src/main.jsx", "src/index.tsx", "src/index.jsx", "main.tsx", "main.jsx", "src/main.ts", "src/main.js"} {
		rel := filepath.ToSlash(filepath.Join(systemRoot, candidate))
		if index.has(rel) {
			bootstrap = rel
			break
		}
	}
	if bootstrap == "" {
		bootstrap = analysisFirstFileContaining(index, files, "createRoot(")
	}
	if bootstrap != "" {
		appendEntryPoint(&bullets, seen, analysisRelativePath(systemRoot, bootstrap), analysisSummarizeEntryPoint(index, bootstrap))
		if appShell := analysisResolveNodeAppShell(root, bootstrap, index); appShell != "" {
			appendEntryPoint(&bullets, seen, analysisRelativePath(systemRoot, appShell), analysisSummarizeEntryPoint(index, appShell))
			if routerModule := analysisFirstRouterLikeJSImport(root, appShell, index); routerModule != "" {
				appendEntryPoint(&bullets, seen, analysisRelativePath(systemRoot, routerModule), analysisSummarizeEntryPoint(index, routerModule))
			}
		}
	}

	for _, candidate := range []string{"src/app/routes.ts", "src/app/routes.tsx", "src/routes.ts", "src/routes.tsx", "src/router.ts", "src/router.tsx"} {
		rel := filepath.ToSlash(filepath.Join(systemRoot, candidate))
		if index.has(rel) {
			appendEntryPoint(&bullets, seen, analysisRelativePath(systemRoot, rel), analysisSummarizeEntryPoint(index, rel))
		}
	}

	if routerModule := analysisFirstFileContaining(index, files, "createBrowserRouter("); routerModule != "" {
		appendEntryPoint(&bullets, seen, analysisRelativePath(systemRoot, routerModule), analysisSummarizeEntryPoint(index, routerModule))
	}
	return bullets
}

func analysisSystemDependencies(index analysisIndex, system analysisSystem) []string {
	switch system.Kind {
	case "frontend":
		return analysisBuildNodeDependencies(system.Root, index)
	case "backend", "go-service":
		return analysisBuildGoDependencies(system.Root, index)
	case "python-service":
		return analysisBuildPythonDependencies(system.Root, index)
	default:
		if index.has(filepath.ToSlash(filepath.Join(system.Root, "package.json"))) {
			return analysisBuildNodeDependencies(system.Root, index)
		}
		if index.has(filepath.ToSlash(filepath.Join(system.Root, "go.mod"))) {
			return analysisBuildGoDependencies(system.Root, index)
		}
		return nil
	}
}

func analysisBuildNodeDependencies(systemRoot string, index analysisIndex) []string {
	pkg, err := analysisLoadPackageManifest(filepath.ToSlash(filepath.Join(systemRoot, "package.json")), index)
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

func analysisBuildGoDependencies(systemRoot string, index analysisIndex) []string {
	if module := analysisReadGoModule(filepath.ToSlash(filepath.Join(systemRoot, "go.mod")), index); module != "" {
		return []string{
			fmt.Sprintf("- Module: `%s`", module),
			"- Runtime: Go standard library",
			"- External runtime: Codex CLI",
			"- External runtime: Git",
		}
	}
	return nil
}

func analysisBuildPythonDependencies(systemRoot string, index analysisIndex) []string {
	var bullets []string
	if index.has(filepath.ToSlash(filepath.Join(systemRoot, "pyproject.toml"))) {
		bullets = append(bullets, "- Dependency manifest: `pyproject.toml`")
	}
	if index.has(filepath.ToSlash(filepath.Join(systemRoot, "requirements.txt"))) {
		bullets = append(bullets, "- Dependency manifest: `requirements.txt`")
	}
	return bullets
}

func analysisLoadPackageManifest(rel string, index analysisIndex) (packageManifest, error) {
	data, ok := index.readText(rel)
	if !ok {
		return packageManifest{}, fmt.Errorf("package manifest not found")
	}
	var pkg packageManifest
	if err := json.Unmarshal([]byte(data), &pkg); err != nil {
		return packageManifest{}, err
	}
	return pkg, nil
}

func analysisReadGoModule(rel string, index analysisIndex) string {
	data, ok := index.readText(rel)
	if !ok {
		return ""
	}
	for _, line := range strings.Split(data, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "module "))
		}
	}
	return ""
}

func analysisSummarizeEntryPoint(index analysisIndex, rel string) string {
	data, ok := index.readText(rel)
	if !ok {
		return "application module"
	}
	text := strings.ToLower(data)
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

func analysisResolveNodeAppShell(root, bootstrap string, index analysisIndex) string {
	data, ok := index.readText(bootstrap)
	if !ok {
		return ""
	}

	imports := analysisOrderedResolvedLocalJSImports(root, bootstrap, index)
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

	for _, target := range renderTargetIdentifiers(data) {
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

	return analysisFirstResolvedLocalJSImport(root, bootstrap, index)
}

func analysisFirstResolvedLocalJSImport(root, rel string, index analysisIndex) string {
	imports := analysisOrderedResolvedLocalJSImports(root, rel, index)
	if len(imports) == 0 {
		return ""
	}
	return imports[0].Resolved
}

type analysisJSImportInfo struct {
	Resolved string
	Bindings []string
}

func analysisOrderedResolvedLocalJSImports(root, rel string, index analysisIndex) []analysisJSImportInfo {
	data, ok := index.readText(rel)
	if !ok {
		return nil
	}

	var infos []analysisJSImportInfo
	for _, match := range localJSImportPattern.FindAllStringSubmatch(data, -1) {
		specifier := strings.TrimSpace(match[2])
		if !strings.HasPrefix(specifier, ".") {
			continue
		}
		resolved := analysisResolveJSImport(root, filepath.Dir(rel), specifier, index)
		if resolved == "" {
			continue
		}
		infos = append(infos, analysisJSImportInfo{
			Resolved: resolved,
			Bindings: parseJSImportBindings(match[1]),
		})
	}
	return infos
}

func analysisResolveJSImport(root, fromDir, specifier string, index analysisIndex) string {
	base := filepath.Clean(filepath.Join(fromDir, specifier))
	candidates := []string{base}
	for _, ext := range []string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"} {
		candidates = append(candidates, base+ext)
		candidates = append(candidates, filepath.Join(base, "index"+ext))
	}

	for _, candidate := range candidates {
		rel := filepath.ToSlash(candidate)
		if !index.has(rel) {
			continue
		}
		switch strings.ToLower(filepath.Ext(candidate)) {
		case ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs":
			return rel
		}
	}
	return ""
}

func analysisFirstRouterLikeJSImport(root, rel string, index analysisIndex) string {
	for _, info := range analysisOrderedResolvedLocalJSImports(root, rel, index) {
		if analysisLooksLikeRouterModule(info.Resolved, index) {
			return info.Resolved
		}
	}
	return ""
}

func analysisLooksLikeRouterModule(rel string, index analysisIndex) bool {
	if rel == "" {
		return false
	}
	data, ok := index.readText(rel)
	if !ok {
		return false
	}
	text := strings.ToLower(data)
	return strings.Contains(text, "createbrowserrouter(") || strings.Contains(text, "routerprovider") || strings.Contains(text, "routeobject")
}

func analysisFirstFileContaining(index analysisIndex, files []analysisFile, needle string) string {
	for _, file := range files {
		if analysisShouldSkipDiscoveryPath(file.Path) {
			continue
		}
		switch strings.ToLower(filepath.Ext(file.Path)) {
		case ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs":
		default:
			continue
		}
		data, ok := index.readText(file.Path)
		if !ok {
			continue
		}
		if strings.Contains(data, needle) {
			return file.Path
		}
	}
	return ""
}

func analysisShouldSkipDiscoveryPath(rel string) bool {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	if rel == "" || rel == "." {
		return false
	}
	for _, segment := range strings.Split(rel, "/") {
		switch segment {
		case ".git", ".namba", ".codex", ".agents", "dist", "external", "node_modules":
			return true
		}
	}
	return false
}
