package namba

import (
	"os"
	"path/filepath"
	"strings"
)

func extractAcceptanceTasks(text string) []string {
	var tasks []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- [ ]") {
			tasks = append(tasks, strings.TrimSpace(strings.TrimPrefix(line, "- [ ]")))
		}
	}
	return tasks
}

func chunkTasks(tasks []string, workers int) [][]string {
	if workers <= 1 || len(tasks) <= 1 {
		return [][]string{tasks}
	}
	chunks := make([][]string, workers)
	for i, task := range tasks {
		idx := i % workers
		chunks[idx] = append(chunks[idx], task)
	}
	var filtered [][]string
	for _, chunk := range chunks {
		if len(chunk) > 0 {
			filtered = append(filtered, chunk)
		}
	}
	return filtered
}

func countDirectories(root, prefix string) int {
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			count++
		}
	}
	return count
}

func firstNonEmptyLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func containsValue(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func normalizeProjectName(name string) string {
	name = strings.TrimSpace(name)
	if name == "." || name == string(filepath.Separator) {
		return ""
	}
	return name
}

func normalizeFramework(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "none"
	}
	return value
}
